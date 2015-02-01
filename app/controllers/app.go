package controllers

import (
	"fmt"
	mandrill "github.com/keighl/mandrill"
	"github.com/revel/revel"
	"os"
)

type App struct {
	*revel.Controller
}

type Form struct {
	Name, Email, Subject, Message string
}

func (c App) Index() revel.Result {
	return c.Render()
}

func (c App) Contact() revel.Result {
	email := c.Params.Get("id")
	form := Form{Name: c.Params.Get("name"),
		Email:   c.Params.Get("email"),
		Subject: c.Params.Get("subject"),
		Message: c.Params.Get("message")}
	//invalid_id = validate_destination_email(email)
	//if invalid_id {
	//	return c.Redirect(App.InvalidEmail)
	//}
	//invalid_form = validate_form(form)
	//if valid_form {
	//
	//}
	successfully_sent := send_email(email, form)
	if successfully_sent {
		return c.Redirect(App.Success)
	} else {
		return c.Redirect(App.Failure)
	}
	return c.Render(email, form, successfully_sent)
}

func (c App) Success() revel.Result {
	c.Response.Status = 200
	return c.Render()
}

func (c App) Failure() revel.Result {
	c.Response.Status = 400
	return c.Render()
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
	fmt.Printf("%s %s %s\n", responses, apiError, err)
	return true
}
