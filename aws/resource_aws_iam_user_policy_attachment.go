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

func resourceAwsIamUserPolicyAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIamUserPolicyAttachmentCreate,
		Read:   resourceAwsIamUserPolicyAttachmentRead,
		Delete: resourceAwsIamUserPolicyAttachmentDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsIamUserPolicyAttachmentImport,
		},

		Schema: map[string]*schema.Schema{
			"user": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"policy_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsIamUserPolicyAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn

	user := d.Get("user").(string)
	arn := d.Get("policy_arn").(string)

	err := attachPolicyToUser(conn, user, arn)
	if err != nil {
		return fmt.Errorf("Error attaching policy %s to IAM User %s: %v", arn, user, err)
	}

	//lintignore:R016 // Allow legacy unstable ID usage in managed resource
	d.SetId(resource.PrefixedUniqueId(fmt.Sprintf("%s-", user)))

	return resourceAwsIamUserPolicyAttachmentRead(d, meta)
}

func resourceAwsIamUserPolicyAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn
	user := d.Get("user").(string)
	arn := d.Get("policy_arn").(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive
	id := fmt.Sprintf("%s:%s", user, arn)

	var attachedPolicy *iam.AttachedPolicy

	err := resource.Retry(waiter.PropagationTimeout, func() *resource.RetryError {
		var err error

		attachedPolicy, err = finder.UserAttachedPolicy(conn, user, arn)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if d.IsNewResource() && attachedPolicy == nil {
			return resource.RetryableError(&resource.NotFoundError{
				LastError: fmt.Errorf("IAM User Managed Policy Attachment (%s) not found", id),
			})
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		attachedPolicy, err = finder.UserAttachedPolicy(conn, user, arn)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		log.Printf("[WARN] IAM User Managed Policy Attachment (%s) not found, removing from state", id)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading IAM User Managed Policy Attachment (%s): %w", id, err)
	}

	if attachedPolicy == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error reading IAM User Managed Policy Attachment (%s): not found after creation", id)
		}

		log.Printf("[WARN] IAM User Managed Policy Attachment (%s) not found, removing from state", id)
		d.SetId("")
		return nil
	}

	return nil
}

func resourceAwsIamUserPolicyAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn
	user := d.Get("user").(string)
	arn := d.Get("policy_arn").(string)

	err := detachPolicyFromUser(conn, user, arn)
	if err != nil {
		return fmt.Errorf("Error removing policy %s from IAM User %s: %v", arn, user, err)
	}
	return nil
}

func resourceAwsIamUserPolicyAttachmentImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.SplitN(d.Id(), "/", 2)
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected <user-name>/<policy_arn>", d.Id())
	}

	userName := idParts[0]
	policyARN := idParts[1]

	d.Set("user", userName)
	d.Set("policy_arn", policyARN)
	d.SetId(fmt.Sprintf("%s-%s", userName, policyARN))

	return []*schema.ResourceData{d}, nil
}

func attachPolicyToUser(conn *iam.IAM, user string, arn string) error {
	_, err := conn.AttachUserPolicy(&iam.AttachUserPolicyInput{
		UserName:  aws.String(user),
		PolicyArn: aws.String(arn),
	})
	return err
}

func detachPolicyFromUser(conn *iam.IAM, user string, arn string) error {
	_, err := conn.DetachUserPolicy(&iam.DetachUserPolicyInput{
		UserName:  aws.String(user),
		PolicyArn: aws.String(arn),
	})
	return err
}
