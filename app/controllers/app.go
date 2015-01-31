package controllers

import (
		"github.com/revel/revel"
)

type App struct {
	*revel.Controller
}

func (c App) Index() revel.Result {
	return c.Render()
}

func send_email(destination_email string, sender_name string, sender_email string, sender_subject string, sender_message string) {

}

// Show the hotel information.
func (c App) Show() revel.Result {
	var id string
	c.Params.Bind(&id, "id")
	return c.Render(id)
}

// Show the hotel information.
func (c App) Contact() revel.Result {
	var id string
	c.Params.Bind(&id, "id")
	return c.Render(id)
}