# Contact API Server in Golang

## The Form

Your API Endpoint is `http://{your_server}/{api-key}/send`

For historical purposes, GET is the default request method. If you wish to be RESTful, POST request methods are also supported.

Here is a sample form:

``` HTML
<form id="contact-form" action="http://{your_server}/{your_api_key_that_you_makeup_yourself}/send" method="POST">
    <input type="text" name="name" placeholder="name" required/><br/>
    <input type="text" name="email" placeholder="email" required/><br/>
    <input type="text" name="subject" placeholder="subject"/><br/>
    <input type="text" name="message" placeholder="message" required/><br/>
    <input id="contact-submit" type="submit" value="Say hello!" />
</form>
```

## Installation

This guide assumes a [Go environment](http://golang.org/doc/install) is already set up.

### Install with ```go get```

```
go get -v -u git.earthbot.net/aerth/casgo
```

### If error, get the relevant dependencies
```
go get -v -u github.com/keighl/mandrill
```

## Usage

```
export MANDRILL_KEY=123456789
export CASGO_DESTINATION=myemail@gmail.com
nohup staticcontact &
< press Ctrl+C >
tail -f debug.log

```

## Heroku App

###Get accounts for services
This requires an account to [Heroku](https://heroku.com), and an API key for [Mandrill](https://mandrillapp.com).

Visit those sites and create accounts if you do not yet have them.

### Configure environment variables
`cd` into your local staticcontact directory. The linked buildpack is for revel applications.
```
$ heroku create -b https://github.com/robfig/heroku-buildpack-go-revel.git
$ heroku config:set MANDRILL_KEY={You get this from the Mandrill dashboard}
$ git push heroku master
```

Voila! Heroku should print out the URL of the corresponding application. Now, for a given contact form, point the action method to be `{heroku_url}/{your_email}`.

### Historical Information

* Casgo is short for "Contact API server in Golang"
* Casgo is not to be confused with Costgo, the warehouse-style superstore.
* It began as a fork of https://github.com/munrocape/staticcontact
