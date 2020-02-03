package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloud9"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsCloud9EnvironmentMembership() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloud9EnvironmentMembershipCreate,
		Read:   resourceAwsCloud9EnvironmentMembershipRead,
		Update: resourceAwsCloud9EnvironmentMembershipUpdate,
		Delete: resourceAwsCloud9EnvironmentMembershipDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "#")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected ENVIRONMENT-ID/USER-ARN", d.Id())
				}
				envId := idParts[0]
				userArn := idParts[1]
				d.Set("environment_id", envId)
				d.Set("user_arn", userArn)
				d.SetId(fmt.Sprintf("%s#%s", envId, userArn))
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"environment_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"permissions": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					cloud9.MemberPermissionsReadOnly,
					cloud9.MemberPermissionsReadWrite,
				}, false),
			},
			"user_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"user_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsCloud9EnvironmentMembershipCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloud9conn

	params := &cloud9.CreateEnvironmentMembershipInput{
		EnvironmentId: aws.String(d.Get("environment_id").(string)),
		Permissions:   aws.String(d.Get("permissions").(string)),
		UserArn:       aws.String(d.Get("user_arn").(string)),
	}

	resp, err := conn.CreateEnvironmentMembership(params)

	if err != nil {
		return fmt.Errorf("Error creating Cloud9 Environment Membership: %s", err)
	}

	d.SetId(fmt.Sprintf("%s#%s", *resp.Membership.EnvironmentId, *resp.Membership.UserArn))

	return resourceAwsCloud9EnvironmentMembershipRead(d, meta)
}

func resourceAwsCloud9EnvironmentMembershipRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloud9conn

	log.Printf("[INFO] Reading Cloud9 Environment Membership %s", d.Id())

	out, err := conn.DescribeEnvironmentMemberships(&cloud9.DescribeEnvironmentMembershipsInput{
		EnvironmentId: aws.String(d.Get("environment_id").(string)),
		UserArn:       aws.String(d.Get("user_arn").(string)),
	})
	if err != nil {
		if isAWSErr(err, cloud9.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] Cloud9 Environment Membership (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}
	if len(out.Memberships) == 0 {
		log.Printf("[WARN] Cloud9 Environment Membership (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	env := out.Memberships[0]

	d.Set("environment_id", env.EnvironmentId)
	d.Set("user_arn", env.UserArn)
	d.Set("permissions", env.Permissions)
	d.Set("user_id", env.UserId)

	log.Printf("[DEBUG] Received Cloud9 Environment Membership: %s", env)

	return nil
}

func resourceAwsCloud9EnvironmentMembershipUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloud9conn

	input := cloud9.UpdateEnvironmentMembershipInput{
		EnvironmentId: aws.String(d.Get("environment_id").(string)),
		Permissions:   aws.String(d.Get("permissions").(string)),
		UserArn:       aws.String(d.Get("user_arn").(string)),
	}

	log.Printf("[INFO] Updating Cloud9 Environment Membership: %s", input)

	out, err := conn.UpdateEnvironmentMembership(&input)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Cloud9 Environment Membership updated: %s", out)

	return resourceAwsCloud9EnvironmentMembershipRead(d, meta)
}

func resourceAwsCloud9EnvironmentMembershipDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloud9conn

	_, err := conn.DeleteEnvironmentMembership(&cloud9.DeleteEnvironmentMembershipInput{
		EnvironmentId: aws.String(d.Get("environment_id").(string)),
		UserArn:       aws.String(d.Get("user_arn").(string)),
	})
	if err != nil {
		return fmt.Errorf("error deleting Cloud9 Environment Membership (%s): %s", d.Id(), err)
	}
	return nil
}
