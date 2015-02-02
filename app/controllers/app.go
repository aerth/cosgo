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

func (c App) Success() revel.Result {
	c.Response.Status = 200
	return c.Render()
}

func (c App) Failure() revel.Result {
	c.Response.Status = 400
	return c.Render()
}