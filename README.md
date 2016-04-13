## cosgo
when all you needed was a contact form, anyways.

* Save the mail: Contact form saves messages to a local mbox file
* Read the mail with: `mutt -Rf cosgo.mbox`
* Option to send mail using SMTP, with Sendgrid (free) or Mandrill (not free).
* Customize: Uses Go style templates, `templates/index.html` and `templates/error.html`
* Static Files: /sitemap.xml /favicon.ico, /robots.txt. Limited to .woff, .ttf, .css, .js .png .jpg
* You can stuff .zip, .tgz. .tar, .tar.gz files in /files/ and cosgo will serve them, too.
* Tested on NetBSD and Debian servers, even runs on Windows. Probably runs great on anything else, too.
* This needs some testing on other setups! If you use cosgo, I would love to hear from you! https://isupon.us
* Now with -pages and -nolog and more in `cosgo --help`

```

The MIT License (MIT)

Copyright (c) 2016 aerth

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

```

## Usage Examples



```
cosgo -h # help
cosgo -noredirect # show 404 pages instead of redirecting to /
cosgo -debug # more messages and don't switch to cosgo.log
cosgo -static=false # don't serve /static, /files or /page
cosgo -custom "MyApp" -nolog # run our MyApp config with nolog
cosgo --nolog # no output whatsoever
cosgo -bind 127.0.0.1 -port 8000 -fastcgi # nginx setup

```

-------

## Installation

Releases:

1. Extract the archive.
2. Copy the binary to /usr/local/bin/ ( or extract everything to C:/cosgo/ )
3. Copy the templates/ and static/ folder to where you are running cosgo
4. Run `cosgo -config` from the command line.
5. Edit templates, static files

Building from source:

```
export GOPATH=/tmp/cosgo
go get -v -u github.com/aerth/cosgo
cd $GOPATH/src/github.com/aerth/cosgo
go build
./cosgo -h # list flags
./cosgo -config # Creates the config ( or runs if exists )
./cosgo -config -debug &
firefox http://127.0.0.1:8080 # Check it out.

```

-------

## Theme

Customize the theme by modifying ./templates/index.html and ./templates/error.html

Included is the bare minimal.

Upload your static .css .js .png .jpeg .woff and .ttf files in ./static/ like /static/{foldername}/whatever.css, they will be available like any typical static file server. Our routing system disables indexing.

Modify your theme if it looks in /assets/.

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
