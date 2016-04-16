package main

import (
	"errors"
	"log"
	"net/http"
	"net/url"

	"github.com/goware/emailx"
	sendgrid "github.com/sendgrid/sendgrid-go"
)

// Form is our email
type Form struct {
	Name, Email, Subject, Message string
}

// sendgridSend connects to the Sendgrid API and processes the form.
func sendgridSend(destinationEmail string, form *Form) (ok bool, msg string) {
	//log.Println("Key: " + sendgridKey) // debug sendgrid
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
		if err == emailx.ErrInvalidFormat {
			return errors.New("Bad email format.")
		}
		if err == emailx.ErrUnresolvableHost {
			return errors.New("Bad email provider.")
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
		log.Printf("SUCCESS-contact: %s at %s", r.UserAgent(), r.RemoteAddr)
		log.Printf(msg+" %s at %s", r.UserAgent(), r.RemoteAddr)
		return nil
	} else {
		log.Printf(msg+" %s at %s", r.UserAgent(), r.RemoteAddr)
		return errors.New(msg)
	}
}
