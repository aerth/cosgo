# cosco 
# Copyright (c)2016 aerth <aerth@riseup.net>
# https://github.com/aerth
NAME=cosgo
VERSION=0.9.1
COMMIT=$(shell git rev-parse --verify --short HEAD)
RELEASE:=${VERSION}.X${COMMIT}


# Build a static linked binary
export CGO_ENABLED=0

# Embed commit version into binary
GO_LDFLAGS=-ldflags "-X main.version=$(RELEASE)"

# Install to /usr/local/
#PREFIX=/usr/local
PREFIX?=/usr/local

# Set temp gopath if none exists
ifeq (,${GOPATH})
export GOPATH=/tmp/gopath
endif

all: build 

build:
	set -e
	go fmt
	mkdir -p bin
	go build -v ${GO_LDFLAGS} -o bin/${NAME}-v${RELEASE}
	@echo Built ${NAME}-${RELEASE}

install:
	@echo installing to ${PREFIX}
	mkdir -p ${PREFIX}
	mv -v ./bin/${NAME}-v${RELEASE} ${PREFIX}/bin/${NAME}
	chmod 755 ${PREFIX}/bin/${NAME}

run:
	bin/${NAME}-v${RELEASE} 

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

clean:
	rm -Rf bin pkg templates static cosgo.mbox cosgo.log HASH HASH.old

deps:
	go get -v -d .