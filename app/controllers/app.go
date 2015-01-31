package controllers

import (
		"os"
		"fmt"
		"github.com/revel/revel"
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
	id := c.Params.Get("id")
	form := Form{Name: c.Params.Get("name"), 
				 Email: c.Params.Get("email"),
				 Subject: c.Params.Get("subject"),
				 Message: c.Params.Get("message")}
    successfully_sent := send_email(form)
    if successfully_sent	{
    	return c.Redirect(App.Success)
    }	else	{
    	return c.Redirect(App.Failure)
    }
	return c.Render(id, form, successfully_sent)
}

func (c App) Success() revel.Result {
	c.Response.Status = 200
	return c.Render()
}

func (c App) Failure() revel.Result {
	c.Response.Status = 400
	return c.Render()
}

func send_email(form Form) bool {
	os_val := os.Getenv("")
	fmt.Printf("%s\n", os_val)
	return true
}