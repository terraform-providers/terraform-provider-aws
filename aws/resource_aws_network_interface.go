package aws

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/hashcode"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsNetworkInterface() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsNetworkInterfaceCreate,
		Read:   resourceAwsNetworkInterfaceRead,
		Update: resourceAwsNetworkInterfaceUpdate,
		Delete: resourceAwsNetworkInterfaceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{

			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"mac_address": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"private_ip": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"private_dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"private_ips": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"private_ips_count": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"security_groups": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"source_dest_check": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"outpost_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"attachment": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"instance": {
							Type:     schema.TypeString,
							Required: true,
						},
						"device_index": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"attachment_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Set: resourceAwsEniAttachmentHash,
			},

			"tags": tagsSchema(),
			"ipv6_address_count": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"ipv6_addresses"},
			},
			"ipv6_addresses": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsIPv6Address,
				},
				ConflictsWith: []string{"ipv6_address_count"},
			},
		},
	}
}

func resourceAwsNetworkInterfaceCreate(d *schema.ResourceData, meta interface{}) error {

	conn := meta.(*AWSClient).ec2conn

	request := &ec2.CreateNetworkInterfaceInput{
		SubnetId:          aws.String(d.Get("subnet_id").(string)),
		TagSpecifications: ec2TagSpecificationsFromMap(d.Get("tags").(map[string]interface{}), ec2.ResourceTypeNetworkInterface),
	}

	if v, ok := d.GetOk("security_groups"); ok && v.(*schema.Set).Len() > 0 {
		request.Groups = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("private_ips"); ok && v.(*schema.Set).Len() > 0 {
		request.PrivateIpAddresses = expandPrivateIPAddresses(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("description"); ok {
		request.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("private_ips_count"); ok {
		request.SecondaryPrivateIpAddressCount = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("ipv6_address_count"); ok {
		request.Ipv6AddressCount = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("ipv6_addresses"); ok && v.(*schema.Set).Len() > 0 {
		request.Ipv6Addresses = expandIP6Addresses(v.(*schema.Set).List())
	}

	log.Printf("[DEBUG] Creating network interface")
	resp, err := conn.CreateNetworkInterface(request)
	if err != nil {
		return fmt.Errorf("Error creating ENI: %s", err)
	}

	d.SetId(*resp.NetworkInterface.NetworkInterfaceId)

	if err := waitForNetworkInterfaceCreation(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for Network Interface (%s) creation: %s", d.Id(), err)
	}

	//Default value is enabled
	if !d.Get("source_dest_check").(bool) {
		request := &ec2.ModifyNetworkInterfaceAttributeInput{
			NetworkInterfaceId: aws.String(d.Id()),
			SourceDestCheck:    &ec2.AttributeBooleanValue{Value: aws.Bool(false)},
		}

		_, err := conn.ModifyNetworkInterfaceAttribute(request)
		if err != nil {
			return fmt.Errorf("Failure updating SourceDestCheck: %s", err)
		}
	}

	if v, ok := d.GetOk("attachment"); ok && v.(*schema.Set).Len() > 0 {
		attachment := v.(*schema.Set).List()[0].(map[string]interface{})
		di := attachment["device_index"].(int)
		attachReq := &ec2.AttachNetworkInterfaceInput{
			DeviceIndex:        aws.Int64(int64(di)),
			InstanceId:         aws.String(attachment["instance"].(string)),
			NetworkInterfaceId: aws.String(d.Id()),
		}
		_, err := conn.AttachNetworkInterface(attachReq)
		if err != nil {
			return fmt.Errorf("Error attaching ENI: %s", err)
		}
	}

	return resourceAwsNetworkInterfaceRead(d, meta)
}

func resourceAwsNetworkInterfaceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	describe_network_interfaces_request := &ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []*string{aws.String(d.Id())},
	}
	describeResp, err := conn.DescribeNetworkInterfaces(describe_network_interfaces_request)

	if err != nil {
		if isAWSErr(err, "InvalidNetworkInterfaceID.NotFound", "") {
			// The ENI is gone now, so just remove it from the state
			log.Printf("[WARN] EC2 Network Interface (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error retrieving ENI: %s", err)
	}
	if len(describeResp.NetworkInterfaces) != 1 {
		return fmt.Errorf("Unable to find ENI: %#v", describeResp.NetworkInterfaces)
	}

	eni := describeResp.NetworkInterfaces[0]

	attachment := []map[string]interface{}{}

	if eni.Attachment != nil {
		attachment = []map[string]interface{}{flattenAttachment(eni.Attachment)}
	}

	if err := d.Set("attachment", attachment); err != nil {
		return fmt.Errorf("error setting attachment: %s", err)
	}

	d.Set("description", eni.Description)
	d.Set("private_dns_name", eni.PrivateDnsName)
	d.Set("mac_address", eni.MacAddress)
	d.Set("private_ip", eni.PrivateIpAddress)
	d.Set("outpost_arn", eni.OutpostArn)

	if err := d.Set("private_ips", flattenNetworkInterfacesPrivateIPAddresses(eni.PrivateIpAddresses)); err != nil {
		return fmt.Errorf("error setting private_ips: %s", err)
	}

	d.Set("private_ips_count", len(eni.PrivateIpAddresses)-1)

	if err := d.Set("security_groups", flattenGroupIdentifiers(eni.Groups)); err != nil {
		return fmt.Errorf("error setting security_groups: %s", err)
	}

	d.Set("source_dest_check", eni.SourceDestCheck)
	d.Set("subnet_id", eni.SubnetId)
	d.Set("ipv6_address_count", len(eni.Ipv6Addresses))

	if err := d.Set("ipv6_addresses", flattenEc2NetworkInterfaceIpv6Address(eni.Ipv6Addresses)); err != nil {
		return fmt.Errorf("error setting ipv6 addresses: %s", err)
	}

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(eni.TagSet).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func networkInterfaceAttachmentRefreshFunc(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		describe_network_interfaces_request := &ec2.DescribeNetworkInterfacesInput{
			NetworkInterfaceIds: []*string{aws.String(id)},
		}
		describeResp, err := conn.DescribeNetworkInterfaces(describe_network_interfaces_request)

		if err != nil {
			log.Printf("[ERROR] Could not find network interface %s. %s", id, err)
			return nil, "", err
		}

		eni := describeResp.NetworkInterfaces[0]
		hasAttachment := strconv.FormatBool(eni.Attachment != nil)
		log.Printf("[DEBUG] ENI %s has attachment state %s", id, hasAttachment)
		return eni, hasAttachment, nil
	}
}

func resourceAwsNetworkInterfaceDetach(oa *schema.Set, meta interface{}, eniId string) error {
	// if there was an old attachment, remove it
	if oa != nil && len(oa.List()) > 0 {
		old_attachment := oa.List()[0].(map[string]interface{})
		detach_request := &ec2.DetachNetworkInterfaceInput{
			AttachmentId: aws.String(old_attachment["attachment_id"].(string)),
			Force:        aws.Bool(true),
		}
		conn := meta.(*AWSClient).ec2conn
		_, detach_err := conn.DetachNetworkInterface(detach_request)
		if detach_err != nil {
			if !isAWSErr(detach_err, "InvalidAttachmentID.NotFound", "") {
				return fmt.Errorf("Error detaching ENI: %s", detach_err)
			}
		}

		log.Printf("[DEBUG] Waiting for ENI (%s) to become detached", eniId)
		stateConf := &resource.StateChangeConf{
			Pending: []string{"true"},
			Target:  []string{"false"},
			Refresh: networkInterfaceAttachmentRefreshFunc(conn, eniId),
			Timeout: 10 * time.Minute,
		}
		if _, err := stateConf.WaitForState(); err != nil {
			return fmt.Errorf(
				"Error waiting for ENI (%s) to become detached: %s", eniId, err)
		}
	}

	return nil
}

func resourceAwsNetworkInterfaceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	if d.HasChange("attachment") {
		oa, na := d.GetChange("attachment")

		detach_err := resourceAwsNetworkInterfaceDetach(oa.(*schema.Set), meta, d.Id())
		if detach_err != nil {
			return detach_err
		}

		// if there is a new attachment, attach it
		if na != nil && len(na.(*schema.Set).List()) > 0 {
			new_attachment := na.(*schema.Set).List()[0].(map[string]interface{})
			di := new_attachment["device_index"].(int)
			attach_request := &ec2.AttachNetworkInterfaceInput{
				DeviceIndex:        aws.Int64(int64(di)),
				InstanceId:         aws.String(new_attachment["instance"].(string)),
				NetworkInterfaceId: aws.String(d.Id()),
			}
			_, attach_err := conn.AttachNetworkInterface(attach_request)
			if attach_err != nil {
				return fmt.Errorf("Error attaching ENI: %s", attach_err)
			}
		}
	}

	if d.HasChange("private_ips") {
		o, n := d.GetChange("private_ips")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		// Unassign old IP addresses
		unassignIps := os.Difference(ns)
		if unassignIps.Len() != 0 {
			input := &ec2.UnassignPrivateIpAddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				PrivateIpAddresses: expandStringSet(unassignIps),
			}
			_, err := conn.UnassignPrivateIpAddresses(input)
			if err != nil {
				return fmt.Errorf("Failure to unassign Private IPs: %s", err)
			}
		}

		// Assign new IP addresses
		assignIps := ns.Difference(os)
		if assignIps.Len() != 0 {
			input := &ec2.AssignPrivateIpAddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				PrivateIpAddresses: expandStringSet(assignIps),
			}
			_, err := conn.AssignPrivateIpAddresses(input)
			if err != nil {
				return fmt.Errorf("Failure to assign Private IPs: %s", err)
			}
		}
	}

	if d.HasChange("ipv6_addresses") {
		o, n := d.GetChange("ipv6_addresses")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		// Unassign old IPV6 addresses
		unassignIps := os.Difference(ns)
		if unassignIps.Len() != 0 {
			input := &ec2.UnassignIpv6AddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				Ipv6Addresses:      expandStringSet(unassignIps),
			}
			_, err := conn.UnassignIpv6Addresses(input)
			if err != nil {
				return fmt.Errorf("failure to unassign IPV6 Addresses: %s", err)
			}
		}

		// Assign new IPV6 addresses
		assignIps := ns.Difference(os)
		if assignIps.Len() != 0 {
			input := &ec2.AssignIpv6AddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				Ipv6Addresses:      expandStringSet(assignIps),
			}
			_, err := conn.AssignIpv6Addresses(input)
			if err != nil {
				return fmt.Errorf("Failure to assign IPV6 Addresses: %s", err)
			}
		}
	}

	if d.HasChange("ipv6_address_count") {
		o, n := d.GetChange("ipv6_address_count")
		ipv6Addresses := d.Get("ipv6_addresses").(*schema.Set).List()

		if o != nil && n != nil && n != len(ipv6Addresses) {

			diff := n.(int) - o.(int)

			// Surplus of IPs, add the diff
			if diff > 0 {
				input := &ec2.AssignIpv6AddressesInput{
					NetworkInterfaceId: aws.String(d.Id()),
					Ipv6AddressCount:   aws.Int64(int64(diff)),
				}
				_, err := conn.AssignIpv6Addresses(input)
				if err != nil {
					return fmt.Errorf("failure to assign IPV6 Addresses: %s", err)
				}
			}

			if diff < 0 {
				input := &ec2.UnassignIpv6AddressesInput{
					NetworkInterfaceId: aws.String(d.Id()),
					Ipv6Addresses:      expandStringList(ipv6Addresses[0:int(math.Abs(float64(diff)))]),
				}
				_, err := conn.UnassignIpv6Addresses(input)
				if err != nil {
					return fmt.Errorf("failure to unassign IPV6 Addresses: %s", err)
				}
			}
		}
	}

	if d.HasChange("source_dest_check") {
		request := &ec2.ModifyNetworkInterfaceAttributeInput{
			NetworkInterfaceId: aws.String(d.Id()),
			SourceDestCheck:    &ec2.AttributeBooleanValue{Value: aws.Bool(d.Get("source_dest_check").(bool))},
		}

		_, err := conn.ModifyNetworkInterfaceAttribute(request)
		if err != nil {
			return fmt.Errorf("failure updating Source Dest Check on ENI: %s", err)
		}
	}

	if d.HasChange("private_ips_count") {
		o, n := d.GetChange("private_ips_count")
		privateIPs := d.Get("private_ips").(*schema.Set).List()
		privateIPsFiltered := privateIPs[:0]
		primaryIP := d.Get("private_ip")

		for _, ip := range privateIPs {
			if ip != primaryIP {
				privateIPsFiltered = append(privateIPsFiltered, ip)
			}
		}

		if o != nil && n != nil && n != len(privateIPsFiltered) {

			diff := n.(int) - o.(int)

			// Surplus of IPs, add the diff
			if diff > 0 {
				input := &ec2.AssignPrivateIpAddressesInput{
					NetworkInterfaceId:             aws.String(d.Id()),
					SecondaryPrivateIpAddressCount: aws.Int64(int64(diff)),
				}
				_, err := conn.AssignPrivateIpAddresses(input)
				if err != nil {
					return fmt.Errorf("Failure to assign Private IPs: %s", err)
				}
			}

			if diff < 0 {
				input := &ec2.UnassignPrivateIpAddressesInput{
					NetworkInterfaceId: aws.String(d.Id()),
					PrivateIpAddresses: expandStringList(privateIPsFiltered[0:int(math.Abs(float64(diff)))]),
				}
				_, err := conn.UnassignPrivateIpAddresses(input)
				if err != nil {
					return fmt.Errorf("Failure to unassign Private IPs: %s", err)
				}
			}
		}
	}

	if d.HasChange("security_groups") {
		request := &ec2.ModifyNetworkInterfaceAttributeInput{
			NetworkInterfaceId: aws.String(d.Id()),
			Groups:             expandStringSet(d.Get("security_groups").(*schema.Set)),
		}

		_, err := conn.ModifyNetworkInterfaceAttribute(request)
		if err != nil {
			return fmt.Errorf("Failure updating ENI: %s", err)
		}
	}

	if d.HasChange("description") {
		request := &ec2.ModifyNetworkInterfaceAttributeInput{
			NetworkInterfaceId: aws.String(d.Id()),
			Description:        &ec2.AttributeValue{Value: aws.String(d.Get("description").(string))},
		}

		_, err := conn.ModifyNetworkInterfaceAttribute(request)
		if err != nil {
			return fmt.Errorf("Failure updating ENI: %s", err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.Ec2UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Network Interface (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsNetworkInterfaceRead(d, meta)
}

func resourceAwsNetworkInterfaceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	log.Printf("[INFO] Deleting ENI: %s", d.Id())

	if err := resourceAwsNetworkInterfaceDetach(d.Get("attachment").(*schema.Set), meta, d.Id()); err != nil {
		return err
	}

	deleteEniOpts := ec2.DeleteNetworkInterfaceInput{
		NetworkInterfaceId: aws.String(d.Id()),
	}
	if _, err := conn.DeleteNetworkInterface(&deleteEniOpts); err != nil {
		return fmt.Errorf("Error deleting ENI: %s", err)
	}

	return nil
}

func resourceAwsEniAttachmentHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["instance"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["device_index"].(int)))
	return hashcode.String(buf.String())
}

func deleteNetworkInterface(conn *ec2.EC2, eniId string) error {
	_, err := conn.DeleteNetworkInterface(&ec2.DeleteNetworkInterfaceInput{
		NetworkInterfaceId: aws.String(eniId),
	})

	if isAWSErr(err, "InvalidNetworkInterfaceID.NotFound", "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting ENI (%s): %s", eniId, err)
	}

	return nil
}

func detachNetworkInterface(conn *ec2.EC2, eni *ec2.NetworkInterface, timeout time.Duration) error {
	if eni == nil {
		return nil
	}

	eniId := aws.StringValue(eni.NetworkInterfaceId)
	if eni.Attachment == nil {
		log.Printf("[DEBUG] ENI %s is already detached", eniId)
		return nil
	}

	_, err := conn.DetachNetworkInterface(&ec2.DetachNetworkInterfaceInput{
		AttachmentId: eni.Attachment.AttachmentId,
		Force:        aws.Bool(true),
	})

	if isAWSErr(err, "InvalidAttachmentID.NotFound", "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error detaching ENI (%s): %s", eniId, err)
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{
			ec2.AttachmentStatusAttaching,
			ec2.AttachmentStatusAttached,
			ec2.AttachmentStatusDetaching,
		},
		Target: []string{
			ec2.AttachmentStatusDetached,
		},
		Refresh:        networkInterfaceAttachmentStateRefresh(conn, eniId),
		Timeout:        timeout,
		Delay:          10 * time.Second,
		MinTimeout:     5 * time.Second,
		NotFoundChecks: 1,
	}

	log.Printf("[DEBUG] Waiting for ENI (%s) to become detached", eniId)
	_, err = stateConf.WaitForState()

	if isResourceNotFoundError(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error waiting for ENI (%s) to become detached: %s", eniId, err)
	}

	return nil
}

func networkInterfaceAttachmentStateRefresh(conn *ec2.EC2, eniId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{
			NetworkInterfaceIds: aws.StringSlice([]string{eniId}),
		})

		if isAWSErr(err, "InvalidNetworkInterfaceID.NotFound", "") {
			return nil, ec2.AttachmentStatusDetached, nil
		}

		if err != nil {
			return nil, "", fmt.Errorf("error describing ENI (%s): %s", eniId, err)
		}

		n := len(resp.NetworkInterfaces)
		switch n {
		case 0:
			return nil, ec2.AttachmentStatusDetached, nil

		case 1:
			attachment := resp.NetworkInterfaces[0].Attachment
			if attachment == nil {
				return nil, ec2.AttachmentStatusDetached, nil
			}
			return attachment, aws.StringValue(attachment.Status), nil

		default:
			return nil, "", fmt.Errorf("found %d ENIs for %s, expected 1", n, eniId)
		}
	}
}

func networkInterfaceStateRefresh(conn *ec2.EC2, eniId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{
			NetworkInterfaceIds: aws.StringSlice([]string{eniId}),
		})

		if isAWSErr(err, "InvalidNetworkInterfaceID.NotFound", "") {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", fmt.Errorf("error describing ENI (%s): %s", eniId, err)
		}

		n := len(resp.NetworkInterfaces)
		switch n {
		case 0:
			return nil, "", nil

		case 1:
			eni := resp.NetworkInterfaces[0]
			return eni, aws.StringValue(eni.Status), nil

		default:
			return nil, "", fmt.Errorf("found %d ENIs for %s, expected 1", n, eniId)
		}
	}
}

func waitForNetworkInterfaceCreation(conn *ec2.EC2, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{ec2.NetworkInterfaceStatusAvailable},
		Refresh: networkInterfaceStateRefresh(conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}
