package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
)

// StageByName returns the Stage corresponding to the specified name.
func StageByName(conn *apigateway.APIGateway, restApiId, name string) (*apigateway.Stage, error) {
	input := &apigateway.GetStageInput{
		RestApiId: aws.String(restApiId),
		StageName: aws.String(name),
	}

	output, err := conn.GetStage(input)
	if err != nil {
		return nil, err
	}

	return output, nil
}
