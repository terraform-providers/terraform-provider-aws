---
subcategory: "Lex"
layout: "aws"
page_title: "AWS: aws_lex_bot"
description: |-
  Provides an Amazon Lex bot resource.
---

# Resource: aws_lex_bot

Provides an Amazon Lex Bot resource. For more information see
[Amazon Lex: How It Works](https://docs.aws.amazon.com/lex/latest/dg/how-it-works.html)

## Example Usage

```hcl
resource "aws_lex_bot" "order_flowers_bot" {
  abort_statement {
    message {
      content      = "Sorry, I am not able to assist at this time"
      content_type = "PlainText"
    }
  }

  child_directed = false

  clarification_prompt {
    max_attempts = 2

    message {
      content      = "I didn't understand you, what would you like to do?"
      content_type = "PlainText"
    }
  }

  create_version              = false
  description                 = "Bot to order flowers on the behalf of a user"
  idle_session_ttl_in_seconds = 600

  intent {
    intent_name    = "OrderFlowers"
    intent_version = "1"
  }

  locale           = "en-US"
  name             = "OrderFlowers"
  process_behavior = "BUILD"
  voice_id         = "Salli"
}
```

## Argument Reference

The following arguments are supported:

* `abort_statement` - (Required) The message that Amazon Lex uses to abort a conversation. Attributes are documented under [statement](#statement).
* `child_directed` - (Required) By specifying true, you confirm that your use of Amazon Lex is related to a website, program, or other application that is directed or targeted, in whole or in part, to children under age 13 and subject to COPPA. For more information see the [Amazon Lex FAQ](https://aws.amazon.com/lex/faqs#data-security) and the [Amazon Lex PutBot API Docs](https://docs.aws.amazon.com/lex/latest/dg/API_PutBot.html#lex-PutBot-request-childDirected).
* `clarification_prompt` - (Required) The message that Amazon Lex uses when it doesn't understand the user's request. Attributes are documented under [prompt](#prompt).
* `create_version` - (Optional) Determines if a new bot version is created when the initial resource is created and on each update. Defaults to `false`.
* `description` - (Optional) A description of the bot. Must be less than or equal to 200 characters in length.
* `detect_sentiment` - (Optional) When set to true user utterances are sent to Amazon Comprehend for sentiment analysis. If you don't specify detectSentiment, the default is `false`.
* `enable_model_improvements` - (Optional) Set to `true` to enable access to natural language understanding improvements. When you set the `enable_model_improvements` parameter to true you can use the `nlu_intent_confidence_threshold` parameter to configure confidence scores. For more information, see [Confidence Scores](https://docs.aws.amazon.com/lex/latest/dg/confidence-scores.html). You can only set the `enable_model_improvements` parameter in certain Regions. If you set the parameter to true, your bot has access to accuracy improvements. For more information see the [Amazon Lex Bot PutBot API Docs](https://docs.aws.amazon.com/lex/latest/dg/API_PutBot.html#lex-PutBot-request-enableModelImprovements).
* `idle_session_ttl_in_seconds` - (Optional) The maximum time in seconds that Amazon Lex retains the data gathered in a conversation. Default is `300`. Must be a number between 60 and 86400 (inclusive).
* `locale` - (Optional) Specifies the target locale for the bot. Any intent used in the bot must be compatible with the locale of the bot. For available locales, see [Amazon Lex Bot PutBot API Docs](https://docs.aws.amazon.com/lex/latest/dg/API_PutBot.html#lex-PutBot-request-locale). Default is `en-US`.
* `intent` - (Required) A set of Intent objects. Each intent represents a command that a user can express. Attributes are documented under [intent](#intent). Can have up to 100 Intent objects.
* `name` - (Required) The name of the bot that you want to create, case sensitive. Must be between 2 and 50 characters in length.
* `nlu_intent_confidence_threshold` - (Optional) Determines the threshold where Amazon Lex will insert the AMAZON.FallbackIntent, AMAZON.KendraSearchIntent, or both when returning alternative intents in a PostContent or PostText response. AMAZON.FallbackIntent and AMAZON.KendraSearchIntent are only inserted if they are configured for the bot. For more information see [Amazon Lex Bot PutBot API Docs](https://docs.aws.amazon.com/lex/latest/dg/API_PutBot.html#lex-PutBot-request-nluIntentConfidenceThreshold) This value requires `enable_model_improvements` to be set to `true` and the default is `0`. Must be a float between 0 and 1.
* `process_behavior` - (Optional) If you set the `process_behavior` element to `BUILD`, Amazon Lex builds the bot so that it can be run. If you set the element to `SAVE` Amazon Lex saves the bot, but doesn't build it. Default is `SAVE`.
* `voice_id` - (Optional) The Amazon Polly voice ID that you want Amazon Lex to use for voice interactions with the user. The locale configured for the voice must match the locale of the bot. For more information, see [Available Voices](http://docs.aws.amazon.com/polly/latest/dg/voicelist.html) in the Amazon Polly Developer Guide.

### intent

Identifies the specific version of an intent.

* `intent_name` - (Required) The name of the intent. Must be less than or equal to 100 characters in length.
* `intent_version` - (Required) The version of the intent. Must be less than or equal to 64 characters in length.

### message

The message object that provides the message text and its type.

* `content` - (Required) The text of the message.
* `content_type` - (Required) The content type of the message string.
* `group_number` - (Optional) Identifies the message group that the message belongs to. When a group
is assigned to a message, Amazon Lex returns one message from each group in the response.

### prompt

Obtains information from the user. To define a prompt, provide one or more messages and specify the
number of attempts to get information from the user. If you provide more than one message, Amazon
Lex chooses one of the messages to use to prompt the user.

* `max_attempts` - (Required) The number of times to prompt the user for information.
* `message` - (Required) A set of messages, each of which provides a message string and its type.
You can specify the message string in plain text or in Speech Synthesis Markup Language (SSML).
Attributes are documented under [message](#message).
* `response_card` - (Optional) The response card. Amazon Lex will substitute session attributes and
slot values into the response card. For more information, see
[Example: Using a Response Card](https://docs.aws.amazon.com/lex/latest/dg/ex-resp-card.html).

### statement

A statement is a map with a set of message maps and an optional response card string. Messages
convey information to the user. At runtime, Amazon Lex selects the message to convey.

* `message` - (Required) A set of messages, each of which provides a message string and its type. You
can specify the message string in plain text or in Speech Synthesis Markup Language (SSML). Attributes
are documented under [message](#message).
* `response_card` - (Optional) The response card. Amazon Lex will substitute session attributes and
slot values into the response card. For more information, see
[Example: Using a Response Card](https://docs.aws.amazon.com/lex/latest/dg/ex-resp-card.html).

### Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) for certain actions:

* `create` - (Defaults to 1 mins) Used when creating the bot
* `update` - (Defaults to 1 mins) Used when updating the bot
* `delete` - (Defaults to 5 mins) Used when deleting the bot

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `checksum` - Checksum identifying the version of the bot that was created. The checksum is not
included as an argument because the resource will add it automatically when updating the bot.
* `created_date` - The date when the bot version was created.
* `failure_reason` - If status is FAILED, Amazon Lex provides the reason that it failed to build the bot.
* `last_updated_date` - The date when the $LATEST version of this bot was updated.
* `status` - When you send a request to create or update a bot, Amazon Lex sets the status response
element to BUILDING. After Amazon Lex builds the bot, it sets status to READY. If Amazon Lex can't
build the bot, it sets status to FAILED. Amazon Lex returns the reason for the failure in the
failure_reason response element.
* `version` - The version of the bot.

## Import

Bots can be imported using their name.

```
$ terraform import aws_lex_bot.order_flowers_bot OrderFlowers
```
