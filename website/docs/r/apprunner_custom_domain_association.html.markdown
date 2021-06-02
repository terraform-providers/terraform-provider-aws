---
subcategory: "App Runner"
layout: "aws"
page_title: "AWS: aws_apprunner_custom_domain_association"
description: |-
  Manages an App Runner Custom Domain association.
---

# Resource: aws_apprunner_custom_domain_association

Manages an App Runner Custom Domain association.

~> **NOTE:** After creation, you must use the information in the `certification_validation_records` attribute to add CNAME records to your Domain Name System (DNS). For each mapped domain name, add a mapping to the target App Runner subdomain (found in the `dns_target` attribute) and one or more certificate validation records. App Runner then performs DNS validation to verify that you own or control the domain name you associated. App Runner tracks domain validity in a certificate stored in AWS Certificate Manager (ACM).

## Example Usage

```terraform
resource "aws_apprunner_custom_domain_association" "example" {
  domain_name = "example.com"
  service_arn = aws_apprunner_service.example.arn
}
```

## Argument Reference

The following arguments supported:

* `domain_name` - (Required) The custom domain endpoint to association. Specify a base domain e.g. `example.com` or a subdomain e.g. `subdomain.example.com`.
* `enable_www_subdomain` (Optional) Whether to associate the subdomain with the App Runner service in addition to the base domain. Defaults to `true`.
* `service_arn` - (Required) The ARN of the App Runner service.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The `domain_name` and `service_arn` separated by a comma (`,`).
* `certificate_validation_records` - A set of certificate CNAME records used for this domain name. See [Certificate Validation Records](#certificate-validation-records) below for more details.
* `dns_target` - The App Runner subdomain of the App Runner service. The custom domain name is mapped to this target name. Attribute only available if resource created (not imported) with Terraform.

### Certificate Validation Records

The configuration block consists of the following arguments:

* `name` - The certificate CNAME record name.
* `status` - The current state of the certificate CNAME record validation. It should change to `SUCCESS` after App Runner completes validation with your DNS.
* `type` - The record type, always `CNAME`.
* `value` - The certificate CNAME record value.

## Import

App Runner Custom Domain Associations can be imported by using the `domain_name` and `service_arn` separated by a comma (`,`), e.g.

```
$ terraform import aws_apprunner_custom_domain_association.example example.com,arn:aws:apprunner:us-east-1:123456789012:service/example-
app/8fe1e10304f84fd2b0df550fe98a71fa
```
