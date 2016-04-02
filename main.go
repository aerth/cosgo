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
	"strings"
	"time"

	"github.com/aerth/seconf"
	"github.com/dchest/captcha"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/goware/emailx"
	"github.com/microcosm-cc/bluemonday"
)

var smtpstyle = os.Getenv("COSGO_MODE")

var (
	//	mandrillAPIUrl   string
	mandrillKey      string
	sendgridKey      string
	cosgoDestination string
	cosgoAPIKey      string
	CSRFKey          []byte
	Mail             *log.Logger
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
	fmt.Println("\nExample: cosgo -insecure -port 8080 -fastcgi -debug")
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
	help       = flag.Bool("help", false, "Show usage help and quit")
	port       = flag.String("port", "8080", "Port to listen on")
	bind       = flag.String("bind", "127.0.0.1", "Default: 127.0.0.1, try 0.0.0.0")
	debug      = flag.Bool("debug", false, "Send logs to stdout. Dont switch to cosgo.log")
	api        = flag.Bool("api", false, "Show error.html for /")
	insecure   = flag.Bool("insecure", false, "Accept insecure cookie transfer (http/80). Otherwise, you must use SSL.")
	mailbox    = flag.Bool("mailbox", true, "Use local cosgo.mbox file. Disables SMTP and outgoing traffic.")
	fastcgi    = flag.Bool("fastcgi", false, "Use fastcgi (for with nginx etc)")
	static     = flag.Bool("static", true, "Serve /static/ files. Use -static=false to disable")
	noredirect = flag.Bool("noredirect", false, "Default is to redirect all 404s back to /. Set true to enable error.html template")
	love       = flag.Bool("love", false, "Fun. Show I love ___ instead of redirect")
	config     = flag.Bool("config", false, "Use config file at ~/.cosgo")
	custom     = flag.String("custom", "cosgo", "Use config file at ~/.cosgo")
)

func main() {

	// Copyright 2016 aerth and contributors. Source code at https://github.com/aerth/cosgo
	// You should recieve a copy of the MIT license with this software.
	log.Println("\n\n\tcosgo v0.4\n\tCopyright 2016 aerth\n\tSource code at https://github.com/aerth/cosgo")
	// Set flags from command line
	//flag.Usage = usage
	flag.Parse()
	args := flag.Args()
	if len(args) > 1 {
		usage()
		os.Exit(2)
	}

	// -custom="anything" sets -config=true
	if *custom != "cosgo" {
		*config = true
	}

	// Load Configuration from seconf/secenv
	if *config == true {
		DoConfig()
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

	QuickSelfTest(*mailbox)

	if !*config {

		// Print API Key
		if os.Getenv("COSGO_API_KEY") == "" {
			log.Println("Generating Random API Key...")
			// The length of the API key can be modified here.
			cosgoAPIKey = GenerateAPIKey(20)
			// Print new GenerateAPIKey
			log.Println("COSGO_API_KEY:", getKey())
		} else {
			cosgoAPIKey = os.Getenv("COSGO_API_KEY")
			// Print selected COSGO_API_KEY
			log.Println("COSGO_API_KEY:", getKey())
		}
	}
	//Begin Routing
	r := mux.NewRouter()

	//Redirect or show custom error?
	if *noredirect == false {
		r.NotFoundHandler = http.HandlerFunc(RedirectHomeHandler)
	} else {
		r.NotFoundHandler = http.HandlerFunc(CustomErrorHandler)
	}

	//If -api flag is set, show custom error.html template on / (and every page)
	if *api == false {
		r.HandleFunc("/", HomeHandler)
	} else {
		r.HandleFunc("/", CustomErrorHandler)
	}

	//The Magic
	r.HandleFunc("/"+cosgoAPIKey+"/form", HomeHandler)
	r.HandleFunc("/"+cosgoAPIKey+"/form/", HomeHandler)
	r.HandleFunc("/"+cosgoAPIKey+"/send", EmailHandler)

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

	if *mailbox == true {
		log.Println("mailbox mode.")
		f, err := os.OpenFile("./cosgo.mbox", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			log.Printf("error opening file: %v", err)
			log.Fatal("Hint: touch ./cosgo.mbox, or chown/chmod it so that the cosgo process can access it.")
			os.Exit(1)
		}
		Mail = log.New(f, "", 0)

	}

	if *debug == false {
		log.Println("[switching logs to cosgo.log]")
		OpenLogFile()
	} else {
		log.Println("[debug on: not using cosgo.log]")
	}

	log.Printf("cosgo is live on " + getLink(*fastcgi, *bind, *port))
	// Start Serving!
	if *fastcgi == true {
		listener, err := net.Listen("tcp", *bind+":"+*port)
		if err != nil {
			log.Fatal("Could not bind: ", err)
		}
		if *insecure == true {
			log.Fatal(fcgi.Serve(listener, csrf.Protect(CSRFKey, csrf.HttpOnly(true), csrf.Secure(false))(r)))
		} else {
			log.Println("info: https:// only")
			log.Fatal(fcgi.Serve(listener, csrf.Protect(CSRFKey, csrf.HttpOnly(true), csrf.Secure(true))(r)))
		}
	} else if *fastcgi == false && *insecure == true {
		log.Fatal(http.ListenAndServe(":"+*port, csrf.Protect(CSRFKey, csrf.HttpOnly(true), csrf.Secure(false))(r)))
	} else if *fastcgi == false && *insecure == false {
		log.Println("info: https:// only")
		// Change this CSRF auth token in production!
		log.Fatal(http.ListenAndServe(":"+*port, csrf.Protect(CSRFKey, csrf.HttpOnly(true), csrf.Secure(true))(r)))
	}

}

// End main function

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
	return cosgoAPIKey
}
func getDestination() string {
	return cosgoDestination
}
func getMandrillKey() string {
	return mandrillKey
}

func QuickSelfTest(mailbox bool) {
	log.Println("Starting self test...")

	if !*config {

		if mailbox != true {

			switch smtpstyle {
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

				log.Println("Warning: environmental variable `COSGO_DESTINATION` is not set. Using default.")
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

	log.Println("Passed self test.")
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
	log.Printf("api: %s - %s at %s", r.Host, r.UserAgent(), r.RemoteAddr)
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
		log.Printf("api template error: %s at %s", r.UserAgent(), r.RemoteAddr)
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
	if r.Method == "POST" {
		if !captcha.VerifyString(r.FormValue("captchaId"), r.FormValue("captchaSolution")) {
			fmt.Fprintf(rw, "You may be a robot. Can you go back and try again?")
			log.Printf("FAIL-contact: %s at %s", r.UserAgent(), r.RemoteAddr)
			http.Redirect(rw, r, "/", 301)
		} else {
			r.ParseForm()
			query = r.Form
			if *mailbox == true {
				EmailSaver(rw, r, destination, query)
				log.Printf("SUCCESS-contact: %s at %s", r.UserAgent(), r.RemoteAddr)
				fmt.Fprintln(rw, "<html><p>Thanks! Would you like to go <a href=\"/\">back</a>?</p></html>")
			} else {
				// Phasing Mandrill out
				switch smtpstyle {
				case "mandrill":
					MandrillSender(rw, r, destination, query)
				case "sendgrid":
					SendgridSender(rw, r, destination, query)
				}

			}
		}
	} else {
		http.Redirect(rw, r, "/", 301)
	}

}

// MandrillSender always returns success for the visitor. This function needs some work.
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

func DoConfig() {
	if !seconf.Detect("cosgo") {
		seconf.Create("cosgo",
			"cosgo config generator",
			"32 bit CSRF Key, can be 1 for auto generated.",
			"COSGO_KEY: can be 1 for auto generated.\nIf auto-generated, the key will change every time cosgo restarts.\nThis is a spam prevention technique,\nit changes the form's POST end point on startup.",
			"COSGO_DESTINATION, where SMTP mails will be sent.\n In mailbox mode, COSGO_DESTINATION is where all mail is addressed.\nFor good time, set this to the email address you will be replying from.",
			"Please select from the following mailbox options. \n\n\t\tmandrill\tsendgrid. \n\nUse 0 for local mailbox mode.",
			"MANDRILL_KEY, can be 0 if local or sendgrid.",
			"SENDGRID_KEY, can be 0 if local or mandrill.")
	}

	configdecoded, err := seconf.Read("cosgo")
	if err != nil {
		fmt.Println("error:")
		fmt.Println(err)
		os.Exit(1)
	}
	configarray := strings.Split(configdecoded, "::::")
	if len(configarray) < 2 {
		fmt.Println("Broken config file. Create a new one.")
		os.Exit(1)
	}
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	CSRFKey = []byte(configarray[0])
	cosgoAPIKey = configarray[1]
	cosgoDestination = configarray[2]
	mandrillKey = configarray[3]
	sendgridKey = configarray[4]

	if configarray[0] == "1" {
		CSRFKey = []byte("LI80POC1xcT01jmQBsEyxyrDCrbyyFPjPU8CKnxwmCruxNijgnyb3hXXD3p1RBc0+LIRQUUbTtis6hc6LD4I/A==")
	}

	if configarray[1] == "1" {
		cosgoAPIKey = ""
	}

	if configarray[2] == "1" {
		cosgoDestination = "user@example.com"
	}

	if configarray[3] == "0" {
		mandrillKey = ""
	}
	if configarray[4] == "0" {
		sendgridKey = ""
	}
	if *mailbox {
		log.Println("Saving mail (cosgo.mbox) addressed to " + cosgoDestination)
	} else {
		log.Println("Sending via mandrill to " + cosgoDestination)

	}
}
