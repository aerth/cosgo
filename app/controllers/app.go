package controllers

import "github.com/revel/revel"

type App struct {
	*revel.Controller
}

func (c App) Index() revel.Result {
	return c.Render()
}

func (c App) Contact(name string, email string, message string) revel.Result {
	
	var id string = c.Params.Get("id")

	if len(id) > 0 {
		// Some email was provided
	}else{
		var token string = c.Params.Get("token")
		if len(token) > 0{
			// A token was provided. Attempt to convert the token into a user email
			// If the email exists, continue. Else, return an error
			c.Response.Status = 418
			return c.Render()
		}
	}
	c.Response.Status = 418
	return c.Render()
	//send the email
	//var email_status_code int = pass_along_form(destination_email, message_email, message_subject, message_body)
	//return email_status_code
}
