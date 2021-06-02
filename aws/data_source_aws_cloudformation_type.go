package aws

import (
	"context"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceAwsCloudFormationType() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAwsCloudFormationTypeRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"default_version_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deprecated_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"documentation_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"execution_role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_default_version": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"logging_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"log_group_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"log_role_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"provisioning_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"schema": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(cloudformation.RegistryType_Values(), false),
			},
			"type_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(10, 204),
					validation.StringMatch(regexp.MustCompile(`[A-Za-z0-9]{2,64}::[A-Za-z0-9]{2,64}::[A-Za-z0-9]{2,64}(::MODULE){0,1}`), "three alphanumeric character sections separated by double colons (::)"),
				),
			},
			"version_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"visibility": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsCloudFormationTypeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).cfconn

	input := &cloudformation.DescribeTypeInput{}

	if v, ok := d.GetOk("arn"); ok {
		input.Arn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("type"); ok {
		input.Type = aws.String(v.(string))
	}

	if v, ok := d.GetOk("type_name"); ok {
		input.TypeName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("version_id"); ok {
		input.VersionId = aws.String(v.(string))
	}

	output, err := conn.DescribeTypeWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading CloudFormation Type: %w", err))
	}

	if output == nil {
		return diag.FromErr(fmt.Errorf("error reading CloudFormation Type: empty response"))
	}

	d.SetId(aws.StringValue(output.Arn))

	d.Set("arn", output.Arn)
	d.Set("default_version_id", output.DefaultVersionId)
	d.Set("deprecated_status", output.DeprecatedStatus)
	d.Set("description", output.Description)
	d.Set("documentation_url", output.DocumentationUrl)
	d.Set("execution_role_arn", output.ExecutionRoleArn)
	d.Set("is_default_version", output.IsDefaultVersion)
	if output.LoggingConfig != nil {
		if err := d.Set("logging_config", []interface{}{flattenCloudformationLoggingConfig(output.LoggingConfig)}); err != nil {
			return diag.FromErr(fmt.Errorf("error setting logging_config: %w", err))
		}
	} else {
		d.Set("logging_config", nil)
	}
	d.Set("provisioning_type", output.ProvisioningType)
	d.Set("schema", output.Schema)
	d.Set("source_url", output.SourceUrl)
	d.Set("type", output.Type)
	d.Set("type_name", output.TypeName)
	d.Set("visibility", output.Visibility)

	return nil
}
