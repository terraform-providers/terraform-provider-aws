package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/lex/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

const (
	LexIntentCreateTimeout = 1 * time.Minute
	LexIntentUpdateTimeout = 1 * time.Minute
	LexIntentDeleteTimeout = 5 * time.Minute
	LexIntentVersionLatest = "$LATEST"
)

func resourceAwsLexIntent() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLexIntentCreate,
		Read:   resourceAwsLexIntentRead,
		Update: resourceAwsLexIntentUpdate,
		Delete: resourceAwsLexIntentDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(LexIntentCreateTimeout),
			Update: schema.DefaultTimeout(LexIntentUpdateTimeout),
			Delete: schema.DefaultTimeout(LexIntentDeleteTimeout),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"checksum": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"conclusion_statement": {
				Type:          schema.TypeList,
				Optional:      true,
				MinItems:      1,
				MaxItems:      1,
				ConflictsWith: []string{"follow_up_prompt"},
				Elem:          lexStatementResource,
			},
			"confirmation_prompt": {
				Type:         schema.TypeList,
				Optional:     true,
				MinItems:     1,
				MaxItems:     1,
				RequiredWith: []string{"rejection_statement"},
				Elem:         lexPromptResource,
			},
			"create_version": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 200),
			},
			"dialog_code_hook": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem:     lexCodeHookResource,
			},
			"follow_up_prompt": {
				Type:          schema.TypeList,
				Optional:      true,
				MinItems:      1,
				MaxItems:      1,
				ConflictsWith: []string{"conclusion_statement"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"prompt": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							MaxItems: 1,
							Elem:     lexPromptResource,
						},
						"rejection_statement": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							MaxItems: 1,
							Elem:     lexStatementResource,
						},
					},
				},
			},
			"fulfillment_activity": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"code_hook": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							MaxItems: 1,
							Elem:     lexCodeHookResource,
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(lexmodelbuildingservice.FulfillmentActivityType_Values(), false),
						},
					},
				},
			},
			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 100),
					validation.StringMatch(regexp.MustCompile(`^([A-Za-z]_?)+$`), ""),
				),
			},
			"parent_intent_signature": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"rejection_statement": {
				Type:         schema.TypeList,
				Optional:     true,
				MinItems:     1,
				MaxItems:     1,
				RequiredWith: []string{"confirmation_prompt"},
				Elem:         lexStatementResource,
			},
			"sample_utterances": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 0,
				MaxItems: 1500,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(1, 200),
				},
			},
			"slot": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 0,
				MaxItems: 100,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "",
							ValidateFunc: validation.StringLenBetween(0, 200),
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 100),
								validation.StringMatch(regexp.MustCompile(`^([A-Za-z]_?)+$`), ""),
							),
						},
						"priority": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      0,
							ValidateFunc: validation.IntBetween(0, 100),
						},
						"response_card": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 50000),
						},
						"sample_utterances": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							MaxItems: 10,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(1, 200),
							},
						},
						"slot_constraint": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(lexmodelbuildingservice.SlotConstraint_Values(), false),
						},
						"slot_type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 100),
								validation.StringMatch(regexp.MustCompile(`^((AMAZON\.)_?|[A-Za-z]_?)+`), ""),
							),
						},
						"slot_type_version": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 64),
								validation.StringMatch(regexp.MustCompile(`\$LATEST|[0-9]+`), ""),
							),
						},
						"value_elicitation_prompt": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							MaxItems: 1,
							Elem:     lexPromptResource,
						},
					},
				},
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsLexIntentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn
	name := d.Get("name").(string)

	input := &lexmodelbuildingservice.PutIntentInput{
		CreateVersion: aws.Bool(d.Get("create_version").(bool)),
		Description:   aws.String(d.Get("description").(string)),
		Name:          aws.String(name),
	}

	if v, ok := d.GetOk("conclusion_statement"); ok {
		input.ConclusionStatement = expandLexStatement(v)
	}

	if v, ok := d.GetOk("confirmation_prompt"); ok {
		input.ConfirmationPrompt = expandLexPrompt(v)
	}

	if v, ok := d.GetOk("dialog_code_hook"); ok {
		input.DialogCodeHook = expandLexCodeHook(v)
	}

	if v, ok := d.GetOk("follow_up_prompt"); ok {
		input.FollowUpPrompt = expandLexFollowUpPrompt(v)
	}

	if v, ok := d.GetOk("fulfillment_activity"); ok {
		input.FulfillmentActivity = expandLexFulfilmentActivity(v)
	}

	if v, ok := d.GetOk("parent_intent_signature"); ok {
		input.ParentIntentSignature = aws.String(v.(string))
	}

	if v, ok := d.GetOk("rejection_statement"); ok {
		input.RejectionStatement = expandLexStatement(v)
	}

	if v, ok := d.GetOk("sample_utterances"); ok {
		input.SampleUtterances = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("slot"); ok {
		input.Slots = expandLexSlots(v.(*schema.Set).List())
	}

	err := resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		output, err := conn.PutIntent(input)

		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeConflictException) {
			input.Checksum = output.Checksum
			return resource.RetryableError(fmt.Errorf("%q intent still creating, another operation is pending: %s", d.Id(), err))
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.PutIntent(input)
	}

	if err != nil {
		return fmt.Errorf("error creating intent %s: %w", name, err)
	}

	d.SetId(name)

	return resourceAwsLexIntentRead(d, meta)
}

func resourceAwsLexIntentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	resp, err := conn.GetIntent(&lexmodelbuildingservice.GetIntentInput{
		Name:    aws.String(d.Id()),
		Version: aws.String(LexIntentVersionLatest),
	})
	if isAWSErr(err, lexmodelbuildingservice.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] Intent (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error getting intent %s: %w", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "lex",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("intent:%s", d.Id()),
	}
	d.Set("arn", arn.String())

	d.Set("checksum", resp.Checksum)
	d.Set("created_date", resp.CreatedDate.Format(time.RFC3339))
	d.Set("description", resp.Description)
	d.Set("last_updated_date", resp.LastUpdatedDate.Format(time.RFC3339))
	d.Set("name", resp.Name)

	version, err := getLatestLexIntentVersion(conn, &lexmodelbuildingservice.GetIntentVersionsInput{
		Name: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("error reading version of intent %s: %w", d.Id(), err)
	}
	d.Set("version", version)

	if resp.ConclusionStatement != nil {
		d.Set("conclusion_statement", flattenLexStatement(resp.ConclusionStatement))
	}

	if resp.ConfirmationPrompt != nil {
		d.Set("confirmation_prompt", flattenLexPrompt(resp.ConfirmationPrompt))
	}

	if resp.DialogCodeHook != nil {
		d.Set("dialog_code_hook", flattenLexCodeHook(resp.DialogCodeHook))
	}

	if resp.FollowUpPrompt != nil {
		d.Set("follow_up_prompt", flattenLexFollowUpPrompt(resp.FollowUpPrompt))
	}

	if resp.FulfillmentActivity != nil {
		d.Set("fulfillment_activity", flattenLexFulfilmentActivity(resp.FulfillmentActivity))
	}

	if resp.ParentIntentSignature != nil {
		d.Set("parent_intent_signature", resp.ParentIntentSignature)
	}

	if resp.RejectionStatement != nil {
		d.Set("rejection_statement", flattenLexStatement(resp.RejectionStatement))
	}

	if resp.SampleUtterances != nil {
		d.Set("sample_utterances", resp.SampleUtterances)
	}

	if resp.Slots != nil {
		d.Set("slot", flattenLexSlots(resp.Slots))
	}

	return nil
}

func resourceAwsLexIntentUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	input := &lexmodelbuildingservice.PutIntentInput{
		Checksum:      aws.String(d.Get("checksum").(string)),
		CreateVersion: aws.Bool(d.Get("create_version").(bool)),
		Description:   aws.String(d.Get("description").(string)),
		Name:          aws.String(d.Id()),
	}

	if v, ok := d.GetOk("conclusion_statement"); ok {
		input.ConclusionStatement = expandLexStatement(v)
	}

	if v, ok := d.GetOk("confirmation_prompt"); ok {
		input.ConfirmationPrompt = expandLexPrompt(v)
	}

	if v, ok := d.GetOk("dialog_code_hook"); ok {
		input.DialogCodeHook = expandLexCodeHook(v)
	}

	if v, ok := d.GetOk("follow_up_prompt"); ok {
		input.FollowUpPrompt = expandLexFollowUpPrompt(v)
	}

	if v, ok := d.GetOk("fulfillment_activity"); ok {
		input.FulfillmentActivity = expandLexFulfilmentActivity(v)
	}

	if v, ok := d.GetOk("parent_intent_signature"); ok {
		input.ParentIntentSignature = aws.String(v.(string))
	}

	if v, ok := d.GetOk("rejection_statement"); ok {
		input.RejectionStatement = expandLexStatement(v)
	}

	if v, ok := d.GetOk("sample_utterances"); ok {
		input.SampleUtterances = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("slot"); ok {
		input.Slots = expandLexSlots(v.(*schema.Set).List())
	}

	err := resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err := conn.PutIntent(input)

		if isAWSErr(err, lexmodelbuildingservice.ErrCodeConflictException, "") {
			return resource.RetryableError(fmt.Errorf("%q: intent still updating", d.Id()))
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.PutIntent(input)
	}

	if err != nil {
		return fmt.Errorf("error updating intent %s: %w", d.Id(), err)
	}

	return resourceAwsLexIntentRead(d, meta)
}

func resourceAwsLexIntentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	input := &lexmodelbuildingservice.DeleteIntentInput{
		Name: aws.String(d.Id()),
	}

	err := resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := conn.DeleteIntent(input)

		if isAWSErr(err, lexmodelbuildingservice.ErrCodeConflictException, "") {
			return resource.RetryableError(fmt.Errorf("%q: there is a pending operation, intent still deleting", d.Id()))
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteIntent(input)
	}

	if err != nil {
		return fmt.Errorf("error deleting intent %s: %w", d.Id(), err)
	}

	_, err = waiter.LexIntentDeleted(conn, d.Id())

	return err
}

var lexCodeHookResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"message_version": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringLenBetween(1, 5),
		},
		"uri": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validateArn,
		},
	},
}

var lexMessageResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"content": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringLenBetween(1, 1000),
		},
		"content_type": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice(lexmodelbuildingservice.ContentType_Values(), false),
		},
		"group_number": {
			Type:         schema.TypeInt,
			Optional:     true,
			ValidateFunc: validation.IntBetween(1, 5),
		},
	},
}

var lexPromptResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"max_attempts": {
			Type:         schema.TypeInt,
			Required:     true,
			ValidateFunc: validation.IntBetween(1, 5),
		},
		"message": {
			Type:     schema.TypeSet,
			Required: true,
			MinItems: 1,
			MaxItems: 15,
			Elem:     lexMessageResource,
		},
		"response_card": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringLenBetween(1, 50000),
		},
	},
}

var lexStatementResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"message": {
			Type:     schema.TypeSet,
			Required: true,
			MinItems: 1,
			MaxItems: 15,
			Elem:     lexMessageResource,
		},
		"response_card": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringLenBetween(1, 50000),
		},
	},
}

func getLatestLexIntentVersion(conn *lexmodelbuildingservice.LexModelBuildingService, input *lexmodelbuildingservice.GetIntentVersionsInput) (string, error) {
	version := LexIntentVersionLatest

	for {
		page, err := conn.GetIntentVersions(input)
		if err != nil {
			return "", err
		}

		// At least 1 version will always be returned.
		if len(page.Intents) == 1 {
			break
		}

		for _, intent := range page.Intents {
			if *intent.Version == LexIntentVersionLatest {
				continue
			}
			if *intent.Version > version {
				version = *intent.Version
			}
		}

		if page.NextToken == nil {
			break
		}
		input.NextToken = page.NextToken
	}

	return version, nil
}

func flattenLexCodeHook(hook *lexmodelbuildingservice.CodeHook) (flattened []map[string]interface{}) {
	return []map[string]interface{}{
		{
			"message_version": aws.StringValue(hook.MessageVersion),
			"uri":             aws.StringValue(hook.Uri),
		},
	}
}

func expandLexCodeHook(rawObject interface{}) (hook *lexmodelbuildingservice.CodeHook) {
	m := rawObject.([]interface{})[0].(map[string]interface{})

	return &lexmodelbuildingservice.CodeHook{
		MessageVersion: aws.String(m["message_version"].(string)),
		Uri:            aws.String(m["uri"].(string)),
	}
}

func flattenLexFollowUpPrompt(followUp *lexmodelbuildingservice.FollowUpPrompt) (flattened []map[string]interface{}) {
	return []map[string]interface{}{
		{
			"prompt":              flattenLexPrompt(followUp.Prompt),
			"rejection_statement": flattenLexStatement(followUp.RejectionStatement),
		},
	}
}

func expandLexFollowUpPrompt(rawObject interface{}) (followUp *lexmodelbuildingservice.FollowUpPrompt) {
	m := rawObject.([]interface{})[0].(map[string]interface{})

	return &lexmodelbuildingservice.FollowUpPrompt{
		Prompt:             expandLexPrompt(m["prompt"]),
		RejectionStatement: expandLexStatement(m["rejection_statement"]),
	}
}

func flattenLexFulfilmentActivity(activity *lexmodelbuildingservice.FulfillmentActivity) (flattened []map[string]interface{}) {
	flattened = []map[string]interface{}{
		{
			"type": aws.StringValue(activity.Type),
		},
	}

	if activity.CodeHook != nil {
		flattened[0]["code_hook"] = flattenLexCodeHook(activity.CodeHook)
	}

	return
}

func expandLexFulfilmentActivity(rawObject interface{}) (activity *lexmodelbuildingservice.FulfillmentActivity) {
	m := rawObject.([]interface{})[0].(map[string]interface{})

	activity = &lexmodelbuildingservice.FulfillmentActivity{}
	activity.Type = aws.String(m["type"].(string))

	if v, ok := m["code_hook"]; ok && len(v.([]interface{})) != 0 {
		activity.CodeHook = expandLexCodeHook(v)
	}

	return
}

func flattenLexMessages(messages []*lexmodelbuildingservice.Message) (flattenedMessages []map[string]interface{}) {
	for _, message := range messages {
		flattenedMessages = append(flattenedMessages, map[string]interface{}{
			"content":      aws.StringValue(message.Content),
			"content_type": aws.StringValue(message.ContentType),
			"group_number": aws.Int64Value(message.GroupNumber),
		})
	}

	return
}

// Expects a slice of maps representing the Lex objects.
// The value passed into this function should have been run through the expandLexSet function.
// Example: []map[content: test content_type: PlainText group_number: 1]
func expandLexMessages(rawValues []interface{}) []*lexmodelbuildingservice.Message {
	messages := make([]*lexmodelbuildingservice.Message, 0, len(rawValues))

	for _, rawValue := range rawValues {
		value, ok := rawValue.(map[string]interface{})
		if !ok {
			continue
		}

		message := &lexmodelbuildingservice.Message{
			Content:     aws.String(value["content"].(string)),
			ContentType: aws.String(value["content_type"].(string)),
		}

		if v, ok := value["group_number"]; ok && v != 0 {
			message.GroupNumber = aws.Int64(int64(v.(int)))
		}

		messages = append(messages, message)
	}

	return messages
}

func flattenLexPrompt(prompt *lexmodelbuildingservice.Prompt) (flattened []map[string]interface{}) {
	flattened = []map[string]interface{}{
		{
			"max_attempts": aws.Int64Value(prompt.MaxAttempts),
			"message":      flattenLexMessages(prompt.Messages),
		},
	}

	if prompt.ResponseCard != nil {
		flattened[0]["response_card"] = aws.StringValue(prompt.ResponseCard)
	}

	return
}

func expandLexPrompt(rawObject interface{}) (prompt *lexmodelbuildingservice.Prompt) {
	m := rawObject.([]interface{})[0].(map[string]interface{})

	prompt = &lexmodelbuildingservice.Prompt{}
	prompt.MaxAttempts = aws.Int64(int64(m["max_attempts"].(int)))
	prompt.Messages = expandLexMessages(m["message"].(*schema.Set).List())

	if v, ok := m["response_card"]; ok && v != "" {
		prompt.ResponseCard = aws.String(v.(string))
	}

	return
}

func flattenLexSlots(slots []*lexmodelbuildingservice.Slot) (flattenedSlots []map[string]interface{}) {
	for _, slot := range slots {
		flattenedSlot := map[string]interface{}{
			"name":            aws.StringValue(slot.Name),
			"priority":        aws.Int64Value(slot.Priority),
			"slot_constraint": aws.StringValue(slot.SlotConstraint),
			"slot_type":       aws.StringValue(slot.SlotType),
		}

		if slot.Description != nil {
			flattenedSlot["description"] = aws.StringValue(slot.Description)
		}

		if slot.ResponseCard != nil {
			flattenedSlot["response_card"] = aws.StringValue(slot.ResponseCard)
		}

		if slot.SampleUtterances != nil {
			flattenedSlot["sample_utterances"] = flattenStringList(slot.SampleUtterances)
		}

		if slot.SlotTypeVersion != nil {
			flattenedSlot["slot_type_version"] = aws.StringValue(slot.SlotTypeVersion)
		}

		if slot.ValueElicitationPrompt != nil {
			flattenedSlot["value_elicitation_prompt"] = flattenLexPrompt(slot.ValueElicitationPrompt)
		}

		flattenedSlots = append(flattenedSlots, flattenedSlot)
	}

	return flattenedSlots
}

// Expects a slice of maps representing the Lex objects.
// The value passed into this function should have been run through the expandLexSet function.
// Example: []map[name: test priority: 0 ...]
func expandLexSlots(rawValues []interface{}) []*lexmodelbuildingservice.Slot {
	slots := make([]*lexmodelbuildingservice.Slot, 0, len(rawValues))

	for _, rawValue := range rawValues {
		value, ok := rawValue.(map[string]interface{})
		if !ok {
			continue
		}

		slot := &lexmodelbuildingservice.Slot{
			Name:           aws.String(value["name"].(string)),
			Priority:       aws.Int64(int64(value["priority"].(int))),
			SlotConstraint: aws.String(value["slot_constraint"].(string)),
			SlotType:       aws.String(value["slot_type"].(string)),
		}

		if v, ok := value["description"]; ok && v != "" {
			slot.Description = aws.String(v.(string))
		}

		if v, ok := value["response_card"]; ok && v != "" {
			slot.ResponseCard = aws.String(v.(string))
		}

		if v, ok := value["response_card"]; ok && v != "" {
			slot.ResponseCard = aws.String(v.(string))
		}

		if v, ok := value["sample_utterances"]; ok && len(v.([]interface{})) != 0 {
			slot.SampleUtterances = expandStringList(v.([]interface{}))
		}

		if v, ok := value["slot_type_version"]; ok && v != "" {
			slot.SlotTypeVersion = aws.String(v.(string))
		}

		if v, ok := value["value_elicitation_prompt"]; ok && len(v.([]interface{})) != 0 {
			slot.ValueElicitationPrompt = expandLexPrompt(v)
		}

		slots = append(slots, slot)
	}

	return slots
}

func flattenLexStatement(statement *lexmodelbuildingservice.Statement) (flattened []map[string]interface{}) {
	flattened = []map[string]interface{}{
		{
			"message": flattenLexMessages(statement.Messages),
		},
	}

	if statement.ResponseCard != nil {
		flattened[0]["response_card"] = aws.StringValue(statement.ResponseCard)
	}

	return
}

func expandLexStatement(rawObject interface{}) (statement *lexmodelbuildingservice.Statement) {
	m := rawObject.([]interface{})[0].(map[string]interface{})

	statement = &lexmodelbuildingservice.Statement{}
	statement.Messages = expandLexMessages(m["message"].(*schema.Set).List())

	if v, ok := m["response_card"]; ok && v != "" {
		statement.ResponseCard = aws.String(v.(string))
	}

	return
}
