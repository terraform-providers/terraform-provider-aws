package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// ConnectionByID returns the connection corresponding to the specified ID.
// Returns NotFoundError if no connection is found.
func ConnectionByID(conn *directconnect.DirectConnect, connID string) (*directconnect.Connection, error) {
	input := &directconnect.DescribeConnectionsInput{
		ConnectionId: aws.String(connID),
	}

	connections, err := Connections(conn, input)
	if len(connections) == 0 {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	// Handle any empty result
	if connections == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	connection := connections[0]

	return connection, nil
}

// Connections returns the connections corresponding to the specified input.
// Returns an empty slice if no APIs are found.
func Connections(conn *directconnect.DirectConnect, input *directconnect.DescribeConnectionsInput) ([]*directconnect.Connection, error) {
	output, err := conn.DescribeConnections(input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output.Connections, nil
}
