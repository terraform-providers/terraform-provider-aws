package aws

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func dataSourceAwsIAMPolicy() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsIAMPolicyRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateArn,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"path": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"path_prefix": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policy_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsIAMPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	arn := d.Get("arn").(string)
	name := d.Get("name").(string)
	pathPrefix := d.Get("path_prefix").(string)

	var results []*iam.Policy

	// Handle IAM eventual consistency
	err := resource.Retry(waiter.PropagationTimeout, func() *resource.RetryError {
		var err error
		results, err = finder.Policies(conn, arn, name, pathPrefix)

		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		results, err = finder.Policies(conn, arn, name, pathPrefix)
	}

	if err != nil {
		return fmt.Errorf("error reading IAM policy (%s): %w", PolicySearchDetails(arn, name, pathPrefix), err)
	}

	if len(results) == 0 {
		return fmt.Errorf("no IAM policy found matching criteria (%s); try different search", PolicySearchDetails(arn, name, pathPrefix))
	}

	if len(results) > 1 {
		return fmt.Errorf("multiple IAM policies found matching criteria (%s); try different search", PolicySearchDetails(arn, name, pathPrefix))
	}

	policy := results[0]
	policyArn := aws.StringValue(policy.Arn)

	d.SetId(policyArn)

	d.Set("arn", policyArn)
	d.Set("name", policy.PolicyName)
	d.Set("path", policy.Path)
	d.Set("policy_id", policy.PolicyId)

	if err := d.Set("tags", keyvaluetags.IamKeyValueTags(policy.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	// Retrieve policy description
	policyInput := &iam.GetPolicyInput{
		PolicyArn: policy.Arn,
	}

	policyOutput, err := conn.GetPolicy(policyInput)

	if err != nil {
		return fmt.Errorf("error reading IAM policy (%s): %w", policyArn, err)
	}

	if policyOutput == nil || policyOutput.Policy == nil {
		return fmt.Errorf("error reading IAM policy (%s): empty output", policyArn)
	}

	d.Set("description", policyOutput.Policy.Description)

	// Retrieve policy
	policyVersionInput := &iam.GetPolicyVersionInput{
		PolicyArn: policy.Arn,
		VersionId: policy.DefaultVersionId,
	}

	// Handle IAM eventual consistency
	var policyVersionOutput *iam.GetPolicyVersionOutput
	err = resource.Retry(waiter.PropagationTimeout, func() *resource.RetryError {
		var err error
		policyVersionOutput, err = conn.GetPolicyVersion(policyVersionInput)

		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		policyVersionOutput, err = conn.GetPolicyVersion(policyVersionInput)
	}

	if err != nil {
		return fmt.Errorf("error reading IAM Policy (%s) version: %w", policyArn, err)
	}

	if policyVersionOutput == nil || policyVersionOutput.PolicyVersion == nil {
		return fmt.Errorf("error reading IAM Policy (%s) version: empty output", policyArn)
	}

	policyVersion := policyVersionOutput.PolicyVersion

	var policyDocument string
	if policyVersion != nil {
		policyDocument, err = url.QueryUnescape(aws.StringValue(policyVersion.Document))
		if err != nil {
			return fmt.Errorf("error parsing IAM Policy (%s) document: %w", policyArn, err)
		}
	}

	d.Set("policy", policyDocument)

	return nil
}

// PolicySearchDetails returns the configured search criteria as a printable string
func PolicySearchDetails(arn, name, pathPrefix string) string {
	var policyDetails []string
	if arn != "" {
		policyDetails = append(policyDetails, fmt.Sprintf("ARN: %s", arn))
	}
	if name != "" {
		policyDetails = append(policyDetails, fmt.Sprintf("Name: %s", name))
	}
	if pathPrefix != "" {
		policyDetails = append(policyDetails, fmt.Sprintf("PathPrefix: %s", pathPrefix))
	}

	return strings.Join(policyDetails, ", ")
}
