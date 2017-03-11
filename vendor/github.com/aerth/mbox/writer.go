package mbox

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/goware/emailx"
)

var Writer = make(chan *Form)

func init() {
}

// Open mbox file, rw+create+append mode ( step 1 )
func Open(file string) (err error) {
	// Writer receives one email at a time
	go func() {
		var wg = make(chan int, 1)
		for {
			select {
			case form := <-Writer:
				wg <- 1
				//print("Writing email: ", form.Subject+"\n")
				// mbox files use two different date formats apparently.
				mailtime := form.Received.Format("Mon Jan 2 15:04:05 2006")
				mailtime2 := form.Received.Format("Mon, 2 Jan 2006 15:04:05 -0700")
				if form.From != "" {
					form.From = "unknown@unknown"
				}
				space := string([]byte{0x20})
				Mail.WriteString("From"+ space + form.From + space + mailtime + "\n")
				Mail.WriteString("Return-path: <" + form.From + ">" + "\n")
				Mail.WriteString("Envelope-to: " + Destination + "\n")
				Mail.WriteString("Delivery-date: " + mailtime2 + "\n")
				Mail.WriteString("To: " + Destination + "\n")
				Mail.WriteString("Subject: " + form.Subject + "\n")
				Mail.WriteString("From: " + form.From + "\n")
				Mail.WriteString("Date: " + mailtime2 + "\n")
			if form.Message != "" {
				Mail.WriteString("\n" + form.Message + "\n\n\n")
			}
			if form.Body != nil {
				Mail.WriteString("\n")
				Mail.Write(form.Body)
				Mail.WriteString("\n\n")
			}
				<- wg
			}
		}
	}()

	Mail, err = os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	return err
}

// Save sends an entire email to the writer
func Save(form *Form) error {
	form.Received = time.Now()
	if er := form.Normalize(); er != nil {
		return er
	}
	Writer <- form
	return nil
}

func (form *Form) Normalize() error {
	if ValidationLevel != 1 {
		if form.From == "@" || form.From == " " || !strings.ContainsAny(form.From, "@") {
			return errors.New("Blank email address.")
		}

		if ValidationLevel > 2 {
			err := emailx.Validate(form.From)
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
	}
	// Normalize email address capitalization
	form.From = emailx.Normalize(form.From)
	return nil
}
