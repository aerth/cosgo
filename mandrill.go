package main

import (
	"bytes"
	mandrill "github.com/keighl/mandrill"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Form struct {
	Name, Email, Subject, Message string
}

func canSendEmail() bool {
	user_count_url := mandrillApiUrl + "users/info.json"
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

func sendEmail(destinationEmail string, form *Form) bool {
	//if !canSendEmail() {

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
