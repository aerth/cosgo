// This is a simple example for using the mbox library
package main

import (
	"fmt"
	"os"

	"github.com/aerth/mbox"
)

func main() {

	/*
	 *
	 * Configure and open/create mbox
	 *
	 */

	// Level 3 requires a net connection to validate email addresses hostnames
	mbox.ValidationLevel = 1

	// This is where mail gets "sent to"
	mbox.Destination = "me@localhost"

	// Choose file name
	err := mbox.Open("my.mbox")

	// Report errors
	if err != nil {
		fmt.Println(err) // "Validation errors"
		os.Exit(1)
	}
for i:= 0; i<100; i++ {
	// Build the email
	form := new(mbox.Form)
	form.Email = "root@localhost"
	form.Subject = "As seen on TV!!!"
	form.Message = "This really works!"

	err = mbox.Save(form) // Add the message to the mailbox

	// Report errors
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

	// Report success
	fmt.Println("You've got mail!\nUse 'mutt -Rf my.mbox' to read it.")

}
