#[Static Contact](http://www.staticcontact.com)

A form forwarder. Static Contact can replace the backend of a contact form so that you can put one on a static site and receive the email in your inbox.

##How-to

Simply change the action and method of a form to be `http://www.staticcontact.com/{your_email}` and `POST`, respectively.

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

Typically, a form submission redirects a user. The below javascript submits the form but does not redirect the user.

``` js
$(function() {
  $("#contact-form").on("submit", function(e) {
    e.preventDefault();
    $.ajax({
        url: $(this).attr("action"),
        type: 'POST',
        data: $(this).serialize(),
        success: function(data) {
            $("#contact-submit").val('Form successfully submitted!')
        }
    });
  });
});
```

## Roll your own

This guide assumes a [Go environment](http://golang.org/doc/install) is already set up.

###First, clone the repo

```
$ git clone https://github.com/munrocape/staticcontact
```

###Next, get the relevant dependencies
```
$ go get github.com/revel/revel
$ go get github.com/keighl/mandrill
$ go get gopkg.in/mgo.v2
$ go get gopkg.in/mgo.v2/bson
```

[Revel](http://revel.github.io/) is the web framework used for the applicaiton. Mandrill is the mandrill API wrapper used to send emails. Mgo is a mongo driver for Go.

###Get accounts for services
This requires an account to [Heroku](https://heroku.com), an API key for [Mandrill](https://mandrillapp.com), and a [MongoLab](https://mongolab.com) connection. 

Visit those sites and create accounts if you do not yet have them.

###Configure environment variables
`cd` into your local staticcontact directory. The linked buildpack is for revel applications.
```
$ heroku create -b https://github.com/robfig/heroku-buildpack-go-revel.git
$ heroku config:set MANDRILL_KEY={You get this from the Mandrill dashboard}
$ heroku config:set STATIC_CONTACT_MONGO_URI={Mongo lab URI, with username and password}
$ heroku config:set STATIC_CONTACT_APP_SECRET={Some random string}
$ git push heroku master
```

Voila! Heroku should print out the URL of the corresponding application. Now, for a given contact form, point the action method to be `{heroku_url}/{your_email}`.
