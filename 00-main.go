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
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/aerth/mbox"

	"github.com/gorilla/csrf"
	sl "github.com/hydrogen18/stoppableListener"
)

var (
	version          = "0.9.G" // Use makefile for version hash
	destinationEmail = "cosgo@localhost"
	antiCSRFkey      = []byte("LI80PNK1xcT01jmQBsEyxyrNCrbyyFPjPU8CKnxwmCruxNijgnyb3hXXD3p1RBc0+LIRQUUbTtis6hc6LD4I/A==")
	cosgoRefresh     = 42 * time.Minute
	err              error
	hitcounter       int
	timeboot         = time.Now()
	templateDir      = "./templates/"
	staticDir        = "./static/"
	cwd              string
	publicKey        []byte
	cosgo            = new(Cosgo)
	// flags
	port        = flag.String("port", "8080", "Server: `port` to listen on\n\t")
	bind        = flag.String("bind", "0.0.0.0", "Server: `interface` to bind to\n\t")
	debug       = flag.Bool("debug", false, "Logging: More verbose.\n")
	quiet       = flag.Bool("quiet", false, "Logging: Less output. See -nolog\n")
	nolog       = flag.Bool("nolog", false, "Logging: Logs to /dev/null")
	logfile     = flag.String("log", "", "Logging: Use a log `file` instead of stdout\n\tExample: cosgo -log cosgo.log -debug\n")
	gpg         = flag.String("gpg", "", "GPG: Path to ascii-armored `public-key` to encrypt mbox\n)")
	sendgridKey = flag.String("sg", "", "Sendgrid: Sendgrid API `key` (disables mbox)\n")
	dest        = flag.String("to", "", "Email: Your email `address` (-sg flag required)\n")
	mboxfile    = flag.String("mbox", "cosgo.mbox", "Email: Custom mbox file `name`\n\tExample: cosgo -mbox custom.mbox\n\t")
)

func setup() {
	if !*quiet {
		fmt.Println(logo)
		fmt.Printf("\n\tcosgo v" + version + "\n")
		fmt.Printf("\tCopyright 2016 aerth\n\tSource code at https://github.com/aerth/cosgo\n")
	}
	_, cwd, staticDir, templateDir = initialize()

	//	http.Handle("/", r)

	if *debug && !*quiet {
		log.Println("booted:", timeboot)
		log.Println("working dir:", cwd)
		log.Println("css/js/img dir:", staticDir)
		log.Println("cosgo templates dir:", templateDir)
		log.Printf("binding to: %s:%s", *bind, *port)
	}

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
			// every X minutes
			time.Sleep(cosgoRefresh)
		}
	}()

	// Disable mbox if using sendgrid
	if *sendgridKey != "" {
		*mboxfile = os.DevNull
	}

	// Open mbox or dev/null
	f, err := os.OpenFile(*mboxfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Printf("error opening file: %v", err)
		log.Fatal("Hint: touch ./cosgo.mbox, or chown/chmod it so that the cosgo process can access it.")
		os.Exit(1)
	}

	// Set destination email for mbox and sendgrid
	if *dest != "" {
		destinationEmail = *dest
	}
	mbox.Destination = destinationEmail
	mbox.Mail = log.New(f, "", 0)

	// Is nolog enabled?
	if *nolog {
		*logfile = os.DevNull
	}
	// Switch to cosgo.log if not debug
	if *logfile != "" {
		openLogFile()
	}
	go fortuneInit()

}

func main() {
	setup()
	r := route(cwd, staticDir)
	// Try to bind
	oglistener, binderr := net.Listen("tcp", *bind+":"+*port)
	if binderr != nil {
		log.Println(binderr)
		os.Exit(1)
	}

	// Bind successful. Creating a stoppable listener.
	listener, stoperr := sl.New(oglistener)
	if stoperr != nil {
		log.Println(stoperr)
		os.Exit(1)
	}
	if *debug && !*quiet {
		log.Printf("Info: Got listener %s %s", listener.Addr().String(), listener.Addr().Network())
	}

	// Start Serving Loop
	for {
		listener, err = sl.New(oglistener)
		if err != nil {
			log.Fatalln(err)
		}
		go func() {
			if listener != nil {
				go http.Serve(listener,
					csrf.Protect(antiCSRFkey,
						csrf.HttpOnly(true),
						csrf.FieldName("cosgo"),
						csrf.CookieName("cosgo"),
						csrf.Secure(false))(r))
			} else {
				log.Fatalln("nil listener")
			}

		}()
		// End the for-loop with a timer/hit counter every 30 minutes.
		select {
		default:
			if !*quiet {
				log.Printf("Uptime: %s", time.Since(timeboot))
				log.Printf("Hits: %s", strconv.Itoa(hitcounter))
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
			fmt.Println("Got signal:", signal)
			fmt.Println("Dying")
			os.Exit(0)
		case signal := <-reload:
			fmt.Println("Got signal:", signal)
			fmt.Println("Dying")
			os.Exit(0)
		case signal := <-stop:
			fmt.Printf("Got signal:%v\n", signal)
			fmt.Println("Dying")
			os.Exit(0)
		}
	}()
	flag.Parse()

}
