package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsIamGroupPolicyAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIamGroupPolicyAttachmentCreate,
		Read:   resourceAwsIamGroupPolicyAttachmentRead,
		Delete: resourceAwsIamGroupPolicyAttachmentDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsIamGroupPolicyAttachmentImport,
		},

		Schema: map[string]*schema.Schema{
			"group": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"policy_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsIamGroupPolicyAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn

	group := d.Get("group").(string)
	arn := d.Get("policy_arn").(string)

	err := attachPolicyToGroup(conn, group, arn)
	if err != nil {
		return fmt.Errorf("Error attaching policy %s to IAM group %s: %v", arn, group, err)
	}

	//lintignore:R016 // Allow legacy unstable ID usage in managed resource
	d.SetId(resource.PrefixedUniqueId(fmt.Sprintf("%s-", group)))

	return resourceAwsIamGroupPolicyAttachmentRead(d, meta)
}

func resourceAwsIamGroupPolicyAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn
	group := d.Get("group").(string)
	arn := d.Get("policy_arn").(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive
	id := fmt.Sprintf("%s:%s", group, arn)

	var attachedPolicy *iam.AttachedPolicy

	err := resource.Retry(waiter.PropagationTimeout, func() *resource.RetryError {
		var err error

		attachedPolicy, err = finder.GroupAttachedPolicy(conn, group, arn)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if d.IsNewResource() && attachedPolicy == nil {
			return resource.RetryableError(&resource.NotFoundError{
				LastError: fmt.Errorf("IAM Group Managed Policy Attachment (%s) not found", id),
			})
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		attachedPolicy, err = finder.GroupAttachedPolicy(conn, group, arn)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		log.Printf("[WARN] IAM User Managed Policy Attachment (%s) not found, removing from state", id)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading IAM Group Managed Policy Attachment (%s): %w", id, err)
	}

	if attachedPolicy == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error reading IAM User Managed Policy Attachment (%s): not found after creation", id)
		}

		log.Printf("[WARN] IAM Group Managed Policy Attachment (%s) not found, removing from state", id)
		d.SetId("")
		return nil
	}

	return nil
}

func resourceAwsIamGroupPolicyAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn
	group := d.Get("group").(string)
	arn := d.Get("policy_arn").(string)

	err := detachPolicyFromGroup(conn, group, arn)
	if err != nil {
		return fmt.Errorf("Error removing policy %s from IAM Group %s: %v", arn, group, err)
	}
	return nil
}

func resourceAwsIamGroupPolicyAttachmentImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.SplitN(d.Id(), "/", 2)
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected <group-name>/<policy_arn>", d.Id())
	}
	groupName := idParts[0]
	policyARN := idParts[1]
	d.Set("group", groupName)
	d.Set("policy_arn", policyARN)
	d.SetId(fmt.Sprintf("%s-%s", groupName, policyARN))
	return []*schema.ResourceData{d}, nil
}

func attachPolicyToGroup(conn *iam.IAM, group string, arn string) error {
	_, err := conn.AttachGroupPolicy(&iam.AttachGroupPolicyInput{
		GroupName: aws.String(group),
		PolicyArn: aws.String(arn),
	})
	return err
}

func detachPolicyFromGroup(conn *iam.IAM, group string, arn string) error {
	_, err := conn.DetachGroupPolicy(&iam.DetachGroupPolicyInput{
		GroupName: aws.String(group),
		PolicyArn: aws.String(arn),
	})
	return err
}
