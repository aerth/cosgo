## cosgo
when all you needed was a contact form, anyways.

[![Go Report Card](https://goreportcard.com/badge/github.com/aerth/cosgo)](https://goreportcard.com/report/github.com/aerth/cosgo)

* Save the mail: Contact form saves messages to a local .mbox file
* Read the mail with something like: `mutt -Rf cosgo.mbox`
* Option to send mail using Sendgrid
* Option to use GPG for encrypting the messages
* Customize: Uses two Go style templates, `templates/index.html` and `templates/error.html`
* Static Files: /sitemap.xml /favicon.ico, /robots.txt. Extensions limited to .woff, .ttf, .css, .js .png .jpg and custom -ext flag
* Easy to convert static web sites to cosgo theme.
* No database, from no log to (very verbose) debug log
* No dependencies. Runs 'out of the box'

```

The MIT License (MIT)

Copyright (c) 2016 aerth

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

```

## Usage Examples

	First, cd into a new empty directory such as $HOME/cosgo, 
	this will allow you to easily see which files cosgo creates on boot.

```
cosgo -h # show flags
cosgo # normal boot and serve
cosgo -debug # more verbosity
cosgo -nolog -quiet # no output whatsoever
cosgo -bind 127.0.0.1 -port 8000 # nginx or hidden service setup (bind only to localhost)
cosgo -bind 0.0.0.0 -port 8000 # listen and serve (bind to all interfaces)
cosgo -gpg publickey.asc # loads publickey.asc into memory

```

See newest version in action! It is automatically deployed at https://cosgo.herokuapp.com/

-------

## Installation

Releases:

1. Extract the archive.
2. Copy the binary to /usr/local/bin/ ( or extract everything to C:/cosgo/ )
3. Run `cosgo` from the command line. This will listen on all interfaces, port 8080 and create static/templates directories.
4. Edit templates, static files

Building from source:

```
# Grab latest source code ( also check https://isupon.us for more info )
git clone https://github.com/aerth/cosgo.git && cd cosgo
# make deps # optional, cosgo comes with vendor directory.
# Build a static binary
make

# Install to /usr/local/bin/cosgo
sudo make install

```

-------

## Theme

Customize the theme by modifying ./templates/index.html and ./templates/error.html
Included in the binary is the *bare minimal* and is not intended to look very fancy.
If they don't exist, the templates and static directories will be created where you run cosgo from. (since v0.9)
You can upload your static .css .js .png .jpeg .woff and .ttf files in ./static/foldername/file
You can include them in your templates as /static/{foldername}/whatever.css
They will be available like any typical static file server. 
Some web design themes look in /assets, which won't be served by cosgo. You can use the -static flag to change this.

Copy favicon.ico into static/. it will be served as if it were located at /favicon.ico

Also copy robot.txt into static/, it will be served at /robots.txt

/static
/static/js/jquery.js
/static/css/style.css
/static/favicon.ico
/static/robots.txt

/templates/index.html
/templates/error.html

-------

## cron
```cron

COSGO_DESTINATION=your@email.com
COSGOPORT=8080

@reboot /demon/$USER/cosgo -port $COSGOPORT> /dev/null 2>&1

```

-------


## Sample Nginx Config

```nginx
server {
        server_name default_address example.com example.net;
        listen 80;
        location / {
        proxy_pass http://127.0.0.1:8080; # Change to your cosgo port
        }
}

```

## Fortunes

Cosgo now has a random fortune for each visitor.
Default is disabled.
To enable it, make a file called "fortunes.txt" with **double-line separated pre-formatted text.**

Such as

```
Welcome!


Welcome!!
```
	 
Here is a `shell` script that populates the file with standard unix "fortune" command.
```

#!/bin/sh
# double line separated fortunes. some get cut off but oh well.
echo "Populating fortunes.txt, press Ctrl C when you think its big enough."
for i in $(cat fortunes.txt); do 
fortune -o >> fortunes.txt && echo "" >> fortunes.txt;

```

# More information at Wiki

More code examples at https://github.com/aerth/cosgo/wiki

Please add your use case. It could make cosgo better. Report any bugs or feature requests directly at https://isupon.us

-------

### Historical Information

* Casgo is short for "Contact API server in Golang"
* Turns out theres another casgo. Now this is cosgo.
* Cosgo is still not to be confused with Costco, the warehouse-style superstore.
* It began as a fork of https://github.com/munrocape/staticcontact
