---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_client_vpn_endpoint"
description: |-
  Provides an AWS Client VPN endpoint for OpenVPN clients.
---

# Resource: aws_ec2_client_vpn_endpoint

Provides an AWS Client VPN endpoint for OpenVPN clients. For more information on usage, please see the
[AWS Client VPN Administrator's Guide](https://docs.aws.amazon.com/vpn/latest/clientvpn-admin/what-is.html).

## Example Usage

```terraform
resource "aws_ec2_client_vpn_endpoint" "example" {
  description            = "terraform-clientvpn-example"
  server_certificate_arn = aws_acm_certificate.cert.arn
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = aws_acm_certificate.root_cert.arn
  }

  connection_log_options {
    enabled               = true
    cloudwatch_log_group  = aws_cloudwatch_log_group.lg.name
    cloudwatch_log_stream = aws_cloudwatch_log_stream.ls.name
  }
}
```

## Argument Reference

The following arguments are supported:

* `authentication_options` - (Required) Information about the authentication method to be used to authenticate clients.
* `client_cidr_block` - (Required) The IPv4 address range, in CIDR notation, from which to assign client IP addresses. The address range cannot overlap with the local CIDR of the VPC in which the associated subnet is located, or the routes that you add manually. The address range cannot be changed after the Client VPN endpoint has been created. The CIDR block should be /22 or greater.
* `connection_log_options` - (Required) Information about the client connection logging options.
* `description` - (Optional) Name of the repository.
* `dns_servers` - (Optional) Information about the DNS servers to be used for DNS resolution. A Client VPN endpoint can have up to two DNS servers. If no DNS server is specified, the DNS address of the VPC that is to be associated with Client VPN endpoint is used as the DNS server.
* `server_certificate_arn` - (Required) The ARN of the ACM server certificate.
* `split_tunnel` - (Optional) Indicates whether split-tunnel is enabled on VPN endpoint. Default value is `false`.
* `tags` - (Optional) A mapping of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `transport_protocol` - (Optional) The transport protocol to be used by the VPN session. Default value is `udp`.

### `authentication_options` Argument Reference

One of the following arguments must be supplied:

* `active_directory_id` - (Optional) The ID of the Active Directory to be used for authentication if type is `directory-service-authentication`.
* `root_certificate_chain_arn` - (Optional) The ARN of the client certificate. The certificate must be signed by a certificate authority (CA) and it must be provisioned in AWS Certificate Manager (ACM). Only necessary when type is set to `certificate-authentication`.
* `saml_provider_arn` - (Optional) The ARN of the IAM SAML identity provider if type is `federated-authentication`.
* `type` - (Required) The type of client authentication to be used. Specify `certificate-authentication` to use certificate-based authentication, `directory-service-authentication` to use Active Directory authentication, or `federated-authentication` to use Federated Authentication via SAML 2.0.

### `connection_log_options` Argument Reference

One of the following arguments must be supplied:

* `cloudwatch_log_group` - (Optional) The name of the CloudWatch Logs log group.
* `cloudwatch_log_stream` - (Optional) The name of the CloudWatch Logs log stream to which the connection data is published.
* `enabled` - (Required) Indicates whether connection logging is enabled.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the Client VPN endpoint.
* `dns_name` - The DNS name to be used by clients when establishing their VPN session.
* `id` - The ID of the Client VPN endpoint.
* `status` - The current state of the Client VPN endpoint.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

AWS Client VPN endpoints can be imported using the `id` value found via `aws ec2 describe-client-vpn-endpoints`, e.g.

```
$ terraform import aws_ec2_client_vpn_endpoint.example cvpn-endpoint-0ac3a1abbccddd666
```
