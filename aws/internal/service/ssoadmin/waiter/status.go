package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	AccountAssignmentStatusUnknown          = "Unknown"
	AccountAssignmentStatusNotFound         = "NotFound"
	PermissionSetProvisioningStatusUnknown  = "Unknown"
	PermissionSetProvisioningStatusNotFound = "NotFound"
)

func AccountAssignmentCreationStatus(conn *ssoadmin.SSOAdmin, instanceArn, requestID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &ssoadmin.DescribeAccountAssignmentCreationStatusInput{
			AccountAssignmentCreationRequestId: aws.String(requestID),
			InstanceArn:                        aws.String(instanceArn),
		}

		resp, err := conn.DescribeAccountAssignmentCreationStatus(input)

		if err != nil {
			return nil, AccountAssignmentStatusUnknown, err
		}

		if resp == nil || resp.AccountAssignmentCreationStatus == nil {
			return nil, AccountAssignmentStatusNotFound, nil
		}

		return resp.AccountAssignmentCreationStatus, aws.StringValue(resp.AccountAssignmentCreationStatus.Status), nil
	}
}

func AccountAssignmentDeletionStatus(conn *ssoadmin.SSOAdmin, instanceArn, requestID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &ssoadmin.DescribeAccountAssignmentDeletionStatusInput{
			AccountAssignmentDeletionRequestId: aws.String(requestID),
			InstanceArn:                        aws.String(instanceArn),
		}

		resp, err := conn.DescribeAccountAssignmentDeletionStatus(input)

		if err != nil {
			return nil, AccountAssignmentStatusUnknown, err
		}

		if resp == nil || resp.AccountAssignmentDeletionStatus == nil {
			return nil, AccountAssignmentStatusNotFound, nil
		}

		return resp.AccountAssignmentDeletionStatus, aws.StringValue(resp.AccountAssignmentDeletionStatus.Status), nil
	}
}

func PermissionSetProvisioningStatus(conn *ssoadmin.SSOAdmin, instanceArn, requestID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &ssoadmin.DescribePermissionSetProvisioningStatusInput{
			InstanceArn:                     aws.String(instanceArn),
			ProvisionPermissionSetRequestId: aws.String(requestID),
		}

		resp, err := conn.DescribePermissionSetProvisioningStatus(input)

		if err != nil {
			return nil, PermissionSetProvisioningStatusUnknown, err
		}

		if resp == nil || resp.PermissionSetProvisioningStatus == nil {
			return nil, PermissionSetProvisioningStatusNotFound, nil
		}

		return resp.PermissionSetProvisioningStatus, aws.StringValue(resp.PermissionSetProvisioningStatus.Status), nil
	}
}
