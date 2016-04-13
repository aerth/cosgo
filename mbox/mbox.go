// Package mbox saves a form to a local .mbox file
package mbox

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/goware/emailx"
	"github.com/microcosm-cc/bluemonday"
)

// Form is our email. No Attachments
type Form struct {
	Name, Email, Subject, Message string
}

// Level should be set to something other than 1 to resolve and check email addresses
var Level = 1
var Destination = "mbox@localhost"
var (
	Mail *log.Logger // local mbox
)

// Save saves an mbox file from a submitted query! Epic!
func Save(rw http.ResponseWriter, r *http.Request, destination string, query url.Values) error {
	form := parseQuery(query)
	t := time.Now()
	if form.Email == "@" || form.Email == " " || !strings.ContainsAny(form.Email, "@") || !strings.ContainsAny(form.Email, ".") {
		return errors.New("Bad email address.")
	}
	if Level != 1 {
		err := emailx.Validate(form.Email)
		if err != nil {
			if err == emailx.ErrInvalidFormat {
				fmt.Fprintln(rw, "<html><p>Email is not valid format.</p></html>")
				return errors.New("Email is not valid format.")
			}

			if err == emailx.ErrUnresolvableHost {
				fmt.Fprintln(rw, "<html><p>We don't recognize that email provider.</p></html>")
				return errors.New("Email is not valid format.")
			}

			fmt.Fprintln(rw, "<html><p>Email is not valid. Would you like to go <a href=\"/\">back</a>?</p></html>")
			return errors.New("Email is not valid format." + err.Error())

		}
	}
	//Normalize email address
	form.Email = emailx.Normalize(form.Email)
	// mbox files use two different date formats apparently.
	mailtime := t.Format("Mon Jan 2 15:04:05 2006")
	mailtime2 := t.Format("Mon, 2 Jan 2006 15:04:05 -0700")
	Mail.Println("From " + form.Email + " " + mailtime)
	Mail.Println("Return-path: <" + form.Email + ">")
	Mail.Println("Envelope-to: " + destination)
	Mail.Println("Delivery-date: " + mailtime2)
	Mail.Println("To: " + destination)
	Mail.Println("Subject: " + form.Subject)
	Mail.Println("From: " + form.Email)
	Mail.Println("Date: " + mailtime2)
	Mail.Println("\n" + form.Message + "\n\n")
	return nil
}

func parseQuery(query url.Values) *Form {
	p := bluemonday.StrictPolicy()
	form := new(Form)
	additionalFields := ""
	for k, v := range query {
		k = strings.ToLower(k)
		if k == "email" || k == "name" {
			form.Email = v[0]
			form.Email = p.Sanitize(form.Email)
		} else if k == "subject" {
			form.Subject = v[0]
			form.Subject = p.Sanitize(form.Subject)
		} else if k == "message" {
			form.Message = k + ": " + v[0] + "<br>\n"
			form.Message = p.Sanitize(form.Message)
		} else if k != "cosgo" && k != "captchaid" && k != "captchasolution" {
			additionalFields = additionalFields + k + ": " + v[0] + "<br>\n"
		}
	}
	if form.Subject == "" || form.Subject == " " {
		form.Subject = "[New Message]"
	}
	if additionalFields != "" {
		if form.Message == "" {
			form.Message = form.Message + "Message:\n<br>" + additionalFields
			//form.Message = form.Message
		} else {
			form.Message = form.Message + "\n<br>Additional:\n<br>" + additionalFields
			//form.Message = form.Message
		}
	}
	return form
}
