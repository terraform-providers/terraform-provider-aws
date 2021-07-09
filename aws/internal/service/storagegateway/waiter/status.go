package waiter

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	StorageGatewayGatewayStatusConnected = "GatewayConnected"
	StoredIscsiVolumeStatusNotFound      = "NotFound"
	StoredIscsiVolumeStatusUnknown       = "Unknown"
	NfsFileShareStatusNotFound           = "NotFound"
	NfsFileShareStatusUnknown            = "Unknown"
	SmbFileShareStatusNotFound           = "NotFound"
	SmbFileShareStatusUnknown            = "Unknown"
	FsxFileSystemStatusNotFound          = "NotFound"
	FsxFileSystemStatusUnknown           = "Unknown"
)

func StorageGatewayGatewayStatus(conn *storagegateway.StorageGateway, gatewayARN string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &storagegateway.DescribeGatewayInformationInput{
			GatewayARN: aws.String(gatewayARN),
		}

		output, err := conn.DescribeGatewayInformation(input)

		if tfawserr.ErrMessageContains(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified gateway is not connected") {
			return output, storagegateway.ErrorCodeGatewayNotConnected, nil
		}

		if err != nil {
			return output, "", err
		}

		return output, StorageGatewayGatewayStatusConnected, nil
	}
}

func StorageGatewayGatewayJoinDomainStatus(conn *storagegateway.StorageGateway, gatewayARN string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &storagegateway.DescribeSMBSettingsInput{
			GatewayARN: aws.String(gatewayARN),
		}

		output, err := conn.DescribeSMBSettings(input)

		log.Printf("[DEBUG] Storage Gateway Gateway Join Domain Status: %s", *output.ActiveDirectoryStatus)

		if tfawserr.ErrMessageContains(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified gateway is not connected") {
			return output, storagegateway.ActiveDirectoryStatusUnknownError, nil
		}

		if err != nil {
			return output, storagegateway.ActiveDirectoryStatusUnknownError, err
		}

		return output, aws.StringValue(output.ActiveDirectoryStatus), nil
	}
}

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

func NfsFileShareStatus(conn *storagegateway.StorageGateway, fileShareArn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &storagegateway.DescribeNFSFileSharesInput{
			FileShareARNList: []*string{aws.String(fileShareArn)},
		}

		log.Printf("[DEBUG] Reading Storage Gateway NFS File Share: %s", input)
		output, err := conn.DescribeNFSFileShares(input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified file share was not found.") {
				return nil, NfsFileShareStatusNotFound, nil
			}
			return nil, NfsFileShareStatusUnknown, fmt.Errorf("error reading Storage Gateway NFS File Share: %w", err)
		}

		if output == nil || len(output.NFSFileShareInfoList) == 0 || output.NFSFileShareInfoList[0] == nil {
			return nil, NfsFileShareStatusNotFound, nil
		}

		fileshare := output.NFSFileShareInfoList[0]

		return fileshare, aws.StringValue(fileshare.FileShareStatus), nil
	}
}

func SmbFileShareStatus(conn *storagegateway.StorageGateway, fileShareArn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &storagegateway.DescribeSMBFileSharesInput{
			FileShareARNList: []*string{aws.String(fileShareArn)},
		}

		log.Printf("[DEBUG] Reading Storage Gateway SMB File Share: %s", input)
		output, err := conn.DescribeSMBFileShares(input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified file share was not found.") {
				return nil, SmbFileShareStatusNotFound, nil
			}
			return nil, SmbFileShareStatusUnknown, fmt.Errorf("error reading Storage Gateway SMB File Share: %w", err)
		}

		if output == nil || len(output.SMBFileShareInfoList) == 0 || output.SMBFileShareInfoList[0] == nil {
			return nil, SmbFileShareStatusNotFound, nil
		}

		fileshare := output.SMBFileShareInfoList[0]

		return fileshare, aws.StringValue(fileshare.FileShareStatus), nil
	}
}

func FsxFileSystemStatus(conn *storagegateway.StorageGateway, fileSystemArn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &storagegateway.DescribeFileSystemAssociationsInput{
			FileSystemAssociationARNList: []*string{aws.String(fileSystemArn)},
		}

		log.Printf("[DEBUG] Reading Storage Gateway FSx File System: %s", input)
		output, err := conn.DescribeFileSystemAssociations(input)
		if err != nil {
			// currently verbose, can update for clarity pending: https://github.com/hashicorp/aws-sdk-go-base/issues/59
			if tfawserr.ErrCodeEquals(err, storagegateway.ErrCodeInvalidGatewayRequestException) {

				var igrex *storagegateway.InvalidGatewayRequestException
				if ok := errors.As(err, &igrex); ok {
					if err := igrex.Error_; err != nil {
						if aws.StringValue(err.ErrorCode) == "FileSystemAssociationNotFound" {
							return nil, FsxFileSystemStatusNotFound, nil
						}
					}
				}
			}
			return nil, FsxFileSystemStatusUnknown, fmt.Errorf("error reading Storage Gateway FSx File System: %w", err)
		}

		if output == nil || len(output.FileSystemAssociationInfoList) == 0 || output.FileSystemAssociationInfoList[0] == nil {
			return nil, FsxFileSystemStatusNotFound, nil
		}

		filesystem := output.FileSystemAssociationInfoList[0]

		return filesystem, aws.StringValue(filesystem.FileSystemAssociationStatus), nil
	}
}
