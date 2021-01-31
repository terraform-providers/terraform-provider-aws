package aws

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsOrganizationsPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsOrganizationsPolicyCreate,
		ReadContext:   resourceAwsOrganizationsPolicyRead,
		UpdateContext: resourceAwsOrganizationsPolicyUpdate,
		DeleteContext: resourceAwsOrganizationsPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAwsOrganizationsPolicyImport,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"content": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
				ValidateFunc:     validation.StringIsJSON,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      organizations.PolicyTypeServiceControlPolicy,
				ValidateFunc: validation.StringInSlice(organizations.PolicyType_Values(), false),
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsOrganizationsPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).organizationsconn

	name := d.Get("name").(string)

	input := &organizations.CreatePolicyInput{
		Content:     aws.String(d.Get("content").(string)),
		Description: aws.String(d.Get("description").(string)),
		Name:        aws.String(name),
		Type:        aws.String(d.Get("type").(string)),
		Tags:        keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().OrganizationsTags(),
	}

	log.Printf("[DEBUG] Creating Organizations Policy (%s): %v", name, input)

	var err error
	var resp *organizations.CreatePolicyOutput
	err = resource.Retry(4*time.Minute, func() *resource.RetryError {
		resp, err = conn.CreatePolicy(input)

		if err != nil {
			if isAWSErr(err, organizations.ErrCodeFinalizingOrganizationException, "") {
				log.Printf("[DEBUG] Retrying creating Organizations Policy (%s): %s", name, err)
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})
	if isResourceTimeoutError(err) {
		resp, err = conn.CreatePolicy(input)
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Organizations Policy (%s): %w", name, err))
	}

	d.SetId(aws.StringValue(resp.Policy.PolicySummary.Id))

	return resourceAwsOrganizationsPolicyRead(ctx, d, meta)
}

func resourceAwsOrganizationsPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).organizationsconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &organizations.DescribePolicyInput{
		PolicyId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Organizations policy: %s", input)
	resp, err := conn.DescribePolicy(input)
	if err != nil {
		if isAWSErr(err, organizations.ErrCodePolicyNotFoundException, "") {
			log.Printf("[WARN] Organizations policy does not exist, removing from state: %s", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf("error reading Organizations Policy (%s): %w", d.Id(), err))
	}

	if resp.Policy == nil || resp.Policy.PolicySummary == nil {
		log.Printf("[WARN] Organizations policy does not exist, removing from state: %s", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", resp.Policy.PolicySummary.Arn)
	d.Set("content", resp.Policy.Content)
	d.Set("description", resp.Policy.PolicySummary.Description)
	d.Set("name", resp.Policy.PolicySummary.Name)
	d.Set("type", resp.Policy.PolicySummary.Type)

	if aws.BoolValue(resp.Policy.PolicySummary.AwsManaged) {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "AWS-managed Organizations policies cannot be imported",
				Detail:   fmt.Sprintf("This resource should be removed from your Terraform state using `terraform state rm` (https://www.terraform.io/docs/commands/state/rm.html) and references should use the ID (%s) directly.", d.Id()),
			},
		}
	}

	tags, err := keyvaluetags.OrganizationsListTags(conn, d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("error listing tags for Organizations policy (%s): %w", d.Id(), err))
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags for Organizations policy (%s): %w", d.Id(), err))
	}

	return nil
}

func resourceAwsOrganizationsPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).organizationsconn

	input := &organizations.UpdatePolicyInput{
		PolicyId: aws.String(d.Id()),
	}

	if d.HasChange("content") {
		input.Content = aws.String(d.Get("content").(string))
	}

	if d.HasChange("description") {
		input.Description = aws.String(d.Get("description").(string))
	}

	if d.HasChange("name") {
		input.Name = aws.String(d.Get("name").(string))
	}

	log.Printf("[DEBUG] Updating Organizations Policy: %s", input)
	_, err := conn.UpdatePolicy(input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating Organizations policy (%s): %w", d.Id(), err))
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.OrganizationsUpdateTags(conn, d.Id(), o, n); err != nil {
			return diag.FromErr(fmt.Errorf("error updating tags for Organizations policy (%s): %w", d.Id(), err))
		}
	}

	return resourceAwsOrganizationsPolicyRead(ctx, d, meta)
}

func resourceAwsOrganizationsPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).organizationsconn

	input := &organizations.DeletePolicyInput{
		PolicyId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Organizations Policy: %s", input)
	_, err := conn.DeletePolicy(input)
	if err != nil {
		if isAWSErr(err, organizations.ErrCodePolicyNotFoundException, "") {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting Organizations policy (%s): %w", d.Id(), err))
	}
	return nil
}

func resourceAwsOrganizationsPolicyImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*AWSClient).organizationsconn

	input := &organizations.DescribePolicyInput{
		PolicyId: aws.String(d.Id()),
	}
	resp, err := conn.DescribePolicyWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	if aws.BoolValue(resp.Policy.PolicySummary.AwsManaged) {
		return nil, fmt.Errorf("AWS-managed Organizations policy (%s) cannot be imported. Use the policy ID directly in your configuration.", d.Id())
	}

	return []*schema.ResourceData{d}, nil
}
