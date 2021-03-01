package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// LocationByCode returns the locations corresponding to the specified location code.
// Returns NotFoundError if no location is found.
func LocationByCode(conn *directconnect.DirectConnect, locationCode string) (*directconnect.Location, error) {
	input := &directconnect.DescribeLocationsInput{}

	output, err := conn.DescribeLocations(input)

	if err != nil {
		return nil, err
	}

	// Handle any empty result.
	if output == nil || len(output.Locations) == 0 {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	for _, location := range output.Locations {
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
