// cosgo is an easy to use contact form *server*.
package main

/*
                           _
  ___ ___  ___  __ _  ___ | |
 / __/ _ \/ __|/ _` |/ _ \| |
| (_| (_) \__ \ (_| | (_) |_|
 \___\___/|___/\__, |\___/(_)
               |___/

https://github.com/aerth/cosgo

*/

// The MIT License (MIT)
//
// Copyright (c) 2016 aerth
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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

var (
	version    = "0.9.3" // Use Makefile for precise version
	err        error
	hitcounter int
	inboxcount int
	timeboot   = time.Now()
	cwd        = os.Getenv("PWD")
	// flags
	port           = flag.String("port", "8080", "Server: `port` to listen on\n\t")
	bind           = flag.String("bind", "0.0.0.0", "Server: `interface` to bind to\n\t")
	debug          = flag.Bool("debug", false, "Logging: More verbose.\n")
	sitename       = flag.String("title", "Contact Form", "Config: Site name. Available as a template variable.\n")
	configcreate   = flag.Bool("new", false, "Config: Create.\n")
	configlocation = flag.String("config", ".cosgorc", "Config: Location.\n")
	quiet          = flag.Bool("quiet", false, "Logging: Less output. See -nolog\n")
	nolog          = flag.Bool("nolog", false, "Logging: Logs to /dev/null")
	fastcgi        = flag.Bool("fastcgi", false, "Use fastcgi (for with nginx etc)")
	logfile        = flag.String("log", "", "Logging: Use a log `file` instead of stdout\n\tExample: cosgo -log cosgo.log -debug\n")
	gpg            = flag.String("gpg", "", "GPG: Path to ascii-armored `public-key` to encrypt mbox\n)")
	sendgridKey    = flag.String("sg", "", "Sendgrid: Sendgrid API `key` (disables mbox)\n")
	dest           = flag.String("to", "", "Email: Your email `address` (-sg flag required)\n")
	mboxfile       = flag.String("mbox", "cosgo.mbox", "Email: Custom mbox file `name`\n\tExample: cosgo -mbox custom.mbox\n\t")
)

func setup() *Cosgo {
	if !flag.Parsed() {
		flag.Parse()
	}
	cosgo := new(Cosgo)
	cosgo.Refresh = (time.Minute * 42)
	cosgo.boot = time.Now()
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
		if config.Fields["bind"] != nil {
			*bind = config.Fields["bind"].(string)
		}
		if config.Fields["name"] != nil {
			*sitename = config.Fields["name"].(string)
		}
		if config.Fields["port"] != nil {
			*port = config.Fields["port"].(string)
		}
		if config.Fields["gpg"] != nil {
			*gpg = config.Fields["gpg"].(string)
		}
		if config.Fields["cookie-key"] != nil {
			cosgo.antiCSRFkey = []byte(config.Fields["cookie-key"].(string))
		}

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
	}

	// Version information
	if !*quiet {
		fmt.Println(logo)
		fmt.Printf("\n\tcosgo v" + version + "\n")
		fmt.Printf("\tCopyright 2016 aerth\n\tSource code at https://github.com/aerth/cosgo\n")
	}
	cosgo.initialize()

	log.Println("booted:", timeboot)
	log.Println("working dir:", cwd)
	log.Println("css/js/img dir:", cosgo.staticDir)
	log.Println("cosgo templates dir:", cosgo.templatesDir)
	log.Printf("binding to: %s:%s", *bind, *port)

	// Fire up the cosgo engine
	go func() {
		for {
			if *debug && !*quiet {
				log.Println("Info: Generating Random 40 URL Key...")
			}
			// set a random URL key (40 char length).
			// Future: Maybe use cookie to store the key, so each visitor gets a unique key.
			cosgo.URLKey = generateURLKey(40)
			if *debug && !*quiet {
				log.Printf("Info: URL Key is " + cosgo.URLKey + "\n")
			}
			// every X minutes change the URL key (default 42 minutes)
			time.Sleep(cosgo.Refresh)
		}
	}()

	// Disable mbox completely if using sendgrid
	if *sendgridKey != "" {
		*mboxfile = os.DevNull
	}

	// Open mbox or dev/null
	f, ferr := os.OpenFile(*mboxfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if ferr != nil {
		log.Printf("error opening file: %v", ferr)
		log.Fatal("Hint: touch ./cosgo.mbox, or chown/chmod it so that the cosgo process can access it.")
		os.Exit(1)
	}

	// Set destination email for mbox and sendgrid
	if *dest != "" {
		cosgo.Destination = *dest
	}
	mbox.Destination = cosgo.Destination
	mbox.Mail = log.New(f, "", 0)

	// Is nolog enabled?
	if *nolog {
		*logfile = os.DevNull
	}
	// Switch to cosgo.log if not debug
	if *logfile != "" {
		openLogFile()
	}

	go fortuneInit() // Spin fortunes

	return cosgo
}

func main() {
	cosgo := setup()
	r := cosgo.route(cwd)
	// Try to bind
	listener, binderr := net.Listen("tcp", *bind+":"+*port)
	if binderr != nil {
		log.Println(binderr)
		os.Exit(1)
	}

	// final check
	if cosgo.antiCSRFkey == nil {
		log.Println("Generating antiCSRFkey based on \"today\"")
		/*
			CSRF key although important, doesn't need to change often.
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
		// 64 bit
		sixtyfour := []byte(today[0:64])
		log.Println("CSRF:", string(sixtyfour))
		// Example:  CSRF: 1476230400/tmp/tstcosgo/tmp/tstcosgo1476230400147623040014762304
		// Or: 			 CSRF: 1476230400hub.com/aerth/cosgo/bi14762304001476230400147623040014
		// In /home: CSRF: 1476230400/home/ftp/home/ftp147623040014762304001476230400147623
		// All are 64 in length, reasonably secure, and unique. And none will change before tomorrow.
		cosgo.antiCSRFkey = sixtyfour
	}

	// Start Serving Loop
	for {

		// Here we either use fastcgi or normal http
		if !*fastcgi {
			go func() {
				if listener != nil {
					go http.Serve(listener,
						csrf.Protect(cosgo.antiCSRFkey,
							csrf.HttpOnly(true),
							csrf.FieldName("cosgo"),
							csrf.CookieName("cosgo"),
							csrf.Secure(false), csrf.ErrorHandler(http.HandlerFunc(csrfErrorHandler)))(r))
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
							csrf.FieldName("cosgo"),
							csrf.CookieName("cosgo"),
							csrf.Secure(false), csrf.ErrorHandler(http.HandlerFunc(csrfErrorHandler)))(r))
				} else {
					log.Fatalln("nil listener")
				}
			}()
		}

		// End the for-loop with a timer/hit counter every 30 minutes.
		select {
		default:

			if !*quiet {
				log.Printf("Uptime: %s (%s)", time.Since(timeboot), humanize(time.Since(timeboot)))
				log.Printf("Hits: %v", hitcounter)
				log.Printf("Messages: %v", inboxcount)
			}
			time.Sleep(time.Minute * 30)
		}

	}
}

var logo = `
                           _
  ___ ___  ___  __ _  ___ | |
 / __/ _ \/ __|/ _  |/ _ \| |
| (_| (_) \__ \ (_| | (_) |_|
 \___\___/|___/\__, |\___/(_)
               |___/
`

func die(signal os.Signal) {
	fmt.Println()
	log.Printf("Got %s signal, Goodbye!", signal)
	log.Println("Boot time:", humanize(time.Since(timeboot)))
	log.Println("Messages saved:", inboxcount)
	log.Println("Hits:", hitcounter)
}
func init() {

	// Key Generator
	rand.Seed(time.Now().UnixNano())

	// Hopefully a clean exit if we get a sig
	interrupt := make(chan os.Signal, 1)
	stop := make(chan os.Signal, 1)
	reload := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	signal.Notify(stop, syscall.SIGINT)
	signal.Notify(reload, syscall.SIGHUP)

	go func() {
		select {
		case signal := <-interrupt:
			die(signal)
			os.Exit(0)
		case signal := <-reload:
			die(signal)
			os.Exit(0)
		case signal := <-stop:
			die(signal)
			os.Exit(0)
		}
	}()
	flag.Parse()

}
