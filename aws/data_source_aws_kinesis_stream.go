package aws

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsKinesisStream() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsKinesisStreamRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"creation_timestamp": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"retention_period": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"open_shards": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"closed_shards": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"shard_level_metrics": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsKinesisStreamRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	sn := d.Get("name").(string)

	state, err := readKinesisStreamState(conn, sn)
	if err != nil {
		return err
	}
	d.SetId(state.arn)
	d.Set("arn", state.arn)
	d.Set("name", sn)
	d.Set("open_shards", state.openShards)
	d.Set("closed_shards", state.closedShards)
	d.Set("status", state.status)
	d.Set("creation_timestamp", state.creationTimestamp)
	d.Set("retention_period", state.retentionPeriod)
	d.Set("shard_level_metrics", state.shardLevelMetrics)

	tags, err := keyvaluetags.KinesisListTags(conn, sn)

	if err != nil {
		return fmt.Errorf("error listing tags for Kinesis Stream (%s): %s", sn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}
