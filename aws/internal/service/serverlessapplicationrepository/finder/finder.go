package finder

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	serverlessrepository "github.com/aws/aws-sdk-go/service/serverlessapplicationrepository"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func Application(conn *serverlessrepository.ServerlessApplicationRepository, applicationID, version string) (*serverlessrepository.GetApplicationOutput, error) {
	input := &serverlessrepository.GetApplicationInput{
		ApplicationId: aws.String(applicationID),
	}
	if version != "" {
		input.SemanticVersion = aws.String(version)
	}

	log.Printf("[DEBUG] Getting Serverless Application Repository Application: %s", input)
	resp, err := conn.GetApplication(input)
	if tfawserr.ErrCodeEquals(err, serverlessrepository.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:    err,
			LastRequest:  input,
			LastResponse: resp,
		}
	}
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return nil, &resource.NotFoundError{
			LastRequest:  input,
			LastResponse: resp,
			Message:      "returned empty response",
		}
	}

	return resp, nil

}
