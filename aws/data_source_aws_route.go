package aws

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
)

func dataSourceAwsRoute() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsRouteRead,

		Schema: map[string]*schema.Schema{
			"route_table_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			///
			// Destinations.
			///
			"destination_cidr_block": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"destination_ipv6_cidr_block": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"destination_prefix_list_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			//
			// Targets.
			//
			"carrier_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"egress_only_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"instance_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"local_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"nat_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"network_interface_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"transit_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"vpc_peering_connection_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsRouteRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	routeTableID := d.Get("route_table_id").(string)

	routeTable, err := finder.RouteTableByID(conn, routeTableID)

	if err != nil {
		return fmt.Errorf("error reading Route Table (%s): %w", routeTableID, err)
	}

	routes := []*ec2.Route{}

	for _, r := range routeTable.Routes {
		if aws.StringValue(r.Origin) == ec2.RouteOriginEnableVgwRoutePropagation {
			continue
		}

		if r.DestinationPrefixListId != nil && strings.HasPrefix(aws.StringValue(r.GatewayId), "vpce-") {
			// Skipping because VPC endpoint routes are handled separately
			// See aws_vpc_endpoint
			continue
		}

		if v, ok := d.GetOk("destination_cidr_block"); ok && aws.StringValue(r.DestinationCidrBlock) != v.(string) {
			continue
		}

		if v, ok := d.GetOk("destination_ipv6_cidr_block"); ok && aws.StringValue(r.DestinationIpv6CidrBlock) != v.(string) {
			continue
		}

		if v, ok := d.GetOk("destination_prefix_list_id"); ok && aws.StringValue(r.DestinationPrefixListId) != v.(string) {
			continue
		}

		if v, ok := d.GetOk("carrier_gateway_id"); ok && aws.StringValue(r.CarrierGatewayId) != v.(string) {
			continue
		}

		if v, ok := d.GetOk("egress_only_gateway_id"); ok && aws.StringValue(r.EgressOnlyInternetGatewayId) != v.(string) {
			continue
		}

		if v, ok := d.GetOk("gateway_id"); ok && aws.StringValue(r.GatewayId) != v.(string) {
			continue
		}

		if v, ok := d.GetOk("instance_id"); ok && aws.StringValue(r.InstanceId) != v.(string) {
			continue
		}

		if v, ok := d.GetOk("local_gateway_id"); ok && aws.StringValue(r.LocalGatewayId) != v.(string) {
			continue
		}

		if v, ok := d.GetOk("nat_gateway_id"); ok && aws.StringValue(r.NatGatewayId) != v.(string) {
			continue
		}

		if v, ok := d.GetOk("network_interface_id"); ok && aws.StringValue(r.NetworkInterfaceId) != v.(string) {
			continue
		}

		if v, ok := d.GetOk("transit_gateway_id"); ok && aws.StringValue(r.TransitGatewayId) != v.(string) {
			continue
		}

		if v, ok := d.GetOk("vpc_peering_connection_id"); ok && aws.StringValue(r.VpcPeeringConnectionId) != v.(string) {
			continue
		}

		routes = append(routes, r)
	}

	if len(routes) == 0 {
		return fmt.Errorf("No routes matching supplied arguments found in Route Table (%s)", routeTableID)
	}

	if len(routes) > 1 {
		return fmt.Errorf("%d routes matched in Route Table (%s); use additional constraints to reduce matches to a single route", len(routes), routeTableID)
	}

	route := routes[0]

	if destination := aws.StringValue(route.DestinationCidrBlock); destination != "" {
		d.SetId(tfec2.RouteCreateID(routeTableID, destination))
	} else if destination := aws.StringValue(route.DestinationIpv6CidrBlock); destination != "" {
		d.SetId(tfec2.RouteCreateID(routeTableID, destination))
	} else if destination := aws.StringValue(route.DestinationPrefixListId); destination != "" {
		d.SetId(tfec2.RouteCreateID(routeTableID, destination))
	}

	d.Set("carrier_gateway_id", route.CarrierGatewayId)
	d.Set("destination_cidr_block", route.DestinationCidrBlock)
	d.Set("destination_ipv6_cidr_block", route.DestinationIpv6CidrBlock)
	d.Set("destination_prefix_list_id", route.DestinationPrefixListId)
	d.Set("egress_only_gateway_id", route.EgressOnlyInternetGatewayId)
	d.Set("gateway_id", route.GatewayId)
	d.Set("instance_id", route.InstanceId)
	d.Set("local_gateway_id", route.LocalGatewayId)
	d.Set("nat_gateway_id", route.NatGatewayId)
	d.Set("network_interface_id", route.NetworkInterfaceId)
	d.Set("transit_gateway_id", route.TransitGatewayId)
	d.Set("vpc_peering_connection_id", route.VpcPeeringConnectionId)

	return nil
}
