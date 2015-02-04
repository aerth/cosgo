package controllers

import (
	"fmt"
	"math/rand"
	//mandrill "github.com/keighl/mandrill"
	"time"
	"github.com/revel/revel"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

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

func (c Contact) Contact() revel.Result {
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

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var tokenLength = 8

func generateToken() string {
	b := make([]rune, tokenLength)
	for i := range b {
	    b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func createTokenRecord(email string) bool {
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
									 Email:email, 
									 MonthlyCount: 0, 
									 CurrentMonth: int(month)}
	
	// Add the user to the DB
	mongoUri, mongoFound := revel.Config.String("mongoLabUri")
	if !mongoFound {
		panic("Mongo URI not set in app.conf.")
	}
	if len(mongoUri) == 0 {
		panic("Mongo URI is invalid with length 0")
	}

	session, err := mgo.Dial(mongoUri)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	c := session.DB("static-contact").C("static-contact-tokens")
	err = c.Insert(&newRecipient)
	if err != nil {
	    fmt.Printf("%s\n", err)
	}
	fmt.Printf("%s %s %d %d\n", newRecipient.Token, newRecipient.Email, newRecipient.MonthlyCount, newRecipient.CurrentMonth)
	return true
}

func tokenExistsInDB(token string) bool {
	mongoUri, mongoFound := revel.Config.String("mongoLabUri")
	
	if !mongoFound {
		panic("Mongo URI not set in app.conf.")
	}
	if len(mongoUri) == 0 {
		panic("Mongo URI is invalid with length 0")
	}
	session, err := mgo.Dial(mongoUri)
	if err != nil {
		panic(err)
	}
	
	defer session.Close()
	c := session.DB("static-contact").C("static-contact-tokens")
	
	result := recipientRecord{}
	
	err = c.Find(bson.M{"Token": token}).One(&result)
	
	if err != nil {
		// The token does not exist
		return false
	}
	return true
}

func incrementSentEmailCount(token string) bool {
	
	mongoUri, mongoFound := revel.Config.String("mongoLabUri")
	
	if !mongoFound {
		panic("Mongo URI not set in app.conf.")
	}
	if len(mongoUri) == 0 {
		panic("Mongo URI is invalid with length 0")
	}
	session, err := mgo.Dial(mongoUri)
	if err != nil {
		panic(err)
	}
	
	defer session.Close()
	c := session.DB("static-contact").C("static-contact-tokens")
	
	oldRecord := recipientRecord{}
	
	err = c.Find(bson.M{"token": token}).One(&oldRecord)
	
	newRecord := oldRecord
	newRecord.MonthlyCount += 1
	err = c.Update(bson.M{"token": token}, &newRecord)

	return true
}

func sendEmail(token string, form Form) bool {
	
	mongoUri, mongoFound := revel.Config.String("mongoLabUri")
	if !mongoFound {
		panic("Mongo URI not set in app.conf.")
	}
	if len(mongoUri) == 0 {
		panic("Mongo URI is invalid with length 0")
	}
	session, err := mgo.Dial(mongoUri)
	if err != nil {
		panic(err)
	}

	defer session.Close()
	c := session.DB("static-contact").C("static-contact-tokens")

	result := recipientRecord{}
	
	err = c.Find(bson.M{"token": token}).One(&result)
	if err != nil {
		// The token does not exist
		fmt.Printf("The token does not exist!! %s\n", token)
		return false
	}

	// destinationEmail := result.Email

	// mandrillKey, found := revel.Config.String("mandrillKey")
	// if !found {
	// 	panic("Mandrill API key not set in app.conf.")
	// }
	// if len(mandrillKey) == 0 {
	// 	panic("Mandrill API key is empty.")
	// }
	
	// client := mandrill.ClientWithKey(mandrillKey)

	// message := &mandrill.Message{}
	// message.AddRecipient(destinationEmail, destinationEmail, "to")
	// message.FromEmail = form.Email
	// message.FromName = form.Name
	// if len(form.Subject) == 0{
	// 	form.Subject = "New contact form submission!"
	// }
	// message.Subject = form.Subject
	// message.HTML = "<p>" + form.Message + "<p>"
	// message.Text = form.Message

	// responses, apiError, err := client.MessagesSend(message)
	
	// if err != nil {
	// 	panic(err)
	// }
	// if apiError != nil {
	// 	fmt.Printf("\n\n\n\nAPI ERROR\n%s\n\n\n\n", apiError)
	// 	return false
	// }
	
	// length := len(responses)
	// for i := 0; i < length; i++ {
	// 	if responses[i].Status == "rejected" {
	// 		return false
	// 	} else if responses[i].Status == "invalid" {
	// 		return false
	// 	}
	// }
	// increment record
	incrementSentEmailCount(token)
	return true
}
