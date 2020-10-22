---
subcategory: "Lex"
layout: "aws"
page_title: "AWS: aws_lex_bot"
description: |-
  Provides details about a specific Lex Bot
---

# Data Source: aws_lex_bot

Provides details about a specific Amazon Lex Bot.

## Example Usage

```hcl
data "aws_lex_bot" "order_flowers_bot" {
  name    = "OrderFlowers"
  version = "$LATEST"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the bot. The name is case sensitive.
* `version` - (Optional) The version or alias of the bot.

## Attributes Reference

The following attributes are exported.

* `child_directed` - Specifies if this Amazon Lex Bot is related to a website, program, or other application that is directed or targeted, in whole or in part, to children under age 13 and subject to COPPA.
* `description` - A description of the bot.
* `detect_sentiment` - When set to true user utterances are sent to Amazon Comprehend for sentiment analysis.
* `enable_model_improvements` - Set to true if natural language understanding improvements are enabled.
* `idle_session_ttl_in_seconds` - The maximum time in seconds that Amazon Lex retains the data gathered in a conversation.
* `locale` - Specifies the target locale for the bot. Any intent used in the bot must be compatible with the locale of the bot.
* `name` - The name of the bot, case sensitive.
* `nlu_intent_confidence_threshold` - The threshold where Amazon Lex will insert the AMAZON.FallbackIntent, AMAZON.KendraSearchIntent, or both when returning alternative intents in a PostContent or PostText response. AMAZON.FallbackIntent and AMAZON.KendraSearchIntent are only inserted if they are configured for the bot.
* `voice_id` - The Amazon Polly voice ID that the Amazon Lex Bot uses for voice interactions with the user.
