# aerth [https://isupon.us]
# https://github.com/aerth
NAME=cosgo
VERSION=0.9
RELEASE:=${VERSION}.X${COMMIT}
COMMIT=$(shell git rev-parse --verify --short HEAD)
GO_LDFLAGS=-ldflags "-X main.version=$(RELEASE)"
PREFIX=/usr/local
PREFIX?=$(shell pwd)
ifeq (,${GOPATH})
export GOPATH=/tmp/gopath
endif

all: build 

build:
	set -e
	go fmt
	mkdir -p bin
	go build -v ${GO_LDFLAGS} -o bin/${NAME}-v${RELEASE} && echo Built ${NAME}-${RELEASE}
	chmod 755 bin/${NAME}-v${RELEASE}

install:
	@echo installing to /usr/local/bin/
	mv -v ./bin/${NAME}-v${RELEASE} ${PREFIX}/bin/${NAME}
	chmod 755 ${PREFIX}/bin/${NAME}

run:
	bin/${NAME}-v${RELEASE} 

cross:
	mkdir -p bin
	GOOS=windows GOARCH=386 go build -v ${GO_LDFLAGS} -o bin/${NAME}-v${RELEASE}-WIN32.exe
	GOOS=windows GOARCH=amd64 go build -v ${GO_LDFLAGS} -o bin/${NAME}-v${RELEASE}-WIN64.exe
	GOOS=darwin GOARCH=386 go build -v ${GO_LDFLAGS} -o bin/${NAME}-v${RELEASE}-OSX-x86
	GOOS=darwin GOARCH=amd64 go build -v ${GO_LDFLAGS} -o bin/${NAME}-v${RELEASE}-OSX-amd64
	GOOS=linux GOARCH=amd64 go build -v ${GO_LDFLAGS} -o bin/${NAME}-v${RELEASE}-linux-amd64
	GOOS=linux GOARCH=386 go build -v ${GO_LDFLAGS} -o bin/${NAME}-v${RELEASE}-linux-x86
	GOOS=freebsd GOARCH=amd64 go build -v ${GO_LDFLAGS} -o bin/${NAME}-v${RELEASE}-freebsd-amd64
	GOOS=freebsd GOARCH=386 go build -v ${GO_LDFLAGS} -o bin/${NAME}-v${RELEASE}-freebsd-x86
	GOOS=openbsd GOARCH=amd64 go build -v ${GO_LDFLAGS} -o bin/${NAME}-v${RELEASE}-openbsd-amd64
	GOOS=openbsd GOARCH=386 go build -v ${GO_LDFLAGS} -o bin/${NAME}-v${RELEASE}-openbsd-x86
	GOOS=netbsd GOARCH=amd64 go build -v ${GO_LDFLAGS} -o bin/${NAME}-v${RELEASE}-netbsd-amd64
	GOOS=netbsd GOARCH=386 go build -v ${GO_LDFLAGS} -o bin/${NAME}-v${RELEASE}-netbsd-x86
	echo ${RELEASE} > bin/VERSION
	for i in $(ls ./bin/); do sha384sum $i >> bin/HASH; done

package: cross
	for i in $(ls ./bin/ | grep "-v"); do zip $i.zip $i README.md LICENSE.md HASH; done

