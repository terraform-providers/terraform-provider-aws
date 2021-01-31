package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker/waiter"
)

func resourceAwsSagemakerImage() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSagemakerImageCreate,
		Read:   resourceAwsSagemakerImageRead,
		Update: resourceAwsSagemakerImageUpdate,
		Delete: resourceAwsSagemakerImageDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"image_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9])*$`), "Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"display_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 512),
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsSagemakerImageCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	name := d.Get("image_name").(string)
	input := &sagemaker.CreateImageInput{
		ImageName: aws.String(name),
		RoleArn:   aws.String(d.Get("role_arn").(string)),
	}

	if v, ok := d.GetOk("display_name"); ok {
		input.DisplayName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tags"); ok {
		input.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().SagemakerTags()
	}

	// for some reason even if the operation is retried the same error response is given even though the role is valid. a short sleep before creation solves it.
	time.Sleep(1 * time.Minute)
	_, err := conn.CreateImage(input)
	if err != nil {
		return fmt.Errorf("error creating SageMaker Image %s: %w", name, err)
	}

	d.SetId(name)

	if _, err := waiter.ImageCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for SageMaker Image (%s) to be created: %w", d.Id(), err)
	}

	return resourceAwsSagemakerImageRead(d, meta)
}

func resourceAwsSagemakerImageRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	image, err := finder.ImageByName(conn, d.Id())
	if err != nil {
		if isAWSErr(err, sagemaker.ErrCodeResourceNotFound, "does not exist") {
			d.SetId("")
			log.Printf("[WARN] Unable to find SageMaker Image (%s); removing from state", d.Id())
			return nil
		}
		return fmt.Errorf("error reading SageMaker Image (%s): %w", d.Id(), err)

	}

	arn := aws.StringValue(image.ImageArn)
	d.Set("image_name", image.ImageName)
	d.Set("arn", arn)
	d.Set("role_arn", image.RoleArn)
	d.Set("display_name", image.DisplayName)
	d.Set("description", image.Description)

	tags, err := keyvaluetags.SagemakerListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for SageMaker Image (%s): %w", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceAwsSagemakerImageUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn
	needsUpdate := false

	input := &sagemaker.UpdateImageInput{
		ImageName: aws.String(d.Id()),
	}

	var deleteProperties []*string

	if d.HasChange("description") {
		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		} else {
			deleteProperties = append(deleteProperties, aws.String("Description"))
			input.DeleteProperties = deleteProperties
		}
		needsUpdate = true
	}

	if d.HasChange("display_name") {
		if v, ok := d.GetOk("display_name"); ok {
			input.DisplayName = aws.String(v.(string))
		} else {
			deleteProperties = append(deleteProperties, aws.String("DisplayName"))
			input.DeleteProperties = deleteProperties
		}
		needsUpdate = true
	}

	if needsUpdate {
		log.Printf("[DEBUG] sagemaker Image update config: %#v", *input)
		_, err := conn.UpdateImage(input)
		if err != nil {
			return fmt.Errorf("error updating SageMaker Image: %w", err)
		}

		if _, err := waiter.ImageCreated(conn, d.Id()); err != nil {
			return fmt.Errorf("error waiting for SageMaker Image (%s) to update: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.SagemakerUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating SageMaker Image (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsSagemakerImageRead(d, meta)
}

func resourceAwsSagemakerImageDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	input := &sagemaker.DeleteImageInput{
		ImageName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteImage(input); err != nil {
		if isAWSErr(err, sagemaker.ErrCodeResourceNotFound, "No Image with the name") {
			return nil
		}
		return fmt.Errorf("error deleting SageMaker Image (%s): %w", d.Id(), err)
	}

	if _, err := waiter.ImageDeleted(conn, d.Id()); err != nil {
		if isAWSErr(err, sagemaker.ErrCodeResourceNotFound, "does not exist") {
			return nil
		}
		return fmt.Errorf("error waiting for SageMaker Image (%s) to delete: %w", d.Id(), err)
	}

	return nil
}
