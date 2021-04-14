package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
)

// AccountAssignment returns the account assigned to a permission set within a specified SSO instance.
// Returns an error if no account assignment is found.
func AccountAssignment(conn *ssoadmin.SSOAdmin, principalId, principalType, accountId, permissionSetArn, instanceArn string) (*ssoadmin.AccountAssignment, error) {
	input := &ssoadmin.ListAccountAssignmentsInput{
		AccountId:        aws.String(accountId),
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
	}

	var accountAssignment *ssoadmin.AccountAssignment
	err := conn.ListAccountAssignmentsPages(input, func(page *ssoadmin.ListAccountAssignmentsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, a := range page.AccountAssignments {
			if a == nil {
				continue
			}

			if aws.StringValue(a.PrincipalType) != principalType {
				continue
			}
			if aws.StringValue(a.PrincipalId) == principalId {
				accountAssignment = a
				return false
			}
		}

		return !lastPage
	})

	return accountAssignment, err
}

// ManagedPolicy returns the managed policy attached to a permission set within a specified SSO instance.
// Returns an error if no managed policy is found.
func ManagedPolicy(conn *ssoadmin.SSOAdmin, managedPolicyArn, permissionSetArn, instanceArn string) (*ssoadmin.AttachedManagedPolicy, error) {
	input := &ssoadmin.ListManagedPoliciesInPermissionSetInput{
		PermissionSetArn: aws.String(permissionSetArn),
		InstanceArn:      aws.String(instanceArn),
	}

	var attachedPolicy *ssoadmin.AttachedManagedPolicy
	err := conn.ListManagedPoliciesInPermissionSetPages(input, func(page *ssoadmin.ListManagedPoliciesInPermissionSetOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, policy := range page.AttachedManagedPolicies {
			if policy == nil {
				continue
			}

			if aws.StringValue(policy.Arn) == managedPolicyArn {
				attachedPolicy = policy
				return false
			}
		}
		return !lastPage
	})

	return attachedPolicy, err
}
