## cosgo
when all you needed was a contact form, anyways.


* Contact form saves messages to a local mbox file, or sends mail with Sendgrid for free. Mandrill option.
* Uses Go style templates, `templates/index.html` and `templates/error.html`
* ...while serving /static/* files, /favicon.ico, /robots.txt.
* You can stuff things in /files/ and cosgo will serve them, too.
* Tested on NetBSD and Debian servers, even runs on Windows. Probably runs great on anything else, too.
* Contact the author using **cosgo** right now. https://isupon.us

```

The MIT License (MIT)

Copyright (c) 2016 aerth

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

```

## Installation

Using config file (new)

```
export GOPATH=/tmp/cosgo
go get -v -u github.com/aerth/cosgo
cd $GOPATH/src/github.com/aerth/cosgo
go build
./cosgo -h
./cosgo -config # Creates the config
./cosgo -config -debug &
firefox http://127.0.0.1:8080 # Check it out.

Should return error, because we don't accept POST without the CSRF token:
curl --data /tmp/cosgo.gif http://127.0.0.1:8080/upload

```

## Theme

Customize the theme by modifying ./templates/index.html and ./templates/error.html

Included is the bare minimal.

Upload your static .css .js .png .jpeg .woff and .tff files in ./static/ like /static/{foldername}/whatever.css, they will be available like any normal server.

Modify your theme if it looks in /assets/.

cp favicon into static too. it well be served at /favicon.ico
cp robot.txt into static too. it well be served at /robots.txt

Make sure your template starts with `{{define "Index"}}`
Make sure your template ends with `{{end}}`


## Upgrading

Open up script/upgrade in $EDITOR
Modify to suit your needs

```
$EDITOR script/*

mv script/launch /usr/local/bin/cosgo-launch
mv script/upgrade /usr/local/bin/cosgo-upgrade
```

## Cron


```cron

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
Please add your use case. It could make cosgo better.

# Historical Information

* Casgo is short for "Contact API server in Golang"
* Turns out theres another casgo. Now this is cosgo.
* Cosgo is still not to be confused with Costco, the warehouse-style superstore.
* It began as a fork of https://github.com/munrocape/staticcontact
