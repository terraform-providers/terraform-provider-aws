package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsImageBuilderDistributionConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsImageBuilderDistributionConfigurationCreate,
		Read:   resourceAwsImageBuilderDistributionConfigurationRead,
		Update: resourceAwsImageBuilderDistributionConfigurationUpdate,
		Delete: resourceAwsImageBuilderDistributionConfigurationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"date_created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"date_updated": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"distribution": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ami_distribution_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ami_tags": tagsSchema(),
									"description": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 1024),
									},
									"kms_key_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 1024),
									},
									"launch_permission": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"user_groups": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validation.StringLenBetween(1, 1024),
													},
												},
												"user_ids": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validateAwsAccountId,
													},
												},
											},
										},
									},
									"name": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 127),
											validation.StringMatch(regexp.MustCompile(`^[-_A-Za-z0-9{][-_A-Za-z0-9\s:{}]+[-_A-Za-z0-9}]$`), "must contain only alphanumeric characters, periods, underscores, and hyphens"),
										),
									},
									"target_account_ids": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validateAwsAccountId,
										},
										MaxItems: 50,
									},
								},
							},
						},
						"license_configuration_arns": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validateArn,
							},
						},
						"region": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 1024),
						},
					},
				},
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 126),
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsImageBuilderDistributionConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	input := &imagebuilder.CreateDistributionConfigurationInput{
		ClientToken: aws.String(resource.UniqueId()),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("distribution"); ok && v.(*schema.Set).Len() > 0 {
		input.Distributions = expandImageBuilderDistributions(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("name"); ok {
		input.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tags"); ok && len(v.(map[string]interface{})) > 0 {
		input.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().ImagebuilderTags()
	}

	output, err := conn.CreateDistributionConfiguration(input)

	if err != nil {
		return fmt.Errorf("error creating Image Builder Distribution Configuration: %w", err)
	}

	if output == nil {
		return fmt.Errorf("error creating Image Builder Distribution Configuration: empty response")
	}

	d.SetId(aws.StringValue(output.DistributionConfigurationArn))

	return resourceAwsImageBuilderDistributionConfigurationRead(d, meta)
}

func resourceAwsImageBuilderDistributionConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &imagebuilder.GetDistributionConfigurationInput{
		DistributionConfigurationArn: aws.String(d.Id()),
	}

	output, err := conn.GetDistributionConfiguration(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Image Builder Distribution Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting Image Builder Distribution Configuration (%s): %w", d.Id(), err)
	}

	if output == nil || output.DistributionConfiguration == nil {
		return fmt.Errorf("error getting Image Builder Distribution Configuration (%s): empty response", d.Id())
	}

	distributionConfiguration := output.DistributionConfiguration

	d.Set("arn", distributionConfiguration.Arn)
	d.Set("date_created", distributionConfiguration.DateCreated)
	d.Set("date_updated", distributionConfiguration.DateUpdated)
	d.Set("description", distributionConfiguration.Description)
	d.Set("distribution", flattenImageBuilderDistributions(distributionConfiguration.Distributions))
	d.Set("name", distributionConfiguration.Name)
	d.Set("tags", keyvaluetags.ImagebuilderKeyValueTags(distributionConfiguration.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map())

	return nil
}

func resourceAwsImageBuilderDistributionConfigurationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	if d.HasChanges("description", "distribution") {
		input := &imagebuilder.UpdateDistributionConfigurationInput{
			DistributionConfigurationArn: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("distribution"); ok && v.(*schema.Set).Len() > 0 {
			input.Distributions = expandImageBuilderDistributions(v.(*schema.Set).List())
		}

		log.Printf("[DEBUG] UpdateDistributionConfiguration: %#v", input)
		_, err := conn.UpdateDistributionConfiguration(input)

		if err != nil {
			return fmt.Errorf("error updating Image Builder Distribution Configuration (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.ImagebuilderUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags for Image Builder Distribution Configuration (%s): %w", d.Id(), err)
		}
	}

	return resourceAwsImageBuilderDistributionConfigurationRead(d, meta)
}

func resourceAwsImageBuilderDistributionConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	input := &imagebuilder.DeleteDistributionConfigurationInput{
		DistributionConfigurationArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteDistributionConfiguration(input)

	if tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Image Builder Distribution Config (%s): %w", d.Id(), err)
	}

	return nil
}

func expandImageBuilderAmiDistributionConfiguration(tfMap map[string]interface{}) *imagebuilder.AmiDistributionConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &imagebuilder.AmiDistributionConfiguration{}

	if v, ok := tfMap["ami_tags"].(map[string]interface{}); ok && len(v) > 0 {
		apiObject.AmiTags = stringMapToPointers(v)
	}

	if v, ok := tfMap["description"].(string); ok && v != "" {
		apiObject.Description = aws.String(v)
	}

	if v, ok := tfMap["kms_key_id"].(string); ok && v != "" {
		apiObject.KmsKeyId = aws.String(v)
	}

	if v, ok := tfMap["launch_permission"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.LaunchPermission = expandImageBuilderLaunchPermissionConfiguration(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap["target_account_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.TargetAccountIds = expandStringSet(v)
	}

	return apiObject
}

func expandImageBuilderDistribution(tfMap map[string]interface{}) *imagebuilder.Distribution {
	if tfMap == nil {
		return nil
	}

	apiObject := &imagebuilder.Distribution{}

	if v, ok := tfMap["ami_distribution_configuration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.AmiDistributionConfiguration = expandImageBuilderAmiDistributionConfiguration(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["license_configuration_arns"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.LicenseConfigurationArns = expandStringSet(v)
	}

	if v, ok := tfMap["region"].(string); ok && v != "" {
		apiObject.Region = aws.String(v)
	}

	return apiObject
}

func expandImageBuilderDistributions(tfList []interface{}) []*imagebuilder.Distribution {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*imagebuilder.Distribution

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandImageBuilderDistribution(tfMap)

		if apiObject == nil {
			continue
		}

		// Prevent error: InvalidParameter: 1 validation error(s) found.
		//  - missing required field, UpdateDistributionConfigurationInput.Distributions[0].Region
		// Reference: https://github.com/hashicorp/terraform-plugin-sdk/issues/588
		if apiObject.Region == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandImageBuilderLaunchPermissionConfiguration(tfMap map[string]interface{}) *imagebuilder.LaunchPermissionConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &imagebuilder.LaunchPermissionConfiguration{}

	if v, ok := tfMap["user_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.UserIds = expandStringSet(v)
	}

	if v, ok := tfMap["user_groups"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.UserGroups = expandStringSet(v)
	}

	return apiObject
}

func flattenImageBuilderAmiDistributionConfiguration(apiObject *imagebuilder.AmiDistributionConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AmiTags; v != nil {
		tfMap["ami_tags"] = aws.StringValueMap(v)
	}

	if v := apiObject.Description; v != nil {
		tfMap["description"] = aws.StringValue(v)
	}

	if v := apiObject.KmsKeyId; v != nil {
		tfMap["kms_key_id"] = aws.StringValue(v)
	}

	if v := apiObject.LaunchPermission; v != nil {
		tfMap["launch_permission"] = []interface{}{flattenImageBuilderLaunchPermissionConfiguration(v)}
	}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.StringValue(v)
	}

	if v := apiObject.TargetAccountIds; v != nil {
		tfMap["target_account_ids"] = aws.StringValueSlice(v)
	}

	return tfMap
}

func flattenImageBuilderDistribution(apiObject *imagebuilder.Distribution) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AmiDistributionConfiguration; v != nil {
		tfMap["ami_distribution_configuration"] = []interface{}{flattenImageBuilderAmiDistributionConfiguration(v)}
	}

	if v := apiObject.LicenseConfigurationArns; v != nil {
		tfMap["license_configuration_arns"] = aws.StringValueSlice(v)
	}

	if v := apiObject.Region; v != nil {
		tfMap["region"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenImageBuilderDistributions(apiObjects []*imagebuilder.Distribution) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenImageBuilderDistribution(apiObject))
	}

	return tfList
}

func flattenImageBuilderLaunchPermissionConfiguration(apiObject *imagebuilder.LaunchPermissionConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.UserGroups; v != nil {
		tfMap["user_groups"] = aws.StringValueSlice(v)
	}

	if v := apiObject.UserIds; v != nil {
		tfMap["user_ids"] = aws.StringValueSlice(v)
	}

	return tfMap
}
