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
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aerth/cosgo/mbox"
	"github.com/dchest/captcha"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	sl "github.com/hydrogen18/stoppableListener"
	"github.com/microcosm-cc/bluemonday"
)

var (
	mandrillKey      string
	sendgridKey      string
	cosgoDestination string
	antiCSRFkey      []byte

	cosgoRefresh = 42 * time.Minute
	err          error
)

// Cosgo. This changes every [cosgoRefresh] minutes
type Cosgo struct {
	PostKey string
}

var cosgo = new(Cosgo)

type C struct {
	CaptchaID string
}

const (

	/* With these settings, the captcha string will be from 3 to 5 characters. */

	// CaptchaLength is the minimum captcha string length.
	CaptchaLength = 3
	// Captcha will add *up to* CaptchaVariation to the CaptchaLength
	CaptchaVariation = 2
	CollectNum       = 100
	Expiration       = 10 * time.Minute
	StdWidth         = 240
	StdHeight        = 90
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
	// TODO: dont use flags. Will be using "cosgo __action__" and env/seconf only.
	help        = flag.Bool("help", false, "Show this and quit")
	port        = flag.String("port", "8080", "Port to listen on")
	bind        = flag.String("bind", "0.0.0.0", "Default: 0.0.0.0 (all interfaces)... Try 127.0.0.1")
	debug       = flag.Bool("debug", false, "Send logs to stdout. Dont switch to cosgo.log")
	secure      = flag.Bool("secure", false, "PRODUCTION MODE - Accept only https secure cookie transfer.")
	mailmode    = flag.String("mailmode", localmail, "Choose one: mailbox, mandrill, sendgrid")
	fastcgi     = flag.Bool("fastcgi", false, "Use fastcgi (for with nginx etc)")
	static      = flag.Bool("static", true, "Serve /static/ files. Use -static=false to disable")
	noredirect  = flag.Bool("noredirect", false, "Default is to redirect all 404s back to /. Set true to enable error.html template")
	love        = flag.Bool("love", false, "Fun. Show I love ___ instead of redirect")
	config      = flag.Bool("config", false, "Use config file at ~/.cosgo")
	custom      = flag.String("custom", "default", "Example: cosgo2 ...creates $HOME/.cosgo2")
	logpath     = flag.String("log", "cosgo.log", "Example: /dev/null or /var/log/cosgo/log")
	quiet       = flag.Bool("quiet", false, "No output to stdout. For use with cron and -log flag such as: cosgo -quiet -log=/dev/null or cosgo -quiet -log=/var/log/cosgo/log")
	nolog       = flag.Bool("nolog", false, "No Output Whatsoever")
	pages       = flag.Bool("pages", true, "Serve /pages/")
	custompages = flag.String("custompages", "page", "Serve pages from X dir")
	mailbox     = true
)

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
		} else {
			additionalFields = additionalFields + k + ": " + v[0] + "<br>\n"
		}
	}
	if form.Subject == "" {
		form.Subject = "[New Message]"
	}
	if additionalFields != "" {
		/*if form.Message == "" {
			//form.Message = form.Message + "Message:\n<br>" + additionalFields
			form.Message = form.Message
		} else {
			//form.Message = form.Message + "\n<br>Additional:\n<br>" + additionalFields
			form.Message = form.Message
		}*/
	}
	return form
}
func getDomain(r *http.Request) string {
	type Domains map[string]http.Handler
	hostparts := strings.Split(r.Host, ":")
	requesthost := hostparts[0]
	return requesthost
}
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

/*













 */

func main() {

	// Yay for signal handling. For now just quit.
	interrupt := make(chan os.Signal, 1)
	stop := make(chan os.Signal, 1)
	reload := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	signal.Notify(stop, syscall.SIGINT)
	signal.Notify(reload, syscall.SIGHUP)
	signal.Ignore(syscall.SIGSTOP)

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

	// Copyright 2016 aerth and contributors. Source code at https://github.com/aerth/cosgo
	// There should be a copy of the MIT license bundled with this software.

	// Future: dont use flags pkg
	flag.Parse()
	if *nolog {
		*quiet = true
		*logpath = "/dev/null" // fix for windows soon
	}

	if !*quiet {
		//fmt.Println(logo)
		fmt.Printf("\n\tcosgo v0.5\n\tCopyright 2016 aerth\n\tSource code at https://github.com/aerth/cosgo\n\tNow with Sendgrid, seconf, and a local mbox feature.\n\n")
	}
	args := flag.Args()

	if len(args) > 1 {
		usage()
		os.Exit(1)
	}
	if len(os.Args) > 1 {
		if os.Args[1] == "config" {
			loadConfig()
			os.Exit(1)
		}
		if os.Args[1] == "help" {
			usage()
			os.Exit(1)
		}
	}

	// -custom="anything" sets -config=true
	if *custom != "default" && *config == false {
		*config = true
	}

	// Load Configuration from seconf/secenv
	if *config == true {
		if *custom == "default" {
			*custom = "cosgo"
		}

		fmt.Println("Boot: Reading config file..." + *custom)

		if !loadConfig() {
			fmt.Println("Fatal: Can't load configuration file.")
			os.Exit(1)
		} else {
			fmt.Printf("done.")
		}
	}

	if os.Getenv("COSGOPAGEDIR") != "" {
		*custompages = os.Getenv("cosgopages")
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
		log.Fatal("Failed test.")
	}

	if os.Getenv("COSGO_API_KEY") == "" {

		// The length of the API key can be modified here.

		// Internal cron!!!
		go func() {
			for {
				if !*quiet {
					log.Println("Info: Generating Random POST Key...")
				}
				cosgo.PostKey = generateAPIKey(40)
				if *debug && !*quiet {
					log.Printf("Info: POST Key is " + cosgo.PostKey + "\n")
				}
				time.Sleep(cosgoRefresh)
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

	//The Magic
	r.HandleFunc("/{{whatever}}/send", emailHandler)

	//Defaults to true. We are serving out of /static/ for now
	if *static == true {
		s := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
		ss := http.FileServer(http.Dir("./static/"))
		sf := http.FileServer(http.Dir("./files/"))
		sp := http.StripPrefix("/"+*custompages+"/", http.FileServer(http.Dir("./"+*custompages+"/")))
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
		r.Methods("GET").Path("/static/{dir}/{whatever}.").Handler(s)
		r.Methods("GET").Path("/static/{dir}/{whatever}.md").Handler(s)
		r.Methods("GET").Path("/static/{dir}/{whatever}.md").Handler(s)
		r.Methods("GET").Path("/files/{whatever}").Handler(sf)

		r.Methods("GET").Path("/" + *custompages + "/{whatever}.html").Handler(sp)
		r.Methods("GET").Path("/" + *custompages + "/{dir}/{whatever}.html").Handler(sp)
	}

	if *love == true {
		r.HandleFunc("/{whatever}", loveHandler)
	}
	if *pages == true {
		r.HandleFunc("/"+*custompages+"{whatever}", pageHandler)
	}

	// Retrieve Captcha IMG and WAV
	r.Methods("GET").Path("/captcha/{captchacode}.png").Handler(captcha.Server(StdWidth, StdHeight))
	r.Methods("GET").Path("/captcha/download/{captchacode}.wav").Handler(captcha.Server(StdWidth, StdHeight))

	http.Handle("/", r)
	//End Routing

	// Start Runtime Info
	if *secure == false && !*quiet {
		log.Println("Warning: Running in *insecure* mode.")
		log.Println("Hint: Use -secure flag for https only.")
	}
	if mailbox == true {
		if !*quiet {
			log.Println("Info: local mode (read with mutt -Rf cosgo.mbox)")
		}
		f, err := os.OpenFile("./cosgo.mbox", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			log.Printf("error opening file: %v", err)
			log.Fatal("Hint: touch ./cosgo.mbox, or chown/chmod it so that the cosgo process can access it.")
			os.Exit(1)
		}
		mbox.Destination = cosgoDestination
		mbox.Mail = log.New(f, "", 0)

	}

	if *debug == false {
		if !*quiet {
			log.Println("[switching logs to " + *logpath + "]")
		}
		openLogFile()
	} else {
		if !*quiet {
			log.Println("[debug on: logs to stdout]")
		}
	}
	// Define listener
	if *debug && !*quiet {
		log.Printf("Link: " + getLink(*fastcgi, *bind, *port))
	}
	oglistener, err := net.Listen("tcp", *bind+":"+*port)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	listener, err := sl.New(oglistener)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	if *debug && !*quiet {
		log.Printf("Info: Got listener %s %s", listener.Addr().String(), listener.Addr().Network())
	}
	boottime := time.Now()
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
				case true: //
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
				case true: // https://, no fastcgi
					if listener != nil {
						go http.Serve(listener,
							csrf.Protect(antiCSRFkey,
								csrf.HttpOnly(true),
								csrf.FieldName("cosgo"),
								csrf.CookieName("cosgo"), csrf.Secure(true))(r))
					} else {
						log.Fatalln("nil listener")
					}
				case false: // This is using http://, no fastcgi.
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

		select {

		default:
			log.Printf("Uptime: %s", time.Since(boottime))
			time.Sleep(cosgoRefresh)
			//Do reload of server here so mux gets the updated routing info
			// apparently this works great as is
		}
	}
} // end loop
