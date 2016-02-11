package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var (
	mandrillApiUrl string
	mandrillKey    string
)

func main() {


	port := flag.String("port", "8080", "HTTP Port to listen on")
	flag.Parse()

	mandrillApiUrl = "https://mandrillapp.com/api/1.0/"
	mandrillKey = os.Getenv("MANDRILL_KEY")
	if mandrillKey == "" {
		log.Fatal("MANDRILL_KEY is Crucial.")
		os.Exit(1)
	}

	log.Println("Starting Server on", *port)
	log.Println("Switching Logs to debug.log")
	OpenLog()
	log.Println("Server on", *port)
	r := mux.NewRouter()

	r.NotFoundHandler = http.HandlerFunc(RedirectHomeHandler)
	r.HandleFunc("/", HomeHandler)

	r.HandleFunc("/contact", ContactHandler)
	r.HandleFunc("/contact/", ContactHandler)

	r.HandleFunc("/{whatever}", LoveHandler)

	r.HandleFunc("/em/{email}/", EmailHandler)
	http.Handle("/", r)
	//http.NotFound() NotFoundHandler()
	http.ListenAndServe(":"+*port, r)

}


// This function opens a log file. "debug.log"

func OpenLog(){
f, err := os.OpenFile("./debug.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
if err != nil {
    log.Fatal("error opening file: %v", err)
		log.Fatal("MANDRILL_KEY is Crucial.")
		os.Exit(1)
}
//defer log.Close()
log.SetOutput(f)
}

// This is the home page it is blank. "This server is broken"

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	 fmt.Fprint(w, "")
}

// I love lamp. This displays affection for r.URL.Path[1:]

func LoveHandler(w http.ResponseWriter, r *http.Request) {
	 fmt.Fprintf(w, "I love %s!", r.URL.Path[1:])
	 log.Printf("I love %s says %s at %s", r.URL.Path[1:], r.UserAgent(), r.RemoteAddr)
}

// Display contact form with CSRF and a Cookie. And maybe a captcha and drawbridge.
func ContactHandler(w http.ResponseWriter, r *http.Request) {
	 fmt.Fprint(w, "Contact Form is moving in right here")
	 log.Printf("pre-contact: %s at %s", r.UserAgent(), r.RemoteAddr)
}

// Redirect everything /
func RedirectHomeHandler(rw http.ResponseWriter, r *http.Request) {
	http.Redirect(rw, r, "/", 301)
}


// Uses mandrillapp.com default sender address.
func EmailHandler(rw http.ResponseWriter, r *http.Request) {
	destination := r.URL.Path[2:]
	var query url.Values
	if r.Method == "GET" {
		query = r.URL.Query()
	} else if r.Method == "POST" {
		r.ParseForm()
		query = r.Form
	} else {
		fmt.Fprintln(rw, "Please submit via GET or POST.")
	}
	EmailSender(rw, r, destination, query)

}


func EmailSender(rw http.ResponseWriter, r *http.Request, destination string, query url.Values) {
	form := ParseQuery(query)
	if form.Email == "" {
		http.Redirect(rw, r, "/", 301)
		return
	}
	if sendEmail(destination, form) {
		fmt.Fprintln(rw, "0 Success! Your message has been delivered.")
		log.Printf("SUCCESS-contact: %s at %s", r.UserAgent(), r.RemoteAddr)
	} else {
		fmt.Fprintln(rw, "1 Uh-oh! We were unable to deliver your message. Please confirm that you entered a valid email address.")
		log.Printf("FAIL-contact: %s at %s", r.UserAgent(), r.RemoteAddr)
	}
}

func ParseQuery(query url.Values) *Form {
	form := new(Form)
	additionalFields := ""
	for k, v := range query {
		k = strings.ToLower(k)
		if (k == "email") {
			form.Email = v[0]
		} else if (k == "name") {
			form.Name = v[0]
		} else if (k == "subject") {
			form.Subject = v[0]
		} else if (k == "message") {
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
