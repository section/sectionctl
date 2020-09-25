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
	errcheck -exclude .lint/errcheck-excludes -blank -ignoretests -ignore 'github.com/section/sectionctl/analytics:^Print' ./...

export GOARCH := amd64
build: clean
	@if [ -z "$(VERSION)" ]; then echo "Missing VERSION"; exit 1 ; fi
	@if [ -z "$(GOOS)" ]; then echo "Missing GOOS"; exit 1 ; fi
	@if [ -z "$(GOARCH)" ]; then echo "Missing GOARCH"; exit 1 ; fi
	go build -ldflags "-X 'github.com/section/sectionctl/analytics.HeapAppID=$(HEAPAPPID)'" -o build/sectionctl sectionctl.go
	cp README.md LICENSE build/
	mkdir -p dist
	tar --create --gzip --verbose --strip-components 1 -f dist/sectionctl-$(VERSION)-$(GOOS)-$(GOARCH).tar.gz build/

clean:
	rm -rf build
