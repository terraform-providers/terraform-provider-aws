package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsBackupVaultPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsBackupVaultPolicyPut,
		Update: resourceAwsBackupVaultPolicyPut,
		Read:   resourceAwsBackupVaultPolicyRead,
		Delete: resourceAwsBackupVaultPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"backup_vault_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			},
			"backup_vault_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsBackupVaultPolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	input := &backup.PutBackupVaultAccessPolicyInput{
		BackupVaultName: aws.String(d.Get("backup_vault_name").(string)),
		Policy:          aws.String(d.Get("policy").(string)),
	}

	_, err := conn.PutBackupVaultAccessPolicy(input)
	if err != nil {
		return fmt.Errorf("error creating Backup Vault Policy (%s): %w", d.Id(), err)
	}

	d.SetId(d.Get("backup_vault_name").(string))

	return resourceAwsBackupVaultPolicyRead(d, meta)
}

func resourceAwsBackupVaultPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	input := &backup.GetBackupVaultAccessPolicyInput{
		BackupVaultName: aws.String(d.Id()),
	}

	resp, err := conn.GetBackupVaultAccessPolicy(input)
	if isAWSErr(err, backup.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Backup Vault Policy %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Backup Vault Policy (%s): %w", d.Id(), err)
	}
	d.Set("backup_vault_name", resp.BackupVaultName)
	d.Set("policy", resp.Policy)
	d.Set("backup_vault_arn", resp.BackupVaultArn)

	return nil
}

func resourceAwsBackupVaultPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	input := &backup.DeleteBackupVaultAccessPolicyInput{
		BackupVaultName: aws.String(d.Id()),
	}

	_, err := conn.DeleteBackupVaultAccessPolicy(input)
	if err != nil {
		if isAWSErr(err, backup.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error deleting Backup Vault Policy (%s): %w", d.Id(), err)
	}

	return nil
}
