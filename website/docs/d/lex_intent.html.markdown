---
subcategory: "Lex"
layout: "aws"
page_title: "AWS: aws_lex_intent"
description: |-
  Provides details about a specific Amazon Lex Intent
---

# Data Source: aws_lex_intent

Provides details about a specific Amazon Lex Intent.

## Example Usage

```hcl
data "aws_lex_intent" "order_flowers" {
  name    = "OrderFlowers"
  version = "$LATEST"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the intent. The name is case sensitive.
* `version` - (Optional) The version of the intent.

## Attributes Reference

The following attributes are exported.

* `arn` - The ARN of the Lex intent.
* `checksum` - Checksum identifying the version of the intent that was created. The checksum is not
included as an argument because the resource will add it automatically when updating the intent.
* `created_date` - The date when the intent version was created.
* `description` - A description of the intent.
* `last_updated_date` - The date when the $LATEST version of this intent was updated.
* `name` - The name of the intent, not case sensitive.
* `parent_intent_signature` - A unique identifier for the built-in intent to base this
intent on. To find the signature for an intent, see
[Standard Built-in Intents](https://developer.amazon.com/public/solutions/alexa/alexa-skills-kit/docs/built-in-intent-ref/standard-intents)
in the Alexa Skills Kit.
* `version` - The version of the bot.
