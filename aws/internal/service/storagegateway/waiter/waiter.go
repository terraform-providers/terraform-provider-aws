package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	StoredIscsiVolumeAvailableTimeout = 5 * time.Minute
)

// StoredIscsiVolumeAvailable waits for a StoredIscsiVolume to return Available
func StoredIscsiVolumeAvailable(conn *storagegateway.StorageGateway, volumeARN string) (*storagegateway.DescribeStorediSCSIVolumesOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"BOOTSTRAPPING", "CREATING", "RESTORING"},
		Target:  []string{"AVAILABLE"},
		Refresh: StoredIscsiVolumeStatus(conn, volumeARN),
		Timeout: StoredIscsiVolumeAvailableTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*storagegateway.DescribeStorediSCSIVolumesOutput); ok {
		return output, err
	}

	return nil, err
}
