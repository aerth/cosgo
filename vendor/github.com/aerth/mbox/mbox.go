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

// Package mbox saves a form to a local .mbox file (opengpg option)
/*

Usage of mbox library is as follows:

Define mbox.Destination variable in your program

Accept an email, populate the mbox.Form struct like this:
	mbox.From = "joe"
	mbox.Email = "joe@blowtorches.info
	mbox.Message = "hello world"
	mbox.Subject = "re: hello joe"
	mbox.Save()


*/
package mbox

import (
	"bytes"
	"errors"
//	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/openpgp"       // pgp support
	"golang.org/x/crypto/openpgp/armor" // armorized public key

	"github.com/goware/emailx"           // email validation
	"github.com/microcosm-cc/bluemonday" // input sanitizaation
)

// Form is our email. No Attachments.
type Form struct {
	Name, Email, Subject, Message string
}

var (
	// ValidationLevel should be set to something other than 1 to resolve hostnames and validate emails
	ValidationLevel = 1
	// Destination is the address where mail is "sent", its useful to change this to the address you will be replying to.
	Destination = "mbox@localhost"

	// Mail is the local mbox, implemented as a logger
	Mail *log.Logger
)

// Open sets the logger up and allows an application to customize the mailbox name. ( step 1 )
func Open(file string) error {

	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	Mail = log.New(f, "", 0)
	return nil
}

// ParseFormGPG parses a url submitted query and returns a mbox.Form
func ParseFormGPG(destination string, query url.Values, publicKey []byte) (form *Form) {
	Destination = destination
	form = ParseQueryGPG(query, publicKey)
	return form
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

// ParseQuery returns a mbox.Form from url.Values
func ParseQuery(query url.Values) *Form {
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
			form.Message = form.Message + "Message:\n<br>" + p.Sanitize(additionalFields)
		} else {
			form.Message = form.Message + "\n<br>Additional:\n<br>" + p.Sanitize(additionalFields)
		}
	}

	return form
}

// ParseQueryGPG returns a mbox.Form from a url.Values but encodes the form.Message if publicKey is not nil
func ParseQueryGPG(query url.Values, publicKey []byte) *Form {
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
			form.Message = form.Message + "Message:\n<br>" + p.Sanitize(additionalFields)
		} else {
			form.Message = form.Message + "\n<br>Additional:\n<br>" + p.Sanitize(additionalFields)
		}
	}

	if publicKey != nil {
		tmpmsg, err := PGPEncode(form.Message, publicKey)
		if err != nil {
			log.Println("gpg error.")
			log.Println(err)
		} else {

			form.Message = tmpmsg
		}

	}
	return form
}

// rel2real Relative to Real path name
func rel2real(file string) (realpath string) {
	pathdir, _ := path.Split(file)

	if pathdir == "" {
		realpath, _ = filepath.Abs(file)
	} else {
		realpath = file
	}
	return realpath
}

// PGPEncode handles the actual encrypting of the message. Outputs ascii armored gpg message or an error.
func PGPEncode(plain string, publicKey []byte) (encStr string, err error) {

	entitylist, err := openpgp.ReadArmoredKeyRing(bytes.NewBuffer(publicKey))
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)

	abuf, err := armor.Encode(buf, "PGP MESSAGE", map[string]string{
		"Version": "OpenPGP",
	})
	if err != nil {
		return "", err
	}
	w, err := openpgp.Encrypt(abuf, entitylist, nil, nil, nil)
	defer w.Close()
	if err != nil {
		return "", err
	}
	defer w.Close()
	_, err = w.Write([]byte(plain))
	if err != nil {
		return "", err
	}

	err = w.Close()

	if err != nil {
		return "", err
	}
	abuf.Close()

	bytes, err := ioutil.ReadAll(buf)

	encStr = string(bytes)

	return encStr, nil
}
