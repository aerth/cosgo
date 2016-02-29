## cosgo
when all you needed was a contact form, anyways.


* Contact form saves messages to a local mbox file, or sends mail with mandrill for free.
* Uses `templates/index.html` and `templates/error.html` while serving `static/` files
* Tested on NetBSD and Debian servers.
* Try Cosgo right now. https://isupon.us

[![Build Status](https://travis-ci.org/aerth/cosgo.svg?branch=master)](https://travis-ci.org/aerth/cosgo) 


```

The MIT License (MIT)

Copyright (c) 2016 aerth

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

```

## Installation

```

go get -v -u github.com/aerth/cosgo
cd $GOPATH/src/github.com/aerth/cosgo
COSGO_DESTINATION="test@localhost" ./cosgo -debug -mailbox &
firefox http://127.0.0.1:8080
step 4.. ?
step 5.. PROFIT!

```

## Theme

Customize the theme by modifying ./templates/index.html and ./templates/error.html

Upload your static .css .js .png .jpeg .woff and .tff files in ./static/, they will be available like any normal server.

Upload to github and link from wiki!


## Upgrading

Open up script/upgrade in $EDITOR
Modify to suit your needs
Run often. Should be only one moment of downtime if using fastcgi..
```
$EDITOR script/*

mv script/launch /usr/local/bin/cosgo-launch
mv script/upgrade /usr/local/bin/cosgo-upgrade
```

## Usage

This is a sample launch script to get started. It doesn't send emails, but can be used to work on templates.

```shell
cosgo -h
MANDRILL_KEY=134 COSGO_DESTINATION=1345 COSGO_API_KEY=contact cosgo -debug -insecure -mailbox
```

###################
## Be sure to copy the templates and static folders to where your binary will exist.
##################
## Sample Cron


```
MANDRILL_KEY=yourK3YgoesH3R3
COSGO_DESTINATION=your@email.com
#COSGO_API_KEY=contact
COSGOPORT=8080
*/30 * * * * /usr/bin/pkill -u $USER cosgo;/demon/$USER/cosgo -insecure -fastcgi -port $COSGOPORT > /dev/null 2>&1
20 4 * * * /usr/bin/pkill -u $USER cosgo;/demon/$USER/cosgo -insecure -fastcgi -port $COSGOPORT > /dev/null 2>&1
@reboot /demon/$USER/cosgo -insecure -fastcgi -port $COSGOPORT> /dev/null 2>&1

```

## Sample Nginx Config

```nginx
server {
        server_name default_address;
        listen 80;

        location / {

        proxy_pass http://127.0.0.1:8081; # Change to your static site's port

        }
        location /contact/form/send {
        proxy_pass http://127.0.0.1:8080; # Change using "cosgo -port XXX"
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
More code examples at https://github.com/aerth/cosgo/wiki
Add your use case

# Historical Information

* Casgo is short for "Contact API server in Golang"
* Turns out theres another casgo. Now this is cosgo.
* Cosgo is still not to be confused with Costco, the warehouse-style superstore.
* It began as a fork of https://github.com/munrocape/staticcontact
