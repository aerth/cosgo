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
	"io/ioutil"
	"log"
	"net/url"
	"strings"

	"golang.org/x/crypto/openpgp"       // pgp support
	"golang.org/x/crypto/openpgp/armor" // armorized public key

	//"github.com/goware/emailx"           // email validation
	"github.com/microcosm-cc/bluemonday" // input sanitizaation
)

// ParseFormGPG parses a url submitted query and returns a mbox.Form
func ParseFormGPG(destination string, query url.Values, publicKey []byte) (form *Form) {
	Destination = destination
	form = ParseQueryGPG(query, publicKey)
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
			form.From = v[0]
			form.From = p.Sanitize(form.From)
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
