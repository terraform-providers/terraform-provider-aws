package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsImageBuilderImageRecipe() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsImageBuilderImageRecipeRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"block_device_mapping": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"device_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ebs": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"delete_on_termination": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"encrypted": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"iops": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"kms_key_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"snapshot_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"volume_size": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"volume_type": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"no_device": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"virtual_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"component": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"component_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"date_created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parent_image": {
				Type:     schema.TypeString,
				Computed: true,
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
			"working_directory": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsImageBuilderImageRecipeRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &imagebuilder.GetImageRecipeInput{}

	if v, ok := d.GetOk("arn"); ok {
		input.ImageRecipeArn = aws.String(v.(string))
	}

	output, err := conn.GetImageRecipe(input)

	if err != nil {
		return fmt.Errorf("error reading Image Builder Image Recipe (%s): %w", aws.StringValue(input.ImageRecipeArn), err)
	}

	if output == nil || output.ImageRecipe == nil {
		return fmt.Errorf("error reading Image Builder Image Recipe (%s): empty response", aws.StringValue(input.ImageRecipeArn))
	}

	imageRecipe := output.ImageRecipe

	d.SetId(aws.StringValue(imageRecipe.Arn))
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
