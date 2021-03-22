package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/apigatewayv2/lister"
)

// ApiByID returns the API corresponding to the specified ID.
// Returns NotFoundError if no API is found.
func ApiByID(conn *apigatewayv2.ApiGatewayV2, apiID string) (*apigatewayv2.GetApiOutput, error) {
	input := &apigatewayv2.GetApiInput{
		ApiId: aws.String(apiID),
	}

	return Api(conn, input)
}

// Api returns the API corresponding to the specified input.
// Returns NotFoundError if no API is found.
func Api(conn *apigatewayv2.ApiGatewayV2, input *apigatewayv2.GetApiInput) (*apigatewayv2.GetApiOutput, error) {
	output, err := conn.GetApi(input)

	if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	// Handle any empty result.
	if output == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output, nil
}

// Apis returns the APIs corresponding to the specified input.
// Returns an empty slice if no APIs are found.
func Apis(conn *apigatewayv2.ApiGatewayV2, input *apigatewayv2.GetApisInput) ([]*apigatewayv2.Api, error) {
	var apis []*apigatewayv2.Api

	err := lister.GetApisPages(conn, input, func(page *apigatewayv2.GetApisOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, item := range page.Items {
			if item == nil {
				continue
			}

			apis = append(apis, item)
		}

		return !isLast
	})

	if err != nil {
		return nil, err
	}

	return apis, nil
}
