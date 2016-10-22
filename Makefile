# cosco
# Copyright (c)2016 aerth <aerth@riseup.net>
# https://github.com/aerth
NAME=cosgo
VERSION=0.9.2
COMMIT=$(shell git rev-parse --verify --short HEAD)
RELEASE:=${VERSION}.X${COMMIT}
DEBUG=${VERSION}${COMMIT}DEBUG
GOFILES=${shell ls | grep '.go' | grep -v '_test'}

# Build a static linked binary
export CGO_ENABLED?=0

# Embed commit version into binary
GO_LDFLAGS=-ldflags "-s -X main.version=${RELEASE}"
DEBUGLDFLAGS=-ldflags "-X main.version=${DEBUG}"

# Install to /usr/local/
#PREFIX=/usr/local
PREFIX?=/usr/local

# Set temp gopath if none exists
ifeq (,${GOPATH})
export GOPATH=/tmp/gopath
endif

all: build
run: build
	bin/${NAME}-v${RELEASE} -port 8088
	
help:
		@echo 'Welcome and thanks for building COSGO!'
		@echo 'Here are some examples:'
		@echo
		@echo 'Normal build and install to /usr/local/bin, static linked with no debug symbols'
		@echo '		make'
		@echo '		sudo make install'
		@echo
		@echo 'Debug build with extra verbose logs and /debug/pprof URLs'
		@echo '		make debug'
		@echo
		@echo 'Custom install path'
		@echo '		PREFIX=$$HOME/bin make install'
		@echo
		@echo 'Generate 10 fortunes, append to fortunes.txt. These are available as a template variable.'
		@echo '		make fortune'


debug:
	set -e
	go fmt
	mkdir -p bin
	CGO_ENABLED=1	go build -v -x -tags debug ${DEBUGLDFLAGS} -o bin/${NAME}-v${DEBUG}
build: fortune
	set -e
	go fmt
	mkdir -p bin
	go build -v ${GO_LDFLAGS} -o bin/${NAME}-v${RELEASE}
	@echo Built bin/${NAME}-${RELEASE}

install:
	@echo installing to ${PREFIX}
	mkdir -p ${PREFIX}
	mv -v ./bin/${NAME}-v${RELEASE} ${PREFIX}/bin/${NAME}
	chmod 755 ${PREFIX}/bin/${NAME}

#run:

cross:
	mkdir -p bin
	@echo "Building for target: Windows"
	GOOS=windows GOARCH=386 go build -v ${GO_LDFLAGS} -o bin/${NAME}-v${RELEASE}-WIN32.exe
	GOOS=windows GOARCH=amd64 go build -v ${GO_LDFLAGS} -o bin/${NAME}-v${RELEASE}-WIN64.exe
	@echo "Building for target: OS X"
	GOOS=darwin GOARCH=386 go build -v ${GO_LDFLAGS} -o bin/${NAME}-v${RELEASE}-OSX-x86
	GOOS=darwin GOARCH=amd64 go build -v ${GO_LDFLAGS} -o bin/${NAME}-v${RELEASE}-OSX-amd64
	@echo "Building for target: Linux"
	GOOS=linux GOARCH=amd64 go build -v ${GO_LDFLAGS} -o bin/${NAME}-v${RELEASE}-linux-amd64
	GOOS=linux GOARCH=386 go build -v ${GO_LDFLAGS} -o bin/${NAME}-v${RELEASE}-linux-x86
	GOOS=linux GOARCH=arm go build -v ${GO_LDFLAGS} -o bin/${NAME}-v${RELEASE}-linux-arm
	@echo "Building for target: FreeBSD"
	GOOS=freebsd GOARCH=amd64 go build -v ${GO_LDFLAGS} -o bin/${NAME}-v${RELEASE}-freebsd-amd64
	GOOS=freebsd GOARCH=386 go build -v ${GO_LDFLAGS} -o bin/${NAME}-v${RELEASE}-freebsd-x86
	@echo "Building for target: OpenBSD"
	GOOS=openbsd GOARCH=amd64 go build -v ${GO_LDFLAGS} -o bin/${NAME}-v${RELEASE}-openbsd-amd64
	GOOS=openbsd GOARCH=386 go build -v ${GO_LDFLAGS} -o bin/${NAME}-v${RELEASE}-openbsd-x86
	@echo "Building for target: NetBSD"
	GOOS=netbsd GOARCH=amd64 go build -v ${GO_LDFLAGS} -o bin/${NAME}-v${RELEASE}-netbsd-amd64
	GOOS=netbsd GOARCH=386 go build -v ${GO_LDFLAGS} -o bin/${NAME}-v${RELEASE}-netbsd-x86
	@echo ${RELEASE} > bin/VERSION
	@echo "Now run ./pack.bash"

# package target is not working out, moved to a shell script named "pack.bash"
package:
	@echo "Run ./pack.bash"

# clean all remainders of testing
clean:
	rm -Rf bin pkg templates static cosgo.mbox cosgo.log HASH HASH.old debug.log

# download the dependencies
deps:
	go get -u -v -d .

# fortune activates random fortunes available as a template variable.
fortune:
	@set -e
	$(shell fortune >> fortunes.txt && echo "" >> fortunes.txt)
	$(shell fortune >> fortunes.txt && echo "" >> fortunes.txt)
	$(shell fortune >> fortunes.txt && echo "" >> fortunes.txt)
	$(shell fortune >> fortunes.txt && echo "" >> fortunes.txt)
	$(shell fortune >> fortunes.txt && echo "" >> fortunes.txt)
	$(shell fortune >> fortunes.txt && echo "" >> fortunes.txt)
	$(shell fortune >> fortunes.txt && echo "" >> fortunes.txt)
	$(shell fortune >> fortunes.txt && echo "" >> fortunes.txt)
	$(shell fortune >> fortunes.txt && echo "" >> fortunes.txt)
	$(shell fortune >> fortunes.txt && echo "" >> fortunes.txt)
	$(shell fortune >> fortunes.txt && echo "" >> fortunes.txt)
	$(shell fortune >> fortunes.txt && echo "" >> fortunes.txt)
	$(shell fortune >> fortunes.txt && echo "" >> fortunes.txt)
	@echo '10 Fortunes added'

test:
	go test -v -x
	echo ${GOFILES}
	@CGO_ENABLED=1 go run -race ${GOFILES} -- -key testdata/key.pem -cert testdata/cert.pem -port 8089 -tlsport 8777