// This is a simple example for using the mbox library
package main

import (
	"fmt"
	"os"

	"github.com/aerth/cosgo/mbox"
)

var err error

func main() {

	// Configure and open/create mbox
	mbox.ValidationLevel = 1            // Level 2 requires a net connection to validate email addresses
	mbox.Destination = "me@example.com" // This is where mail gets "sent to"
	err = mbox.Open("my.mbox")          // Choose file name

	// Report errors
	if err != nil {
		fmt.Println(err) // "Validation errors"
		os.Exit(1)
	}

	// Build the email
	form := new(mbox.Form)
	form.Email = "from@somebody.com"
	form.Subject = "As seen on TV!!!"
	form.Message = "This really works!"
	err = mbox.Save(form) // Add the message to the mailbox

	// Report errors
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Report success
	fmt.Println("You've got mail!")
}
