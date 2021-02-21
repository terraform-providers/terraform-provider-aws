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
