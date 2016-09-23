package main

import (
	"bufio"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/dchest/captcha"
)

//openLogFile switches the log engine to a file, rather than stdout.
func openLogFile() {
	if *logfile == "" {
		return
	}
	f, err := os.OpenFile(*logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Printf("error opening file: %v", err)
		log.Fatal("Hint: touch " + *logfile + ", or chown/chmod it so that the cosgo process can access it.")
		os.Exit(1)
	}
	log.SetOutput(f)
}

func initialize() (time.Time, string, string, string) {

	// Load environmental variables as flags
	if os.Getenv("COSGO_PORT") != "" {
		*port = os.Getenv("COSGO_PORT")
	}

	if os.Getenv("COSGO_BIND") != "" {
		*bind = os.Getenv("COSGO_BIND")
	}

	if os.Getenv("COSGO_REFRESH") != "" {
		cosgoRefresh, err = time.ParseDuration(os.Getenv("COSGO_REFRESH"))
		if err != nil {
			log.Fatalln(err)
		}
	}

	if os.Getenv("COSGO_MBOX") != "" {
		*mboxfile = os.Getenv("COSGO_MBOX")
	}
	if os.Getenv("COSGO_LOG") != "" {
		*logfile = os.Getenv("COSGO_LOG")
	}
	if os.Getenv("COSGO_GPG") != "" {
		*gpg = os.Getenv("COSGO_GPG")
	}
	if *gpg != "" {
		log.Println("\t[gpg pubkey: " + *gpg + "]")
		publicKey = read2mem(rel2real(*gpg))
	}
	timeboot := time.Now()

	staticDir := staticFinder()
	templatesDir := templateFinder()

	return timeboot, cwd, staticDir, templatesDir

}

// templateFinder returns the template directory we will use, if one isn't found, the error is fatal.
func templateFinder() string {

	if *debug && !*quiet {
		log.Printf("Trying template directory %q", templateDir)
	}
	_, err = template.New("Index").Funcs(funcMap).ParseFiles(templateDir + "index.html")
	if err == nil {
		return templateDir
	}

	if strings.Contains(err.Error(), "no such file") {
		log.Println("No such file...", err.Error())
		log.Println("Creating!")
		err = RestoreAssets(".", "templates")
		if err != nil {
			log.Fatalln(err)
		}

		_, err = template.New("Index").Funcs(funcMap).ParseFiles(templateDir + "index.html")
		if err == nil {
			return templateDir
		}

	}

	if strings.Contains(err.Error(), "not defined") {
		log.Println("Template is bad.", err.Error())
		os.Exit(1)
	}

	return "./templates/"
}

// staticFinder returns the static directory. If none is found, static files are disabled.
func staticFinder() string {
	staticDir := "./static/"
	_, err = os.Open(staticDir)
	if err != nil {
		if os.IsNotExist(err) {
			staticDir = "/usr/local/share/cosgo/static/"
			_, err = os.Open(staticDir)
			if err != nil {
				if os.IsNotExist(err) {
					if *debug {
						log.Printf("No staticDir. Creating one.")
					}
					RestoreAssets(".", "static")
					staticDir = "./static/"
					return staticDir
				}
			}
		}
	}

	return staticDir
}

func getKey() string {
	return cosgo.URLKey
}
func getDestination() string {
	return destinationEmail
}
func newfortune() string {
	if len(fortunes) == 0 {
		return ""
	}
	n := rand.Intn(len(fortunes))
	log.Println("Fortune #", n)
	return fortunes[n]
}

var fortunes = map[int]string{}

func fortuneInit() {
	file, err := os.Open("fortunes.txt")
	if err != nil {
		log.Println("New feature: fortunes. Looks for fortunes.txt. Format double line separated. available as a template variable {{.Fortune}}")
		return
	}
	b := make([]byte, 1024*1000)
	n, err := file.Read(b)
	str := string(b[:n])
	scanner := bufio.NewScanner(strings.NewReader(str))
	var i = 1
	var buf string
	for scanner.Scan() {
		if scanner.Text() == "" {
			if buf != "" {

				fortunes[i] = buf
				i++
			}
			buf = ""
			continue
		}
		buf = buf + scanner.Text() + "\n"
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}

	log.Println(len(fortunes), "Fortunes")
}
func read2mem(abspath string) []byte {
	file, err := os.Open(abspath) // For read access.
	if err != nil {
		log.Fatal(err)
	}

	data := make([]byte, 4096)
	i, err := file.Read(data)
	if err != nil {
		log.Fatal(err)
	}

	return data[:i]

}
func rel2real(file string) (realpath string) {
	pathdir, _ := path.Split(file)
	if pathdir == "" {
		realpath, _ = filepath.Abs(file)
	} else {
		realpath = file
	}
	return realpath
}

// getDomain returns the domain name (without port) of a request.
func getDomain(r *http.Request) string {
	type Domains map[string]http.Handler
	hostparts := strings.Split(r.Host, ":")
	requesthost := hostparts[0]
	return requesthost
}

// getSubdomain returns the subdomain (if any)
func getSubdomain(r *http.Request) string {
	type Subdomains map[string]http.Handler
	hostparts := strings.Split(r.Host, ":")
	requesthost := hostparts[0]
	if net.ParseIP(requesthost) == nil {
		//log.Println("Info: " + requesthost)
		domainParts := strings.Split(requesthost, ".")
		if len(domainParts) > 2 {
			if domainParts[0] != "127" {
				return domainParts[0]
			}
		}
	}
	return ""
}

//generateURLKey creates a new key, with the given runes, n length.
func generateURLKey(n int) string {
	runes := []rune("____ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890123456789012345678901234567890")
	b := make([]rune, n)
	for i := range b {
		b[i] = runes[rand.Intn(len(runes))]
	}
	return strings.TrimSpace(string(b))
}

// Returns human time since ( Example 3 weeks ago or 11 hours ago)
func timesince(anything interface{}) string {
	var date string
	var t time.Time
	switch reflect.TypeOf(anything).Kind() {
	case reflect.String:
		date = anything.(string)
		const longForm = "Jan 2, 2006 at 3:04pm (MST)"
		t, err = time.Parse(longForm, date)
		if err != nil {
			log.Println(err)
			return "Unknown"
		}

	default:
		t = anything.(time.Time)

	}

	since := time.Since(t)
	const year = 365 * time.Hour * 24

	// Minutes ago
	if since < time.Hour*1 {
		return strconv.FormatFloat(since.Minutes(), 'f', 0, 64) + " minutes ago"
	}
	// Hours ago
	if since < time.Hour*24 {
		return strconv.FormatFloat(since.Hours(), 'f', 0, 64) + " hours ago"
	}
	// Days ago
	if since < time.Hour*24*7 {
		return strconv.FormatFloat(since.Hours()/24, 'f', 0, 64) + " days ago"
	}

	// Weeks ago
	if since < time.Hour*24*7*4 {
		return strconv.FormatFloat(since.Hours()/(24*7), 'f', 0, 64) + " weeks ago"
	}
	// Months ago

	if since < time.Hour*24*7*4*12 {
		return strconv.FormatFloat(since.Hours()/(24*7*4), 'f', 0, 64) + " months ago"
	}
	// Years ago
	if t.Year() != time.Now().Year() {
		return strconv.Itoa(time.Now().Year()-t.Year()) + " years ago"
	}

	if time.Hour*24 < since && since < time.Hour*48 {
		return "Yesterday"
	}

	//
	floaty := (since.Hours()) / 24

	ago := strconv.FormatFloat(floaty, 'f', 0, 64) + " days ago!"
	//ago := strconv.FormatFloat(floaty, 'E', -1, 32) + " days ago!"
	fmt.Println(ago)
	return ago

}
func verifyKey(r *http.Request) bool {

	// userkey is the user submitted URL key
	userkey := strings.TrimLeft(r.URL.Path, "/")
	userkey = strings.TrimRight(userkey, "/send")
	// Check URL Key
	if *debug {
		log.Printf("\nComparing... \n\t" + userkey + "\n\t" + cosgo.URLKey)
	}
	if userkey != cosgo.URLKey {
		log.Println("Key Mismatch. ", r.UserAgent(), r.RemoteAddr, r.RequestURI+"\n")
		return false
	}
	return true
}

func verifyCaptcha(r *http.Request) bool {
	if !captcha.VerifyString(r.FormValue("captchaId"), r.FormValue("captchaSolution")) {
		log.Printf("User Error: CAPTCHA %s at %s", r.UserAgent(), r.RemoteAddr)
		return false
	}

	return true
}
