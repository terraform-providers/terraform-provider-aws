# awsproviderlint

The `awsproviderlint` tool is a Terraform Provider code linting tool, specifically tailored for the Terraform AWS Provider.

## Lint Checks

For additional information about each check, you can run `awsproviderlint help NAME`.

### tfproviderlint Checks

The `awsproviderlint` tool extends the `tfproviderlint` tool and its checks. See the [`tfproviderlint` documentation](https://github.com/bflad/tfproviderlint) for additional information about the checks it provides.

### AWS Acceptance Test Checks

| Check | Description |
|---|---|
| [AWSAT001](passes/AWSAT001/README.md) | check for `resource.TestMatchResourceAttr()` calls against ARN attributes |

### AWS Resource Checks

| Check | Description |
|---|---|
| [AWSR001](passes/AWSR001/README.md) | check for `fmt.Sprintf()` calls using `.amazonaws.com` domain suffix |

## Development and Testing

This project is built on the [`tfproviderlint`](https://github.com/bflad/tfproviderlint) project and the [`go/analysis`](https://godoc.org/golang.org/x/tools/go/analysis) framework.

Helpful tooling for development:

* [`astdump`](https://github.com/wingyplus/astdump): a tool for displaying the AST form of Go file

### Unit Testing

```console
$ go test ./...
```

### Adding an Analyzer

NOTE: Provider-specific analyzers should implement their own namespace outside `tfproviderlint`'s AT### (acceptance testing), R### (resource), and S### (schema) to prevent naming collisions.

* Create new analyzer directory in `passes/`. The new directory name should match the name of the new analyzer.
  * Add `passes/NAME/README.md` which documents at least a description of analyzer.
  * Add `passes/NAME/NAME.go` which implements `Analyzer`.
  * If analyzer is a full check:
    * Include passing and failing example code in `passes/NAME/README.md`.
    * Add `passes/NAME/NAME_test.go` which implements `analysistest.TestData()` and `analysistest.Run()`.
    * Add `passes/NAME/testdata/src/a` directory with Go source files that implement passing and failing code based on `analysistest` framework.
    * Since the [`analysistest` package](https://godoc.org/golang.org/x/tools/go/analysis/analysistest) does not support Go Modules currently, each analyzer that implements testing must add a symlink to the top level `vendor` directory in the `testdata/src/a` directory. e.g. `ln -s ../../../../../../vendor passes/NAME/testdata/src/a/vendor`.
* Add new `Analyzer` in `main.go`.
