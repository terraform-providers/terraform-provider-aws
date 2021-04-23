package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/go-multierror"
)

const (
	ErrCodeInvalidParameterValue = "InvalidParameterValue"
)

const (
	ErrCodeInvalidCarrierGatewayIDNotFound = "InvalidCarrierGatewayID.NotFound"
)

const (
	ErrCodeInvalidPrefixListIDNotFound = "InvalidPrefixListID.NotFound"
)

const (
	ErrCodeInvalidRouteTableIDNotFound = "InvalidRouteTableID.NotFound"
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
	ErrCodeInvalidSubnetIDNotFound = "InvalidSubnetID.NotFound"
)

const (
	ErrCodeInvalidVpcIDNotFound = "InvalidVpcID.NotFound"
)

const (
	ErrCodeInvalidVpcEndpointIdNotFound        = "InvalidVpcEndpointId.NotFound"
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

	return fmt.Errorf("%s: %s", aws.StringValue(apiObject.Code), aws.StringValue(apiObject.Message))
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
