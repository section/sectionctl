.PHONY: test

export PATH := bin:$(PATH)

staticcheck := /go/bin/staticcheck

all: test

cidep:
	wget https://github.com/dominikh/go-tools/releases/download/2019.1.1/staticcheck_linux_amd64 -O $(staticcheck)
	chmod +x $(staticcheck)
	go get -u github.com/kisielk/errcheck

test: gotest gostaticcheck goerrcheck

gotest:
	go test ./... -v -timeout=45s -failfast

gostaticcheck:
	staticcheck ./...

goerrcheck:
	errcheck -exclude .lint/errcheck-excludes -ignoretests ./...

build:
	go build -o bin/section section.go
