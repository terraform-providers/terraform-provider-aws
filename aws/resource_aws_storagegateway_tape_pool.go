package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsStorageGatewayTapePool() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsStorageGatewayTapePoolCreate,
		Read:   resourceAwsStorageGatewayTapePoolRead,
		Update: resourceAwsStorageGatewayTapePoolUpdate,
		Delete: resourceAwsStorageGatewayTapePoolDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"pool_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"storage_class": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(storagegateway.TapeStorageClass_Values(), false),
			},
			"retention_lock_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      storagegateway.RetentionLockTypeNone,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(storagegateway.RetentionLockType_Values(), false),
			},
			"retention_lock_time_in_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Default:      0,
				ValidateFunc: validation.IntBetween(0, 36500),
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsStorageGatewayTapePoolCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).storagegatewayconn

	input := &storagegateway.CreateTapePoolInput{
		PoolName:                aws.String(d.Get("pool_name").(string)),
		StorageClass:            aws.String(d.Get("storage_class").(string)),
		RetentionLockType:       aws.String(d.Get("retention_lock_type").(string)),
		RetentionLockTimeInDays: aws.Int64(int64(d.Get("retention_lock_time_in_days").(int))),
		Tags:                    keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().StoragegatewayTags(),
	}

	log.Printf("[DEBUG] Creating Storage Gateway Tape Pool: %s", input)
	output, err := conn.CreateTapePool(input)
	if err != nil {
		return fmt.Errorf("error creating Storage Gateway Tape Pool: %w", err)
	}

	d.SetId(aws.StringValue(output.PoolARN))

	return resourceAwsStorageGatewayTapePoolRead(d, meta)
}

func resourceAwsStorageGatewayTapePoolUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).storagegatewayconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.StoragegatewayUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceAwsStorageGatewayTapePoolRead(d, meta)
}

func resourceAwsStorageGatewayTapePoolRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).storagegatewayconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &storagegateway.ListTapePoolsInput{
		PoolARNs: []*string{aws.String(d.Id())},
	}

	log.Printf("[DEBUG] Reading Storage Gateway Tape Pool: %s", input)
	output, err := conn.ListTapePools(input)

	if err != nil {
		return fmt.Errorf("[ERROR] Listing Storage Gateway Tape Pools %w", err)
	}

	if output == nil || len(output.PoolInfos) == 0 || output.PoolInfos[0] == nil || aws.StringValue(output.PoolInfos[0].PoolARN) != d.Id() {
		log.Printf("[WARN] Storage Gateway Tape Pool %q not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	pool := output.PoolInfos[0]

	poolArn := aws.StringValue(pool.PoolARN)
	d.Set("arn", poolArn)
	d.Set("pool_name", aws.StringValue(pool.PoolName))
	d.Set("retention_lock_time_in_days", aws.Int64Value(pool.RetentionLockTimeInDays))
	d.Set("retention_lock_type", aws.StringValue(pool.RetentionLockType))
	d.Set("storage_class", aws.StringValue(pool.StorageClass))

	tags, err := keyvaluetags.StoragegatewayListTags(conn, poolArn)
	if err != nil {
		return fmt.Errorf("error listing tags for resource (%s): %w", poolArn, err)
	}
	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceAwsStorageGatewayTapePoolDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).storagegatewayconn

	input := &storagegateway.DeleteTapePoolInput{
		PoolARN: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Storage Gateway Tape Pool: %s", input)
	_, err := conn.DeleteTapePool(input)
	if err != nil {
		return fmt.Errorf("error deleting Storage Gateway Tape Pool %q: %w", d.Id(), err)
	}

	return nil
}
