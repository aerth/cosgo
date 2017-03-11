package mbox

import "testing"
import "strconv"

func TestMboxConcurrency(t *testing.T) {
	// Choose file name
	err := Open("test.mbox")

	// Report errors
	if err != nil {
		t.Log(err)
	}
	for i := 0; i <= 128; i++ {
		go func() { // Build the email
			var form Form
			form.Email = "aerth@localhost"
			form.Subject = "As seen on TV!!! "+strconv.Itoa(i)
			form.Message = "This really works! "+strconv.Itoa(i)
			err = Save(&form) // Add the message to the mailbox
			// Report errors
			if err != nil {
				t.Log(err)
			}
		}()
	}
}

func dummymail() {

}
