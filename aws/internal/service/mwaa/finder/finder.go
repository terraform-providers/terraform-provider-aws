package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mwaa"
)

// EnvironmentByName returns the MWAA Environment corresponding to the specified Name.
// Returns nil if no environment is found.
func EnvironmentByName(conn *mwaa.MWAA, name string) (*mwaa.Environment, error) {
	input := &mwaa.GetEnvironmentInput{
		Name: aws.String(name),
	}

	output, err := conn.GetEnvironment(input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output.Environment, nil
}
