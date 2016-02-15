package main

import (
	"flag"
	"github.com/dchest/captcha"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"time"
)

var (
	mandrillApiUrl   string
	mandrillKey      string
	casgoDestination string
	casgoAPIKey      string
)

type C struct {
CaptchaId string
}


const (
    // Default number of digits in captcha solution.
    DefaultLen = 6
    // The number of captchas created that triggers garbage collection used
    // by default store.
    CollectNum = 100
    // Expiration time of captchas used by default store.
    Expiration = 10 * time.Minute
)

const (
    // Standard width and height of a captcha image.
    StdWidth  = 240
    StdHeight = 80
)

func main() {

	// We can set the CASGO_API_KEY environment variable, or it defaults to a new random one!

	if os.Getenv("CASGO_API_KEY") == "" {
		log.Println("Generating Random API Key...")

		// The length of the API key can be modified here.
		casgoAPIKey = GenerateAPIKey(20)

		// Print new GenerateAPIKey
		log.Println("CASGO_API_KEY:", getKey())
	} else {
		casgoAPIKey = os.Getenv("CASGO_API_KEY")
		// Print selected CASGO_API_KEY
		log.Println("CASGO_API_KEY:", getKey())
	}
	//
	port := flag.String("port", "8080", "HTTP Port to listen on")
	debug := flag.Bool("debug", false, "be verbose, dont switch to logfile")
	insecure := flag.Bool("insecure", false, "accept insecure cookie transfer")
	mailbox := flag.Bool("mailbox", false, "save messages to an local mbox file")
	fastcgi := flag.Bool("fastcgi", false, "use fastcgi")
	bind := flag.String("bind", "127.0.0.1", "default: 127.0.0.1")
	flag.Parse()

	mandrillApiUrl = "https://mandrillapp.com/api/1.0/"
	mandrillKey = os.Getenv("MANDRILL_KEY")
	if mandrillKey == "" {
		log.Fatal("MANDRILL_KEY is Crucial. Type: export MANDRILL_KEY=123456789")
		os.Exit(1)
	}

	casgoDestination = os.Getenv("CASGO_DESTINATION")
	if casgoDestination == "" {
		log.Fatal("CASGO_DESTINATION is Crucial. Type: export CASGO_DESTINATION=\"your@email.com\"")
		os.Exit(1)
	}

	log.Printf("listening on http://127.0.0.1:%s", *port)

	r := mux.NewRouter()

	// Custom 404 redirect to /
	r.NotFoundHandler = http.HandlerFunc(RedirectHomeHandler)

	// Should be called BlankPageHandler
	r.HandleFunc("/", HomeHandler)

	// This is the meat, for behind a reverse proxy.
	r.HandleFunc("/"+casgoAPIKey+"/form", ContactHandler)
	r.HandleFunc("/"+casgoAPIKey+"/form/", ContactHandler)
	//	r.HandleFunc("/contact/", ContactHandler)

	// Magic URL Generator for API endpoint
	r.HandleFunc("/"+casgoAPIKey+"/send", EmailHandler)
	//r.Methods("GET").PathPrefix("/captcha2").Handler(captcha.Server(captcha.StdWidth, captcha.StdHeight))

	// Fun for 404s
	r.HandleFunc("/{whatever}", LoveHandler)
	r.Methods("GET").PathPrefix("/captcha/").Handler(captcha.Server(captcha.StdWidth, captcha.StdHeight))

	//http.Handle("/captcha/", captcha.Server(captcha.StdWidth, captcha.StdHeight))
	http.Handle("/", r)
	//r.HandleFunc("/captcha/",captcha.Server(captcha.StdWidth, captcha.StdHeight))

	// Switch to file log so we can ctrl+c and launch another instance :)

	if *mailbox == true {
		log.Println("mailbox mode: [sending mail to casgo.mbox]")
		//CreateMailBox()
	}

	if *debug == false {
		log.Println("quiet mode: [switching logs to casgo.log]")
		OpenLogFile()
	} else {
		log.Println("debug on: [not using casgo.log]")
	}

	if *fastcgi == true {
		log.Println("fastcgi [on]")
		log.Println("secure [off]")
		listener, err := net.Listen("tcp", *bind+":"+*port)
		if err != nil {
			log.Fatal("Could not bind: ", err)
		}
		log.Println("info: Listening on", *port)
	//	fcgi.Serve(listener, nil) // this works but without csrf..!
		fcgi.Serve(listener, csrf.Protect([]byte("LI80PNK1xcT01jmQBsEyxyrNCrbyyFPjPU8CKnxwmCruxNijgnyb3hXXD3p1RBc0+LIRQUUbTtis6hc6LD4I/A=="), csrf.HttpOnly(true), csrf.Secure(false))(r))
		//log.Fatal(fcgi.Serve( listener, csrf.Protect([]byte("LI80PNK1xcT01jmQBsEyxyrNCrbyyFPjPU8CKnxwmCruxNijgnyb3hXXD3p1RBc0+LIRQUUbTtis6hc6LD4I/A=="), csrf.HttpOnly(true), csrf.Secure(false))(r)))

	} else if *insecure == true {
		log.Println("info: Listening on", *port)
		log.Println("secure [off]")
		log.Fatal(http.ListenAndServe(":"+*port, csrf.Protect([]byte("LI80PNK1xcT01jmQBsEyxyrNCrbyyFPjPU8CKnxwmCruxNijgnyb3hXXD3p1RBc0+LIRQUUbTtis6hc6LD4I/A=="), csrf.HttpOnly(true), csrf.Secure(false))(r)))
	} else {
		log.Println("info: Listening on", *port)
		// Change this CSRF auth token in production!
		log.Println("secure [on]")
		log.Fatal(http.ListenAndServe(":"+*port, csrf.Protect([]byte("LI80PNK1xcT01jmQBsEyxyrNCrbyyFPjPU8CKnxwmCruxNijgnyb3hXXD3p1RBc0+LIRQUUbTtis6hc6LD4I/A=="), csrf.HttpOnly(true), csrf.Secure(true))(r)))
	}

}

// Key Generator
func init() {
	rand.Seed(time.Now().UnixNano())
}

var runes = []rune("____ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890123456789012345678901234567890")

func GenerateAPIKey(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = runes[rand.Intn(len(runes))]
	}
	return string(b)
}

// Which Key are we using again?
func getKey() string {
	return casgoAPIKey
}

// This function opens a log file. "debug.log"
func OpenLogFile() {
	f, err := os.OpenFile("./casgo.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0660)
	if err != nil {
		log.Fatal("error opening file: %v", err)
		os.Exit(1)
	}
	log.SetOutput(f)
}

// This is the home page it is blank. "This server is broken"
