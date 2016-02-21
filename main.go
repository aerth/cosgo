package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/url"
	"net/http"
	"net/http/fcgi"
	"os"
	"time"
	"html/template"
	"strings"
	"github.com/dchest/captcha"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/microcosm-cc/bluemonday"
)

var (
	mandrillApiUrl   string
	mandrillKey      string
	cosgoDestination string
	cosgoAPIKey      string
)

type C struct {
	CaptchaId string
}

const (
	// Default number of digits in captcha solution.
	DefaultLen = 4
	// The number of captchas created that triggers garbage collection used
	// by default store.
	CollectNum = 100
	// Expiration time of captchas used by default store.
	Expiration = 10 * time.Minute
)

const (
	// Standard width and height of a captcha image.
	StdWidth  = 240
	StdHeight = 120
)

func main() {

	// Copyright 2016 aerth and contributors. Source code at https://github.com/aerth/cosgo
	// You should recieve a copy of the MIT license with this software.
	log.Println("\n\n\tcosgo v0.4\n\tCopyright 2016 aerth\n\tSource code at https://github.com/aerth/cosgo\n\n")

	// Set flags from command line
	port := flag.String("port", "8080", "HTTP Port to listen on")
	Debug := flag.Bool("debug", false, "be verbose, dont switch to cosgo.log")
	insecure := flag.Bool("insecure", false, "accept insecure cookie transfer (http/80)")
	mailbox := flag.Bool("mailbox", false, "disable mandrill send")
	fastcgi := flag.Bool("fastcgi", false, "use fastcgi with nginx")
	static := flag.Bool("static", true, "use -static=false to disable")
	redirect := flag.Bool("redirect", false, "disable error.html template")
	bind := flag.String("bind", "127.0.0.1", "default: 127.0.0.1 - maybe 0.0.0.0 ?")
	flag.Parse()


	mandrillApiUrl = "https://mandrillapp.com/api/1.0/"

	// For backwards compatibility
if os.Getenv("CASGO_API_KEY") != "" && os.Getenv("COSGO_API_KEY") == "" {
os.Setenv("COSGO_API_KEY",os.Getenv("CASGO_API_KEY"))
			log.Println("Please use COSGO_API_KEY...")
}
if os.Getenv("CASGO_DESTINATION") != "" && os.Getenv("COSGO_DESTINATION") == "" {
os.Setenv("COSGO_DESTINATION",os.Getenv("CASGO_DESTINATION"))
			log.Println("Please use COSGO_DESTINATION...")
}


	// Test environmental variables, if we aren't in -mailbox mode.
	if *mailbox != true {
		QuickSelfTest()
	}


		//Print API Key
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

	log.Printf("cosgo is booting up on "+getLink(*fastcgi, *bind, *port))
	if *fastcgi == true {
		log.Printf("[fastcgi mode on]")
	}

//Begin Routing
	r := mux.NewRouter()

		if *redirect == true {
			r.NotFoundHandler = http.HandlerFunc(RedirectHomeHandler)
		}else{
			r.NotFoundHandler = http.HandlerFunc(CustomErrorHandler)
		}

//	r.NotFoundHandler = http.HandlerFunc(CustomErrorHandler)
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/"+cosgoAPIKey+"/form", ContactHandler)
	r.HandleFunc("/"+cosgoAPIKey+"/form/", ContactHandler)
	r.HandleFunc("/"+cosgoAPIKey+"/send", EmailHandler)

	if *static == true 	{
			s := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
			ss := http.FileServer(http.Dir("./static/"))
			// Serve /static folder and favicon etc
			r.Path("/favicon.ico").Handler(ss)
			r.Path("/robots.txt").Handler(ss)
			r.Path("/sitemap.xml").Handler(ss)
			r.Path("/static/{dir}/{whatever}.css").Handler(s)
			r.Path("/static/{dir}/{whatever}.js").Handler(s)
	}
//	r.HandleFunc("/{whatever}", LoveHandler)

	r.HandleFunc("/{whatever}", RedirectHomeHandler)

	// Retrieve Captcha IMG and WAV
	r.Methods("GET").PathPrefix("/captcha/").Handler(captcha.Server(captcha.StdWidth, captcha.StdHeight))

	//http.NotFoundHandler = r.HandlerFunc(CustomErrorHandler)
	http.Handle("/", r)
	//End Routing
	// Switch to file log so we can ctrl+c and launch another instance :)
	if *mailbox == true {
		log.Println("mailbox mode: not enabled just saying")
		//CreateMailBox()
	}

	if *Debug == false {
		log.Println("quiet mode: [switching logs to cosgo.log]")
		OpenLogFile()
	} else {
		log.Println("Debug on: [not using cosgo.log]")
	}

log.Printf("cosgo is live on "+getLink(*fastcgi, *bind, *port))
// Start Serving!
	if *fastcgi == true {
			listener, err := net.Listen("tcp", *bind+":"+*port)
			if err != nil {
					log.Fatal("Could not bind: ", err)
		}
		if *insecure == true {
					log.Fatal(fcgi.Serve(listener, csrf.Protect([]byte("LI80PNK1xcT01jmQBsEyxyrNCrbyyFPjPU8CKnxwmCruxNijgnyb3hXXD3p1RBc0+LIRQUUbTtis6hc6LD4I/A=="), csrf.HttpOnly(true), csrf.Secure(false))(r)))
				}else {
					log.Println("info: https:// only")
					log.Fatal(fcgi.Serve(listener, csrf.Protect([]byte("LI80PNK1xcT01jmQBsEyxyrNCrbyyFPjPU8CKnxwmCruxNijgnyb3hXXD3p1RBc0+LIRQUUbTtis6hc6LD4I/A=="), csrf.HttpOnly(true), csrf.Secure(true))(r)))
				}
	} else if *fastcgi == false && *insecure == true {
			log.Fatal(http.ListenAndServe(":"+*port, csrf.Protect([]byte("LI80PNK1xcT01jmQBsEyxyrNCrbyyFPjPU8CKnxwmCruxNijgnyb3hXXD3p1RBc0+LIRQUUbTtis6hc6LD4I/A=="), csrf.HttpOnly(true), csrf.Secure(false))(r)))
	} else if *fastcgi == false && *insecure == false {
					log.Println("info: https:// only")
					// Change this CSRF auth token in production!
					log.Fatal(http.ListenAndServe(":"+*port, csrf.Protect([]byte("LI80PNK1xcT01jmQBsEyxyrNCrbyyFPjPU8CKnxwmCruxNijgnyb3hXXD3p1RBc0+LIRQUUbTtis6hc6LD4I/A=="), csrf.HttpOnly(true), csrf.Secure(true))(r)))
		}

}

// End main function

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

func QuickSelfTest(){
	log.Println("Starting self test...")
	mandrillKey = os.Getenv("MANDRILL_KEY")
	if mandrillKey == ""{
		log.Fatal("Fatal: MANDRILL_KEY is Crucial.\nHint: export MANDRILL_KEY=123456789")
		os.Exit(1)
	}
	cosgoDestination = os.Getenv("COSGO_DESTINATION")
	if cosgoDestination == "" {
		log.Fatal("Fatal: COSGO_DESTINATION is Crucial.\nHint: export COSGO_DESTINATION=\"your@email.com\"")
		os.Exit(1)
	}
	_, err := template.New("Index").ParseFiles("./templates/index.html")
	if err != nil {
		log.Fatal("Fatal: Template Error\nHint: Copy ./templates and ./static from $GOPATH/src/github.com/aerth/cosgo/ to the location of your binary.")
	}
	log.Println("Passed self test.")
}

// HomeHandler parses the ./templates/index.html template file.
// This returns a web page with a form, captcha, CSRF token, and the cosgo API key to send the message.
func HomeHandler(w http.ResponseWriter, r *http.Request) {

	t, err := template.New("Index").ParseFiles("./templates/index.html")
	if err != nil {
		// Do Something
		log.Println(err)

	} else {
							data := map[string]interface{}{
								"Key":            getKey(),
								csrf.TemplateTag: csrf.TemplateField(r),
								"CaptchaId":      captcha.New(),
		}
		t.ExecuteTemplate(w, "Index", data)
	}
	log.Printf("visitor: %s - %s - %s", r.UserAgent(), r.RemoteAddr, r.Host)
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

// CustomErrorHandler allows cosgo administrator to customize the 404 Error page
// Parses the ./templates/error.html file.
func CustomErrorHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("Error").ParseFiles("./templates/error.html")
	if err == nil {
		data := map[string]interface{}{
			"Key":            getKey(),
			csrf.TemplateTag: csrf.TemplateField(r),
		}
		t.ExecuteTemplate(w, "Error", data)
	}else
	{
	log.Printf("template error: %s at %s", r.UserAgent(), r.RemoteAddr)
	log.Println(err)
}
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

	p := bluemonday.UGCPolicy()
	domain := getDomain(r)
	lol := p.Sanitize(r.URL.Path[1:])
	log.Printf("Redirecting %s back home on %s", lol, domain)
	http.Redirect(rw, r, "/", 301)

}

// EmailHandler checks the Captcha string, and calls EmailSender
func EmailHandler(rw http.ResponseWriter, r *http.Request) {
	destination := cosgoDestination
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

//getKey returns the current instance's API key as string
//func getKey() string {
//	return cosgoAPIKey
//}

//OpenLogFile switches the log engine to a file, rather than stdout
func OpenLogFile() {
	f, err := os.OpenFile("./cosgo.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Fatal("error opening file: %v", err)
		os.Exit(1)
	}
	log.SetOutput(f)
}

func getLink(fastcgi bool, bind string, port string) string {
	if fastcgi == true {
		link := bind+":"+port
		return link
	}else{
		link := "http://"+bind+":"+port
		return link
	}
}
