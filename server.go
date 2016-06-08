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
	var status, reason string
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
				reason = "Bad error!"
			}
			status = "Your message was not sent: " + reason
		}
	}
	thyme := time.Now()
	nowtime := thyme.Format("Mon Jan 2 15:04:05 2006")
	uptime := time.Since(timeboot).String()

	t, templateerr := template.New("Index").ParseFiles(templateDir + "index.html")
	if templateerr != nil {
		// Something happened to the template since booting successfully. Must be 100% correct HTML syntax.
		log.Println("Almost fatal")
		log.Println(templateerr)
		fmt.Fprintf(w, "We are experiencing some technical difficulties. Please come back soon!")
	} else {
		data := map[string]interface{}{
			"Now":            nowtime,
			"Status":         status,
			"Version":        version,
			"Hits":           hitcounter,
			"Uptime":         uptime,
			"Key":            getKey(),
			csrf.TemplateTag: csrf.TemplateField(r),
			"CaptchaId":      captcha.NewLen(CaptchaLength + rand.Intn(CaptchaVariation)),
		}
		t.ExecuteTemplate(w, "Index", data)

	}
}

// customErrorHandler allows cosgo administrator to customize the 404 Error page
// Using the ./templates/error.html file.
func customErrorHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("visitor: 404 %s - %s at %s", r.Host, r.UserAgent(), r.RemoteAddr)
	p := bluemonday.UGCPolicy()
	domain := getDomain(r)
	cleanURL := p.Sanitize(r.URL.Path[1:])
	log.Printf("404 on %s/%s", cleanURL, domain)
	t, err := template.New("Error").ParseFiles(templateDir + "error.html")
	if err == nil {
		data := map[string]interface{}{
			"err":            "404",
			"Key":            getKey(),
			csrf.TemplateTag: csrf.TemplateField(r),
		}
		t.ExecuteTemplate(w, "Error", data)
	} else {
		log.Printf("template error: %q at %s", r.UserAgent(), r.RemoteAddr)
		log.Println(err)
		http.Redirect(w, r, "/", 301)
	}
}

// redirecthomeHandler redirects everyone home ("/") with a 301 redirect.
func redirecthomeHandler(rw http.ResponseWriter, r *http.Request) {
	p := bluemonday.UGCPolicy()
	domain := getDomain(r)
	cleanURL := p.Sanitize(r.URL.Path[1:])
	log.Printf("%q from %s hit %q on domain: %q", r.UserAgent(), r.RemoteAddr, cleanURL, domain)
	http.Redirect(rw, r, "/", 301)

}

// make sure user submitted a POST request
func verifyMethod(r *http.Request) bool {
	if r.Method != "POST" {
		return false
	}
	return true
}

func srvError(r *http.Request, rw http.ResponseWriter, e int) {
	http.Redirect(rw, r, "/?status=0&r=1#contact", http.StatusFound)
	return
	log.Fatalln("[srvError] Couldn't redirect.")
}

func verifyKey(r *http.Request) bool {

	// userkey is the user submitted URL key
	userkey := strings.TrimLeft(r.URL.Path, "/")
	userkey = strings.TrimRight(userkey, "/send")
	// Check URL Key
	if *debug {
		log.Printf("\nComparing... \n\t" + userkey + "\n\t" + cosgo.PostKey)
	}
	if userkey != cosgo.PostKey {
		log.Println("Key Mismatch. ", r.UserAgent(), r.RemoteAddr, r.RequestURI+"\n")
		return false
	}

	return true
}

func verifyCaptcha(r *http.Request) bool {
	if !captcha.VerifyString(r.FormValue("captchaId"), r.FormValue("captchaSolution")) {
		log.Printf("User Error: CAPTCHA %s at %s", r.UserAgent(), r.RemoteAddr)
		return false
	}

	return true
}

// emailHandler checks the Captcha string, and the POST key, and sends on its way.
func emailHandler(rw http.ResponseWriter, r *http.Request) {
	if !verifyMethod(r) {
		srvError(r, rw, 1)
		return
	}

	if !verifyKey(r) {
		srvError(r, rw, 2)
		return
	}

	if !verifyCaptcha(r) {
		srvError(r, rw, 3)
		return
	}

	if *debug {
		log.Printf("Human Visitor: %s at %s", r.UserAgent(), r.RemoteAddr)
	}

	r.ParseForm()

	// normalize and validate email
	mailform := mbox.ParseFormGPG(destinationEmail, r.Form, publicKey)

	// quick offline email address validation
	if mailform.Email == "@" || mailform.Email == " " || !strings.ContainsAny(mailform.Email, "@") || !strings.ContainsAny(mailform.Email, ".") {
		srvError(r, rw, 4)
		return
	}

	if *sendgridKey != "" {
		ok, str := sendgridder(destinationEmail, mailform)
		if str != "" {
			log.Println(str)
		}
		if !ok {
			log.Printf("FAILURE-contact: %q at %s\n\t%q", r.UserAgent(), r.RemoteAddr, r.Form)
			srvError(r, rw, 5)
			return
		}

	} else {
		// save the message
		saving := mbox.Save(mailform)
		if saving != nil {
			log.Println(saving)
			srvError(r, rw, 6)
			return
		}

	}

	log.Printf("SUCCESS-contact: %s at %s", r.UserAgent(), r.RemoteAddr)
	srvSuccess(r, rw, 1)
}

func srvSuccess(r *http.Request, rw http.ResponseWriter, status int) {
	http.Redirect(rw, r, "/?status=1", http.StatusFound)
}

// serverSingle just shows one file.
func serveSingle(pattern string, filename string) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, filename, time.Now(), nil)
	})
}
