package controllers

import (
	"fmt"
	"bytes"
	"strconv"
	"strings"
	"io/ioutil"
	"math/rand"
	"net/http"
	mandrill "github.com/keighl/mandrill"
	"time"
	"github.com/revel/revel"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
    mgoSession     *mgo.Session
    tokenCollection *mgo.Collection
    mandrillKey string
    mandrillApiUrl string
)

func getSession () *mgo.Session {
    if mgoSession == nil {
    	mgoSession = dialMongo()
    }
    return mgoSession.Clone()
}

func canSendEmail () bool {
	if mandrillApiUrl == "" {
		mandrillApiUrl = "https://mandrillapp.com/api/1.0/"
	}
	user_count_url := mandrillApiUrl + "users/info.json"
	var jsonStr = []byte(`
	{
	    "key": "` + getMandrillKey() + `"
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
    send_count, _ := strconv.Atoi(sent_str[:len(sent_str) - 1])
    if send_count >= 10000 {
    	return false
    }
    return true
}

func getUserCollection () *mgo.Collection {
	if tokenCollection == nil {
		session := getSession()
		tokenCollection = session.DB("static-contact").C("static-contact-tokens")
	}
	return tokenCollection
}

func getMandrillKey () string {
	if mandrillKey == "" {
		var found bool
		mandrillKey, found = revel.Config.String("mandrillKey")
		if !found {
			panic("Mandrill API key not set in app.conf.")
		}
		if len(mandrillKey) == 0 {
			panic("Mandrill API key is empty.")
		}
	}
	return mandrillKey
}

func dialMongo () *mgo.Session {
	mongoUri, mongoFound := revel.Config.String("mongoLabUri")
	
	if !mongoFound {
		panic("Mongo URI not set in app.conf.")
	}
	if len(mongoUri) == 0 {
		panic("Mongo URI is invalid with length 0")
	}

    mgoSession, err := mgo.Dial(mongoUri)
    if err != nil {
         panic(err)
    }
    return mgoSession.Clone()
}

type Contact struct {
	*revel.Controller
}

type Form struct {
	Name, Email, Subject, Message string
}

type recipientRecord struct {
	Token string
	Email string
	MonthlyCount int
	CurrentMonth int
}

func (c Contact) ContactA() revel.Result {
	email := c.Params.Get("id")
	form := Form{Name: c.Params.Get("name"),
		Email:   c.Params.Get("email"),
		Subject: c.Params.Get("subject"),
		Message: c.Params.Get("message")}
	sent := sendEmail(email, form)
	if sent {
		return c.Redirect(App.Success)
	} else {
		return c.Redirect(App.Failure)
	}
}

func (c Contact) Contact(id string, name string, email string, subject string, message string) revel.Result {
	form := Form{Name: name,
		Email:   email,
		Subject: subject,
		Message: message}
	fmt.Printf("destination: %s\nname: %s\nemail: %s\nsubject: %s\nmessage: %s\n", id, name, email, subject, message)
	sent := sendEmail(id, form)
	if sent {
		return c.Redirect(App.Success)
	} else {
		return c.Redirect(App.Failure)
	}
}

func (c Contact) Register() revel.Result {
	email := c.Params.Get("email")
	return c.Render(email)
}

func (c Contact) RegisteredEmail() revel.Result {
	email := c.Params.Get("ids")
	return c.Render(email)
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var tokenLength = 8

func generateToken() string {
	b := make([]rune, tokenLength)
	for i := range b {
	    b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func createEmailRecord(email string) bool {
	// Create a new user
	var token string
	for {
		token = generateToken()
		if !tokenExistsInDB(token){
			break
		}
	}
	
	_, month, _ := time.Now().Date()

	newRecipient := recipientRecord {Token: token, 
									 Email: email, 
									 MonthlyCount: 1, 
									 CurrentMonth: int(month)}
	// Add the user to the DB
	c := getUserCollection()
	err := c.Insert(&newRecipient)
	if err != nil {
	    fmt.Printf("%s\n", err)
	}
	return true
}

func tokenExistsInDB(token string) bool {
	c := getUserCollection()
	result := recipientRecord{}
	err := c.Find(bson.M{"Token": token}).One(&result)
	if err != nil {
		// The token does not exist
		return false
	}
	return true
}

func incrementSentEmailCount(email string) bool {
	c := getUserCollection()
	oldRecord := recipientRecord{}
	err := c.Find(bson.M{"email": email}).One(&oldRecord)
	if err != nil {
		createEmailRecord(email)
		return true
	}
	newRecord := oldRecord
	newRecord.MonthlyCount += 1
	err = c.Update(bson.M{"email": email}, &newRecord)
	if err != nil {
		panic(err)
	}
	return true
}

func sendEmail(destinationEmail string, form Form) bool {
	if !canSendEmail() {
		return false
	}
	emailKey := getMandrillKey()

	client := mandrill.ClientWithKey(emailKey)

	message := &mandrill.Message{}
	message.AddRecipient(destinationEmail, destinationEmail, "to")
	message.FromEmail = form.Email
	message.FromName = form.Name
	if len(form.Subject) == 0{
		form.Subject = "New contact form submission!"
	}
	message.Subject = form.Subject
	message.HTML = "<p>" + form.Message + "<p>"
	message.Text = form.Message

	responses, apiError, err := client.MessagesSend(message)
	
	if err != nil {
		panic(err)
	}
	if apiError != nil {
		fmt.Printf("\n\n\n\nAPI ERROR\n%s\n\n\n\n", apiError)
		return false
	}
	
	length := len(responses)
	for i := 0; i < length; i++ {
		if responses[i].Status == "rejected" {
			return false
		} else if responses[i].Status == "invalid" {
			return false
		}
	}
	incrementSentEmailCount(destinationEmail)
	return true
}
