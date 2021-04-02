package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloud9"
)

// EnvironmentByID returns the Environment corresponding to the specified ID.
func EnvironmentByID(conn *cloud9.Cloud9, id string) (*cloud9.Environment, error) {
	out, err := conn.DescribeEnvironments(&cloud9.DescribeEnvironmentsInput{
		EnvironmentIds: []*string{aws.String(id)},
	})
	if err != nil {
		return nil, err
	}

	envs := out.Environments
	if len(envs) == 0 {
		return nil, nil
	}

	env := envs[0]

	return env, nil
}
