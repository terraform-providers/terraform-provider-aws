package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/schemas"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	tfschemas "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/schemas"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/schemas/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsSchemasSchema() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSchemasSchemaCreate,
		Read:   resourceAwsSchemasSchemaRead,
		Update: resourceAwsSchemasSchemaUpdate,
		Delete: resourceAwsSchemasSchemaDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"content": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: suppressEquivalentJsonDiffs,
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

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 385),
					validation.StringMatch(regexp.MustCompile(`^[\.\-_A-Za-z@]+`), ""),
				),
			},

			"registry_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(schemas.Type_Values(), true),
			},

			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"version_created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsSchemasSchemaCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).schemasconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	registryName := d.Get("registry_name").(string)
	input := &schemas.CreateSchemaInput{
		Content:      aws.String(d.Get("content").(string)),
		RegistryName: aws.String(registryName),
		SchemaName:   aws.String(name),
		Type:         aws.String(d.Get("type").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().SchemasTags()
	}

	id := tfschemas.SchemaCreateResourceID(name, registryName)

	log.Printf("[DEBUG] Creating EventBridge Schemas Schema: %s", input)
	_, err := conn.CreateSchema(input)

	if err != nil {
		return fmt.Errorf("error creating EventBridge Schemas Schema (%s): %w", id, err)
	}

	d.SetId(id)

	return resourceAwsSchemasSchemaRead(d, meta)
}

func resourceAwsSchemasSchemaRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).schemasconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	name, registryName, err := tfschemas.SchemaParseResourceID(d.Id())

	if err != nil {
		return fmt.Errorf("error parsing EventBridge Schemas Schema ID: %w", err)
	}

	output, err := finder.SchemaByNameAndRegistryName(conn, name, registryName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EventBridge Schemas Schema (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EventBridge Schemas Schema (%s): %w", d.Id(), err)
	}

	d.Set("arn", output.SchemaArn)
	d.Set("content", output.Content)
	d.Set("description", output.Description)
	if output.LastModified != nil {
		d.Set("last_modified", aws.TimeValue(output.LastModified).Format(time.RFC3339))
	} else {
		d.Set("last_modified", nil)
	}
	d.Set("name", output.SchemaName)
	d.Set("registry_name", registryName)
	d.Set("type", output.Type)
	d.Set("version", output.SchemaVersion)
	if output.VersionCreatedDate != nil {
		d.Set("version_created_date", aws.TimeValue(output.VersionCreatedDate).Format(time.RFC3339))
	} else {
		d.Set("version_created_date", nil)
	}

	tags, err := keyvaluetags.SchemasListTags(conn, d.Get("arn").(string))

	if err != nil {
		return fmt.Errorf("error listing tags for EventBridge Schemas Schema (%s): %w", d.Id(), err)
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

func resourceAwsSchemasSchemaUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).schemasconn

	if d.HasChanges("content", "description", "type") {
		name, registryName, err := tfschemas.SchemaParseResourceID(d.Id())

		if err != nil {
			return fmt.Errorf("error parsing EventBridge Schemas Schema ID: %w", err)
		}

		input := &schemas.UpdateSchemaInput{
			RegistryName: aws.String(registryName),
			SchemaName:   aws.String(name),
		}

		if d.HasChanges("content", "type") {
			input.Content = aws.String(d.Get("content").(string))
			input.Type = aws.String(d.Get("type").(string))
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}

		log.Printf("[DEBUG] Updating EventBridge Schemas Schema: %s", input)
		_, err = conn.UpdateSchema(input)

		if err != nil {
			return fmt.Errorf("error updating EventBridge Schemas Schema (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := keyvaluetags.SchemasUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceAwsSchemasSchemaRead(d, meta)
}

func resourceAwsSchemasSchemaDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).schemasconn

	name, registryName, err := tfschemas.SchemaParseResourceID(d.Id())

	if err != nil {
		return fmt.Errorf("error parsing EventBridge Schemas Schema ID: %w", err)
	}

	log.Printf("[INFO] Deleting EventBridge Schemas Schema (%s)", d.Id())
	_, err = conn.DeleteSchema(&schemas.DeleteSchemaInput{
		RegistryName: aws.String(registryName),
		SchemaName:   aws.String(name),
	})

	if tfawserr.ErrCodeEquals(err, schemas.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EventBridge Schemas Schema (%s): %w", d.Id(), err)
	}

	return nil
}
