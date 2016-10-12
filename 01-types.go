package main

import "time"

// Cosgo struct holds the changing URLKey and hit counter variables
type Cosgo struct {
	boot                    time.Time
	URLKey                  string // This changes every X minutes
	Refresh                 time.Duration
	Name                    string
	Visitors                int    // hit counter
	publicKey               []byte // gpg key bytes (ascii armored)
	antiCSRFkey             []byte // needed for POST method
	Destination             string // you@yours.com
	staticDir, templatesDir string
}
