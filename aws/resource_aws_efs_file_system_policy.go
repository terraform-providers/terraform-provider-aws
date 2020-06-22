package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsEfsFileSystemPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEfsFileSystemPolicyPut,
		Read:   resourceAwsEfsFileSystemPolicyRead,
		Update: resourceAwsEfsFileSystemPolicyPut,
		Delete: resourceAwsEfsFileSystemPolicyDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"file_system_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentJsonDiffs,
			},
		},
	}
}

func resourceAwsEfsFileSystemPolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).efsconn

	fsId := d.Get("file_system_id").(string)
	input := &efs.PutFileSystemPolicyInput{
		FileSystemId: aws.String(fsId),
		Policy:       aws.String(d.Get("policy").(string)),
	}
	log.Printf("[DEBUG] Adding EFS File System Policy: %#v", input)
	_, err := conn.PutFileSystemPolicy(input)
	if err != nil {
		return fmt.Errorf("error creating EFS File System Policy %q: %s", d.Id(), err.Error())
	}

	d.SetId(fsId)

	return resourceAwsEfsFileSystemPolicyRead(d, meta)
}

func resourceAwsEfsFileSystemPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).efsconn

	var policyRes *efs.DescribeFileSystemPolicyOutput
	policyRes, err := conn.DescribeFileSystemPolicy(&efs.DescribeFileSystemPolicyInput{
		FileSystemId: aws.String(d.Id()),
	})
	if err != nil {
		if isAWSErr(err, efs.ErrCodeFileSystemNotFound, "") {
			log.Printf("[WARN] EFS File System (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		if isAWSErr(err, efs.ErrCodePolicyNotFound, "") {
			log.Printf("[WARN] EFS File System Policy (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error describing policy for EFS file system (%s): %s", d.Id(), err)
	}

	d.Set("file_system_id", policyRes.FileSystemId)
	d.Set("policy", policyRes.Policy)

	return nil
}

func resourceAwsEfsFileSystemPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).efsconn

	log.Printf("[DEBUG] Deleting EFS File System Policy: %s", d.Id())
	_, err := conn.DeleteFileSystemPolicy(&efs.DeleteFileSystemPolicyInput{
		FileSystemId: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("error deleting EFS File System Policy: %s with err %s", d.Id(), err.Error())
	}

	log.Printf("[DEBUG] EFS File System Policy %q deleted.", d.Id())

	return nil
}
