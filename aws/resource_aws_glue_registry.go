package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	tfglue "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/glue"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/glue/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/glue/waiter"
)

func resourceAwsGlueRegistry() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGlueRegistryCreate,
		Read:   resourceAwsGlueRegistryRead,
		Update: resourceAwsGlueRegistryUpdate,
		Delete: resourceAwsGlueRegistryDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 2048),
			},
			"registry_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexp.MustCompile(`[a-zA-Z0-9-_$#]+$`), ""),
				),
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsGlueRegistryCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	input := &glue.CreateRegistryInput{
		RegistryName: aws.String(d.Get("registry_name").(string)),
		Tags:         keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().GlueTags(),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Glue Registry: %s", input)
	output, err := conn.CreateRegistry(input)
	if err != nil {
		return fmt.Errorf("error creating Glue Registry: %w", err)
	}
	d.SetId(aws.StringValue(output.RegistryArn))

	return resourceAwsGlueRegistryRead(d, meta)
}

func resourceAwsGlueRegistryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	output, err := finder.RegistryByID(conn, d.Id())
	if err != nil {
		if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
			log.Printf("[WARN] Glue Registry (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading Glue Registry (%s): %w", d.Id(), err)
	}

	if output == nil {
		log.Printf("[WARN] Glue Registry (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	arn := aws.StringValue(output.RegistryArn)
	d.Set("arn", arn)
	d.Set("description", output.Description)
	d.Set("registry_name", output.RegistryName)

	tags, err := keyvaluetags.GlueListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Glue Registry (%s): %w", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceAwsGlueRegistryUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	if d.HasChanges("description") {
		input := &glue.UpdateRegistryInput{
			RegistryId: tfglue.CreateAwsGlueRegistryID(d.Id()),
		}

		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}

		log.Printf("[DEBUG] Updating Glue Registry: %#v", input)
		_, err := conn.UpdateRegistry(input)
		if err != nil {
			return fmt.Errorf("error updating Glue Registry (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.GlueUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsGlueRegistryRead(d, meta)
}

func resourceAwsGlueRegistryDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	log.Printf("[DEBUG] Deleting Glue Registry: %s", d.Id())
	input := &glue.DeleteRegistryInput{
		RegistryId: tfglue.CreateAwsGlueRegistryID(d.Id()),
	}

	_, err := conn.DeleteRegistry(input)
	if err != nil {
		if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error deleting Glue Registry (%s): %w", d.Id(), err)
	}

	_, err = waiter.RegistryDeleted(conn, d.Id())
	if err != nil {
		if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error waiting for Glue Registry (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}
