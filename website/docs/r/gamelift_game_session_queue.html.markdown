---
subcategory: "Gamelift"
layout: "aws"
page_title: "AWS: aws_gamelift_game_session_queue"
description: |-
  Provides a Gamelift Game Session Queue resource.
---

# Resource: aws_gamelift_game_session_queue

Provides an Gamelift Game Session Queue resource.

## Example Usage

```terraform
resource "aws_gamelift_game_session_queue" "test" {
  name = "example-session-queue"

  destinations = [
    aws_gamelift_fleet.us_west_2_fleet.arn,
    aws_gamelift_fleet.eu_central_1_fleet.arn,
  ]

  player_latency_policy {
    maximum_individual_player_latency_milliseconds = 100
    policy_duration_seconds                        = 5
  }

  player_latency_policy {
    maximum_individual_player_latency_milliseconds = 200
  }

  timeout_in_seconds = 60
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the session queue.
* `timeout_in_seconds` - (Required) Maximum time a game session request can remain in the queue.
* `destinations` - (Optional) List of fleet/alias ARNs used by session queue for placing game sessions.
* `player_latency_policy` - (Optional) One or more policies used to choose fleet based on player latency. See below.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Nested Fields

#### `player_latency_policy`

* `maximum_individual_player_latency_milliseconds` - (Required) Maximum latency value that is allowed for any player.
* `policy_duration_seconds` - (Optional) Length of time that the policy is enforced while placing a new game session. Absence of value for this attribute means that the policy is enforced until the queue times out.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Game Session Queue ARN.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

Gamelift Game Session Queues can be imported by their `name`, e.g.

```
$ terraform import aws_gamelift_game_session_queue.example example
```
