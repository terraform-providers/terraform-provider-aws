package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	multierror "github.com/hashicorp/go-multierror"
)

const (
	ErrCodeGatewayNotAttached           = "Gateway.NotAttached"
	ErrCodeInvalidAssociationIDNotFound = "InvalidAssociationID.NotFound"
	ErrCodeInvalidParameter             = "InvalidParameter"
	ErrCodeInvalidParameterException    = "InvalidParameterException"
	ErrCodeInvalidParameterValue        = "InvalidParameterValue"
)

const (
	ErrCodeInvalidCarrierGatewayIDNotFound = "InvalidCarrierGatewayID.NotFound"
)

const (
	ErrCodeInvalidNetworkInterfaceIDNotFound = "InvalidNetworkInterfaceID.NotFound"
)

const (
	ErrCodeInvalidPrefixListIDNotFound = "InvalidPrefixListID.NotFound"
)

const (
	ErrCodeInvalidRouteNotFound        = "InvalidRoute.NotFound"
	ErrCodeInvalidRouteTableIdNotFound = "InvalidRouteTableId.NotFound"
	ErrCodeInvalidRouteTableIDNotFound = "InvalidRouteTableID.NotFound"
)

const (
	ErrCodeInvalidTransitGatewayIDNotFound = "InvalidTransitGatewayID.NotFound"
)

const (
	ErrCodeClientVpnEndpointIdNotFound        = "InvalidClientVpnEndpointId.NotFound"
	ErrCodeClientVpnAuthorizationRuleNotFound = "InvalidClientVpnEndpointAuthorizationRuleNotFound"
	ErrCodeClientVpnAssociationIdNotFound     = "InvalidClientVpnAssociationId.NotFound"
	ErrCodeClientVpnRouteNotFound             = "InvalidClientVpnRouteNotFound"
)

const (
	ErrCodeInvalidInstanceIDNotFound = "InvalidInstanceID.NotFound"
)

const (
	InvalidSecurityGroupIDNotFound = "InvalidSecurityGroupID.NotFound"
	InvalidGroupNotFound           = "InvalidGroup.NotFound"
)

const (
	ErrCodeInvalidSpotInstanceRequestIDNotFound = "InvalidSpotInstanceRequestID.NotFound"
)

const (
	ErrCodeInvalidSubnetIdNotFound = "InvalidSubnetId.NotFound"
	ErrCodeInvalidSubnetIDNotFound = "InvalidSubnetID.NotFound"
)

const (
	ErrCodeInvalidVpcIDNotFound = "InvalidVpcID.NotFound"
)

const (
	ErrCodeInvalidVpcEndpointIdNotFound        = "InvalidVpcEndpointId.NotFound"
	ErrCodeInvalidVpcEndpointNotFound          = "InvalidVpcEndpoint.NotFound"
	ErrCodeInvalidVpcEndpointServiceIdNotFound = "InvalidVpcEndpointServiceId.NotFound"
)

const (
	ErrCodeInvalidVpcPeeringConnectionIDNotFound = "InvalidVpcPeeringConnectionID.NotFound"
)

const (
	InvalidVpnGatewayAttachmentNotFound = "InvalidVpnGatewayAttachment.NotFound"
	InvalidVpnGatewayIDNotFound         = "InvalidVpnGatewayID.NotFound"
)

func UnsuccessfulItemError(apiObject *ec2.UnsuccessfulItemError) error {
	if apiObject == nil {
		return nil
	}

	return awserr.New(aws.StringValue(apiObject.Code), aws.StringValue(apiObject.Message), nil)
}

func UnsuccessfulItemsError(apiObjects []*ec2.UnsuccessfulItem) error {
	var errors *multierror.Error

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		err := UnsuccessfulItemError(apiObject.Error)

		if err != nil {
			errors = multierror.Append(errors, fmt.Errorf("%s: %w", aws.StringValue(apiObject.ResourceId), err))
		}
	}

	return errors.ErrorOrNil()
}
