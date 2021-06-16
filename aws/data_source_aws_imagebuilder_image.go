package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsImageBuilderImage() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsImageBuilderImageRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"build_version_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"date_created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"distribution_configuration_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enhanced_image_metadata_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"image_recipe_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"image_tests_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"image_tests_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"timeout_minutes": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"infrastructure_configuration_arn": {
				Type:     schema.TypeString,
				Computed: true,
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
			"tags": tagsSchemaComputed(),
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsImageBuilderImageRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	input := &imagebuilder.GetImageInput{}

	if v, ok := d.GetOk("arn"); ok {
		input.ImageBuildVersionArn = aws.String(v.(string))
	}

	output, err := conn.GetImage(input)

	if err != nil {
		return fmt.Errorf("error getting Image Builder Image: %w", err)
	}

	if output == nil || output.Image == nil {
		return fmt.Errorf("error getting Image Builder Image: empty response")
	}

	image := output.Image

	d.SetId(aws.StringValue(image.Arn))

	// To prevent Terraform errors, only reset arn if not configured.
	// The configured ARN may contain x.x.x wildcards while the API returns
	// the full build version #.#.#/# suffix.
	if _, ok := d.GetOk("arn"); !ok {
		d.Set("arn", image.Arn)
	}

	d.Set("build_version_arn", image.Arn)
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
