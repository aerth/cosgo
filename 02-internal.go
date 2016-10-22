package main

import (
	"bufio"
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
	"github.com/gorilla/mux"
)

//openLogFile switches the log engine to a file, rather than stdout.
func openLogFile() {
	if *logfile == "" {
		return
	}
	if *logfile == "stdout" {
		*logfile = os.Stdout.Name()
	}
	if *logfile == "stderr" {
		*logfile = os.Stderr.Name()
	}
	f, ferr := os.OpenFile(*logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if ferr != nil {
		log.Printf("error opening file: %v", ferr)
		log.Fatal("Hint: touch " + *logfile + ", or chown/chmod it so that the cosgo process can access it.")
		os.Exit(1)
	}
	log.SetOutput(f)

}

func (c *Cosgo) initialize() error {

	// Load environmental variables as flags
	if os.Getenv("COSGO_PORT") != "" {
		c.Port = os.Getenv("COSGO_PORT")
	}

	if os.Getenv("COSGO_BIND") != "" {
		c.Bind = os.Getenv("COSGO_BIND")
	}

	if os.Getenv("COSGO_REFRESH") != "" {
		*refreshTime, err = time.ParseDuration(os.Getenv("COSGO_REFRESH"))
		if err != nil {
			return err
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

	// GPG is a flag, but can be overrided by ENV
	if *gpg != "" {
		if !*quiet {
			log.Println("[+gpg]")
		}
		c.publicKey = read2mem(rel2real(*gpg))
	}
	c.boot = time.Now()
	c.staticDir = staticFinder()
	c.templatesDir = templateFinder()
	c.r = mux.NewRouter()

	return nil
}

// templateFinder returns a directory string where template "index.html" is located.
// We also parse the template to test whether or not we should boot any further.
func templateFinder() string {
	templateDir := "./templates/"
	// Try to parse
	_, err = template.New("Index").Funcs(funcMap).ParseFiles(templateDir + "index.html")
	if err == nil {
		return templateDir
	}

	// Does not exist
	if strings.Contains(err.Error(), "no such file") {
		log.Println("Creating ./templates")
		err = RestoreAssets(".", "templates")
		if err != nil {
			log.Fatalln(err)
		}

		// Try to parse
		_, err = template.New("Index").Funcs(funcMap).ParseFiles(templateDir + "index.html")
		if err == nil {
			return templateDir
		}

	} else if strings.Contains(err.Error(), "not defined") {
		log.Println("Template is bad.", err.Error())
		os.Exit(1)
	}

	// The error is probably permissions and theres nothing more to be done
	log.Fatalln("Template:", err)
	return ""
}

// staticFinder returns the directory path where static files will be served from.
// If it doesn't exist, it is created at ./static
func staticFinder() string {
	staticDir := "./static/"
	_, err = os.Open(staticDir)
	if err == nil {
		return staticDir
	}

	// try global
	staticDir = "/usr/local/share/cosgo/static/"
	_, err = os.Open(staticDir)
	if err == nil {
		return staticDir
	}

	// create
	log.Printf("Creating ./static")
	RestoreAssets(".", "static")
	staticDir = "./static/"
	return staticDir

}

func (c *Cosgo) getKey() string {
	return c.URLKey
}
func (c *Cosgo) getDestination() string {
	return c.Destination
}

// Open file into bytes
func read2mem(abspath string) []byte {
	file, ferr := os.Open(abspath) // For read access.
	if ferr != nil {
		log.Fatal("Cosgo fatal:", ferr)
	}

	var data []byte
	i, rerr := file.Read(data)
	if rerr != nil {
		log.Fatal("Cosgo fatal:", rerr)
	}

	return data[:i]

}

// Relative path to real path
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
	hostparts := strings.Split(r.Host, ":")
	requesthost := hostparts[0]
	return requesthost
}

func getSubdomain(r *http.Request) string {
	hostparts := strings.Split(r.Host, ":")
	dots := strings.Count(hostparts[0], ".")
	if dots < 2 {
		return "" // is a top level domain name
	}

	if net.ParseIP(hostparts[0]) != nil {
		return "" // is an IP address
	}

	domainParts := strings.Split(hostparts[0], ".")

	if len(domainParts) > 2 {
		return strings.Join(domainParts[0:len(domainParts)-2], ".")
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

	// Years ago
	if t.Year() != time.Now().Year() {
		str := strconv.Itoa(time.Now().Year() - t.Year())
		if str == "1" {
			return str + " year ago"
		}
		return str + " years ago"

	}

	return humanize(since)

}

// Humanize a duration of time
func humanize(since time.Duration) string {

	// Minutes ago
	if since < time.Hour*1 {
		str := strconv.FormatFloat(since.Minutes(), 'f', 0, 64)
		if str == "1" {
			return str + " minute ago"
		}
		return str + " minutes ago"
	}
	// Hours ago
	if since < time.Hour*24 {
		str := strconv.FormatFloat(since.Hours(), 'f', 0, 64)
		if str == "1" {
			return str + " hour ago"
		}
		return str + " hours ago"
	}
	// Days ago
	if since < time.Hour*24*7 {
		str := strconv.FormatFloat(since.Hours()/24, 'f', 0, 64)
		if str == "1" {
			return "yesterday"
		}
		return str + " days ago"
	}

	// Weeks ago
	if since < time.Hour*24*7*4 {
		str := strconv.FormatFloat(since.Hours()/(24*7), 'f', 0, 64)
		if str == "1" {
			return str + " week ago"
		}
		return str + " weeks ago"
	}

	// Months ago
	if since < time.Hour*24*7*4*12 {
		str := strconv.FormatFloat(since.Hours()/(24*7*4), 'f', 0, 64)
		if str == "1" {
			return str + " month ago"
		}
		return str + " months ago"
	}

	const year = time.Hour * 24 * 365
	if since < year {
		str := strconv.FormatFloat(since.Hours()/(24*7*4), 'f', 0, 64)
		if str == "1" {
			return str + " month ago"
		}
		return str + " months ago"
	}

	//
	floaty := (since.Hours()) / 24

	ago := strconv.FormatFloat(floaty, 'f', 0, 64) + " days ago"

	return ago

}
func (c *Cosgo) verifyKey(r *http.Request) bool {

	// userkey is the user submitted URL key
	userkey := strings.TrimLeft(r.URL.Path, "/")
	userkey = strings.TrimRight(userkey, "/send")
	// Check URL Key
	if *debug {
		log.Printf("\nComparing... \n\t" + userkey + "\n\t" + c.URLKey)
	}
	if userkey != c.URLKey {
		log.Println("Key Mismatch. ", r.UserAgent(), r.RemoteAddr, r.RequestURI+"\n")
		return false
	}
	return true
}

// verifyCaptcha is a variable func for testing
var verifyCaptcha = func(r *http.Request) bool {
	if !captcha.VerifyString(r.FormValue("captchaId"), r.FormValue("captchaSolution")) {
		log.Printf("User Error: CAPTCHA %s at %s", r.UserAgent(), r.RemoteAddr)
		log.Println(r.FormValue("captchaId"))
		log.Println(r.FormValue("captchaSolution"))
		return false
	}
	return true
}

func newfortune() string {
	if len(fortunes) == 0 {
		return ""
	}
	n := rand.Intn(len(fortunes))
	log.Println("Fortune #", n)
	return fortunes[n]
}

// Fortunes!
var fortunes = map[int]string{}

func fortuneInit() {
	file, err := os.Open("fortunes.txt")
	if err != nil {
		if !*quiet {
			log.Println("No 'fortunes.txt' file.")
		}
		return
	}
	var b []byte
	n, err := file.Read(b)
	if err != nil {
		log.Println("Fortunes:", err)
		return
	}
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
		buf += scanner.Text() + "\n"
	}
	if err := scanner.Err(); err != nil {
		log.Fatalln("Can't read fortunes.txt somewhere near line #", i)
	}
	if !*quiet {
		log.Println(len(fortunes), "Fortunes")
	}
}
