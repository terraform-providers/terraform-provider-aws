package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iotevents"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func generateActionSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"clear_timer": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"firehose": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"delivery_stream_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"separator": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"iot_events": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"input_name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"iot_topic_publish": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mqtt_topic": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"lambda": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"function_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},
			"reset_timer": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"set_timer": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"seconds": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
					},
				},
			},
			"set_variable": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 128),
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"sns": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"target_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},
			"sqs": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"queue_url": {
							Type:     schema.TypeString,
							Required: true,
						},
						"use_base64": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func generateEventSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"action": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     generateActionSchema(),
			},
			"condition": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func generateTransitionEventSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"action": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     generateActionSchema(),
			},
			"condition": {
				Type:     schema.TypeString,
				Required: true,
			},
			"next_state": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func generateStateSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"on_enter": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"event": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     generateEventSchema(),
						},
					},
				},
			},
			"on_exit": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"event": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     generateEventSchema(),
						},
					},
				},
			},
			"on_input": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"event": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     generateEventSchema(),
						},
						"transition_event": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     generateTransitionEventSchema(),
						},
					},
				},
			},
		},
	}
}

func resourceAwsIotEventsDetectorModel() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIotDetectorCreate,
		Read:   resourceAwsIotDetectorRead,
		Update: resourceAwsIotDetectorUpdate,
		Delete: resourceAwsIotDetectorDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"definition": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"initial_state_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"state": {
							Type:     schema.TypeSet,
							MinItems: 1,
							Required: true,
							Elem:     generateStateSchema(),
						},
					},
				},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"tags": tagsSchema(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func parseAction(rawAction map[string]interface{}) *iotevents.ActionData {
	action := &iotevents.ActionData{}

	if v, ok := rawAction["clear_timer"].(*schema.Set); ok && len(v.List()) != 0 {

		rawClearTimer := v.List()[0].(map[string]interface{})
		action.ClearTimer = &iotevents.ClearTimerAction{
			TimerName: aws.String(rawClearTimer["name"].(string)),
		}
	}

	if v, ok := rawAction["firehose"].(*schema.Set); ok && len(v.List()) != 0 {

		rawFirehose := v.List()[0].(map[string]interface{})
		action.Firehose = &iotevents.FirehoseAction{
			DeliveryStreamName: aws.String(rawFirehose["delivery_stream_name"].(string)),
		}

		if fv, ok := rawFirehose["separator"].(string); ok && fv != "" {
			action.Firehose.Separator = aws.String(fv)
		}
	}

	if v, ok := rawAction["iot_events"].(*schema.Set); ok && len(v.List()) != 0 {

		rawIotEvents := v.List()[0].(map[string]interface{})
		action.IotEvents = &iotevents.Action{
			InputName: aws.String(rawIotEvents["input_name"].(string)),
		}
	}

	if v, ok := rawAction["iot_topic_publish"].(*schema.Set); ok && len(v.List()) != 0 {

		rawIotTopicPublish := v.List()[0].(map[string]interface{})
		action.IotTopicPublish = &iotevents.IotTopicPublishAction{
			MqttTopic: aws.String(rawIotTopicPublish["mqtt_topic"].(string)),
		}
	}

	if v, ok := rawAction["lambda"].(*schema.Set); ok && len(v.List()) != 0 {

		rawLambda := v.List()[0].(map[string]interface{})
		action.Lambda = &iotevents.LambdaAction{
			FunctionArn: aws.String(rawLambda["function_arn"].(string)),
		}
	}

	if v, ok := rawAction["reset_timer"].(*schema.Set); ok && len(v.List()) != 0 {

		rawResetTimer := v.List()[0].(map[string]interface{})
		action.ResetTimer = &iotevents.ResetTimerAction{
			TimerName: aws.String(rawResetTimer["name"].(string)),
		}
	}

	if v, ok := rawAction["set_timer"].(*schema.Set); ok && len(v.List()) != 0 {

		rawSetTimer := v.List()[0].(map[string]interface{})
		action.SetTimer = &iotevents.SetTimerAction{
			TimerName: aws.String(rawSetTimer["name"].(string)),
			Seconds:   aws.Int64(int64(rawSetTimer["seconds"].(int))),
		}
	}

	if v, ok := rawAction["set_variable"].(*schema.Set); ok && len(v.List()) != 0 {

		rawSetVariable := v.List()[0].(map[string]interface{})
		action.SetVariable = &iotevents.SetVariableAction{
			VariableName: aws.String(rawSetVariable["name"].(string)),
			Value:        aws.String(rawSetVariable["value"].(string)),
		}
	}

	if v, ok := rawAction["sns"].(*schema.Set); ok && len(v.List()) != 0 {

		rawSns := v.List()[0].(map[string]interface{})
		action.Sns = &iotevents.SNSTopicPublishAction{
			TargetArn: aws.String(rawSns["target_arn"].(string)),
		}
	}

	if v, ok := rawAction["sqs"].(*schema.Set); ok && len(v.List()) != 0 {

		rawSqs := v.List()[0].(map[string]interface{})
		action.Sqs = &iotevents.SqsAction{
			QueueUrl: aws.String(rawSqs["queue_url"].(string)),
		}

		if sqsv, ok := rawSqs["use_base64"].(bool); ok {
			action.Sqs.UseBase64 = aws.Bool(sqsv)
		}
	}

	return action
}

func parseEvent(rawEvent map[string]interface{}) *iotevents.Event {
	event := &iotevents.Event{
		EventName: aws.String(rawEvent["name"].(string)),
	}

	actions := make([]*iotevents.ActionData, 0)
	for _, rawAct := range rawEvent["action"].(*schema.Set).List() {
		actions = append(actions, parseAction(rawAct.(map[string]interface{})))
	}

	event.Actions = actions

	if v, ok := rawEvent["condition"].(string); ok && v != "" {
		event.Condition = aws.String(v)
	}

	return event
}

func parseTransitionEvent(rawTransitionEvent map[string]interface{}) *iotevents.TransitionEvent {
	event := &iotevents.TransitionEvent{
		EventName: aws.String(rawTransitionEvent["name"].(string)),
		Condition: aws.String(rawTransitionEvent["condition"].(string)),
		NextState: aws.String(rawTransitionEvent["next_state"].(string)),
	}

	actions := make([]*iotevents.ActionData, 0)
	for _, rawAct := range rawTransitionEvent["action"].(*schema.Set).List() {
		actions = append(actions, parseAction(rawAct.(map[string]interface{})))
	}

	event.Actions = actions

	return event
}

func parseEventsList(rawEvents []interface{}) []*iotevents.Event {
	events := make([]*iotevents.Event, 0)

	for _, rawEvent := range rawEvents {
		events = append(events, parseEvent(rawEvent.(map[string]interface{})))
	}

	return events
}

func parseOnEnter(rawOnEnter map[string]interface{}) *iotevents.OnEnterLifecycle {
	events := parseEventsList(rawOnEnter["event"].(*schema.Set).List())

	if len(events) != 0 {
		onEnter := &iotevents.OnEnterLifecycle{
			Events: events,
		}
		return onEnter
	}

	return nil
}

func parseOnExit(rawOnExit map[string]interface{}) *iotevents.OnExitLifecycle {
	events := parseEventsList(rawOnExit["event"].(*schema.Set).List())

	if len(events) != 0 {
		onExit := &iotevents.OnExitLifecycle{
			Events: events,
		}
		return onExit
	}
	return nil
}

func parseOnInput(rawOnInput map[string]interface{}) *iotevents.OnInputLifecycle {
	var onInput *iotevents.OnInputLifecycle

	events := parseEventsList(rawOnInput["event"].(*schema.Set).List())
	if len(events) != 0 {
		onInput = &iotevents.OnInputLifecycle{
			Events: events,
		}
	}

	transitionEvents := make([]*iotevents.TransitionEvent, 0)
	for _, rawEvent := range rawOnInput["transition_event"].(*schema.Set).List() {
		transitionEvents = append(transitionEvents, parseTransitionEvent(rawEvent.(map[string]interface{})))
	}
	if len(transitionEvents) != 0 {
		if onInput == nil {
			onInput = &iotevents.OnInputLifecycle{
				TransitionEvents: transitionEvents,
			}
		} else {
			onInput.TransitionEvents = transitionEvents
		}
	}

	if onInput != nil {
		return onInput
	}

	return nil
}

func parseState(rawState map[string]interface{}) *iotevents.State {
	state := &iotevents.State{
		StateName: aws.String(rawState["name"].(string)),
	}

	if v, ok := rawState["on_enter"].(*schema.Set); ok && len(v.List()) != 0 {
		rawOnEnter := v.List()[0].(map[string]interface{})
		state.OnEnter = parseOnEnter(rawOnEnter)
	}

	if v, ok := rawState["on_exit"].(*schema.Set); ok && len(v.List()) != 0 {
		rawOnExit := v.List()[0].(map[string]interface{})
		state.OnExit = parseOnExit(rawOnExit)
	}

	if v, ok := rawState["on_input"].(*schema.Set); ok && len(v.List()) != 0 {
		rawOnInput := v.List()[0].(map[string]interface{})
		state.OnInput = parseOnInput(rawOnInput)
	}

	return state
}

func parseDetectorModelDefinition(rawDetectorModelDefinition map[string]interface{}) *iotevents.DetectorModelDefinition {
	states := make([]*iotevents.State, 0)

	for _, rawState := range rawDetectorModelDefinition["state"].(*schema.Set).List() {
		states = append(states, parseState(rawState.(map[string]interface{})))
	}

	detectorDefinitionParams := &iotevents.DetectorModelDefinition{
		InitialStateName: aws.String(rawDetectorModelDefinition["initial_state_name"].(string)),
		States:           states,
	}

	return detectorDefinitionParams
}

func resourceAwsIotDetectorCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ioteventsconn

	detectorDefinition := d.Get("definition").(*schema.Set).List()[0].(map[string]interface{})
	detectorDefinitionParams := parseDetectorModelDefinition(detectorDefinition)

	detectorName := d.Get("name").(string)
	roleArn := d.Get("role_arn").(string)

	params := &iotevents.CreateDetectorModelInput{
		DetectorModelName:       aws.String(detectorName),
		DetectorModelDefinition: detectorDefinitionParams,
		RoleArn:                 aws.String(roleArn),
		Tags:                    keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().IoteventsTags(),
	}

	if v, ok := d.GetOk("description"); ok {
		params.DetectorModelDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("key"); ok {
		params.Key = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating IoT Model Detector: %s", params)

	retrySecondsList := [6]int{1, 2, 5, 8, 10, 0}

	var err error

	// Primitive retry.
	// During testing detector model, problem was detected.
	// When we try to create detector model and role arn that
	// will be assumed by detector during one apply we get:
	// 'Unable to assume role, role ARN' error. However if we run apply
	// second time(when all required resources are created) detector will be created successfully.
	// So we suppose that problem is that AWS return response of successful role arn creation before
	// process of creation is really ended, and then creation of detector model fails.
	for index, sleepSeconds := range retrySecondsList {
		_, err = conn.CreateDetectorModel(params)
		if err == nil {
			break
		} else if err != nil && index != len(retrySecondsList)-1 {
			err = nil
		}
		time.Sleep(time.Duration(sleepSeconds) * time.Second)
	}

	if err != nil {
		return err
	}

	d.SetId(detectorName)

	return resourceAwsIotDetectorRead(d, meta)
}

func flattenAction(action *iotevents.ActionData) map[string]interface{} {
	rawAction := make(map[string]interface{})

	if v := action.ClearTimer; v != nil {
		clearTimer := make(map[string]interface{})
		clearTimer["name"] = aws.StringValue(v.TimerName)
		rawAction["clear_timer"] = []map[string]interface{}{clearTimer}
	}

	if v := action.Firehose; v != nil {
		firehose := make(map[string]interface{})
		firehose["delivery_stream_name"] = aws.StringValue(v.DeliveryStreamName)

		if v.Separator != nil {
			firehose["separator"] = aws.StringValue(v.Separator)
		}

		rawAction["firehose"] = []map[string]interface{}{firehose}
	}

	if v := action.IotEvents; v != nil {
		iotEvents := make(map[string]interface{})
		iotEvents["input_name"] = aws.StringValue(v.InputName)
		rawAction["iot_events"] = []map[string]interface{}{iotEvents}
	}

	if v := action.IotTopicPublish; v != nil {
		iotTopicPublish := make(map[string]interface{})
		iotTopicPublish["mqtt_topic"] = aws.StringValue(v.MqttTopic)
		rawAction["iot_topic_publish"] = []map[string]interface{}{iotTopicPublish}
	}

	if v := action.Lambda; v != nil {
		lambda := make(map[string]interface{})
		lambda["function_arn"] = aws.StringValue(v.FunctionArn)
		rawAction["lambda"] = []map[string]interface{}{lambda}
	}

	if v := action.ResetTimer; v != nil {
		resetTimer := make(map[string]interface{})
		resetTimer["name"] = aws.StringValue(v.TimerName)
		rawAction["reset_timer"] = []map[string]interface{}{resetTimer}
	}

	if v := action.SetTimer; v != nil {
		setTimer := make(map[string]interface{})
		setTimer["name"] = aws.StringValue(v.TimerName)
		setTimer["seconds"] = aws.Int64Value(v.Seconds)
		rawAction["set_timer"] = []map[string]interface{}{setTimer}
	}

	if v := action.SetVariable; v != nil {
		setVariable := make(map[string]interface{})
		setVariable["name"] = aws.StringValue(v.VariableName)
		setVariable["value"] = aws.StringValue(v.Value)
		rawAction["set_variable"] = []map[string]interface{}{setVariable}
	}

	if v := action.Sns; v != nil {
		sns := make(map[string]interface{})
		sns["target_arn"] = aws.StringValue(v.TargetArn)
		rawAction["sns"] = []map[string]interface{}{sns}
	}

	if v := action.Sqs; v != nil {
		sqs := make(map[string]interface{})
		sqs["queue_url"] = aws.StringValue(v.QueueUrl)

		if v.UseBase64 != nil {
			sqs["use_base64"] = aws.BoolValue(v.UseBase64)
		}

		rawAction["sqs"] = []map[string]interface{}{sqs}
	}

	return rawAction
}

func flattenEvent(event *iotevents.Event) map[string]interface{} {
	rawEvent := make(map[string]interface{})
	rawEvent["name"] = aws.StringValue(event.EventName)

	if event.Condition != nil {
		rawEvent["condition"] = aws.StringValue(event.Condition)
	}

	rawActions := make([]map[string]interface{}, 0)
	for _, act := range event.Actions {
		rawActions = append(rawActions, flattenAction(act))
	}

	rawEvent["action"] = rawActions

	return rawEvent
}

func flattenTransitionEvent(transitionEvent *iotevents.TransitionEvent) map[string]interface{} {
	rawTransitionEvent := make(map[string]interface{})
	rawTransitionEvent["name"] = aws.StringValue(transitionEvent.EventName)
	rawTransitionEvent["condition"] = aws.StringValue(transitionEvent.Condition)
	rawTransitionEvent["next_state"] = aws.StringValue(transitionEvent.NextState)

	rawActions := make([]map[string]interface{}, 0)
	for _, act := range transitionEvent.Actions {
		rawActions = append(rawActions, flattenAction(act))
	}

	rawTransitionEvent["action"] = rawActions

	return rawTransitionEvent
}

func flattenOnEnter(onEnter *iotevents.OnEnterLifecycle) map[string]interface{} {
	rawEvents := make([]map[string]interface{}, 0)
	for _, event := range onEnter.Events {
		rawEvents = append(rawEvents, flattenEvent(event))
	}

	rawOnEnter := make(map[string]interface{})
	rawOnEnter["event"] = rawEvents
	return rawOnEnter
}

func flattenOnExit(onExit *iotevents.OnExitLifecycle) map[string]interface{} {
	rawEvents := make([]map[string]interface{}, 0)
	for _, event := range onExit.Events {
		rawEvents = append(rawEvents, flattenEvent(event))
	}

	rawOnExit := make(map[string]interface{})
	rawOnExit["event"] = rawEvents
	return rawOnExit
}

func flattenOnInput(onInput *iotevents.OnInputLifecycle) map[string]interface{} {
	rawOnInput := make(map[string]interface{})

	rawEvents := make([]map[string]interface{}, 0)
	for _, event := range onInput.Events {
		rawEvents = append(rawEvents, flattenEvent(event))
	}
	rawOnInput["event"] = rawEvents

	rawTransitionEvents := make([]map[string]interface{}, 0)
	for _, transitionEvent := range onInput.TransitionEvents {
		rawTransitionEvents = append(rawTransitionEvents, flattenTransitionEvent(transitionEvent))
	}
	rawOnInput["transition_event"] = rawTransitionEvents

	return rawOnInput
}

func flattenState(state *iotevents.State) map[string]interface{} {
	rawState := make(map[string]interface{})

	rawState["name"] = aws.StringValue(state.StateName)
	rawState["on_enter"] = []map[string]interface{}{flattenOnEnter(state.OnEnter)}
	rawState["on_exit"] = []map[string]interface{}{flattenOnExit(state.OnExit)}
	rawState["on_input"] = []map[string]interface{}{flattenOnInput(state.OnInput)}

	return rawState
}

func flattenDetectorModelDefinition(detectorModelDefinition *iotevents.DetectorModelDefinition) map[string]interface{} {
	rawStates := make([]interface{}, 0)
	for _, state := range detectorModelDefinition.States {
		rawStates = append(rawStates, flattenState(state))
	}

	rawDetectorModelDefinition := make(map[string]interface{})
	rawDetectorModelDefinition["state"] = rawStates
	rawDetectorModelDefinition["initial_state_name"] = aws.StringValue(detectorModelDefinition.InitialStateName)

	return rawDetectorModelDefinition
}

func resourceAwsIotDetectorRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ioteventsconn

	params := &iotevents.DescribeDetectorModelInput{
		DetectorModelName: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Reading IoT Events Detector Model: %s", params)
	out, err := conn.DescribeDetectorModel(params)

	if err != nil {
		return err
	}

	d.Set("name", out.DetectorModel.DetectorModelConfiguration.DetectorModelName)
	d.Set("description", out.DetectorModel.DetectorModelConfiguration.DetectorModelDescription)
	d.Set("key", out.DetectorModel.DetectorModelConfiguration.Key)
	d.Set("role_arn", out.DetectorModel.DetectorModelConfiguration.RoleArn)
	detectorModelDefinition := []map[string]interface{}{flattenDetectorModelDefinition(out.DetectorModel.DetectorModelDefinition)}
	d.Set("definition", detectorModelDefinition)
	d.Set("arn", out.DetectorModel.DetectorModelConfiguration.DetectorModelArn)

	arn := *out.DetectorModel.DetectorModelConfiguration.DetectorModelArn
	tags, err := keyvaluetags.IoteventsListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for resource (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsIotDetectorUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ioteventsconn

	detectorDefinition := d.Get("definition").(*schema.Set).List()[0].(map[string]interface{})
	detectorDefinitionParams := parseDetectorModelDefinition(detectorDefinition)

	detectorName := d.Get("name").(string)
	roleArn := d.Get("role_arn").(string)

	params := &iotevents.UpdateDetectorModelInput{
		DetectorModelName:       aws.String(detectorName),
		DetectorModelDefinition: detectorDefinitionParams,
		RoleArn:                 aws.String(roleArn),
	}

	if v, ok := d.GetOk("description"); ok {
		params.DetectorModelDescription = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Updating IoT Events Detector Model: %s", params)

	retrySecondsList := [6]int{1, 2, 5, 8, 10, 0}
	var err error
	// Primitive retry.
	// Full explanation can be found in function `resourceAwsIotDetectorCreate`.
	// We suppose that such error can appear during update also, if you update
	// role arn.
	for index, sleepSeconds := range retrySecondsList {
		_, err = conn.UpdateDetectorModel(params)
		if err == nil {
			break
		} else if err != nil && index != len(retrySecondsList)-1 {
			err = nil
		}
		time.Sleep(time.Duration(sleepSeconds) * time.Second)
	}

	if err != nil {
		return err
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.IoteventsUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsIotDetectorRead(d, meta)
}

func resourceAwsIotDetectorDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ioteventsconn

	params := &iotevents.DeleteDetectorModelInput{
		DetectorModelName: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Deleting IoT Events Detector Model: %s", params)
	_, err := conn.DeleteDetectorModel(params)

	return err
}
