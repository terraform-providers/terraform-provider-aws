package aws

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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

	statementID := d.Get("statement_id").(string)

	input := events.PutPermissionInput{
		Action:      aws.String(d.Get("action").(string)),
		Condition:   expandCloudWatchEventsCondition(d.Get("condition").([]interface{})),
		Principal:   aws.String(d.Get("principal").(string)),
		StatementId: aws.String(statementID),
	}

	log.Printf("[DEBUG] Creating CloudWatch Events permission: %s", input)
	_, err := conn.PutPermission(&input)
	if err != nil {
		return fmt.Errorf("Creating CloudWatch Events permission failed: %s", err.Error())
	}

	d.SetId(statementID)

	return resourceAwsCloudWatchEventPermissionRead(d, meta)
}

// See also: https://docs.aws.amazon.com/AmazonCloudWatchEvents/latest/APIReference/API_DescribeEventBus.html
func resourceAwsCloudWatchEventPermissionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatcheventsconn
	input := events.DescribeEventBusInput{}
	var output *events.DescribeEventBusOutput
	var policyStatement *CloudWatchEventPermissionPolicyStatement

	// Especially with concurrent PutPermission calls there can be a slight delay
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		log.Printf("[DEBUG] Reading CloudWatch Events bus: %s", input)
		output, err := conn.DescribeEventBus(&input)
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Reading CloudWatch Events permission '%s' failed: %s", d.Id(), err.Error()))
		}

		policyStatement, err = getPolicyStatement(output, d.Id())
		if err != nil {
			return resource.RetryableError(err)
		}
		return nil
	})

	if isResourceTimeoutError(err) {
		output, err = conn.DescribeEventBus(&input)
		if output != nil {
			policyStatement, err = getPolicyStatement(output, d.Id())
		}
	}

	if isResourceNotFoundError(err) {
		log.Printf("[WARN] %s", err)
		d.SetId("")
		return nil
	}
	if err != nil {
		// Missing statement inside valid policy
		return err
	}

	d.Set("action", policyStatement.Action)

	if err := d.Set("condition", flattenCloudWatchEventPermissionPolicyStatementCondition(policyStatement.Condition)); err != nil {
		return fmt.Errorf("error setting condition: %s", err)
	}

	principalString, ok := policyStatement.Principal.(string)
	if ok && (principalString == "*") {
		d.Set("principal", "*")
	} else {
		principalMap := policyStatement.Principal.(map[string]interface{})
		policyARN, err := arn.Parse(principalMap["AWS"].(string))
		if err != nil {
			return fmt.Errorf("Reading CloudWatch Events permission '%s' failed: %s", d.Id(), err)
		}
		d.Set("principal", policyARN.AccountID)
	}
	d.Set("statement_id", policyStatement.Sid)

	return nil
}

func getPolicyStatement(output *events.DescribeEventBusOutput, statementID string) (*CloudWatchEventPermissionPolicyStatement, error) {
	var policyDoc CloudWatchEventPermissionPolicyDoc

	if output == nil || output.Policy == nil {
		return nil, &resource.NotFoundError{
			Message: fmt.Sprintf("CloudWatch Events permission %q not found"+
				"in given results from DescribeEventBus", statementID),
			LastResponse: output,
		}
	}

	err := json.Unmarshal([]byte(*output.Policy), &policyDoc)
	if err != nil {
		return nil, fmt.Errorf("Reading CloudWatch Events permission '%s' failed: %s", statementID, err)
	}

	return findCloudWatchEventPermissionPolicyStatementByID(&policyDoc, statementID)
}

func resourceAwsCloudWatchEventPermissionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatcheventsconn

	input := events.PutPermissionInput{
		Action:      aws.String(d.Get("action").(string)),
		Condition:   expandCloudWatchEventsCondition(d.Get("condition").([]interface{})),
		Principal:   aws.String(d.Get("principal").(string)),
		StatementId: aws.String(d.Get("statement_id").(string)),
	}

	log.Printf("[DEBUG] Update CloudWatch Events permission: %s", input)
	_, err := conn.PutPermission(&input)
	if isAWSErr(err, events.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] CloudWatch Events permission %q not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Updating CloudWatch Events permission '%s' failed: %s", d.Id(), err.Error())
	}

	return resourceAwsCloudWatchEventPermissionRead(d, meta)
}

func resourceAwsCloudWatchEventPermissionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatcheventsconn
	input := events.RemovePermissionInput{
		StatementId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Delete CloudWatch Events permission: %s", input)
	_, err := conn.RemovePermission(&input)
	if isAWSErr(err, events.ErrCodeResourceNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("Deleting CloudWatch Events permission '%s' failed: %s", d.Id(), err.Error())
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
