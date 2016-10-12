package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	clr "github.com/mewkiz/pkg/term"
)

/*

cosgo test suite:


So far, testing for a few things:
	1. POST URL Key MUST ALWAYS be served in the normal home page
	2. POST URL Key MUST ALWAYS be present in user's POST message submissions
	3. mbox gets bytes written to it on successful message send
	4. Asset directories are restored (static, template) if they dont exist
	5. Mbox is created on boot
	6. CRONJOBs will not be broken. CLI Flags will not change for a while.

TODO:
	1. Config tests
	2. Benchmarks
	3. Stress test
	4. Template fields testing
*/

var times []time.Duration
var redwanted = clr.Red("Wanted:")
var rednotfound = ""

// thanks strings
func grep(source interface{}, target string) string {
	var str string
	switch source.(type) {
	case []byte:
		str = string(source.([]byte))
	case string:
		str = source.(string)
	default:
		// working on it
	}
	defer func() {
		backup([]byte(str))
	}()

	lines := strings.Split(str, "\n")

	for _, line := range lines {
		if strings.Contains(line, target) {
			return clr.MagentaBold("Found: ") + clr.MagentaBold(line)
		}
	}
	return ""
}

func backup(b []byte) {
	tmpfile, _ := ioutil.TempFile(os.TempDir(), "c_test_"+strconv.Itoa(int(time.Now().Truncate(1*time.Minute).Unix())))
	n, e := tmpfile.Write(b)
	if e != nil {
		panic(e)
	}

	go func() {

		fmt.Println("\t" + clr.Green(strconv.Itoa(n)+" bytes saved to "+tmpfile.Name()))

	}()
}
func init() {
	rand.NewSource(time.Now().UnixNano())

	*quiet = true
	*nolog = true
	*mboxfile = "testing.mbox"

}
func randomDuration() (times []time.Duration) {
	t := uint64(rand.Uint32()*4) * uint64(rand.Uint32()) * uint64(rand.Uint32()*rand.Uint32()) / uint64(rand.Intn(10000)+1)
	times = append(times, time.Duration(t))
	return times
}

func BenchmarkHumanize(b *testing.B) {
	for i := 0; i < 1000; i++ {
		rands := randomDuration()
		for _, t := range rands {
			b.StartTimer()
			humanize(t)
			b.StopTimer()
		}
	}
}

func TestHumanize(t *testing.T) {
	d := time.Duration(time.Hour * 24 * 7 * 4 * 8)
	s := humanize(d)
	if s != "8 months ago" {
		fmt.Println("\tGot:", s)
		fmt.Println("\t"+redwanted, "8 months ago")
		t.FailNow()
	}
	// Output: 8 months ago
}

// TestHomeHandler tests the home page, and makes sure the URLKey is in the output.
func TestHomeHandler(t *testing.T) {
	c := setup()
	cwd, _ := os.Getwd()
	c.route(cwd)

	ts := httptest.NewServer(http.HandlerFunc(c.homeHandler))
	defer ts.Close()

	time.Sleep(100 * time.Millisecond)
	key := c.URLKey
	res, err := http.Get(ts.URL)
	if err != nil {
		log.Fatal(err)
	}
	greeting, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	keyline := `<form id="contact-form" action="/` + key + `/send" method="POST">`
	if !strings.Contains(string(greeting), keyline) {
		fmt.Println("\t"+redwanted, keyline)

		fmt.Println("\t" + grep(greeting, `<form id="contact-form" action="/`))
		t.FailNow()
	}

	return
}

// TestEmailHandler tests that the URLKey allows sending message (later we test the mbox)
func TestEmailHandler(t *testing.T) {
	flag.Parse()
	verifyCaptcha = func(r *http.Request) bool { return true } // hack the captcha
	//*logfile = os.DevNull

	c2 := setup()
	cwd, _ := os.Getwd()
	c2.route(cwd)

	ts := httptest.NewServer(c2.r)
	defer ts.Close()

	time.Sleep(100 * time.Millisecond)
	res, err := http.Get(ts.URL)
	if err != nil {
		log.Fatal(err)
	}
	greeting, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	cut := strings.Split(string(greeting), `<p><img id="image" src="/captcha/`)
	captcha := cut[1][:20]
	v := &url.Values{}
	v.Add("email", "joe@joe.com")
	v.Add("subject", "hello world")
	v.Add("message", "from test")
	v.Add("cosgo", "123")
	v.Add("captchaId", captcha)

	res, err = http.PostForm(ts.URL+"/"+c2.URLKey+"/send", *v)
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	greeting, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	keyline := `Thanks! Your message was sent.`
	if !strings.Contains(string(greeting), keyline) {
		fmt.Println("\t"+redwanted, keyline)

		fmt.Println("\t" + grep(greeting, `Your message`))
		if *debug {
			fmt.Println(string(greeting))
		}
		t.FailNow()
	}
	return
}

// TestEmailHandlerIncorrectKey tests that an incorrect URLKey doesn't work
func TestEmailHandlerIncorrectKey(t *testing.T) {
	verifyCaptcha = func(r *http.Request) bool { return true } // hack the captcha
	//*logfile = os.DevNull

	c2 := setup()
	cwd, _ := os.Getwd()
	c2.route(cwd)

	ts := httptest.NewServer(c2.r)
	defer ts.Close()

	time.Sleep(100 * time.Millisecond)
	res, err := http.Get(ts.URL)
	if err != nil {
		log.Fatal(err)
	}
	greeting, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	cut := strings.Split(string(greeting), `<p><img id="image" src="/captcha/`)

	captcha := cut[1][:20]

	v := &url.Values{}
	v.Add("email", "joe@joe.com")
	v.Add("subject", "hello world")
	v.Add("message", "bad user")
	v.Add("cosgo", "123")
	v.Add("captchaId", captcha)

	time.Sleep(100 * time.Millisecond)
	res, err = http.PostForm(ts.URL+"/fake"+c2.URLKey+"/send", *v)
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	greeting, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	keyline := `Your message was not sent: Bad endpoint.`
	if !strings.Contains(string(greeting), keyline) {
		fmt.Println("\t"+redwanted, keyline)

		fmt.Println("\t" + grep(greeting, `Your message`))
		if *debug {
			fmt.Println(string(greeting))
		}
		t.FailNow()
	}
	//fmt.Println("\tFound it :)")
	//fmt.Println()
	return
}

// TestEmailHandlerIncorrectEmail tests that an incorrect email doesn't work
func TestEmailHandlerIncorrectEmail(t *testing.T) {
	verifyCaptcha = func(r *http.Request) bool { return true } // hack the captcha
	//*logfile = os.DevNull

	c2 := setup()
	cwd, _ := os.Getwd()
	c2.route(cwd)

	ts := httptest.NewServer(c2.r)
	defer ts.Close()

	time.Sleep(100 * time.Millisecond)
	res, err := http.Get(ts.URL)
	if err != nil {
		log.Fatal(err)
	}
	greeting, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	cut := strings.Split(string(greeting), `<p><img id="image" src="/captcha/`)

	captcha := cut[1][:20]

	v := &url.Values{}
	v.Add("email", "joejoe.com")
	v.Add("subject", "hello world")
	v.Add("message", "bad user")
	v.Add("cosgo", "123")
	v.Add("captchaId", captcha)

	res, err = http.PostForm(ts.URL+"/"+c2.URLKey+"/send", *v)
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	greeting, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	keyline := `Your message was not sent: Bad email address.`
	if !strings.Contains(string(greeting), keyline) {
		fmt.Println("\t"+redwanted, keyline)

		fmt.Println("\t" + grep(greeting, `Your message`))
		if *debug {
			fmt.Println(string(greeting))
		}
		t.FailNow()
	}
	// fmt.Println("\tFound it :)")
	// fmt.Println()
	return
}

// TestMbox tests only that the amount of bytes in the test mbox is 235.
// This test needs some work
func TestMbox(t *testing.T) {

	b, e := ioutil.ReadFile(*mboxfile)
	if e != nil {
		fmt.Println(e)
		os.Exit(2)
	}

	// Remove test mbox file
	defer func() {
		e = os.Remove(*mboxfile)
		if e != nil {
			t.FailNow()
		}
	}()

	if len(b) != 235 {
		fmt.Println(len(b), "!= 245")
		t.FailNow()
	}

	//	fmt.Println("\tContents of test mbox:", *mboxfile)
	//fmt.Println(string(b))
	fmt.Println("\tRemoving test mbox:", *mboxfile)

}
func TestAssets(t *testing.T) {
	tmpdir, er := ioutil.TempDir(os.TempDir(), "cosgoTest")
	if er != nil {
		panic(er)
	}
	os.Chdir(tmpdir)
	c2 := setup()
	cwd, _ := os.Getwd()
	c2.route(cwd)
	files, e := ioutil.ReadDir(tmpdir)
	if e != nil {
		panic(e)
	}

	var static, template, mbox bool
	for _, i := range files {
		if i.Name() == "static" {
			static = true
		}
		if i.Name() == "templates" {
			template = true
		}
		if i.Name() == "testing.mbox" {
			mbox = true
		}
	}
	fmt.Println("\tCreated mbox:", mbox)
	fmt.Println("\tCreated static:", static)
	fmt.Println("\tCreated template dir:", template)

	if !mbox || !template || !static {
		t.FailNow()
	}
}

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

// TestEmailHandlerTimeout tests what happens after a refresh
func TestEmailHandlerTimeout(t *testing.T) {
	verifyCaptcha = func(r *http.Request) bool { return true } // hack the captcha
	//*logfile = os.DevNull
	*refreshTime = 1 * time.Nanosecond // Really quick...
	flag.Parse()
	c3 := setup() // Timer is started
	cwd, _ := os.Getwd()
	c3.route(cwd)
	ts := httptest.NewServer(c3.r)
	t1 := time.Now()
	defer ts.Close()
	res, err := http.Get(ts.URL)
	if err != nil {
		log.Fatal(err)
	}
	greeting, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	cut := strings.Split(string(greeting), `<p><img id="image" src="/captcha/`)
	captcha := cut[1][:20]
	v := &url.Values{}
	v.Add("email", "joe@joe.com")
	v.Add("subject", "hello world")
	v.Add("message", "from test")
	v.Add("cosgo", "123")
	v.Add("captchaId", captcha)
	oldkey := c3.URLKey
	var t2 time.Time
	for {
		time.Sleep(1 * time.Nanosecond)
		if oldkey != c3.URLKey {
			t2 = time.Now()
			break
		}
	}
	fmt.Printf("Took %s to change keys\n", t2.Sub(t1))
	res, err = http.PostForm(ts.URL+"/"+oldkey+"/send", *v)
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	greeting, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	keyline := `Your message was not sent: Bad endpoint.`
	if !strings.Contains(string(greeting), keyline) {

		fmt.Println("\t"+redwanted, keyline)

		fmt.Println("\t" + grep(greeting, `Your message`))
		if *debug {
			fmt.Println(string(greeting))
		}
		t.FailNow()
	}
	//	fmt.Println("\tFound it :)")
	//	fmt.Println()
	return
}

// Test that this line does not break:
// 'cosgo -gpg /etc/aerth.asc -fastcgi -log cosgo.log -port 8089'
// Lots of flags: bind config debug fastcgi gpg log mbox new nolog port quiet sg title to
// No more breaking CRONTABS on cosgo upgrades!
// These flag tests need to be at the bottom of the test suite
func TestFlags(t *testing.T) {
	const PanicOnError = true
	devnull, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0700)
	flag.CommandLine.SetOutput(devnull)
	flag.CommandLine.Init("testflags", flag.ErrorHandling(2))
	flag.Usage = func() {}
	orig := os.Args
	os.Args = []string{"cosgo", "-gpg", "/etc/aerth.asc", "-fastcgi", "-log", "cosgo.log", "-port", "8089"}
	flag.Parse()
	os.Args = []string{"cosgo", "-mbox", "test.mbox", "-bind", "0.0.0.0", "-debug"}
	flag.Parse()
	os.Args = []string{"cosgo", "-title", "Test Page", "-fastcgi", "-log", "cosgo.log", "-port", "8089"}
	flag.Parse()
	os.Args = []string{"cosgo", "-config", "/tmp/config.cosgo", "-nolog", "-sg", "sgkey123", "-to", "me@mine.com"}
	flag.Parse()
	os.Args = orig
	flag.Parse()
	return
}
func TestBadFlags(t *testing.T) {
	// Test that this line *does* break:
	teststring := []string{"-bad"}
	const ContinueOnError = true
	devnull, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0700)
	flag.CommandLine.SetOutput(devnull)
	flag.Usage = func() {}
	flag.CommandLine.Init("testflags", flag.ErrorHandling(0))
	e := flag.CommandLine.Parse(teststring)
	if e == nil {
		fmt.Println("\tExpected Error. Got nil.")
		t.FailNow()
	} else if e.Error() == "flag provided but not defined: -bad" {
		//
	} else {
		fmt.Println("\tExpected a different error:", e)
	}
	return
}
