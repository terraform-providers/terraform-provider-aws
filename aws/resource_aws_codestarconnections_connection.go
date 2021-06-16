package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codestarconnections"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/codestarconnections/finder"
)

func resourceAwsCodeStarConnectionsConnection() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCodeStarConnectionsConnectionCreate,
		Read:   resourceAwsCodeStarConnectionsConnectionRead,
		Update: resourceAwsCodeStarConnectionsConnectionUpdate,
		Delete: resourceAwsCodeStarConnectionsConnectionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"connection_status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"host_arn": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"provider_type"},
				ValidateFunc:  validateArn,
			},

			"provider_type": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Computed:      true,
				ConflictsWith: []string{"host_arn"},
				ValidateFunc:  validation.StringInSlice(codestarconnections.ProviderType_Values(), false),
			},

			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsCodeStarConnectionsConnectionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codestarconnectionsconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	params := &codestarconnections.CreateConnectionInput{
		ConnectionName: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("provider_type"); ok {
		params.ProviderType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("host_arn"); ok {
		params.HostArn = aws.String(v.(string))
	}

	if len(tags) > 0 {
		params.Tags = tags.IgnoreAws().CodestarconnectionsTags()
	}

	resp, err := conn.CreateConnection(params)
	if err != nil {
		return fmt.Errorf("error creating CodeStar connection: %w", err)
	}

	d.SetId(aws.StringValue(resp.ConnectionArn))

	return resourceAwsCodeStarConnectionsConnectionRead(d, meta)
}

func resourceAwsCodeStarConnectionsConnectionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codestarconnectionsconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	connection, err := finder.ConnectionByArn(conn, d.Id())
	if tfawserr.ErrCodeEquals(err, codestarconnections.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] CodeStar connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading CodeStar connection: %w", err)
	}

	if connection == nil {
		return fmt.Errorf("error reading CodeStar connection (%s): empty response", d.Id())
	}

	arn := aws.StringValue(connection.ConnectionArn)
	d.SetId(arn)
	d.Set("arn", connection.ConnectionArn)
	d.Set("connection_status", connection.ConnectionStatus)
	d.Set("name", connection.ConnectionName)
	d.Set("host_arn", connection.HostArn)
	d.Set("provider_type", connection.ProviderType)

	tags, err := keyvaluetags.CodestarconnectionsListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for CodeStar Connection (%s): %w", arn, err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsCodeStarConnectionsConnectionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codestarconnectionsconn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.CodestarconnectionsUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error Codestar Connection (%s) tags: %w", d.Id(), err)
		}
	}

	return nil
}

func resourceAwsCodeStarConnectionsConnectionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codestarconnectionsconn

	_, err := conn.DeleteConnection(&codestarconnections.DeleteConnectionInput{
		ConnectionArn: aws.String(d.Id()),
	})
	if tfawserr.ErrCodeEquals(err, codestarconnections.ErrCodeResourceNotFoundException) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting CodeStar connection: %w", err)
	}

	return nil
}
