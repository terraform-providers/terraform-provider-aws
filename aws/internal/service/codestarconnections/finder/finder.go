package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codestarconnections"
)

// ConnectionByArn returns the Connection corresponding to the specified Arn.
func ConnectionByArn(conn *codestarconnections.CodeStarConnections, arn string) (*codestarconnections.Connection, error) {
	output, err := conn.GetConnection(&codestarconnections.GetConnectionInput{
		ConnectionArn: aws.String(arn),
	})
	if err != nil {
		return nil, err
	}

	if output == nil || output.Connection == nil {
		return nil, nil
	}

	return output.Connection, nil
}

func ConnectionByName(conn *codestarconnections.CodeStarConnections, name string) (*codestarconnections.Connection, error) {
	var result *codestarconnections.Connection

	err := conn.ListConnectionsPages(&codestarconnections.ListConnectionsInput{}, func(page *codestarconnections.ListConnectionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, connection := range page.Connections {
			if connection == nil {
				continue
			}

			if aws.StringValue(connection.ConnectionName) == name {
				result = connection
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
