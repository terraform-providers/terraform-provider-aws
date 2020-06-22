package aws

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsDxHostedPublicVirtualInterface() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDxHostedPublicVirtualInterfaceCreate,
		Read:   resourceAwsDxHostedPublicVirtualInterfaceRead,
		Delete: resourceAwsDxHostedPublicVirtualInterfaceDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsDxHostedPublicVirtualInterfaceImport,
		},
		CustomizeDiff: resourceAwsDxHostedPublicVirtualInterfaceCustomizeDiff,

		Schema: map[string]*schema.Schema{
			"address_family": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					directconnect.AddressFamilyIpv4,
					directconnect.AddressFamilyIpv6,
				}, false),
			},
			"amazon_address": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"amazon_side_asn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"aws_device": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bgp_asn": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"bgp_auth_key": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"connection_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"customer_address": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"owner_account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsAccountId,
			},
			"route_filter_prefixes": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				MinItems: 1,
			},
			"vlan": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(1, 4094),
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceAwsDxHostedPublicVirtualInterfaceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	req := &directconnect.AllocatePublicVirtualInterfaceInput{
		ConnectionId: aws.String(d.Get("connection_id").(string)),
		NewPublicVirtualInterfaceAllocation: &directconnect.NewPublicVirtualInterfaceAllocation{
			AddressFamily:        aws.String(d.Get("address_family").(string)),
			Asn:                  aws.Int64(int64(d.Get("bgp_asn").(int))),
			VirtualInterfaceName: aws.String(d.Get("name").(string)),
			Vlan:                 aws.Int64(int64(d.Get("vlan").(int))),
		},
		OwnerAccount: aws.String(d.Get("owner_account_id").(string)),
	}
	if v, ok := d.GetOk("amazon_address"); ok && v.(string) != "" {
		req.NewPublicVirtualInterfaceAllocation.AmazonAddress = aws.String(v.(string))
	}
	if v, ok := d.GetOk("bgp_auth_key"); ok && v.(string) != "" {
		req.NewPublicVirtualInterfaceAllocation.AuthKey = aws.String(v.(string))
	}
	if v, ok := d.GetOk("customer_address"); ok && v.(string) != "" {
		req.NewPublicVirtualInterfaceAllocation.CustomerAddress = aws.String(v.(string))
	}
	if v, ok := d.GetOk("route_filter_prefixes"); ok {
		req.NewPublicVirtualInterfaceAllocation.RouteFilterPrefixes = expandDxRouteFilterPrefixes(v.(*schema.Set))
	}

	log.Printf("[DEBUG] Allocating Direct Connect hosted public virtual interface: %s", req)
	resp, err := conn.AllocatePublicVirtualInterface(req)
	if err != nil {
		return fmt.Errorf("error allocating Direct Connect hosted public virtual interface: %s", err)
	}

	d.SetId(aws.StringValue(resp.VirtualInterfaceId))

	if err := dxHostedPublicVirtualInterfaceWaitUntilAvailable(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return err
	}

	return resourceAwsDxHostedPublicVirtualInterfaceRead(d, meta)
}

func resourceAwsDxHostedPublicVirtualInterfaceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	vif, err := dxVirtualInterfaceRead(d.Id(), conn)
	if err != nil {
		return err
	}
	if vif == nil {
		log.Printf("[WARN] Direct Connect virtual interface (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("address_family", vif.AddressFamily)
	d.Set("amazon_address", vif.AmazonAddress)
	d.Set("amazon_side_asn", strconv.FormatInt(aws.Int64Value(vif.AmazonSideAsn), 10))
	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "directconnect",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("dxvif/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("aws_device", vif.AwsDeviceV2)
	d.Set("bgp_asn", vif.Asn)
	d.Set("bgp_auth_key", vif.AuthKey)
	d.Set("connection_id", vif.ConnectionId)
	d.Set("customer_address", vif.CustomerAddress)
	d.Set("name", vif.VirtualInterfaceName)
	d.Set("owner_account_id", vif.OwnerAccount)
	if err := d.Set("route_filter_prefixes", flattenDxRouteFilterPrefixes(vif.RouteFilterPrefixes)); err != nil {
		return fmt.Errorf("error setting route_filter_prefixes: %s", err)
	}
	d.Set("vlan", vif.Vlan)

	return nil
}

func resourceAwsDxHostedPublicVirtualInterfaceDelete(d *schema.ResourceData, meta interface{}) error {
	return dxVirtualInterfaceDelete(d, meta)
}

func resourceAwsDxHostedPublicVirtualInterfaceImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*AWSClient).dxconn

	vif, err := dxVirtualInterfaceRead(d.Id(), conn)
	if err != nil {
		return nil, err
	}
	if vif == nil {
		return nil, fmt.Errorf("virtual interface (%s) not found", d.Id())
	}

	if vifType := aws.StringValue(vif.VirtualInterfaceType); vifType != "public" {
		return nil, fmt.Errorf("virtual interface (%s) has incorrect type: %s", d.Id(), vifType)
	}

	return []*schema.ResourceData{d}, nil
}

func resourceAwsDxHostedPublicVirtualInterfaceCustomizeDiff(diff *schema.ResourceDiff, meta interface{}) error {
	if diff.Id() == "" {
		// New resource.
		if addressFamily := diff.Get("address_family").(string); addressFamily == directconnect.AddressFamilyIpv4 {
			if _, ok := diff.GetOk("customer_address"); !ok {
				return fmt.Errorf("'customer_address' must be set when 'address_family' is '%s'", addressFamily)
			}
			if _, ok := diff.GetOk("amazon_address"); !ok {
				return fmt.Errorf("'amazon_address' must be set when 'address_family' is '%s'", addressFamily)
			}
		}
	}

	return nil
}

func dxHostedPublicVirtualInterfaceWaitUntilAvailable(conn *directconnect.DirectConnect, vifId string, timeout time.Duration) error {
	return dxVirtualInterfaceWaitUntilAvailable(
		conn,
		vifId,
		timeout,
		[]string{
			directconnect.VirtualInterfaceStatePending,
		},
		[]string{
			directconnect.VirtualInterfaceStateAvailable,
			directconnect.VirtualInterfaceStateConfirming,
			directconnect.VirtualInterfaceStateDown,
			directconnect.VirtualInterfaceStateVerifying,
		})
}
