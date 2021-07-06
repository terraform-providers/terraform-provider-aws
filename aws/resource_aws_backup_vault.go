package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsBackupVault() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsBackupVaultCreate,
		Read:   resourceAwsBackupVaultRead,
		Update: resourceAwsBackupVaultUpdate,
		Delete: resourceAwsBackupVaultDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9\-\_\.]{1,50}$`), "must consist of lowercase letters, numbers, and hyphens."),
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
			"kms_key_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"recovery_points": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsBackupVaultCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := &backup.CreateBackupVaultInput{
		BackupVaultName: aws.String(d.Get("name").(string)),
		BackupVaultTags: tags.IgnoreAws().BackupTags(),
	}

	if v, ok := d.GetOk("kms_key_arn"); ok {
		input.EncryptionKeyArn = aws.String(v.(string))
	}

	_, err := conn.CreateBackupVault(input)
	if err != nil {
		return fmt.Errorf("error creating Backup Vault (%s): %s", d.Id(), err)
	}

	d.SetId(d.Get("name").(string))

	return resourceAwsBackupVaultRead(d, meta)
}

func resourceAwsBackupVaultRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &backup.DescribeBackupVaultInput{
		BackupVaultName: aws.String(d.Id()),
	}

	resp, err := conn.DescribeBackupVault(input)
	if isAWSErr(err, backup.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Backup Vault %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if isAWSErr(err, "AccessDeniedException", "") {
		log.Printf("[WARN] Backup Vault %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Backup Vault (%s): %s", d.Id(), err)
	}
	d.Set("name", resp.BackupVaultName)
	d.Set("kms_key_arn", resp.EncryptionKeyArn)
	d.Set("arn", resp.BackupVaultArn)
	d.Set("recovery_points", resp.NumberOfRecoveryPoints)

	tags, err := keyvaluetags.BackupListTags(conn, d.Get("arn").(string))
	if err != nil {
		return fmt.Errorf("error listing tags for Backup Vault (%s): %s", d.Id(), err)
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

func resourceAwsBackupVaultUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := keyvaluetags.BackupUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags for Backup Vault (%s): %s", d.Id(), err)
		}
	}

	return resourceAwsBackupVaultRead(d, meta)
}

func resourceAwsBackupVaultDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	log.Printf("[DEBUG] Deleting Backup Vault: %s", d.Id())
	_, err := conn.DeleteBackupVault(&backup.DeleteBackupVaultInput{
		BackupVaultName: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("error deleting Backup Vault (%s): %w", d.Id(), err)
	}

	return nil
}
