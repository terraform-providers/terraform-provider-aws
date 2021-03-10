package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsSnapshotCreateVolumePermission() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSnapshotCreateVolumePermissionCreate,
		Read:   resourceAwsSnapshotCreateVolumePermissionRead,
		Delete: resourceAwsSnapshotCreateVolumePermissionDelete,

		Schema: map[string]*schema.Schema{
			"snapshot_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"account_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"group"},
			},
			"group": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"account_id"},
			},
		},
	}
}

func resourceAwsSnapshotCreateVolumePermissionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	snapshot_id := d.Get("snapshot_id").(string)
	account_id := ""
	group := ""

	if v, ok := d.GetOk("account_id"); ok {
		account_id = v.(string)
	} else if v, ok := d.GetOk("group"); ok {
		group = v.(string)
	}

	if len(group) > 0 {
		_, err := conn.ModifySnapshotAttribute(&ec2.ModifySnapshotAttributeInput{
			SnapshotId: aws.String(snapshot_id),
			Attribute:  aws.String("createVolumePermission"),
			CreateVolumePermission: &ec2.CreateVolumePermissionModifications{
				Add: []*ec2.CreateVolumePermission{
					{Group: aws.String(group)},
				},
			},
		})

		if err != nil {
			return fmt.Errorf("Error adding snapshot createVolumePermission: %s", err)
		}
	} else {
		_, err := conn.ModifySnapshotAttribute(&ec2.ModifySnapshotAttributeInput{
			SnapshotId: aws.String(snapshot_id),
			Attribute:  aws.String("createVolumePermission"),
			CreateVolumePermission: &ec2.CreateVolumePermissionModifications{
				Add: []*ec2.CreateVolumePermission{
					{UserId: aws.String(account_id)},
				},
			},
		})

		if err != nil {
			return fmt.Errorf("Error adding snapshot createVolumePermission: %s", err)
		}
	}

	d.SetId(fmt.Sprintf("%s-%s-%s", snapshot_id, account_id, group))

	// Wait for the account to appear in the permission list
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"denied"},
		Target:     []string{"granted"},
		Refresh:    resourceAwsSnapshotCreateVolumePermissionStateRefreshFunc(conn, snapshot_id, account_id, group),
		Timeout:    20 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf(
			"Error waiting for snapshot createVolumePermission (%s) to be added: %s",
			d.Id(), err)
	}

	return nil
}

func resourceAwsSnapshotCreateVolumePermissionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	snapshotID, accountID, group, err := resourceAwsSnapshotCreateVolumePermissionParseID(d.Id())
	if err != nil {
		return err
	}

	exists, err := hasCreateVolumePermission(conn, snapshotID, accountID, group)
	if err != nil {
		return err
	}
	if !exists {
		log.Printf("[WARN] snapshot createVolumePermission (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	return nil
}

func resourceAwsSnapshotCreateVolumePermissionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	snapshotID, accountID, group, err := resourceAwsSnapshotCreateVolumePermissionParseID(d.Id())
	if err != nil {
		return err
	}

	if len(group) > 0 {
		_, err := conn.ModifySnapshotAttribute(&ec2.ModifySnapshotAttributeInput{
			SnapshotId: aws.String(snapshotID),
			Attribute:  aws.String("createVolumePermission"),
			CreateVolumePermission: &ec2.CreateVolumePermissionModifications{
				Remove: []*ec2.CreateVolumePermission{
					{Group: aws.String(group)},
				},
			},
		})

		if err != nil {
			return fmt.Errorf("Error removing snapshot createVolumePermission: %s", err)
		}
	} else {
		_, err := conn.ModifySnapshotAttribute(&ec2.ModifySnapshotAttributeInput{
			SnapshotId: aws.String(snapshotID),
			Attribute:  aws.String("createVolumePermission"),
			CreateVolumePermission: &ec2.CreateVolumePermissionModifications{
				Remove: []*ec2.CreateVolumePermission{
					{UserId: aws.String(accountID)},
				},
			},
		})

		if err != nil {
			return fmt.Errorf("Error removing snapshot createVolumePermission: %s", err)
		}
	}

	// Wait for the account to disappear from the permission list
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"granted"},
		Target:     []string{"denied"},
		Refresh:    resourceAwsSnapshotCreateVolumePermissionStateRefreshFunc(conn, snapshotID, accountID, group),
		Timeout:    5 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf(
			"Error waiting for snapshot createVolumePermission (%s) to be removed: %s",
			d.Id(), err)
	}

	return nil
}

func hasCreateVolumePermission(conn *ec2.EC2, snapshot_id string, account_id string, group string) (bool, error) {
	_, state, err := resourceAwsSnapshotCreateVolumePermissionStateRefreshFunc(conn, snapshot_id, account_id, group)()
	if err != nil {
		return false, err
	}
	if state == "granted" {
		return true, nil
	} else {
		return false, nil
	}
}

func resourceAwsSnapshotCreateVolumePermissionStateRefreshFunc(conn *ec2.EC2, snapshot_id string, account_id string, group string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		attrs, err := conn.DescribeSnapshotAttribute(&ec2.DescribeSnapshotAttributeInput{
			SnapshotId: aws.String(snapshot_id),
			Attribute:  aws.String("createVolumePermission"),
		})
		if err != nil {
			return nil, "", fmt.Errorf("Error refreshing snapshot createVolumePermission state: %s", err)
		}

		for _, vp := range attrs.CreateVolumePermissions {
			if (aws.StringValue(vp.UserId) == account_id) || (aws.StringValue(vp.Group) == group) {
				return attrs, "granted", nil
			}
		}
		return attrs, "denied", nil
	}
}

func resourceAwsSnapshotCreateVolumePermissionParseID(id string) (string, string, string, error) {
	idParts := strings.SplitN(id, "-", 4)
	if len(idParts) != 4 || idParts[0] != "snap" || idParts[1] == "" || (idParts[2] == "" && idParts[3] == "") {
		return "", "", "", fmt.Errorf("unexpected format of ID (%s), expected SNAPSHOT_ID-ACCOUNT_ID-GROUP", id)
	}
	return fmt.Sprintf("%s-%s", idParts[0], idParts[1]), idParts[2], idParts[3], nil
}
