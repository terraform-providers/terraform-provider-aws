package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
)

// ApiByID returns the API corresponding to the specified ID.
func ApiByID(conn *apigatewayv2.ApiGatewayV2, apiID string) (*apigatewayv2.GetApiOutput, error) {
	input := &apigatewayv2.GetApiInput{
		ApiId: aws.String(apiID),
	}

	output, err := conn.GetApi(input)
	if err != nil {
		return nil, err
	}

	return output, nil
}
