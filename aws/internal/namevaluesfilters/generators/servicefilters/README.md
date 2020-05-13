# servicefilters

This package contains a code generator to consistently handle the various AWS Go SDK service implementations for converting service filter types to/from `NameValuesFilters`. Not all AWS Go SDK services that support tagging are generated in this manner.

To run this code generator, execute `go generate ./...` from the root of the repository. The general workflow for the generator is:

- Generate Go file contents via template from local variables and functions
- Go format file contents
- Write file contents to `service_filters_gen.go` file
