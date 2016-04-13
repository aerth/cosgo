// Package mbox saves a form to a local .mbox file

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
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//

package mbox

import (
	"errors"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/goware/emailx"
	"github.com/microcosm-cc/bluemonday"
)

// Form is our email. No Attachments
type Form struct {
	Name, Email, Subject, Message string
}

// ValidationLevel should be set to something other than 1 to resolve and check email addresses
var ValidationLevel = 1

// Destination is the address where mail is "sent", its useful to change this to the address you will be replying to.
var Destination = "mbox@localhost"
var (
	Mail *log.Logger // local mbox
)

// Open sets the logger up and allows an application to customize the mailbox name. ( step 1 )
func Open(file string) error {

	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Printf("error opening file: %v", err)
		log.Fatal("Hint: touch " + file + ", or chown/chmod it so that the cosgo process can access it.")
		os.Exit(1)
	}
	Mail = log.New(f, "", 0)
	return nil
}

// PraseForm parses a url submitted query and returns a mbox.Form
func ParseForm(destination string, query url.Values) (form *Form) {
	Destination = destination
	form = parseQuery(query)
	return form
}

// ParseAndSave parses a url submitted query and sends it to Send as a mbox.Form
func ParseAndSave(destination string, query url.Values) (err error) {
	Destination = destination
	form := parseQuery(query)
	err = Save(form)
	return err
}

// Save saves an mbox file from a mbox.Form!
func Save(form *Form) error {
	t := time.Now()
	if form.Email == "@" || form.Email == " " || !strings.ContainsAny(form.Email, "@") || !strings.ContainsAny(form.Email, ".") {
		return errors.New("Blank email address.")
	}
	if ValidationLevel != 1 {
		err := emailx.Validate(form.Email)
		if err != nil {
			if err == emailx.ErrInvalidFormat {
				return errors.New("Email is not valid format.")
			}
			if err == emailx.ErrUnresolvableHost {
				return errors.New("Email is not valid format.")
			}
			return errors.New("Email is not valid format." + err.Error())
		}
	}
	// Normalize email address capitalization
	form.Email = emailx.Normalize(form.Email)
	// mbox files use two different date formats apparently.
	mailtime := t.Format("Mon Jan 2 15:04:05 2006")
	mailtime2 := t.Format("Mon, 2 Jan 2006 15:04:05 -0700")
	Mail.Println("From " + form.Email + " " + mailtime)
	Mail.Println("Return-path: <" + form.Email + ">")
	Mail.Println("Envelope-to: " + Destination)
	Mail.Println("Delivery-date: " + mailtime2)
	Mail.Println("To: " + Destination)
	Mail.Println("Subject: " + form.Subject)
	Mail.Println("From: " + form.Email)
	Mail.Println("Date: " + mailtime2)
	Mail.Println("\n" + form.Message + "\n\n")
	return nil
}

// parseQuery returns a mbox.Form from a url.Values
func parseQuery(query url.Values) *Form {
	p := bluemonday.StrictPolicy()
	form := new(Form)
	additionalFields := ""
	for k, v := range query {
		k = strings.ToLower(k)
		if k == "email" || k == "name" {
			form.Email = v[0]
			form.Email = p.Sanitize(form.Email)
		} else if k == "subject" {
			form.Subject = v[0]
			form.Subject = p.Sanitize(form.Subject)
		} else if k == "message" {
			form.Message = k + ": " + v[0] + "<br>\n"
			form.Message = p.Sanitize(form.Message)
		} else if k != "cosgo" && k != "captchaid" && k != "captchasolution" {
			additionalFields = additionalFields + k + ": " + v[0] + "<br>\n"
		}
	}
	if form.Subject == "" || form.Subject == " " {
		form.Subject = "[New Message]"
	}
	if additionalFields != "" {
		if form.Message == "" {
			form.Message = form.Message + "Message:\n<br>" + additionalFields
		} else {
			form.Message = form.Message + "\n<br>Additional:\n<br>" + additionalFields
		}
	}
	return form
}
