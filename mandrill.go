package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/aerth/cosgo/mbox"
	"github.com/goware/emailx"
	mandrill "github.com/keighl/mandrill"
)

func canMandrill() bool {
	uinfo := mandrillAPIUrl + "users/info.json"
	var jsonStr = []byte(`
	{
	    "key": "` + mandrillKey + `"
	}`)

	req, err := http.NewRequest("POST", uinfo, bytes.NewBuffer(jsonStr))
	req.Header.Set("X-Custom-Header", "key")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	split := strings.SplitAfter(string(body), "last_30_days\":{\"sent\":")
	sentString := strings.SplitAfter(split[1], ",")[0]
	sendCount, _ := strconv.Atoi(sentString[:len(sentString)-1])
	if sendCount >= 10000 {
		log.Fatal("Monthly quota reached")
		return false
	}
	return true
}

func sendMandrill(destinationEmail string, form *mbox.Form) bool {
	//if !canMandrill() {

	//	return false
	//}

	client := mandrill.ClientWithKey(mandrillKey)

	message := &mandrill.Message{}
	message.AddRecipient(destinationEmail, destinationEmail, "to")
	message.FromEmail = form.Email
	message.FromName = form.Name
	message.Subject = form.Subject
	message.HTML = "<p>" + form.Message + "<p>"
	message.Text = form.Message
	responses, err := client.MessagesSend(message)
	//log.Println(responses)
	if err != nil {
		log.Println(err)
		return false
	}
	length := len(responses)
	for i := 0; i < length; i++ {
		//if responses[i].Status == "sent" {
		//	return true
		// } else if responses[i].Status == "rejected" {
		if responses[i].Status == "rejected" {
			return false
		} else if responses[i].Status == "invalid" {
			return false
		} else if responses[i].Status == "must" {
			return false
		}
	}
	return true
}

// mandrillSender always returns success for the visitor. This function needs some work.
func mandrillSender(rw http.ResponseWriter, r *http.Request, destination string, query url.Values) error {
	form := mbox.ParseQueryGPG(query, publicKey)
	//Validate user submitted email address
	err = emailx.Validate(form.Email)
	if err != nil {
		if err == emailx.ErrInvalidFormat {
			return errors.New("Bad email format.")
		}
		if err == emailx.ErrUnresolvableHost {
			return errors.New("Bad email provider.")
		}
		return errors.New("Bad email.")
	}
	//Normalize email address
	form.Email = emailx.Normalize(form.Email)
	//Is it empty?
	if form.Email == "" || form.Email == "@" {
		http.Redirect(rw, r, "/", 301)
		return errors.New("Bad Email")
	}

	if sendMandrill(destination, form) {
		return nil
	}
	return errors.New("Can't send email")

}
