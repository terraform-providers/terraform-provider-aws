---
subcategory: "Lex"
layout: "aws"
page_title: "AWS: aws_lex_intent"
description: |-
  Provides an Amazon Lex intent resource.
---

# Resource: aws_lex_intent

Provides an Amazon Lex Intent resource. For more information see
[Amazon Lex: How It Works](https://docs.aws.amazon.com/lex/latest/dg/how-it-works.html)

## Example Usage

```hcl
resource "aws_lex_intent" "order_flowers_intent" {
  confirmation_prompt {
    max_attempts = 2

    message {
      content      = "Okay, your {FlowerType} will be ready for pickup by {PickupTime} on {PickupDate}.  Does this sound okay?"
      content_type = "PlainText"
    }
  }

  create_version = false
  description    = "Intent to order a bouquet of flowers for pick up"

  fulfillment_activity {
    type = "ReturnIntent"
  }

  rejection_statement {
    message {
      content      = "Okay, I will not place your order."
      content_type = "PlainText"
    }
  }

  sample_utterances = [
    "I would like to order some flowers",
    "I would like to pick up flowers",
  ]

  slot {
    description = "The type of flowers to pick up"
    name        = "FlowerType"
    priority    = 1

    sample_utterances = [
      "I would like to order {FlowerType}",
    ]

    slot_constraint   = "Required"
    slot_type         = "FlowerTypes"
    slot_type_version = "$$LATEST"

    value_elicitation_prompt {
      max_attempts = 2

      message {
        content      = "What type of flowers would you like to order?"
        content_type = "PlainText"
      }
    }
  }

  slot {
    description = "The date to pick up the flowers"
    name        = "PickupDate"
    priority    = 2

    sample_utterances = [
      "I would like to order {FlowerType}",
    ]

    slot_constraint   = "Required"
    slot_type         = "AMAZON.DATE"
    slot_type_version = "$$LATEST"

    value_elicitation_prompt {
      max_attempts = 2

      message {
        content      = "What day do you want the {FlowerType} to be picked up?"
        content_type = "PlainText"
      }
    }
  }

  slot {
    description = "The time to pick up the flowers"
    name        = "PickupTime"
    priority    = 3

    sample_utterances = [
      "I would like to order {FlowerType}",
    ]

    slot_constraint   = "Required"
    slot_type         = "AMAZON.TIME"
    slot_type_version = "$$LATEST"

    value_elicitation_prompt {
      max_attempts = 2

      message {
        content      = "Pick up the {FlowerType} at what time on {PickupDate}?"
        content_type = "PlainText"
      }
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `conclusion_statement` - (Optional) The statement that you want Amazon Lex to convey to the user
after the intent is successfully fulfilled by the Lambda function. This element is relevant only if
you provide a Lambda function in the `fulfillment_activity`. If you return the intent to the client
application, you can't specify this element. The `follow_up_prompt` and `conclusion_statement` are
mutually exclusive. You can specify only one. Attributes are documented under [statement](#statement).
* `confirmation_prompt` - (Optional) Prompts the user to confirm the intent. This question should
have a yes or no answer. You you must provide both the `rejection_statement` and `confirmation_prompt`,
or neither. Attributes are documented under [prompt](#prompt).
* `create_version` - (Optional) Determines if a new slot type version is created when the initial
resource is created and on each update. Defaults to `false`.
* `description` - (Optional) A description of the intent. Must be less than or equal to 200 characters in length.
* `dialog_code_hook` - (Optional) Specifies a Lambda function to invoke for each user input. You can
invoke this Lambda function to personalize user interaction. Attributes are documented under [code_hook](#code_hook).
* `follow_up_prompt` - (Optional) Amazon Lex uses this prompt to solicit additional activity after
fulfilling an intent. For example, after the OrderPizza intent is fulfilled, you might prompt the
user to order a drink. The `follow_up_prompt` field and the `conclusion_statement` field are mutually
exclusive. You can specify only one. Attributes are documented under [follow_up_prompt](#follow_up_prompt).
* `fulfillment_activity` - (Required) Describes how the intent is fulfilled. For example, after a
user provides all of the information for a pizza order, `fulfillment_activity` defines how the bot
places an order with a local pizza store. Attributes are documented under [fulfillment_activity](#fulfillment_activity).
* `name` - (Required) The name of the intent, not case sensitive. Must be less than or equal to 100 characters in length.
* `parent_intent_signature` - (Optional) A unique identifier for the built-in intent to base this
intent on. To find the signature for an intent, see
[Standard Built-in Intents](https://developer.amazon.com/public/solutions/alexa/alexa-skills-kit/docs/built-in-intent-ref/standard-intents)
in the Alexa Skills Kit.
* `rejection_statement` - (Optional) When the user answers "no" to the question defined in
`confirmation_prompt`, Amazon Lex responds with this statement to acknowledge that the intent was
canceled. You must provide both the `rejection_statement` and the `confirmation_prompt`, or neither.
Attributes are documented under [statement](#statement).
* `sample_utterances` - (Optional) An array of utterances (strings) that a user might say to signal
the intent. For example, "I want {PizzaSize} pizza", "Order {Quantity} {PizzaSize} pizzas".
In each utterance, a slot name is enclosed in curly braces. Must have between 1 and 10 items in the list, and each item must be less than or equal to 200 characters in length.
* `slot` - (Optional) An list of intent slots. At runtime, Amazon Lex elicits required slot values
from the user using prompts defined in the slots. Attributes are documented under [slot](#slot).

### code_hook

Specifies a Lambda function that verifies requests to a bot or fulfills the user's request to a bot.

* `message_version` - (Required) The version of the request-response that you want Amazon Lex to use
to invoke your Lambda function. For more information, see
[Using Lambda Functions](https://docs.aws.amazon.com/lex/latest/dg/using-lambda.html). Must be less than or equal to 5 characters in length.
* `uri` - (Required) The Amazon Resource Name (ARN) of the Lambda function.

### follow_up_prompt

A prompt for additional activity after an intent is fulfilled. For example, after the OrderPizza
intent is fulfilled, you might prompt the user to find out whether the user wants to order drinks.

* `prompt` - (Required) Prompts for information from the user. Attributes are documented under [prompt](#prompt).
* `rejection_statement` - (Optional) If the user answers "no" to the question defined in the prompt field,
Amazon Lex responds with this statement to acknowledge that the intent was canceled. Attributes are
documented below under [statement](#statement).

### fulfillment_activity

Describes how the intent is fulfilled after the user provides all of the information required for the intent.

* `type` - (Required) How the intent should be fulfilled, either by running a Lambda function or by
returning the slot data to the client application.
* `code_hook` - (Optional) A description of the Lambda function that is run to fulfill the intent.
Required if type is CodeHook. Attributes are documented under [code_hook](#code_hook).

### message

The message object that provides the message text and its type.

* `content` - (Required) The text of the message. Must be less than or equal to 1000 characters in length.
* `content_type` - (Required) The content type of the message string.
* `group_number` - (Optional) Identifies the message group that the message belongs to. When a group
is assigned to a message, Amazon Lex returns one message from each group in the response. Must be a number between 1 and 5 (inclusive).

### prompt

Obtains information from the user. To define a prompt, provide one or more messages and specify the
number of attempts to get information from the user. If you provide more than one message, Amazon
Lex chooses one of the messages to use to prompt the user.

* `max_attempts` - (Required) The number of times to prompt the user for information. Must be a number between 1 and 5 (inclusive).
* `message` - (Required) A set of messages, each of which provides a message string and its type.
You can specify the message string in plain text or in Speech Synthesis Markup Language (SSML).
Attributes are documented under [message](#message). Must contain between 1 and 15 messages.
* `response_card` - (Optional) The response card. Amazon Lex will substitute session attributes and
slot values into the response card. For more information, see
[Example: Using a Response Card](https://docs.aws.amazon.com/lex/latest/dg/ex-resp-card.html). Must be less than or equal to 50000 characters in length.

### slot

Identifies the version of a specific slot.

* `name` - (Required) The name of the intent slot that you want to create. The name is case sensitive. Must be less than or equal to 100 characters in length.
* `slot_constraint` - (Required) Specifies whether the slot is required or optional.
* `description` - (Optional) A description of the bot. Must be less than or equal to 200 characters in length.
* `priority` - (Optional) Directs Lex the order in which to elicit this slot value from the user.
For example, if the intent has two slots with priorities 1 and 2, AWS Lex first elicits a value for
the slot with priority 1. If multiple slots share the same priority, the order in which Lex elicits
values is arbitrary. Must be between 1 and 100.
* `response_card` - (Optional) The response card. Amazon Lex will substitute session attributes and
slot values into the response card. For more information, see
[Example: Using a Response Card](https://docs.aws.amazon.com/lex/latest/dg/ex-resp-card.html). Must be less than or equal to 50000 characters in length.
* `sample_utterances` - (Optional) If you know a specific pattern with which users might respond to
an Amazon Lex request for a slot value, you can provide those utterances to improve accuracy. This
is optional. In most cases, Amazon Lex is capable of understanding user utterances. Must have between 1 and 10 items in the list, and each item must be less than or equal to 200 characters in length.
* `slot_type` - (Optional) The type of the slot, either a custom slot type that you defined or one of
the built-in slot types. Must be less than or equal to 100 characters in length.
* `slot_type_version` - (Optional) The version of the slot type. Must be less than or equal to 64 characters in length.
* `value_elicitation_prompt` - (Optional) The prompt that Amazon Lex uses to elicit the slot value
from the user. Attributes are documented under [prompt](#prompt).

### statement

A statement is a map with a set of message maps and an optional response card string. Messages
convey information to the user. At runtime, Amazon Lex selects the message to convey.

* `message` - (Required) A set of messages, each of which provides a message string and its type.
You can specify the message string in plain text or in Speech Synthesis Markup Language (SSML).
Attributes are documented under [message](#message). Must contain between 1 and 15 messages.
* `response_card` - (Optional) The response card. Amazon Lex will substitute session attributes and
slot values into the response card. For more information, see
[Example: Using a Response Card](https://docs.aws.amazon.com/lex/latest/dg/ex-resp-card.html). Must be less than or equal to 50000 characters in length.

### Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) for certain actions:

* `create` - (Defaults to 1 min) Used when creating the intent
* `update` - (Defaults to 1 min) Used when updating the intent
* `delete` - (Defaults to 5 mins) Used when deleting the intent


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the Lex intent.
* `checksum` - Checksum identifying the version of the intent that was created. The checksum is not
included as an argument because the resource will add it automatically when updating the intent.
* `created_date` - The date when the intent version was created.
* `last_updated_date` - The date when the $LATEST version of this intent was updated.
* `version` - The version of the bot.

## Import

Intents can be imported using their name.

```
$ terraform import aws_lex_intent.order_flowers_intent OrderFlowers
```
