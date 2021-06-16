// +build !generate

package keyvaluetags

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
)

// Custom IAM tag service update functions using the same format as generated code.

// IamRoleUpdateTags updates IAM role tags.
// The identifier is the role name.
func IamRoleUpdateTags(conn *iam.IAM, identifier string, oldTagsMap interface{}, newTagsMap interface{}) error {
	oldTags := New(oldTagsMap)
	newTags := New(newTagsMap)

	if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
		input := &iam.UntagRoleInput{
			RoleName: aws.String(identifier),
			TagKeys:  aws.StringSlice(removedTags.Keys()),
		}

		_, err := conn.UntagRole(input)

		if err != nil {
			return fmt.Errorf("error untagging resource (%s): %w", identifier, err)
		}
	}

	if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
		input := &iam.TagRoleInput{
			RoleName: aws.String(identifier),
			Tags:     updatedTags.IgnoreAws().IamTags(),
		}

		_, err := conn.TagRole(input)

		if err != nil {
			return fmt.Errorf("error tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}

// IamUserUpdateTags updates IAM user tags.
// The identifier is the user name.
func IamUserUpdateTags(conn *iam.IAM, identifier string, oldTagsMap interface{}, newTagsMap interface{}) error {
	oldTags := New(oldTagsMap)
	newTags := New(newTagsMap)

	if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
		input := &iam.UntagUserInput{
			UserName: aws.String(identifier),
			TagKeys:  aws.StringSlice(removedTags.Keys()),
		}

		_, err := conn.UntagUser(input)

		if err != nil {
			return fmt.Errorf("error untagging resource (%s): %w", identifier, err)
		}
	}

	if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
		input := &iam.TagUserInput{
			UserName: aws.String(identifier),
			Tags:     updatedTags.IgnoreAws().IamTags(),
		}

		_, err := conn.TagUser(input)

		if err != nil {
			return fmt.Errorf("error tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}

// IamInstanceProfileUpdateTags updates IAM Instance Profile tags.
// The identifier is the Instance Profile name.
func IamInstanceProfileUpdateTags(conn *iam.IAM, identifier string, oldTagsMap interface{}, newTagsMap interface{}) error {
	oldTags := New(oldTagsMap)
	newTags := New(newTagsMap)

	if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
		input := &iam.UntagInstanceProfileInput{
			InstanceProfileName: aws.String(identifier),
			TagKeys:             aws.StringSlice(removedTags.Keys()),
		}

		_, err := conn.UntagInstanceProfile(input)

		if err != nil {
			return fmt.Errorf("error untagging resource (%s): %w", identifier, err)
		}
	}

	if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
		input := &iam.TagInstanceProfileInput{
			InstanceProfileName: aws.String(identifier),
			Tags:                updatedTags.IgnoreAws().IamTags(),
		}

		_, err := conn.TagInstanceProfile(input)

		if err != nil {
			return fmt.Errorf("error tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}

// IamOpenIDConnectProviderUpdateTags updates IAM OpenID Connect Provider tags.
// The identifier is the OpenID Connect Provider ARN.
func IamOpenIDConnectProviderUpdateTags(conn *iam.IAM, identifier string, oldTagsMap interface{}, newTagsMap interface{}) error {
	oldTags := New(oldTagsMap)
	newTags := New(newTagsMap)

	if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
		input := &iam.UntagOpenIDConnectProviderInput{
			OpenIDConnectProviderArn: aws.String(identifier),
			TagKeys:                  aws.StringSlice(removedTags.Keys()),
		}

		_, err := conn.UntagOpenIDConnectProvider(input)

		if err != nil {
			return fmt.Errorf("error untagging resource (%s): %w", identifier, err)
		}
	}

	if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
		input := &iam.TagOpenIDConnectProviderInput{
			OpenIDConnectProviderArn: aws.String(identifier),
			Tags:                     updatedTags.IgnoreAws().IamTags(),
		}

		_, err := conn.TagOpenIDConnectProvider(input)

		if err != nil {
			return fmt.Errorf("error tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}

// IamPolicyUpdateTags updates IAM Policy tags.
// The identifier is the Policy ARN.
func IamPolicyUpdateTags(conn *iam.IAM, identifier string, oldTagsMap interface{}, newTagsMap interface{}) error {
	oldTags := New(oldTagsMap)
	newTags := New(newTagsMap)

	if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
		input := &iam.UntagPolicyInput{
			PolicyArn: aws.String(identifier),
			TagKeys:   aws.StringSlice(removedTags.Keys()),
		}

		_, err := conn.UntagPolicy(input)

		if err != nil {
			return fmt.Errorf("error untagging resource (%s): %w", identifier, err)
		}
	}

	if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
		input := &iam.TagPolicyInput{
			PolicyArn: aws.String(identifier),
			Tags:      updatedTags.IgnoreAws().IamTags(),
		}

		_, err := conn.TagPolicy(input)

		if err != nil {
			return fmt.Errorf("error tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}

// IamSAMLProviderUpdateTags updates IAM SAML Provider tags.
// The identifier is the SAML Provider ARN.
func IamSAMLProviderUpdateTags(conn *iam.IAM, identifier string, oldTagsMap interface{}, newTagsMap interface{}) error {
	oldTags := New(oldTagsMap)
	newTags := New(newTagsMap)

	if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
		input := &iam.UntagSAMLProviderInput{
			SAMLProviderArn: aws.String(identifier),
			TagKeys:         aws.StringSlice(removedTags.Keys()),
		}

		_, err := conn.UntagSAMLProvider(input)

		if err != nil {
			return fmt.Errorf("error untagging resource (%s): %w", identifier, err)
		}
	}

	if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
		input := &iam.TagSAMLProviderInput{
			SAMLProviderArn: aws.String(identifier),
			Tags:            updatedTags.IgnoreAws().IamTags(),
		}

		_, err := conn.TagSAMLProvider(input)

		if err != nil {
			return fmt.Errorf("error tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}

// IamServerCertificateUpdateTags updates IAM Server Certificate tags.
// The identifier is the Server Certificate name.
func IamServerCertificateUpdateTags(conn *iam.IAM, identifier string, oldTagsMap interface{}, newTagsMap interface{}) error {
	oldTags := New(oldTagsMap)
	newTags := New(newTagsMap)

	if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
		input := &iam.UntagServerCertificateInput{
			ServerCertificateName: aws.String(identifier),
			TagKeys:               aws.StringSlice(removedTags.Keys()),
		}

		_, err := conn.UntagServerCertificate(input)

		if err != nil {
			return fmt.Errorf("error untagging resource (%s): %w", identifier, err)
		}
	}

	if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
		input := &iam.TagServerCertificateInput{
			ServerCertificateName: aws.String(identifier),
			Tags:                  updatedTags.IgnoreAws().IamTags(),
		}

		_, err := conn.TagServerCertificate(input)

		if err != nil {
			return fmt.Errorf("error tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}
