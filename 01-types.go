package main

import (
	"time"

	"github.com/gorilla/mux"
)

// Cosgo struct holds the changing URLKey and hit counter variables
type Cosgo struct {
	boot                    time.Time
	URLKey                  string // This changes every X minutes
	Name                    string
	Visitors                int    // hit counter
	publicKey               []byte // gpg key bytes (ascii armored)
	antiCSRFkey             []byte // needed for POST method
	Destination             string // you@yours.com
	staticDir, templatesDir string
	r                       *mux.Router
	routed                  bool
	Port, Bind              string
}
