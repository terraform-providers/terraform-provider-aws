package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsEbsEncryptionByDefault() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEbsEncryptionByDefaultCreate,
		Read:   resourceAwsEbsEncryptionByDefaultRead,
		Update: resourceAwsEbsEncryptionByDefaultUpdate,
		Delete: resourceAwsEbsEncryptionByDefaultDelete,

		Schema: map[string]*schema.Schema{
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceAwsEbsEncryptionByDefaultCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	enabled := d.Get("enabled").(bool)
	if err := setEbsEncryptionByDefault(conn, enabled); err != nil {
		return fmt.Errorf("error creating EBS encryption by default (%t): %s", enabled, err)
	}

	//lintignore:R015 // Allow legacy unstable ID usage in managed resource
	d.SetId(resource.UniqueId())

	return resourceAwsEbsEncryptionByDefaultRead(d, meta)
}

func resourceAwsEbsEncryptionByDefaultRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	resp, err := conn.GetEbsEncryptionByDefault(&ec2.GetEbsEncryptionByDefaultInput{})
	if err != nil {
		return fmt.Errorf("error reading EBS encryption by default: %s", err)
	}

	d.Set("enabled", aws.BoolValue(resp.EbsEncryptionByDefault))

	return nil
}

func resourceAwsEbsEncryptionByDefaultUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	enabled := d.Get("enabled").(bool)
	if err := setEbsEncryptionByDefault(conn, enabled); err != nil {
		return fmt.Errorf("error updating EBS encryption by default (%t): %s", enabled, err)
	}

	return resourceAwsEbsEncryptionByDefaultRead(d, meta)
}

func resourceAwsEbsEncryptionByDefaultDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	// Removing the resource disables default encryption.
	if err := setEbsEncryptionByDefault(conn, false); err != nil {
		return fmt.Errorf("error disabling EBS encryption by default: %s", err)
	}

	return nil
}

func setEbsEncryptionByDefault(conn *ec2.EC2, enabled bool) error {
	var err error

	if enabled {
		_, err = conn.EnableEbsEncryptionByDefault(&ec2.EnableEbsEncryptionByDefaultInput{})
	} else {
		_, err = conn.DisableEbsEncryptionByDefault(&ec2.DisableEbsEncryptionByDefaultInput{})
	}

	return err
}
