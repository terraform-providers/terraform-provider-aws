package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// ConnectionByID returns the connections corresponding to the specified connection ID.
// Returns NotFoundError if no connection is found.
func ConnectionByID(conn *directconnect.DirectConnect, connectionID string) (*directconnect.Connection, error) {
	input := &directconnect.DescribeConnectionsInput{
		ConnectionId: aws.String(connectionID),
	}

	connections, err := Connections(conn, input)

	if tfawserr.ErrMessageContains(err, directconnect.ErrCodeClientException, "Could not find Connection with ID") {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	// Handle any empty result.
	if len(connections) == 0 {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	if state := aws.StringValue(connections[0].ConnectionState); state == directconnect.ConnectionStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	return connections[0], nil
}

// Connections returns the connections corresponding to the specified input.
// Returns an empty slice if no connections are found.
func Connections(conn *directconnect.DirectConnect, input *directconnect.DescribeConnectionsInput) ([]*directconnect.Connection, error) {
	output, err := conn.DescribeConnections(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return []*directconnect.Connection{}, nil
	}

	return output.Connections, nil
}

// LagByID returns the locations corresponding to the specified LAG ID.
// Returns NotFoundError if no LAG is found.
func LagByID(conn *directconnect.DirectConnect, lagID string) (*directconnect.Lag, error) {
	input := &directconnect.DescribeLagsInput{
		LagId: aws.String(lagID),
	}

	lags, err := Lags(conn, input)

	if tfawserr.ErrMessageContains(err, directconnect.ErrCodeClientException, "Could not find Lag with ID") {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	// Handle any empty result.
	if len(lags) == 0 {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	if state := aws.StringValue(lags[0].LagState); state == directconnect.LagStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	return lags[0], nil
}

// Lags returns the LAGs corresponding to the specified input.
// Returns an empty slice if no LAGs are found.
func Lags(conn *directconnect.DirectConnect, input *directconnect.DescribeLagsInput) ([]*directconnect.Lag, error) {
	output, err := conn.DescribeLags(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return []*directconnect.Lag{}, nil
	}

	return output.Lags, nil
}

// LocationByCode returns the locations corresponding to the specified location code.
// Returns NotFoundError if no location is found.
func LocationByCode(conn *directconnect.DirectConnect, locationCode string) (*directconnect.Location, error) {
	input := &directconnect.DescribeLocationsInput{}

	locations, err := Locations(conn, input)

	if err != nil {
		return nil, err
	}

	for _, location := range locations {
		if aws.StringValue(location.LocationCode) == locationCode {
			return location, nil
		}
	}

	return nil, &resource.NotFoundError{
		Message:     "Empty result",
		LastRequest: input,
	}
}

// Locations returns the locations corresponding to the specified input.
// Returns an empty slice if no locations are found.
func Locations(conn *directconnect.DirectConnect, input *directconnect.DescribeLocationsInput) ([]*directconnect.Location, error) {
	output, err := conn.DescribeLocations(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return []*directconnect.Location{}, nil
	}

	return output.Locations, nil
}
