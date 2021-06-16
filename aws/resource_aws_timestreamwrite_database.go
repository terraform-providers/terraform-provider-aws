package aws

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/timestreamwrite"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsTimestreamWriteDatabase() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAwsTimestreamWriteDatabaseCreate,
		ReadWithoutTimeout:   resourceAwsTimestreamWriteDatabaseRead,
		UpdateWithoutTimeout: resourceAwsTimestreamWriteDatabaseUpdate,
		DeleteWithoutTimeout: resourceAwsTimestreamWriteDatabaseDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"database_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 64),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`), "must only include alphanumeric, underscore, period, or hyphen characters"),
				),
			},

			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				// The Timestream API accepts the KmsKeyId as an ID, ARN, alias, or alias ARN but always returns the ARN of the key.
				// The ARN is of the format 'arn:aws:kms:REGION:ACCOUNT_ID:key/KMS_KEY_ID'. Appropriate diff suppression
				// would require an extra API call to the kms service's DescribeKey method to decipher aliases.
				// To avoid importing an extra service in this resource, input here is restricted to only ARNs.
				ValidateFunc: validateArn,
			},

			"table_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"tags": tagsSchema(),

			"tags_all": tagsSchemaComputed(),
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsTimestreamWriteDatabaseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).timestreamwriteconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	dbName := d.Get("database_name").(string)

	input := &timestreamwrite.CreateDatabaseInput{
		DatabaseName: aws.String(dbName),
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().TimestreamwriteTags()
	}

	resp, err := conn.CreateDatabaseWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Timestream Database (%s): %w", dbName, err))
	}

	if resp == nil || resp.Database == nil {
		return diag.FromErr(fmt.Errorf("error creating Timestream Database (%s): empty output", dbName))
	}

	d.SetId(aws.StringValue(resp.Database.DatabaseName))

	return resourceAwsTimestreamWriteDatabaseRead(ctx, d, meta)
}

func resourceAwsTimestreamWriteDatabaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).timestreamwriteconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &timestreamwrite.DescribeDatabaseInput{
		DatabaseName: aws.String(d.Id()),
	}

	resp, err := conn.DescribeDatabaseWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, timestreamwrite.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Timestream Database %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading Timestream Database (%s): %w", d.Id(), err))
	}

	if resp == nil || resp.Database == nil {
		return diag.FromErr(fmt.Errorf("error reading Timestream Database (%s): empty output", d.Id()))
	}

	db := resp.Database
	arn := aws.StringValue(db.Arn)

	d.Set("arn", arn)
	d.Set("database_name", db.DatabaseName)
	d.Set("kms_key_id", db.KmsKeyId)
	d.Set("table_count", db.TableCount)

	tags, err := keyvaluetags.TimestreamwriteListTags(conn, arn)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error listing tags for Timestream Database (%s): %w", arn, err))
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

func resourceAwsTimestreamWriteDatabaseUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).timestreamwriteconn

	if d.HasChange("kms_key_id") {
		input := &timestreamwrite.UpdateDatabaseInput{
			DatabaseName: aws.String(d.Id()),
			KmsKeyId:     aws.String(d.Get("kms_key_id").(string)),
		}

		_, err := conn.UpdateDatabaseWithContext(ctx, input)

		if err != nil {
			return diag.FromErr(fmt.Errorf("error updating Timestream Database (%s): %w", d.Id(), err))
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.TimestreamwriteUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return diag.FromErr(fmt.Errorf("error updating Timestream Database (%s) tags: %w", d.Get("arn").(string), err))
		}
	}

	return resourceAwsTimestreamWriteDatabaseRead(ctx, d, meta)
}

func resourceAwsTimestreamWriteDatabaseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).timestreamwriteconn

	input := &timestreamwrite.DeleteDatabaseInput{
		DatabaseName: aws.String(d.Id()),
	}

	_, err := conn.DeleteDatabaseWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, timestreamwrite.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting Timestream Database (%s): %w", d.Id(), err))
	}

	return nil
}
