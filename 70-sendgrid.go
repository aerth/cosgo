package main

import (
	"github.com/aerth/mbox"

	sendgrid "gopkg.in/sendgrid/sendgrid-go.v2" //sendgrid "github.com/sendgrid/sendgrid-go"
)

// sendgridder connects to the Sendgrid API and sends the form as an email to cosgo.Destination.
func (c *Cosgo) sendgridder(form *mbox.Form) (msg string, ok bool) {
	sg := sendgrid.NewSendGridClientWithApiKey(*sendgridKey)
	message := sendgrid.NewMail()
	message.AddTo(c.Destination)
	message.SetFrom(form.From)
	message.SetFromName(form.From)
	message.SetSubject(form.Subject)
	message.SetText(form.Message)
	r := sg.Send(message)
	if r == nil {
		return string("Sendgrid: Email sent to " + c.Destination), true
	}
	return string("Sendgrid Error:" + r.Error()), false
}
