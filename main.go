package main

import (
	"flag"

	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"time"

	"html/template"

	//http "net/http"
	"net/url"
	"strings"

	"github.com/dchest/captcha"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/microcosm-cc/bluemonday"
)

var (
	mandrillApiUrl   string
	mandrillKey      string
	casgoDestination string
	casgoAPIKey      string
)

type C struct {
	CaptchaId string
}

const (
	// Default number of digits in captcha solution.
	DefaultLen = 6
	// The number of captchas created that triggers garbage collection used
	// by default store.
	CollectNum = 100
	// Expiration time of captchas used by default store.
	Expiration = 10 * time.Minute
)

const (
	// Standard width and height of a captcha image.
	StdWidth  = 240
	StdHeight = 80
)

func main() {

	// Copyright 2016 aerth and contributors. Source code at https://github.com/aerth
	// You should recieve a copy of the MIT license with this software.
	log.Println("\n\n\tcasgo v0.4\n\n\tCopyright 2016 aerth\n\n\tSource code at https://github.com/aerth/casgo\n\n")

	// We can set the CASGO_API_KEY environment variable, or it defaults to a new random one!

	//
	port := flag.String("port", "8080", "HTTP Port to listen on")
	Debug := flag.Bool("debug", false, "be verbose, dont switch to casgo.log")
	insecure := flag.Bool("insecure", false, "accept insecure cookie transfer (http/80)")
	mailbox := flag.Bool("mailbox", false, "disable mandrill send")
	fastcgi := flag.Bool("fastcgi", false, "use fastcgi with nginx")
	bind := flag.String("bind", "127.0.0.1", "default: 127.0.0.1 - maybe 0.0.0.0 ?")
	flag.Parse()

	if os.Getenv("CASGO_API_KEY") == "" {
		log.Println("Generating Random API Key...")
		// The length of the API key can be modified here.
		casgoAPIKey = GenerateAPIKey(20)
		// Print new GenerateAPIKey
		log.Println("CASGO_API_KEY:", getKey())
	} else {
		casgoAPIKey = os.Getenv("CASGO_API_KEY")
		// Print selected CASGO_API_KEY
		log.Println("CASGO_API_KEY:", getKey())
	}
	//For now...
	mandrillApiUrl = "https://mandrillapp.com/api/1.0/"
	//From environmental variable.
	mandrillKey = os.Getenv("MANDRILL_KEY")
	if mandrillKey == "" && *mailbox == false{
		log.Fatal("MANDRILL_KEY is Crucial. Type: export MANDRILL_KEY=123456789")
		os.Exit(1)
	}
	casgoDestination = os.Getenv("CASGO_DESTINATION")
	if casgoDestination == "" && *mailbox == false {
		log.Fatal("CASGO_DESTINATION is Crucial. Type: export CASGO_DESTINATION=\"your@email.com\"")
		os.Exit(1)
	}

// Right-clickable for preview
	log.Printf("listening on http://%s:%s", *bind, *port)

//Begin Routing
	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(CustomErrorHandler)
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/"+casgoAPIKey+"/form", ContactHandler)
	r.HandleFunc("/"+casgoAPIKey+"/form/", ContactHandler)
	// Magic URL Generator for API endpoint
	r.HandleFunc("/"+casgoAPIKey+"/send", EmailHandler)
	//r.Handle("/static/{static}", http.FileServer(http.Dir("./static")))
	s := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	ss := http.FileServer(http.Dir("./static/"))
	// Serve /static folder and favicon etc
	r.Path("/favicon.ico").Handler(ss)
	r.Path("/robots.txt").Handler(ss)
	r.Path("/sitemap.xml").Handler(ss)
	r.PathPrefix("/static/{whatever}").Handler(s)
	r.HandleFunc("/{whatever}", LoveHandler)
	// Retrieve Captcha IMG and WAV
	r.Methods("GET").PathPrefix("/captcha/").Handler(captcha.Server(captcha.StdWidth, captcha.StdHeight))
	r.NotFoundHandler = http.HandlerFunc(CustomErrorHandler)
	//http.NotFoundHandler = r.HandlerFunc(CustomErrorHandler)
	http.Handle("/", r)
//End Routing


	// Switch to file log so we can ctrl+c and launch another instance :)
	if *mailbox == true {
		log.Println("mailbox mode: not enabled just saying")
		//CreateMailBox()
	}

	if *Debug == false {
		log.Println("quiet mode: [switching logs to casgo.log]")
		OpenLogFile()
	} else {
		log.Println("Debug on: [not using casgo.log]")
	}

	if *fastcgi == true {
		log.Println("fastcgi [on]")
		log.Println("secure [off]")
		listener, err := net.Listen("tcp", *bind+":"+*port)
		if err != nil {
			log.Fatal("Could not bind: ", err)
		}
		log.Println("info: Listening on", *port)
		//	fcgi.Serve(listener, nil) // this works but without csrf..!
		fcgi.Serve(listener, csrf.Protect([]byte("LI80PNK1xcT01jmQBsEyxyrNCrbyyFPjPU8CKnxwmCruxNijgnyb3hXXD3p1RBc0+LIRQUUbTtis6hc6LD4I/A=="), csrf.HttpOnly(true), csrf.Secure(false))(r))
		//log.Fatal(fcgi.Serve( listener, csrf.Protect([]byte("LI80PNK1xcT01jmQBsEyxyrNCrbyyFPjPU8CKnxwmCruxNijgnyb3hXXD3p1RBc0+LIRQUUbTtis6hc6LD4I/A=="), csrf.HttpOnly(true), csrf.Secure(false))(r)))

	} else if *insecure == true {
		log.Println("info: Listening on", *port)
		log.Println("secure [off]")
		log.Fatal(http.ListenAndServe(":"+*port, csrf.Protect([]byte("LI80PNK1xcT01jmQBsEyxyrNCrbyyFPjPU8CKnxwmCruxNijgnyb3hXXD3p1RBc0+LIRQUUbTtis6hc6LD4I/A=="), csrf.HttpOnly(true), csrf.Secure(false))(r)))
	} else {
		log.Println("info: Listening on", *port)
		// Change this CSRF auth token in production!
		log.Println("secure [on]")
		log.Fatal(http.ListenAndServe(":"+*port, csrf.Protect([]byte("LI80PNK1xcT01jmQBsEyxyrNCrbyyFPjPU8CKnxwmCruxNijgnyb3hXXD3p1RBc0+LIRQUUbTtis6hc6LD4I/A=="), csrf.HttpOnly(true), csrf.Secure(true))(r)))
	}

}

// HomeHandler parses the ./templates/index.html template file.
// This returns a web page with a form, captcha, CSRF token, and the casgo API key to send the message.
func HomeHandler(w http.ResponseWriter, r *http.Request) {

	t, err := template.New("Index").ParseFiles("./templates/index.html")
	if err != nil {

		data := map[string]interface{}{
			"Key":            getKey(),
			csrf.TemplateTag: csrf.TemplateField(r),
			//		"Captcha":
		}

		t.ExecuteTemplate(w, "Index", data)
	} else {

		data := map[string]interface{}{
			"Key":            getKey(),
			csrf.TemplateTag: csrf.TemplateField(r),
			"CaptchaId":      captcha.New(),
		}

		t.ExecuteTemplate(w, "Index", data)

	}


	log.Printf("pre-contact: %s at %s", r.UserAgent(), r.RemoteAddr)
}

// LoveHandler is just for fun.
// I love lamp. This displays affection for r.URL.Path[1:]
func LoveHandler(w http.ResponseWriter, r *http.Request) {
	p := bluemonday.UGCPolicy()
	subdomain := getSubdomain(r)
	lol := p.Sanitize(r.URL.Path[1:])
	if subdomain == "" {
		fmt.Fprintf(w, "I love %s!", lol)
		log.Printf("I love %s says %s at %s", lol, r.UserAgent(), r.RemoteAddr)
	} else {
		fmt.Fprintf(w, "%s loves %s!", subdomain, lol)
		log.Printf("I love %s says %s at %s", subdomain, r.UserAgent(), r.RemoteAddr)
	}

}

// CustomErrorHandler allows casgo administrator to customize the 404 Error page
// Parses the ./templates/error.html file.
func CustomErrorHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("Error").ParseFiles("./templates/error.html")
	if err != nil {
		data := map[string]interface{}{
			"Key":            getKey(),
			csrf.TemplateTag: csrf.TemplateField(r),
		}
		t.ExecuteTemplate(w, "Error", data)
	} else {
		data := map[string]interface{}{
			"Key": getKey(),

			csrf.TemplateTag: csrf.TemplateField(r),
			"CaptchaId":      captcha.New(),
		}

		t.ExecuteTemplate(w, "Error", data)

	}

	log.Printf("error: %s at %s", r.UserAgent(), r.RemoteAddr)

}

// ContactHandler displays a contact form with CSRF and a Cookie. And maybe a captcha and drawbridge.
func ContactHandler(w http.ResponseWriter, r *http.Request) {

	t, err := template.New("Contact").ParseFiles("./templates/form.html")
	if err != nil {

		data := map[string]interface{}{

			"Key":            getKey(),
			csrf.TemplateTag: csrf.TemplateField(r),
		}

		t.ExecuteTemplate(w, "Contact", data)
	} else {

		data := map[string]interface{}{

			"Key":            getKey(),
			csrf.TemplateTag: csrf.TemplateField(r),
			"CaptchaId":      captcha.New(),
		}

		t.ExecuteTemplate(w, "Contact", data)

	}

	log.Printf("pre-contact: %s at %s", r.UserAgent(), r.RemoteAddr)

}

// RedirectHomeHandler redirects everyone home ("/") with a 301 redirect.
func RedirectHomeHandler(rw http.ResponseWriter, r *http.Request) {
	http.Redirect(rw, r, "/", 301)
}

// EmailHandler checks the Captcha string, and calls EmailSender
func EmailHandler(rw http.ResponseWriter, r *http.Request) {
	destination := casgoDestination
	var query url.Values
	if r.Method == "POST" {
		if !captcha.VerifyString(r.FormValue("captchaId"), r.FormValue("captchaSolution")) {
			fmt.Fprintf(rw, "You may be a robot. Can you go back and try again?")
			http.Redirect(rw, r, "/", 301)
		} else {
			r.ParseForm()
			query = r.Form
			EmailSender(rw, r, destination, query)
		}
	} else {
		http.Redirect(rw, r, "/", 301)
	}

}

// EmailSender always returns success for the visitor. This function needs some work.
func EmailSender(rw http.ResponseWriter, r *http.Request, destination string, query url.Values) {
	form := ParseQuery(query)
	if form.Email == "" {
		http.Redirect(rw, r, "/", 301)
		return
	}
	if sendEmail(destination, form) {
		fmt.Fprintln(rw, "<html><p>Thanks! Would you like to go <a href=\"/\">back</a>?</p></html>")
		http.Redirect(rw, r, "/", 301)
		log.Printf("SUCCESS-contact: %s at %s", r.UserAgent(), r.RemoteAddr)
	} else {
		log.Printf("Debug: %s at %s", form, destination)
		fmt.Fprintln(rw, "Uh-oh! Check your mandrill settings/api-logs!")
		log.Printf("FAIL-contact: %s at %s", r.UserAgent(), r.RemoteAddr)
	}
}

func ParseQuery(query url.Values) *Form {
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
		} else if k == "message" {
			form.Message = k + ": " + v[0] + "<br>\n"
		} else {
			additionalFields = additionalFields + k + ": " + v[0] + "<br>\n"
		}
	}
	if form.Subject == "" {
		form.Subject = "You have mail!"
	}
	if additionalFields != "" {
		if form.Message == "" {
			form.Message = form.Message + "Message:\n<br>" + additionalFields
		} else {
			form.Message = form.Message + "\n<br>Additional:\n<br>" + additionalFields
		}
	}
	return form
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

//getKey returns the current instance's API key as string
func getKey() string {
	return casgoAPIKey
}

//OpenLogFile switches the log engine to a file, rather than stdout
func OpenLogFile() {
	f, err := os.OpenFile("./casgo.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Fatal("error opening file: %v", err)
		os.Exit(1)
	}
	log.SetOutput(f)
}
