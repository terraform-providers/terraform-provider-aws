package aws

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tfcloudformation "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudformation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudformation/waiter"
)

func resourceAwsCloudFormationType() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAwsCloudFormationTypeCreate,
		DeleteWithoutTimeout: resourceAwsCloudFormationTypeDelete,
		ReadWithoutTimeout:   resourceAwsCloudFormationTypeRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
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
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"is_default_version": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"logging_config": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"log_group_name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 512),
								validation.StringMatch(regexp.MustCompile(`[\.\-_/#A-Za-z0-9]+`), "must contain only alphanumeric, period, hyphen, forward slash, and octothorp characters"),
							),
						},
						"log_role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validateArn,
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
			"schema_handler_package": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^(https|s3)\:\/\/.+`), "must begin with s3:// or https://"),
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(cloudformation.RegistryType_Values(), false),
			},
			"type_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(10, 204),
					validation.StringMatch(regexp.MustCompile(`[A-Za-z0-9]{2,64}::[A-Za-z0-9]{2,64}::[A-Za-z0-9]{2,64}(::MODULE){0,1}`), "three alphanumeric character sections separated by double colons (::)"),
				),
			},
			"version_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"visibility": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsCloudFormationTypeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).cfconn
	typeName := d.Get("type_name").(string)
	input := &cloudformation.RegisterTypeInput{
		ClientRequestToken: aws.String(resource.UniqueId()),
		TypeName:           aws.String(typeName),
	}

	if v, ok := d.GetOk("execution_role_arn"); ok {
		input.ExecutionRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("logging_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.LoggingConfig = expandCloudformationLoggingConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("schema_handler_package"); ok {
		input.SchemaHandlerPackage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("type"); ok {
		input.Type = aws.String(v.(string))
	}

	output, err := conn.RegisterTypeWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error registering CloudFormation Type (%s): %w", typeName, err))
	}

	if output == nil || output.RegistrationToken == nil {
		return diag.FromErr(fmt.Errorf("error registering CloudFormation Type (%s): empty response", typeName))
	}

	registrationOutput, err := waiter.TypeRegistrationProgressStatusComplete(ctx, conn, aws.StringValue(output.RegistrationToken))

	if err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for CloudFormation Type (%s) registration: %w", typeName, err))
	}

	// Type Version ARN is not available until after registration is complete
	d.SetId(aws.StringValue(registrationOutput.TypeVersionArn))

	return resourceAwsCloudFormationTypeRead(ctx, d, meta)
}

func resourceAwsCloudFormationTypeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).cfconn

	input := &cloudformation.DescribeTypeInput{
		Arn: aws.String(d.Id()),
	}

	output, err := conn.DescribeTypeWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, cloudformation.ErrCodeTypeNotFoundException) {
		log.Printf("[WARN] CloudFormation Type (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading CloudFormation Type (%s): %w", d.Id(), err))
	}

	if output == nil {
		return diag.FromErr(fmt.Errorf("error reading CloudFormation Type (%s): empty response", d.Id()))
	}

	if aws.StringValue(output.DeprecatedStatus) == cloudformation.DeprecatedStatusDeprecated {
		if d.IsNewResource() {
			return diag.FromErr(fmt.Errorf("error reading CloudFormation Type (%s): deprecated after creation", d.Id()))
		}

		log.Printf("[WARN] CloudFormation Type (%s) %s, removing from state", d.Id(), aws.StringValue(output.DeprecatedStatus))
		d.SetId("")
		return nil
	}

	typeARN, versionID, err := tfcloudformation.TypeVersionARNToTypeARNAndVersionID(d.Id())

	if err != nil {
		return diag.FromErr(fmt.Errorf("error parsing CloudFormation Type (%s) ARN: %w", d.Id(), err))
	}

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
	d.Set("type_arn", typeARN)
	d.Set("type_name", output.TypeName)
	d.Set("version_id", versionID)
	d.Set("visibility", output.Visibility)

	return nil
}

func resourceAwsCloudFormationTypeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).cfconn

	input := &cloudformation.DeregisterTypeInput{
		Arn: aws.String(d.Id()),
	}

	_, err := conn.DeregisterTypeWithContext(ctx, input)

	// Must deregister type if removing final LIVE version. This error can also occur
	// when the type is already DEPRECATED.
	if tfawserr.ErrMessageContains(err, cloudformation.ErrCodeCFNRegistryException, "is the default version and cannot be deregistered") {
		typeARN, _, err := tfcloudformation.TypeVersionARNToTypeARNAndVersionID(d.Id())

		if err != nil {
			return diag.FromErr(fmt.Errorf("error parsing CloudFormation Type (%s) ARN: %w", d.Id(), err))
		}

		input := &cloudformation.ListTypeVersionsInput{
			Arn:              aws.String(typeARN),
			DeprecatedStatus: aws.String(cloudformation.DeprecatedStatusLive),
		}

		var typeVersionSummaries []*cloudformation.TypeVersionSummary

		err = conn.ListTypeVersionsPagesWithContext(ctx, input, func(page *cloudformation.ListTypeVersionsOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			typeVersionSummaries = append(typeVersionSummaries, page.TypeVersionSummaries...)

			if len(typeVersionSummaries) > 1 {
				return false
			}

			return !lastPage
		})

		if err != nil {
			return diag.FromErr(fmt.Errorf("error listing CloudFormation Type (%s) Versions: %w", d.Id(), err))
		}

		if len(typeVersionSummaries) <= 1 {
			input := &cloudformation.DeregisterTypeInput{
				Arn: aws.String(typeARN),
			}

			_, err := conn.DeregisterTypeWithContext(ctx, input)

			if tfawserr.ErrCodeEquals(err, cloudformation.ErrCodeTypeNotFoundException) {
				return nil
			}

			if err != nil {
				return diag.FromErr(fmt.Errorf("error deregistering CloudFormation Type (%s): %w", d.Id(), err))
			}

			return nil
		}
	}

	if tfawserr.ErrCodeEquals(err, cloudformation.ErrCodeTypeNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deregistering CloudFormation Type (%s): %w", d.Id(), err))
	}

	return nil
}

func expandCloudformationLoggingConfig(tfMap map[string]interface{}) *cloudformation.LoggingConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudformation.LoggingConfig{}

	if v, ok := tfMap["log_group_name"].(string); ok && v != "" {
		apiObject.LogGroupName = aws.String(v)
	}

	if v, ok := tfMap["log_role_arn"].(string); ok && v != "" {
		apiObject.LogRoleArn = aws.String(v)
	}

	return apiObject
}

func flattenCloudformationLoggingConfig(apiObject *cloudformation.LoggingConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.LogGroupName; v != nil {
		tfMap["log_group_name"] = aws.StringValue(v)
	}

	if v := apiObject.LogRoleArn; v != nil {
		tfMap["log_role_arn"] = aws.StringValue(v)
	}

	return tfMap
}
