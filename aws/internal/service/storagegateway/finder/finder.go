package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
)

func LocalDiskByDiskId(conn *storagegateway.StorageGateway, gatewayARN string, diskID string) (*storagegateway.Disk, error) {
	input := &storagegateway.ListLocalDisksInput{
		GatewayARN: aws.String(gatewayARN),
	}

	output, err := conn.ListLocalDisks(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	for _, disk := range output.Disks {
		if aws.StringValue(disk.DiskId) == diskID {
			return disk, nil
		}
	}

	return nil, nil
}

func LocalDiskByDiskPath(conn *storagegateway.StorageGateway, gatewayARN string, diskPath string) (*storagegateway.Disk, error) {
	input := &storagegateway.ListLocalDisksInput{
		GatewayARN: aws.String(gatewayARN),
	}

	output, err := conn.ListLocalDisks(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	for _, disk := range output.Disks {
		if aws.StringValue(disk.DiskPath) == diskPath {
			return disk, nil
		}
	}

	return nil, nil
}

func UploadBufferDisk(conn *storagegateway.StorageGateway, gatewayARN string, diskID string) (*string, error) {
	input := &storagegateway.DescribeUploadBufferInput{
		GatewayARN: aws.String(gatewayARN),
	}

	var result *string

	output, err := conn.DescribeUploadBuffer(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	for _, diskId := range output.DiskIds {
		if aws.StringValue(diskId) == diskID {
			result = diskId
			break
		}
	}

	return result, err
}
