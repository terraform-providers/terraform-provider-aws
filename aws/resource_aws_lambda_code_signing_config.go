package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsLambdaCodeSigningConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLambdaCodeSigningConfigCreate,
		Read:   resourceAwsLambdaCodeSigningConfigRead,
		Update: resourceAwsLambdaCodeSigningConfigUpdate,
		Delete: resourceAwsLambdaCodeSigningConfigDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"allowed_publishers": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"signing_profile_version_arns": {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							MaxItems: 20,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validateArn,
							},
							Set: schema.HashString,
						},
					},
				},
			},
			"policies": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"untrusted_artifact_on_deployment": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice(
								lambda.CodeSigningPolicy_Values(),
								false),
						},
					},
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"config_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsLambdaCodeSigningConfigCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn

	log.Printf("[DEBUG] Creating Lambda code signing config")

	configInput := &lambda.CreateCodeSigningConfigInput{
		AllowedPublishers: expandLambdaCodeSigningConfigAllowedPublishers(d.Get("allowed_publishers").([]interface{})),
		Description:       aws.String(d.Get("description").(string)),
	}

	if v, ok := d.GetOk("policies"); ok {
		codeSigningPolicies := v.([]interface{})
		policies := codeSigningPolicies[0].(map[string]interface{})
		configInput.CodeSigningPolicies = &lambda.CodeSigningPolicies{
			UntrustedArtifactOnDeployment: aws.String(policies["untrusted_artifact_on_deployment"].(string)),
		}
	}

	configOutput, err := conn.CreateCodeSigningConfig(configInput)
	if err != nil {
		return fmt.Errorf("error creating Lambda code signing config: %s", err)
	}

	if configOutput == nil || configOutput.CodeSigningConfig == nil {
		return fmt.Errorf("error creating Lambda code signing config: empty output")
	}
	d.SetId(aws.StringValue(configOutput.CodeSigningConfig.CodeSigningConfigArn))

	return resourceAwsLambdaCodeSigningConfigRead(d, meta)
}

func resourceAwsLambdaCodeSigningConfigRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn

	configOutput, err := conn.GetCodeSigningConfig(&lambda.GetCodeSigningConfigInput{
		CodeSigningConfigArn: aws.String(d.Id()),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Lambda Code Signing Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Lambda code signing config (%s): %s", d.Id(), err)
	}

	codeSigningConfig := configOutput.CodeSigningConfig
	if codeSigningConfig == nil {
		return fmt.Errorf("error getting Lambda code signing config (%s): empty CodeSigningConfig", d.Id())
	}

	if err := d.Set("arn", codeSigningConfig.CodeSigningConfigArn); err != nil {
		return fmt.Errorf("error setting lambda code signing config arn: %s", err)
	}

	if err := d.Set("config_id", codeSigningConfig.CodeSigningConfigId); err != nil {
		return fmt.Errorf("error setting lambda code signing config id: %s", err)
	}

	if err := d.Set("description", codeSigningConfig.Description); err != nil {
		return fmt.Errorf("error setting lambda code signing config description: %s", err)
	}

	if err := d.Set("last_modified", codeSigningConfig.LastModified); err != nil {
		return fmt.Errorf("error setting lambda code signing config last modified: %s", err)
	}

	if err := d.Set("allowed_publishers", flattenLambdaCodeSigningConfigAllowedPublishers(codeSigningConfig.AllowedPublishers)); err != nil {
		return fmt.Errorf("error setting lambda code signing config allowed publishers: %s", err)
	}

	if err := d.Set("policies", flattenCodeSigningPolicies(codeSigningConfig.CodeSigningPolicies)); err != nil {
		return fmt.Errorf("error setting lambda code signing config code signing policies: %s", err)
	}

	return nil
}

func resourceAwsLambdaCodeSigningConfigUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn

	configInput := &lambda.UpdateCodeSigningConfigInput{
		CodeSigningConfigArn: aws.String(d.Id()),
	}

	configUpdate := false
	if d.HasChange("allowed_publishers") {
		configInput.AllowedPublishers = expandLambdaCodeSigningConfigAllowedPublishers(d.Get("allowed_publishers").([]interface{}))
		configUpdate = true
	}
	if d.HasChange("policies") {
		codeSigningPolicies := d.Get("policies").([]interface{})
		policies := codeSigningPolicies[0].(map[string]interface{})
		configInput.CodeSigningPolicies = &lambda.CodeSigningPolicies{
			UntrustedArtifactOnDeployment: aws.String(policies["untrusted_artifact_on_deployment"].(string)),
		}
		configUpdate = true
	}
	if d.HasChange("description") {
		configInput.Description = aws.String(d.Get("description").(string))
		configUpdate = true
	}

	if configUpdate {
		log.Printf("[DEBUG] Updating Lambda code signing config: %#v", configInput)

		_, err := conn.UpdateCodeSigningConfig(configInput)
		if err != nil {
			return fmt.Errorf("error updating Lambda code signing config (%s): %s", d.Id(), err)
		}
	}

	return resourceAwsLambdaCodeSigningConfigRead(d, meta)
}

func resourceAwsLambdaCodeSigningConfigDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn

	_, err := conn.DeleteCodeSigningConfig(&lambda.DeleteCodeSigningConfigInput{
		CodeSigningConfigArn: aws.String(d.Id()),
	})

	if err != nil {
		if tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceNotFoundException) {
			return nil
		}
		return fmt.Errorf("error deleting Lambda code signing config (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Lambda code signing config %q deleted", d.Id())
	return nil
}

func expandLambdaCodeSigningConfigAllowedPublishers(allowedPublishers []interface{}) *lambda.AllowedPublishers {
	if len(allowedPublishers) == 0 || allowedPublishers[0] == nil {
		return nil
	}

	mAllowedPublishers := allowedPublishers[0].(map[string]interface{})

	return &lambda.AllowedPublishers{
		SigningProfileVersionArns: expandStringSet(mAllowedPublishers["signing_profile_version_arns"].(*schema.Set)),
	}
}

func flattenLambdaCodeSigningConfigAllowedPublishers(allowedPublishers *lambda.AllowedPublishers) []interface{} {
	if allowedPublishers == nil {
		return []interface{}{}
	}

	return []interface{}{map[string]interface{}{
		"signing_profile_version_arns": flattenStringSet(allowedPublishers.SigningProfileVersionArns),
	}}
}

func flattenCodeSigningPolicies(p *lambda.CodeSigningPolicies) []interface{} {
	if p == nil {
		return nil
	}
	m := map[string]interface{}{
		"untrusted_artifact_on_deployment": p.UntrustedArtifactOnDeployment,
	}

	return []interface{}{m}
}
