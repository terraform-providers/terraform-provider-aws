package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsDxPrivateVirtualInterface() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsDxPrivateVirtualInterfaceRead,

		Schema: map[string]*schema.Schema{
			"virtual_interface_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpn_gateway_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dx_gateway_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vlan": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"bgp_asn": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"bgp_auth_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"address_family": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"customer_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"amazon_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"aws_device": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"mtu": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"jumbo_frame_capable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsDxPrivateVirtualInterfaceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	vif, err := dxVirtualInterfaceRead(d.Get("virtual_interface_id").(string), conn)
	if err != nil {
		return err
	}
	if vif == nil {
		return fmt.Errorf("[ERR] Direct Connect virtual interface (%s) not found", d.Get("virtual_interface_id").(string))
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "directconnect",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("dxvif/%s", d.Get("virtual_interface_id")),
	}.String()

	d.SetId(aws.StringValue(vif.VirtualInterfaceId))
	d.Set("arn", arn)
	d.Set("connection_id", vif.ConnectionId)
	d.Set("name", vif.VirtualInterfaceName)
	d.Set("vlan", vif.Vlan)
	d.Set("bgp_asn", vif.Asn)
	d.Set("bgp_auth_key", vif.AuthKey)
	d.Set("address_family", vif.AddressFamily)
	d.Set("customer_address", vif.CustomerAddress)
	d.Set("amazon_address", vif.AmazonAddress)
	d.Set("aws_device", vif.AwsDeviceV2)
	d.Set("vpn_gateway_id", vif.VirtualGatewayId)
	d.Set("dx_gateway_id", vif.DirectConnectGatewayId)
	d.Set("mtu", vif.Mtu)
	d.Set("jumbo_frame_capable", vif.JumboFrameCapable)

	err1 := getTagsDX(conn, d, arn)

	return err1
}
