---
subcategory: "Macie"
layout: "aws"
page_title: "AWS: aws_macie2_custom_data_identifier"
description: |-
  Provides a resource to manage an AWS Macie Custom Data Identifier.
---

# Resource: aws_macie2_custom_data_identifier

Provides a resource to manage an [AWS Macie Custom Data Identifier](https://docs.aws.amazon.com/macie/latest/APIReference/custom-data-identifiers-id.html).

## Example Usage

```terraform
resource "aws_macie2_account" "example" {}

resource "aws_macie2_custom_data_identifier" "example" {
  name                   = "NAME OF CUSTOM DATA IDENTIFIER"
  regex                  = "[0-9]{3}-[0-9]{2}-[0-9]{4}"
  description            = "DESCRIPTION"
  maximum_match_distance = 10
  keywords               = ["keyword"]
  ignore_words           = ["ignore"]

  depends_on = [aws_macie2_account.test]
}
```

## Argument Reference

The following arguments are supported:

* `regex` - (Optional) The regular expression (regex) that defines the pattern to match. The expression can contain as many as 512 characters.
* `keywords` -  (Optional) An array that lists specific character sequences (keywords), one of which must be within proximity (`maximum_match_distance`) of the regular expression to match. The array can contain as many as 50 keywords. Each keyword can contain 3 - 90 characters. Keywords aren't case sensitive.
* `ignore_words` - (Optional) An array that lists specific character sequences (ignore words) to exclude from the results. If the text matched by the regular expression is the same as any string in this array, Amazon Macie ignores it. The array can contain as many as 10 ignore words. Each ignore word can contain 4 - 90 characters. Ignore words are case sensitive.
* `name` - (Optional) A custom name for the custom data identifier. The name can contain as many as 128 characters. If omitted, Terraform will assign a random, unique name. Conflicts with `name_prefix`.
* `name_prefix` -  (Optional) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `description` - (Optional) A custom description of the custom data identifier. The description can contain as many as 512 characters.
* `maximum_match_distance` - (Optional) The maximum number of characters that can exist between text that matches the regex pattern and the character sequences specified by the keywords array. Macie includes or excludes a result based on the proximity of a keyword to text that matches the regex pattern. The distance can be 1 - 300 characters. The default value is 50.
* `tags` - (Optional) A map of key-value pairs that specifies the tags to associate with the custom data identifier.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier (ID) of the macie custom data identifier.
* `deleted` - Specifies whether the custom data identifier was deleted. If you delete a custom data identifier, Amazon Macie doesn't delete it permanently. Instead, it soft deletes the identifier.
* `created_at` - The date and time, in UTC and extended RFC 3339 format, when the Amazon Macie account was created.
* `arn` - The Amazon Resource Name (ARN) of the custom data identifier.

## Import

`aws_macie2_custom_data_identifier` can be imported using the id, e.g.

```
$ terraform import aws_macie2_custom_data_identifier.example abcd1
```
