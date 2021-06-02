package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
)

func resourceAwsDefaultRouteTable() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDefaultRouteTableCreate,
		Read:   resourceAwsDefaultRouteTableRead,
		Update: resourceAwsRouteTableUpdate,
		Delete: resourceAwsDefaultRouteTableDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("vpc_id", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"default_route_table_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"propagating_vgws": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"route": {
				Type:       schema.TypeSet,
				ConfigMode: schema.SchemaConfigModeAttr,
				Computed:   true,
				Optional:   true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						///
						// Destinations.
						///
						"cidr_block": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.Any(
								validation.StringIsEmpty,
								validateIpv4CIDRNetworkAddress,
							),
						},

						"ipv6_cidr_block": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.Any(
								validation.StringIsEmpty,
								validateIpv6CIDRNetworkAddress,
							),
						},

						"destination_prefix_list_id": {
							Type:     schema.TypeString,
							Optional: true,
						},

						//
						// Targets.
						//
						"egress_only_gateway_id": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"gateway_id": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"instance_id": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"nat_gateway_id": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"network_interface_id": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"transit_gateway_id": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"vpc_endpoint_id": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"vpc_peering_connection_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
				Set: resourceAwsRouteTableHash,
			},

			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsDefaultRouteTableCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))
	routeTableID := d.Get("default_route_table_id").(string)

	routeTable, err := finder.RouteTableByID(conn, routeTableID)

	if err != nil {
		return fmt.Errorf("error reading EC2 Default Route Table (%s): %w", routeTableID, err)
	}

	d.SetId(routeTableID)
	d.Set("vpc_id", routeTable.VpcId)

	// revoke all default and pre-existing routes on the default route table.
	// In the UPDATE method, we'll apply only the rules in the configuration.
	log.Printf("[DEBUG] Revoking default routes for Default Route Table for %s", d.Id())
	if err := revokeAllRouteTableRules(conn, routeTable); err != nil {
		return err
	}

	if len(tags) > 0 {
		if err := keyvaluetags.Ec2CreateTags(conn, d.Id(), tags); err != nil {
			return fmt.Errorf("error adding tags: %w", err)
		}
	}

	return resourceAwsRouteTableUpdate(d, meta)
}

func resourceAwsDefaultRouteTableRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	// look up default route table for VPC
	filter1 := &ec2.Filter{
		Name:   aws.String("association.main"),
		Values: []*string{aws.String("true")},
	}
	filter2 := &ec2.Filter{
		Name:   aws.String("vpc-id"),
		Values: []*string{aws.String(d.Get("vpc_id").(string))},
	}

	findOpts := &ec2.DescribeRouteTablesInput{
		Filters: []*ec2.Filter{filter1, filter2},
	}

	resp, err := conn.DescribeRouteTables(findOpts)
	if err != nil {
		return err
	}

	if len(resp.RouteTables) < 1 || resp.RouteTables[0] == nil {
		log.Printf("[WARN] EC2 Default Route Table (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	rt := resp.RouteTables[0]

	d.Set("default_route_table_id", rt.RouteTableId)
	d.SetId(aws.StringValue(rt.RouteTableId))

	// re-use regular AWS Route Table READ. This is an extra API call but saves us
	// from trying to manually keep parity
	return resourceAwsRouteTableRead(d, meta)
}

func resourceAwsDefaultRouteTableDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] Cannot destroy Default Route Table. Terraform will remove this resource from the state file, however resources may remain.")
	return nil
}

// revokeAllRouteTableRules revoke all routes on the Default Route Table
// This should only be ran once at creation time of this resource
func revokeAllRouteTableRules(conn *ec2.EC2, routeTable *ec2.RouteTable) error {
	// Remove all Gateway association
	for _, r := range routeTable.PropagatingVgws {
		_, err := conn.DisableVgwRoutePropagation(&ec2.DisableVgwRoutePropagationInput{
			RouteTableId: routeTable.RouteTableId,
			GatewayId:    r.GatewayId,
		})

		if err != nil {
			return err
		}
	}

	// Delete all routes
	for _, r := range routeTable.Routes {
		// you cannot delete the local route
		if aws.StringValue(r.GatewayId) == "local" {
			continue
		}

		if aws.StringValue(r.Origin) == ec2.RouteOriginEnableVgwRoutePropagation {
			continue
		}

		if r.DestinationPrefixListId != nil && strings.HasPrefix(aws.StringValue(r.GatewayId), "vpce-") {
			// Skipping because VPC endpoint routes are handled separately
			// See aws_vpc_endpoint
			continue
		}

		if r.DestinationCidrBlock != nil {
			_, err := conn.DeleteRoute(&ec2.DeleteRouteInput{
				RouteTableId:         routeTable.RouteTableId,
				DestinationCidrBlock: r.DestinationCidrBlock,
			})

			if err != nil {
				return err
			}
		}

		if r.DestinationIpv6CidrBlock != nil {
			_, err := conn.DeleteRoute(&ec2.DeleteRouteInput{
				RouteTableId:             routeTable.RouteTableId,
				DestinationIpv6CidrBlock: r.DestinationIpv6CidrBlock,
			})

			if err != nil {
				return err
			}
		}

		if r.DestinationPrefixListId != nil {
			_, err := conn.DeleteRoute(&ec2.DeleteRouteInput{
				RouteTableId:            routeTable.RouteTableId,
				DestinationPrefixListId: r.DestinationPrefixListId,
			})

			if err != nil {
				return err
			}
		}
	}

	return nil
}
