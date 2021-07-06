package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/efs/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/efs/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsEfsBackupPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEfsBackupPolicyCreate,
		Read:   resourceAwsEfsBackupPolicyRead,
		Update: resourceAwsEfsBackupPolicyUpdate,
		Delete: resourceAwsEfsBackupPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"backup_policy": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"status": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								efs.StatusDisabled,
								efs.StatusEnabled,
							}, false),
						},
					},
				},
			},

			"file_system_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsEfsBackupPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).efsconn

	fsID := d.Get("file_system_id").(string)

	if err := efsBackupPolicyPut(conn, fsID, d.Get("backup_policy").([]interface{})[0].(map[string]interface{})); err != nil {
		return err
	}

	d.SetId(fsID)

	return resourceAwsEfsBackupPolicyRead(d, meta)
}

func resourceAwsEfsBackupPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).efsconn

	output, err := finder.BackupPolicyByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EFS Backup Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EFS Backup Policy (%s): %w", d.Id(), err)
	}

	if err := d.Set("backup_policy", []interface{}{flattenEfsBackupPolicy(output)}); err != nil {
		return fmt.Errorf("error setting backup_policy: %w", err)
	}

	d.Set("file_system_id", d.Id())

	return nil
}

func resourceAwsEfsBackupPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).efsconn

	if err := efsBackupPolicyPut(conn, d.Id(), d.Get("backup_policy").([]interface{})[0].(map[string]interface{})); err != nil {
		return err
	}

	return resourceAwsEfsBackupPolicyRead(d, meta)
}

func resourceAwsEfsBackupPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).efsconn

	err := efsBackupPolicyPut(conn, d.Id(), map[string]interface{}{
		"status": efs.StatusDisabled,
	})

	if tfawserr.ErrCodeEquals(err, efs.ErrCodeFileSystemNotFound) {
		return nil
	}

	if err != nil {
		return err
	}

	return nil
}

// efsBackupPolicyPut attempts to update the file system's backup policy.
// Any error is returned.
func efsBackupPolicyPut(conn *efs.EFS, fsID string, tfMap map[string]interface{}) error {
	input := &efs.PutBackupPolicyInput{
		BackupPolicy: expandEfsBackupPolicy(tfMap),
		FileSystemId: aws.String(fsID),
	}

	log.Printf("[DEBUG] Putting EFS Backup Policy: %s", input)
	_, err := conn.PutBackupPolicy(input)

	if err != nil {
		return fmt.Errorf("error putting EFS Backup Policy (%s): %w", fsID, err)
	}

	if aws.StringValue(input.BackupPolicy.Status) == efs.StatusEnabled {
		if _, err := waiter.BackupPolicyEnabled(conn, fsID); err != nil {
			return fmt.Errorf("error waiting for EFS Backup Policy (%s) to enable: %w", fsID, err)
		}
	} else {
		if _, err := waiter.BackupPolicyDisabled(conn, fsID); err != nil {
			return fmt.Errorf("error waiting for EFS Backup Policy (%s) to disable: %w", fsID, err)
		}
	}

	return nil
}

func expandEfsBackupPolicy(tfMap map[string]interface{}) *efs.BackupPolicy {
	if tfMap == nil {
		return nil
	}

	apiObject := &efs.BackupPolicy{}

	if v, ok := tfMap["status"].(string); ok && v != "" {
		apiObject.Status = aws.String(v)
	}

	return apiObject
}

func flattenEfsBackupPolicy(apiObject *efs.BackupPolicy) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Status; v != nil {
		tfMap["status"] = aws.StringValue(v)
	}

	return tfMap
}
