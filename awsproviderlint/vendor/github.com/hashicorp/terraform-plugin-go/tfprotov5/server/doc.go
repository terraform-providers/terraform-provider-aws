// Package tf5server implementations a server implementation for running
// tfprotov5.ProviderServers as gRPC servers.
//
// Providers will likely be calling tf5server.Serve from their main function to
// start the server so Terraform can connect to it.
package tf5server
