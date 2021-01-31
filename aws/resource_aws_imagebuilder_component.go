package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsImageBuilderComponent() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsImageBuilderComponentCreate,
		Read:   resourceAwsImageBuilderComponentRead,
		Update: resourceAwsImageBuilderComponentUpdate,
		Delete: resourceAwsImageBuilderComponentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"change_description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"data": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"data", "uri"},
				ValidateFunc: validation.StringLenBetween(1, 16000),
			},
			"date_created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"encrypted": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"kms_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
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
			"platform": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(imagebuilder.Platform_Values(), false),
			},
			"supported_os_versions": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringIsNotEmpty,
				},
				MinItems: 1,
				MaxItems: 25,
			},
			"tags": tagsSchema(),
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"uri": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"data", "uri"},
			},
			"version": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
		},
	}
}

func resourceAwsImageBuilderComponentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	input := &imagebuilder.CreateComponentInput{
		ClientToken: aws.String(resource.UniqueId()),
	}

	if v, ok := d.GetOk("change_description"); ok {
		input.ChangeDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("data"); ok {
		input.Data = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("name"); ok {
		input.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("platform"); ok {
		input.Platform = aws.String(v.(string))
	}

	if v, ok := d.GetOk("supported_os_versions"); ok && v.(*schema.Set).Len() > 0 {
		input.SupportedOsVersions = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("tags"); ok && len(v.(map[string]interface{})) > 0 {
		input.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().ImagebuilderTags()
	}

	if v, ok := d.GetOk("uri"); ok {
		input.Uri = aws.String(v.(string))
	}

	if v, ok := d.GetOk("version"); ok {
		input.SemanticVersion = aws.String(v.(string))
	}

	output, err := conn.CreateComponent(input)

	if err != nil {
		return fmt.Errorf("error creating Image Builder Component: %w", err)
	}

	if output == nil {
		return fmt.Errorf("error creating Image Builder Component: empty result")
	}

	d.SetId(aws.StringValue(output.ComponentBuildVersionArn))

	return resourceAwsImageBuilderComponentRead(d, meta)
}

func resourceAwsImageBuilderComponentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &imagebuilder.GetComponentInput{
		ComponentBuildVersionArn: aws.String(d.Id()),
	}

	output, err := conn.GetComponent(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Image Builder Component (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting Image Builder Component (%s): %w", d.Id(), err)
	}

	if output == nil || output.Component == nil {
		return fmt.Errorf("error getting Image Builder Component (%s): empty result", d.Id())
	}

	component := output.Component

	d.Set("arn", component.Arn)
	d.Set("change_description", component.ChangeDescription)
	d.Set("data", component.Data)
	d.Set("date_created", component.DateCreated)
	d.Set("description", component.Description)
	d.Set("encrypted", component.Encrypted)
	d.Set("kms_key_id", component.KmsKeyId)
	d.Set("name", component.Name)
	d.Set("owner", component.Owner)
	d.Set("platform", component.Platform)
	d.Set("supported_os_versions", aws.StringValueSlice(component.SupportedOsVersions))

	if err := d.Set("tags", keyvaluetags.ImagebuilderKeyValueTags(component.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	d.Set("type", component.Type)
	d.Set("version", component.Version)

	return nil
}

func resourceAwsImageBuilderComponentUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.ImagebuilderUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags for Image Builder Component (%s): %w", d.Id(), err)
		}
	}

	return resourceAwsImageBuilderComponentRead(d, meta)
}

func resourceAwsImageBuilderComponentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	input := &imagebuilder.DeleteComponentInput{
		ComponentBuildVersionArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteComponent(input)

	if tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Image Builder Component (%s): %w", d.Id(), err)
	}

	return nil
}
