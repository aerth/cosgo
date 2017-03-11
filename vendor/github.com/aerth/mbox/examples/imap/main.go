package main

import (
	"bytes"
	"fmt"
	"net/mail"
	"os"
	"time"

	"github.com/aerth/mbox"
	"github.com/xarg/imap"
)

func main() {
	var fetchAll bool
	if os.Getenv("IMAP_ALL") != "" {
		fetchAll = true
	}
	var c *imap.Client
	var cmd *imap.Command
	var rsp *imap.Response
	var err error

	// Connect to the server
	c, err = imap.DialTLS(os.Getenv("IMAP_HOST"), nil)

	if c == nil || err != nil {
		fmt.Println("Error: can't connect.")
		if err != nil {
			fmt.Println(err)
		}
		return
	}
	// Remember to log out and close the connection when finished
	defer c.Logout(30 * time.Second)

	// Print server greeting (first response in the unilateral server data queue)
	fmt.Println("Server says hello:", c.Data[0].Info)
	c.Data = nil

	// Enable encryption, if supported by the server
	if c.Caps["STARTTLS"] {
		c.StartTLS(nil)
	}

	var user, pass = os.Getenv("IMAP_USER"), os.Getenv("IMAP_PASS")

	if user == "" || pass == "" {
		fmt.Println("Use IMAP_USER, IMAP_PASS, and IMAP_HOST ( optional IMAP_POST ) variables")
		return
	}
	// Authenticate
	if c.State() == imap.Login {
		c.Login(user, pass)
	}

	// List all top-level mailboxes, wait for the command to finish
	cmd, err = imap.Wait(c.List("", "%"))
	if err != nil {
		fmt.Println(err)
		return
	}

	// Open a mailbox (synchronous command - no need for imap.Wait)
	c.Select("INBOX", true)
	fmt.Println("\nMailbox status:\n", c.Mailbox)

	fmt.Println("Saving messages to mbox file: 'imap.mbox'")
	// fetch all messages
	var set *imap.SeqSet
	if fetchAll {
		set, _ = imap.NewSeqSet("1:*")
	} else {
		set, _ = imap.NewSeqSet("1:5")
	}
	// if c.Mailbox.Messages >= 10 {
	// 	set.AddRange(c.Mailbox.Messages-9, c.Mailbox.Messages)
	// } else {
	// 	set.Add("1:*")
	// }
	// cmd, err = c.Fetch(set, "RFC822.HEADER")
	// if err != nil {
	// 	fmt.Println("Error processing:", err)
	// 	return
	// }

	cmd, err = c.Fetch(set, "RFC822.HEADER", "RFC822.TEXT")
	if err != nil {
		fmt.Println(err)
		return
	}

	var i int = 1
	for cmd.InProgress() {
		c.Recv(-1)
		mbox.Open("imap.mbox")
		for _, rsp = range cmd.Data {
			header := imap.AsBytes(rsp.MessageInfo().Attrs["RFC822.HEADER"])
			if msg, _ := mail.ReadMessage(bytes.NewReader(header)); msg != nil {
				body := imap.AsBytes(rsp.MessageInfo().Attrs["RFC822.TEXT"])

				if err != nil {
					fmt.Println(err)
					return
				}

				if len(body) > 0 {
					var form mbox.Form
					form.From = msg.Header.Get("Return-path")
					form.Subject = msg.Header.Get("Subject")
					if form.Subject == "" { form.Subject = "<no subject>" }
					form.Message = string(body)
					mbox.Save(&form)
					fmt.Printf("Message #%v saved to mbox\n", i)
					i++
				}
			}
		}
		cmd.Data = nil

		// Process unilateral server data
		for _, rsp = range c.Data {
			//fmt.Println("Server data:", rsp)
		}
		c.Data = nil
	}

	// Check command completion status
	if rsp, err := cmd.Result(imap.OK); err != nil {
		if err == imap.ErrAborted {
			fmt.Println("Fetch command aborted")
		} else {
			fmt.Println("Fetch error:", rsp.Info)
		}
	}
}
