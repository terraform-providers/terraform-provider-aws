---
subcategory: "Lex"
layout: "aws"
page_title: "AWS: aws_lex_bot_alias"
description: |-
  Provides details about a specific Lex Bot Alias
---

# Data Source: aws_lex_bot_alias

Provides details about a specific Amazon Lex Bot Alias.

## Example Usage

```hcl
data "aws_lex_bot_alias" "order_flowers_prod" {
  bot_name = "OrderFlowers"
  name     = "OrderFlowersProd"
}
```

## Argument Reference

The following arguments are supported:

* `bot_name` - (Required) The name of the bot.
* `name` - (Required) The name of the bot alias. The name is case sensitive.

## Attributes Reference

The following attributes are exported.

* `arn` - The ARN of the bot alias.
* `bot_name` - The name of the bot.
* `bot_version` - The version of the bot that the alias points to.
* `checksum` - Checksum of the bot alias.
* `created_date` - The date that the bot alias was created.
* `description` - A description of the alias.
* `last_updated_date` - The date that the bot alias was updated. When you create a resource, the creation date and the last updated date are the same.
* `name` - The name of the alias. The name is not case sensitive.
