package aws

import (
	"fmt"
	"log"
	"net/url"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
)

func resourceAwsIamPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIamPolicyCreate,
		Read:   resourceAwsIamPolicyRead,
		Update: resourceAwsIamPolicyUpdate,
		Delete: resourceAwsIamPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"path": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "/",
				ForceNew: true,
			},
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validateIAMPolicyJson,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexp.MustCompile(`^[\w+=,.@-]*$`), "must match [\\w+=,.@-]"),
				),
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 96),
					validation.StringMatch(regexp.MustCompile(`^[\w+=,.@-]*$`), "must match [\\w+=,.@-]"),
				),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policy_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsIamPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn

	var name string
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		name = resource.PrefixedUniqueId(v.(string))
	} else {
		name = resource.UniqueId()
	}

	request := &iam.CreatePolicyInput{
		Description:    aws.String(d.Get("description").(string)),
		Path:           aws.String(d.Get("path").(string)),
		PolicyDocument: aws.String(d.Get("policy").(string)),
		PolicyName:     aws.String(name),
		Tags:           keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().IamTags(),
	}

	response, err := conn.CreatePolicy(request)
	if err != nil {
		return fmt.Errorf("Error creating IAM policy %s: %w", name, err)
	}

	d.SetId(aws.StringValue(response.Policy.Arn))

	return resourceAwsIamPolicyRead(d, meta)
}

func resourceAwsIamPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	getPolicyRequest := &iam.GetPolicyInput{
		PolicyArn: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Getting IAM Policy: %s", getPolicyRequest)

	// Handle IAM eventual consistency
	var getPolicyResponse *iam.GetPolicyOutput
	err := resource.Retry(waiter.PropagationTimeout, func() *resource.RetryError {
		var err error
		getPolicyResponse, err = conn.GetPolicy(getPolicyRequest)

		if d.IsNewResource() && isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if isResourceTimeoutError(err) {
		getPolicyResponse, err = conn.GetPolicy(getPolicyRequest)
	}
	if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
		log.Printf("[WARN] IAM Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error reading IAM policy %s: %w", d.Id(), err)
	}

	if getPolicyResponse == nil || getPolicyResponse.Policy == nil {
		log.Printf("[WARN] IAM Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	policyRes := getPolicyResponse.Policy
	d.Set("arn", policyRes.Arn)
	d.Set("description", policyRes.Description)
	d.Set("name", policyRes.PolicyName)
	d.Set("path", policyRes.Path)
	d.Set("policy_id", policyRes.PolicyId)

	if err := d.Set("tags", keyvaluetags.IamKeyValueTags(policyRes.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	// Retrieve policy

	getPolicyVersionRequest := &iam.GetPolicyVersionInput{
		PolicyArn: aws.String(d.Id()),
		VersionId: policyRes.DefaultVersionId,
	}
	log.Printf("[DEBUG] Getting IAM Policy Version: %s", getPolicyVersionRequest)

	// Handle IAM eventual consistency
	var getPolicyVersionResponse *iam.GetPolicyVersionOutput
	err = resource.Retry(waiter.PropagationTimeout, func() *resource.RetryError {
		var err error
		getPolicyVersionResponse, err = conn.GetPolicyVersion(getPolicyVersionRequest)

		if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if isResourceTimeoutError(err) {
		getPolicyVersionResponse, err = conn.GetPolicyVersion(getPolicyVersionRequest)
	}
	if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
		log.Printf("[WARN] IAM Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error reading IAM policy version %s: %w", d.Id(), err)
	}

	policy := ""
	if getPolicyVersionResponse != nil && getPolicyVersionResponse.PolicyVersion != nil {
		var err error
		policy, err = url.QueryUnescape(aws.StringValue(getPolicyVersionResponse.PolicyVersion.Document))
		if err != nil {
			return fmt.Errorf("error parsing policy: %w", err)
		}
	}

	d.Set("policy", policy)

	return nil
}

func resourceAwsIamPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn

	if d.HasChangeExcept("tags") {

		if err := iamPolicyPruneVersions(d.Id(), conn); err != nil {
			return err
		}

		request := &iam.CreatePolicyVersionInput{
			PolicyArn:      aws.String(d.Id()),
			PolicyDocument: aws.String(d.Get("policy").(string)),
			SetAsDefault:   aws.Bool(true),
		}

		if _, err := conn.CreatePolicyVersion(request); err != nil {
			return fmt.Errorf("Error updating IAM policy %s: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.IamPolicyUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags for IAM Policy (%s): %w", d.Id(), err)
		}
	}

	return resourceAwsIamPolicyRead(d, meta)
}

func resourceAwsIamPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn

	if err := iamPolicyDeleteNondefaultVersions(d.Id(), conn); err != nil {
		return err
	}

	request := &iam.DeletePolicyInput{
		PolicyArn: aws.String(d.Id()),
	}

	if _, err := conn.DeletePolicy(request); err != nil {
		if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
			return nil
		}
		return fmt.Errorf("Error deleting IAM policy %s: %w", d.Id(), err)
	}

	return nil
}

// iamPolicyPruneVersions deletes the oldest versions.
//
// Old versions are deleted until there are 4 or less remaining, which means at
// least one more can be created before hitting the maximum of 5.
//
// The default version is never deleted.

func iamPolicyPruneVersions(arn string, conn *iam.IAM) error {
	versions, err := iamPolicyListVersions(arn, conn)
	if err != nil {
		return err
	}
	if len(versions) < 5 {
		return nil
	}

	var oldestVersion *iam.PolicyVersion

	for _, version := range versions {
		if *version.IsDefaultVersion {
			continue
		}
		if oldestVersion == nil ||
			version.CreateDate.Before(*oldestVersion.CreateDate) {
			oldestVersion = version
		}
	}

	err1 := iamPolicyDeleteVersion(arn, aws.StringValue(oldestVersion.VersionId), conn)
	return err1
}

func iamPolicyDeleteNondefaultVersions(arn string, conn *iam.IAM) error {
	versions, err := iamPolicyListVersions(arn, conn)
	if err != nil {
		return err
	}

	for _, version := range versions {
		if *version.IsDefaultVersion {
			continue
		}
		if err := iamPolicyDeleteVersion(arn, aws.StringValue(version.VersionId), conn); err != nil {
			return err
		}
	}

	return nil
}

func iamPolicyDeleteVersion(arn, versionID string, conn *iam.IAM) error {
	request := &iam.DeletePolicyVersionInput{
		PolicyArn: aws.String(arn),
		VersionId: aws.String(versionID),
	}

	_, err := conn.DeletePolicyVersion(request)
	if err != nil {
		return fmt.Errorf("Error deleting version %s from IAM policy %s: %w", versionID, arn, err)
	}
	return nil
}

func iamPolicyListVersions(arn string, conn *iam.IAM) ([]*iam.PolicyVersion, error) {
	request := &iam.ListPolicyVersionsInput{
		PolicyArn: aws.String(arn),
	}

	response, err := conn.ListPolicyVersions(request)
	if err != nil {
		return nil, fmt.Errorf("Error listing versions for IAM policy %s: %w", arn, err)
	}
	return response.Versions, nil
}
