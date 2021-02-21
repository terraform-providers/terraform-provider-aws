package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	AWSSSOAdminAccountAssignmentCreateTimeout      = 5 * time.Minute
	AWSSSOAdminAccountAssignmentDeleteTimeout      = 5 * time.Minute
	AWSSSOAdminAccountAssignmentDelay              = 5 * time.Second
	AWSSSOAdminAccountAssignmentMinTimeout         = 3 * time.Second
	AWSSSOAdminPermissionSetProvisioningRetryDelay = 5 * time.Second
	AWSSSOAdminPermissionSetProvisionTimeout       = 10 * time.Minute
)

func AccountAssignmentCreated(conn *ssoadmin.SSOAdmin, instanceArn, requestID string) (*ssoadmin.AccountAssignmentOperationStatus, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ssoadmin.StatusValuesInProgress},
		Target:     []string{ssoadmin.StatusValuesSucceeded},
		Refresh:    AccountAssignmentCreationStatus(conn, instanceArn, requestID),
		Timeout:    AWSSSOAdminAccountAssignmentCreateTimeout,
		Delay:      AWSSSOAdminAccountAssignmentDelay,
		MinTimeout: AWSSSOAdminAccountAssignmentMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()
	if v, ok := outputRaw.(*ssoadmin.AccountAssignmentOperationStatus); ok {
		return v, err
	}

	return nil, err
}

func AccountAssignmentDeleted(conn *ssoadmin.SSOAdmin, instanceArn, requestID string) (*ssoadmin.AccountAssignmentOperationStatus, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ssoadmin.StatusValuesInProgress},
		Target:     []string{ssoadmin.StatusValuesSucceeded},
		Refresh:    AccountAssignmentDeletionStatus(conn, instanceArn, requestID),
		Timeout:    AWSSSOAdminAccountAssignmentDeleteTimeout,
		Delay:      AWSSSOAdminAccountAssignmentDelay,
		MinTimeout: AWSSSOAdminAccountAssignmentMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()
	if v, ok := outputRaw.(*ssoadmin.AccountAssignmentOperationStatus); ok {
		return v, err
	}

	return nil, err
}

func PermissionSetProvisioned(conn *ssoadmin.SSOAdmin, instanceArn, requestID string) (*ssoadmin.PermissionSetProvisioningStatus, error) {
	stateConf := resource.StateChangeConf{
		Delay:   AWSSSOAdminPermissionSetProvisioningRetryDelay,
		Pending: []string{ssoadmin.StatusValuesInProgress},
		Target:  []string{ssoadmin.StatusValuesSucceeded},
		Refresh: PermissionSetProvisioningStatus(conn, instanceArn, requestID),
		Timeout: AWSSSOAdminPermissionSetProvisionTimeout,
	}
	outputRaw, err := stateConf.WaitForState()
	if v, ok := outputRaw.(*ssoadmin.PermissionSetProvisioningStatus); ok {
		return v, err
	}
	return nil, err
}
