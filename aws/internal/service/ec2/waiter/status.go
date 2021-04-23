package waiter

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
)

const (
	carrierGatewayStateNotFound = "NotFound"
	carrierGatewayStateUnknown  = "Unknown"
)

// CarrierGatewayState fetches the CarrierGateway and its State
func CarrierGatewayState(conn *ec2.EC2, carrierGatewayID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		carrierGateway, err := finder.CarrierGatewayByID(conn, carrierGatewayID)
		if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidCarrierGatewayIDNotFound) {
			return nil, carrierGatewayStateNotFound, nil
		}
		if err != nil {
			return nil, carrierGatewayStateUnknown, err
		}

		if carrierGateway == nil {
			return nil, carrierGatewayStateNotFound, nil
		}

		state := aws.StringValue(carrierGateway.State)

		if state == ec2.CarrierGatewayStateDeleted {
			return nil, carrierGatewayStateNotFound, nil
		}

		return carrierGateway, state, nil
	}
}

// LocalGatewayRouteTableVpcAssociationState fetches the LocalGatewayRouteTableVpcAssociation and its State
func LocalGatewayRouteTableVpcAssociationState(conn *ec2.EC2, localGatewayRouteTableVpcAssociationID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &ec2.DescribeLocalGatewayRouteTableVpcAssociationsInput{
			LocalGatewayRouteTableVpcAssociationIds: aws.StringSlice([]string{localGatewayRouteTableVpcAssociationID}),
		}

		output, err := conn.DescribeLocalGatewayRouteTableVpcAssociations(input)

		if err != nil {
			return nil, "", err
		}

		var association *ec2.LocalGatewayRouteTableVpcAssociation

		for _, outputAssociation := range output.LocalGatewayRouteTableVpcAssociations {
			if outputAssociation == nil {
				continue
			}

			if aws.StringValue(outputAssociation.LocalGatewayRouteTableVpcAssociationId) == localGatewayRouteTableVpcAssociationID {
				association = outputAssociation
				break
			}
		}

		if association == nil {
			return association, ec2.RouteTableAssociationStateCodeDisassociated, nil
		}

		return association, aws.StringValue(association.State), nil
	}
}

const (
	ClientVpnEndpointStatusNotFound = "NotFound"

	ClientVpnEndpointStatusUnknown = "Unknown"
)

// ClientVpnEndpointStatus fetches the Client VPN endpoint and its Status
func ClientVpnEndpointStatus(conn *ec2.EC2, endpointID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		result, err := conn.DescribeClientVpnEndpoints(&ec2.DescribeClientVpnEndpointsInput{
			ClientVpnEndpointIds: aws.StringSlice([]string{endpointID}),
		})
		if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeClientVpnEndpointIdNotFound) {
			return nil, ClientVpnEndpointStatusNotFound, nil
		}
		if err != nil {
			return nil, ClientVpnEndpointStatusUnknown, err
		}

		if result == nil || len(result.ClientVpnEndpoints) == 0 || result.ClientVpnEndpoints[0] == nil {
			return nil, ClientVpnEndpointStatusNotFound, nil
		}

		endpoint := result.ClientVpnEndpoints[0]
		if endpoint.Status == nil || endpoint.Status.Code == nil {
			return endpoint, ClientVpnEndpointStatusUnknown, nil
		}

		return endpoint, aws.StringValue(endpoint.Status.Code), nil
	}
}

const (
	ClientVpnAuthorizationRuleStatusNotFound = "NotFound"

	ClientVpnAuthorizationRuleStatusUnknown = "Unknown"
)

// ClientVpnAuthorizationRuleStatus fetches the Client VPN authorization rule and its Status
func ClientVpnAuthorizationRuleStatus(conn *ec2.EC2, authorizationRuleID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		result, err := finder.ClientVpnAuthorizationRuleByID(conn, authorizationRuleID)
		if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeClientVpnAuthorizationRuleNotFound) {
			return nil, ClientVpnAuthorizationRuleStatusNotFound, nil
		}
		if err != nil {
			return nil, ClientVpnAuthorizationRuleStatusUnknown, err
		}

		if result == nil || len(result.AuthorizationRules) == 0 || result.AuthorizationRules[0] == nil {
			return nil, ClientVpnAuthorizationRuleStatusNotFound, nil
		}

		if len(result.AuthorizationRules) > 1 {
			return nil, ClientVpnAuthorizationRuleStatusUnknown, fmt.Errorf("internal error: found %d results for Client VPN authorization rule (%s) status, need 1", len(result.AuthorizationRules), authorizationRuleID)
		}

		rule := result.AuthorizationRules[0]
		if rule.Status == nil || rule.Status.Code == nil {
			return rule, ClientVpnAuthorizationRuleStatusUnknown, nil
		}

		return rule, aws.StringValue(rule.Status.Code), nil
	}
}

const (
	ClientVpnNetworkAssociationStatusNotFound = "NotFound"

	ClientVpnNetworkAssociationStatusUnknown = "Unknown"
)

func ClientVpnNetworkAssociationStatus(conn *ec2.EC2, cvnaID string, cvepID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		result, err := conn.DescribeClientVpnTargetNetworks(&ec2.DescribeClientVpnTargetNetworksInput{
			ClientVpnEndpointId: aws.String(cvepID),
			AssociationIds:      []*string{aws.String(cvnaID)},
		})

		if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeClientVpnAssociationIdNotFound) || tfawserr.ErrCodeEquals(err, tfec2.ErrCodeClientVpnEndpointIdNotFound) {
			return nil, ClientVpnNetworkAssociationStatusNotFound, nil
		}
		if err != nil {
			return nil, ClientVpnNetworkAssociationStatusUnknown, err
		}

		if result == nil || len(result.ClientVpnTargetNetworks) == 0 || result.ClientVpnTargetNetworks[0] == nil {
			return nil, ClientVpnNetworkAssociationStatusNotFound, nil
		}

		network := result.ClientVpnTargetNetworks[0]
		if network.Status == nil || network.Status.Code == nil {
			return network, ClientVpnNetworkAssociationStatusUnknown, nil
		}

		return network, aws.StringValue(network.Status.Code), nil
	}
}

const (
	ClientVpnRouteStatusNotFound = "NotFound"

	ClientVpnRouteStatusUnknown = "Unknown"
)

// ClientVpnRouteStatus fetches the Client VPN route and its Status
func ClientVpnRouteStatus(conn *ec2.EC2, routeID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		result, err := finder.ClientVpnRouteByID(conn, routeID)
		if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeClientVpnRouteNotFound) {
			return nil, ClientVpnRouteStatusNotFound, nil
		}
		if err != nil {
			return nil, ClientVpnRouteStatusUnknown, err
		}

		if result == nil || len(result.Routes) == 0 || result.Routes[0] == nil {
			return nil, ClientVpnRouteStatusNotFound, nil
		}

		if len(result.Routes) > 1 {
			return nil, ClientVpnRouteStatusUnknown, fmt.Errorf("internal error: found %d results for Client VPN route (%s) status, need 1", len(result.Routes), routeID)
		}

		rule := result.Routes[0]
		if rule.Status == nil || rule.Status.Code == nil {
			return rule, ClientVpnRouteStatusUnknown, nil
		}

		return rule, aws.StringValue(rule.Status.Code), nil
	}
}

const (
	SecurityGroupStatusCreated = "Created"

	SecurityGroupStatusNotFound = "NotFound"

	SecurityGroupStatusUnknown = "Unknown"
)

// SecurityGroupStatus fetches the security group and its status
func SecurityGroupStatus(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		group, err := finder.SecurityGroupByID(conn, id)
		if tfawserr.ErrCodeEquals(err, tfec2.InvalidSecurityGroupIDNotFound) ||
			tfawserr.ErrCodeEquals(err, tfec2.InvalidGroupNotFound) {
			return nil, SecurityGroupStatusNotFound, nil
		}
		if err != nil {
			return nil, SecurityGroupStatusUnknown, err
		}

		if group == nil {
			return nil, SecurityGroupStatusNotFound, nil
		}

		return group, SecurityGroupStatusCreated, nil
	}
}

const (
	vpcPeeringConnectionStatusNotFound = "NotFound"
	vpcPeeringConnectionStatusUnknown  = "Unknown"
)

// VpcPeeringConnectionStatus fetches the VPC peering connection and its status
func VpcPeeringConnectionStatus(conn *ec2.EC2, vpcPeeringConnectionID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		vpcPeeringConnection, err := finder.VpcPeeringConnectionByID(conn, vpcPeeringConnectionID)
		if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidVpcPeeringConnectionIDNotFound) {
			return nil, vpcPeeringConnectionStatusNotFound, nil
		}
		if err != nil {
			return nil, vpcPeeringConnectionStatusUnknown, err
		}

		// Sometimes AWS just has consistency issues and doesn't see
		// our peering connection yet. Return an empty state.
		if vpcPeeringConnection == nil || vpcPeeringConnection.Status == nil {
			return nil, vpcPeeringConnectionStatusNotFound, nil
		}

		statusCode := aws.StringValue(vpcPeeringConnection.Status.Code)

		// https://docs.aws.amazon.com/vpc/latest/peering/vpc-peering-basics.html#vpc-peering-lifecycle
		switch statusCode {
		case ec2.VpcPeeringConnectionStateReasonCodeFailed:
			log.Printf("[WARN] VPC Peering Connection (%s): %s: %s", vpcPeeringConnectionID, statusCode, aws.StringValue(vpcPeeringConnection.Status.Message))
			fallthrough
		case ec2.VpcPeeringConnectionStateReasonCodeDeleted, ec2.VpcPeeringConnectionStateReasonCodeExpired, ec2.VpcPeeringConnectionStateReasonCodeRejected:
			return nil, vpcPeeringConnectionStatusNotFound, nil
		}

		return vpcPeeringConnection, statusCode, nil
	}
}

const (
	attachmentStateNotFound = "NotFound"
	attachmentStateUnknown  = "Unknown"
)

// VpnGatewayVpcAttachmentState fetches the attachment between the specified VPN gateway and VPC and its state
func VpnGatewayVpcAttachmentState(conn *ec2.EC2, vpnGatewayID, vpcID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		vpcAttachment, err := finder.VpnGatewayVpcAttachment(conn, vpnGatewayID, vpcID)
		if tfawserr.ErrCodeEquals(err, tfec2.InvalidVpnGatewayIDNotFound) {
			return nil, attachmentStateNotFound, nil
		}
		if err != nil {
			return nil, attachmentStateUnknown, err
		}

		if vpcAttachment == nil {
			return nil, attachmentStateNotFound, nil
		}

		return vpcAttachment, aws.StringValue(vpcAttachment.State), nil
	}
}

const (
	managedPrefixListStateNotFound = "NotFound"
	managedPrefixListStateUnknown  = "Unknown"
)

func ManagedPrefixListState(conn *ec2.EC2, prefixListId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		managedPrefixList, err := finder.ManagedPrefixListByID(conn, prefixListId)
		if err != nil {
			return nil, managedPrefixListStateUnknown, err
		}
		if managedPrefixList == nil {
			return nil, managedPrefixListStateNotFound, nil
		}

		return managedPrefixList, aws.StringValue(managedPrefixList.State), nil
	}
}
