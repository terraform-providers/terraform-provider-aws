package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

const (
	StackSetInstanceCreateTimeout  = 30 * time.Minute
	StackSetInstanceUpdateTimeout  = 30 * time.Minute
	StackSetInstanceDeletedTimeout = 30 * time.Minute
)

func resourceAwsCloudFormationStackSetInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloudFormationStackSetInstanceCreate,
		Read:   resourceAwsCloudFormationStackSetInstanceRead,
		Update: resourceAwsCloudFormationStackSetInstanceUpdate,
		Delete: resourceAwsCloudFormationStackSetInstanceDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(StackSetInstanceCreateTimeout),
			Update: schema.DefaultTimeout(StackSetInstanceUpdateTimeout),
			Delete: schema.DefaultTimeout(StackSetInstanceDeletedTimeout),
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsAccountId,
			},
			"parameter_overrides": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"retain_stack": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"stack_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"stack_set_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
		},
	}
}

func resourceAwsCloudFormationStackSetInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cfconn

	accountID := meta.(*AWSClient).accountid
	if v, ok := d.GetOk("account_id"); ok {
		accountID = v.(string)
	}

	region := meta.(*AWSClient).region
	if v, ok := d.GetOk("region"); ok {
		region = v.(string)
	}

	stackSetName := d.Get("stack_set_name").(string)

	input := &cloudformation.CreateStackInstancesInput{
		Accounts:     aws.StringSlice([]string{accountID}),
		OperationId:  aws.String(resource.UniqueId()),
		Regions:      aws.StringSlice([]string{region}),
		StackSetName: aws.String(stackSetName),
	}

	if v, ok := d.GetOk("parameter_overrides"); ok {
		input.ParameterOverrides = expandCloudFormationParameters(v.(map[string]interface{}))
	}

	log.Printf("[DEBUG] Creating CloudFormation StackSet Instance: %s", input)
	output, err := conn.CreateStackInstances(input)

	if err != nil {
		return fmt.Errorf("error creating CloudFormation StackSet Instance: %s", err)
	}

	d.SetId(resourceAwsCloudFormationStackSetInstanceCreateId(stackSetName, accountID, region))

	if err := waitForCloudFormationStackSetOperation(conn, stackSetName, aws.StringValue(output.OperationId), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for CloudFormation StackSet Instance (%s) creation: %s", d.Id(), err)
	}

	return resourceAwsCloudFormationStackSetInstanceRead(d, meta)
}

func resourceAwsCloudFormationStackSetInstanceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cfconn

	stackSetName, accountID, region, err := resourceAwsCloudFormationStackSetInstanceParseId(d.Id())

	if err != nil {
		return err
	}

	input := &cloudformation.DescribeStackInstanceInput{
		StackInstanceAccount: aws.String(accountID),
		StackInstanceRegion:  aws.String(region),
		StackSetName:         aws.String(stackSetName),
	}

	log.Printf("[DEBUG] Reading CloudFormation StackSet Instance: %s", d.Id())
	output, err := conn.DescribeStackInstance(input)

	if isAWSErr(err, cloudformation.ErrCodeStackInstanceNotFoundException, "") {
		log.Printf("[WARN] CloudFormation StackSet Instance (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if isAWSErr(err, cloudformation.ErrCodeStackSetNotFoundException, "") {
		log.Printf("[WARN] CloudFormation StackSet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading CloudFormation StackSet Instance (%s): %s", d.Id(), err)
	}

	if output == nil || output.StackInstance == nil {
		return fmt.Errorf("error reading CloudFormation StackSet Instance (%s): empty response", d.Id())
	}

	stackInstance := output.StackInstance

	d.Set("account_id", stackInstance.Account)

	if err := d.Set("parameter_overrides", flattenAllCloudFormationParameters(stackInstance.ParameterOverrides)); err != nil {
		return fmt.Errorf("error setting parameters: %s", err)
	}

	d.Set("region", stackInstance.Region)
	d.Set("stack_id", stackInstance.StackId)
	d.Set("stack_set_name", stackSetName)

	return nil
}

func resourceAwsCloudFormationStackSetInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cfconn

	if d.HasChange("parameter_overrides") {
		stackSetName, accountID, region, err := resourceAwsCloudFormationStackSetInstanceParseId(d.Id())

		if err != nil {
			return err
		}

		input := &cloudformation.UpdateStackInstancesInput{
			Accounts:           aws.StringSlice([]string{accountID}),
			OperationId:        aws.String(resource.UniqueId()),
			ParameterOverrides: []*cloudformation.Parameter{},
			Regions:            aws.StringSlice([]string{region}),
			StackSetName:       aws.String(stackSetName),
		}

		if v, ok := d.GetOk("parameter_overrides"); ok {
			input.ParameterOverrides = expandCloudFormationParameters(v.(map[string]interface{}))
		}

		log.Printf("[DEBUG] Updating CloudFormation StackSet Instance: %s", input)
		output, err := conn.UpdateStackInstances(input)

		if err != nil {
			return fmt.Errorf("error updating CloudFormation StackSet Instance (%s): %s", d.Id(), err)
		}

		if err := waitForCloudFormationStackSetOperation(conn, stackSetName, aws.StringValue(output.OperationId), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("error waiting for CloudFormation StackSet Instance (%s) update: %s", d.Id(), err)
		}
	}

	return resourceAwsCloudFormationStackSetInstanceRead(d, meta)
}

func resourceAwsCloudFormationStackSetInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cfconn

	log.Printf("[DEBUG] Deleting CloudFormation StackSet Instance: %s", d.Id())
	input, err := deleteCloudFormationStackSetInstanceInputFromResourceData(d)
	if err != nil {
		return err
	}
	return deleteCloudFormationStackSetInstance(conn, input, d.Timeout(schema.TimeoutDelete))
}

func deleteCloudFormationStackSetInstanceInputFromResourceData(d *schema.ResourceData) (*cloudformation.DeleteStackInstancesInput, error) {
	stackSetName, accountID, region, err := resourceAwsCloudFormationStackSetInstanceParseId(d.Id())
	if err != nil {
		return nil, err
	}

	return &cloudformation.DeleteStackInstancesInput{
		OperationId:  aws.String(resource.UniqueId()),
		Accounts:     aws.StringSlice([]string{accountID}),
		Regions:      aws.StringSlice([]string{region}),
		StackSetName: aws.String(stackSetName),
		RetainStacks: aws.Bool(d.Get("retain_stack").(bool)),
	}, nil
}

func deleteCloudFormationStackSetInstanceInputFromAPIResource(p *cloudformation.StackSetSummary, r *cloudformation.StackInstanceSummary) *cloudformation.DeleteStackInstancesInput {
	return &cloudformation.DeleteStackInstancesInput{
		OperationId:  aws.String(resource.UniqueId()),
		Accounts:     []*string{r.Account},
		Regions:      []*string{r.Region},
		StackSetName: p.StackSetName,
		RetainStacks: aws.Bool(false),
	}
}

func deleteCloudFormationStackSetInstance(conn *cloudformation.CloudFormation, input *cloudformation.DeleteStackInstancesInput, timeout time.Duration) error {
	stackSetName := aws.StringValue(input.StackSetName)
	accountID := aws.StringValue(input.Accounts[0])
	region := aws.StringValue(input.Regions[0])
	id := resourceAwsCloudFormationStackSetInstanceCreateId(stackSetName, accountID, region)

	output, err := conn.DeleteStackInstances(input)
	if isAWSErr(err, cloudformation.ErrCodeStackInstanceNotFoundException, "") {
		return nil
	}
	if isAWSErr(err, cloudformation.ErrCodeStackSetNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting CloudFormation StackSet Instance (%s): %w", id, err)
	}

	if err := waitForCloudFormationStackSetOperation(conn, stackSetName, aws.StringValue(output.OperationId), timeout); err != nil {
		return fmt.Errorf("error waiting for CloudFormation StackSet Instance (%s) deletion: %w", id, err)
	}

	return nil
}

func resourceAwsCloudFormationStackSetInstanceCreateId(stackSetName, accountID, region string) string {
	return fmt.Sprintf("%s,%s,%s", stackSetName, accountID, region)
}

func resourceAwsCloudFormationStackSetInstanceParseId(id string) (string, string, string, error) {
	idFormatErr := fmt.Errorf("unexpected format of ID (%s), expected NAME,ACCOUNT,REGION", id)

	parts := strings.SplitN(id, ",", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", idFormatErr
	}

	return parts[0], parts[1], parts[2], nil
}

func refreshCloudformationStackSetOperation(conn *cloudformation.CloudFormation, stackSetName, operationID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &cloudformation.DescribeStackSetOperationInput{
			OperationId:  aws.String(operationID),
			StackSetName: aws.String(stackSetName),
		}

		output, err := conn.DescribeStackSetOperation(input)

		if isAWSErr(err, cloudformation.ErrCodeOperationNotFoundException, "") {
			return nil, cloudformation.StackSetOperationStatusRunning, nil
		}

		if err != nil {
			return nil, cloudformation.StackSetOperationStatusFailed, err
		}

		if output == nil || output.StackSetOperation == nil {
			return nil, cloudformation.StackSetOperationStatusRunning, nil
		}

		if aws.StringValue(output.StackSetOperation.Status) == cloudformation.StackSetOperationStatusFailed {
			allResults := make([]string, 0)
			listOperationResultsInput := &cloudformation.ListStackSetOperationResultsInput{
				OperationId:  aws.String(operationID),
				StackSetName: aws.String(stackSetName),
			}

			for {
				listOperationResultsOutput, err := conn.ListStackSetOperationResults(listOperationResultsInput)

				if err != nil {
					return output.StackSetOperation, cloudformation.StackSetOperationStatusFailed, fmt.Errorf("error listing Operation (%s) errors: %s", operationID, err)
				}

				if listOperationResultsOutput == nil {
					continue
				}

				for _, summary := range listOperationResultsOutput.Summaries {
					allResults = append(allResults, fmt.Sprintf("Account (%s) Region (%s) Status (%s) Status Reason: %s", aws.StringValue(summary.Account), aws.StringValue(summary.Region), aws.StringValue(summary.Status), aws.StringValue(summary.StatusReason)))
				}

				if aws.StringValue(listOperationResultsOutput.NextToken) == "" {
					break
				}

				listOperationResultsInput.NextToken = listOperationResultsOutput.NextToken
			}

			return output.StackSetOperation, cloudformation.StackSetOperationStatusFailed, fmt.Errorf("Operation (%s) Results:\n%s", operationID, strings.Join(allResults, "\n"))
		}

		return output.StackSetOperation, aws.StringValue(output.StackSetOperation.Status), nil
	}
}

func waitForCloudFormationStackSetOperation(conn *cloudformation.CloudFormation, stackSetName, operationID string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{cloudformation.StackSetOperationStatusRunning},
		Target:  []string{cloudformation.StackSetOperationStatusSucceeded},
		Refresh: refreshCloudformationStackSetOperation(conn, stackSetName, operationID),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	log.Printf("[DEBUG] Waiting for CloudFormation StackSet (%s) operation: %s", stackSetName, operationID)
	_, err := stateConf.WaitForState()

	return err
}
