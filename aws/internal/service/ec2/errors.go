package ec2

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws/awserr"
)

// Copied from aws-sdk-go-base
// Can be removed when aws-sdk-go-base v0.6+ is merged
// TODO:
func ErrCodeEquals(err error, code string) bool {
	var awsErr awserr.Error
	if errors.As(err, &awsErr) {
		return awsErr.Code() == code
	}
	return false
}

const (
	ErrCodeClientVpnEndpointIdNotFound        = "InvalidClientVpnEndpointId.NotFound"
	ErrCodeClientVpnAuthorizationRuleNotFound = "InvalidClientVpnEndpointAuthorizationRuleNotFound"
	ErrCodeClientVpnAssociationIdNotFound     = "InvalidClientVpnAssociationId.NotFound"
	ErrCodeClientVpnRouteNotFound             = "InvalidClientVpnRouteNotFound"
)

const (
	InvalidSecurityGroupIDNotFound = "InvalidSecurityGroupID.NotFound"
	InvalidGroupNotFound           = "InvalidGroup.NotFound"
)
