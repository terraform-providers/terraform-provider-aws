package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// ImageStatusAvailable waits for an Image to return Available
func ImageStatusAvailable(conn *imagebuilder.Imagebuilder, imageBuildVersionArn string, timeout time.Duration) (*imagebuilder.Image, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			imagebuilder.ImageStatusBuilding,
			imagebuilder.ImageStatusCreating,
			imagebuilder.ImageStatusDistributing,
			imagebuilder.ImageStatusIntegrating,
			imagebuilder.ImageStatusPending,
			imagebuilder.ImageStatusTesting,
		},
		Target:  []string{imagebuilder.ImageStatusAvailable},
		Refresh: ImageStatus(conn, imageBuildVersionArn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*imagebuilder.Image); ok {
		return v, err
	}

	return nil, err
}
