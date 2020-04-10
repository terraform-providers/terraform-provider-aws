# AWSR001

The `AWSR001` analyzer reports when a fmt.Sprintf() call contains the format string ending `.amazonaws.com`. This domain suffix is only valid in the AWS Commercial and GovCloud (US) partitions.

To ensure the correct domain suffix is used in all partitions, the `*AWSClient` available to all resources provides the `PartitionHostname()` and `RegionalHostname()` receiver methods.

## Flagged Code

```go
fmt.Sprintf("%s.amazonaws.com", d.Id())

fmt.Sprintf("%s.%s.amazonaws.com", d.Id(), meta.(*AWSClient).region)
```

## Passing Code

```go
meta.(*AWSClient).PartitionHostname(d.Id())

meta.(*AWSClient).RegionalHostname(d.Id())
```

## Ignoring Check

The check can be ignored for a certain line via a `//lintignore:AWSR001` comment on the previous line or at the end of the offending line, e.g.

```go
//lintignore:AWSR001
fmt.Sprintf("%s.amazonaws.com", d.Id())
```
