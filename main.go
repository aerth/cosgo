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
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/aerth/mbox"

	"github.com/gorilla/csrf"
	sl "github.com/hydrogen18/stoppableListener"
)

var (
	version          = "0.9.G" // Use makefile for version hash
	destinationEmail = "cosgo@localhost"
	antiCSRFkey      = []byte("LI80PNK1xcT01jmQBsEyxyrNCrbyyFPjPU8CKnxwmCruxNijgnyb3hXXD3p1RBc0+LIRQUUbTtis6hc6LD4I/A==")
	cosgoRefresh     = 42 * time.Minute
	err              error
	hitcounter       int
	timeboot         = time.Now()
	templateDir      = "./templates/"
	cwd              string
	publicKey        []byte
	cosgo            = new(Cosgo)
	// flags
	port            = flag.String("port", "8080", "Port to listen on")
	bind            = flag.String("bind", "0.0.0.0", "Default: 0.0.0.0 (all interfaces)...\n\tTry -bind=127.0.0.1")
	debug           = flag.Bool("debug", false, "More verbose.")
	quiet           = flag.Bool("quiet", false, "Less output. Can be combined with -nolog for absolute silence.")
	nolog           = flag.Bool("nolog", false, "Logs to /dev/null")
	gpg             = flag.String("gpg", "", "Path to ascii-armored GPG public key (for encrypting messages.)")
	customExtension = flag.String("ext", "", "Serve extra static files. Uses regex. \n\tExample: -ext \"pdf|txt|html\"")
	sendgridKey     = flag.String("sg", "", "Sendgrid API key (disables mbox)")
	logfile         = flag.String("log", "", "Use a log file instead of stdout\n\tExample: -log cosgo.log -debug")
	mboxfile        = flag.String("mbox", "cosgo.mbox", "Custom mbox file name\n\tExample: -mbox custom.mbox")
)

// Cosgo struct holds the Boottime and
type Cosgo struct {
	URLKey   string
	Visitors int
}

func main() {
	if !*quiet {
		fmt.Println(logo)
		fmt.Printf("\n\tcosgo v" + version + "\n")
		fmt.Printf("\tCopyright 2016 aerth\n\tSource code at https://github.com/aerth/cosgo\n")
	}
	timeboot, cwd, staticDir, templatesDir := initialize()
	r := route(cwd, staticDir)
	//	http.Handle("/", r)

	if *debug && !*quiet {
		log.Println("booted:", timeboot)
		log.Println("working dir:", cwd)
		log.Println("css/js/img dir:", staticDir)
		log.Println("cosgo templates dir:", templatesDir)
	}
	// Fire up the cosgo engine
	go func() {
		for {
			if *debug && !*quiet {
				log.Println("Info: Generating Random 40 URL Key...")
			}
			// set a random URL key (40 char length)
			cosgo.URLKey = generateURLKey(40)
			if *debug && !*quiet {
				log.Printf("Info: URL Key is " + cosgo.URLKey + "\n")
			}
			// every X minutes
			time.Sleep(cosgoRefresh)
		}
	}()

	if *sendgridKey != "" {
		*mboxfile = os.DevNull
	}
	f, err := os.OpenFile(*mboxfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Printf("error opening file: %v", err)
		log.Fatal("Hint: touch ./cosgo.mbox, or chown/chmod it so that the cosgo process can access it.")
		os.Exit(1)
	}
	mbox.Destination = destinationEmail
	mbox.Mail = log.New(f, "", 0)
	if *nolog {
		*logfile = os.DevNull
	}
	// Switch to cosgo.log if not debug
	if *logfile != "" {
		openLogFile()
	}

	// Start serving
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
	// Start Serving Loop
	for {
		listener, err = sl.New(oglistener)
		if err != nil {
			log.Fatalln(err)
		}
		go func() {
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

		}()
		// End the for-loop with a timer/hit counter
		select {
		default:
			if !*quiet {
				log.Printf("Uptime: %s", time.Since(timeboot))
				log.Printf("Hits: %s", strconv.Itoa(hitcounter))
			}
			time.Sleep(time.Minute * 30)
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

func init() {

	// Key Generator
	rand.Seed(time.Now().UnixNano())

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
	flag.Parse()

}

//openLogFile switches the log engine to a file, rather than stdout.
func openLogFile() {
	if *logfile == "" {
		return
	}
	f, err := os.OpenFile(*logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Printf("error opening file: %v", err)
		log.Fatal("Hint: touch " + *logfile + ", or chown/chmod it so that the cosgo process can access it.")
		os.Exit(1)
	}
	log.SetOutput(f)
}

func initialize() (time.Time, string, string, string) {

	// Load environmental variables as flags
	if os.Getenv("COSGO_PORT") != "" {
		*port = os.Getenv("COSGO_PORT")
	}

	if os.Getenv("COSGO_BIND") != "" {
		*bind = os.Getenv("COSGO_BIND")
	}

	if os.Getenv("COSGO_REFRESH") != "" {
		cosgoRefresh, err = time.ParseDuration(os.Getenv("COSGO_REFRESH"))
		if err != nil {
			log.Fatalln(err)
		}
	}

	if os.Getenv("COSGO_MBOX") != "" {
		*mboxfile = os.Getenv("COSGO_MBOX")
	}
	if os.Getenv("COSGO_LOG") != "" {
		*logfile = os.Getenv("COSGO_LOG")
	}
	if os.Getenv("COSGO_GPG") != "" {
		*gpg = os.Getenv("COSGO_GPG")
	}
	if *gpg != "" {
		log.Println("\t[gpg pubkey: " + *gpg + "]")
		publicKey = read2mem(rel2real(*gpg))
	}
	timeboot := time.Now()
	cwd, _ := os.Getwd()
	staticDir := staticFinder(cwd)
	templatesDir := templateFinder()

	return timeboot, cwd, staticDir, templatesDir

}

// templateFinder returns the template directory we will use, if one isn't found, the error is fatal.
func templateFinder() string {
	templateDir := "./templates/"
	if *debug && !*quiet {
		log.Printf("Trying ./templates/")
	}
	_, notlocal := template.New("Index").ParseFiles(templateDir + "index.html")
	if notlocal != nil {
		if *debug && !*quiet {
			log.Println(notlocal)
			log.Printf("...Trying /usr/local/share/cosgo/templates/")
		}
		_, notglobal := template.New("Index").ParseFiles("/usr/local/share/cosgo/templates/index.html")
		if notglobal != nil {
			if *debug && !*quiet {
				log.Printf("...No templatesDir found, creating one.")
			}
			RestoreAssets(".", "templates")
		} else {
			if *debug && !*quiet {
				log.Println("Using /usr/local/share/cosgo/templates directory")
			}
			return "/usr/local/share/cosgo/templates/"
		}
	}
	if *debug && !*quiet {
		log.Printf("Using %s directory\n", templateDir)
	}
	return templateDir
}

// staticFinder returns the static directory. If none is found, static files are disabled.
func staticFinder(cwd string) string {
	staticDir := "./static/"
	_, err = os.Open(staticDir)
	if err != nil {
		if os.IsNotExist(err) {
			staticDir = "/usr/local/share/cosgo/static/"
			_, err = os.Open(staticDir)
			if err != nil {
				if os.IsNotExist(err) {
					if *debug {
						log.Printf("No staticDir. Creating one.")
					}
					RestoreAssets(".", "static")
					staticDir = "./static/"
				}
			}
		}
	}

	return staticDir
}

func getKey() string {
	return cosgo.URLKey
}
func getDestination() string {
	return destinationEmail
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
func rel2real(file string) (realpath string) {
	pathdir, _ := path.Split(file)

	if pathdir == "" {
		realpath, _ = filepath.Abs(file)
	} else {
		realpath = file
	}
	return realpath
}

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

//generateURLKey creates a new key, with the given runes, n length.
func generateURLKey(n int) string {
	runes := []rune("____ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890123456789012345678901234567890")
	b := make([]rune, n)
	for i := range b {
		b[i] = runes[rand.Intn(len(runes))]
	}
	return strings.TrimSpace(string(b))
}
