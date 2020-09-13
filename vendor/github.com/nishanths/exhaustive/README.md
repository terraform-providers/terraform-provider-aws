# exhaustive

[![Godoc](https://godoc.org/github.com/nishanths/exhaustive?status.svg)](https://godoc.org/github.com/nishanths/exhaustive)

The `exhaustive` package and command line program can be used to detect
enum switch statements that are not exhaustive.

An enum switch statement is exhaustive if it has cases for each of the enum's members. See godoc for the definition of enum used by the program.

The `exhaustive` package provides an `Analyzer` that follows the guidelines
described in the [go/analysis](https://godoc.org/golang.org/x/tools/go/analysis) package; this makes
it possible to integrate into existing analysis driver programs.

## Install

```
go get github.com/nishanths/exhaustive/...
```

## Docs

https://godoc.org/github.com/nishanths/exhaustive

## Usage

The command line usage is:

```
Usage: exhaustive [-flags] [packages...]

Flags:
  -default-signifies-exhaustive
    	indicates that switch statements are to be considered exhaustive if a 'default' case
    	is present, even if all enum members aren't listed in the switch (default false)
  -fix
    	apply all suggested fixes (default false)

Examples:
  exhaustive code.org/proj/...
  exhaustive -fix example.org/foo/pkg example.org/foo/bar
```

## Example

Given the code:

```diff
package token

type Token int

const (
	Add Token = iota
	Subtract
	Multiply
+	Quotient
+	Remainder
)
```
```
package calc

import "token"

func processToken(t token.Token) {
	switch t {
	case token.Add:
		// ...
	case token.Subtract:
		// ...
	case token.Multiply:
		// ...
	}
}
```

Running the `exhaustive` command will print:

```
calc.go:6:2: missing cases in switch of type token.Token: Quotient, Remainder
```

## License

BSD 2-Clause
