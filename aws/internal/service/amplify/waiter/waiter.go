package waiter

import (
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/amplify"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

const (
	DomainAssociationCreatedTimeout  = 5 * time.Minute
	DomainAssociationVerifiedTimeout = 15 * time.Minute
)

func DomainAssociationCreated(conn *amplify.Amplify, appID, domainName string) (*amplify.DomainAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{amplify.DomainStatusCreating, amplify.DomainStatusInProgress, amplify.DomainStatusRequestingCertificate},
		Target:  []string{amplify.DomainStatusPendingVerification, amplify.DomainStatusPendingDeployment, amplify.DomainStatusAvailable},
		Refresh: DomainAssociationStatus(conn, appID, domainName),
		Timeout: DomainAssociationCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*amplify.DomainAssociation); ok {
		if status := aws.StringValue(v.DomainStatus); status == amplify.DomainStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(v.StatusReason)))
		}

		return v, err
	}

	return nil, err
}

func DomainAssociationVerified(conn *amplify.Amplify, appID, domainName string) (*amplify.DomainAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{amplify.DomainStatusUpdating, amplify.DomainStatusInProgress, amplify.DomainStatusPendingVerification},
		Target:  []string{amplify.DomainStatusPendingDeployment, amplify.DomainStatusAvailable},
		Refresh: DomainAssociationStatus(conn, appID, domainName),
		Timeout: DomainAssociationVerifiedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*amplify.DomainAssociation); ok {
		if v != nil && aws.StringValue(v.DomainStatus) == amplify.DomainStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(v.StatusReason)))
		}

		return v, err
	}

	return nil, err
}
