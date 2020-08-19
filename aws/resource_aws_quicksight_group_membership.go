package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
)

func resourceAwsQuickSightGroupMembership() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsQuickSightGroupMembershipCreate,
		Read:   resourceAwsQuickSightGroupMembershipRead,
		Delete: resourceAwsQuickSightGroupMembershipDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"aws_account_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"member_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"group_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"namespace": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "default",
				ValidateFunc: validation.StringInSlice([]string{
					"default",
				}, false),
			},
		},
	}
}

func resourceAwsQuickSightGroupMembershipCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).quicksightconn

	awsAccountID := meta.(*AWSClient).accountid
	namespace := d.Get("namespace").(string)
	groupName := d.Get("group_name").(string)

	if v, ok := d.GetOk("aws_account_id"); ok {
		awsAccountID = v.(string)
	}

	createOpts := &quicksight.CreateGroupMembershipInput{
		AwsAccountId: aws.String(awsAccountID),
		GroupName:    aws.String(groupName),
		MemberName:   aws.String(d.Get("member_name").(string)),
		Namespace:    aws.String(namespace),
	}

	resp, err := conn.CreateGroupMembership(createOpts)
	if err != nil {
		return fmt.Errorf("Error adding QuickSight user to group: %s", err)
	}

	d.SetId(fmt.Sprintf("%s/%s/%s/%s", awsAccountID, namespace, groupName, aws.StringValue(resp.GroupMember.MemberName)))

	return resourceAwsQuickSightGroupMembershipRead(d, meta)
}

func resourceAwsQuickSightGroupMembershipRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).quicksightconn

	awsAccountID, namespace, groupName, userName, err := resourceAwsQuickSightGroupMembershipParseID(d.Id())
	if err != nil {
		return err
	}

	listOpts := &quicksight.ListUserGroupsInput{
		AwsAccountId: aws.String(awsAccountID),
		Namespace:    aws.String(namespace),
		UserName:     aws.String(userName),
	}

	found := false

	for {
		resp, err := conn.ListUserGroups(listOpts)
		if isAWSErr(err, quicksight.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] QuickSight User %s is not found", d.Id())
			d.SetId("")
			return nil
		}
		if err != nil {
			return fmt.Errorf("Error listing QuickSight User groups (%s): %s", d.Id(), err)
		}

		for _, group := range resp.GroupList {
			if *group.GroupName == groupName {
				found = true
				break
			}
		}

		if found || resp.NextToken == nil {
			break
		}

		listOpts.NextToken = resp.NextToken
	}

	if found {
		d.Set("aws_account_id", awsAccountID)
		d.Set("namespace", namespace)
		d.Set("member_name", userName)
		d.Set("group_name", groupName)
	} else {
		log.Printf("[WARN] QuickSight User-group membership %s is not found", d.Id())
		d.SetId("")
	}

	return nil
}

func resourceAwsQuickSightGroupMembershipDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).quicksightconn

	awsAccountID, namespace, groupName, userName, err := resourceAwsQuickSightGroupMembershipParseID(d.Id())
	if err != nil {
		return err
	}

	deleteOpts := &quicksight.DeleteGroupMembershipInput{
		AwsAccountId: aws.String(awsAccountID),
		Namespace:    aws.String(namespace),
		MemberName:   aws.String(userName),
		GroupName:    aws.String(groupName),
	}

	if _, err := conn.DeleteGroupMembership(deleteOpts); err != nil {
		if isAWSErr(err, quicksight.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("Error deleting QuickSight User-group membership %s: %s", d.Id(), err)
	}

	return nil
}

func resourceAwsQuickSightGroupMembershipParseID(id string) (string, string, string, string, error) {
	parts := strings.SplitN(id, "/", 4)
	if len(parts) < 4 || parts[0] == "" || parts[1] == "" || parts[2] == "" || parts[3] == "" {
		return "", "", "", "", fmt.Errorf("unexpected format of ID (%s), expected AWS_ACCOUNT_ID/NAMESPACE/GROUP_NAME/USER_NAME", id)
	}
	return parts[0], parts[1], parts[2], parts[3], nil
}
