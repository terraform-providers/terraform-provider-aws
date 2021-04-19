package aws

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
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
				Type:          schema.TypeSet,
				Optional:      true,
				Computed:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"private_ip_list"},
			},

			"private_ips_count": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"private_ip_list"},
			},

			"private_ip_list": {
				Type:          schema.TypeList,
				Optional:      true,
				Computed:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"private_ips", "private_ips_count"},
			},

			"private_ip_list_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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
				ConflictsWith: []string{"ipv6_addresses", "ipv6_address_list"},
			},
			"ipv6_addresses": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsIPv6Address,
				},
				ConflictsWith: []string{"ipv6_address_count", "ipv6_address_list"},
			},
			"ipv6_address_list": {
				Type:          schema.TypeList,
				Optional:      true,
				Computed:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"ipv6_addresses", "ipv6_address_count"},
			},
			"ipv6_address_list_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
		CustomizeDiff: customdiff.Sequence(
			customdiff.ForceNewIf("private_ips", func(_ context.Context, d *schema.ResourceDiff, meta interface{}) bool {
				private_ip_list_enabled := d.Get("private_ip_list_enabled").(bool)
				if private_ip_list_enabled {
					return false
				}
				_, new := d.GetChange("private_ips")
				if new != nil {
					old_primary_ip := ""
					if v, ok := d.GetOk("private_ip_list"); ok {
						for _, ip := range v.([]interface{}) {
							old_primary_ip = ip.(string)
							break
						}
					}
					for _, ip := range new.(*schema.Set).List() {
						// no need for new resource if we still have the primary ip
						if old_primary_ip == ip.(string) {
							return false
						}
					}
					// new primary ip requires a new resource
					return true
				} else {
					return false
				}
			}),
			customdiff.ForceNewIf("private_ip_list", func(_ context.Context, d *schema.ResourceDiff, meta interface{}) bool {
				private_ip_list_enabled := d.Get("private_ip_list_enabled").(bool)
				if !private_ip_list_enabled {
					return false
				}
				old, new := d.GetChange("private_ip_list")
				if old != nil && new != nil {
					old_primary_ip := ""
					new_primary_ip := ""
					for _, ip := range old.([]interface{}) {
						old_primary_ip = ip.(string)
						break
					}
					for _, ip := range new.([]interface{}) {
						new_primary_ip = ip.(string)
						break
					}

					// change in primary private ip requires a new resource
					return old_primary_ip != new_primary_ip
				} else {
					return false
				}
			}),
			customdiff.ComputedIf("private_ips", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				if !diff.Get("private_ip_list_enabled").(bool) {
					// it is not computed if we are actively updating it
					if diff.HasChange("private_ips") {
						return false
					} else {
						return diff.HasChange("private_ips_count")
					}
				} else {
					return diff.HasChange("private_ip_list")
				}
			}),
			customdiff.ComputedIf("private_ips_count", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				if !diff.Get("private_ip_list_enabled").(bool) {
					// it is not computed if we are actively updating it
					if diff.HasChange("private_ips_count") {
						return false
					} else {
						// compute the new count if private_ips change
						return diff.HasChange("private_ips")
					}
				} else {
					// compute the new count if private_ip_list changes
					return diff.HasChange("private_ip_list")
				}
			}),
			customdiff.ComputedIf("private_ip_list", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				if diff.Get("private_ip_list_enabled").(bool) {
					// if the list is controlling it does not need to be computed
					return false
				} else {
					// list is not controlling so compute new list if private_ips or private_ips_count changes
					return diff.HasChange("private_ips") || diff.HasChange("private_ips_count") || diff.HasChange("private_ip_list")
				}
			}),
			customdiff.ComputedIf("ipv6_addresses", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				if !diff.Get("ipv6_address_list_enabled").(bool) {
					// it is not computed if we are actively updating it
					if diff.HasChange("private_ips") {
						return false
					} else {
						return diff.HasChange("ipv6_address_count")
					}
				} else {
					return diff.HasChange("ipv6_address_list")
				}
			}),
			customdiff.ComputedIf("ipv6_address_count", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				if !diff.Get("ipv6_address_list_enabled").(bool) {
					// it is not computed if we are actively updating it
					if diff.HasChange("ipv6_address_count") {
						return false
					} else {
						// compute the new count if ipv6_addresses change
						return diff.HasChange("ipv6_addresses")
					}
				} else {
					// compute the new count if ipv6_address_list changes
					return diff.HasChange("ipv6_address_list")
				}
			}),
			customdiff.ComputedIf("ipv6_address_list", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				if diff.Get("ipv6_address_list_enabled").(bool) {
					// if the list is controlling it does not need to be computed
					return false
				} else {
					// list is not controlling so compute new list if anything changes
					return diff.HasChange("ipv6_addresses") || diff.HasChange("ipv6_address_count") || diff.HasChange("ipv6_address_list")
				}
			}),
		),
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

	if d.Get("private_ip_list_enabled").(bool) {
		if v, ok := d.GetOk("private_ip_list"); ok && len(v.([]interface{})) > 0 {
			request.PrivateIpAddresses = expandPrivateIPAddresses(v.([]interface{}))
		}
	} else {
		if v, ok := d.GetOk("private_ips"); ok && v.(*schema.Set).Len() > 0 {
			private_ips := v.(*schema.Set).List()
			// total includes the primary
			total_private_ips := len(private_ips)
			// private_ips_count is for secondaries
			if v, ok := d.GetOk("private_ips_count"); ok {
				// reduce total count if necessary
				if v.(int)+1 < total_private_ips {
					total_private_ips = v.(int) + 1
				}
			}
			// truncate the list
			count_limited_ips := make([]interface{}, total_private_ips)
			for i, ip := range private_ips {
				count_limited_ips[i] = ip.(string)
				if i == total_private_ips-1 {
					break
				}
			}
			request.PrivateIpAddresses = expandPrivateIPAddresses(count_limited_ips)
		} else {
			if v, ok := d.GetOk("private_ips_count"); ok {
				request.SecondaryPrivateIpAddressCount = aws.Int64(int64(v.(int)))
			}
		}
	}

	if v, ok := d.GetOk("description"); ok {
		request.Description = aws.String(v.(string))
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

	d.SetId(aws.StringValue(resp.NetworkInterface.NetworkInterfaceId))

	if err := waitForNetworkInterfaceCreation(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for Network Interface (%s) creation: %s", d.Id(), err)
	}

	if !d.Get("private_ip_list_enabled").(bool) {
		// add more ips to match the count
		if v, ok := d.GetOk("private_ips"); ok && v.(*schema.Set).Len() > 0 {
			total_private_ips := v.(*schema.Set).Len()
			if private_ips_count, ok := d.GetOk("private_ips_count"); ok {
				if private_ips_count.(int)+1 > total_private_ips {
					input := &ec2.AssignPrivateIpAddressesInput{
						NetworkInterfaceId:             aws.String(d.Id()),
						SecondaryPrivateIpAddressCount: aws.Int64(int64(private_ips_count.(int) + 1 - total_private_ips)),
					}
					_, err := conn.AssignPrivateIpAddresses(input)
					if err != nil {
						return fmt.Errorf("Failure to assign Private IPs: %s", err)
					}
				}
			}
		}
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

	if err := d.Set("private_ip_list", flattenNetworkInterfacesPrivateIPAddresses(eni.PrivateIpAddresses)); err != nil {
		return fmt.Errorf("error setting private_ip_list: %s", err)
	}

	if err := d.Set("security_groups", flattenGroupIdentifiers(eni.Groups)); err != nil {
		return fmt.Errorf("error setting security_groups: %s", err)
	}

	d.Set("source_dest_check", eni.SourceDestCheck)
	d.Set("subnet_id", eni.SubnetId)
	d.Set("ipv6_address_count", len(eni.Ipv6Addresses))

	if err := d.Set("ipv6_addresses", flattenEc2NetworkInterfaceIpv6Address(eni.Ipv6Addresses)); err != nil {
		return fmt.Errorf("error setting ipv6 addresses: %s", err)
	}

	if err := d.Set("ipv6_address_list", flattenEc2NetworkInterfaceIpv6Address(eni.Ipv6Addresses)); err != nil {
		return fmt.Errorf("error setting ipv6 address list: %s", err)
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
	private_ips_net_change := 0

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

	if d.HasChange("private_ips") && !d.Get("private_ip_list_enabled").(bool) {
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
			private_ips_net_change -= unassignIps.Len()
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
			private_ips_net_change += assignIps.Len()
		}
	}

	if d.HasChange("private_ip_list") && d.Get("private_ip_list_enabled").(bool) {
		o, n := d.GetChange("private_ip_list")
		if o == nil {
			o = make([]string, 0)
		}
		if n == nil {
			n = make([]string, 0)
		}
		if len(o.([]interface{}))-1 > 0 {
			privateIpsToUnassign := make([]interface{}, len(o.([]interface{}))-1)
			idx := 0
			for i, ip := range o.([]interface{}) {
				// skip primary private ip address
				if i == 0 {
					continue
				}
				privateIpsToUnassign[idx] = ip
				log.Printf("[INFO] Unassigning private ip %s", ip)
				idx += 1
			}

			// Unassign the secondary IP addresses
			input := &ec2.UnassignPrivateIpAddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				PrivateIpAddresses: expandStringList(privateIpsToUnassign),
			}
			_, err := conn.UnassignPrivateIpAddresses(input)
			if err != nil {
				return fmt.Errorf("Failure to unassign Private IPs: %s", err)
			}
		}

		// Assign each ip one-by-one in order to retain order
		for i, ip := range n.([]interface{}) {
			// skip primary private ip address
			if i == 0 {
				continue
			}
			privateIpToAssign := []interface{}{ip}
			log.Printf("[INFO] Assigning private ip %s", ip)

			input := &ec2.AssignPrivateIpAddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				PrivateIpAddresses: expandStringList(privateIpToAssign),
			}
			_, err := conn.AssignPrivateIpAddresses(input)
			if err != nil {
				return fmt.Errorf("Failure to assign Private IPs: %s", err)
			}
		}
	}

	if d.HasChange("ipv6_addresses") && !d.Get("ipv6_address_list_enabled").(bool) {
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

	if d.HasChange("ipv6_address_count") && !d.Get("ipv6_address_list_enabled").(bool) {
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

	if d.HasChange("ipv6_address_list") && d.Get("ipv6_address_list_enabled").(bool) {
		o, n := d.GetChange("ipv6_address_list")
		if o == nil {
			o = make([]string, 0)
		}
		if n == nil {
			n = make([]string, 0)
		}

		// Unassign old IPV6 addresses
		if len(o.([]interface{})) > 0 {
			unassignIps := make([]interface{}, len(o.([]interface{})))
			for i, ip := range o.([]interface{}) {
				unassignIps[i] = ip
				log.Printf("[INFO] Unassigning ipv6 address %s", ip)
			}

			log.Printf("[INFO] Unassigning ipv6 addresses")
			input := &ec2.UnassignIpv6AddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				Ipv6Addresses:      expandStringList(unassignIps),
			}
			_, err := conn.UnassignIpv6Addresses(input)
			if err != nil {
				return fmt.Errorf("failure to unassign IPV6 Addresses: %s", err)
			}
		}

		// Assign each ip one-by-one in order to retain order
		for _, ip := range n.([]interface{}) {
			privateIpToAssign := []interface{}{ip}
			log.Printf("[INFO] Assigning ipv6 address %s", ip)

			input := &ec2.AssignIpv6AddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				Ipv6Addresses:      expandStringList(privateIpToAssign),
			}
			_, err := conn.AssignIpv6Addresses(input)
			if err != nil {
				return fmt.Errorf("Failure to assign IPV6 Addresses: %s", err)
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

	if d.HasChange("private_ips_count") && !d.Get("private_ip_list_enabled").(bool) {
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

			diff := n.(int) - o.(int) - private_ips_net_change

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
