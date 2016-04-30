# aerth [https://isupon.us]
# https://github.com/aerth
# This should do it: `make deps && make && sudo make install`
# yo type "make cross" to cross compile!
# for releases i use https://github.com/aerth/hashsum
NAME=cosgo
VERSION=0.6
RELEASE:=v${VERSION}.$(shell git rev-parse --verify --short HEAD)
PREFIX=/usr/local/bin
all:
	set -e
	go fmt
	go vet
	go build -v -o ${NAME}-${RELEASE} && echo Built ${NAME}-${RELEASE}

install:
	mv ${NAME}-${RELEASE} ${PREFIX}/${NAME}

deps:
	go get -v -d .

update:
	rm ${PREFIX}/${NAME} || true
	mv ${NAME}-${RELEASE} ${PREFIX}/${NAME}

cross:
	mkdir bin || true
	GOOS=windows GOARCH=386 go build -v -o bin/${NAME}-${RELEASE}-WIN32.exe
	GOOS=windows GOARCH=amd64 go build -v -o bin/${NAME}-${RELEASE}-WIN64.exe
	GOOS=darwin GOARCH=386 go build -v -o bin/${NAME}-${RELEASE}-OSX-x86
	GOOS=darwin GOARCH=amd64 go build -v -o bin/${NAME}-${RELEASE}-OSX-amd64
	GOOS=linux GOARCH=amd64 go build -v -o bin/${NAME}-${RELEASE}-linux-amd64
	GOOS=linux GOARCH=386 go build -v -o bin/${NAME}-${RELEASE}-linux-x86
	GOOS=freebsd GOARCH=amd64 go build -v -o bin/${NAME}-${RELEASE}-freebsd-amd64
	GOOS=freebsd GOARCH=386 go build -v -o bin/${NAME}-${RELEASE}-freebsd-x86
	GOOS=openbsd GOARCH=amd64 go build -v -o bin/${NAME}-${RELEASE}-openbsd-amd64
	GOOS=openbsd GOARCH=386 go build -v -o bin/${NAME}-${RELEASE}-openbsd-x86
	GOOS=netbsd GOARCH=amd64 go build -v -o bin/${NAME}-${RELEASE}-netbsd-amd64
	GOOS=netbsd GOARCH=386 go build -v -o bin/${NAME}-${RELEASE}-netbsd-x86
