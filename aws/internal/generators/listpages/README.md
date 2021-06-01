# listpages

The `listpages` generator creates paginated variants of AWS Go SDK functions that return collections of objects where the SDK does not define them. It should typically be called using [`go generate`](https://golang.org/cmd/go/#hdr-Generate_Go_files_by_processing_source).

For example, the EC2 API defines both [`DescribeInstancesPages`](https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.DescribeInstancesPages) and  [`DescribeInstances`](https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.DescribeInstances), whereas the CloudWatch Events API defines only [`ListEventBuses`](https://docs.aws.amazon.com/sdk-for-go/api/service/cloudwatchevents/#CloudWatchEvents.ListEventBuses).

The `listpages` executable is called as follows:

```console
$ go run main.go -function <function-name>[,<function-name>] <source-package>
```

* `<source-package>`: The full Go package name of the AWS Go SDK package to be extended, e.g. `github.com/aws/aws-sdk-go/service/cloudwatchevents`
* `<function-name>`: Name of a function to wrap

Optional Flags:

* `-paginator`: Name of the pagination token field (default `NextToken`)
* `-package`: Override the package name for the generated code (By default, uses the environment variable `$GOPACKAGE` set by `go generate`)

To use with `go generate`, add the following directive to a Go file

```go
//go:generate go run <relative-path-to-generators>/generators/listpages/main.go -function=<comma-separated-list-of-functions> <aws-sdk-package>
```

For example, in the file `aws/internal/service/cloudwatchevents/lister/list.go`

```go
//go:generate go run ../../../generators/listpages/main.go -function=ListEventBuses,ListRules,ListTargetsByRule github.com/aws/aws-sdk-go/service/cloudwatchevents

package lister
```

Generates the file `aws/internal/service/cloudwatchevents/lister/list_pages_gen.go` with the functions `ListEventBusesPages`, `ListRulesPages`, and `ListTargetsByRulePages`.
