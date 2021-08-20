# sectionctl

[![Test status](https://github.com/section/sectionctl/workflows/Test/badge.svg?event=push)](https://github.com/section/sectionctl/actions?query=workflow%3ATest+event%3Apush)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fsection%2Fsectionctl.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fsection%2Fsectionctl?ref=badge_shield)

Section command line tool.

## Usage

Run the command without any arguments to see the help:

```
sectionctl
```

To set up credentials so the CLI tool works, run:

```
sectionctl login
```

Install bash shell completions with:

```
sectionctl install-completions
```

### Using `sectionctl` in CI/CD

You can set credentials via the `SECTION_TOKEN` environment variable:

```bash
SECTION_TOKEN=s3cr3t sectionctl accounts list
```

## Installing

### Mac

```
brew tap section/brews
brew install sectionctl
```

### Linux

Install the .deb or .rpm, or download the binaries for your system and put sectionctl in your PATH.

### Windows

Install with the installer exe on [the latest release](https://github.com/section/sectionctl/releases/latest).

### Manual Installation

The easiest way to install `sectionctl` is by downloading [the latest release](https://github.com/section/sectionctl/releases/latest) and putting it on your `PATH`.

## Developing

Please ensure you're running at least Go 1.16.

To run tests:

```
git clone https://github.com/section/sectionctl
cd sectionctl
make test
```

To run a development version of `sectionctl`:

```
go run sectionctl.go
```

Add whatever flags and arguments you need at the end of above command.

### Building

To build a binary at `./sectionctl`:

```
make build
```

## Releasing

Run `make release` and specify VERSION string prefaced with a `v`, like `v1.0.1`.

This triggers [a GitHub Actions workflow](https://github.com/section/sectionctl/actions?query=workflow%3A%22Build+and+release+sectionctl+binaries%22) that does cross platform builds, and publishes [a draft release](https://github.com/section/sectionctl/releases).


## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fsection%2Fsectionctl.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fsection%2Fsectionctl?ref=badge_large)