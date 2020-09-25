# sectionctl

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
