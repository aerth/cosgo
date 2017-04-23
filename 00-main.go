/*
The MIT License (MIT)

Copyright (c) 2016 aerth

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

// cosgo is an easy to use contact form server for any purpose
package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/http/fcgi"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/aerth/mbox"
	"github.com/aerth/seconf"

	"github.com/gorilla/csrf"
)

var logo = `
                           _
  ___ ___  ___  __ _  ___ | |
 / __/ _ \/ __|/ _  |/ _ \| |
| (_| (_) \__ \ (_| | (_) |_|
 \___\___/|___/\__, |\___/(_)
               |___/
`

/*

https://github.com/aerth/cosgo

*/

var (
	version    = "0.9.3" // Use Makefile for precise version
	err        error
	hitcounter int
	inboxcount int
	timeboot   = time.Now()
	cwd        = os.Getenv("PWD")
	// flags (all 20 of them)
	portnum        = flag.Int("port", 8080, "Server: `port` to listen on\n\t")
	bind           = flag.String("bind", "0.0.0.0", "Server: `interface` to bind to\n\t")
	sslport        = flag.String("tls", "0.0.0.0:443", "Server: `interface:port` to bind (SSL/TLS)o\n\t")
	debug          = flag.Bool("debug", false, "Logging: More verbose.\n")
	sitename       = flag.String("title", "Contact Form", "Config: Site name. Available as a template variable.\n")
	configcreate   = flag.Bool("new", false, "Config: Create.\n")
	configlocation = flag.String("config", ".cosgorc", "Config: Location.\n")
	quiet          = flag.Bool("quiet", false, "Logging: Less output. See -nolog\n")
	nolog          = flag.Bool("nolog", false, "Logging: Logs to /dev/null")
	fastcgi        = flag.Bool("fastcgi", false, "Use fastcgi (for with nginx etc)")
	secure         = flag.Bool("secure", false, "HTTPS only.")
	logfile        = flag.String("log", "stdout", "Logging: Use a log `file` instead of stdout\n\tExample: cosgo -log cosgo.log -debug\n")
	cookie         = flag.String("cookie", "cosgo", "Custom cookie+form field name\n")
	gpg            = flag.String("gpg", "", "GPG: Path to ascii-armored `public-key` to encrypt mbox\n)")
	sendgridKey    = flag.String("sg", "", "Sendgrid: Sendgrid API `key` (disables mbox)\n")
	dest           = flag.String("to", "", "Email: Your email `address` (-sg flag required)\n")
	mboxfile       = flag.String("mbox", "cosgo.mbox", "Email: Custom mbox file `name`\n\tExample: cosgo -mbox custom.mbox\n\t")
	refreshTime    = flag.Duration("refresh", time.Hour, "How often to change the POST URL Key")
	path2key       = flag.String("key", "", "Path to SSL Key")
	path2cert      = flag.String("cert", "", "Path to SSL Cert")
)

func setup() *Cosgo {
	log.SetPrefix("cosgo>")

	cosgo := new(Cosgo)
	cosgo.r = nil

	cosgo.boot = time.Now()
	cosgo.rw.Lock()
	t1 := time.Now()
	kee := generateURLKey(40)
	cosgo.URLKey = kee
	if *debug && !*quiet {
		log.Printf("Generated URL Key %q in %v", cosgo.URLKey, time.Now().Sub(t1))
	}
	cosgo.rw.Unlock()

	// Seconf encoded configuration file (recommended)
	if seconf.Exists(*configlocation) {
		config, errar := seconf.ReadJSON(*configlocation)
		if errar != nil {
			log.Println(errar)
			log.Fatalf("Bad config. Please remove the %q file and try again.", *configlocation)
			os.Exit(1)
		}

		if config.Fields["bind"] == nil || config.Fields["gpg"] == nil || config.Fields["port"] == nil || config.Fields["name"] == nil || config.Fields["cookie-key"] == nil {
			log.Fatalf("Bad config. Please remove the %q file and try again.", *configlocation)
		}

		cosgo.Bind = config.Fields["bind"].(string)
		*sitename = config.Fields["name"].(string)
		cosgo.Port = config.Fields["port"].(string)
		*gpg = config.Fields["gpg"].(string)
		cosgo.antiCSRFkey = []byte(config.Fields["cookie-key"].(string))

	} else if *configcreate { // Config does not exist, user is asking to make one.
		fmt.Println("Welcome to Cosgo!")
		seconf.LockJSON(*configlocation, "", map[string]string{
			"bind":        "Bind to what address? \n0.0.0.0 for all, 127.0.0.1 for local/tor only",
			"port":        "Which port to listen on?",
			"cookie-code": "32/64 bit cookie key. Good to make this random. Important not to change across reboots.",
			"name":        "What is your site called? Available as a template variable",
			"gpg":         "Location of GPG public key. Enter \"none\" to disable",
		})
		os.Exit(0)
	} else {
		// No configuration detected
	}

	// Version information
	if !*quiet {
		fmt.Println(logo)
		fmt.Printf("\n\tcosgo v" + version + "\n")
		fmt.Printf("\tCopyright 2016 aerth\n\tSource code at https://github.com/aerth/cosgo\n")
		fmt.Println(os.Args)
	}
	cosgo.initialize()
	if !*quiet {
		log.Println("booted:", timeboot)
		log.Println("working dir:", cwd)
		log.Println("css/js/img dir:", cosgo.staticDir)
		log.Println("cosgo templates dir:", cosgo.templatesDir)
	}

	// Disable mbox completely if using sendgrid
	if *sendgridKey != "" {
		*mboxfile = os.DevNull
	}
	mbox.Destination = cosgo.Destination
	mbox.Open(*mboxfile)

	// Set destination email for mbox and sendgrid
	if *dest != "" {
		cosgo.Destination = *dest
	}

	go fortuneInit() // Spin fortunes

	return cosgo
}

func main() {

	// Create the server, load mbox and fortunes and run initialize
	cosgo := setup()

	// Set all the needed /url paths
	e := cosgo.route(cwd)
	if e != nil {
		log.Fatalln(e)
	}

	// Needs to be compiled with build tag 'debug' to be redefined, and -debug CLI flag to be activated
	if *debug {
		cosgo.debug()
	}
	cosgo.Bind = *bind
	cosgo.Port = strconv.Itoa(*portnum)
	log.Println("Refreshing every", *refreshTime)
	go func() {
		time.Sleep(100 * time.Millisecond)
		log.Println("Listening on", cosgo.Bind+":"+cosgo.Port)
	}()
	// Try to bind
	listener, binderr := net.Listen("tcp", cosgo.Bind+":"+cosgo.Port)
	if binderr != nil {
		log.Println(binderr)
		os.Exit(1)
	}

	if cosgo.antiCSRFkey == nil {
		cosgo.antiCSRFkey = anticsrfGen()
	}
	if *path2cert != *path2key {
		go cosgo.ServeSSL()
	}

	// Is nolog enabled?
	if *nolog {
		*logfile = os.DevNull
	}
	// stdout or a filename
	openLogFile()

	// Start Serving
	// Here we either use fastcgi or normal http server, using csrf and mux.
	// with custom csrf error handler and 10 minute cookie.
	if !*fastcgi {

		go func() {
			if listener != nil {
				go http.Serve(listener,
					csrf.Protect(cosgo.antiCSRFkey,
						csrf.HttpOnly(true),
						csrf.FieldName(*cookie),
						csrf.CookieName(*cookie),
						csrf.Secure(*secure), csrf.MaxAge(600), csrf.ErrorHandler(http.HandlerFunc(csrfErrorHandler)))(cosgo.r))
			} else {
				log.Fatalln("nil listener")
			}

		}()
	} else {
		go func() {
			if listener != nil {
				go fcgi.Serve(listener,
					csrf.Protect(cosgo.antiCSRFkey,
						csrf.HttpOnly(true),
						csrf.FieldName(*cookie),
						csrf.CookieName(*cookie),
						csrf.Secure(*secure), csrf.MaxAge(600), csrf.ErrorHandler(http.HandlerFunc(csrfErrorHandler)))(cosgo.r))
			} else {
				log.Fatalln("nil listener")
			}
		}()
	}
	for {
		select {
		// Fire up the cosgo engine

		case <-time.After(*refreshTime):
			cosgo.rw.Lock()
			if *debug && !*quiet {
				log.Println("Info: Generating Random 40 URL Key...")
			}
			t1 := time.Now()
			// set a random URL key (40 char length).
			kee := generateURLKey(40)
			cosgo.URLKey = kee
			if *debug && !*quiet {
				log.Printf("Generated URL Key %q in %v", cosgo.URLKey, time.Now().Sub(t1))
			}
			cosgo.rw.Unlock()

			// every X minutes change the URL key (default 42 minutes)
			// break tests uncomment next line
			//*refreshTime = time.Nanosecond

			if !*quiet {
				log.Printf("Uptime: %s (%s)", time.Since(timeboot), humanize(time.Since(timeboot)))
				log.Printf("Hits: %v", hitcounter)
				log.Printf("Messages: %v", inboxcount)
				if *debug {
					log.Printf("Port: %v", cosgo.Port)
				}
				if *path2cert != "" {
					log.Println("TLS: ON")
				}
			}
			// loop
		}
	}
}
func init() {
	// Key Generator
	rand.Seed(time.Now().UnixNano())

	// Hopefully a clean exit if we get a sig
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGINT, syscall.SIGHUP, syscall.SIGKILL)
	go func() {
		select {
		case signal := <-interrupt:
			fmt.Println()
			log.Printf("Got %s signal, Goodbye!", signal)
			log.Println("Boot time:", humanize(time.Since(timeboot)))
			log.Println("Messages saved:", inboxcount)
			log.Println("Hits:", hitcounter)
			os.Exit(0)
		}
	}()
	flag.Parse()

}

func anticsrfGen() []byte {
	log.Println("Generating antiCSRFkey based on \"today\"")
	/*
		CSRF key is important, but doesn't need to change often.
		And I don't expect users to set their own.
		This method changes at most once a day, and makes sure each cosgo instance's key is unique.
	*/

	today := strconv.Itoa(int(time.Now().Truncate(24 * time.Hour).Unix()))
	here, _ := os.Getwd()

	// Add working directory
	if len(here) < 23 { // short cwd
		today += here + here
	} else { // long cwd
		today += here[len(here)-23 : len(here)-1]
	}
	// Increase length
	for {
		if len([]byte(today)) < 64 { // short cwd, add today
			today += strconv.Itoa(int(time.Now().Truncate(24 * time.Hour).Unix()))
		} else {
			break
		}
	}
	sixtyfour := []byte(today[0:64])
	log.Printf("CSRF: %q", string(sixtyfour))
	return sixtyfour
	// 64 bit
	// Example:  CSRF: 1476230400/tmp/tstcosgo/tmp/tstcosgo1476230400147623040014762304
	// Or: 			 CSRF: 1476230400hub.com/aerth/cosgo/bi14762304001476230400147623040014
	// In /home: CSRF: 1476230400/home/ftp/home/ftp147623040014762304001476230400147623
	// All are 64 in length, reasonably secure, and unique. And none will change before tomorrow.

}

// ServeSSL serves cosgo on port 443 with attached key+cert
func (c *Cosgo) ServeSSL() {
	go func() {
		time.Sleep(100 * time.Millisecond)
		log.Println("Cosgo: Serving TLS on", *sslport)
	}()

	log.Fatalln(http.ListenAndServeTLS(*sslport, *path2cert, *path2key,
		csrf.Protect(c.antiCSRFkey,
			csrf.HttpOnly(true),
			csrf.FieldName(*cookie),
			csrf.CookieName(*cookie),
			csrf.Secure(true),
			csrf.MaxAge(600),
			csrf.ErrorHandler(http.HandlerFunc(csrfErrorHandler)),
		)(c.r)))

}
