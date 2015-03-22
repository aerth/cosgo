package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"net/url"
	"strings"
	"os"
)

var (
	mandrillApiUrl string
	mandrillKey    string
)

func main() {
	// Get Port
	port := flag.String("port", "8080", "HTTP Port to listen on")
	flag.Parse()

	// Get Mandrill Keys
	mandrillApiUrl = "https://mandrillapp.com/api/1.0/"
	mandrillKey = os.Getenv("MANDRILL_KEY")
	if mandrillKey == "" {
		log.Fatal("MANDRILL_KEY envrionment variable is not set.")
		os.Exit(1)
	}

	// Serve
	log.Println("Starting Server on", *port)
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/{email}", EmailHandler)
	http.Handle("/", r)
	http.ListenAndServe(":8080", r)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}

func HomeHandler(rw http.ResponseWriter, r *http.Request) {
	http.Redirect(rw, r, "http://www.github.com/munrocape/staticcontact", 301)
}

func EmailHandler(rw http.ResponseWriter, r *http.Request) {
	Log.Info("Sending email")
	query := r.URL.Query()
	destination := r.URL.Path[1:]
	if r.Method == "GET" {
		EmailGetHandler(rw, r, destination, query)
	} else if r.Method == "POST" {
		fmt.Fprintln(rw, "POST")
	} else {
		fmt.Fprintln(rw, "Please submit via GET or POST. See www.staticcontact.com for further instructions.")
	}
}

func EmailGetHandler(rw http.ResponseWriter, r *http.Request, destination string, query url.Values) {
	form := ParseQuery(query)
	if (form.Email == "") {
		fmt.Fprintln(rw, "Please provide a sender email.")
		return
	}
	sendEmail(destination, form)
}

func ParseQuery(query url.Values) *Form {
	form := new(Form)
	additionalFields := ""
	for k, v := range query {
		k = strings.ToLower(k)
		if (k == "email"){
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
	if (form.Subject == ""){
		form.Subject = "New Form Submission!"
	}
	if (additionalFields != "") {
		if (form.Message == ""){
			form.Message = form.Message + "The following fields were also entered:\n<br>" + additionalFields
		} else {
			form.Message = form.Message + "\n<br>The following additional fields were also entered:\n<br>" + additionalFields
		}
	}
	return form
}

func EmailPostHandler(rw http.ResponseWriter, r *http.Request) {

}
