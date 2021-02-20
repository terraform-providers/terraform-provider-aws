package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	DocumentDeleteTimeout = 2 * time.Minute
	DocumentActiveTimeout = 2 * time.Minute
)

// DocumentDeleted waits for an Document to return Deleted
func DocumentDeleted(conn *ssm.SSM, name string) (*ssm.DocumentDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ssm.DocumentStatusDeleting},
		Target:  []string{},
		Refresh: DocumentStatus(conn, name),
		Timeout: DocumentDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ssm.DocumentDescription); ok {
		return output, err
	}

	return nil, err
}

// DocumentActive waits for an Document to return Active
func DocumentActive(conn *ssm.SSM, name string) (*ssm.DocumentDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ssm.DocumentStatusCreating, ssm.DocumentStatusUpdating},
		Target:  []string{ssm.DocumentStatusActive},
		Refresh: DocumentStatus(conn, name),
		Timeout: DocumentActiveTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ssm.DocumentDescription); ok {
		return output, err
	}

	return nil, err
}

// AssociationSuccess waits for an Association to return Success
func AssociationSuccess(conn *ssm.SSM, id string, timeout time.Duration) (*ssm.AssociationDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ssm.AssociationStatusNamePending},
		Target:  []string{ssm.AssociationStatusNameSuccess},
		Refresh: AssociationStatus(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ssm.AssociationDescription); ok {
		return output, err
	}

	return nil, err
}
