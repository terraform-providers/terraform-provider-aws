package aws

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/hashcode"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	tfnet "github.com/terraform-providers/terraform-provider-aws/aws/internal/net"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

var routeTableValidDestinations = []string{
	"cidr_block",
	"ipv6_cidr_block",
	"destination_prefix_list_id",
}

var routeTableValidTargets = []string{
	"carrier_gateway_id",
	"egress_only_gateway_id",
	"gateway_id",
	"instance_id",
	"local_gateway_id",
	"nat_gateway_id",
	"network_interface_id",
	"transit_gateway_id",
	"vpc_endpoint_id",
	"vpc_peering_connection_id",
}

func resourceAwsRouteTable() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRouteTableCreate,
		Read:   resourceAwsRouteTableRead,
		Update: resourceAwsRouteTableUpdate,
		Delete: resourceAwsRouteTableDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"propagating_vgws": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"route": {
				Type:       schema.TypeSet,
				Computed:   true,
				Optional:   true,
				ConfigMode: schema.SchemaConfigModeAttr,
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
						//
						"carrier_gateway_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
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
						"local_gateway_id": {
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
				Required: true,
				ForceNew: true,
			},
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsRouteTableCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateRouteTableInput{
		VpcId:             aws.String(d.Get("vpc_id").(string)),
		TagSpecifications: ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeRouteTable),
	}

	log.Printf("[DEBUG] Creating Route Table: %s", input)
	output, err := conn.CreateRouteTable(input)

	if err != nil {
		return fmt.Errorf("error creating Route Table: %w", err)
	}

	d.SetId(aws.StringValue(output.RouteTable.RouteTableId))

	if _, err := waiter.RouteTableReady(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Route Table (%s) to become available: %w", d.Id(), err)
	}

	if v, ok := d.GetOk("propagating_vgws"); ok && v.(*schema.Set).Len() > 0 {
		for _, v := range v.(*schema.Set).List() {
			v := v.(string)

			if err := ec2RouteTableEnableVgwRoutePropagation(conn, d.Id(), v); err != nil {
				return err
			}
		}
	}

	if v, ok := d.GetOk("route"); ok && v.(*schema.Set).Len() > 0 {
		for _, v := range v.(*schema.Set).List() {
			v := v.(map[string]interface{})

			if err := ec2RouteTableAddRoute(conn, d.Id(), v); err != nil {
				return err
			}
		}
	}

	return resourceAwsRouteTableRead(d, meta)
}

func resourceAwsRouteTableRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	routeTable, err := finder.RouteTableByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route Table (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Route Table (%s): %w", d.Id(), err)
	}

	d.Set("vpc_id", routeTable.VpcId)

	propagatingVGWs := make([]string, 0, len(routeTable.PropagatingVgws))
	for _, v := range routeTable.PropagatingVgws {
		propagatingVGWs = append(propagatingVGWs, aws.StringValue(v.GatewayId))
	}
	if err := d.Set("propagating_vgws", propagatingVGWs); err != nil {
		return fmt.Errorf("error setting propagating_vgws: %w", err)
	}

	if err := d.Set("route", flattenEc2Routes(routeTable.Routes)); err != nil {
		return fmt.Errorf("error setting route: %w", err)
	}

	tags := keyvaluetags.Ec2KeyValueTags(routeTable.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	ownerID := aws.StringValue(routeTable.OwnerId)
	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*AWSClient).region,
		AccountID: ownerID,
		Resource:  fmt.Sprintf("route-table/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("owner_id", ownerID)

	return nil
}

func resourceAwsRouteTableUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	if d.HasChange("propagating_vgws") {
		o, n := d.GetChange("propagating_vgws")
		os := o.(*schema.Set)
		ns := n.(*schema.Set)
		del := os.Difference(ns).List()
		add := ns.Difference(os).List()

		for _, v := range del {
			v := v.(string)

			if err := ec2RouteTableDisableVgwRoutePropagation(conn, d.Id(), v); err != nil {
				return err
			}
		}

		for _, v := range add {
			v := v.(string)

			if err := ec2RouteTableEnableVgwRoutePropagation(conn, d.Id(), v); err != nil {
				return err
			}
		}
	}

	if d.HasChange("route") {
		o, n := d.GetChange("route")

		for _, new := range n.(*schema.Set).List() {
			vNew := new.(map[string]interface{})

			_, newDestination := routeTableRouteDestinationAttribute(vNew)
			_, newTarget := routeTableRouteTargetAttribute(vNew)

			addRoute := true

			for _, old := range o.(*schema.Set).List() {
				vOld := old.(map[string]interface{})

				_, oldDestination := routeTableRouteDestinationAttribute(vOld)
				_, oldTarget := routeTableRouteTargetAttribute(vOld)

				if oldDestination == newDestination {
					addRoute = false

					if oldTarget != newTarget {
						if err := ec2RouteTableUpdateRoute(conn, d.Id(), vNew); err != nil {
							return err
						}
					}
				}
			}

			if addRoute {
				if err := ec2RouteTableAddRoute(conn, d.Id(), vNew); err != nil {
					return err
				}
			}
		}

		for _, old := range o.(*schema.Set).List() {
			vOld := old.(map[string]interface{})

			_, oldDestination := routeTableRouteDestinationAttribute(vOld)

			delRoute := true

			for _, new := range n.(*schema.Set).List() {
				vNew := new.(map[string]interface{})

				_, newDestination := routeTableRouteDestinationAttribute(vNew)

				if newDestination == oldDestination {
					delRoute = false
				}
			}

			if delRoute {
				if err := ec2RouteTableDeleteRoute(conn, d.Id(), vOld); err != nil {
					return err
				}
			}
		}
	}

	if d.HasChange("tags_all") && !d.IsNewResource() {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.Ec2UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Route Table (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAwsRouteTableRead(d, meta)
}

func resourceAwsRouteTableDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	routeTable, err := finder.RouteTableByID(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error reading Route Table (%s): %w", d.Id(), err)
	}

	// Do all the disassociations
	for _, v := range routeTable.Associations {
		v := aws.StringValue(v.RouteTableAssociationId)

		if err := ec2RouteTableAssociationDelete(conn, v); err != nil {
			return err
		}
	}

	log.Printf("[INFO] Deleting Route Table: %s", d.Id())
	_, err = conn.DeleteRouteTable(&ec2.DeleteRouteTableInput{
		RouteTableId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidRouteTableIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Route Table (%s): %w", d.Id(), err)
	}

	// Wait for the route table to really destroy
	log.Printf("[DEBUG] Waiting for route table (%s) deletion", d.Id())
	if _, err := waiter.RouteTableDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Route Table (%s) deletion: %w", d.Id(), err)
	}

	return nil
}

func resourceAwsRouteTableHash(v interface{}) int {
	var buf bytes.Buffer
	m, castOk := v.(map[string]interface{})
	if !castOk {
		return 0
	}

	if v, ok := m["ipv6_cidr_block"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", tfnet.CanonicalCIDRBlock(v.(string))))
	}

	if v, ok := m["cidr_block"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["destination_prefix_list_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["gateway_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["carrier_gateway_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["egress_only_gateway_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	natGatewaySet := false
	if v, ok := m["nat_gateway_id"]; ok {
		natGatewaySet = v.(string) != ""
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	instanceSet := false
	if v, ok := m["instance_id"]; ok {
		instanceSet = v.(string) != ""
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["transit_gateway_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["local_gateway_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["vpc_endpoint_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["vpc_peering_connection_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["network_interface_id"]; ok && !(instanceSet || natGatewaySet) {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	return hashcode.String(buf.String())
}

// ec2RouteTableAddRoute adds a route to the specified route table.
func ec2RouteTableAddRoute(conn *ec2.EC2, routeTableID string, tfMap map[string]interface{}) error {
	if err := validateNestedExactlyOneOf(tfMap, routeTableValidDestinations); err != nil {
		return fmt.Errorf("error creating route: %w", err)
	}
	if err := validateNestedExactlyOneOf(tfMap, routeTableValidTargets); err != nil {
		return fmt.Errorf("error creating route: %w", err)
	}

	destinationAttributeKey, destination := routeTableRouteDestinationAttribute(tfMap)

	var routeFinder finder.RouteFinder

	switch destinationAttributeKey {
	case "cidr_block":
		routeFinder = finder.RouteByIPv4Destination
	case "ipv6_cidr_block":
		routeFinder = finder.RouteByIPv6Destination
	case "destination_prefix_list_id":
		routeFinder = finder.RouteByPrefixListIDDestination
	default:
		return fmt.Errorf("error creating Route: unexpected route destination attribute: %q", destinationAttributeKey)
	}

	input := expandEc2CreateRouteInput(tfMap)

	if input == nil {
		return nil
	}

	input.RouteTableId = aws.String(routeTableID)

	log.Printf("[DEBUG] Creating Route: %s", input)
	_, err := tfresource.RetryWhenAwsErrCodeEquals(
		waiter.PropagationTimeout,
		func() (interface{}, error) {
			return conn.CreateRoute(input)
		},
		tfec2.ErrCodeInvalidParameterException,
		tfec2.ErrCodeInvalidTransitGatewayIDNotFound,
	)

	if err != nil {
		return fmt.Errorf("error creating Route in Route Table (%s) with destination (%s): %w", routeTableID, destination, err)
	}

	_, err = waiter.RouteReady(conn, routeFinder, routeTableID, destination)

	if err != nil {
		return fmt.Errorf("error waiting for Route in Route Table (%s) with destination (%s) to become available: %w", routeTableID, destination, err)
	}

	return nil
}

// ec2RouteTableDeleteRoute deletes a route from the specified route table.
func ec2RouteTableDeleteRoute(conn *ec2.EC2, routeTableID string, tfMap map[string]interface{}) error {
	destinationAttributeKey, destination := routeTableRouteDestinationAttribute(tfMap)

	input := &ec2.DeleteRouteInput{
		RouteTableId: aws.String(routeTableID),
	}

	var routeFinder finder.RouteFinder

	switch destination := aws.String(destination); destinationAttributeKey {
	case "cidr_block":
		input.DestinationCidrBlock = destination
		routeFinder = finder.RouteByIPv4Destination
	case "ipv6_cidr_block":
		input.DestinationIpv6CidrBlock = destination
		routeFinder = finder.RouteByIPv6Destination
	case "destination_prefix_list_id":
		input.DestinationPrefixListId = destination
		routeFinder = finder.RouteByPrefixListIDDestination
	default:
		return fmt.Errorf("error deleting Route: unexpected route destination attribute: %q", destinationAttributeKey)
	}

	log.Printf("[DEBUG] Deleting Route: %s", input)
	_, err := conn.DeleteRoute(input)

	if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidRouteNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Route in Route Table (%s) with destination (%s): %w", routeTableID, destination, err)
	}

	_, err = waiter.RouteDeleted(conn, routeFinder, routeTableID, destination)

	if err != nil {
		return fmt.Errorf("error waiting for Route in Route Table (%s) with destination (%s) to delete: %w", routeTableID, destination, err)
	}

	return nil
}

// ec2RouteTableUpdateRoute updates a route in the specified route table.
func ec2RouteTableUpdateRoute(conn *ec2.EC2, routeTableID string, tfMap map[string]interface{}) error {
	if err := validateNestedExactlyOneOf(tfMap, routeTableValidDestinations); err != nil {
		return fmt.Errorf("error updating route: %w", err)
	}
	if err := validateNestedExactlyOneOf(tfMap, routeTableValidTargets); err != nil {
		return fmt.Errorf("error updating route: %w", err)
	}

	destinationAttributeKey, destination := routeTableRouteDestinationAttribute(tfMap)

	var routeFinder finder.RouteFinder

	switch destinationAttributeKey {
	case "cidr_block":
		routeFinder = finder.RouteByIPv4Destination
	case "ipv6_cidr_block":
		routeFinder = finder.RouteByIPv6Destination
	case "destination_prefix_list_id":
		routeFinder = finder.RouteByPrefixListIDDestination
	default:
		return fmt.Errorf("error creating Route: unexpected route destination attribute: %q", destinationAttributeKey)
	}

	input := expandEc2ReplaceRouteInput(tfMap)

	if input == nil {
		return nil
	}

	input.RouteTableId = aws.String(routeTableID)

	log.Printf("[DEBUG] Updating Route: %s", input)
	_, err := conn.ReplaceRoute(input)

	if err != nil {
		return fmt.Errorf("error updating Route in Route Table (%s) with destination (%s): %w", routeTableID, destination, err)
	}

	_, err = waiter.RouteReady(conn, routeFinder, routeTableID, destination)

	if err != nil {
		return fmt.Errorf("error waiting for Route in Route Table (%s) with destination (%s) to become available: %w", routeTableID, destination, err)
	}

	return nil
}

// ec2RouteTableDisableVgwRoutePropagation attempts to disable VGW route propagation.
// Any error is returned.
func ec2RouteTableDisableVgwRoutePropagation(conn *ec2.EC2, routeTableID, gatewayID string) error {
	input := &ec2.DisableVgwRoutePropagationInput{
		GatewayId:    aws.String(gatewayID),
		RouteTableId: aws.String(routeTableID),
	}

	log.Printf("[DEBUG] Disabling Route Table (%s) VPN Gateway (%s) route propagation", routeTableID, gatewayID)
	_, err := conn.DisableVgwRoutePropagation(input)

	if err != nil {
		return fmt.Errorf("error disabling Route Table (%s) VPN Gateway (%s) route propagation: %w", routeTableID, gatewayID, err)
	}

	return nil
}

// ec2RouteTableEnableVgwRoutePropagation attempts to enable VGW route propagation.
// The specified eventual consistency timeout is respected.
// Any error is returned.
func ec2RouteTableEnableVgwRoutePropagation(conn *ec2.EC2, routeTableID, gatewayID string) error {
	input := &ec2.EnableVgwRoutePropagationInput{
		GatewayId:    aws.String(gatewayID),
		RouteTableId: aws.String(routeTableID),
	}

	log.Printf("[DEBUG] Enabling Route Table (%s) VPN Gateway (%s) route propagation", routeTableID, gatewayID)
	_, err := tfresource.RetryWhenAwsErrCodeEquals(
		waiter.PropagationTimeout,
		func() (interface{}, error) {
			return conn.EnableVgwRoutePropagation(input)
		},
		tfec2.ErrCodeGatewayNotAttached,
	)

	if err != nil {
		return fmt.Errorf("error enabling Route Table (%s) VPN Gateway (%s) route propagation: %w", routeTableID, gatewayID, err)
	}

	return nil
}

func expandEc2CreateRouteInput(tfMap map[string]interface{}) *ec2.CreateRouteInput {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.CreateRouteInput{}

	if v, ok := tfMap["cidr_block"].(string); ok && v != "" {
		apiObject.DestinationCidrBlock = aws.String(v)
	}

	if v, ok := tfMap["ipv6_cidr_block"].(string); ok && v != "" {
		apiObject.DestinationIpv6CidrBlock = aws.String(v)
	}

	if v, ok := tfMap["destination_prefix_list_id"].(string); ok && v != "" {
		apiObject.DestinationPrefixListId = aws.String(v)
	}

	if v, ok := tfMap["carrier_gateway_id"].(string); ok && v != "" {
		apiObject.CarrierGatewayId = aws.String(v)
	}

	if v, ok := tfMap["egress_only_gateway_id"].(string); ok && v != "" {
		apiObject.EgressOnlyInternetGatewayId = aws.String(v)
	}

	if v, ok := tfMap["gateway_id"].(string); ok && v != "" {
		apiObject.GatewayId = aws.String(v)
	}

	if v, ok := tfMap["instance_id"].(string); ok && v != "" {
		apiObject.InstanceId = aws.String(v)
	}

	if v, ok := tfMap["local_gateway_id"].(string); ok && v != "" {
		apiObject.LocalGatewayId = aws.String(v)
	}

	if v, ok := tfMap["nat_gateway_id"].(string); ok && v != "" {
		apiObject.NatGatewayId = aws.String(v)
	}

	if v, ok := tfMap["network_interface_id"].(string); ok && v != "" {
		apiObject.NetworkInterfaceId = aws.String(v)
	}

	if v, ok := tfMap["transit_gateway_id"].(string); ok && v != "" {
		apiObject.TransitGatewayId = aws.String(v)
	}

	if v, ok := tfMap["vpc_endpoint_id"].(string); ok && v != "" {
		apiObject.VpcEndpointId = aws.String(v)
	}

	if v, ok := tfMap["vpc_peering_connection_id"].(string); ok && v != "" {
		apiObject.VpcPeeringConnectionId = aws.String(v)
	}

	return apiObject
}

func expandEc2ReplaceRouteInput(tfMap map[string]interface{}) *ec2.ReplaceRouteInput {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.ReplaceRouteInput{}

	if v, ok := tfMap["cidr_block"].(string); ok && v != "" {
		apiObject.DestinationCidrBlock = aws.String(v)
	}

	if v, ok := tfMap["ipv6_cidr_block"].(string); ok && v != "" {
		apiObject.DestinationIpv6CidrBlock = aws.String(v)
	}

	if v, ok := tfMap["destination_prefix_list_id"].(string); ok && v != "" {
		apiObject.DestinationPrefixListId = aws.String(v)
	}

	if v, ok := tfMap["carrier_gateway_id"].(string); ok && v != "" {
		apiObject.CarrierGatewayId = aws.String(v)
	}

	if v, ok := tfMap["egress_only_gateway_id"].(string); ok && v != "" {
		apiObject.EgressOnlyInternetGatewayId = aws.String(v)
	}

	if v, ok := tfMap["gateway_id"].(string); ok && v != "" {
		apiObject.GatewayId = aws.String(v)
	}

	if v, ok := tfMap["instance_id"].(string); ok && v != "" {
		apiObject.InstanceId = aws.String(v)
	}

	if v, ok := tfMap["local_gateway_id"].(string); ok && v != "" {
		apiObject.LocalGatewayId = aws.String(v)
	}

	if v, ok := tfMap["nat_gateway_id"].(string); ok && v != "" {
		apiObject.NatGatewayId = aws.String(v)
	}

	if v, ok := tfMap["network_interface_id"].(string); ok && v != "" {
		apiObject.NetworkInterfaceId = aws.String(v)
	}

	if v, ok := tfMap["transit_gateway_id"].(string); ok && v != "" {
		apiObject.TransitGatewayId = aws.String(v)
	}

	if v, ok := tfMap["vpc_endpoint_id"].(string); ok && v != "" {
		apiObject.VpcEndpointId = aws.String(v)
	}

	if v, ok := tfMap["vpc_peering_connection_id"].(string); ok && v != "" {
		apiObject.VpcPeeringConnectionId = aws.String(v)
	}

	return apiObject
}

func flattenEc2Route(apiObject *ec2.Route) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DestinationCidrBlock; v != nil {
		tfMap["cidr_block"] = aws.StringValue(v)
	}

	if v := apiObject.DestinationIpv6CidrBlock; v != nil {
		tfMap["ipv6_cidr_block"] = aws.StringValue(v)
	}

	if v := apiObject.DestinationPrefixListId; v != nil {
		tfMap["destination_prefix_list_id"] = aws.StringValue(v)
	}

	if v := apiObject.CarrierGatewayId; v != nil {
		tfMap["carrier_gateway_id"] = aws.StringValue(v)
	}

	if v := apiObject.EgressOnlyInternetGatewayId; v != nil {
		tfMap["egress_only_gateway_id"] = aws.StringValue(v)
	}

	if v := apiObject.GatewayId; v != nil {
		if strings.HasPrefix(aws.StringValue(v), "vpce-") {
			tfMap["vpc_endpoint_id"] = aws.StringValue(v)
		} else {
			tfMap["gateway_id"] = aws.StringValue(v)
		}
	}

	if v := apiObject.InstanceId; v != nil {
		tfMap["instance_id"] = aws.StringValue(v)
	}

	if v := apiObject.LocalGatewayId; v != nil {
		tfMap["local_gateway_id"] = aws.StringValue(v)
	}

	if v := apiObject.NatGatewayId; v != nil {
		tfMap["nat_gateway_id"] = aws.StringValue(v)
	}

	if v := apiObject.NetworkInterfaceId; v != nil {
		tfMap["network_interface_id"] = aws.StringValue(v)
	}

	if v := apiObject.TransitGatewayId; v != nil {
		tfMap["transit_gateway_id"] = aws.StringValue(v)
	}

	if v := apiObject.VpcPeeringConnectionId; v != nil {
		tfMap["vpc_peering_connection_id"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenEc2Routes(apiObjects []*ec2.Route) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		if aws.StringValue(apiObject.GatewayId) == "local" {
			continue
		}

		if aws.StringValue(apiObject.Origin) == ec2.RouteOriginEnableVgwRoutePropagation {
			continue
		}

		if apiObject.DestinationPrefixListId != nil && strings.HasPrefix(aws.StringValue(apiObject.GatewayId), "vpce-") {
			// Skipping because VPC endpoint routes are handled separately
			// See aws_vpc_endpoint
			continue
		}

		tfList = append(tfList, flattenEc2Route(apiObject))
	}

	return tfList
}

// routeTableRouteDestinationAttribute returns the attribute key and value of the route table route's destination.
func routeTableRouteDestinationAttribute(m map[string]interface{}) (string, string) {
	for _, key := range routeTableValidDestinations {
		if v, ok := m[key].(string); ok && v != "" {
			return key, v
		}
	}

	return "", ""
}

// routeTableRouteTargetAttribute returns the attribute key and value of the route table route's target.
func routeTableRouteTargetAttribute(m map[string]interface{}) (string, string) { //nolint:unparam
	for _, key := range routeTableValidTargets {
		if v, ok := m[key].(string); ok && v != "" {
			return key, v
		}
	}

	return "", ""
}
