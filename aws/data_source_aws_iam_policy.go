package aws

import (
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func dataSourceAwsIAMPolicy() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsIAMPolicyRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"path": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
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

	input := &iam.GetPolicyInput{
		PolicyArn: aws.String(arn),
	}

	// Handle IAM eventual consistency
	var output *iam.GetPolicyOutput
	err := resource.Retry(waiter.PropagationTimeout, func() *resource.RetryError {
		var err error
		output, err = conn.GetPolicy(input)

		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.GetPolicy(input)
	}

	if err != nil {
		return fmt.Errorf("error reading IAM policy %s: %w", arn, err)
	}

	if output == nil || output.Policy == nil {
		return fmt.Errorf("error reading IAM policy %s: empty output", arn)
	}

	policy := output.Policy

	d.SetId(aws.StringValue(policy.Arn))

	d.Set("arn", policy.Arn)
	d.Set("description", policy.Description)
	d.Set("name", policy.PolicyName)
	d.Set("path", policy.Path)
	d.Set("policy_id", policy.PolicyId)

	if err := d.Set("tags", keyvaluetags.IamKeyValueTags(policy.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	// Retrieve policy

	policyVersionInput := &iam.GetPolicyVersionInput{
		PolicyArn: aws.String(arn),
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
		return fmt.Errorf("error reading IAM Policy (%s) version: %w", arn, err)
	}

	if policyVersionOutput == nil || policyVersionOutput.PolicyVersion == nil {
		return fmt.Errorf("error reading IAM Policy (%s) version: empty output", arn)
	}

	policyVersion := policyVersionOutput.PolicyVersion

	var policyDocument string
	if policyVersion != nil {
		policyDocument, err = url.QueryUnescape(aws.StringValue(policyVersion.Document))
		if err != nil {
			return fmt.Errorf("error parsing IAM Policy (%s) document: %w", arn, err)
		}
	}

	d.Set("policy", policyDocument)

	return nil
}
