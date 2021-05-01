package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	PrincipalAssociationTimeout    = 3 * time.Minute
	PrincipalDisassociationTimeout = 3 * time.Minute
)

// ResourceShareInvitationAccepted waits for a ResourceShareInvitation to return ACCEPTED
func ResourceShareInvitationAccepted(conn *ram.RAM, arn string, timeout time.Duration) (*ram.ResourceShareInvitation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ram.ResourceShareInvitationStatusPending},
		Target:  []string{ram.ResourceShareInvitationStatusAccepted},
		Refresh: ResourceShareInvitationStatus(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*ram.ResourceShareInvitation); ok {
		return v, err
	}

	return nil, err
}

// ResourceShareOwnedBySelfDisassociated waits for a ResourceShare owned by own account to be disassociated
func ResourceShareOwnedBySelfDisassociated(conn *ram.RAM, arn string, timeout time.Duration) (*ram.ResourceShare, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ram.ResourceShareAssociationStatusAssociated},
		Target:  []string{},
		Refresh: ResourceShareOwnerSelfStatus(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*ram.ResourceShare); ok {
		return v, err
	}

	return nil, err
}

// ResourceShareOwnedBySelfActive waits for a ResourceShare owned by own account to return ACTIVE
func ResourceShareOwnedBySelfActive(conn *ram.RAM, arn string, timeout time.Duration) (*ram.ResourceShare, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ram.ResourceShareStatusPending},
		Target:  []string{ram.ResourceShareStatusActive},
		Refresh: ResourceShareOwnerSelfStatus(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*ram.ResourceShare); ok {
		return v, err
	}

	return nil, err
}

// ResourceShareOwnedBySelfDeleted waits for a ResourceShare owned by own account to return DELETED
func ResourceShareOwnedBySelfDeleted(conn *ram.RAM, arn string, timeout time.Duration) (*ram.ResourceShare, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ram.ResourceShareStatusDeleting},
		Target:  []string{ram.ResourceShareStatusDeleted},
		Refresh: ResourceShareOwnerSelfStatus(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*ram.ResourceShare); ok {
		return v, err
	}

	return nil, err
}

func ResourceSharePrincipalAssociated(conn *ram.RAM, resourceShareARN, principal string) (*ram.ResourceShareAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ram.ResourceShareAssociationStatusAssociating, PrincipalAssociationStatusNotFound},
		Target:  []string{ram.ResourceShareAssociationStatusAssociated},
		Refresh: ResourceSharePrincipalAssociationStatus(conn, resourceShareARN, principal),
		Timeout: PrincipalAssociationTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*ram.ResourceShareAssociation); ok {
		return v, err
	}

	return nil, err
}

func ResourceSharePrincipalDisassociated(conn *ram.RAM, resourceShareARN, principal string) (*ram.ResourceShareAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ram.ResourceShareAssociationStatusAssociated, ram.ResourceShareAssociationStatusDisassociating},
		Target:  []string{ram.ResourceShareAssociationStatusDisassociated, PrincipalAssociationStatusNotFound},
		Refresh: ResourceSharePrincipalAssociationStatus(conn, resourceShareARN, principal),
		Timeout: PrincipalDisassociationTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*ram.ResourceShareAssociation); ok {
		return v, err
	}

	return nil, err
}
