package controllers

import (
	"os"
	"fmt"
	mandrill "github.com/keighl/mandrill"
	"github.com/revel/revel"
)

type Contact struct {
	*revel.Controller
}

type Form struct {
	Name, Email, Subject, Message string
}

func (c Contact) Contact() revel.Result {
	email := c.Params.Get("id")
	form := Form{Name: c.Params.Get("name"),
		Email:   c.Params.Get("email"),
		Subject: c.Params.Get("subject"),
		Message: c.Params.Get("message")}
	successfully_sent := send_email(email, form)
	if successfully_sent {
		return c.Redirect(App.Success)
	} else {
		return c.Redirect(App.Failure)
	}
}

func send_email(destination string, form Form) bool {
	MANDRILL_KEY := os.Getenv("MANDRILL_KEY")
	if len(MANDRILL_KEY) == 0 {
		fmt.Printf("API Key for Mandrill was not found.\nSet the environment variable MANDRILL_KEY with a valid API key.")
		return false
	}
	client := mandrill.ClientWithKey(MANDRILL_KEY)

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
		return false
	}
	if apiError != nil {
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
