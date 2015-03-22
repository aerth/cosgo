#[Static Contact](http://www.staticcontact.com)

A form forwarder. Static Contact can replace the backend of a contact form so that you can put one on a static site and receive the email in your inbox.

##How-to

Simply change the action and method of a form to be `http://www.staticcontact.com/{your_email}`.

For historical purposes, GET is the default request method. If you wish to be RESTful, POST request methods are also supported.

Here is a sample form:

``` HTML
<form id="contact-form" action="http://www.staticcontact.com/contact/{your_email}" method="POST">
    <input type="text" name="name" placeholder="name" required/><br/>
    <input type="text" name="email" placeholder="email" required/><br/>
    <input type="text" name="subject" placeholder="subject"/><br/>
    <input type="text" name="message" placeholder="message" required/><br/>
    <input id="contact-submit" type="submit" value="Say hello!" />
</form>
```

## Roll your own

This guide assumes a [Go environment](http://golang.org/doc/install) is already set up.

###First, clone the repo

```
$ git clone https://github.com/munrocape/staticcontact
```

###Next, get the relevant dependencies
```
$ go get github.com/keighl/mandrill
```

###Get accounts for services
This requires an account to [Heroku](https://heroku.com), and an API key for [Mandrill](https://mandrillapp.com).

Visit those sites and create accounts if you do not yet have them.

###Configure environment variables
`cd` into your local staticcontact directory. The linked buildpack is for revel applications.
```
$ heroku create -b https://github.com/robfig/heroku-buildpack-go-revel.git
$ heroku config:set MANDRILL_KEY={You get this from the Mandrill dashboard}
$ git push heroku master
```

Voila! Heroku should print out the URL of the corresponding application. Now, for a given contact form, point the action method to be `{heroku_url}/{your_email}`.
