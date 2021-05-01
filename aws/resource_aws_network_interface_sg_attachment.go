package aws

import (
	"fmt"
	"log"
	"reflect"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsNetworkInterfaceSGAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsNetworkInterfaceSGAttachmentCreate,
		Read:   resourceAwsNetworkInterfaceSGAttachmentRead,
		Delete: resourceAwsNetworkInterfaceSGAttachmentDelete,
		Schema: map[string]*schema.Schema{
			"security_group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"network_interface_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsNetworkInterfaceSGAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	mk := "network_interface_sg_attachment_" + d.Get("network_interface_id").(string)
	awsMutexKV.Lock(mk)
	defer awsMutexKV.Unlock(mk)

	sgID := d.Get("security_group_id").(string)
	interfaceID := d.Get("network_interface_id").(string)

	conn := meta.(*AWSClient).ec2conn

	iface, err := finder.NetworkInterfaceByID(conn, interfaceID)

	if err != nil {
		return fmt.Errorf("error reading EC2 Network Interface (%s): %w", interfaceID, err)
	}

	groupIDs := []string{sgID}

	for _, group := range iface.Groups {
		if group == nil {
			continue
		}

		if aws.StringValue(group.GroupId) == sgID {
			return fmt.Errorf("EC2 Security Group (%s) already attached to EC2 Network Interface ID: %s", sgID, interfaceID)
		}

		groupIDs = append(groupIDs, aws.StringValue(group.GroupId))
	}

	params := &ec2.ModifyNetworkInterfaceAttributeInput{
		NetworkInterfaceId: iface.NetworkInterfaceId,
		Groups:             aws.StringSlice(groupIDs),
	}

	_, err = conn.ModifyNetworkInterfaceAttribute(params)

	if err != nil {
		return fmt.Errorf("error modifying EC2 Network Interface (%s): %w", interfaceID, err)
	}

	d.SetId(fmt.Sprintf("%s_%s", sgID, interfaceID))

	return resourceAwsNetworkInterfaceSGAttachmentRead(d, meta)
}

func resourceAwsNetworkInterfaceSGAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	sgID := d.Get("security_group_id").(string)
	interfaceID := d.Get("network_interface_id").(string)

	log.Printf("[DEBUG] Checking association of security group %s to network interface ID %s", sgID, interfaceID)

	conn := meta.(*AWSClient).ec2conn

	var groupIdentifier *ec2.GroupIdentifier

	err := resource.Retry(waiter.PropagationTimeout, func() *resource.RetryError {
		var err error

		groupIdentifier, err = finder.NetworkInterfaceSecurityGroup(conn, interfaceID, sgID)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidNetworkInterfaceIDNotFound) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if d.IsNewResource() && groupIdentifier == nil {
			return resource.RetryableError(&resource.NotFoundError{
				LastError: fmt.Errorf("EC2 Network Interface Security Group Attachment (%s) not found", d.Id()),
			})
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		groupIdentifier, err = finder.NetworkInterfaceSecurityGroup(conn, interfaceID, sgID)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidNetworkInterfaceIDNotFound) {
		log.Printf("[WARN] EC2 Network Interface Security Group Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Network Interface Security Group Attachment (%s): %w", d.Id(), err)
	}

	if groupIdentifier == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error reading EC2 Network Interface Security Group Attachment (%s): not found after creation", d.Id())
		}

		log.Printf("[WARN] EC2 Network Interface Security Group Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("network_interface_id", interfaceID)
	d.Set("security_group_id", groupIdentifier.GroupId)

	return nil
}

func resourceAwsNetworkInterfaceSGAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	mk := "network_interface_sg_attachment_" + d.Get("network_interface_id").(string)
	awsMutexKV.Lock(mk)
	defer awsMutexKV.Unlock(mk)

	sgID := d.Get("security_group_id").(string)
	interfaceID := d.Get("network_interface_id").(string)

	log.Printf("[DEBUG] Removing security group %s from interface ID %s", sgID, interfaceID)

	conn := meta.(*AWSClient).ec2conn

	iface, err := finder.NetworkInterfaceByID(conn, interfaceID)

	if isAWSErr(err, "InvalidNetworkInterfaceID.NotFound", "") {
		return nil
	}

	if err != nil {
		return err
	}

	return delSGFromENI(conn, sgID, iface)
}

func delSGFromENI(conn *ec2.EC2, sgID string, iface *ec2.NetworkInterface) error {
	old := iface.Groups
	var new []*string
	for _, v := range iface.Groups {
		if *v.GroupId == sgID {
			continue
		}
		new = append(new, v.GroupId)
	}
	if reflect.DeepEqual(old, new) {
		// The interface already didn't have the security group, nothing to do
		return nil
	}

	params := &ec2.ModifyNetworkInterfaceAttributeInput{
		NetworkInterfaceId: iface.NetworkInterfaceId,
		Groups:             new,
	}

	_, err := conn.ModifyNetworkInterfaceAttribute(params)

	if isAWSErr(err, "InvalidNetworkInterfaceID.NotFound", "") {
		return nil
	}

	return err
}
