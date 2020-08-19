package aws

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
)

var resourceAwsQuickSighIAMPolicyAttachmentCUPendingStates = []string{
	quicksight.AssignmentStatusDisabled,
	quicksight.AssignmentStatusDraft,
	"",
}

func resourceAwsQuickSightIAMPolicyAssignment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsQuickSightIAMPolicyAssignmentCreate,
		Read:   resourceAwsQuickSightIAMPolicyAssignmentRead,
		Update: resourceAwsQuickSightIAMPolicyAssignmentUpdate,
		Delete: resourceAwsQuickSightIAMPolicyAssignmentDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Second),
			Read:   schema.DefaultTimeout(60 * time.Second),
			Update: schema.DefaultTimeout(60 * time.Second),
			Delete: schema.DefaultTimeout(60 * time.Second),
		},

		Schema: map[string]*schema.Schema{
			"assignment_name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile("^[a-zA-Z0-9]*$"),
					"The value may only contain alphanumeric value. Special chars not allowed"),
			},

			"assignment_status": {
				Type:     schema.TypeString,
				Required: true,
			},

			"aws_account_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"groups": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Optional: true,
			},

			"users": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Optional: true,
			},

			"policy_arn": {
				Type:     schema.TypeString,
				Required: true,
			},

			"namespace": {
				Type:     schema.TypeString,
				Default:  "default",
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"default",
				}, false),
			},
		},
	}
}

func resourceAwsQuickSightIAMPolicyAssignmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).quicksightconn

	awsAccountId := d.Get("aws_account_id").(string)
	namespace := d.Get("namespace").(string)
	assignmentName := d.Get("assignment_name").(string)

	if v, ok := d.GetOk("aws_account_id"); ok {
		awsAccountId = v.(string)
	}

	identities := make(map[string][]*string)
	if groupAttr := d.Get("groups").(*schema.Set); groupAttr.Len() > 0 {
		identities["Group"] = expandStringList(groupAttr.List())
	}

	if userAttr := d.Get("users").(*schema.Set); userAttr.Len() > 0 {
		identities["User"] = expandStringList(userAttr.List())
	}

	createOpts := &quicksight.CreateIAMPolicyAssignmentInput{
		AssignmentName:   aws.String(assignmentName),
		AssignmentStatus: aws.String(d.Get("assignment_status").(string)),
		AwsAccountId:     aws.String(awsAccountId),
		Identities:       identities,
		Namespace:        aws.String(namespace),
		PolicyArn:        aws.String(d.Get("policy_arn").(string)),
	}

	resp, err := conn.CreateIAMPolicyAssignment(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating QuickSight IAM Policy Assignment: %s", err)
	}

	_, err = waitIAMPolicyAttachmentCreate(conn, awsAccountId, assignmentName, namespace, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return fmt.Errorf("Error waiting for Data Source (%s) to become available: %s", assignmentName, err)
	}

	d.SetId(fmt.Sprintf("%s/%s/%s", awsAccountId, namespace, aws.StringValue(resp.AssignmentName)))
	return resourceAwsQuickSightIAMPolicyAssignmentRead(d, meta)
}

func waitIAMPolicyAttachmentCreate(conn *quicksight.QuickSight, awsAccountId, assignmentName, namespace string, timeout time.Duration) (interface{}, error) {
	stateChangeConf := &resource.StateChangeConf{
		Pending: resourceAwsQuickSighIAMPolicyAttachmentCUPendingStates,
		Target:  []string{quicksight.AssignmentStatusEnabled},
		Refresh: iamPolicyAttachmentStateRefreshFunc(conn, awsAccountId, assignmentName, namespace),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}
	return stateChangeConf.WaitForState()
}

func iamPolicyAttachmentStateRefreshFunc(conn *quicksight.QuickSight, awsAccountId, assignmentName, namespace string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		req := &quicksight.DescribeIAMPolicyAssignmentInput{
			AssignmentName: aws.String(assignmentName),
			AwsAccountId:   aws.String(awsAccountId),
			Namespace:      aws.String(namespace),
		}
		resp, err := conn.DescribeIAMPolicyAssignment(req)
		if err != nil {
			return nil, "", err
		}

		assignmentId := resp.IAMPolicyAssignment.AssignmentId
		state := ""
		if aws.StringValue(resp.IAMPolicyAssignment.AssignmentStatus) == quicksight.AssignmentStatusEnabled {
			state = quicksight.AssignmentStatusEnabled
		}
		return assignmentId, state, nil
	}
}

func resourceAwsQuickSightIAMPolicyAssignmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).quicksightconn

	awsAccountID, namespace, assignmentName, err := resourceAwsQuickSightIAMPolicyAssignmentParseID(d.Id())
	if err != nil {
		return err
	}

	descOpts := &quicksight.DescribeIAMPolicyAssignmentInput{
		AssignmentName: aws.String(assignmentName),
		AwsAccountId:   aws.String(awsAccountID),
		Namespace:      aws.String(namespace),
	}

	resp, err := conn.DescribeIAMPolicyAssignment(descOpts)
	if isAWSErr(err, quicksight.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] QuickSight IAM Policy Assignment %s not found", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error describing QuickSight IAM Policy Assignment (%s): %s", d.Id(), err)
	}

	d.Set("aws_account_id", resp.IAMPolicyAssignment.AwsAccountId)
	d.Set("namespace", namespace)
	d.Set("assignment_id", resp.IAMPolicyAssignment.AssignmentId)
	d.Set("assignment_name", resp.IAMPolicyAssignment.AssignmentName)
	d.Set("assignment_status", resp.IAMPolicyAssignment.AssignmentStatus)
	d.Set("identities", resp.IAMPolicyAssignment.Identities)
	d.Set("groups", resp.IAMPolicyAssignment.Identities["group"])
	d.Set("users", resp.IAMPolicyAssignment.Identities["user"])
	d.Set("policy_arn", resp.IAMPolicyAssignment.PolicyArn)

	return nil
}

func resourceAwsQuickSightIAMPolicyAssignmentUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).quicksightconn
	awsAccountID, namespace, assignmentName, err := resourceAwsQuickSightIAMPolicyAssignmentParseID(d.Id())
	if err != nil {
		return err
	}

	identities := make(map[string][]*string)
	if groupAttr := d.Get("groups").(*schema.Set); groupAttr.Len() > 0 {
		identities["Group"] = expandStringList(groupAttr.List())
	}
	if userAttr := d.Get("users").(*schema.Set); userAttr.Len() > 0 {
		identities["User"] = expandStringList(userAttr.List())
	}

	updateOpts := &quicksight.UpdateIAMPolicyAssignmentInput{
		AssignmentName:   aws.String(assignmentName),
		AssignmentStatus: aws.String(d.Get("assignment_status").(string)),
		AwsAccountId:     aws.String(awsAccountID),
		Identities:       identities,
		Namespace:        aws.String(namespace),
		PolicyArn:        aws.String(d.Get("policy_arn").(string)),
	}

	_, err = conn.UpdateIAMPolicyAssignment(updateOpts)
	if isAWSErr(err, quicksight.ErrCodeResourceNotFoundException, "") {
		log.Printf("[ERROR] QuickSight IAM Policy Assignment %s not found", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error updating QuickSight IAM Policy Assignment %s: %s", d.Id(), err)
	}
	return resourceAwsQuickSightIAMPolicyAssignmentRead(d, meta)
}

func resourceAwsQuickSightIAMPolicyAssignmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).quicksightconn

	awsAccountID, namespace, assignmentName, err := resourceAwsQuickSightIAMPolicyAssignmentParseID(d.Id())
	if err != nil {
		return err
	}

	deleteOpts := &quicksight.DeleteIAMPolicyAssignmentInput{
		AssignmentName: aws.String(assignmentName),
		AwsAccountId:   aws.String(awsAccountID),
		Namespace:      aws.String(namespace),
	}

	if _, err := conn.DeleteIAMPolicyAssignment(deleteOpts); err != nil {
		if isAWSErr(err, quicksight.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("Error deleting QuickSight IAM Policy Assignment %s: %s", d.Id(), err)
	}

	return nil
}

func resourceAwsQuickSightIAMPolicyAssignmentParseID(id string) (string, string, string, error) {
	parts := strings.SplitN(id, "/", 3)
	if len(parts) < 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of ID (%s), expected AWS_ACCOUNT_ID/NAMESPACE/GROUP_NAME", id)
	}
	return parts[0], parts[1], parts[2], nil
}
