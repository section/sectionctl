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

export GOARCH := amd64
build: clean
	@if [ -z "$(VERSION)" ]; then echo "Missing VERSION"; exit 1 ; fi
	@if [ -z "$(GOOS)" ]; then echo "Missing GOOS"; exit 1 ; fi
	@if [ -z "$(GOARCH)" ]; then echo "Missing GOARCH"; exit 1 ; fi
	go build -ldflags "-X 'github.com/section/sectionctl/analytics.HeapAppID=$(HEAPAPPID)'" -o bin/sectionctl sectionctl.go
	mkdir -p dist/sectionctl-$(VERSION)-$(GOOS)-$(GOARCH)/
	cp README.md LICENSE bin/sectionctl dist/sectionctl-$(VERSION)-$(GOOS)-$(GOARCH)/
	tar --create --gzip --verbose --file dist/sectionctl-$(VERSION)-$(GOOS)-$(GOARCH).tar.gz --directory dist/sectionctl-$(VERSION)-$(GOOS)-$(GOARCH) .

clean:
	rm -rf build

release:
	@if [ -z "$(VERSION)" ]; then echo "Missing VERSION"; exit 1 ; fi
	@if [ "$(shell echo $(VERSION) | cut -c 1)" != "v" ]; then echo "VERSION must be in the format v0.0.5"; exit 1 ; fi
	git tag -f -a $(VERSION) -m ''
	git push --all
	git push --tags
