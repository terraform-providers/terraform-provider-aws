# tfproviderdocs

A documentation tool for [Terraform Provider](https://www.terraform.io/docs/providers/index.html) code.

## Install

### Local Install

Release binaries are available in the [Releases](https://github.com/bflad/tfproviderdocs/releases) section.

To instead use Go to install into your `$GOBIN` directory (e.g. `$GOPATH/bin`):

```console
$ go get github.com/bflad/tfproviderdocs
```

### Docker Install

```console
$ docker pull bflad/tfproviderdocs
```

### Homebrew Install

```console
$ brew install bflad/tap/tfproviderdocs
```

## Usage

Additional information about usage and configuration options can be found by passing the `help` argument:

```console
$ tfproviderdocs help
```

### Local Usage

Change into the directory of the Terraform Provider code and run:

```console
$ tfproviderdocs
```

### Docker Usage

Change into the directory of the Terraform Provider code and run:

```console
$ docker run -v $(pwd):/src bflad/tfproviderdocs
```

## Available Commands

### check Command

The `tfproviderdocs check` command verifies the Terraform Provider documentation against the [specifications from Terraform Registry documentation](https://www.terraform.io/docs/registry/providers/docs.html) and common practices across official Terraform Providers. This includes the following checks:

- Verifies that no invalid directories are found in the documentation directory structure.
- Ensures that there is not a mix (legacy and Terraform Registry) of directory structures, which is not supported during Terraform Registry documentation ingress.
- Verifies side navigation for missing links and mismatched link text (if legacy directory structure).
- Verifies number of documentation files is below Terraform Registry storage limits.
- Verifies all known data sources and resources have an associated documentation file (if `-providers-schema-json` is provided)
- Verifies no extraneous or incorrectly named documentation files exist (if `-providers-schema-json` is provided)
- Verifies each file in the documentation directories is valid.

The validity of files is checked with the following rules:

- Proper file extensions are used (e.g. `.md` for Terraform Registry).
- Verifies size of file is below Terraform Registry storage limits.
- YAML frontmatter can be parsed and matches expectations.

The YAML frontmatter checks include some defaults (e.g. no `layout` field for Terraform Registry), but there are some useful flags that can be passed to the command to tune the behavior, especially for larger Terraform Providers.

For additional information about check flags, you can run `tfproviderdocs check -help`.

## Development and Testing

This project uses [Go Modules](https://github.com/golang/go/wiki/Modules) for dependency management.

### Updating Dependencies

```console
$ go get URL
$ go mod tidy
$ go mod vendor
```

### Unit Testing

```console
$ go test ./...
```

### Local Install Testing

```console
$ go install .
```
