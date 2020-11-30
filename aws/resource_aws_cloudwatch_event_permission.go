package aws

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tfevents "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudwatchevents"
	iamwaiter "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
)

func resourceAwsCloudWatchEventPermission() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloudWatchEventPermissionCreate,
		Read:   resourceAwsCloudWatchEventPermissionRead,
		Update: resourceAwsCloudWatchEventPermissionUpdate,
		Delete: resourceAwsCloudWatchEventPermissionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"action": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "events:PutEvents",
				ValidateFunc: validateCloudWatchEventPermissionAction,
			},
			"condition": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"aws:PrincipalOrgID"}, false),
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"StringEquals"}, false),
						},
						"value": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.NoZeroValues,
						},
					},
				},
			},
			"event_bus_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateCloudWatchEventBusName,
				Default:      tfevents.DefaultEventBusName,
			},
			"principal": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateCloudWatchEventPermissionPrincipal,
			},
			"statement_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateCloudWatchEventPermissionStatementID,
			},
		},
	}
}

func resourceAwsCloudWatchEventPermissionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatcheventsconn

	eventBusName := d.Get("event_bus_name").(string)
	statementID := d.Get("statement_id").(string)

	input := events.PutPermissionInput{
		Action:       aws.String(d.Get("action").(string)),
		Condition:    expandCloudWatchEventsCondition(d.Get("condition").([]interface{})),
		EventBusName: aws.String(eventBusName),
		Principal:    aws.String(d.Get("principal").(string)),
		StatementId:  aws.String(statementID),
	}

	log.Printf("[DEBUG] Creating CloudWatch Events permission: %s", input)
	_, err := conn.PutPermission(&input)
	if err != nil {
		return fmt.Errorf("Creating CloudWatch Events permission failed: %w", err)
	}

	id := tfevents.PermissionCreateID(eventBusName, statementID)
	d.SetId(id)

	return resourceAwsCloudWatchEventPermissionRead(d, meta)
}

// See also: https://docs.aws.amazon.com/AmazonCloudWatchEvents/latest/APIReference/API_DescribeEventBus.html
func resourceAwsCloudWatchEventPermissionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatcheventsconn

	eventBusName, statementID, err := tfevents.PermissionParseID(d.Id())
	if err != nil {
		return fmt.Errorf("error reading CloudWatch Events permission (%s): %w", d.Id(), err)
	}
	input := events.DescribeEventBusInput{
		Name: aws.String(eventBusName),
	}
	var output *events.DescribeEventBusOutput
	var policyStatement *CloudWatchEventPermissionPolicyStatement

	// Especially with concurrent PutPermission calls there can be a slight delay
	err = resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		log.Printf("[DEBUG] Reading CloudWatch Events bus: %s", input)
		output, err = conn.DescribeEventBus(&input)
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("reading CloudWatch Events permission (%s) failed: %w", d.Id(), err))
		}

		policyStatement, err = getPolicyStatement(output, statementID)
		if err != nil {
			return resource.RetryableError(err)
		}
		return nil
	})

	if isResourceTimeoutError(err) {
		output, err = conn.DescribeEventBus(&input)
		if output != nil {
			policyStatement, err = getPolicyStatement(output, statementID)
		}
	}

	if isResourceNotFoundError(err) {
		log.Printf("[WARN] CloudWatch Events permission (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading CloudWatch Events permission (%s): %w", d.Id(), err)
	}

	d.Set("action", policyStatement.Action)
	busName := aws.StringValue(output.Name)
	if busName == "" {
		busName = tfevents.DefaultEventBusName
	}
	d.Set("event_bus_name", busName)

	if err := d.Set("condition", flattenCloudWatchEventPermissionPolicyStatementCondition(policyStatement.Condition)); err != nil {
		return fmt.Errorf("error setting condition: %w", err)
	}

	switch principal := policyStatement.Principal.(type) {
	case string:
		d.Set("principal", principal)
	case map[string]interface{}:
		if v, ok := principal["AWS"].(string); ok {
			if arn.IsARN(v) {
				principalARN, err := arn.Parse(v)

				if err != nil {
					return fmt.Errorf("error parsing CloudWatch Events Permission (%s) principal as ARN (%s): %w", d.Id(), v, err)
				}

				d.Set("principal", principalARN.AccountID)
			} else {
				d.Set("principal", v)
			}
		}
	}

	d.Set("statement_id", policyStatement.Sid)

	return nil
}

func getPolicyStatement(output *events.DescribeEventBusOutput, statementID string) (*CloudWatchEventPermissionPolicyStatement, error) {
	var policyDoc CloudWatchEventPermissionPolicyDoc

	if output == nil || output.Policy == nil {
		return nil, &resource.NotFoundError{
			Message:      fmt.Sprintf("CloudWatch Events permission %q not found", statementID),
			LastResponse: output,
		}
	}

	err := json.Unmarshal([]byte(*output.Policy), &policyDoc)
	if err != nil {
		return nil, fmt.Errorf("error reading CloudWatch Events permission (%s): %w", statementID, err)
	}

	return findCloudWatchEventPermissionPolicyStatementByID(&policyDoc, statementID)
}

func resourceAwsCloudWatchEventPermissionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatcheventsconn

	eventBusName, statementID, err := tfevents.PermissionParseID(d.Id())
	if err != nil {
		return fmt.Errorf("error updating CloudWatch Events permission (%s): %w", d.Id(), err)
	}
	input := events.PutPermissionInput{
		Action:       aws.String(d.Get("action").(string)),
		Condition:    expandCloudWatchEventsCondition(d.Get("condition").([]interface{})),
		EventBusName: aws.String(eventBusName),
		Principal:    aws.String(d.Get("principal").(string)),
		StatementId:  aws.String(statementID),
	}

	log.Printf("[DEBUG] Update CloudWatch Events permission: %s", input)
	_, err = conn.PutPermission(&input)
	if isAWSErr(err, events.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] CloudWatch Events permission %q not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error updating CloudWatch Events permission (%s): %w", d.Id(), err)
	}

	return resourceAwsCloudWatchEventPermissionRead(d, meta)
}

func resourceAwsCloudWatchEventPermissionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatcheventsconn

	eventBusName, statementID, err := tfevents.PermissionParseID(d.Id())
	if err != nil {
		return fmt.Errorf("error deleting CloudWatch Events permission (%s): %w", d.Id(), err)
	}
	input := events.RemovePermissionInput{
		EventBusName: aws.String(eventBusName),
		StatementId:  aws.String(statementID),
	}

	log.Printf("[DEBUG] Delete CloudWatch Events permission: %s", input)
	_, err = conn.RemovePermission(&input)
	if isAWSErr(err, events.ErrCodeResourceNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting CloudWatch Events permission (%s): %w", d.Id(), err)
	}
	return nil
}

// https://docs.aws.amazon.com/AmazonCloudWatchEvents/latest/APIReference/API_PutPermission.html#API_PutPermission_RequestParameters
func validateCloudWatchEventPermissionAction(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)
	if (len(value) < 1) || (len(value) > 64) {
		es = append(es, fmt.Errorf("%q must be between 1 and 64 characters", k))
	}

	if !regexp.MustCompile(`^events:[a-zA-Z]+$`).MatchString(value) {
		es = append(es, fmt.Errorf("%q must be: events: followed by one or more alphabetic characters", k))
	}
	return
}

// https://docs.aws.amazon.com/AmazonCloudWatchEvents/latest/APIReference/API_PutPermission.html#API_PutPermission_RequestParameters
func validateCloudWatchEventPermissionPrincipal(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)
	if !regexp.MustCompile(`^(\d{12}|\*)$`).MatchString(value) {
		es = append(es, fmt.Errorf("%q must be * or a 12 digit AWS account ID", k))
	}
	return
}

// https://docs.aws.amazon.com/AmazonCloudWatchEvents/latest/APIReference/API_PutPermission.html#API_PutPermission_RequestParameters
func validateCloudWatchEventPermissionStatementID(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)
	if (len(value) < 1) || (len(value) > 64) {
		es = append(es, fmt.Errorf("%q must be between 1 and 64 characters", k))
	}

	if !regexp.MustCompile(`^[a-zA-Z0-9-_]+$`).MatchString(value) {
		es = append(es, fmt.Errorf("%q must be one or more alphanumeric, hyphen, or underscore characters", k))
	}
	return
}

// CloudWatchEventPermissionPolicyDoc represents the Policy attribute of DescribeEventBus
// See also: https://docs.aws.amazon.com/AmazonCloudWatchEvents/latest/APIReference/API_DescribeEventBus.html
type CloudWatchEventPermissionPolicyDoc struct {
	Version    string
	ID         string                                     `json:"Id,omitempty"`
	Statements []CloudWatchEventPermissionPolicyStatement `json:"Statement"`
}

// String returns the string representation
func (d CloudWatchEventPermissionPolicyDoc) String() string {
	return awsutil.Prettify(d)
}

// GoString returns the string representation
func (d CloudWatchEventPermissionPolicyDoc) GoString() string {
	return d.String()
}

// CloudWatchEventPermissionPolicyStatement represents the Statement attribute of CloudWatchEventPermissionPolicyDoc
// See also: https://docs.aws.amazon.com/AmazonCloudWatchEvents/latest/APIReference/API_DescribeEventBus.html
type CloudWatchEventPermissionPolicyStatement struct {
	Sid       string
	Effect    string
	Action    string
	Condition *CloudWatchEventPermissionPolicyStatementCondition `json:"Condition,omitempty"`
	Principal interface{}                                        // "*" or {"AWS": "arn:aws:iam::111111111111:root"}
	Resource  string
}

// String returns the string representation
func (s CloudWatchEventPermissionPolicyStatement) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s CloudWatchEventPermissionPolicyStatement) GoString() string {
	return s.String()
}

// CloudWatchEventPermissionPolicyStatementCondition represents the Condition attribute of CloudWatchEventPermissionPolicyStatement
// See also: https://docs.aws.amazon.com/AmazonCloudWatchEvents/latest/APIReference/API_DescribeEventBus.html
type CloudWatchEventPermissionPolicyStatementCondition struct {
	Key   string
	Type  string
	Value string
}

// String returns the string representation
func (c CloudWatchEventPermissionPolicyStatementCondition) String() string {
	return awsutil.Prettify(c)
}

// GoString returns the string representation
func (c CloudWatchEventPermissionPolicyStatementCondition) GoString() string {
	return c.String()
}

func (c *CloudWatchEventPermissionPolicyStatementCondition) UnmarshalJSON(b []byte) error {
	var out CloudWatchEventPermissionPolicyStatementCondition

	// JSON representation: \"Condition\":{\"StringEquals\":{\"aws:PrincipalOrgID\":\"o-0123456789\"}}
	var data map[string]map[string]string
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}

	for typeKey, typeValue := range data {
		for conditionKey, conditionValue := range typeValue {
			out = CloudWatchEventPermissionPolicyStatementCondition{
				Key:   conditionKey,
				Type:  typeKey,
				Value: conditionValue,
			}
		}
	}

	*c = out
	return nil
}

func findCloudWatchEventPermissionPolicyStatementByID(policy *CloudWatchEventPermissionPolicyDoc, id string) (
	*CloudWatchEventPermissionPolicyStatement, error) {

	log.Printf("[DEBUG] Finding statement (%s) in CloudWatch Events permission policy: %s", id, policy)
	for _, statement := range policy.Statements {
		if statement.Sid == id {
			return &statement, nil
		}
	}

	return nil, &resource.NotFoundError{
		LastRequest:  id,
		LastResponse: policy,
		Message:      fmt.Sprintf("Failed to find statement (%s) in CloudWatch Events permission policy: %s", id, policy),
	}
}

func expandCloudWatchEventsCondition(l []interface{}) *events.Condition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	condition := &events.Condition{
		Key:   aws.String(m["key"].(string)),
		Type:  aws.String(m["type"].(string)),
		Value: aws.String(m["value"].(string)),
	}

	return condition
}

func flattenCloudWatchEventPermissionPolicyStatementCondition(c *CloudWatchEventPermissionPolicyStatementCondition) []interface{} {
	if c == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"key":   c.Key,
		"type":  c.Type,
		"value": c.Value,
	}

	return []interface{}{m}
}
