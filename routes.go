package main

import (
	"net/http"

	"github.com/dchest/captcha"
	"github.com/gorilla/mux"
)

func route(cwd string, staticDir string) (r *mux.Router) {
	r = mux.NewRouter()

	//Redirect 404 errors home
	r.NotFoundHandler = http.HandlerFunc(redirecthomeHandler)

	// Home Page
	r.HandleFunc("/", homeHandler)

	// POST endpoint (emailHandler checks the key)
	r.HandleFunc("/{{whatever}}/send", emailHandler)

	s := http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir)))
	ss := http.FileServer(http.Dir(staticDir))
	sf := http.FileServer(http.Dir(cwd + "files"))

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
		r.Methods("GET").Path("/files/").Handler(sf)
		/*
			r.Methods("GET").Path("/files/{whatever}.tgz").Handler(ss)
			r.Methods("GET").Path("/files/{whatever}.txz").Handler(ss)
			r.Methods("GET").Path("/files/{whatever}.txt").Handler(ss)
			r.Methods("GET").Path("/files/{whatever}.tar").Handler(ss)
			r.Methods("GET").Path("/files/{whatever}.zip").Handler(ss)
			r.Methods("GET").Path("/files/{whatever}.tar.gz").Handler(ss)
			r.Methods("GET").Path("/files/{whatever}.tar.bz2").Handler(ss)
		*/
		if *customExtension != "" {
			r.Methods("GET").Path("/files/{whatever}." + *customExtension).Handler(sf)
		}
	}

	// Serve Captcha IMG and WAV
	r.Methods("GET").Path("/captcha/{captchacode}.png").Handler(captcha.Server(StdWidth, StdHeight))
	r.Methods("GET").Path("/captcha/download/{captchacode}.wav").Handler(captcha.Server(StdWidth, StdHeight))
	r.Methods("GET").Path("/captcha/{captchacode}.wav").Handler(captcha.Server(StdWidth, StdHeight))

	// Set routing
	http.Handle("/", r)
	return r
}
