package main

import (
	"github.com/aerth/mbox"

	sendgrid "gopkg.in/sendgrid/sendgrid-go.v2" //sendgrid "github.com/sendgrid/sendgrid-go"
)

// Form is sent by a user
type Form struct {
	Name, Email, Subject, Message string
}

// sendgridder connects to the Sendgrid API and sends the form as an email to cosgo.Destination.
func (c *Cosgo) sendgridder(form *mbox.Form) (msg string, ok bool) {
	sg := sendgrid.NewSendGridClientWithApiKey(*sendgridKey)
	message := sendgrid.NewMail()
	message.AddTo(c.Destination)
	message.SetFrom(form.Email)
	message.SetFromName(form.Name)
	message.SetSubject(form.Subject)
	message.SetText(form.Message)
	r := sg.Send(message)
	if r == nil {
		return string("Sendgrid: Email sent to " + c.Destination), true
	}
	return string("Sendgrid Error:" + r.Error()), false
}

//
// // sendgridSender format for sendgrid and send it
// func (c *Cosgo) sendgridSender(rw http.ResponseWriter, r *http.Request, destination string, query url.Values) error {
// 	form := mbox.ParseQuery(query)
// 	//Validate user submitted email address
// 	err := emailx.Validate(form.Email)
// 	if err != nil {
// 		if err == emailx.ErrInvalidFormat {
// 			return errors.New("Bad email format.")
// 		}
// 		if err == emailx.ErrUnresolvableHost {
// 			return errors.New("Bad email provider.")
// 		}
// 		return errors.New("Bad email address.")
// 	}
// 	//Normalize email address
// 	form.Email = emailx.Normalize(form.Email)
// 	// Is it empty?
// 	if form.Email == "" || form.Email == "@" {
// 		return errors.New("Bad email address.")
// 	}
//
// 	// Looks good! Lets send it to sendgrid!
// 	msg, ok := c.sendgridder(form)
//
// 	// Is good send!
// 	if ok == true {
// 		log.Printf("SUCCESS-contact: %s at %s", r.UserAgent(), r.RemoteAddr)
// 		log.Printf(msg+" %s at %s", r.UserAgent(), r.RemoteAddr)
// 		return nil
// 	}
//
// 	// The send wasn't good. Sending msg to log just in case it was important.
// 	log.Printf("Bad Message: %q from a %s at %s", msg, r.UserAgent(), r.RemoteAddr)
// 	return errors.New(msg)
// }
