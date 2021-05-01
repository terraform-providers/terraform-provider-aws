package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsIamGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIamGroupCreate,
		Read:   resourceAwsIamGroupRead,
		Update: resourceAwsIamGroupUpdate,
		Delete: resourceAwsIamGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"unique_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^[0-9A-Za-z=,.@\-_+]+$`),
					"must only contain alphanumeric characters, hyphens, underscores, commas, periods, @ symbols, plus and equals signs",
				),
			},
			"path": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "/",
			},
		},
	}
}

func resourceAwsIamGroupCreate(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn
	name := d.Get("name").(string)
	path := d.Get("path").(string)

	request := &iam.CreateGroupInput{
		Path:      aws.String(path),
		GroupName: aws.String(name),
	}

	createResp, err := iamconn.CreateGroup(request)
	if err != nil {
		return fmt.Errorf("Error creating IAM Group %s: %s", name, err)
	}
	d.SetId(aws.StringValue(createResp.Group.GroupName))

	return resourceAwsIamGroupRead(d, meta)
}

func resourceAwsIamGroupRead(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

	request := &iam.GetGroupInput{
		GroupName: aws.String(d.Id()),
	}

	var getResp *iam.GetGroupOutput

	err := resource.Retry(waiter.PropagationTimeout, func() *resource.RetryError {
		var err error

		getResp, err = iamconn.GetGroup(request)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		getResp, err = iamconn.GetGroup(request)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		log.Printf("[WARN] IAM Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading IAM Group (%s): %w", d.Id(), err)
	}

	if getResp == nil || getResp.Group == nil {
		return fmt.Errorf("error reading IAM Group (%s): empty response", d.Id())
	}

	group := getResp.Group

	if err := d.Set("name", group.GroupName); err != nil {
		return err
	}
	if err := d.Set("arn", group.Arn); err != nil {
		return err
	}
	if err := d.Set("path", group.Path); err != nil {
		return err
	}
	if err := d.Set("unique_id", group.GroupId); err != nil {
		return err
	}
	return nil
}

func resourceAwsIamGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	if d.HasChanges("name", "path") {
		iamconn := meta.(*AWSClient).iamconn
		on, nn := d.GetChange("name")
		_, np := d.GetChange("path")

		request := &iam.UpdateGroupInput{
			GroupName:    aws.String(on.(string)),
			NewGroupName: aws.String(nn.(string)),
			NewPath:      aws.String(np.(string)),
		}
		_, err := iamconn.UpdateGroup(request)
		if err != nil {
			return fmt.Errorf("Error updating IAM Group %s: %s", d.Id(), err)
		}
		d.SetId(nn.(string))
		return resourceAwsIamGroupRead(d, meta)
	}
	return nil
}

func resourceAwsIamGroupDelete(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

	request := &iam.DeleteGroupInput{
		GroupName: aws.String(d.Id()),
	}

	if _, err := iamconn.DeleteGroup(request); err != nil {
		return fmt.Errorf("Error deleting IAM Group %s: %s", d.Id(), err)
	}
	return nil
}

func deleteAwsIamGroupPolicyAttachments(conn *iam.IAM, groupName string) error {
	var attachedPolicies []*iam.AttachedPolicy
	input := &iam.ListAttachedGroupPoliciesInput{
		GroupName: aws.String(groupName),
	}

	err := conn.ListAttachedGroupPoliciesPages(input, func(page *iam.ListAttachedGroupPoliciesOutput, lastPage bool) bool {
		attachedPolicies = append(attachedPolicies, page.AttachedPolicies...)

		return !lastPage
	})

	if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing IAM Group (%s) policy attachments for deletion: %w", groupName, err)
	}

	for _, attachedPolicy := range attachedPolicies {
		input := &iam.DetachGroupPolicyInput{
			GroupName: aws.String(groupName),
			PolicyArn: attachedPolicy.PolicyArn,
		}

		_, err := conn.DetachGroupPolicy(input)

		if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
			continue
		}

		if err != nil {
			return fmt.Errorf("error detaching IAM Group (%s) policy (%s): %w", groupName, aws.StringValue(attachedPolicy.PolicyArn), err)
		}
	}

	return nil
}

func deleteAwsIamGroupPolicies(conn *iam.IAM, groupName string) error {
	var inlinePolicies []*string
	input := &iam.ListGroupPoliciesInput{
		GroupName: aws.String(groupName),
	}

	err := conn.ListGroupPoliciesPages(input, func(page *iam.ListGroupPoliciesOutput, lastPage bool) bool {
		inlinePolicies = append(inlinePolicies, page.PolicyNames...)
		return !lastPage
	})

	if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing IAM Group (%s) inline policies for deletion: %w", groupName, err)
	}

	for _, policyName := range inlinePolicies {
		input := &iam.DeleteGroupPolicyInput{
			GroupName:  aws.String(groupName),
			PolicyName: policyName,
		}

		_, err := conn.DeleteGroupPolicy(input)

		if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
			continue
		}

		if err != nil {
			return fmt.Errorf("error deleting IAM Group (%s) inline policy (%s): %w", groupName, aws.StringValue(policyName), err)
		}
	}

	return nil
}
