package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsAmiLaunchPermission() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAmiLaunchPermissionCreate,
		Read:   resourceAwsAmiLaunchPermissionRead,
		Delete: resourceAwsAmiLaunchPermissionDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected ACCOUNT-ID/IMAGE-ID", d.Id())
				}
				accountId := idParts[0]
				imageId := idParts[1]
				d.Set("account_id", accountId)
				d.Set("image_id", imageId)
				d.SetId(fmt.Sprintf("%s-%s", imageId, accountId))
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"image_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"account_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsAmiLaunchPermissionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	image_id := d.Get("image_id").(string)
	account_id := d.Get("account_id").(string)

	_, err := conn.ModifyImageAttribute(&ec2.ModifyImageAttributeInput{
		ImageId:   aws.String(image_id),
		Attribute: aws.String(ec2.ImageAttributeNameLaunchPermission),
		LaunchPermission: &ec2.LaunchPermissionModifications{
			Add: []*ec2.LaunchPermission{
				{UserId: aws.String(account_id)},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("error creating AMI launch permission: %w", err)
	}

	d.SetId(fmt.Sprintf("%s-%s", image_id, account_id))
	return nil
}

func resourceAwsAmiLaunchPermissionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	exists, err := hasLaunchPermission(conn, d.Get("image_id").(string), d.Get("account_id").(string))
	if err != nil {
		return fmt.Errorf("error reading AMI launch permission (%s): %w", d.Id(), err)
	}
	if !exists {
		log.Printf("[WARN] AMI launch permission (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	return nil
}

func resourceAwsAmiLaunchPermissionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	image_id := d.Get("image_id").(string)
	account_id := d.Get("account_id").(string)

	_, err := conn.ModifyImageAttribute(&ec2.ModifyImageAttributeInput{
		ImageId:   aws.String(image_id),
		Attribute: aws.String(ec2.ImageAttributeNameLaunchPermission),
		LaunchPermission: &ec2.LaunchPermissionModifications{
			Remove: []*ec2.LaunchPermission{
				{UserId: aws.String(account_id)},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("error deleting AMI launch permission (%s): %w", d.Id(), err)
	}

	return nil
}

func hasLaunchPermission(conn *ec2.EC2, image_id string, account_id string) (bool, error) {
	attrs, err := conn.DescribeImageAttribute(&ec2.DescribeImageAttributeInput{
		ImageId:   aws.String(image_id),
		Attribute: aws.String(ec2.ImageAttributeNameLaunchPermission),
	})
	if err != nil {
		// When an AMI disappears out from under a launch permission resource, we will
		// see either InvalidAMIID.NotFound or InvalidAMIID.Unavailable.
		if ec2err, ok := err.(awserr.Error); ok && strings.HasPrefix(ec2err.Code(), "InvalidAMIID") {
			log.Printf("[DEBUG] %s no longer exists, so we'll drop launch permission for %s from the state", image_id, account_id)
			return false, nil
		}
		return false, err
	}

	for _, lp := range attrs.LaunchPermissions {
		if aws.StringValue(lp.UserId) == account_id {
			return true, nil
		}
	}
	return false, nil
}
