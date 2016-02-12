# Contact API Server in Golang

## Installation

```
go get -v -u github.com/aerth/casgo

```

If error, get the relevant dependencies

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

Repeat for each virtual host. nginx server block coming soon.


# Future

* Option to use different mail handler (not only mandrill)

* .config File

* templates/ folder




# Other




## Heroku App (not tested with Casgo yet)

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
