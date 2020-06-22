# keyvaluetags

The `keyvaluetags` package is designed to provide a consistent interface for handling AWS resource key-value tags. Many of the AWS Go SDK services, implement their own Go struct with `Key` and `Value` fields (e.g. `athena.Tag`) while others simply use a map (e.g. `map[string]string`). These inconsistent implementations and numerous Go types makes the process of correctly working with each of the services a tedius, previously copy-paste-modify process.

This package instead implements a single `KeyValueTags` type, which covers all key-value handling logic such as merging tags and ignoring keys via functions on the single type. The underlying implementation is compatible with Go operations such as `len()`.

Full documentation for this package can be found on [GoDoc](https://godoc.org/github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags).

Many AWS Go SDK services that support tagging have their service-specific Go type conversion functions to and from `KeyValueTags` code generated. Converting from `KeyValueTags` to AWS Go SDK types is done via `{SERVICE}Tags()` functions on the type, while converting from AWS Go SDK types to the `KeyValueTags` type is done via `{SERVICE}KeyValueTags()` functions. For more information about this code generation, see the [`generators/servicetags` README](generators/servicetags/README.md).

Some AWS Go SDK services that have common tag listing functionality (such as `ListTagsForResource` API call), also have auto-generated list functions. For more information about this code generation, see the [`generators/listtags` README](generators/listtags/README.md).

Some AWS Go SDK services that have common tagging update functionality (such as `TagResource` and `UntagResource` API calls), also have auto-generated update functions. For more information about this code generation, see the [`generators/updatetags` README](generators/updatetags/README.md).

Any tagging functions that cannot be generated should be hand implemented in a service-specific source file (e.g. `iam_tags.go`) and follow the format of similar generated code wherever possible. The first line of the source file should be `// +build !generate`. This prevents the file's inclusion during the code generation phase.

## Code Structure

```text
aws/internal/keyvaluetags
├── generators
│   ├── createtags (generates create_tags_gen.go)
│   ├── gettag (generates get_tag_gen.go)
│   ├── listtags (generates list_tags_gen.go)
│   ├── servicetags (generates service_tags_gen.go)
│   └── updatetags (generates update_tags_gen.go)
├── key_value_tags_test.go (unit tests for core logic)
├── key_value_tags.go (core logic)
├── list_tags_gen.go (generated AWS Go SDK service list tag functions)
├── service_generation_customizations.go (shared AWS Go SDK service customizations for generators)
├── service_tags_gen.go (generated AWS Go SDK service conversion functions)
├── update_tags_gen.go (generated AWS Go SDK service tagging update functions)
└── <service name>_tags.go (any service-specific functions that cannot be generated)
```
