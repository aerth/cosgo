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

package main

import (
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/http/fcgi"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/aerth/seconf"
	"github.com/dchest/captcha"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/goware/emailx"
	sl "github.com/hydrogen18/stoppableListener"
	"github.com/microcosm-cc/bluemonday"
)

var (
	//	mandrillAPIUrl   string
	mandrillKey      string
	sendgridKey      string
	cosgoDestination string
	cosgoKey         string
	CSRFKey          []byte
	Mail             *log.Logger
	cosgoRefresh     = 42 * time.Second // Will change in a few commits.
)

type C struct {
	CaptchaId string
}

const (
	// Minimum Captcha Length
	CaptchaLength = 3
	// Captcha will add *up to* CaptchaVariation to the CaptchaLength
	CaptchaVariation = 2
	CollectNum       = 100
	Expiration       = 10 * time.Minute
	StdWidth         = 240
	StdHeight        = 90
)

//usage shows how available flags.
func usage() {
	fmt.Println("\nusage: cosgo [flags]")
	fmt.Println("\nflags:")
	time.Sleep(1000 * time.Millisecond)
	flag.PrintDefaults()
	time.Sleep(1000 * time.Millisecond)
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
var (
	// TODO: dont use flags. Will be using "cosgo __action__" and env/seconf only.
	help       = flag.Bool("help", false, "Show this and quit")
	port       = flag.String("port", "8080", "Port to listen on")
	bind       = flag.String("bind", "127.0.0.1", "Default: 127.0.0.1, try 0.0.0.0")
	debug      = flag.Bool("debug", false, "Send logs to stdout. Dont switch to cosgo.log")
	api        = flag.Bool("api", false, "Show error.html for /")
	secure     = flag.Bool("secure", false, "PRODUCTION MODE - Accept only https secure cookie transfer.")
	mailmode   = flag.String("mailmode", "mailbox", "Choose one: mailbox, mandrill, sendgrid")
	fastcgi    = flag.Bool("fastcgi", false, "Use fastcgi (for with nginx etc)")
	static     = flag.Bool("static", true, "Serve /static/ files. Use -static=false to disable")
	noredirect = flag.Bool("noredirect", false, "Default is to redirect all 404s back to /. Set true to enable error.html template")
	love       = flag.Bool("love", false, "Fun. Show I love ___ instead of redirect")
	config     = flag.Bool("config", false, "Use config file at ~/.cosgo")
	custom     = flag.String("custom", "default", "Example: cosgo2 ...creates $HOME/.cosgo2")
	mailbox    = true
)

/*
TODO:
cosgo config
cosgo -h, cosgo help
cosgo fastcgi, cosgo http, cosgo
cosgo reconfig
cosgo custom custom-config-path
*/

var logo = `
                           _
  ___ ___  ___  __ _  ___ | |
 / __/ _ \/ __|/ _  |/ _ \| |
| (_| (_) \__ \ (_| | (_) |_|
 \___\___/|___/\__, |\___/(_)
               |___/
`

func main() {

	// Copyright 2016 aerth and contributors. Source code at https://github.com/aerth/cosgo
	// You should recieve a copy of the MIT license with this software.
	fmt.Println(logo)
	fmt.Println("\tcosgo v0.5\n\tCopyright 2016 aerth\n\tSource code at https://github.com/aerth/cosgo\n\tNow with Sendgrid, seconf, and local mbox ability.\n")

	// Set flags from command line
	//flag.Usage = usage
	flag.Parse()
	args := flag.Args()
	if len(args) > 1 {
		usage()
		os.Exit(1)
	}
	if len(os.Args) > 1 {
		if os.Args[1] == "config" {
			LoadConfig()
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
		fmt.Println("Boot: Reading config file...")
		if !LoadConfig() {
			fmt.Println("Fatal: Can't load configuration file.")
			os.Exit(1)
		} else {
			fmt.Printf("done.")
		}
	}

	// If user is still using CASGO_DESTINATION or CASGO_API_KEY (instead of COSGO)
	backwardsComp()
	// Define CSRFKey with env var, or set default.
	if !*config {
		if os.Getenv("COSGO_CSRF_KEY") == "" && string(CSRFKey) == "" {
			//	log.Println("You can now set COSGO_CSRF_KEY environmental variable. Using default.")
			CSRFKey = []byte("LI80PNK1xcT01jmQBsEyxyrNCrbyyFPjPU8CKnxwmCruxNijgnyb3hXXD3p1RBc0+LIRQUUbTtis6hc6LD4I/A==")
		} else {
			log.Println("CSRF key OK", os.Getenv("COSGO_CSRF_KEY"))
			CSRFKey = []byte(os.Getenv("COSGO_CSRF_KEY"))
		}
	}
	// Test environmental variables, if we aren't in -mailbox mode.
	QuickSelfTest()

	if os.Getenv("COSGO_API_KEY") == "" {
		// The length of the API key can be modified here.
		// Internal cron!!!
		go func() {
			for {
				log.Println("Boot: Generating Random POST Key...")
				cosgoKey = GenerateAPIKey(20)
				time.Sleep(cosgoRefresh)
			}
		}()
	} else {
		cosgoKey = os.Getenv("COSGO_API_KEY")
		// Print selected COSGO_API_KEY
		log.Println("COSGO_API_KEY:", getKey())
	}

	//Begin Routing
	r := mux.NewRouter()

	//Redirect or show custom error?
	if *noredirect == false {
		r.NotFoundHandler = http.HandlerFunc(RedirectHomeHandler)
	} else {
		r.NotFoundHandler = http.HandlerFunc(CustomErrorHandler)
	}
	r.HandleFunc("/", HomeHandler)

	//The Magic
	r.HandleFunc("/{{whatever}}/send", EmailHandler)

	//Defaults to true. We are serving out of /static/ for now
	if *static == true {
		s := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
		ss := http.FileServer(http.Dir("./static/"))
		sf := http.FileServer(http.Dir("./files/"))
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
	}

	if *love == true {
		r.HandleFunc("/{whatever}", LoveHandler)
	}

	// Retrieve Captcha IMG and WAV
	r.Methods("GET").Path("/captcha/{captchacode}.png").Handler(captcha.Server(StdWidth, StdHeight))
	r.Methods("GET").Path("/captcha/{captchacode}.wav").Handler(captcha.Server(StdWidth, StdHeight))

	http.Handle("/", r)
	//End Routing

	// Start Runtime Info
	fmt.Println("")
	if *secure == false {
		log.Println("Warning: Running in *insecure* mode.")
		log.Println("Hint: Use -secure flag for https only.")
	}
	if mailbox == true {
		log.Println("Mode: mailbox (read with mutt -Rf cosgo.mbox)")
		f, err := os.OpenFile("./cosgo.mbox", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			log.Printf("error opening file: %v", err)
			log.Fatal("Hint: touch ./cosgo.mbox, or chown/chmod it so that the cosgo process can access it.")
			os.Exit(1)
		}
		Mail = log.New(f, "", 0)

	}

	if *debug == false {
		log.Printf("Link: " + getLink(*fastcgi, *bind, *port))
		log.Println("[switching logs to cosgo.log]")

		OpenLogFile()
	} else {
		log.Println("[debug on: logs to stdout]")
	}

	log.Printf("Link: " + getLink(*fastcgi, *bind, *port))

	// Define listener

	fmt.Println("Trying to listen on " + getLink(*fastcgi, *bind, *port))

	oglistener, err := net.Listen("tcp", *bind+":"+*port)
	if err != nil {
		log.Println(err)
	}

	fmt.Println("Got oglistener")
	// listener, err := sl.New(oglistener)
	//
	// if err != nil {
	// 	log.Println(err)
	// }

	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGINT)

	listener, err := sl.New(oglistener)

	if err != nil {
		log.Println(err)
	}
	fmt.Println("Got listener")
	var wg sync.WaitGroup
	wg.Add(1)
	// Start Serving!

	// TODO: switch if nonsense to switch case
	// switch *fastcgi {
	// case true:
	// 	switch *secure {
	// 	case false:
	// 		fcgi.Serve(listener,
	// 			csrf.Protect(CSRFKey,
	// 				csrf.HttpOnly(true),
	// 				csrf.Secure(false))(r))
	// 		return
	// 	case true:
	// 		fcgi.Serve(listener,
	// 			csrf.Protect(CSRFKey,
	// 				csrf.HttpOnly(true),
	// 				csrf.Secure(true))(r))
	// 		return
	// 	}
	// case false:
	// 	switch *secure {
	// 	case true:
	// 		http.Serve(listener,
	// 			csrf.Protect(CSRFKey,
	// 				csrf.HttpOnly(true),
	// 				csrf.Secure(true))(r))
	// 		return
	// 	case false:
	// 		fmt.Println("Got switcher")
	// 		http.Serve(listener,
	// 			csrf.Protect(CSRFKey,
	// 				csrf.HttpOnly(true),
	// 				csrf.Secure(false))(r))
	//
	// 		return
	// 	}
	// }
	// fmt.Println("Got past switcher")

	//listener.Stop()
	// _, err = listener.Accept()
	// fmt.Println(err)
	fmt.Printf("Serving HTTP\n")
	//go func() {
	for {

		listener, err = sl.New(oglistener)

		//fmt.Printf("Waiting on server\n")
		//go wg.Wait()
		// TODO: switch if nonsense to switch case
		switch *fastcgi {
		case true:
			switch *secure {
			case false:
				go fcgi.Serve(listener,
					csrf.Protect(CSRFKey,
						csrf.HttpOnly(true),
						csrf.Secure(false))(r))
				return
			default:
				go fcgi.Serve(listener,
					csrf.Protect(CSRFKey,
						csrf.HttpOnly(true),
						csrf.Secure(true))(r))
				return
			}
		case false:
			switch *secure {
			case true:
				go http.Serve(listener,
					csrf.Protect(CSRFKey,
						csrf.HttpOnly(true),
						csrf.Secure(true))(r))
				return
			default:
				fmt.Println("made it ")
				if listener != nil {

					go http.Serve(listener,
						csrf.Protect(CSRFKey,
							csrf.HttpOnly(true),
							csrf.Secure(false))(r))
				}
				//return
			}

			fmt.Printf("Serving HTTP\n")
			select {
			case signal := <-stop:
				fmt.Printf("Got signal:%v\n", signal)
				listener.Close()
				listener.Stop()
			default:
				for {
					fmt.Println("loop")
					time.Sleep(cosgoRefresh)

				}
			}

			return
		}
		listener.Close()
		listener.Stop()
		os.Exit(1)
	} // end loop
} // End main function

func backwardsComp() {

	// For backwards compatibility
	if os.Getenv("CASGO_API_KEY") != "" && os.Getenv("COSGO_API_KEY") == "" {
		os.Setenv("COSGO_API_KEY", os.Getenv("CASGO_API_KEY"))
		log.Println("Please use COSGO_API_KEY instead of depreciated CASGO_API_KEY")
	}
	if os.Getenv("CASGO_DESTINATION") != "" && os.Getenv("COSGO_DESTINATION") == "" {
		os.Setenv("COSGO_DESTINATION", os.Getenv("CASGO_DESTINATION"))
		log.Println("Please use COSGO_DESTINATION instead of depreciated CASGO_DESTINATION")
	}

}

// Hello functions
func getKey() string {
	return cosgoKey
}
func getDestination() string {
	return cosgoDestination
}
func getMandrillKey() string {
	return mandrillKey
}

// HomeHandler parses the ./templates/index.html template file.
// This returns a web page with a form, captcha, CSRF token, and the cosgo API key to send the message.
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("home visitor: %s - %s - %s", r.UserAgent(), r.RemoteAddr, r.Host)
	t, err := template.New("Index").ParseFiles("./templates/index.html")
	if err != nil {
		// Do Something
		log.Println("Almost fatal: Cant load index.html template!")
		log.Println(err)
		fmt.Fprintf(w, "We are experiencing some technical difficulties. Please come back soon!")
	} else {
		data := map[string]interface{}{
			"Key":            getKey(),
			csrf.TemplateTag: csrf.TemplateField(r),
			"CaptchaId":      captcha.NewLen(CaptchaLength + rand.Intn(CaptchaVariation)),
		}
		t.ExecuteTemplate(w, "Index", data)

	}
}

// LoveHandler is just for fun.
// I love lamp. This displays affection for r.URL.Path[1:]
func LoveHandler(w http.ResponseWriter, r *http.Request) {

	p := bluemonday.UGCPolicy()
	subdomain := getSubdomain(r)
	lol := p.Sanitize(r.URL.Path[1:])
	if r.Method == "POST" {
		log.Printf("Something tried POST on %s", lol)
		http.Redirect(w, r, "/", 301)
	}
	if subdomain == "" {
		fmt.Fprintf(w, "I love %s!", lol)
		log.Printf("I love %s says %s at %s", lol, r.UserAgent(), r.RemoteAddr)
	} else {
		fmt.Fprintf(w, "%s loves %s!", subdomain, lol)
		log.Printf("I love %s says %s at %s", subdomain, r.UserAgent(), r.RemoteAddr)
	}

}

// CustomErrorHandler allows cosgo administrator to customize the 404 Error page
// Using the ./templates/error.html file.
func CustomErrorHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("visitor: 404 %s - %s at %s", r.Host, r.UserAgent(), r.RemoteAddr)
	p := bluemonday.UGCPolicy()
	domain := getDomain(r)
	lol := p.Sanitize(r.URL.Path[1:])
	log.Printf("404 on %s/%s", lol, domain)
	t, err := template.New("Error").ParseFiles("./templates/error.html")
	if err == nil {
		data := map[string]interface{}{
			"err":            "404",
			"Key":            getKey(),
			csrf.TemplateTag: csrf.TemplateField(r),
		}
		t.ExecuteTemplate(w, "Error", data)
	} else {
		log.Printf("template error: %s at %s", r.UserAgent(), r.RemoteAddr)
		log.Println(err)
		http.Redirect(w, r, "/", 301)
	}
}

// ContactHandler displays a contact form with CSRF and a Cookie. And maybe a captcha and drawbridge.
func ContactHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("contact: %s - %s at %s", r.Host, r.UserAgent(), r.RemoteAddr)
	t, err := template.New("Contact").ParseFiles("./templates/form.html")
	if err == nil {
		// Allow form in error page
		data := map[string]interface{}{
			"Key":            getKey(),
			csrf.TemplateTag: csrf.TemplateField(r),
			"CaptchaId":      captcha.New(),
		}

		t.ExecuteTemplate(w, "Contact", data)
	} else {
		log.Printf("Error: form template error: %s at %s", r.UserAgent(), r.RemoteAddr)
		log.Printf("Hint: Check ./templates/form.html")
		log.Println(err)
		http.Redirect(w, r, "/", 301)
	}

}

// RedirectHomeHandler redirects everyone home ("/") with a 301 redirect.
func RedirectHomeHandler(rw http.ResponseWriter, r *http.Request) {
	p := bluemonday.UGCPolicy()
	domain := getDomain(r)
	lol := p.Sanitize(r.URL.Path[1:])
	log.Printf("Redirecting %s back home on %s", lol, domain)
	http.Redirect(rw, r, "/", 301)

}

// EmailHandler checks the Captcha string, and calls MandrillSender
func EmailHandler(rw http.ResponseWriter, r *http.Request) {

	destination := cosgoDestination
	var query url.Values
	fmt.Println(r.URL.Path)
	ourpath := strings.TrimLeft(r.URL.Path, "/")
	ourpath = strings.TrimRight(ourpath, "/send")

	fmt.Println(ourpath)
	fmt.Println(cosgoKey)
	if r.Method == "POST" && strings.ContainsAny(r.URL.Path, cosgoKey) {
		// Method is POST, URL KEY is correct.
		if !captcha.VerifyString(r.FormValue("captchaId"), r.FormValue("captchaSolution")) {
			fmt.Fprintf(rw, "You may be a robot. Can you go back and try again?")
			log.Printf("User Error: CAPTCHA %s at %s", r.UserAgent(), r.RemoteAddr)
			http.Redirect(rw, r, "/", 301)
		} else {
			// Captcha is correct.
			log.Printf("User Human: %s at %s", r.UserAgent(), r.RemoteAddr)
			r.ParseForm()
			query = r.Form

			// Phasing Mandrill out
			switch *mailmode {
			case "mandrill":
				MandrillSender(rw, r, destination, query)
			case "sendgrid":
				SendgridSender(rw, r, destination, query)

			default:
				EmailSaver(rw, r, destination, query)
				log.Printf("SUCCESS-contact: %s at %s", r.UserAgent(), r.RemoteAddr)
				fmt.Fprintln(rw, "<html><p>Thanks! Would you like to go <a href=\"/\">back</a>?</p></html>")
			}

		}
	} else {
		log.Printf("User Error: KEY %s at %s", r.UserAgent(), r.RemoteAddr)
		http.Redirect(rw, r, "/", 301)
	}

}

// EmailSaver always returns success for the visitor. This function needs some work.
func EmailSaver(rw http.ResponseWriter, r *http.Request, destination string, query url.Values) {
	form := ParseQuery(query)
	t := time.Now()
	mailtime := t.Format("Mon Jan 2 15:04:05 2006")
	mailtime2 := t.Format("Mon, 2 Jan 2006 15:04:05 -0700")
	Mail.Println("From " + form.Email + " " + mailtime)
	Mail.Println("Return-path: <" + form.Email + ">")
	Mail.Println("Envelope-to: " + destination)
	Mail.Println("Delivery-date: " + mailtime2)
	Mail.Println("To: " + destination)
	Mail.Println("Subject: " + form.Subject)
	Mail.Println("From: " + form.Email)
	Mail.Println("Date: " + mailtime2)

	Mail.Println("\n" + form.Message + "\n\n")

}

// MandrillSender always returns success for the visitor. This function needs some work.
func MandrillSender(rw http.ResponseWriter, r *http.Request, destination string, query url.Values) {
	form := ParseQuery(query)
	//Validate user submitted email address
	err := emailx.Validate(form.Email)
	if err != nil {
		fmt.Fprintln(rw, "<html><p>Email is not valid. Would you like to go <a href=\"/\">back</a>?</p></html>")

		if err == emailx.ErrInvalidFormat {
			fmt.Fprintln(rw, "<html><p>Email is not valid format.</p></html>")
		}
		if err == emailx.ErrUnresolvableHost {
			fmt.Fprintln(rw, "<html><p>We don't recognize that email provider.</p></html>")
		}
	}
	//Normalize email address
	form.Email = emailx.Normalize(form.Email)
	//Is it empty?
	if form.Email == "" {
		http.Redirect(rw, r, "/", 301)
		return
	}

	if sendMandrill(destination, form) {
		fmt.Fprintln(rw, "<html><p>Thanks! Would you like to go <a href=\"/\">back</a>?</p></html>")
		http.Redirect(rw, r, "/", 301)
		log.Printf("SUCCESS-contact: %s at %s", r.UserAgent(), r.RemoteAddr)
	} else {
		log.Printf("FAIL-contact: %s at %s", r.UserAgent(), r.RemoteAddr)
		log.Printf("debug: %s to mandrill %s", form, destination)
		log.Printf("debug: %s to mandrill %s", form.Message, destination)

		t, err := template.New("Error").ParseFiles("./templates/error.html")
		if err == nil {
			data := map[string]interface{}{
				"err":            "Mail System",
				"Key":            getKey(),
				csrf.TemplateTag: csrf.TemplateField(r),
			}
			t.ExecuteTemplate(rw, "Error", data)
		} else {
			log.Printf("template error: %s at %s", r.UserAgent(), r.RemoteAddr)
			log.Println(err)
			http.Redirect(rw, r, "/", 301)
		}
	}
}

// SendgridSender always returns success for the visitor. This function needs some work.
func SendgridSender(rw http.ResponseWriter, r *http.Request, destination string, query url.Values) {
	form := ParseQuery(query)
	//Validate user submitted email address
	err := emailx.Validate(form.Email)
	if err != nil {
		fmt.Fprintln(rw, "<html><p>Email is not valid. Would you like to go <a href=\"/\">back</a>?</p></html>")

		if err == emailx.ErrInvalidFormat {
			fmt.Fprintln(rw, "<html><p>Email is not valid format.</p></html>")
		}
		if err == emailx.ErrUnresolvableHost {
			fmt.Fprintln(rw, "<html><p>We don't recognize that email provider.</p></html>")
		}
	}
	//Normalize email address
	form.Email = emailx.Normalize(form.Email)
	//Is it empty?
	if form.Email == "" {
		http.Redirect(rw, r, "/", 301)
		return
	}

	if sendgridSend(destination, form) {
		fmt.Fprintln(rw, "<html><p>Thanks! Would you like to go <a href=\"/\">back</a>?</p></html>")
		http.Redirect(rw, r, "/", 301)
		log.Printf("SUCCESS-contact: %s at %s", r.UserAgent(), r.RemoteAddr)
	} else {
		log.Printf("FAIL-contact: %s at %s", r.UserAgent(), r.RemoteAddr)
		log.Printf("debug: %s to mandrill %s", form, destination)
		log.Printf("debug: %s to mandrill %s", form.Message, destination)

		t, err := template.New("Error").ParseFiles("./templates/error.html")
		if err == nil {
			data := map[string]interface{}{
				"err":            "Mail System",
				"Key":            getKey(),
				csrf.TemplateTag: csrf.TemplateField(r),
			}
			t.ExecuteTemplate(rw, "Error", data)
		} else {
			log.Printf("template error: %s at %s", r.UserAgent(), r.RemoteAddr)
			log.Println(err)
			http.Redirect(rw, r, "/", 301)
		}
	}
}

func ParseQuery(query url.Values) *Form {
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
		log.Println("Requested domain: " + requesthost)
		domainParts := strings.Split(requesthost, ".")
		log.Println("Subdomain:" + domainParts[0])
		if len(domainParts) > 2 {
			if domainParts[0] != "127" {
				return domainParts[0]
			}
		}
	}
	return ""
}

// serverSingle just shows one file.
func serveSingle(pattern string, filename string) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filename)
	})
}

// Key Generator
func init() {
	rand.Seed(time.Now().UnixNano())
}

var runes = []rune("____ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890123456789012345678901234567890")

//GenerateAPIKey does API Key Generation with the given runes.
func GenerateAPIKey(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = runes[rand.Intn(len(runes))]
	}
	return string(b)
}

//OpenLogFile switches the log engine to a file, rather than stdout
func OpenLogFile() {
	f, err := os.OpenFile("./cosgo.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Printf("error opening file: %v", err)
		log.Fatal("Hint: touch ./cosgo.log, or chown/chmod it so that the cosgo process can access it.")
		os.Exit(1)
	}
	log.SetOutput(f)
}

//getLink returns the bind:port or http://bind:port string
func getLink(fastcgi bool, bind string, port string) string {
	if fastcgi == true {
		link := bind + ":" + port
		return link
	} else {
		link := "http://" + bind + ":" + port
		return link
	}
}

func LoadConfig() bool {

	// Detect seconf file. Create if it doesn't exist.
	if !seconf.Detect(*custom) {
		seconf.Create(*custom,
			"cosgo config generator", // Title
			"32 bit CSRF Key, can be 1 for auto generated.",
			"COSGO_KEY: can be 1 for auto generated.\nIf auto-generated, the key will change every time cosgo restarts.\nThis is a spam prevention technique,\nit changes the form's POST end point on startup.",
			"COSGO_DESTINATION, where SMTP mails will be sent.\n In mailbox mode, COSGO_DESTINATION is where all mail is addressed.\nFor good time, set this to the email address you will be replying from.",
			"Please select from the following mailbox options. \n\n\t\tmandrill\tsendgrid. \n\nUse 0 for local mailbox mode.",
			"pass MANDRILL_KEY, can be 0 if local or sendgrid.",
			"pass SENDGRID_KEY, can be 0 if local or mandrill.")
	}

	// Now that a config file exists, unlock it.
	configdecoded, err := seconf.Read("cosgo")
	if err != nil {
		fmt.Println("error:")
		fmt.Println(err)
		return false
	}
	configarray := strings.Split(configdecoded, "::::")

	// Cosgo 0.5 uses new config file!
	if len(configarray) < 5 {
		fmt.Println("Broken config file. Create a new one.")
		return false
	}
	if err != nil {
		fmt.Println(err)
		return false
	}
	CSRFKey = []byte(configarray[0])
	cosgoKey = configarray[1]
	cosgoDestination = configarray[2]
	*mailmode = configarray[3]
	mandrillKey = configarray[4]
	sendgridKey = configarray[5]

	if configarray[0] == "1" {
		CSRFKey = []byte("LI80POC1xcT01jmQBsEyxyrDCrbyyFPjPU8CKnxwmCruxNijgnyb3hXXD3p1RBc0+LIRQUUbTtis6hc6LD4I/A==")
	}

	if cosgoKey == "1" {
		cosgoKey = ""
	}

	if cosgoDestination == "1" {
		cosgoDestination = "user@example.com"
	}

	if configarray[3] == "0" {
		*mailmode = "mailbox"
	}

	if configarray[3] == "0" {
		mandrillKey = ""
	}
	if configarray[4] == "0" {
		sendgridKey = ""
	}

	switch *mailmode {
	case "mailbox":
		log.Println("Saving mail (cosgo.mbox) addressed to " + cosgoDestination)
	case "mandrill":
		log.Println("Sending via Mandrill to " + cosgoDestination)
	case "sendgrid":
		log.Println("Sending via Sendgrid to " + cosgoDestination)
	default:
		log.Fatalln("No mailmode.")
	}

	return true
}

func QuickSelfTest() {
	return
	if !*config {
		if mailbox != true {
			switch *mailmode {
			case "mandrill":
				mandrillKey = os.Getenv("MANDRILL_KEY")
				if mandrillKey == "" {
					log.Fatal("Fatal: environmental variable `MANDRILL_KEY` is Crucial.\n\n\t\tHint: export MANDRILL_KEY=123456789")
					os.Exit(1)
				}
			case "sendgrid":
				sendgridKey = os.Getenv("SENDGRID_KEY")
				if mandrillKey == "" {
					log.Fatal("Fatal: environmental variable `SENDGRID_KEY` is Crucial.\n\n\t\tHint: export SENDGRID_KEY=123456789")
					os.Exit(1)
				}
			default:
				mailbox = true
			}

			cosgoDestination = os.Getenv("COSGO_DESTINATION")
			if cosgoDestination == "" {
				log.Fatal("Fatal: environmental variable `COSGO_DESTINATION` is Crucial.\n\n\t\tHint: export COSGO_DESTINATION=\"your@email.com\"")
				os.Exit(1)
			}

		} else {
			cosgoDestination = os.Getenv("COSGO_DESTINATION")
			if cosgoDestination == "" {

				log.Println("Info: COSGO_DESTINATION not set. Using user@example.com")
				log.Println("Hint: export COSGO_DESTINATION=\"your@email.com\"")
			}
		}
	}
	_, err := template.New("Index").ParseFiles("./templates/index.html")
	if err != nil {
		log.Println("Fatal: Template Error:", err)
		log.Fatal("Fatal: Template Error\n\n\t\tHint: Copy ./templates and ./static from $GOPATH/src/github.com/aerth/cosgo/ to the location of your binary.")
	}

	_, err = template.New("Contact").ParseFiles("./templates/form.html")
	if err != nil {
		log.Println("Fatal: Template Error:", err)
		log.Fatal("\t\tHint: Copy ./templates and ./static from $GOPATH/src/github.com/aerth/cosgo/ to the location of your binary.")
	}

	_, err = template.New("Error").ParseFiles("./templates/error.html")
	if err != nil {
		log.Println("Fatal: Template Error:", err)
		log.Fatal("Fatal: Template Error\nHint: Copy ./templates and ./static from $GOPATH/src/github.com/aerth/cosgo/ to the location of your binary.")
	}

}
