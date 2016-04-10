package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/goware/emailx"
	mandrill "github.com/keighl/mandrill"
)

func canMandrill() bool {
	user_count_url := mandrillAPIUrl + "users/info.json"
	var jsonStr = []byte(`
	{
	    "key": "` + mandrillKey + `"
	}`)

	req, err := http.NewRequest("POST", user_count_url, bytes.NewBuffer(jsonStr))
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
	sent_str := strings.SplitAfter(split[1], ",")[0]
	send_count, _ := strconv.Atoi(sent_str[:len(sent_str)-1])
	if send_count >= 10000 {
		log.Fatal("Monthly quota reached")
		return false
	}
	return true
}

func sendMandrill(destinationEmail string, form *Form) bool {
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
	form := parseQuery(query)
	//Validate user submitted email address
	err = emailx.Validate(form.Email)
	if err != nil {
		fmt.Fprintln(rw, "<html><p>Email is not valid. Would you like to go <a href=\"/\">back</a>?</p></html>")

		if err == emailx.ErrInvalidFormat {
			fmt.Fprintln(rw, "<html><p>Email is not valid format.</p></html>")
		}
		if err == emailx.ErrUnresolvableHost {
			fmt.Fprintln(rw, "<html><p>We don't recognize that email provider.</p></html>")
		}
	}
	//Normalize email address
	form.Email = emailx.Normalize(form.Email)
	//Is it empty?
	if form.Email == "" || form.Email == "@" {
		http.Redirect(rw, r, "/", 301)
		return errors.New("Blank Email")
	}

	if sendMandrill(destination, form) {
		fmt.Fprintln(rw, "<html><p>Thanks! Would you like to go <a href=\"/\">back</a>?</p></html>")
		log.Printf("SUCCESS-contact: %s at %s", r.UserAgent(), r.RemoteAddr)
		return nil
	} else {
		log.Printf("FAIL-contact: %s at %s", r.UserAgent(), r.RemoteAddr)
		log.Printf("debug: %s to mandrill %s", form, destination)
		log.Printf("debug: %s to mandrill %s", form.Message, destination)

		t, err := template.New("Error").ParseFiles("./templates/error.html")
		if err == nil {
			data := map[string]interface{}{
				"err":            "Mail System",
				"Key":            getKey(),
				csrf.TemplateTag: csrf.TemplateField(r),
			}
			t.ExecuteTemplate(rw, "Error", data)
			return err
		} else {
			log.Printf("template error: %s at %s", r.UserAgent(), r.RemoteAddr)
			log.Println(err)
			http.Redirect(rw, r, "/", 301)
			return errors.New("error.html template error.")

		}
	}
}
