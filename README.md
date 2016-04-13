## cosgo
when all you needed was a contact form, anyways.

* Mailing: Contact form saves messages to a local mbox file, or sends mail with Sendgrid for free. Mandrill option.
* Customize: Uses Go style templates, `templates/index.html` and `templates/error.html`
* Static Files: /sitemap.xml /favicon.ico, /robots.txt. Limited to .woff, .ttf, .css, .js .png .jpg
* You can stuff .zip, .tgz. .tar, .tar.gz files in /files/ and cosgo will serve them, too.
* Tested on NetBSD and Debian servers, even runs on Windows. Probably runs great on anything else, too.
* This needs some testing! If you use cosgo, I would love to hear from you! https://isupon.us

```

The MIT License (MIT)

Copyright (c) 2016 aerth

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

```

## Usage

```
cosgo # serve contact form in local mbox mode
cosgo -noredirect # show 404 pages instead of redirecting to /
cosgo -debug # don't switch to cosgo.log
cosgo -static=false # don't serve /static, /files or /page
cosgo -custom "cosgo3" -port 8080 -bind 192.16.1.10

```




## Installation

Using config file (new)

```
export GOPATH=/tmp/cosgo
go get -v -u github.com/aerth/cosgo
cd $GOPATH/src/github.com/aerth/cosgo
go build
./cosgo -h # list flags
./cosgo -config # Creates the config
./cosgo -config -debug &
firefox http://127.0.0.1:8080 # Check it out.

This curl request hould return an error, because we don't accept POST without the CSRF token:
curl --data /tmp/cosgo.gif http://127.0.0.1:8080/upload

```

## Theme

Customize the theme by modifying ./templates/index.html and ./templates/error.html

Included is the bare minimal.

Upload your static .css .js .png .jpeg .woff and .tff files in ./static/ like /static/{foldername}/whatever.css, they will be available like any normal server.

Modify your theme if it looks in /assets/.

cp favicon into static too. it well be served as if it were located at /favicon.ico
cp robot.txt into static too. it well be served at /robots.txt

### Template Rules:

Make sure your template starts with `{{define "Index"}}`
Make sure your template ends with `{{end}}`


## Upgrading

To update the running instances I have been running something like this:

```
#!/bin/sh

# since i run this with sudo -u cosgo, and cosgo runs as unprivileged demon user...
# you have to modify this! Keep the $USER variables here if you place it in a @reboot cron.

# I use -9 to make sure that thing is dead
# And since multiple cosgo instances are running, I don't want to kill the wrong user's process.# Run it "quiet" so cron doesn't complain. cosgo.log will still be updated.
# If you are using docker or something and you want it all in stdout, just use -debug
# -insecure is no longer needed or supported. it is now default.
# Check the new flags before deploying

pkill -9 -u $USER cosgo || true; /demon/$USER/cosgo -quiet -port 8080 -bind 0.0.0.0

```

```cron

COSGO_DESTINATION=your@email.com

#COSGO_API_KEY=contact

COSGOPORT=8080

@reboot /demon/$USER/cosgo -fastcgi -port $COSGOPORT> /dev/null 2>&1

```

## Sample Nginx Config

```nginx
server {
        server_name default_address;
        listen 80;

        location / {

        proxy_pass http://127.0.0.1:8080; # Change to your static site's port

        }

}

```

or with fastcgi:

```nginx
server {
        server_name default_address;
        listen 80;

        location / {

        fastcgi_pass 127.0.0.1:8080; # no http:// with fastcgi
        include $fastcgi_params;
        }

}

```

More code examples at https://github.com/aerth/cosgo/wiki

Please add your use case. It could make cosgo better. Report any bugs or feature requests directly at https://isupon.us ( running casgo with a pixelarity theme )

# Historical Information

* Casgo is short for "Contact API server in Golang"
* Turns out theres another casgo. Now this is cosgo.
* Cosgo is still not to be confused with Costco, the warehouse-style superstore.
* It began as a fork of https://github.com/munrocape/staticcontact
