package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsBackupPlan() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsBackupPlanRead,

		Schema: map[string]*schema.Schema{
			"plan_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchemaComputed(),
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsBackupPlanRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	id := d.Get("plan_id").(string)

	resp, err := conn.GetBackupPlan(&backup.GetBackupPlanInput{
		BackupPlanId: aws.String(id),
	})
	if err != nil {
		return fmt.Errorf("Error getting Backup Plan: %v", err)
	}

	d.SetId(aws.StringValue(resp.BackupPlanId))
	d.Set("arn", resp.BackupPlanArn)
	d.Set("name", resp.BackupPlan.BackupPlanName)
	d.Set("version", resp.VersionId)

	tags, err := keyvaluetags.BackupListTags(conn, aws.StringValue(resp.BackupPlanArn))
	if err != nil {
		return fmt.Errorf("error listing tags for Backup Plan (%s): %s", id, err)
	}
	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}
