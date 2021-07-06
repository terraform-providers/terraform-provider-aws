// Package tf6server implements a server implementation to run
// tfprotov6.ProviderServers as gRPC servers.
//
// Providers will likely be calling tf6server.Serve from their main function to
// start the server so Terraform can connect to it.
package tf6server
