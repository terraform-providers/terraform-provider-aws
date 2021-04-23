package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker/waiter"
)

func resourceAwsSagemakerModelPackageGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSagemakerModelPackageGroupCreate,
		Read:   resourceAwsSagemakerModelPackageGroupRead,
		Update: resourceAwsSagemakerModelPackageGroupUpdate,
		Delete: resourceAwsSagemakerModelPackageGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"model_package_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9]){0,62}$`),
						"Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
			"model_package_group_description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsSagemakerModelPackageGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	name := d.Get("model_package_group_name").(string)
	input := &sagemaker.CreateModelPackageGroupInput{
		ModelPackageGroupName: aws.String(name),
	}

	if v, ok := d.GetOk("model_package_group_description"); ok {
		input.ModelPackageGroupDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tags"); ok {
		input.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().SagemakerTags()
	}

	_, err := conn.CreateModelPackageGroup(input)
	if err != nil {
		return fmt.Errorf("error creating SageMaker Model Package Group %s: %w", name, err)
	}

	d.SetId(name)

	if _, err := waiter.ModelPackageGroupCompleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Sagemaker Model Package Group (%s) to be created: %w", d.Id(), err)
	}

	return resourceAwsSagemakerModelPackageGroupRead(d, meta)
}

func resourceAwsSagemakerModelPackageGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	mpg, err := finder.ModelPackageGroupByName(conn, d.Id())
	if err != nil {
		if isAWSErr(err, "ValidationException", "does not exist") {
			d.SetId("")
			log.Printf("[WARN] Unable to find Sagemaker Model Package Group (%s); removing from state", d.Id())
			return nil
		}
		return fmt.Errorf("error reading SageMaker Model Package Group (%s): %w", d.Id(), err)

	}

	arn := aws.StringValue(mpg.ModelPackageGroupArn)
	d.Set("model_package_group_name", mpg.ModelPackageGroupName)
	d.Set("arn", arn)
	d.Set("model_package_group_description", mpg.ModelPackageGroupDescription)

	tags, err := keyvaluetags.SagemakerListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for SageMaker Model Package Group (%s): %w", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceAwsSagemakerModelPackageGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.SagemakerUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating SageMaker Model Package Group (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsSagemakerModelPackageGroupRead(d, meta)
}

func resourceAwsSagemakerModelPackageGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	input := &sagemaker.DeleteModelPackageGroupInput{
		ModelPackageGroupName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteModelPackageGroup(input); err != nil {
		if isAWSErr(err, "ValidationException", "does not exist") {
			return nil
		}
		return fmt.Errorf("error deleting SageMaker Model Package Group (%s): %w", d.Id(), err)
	}

	if _, err := waiter.ModelPackageGroupDeleted(conn, d.Id()); err != nil {
		if isAWSErr(err, "ValidationException", "does not exist") {
			return nil
		}
		return fmt.Errorf("error waiting for SageMaker Model Package Group (%s) to delete: %w", d.Id(), err)
	}

	return nil
}
