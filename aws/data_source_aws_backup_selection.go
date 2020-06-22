package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsBackupSelection() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsBackupSelectionRead,

		Schema: map[string]*schema.Schema{
			"plan_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"selection_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"iam_role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resources": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAwsBackupSelectionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	input := &backup.GetBackupSelectionInput{
		BackupPlanId: aws.String(d.Get("plan_id").(string)),
		SelectionId:  aws.String(d.Get("selection_id").(string)),
	}

	resp, err := conn.GetBackupSelection(input)
	if err != nil {
		return fmt.Errorf("Error getting Backup Selection: %s", err)
	}

	d.SetId(aws.StringValue(resp.SelectionId))
	d.Set("iam_role_arn", resp.BackupSelection.IamRoleArn)
	d.Set("name", resp.BackupSelection.SelectionName)

	if resp.BackupSelection.Resources != nil {
		if err := d.Set("resources", aws.StringValueSlice(resp.BackupSelection.Resources)); err != nil {
			return fmt.Errorf("error setting resources: %s", err)
		}
	}

	return nil
}
