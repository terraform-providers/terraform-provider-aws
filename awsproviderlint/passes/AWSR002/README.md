# AWSR002

The AWSR002 analyzer reports when a [(schema.ResourceData).Set()](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema?tab=doc#ResourceData.Set) call with the `tags` key is missing a call to `(keyvaluetags.KeyValueTags).IgnoreConfig()` in the value, which ensures any provider level ignore tags configuration is applied.

## Flagged Code

```go
d.Set("tags", keyvaluetags.Ec2KeyValueTags(subnet.Tags).IgnoreAws().Map())
```

## Passing Code

```go
ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

d.Set("tags", keyvaluetags.Ec2KeyValueTags(subnet.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map())
```

## Ignoring Check

The check can be ignored for a certain line via a `//lintignore:AWSR002` comment on the previous line or at the end of the offending line, e.g.

```go
//lintignore:AWSR002
d.Set("tags", keyvaluetags.Ec2KeyValueTags(subnet.Tags).IgnoreAws().Map())
```
