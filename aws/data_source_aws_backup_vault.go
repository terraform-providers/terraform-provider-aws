package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsBackupVault() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsBackupVaultRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kms_key_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"recovery_points": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsBackupVaultRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	name := d.Get("name").(string)
	input := &backup.DescribeBackupVaultInput{
		BackupVaultName: aws.String(name),
	}

	resp, err := conn.DescribeBackupVault(input)
	if err != nil {
		return fmt.Errorf("Error getting Backup Vault: %v", err)
	}

	d.SetId(aws.StringValue(resp.BackupVaultName))
	d.Set("arn", resp.BackupVaultArn)
	d.Set("kms_key_arn", resp.EncryptionKeyArn)
	d.Set("name", resp.BackupVaultName)
	d.Set("recovery_points", resp.NumberOfRecoveryPoints)

	tags, err := keyvaluetags.BackupListTags(conn, aws.StringValue(resp.BackupVaultArn))
	if err != nil {
		return fmt.Errorf("error listing tags for Backup Vault (%s): %s", name, err)
	}
	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}
