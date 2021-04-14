---
subcategory: "ECR"
layout: "aws"
page_title: "AWS: aws_ecrpublic_repository"
description: |-
  Provides a Public Elastic Container Registry Repository.
---

# Resource: aws_ecrpublic_repository

Provides a Public Elastic Container Registry Repository.

## Example Usage

```terraform
resource "aws_ecrpublic_repository" "foo" {
  repository_name = "bar"

  catalog_data {
    about_text        = "About Text"
    architectures     = ["Linux"]
    description       = "Description"
    logo_image_blob   = filebase64(image.png)
    operating_systems = ["ARM"]
    usage_text        = "Usage Text"
  }
}
```

## Argument Reference

The following arguments are supported:

* `repository_name` - (Required) Name of the repository.
* `catalog_data` - (Optional) Catalog data configuration for the repository. See [below for schema](#catalog_data).


### catalog_data

* `about_text` - (Optional) A detailed description of the contents of the repository. It is publicly visible in the Amazon ECR Public Gallery. The text must be in markdown format.
* `architectures` - (Optional) The system architecture that the images in the repository are compatible with. On the Amazon ECR Public Gallery, the following supported architectures will appear as badges on the repository and are used as search filters: `Linux`, `Windows`
* `description` - (Optional) A short description of the contents of the repository. This text appears in both the image details and also when searching for repositories on the Amazon ECR Public Gallery.
* `logo_image_blob` - (Optional) The base64-encoded repository logo payload. (Only visible for verified accounts) Note that drift detection is disabled for this attribute.
* `operating_systems` -  (Optional) The operating systems that the images in the repository are compatible with. On the Amazon ECR Public Gallery, the following supported operating systems will appear as badges on the repository and are used as search filters. `ARM`, `ARM 64`, `x86`, `x86-64`
* `usage_text` -  (Optional) Detailed information on how to use the contents of the repository. It is publicly visible in the Amazon ECR Public Gallery. The usage text provides context, support information, and additional usage details for users of the repository. The text must be in markdown format.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Full ARN of the repository.
* `id` - The repository name.
* `registry_id` - The registry ID where the repository was created.
* `repository_uri` - The URI of the repository.

## Timeouts

`aws_ecrpublic_repository` provides the following [Timeouts](/docs/configuration/resources.html#timeouts)
configuration options:

- `delete` - (Default `20 minutes`) How long to wait for a repository to be deleted.

## Import

ECR Public Repositories can be imported using the `repository_name`, e.g.

```
$ terraform import aws_ecrpublic_repository.example example
```
