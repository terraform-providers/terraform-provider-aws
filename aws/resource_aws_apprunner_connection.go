package aws

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apprunner"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/apprunner/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/apprunner/waiter"
)

func resourceAwsAppRunnerConnection() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAwsAppRunnerConnectionCreate,
		ReadWithoutTimeout:   resourceAwsAppRunnerConnectionRead,
		UpdateWithoutTimeout: resourceAwsAppRunnerConnectionUpdate,
		DeleteWithoutTimeout: resourceAwsAppRunnerConnectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"connection_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"provider_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(apprunner.ProviderType_Values(), false),
				ForceNew:     true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchema(),

			"tags_all": tagsSchemaComputed(),
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsAppRunnerConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).apprunnerconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("connection_name").(string)

	input := &apprunner.CreateConnectionInput{
		ConnectionName: aws.String(name),
		ProviderType:   aws.String(d.Get("provider_type").(string)),
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().ApprunnerTags()
	}

	output, err := conn.CreateConnectionWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating App Runner Connection (%s): %w", name, err))
	}

	if output == nil || output.Connection == nil {
		return diag.FromErr(fmt.Errorf("error creating App Runner Connection (%s): empty output", name))
	}

	d.SetId(aws.StringValue(output.Connection.ConnectionName))

	return resourceAwsAppRunnerConnectionRead(ctx, d, meta)
}

func resourceAwsAppRunnerConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).apprunnerconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	c, err := finder.ConnectionSummaryByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] App Runner Connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading App Runner Connection (%s): %w", d.Id(), err))
	}

	if c == nil {
		if d.IsNewResource() {
			return diag.FromErr(fmt.Errorf("error reading App Runner Connection (%s): empty output after creation", d.Id()))
		}
		log.Printf("[WARN] App Runner Connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	arn := aws.StringValue(c.ConnectionArn)

	d.Set("arn", arn)
	d.Set("connection_name", c.ConnectionName)
	d.Set("provider_type", c.ProviderType)
	d.Set("status", c.Status)

	tags, err := keyvaluetags.ApprunnerListTags(conn, arn)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error listing tags for App Runner Connection (%s): %w", arn, err))
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags_all: %w", err))
	}

	return nil
}

func resourceAwsAppRunnerConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).apprunnerconn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.ApprunnerUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return diag.FromErr(fmt.Errorf("error updating App Runner Connection (%s) tags: %w", d.Get("arn").(string), err))
		}
	}

	return resourceAwsAppRunnerConnectionRead(ctx, d, meta)
}

func resourceAwsAppRunnerConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).apprunnerconn

	input := &apprunner.DeleteConnectionInput{
		ConnectionArn: aws.String(d.Get("arn").(string)),
	}

	_, err := conn.DeleteConnectionWithContext(ctx, input)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting App Runner Connection (%s): %w", d.Id(), err))
	}

	if err := waiter.ConnectionDeleted(ctx, conn, d.Id()); err != nil {
		if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error waiting for App Runner Connection (%s) deletion: %w", d.Id(), err))
	}

	return nil
}
