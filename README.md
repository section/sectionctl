# sectionctl

[![Test status](https://github.com/section/sectionctl/workflows/Test/badge.svg?event=push)](https://github.com/section/sectionctl/actions?query=workflow%3ATest+event%3Apush)

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

You can set credentials via the `SECTION_USERNAME` and `SECTION_PASSWORD` environment variables:

```
SECTION_USERNAME=me@domain.example SECTION_PASSWORD=s3cr3t sectionctl accounts list
```

## Installing

The easiest way to install `sectionctl` is by downloading [the latest release](https://github.com/section/sectionctl/releases/latest) and putting it on your `PATH`.

## Developing

Please ensure you're running at least Go 1.14.

To run tests:

```
git clone https://github.com/section/sectionctl
cd sectionctl
make test
```

To build a binary in `bin/`

```
make build
```

## Releasing

1. Increment the version number in `version/version.go` and commit.
1. Run `make release` and specify VERSION string prefaced with a `v`, like `v1.0.1`.

This triggers [a GitHub Actions workflow](https://github.com/section/sectionctl/actions?query=workflow%3A%22Build+and+release+sectionctl+binaries%22) that does cross platform builds, and publishes [a draft release](https://github.com/section/sectionctl/releases).
