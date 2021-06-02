package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/schemas"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/schemas/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsSchemasDiscoverer() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSchemasDiscovererCreate,
		Read:   resourceAwsSchemasDiscovererRead,
		Update: resourceAwsSchemasDiscovererUpdate,
		Delete: resourceAwsSchemasDiscovererDelete,
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
				ValidateFunc: validation.StringLenBetween(0, 256),
			},

			"source_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},

			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsSchemasDiscovererCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).schemasconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	sourceARN := d.Get("source_arn").(string)
	input := &schemas.CreateDiscovererInput{
		SourceArn: aws.String(sourceARN),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().SchemasTags()
	}

	log.Printf("[DEBUG] Creating EventBridge Schemas Discoverer: %s", input)
	output, err := conn.CreateDiscoverer(input)

	if err != nil {
		return fmt.Errorf("error creating EventBridge Schemas Discoverer (%s): %w", sourceARN, err)
	}

	d.SetId(aws.StringValue(output.DiscovererId))

	return resourceAwsSchemasDiscovererRead(d, meta)
}

func resourceAwsSchemasDiscovererRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).schemasconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	output, err := finder.DiscovererByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EventBridge Schemas Discoverer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EventBridge Schemas Discoverer (%s): %w", d.Id(), err)
	}

	d.Set("arn", output.DiscovererArn)
	d.Set("description", output.Description)
	d.Set("source_arn", output.SourceArn)

	tags, err := keyvaluetags.SchemasListTags(conn, d.Get("arn").(string))

	if err != nil {
		return fmt.Errorf("error listing tags for EventBridge Schemas Discoverer (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsSchemasDiscovererUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).schemasconn

	if d.HasChange("description") {
		input := &schemas.UpdateDiscovererInput{
			DiscovererId: aws.String(d.Id()),
			Description:  aws.String(d.Get("description").(string)),
		}

		log.Printf("[DEBUG] Updating EventBridge Schemas Discoverer: %s", input)
		_, err := conn.UpdateDiscoverer(input)

		if err != nil {
			return fmt.Errorf("error updating EventBridge Schemas Discoverer (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := keyvaluetags.SchemasUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceAwsSchemasDiscovererRead(d, meta)
}

func resourceAwsSchemasDiscovererDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).schemasconn

	log.Printf("[INFO] Deleting EventBridge Schemas Discoverer (%s)", d.Id())
	_, err := conn.DeleteDiscoverer(&schemas.DeleteDiscovererInput{
		DiscovererId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, schemas.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EventBridge Schemas Discoverer (%s): %w", d.Id(), err)
	}

	return nil
}
