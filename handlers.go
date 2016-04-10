package main

import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aerth/cosgo/mbox"
	"github.com/dchest/captcha"
	"github.com/gorilla/csrf"
	"github.com/microcosm-cc/bluemonday"
)

// homeHandler parses the ./templates/index.html template file.
// This returns a web page with a themeable form, captcha, CSRF token, and the cosgo API key to send the message.
func homeHandler(w http.ResponseWriter, r *http.Request) {
	if *debug {
		log.Printf("Visitor: %s %s %s %s", r.UserAgent(), r.RemoteAddr, r.Host, r.RequestURI)
	}
	thyme := time.Now()
	nowtime := thyme.Format("Mon Jan 2 15:04:05 2006")
	t, err := template.New("Index").ParseFiles("./templates/index.html")
	if err != nil {
		// Do Something
		log.Println("Almost fatal: Cant load index.html template!")
		log.Println(err)
		fmt.Fprintf(w, "We are experiencing some technical difficulties. Please come back soon!")
	} else {
		data := map[string]interface{}{
			"Now":            nowtime,
			"Key":            getKey(),
			csrf.TemplateTag: csrf.TemplateField(r),
			"CaptchaId":      captcha.NewLen(CaptchaLength + rand.Intn(CaptchaVariation)),
		}
		t.ExecuteTemplate(w, "Index", data)

	}
}

// loveHandler is just for fun.
// I love lamp. This displays affection for r.URL.Path[1:]
func loveHandler(w http.ResponseWriter, r *http.Request) {

	p := bluemonday.UGCPolicy()
	subdomain := getSubdomain(r)
	lol := p.Sanitize(r.URL.Path[1:])
	if r.Method != "GET" {
		log.Printf("Something tried %s on %s", r.Method, lol)
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

// customErrorHandler allows cosgo administrator to customize the 404 Error page
// Using the ./templates/error.html file.
func customErrorHandler(w http.ResponseWriter, r *http.Request) {
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

// redirecthomeHandler redirects everyone home ("/") with a 301 redirect.
func redirecthomeHandler(rw http.ResponseWriter, r *http.Request) {
	p := bluemonday.UGCPolicy()
	domain := getDomain(r)
	lol := p.Sanitize(r.URL.Path[1:])
	if *debug {
		log.Printf("Redirecting %s back home on %s", lol, domain)
	}
	http.Redirect(rw, r, "/", 301)

}

// redirecthomeHandler redirects everyone home ("/") with a 301 redirect.
func pageHandler(rw http.ResponseWriter, r *http.Request) {
	p := bluemonday.UGCPolicy()
	domain := getDomain(r)
	lol := p.Sanitize(r.URL.Path[1:])
	if *debug {
		log.Printf("Page %s %s", lol, domain)
	}

}

// emailHandler checks the Captcha string, and the POST key, and sends on its way.
func emailHandler(rw http.ResponseWriter, r *http.Request) {

	destination := cosgoDestination
	var query url.Values
	ourpath := strings.TrimLeft(r.URL.Path, "/")
	ourpath = strings.TrimRight(ourpath, "/send")

	// Check POST
	if r.Method != "POST" {
		log.Printf("\nNot a POST request... \n\t" + r.RemoteAddr + r.RequestURI + r.UserAgent())
		fmt.Fprintln(rw, "<html><p>What are we doing here?</p></html>")
		return
	}

	// Check POST Key
	if *debug {
		log.Printf("\nComparing... \n\t" + ourpath + "\n\t" + cosgo.PostKey)
	}
	if ourpath != cosgo.PostKey {
		log.Println("Key Mismatch. ", r.UserAgent(), r.RemoteAddr, r.RequestURI+"\n")
		fmt.Fprintln(rw, "<html><p>What are we doing here? If you waited too long to send the form, try again. <a href=\"/\">Go back</a>?</p></html>")
		return
	}

	// Method is POST, URL KEY is correct. Check CAPTCHA.
	if !captcha.VerifyString(r.FormValue("captchaId"), r.FormValue("captchaSolution")) {
		fmt.Fprintln(rw, "<html><p>You may be a robot! Please go <a href=\"/\">back</a> and try again!</p></html>")
		log.Printf("User Error: CAPTCHA %s at %s", r.UserAgent(), r.RemoteAddr)
		return
	}

	// Captcha is correct. POST key is correct.
	if *debug {
		log.Printf("User Human: %s at %s", r.UserAgent(), r.RemoteAddr)
		log.Printf("Key Match:\n\t%s\n\t%s", ourpath, cosgo.PostKey)
	}
	r.ParseForm()

	// Given the circumstances, you would think the form is ready.
	query = r.Form

	// Switch mailmode and send it out! Success message may change/be customized in the future.
	switch *mailmode {
	case smtpmandrill:
		mandrillSender(rw, r, destination, query)
		log.Printf("SUCCESS-contact: %s at %s", r.UserAgent(), r.RemoteAddr)
		fmt.Fprintln(rw, "<html><p>Thanks! Would you like to go <a href=\"/\">back</a>?</p></html>")
		return

	case smtpsendgrid:
		err = sendgridSender(rw, r, destination, query)
		if err != nil {
			log.Printf("FAILURE-contact: %s at %s\n\t%s", r.UserAgent(), r.RemoteAddr, query)
			fmt.Fprintln(rw, "<html><p>Thanks! But your email was not sent. Would you like to go <a href=\"/\">back</a>?</p></html>")
			return
		}
		log.Printf("SUCCESS-contact: %s at %s", r.UserAgent(), r.RemoteAddr)
		fmt.Fprintln(rw, "<html><p>Thanks! Would you like to go <a href=\"/\">back</a>?</p></html>")
		return
	default:
		err = mbox.Save(rw, r, destination, query)
		if err != nil {
			log.Printf("FAILURE-contact: %s at %s\n\t%s", r.UserAgent(), r.RemoteAddr, query)
			fmt.Fprintln(rw, "<html><p>Your email was not sent. Would you like to go <a href=\"/\">back</a>?</p></html>")
			return
		}
		log.Printf("SUCCESS-contact: %s at %s", r.UserAgent(), r.RemoteAddr)
		fmt.Fprintln(rw, "<html><p>Thanks! Would you like to go <a href=\"/\">back</a>?</p></html>")
		return

	}

}

// serverSingle just shows one file.
func serveSingle(pattern string, filename string) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filename)
	})
}
