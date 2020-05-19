# servicefilters

This package contains a code generator to consistently handle the various AWS Go SDK service implementations for converting service filter types to/from `NameValuesFilters`. Not all AWS Go SDK services that support tagging are generated in this manner.

To run this code generator, execute `go generate ./...` from the root of the repository. The general workflow for the generator is:

- Generate Go file contents via template from local variables and functions
- Go format file contents
- Write file contents to `service_filters_gen.go` file

## Example Output

```go
// DocdbFilters returns docdb service filters.
func (filters NameValuesFilters) DocdbFilters() []*docdb.Filter {
	nvfm := filters.Map()

	if len(nvfm) == 0 {
		return nil
	}

	result := make([]*docdb.Filter, 0, len(nvfm))

	for k, v := range nvfm {
		filter := &docdb.Filter{
			Name:   aws.String(k),
			Values: aws.StringSlice(v),
		}

		result = append(result, filter)
	}

	return result
}
```

## Implementing a New Generated Service

- In `main.go`: Add service name, e.g. `docdb`, to one of the implementation handlers
  - Use `sliceServiceNames` if the AWS Go SDK service implements a specific Go type such as `Filter`
- Run `go generate ./...` (or `make gen`) from the root of the repository to regenerate the code
- Run `go test ./...` (or `make test`) from the root of the repository to ensure the generated code compiles
