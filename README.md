# cosgo

Copyright (c) 2016 aerth

### Contact on Static in Golang
=======

###############################################################################
###############################################################################
### Contact on Static in Golang
or...
### Contact API Server in Golang
###############################################################################
###############################################################################

[![Build Status](https://travis-ci.org/aerth/cosgo.svg?branch=master)](https://travis-ci.org/aerth/cosgo)

* Run it 100 times for 100 different one page web sites on your server.
* Keep it on @reboot and every couple minutes cronjob.
* Contact form sends mail with mandrill for free.
* Enjoy Cosgo. Ongoing testing on NetBSD and Debian servers.
* Work in progress.

## 

```


The MIT License (MIT)

Copyright (c) 2016 aerth

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

```



## Installation / Updating

Please do this often. If you have recent templates (v0.4) only the binary need update. 
```
go get -v -u github.com/aerth/casgo

```
## Usage

```shell
cd $GOPATH/src/github/aerth/casgo
go build # we need templates and static directory for now
export MANDRILL_KEY=123456789
export CASGO_DESTINATION=changeme@gmail.com
./casgo -debug -mailbox
//or with nginx:
export MANDRILL_KEY=123456789
export CASGO_API_KEY=contact # for easy /contact/form
export CASGO_DESTINATION=myemail@gmail.com
nohup casgo -fastcgi -insecure -port 6060 &
< press Ctrl+C >
tail -f debug.log
// or while testing
MANDRILL_KEY=134 CASGO_DESTINATION=1345 CASGO_API_KEY=contact casgo -debug -insecure -mailbox
```
########



```
$ casgo -h
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
## Be sure to copy the templates and static folders to where your binary will exist.
##################
## Sample Cron
This cron right here, it changes the form action= to whatever the key is. Every 15 minute. More minutes may be better, in case a visitor just stays on the page for a minute or two before sending the message.

```
MANDRILL_KEY=yourK3YgoesH3R3
CASGO_DESTINATION=your@email.com
#CASGO_API_KEY=contact

*/15 * * * * /usr/bin/pkill casgo;/demon/casgo/casgo -insecure -fastcgi -port 8080 > /dev/null 2>&1

20 4 * * * /usr/bin/pkill casgo;/demon/casgo/casgo -insecure -fastcgi -port 8080 > /dev/null 2>&1

@reboot /demon/casgo/casgo -insecure -fastcgi -port 8080 > /dev/null 2>&1

```



## Sample Nginx Config
For use when setting CASGO_API_KEY=contact, and using a different system for serving pages.


```nginx
server {
        server_name default_address;
        listen 80;

        location / {

        proxy_pass http://127.0.0.1:8081; # Change to your static site's port

        }
        location /contact/form/send {
        proxy_pass http://127.0.0.1:8080; # Change using "casgo -port XXX"
        }
}

```

or with fastcgi:


```nginx
server {
        server_name default_address;
        listen 80;

        location / {

        fastcgi_pass 127.0.0.1:8080;
        include $fastcgi_params;
        }

}

```

# Historical Information

* Casgo is short for "Contact API server in Golang"
* Turns out theres another casgo. Now this is cosgo.
* Cosgo is still not to be confused with Costco, the warehouse-style superstore.
* It began as a fork of https://github.com/munrocape/staticcontact
