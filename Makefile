.PHONY: test

export PATH := bin:$(PATH)
export HEAPAPPID := 892134159

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
	errcheck -exclude .lint/errcheck-excludes -blank -ignoretests ./...

build: clean
	go build sectionctl.go
	mkdir -p tmp
	mkdir -p luft
	env GOOS=linux GOARCH=amd64 go build -o tmp/sectionctl sectionctl.go
	tar --create --gzip --verbose --file luft/sectionctl.tar.gz --directory tmp .

build-release: clean check_version
	@if [ -z "$(GOOS)" ]; then echo "Missing GOOS"; exit 1 ; fi
	@if [ -z "$(GOARCH)" ]; then echo "Missing GOARCH"; exit 1 ; fi
	go build -o sectionctl -ldflags "-X 'github.com/section/sectionctl/version.Version=$(shell echo $(VERSION) | cut -c 2-)'" sectionctl.go
	mkdir -p dist/sectionctl-$(VERSION)-$(GOOS)-$(GOARCH)/
	cp README.md LICENSE sectionctl dist/sectionctl-$(VERSION)-$(GOOS)-$(GOARCH)/
	@if [ "$(GOOS)" = "windows" ]; then mv dist/sectionctl-$(VERSION)-$(GOOS)-$(GOARCH)/sectionctl dist/sectionctl-$(VERSION)-$(GOOS)-$(GOARCH)/sectionctl.exe  ; fi
	tar --create --gzip --verbose --file dist/sectionctl-$(VERSION)-$(GOOS)-$(GOARCH).tar.gz --directory dist/sectionctl-$(VERSION)-$(GOOS)-$(GOARCH) .

clean:
	rm -rf dist bin luft tmp

check_version:
	@if [ -z "$(VERSION)" ]; then echo "Missing VERSION"; exit 1 ; fi
	@if [ "$(shell echo $(VERSION) | cut -c 1)" != "v" ]; then echo "VERSION must be in the format v0.0.5"; exit 1 ; fi

release: check_version
	@if [ "$(shell git rev-parse --abbrev-ref HEAD)" != "main" ]; then echo "Must be on the 'main' branch"; exit 1 ; fi
	@git update-index --refresh
	@git diff-index --quiet HEAD --
	git tag -f -a $(VERSION) -m ''
	git push origin main
	git push origin refs/tags/$(VERSION)
