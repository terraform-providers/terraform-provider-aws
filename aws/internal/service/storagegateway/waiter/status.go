package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	StoredIscsiVolumeStatusNotFound = "NotFound"
	StoredIscsiVolumeStatusUnknown  = "Unknown"
)

// StoredIscsiVolumeStatus fetches the Volume and its Status
func StoredIscsiVolumeStatus(conn *storagegateway.StorageGateway, volumeARN string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &storagegateway.DescribeStorediSCSIVolumesInput{
			VolumeARNs: []*string{aws.String(volumeARN)},
		}

		output, err := conn.DescribeStorediSCSIVolumes(input)

		if tfawserr.ErrCodeEquals(err, storagegateway.ErrorCodeVolumeNotFound) ||
			tfawserr.ErrMessageContains(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified volume was not found") {
			return nil, StoredIscsiVolumeStatusNotFound, nil
		}

		if err != nil {
			return nil, StoredIscsiVolumeStatusUnknown, err
		}

		if output == nil || len(output.StorediSCSIVolumes) == 0 {
			return nil, StoredIscsiVolumeStatusNotFound, nil
		}

		return output, aws.StringValue(output.StorediSCSIVolumes[0].VolumeStatus), nil
	}
}
