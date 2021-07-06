---
subcategory: "ElasticSearch"
layout: "aws"
page_title: "AWS: aws_elasticsearch_domain_saml_options"
description: |-
  Terraform resource for managing SAML authentication options for an AWS Elasticsearch Domain.
---

# Resource: aws_elasticsearch_domain_saml_options

Manages SAML authentication options for an AWS Elasticsearch Domain.

## Example Usage

### Basic Usage

```terraform
resource "aws_elasticsearch_domain" "example" {
  domain_name           = "example"
  elasticsearch_version = "1.5"

  cluster_config {
    instance_type = "r4.large.elasticsearch"
  }

  snapshot_options {
    automated_snapshot_start_hour = 23
  }

  tags = {
    Domain = "TestDomain"
  }
}

resource "aws_elasticsearch_domain_saml_options" "example" {
  domain_name = aws_elasticsearch_domain.example.domain_name
  saml_options {
    enabled = true
    idp {
      entity_id        = "https://example.com"
      metadata_content = file("./saml-metadata.xml")
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `domain_name` - (Required) Name of the domain.

The following arguments are optional:

* `saml_options` - (Optional) The SAML authentication options for an AWS Elasticsearch Domain.

### saml_options

* `enabled` - (Required) Whether SAML authentication is enabled.
* `idp` - (Optional) Information from your identity provider.
* `master_backend_role` - (Optional) This backend role from the SAML IdP receives full permissions to the cluster, equivalent to a new master user.
* `master_user_name` - (Optional) This username from the SAML IdP receives full permissions to the cluster, equivalent to a new master user.
* `roles_key` - (Optional) Element of the SAML assertion to use for backend roles. Default is roles.
* `session_timeout_minutes` - (Optional) Duration of a session in minutes after a user logs in. Default is 60. Maximum value is 1,440.
* `subject_key` - (Optional) Element of the SAML assertion to use for username. Default is NameID.


#### idp

* `entity_id` - (Required) The unique Entity ID of the application in SAML Identity Provider.
* `metadata_content` - (Required) The Metadata of the SAML application in xml format.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the domain the SAML options are associated with.

## Import

Elasticsearch domains can be imported using the `domain_name`, e.g.

```
$ terraform import aws_elasticsearch_domain_saml_options.example domain_name
```
