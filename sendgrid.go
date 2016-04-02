package main

import (
	"fmt"

	sendgrid "github.com/sendgrid/sendgrid-go"
)

type Form struct {
	Name, Email, Subject, Message string
}

func canSendgrid() bool {
	return true
}

func sendgridSend(destinationEmail string, form *Form) bool {
	//if !canMandrill() {

	//	return false
	//}

	sg := sendgrid.NewSendGridClientWithApiKey(sendgridKey)
	message := sendgrid.NewMail()
	message.AddTo(destinationEmail)
	message.SetFrom(form.Email)
	message.SetFromName(form.Name)
	message.SetSubject(form.Subject)
	message.SetText(form.Message)
	if r := sg.Send(message); r == nil {
		fmt.Println("Sendgrid Email sent!")
		return true
	} else {
		fmt.Println(r)
		return false
	}

}
