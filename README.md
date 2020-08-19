# section-cli

Section command line tool.

## Usage

Run the command without any arguments to see the help:

```
section
```

To set up credentials so the CLI tool works, run:

```
section login
```

## Developing

Please ensure you're running at least Go 1.14.

To run tests:

```
git clone https://github.com/section/section-cli
cd section-cli
make test
```

To build a binary in `bin/`

```
make build
```
