# updatetags

This package contains a code generator to consistently handle the various AWS Go SDK service implementations for updating resource tags. Not all AWS Go SDK services that support tagging are generated in this manner.

To run this code generator, execute `go generate ./...` from the root of the repository. The general workflow for the generator is:

- Generate Go file contents via template from local variables and functions
- Go format file contents
- Write file contents to `update_tags_gen.go` file

## Example Output

```go
// AthenaUpdateTags updates athena service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func AthenaUpdateTags(conn *athena.Athena, identifier string, oldTagsMap interface{}, newTagsMap interface{}) error {
    oldTags := New(oldTagsMap)
    newTags := New(newTagsMap)

    if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
        input := &athena.UntagResourceInput{
            ResourceARN: aws.String(identifier),
            TagKeys:     aws.StringSlice(removedTags.Keys()),
        }

        _, err := conn.UntagResource(input)

        if err != nil {
            return fmt.Errorf("error untagging resource (%s): %s", identifier, err)
        }
    }

    if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
        input := &athena.TagResourceInput{
            ResourceARN: aws.String(identifier),
            Tags:        updatedTags.IgnoreAws().AthenaTags(),
        }

        _, err := conn.TagResource(input)

        if err != nil {
            return fmt.Errorf("error tagging resource (%s): %s", identifier, err)
        }
    }

    return nil
}
```

## Implementing a New Generated Service

### Requirements

Before a new service can be added to the generator, the new service must:

- Have the `KeyValueTags` conversion functions implemented for the AWS Go SDK service type/map. See also the [`servicetags` generator README](../servicetags/README.md).
- Implement a function for tagging (e.g. `TagResource`) and a function for untagging via keys (e.g. `UntagResource`)
- Have the service included in `aws/internal/keyvaluetags/service_generation_customizations.go`, if not present the following compilation error will be seen:

```text
2019/09/03 09:22:21 error executing template: template: listtags:19:41: executing "updatetags" at <ClientType>: error calling ClientType: unrecognized ServiceClientType: acmpca
```

Once the service has met all the requirements, in `main.go`:

- Add import for new service, e.g. `"github.com/aws/aws-sdk-go/service/athena"`
- Add service name to `serviceNames`, e.g. `athena`
- Add reflection handling to `ServiceClientType()` function, e.g.

```go
case "athena":
    funcType = reflect.TypeOf(athena.New)
```

- Run `go generate ./...` (or `make gen`) from the root of the repository to regenerate the code
- Run `go test ./...` (or `make test`) from the root of the repository to ensure the generated code compiles
- (Optional) Customize the service generation, if necessary (see below)

### Customizations

By default, the generator creates a `{SERVICE}UpdateTags()` function with the following structs and function calls:

- `{SERVICE}.TagResourceInput` struct with `ResourceArn` field and `Tags` field for calling `TagResource()` API call
- `{SERVICE}.UntagResourceInput` struct with `ResourceArn` field and `TagKeys` field for calling `UntagResource()` API call

If these do not match the actual AWS Go SDK service implementation, the generated code will compile with errors. See the sections below for certain errors and how to handle them.

#### ServiceTagFunction

Given the following compilation error:

```text
aws/internal/keyvaluetags/update_tags_gen.go:704:13: undefined: datapipeline.TagResourceInput
aws/internal/keyvaluetags/update_tags_gen.go:709:17: conn.TagResource undefined (type *datapipeline.DataPipeline has no field or method TagResource)
```

The function for resource tagging must be updated. Add an entry within the `ServiceTagFunction()` function of the generator to customize the naming of the `TagResource()` function and matching `TagResourceInput` struct. In the above case:

```go
case "datapipeline":
    return "AddTags"
```

#### ServiceTagInputCustomValue

Given the following compilation errors:

```text
aws/internal/keyvaluetags/update_tags_gen.go:1994:4: cannot use updatedTags.IgnoreAws().KinesisTags() (type []*kinesis.Tag) as type map[string]*string in field value
```

or

```text
aws/internal/keyvaluetags/update_tags_gen.go:2534:4: cannot use updatedTags.IgnoreAws().PinpointTags() (type map[string]*string) as type *pinpoint.TagsModel in field value
```

The value of the tags for tagging must be transformed. Add an entry within the `ServiceTagInputCustomValue()` function of the generator to return the custom value. In the above case:

```go
case "kinesis":
	return "aws.StringMap(chunk.IgnoreAws().Map())"
```

#### ServiceTagInputIdentifierField

Given the following compilation error:

```text
aws/internal/keyvaluetags/update_tags_gen.go:296:4: unknown field 'ResourceArn' in struct literal of type athena.UntagResourceInput (but does have ResourceARN)
aws/internal/keyvaluetags/update_tags_gen.go:309:4: unknown field 'ResourceArn' in struct literal of type athena.TagResourceInput (but does have ResourceARN)
```

The field name to identify the resource for tagging must be updated. Add an entry within the `ServiceTagInputIdentifierField()` function of the generator to customize the naming of the `ResourceArn` field for the tagging and untagging input structs. In the above case:

```go
case "athena":
    return "ResourceARN"
```

#### ServiceTagInputIdentifierRequiresSlice

Given the following compilation error:

```text
aws/internal/keyvaluetags/update_tags_gen.go:1296:4: cannot use aws.String(identifier) (type *string) as type []*string in field value
```

The value to identify the resource for tagging must be passed in a string slice. Add an entry within the `ServiceTagInputIdentifierRequiresSlice()` function of the generator to ensure that the value is passed as expected. In the above case

```go
case "ec2":
	return "yes"
```

#### ServiceTagInputTagsField

Given the following compilation error:

```text
aws/internal/keyvaluetags/update_tags_gen.go:382:4: unknown field 'Tags' in struct literal of type cloudhsmv2.TagResourceInput
```

The field name with the tags for tagging must be updated. Add an entry within the `ServiceTagInputTagsField()` function of the generator to customize the naming of the `Tags` field for the tagging input struct. In the above case:

```go
case "cloudhsmv2":
    return "TagList"
```

#### ServiceUntagFunction

Given the following compilation error:

```text
aws/internal/keyvaluetags/update_tags_gen.go:691:13: undefined: datapipeline.UntagResourceInput
aws/internal/keyvaluetags/update_tags_gen.go:696:17: conn.UntagResource undefined (type *datapipeline.DataPipeline has no field or method UntagResource)
```

The function for resource untagging must be updated. Add an entry within the `ServiceUntagFunction()` function of the generator to customize the naming of the `UntagResource()` function and matching `UntagResourceInput` struct. In the above case:

```go
case "datapipeline":
    return "RemoveTags"
```

#### ServiceUntagInputTagsField

Given the following compilation error:

```text
aws/internal/keyvaluetags/update_tags_gen.go:369:4: unknown field 'TagKeys' in struct literal of type cloudhsmv2.UntagResourceInput
```

The field name with the tag keys for untagging must be updated. Add an entry within the `ServiceUntagInputTagsField()` function of the generator to customize the naming of the `TagKeys` field for the untagging input struct. In the above case:

```go
case "cloudhsmv2":
    return "TagKeyList"
```

#### ServiceUntagInputCustomValue

Given the following compilation error:

```text
aws/internal/keyvaluetags/update_tags_gen.go:523:4: cannot use updatedTags.IgnoreAws().CloudfrontTags() (type []*cloudfront.Tag) as type *cloudfront.Tags in field value
```

The value of the tags for untagging must be transformed. Add an entry within the `ServiceUntagInputCustomValue()` function of the generator to return the custom value. In the above case:

```go
case "cloudfront":
	return "&cloudfront.TagKeys{Items: aws.StringSlice(removedTags.IgnoreAws().Keys())}"
```
