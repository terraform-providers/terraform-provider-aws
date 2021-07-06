package finder

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
)

// GroupAttachedPolicy returns the AttachedPolicy corresponding to the specified group and policy ARN.
func GroupAttachedPolicy(conn *iam.IAM, groupName string, policyARN string) (*iam.AttachedPolicy, error) {
	input := &iam.ListAttachedGroupPoliciesInput{
		GroupName: aws.String(groupName),
	}

	var result *iam.AttachedPolicy

	err := conn.ListAttachedGroupPoliciesPages(input, func(page *iam.ListAttachedGroupPoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, attachedPolicy := range page.AttachedPolicies {
			if attachedPolicy == nil {
				continue
			}

			if aws.StringValue(attachedPolicy.PolicyArn) == policyARN {
				result = attachedPolicy
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// UserAttachedPolicy returns the AttachedPolicy corresponding to the specified user and policy ARN.
func UserAttachedPolicy(conn *iam.IAM, userName string, policyARN string) (*iam.AttachedPolicy, error) {
	input := &iam.ListAttachedUserPoliciesInput{
		UserName: aws.String(userName),
	}

	var result *iam.AttachedPolicy

	err := conn.ListAttachedUserPoliciesPages(input, func(page *iam.ListAttachedUserPoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, attachedPolicy := range page.AttachedPolicies {
			if attachedPolicy == nil {
				continue
			}

			if aws.StringValue(attachedPolicy.PolicyArn) == policyARN {
				result = attachedPolicy
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// Policies returns the Policies corresponding to the specified ARN, name, and/or path-prefix.
func Policies(conn *iam.IAM, arn, name, pathPrefix string) ([]*iam.Policy, error) {
	input := &iam.ListPoliciesInput{}

	if pathPrefix != "" {
		input.PathPrefix = aws.String(pathPrefix)
	}

	var results []*iam.Policy

	err := conn.ListPoliciesPages(input, func(page *iam.ListPoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, p := range page.Policies {
			if p == nil {
				continue
			}

			if arn != "" && arn != aws.StringValue(p.Arn) {
				continue
			}

			if name != "" && name != aws.StringValue(p.PolicyName) {
				continue
			}

			results = append(results, p)
		}

		return !lastPage
	})

	return results, err
}

// Role returns a role's ARN given the role name
func Role(conn *iam.IAM, name string) (*iam.Role, error) {
	input := &iam.GetRoleInput{
		RoleName: aws.String(name),
	}

	output, err := conn.GetRole(input)

	if err != nil {
		return nil, fmt.Errorf("getting IAM Role (%s): %w", name, err)
	}

	if output == nil || output.Role == nil {
		return nil, fmt.Errorf("getting IAM Role (%s): empty response", name)
	}

	return output.Role, nil
}
