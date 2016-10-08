package main

import (
	"net/http"
	"net/http/pprof"

	"github.com/dchest/captcha"
	"github.com/gorilla/mux"
)

func route(cwd string, staticDir string) (r *mux.Router) {
	r = mux.NewRouter()

	//Redirect 404 errors home
	r.NotFoundHandler = http.HandlerFunc(redirecthomeHandler)

	// Home Page
	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/pub.asc", pubkeyHandler)
	r.HandleFunc("/pub.txt", pubkeyHandler)
	if *debug {
		r.Handle("/debug/pprof/{{whatever}}", pprof.Handler("Index"))
		r.Handle("/hacker/{{whatever}}", pprof.Handler("Index"))

	}
	// POST endpoint (emailHandler checks the key)
	r.HandleFunc("/{{whatever}}/send", emailHandler)

	//
	s := http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir)))
	ss := http.FileServer(http.Dir(staticDir))

	if staticDir != "" {
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
