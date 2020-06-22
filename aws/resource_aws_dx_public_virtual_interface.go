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
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsDxPublicVirtualInterface() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDxPublicVirtualInterfaceCreate,
		Read:   resourceAwsDxPublicVirtualInterfaceRead,
		Update: resourceAwsDxPublicVirtualInterfaceUpdate,
		Delete: resourceAwsDxPublicVirtualInterfaceDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsDxPublicVirtualInterfaceImport,
		},
		CustomizeDiff: resourceAwsDxPublicVirtualInterfaceCustomizeDiff,

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
			"route_filter_prefixes": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				MinItems: 1,
			},
			"tags": tagsSchema(),
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

func resourceAwsDxPublicVirtualInterfaceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	req := &directconnect.CreatePublicVirtualInterfaceInput{
		ConnectionId: aws.String(d.Get("connection_id").(string)),
		NewPublicVirtualInterface: &directconnect.NewPublicVirtualInterface{
			AddressFamily:        aws.String(d.Get("address_family").(string)),
			Asn:                  aws.Int64(int64(d.Get("bgp_asn").(int))),
			VirtualInterfaceName: aws.String(d.Get("name").(string)),
			Vlan:                 aws.Int64(int64(d.Get("vlan").(int))),
		},
	}
	if v, ok := d.GetOk("amazon_address"); ok && v.(string) != "" {
		req.NewPublicVirtualInterface.AmazonAddress = aws.String(v.(string))
	}
	if v, ok := d.GetOk("bgp_auth_key"); ok && v.(string) != "" {
		req.NewPublicVirtualInterface.AuthKey = aws.String(v.(string))
	}
	if v, ok := d.GetOk("customer_address"); ok && v.(string) != "" {
		req.NewPublicVirtualInterface.CustomerAddress = aws.String(v.(string))
	}
	if v, ok := d.GetOk("route_filter_prefixes"); ok {
		req.NewPublicVirtualInterface.RouteFilterPrefixes = expandDxRouteFilterPrefixes(v.(*schema.Set))
	}
	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		req.NewPublicVirtualInterface.Tags = keyvaluetags.New(v).IgnoreAws().DirectconnectTags()
	}

	log.Printf("[DEBUG] Creating Direct Connect public virtual interface: %s", req)
	resp, err := conn.CreatePublicVirtualInterface(req)
	if err != nil {
		return fmt.Errorf("error creating Direct Connect public virtual interface: %s", err)
	}

	d.SetId(aws.StringValue(resp.VirtualInterfaceId))

	if err := dxPublicVirtualInterfaceWaitUntilAvailable(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return err
	}

	return resourceAwsDxPublicVirtualInterfaceRead(d, meta)
}

func resourceAwsDxPublicVirtualInterfaceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

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
	d.Set("customer_address", vif.CustomerAddress)
	d.Set("connection_id", vif.ConnectionId)
	d.Set("name", vif.VirtualInterfaceName)
	if err := d.Set("route_filter_prefixes", flattenDxRouteFilterPrefixes(vif.RouteFilterPrefixes)); err != nil {
		return fmt.Errorf("error setting route_filter_prefixes: %s", err)
	}
	d.Set("vlan", vif.Vlan)

	tags, err := keyvaluetags.DirectconnectListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Direct Connect public virtual interface (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsDxPublicVirtualInterfaceUpdate(d *schema.ResourceData, meta interface{}) error {
	if err := dxVirtualInterfaceUpdate(d, meta); err != nil {
		return err
	}

	return resourceAwsDxPublicVirtualInterfaceRead(d, meta)
}

func resourceAwsDxPublicVirtualInterfaceDelete(d *schema.ResourceData, meta interface{}) error {
	return dxVirtualInterfaceDelete(d, meta)
}

func resourceAwsDxPublicVirtualInterfaceImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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

func resourceAwsDxPublicVirtualInterfaceCustomizeDiff(diff *schema.ResourceDiff, meta interface{}) error {
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

func dxPublicVirtualInterfaceWaitUntilAvailable(conn *directconnect.DirectConnect, vifId string, timeout time.Duration) error {
	return dxVirtualInterfaceWaitUntilAvailable(
		conn,
		vifId,
		timeout,
		[]string{
			directconnect.VirtualInterfaceStatePending,
		},
		[]string{
			directconnect.VirtualInterfaceStateAvailable,
			directconnect.VirtualInterfaceStateDown,
			directconnect.VirtualInterfaceStateVerifying,
		})
}
