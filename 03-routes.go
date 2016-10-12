package main

import (
	"net/http"

	"github.com/dchest/captcha"
	"github.com/gorilla/mux"
)

func (c *Cosgo) route(cwd string) (r *mux.Router) {
	r = mux.NewRouter()

	//Redirect 404 errors home
	r.NotFoundHandler = http.HandlerFunc(redirecthomeHandler)

	// Home Page
	r.HandleFunc("/", c.homeHandler)
	r.HandleFunc("/pub.asc", c.pubkeyHandler)
	r.HandleFunc("/pub.txt", c.pubkeyHandler)
	// if *debug {
	// 	r.Handle("/debug/pprof/{{whatever}}", pprof.Handler("Index"))
	// 	r.Handle("/hacker/{{whatever}}", pprof.Handler("Index"))
	// }
	// POST endpoint (emailHandler checks the key)
	r.HandleFunc("/{{whatever}}/send", c.emailHandler)

	//

	if c.staticDir != "" {
		s := http.StripPrefix("/static/", http.FileServer(http.Dir(c.staticDir)))
		ss := http.FileServer(http.Dir(c.staticDir))
		// Serve /static folder and favicon etc
		r.Methods("GET").Path("/favicon.ico").Handler(ss)
		r.Methods("GET").Path("/robots.txt").Handler(ss)
		r.Methods("GET").Path("/sitemap.xml").Handler(ss)
		r.Methods("GET").Path("/static/{dir}/{whatever}.css").Handler(s)
		r.Methods("GET").Path("/static/{dir}/{whatever}.js").Handler(s)
		r.Methods("GET").Path("/static/{dir}/{whatever}.png").Handler(s)
		r.Methods("GET").Path("/static/{dir}/{whatever}.jpg").Handler(s)
		r.Methods("GET").Path("/static/{dir}/{whatever}.jpeg").Handler(s)
		r.Methods("GET").Path("/static/{dir}/{whatever}.woff").Handler(s)
		r.Methods("GET").Path("/static/{dir}/{whatever}.ttf").Handler(s)
		r.Methods("GET").Path("/static/{dir}/{whatever}.txt").Handler(s)
		r.Methods("GET").Path("/static/{dir}/{whatever}.mp3").Handler(s)
		r.Methods("GET").Path("/static/{dir}/{whatever}.m3u").Handler(s)
		r.Methods("GET").Path("/static/{dir}/{whatever}.md").Handler(s)

	}

	// Serve Captcha IMG and WAV
	r.Methods("GET").Path("/captcha/{captchacode}.png").Handler(captcha.Server(StdWidth, StdHeight))
	r.Methods("GET").Path("/captcha/download/{captchacode}.wav").Handler(captcha.Server(StdWidth, StdHeight))
	r.Methods("GET").Path("/captcha/{captchacode}.wav").Handler(captcha.Server(StdWidth, StdHeight))

	// Set routing
	http.Handle("/", r)
	return r
}
