package main

import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aerth/mbox"

	"github.com/dchest/captcha"
	"github.com/gorilla/csrf"
	"github.com/microcosm-cc/bluemonday"
)

var funcMap = template.FuncMap{
	"timesince": timesince,
}

func (c *Cosgo) pubkeyHandler(w http.ResponseWriter, r *http.Request) {
	if c.publicKey == nil {
		redirecthomeHandler(w, r)
		return
	}
	w.Write(c.publicKey)
}

func (c *Cosgo) homeHandler(w http.ResponseWriter, r *http.Request) {
	hitcounter = hitcounter + 1
	c.Visitors = hitcounter
	if !*quiet {
		log.Printf("Visitor #%v: %s %s %s %s", c.Visitors, r.UserAgent(), r.RemoteAddr, r.Host, r.RequestURI)
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

	// Send an error message without using session
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
	fortune := newfortune()
	t, templateerr := template.New("Index").Funcs(funcMap).ParseFiles(c.templatesDir + "index.html")
	if templateerr != nil {
		// Something happened to the template since booting successfully. Must be 100% correct HTML syntax.
		log.Println("Almost fatal")
		log.Println(templateerr)
		fmt.Fprintf(w, "We are experiencing some technical difficulties. Please come back soon!")
	} else {
		data := map[string]interface{}{
			"Now":            nowtime,             // Now
			"Status":         status,              // notify of form success or fail
			"Version":        version,             // Cosgo version
			"Hits":           hitcounter,          // Visitor hits
			"Uptime":         uptime,              // Uptime
			"Boottime":       timeboot,            // Boot time
			"Fortune":        fortune,             // random fortune from fortunes.txt
			"Title":          c.Name,              // Site Name from config
			"PublicKey":      string(c.publicKey), // GPG key
			"Key":            c.getKey(),          // POST URL key
			csrf.TemplateTag: csrf.TemplateField(r),
			"CaptchaId":      captcha.NewLen(CaptchaLength + rand.Intn(CaptchaVariation)),
		}
		t.ExecuteTemplate(w, "Index", data)

	}
}

// redirecthomeHandler redirects everyone home ("/") with a 301 redirect.
func redirecthomeHandler(rw http.ResponseWriter, r *http.Request) {
	domain := getDomain(r)
	p := bluemonday.UGCPolicy()
	cleanURL := p.Sanitize(r.URL.Path)
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

// Very primitive way of flashing a message without session info
func srvError(r *http.Request, rw http.ResponseWriter, e int) {
	http.Redirect(rw, r, "/?status=0&r="+strconv.Itoa(e)+"#contact", http.StatusFound)
	return
}

// emailHandler checks the Captcha string, and the POST key, and sends on its way.
func (c *Cosgo) emailHandler(rw http.ResponseWriter, r *http.Request) {
	if !verifyMethod(r) {
		log.Println("bad method")
		srvError(r, rw, 1)
		return
	}

	if !c.verifyKey(r) {
		srvError(r, rw, 2)
		return
	}

	if !verifyCaptcha(r) {
		srvError(r, rw, 3)
		return
	}

	if *debug {
		log.Printf("Human Visitor: %s at %s %q", r.UserAgent(), r.RemoteAddr, r.URL.Path)
	}

	r.ParseForm()

	// normalize and validate email, message
	mailform := mbox.ParseFormGPG(c.Destination, r.Form, c.publicKey)

	// quick offline email address validation
	if mailform.Email == "@" || mailform.Email == " " || !strings.ContainsAny(mailform.Email, "@") || !strings.ContainsAny(mailform.Email, ".") {
		srvError(r, rw, 4)
		return
	}

	// Sendgrid?
	if *sendgridKey != "" {
		str, ok := c.sendgridder(mailform)
		if str != "" {
			log.Println(str)
		}
		if !ok {
			log.Printf("FAILURE-contact: %q at %s\n\t%q", r.UserAgent(), r.RemoteAddr, r.Form)
			srvError(r, rw, 5)
			return
		}
		log.Printf("SUCCESS-contact: %s at %s", r.UserAgent(), r.RemoteAddr)
		srvSuccess(r, rw, 1)
		inboxcount++
		return
	}

	// save the message
	saving := mbox.Save(mailform)
	if saving != nil {
		log.Println(saving)
		srvError(r, rw, 6)
		return
	}
	log.Printf("SUCCESS-contact: %s at %s", r.UserAgent(), r.RemoteAddr)
	inboxcount++
	srvSuccess(r, rw, 1)
}

func srvSuccess(r *http.Request, rw http.ResponseWriter, status int) {
	http.Redirect(rw, r, "/?status=1", http.StatusFound)
}

// serverSingle just shows one file.
func serveSingle(pattern string, filename string) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Serving %s: %s at %s", pattern, r.UserAgent(), r.RemoteAddr)
		http.ServeContent(w, r, filename, time.Now(), nil)
	})
}
func csrfErrorHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Please clear your cache, or delete any old cookies. We have updated our CSRF token."))
}
