---
subcategory: "IoTEvents"
layout: "aws"
page_title: "AWS: aws_iotevents_detector_model"
sidebar_current: "docs-aws-resource-iotevents-detector-model"
description: |-
    Creates and manages an AWS IoTEvents detector model
---

# Resource: aws_iotevents_detector_model

## Example Usage

```hcl
resource "aws_iotevents_detector_model" "detector" {
  name        = "detector_name"
  role_arn    = "role_arn"

  definition {
    initial_state_name = "state_1"

    state {
      name = "state_1"
      on_enter {
        event {
          name      = "event_name"
          condition = "convert(Decimal, $input.some_sensor.temperature) > 20"

          action {
            clear_timer {
              name = "timer_name"
            }
          }
        }
      }
    }
  }
}
```

## Argument Reference

* `name` - (Required) The name of the detector.
* `description` - (Optional) The description of the detector.
* `key` - (Optional) The input attribute key used to identify a device or system in order to 
create a detector (an instance of the detector model) and then to route each input received to 
the appropriate detector (instance). This parameter uses a JSON-path expression to specify the 
attribute-value pair in the message payload of each input that is used to identify the device associated with 
the input.
* `role_arn` - (Required) .The ARN of the role that grants permission to AWS IoT Events to perform its operations.
* `tags` - (Optional) Map. Map of tags. Metadata that can be used to manage the detector model.

The `definition` object describes information that defines how the detectors operate. It takes such arguments:
* `initial_state_name` - (Required) The state that is entered at the creation of each detector.
* [`state`](#state) - Object (At least 1). The `state` object describes information about state of detector

<a name="state"><a/> The `state` argument reference.
* `name` - (Required) The name of the state.
* [`on_enter`](#on_enter) - Object (Optional) When entering this state, perform these "actions" if the "condition" is TRUE.
* [`on_exit`](#on_exit) - Object (Optional) When exiting this state, perform these "actions" if the specified "condition" is TRUE.
* [`on_input`](#on_input) - Object (Optional) When an input is received and the "condition" is TRUE, perform the specified "actions"

<a name="on_enter"><a/> The `on_enter` argument reference
* [`event`](#event) - Object (Optional) Specifies the actions that are performed when the state is entered and the "condition" is TRUE.

<a name="on_exit"><a/> The `on_exit` argument reference
* [`event`](#event) - Object (Optional) Specifies the actions that are performed when the state is entered and the "condition" is TRUE.

<a name="on_input"><a/> The `on_input` argument reference
* [`event`](#event) - Object (Optional) Specifies the actions that are performed when the state is entered and the "condition" is TRUE.
* [`transition_event`](#transition_event)` - Object (Optional) Specifies the actions performed, and the next state entered, when a "condition" evaluates to TRUE.

<a name="event"><a/> The `event` argument reference
* `name` - (Required) The name of the event.
* `condition` - (Optional) The Boolean expression that when TRUE causes the "actions" to be performed. 
If not present, the actions are performed (=TRUE); if the expression result is not a Boolean value, the actions are NOT performed (=FALSE).
* `action` - Object (Optional) The action to be performed.

<a name="transition_event"><a/>  The `transition_event` argument reference
* `name` - (Required) The name of the event.
* `condition` - (Required) The Boolean expression that when TRUE causes the "actions" to be performed. 
If not present, the actions are performed (=TRUE); if the expression result is not a Boolean value, the actions are NOT performed (=FALSE).
* `action` - Object (Optional) The action to be performed.
* `next_state` - (Required) The next state to enter.

The `action` argument reference
* [`clear_timer`](#clear_timer) - Object (Optional) Information needed to clear the timer.
* [`firehose`](#) - Object (Optional) Sends information about the detector model instance and the event which triggered
the action to a Kinesis Data Firehose stream.
* [`iot_events`](#) - Object (Optional) Sends an IoT Events input, passing in information about the detector model 
instance and the event which triggered the action.
* [`iot_topic_publish`](#) - Object (Optional) Publishes an MQTT message with the given topic to the AWS IoT message broker.
* [`lambda`](#) - Object (Optional) Calls a Lambda function, passing in information about the detector model
instance and the event which triggered the action.
* [`reset_timer`](#) - Object (Optional) Information needed to reset the timer.
* [`set_timer`](#) - Object (Optional) Information needed to set the timer.
* [`set_variable`](#) - Object (Optional) Sets a variable to a specified value.
* [`sns`](#) - Object (Optional) Sends an Amazon SNS message.
* [`sqs`](#) - Object (Optional) Sends information about the detector model instance and the event which triggered
the action to an AWS SQS queue.

<a name="clear_timer"><a/>  The `clear_timer` argument reference
* `name` - (Required) The name of the timer to clear.

<a name="firehose"><a/>  The `firehose` argument reference
* `delivery_stream_name` - (Required) The name of the Kinesis Data Firehose stream where the data is written.
* `separator` - (Optional) A character separator that is used to separate records written to the Kinesis 
Data Firehose stream. Valid values are: '\n' (newline), '\t' (tab), '\r\n' (Windows newline), ',' (comma).

<a name="iot_events"><a/>  The `iot_events` argument reference
* `name` - (Required) The name of the AWS IoT Events input where the data is sent.

<a name="iot_topic_publish"><a/>  The `iot_topic_publish` argument reference
* `mqtt_topic - (Required) The MQTT topic of the message.

<a name="lambda"><a/>  The `lambda` argument reference
* `function_arn` - (Required) The ARN of the Lambda function which is executed.

<a name="reset_timer"><a/>  The `reset_timer` argument reference
* `name` - (Required) The name of the timer to reset.

<a name="set_timer"><a/>  The `set_timer` argument reference
* `name` - (Required) The name of the timer.
* `seconds` - (Required) The number of seconds until the timer expires. The minimum value is 60 seconds to ensure accuracy.

<a name="set_variable"><a/>  The `set_variable` argument reference
* `name` - (Required) The name of the variable.
* `value` - (Required) The new value of the variable.

<a name="sns"><a/>  The `sns` argument reference
* `target_arn` - (Required) The ARN of the Amazon SNS target where the message is sent

<a name="sqs"><a/>  The `sqs` argument reference
* `queue_url` - (Required) The URL of the SQS queue where the data is written.
* `use_base64` - (Optional) Set this to TRUE if you want the data to be Base-64 encoded before it is written to the queue. Otherwise, set this to FALSE.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the detector model.

## Import

IoTEvents Detector Model can be imported using the `name`, e.g.

```
$ terraform import aws_iotevents_detector_model.detector <name>
```
