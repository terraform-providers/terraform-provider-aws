package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/synthetics"
)

// CanaryByName returns the Canary corresponding to the specified Name.
func CanaryByName(conn *synthetics.Synthetics, name string) (*synthetics.GetCanaryOutput, error) {
	input := &synthetics.GetCanaryInput{
		Name: aws.String(name),
	}

	output, err := conn.GetCanary(input)
	if err != nil {
		return nil, err
	}

	return output, nil
}
