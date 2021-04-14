// Package tfprotov5 provides the interfaces and types needed to build a
// Terraform provider server.
//
// All Terraform provider servers should be built on
// these types, to take advantage of the ecosystem and tooling built around
// them.
//
// These types are small wrappers around the Terraform protocol. It is assumed
// that developers using tfprotov5 are familiar with the protocol, its
// requirements, and its semantics. Developers not comfortable working with the
// raw protocol should use the github.com/hashicorp/terraform-plugin-sdk/v2 Go
// module instead, which offers a less verbose, safer way to develop a
// Terraform provider, albeit with less flexibility and power.
//
// Provider developers should start by defining a type that implements the
// `ProviderServer` interface. A struct is recommended, as it will allow you to
// store the configuration information attached to your provider for use in
// requests, but any type is technically possible.
//
// `ProviderServer` implementations will need to implement the composed
// interfaces, `ResourceServer` and `DataSourceServer`. It is recommended, but
// not required, to use an embedded `ResourceRouter` and `DataSourceRouter` in
// your `ProviderServer` to achieve this, which will let you handle requests
// for each resource and data source in a resource-specific or data
// source-specific function.
//
// To serve the `ProviderServer` implementation as a gRPC server that Terraform
// can connect to, use the `tfprotov5/server.Serve` function.
package tfprotov5
