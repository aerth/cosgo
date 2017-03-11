package main

import (
	"net/http"
	"time"

	"github.com/aerth/mbox"
)

func main() {
	server := new(http.Server)
	server.Addr = ":8080"
	server.Handler = http.HandlerFunc(handler)
	server.ListenAndServe()
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		handleForm(w, r)
		return
	}
	w.Write(form)
}

func handleForm(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.Write([]byte(err.Error()))
	}
	name, email := r.FormValue("name"), r.FormValue("email")
	subject, message := r.FormValue("subject"), r.FormValue("message")

		mbox.Open("my.mbox")

	var form mbox.Form
	form.Name = name
	form.Email = email
	form.Subject = subject
	form.Message = message
	err = mbox.Save(&form)
	if err != nil {
		w.Write([]byte(err.Error()+"\n"))
	} else {
		print("Message Received ", time.Now().String()+"\n")
    http.Redirect(w, r, "/?sent", http.StatusFound)
	}

}

var form = []byte(`<html>
  <form method="POST">
    Your Name: <input name="name"><br>
    Your Email: <input name="email"><br>
    Subject: <input name="subject"><br>
    Message: <input name="message"><br><br>
    <input type="submit" value="send mail">
  </form>
  </html>

  `)
