package finder

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
)

// DocumentByName returns the Document corresponding to the specified name.
func DocumentByName(conn *ssm.SSM, name string) (*ssm.DocumentDescription, error) {
	input := &ssm.DescribeDocumentInput{
		Name: aws.String(name),
	}

	output, err := conn.DescribeDocument(input)
	if err != nil {
		return nil, err
	}

	if output == nil || output.Document == nil {
		return nil, fmt.Errorf("error describing SSM Document (%s): empty result", name)
	}

	doc := output.Document

	if aws.StringValue(doc.Status) == ssm.DocumentStatusFailed {
		return nil, fmt.Errorf("Document is in a failed state: %s", aws.StringValue(doc.StatusInformation))
	}

	return output.Document, nil
}

// PatchGroup returns matching SSM Patch Group by Patch Group and BaselineId.
func PatchGroup(conn *ssm.SSM, patchGroup, baselineId string) (*ssm.PatchGroupPatchBaselineMapping, error) {
	input := &ssm.DescribePatchGroupsInput{}
	var result *ssm.PatchGroupPatchBaselineMapping

	err := conn.DescribePatchGroupsPages(input, func(page *ssm.DescribePatchGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, mapping := range page.Mappings {
			if mapping == nil {
				continue
			}

			if aws.StringValue(mapping.PatchGroup) == patchGroup {
				if mapping.BaselineIdentity != nil && aws.StringValue(mapping.BaselineIdentity.BaselineId) == baselineId {
					result = mapping
					return false
				}
			}
		}

		return !lastPage
	})

	return result, err
}
