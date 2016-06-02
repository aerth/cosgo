// cosgo is an easy to use contact form *server*, able to be iframed on a static web site.
package main

/*
                           _
  ___ ___  ___  __ _  ___ | |
 / __/ _ \/ __|/ _` |/ _ \| |
| (_| (_) \__ \ (_| | (_) |_|
 \___\___/|___/\__, |\___/(_)
               |___/

https://github.com/aerth/cosgo

Contact form server. Saves to local mailbox or uses Mandrill or Sendgrid SMTP API.

*/

// The MIT License (MIT)
//
// Copyright (c) 2016 aerth
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"path"
	"path/filepath"

	"github.com/aerth/cosgo/mbox"
	"github.com/dchest/captcha"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	sl "github.com/hydrogen18/stoppableListener"
)

var (
	mandrillKey      string
	sendgridKey      string
	cosgoDestination string
	antiCSRFkey      []byte
	cosgoRefresh     = 42 * time.Minute
	err              error
	hitcounter       int
	boottime         time.Time
	// Version is modified by cgo linker when compiled using "make"
	Version   = "0.8-go-get"
	publicKey []byte
)

// Cosgo - This changes every [cosgoRefresh] minutes
type Cosgo struct {
	PostKey   string
	Boottime  time.Time
	Static    string
	Templates string // Directory where templates are located. Defaults to ./templates and falls back to /usr/local/share/cosgo/templates
	Dir       string // Directory where we are currently located (./)

}

var cosgo = new(Cosgo)

type C struct {
	CaptchaID string
}

/* With these settings, the captcha string will be from 3 to 5 characters. */
const (
	// CaptchaLength is the minimum captcha string length.
	CaptchaLength = 3
	// CaptchaVariation will add *up to* CaptchaVariation to the CaptchaLength
	CaptchaVariation = 2
	// CollectNum triggers a garbage collection routine after X captchas are created.
	CollectNum = 100
	Expiration = 10 * time.Minute
	StdWidth   = 240
	StdHeight  = 90
)

//usage shows how and available flags.
func usage() {
	fmt.Println("\nusage: cosgo [flags]")
	fmt.Println("\nflags:")
	//time.Sleep(1000 * time.Millisecond)
	flag.PrintDefaults()
	//time.Sleep(1000 * time.Millisecond)
	fmt.Println("\nExample: cosgo -secure -port 8080 -fastcgi -debug")
}

var (
	mandrillAPIUrl = "https://mandrillapp.com/api/1.0/"
)
var (
	mailform *mbox.Form

	// ErrNoReferer is returned when a HTTPS request provides an empty Referer
	// header.
	ErrNoReferer = errors.New("referer not supplied")
	// ErrBadReferer is returned when the scheme & host in the URL do not match
	// the supplied Referer header.
	ErrBadReferer = errors.New("referer invalid")
	// ErrNoToken is returned if no CSRF token is supplied in the request.
	ErrNoToken = errors.New("CSRF token not found in request")
	// ErrBadToken is returned if the CSRF token in the request does not match
	// the token in the session, or is otherwise malformed.
	ErrBadToken = errors.New("CSRF token invalid, yo")
)

const (
	localmail    = "mailbox"
	smtpmandrill = "mandrill"
	smtpsendgrid = "sendgrid"
)

var (
	// TODO: dont use flags. Will be using env/seconf only. seconf is an environmental variable based configuration system. Try cosgo -config
	help         = flag.Bool("help", false, "Show this and quit")
	port         = flag.String("port", "8080", "Port to listen on")
	bind         = flag.String("bind", "0.0.0.0", "Default: 0.0.0.0 (all interfaces)...\nTry -bind=127.0.0.1")
	debug        = flag.Bool("debug", false, "Send logs to stdout. \nDont switch to cosgo.log")
	secure       = flag.Bool("secure", false, "Accept only https secure cookie transfer.")
	mailmode     = flag.String("mailmode", localmail, "Choose one: mailbox, mandrill, sendgrid	\nExample: -mailmode=mailbox")
	fastcgi      = flag.Bool("fastcgi", false, "Use fastcgi (for with nginx etc)")
	static       = flag.Bool("static", true, "Serve /static/ files. Use -static=false to disable")
	files        = flag.Bool("files", true, "Serve /files/ files. Use -files=false to disable")
	noredirect   = flag.Bool("noredirect", false, "Default is to redirect all 404s back to /. \nSet true to enable error.html template")
	config       = flag.Bool("config", false, "Use config file at ~/.cosgo")
	custom       = flag.String("custom", "default", "Example: cosgo2 ...creates $HOME/.cosgo2")
	logpath      = flag.String("log", "cosgo.log", "Example: /dev/null or /var/log/cosgo/log")
	quiet        = flag.Bool("quiet", false, "No output to stdout. \nFor use with cron and -log flag such as: cosgo -quiet -log=/dev/null or cosgo -quiet -log=/var/log/cosgo/log")
	nolog        = flag.Bool("nolog", false, "No Output Whatsoever")
	form         = flag.Bool("form", false, "Use /page/index.html and /templates/form.html, \nsets -pages flag automatically.")
	pages        = flag.Bool("pages", false, "Serve /page/")
	sendgridflag = flag.Bool("sendgrid", false, "Set -sendgrid to not use local mailbox. This automatically sets \"-mailmode sendgrid\"")
	resolvemail  = flag.Bool("resolvemail", false, "Set true to check email addresses (Outgoing traffic)")
	custompages  = flag.String("custompages", "page", "Serve pages from X dir")
	gpg          = flag.String("gpg", "", "Path to gpg Public Key (automatically encrypts messages)")
	mailbox      = true
)

/*













 */

func main() {

	// Hopefully a clean exit
	interrupt := make(chan os.Signal, 1)
	stop := make(chan os.Signal, 1)
	reload := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	signal.Notify(stop, syscall.SIGINT)
	signal.Notify(reload, syscall.SIGHUP)
	go func() {
		select {
		case signal := <-interrupt:
			fmt.Println("Got signal:", signal)
			fmt.Println("Dying")
			os.Exit(0)
		case signal := <-reload:
			fmt.Println("Got signal:", signal)
			fmt.Println("Dying")
			os.Exit(0)
		case signal := <-stop:
			fmt.Printf("Got signal:%v\n", signal)
			fmt.Println("Dying")
			os.Exit(0)
		}
	}()

	// Future: dont use flags pkg
	flag.Parse()
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}
	cosgo.Dir = cwd + "/"
	// Custom refresh time
	// Example: COSGOREFRESH=10m cosgo -debug
	if os.Getenv("COSGOREFRESH") != "" {
		cosgoRefresh, err = time.ParseDuration(os.Getenv("COSGOREFRESH"))
		if err != nil {
			log.Println(err.Error() + "... Using default (42m).")
			cosgoRefresh = 42 * time.Minute
		}
	}

	if runtime.GOOS == "windows" && flag.NFlag() == 0 && flag.NArg() == 0 {
		fmt.Println("cosgo is running on Windows with no flags. \nThis may work, but *please* run from cmd.exe!\nClick Start -> Run, Type: cmd.exe\n\nIf this is your first run, you will see a Windows Firewall request in a moment.\nCheck all the boxes and click \"Allow access\" for typical usage. \nSince no flags were set, the default port is 8080. \nSo open up firefox and visit http://localhost:8080 !!!\n\n\tThank you for using cosgo!\n\t-aerth")
		time.Sleep(5 * time.Second)
	}

	// nolog flag = no output whatsoever. Equivalent of -quiet -logpath=/dev/null
	if *nolog {
		*quiet = true
		if runtime.GOOS == "windows" { // fix for windows soon
			fmt.Println("You chose -nolog on Windows. Experimental results!")
		}
		*logpath = os.DevNull

	}

	if *custompages != "page" {
		*pages = true
	}

	if *custom != "default" {
		*config = true
	}

	if *custom == "default" {
		*custom = "cosgorc"
	}
	if *sendgridflag {
		*mailmode = "sendgrid"
	}
	if *mailmode != localmail {
		mailbox = false
	}

	if *resolvemail {
		mbox.ValidationLevel = 2
	}

	// -custom="anything" sets -config=true
	if *custom != "cosgorc" && *config == false {
		*config = true
	}

	// Load Configuration from seconf/secenv
	if *config == true {
		if *custom == "default" {
			*custom = "cosgorc"
		}

		//fmt.Printf("Loading cosgo config: %s.\n", *custom)

		if !loadConfig() {
			fmt.Println("Fatal: Can't load configuration file.")
			os.Exit(1)
		}

		// else {
		// 	//fmt.Printf("done.")
		// }
	}

	if !*quiet {
		fmt.Println(logo)
		fmt.Printf("\n\tcosgo v" + Version + "\n\tCopyright 2016 aerth\n\tSource code at https://github.com/aerth/cosgo\n")

	}

	if !*quiet {
		func() {
			opts := ""
			opts = opts + "[" + *bind + ":" + *port + "]"
			if *files == true {
				opts = opts + "[files]"
			}
			if *static == true {
				opts = opts + "[static]"
			}

			if *fastcgi == true {
				opts = opts + "[fastcgi]"
			}

			opts = opts + "[" + *mailmode + "]"

			if *pages == true {
				opts = opts + "[pages]"
			}
			if *secure == true {
				opts = opts + "[https]"
			} else {
				opts = opts + "[no https]"
			}
			if *custompages != "page" && *pages == true {
				opts = opts + "[" + *custompages + "]"
			}
			opts = opts + "[refresh: " + cosgoRefresh.String() + "]"

			fmt.Println("\t" + opts)
		}()
		if !*nolog {
			fmt.Println("\t[logs: " + *logpath + "]")
		}
		if *gpg != "" {
			fmt.Println("\t[gpg pubkey: " + *gpg + "]")
			publicKey = read2mem(rel2real(*gpg))
		}
		if *config {
			fmt.Println("\t[config: " + *custom + "]")
		}
		if *debug == true {
			fmt.Println("\t[debug: on]")
		}

		fmt.Printf("\n\n")
	}

	args := flag.Args() // arguments that arent -flags
	if len(args) > 1 {
		usage()
		os.Exit(1)
	}
	if len(os.Args) > 1 {
		if os.Args[1] == "config" {
			if detectConfig() {
				fmt.Println("Config file already exists, use -config to read it.")
				os.Exit(1)
			}
			loadConfig()
			os.Exit(1)
		}
		if os.Args[1] == "help" {
			usage()
			os.Exit(1)
		}
	}

	if os.Getenv("COSGOPAGEDIR") != "" {
		*custompages = os.Getenv("COSGOPAGEDIR")
	}

	// // If user is still using CASGO_DESTINATION or CASGO_API_KEY (instead of COSGO)
	// backwardsComp()

	// Define antiCSRFkey with env var, or set default.
	if !*config {
		if os.Getenv("COSGO_CSRF_KEY") == "" && string(antiCSRFkey) == "" {
			//	log.Println("You can now set COSGO_CSRF_KEY environmental variable. Using default.")
			antiCSRFkey = []byte("LI80PNK1xcT01jmQBsEyxyrNCrbyyFPjPU8CKnxwmCruxNijgnyb3hXXD3p1RBc0+LIRQUUbTtis6hc6LD4I/A==")
		} else {
			log.Println("CSRF key OK", os.Getenv("COSGO_CSRF_KEY"))
			antiCSRFkey = []byte(os.Getenv("COSGO_CSRF_KEY"))
		}
	}
	// Check templates and variables
	if err = quickSelfTest(); err != nil {
		log.Fatal(err.Error())
	}

	if os.Getenv("COSGO_API_KEY") == "" {

		// The length of the POST key can be modified here.

		go func() {
			for {
				if *debug && !*quiet {
					log.Println("Info: Generating Random POST Key...")
				}
				cosgo.PostKey = generateAPIKey(40)
				if *debug && !*quiet {
					log.Printf("Info: POST Key is " + cosgo.PostKey + "\n")
				}
				time.Sleep(cosgoRefresh) // Internal cron!!!
			}
		}()
	} else {
		cosgo.PostKey = os.Getenv("COSGO_API_KEY")
		// Print selected COSGO_API_KEY
		log.Println("COSGO_API_KEY:", getKey())
	}

	//Begin Routing
	r := mux.NewRouter()

	//Redirect or show custom error?
	if *noredirect == false {
		r.NotFoundHandler = http.HandlerFunc(redirecthomeHandler)
	} else {
		r.NotFoundHandler = http.HandlerFunc(customErrorHandler)
	}
	r.HandleFunc("/", homeHandler)

	// POST endpoint (emailHandler checks the key)
	r.HandleFunc("/{{whatever}}/send", emailHandler)

	// TODO: maybe allow changing static and files directory names
	// TODO: mime types

	//s := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	s := http.StripPrefix("/static/", http.FileServer(http.Dir(cosgo.Static)))
	ss := http.FileServer(http.Dir(cosgo.Static))
	sf := http.FileServer(http.Dir("./files/"))
	sp := http.StripPrefix("/"+*custompages+"/", http.FileServer(http.Dir("./"+*custompages+"/")))

	// *static defaults to true. We are serving out of /static/ for now
	if *static == true {

		// Serve /static folder and favicon etc
		r.Methods("GET").Path("/favicon.ico").Handler(ss)
		r.Methods("GET").Path("/robots.txt").Handler(ss)
		r.Methods("GET").Path("/sitemap.xml").Handler(ss)
		r.Methods("GET").Path("/static/{dir}/{whatever}.css").Handler(s)
		r.Methods("GET").Path("/static/{dir}/{whatever}.js").Handler(s)
		r.Methods("GET").Path("/static/{dir}/{whatever}.png").Handler(s)
		r.Methods("GET").Path("/static/{dir}/{whatever}.jpg").Handler(s)
		r.Methods("GET").Path("/static/{dir}/{whatever}.jpeg").Handler(s)
		r.Methods("GET").Path("/static/{dir}/{whatever}.woff").Handler(s)
		r.Methods("GET").Path("/static/{dir}/{whatever}.ttf").Handler(s)
		r.Methods("GET").Path("/static/{dir}/{whatever}.txt").Handler(s)
		r.Methods("GET").Path("/static/{dir}/{whatever}.mp3").Handler(s)
		r.Methods("GET").Path("/static/{dir}/{whatever}.m3u").Handler(s)
		r.Methods("GET").Path("/static/{dir}/{whatever}.md").Handler(s)

	}
	if *static == true && *files == true {
		r.Methods("GET").Path("/files/{whatever}.tgz").Handler(sf)
		r.Methods("GET").Path("/files/{whatever}.tar").Handler(sf)
		r.Methods("GET").Path("/files/{whatever}.zip").Handler(sf)
		r.Methods("GET").Path("/files/{whatever}.tar.gz").Handler(sf)
	}
	if *static == true && *pages == true { // No /page/index.html
		r.Methods("GET").Path("/" + *custompages + "/{whatever}.html").Handler(sp)
		r.Methods("GET").Path("/" + *custompages + "/{dir}/{whatever}.html").Handler(sp)
	}

	if *pages == true {
		r.HandleFunc("/"+*custompages+"{whatever}", pageHandler)
	}

	// Retrieve Captcha IMG and WAV
	r.Methods("GET").Path("/captcha/{captchacode}.png").Handler(captcha.Server(StdWidth, StdHeight))
	r.Methods("GET").Path("/captcha/download/{captchacode}.wav").Handler(captcha.Server(StdWidth, StdHeight))
	r.Methods("GET").Path("/captcha/{captchacode}.wav").Handler(captcha.Server(StdWidth, StdHeight))

	http.Handle("/", r)
	//End Routing

	// Start Runtime Info
	if *secure == false && !*quiet && *debug {
		log.Println("Warning: Running in *insecure* mode.")

	}
	if mailbox == true {

		f, err := os.OpenFile("./cosgo.mbox", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			log.Printf("error opening file: %v", err)
			log.Fatal("Hint: touch ./cosgo.mbox, or chown/chmod it so that the cosgo process can access it.")
			os.Exit(1)
		}
		mbox.Destination = cosgoDestination
		mbox.Mail = log.New(f, "", 0)

	}
	// Display right-clickable link
	if *debug && !*quiet {
		log.Printf("Link: " + getLink(*fastcgi, *bind, *port))
	}

	// Switch to cosgo.log if not debug
	if *debug == false {
		if !*quiet && *debug {
			log.Println("[logs at " + *logpath + "]")
		}
		openLogFile()

	}

	oglistener, binderr := net.Listen("tcp", *bind+":"+*port)
	if binderr != nil {
		log.Println(binderr)
		os.Exit(1)
	}
	listener, stoperr := sl.New(oglistener)
	if stoperr != nil {
		log.Println(stoperr)
		os.Exit(1)
	}
	if *debug && !*quiet {
		log.Printf("Info: Got listener %s %s", listener.Addr().String(), listener.Addr().Network())
	}
	go func() {

		boottime := time.Now()
		cosgo.Boottime = boottime

	}()
	// Start Serving Loop
	for {
		listener, err = sl.New(oglistener)
		if err != nil {
			log.Fatalln(err)
		}

		// Start listening in a goroutine switch case on *fastcgi and *secure
		go func() {
			switch *fastcgi {
			case true:
				switch *secure {
				case false: // Fastcgi + http://
					if listener != nil {
						go fcgi.Serve(listener,
							csrf.Protect(antiCSRFkey,
								csrf.HttpOnly(true),
								csrf.FieldName("cosgo"),
								csrf.CookieName("cosgo"),
								csrf.Secure(false))(r))
					} else {
						log.Fatalln("nil listener")
					}
				case true: // Fastcgi + https
					if listener != nil {
						go fcgi.Serve(listener,
							csrf.Protect(antiCSRFkey,
								csrf.HttpOnly(true),
								csrf.FieldName("cosgo"),
								csrf.CookieName("cosgo"),
								csrf.Secure(true))(r))
					} else {
						log.Fatalln("nil listener")
					}
				}
			case false:
				switch *secure {
				case true: // https:// only, no fastcgi
					if listener != nil {
						go http.Serve(listener,
							csrf.Protect(antiCSRFkey,
								csrf.HttpOnly(true),
								csrf.FieldName("cosgo"),
								csrf.CookieName("cosgo"), csrf.Secure(true))(r))
					} else {
						log.Fatalln("nil listener")
					}
				case false: // no https, no fastcgi.
					if listener != nil {
						go http.Serve(listener,
							csrf.Protect(antiCSRFkey,
								csrf.HttpOnly(true),
								csrf.FieldName("cosgo"),
								csrf.CookieName("cosgo"),
								csrf.Secure(false))(r))
					} else {
						log.Fatalln("nil listener")
					}
				}
			}
		}()
		// End the for loop with a select loop with cosgoRefresh sleep timer.
		select {
		default:
			if !*quiet {
				log.Printf("Uptime: %s", time.Since(boottime))
				log.Printf("Hits: %s", strconv.Itoa(hitcounter))
			}
			time.Sleep(cosgoRefresh)
		}
	}
}

var logo = `
                           _
  ___ ___  ___  __ _  ___ | |
 / __/ _ \/ __|/ _  |/ _ \| |
| (_| (_) \__ \ (_| | (_) |_|
 \___\___/|___/\__, |\___/(_)
               |___/
`

// Hello functions
func getKey() string {
	return cosgo.PostKey
}
func getDestination() string {
	return cosgoDestination
}
func getMandrillKey() string {
	return mandrillKey
}

/*
// parseQuery sanitizes inputs and gets ready to save to mbox
func parseQuery(query url.Values) *Form {
	p := bluemonday.UGCPolicy()
	form := new(Form)
	additionalFields := ""
	for k, v := range query {
		k = strings.ToLower(k)
		if k == "email" {
			form.Email = v[0]
			//} else if (k == "name") {
			//	form.Name = v[0]
		} else if k == "subject" {
			form.Subject = v[0]
			form.Subject = p.Sanitize(form.Subject)
		} else if k == "message" {
			form.Message = k + ": " + v[0] + "<br>\n"
			form.Message = p.Sanitize(form.Message)
		} else if k != "cosgo" && k != "captchaid" && k != "captchasolution" {
			additionalFields = additionalFields + k + ": " + v[0] + "<br>\n"
		}
	}
	if form.Subject == "" {
		form.Subject = "[New Message]"
	}
	if additionalFields != "" {
		if form.Message == "" {
			form.Message = form.Message + "Message:\n<br>" + p.Sanitize(additionalFields)

		} else {
			form.Message = form.Message + "\n<br>Additional:\n<br>" + p.Sanitize(additionalFields)

		}
	}
	if publicKey != nil {
		log.Println("Got form. Encoding it.")
		tmpmsg, err := mbox.PGPEncode(form.Message, publicKey)
		if err != nil {
			log.Println("gpg error.")
			log.Println(err)
		} else {
			log.Println("No GPG error.")
			form.Message = tmpmsg
		}
	}
	return form
}*/

// getDomain returns the domain name (without port) of a request.
func getDomain(r *http.Request) string {
	type Domains map[string]http.Handler
	hostparts := strings.Split(r.Host, ":")
	requesthost := hostparts[0]
	return requesthost
}

// getSubdomain returns the subdomain (if any)
func getSubdomain(r *http.Request) string {
	type Subdomains map[string]http.Handler
	hostparts := strings.Split(r.Host, ":")
	requesthost := hostparts[0]
	if net.ParseIP(requesthost) == nil {
		//log.Println("Info: " + requesthost)
		domainParts := strings.Split(requesthost, ".")
		if len(domainParts) > 2 {
			if domainParts[0] != "127" {
				return domainParts[0]
			}
		}
	}
	return ""
}

// Key Generator
func init() {
	rand.Seed(time.Now().UnixNano())
}

var runes = []rune("____ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890123456789012345678901234567890")

//generateAPIKey does API Key Generation with the given runes.
func generateAPIKey(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = runes[rand.Intn(len(runes))]
	}
	return strings.TrimSpace(string(b))
}

func rel2real(file string) (realpath string) {
	pathdir, _ := path.Split(file)

	if pathdir == "" {
		realpath, _ = filepath.Abs(file)
	} else {
		realpath = file
	}
	return realpath
}

func read2mem(abspath string) []byte {
	file, err := os.Open(abspath) // For read access.
	if err != nil {
		log.Fatal(err)
	}

	data := make([]byte, 4096)
	_, err = file.Read(data)
	if err != nil {
		log.Fatal(err)
	}

	return data

}

/*
func pgpEncode(plain string, publicKey []byte) (encStr string, err error) {
	entitylist, err := openpgp.ReadArmoredKeyRing(bytes.NewBuffer(publicKey))
	if err != nil {
	return plain, err
	}

	// Encrypt message using public key
	buf := new(bytes.Buffer)
	w, err := openpgp.Encrypt(buf, entitylist, nil, nil, nil)
	if err != nil {
	return plain, err
	}
	_, err = w.Write([]byte(plain))
	if err != nil {
	return plain, err
	}
	err = w.Close()
	if err != nil {
	}

	// Output as base64 encoded string
	bytes, err := ioutil.ReadAll(buf)
	encStr = base64.StdEncoding.EncodeToString(bytes)

	return encStr, nil
}

*/

// Copyright 2016 aerth. All Rights Reserved.
// Full source code at https://github.com/aerth/cosgo
// There should be a copy of the MIT license bundled with this software.
