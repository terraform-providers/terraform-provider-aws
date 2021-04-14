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

func resourceAwsGlueSchema() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGlueSchemaCreate,
		Read:   resourceAwsGlueSchemaRead,
		Update: resourceAwsGlueSchemaUpdate,
		Delete: resourceAwsGlueSchemaDelete,
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
			"registry_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateArn,
			},
			"registry_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"latest_schema_version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"next_schema_version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"schema_checkpoint": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"compatibility": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(glue.Compatibility_Values(), false),
			},
			"data_format": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(glue.DataFormat_Values(), false),
			},
			"schema_definition": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 170000),
					validation.StringMatch(regexp.MustCompile(`.*\S.*`), ""),
				),
			},
			"schema_name": {
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

func resourceAwsGlueSchemaCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	input := &glue.CreateSchemaInput{
		SchemaName:       aws.String(d.Get("schema_name").(string)),
		SchemaDefinition: aws.String(d.Get("schema_definition").(string)),
		DataFormat:       aws.String(d.Get("data_format").(string)),
		Tags:             keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().GlueTags(),
	}

	if v, ok := d.GetOk("registry_arn"); ok {
		input.RegistryId = tfglue.CreateAwsGlueRegistryID(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("compatibility"); ok {
		input.Compatibility = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Glue Schema: %s", input)
	output, err := conn.CreateSchema(input)
	if err != nil {
		return fmt.Errorf("error creating Glue Schema: %w", err)
	}
	d.SetId(aws.StringValue(output.SchemaArn))

	_, err = waiter.SchemaAvailable(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error waiting for Glue Schema (%s) to be Available: %w", d.Id(), err)
	}

	return resourceAwsGlueSchemaRead(d, meta)
}

func resourceAwsGlueSchemaRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	output, err := finder.SchemaByID(conn, d.Id())
	if err != nil {
		if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
			log.Printf("[WARN] Glue Schema (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading Glue Schema (%s): %w", d.Id(), err)
	}

	if output == nil {
		log.Printf("[WARN] Glue Schema (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	arn := aws.StringValue(output.SchemaArn)
	d.Set("arn", arn)
	d.Set("description", output.Description)
	d.Set("schema_name", output.SchemaName)
	d.Set("compatibility", output.Compatibility)
	d.Set("data_format", output.DataFormat)
	d.Set("latest_schema_version", output.LatestSchemaVersion)
	d.Set("next_schema_version", output.NextSchemaVersion)
	d.Set("registry_arn", output.RegistryArn)
	d.Set("registry_name", output.RegistryName)
	d.Set("schema_checkpoint", output.SchemaCheckpoint)

	tags, err := keyvaluetags.GlueListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Glue Schema (%s): %w", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	schemeDefOutput, err := finder.SchemaVersionByID(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error reading Glue Schema Definition (%s): %w", d.Id(), err)
	}

	d.Set("schema_definition", schemeDefOutput.SchemaDefinition)

	return nil
}

func resourceAwsGlueSchemaUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	input := &glue.UpdateSchemaInput{
		SchemaId: tfglue.CreateAwsGlueSchemaID(d.Id()),
		SchemaVersionNumber: &glue.SchemaVersionNumber{
			LatestVersion: aws.Bool(true),
		},
	}
	update := false

	if d.HasChange("description") {
		input.Description = aws.String(d.Get("description").(string))
		update = true
	}

	if d.HasChange("compatibility") {
		input.Compatibility = aws.String(d.Get("compatibility").(string))
		update = true
	}

	if update {
		log.Printf("[DEBUG] Updating Glue Schema: %#v", input)
		_, err := conn.UpdateSchema(input)
		if err != nil {
			return fmt.Errorf("error updating Glue Schema (%s): %w", d.Id(), err)
		}

		_, err = waiter.SchemaAvailable(conn, d.Id())
		if err != nil {
			return fmt.Errorf("error waiting for Glue Schema (%s) to be Available: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.GlueUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	if d.HasChange("schema_definition") {
		defInput := &glue.RegisterSchemaVersionInput{
			SchemaId:         tfglue.CreateAwsGlueSchemaID(d.Id()),
			SchemaDefinition: aws.String(d.Get("schema_definition").(string)),
		}

		_, err := conn.RegisterSchemaVersion(defInput)
		if err != nil {
			return fmt.Errorf("error updating Glue Schema Definition (%s): %w", d.Id(), err)
		}

		_, err = waiter.SchemaVersionAvailable(conn, d.Id())
		if err != nil {
			return fmt.Errorf("error waiting for Glue Schema Version (%s) to be Available: %w", d.Id(), err)
		}
	}

	return resourceAwsGlueSchemaRead(d, meta)
}

func resourceAwsGlueSchemaDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	log.Printf("[DEBUG] Deleting Glue Schema: %s", d.Id())
	input := &glue.DeleteSchemaInput{
		SchemaId: tfglue.CreateAwsGlueSchemaID(d.Id()),
	}

	_, err := conn.DeleteSchema(input)
	if err != nil {
		if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error deleting Glue Schema (%s): %w", d.Id(), err)
	}

	_, err = waiter.SchemaDeleted(conn, d.Id())
	if err != nil {
		if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error waiting for Glue Schema (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}
