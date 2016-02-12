package main

import (
	"fmt"
	// soon...
	"github.com/gorilla/csrf"
	"html/template"
	"log"
	http "net/http"
	"net/url"
	"strings"
)

var precenter string = `<!-- casgo is free software -->
<div style="margin-top: 10% margin-bottom: 10%; width: 100%; max-width: 100%; text-align: center;">
<a href="https://github.com/aerth/casgo">
<img alt="casgo v0.1beta https://github.com/aerth/casgo" src="`

var casgologo string = `data:image/png;base64,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`

var poscenter string = `">
</a>
</div>`

// Routing URL handlers

func HomeHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Fprint(w, "<!DOCTYPE html><html>" + precenter + casgologo + poscenter + "</html>")


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
	//var key string
	//var err string
	//key = getKey()

	t, err := template.New("Contact").ParseFiles("./templates/form.html")
	//	t = t.Funcs(template.FuncMap{"Key": key})
	//	t = t.Funcs(template.FuncMap{csrf.TemplateTag: csrf.TemplateField(r)})
	if err != nil {
		//	p := Person{Key: key,
		//        csrf.TemplateTag: csrf.TemplateField(r),
		//			}

		data := map[string]interface{}{
			"Key":            getKey(),
			csrf.TemplateTag: csrf.TemplateField(r),
			//	 "Context": &Context{true}, // Set to false will prevent addClassIfActive to print
		}

		t.ExecuteTemplate(w, "Contact", data)
	} else {

		data := map[string]interface{}{
			"Logo":            casgologo,
			"Key":            getKey(),
			csrf.TemplateTag: csrf.TemplateField(r),
			//	 "Context": &Context{true}, // Set to false will prevent addClassIfActive to print
		}

		t.ExecuteTemplate(w, "Contact", data)
		// t.ExecuteTemplate(w, "Contact", key)
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
		//fmt.Fprintln(rw, "Please submit via POST.")
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

var subdomain string

func getSubdomain(rw http.ResponseWriter, r *http.Request, query url.Values) string {
	prefix := strings.Split(r.Host, ".")

	return prefix[0]

}
