# seconf

### This library creates, detects, and reads non-plaintext configuration files.

[![GoDoc](https://godoc.org/github.com/aerth/seconf?status.svg)](https://godoc.org/github.com/aerth/seconf)

[(Example)](https://github.com/aerth/seconf/blob/master/_examples/hello/main.go)


seconf saves the configuration file as a `::::` separated list. for the encryption, i chose [(nacl/secretbox)](http://nacl.cr.yp.to/secretbox.html). theres a simple default pad that allows "empty" passwords. we use [bgentry/speakeasy](https://github.com/bgentry/speakeasy) for accepting user input for fields that contain "pass" or "key".

```
Secretbox uses XSalsa20 and Poly1305 to encrypt and authenticate messages with secret-key cryptography.
```

Future versions will store the values differently, using new functions, but the legacy functions will remain.

  * [go-quitter](https://github.com/aerth/go-quitter) for user, password, node URL.
  * [secenv](https://github.com/aerth/secenv) for user environment export seconf variables to shell
  * My contact form server, [cosgo](https://github.com/aerth/cosgo), uses seconf to store SMTP API keys and configuration fields.

## Attention

This project has not been reviewed by security professionals. Its internals, data formats, and interfaces may change at any time in the future without warning.


```

The MIT License (MIT)

Copyright (c) 2016 aerth

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

```
