## cosgo
when all you needed was a contact form, anyways.

[![Go Report Card](https://goreportcard.com/badge/github.com/aerth/cosgo)](https://goreportcard.com/report/github.com/aerth/cosgo)

* Save the mail: Contact form saves messages to a local .mbox file
* Read the mail with something like: `mutt -Rf cosgo.mbox`
* Option to send mail using Sendgrid
* Option to use GPG for encrypting the messages
* Customize: Uses Go style templates, `templates/index.html` and `templates/error.html`
* Static Files: /sitemap.xml /favicon.ico, /robots.txt. Limited to .woff, .ttf, .css, .js .png .jpg and custom -ext flag
* You can stuff .zip, .tgz. .tar, .tar.gz files in /files/ and cosgo will serve them, too.
* Easy to convert static web sites to cosgo template.

```

The MIT License (MIT)

Copyright (c) 2016 aerth

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

```

## Usage Examples

```
cosgo -h # show flags
cosgo -debug # more messages
cosgo -nolog -quiet # no output whatsoever
cosgo -bind 127.0.0.1 -port 8000 # nginx or hidden service setup (bind only to localhost)
cosgo -bind 0.0.0.0 -port 8000 # listen and serve (bind to all interfaces)
cosgo -gpg publickey.asc -debug # loads publickey.asc into memory (so you can delete it after booting cosgo if you want)

```

See newest version in action at https://cosgo.herokuapp.com/

-------

## Installation

Releases:

1. Extract the archive.
2. Copy the binary to /usr/local/bin/ ( or extract everything to C:/cosgo/ )
3. Run `cosgo` from the command line. This will listen on all interfaces, port 8080.
4. Edit templates, static files

Building from source:

```
# Grab latest source code ( also check https://isupon.us )
git clone https://github.com/aerth/cosgo.git && cd cosgo
# make deps # optional, cosgo comes with vendor directory.

# Build a static binary
CGO_ENABLED=0 make

# Install to /usr/local/bin/cosgo
sudo make install

```

-------

## Theme

Customize the theme by modifying ./templates/index.html and ./templates/error.html
Included in the binary is the bare minimal. o
If they don't exist, the templates and static directories will be created where you run cosgo from. (since v0.9)
Upload your static .css .js .png .jpeg .woff and .ttf files in ./static/ like /static/{foldername}/whatever.css, they will be available like any typical static file server. 
Some web design themes look in /assets, which won't be served by cosgo. `s/assets/static/` 

cp favicon into static/ too. it will be served as if it were located at /favicon.ico

cp robot.txt into static/ too. it will be served at /robots.txt

-------

### Template Rules:

  * Make sure your template starts with `{{define "Index"}}`

  * Make sure your template ends with `{{end}}`

  * And is named index.html

  * Your form needs {{ .csrfField }} and `action="/{{.Key}}/send" method="POST"`
  * The minimal captcha stuff would be:

`
    <p><img id="image" src="/captcha/{{.CaptchaId}}.png" alt="Captcha image"></p>
 ` ,

 `  <input type=hidden name=captchaId value="{{.CaptchaId}}" />`

  and something like:

`
 <input name=captchaSolution placeholder="Enter the code to proceed" />
`

All your static files should be neatly organized in `/static/css/*.css` to be used with the templates.


### Minimal Error template:

error.html:
`{{define "Error"}}
Houston: We have a {{.err}} problem.
{{end}}`

-------

## cron
```cron

COSGO_DESTINATION=your@email.com

# This next line turns cosgo into a normal contact form.
#COSGO_API_KEY=contact

COSGOPORT=8080

@reboot /demon/$USER/cosgo -fastcgi -port $COSGOPORT> /dev/null 2>&1

```

-------


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
-------

# More information at Wiki

More code examples at https://github.com/aerth/cosgo/wiki

Please add your use case. It could make cosgo better. Report any bugs or feature requests directly at https://isupon.us ( running casgo with a pixelarity theme )

-------

### Historical Information

* Casgo is short for "Contact API server in Golang"
* Turns out theres another casgo. Now this is cosgo.
* Cosgo is still not to be confused with Costco, the warehouse-style superstore.
* It began as a fork of https://github.com/munrocape/staticcontact
