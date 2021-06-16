package lister

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
)

// Custom Cloudfront listing functions using similar formatting as other service generated code.

func ListFunctionsPages(conn *cloudfront.CloudFront, input *cloudfront.ListFunctionsInput, fn func(*cloudfront.ListFunctionsOutput, bool) bool) error {
	for {
		output, err := conn.ListFunctions(input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.FunctionList.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.Marker = output.FunctionList.NextMarker
	}
	return nil
}
