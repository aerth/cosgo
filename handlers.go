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

// homeHandler parses the ./templates/index.html template file. (or /templates/form.html)
// This returns a web page with a themeable form, captcha, CSRF token, and the cosgo API key to send the message.
func homeHandler(w http.ResponseWriter, r *http.Request) {
	hitcounter = hitcounter + 1
	if !*quiet {
		log.Printf("Visitor: %s %s %s %s", r.UserAgent(), r.RemoteAddr, r.Host, r.RequestURI)
	}

	query, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		fmt.Println(err)
		return
	}
	status := ""
	reason := ""

	if query.Get("status") != "" {
		if query["status"][0] == "1" {
			status = "Thanks! Your message was sent."
		}
	}
	if query.Get("r") != "" {
		if query["status"][0] == "0" {
			switch query["r"][0] {
			default:
				reason = "Error."
			case "1":
				reason = "Bad method."
			case "2":
				reason = "Bad endpoint."
			case "3":
				reason = "Bad capcha."
			case "4":
				reason = "Bad email address."

			case "5":
				reason = "Bad message."

			case "6":
				reason = "Bad error?"
			}
			status = "Your message was not sent: " + reason
		}
	}
	thyme := time.Now()
	nowtime := thyme.Format("Mon Jan 2 15:04:05 2006")

	t, templateerr := template.New("Index").ParseFiles("./templates/index.html")
	if !*form {
		t, templateerr = template.New("Index").ParseFiles("./templates/index.html")
	}
	if templateerr != nil {
		// Do Something
		log.Println("Almost fatal: Cant load index.html template!")
		log.Println(err)
		fmt.Fprintf(w, "We are experiencing some technical difficulties. Please come back soon!")
	} else {
		data := map[string]interface{}{
			"Now":            nowtime,
			"Status":         status,
			"Pages":          *pages,
			"PagePath":       *custompages,
			"Hits":           hitcounter,
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
		http.Redirect(rw, r, "/?status=0&r=1#contact", http.StatusFound)
		return
	}

	// Check POST Key
	if *debug {
		log.Printf("\nComparing... \n\t" + ourpath + "\n\t" + cosgo.PostKey)
	}
	if ourpath != cosgo.PostKey {
		log.Println("Key Mismatch. ", r.UserAgent(), r.RemoteAddr, r.RequestURI+"\n")
		http.Redirect(rw, r, "/?status=0&r=2#contact", http.StatusFound)
		return
	}

	// Method is POST, URL KEY is correct. Check CAPTCHA.
	if !captcha.VerifyString(r.FormValue("captchaId"), r.FormValue("captchaSolution")) {
		log.Printf("User Error: CAPTCHA %s at %s", r.UserAgent(), r.RemoteAddr)
		http.Redirect(rw, r, "/?status=0&r=3#contact", http.StatusFound)
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
	form := parseQuery(query)
	if form.Email == "@" || form.Email == " " || !strings.ContainsAny(form.Email, "@") || !strings.ContainsAny(form.Email, ".") {
		http.Redirect(rw, r, "/?status=0&r=4#contact ", http.StatusFound)
		return
	}
	// Switch mailmode and send it out! Success message may change/be customized in the future.
	switch *mailmode {
	case smtpmandrill:
		mandrillSender(rw, r, destination, query)
		log.Printf("SUCCESS-contact: %s at %s", r.UserAgent(), r.RemoteAddr)
		http.Redirect(rw, r, "/?status=1", http.StatusFound)
		return

	case smtpsendgrid:
		err = sendgridSender(rw, r, destination, query)
		if err != nil {
			log.Printf("FAILURE-contact: %s at %s\n\t%s", r.UserAgent(), r.RemoteAddr, query)
			http.Redirect(rw, r, "/?status=0&r=5#contact ", http.StatusFound)
			return
		}
		log.Printf("SUCCESS-contact: %s at %s", r.UserAgent(), r.RemoteAddr)
		rw.Header().Add("status", "success")
		homeHandler(rw, r)
		return
	default:
		err = mbox.Save(rw, r, destination, query)
		if err != nil {
			log.Printf("FAILURE-contact: %s at %s\n\t%s %s", r.UserAgent(), r.RemoteAddr, query, err.Error())
			http.Redirect(rw, r, "/?status=0&r=5#contact", http.StatusFound)

			return
		}
		log.Printf("SUCCESS-contact: %s at %s", r.UserAgent(), r.RemoteAddr)
		http.Redirect(rw, r, "/?status=1", http.StatusFound)
		return

	}

}

// serverSingle just shows one file.
func serveSingle(pattern string, filename string) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, filename, time.Now(), nil)
	})
}
