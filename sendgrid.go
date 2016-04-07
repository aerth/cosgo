package main

import sendgrid "github.com/sendgrid/sendgrid-go"

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
