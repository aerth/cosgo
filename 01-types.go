package main

import (
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// Cosgo struct holds the changing URLKey and hit counter variables
type Cosgo struct {
	boot                    time.Time
	rw                      sync.RWMutex
	mu                      sync.Mutex // this protects URLKey
	URLKey                  string     // This changes every X minutes
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
