.PHONY: test

export PATH := bin:$(PATH)

staticcheck := /home/runner/go/bin/staticcheck

all: test

cidep:
	go get -u honnef.co/go/tools/cmd/staticcheck
	go get -u github.com/kisielk/errcheck

test: gotest gostaticcheck goerrcheck

gotest:
	go test ./... -v -timeout=45s -failfast

gostaticcheck:
	staticcheck ./...

goerrcheck:
	errcheck -exclude .lint/errcheck-excludes -blank -ignoretests -ignore 'github.com/section/section-cli/analytics:^Print' -ignore 'github.com/section/section-cli/api/auth:^Print' ./...

build:
	go build -ldflags "-X 'github.com/section/section-cli/analytics.HeapAppID=892134159'" -o bin/section section.go
