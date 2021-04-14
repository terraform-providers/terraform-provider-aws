package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsIamGroupMembership() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIamGroupMembershipCreate,
		Read:   resourceAwsIamGroupMembershipRead,
		Update: resourceAwsIamGroupMembershipUpdate,
		Delete: resourceAwsIamGroupMembershipDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"users": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"group": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsIamGroupMembershipCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn

	group := d.Get("group").(string)
	userList := expandStringSet(d.Get("users").(*schema.Set))

	if err := addUsersToGroup(conn, userList, group); err != nil {
		return err
	}

	d.SetId(d.Get("name").(string))
	return resourceAwsIamGroupMembershipRead(d, meta)
}

func resourceAwsIamGroupMembershipRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn
	group := d.Get("group").(string)

	input := &iam.GetGroupInput{
		GroupName: aws.String(group),
	}

	var ul []string

	err := resource.Retry(waiter.PropagationTimeout, func() *resource.RetryError {
		err := conn.GetGroupPages(input, func(page *iam.GetGroupOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, user := range page.Users {
				ul = append(ul, aws.StringValue(user.UserName))
			}

			return !lastPage
		})

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		err = conn.GetGroupPages(input, func(page *iam.GetGroupOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, user := range page.Users {
				ul = append(ul, aws.StringValue(user.UserName))
			}

			return !lastPage
		})
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		log.Printf("[WARN] IAM Group Membership (%s) not found, removing from state", group)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading IAM Group Membership (%s): %w", group, err)
	}

	if err := d.Set("users", ul); err != nil {
		return fmt.Errorf("Error setting user list from IAM Group Membership (%s), error: %s", group, err)
	}

	return nil
}

func resourceAwsIamGroupMembershipUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn

	if d.HasChange("users") {
		group := d.Get("group").(string)

		o, n := d.GetChange("users")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)
		remove := expandStringSet(os.Difference(ns))
		add := expandStringSet(ns.Difference(os))

		if err := removeUsersFromGroup(conn, remove, group); err != nil {
			return err
		}

		if err := addUsersToGroup(conn, add, group); err != nil {
			return err
		}
	}

	return resourceAwsIamGroupMembershipRead(d, meta)
}

func resourceAwsIamGroupMembershipDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn
	userList := expandStringSet(d.Get("users").(*schema.Set))
	group := d.Get("group").(string)

	err := removeUsersFromGroup(conn, userList, group)
	return err
}

func removeUsersFromGroup(conn *iam.IAM, users []*string, group string) error {
	for _, u := range users {
		_, err := conn.RemoveUserFromGroup(&iam.RemoveUserFromGroupInput{
			UserName:  u,
			GroupName: aws.String(group),
		})

		if err != nil {
			if iamerr, ok := err.(awserr.Error); ok && iamerr.Code() == "NoSuchEntity" {
				return nil
			}
			return err
		}
	}
	return nil
}

func addUsersToGroup(conn *iam.IAM, users []*string, group string) error {
	for _, u := range users {
		_, err := conn.AddUserToGroup(&iam.AddUserToGroupInput{
			UserName:  u,
			GroupName: aws.String(group),
		})

		if err != nil {
			return err
		}
	}
	return nil
}
