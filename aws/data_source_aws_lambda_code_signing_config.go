package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsLambdaCodeSigningConfig() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLambdaCodeSigningConfigRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"allowed_publishers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"signing_profile_version_arns": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Set: schema.HashString,
						},
					},
				},
			},
			"policies": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"untrusted_artifact_on_deployment": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"config_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsLambdaCodeSigningConfigRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn

	arn := d.Get("arn").(string)

	configOutput, err := conn.GetCodeSigningConfig(&lambda.GetCodeSigningConfigInput{
		CodeSigningConfigArn: aws.String(arn),
	})

	if err != nil {
		return fmt.Errorf("error getting Lambda code signing config (%s): %w", arn, err)
	}

	if configOutput == nil {
		return fmt.Errorf("error getting Lambda code signing config (%s): empty response", arn)
	}

	codeSigningConfig := configOutput.CodeSigningConfig
	if codeSigningConfig == nil {
		return fmt.Errorf("error getting Lambda code signing config (%s): empty CodeSigningConfig", arn)
	}

	if err := d.Set("config_id", codeSigningConfig.CodeSigningConfigId); err != nil {
		return fmt.Errorf("error setting lambda code signing config id: %w", err)
	}

	if err := d.Set("description", codeSigningConfig.Description); err != nil {
		return fmt.Errorf("error setting lambda code signing config description: %w", err)
	}

	if err := d.Set("last_modified", codeSigningConfig.LastModified); err != nil {
		return fmt.Errorf("error setting lambda code signing config last modified: %w", err)
	}

	if err := d.Set("allowed_publishers", flattenLambdaCodeSigningConfigAllowedPublishers(codeSigningConfig.AllowedPublishers)); err != nil {
		return fmt.Errorf("error setting lambda code signing config allowed publishers: %w", err)
	}

	if err := d.Set("policies", []interface{}{
		map[string]interface{}{
			"untrusted_artifact_on_deployment": codeSigningConfig.CodeSigningPolicies.UntrustedArtifactOnDeployment,
		},
	}); err != nil {
		return fmt.Errorf("error setting lambda code signing config code signing policies: %w", err)
	}

	d.SetId(aws.StringValue(codeSigningConfig.CodeSigningConfigArn))

	return nil
}
