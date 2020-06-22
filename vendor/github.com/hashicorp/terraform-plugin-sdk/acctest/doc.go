// Package acctest provides the ability to opt in to the new binary test driver. The binary
// test driver allows you to run your acceptance tests with a binary of Terraform instead of
// an emulated version packaged inside the SDK. This allows for a number of important
// enhancements, but most notably a more realistic testing experience and matrix testing
// against multiple versions of Terraform CLI. This also allows the SDK to be completely
// separated, at a dependency level, from the Terraform CLI, as long as it is >= 0.12.0
//
// The new test driver must be enabled by initialising the test helper in your TestMain
// function in all provider packages that run acceptance tests. Most providers have only
// one package.
//
// In v2 of the SDK, the binary test driver will be mandatory.
//
// After importing this package, you can add code similar to the following:
//
//   func TestMain(m *testing.M) {
//     acctest.UseBinaryDriver("provider_name", Provider)
//     resource.TestMain(m)
//   }
//
// Where `Provider` is the function that returns the instance of a configured `terraform.ResourceProvider`
// Some providers already have a TestMain defined, usually for the purpose of enabling test
// sweepers. These additional occurrences should be removed.
//
// Initialising the binary test helper using UseBinaryDriver causes all tests to be run using
// the new binary driver. Until SDK v2, the DisableBinaryDriver boolean property can be used
// to use the legacy test driver for an individual TestCase.
//
// It is no longer necessary to import other Terraform providers as Go modules: these
// imports should be removed.
package acctest
