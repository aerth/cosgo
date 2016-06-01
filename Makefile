# aerth [https://isupon.us]
# https://github.com/aerth
# This should do it: `make deps && make && sudo make install`
# yo type "make cross" to cross compile!
# for releases i use https://github.com/aerth/hashsum
NAME=cosgo
VERSION=0.7
RELEASE:=${VERSION}.$(shell git rev-parse --verify --short HEAD)
GO_LDFLAGS=-ldflags "-X main.Version=$(RELEASE)"
PREFIX=/usr/local
PREFIX?=$(shell pwd)
ifeq (,${GOPATH})
export GOPATH=/tmp/gopath
endif
all: deps
	set -e
	go fmt
#	go vet
	mkdir bin || true
	go build -v ${GO_LDFLAGS} -o bin/${NAME}-v${RELEASE} && echo Built ${NAME}-${RELEASE}
	chmod 755 bin/${NAME}-v${RELEASE}

install:
	echo installing to /usr/local/bin and /usr/local/share/cosgo/templates/
	mv ./bin/${NAME}-v${RELEASE} ${PREFIX}/bin/${NAME}
	set noclobber on
	mkdir -p ${PREFIX}/share/${NAME}/ || true
	cp -R templates ${PREFIX}/share/${NAME}/templates
	cp -R static ${PREFIX}/share/${NAME}/static
	chmod -R 755 ${PREFIX}/share/${NAME}

deps:
	go get -v -d .

update:
	rm ${PREFIX}/${NAME} || true
	mv ${NAME}-v${RELEASE} ${PREFIX}/${NAME}

cross:
	mkdir bin || true
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

package:
	cross
	for i in $(ls ./bin/ | grep "-v"); do zip $i.zip $i README.md LICENSE.md HASH; done

