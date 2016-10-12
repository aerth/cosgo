package main

import (
	"errors"
	"net/http"

	"github.com/dchest/captcha"
	"github.com/gorilla/mux"
)

func (c *Cosgo) route(cwd string) error {
	if c.routed {
		return errors.New("already is routed")
	}
	c.r = mux.NewRouter()
	//Redirect 404 errors home
	c.r.NotFoundHandler = http.HandlerFunc(redirecthomeHandler)

	// Home Page
	c.r.HandleFunc("/", c.homeHandler)
	c.r.HandleFunc("/pub.asc", c.pubkeyHandler)
	c.r.HandleFunc("/pub.txt", c.pubkeyHandler)

	// POST endpoint (emailHandler checks the key)
	c.r.HandleFunc("/{{key}}/send", c.emailHandler)

	//

	if c.staticDir != "" {
		s := http.StripPrefix("/static/", http.FileServer(http.Dir(c.staticDir)))
		ss := http.FileServer(http.Dir(c.staticDir))
		// Serve /static folder and favicon etc
		c.r.Methods("GET").Path("/favicon.ico").Handler(ss)
		c.r.Methods("GET").Path("/robots.txt").Handler(ss)
		c.r.Methods("GET").Path("/sitemap.xml").Handler(ss)
		c.r.Methods("GET").Path("/static/{dir}/{whatever}.css").Handler(s)
		c.r.Methods("GET").Path("/static/{dir}/{whatever}.js").Handler(s)
		c.r.Methods("GET").Path("/static/{dir}/{whatever}.png").Handler(s)
		c.r.Methods("GET").Path("/static/{dir}/{whatever}.jpg").Handler(s)
		c.r.Methods("GET").Path("/static/{dir}/{whatever}.jpeg").Handler(s)
		c.r.Methods("GET").Path("/static/{dir}/{whatever}.woff").Handler(s)
		c.r.Methods("GET").Path("/static/{dir}/{whatever}.ttf").Handler(s)
		c.r.Methods("GET").Path("/static/{dir}/{whatever}.txt").Handler(s)
		c.r.Methods("GET").Path("/static/{dir}/{whatever}.mp3").Handler(s)
		c.r.Methods("GET").Path("/static/{dir}/{whatever}.m3u").Handler(s)
		c.r.Methods("GET").Path("/static/{dir}/{whatever}.md").Handler(s)

	}

	// Serve Captcha IMG and WAV
	c.r.Methods("GET").Path("/captcha/{captchacode}.png").Handler(captcha.Server(StdWidth, StdHeight))
	c.r.Methods("GET").Path("/captcha/download/{captchacode}.wav").Handler(captcha.Server(StdWidth, StdHeight))
	c.r.Methods("GET").Path("/captcha/{captchacode}.wav").Handler(captcha.Server(StdWidth, StdHeight))

	//
	c.routed = true
	return nil
}
