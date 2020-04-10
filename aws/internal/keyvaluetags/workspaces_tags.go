// +build !generate

package keyvaluetags

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
)

// Custom WorkSpaces tag service update functions using the same format as generated code.

// WorkspacesUpdateTags updates WorkSpaces resource tags.
// The identifier is the resource ID.
func WorkspacesUpdateTags(conn *workspaces.WorkSpaces, identifier string, oldTagsMap interface{}, newTagsMap interface{}) error {
	oldTags := New(oldTagsMap)
	newTags := New(newTagsMap)

	// https://docs.aws.amazon.com/workspaces/latest/api/API_CreateTags.html
	// "If you want to add new tags to a set of existing tags, you must submit all of the existing tags along with the new ones."
	// This doesn't in fact seem to be correct...

	if len(newTags) > 0 {
		input := &workspaces.CreateTagsInput{
			ResourceId: aws.String(identifier),
			Tags:       newTags.IgnoreAws().WorkspacesTags(),
		}

		_, err := conn.CreateTags(input)

		if err != nil {
			return fmt.Errorf("error tagging resource (%s): %w", identifier, err)
		}
	} else if len(oldTags) > 0 {
		input := &workspaces.DeleteTagsInput{
			ResourceId: aws.String(identifier),
			TagKeys:    aws.StringSlice(oldTags.Keys()),
		}

		_, err := conn.DeleteTags(input)

		if err != nil {
			return fmt.Errorf("error untagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}
