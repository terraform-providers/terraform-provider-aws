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
