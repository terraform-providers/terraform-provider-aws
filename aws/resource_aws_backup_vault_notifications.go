package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsBackupVaultNotifications() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsBackupVaultNotificationsCreate,
		Read:   resourceAwsBackupVaultNotificationsRead,
		Delete: resourceAwsBackupVaultNotificationsDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"backup_vault_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9\-\_\.]{1,50}$`), "must consist of lowercase letters, numbers, and hyphens."),
			},
			"sns_topic_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"backup_vault_events": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(backup.VaultEvent_Values(), false),
				},
			},
			"backup_vault_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsBackupVaultNotificationsCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	input := &backup.PutBackupVaultNotificationsInput{
		BackupVaultName:   aws.String(d.Get("backup_vault_name").(string)),
		SNSTopicArn:       aws.String(d.Get("sns_topic_arn").(string)),
		BackupVaultEvents: expandStringSet(d.Get("backup_vault_events").(*schema.Set)),
	}

	_, err := conn.PutBackupVaultNotifications(input)
	if err != nil {
		return fmt.Errorf("error creating Backup Vault Notifications (%s): %w", d.Id(), err)
	}

	d.SetId(d.Get("backup_vault_name").(string))

	return resourceAwsBackupVaultNotificationsRead(d, meta)
}

func resourceAwsBackupVaultNotificationsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	input := &backup.GetBackupVaultNotificationsInput{
		BackupVaultName: aws.String(d.Id()),
	}

	resp, err := conn.GetBackupVaultNotifications(input)
	if isAWSErr(err, backup.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Backup Vault Notifcations %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Backup Vault Notifications (%s): %w", d.Id(), err)
	}
	d.Set("backup_vault_name", resp.BackupVaultName)
	d.Set("sns_topic_arn", resp.SNSTopicArn)
	d.Set("backup_vault_arn", resp.BackupVaultArn)
	if err := d.Set("backup_vault_events", flattenStringSet(resp.BackupVaultEvents)); err != nil {
		return fmt.Errorf("error setting backup_vault_events: %w", err)
	}

	return nil
}

func resourceAwsBackupVaultNotificationsDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	input := &backup.DeleteBackupVaultNotificationsInput{
		BackupVaultName: aws.String(d.Id()),
	}

	_, err := conn.DeleteBackupVaultNotifications(input)
	if err != nil {
		if isAWSErr(err, backup.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error deleting Backup Vault Notifications (%s): %w", d.Id(), err)
	}

	return nil
}
