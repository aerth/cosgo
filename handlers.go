package main


import (

	"fmt"
// soon...
//	"github.com/gorilla/csrf"
	"log"
	http "net/http"
	"net/url"
	"strings"
	"html/template"
)


// Routing URL handlers

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	 fmt.Fprint(w, "")
	// http.ServeFile("./templates/form.html")
}

// I love lamp. This displays affection for r.URL.Path[1:]

func LoveHandler(w http.ResponseWriter, r *http.Request) {
	 fmt.Fprintf(w, "I love %s!", r.URL.Path[1:])
	 log.Printf("I love %s says %s at %s", r.URL.Path[1:], r.UserAgent(), r.RemoteAddr)
}

// Display contact form with CSRF and a Cookie. And maybe a captcha and drawbridge.
func ContactHandler(w http.ResponseWriter, r *http.Request) {
		//w.Header().Set("X-CSRF-Token", csrf.Token(r))
		var key string
		//var err string
		key = getKey()
		t, err := template.New("Contact").ParseFiles("./templates/form.html")
		if err != nil {

		t.ExecuteTemplate(w, "Contact", key,)
		}else{
			t.ExecuteTemplate(w, "Contact", key,)
		}
		// log.Println(t.ExecuteTemplate(w, "Contact", key,))

	 log.Printf("pre-contact: %s at %s", r.UserAgent(), r.RemoteAddr)
}

// Redirect everything /
func RedirectHomeHandler(rw http.ResponseWriter, r *http.Request) {
	http.Redirect(rw, r, "/", 301)
}


// Uses environmental variable on launch to determine Destination
func EmailHandler(rw http.ResponseWriter, r *http.Request) {
	destination := casgoDestination
	var query url.Values
//	if r.Method == "GET" {
//		query = r.URL.Query()
//	} else if r.Method == "POST" {
	if r.Method == "POST" {
		r.ParseForm()
		query = r.Form
	} else {
		fmt.Fprintln(rw, "Please submit via POST.")
	}
	EmailSender(rw, r, destination, query)

}


// Will introduce success/fail in the templates soon!
func EmailSender(rw http.ResponseWriter, r *http.Request, destination string, query url.Values) {
	form := ParseQuery(query)
	if form.Email == "" {
		http.Redirect(rw, r, "/", 301)
		return
	}
	if sendEmail(destination, form) {
		fmt.Fprintln(rw, "Success! Check your inbox!")
		log.Printf("SUCCESS-contact: %s at %s", r.UserAgent(), r.RemoteAddr)
	} else {
		log.Printf("debug: %s at %s", form, destination)
		fmt.Fprintln(rw, "Uh-oh! Check your mandrill settings/api-logs!")
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
		//} else if (k == "name") {
		//	form.Name = v[0]
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
