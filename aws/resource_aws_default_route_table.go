package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/waiter"
)

func resourceAwsDefaultRouteTable() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDefaultRouteTableCreate,
		Read:   resourceAwsDefaultRouteTableRead,
		Update: resourceAwsRouteTableUpdate,
		Delete: resourceAwsDefaultRouteTableDelete,

		Importer: &schema.ResourceImporter{
			State: resourceAwsDefaultRouteTableImport,
		},

		//
		// The top-level attributes must be a superset of the aws_route_table resource's attributes as common CRUD handlers are used.
		//
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"default_route_table_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"owner_id": {
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
						"destination_prefix_list_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ipv6_cidr_block": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.Any(
								validation.StringIsEmpty,
								validateIpv6CIDRNetworkAddress,
							),
						},

						//
						// Targets.
						// These target attributes are a subset of the aws_route_table resource's target attributes
						// as there are some targets that are not allowed in the default route table for a VPC.
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

			"vpc_id": {
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

	d.SetId(aws.StringValue(routeTable.RouteTableId))

	// Remove all existing VGW associations.
	for _, v := range routeTable.PropagatingVgws {
		if err := ec2RouteTableDisableVgwRoutePropagation(conn, d.Id(), aws.StringValue(v.GatewayId)); err != nil {
			return err
		}
	}

	// Delete all existing routes.
	for _, v := range routeTable.Routes {
		// you cannot delete the local route
		if aws.StringValue(v.GatewayId) == "local" {
			continue
		}

		if aws.StringValue(v.Origin) == ec2.RouteOriginEnableVgwRoutePropagation {
			continue
		}

		if v.DestinationPrefixListId != nil && strings.HasPrefix(aws.StringValue(v.GatewayId), "vpce-") {
			// Skipping because VPC endpoint routes are handled separately
			// See aws_vpc_endpoint
			continue
		}

		input := &ec2.DeleteRouteInput{
			RouteTableId: aws.String(d.Id()),
		}

		var destination string
		var routeFinder finder.RouteFinder

		if v.DestinationCidrBlock != nil {
			input.DestinationCidrBlock = v.DestinationCidrBlock
			destination = aws.StringValue(v.DestinationCidrBlock)
			routeFinder = finder.RouteByIPv4Destination
		} else if v.DestinationIpv6CidrBlock != nil {
			input.DestinationIpv6CidrBlock = v.DestinationIpv6CidrBlock
			destination = aws.StringValue(v.DestinationIpv6CidrBlock)
			routeFinder = finder.RouteByIPv6Destination
		} else if v.DestinationPrefixListId != nil {
			input.DestinationPrefixListId = v.DestinationPrefixListId
			destination = aws.StringValue(v.DestinationPrefixListId)
			routeFinder = finder.RouteByPrefixListIDDestination
		}

		log.Printf("[DEBUG] Deleting Route: %s", input)
		_, err := conn.DeleteRoute(input)

		if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidRouteNotFound) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error deleting Route in EC2 Default Route Table (%s) with destination (%s): %w", d.Id(), destination, err)
		}

		_, err = waiter.RouteDeleted(conn, routeFinder, routeTableID, destination)

		if err != nil {
			return fmt.Errorf("error waiting for Route in EC2 Default Route Table (%s) with destination (%s) to delete: %w", d.Id(), destination, err)
		}
	}

	// Add new VGW associations.
	if v, ok := d.GetOk("propagating_vgws"); ok && v.(*schema.Set).Len() > 0 {
		for _, v := range v.(*schema.Set).List() {
			v := v.(string)

			if err := ec2RouteTableEnableVgwRoutePropagation(conn, d.Id(), v); err != nil {
				return err
			}
		}
	}

	// Add new routes.
	if v, ok := d.GetOk("route"); ok && v.(*schema.Set).Len() > 0 {
		for _, v := range v.(*schema.Set).List() {
			v := v.(map[string]interface{})

			if err := ec2RouteTableAddRoute(conn, d.Id(), v); err != nil {
				return err
			}
		}
	}

	if len(tags) > 0 {
		if err := keyvaluetags.Ec2CreateTags(conn, d.Id(), tags); err != nil {
			return fmt.Errorf("error adding tags: %w", err)
		}
	}

	return resourceAwsDefaultRouteTableRead(d, meta)
}

func resourceAwsDefaultRouteTableRead(d *schema.ResourceData, meta interface{}) error {
	d.Set("default_route_table_id", d.Id())

	// re-use regular AWS Route Table READ. This is an extra API call but saves us
	// from trying to manually keep parity
	return resourceAwsRouteTableRead(d, meta)
}

func resourceAwsDefaultRouteTableDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] Cannot destroy Default Route Table. Terraform will remove this resource from the state file, however resources may remain.")
	return nil
}

func resourceAwsDefaultRouteTableImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*AWSClient).ec2conn

	routeTable, err := finder.MainRouteTableByVpcID(conn, d.Id())

	if err != nil {
		return nil, err
	}

	d.SetId(aws.StringValue(routeTable.RouteTableId))

	return []*schema.ResourceData{d}, nil
}
