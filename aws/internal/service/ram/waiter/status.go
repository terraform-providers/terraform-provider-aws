package waiter

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ram/finder"
)

const (
	ResourceShareInvitationStatusNotFound = "NotFound"
	ResourceShareInvitationStatusUnknown  = "Unknown"

	ResourceShareStatusNotFound = "NotFound"
	ResourceShareStatusUnknown  = "Unknown"

	PrincipalAssociationStatusNotFound = "NotFound"
)

// ResourceShareInvitationStatus fetches the ResourceShareInvitation and its Status
func ResourceShareInvitationStatus(conn *ram.RAM, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		invitation, err := finder.ResourceShareInvitationByArn(conn, arn)

		if err != nil {
			return nil, ResourceShareInvitationStatusUnknown, err
		}

		if invitation == nil {
			return nil, ResourceShareInvitationStatusNotFound, nil
		}

		return invitation, aws.StringValue(invitation.Status), nil
	}
}

// ResourceShareOwnerSelfStatus fetches the ResourceShare and its Status
func ResourceShareOwnerSelfStatus(conn *ram.RAM, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		share, err := finder.ResourceShareOwnerSelfByArn(conn, arn)

		if tfawserr.ErrCodeEquals(err, ram.ErrCodeUnknownResourceException) {
			return nil, ResourceShareStatusNotFound, nil
		}

		if err != nil {
			return nil, ResourceShareStatusUnknown, err
		}

		if share == nil {
			return nil, ResourceShareStatusNotFound, nil
		}

		return share, aws.StringValue(share.Status), nil
	}
}

func ResourceSharePrincipalAssociationStatus(conn *ram.RAM, resourceShareArn, principal string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		association, err := finder.ResourceSharePrincipalAssociationByShareARNPrincipal(conn, resourceShareArn, principal)

		if tfawserr.ErrCodeEquals(err, ram.ErrCodeUnknownResourceException) {
			return nil, PrincipalAssociationStatusNotFound, err
		}

		if err != nil {
			return nil, ram.ResourceShareAssociationStatusFailed, err
		}

		if association == nil {
			return nil, ram.ResourceShareAssociationStatusDisassociated, nil
		}

		if aws.StringValue(association.Status) == ram.ResourceShareAssociationStatusFailed {
			extendedErr := fmt.Errorf("association status message: %s", aws.StringValue(association.StatusMessage))
			return association, aws.StringValue(association.Status), extendedErr
		}

		return association, aws.StringValue(association.Status), nil
	}
}
