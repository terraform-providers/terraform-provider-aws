package finder

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	tfnet "github.com/terraform-providers/terraform-provider-aws/aws/internal/net"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
)

// CarrierGatewayByID returns the carrier gateway corresponding to the specified identifier.
// Returns nil and potentially an error if no carrier gateway is found.
func CarrierGatewayByID(conn *ec2.EC2, id string) (*ec2.CarrierGateway, error) {
	input := &ec2.DescribeCarrierGatewaysInput{
		CarrierGatewayIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeCarrierGateways(input)
	if err != nil {
		return nil, err
	}

	if output == nil || len(output.CarrierGateways) == 0 {
		return nil, nil
	}

	return output.CarrierGateways[0], nil
}

func ClientVpnAuthorizationRule(conn *ec2.EC2, endpointID, targetNetworkCidr, accessGroupID string) (*ec2.DescribeClientVpnAuthorizationRulesOutput, error) {
	filters := map[string]string{
		"destination-cidr": targetNetworkCidr,
	}
	if accessGroupID != "" {
		filters["group-id"] = accessGroupID
	}

	input := &ec2.DescribeClientVpnAuthorizationRulesInput{
		ClientVpnEndpointId: aws.String(endpointID),
		Filters:             tfec2.BuildAttributeFilterList(filters),
	}

	return conn.DescribeClientVpnAuthorizationRules(input)

}

func ClientVpnAuthorizationRuleByID(conn *ec2.EC2, authorizationRuleID string) (*ec2.DescribeClientVpnAuthorizationRulesOutput, error) {
	endpointID, targetNetworkCidr, accessGroupID, err := tfec2.ClientVpnAuthorizationRuleParseID(authorizationRuleID)
	if err != nil {
		return nil, err
	}

	return ClientVpnAuthorizationRule(conn, endpointID, targetNetworkCidr, accessGroupID)
}

func ClientVpnRoute(conn *ec2.EC2, endpointID, targetSubnetID, destinationCidr string) (*ec2.DescribeClientVpnRoutesOutput, error) {
	filters := map[string]string{
		"target-subnet":    targetSubnetID,
		"destination-cidr": destinationCidr,
	}

	input := &ec2.DescribeClientVpnRoutesInput{
		ClientVpnEndpointId: aws.String(endpointID),
		Filters:             tfec2.BuildAttributeFilterList(filters),
	}

	return conn.DescribeClientVpnRoutes(input)
}

func ClientVpnRouteByID(conn *ec2.EC2, routeID string) (*ec2.DescribeClientVpnRoutesOutput, error) {
	endpointID, targetSubnetID, destinationCidr, err := tfec2.ClientVpnRouteParseID(routeID)
	if err != nil {
		return nil, err
	}

	return ClientVpnRoute(conn, endpointID, targetSubnetID, destinationCidr)
}

// InstanceByID looks up a Instance by ID. When not found, returns nil and potentially an API error.
func InstanceByID(conn *ec2.EC2, id string) (*ec2.Instance, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeInstances(input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Reservations) == 0 || output.Reservations[0] == nil || len(output.Reservations[0].Instances) == 0 || output.Reservations[0].Instances[0] == nil {
		return nil, nil
	}

	return output.Reservations[0].Instances[0], nil
}

// NetworkAclByID looks up a NetworkAcl by ID. When not found, returns nil and potentially an API error.
func NetworkAclByID(conn *ec2.EC2, id string) (*ec2.NetworkAcl, error) {
	input := &ec2.DescribeNetworkAclsInput{
		NetworkAclIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeNetworkAcls(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	for _, networkAcl := range output.NetworkAcls {
		if networkAcl == nil {
			continue
		}

		if aws.StringValue(networkAcl.NetworkAclId) != id {
			continue
		}

		return networkAcl, nil
	}

	return nil, nil
}

// NetworkAclEntry looks up a NetworkAclEntry by Network ACL ID, Egress, and Rule Number. When not found, returns nil and potentially an API error.
func NetworkAclEntry(conn *ec2.EC2, networkAclID string, egress bool, ruleNumber int) (*ec2.NetworkAclEntry, error) {
	input := &ec2.DescribeNetworkAclsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("entry.egress"),
				Values: aws.StringSlice([]string{fmt.Sprintf("%t", egress)}),
			},
			{
				Name:   aws.String("entry.rule-number"),
				Values: aws.StringSlice([]string{fmt.Sprintf("%d", ruleNumber)}),
			},
		},
		NetworkAclIds: aws.StringSlice([]string{networkAclID}),
	}

	output, err := conn.DescribeNetworkAcls(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	for _, networkAcl := range output.NetworkAcls {
		if networkAcl == nil {
			continue
		}

		if aws.StringValue(networkAcl.NetworkAclId) != networkAclID {
			continue
		}

		for _, entry := range output.NetworkAcls[0].Entries {
			if entry == nil {
				continue
			}

			if aws.BoolValue(entry.Egress) != egress || aws.Int64Value(entry.RuleNumber) != int64(ruleNumber) {
				continue
			}

			return entry, nil
		}
	}

	return nil, nil
}

// NetworkInterfaceByID looks up a NetworkInterface by ID. When not found, returns nil and potentially an API error.
func NetworkInterfaceByID(conn *ec2.EC2, id string) (*ec2.NetworkInterface, error) {
	input := &ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeNetworkInterfaces(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	for _, networkInterface := range output.NetworkInterfaces {
		if networkInterface == nil {
			continue
		}

		if aws.StringValue(networkInterface.NetworkInterfaceId) != id {
			continue
		}

		return networkInterface, nil
	}

	return nil, nil
}

// NetworkInterfaceSecurityGroup returns the associated GroupIdentifier if found
func NetworkInterfaceSecurityGroup(conn *ec2.EC2, networkInterfaceID string, securityGroupID string) (*ec2.GroupIdentifier, error) {
	var result *ec2.GroupIdentifier

	networkInterface, err := NetworkInterfaceByID(conn, networkInterfaceID)

	if err != nil {
		return nil, err
	}

	if networkInterface == nil {
		return nil, nil
	}

	for _, groupIdentifier := range networkInterface.Groups {
		if aws.StringValue(groupIdentifier.GroupId) == securityGroupID {
			result = groupIdentifier
			break
		}
	}

	return result, err
}

// MainRouteTableAssociationByID returns the main route table association corresponding to the specified identifier.
// Returns NotFoundError if no route table association is found.
func MainRouteTableAssociationByID(conn *ec2.EC2, associationID string) (*ec2.RouteTableAssociation, error) {
	association, err := RouteTableAssociationByID(conn, associationID)

	if err != nil {
		return nil, err
	}

	if !aws.BoolValue(association.Main) {
		return nil, &resource.NotFoundError{
			Message: fmt.Sprintf("%s is not the association with the main route table", associationID),
		}
	}

	return association, err
}

// MainRouteTableAssociationByVpcID returns the main route table association for the specified VPC.
// Returns NotFoundError if no route table association is found.
func MainRouteTableAssociationByVpcID(conn *ec2.EC2, vpcID string) (*ec2.RouteTableAssociation, error) {
	routeTable, err := MainRouteTableByVpcID(conn, vpcID)

	if err != nil {
		return nil, err
	}

	for _, association := range routeTable.Associations {
		if aws.BoolValue(association.Main) {
			if state := aws.StringValue(association.AssociationState.State); state == ec2.RouteTableAssociationStateCodeDisassociated {
				continue
			}

			return association, nil
		}
	}

	return nil, &resource.NotFoundError{}
}

// RouteTableAssociationByID returns the route table association corresponding to the specified identifier.
// Returns NotFoundError if no route table association is found.
func RouteTableAssociationByID(conn *ec2.EC2, associationID string) (*ec2.RouteTableAssociation, error) {
	input := &ec2.DescribeRouteTablesInput{
		Filters: tfec2.BuildAttributeFilterList(map[string]string{
			"association.route-table-association-id": associationID,
		}),
	}

	routeTable, err := RouteTable(conn, input)

	if err != nil {
		return nil, err
	}

	for _, association := range routeTable.Associations {
		if aws.StringValue(association.RouteTableAssociationId) == associationID {
			if state := aws.StringValue(association.AssociationState.State); state == ec2.RouteTableAssociationStateCodeDisassociated {
				return nil, &resource.NotFoundError{Message: state}
			}

			return association, nil
		}
	}

	return nil, &resource.NotFoundError{}
}

// MainRouteTableByVpcID returns the main route table for the specified VPC.
// Returns NotFoundError if no route table is found.
func MainRouteTableByVpcID(conn *ec2.EC2, vpcID string) (*ec2.RouteTable, error) {
	input := &ec2.DescribeRouteTablesInput{
		Filters: tfec2.BuildAttributeFilterList(map[string]string{
			"association.main": "true",
			"vpc-id":           vpcID,
		}),
	}

	return RouteTable(conn, input)
}

// RouteTableByID returns the route table corresponding to the specified identifier.
// Returns NotFoundError if no route table is found.
func RouteTableByID(conn *ec2.EC2, routeTableID string) (*ec2.RouteTable, error) {
	input := &ec2.DescribeRouteTablesInput{
		RouteTableIds: aws.StringSlice([]string{routeTableID}),
	}

	return RouteTable(conn, input)
}

func RouteTable(conn *ec2.EC2, input *ec2.DescribeRouteTablesInput) (*ec2.RouteTable, error) {
	output, err := conn.DescribeRouteTables(input)

	if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidRouteTableIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.RouteTables) == 0 || output.RouteTables[0] == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.RouteTables[0], nil
}

// RouteFinder returns the route corresponding to the specified destination.
// Returns NotFoundError if no route is found.
type RouteFinder func(*ec2.EC2, string, string) (*ec2.Route, error)

// RouteByIPv4Destination returns the route corresponding to the specified IPv4 destination.
// Returns NotFoundError if no route is found.
func RouteByIPv4Destination(conn *ec2.EC2, routeTableID, destinationCidr string) (*ec2.Route, error) {
	routeTable, err := RouteTableByID(conn, routeTableID)

	if err != nil {
		return nil, err
	}

	for _, route := range routeTable.Routes {
		if tfnet.CIDRBlocksEqual(aws.StringValue(route.DestinationCidrBlock), destinationCidr) {
			return route, nil
		}
	}

	return nil, &resource.NotFoundError{}
}

// RouteByIPv6Destination returns the route corresponding to the specified IPv6 destination.
// Returns NotFoundError if no route is found.
func RouteByIPv6Destination(conn *ec2.EC2, routeTableID, destinationIpv6Cidr string) (*ec2.Route, error) {
	routeTable, err := RouteTableByID(conn, routeTableID)

	if err != nil {
		return nil, err
	}

	for _, route := range routeTable.Routes {
		if tfnet.CIDRBlocksEqual(aws.StringValue(route.DestinationIpv6CidrBlock), destinationIpv6Cidr) {
			return route, nil
		}
	}

	return nil, &resource.NotFoundError{}
}

// RouteByPrefixListIDDestination returns the route corresponding to the specified prefix list destination.
// Returns NotFoundError if no route is found.
func RouteByPrefixListIDDestination(conn *ec2.EC2, routeTableID, prefixListID string) (*ec2.Route, error) {
	routeTable, err := RouteTableByID(conn, routeTableID)
	if err != nil {
		return nil, err
	}

	for _, route := range routeTable.Routes {
		if aws.StringValue(route.DestinationPrefixListId) == prefixListID {
			return route, nil
		}
	}

	return nil, &resource.NotFoundError{}
}

// SecurityGroupByID looks up a security group by ID. When not found, returns nil and potentially an API error.
func SecurityGroupByID(conn *ec2.EC2, id string) (*ec2.SecurityGroup, error) {
	req := &ec2.DescribeSecurityGroupsInput{
		GroupIds: aws.StringSlice([]string{id}),
	}
	result, err := conn.DescribeSecurityGroups(req)
	if err != nil {
		return nil, err
	}

	if result == nil || len(result.SecurityGroups) == 0 || result.SecurityGroups[0] == nil {
		return nil, nil
	}

	return result.SecurityGroups[0], nil
}

// SpotInstanceRequestByID looks up a SpotInstanceRequest by ID. When not found, returns nil and potentially an API error.
func SpotInstanceRequestByID(conn *ec2.EC2, id string) (*ec2.SpotInstanceRequest, error) {
	input := &ec2.DescribeSpotInstanceRequestsInput{
		SpotInstanceRequestIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeSpotInstanceRequests(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	for _, spotInstanceRequest := range output.SpotInstanceRequests {
		if spotInstanceRequest == nil {
			continue
		}

		if aws.StringValue(spotInstanceRequest.SpotInstanceRequestId) != id {
			continue
		}

		return spotInstanceRequest, nil
	}

	return nil, nil
}

// SubnetByID looks up a Subnet by ID. When not found, returns nil and potentially an API error.
func SubnetByID(conn *ec2.EC2, id string) (*ec2.Subnet, error) {
	input := &ec2.DescribeSubnetsInput{
		SubnetIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeSubnets(input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Subnets) == 0 || output.Subnets[0] == nil {
		return nil, nil
	}

	return output.Subnets[0], nil
}

func TransitGatewayPrefixListReference(conn *ec2.EC2, transitGatewayRouteTableID string, prefixListID string) (*ec2.TransitGatewayPrefixListReference, error) {
	filters := map[string]string{
		"prefix-list-id": prefixListID,
	}

	input := &ec2.GetTransitGatewayPrefixListReferencesInput{
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
		Filters:                    tfec2.BuildAttributeFilterList(filters),
	}

	var result *ec2.TransitGatewayPrefixListReference

	err := conn.GetTransitGatewayPrefixListReferencesPages(input, func(page *ec2.GetTransitGatewayPrefixListReferencesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, transitGatewayPrefixListReference := range page.TransitGatewayPrefixListReferences {
			if transitGatewayPrefixListReference == nil {
				continue
			}

			if aws.StringValue(transitGatewayPrefixListReference.PrefixListId) == prefixListID {
				result = transitGatewayPrefixListReference
				return false
			}
		}

		return !lastPage
	})

	return result, err
}

func TransitGatewayPrefixListReferenceByID(conn *ec2.EC2, resourceID string) (*ec2.TransitGatewayPrefixListReference, error) {
	transitGatewayRouteTableID, prefixListID, err := tfec2.TransitGatewayPrefixListReferenceParseID(resourceID)

	if err != nil {
		return nil, fmt.Errorf("error parsing EC2 Transit Gateway Prefix List Reference (%s) identifier: %w", resourceID, err)
	}

	return TransitGatewayPrefixListReference(conn, transitGatewayRouteTableID, prefixListID)
}

func TransitGatewayRouteTablePropagation(conn *ec2.EC2, transitGatewayRouteTableID string, transitGatewayAttachmentID string) (*ec2.TransitGatewayRouteTablePropagation, error) {
	input := &ec2.GetTransitGatewayRouteTablePropagationsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("transit-gateway-attachment-id"),
				Values: aws.StringSlice([]string{transitGatewayAttachmentID}),
			},
		},
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	var result *ec2.TransitGatewayRouteTablePropagation

	err := conn.GetTransitGatewayRouteTablePropagationsPages(input, func(page *ec2.GetTransitGatewayRouteTablePropagationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, transitGatewayRouteTablePropagation := range page.TransitGatewayRouteTablePropagations {
			if transitGatewayRouteTablePropagation == nil {
				continue
			}

			if aws.StringValue(transitGatewayRouteTablePropagation.TransitGatewayAttachmentId) == transitGatewayAttachmentID {
				result = transitGatewayRouteTablePropagation
				return false
			}
		}

		return !lastPage
	})

	return result, err
}

// VpcAttribute looks up a VPC attribute.
func VpcAttribute(conn *ec2.EC2, vpcID string, attribute string) (*bool, error) {
	input := &ec2.DescribeVpcAttributeInput{
		Attribute: aws.String(attribute),
		VpcId:     aws.String(vpcID),
	}

	output, err := conn.DescribeVpcAttribute(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	switch attribute {
	case ec2.VpcAttributeNameEnableDnsHostnames:
		if output.EnableDnsHostnames == nil {
			return nil, nil
		}

		return output.EnableDnsHostnames.Value, nil
	case ec2.VpcAttributeNameEnableDnsSupport:
		if output.EnableDnsSupport == nil {
			return nil, nil
		}

		return output.EnableDnsSupport.Value, nil
	}

	return nil, fmt.Errorf("unimplemented VPC attribute: %s", attribute)
}

// VpcByID looks up a Vpc by ID. When not found, returns nil and potentially an API error.
func VpcByID(conn *ec2.EC2, id string) (*ec2.Vpc, error) {
	input := &ec2.DescribeVpcsInput{
		VpcIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeVpcs(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	for _, vpc := range output.Vpcs {
		if vpc == nil {
			continue
		}

		if aws.StringValue(vpc.VpcId) != id {
			continue
		}

		return vpc, nil
	}

	return nil, nil
}

// VpcEndpointByID returns the VPC endpoint corresponding to the specified identifier.
// Returns NotFoundError if no VPC endpoint is found.
func VpcEndpointByID(conn *ec2.EC2, vpcEndpointID string) (*ec2.VpcEndpoint, error) {
	input := &ec2.DescribeVpcEndpointsInput{
		VpcEndpointIds: aws.StringSlice([]string{vpcEndpointID}),
	}

	vpcEndpoint, err := VpcEndpoint(conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(vpcEndpoint.State); state == tfec2.VpcEndpointStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(vpcEndpoint.VpcEndpointId) != vpcEndpointID {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return vpcEndpoint, nil
}

func VpcEndpoint(conn *ec2.EC2, input *ec2.DescribeVpcEndpointsInput) (*ec2.VpcEndpoint, error) {
	output, err := conn.DescribeVpcEndpoints(input)

	if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidVpcEndpointIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.VpcEndpoints) == 0 || output.VpcEndpoints[0] == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.VpcEndpoints[0], nil
}

// VpcEndpointRouteTableAssociationExists returns NotFoundError if no association for the specified VPC endpoint and route table IDs is found.
func VpcEndpointRouteTableAssociationExists(conn *ec2.EC2, vpcEndpointID string, routeTableID string) error {
	vpcEndpoint, err := VpcEndpointByID(conn, vpcEndpointID)

	if err != nil {
		return err
	}

	for _, vpcEndpointRouteTableID := range vpcEndpoint.RouteTableIds {
		if aws.StringValue(vpcEndpointRouteTableID) == routeTableID {
			return nil
		}
	}

	return &resource.NotFoundError{
		LastError: fmt.Errorf("VPC Endpoint Route Table Association (%s/%s) not found", vpcEndpointID, routeTableID),
	}
}

// VpcEndpointSubnetAssociationExists returns NotFoundError if no association for the specified VPC endpoint and subnet IDs is found.
func VpcEndpointSubnetAssociationExists(conn *ec2.EC2, vpcEndpointID string, subnetID string) error {
	vpcEndpoint, err := VpcEndpointByID(conn, vpcEndpointID)

	if err != nil {
		return err
	}

	for _, vpcEndpointSubnetID := range vpcEndpoint.SubnetIds {
		if aws.StringValue(vpcEndpointSubnetID) == subnetID {
			return nil
		}
	}

	return &resource.NotFoundError{
		LastError: fmt.Errorf("VPC Endpoint Subnet Association (%s/%s) not found", vpcEndpointID, subnetID),
	}
}

// VpcPeeringConnectionByID returns the VPC peering connection corresponding to the specified identifier.
// Returns nil and potentially an error if no VPC peering connection is found.
func VpcPeeringConnectionByID(conn *ec2.EC2, id string) (*ec2.VpcPeeringConnection, error) {
	input := &ec2.DescribeVpcPeeringConnectionsInput{
		VpcPeeringConnectionIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeVpcPeeringConnections(input)
	if err != nil {
		return nil, err
	}

	if output == nil || len(output.VpcPeeringConnections) == 0 {
		return nil, nil
	}

	return output.VpcPeeringConnections[0], nil
}

// VpnGatewayVpcAttachment returns the attachment between the specified VPN gateway and VPC.
// Returns nil and potentially an error if no attachment is found.
func VpnGatewayVpcAttachment(conn *ec2.EC2, vpnGatewayID, vpcID string) (*ec2.VpcAttachment, error) {
	vpnGateway, err := VpnGatewayByID(conn, vpnGatewayID)
	if err != nil {
		return nil, err
	}

	if vpnGateway == nil {
		return nil, nil
	}

	for _, vpcAttachment := range vpnGateway.VpcAttachments {
		if aws.StringValue(vpcAttachment.VpcId) == vpcID {
			return vpcAttachment, nil
		}
	}

	return nil, nil
}

// VpnGatewayByID returns the VPN gateway corresponding to the specified identifier.
// Returns nil and potentially an error if no VPN gateway is found.
func VpnGatewayByID(conn *ec2.EC2, id string) (*ec2.VpnGateway, error) {
	input := &ec2.DescribeVpnGatewaysInput{
		VpnGatewayIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeVpnGateways(input)
	if err != nil {
		return nil, err
	}

	if output == nil || len(output.VpnGateways) == 0 {
		return nil, nil
	}

	return output.VpnGateways[0], nil
}

func ManagedPrefixListByID(conn *ec2.EC2, id string) (*ec2.ManagedPrefixList, error) {
	input := &ec2.DescribeManagedPrefixListsInput{
		PrefixListIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeManagedPrefixLists(input)
	if err != nil {
		return nil, err
	}

	if output == nil || len(output.PrefixLists) == 0 {
		return nil, nil
	}

	return output.PrefixLists[0], nil
}
