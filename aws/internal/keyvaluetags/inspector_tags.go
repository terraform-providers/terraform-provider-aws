// +build !generate

package keyvaluetags

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/inspector"
)

// Custom Inspector tag service update functions using the same format as generated code.

// InspectorUpdateTags updates WorkSpaces resource tags.
// The identifier is the resource ARN.
func InspectorUpdateTags(conn *inspector.Inspector, identifier string, oldTagsMap interface{}, newTagsMap interface{}) error {
	oldTags := New(oldTagsMap)
	newTags := New(newTagsMap)

	if len(newTags) > 0 {
		input := &inspector.SetTagsForResourceInput{
			ResourceArn: aws.String(identifier),
			Tags:        newTags.IgnoreAws().InspectorTags(),
		}

		_, err := conn.SetTagsForResource(input)

		if err != nil {
			return fmt.Errorf("error tagging resource (%s): %w", identifier, err)
		}
	} else if len(oldTags) > 0 {
		input := &inspector.SetTagsForResourceInput{
			ResourceArn: aws.String(identifier),
		}

		_, err := conn.SetTagsForResource(input)

		if err != nil {
			return fmt.Errorf("error untagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}
