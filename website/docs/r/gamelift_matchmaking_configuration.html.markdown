---
subcategory: "Gamelift"
layout: "aws"
page_title: "AWS: aws_gamelift_matchmaking_configuration"
description: |-
  Provides a Gamelift Matchmaking Configuration resource.
---

# Resource: aws_gamelift_matchmaking_configuration

Provides an Gamelift Matchmaking Configuration resource.

## Example Usage

```terraform
resource "aws_gamelift_matchmaking_configuration" "test" {
  name          = "example-configuration"
  
  acceptance_required = false
  backfill_mode = "MANUAL"
  
  custom_event_data = "custom_event_data"
  game_session_data = "game_session_data"  
  
  request_timeout_seconds = 25
  
  rule_set_name = aws_gamelift_matchmaking_rule_set.test.name
  game_session_queue_arns = [aws_gamelift_game_session_queue.test.arn]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the matchmaking configuration.
* `acceptance_required` - (Required) A flag that determines whether a match that was created with this configuration must be accepted by the matched players.
* `request_timeout_seconds` - (Required) The maximum duration, in seconds, that a matchmaking ticket can remain in process before timing out. Requests that fail due to timing out can be resubmitted as needed.
* `rule_set_name` - (Required) A unique identifier for the matchmaking rule set to use with this configuration. You can use either the rule set name or ARN value. A matchmaking configuration can only use rule sets that are defined in the same Region.
* `acceptance_timeout_seconds` - (Optional) The length of time (in seconds) to wait for players to accept a proposed match, if acceptance is required.
* `additional_player_count` - (Optional) The number of player slots in a match to keep open for future players. For example, if the configuration's rule set specifies a match for a single 12-person team, and the additional player count is set to 2, only 10 players are selected for the match.
* `backfill_mode` - (Optional) The method used to backfill game sessions that are created with this matchmaking configuration. Specify MANUAL when your game manages backfill requests manually or does not use the match backfill feature. Specify AUTOMATIC to have GameLift create a StartMatchBackfill request whenever a game session has one or more open slots.
* `custom_event_data` - (Optional) Information to be added to all events related to this matchmaking configuration.
* `description` - (Optional) A human-readable description of the matchmaking configuration.
* `game_property` - (Optional) Key-value map of custom game properties. These properties are passed to a game server process in the GameSession object with a request to start a new game session (see Start a Game Session).
* `game_session_data` - (Optional) A set of custom game session properties, formatted as a single string value. This data is passed to a game server process in the GameSession object with a request to start a new game session.
* `game_session_queue_arns` - (Optional) List of the Amazon Resource Name (ARN) that are assigned to a GameLift game session queue resources. Format is arn:aws:gamelift:<region>::gamesessionqueue/<queue name>. Queues can be located in any Region.
* `notification_target` - (Optional) An SNS topic ARN that is set up to receive matchmaking notifications.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Matchmaking Configuration ARN.
* `creation_time` - Time of matchmaking configuration creation.
* `rule_set_arn` - Rule Set ARN.

## Import

Gamelift Match Making Configurations can be imported by their `name`, e.g.

```
$ terraform import aws_gamelift_matchmaking_configuration.example example
```
