package waiter

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/apigatewayv2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

// DeploymentStatus fetches the Deployment and its Status
func DeploymentStatus(conn *apigatewayv2.ApiGatewayV2, apiId, deploymentId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &apigatewayv2.GetDeploymentInput{
			ApiId:        aws.String(apiId),
			DeploymentId: aws.String(deploymentId),
		}

		output, err := conn.GetDeployment(input)

		if err != nil {
			return nil, apigatewayv2.DeploymentStatusFailed, err
		}

		// Error messages can also be contained in the response with FAILED status

		if aws.StringValue(output.DeploymentStatus) == apigatewayv2.DeploymentStatusFailed {
			return output, apigatewayv2.DeploymentStatusFailed, fmt.Errorf("%s", aws.StringValue(output.DeploymentStatusMessage))
		}

		return output, aws.StringValue(output.DeploymentStatus), nil
	}
}

func DomainNameStatus(conn *apigatewayv2.ApiGatewayV2, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		domainName, err := finder.DomainNameByName(conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if statusMessage := aws.StringValue(domainName.DomainNameConfigurations[0].DomainNameStatusMessage); statusMessage != "" {
			log.Printf("[INFO] API Gateway v2 domain name (%s) status message: %s", name, statusMessage)
		}

		return domainName, aws.StringValue(domainName.DomainNameConfigurations[0].DomainNameStatus), nil
	}
}

// VpcLinkStatus fetches the VPC Link and its Status
func VpcLinkStatus(conn *apigatewayv2.ApiGatewayV2, vpcLinkId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &apigatewayv2.GetVpcLinkInput{
			VpcLinkId: aws.String(vpcLinkId),
		}

		output, err := conn.GetVpcLink(input)

		if err != nil {
			return nil, apigatewayv2.VpcLinkStatusFailed, err
		}

		// Error messages can also be contained in the response with FAILED status

		if aws.StringValue(output.VpcLinkStatus) == apigatewayv2.VpcLinkStatusFailed {
			return output, apigatewayv2.VpcLinkStatusFailed, fmt.Errorf("%s", aws.StringValue(output.VpcLinkStatusMessage))
		}

		return output, aws.StringValue(output.VpcLinkStatus), nil
	}
}
