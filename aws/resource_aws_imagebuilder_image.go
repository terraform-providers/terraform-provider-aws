package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/imagebuilder/waiter"
)

func resourceAwsImageBuilderImage() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsImageBuilderImageCreate,
		Read:   resourceAwsImageBuilderImageRead,
		Update: resourceAwsImageBuilderImageUpdate,
		Delete: resourceAwsImageBuilderImageDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
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
			"distribution_configuration_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^arn:aws[^:]*:imagebuilder:[^:]+:(?:\d{12}|aws):distribution-configuration/[a-z0-9-_]+$`), "valid distribution configuration ARN must be provided"),
			},
			"enhanced_image_metadata_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  true,
			},
			"image_recipe_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^arn:aws[^:]*:imagebuilder:[^:]+:(?:\d{12}|aws):image-recipe/[a-z0-9-_]+/\d+\.\d+\.\d+$`), "valid image recipe ARN must be provided"),
			},
			"image_tests_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"image_tests_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
							Default:  true,
						},
						"timeout_minutes": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							Default:      720,
							ValidateFunc: validation.IntBetween(60, 1440),
						},
					},
				},
			},
			"infrastructure_configuration_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^arn:aws[^:]*:imagebuilder:[^:]+:(?:\d{12}|aws):infrastructure-configuration/[a-z0-9-_]+$`), "valid infrastructure configuration ARN must be provided"),
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"os_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"output_resources": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"amis": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"account_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"description": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"image": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"region": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"platform": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsImageBuilderImageCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	input := &imagebuilder.CreateImageInput{
		ClientToken:                  aws.String(resource.UniqueId()),
		EnhancedImageMetadataEnabled: aws.Bool(d.Get("enhanced_image_metadata_enabled").(bool)),
	}

	if v, ok := d.GetOk("distribution_configuration_arn"); ok {
		input.DistributionConfigurationArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("image_recipe_arn"); ok {
		input.ImageRecipeArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("image_tests_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ImageTestsConfiguration = expandImageBuilderImageTestConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("infrastructure_configuration_arn"); ok {
		input.InfrastructureConfigurationArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tags"); ok && len(v.(map[string]interface{})) > 0 {
		input.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().ImagebuilderTags()
	}

	output, err := conn.CreateImage(input)

	if err != nil {
		return fmt.Errorf("error creating Image Builder Image: %w", err)
	}

	if output == nil {
		return fmt.Errorf("error creating Image Builder Image: empty response")
	}

	d.SetId(aws.StringValue(output.ImageBuildVersionArn))

	if _, err := waiter.ImageStatusAvailable(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for Image Builder Image (%s) to become available: %w", d.Id(), err)
	}

	return resourceAwsImageBuilderImageRead(d, meta)
}

func resourceAwsImageBuilderImageRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	input := &imagebuilder.GetImageInput{
		ImageBuildVersionArn: aws.String(d.Id()),
	}

	output, err := conn.GetImage(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Image Builder Image (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting Image Builder Image (%s): %w", d.Id(), err)
	}

	if output == nil || output.Image == nil {
		return fmt.Errorf("error getting Image Builder Image (%s): empty response", d.Id())
	}

	image := output.Image

	d.Set("arn", image.Arn)
	d.Set("date_created", image.DateCreated)

	if image.DistributionConfiguration != nil {
		d.Set("distribution_configuration_arn", image.DistributionConfiguration.Arn)
	}

	d.Set("enhanced_image_metadata_enabled", image.EnhancedImageMetadataEnabled)

	if image.ImageRecipe != nil {
		d.Set("image_recipe_arn", image.ImageRecipe.Arn)
	}

	if image.ImageTestsConfiguration != nil {
		d.Set("image_tests_configuration", []interface{}{flattenImageBuilderImageTestsConfiguration(image.ImageTestsConfiguration)})
	} else {
		d.Set("image_tests_configuration", nil)
	}

	if image.InfrastructureConfiguration != nil {
		d.Set("infrastructure_configuration_arn", image.InfrastructureConfiguration.Arn)
	}

	d.Set("name", image.Name)
	d.Set("platform", image.Platform)
	d.Set("os_version", image.OsVersion)

	if image.OutputResources != nil {
		d.Set("output_resources", []interface{}{flattenImageBuilderOutputResources(image.OutputResources)})
	} else {
		d.Set("output_resources", nil)
	}

	d.Set("tags", keyvaluetags.ImagebuilderKeyValueTags(image.Tags).IgnoreAws().IgnoreConfig(meta.(*AWSClient).IgnoreTagsConfig).Map())
	d.Set("version", image.Version)

	return nil
}

func resourceAwsImageBuilderImageUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.ImagebuilderUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags for Image Builder Image (%s): %w", d.Id(), err)
		}
	}

	return resourceAwsImageBuilderImageRead(d, meta)
}

func resourceAwsImageBuilderImageDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	input := &imagebuilder.DeleteImageInput{
		ImageBuildVersionArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteImage(input)

	if tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Image Builder Image (%s): %w", d.Id(), err)
	}

	return nil
}

func flattenImageBuilderOutputResources(apiObject *imagebuilder.OutputResources) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Amis; v != nil {
		tfMap["amis"] = flattenImageBuilderAmis(v)
	}

	return tfMap
}

func flattenImageBuilderAmi(apiObject *imagebuilder.Ami) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AccountId; v != nil {
		tfMap["account_id"] = aws.StringValue(v)
	}

	if v := apiObject.Description; v != nil {
		tfMap["description"] = aws.StringValue(v)
	}

	if v := apiObject.Image; v != nil {
		tfMap["image"] = aws.StringValue(v)
	}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.StringValue(v)
	}

	if v := apiObject.Region; v != nil {
		tfMap["region"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenImageBuilderAmis(apiObjects []*imagebuilder.Ami) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenImageBuilderAmi(apiObject))
	}

	return tfList
}
