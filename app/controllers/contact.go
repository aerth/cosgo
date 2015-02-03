package controllers

import (
	mandrill "github.com/keighl/mandrill"
	"github.com/revel/revel"
	"fmt"
)

type Contact struct {
	*revel.Controller
}

type Form struct {
	Name, Email, Subject, Message string
}

type recipientRecord struct {
	Token, Email, MonthlyCount, CurrentMonth string
}

func (c Contact) Contact() revel.Result {
	email := c.Params.Get("id")
	form := Form{Name: c.Params.Get("name"),
		Email:   c.Params.Get("email"),
		Subject: c.Params.Get("subject"),
		Message: c.Params.Get("message")}
	sent := sendEmail(email, form)
	if sent {
		return c.Redirect(App.Success)
	} else {
		return c.Redirect(App.Failure)
	}
}

func sendEmail(destination string, form Form) bool {
	
	mandrillKey, found := revel.Config.String("mandrillKey")
	if !found {
		panic("Mandrill API key not set in app.conf.")
	}
	if len(mandrillKey) == 0 {
		panic("Mandrill API key is empty.")
	}
	
	client := mandrill.ClientWithKey(mandrillKey)

	message := &mandrill.Message{}
	message.AddRecipient(destination, destination, "to")
	message.FromEmail = form.Email
	message.FromName = form.Name
	if len(form.Subject) == 0{
		form.Subject = "New contact form submission!"
	}
	message.Subject = form.Subject
	message.HTML = "<p>" + form.Message + "<p>"
	message.Text = form.Message

	responses, apiError, err := client.MessagesSend(message)
	
	if err != nil {
		panic(err)
	}
	if apiError != nil {
		fmt.Printf("\n\n\n\nAPI ERROR\n%s\n\n\n\n", apiError)
		return false
	}
	
	length := len(responses)
	for i := 0; i < length; i++ {
		if responses[i].Status == "rejected" {
			return false
		} else if responses[i].Status == "invalid" {
			return false
		}
	}
	return true
}
