package finder

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
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
