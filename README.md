# Contact and Static in Golang
or
# Contact API Server in Golang

Copyright (c) 2016 aerth

## Installation / Updating

```
go get -v -u github.com/aerth/casgo

```
## Usage

```shell
export MANDRILL_KEY=123456789
export CASGO_DESTINATION=myemail@gmail.com
casgo -debug

```

or with nginx:

```shell
export MANDRILL_KEY=123456789
export CASGO_API_KEY=contact # for easy /contact/form
export CASGO_DESTINATION=myemail@gmail.com
nohup casgo -fastcgi -insecure -port 6060 &
< press Ctrl+C >
tail -f debug.log

```
or while testing

```
MANDRILL_KEY=134 CASGO_DESTINATION=1345 CASGO_API_KEY=contact casgo -debug -insecure

```
########


## Usage of casgo:
```
  -bind string
    	default: 127.0.0.1 (default "127.0.0.1")
  -debug
    	be verbose, dont switch to logfile
  -fastcgi
    	use fastcgi
  -insecure
    	accept insecure cookie transfer
  -mailbox
    	save messages to an local mbox file
  -port string
    	HTTP Port to listen on (default "8080")
  -static (default on)
      serve static files from ./static directory
```


###################

### Be sure to copy the templates and static folders to whereever your binary will exit.

## Sample Cron

This right here, it changes the form action= to whatever the key is. Every 1 minute. 5 minutes may be better, in case a visitor just stays on the page for a minute or two before sending the message.

```
MANDRILL_KEY=yourK3YgoesH3R3
CASGO_DESTINATION=your@email.com
#CASGO_API_KEY=contact

* * * * * /usr/bin/pkill casgo;/demon/casgo/casgo -insecure -fastcgi -port 6099 > /dev/null 2>&1

20 4 * * * /usr/bin/pkill casgo;/demon/casgo/casgo -insecure -fastcgi -port 6099 > /dev/null 2>&1

@reboot /demon/casgo/casgo -insecure -fastcgi -port 6099 > /dev/null 2>&1

```



## Sample Nginx Config
For use when setting CASGO_API_KEY=contact, and using a different system for serving pages.


```nginx
server {
        server_name default_address;
        listen 80;

        location / {

        proxy_pass http://127.0.0.1:6098; # Change to your static site's port

        }
        location /contact/form/send {
        proxy_pass http://127.0.0.1:6099; # Change using "casgo -port XXX"
        }
}

```

or with fastcgi:


```nginx
server {
        server_name default_address;
        listen 80;

        location / {

        fastcgi_pass 127.0.0.1:6099;
        include $fastcgi_params;
        }

}

```
# Future

* Option to use different mail handler (not only mandrill)

* *remove flags and use env variables!*

* -mailbox flag disables mandrill and uses a local casgo.mbox file

* better 404 handling on /static/

* option for log level 0-5

* Pull requests from strangers are cool!


# Historical Information

* Casgo is short for "Contact API server in Golang"
* Casgo is not to be confused with Costco, the warehouse-style superstore.
* It began as a fork of https://github.com/munrocape/staticcontact
