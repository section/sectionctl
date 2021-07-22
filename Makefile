.PHONY: test

export PATH := bin:$(PATH)
export HEAPAPPID := 892134159

staticcheck := /home/runner/go/bin/staticcheck

ifeq ($(OS),Windows_NT)     # is Windows_NT on XP, 2000, 7, Vista, 10...
    detected_OS := Windows
else
    detected_OS := $(shell uname)  # same as "uname -s"
endif


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
	go build -o sectionctl -ldflags "-X 'github.com/section/sectionctl/version.Version=$(shell echo $(VERSION) | cut -c 2-)'"


build-release: clean check_version
	@if [ -z "$(GOOS)" ]; then echo "Missing GOOS"; exit 1 ; fi
	@if [ -z "$(GOARCH)" ]; then echo "Missing GOARCH"; exit 1 ; fi
	go build -o sectionctl -ldflags "-X 'github.com/section/sectionctl/version.Version=$(shell echo $(VERSION) | cut -c 2-)'"
	mkdir -p dist/sectionctl-$(VERSION)-$(GOOS)-$(GOARCH)/
	cp README.md LICENSE sectionctl dist/sectionctl-$(VERSION)-$(GOOS)-$(GOARCH)/
	chmod +x dist/sectionctl-$(VERSION)-$(GOOS)-$(GOARCH)/sectionctl
	@if [ "$(GOOS)" = "windows" ]; then mv dist/sectionctl-$(VERSION)-$(GOOS)-$(GOARCH)/sectionctl dist/sectionctl-$(VERSION)-$(GOOS)-$(GOARCH)/sectionctl.exe  ; fi
	tar --create --gzip --verbose --file dist/sectionctl-$(VERSION)-$(GOOS)-$(GOARCH).tar.gz --directory dist/sectionctl-$(VERSION)-$(GOOS)-$(GOARCH) .

clean:
	rm -rf dist bin

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

test-release: check_version
	@if [ "$(shell git rev-parse --abbrev-ref HEAD)" != "v0.0.0-test" ]; then echo "Must be on the 'v0.0.0-test' branch"; exit 1 ; fi
	@git update-index --refresh
	@git diff-index --quiet HEAD --
	git tag -f -a $(VERSION) -m ''
	git push origin v0.0.0-test
	git push origin refs/tags/$(VERSION)



ifeq ($(detected_OS),Windows)
windows-installer: nsis-windows
else
windows-installer: nsis-ubuntu
endif

ifeq ($(detected_OS),Linux)
windows-installer: nsis-ubuntu
endif

ifeq (,$(wildcard /usr/share/nsis/Plugins/x86-unicode/EnVar.dll))
nsis-ubuntu: nsis-ubuntu-deps nsis-ubuntu-build
else
nsis-ubuntu: check_version nsis-ubuntu-build

endif

nsis-ubuntu-build: build-release
	version_major=$(shell echo $(VERSION) | cut -c 2- | cut -d. -f1)
	version_minor=$(shell echo $(VERSION) | cut -d. -f2)
	version_patch=$(shell echo $(VERSION) | cut -d. -f3)
	echo $(version_patch)
	echo $(version_minor)
	echo $(version_major)
	makensis -DARCH=amd64 -DOUTPUTFILE=sectionctl-$(VERSION)-windows-amd64-installer.exe -DSECTIONCTL_VERSION=$(VERSION) -DMAJORVERSION=$(version_major) -DDMINORVERSION=$(version_minor) -DBUILDVERSION=$(version_patch) packaging/windows/nsis.sectionctl.nsi 
nsis-ubuntu-deps:
	sudo apt update && sudo apt install -y nsis nsis-pluginapi
	wget https://nsis.sourceforge.io/mediawiki/images/7/7f/EnVar_plugin.zip
	sudo unzip EnVar_plugin.zip -d /usr/share/nsis/
	rm EnVar_plugin.zip

nsis-windows:
	echo "Not available on Windows yet"
