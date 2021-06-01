package aws

import (
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsImageBuilderImageRecipe() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsImageBuilderImageRecipeCreate,
		Read:   resourceAwsImageBuilderImageRecipeRead,
		Update: resourceAwsImageBuilderImageRecipeUpdate,
		Delete: resourceAwsImageBuilderImageRecipeDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"block_device_mapping": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"device_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
						"ebs": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"delete_on_termination": {
										// Use TypeString to allow an "unspecified" value,
										// since TypeBool only has true/false with false default.
										// The conversion from bare true/false values in
										// configurations to TypeString value is currently safe.
										Type:             schema.TypeString,
										Optional:         true,
										ForceNew:         true,
										DiffSuppressFunc: suppressEquivalentTypeStringBoolean,
										ValidateFunc:     validateTypeStringNullableBoolean,
									},
									"encrypted": {
										// Use TypeString to allow an "unspecified" value,
										// since TypeBool only has true/false with false default.
										// The conversion from bare true/false values in
										// configurations to TypeString value is currently safe.
										Type:             schema.TypeString,
										Optional:         true,
										ForceNew:         true,
										DiffSuppressFunc: suppressEquivalentTypeStringBoolean,
										ValidateFunc:     validateTypeStringNullableBoolean,
									},
									"iops": {
										Type:         schema.TypeInt,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntBetween(100, 10000),
									},
									"kms_key_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validateArn,
									},
									"snapshot_id": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
									"volume_size": {
										Type:         schema.TypeInt,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntBetween(1, 16000),
									},
									"volume_type": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringInSlice(imagebuilder.EbsVolumeType_Values(), false),
									},
								},
							},
						},
						"no_device": {
							// Use TypeBool to allow an "unspecified" value of false,
							// since the API uses an empty string ("") as true and
							// this is not compatible with TypeString's zero value.
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
						},
						"virtual_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
					},
				},
			},
			"component": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"component_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},
			"date_created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 126),
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parent_image": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 126),
			},
			"platform": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
			"version": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"working_directory": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
		},
	}
}

func resourceAwsImageBuilderImageRecipeCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	input := &imagebuilder.CreateImageRecipeInput{
		ClientToken: aws.String(resource.UniqueId()),
	}

	if v, ok := d.GetOk("block_device_mapping"); ok && v.(*schema.Set).Len() > 0 {
		input.BlockDeviceMappings = expandImageBuilderInstanceBlockDeviceMappings(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("component"); ok && len(v.([]interface{})) > 0 {
		input.Components = expandImageBuilderComponentConfigurations(v.([]interface{}))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("name"); ok {
		input.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("parent_image"); ok {
		input.ParentImage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tags"); ok {
		input.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().ImagebuilderTags()
	}

	if v, ok := d.GetOk("version"); ok {
		input.SemanticVersion = aws.String(v.(string))
	}
	if v, ok := d.GetOk("working_directory"); ok {
		input.WorkingDirectory = aws.String(v.(string))
	}

	output, err := conn.CreateImageRecipe(input)

	if err != nil {
		return fmt.Errorf("error creating Image Builder Image Recipe: %w", err)
	}

	if output == nil {
		return fmt.Errorf("error creating Image Builder Image Recipe: empty response")
	}

	d.SetId(aws.StringValue(output.ImageRecipeArn))

	return resourceAwsImageBuilderImageRecipeRead(d, meta)
}

func resourceAwsImageBuilderImageRecipeRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &imagebuilder.GetImageRecipeInput{
		ImageRecipeArn: aws.String(d.Id()),
	}

	output, err := conn.GetImageRecipe(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Image Builder Image Recipe (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting Image Builder Image Recipe (%s): %w", d.Id(), err)
	}

	if output == nil || output.ImageRecipe == nil {
		return fmt.Errorf("error getting Image Builder Image Recipe (%s): empty response", d.Id())
	}

	imageRecipe := output.ImageRecipe

	d.Set("arn", imageRecipe.Arn)
	d.Set("block_device_mapping", flattenImageBuilderInstanceBlockDeviceMappings(imageRecipe.BlockDeviceMappings))
	d.Set("component", flattenImageBuilderComponentConfigurations(imageRecipe.Components))
	d.Set("date_created", imageRecipe.DateCreated)
	d.Set("description", imageRecipe.Description)
	d.Set("name", imageRecipe.Name)
	d.Set("owner", imageRecipe.Owner)
	d.Set("parent_image", imageRecipe.ParentImage)
	d.Set("platform", imageRecipe.Platform)
	d.Set("tags", keyvaluetags.ImagebuilderKeyValueTags(imageRecipe.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map())
	d.Set("version", imageRecipe.Version)
	d.Set("working_directory", imageRecipe.WorkingDirectory)

	return nil
}

func resourceAwsImageBuilderImageRecipeUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.ImagebuilderUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags for Image Builder Image Recipe (%s): %w", d.Id(), err)
		}
	}

	return resourceAwsImageBuilderImageRecipeRead(d, meta)
}

func resourceAwsImageBuilderImageRecipeDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	input := &imagebuilder.DeleteImageRecipeInput{
		ImageRecipeArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteImageRecipe(input)

	if tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Image Builder Image Recipe (%s): %w", d.Id(), err)
	}

	return nil
}

func expandImageBuilderComponentConfiguration(tfMap map[string]interface{}) *imagebuilder.ComponentConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &imagebuilder.ComponentConfiguration{}

	if v, ok := tfMap["component_arn"].(string); ok && v != "" {
		apiObject.ComponentArn = aws.String(v)
	}

	return apiObject
}

func expandImageBuilderComponentConfigurations(tfList []interface{}) []*imagebuilder.ComponentConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*imagebuilder.ComponentConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandImageBuilderComponentConfiguration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandImageBuilderEbsInstanceBlockDeviceSpecification(tfMap map[string]interface{}) *imagebuilder.EbsInstanceBlockDeviceSpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &imagebuilder.EbsInstanceBlockDeviceSpecification{}

	if v, ok := tfMap["delete_on_termination"].(string); ok && v != "" {
		vBool, _ := strconv.ParseBool(v) // ignore error as previously validatated
		apiObject.DeleteOnTermination = aws.Bool(vBool)
	}

	if v, ok := tfMap["encrypted"].(string); ok && v != "" {
		vBool, _ := strconv.ParseBool(v) // ignore error as previously validatated
		apiObject.Encrypted = aws.Bool(vBool)
	}

	if v, ok := tfMap["iops"].(int); ok && v != 0 {
		apiObject.Iops = aws.Int64(int64(v))
	}

	if v, ok := tfMap["kms_key_id"].(string); ok && v != "" {
		apiObject.KmsKeyId = aws.String(v)
	}

	if v, ok := tfMap["snapshot_id"].(string); ok && v != "" {
		apiObject.SnapshotId = aws.String(v)
	}

	if v, ok := tfMap["volume_size"].(int); ok && v != 0 {
		apiObject.VolumeSize = aws.Int64(int64(v))
	}

	if v, ok := tfMap["volume_type"].(string); ok && v != "" {
		apiObject.VolumeType = aws.String(v)
	}

	return apiObject
}

func expandImageBuilderInstanceBlockDeviceMapping(tfMap map[string]interface{}) *imagebuilder.InstanceBlockDeviceMapping {
	if tfMap == nil {
		return nil
	}

	apiObject := &imagebuilder.InstanceBlockDeviceMapping{}

	if v, ok := tfMap["device_name"].(string); ok && v != "" {
		apiObject.DeviceName = aws.String(v)
	}

	if v, ok := tfMap["ebs"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Ebs = expandImageBuilderEbsInstanceBlockDeviceSpecification(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["no_device"].(bool); ok && v {
		apiObject.NoDevice = aws.String("")
	}

	if v, ok := tfMap["virtual_name"].(string); ok && v != "" {
		apiObject.VirtualName = aws.String(v)
	}

	return apiObject
}

func expandImageBuilderInstanceBlockDeviceMappings(tfList []interface{}) []*imagebuilder.InstanceBlockDeviceMapping {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*imagebuilder.InstanceBlockDeviceMapping

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandImageBuilderInstanceBlockDeviceMapping(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenImageBuilderComponentConfiguration(apiObject *imagebuilder.ComponentConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ComponentArn; v != nil {
		tfMap["component_arn"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenImageBuilderComponentConfigurations(apiObjects []*imagebuilder.ComponentConfiguration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenImageBuilderComponentConfiguration(apiObject))
	}

	return tfList
}

func flattenImageBuilderEbsInstanceBlockDeviceSpecification(apiObject *imagebuilder.EbsInstanceBlockDeviceSpecification) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DeleteOnTermination; v != nil {
		tfMap["delete_on_termination"] = strconv.FormatBool(aws.BoolValue(v))
	}

	if v := apiObject.Encrypted; v != nil {
		tfMap["encrypted"] = strconv.FormatBool(aws.BoolValue(v))
	}

	if v := apiObject.Iops; v != nil {
		tfMap["iops"] = aws.Int64Value(v)
	}

	if v := apiObject.KmsKeyId; v != nil {
		tfMap["kms_key_id"] = aws.StringValue(v)
	}

	if v := apiObject.SnapshotId; v != nil {
		tfMap["snapshot_id"] = aws.StringValue(v)
	}

	if v := apiObject.VolumeSize; v != nil {
		tfMap["volume_size"] = aws.Int64Value(v)
	}

	if v := apiObject.VolumeType; v != nil {
		tfMap["volume_type"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenImageBuilderInstanceBlockDeviceMapping(apiObject *imagebuilder.InstanceBlockDeviceMapping) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DeviceName; v != nil {
		tfMap["device_name"] = aws.StringValue(v)
	}

	if v := apiObject.Ebs; v != nil {
		tfMap["ebs"] = []interface{}{flattenImageBuilderEbsInstanceBlockDeviceSpecification(v)}
	}

	if v := apiObject.NoDevice; v != nil {
		tfMap["no_device"] = true
	}

	if v := apiObject.VirtualName; v != nil {
		tfMap["virtual_name"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenImageBuilderInstanceBlockDeviceMappings(apiObjects []*imagebuilder.InstanceBlockDeviceMapping) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenImageBuilderInstanceBlockDeviceMapping(apiObject))
	}

	return tfList
}
