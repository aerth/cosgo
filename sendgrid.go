package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/csrf"
	"github.com/goware/emailx"
	sendgrid "github.com/sendgrid/sendgrid-go"
)

// Form is our email
type Form struct {
	Name, Email, Subject, Message string
}

// sendgridSend connects to the Sendgrid API and processes the form.
func sendgridSend(destinationEmail string, form *Form) (ok bool, msg string) {
	sg := sendgrid.NewSendGridClientWithApiKey(sendgridKey)
	message := sendgrid.NewMail()
	message.AddTo(destinationEmail)
	message.SetFrom(form.Email)
	message.SetFromName(form.Name)
	message.SetSubject(form.Subject)
	message.SetText(form.Message)
	r := sg.Send(message)
	if r == nil {
		return true, string("Sendgrid Email sent from" + destinationEmail)
	}
	return false, string("Sendgrid Error: (" + destinationEmail + ")" + r.Error())
}

// sendgridSender always returns success for the visitor. This function needs some work.
func sendgridSender(rw http.ResponseWriter, r *http.Request, destination string, query url.Values) error {
	form := parseQuery(query)
	//Validate user submitted email address
	err := emailx.Validate(form.Email)
	if err != nil {
		fmt.Fprintln(rw, "<html><p>Email is not valid. Would you like to go <a href=\"/\">back</a>?</p></html>")

		if err == emailx.ErrInvalidFormat {
			fmt.Fprintln(rw, "<html><p>Email is not valid format.</p></html>")
		}
		if err == emailx.ErrUnresolvableHost {
			fmt.Fprintln(rw, "<html><p>We don't recognize that email provider.</p></html>")
		}
		return errors.New("Bad email address.")
	}
	//Normalize email address
	form.Email = emailx.Normalize(form.Email)
	//Is it empty?
	if form.Email == "" || form.Email == "@" {
		return errors.New("Bad email address.")
	}

	if ok, msg := sendgridSend(destination, form); ok == true {
		fmt.Fprintln(rw, "<html><p>Thanks! Would you like to go <a href=\"/\">back</a>?</p></html>")
		log.Printf("SUCCESS-contact: %s at %s", r.UserAgent(), r.RemoteAddr)
		log.Printf(msg+" %s at %s", r.UserAgent(), r.RemoteAddr)
		return nil
	} else {
		log.Printf(msg+" %s at %s", r.UserAgent(), r.RemoteAddr)
		// Basic error template
		t, err := template.New("Error").ParseFiles("./templates/error.html")
		if err == nil {
			data := map[string]interface{}{
				"err":            "Mail System",
				"Key":            getKey(),
				csrf.TemplateTag: csrf.TemplateField(r),
			}
			t.ExecuteTemplate(rw, "Error", data)
			return err
		} else {
			log.Printf("template error: %s at %s", r.UserAgent(), r.RemoteAddr)
			log.Println(err)
			http.Redirect(rw, r, "/", 301)
			return errors.New("Bad error.html template.")
		}
	}
}
