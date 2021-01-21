package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/glue/waiter"
)

func resourceAwsGlueTrigger() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGlueTriggerCreate,
		Read:   resourceAwsGlueTriggerRead,
		Update: resourceAwsGlueTriggerUpdate,
		Delete: resourceAwsGlueTriggerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"actions": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arguments": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"crawler_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"job_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"timeout": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"security_configuration": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"notification_property": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"notify_delay_after": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
					},
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"predicate": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"conditions": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"job_name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"crawler_name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"logical_operator": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      glue.LogicalOperatorEquals,
										ValidateFunc: validation.StringInSlice(glue.LogicalOperator_Values(), false),
									},
									"state": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(glue.JobRunState_Values(), false),
									},
									"crawl_state": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(glue.CrawlState_Values(), false),
									},
								},
							},
						},
						"logical": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      glue.LogicalAnd,
							ValidateFunc: validation.StringInSlice(glue.Logical_Values(), false),
						},
					},
				},
			},
			"schedule": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": tagsSchema(),
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(glue.TriggerType_Values(), false),
			},
			"workflow_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsGlueTriggerCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn
	name := d.Get("name").(string)
	triggerType := d.Get("type").(string)

	input := &glue.CreateTriggerInput{
		Actions: expandGlueActions(d.Get("actions").([]interface{})),
		Name:    aws.String(name),
		Tags:    keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().GlueTags(),
		Type:    aws.String(triggerType),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("predicate"); ok {
		input.Predicate = expandGluePredicate(v.([]interface{}))
	}

	if v, ok := d.GetOk("schedule"); ok {
		input.Schedule = aws.String(v.(string))
	}

	if v, ok := d.GetOk("workflow_name"); ok {
		input.WorkflowName = aws.String(v.(string))
	}

	if d.Get("enabled").(bool) && triggerType != glue.TriggerTypeOnDemand {
		input.StartOnCreation = aws.Bool(true)
	}

	if v, ok := d.GetOk("workflow_name"); ok {
		input.WorkflowName = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Glue Trigger: %s", input)
	_, err := conn.CreateTrigger(input)
	if err != nil {
		return fmt.Errorf("error creating Glue Trigger (%s): %w", name, err)
	}

	d.SetId(name)

	log.Printf("[DEBUG] Waiting for Glue Trigger (%s) to create", d.Id())
	if _, err := waiter.TriggerCreated(conn, d.Id()); err != nil {
		if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error waiting for Glue Trigger (%s) to be Created: %w", d.Id(), err)
	}

	return resourceAwsGlueTriggerRead(d, meta)
}

func resourceAwsGlueTriggerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &glue.GetTriggerInput{
		Name: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Glue Trigger: %s", input)
	output, err := conn.GetTrigger(input)
	if err != nil {
		if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
			log.Printf("[WARN] Glue Trigger (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading Glue Trigger (%s): %w", d.Id(), err)
	}

	trigger := output.Trigger
	if trigger == nil {
		log.Printf("[WARN] Glue Trigger (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err := d.Set("actions", flattenGlueActions(trigger.Actions)); err != nil {
		return fmt.Errorf("error setting actions: %w", err)
	}

	triggerARN := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "glue",
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("trigger/%s", d.Id()),
	}.String()
	d.Set("arn", triggerARN)

	d.Set("description", trigger.Description)

	var enabled bool
	state := aws.StringValue(trigger.State)

	if aws.StringValue(trigger.Type) == glue.TriggerTypeOnDemand {
		enabled = (state == glue.TriggerStateCreated || state == glue.TriggerStateCreating)
	} else {
		enabled = (state == glue.TriggerStateActivated || state == glue.TriggerStateActivating)
	}
	d.Set("enabled", enabled)

	if err := d.Set("predicate", flattenGluePredicate(trigger.Predicate)); err != nil {
		return fmt.Errorf("error setting predicate: %w", err)
	}

	d.Set("name", trigger.Name)
	d.Set("schedule", trigger.Schedule)

	tags, err := keyvaluetags.GlueListTags(conn, triggerARN)

	if err != nil {
		return fmt.Errorf("error listing tags for Glue Trigger (%s): %w", triggerARN, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	d.Set("type", trigger.Type)
	d.Set("workflow_name", trigger.WorkflowName)

	return nil
}

func resourceAwsGlueTriggerUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	if d.HasChanges("actions", "description", "predicate", "schedule") {
		triggerUpdate := &glue.TriggerUpdate{
			Actions: expandGlueActions(d.Get("actions").([]interface{})),
		}

		if v, ok := d.GetOk("description"); ok {
			triggerUpdate.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("predicate"); ok {
			triggerUpdate.Predicate = expandGluePredicate(v.([]interface{}))
		}

		if v, ok := d.GetOk("schedule"); ok {
			triggerUpdate.Schedule = aws.String(v.(string))
		}
		input := &glue.UpdateTriggerInput{
			Name:          aws.String(d.Id()),
			TriggerUpdate: triggerUpdate,
		}

		log.Printf("[DEBUG] Updating Glue Trigger: %s", input)
		_, err := conn.UpdateTrigger(input)
		if err != nil {
			return fmt.Errorf("error updating Glue Trigger (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("enabled") {
		if d.Get("enabled").(bool) {
			input := &glue.StartTriggerInput{
				Name: aws.String(d.Id()),
			}

			log.Printf("[DEBUG] Starting Glue Trigger: %s", input)
			_, err := conn.StartTrigger(input)
			if err != nil {
				return fmt.Errorf("error starting Glue Trigger (%s): %w", d.Id(), err)
			}
		} else {
			input := &glue.StopTriggerInput{
				Name: aws.String(d.Id()),
			}

			log.Printf("[DEBUG] Stopping Glue Trigger: %s", input)
			_, err := conn.StopTrigger(input)
			if err != nil {
				return fmt.Errorf("error stopping Glue Trigger (%s): %w", d.Id(), err)
			}
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.GlueUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return nil
}

func resourceAwsGlueTriggerDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	log.Printf("[DEBUG] Deleting Glue Trigger: %s", d.Id())
	err := deleteGlueTrigger(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error deleting Glue Trigger (%s): %w", d.Id(), err)
	}

	log.Printf("[DEBUG] Waiting for Glue Trigger (%s) to delete", d.Id())
	if _, err := waiter.TriggerDeleted(conn, d.Id()); err != nil {
		if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error waiting for Glue Trigger (%s) to be Deleted: %w", d.Id(), err)
	}

	return nil
}

func deleteGlueTrigger(conn *glue.Glue, Name string) error {
	input := &glue.DeleteTriggerInput{
		Name: aws.String(Name),
	}

	_, err := conn.DeleteTrigger(input)
	if err != nil {
		if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
			return nil
		}
		return err
	}

	return nil
}

func expandGlueActions(l []interface{}) []*glue.Action {
	actions := []*glue.Action{}

	for _, mRaw := range l {
		m := mRaw.(map[string]interface{})

		action := &glue.Action{}

		if v, ok := m["crawler_name"].(string); ok && v != "" {
			action.CrawlerName = aws.String(v)
		}

		if v, ok := m["job_name"].(string); ok && v != "" {
			action.JobName = aws.String(v)
		}

		argumentsMap := make(map[string]string)
		for k, v := range m["arguments"].(map[string]interface{}) {
			argumentsMap[k] = v.(string)
		}
		action.Arguments = aws.StringMap(argumentsMap)

		if v, ok := m["timeout"]; ok && v.(int) > 0 {
			action.Timeout = aws.Int64(int64(v.(int)))
		}

		if v, ok := m["security_configuration"].(string); ok && v != "" {
			action.SecurityConfiguration = aws.String(v)
		}

		if v, ok := m["notification_property"].([]interface{}); ok && len(v) > 0 {
			action.NotificationProperty = expandGlueTriggerNotificationProperty(v)
		}

		actions = append(actions, action)
	}

	return actions
}

func expandGlueTriggerNotificationProperty(l []interface{}) *glue.NotificationProperty {
	m := l[0].(map[string]interface{})

	property := &glue.NotificationProperty{}

	if v, ok := m["notify_delay_after"]; ok && v.(int) > 0 {
		property.NotifyDelayAfter = aws.Int64(int64(v.(int)))
	}

	return property
}

func expandGlueConditions(l []interface{}) []*glue.Condition {
	conditions := []*glue.Condition{}

	for _, mRaw := range l {
		m := mRaw.(map[string]interface{})

		condition := &glue.Condition{
			LogicalOperator: aws.String(m["logical_operator"].(string)),
		}

		if v, ok := m["crawler_name"].(string); ok && v != "" {
			condition.CrawlerName = aws.String(v)
		}

		if v, ok := m["crawl_state"].(string); ok && v != "" {
			condition.CrawlState = aws.String(v)
		}

		if v, ok := m["job_name"].(string); ok && v != "" {
			condition.JobName = aws.String(v)
		}

		if v, ok := m["state"].(string); ok && v != "" {
			condition.State = aws.String(v)
		}

		conditions = append(conditions, condition)
	}

	return conditions
}

func expandGluePredicate(l []interface{}) *glue.Predicate {
	m := l[0].(map[string]interface{})

	predicate := &glue.Predicate{
		Conditions: expandGlueConditions(m["conditions"].([]interface{})),
	}

	if v, ok := m["logical"]; ok && v.(string) != "" {
		predicate.Logical = aws.String(v.(string))
	}

	return predicate
}

func flattenGlueActions(actions []*glue.Action) []interface{} {
	l := []interface{}{}

	for _, action := range actions {
		m := map[string]interface{}{
			"arguments": aws.StringValueMap(action.Arguments),
			"timeout":   int(aws.Int64Value(action.Timeout)),
		}

		if v := aws.StringValue(action.CrawlerName); v != "" {
			m["crawler_name"] = v
		}

		if v := aws.StringValue(action.JobName); v != "" {
			m["job_name"] = v
		}

		if v := aws.StringValue(action.SecurityConfiguration); v != "" {
			m["security_configuration"] = v
		}

		if v := action.NotificationProperty; v != nil {
			m["notification_property"] = flattenGlueTriggerNotificationProperty(v)
		}

		l = append(l, m)
	}

	return l
}

func flattenGlueConditions(conditions []*glue.Condition) []interface{} {
	l := []interface{}{}

	for _, condition := range conditions {
		m := map[string]interface{}{
			"logical_operator": aws.StringValue(condition.LogicalOperator),
		}

		if v := aws.StringValue(condition.CrawlerName); v != "" {
			m["crawler_name"] = v
		}

		if v := aws.StringValue(condition.CrawlState); v != "" {
			m["crawl_state"] = v
		}

		if v := aws.StringValue(condition.JobName); v != "" {
			m["job_name"] = v
		}

		if v := aws.StringValue(condition.State); v != "" {
			m["state"] = v
		}

		l = append(l, m)
	}

	return l
}

func flattenGluePredicate(predicate *glue.Predicate) []map[string]interface{} {
	if predicate == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"conditions": flattenGlueConditions(predicate.Conditions),
		"logical":    aws.StringValue(predicate.Logical),
	}

	return []map[string]interface{}{m}
}

func flattenGlueTriggerNotificationProperty(property *glue.NotificationProperty) []map[string]interface{} {
	if property == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"notify_delay_after": aws.Int64Value(property.NotifyDelayAfter),
	}

	return []map[string]interface{}{m}
}
