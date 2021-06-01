package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	StorageGatewayGatewayConnectedMinTimeout                = 10 * time.Second
	StorageGatewayGatewayConnectedContinuousTargetOccurence = 6
	StorageGatewayGatewayJoinDomainJoinedTimeout            = 5 * time.Minute
	StoredIscsiVolumeAvailableTimeout                       = 5 * time.Minute
	NfsFileShareAvailableDelay                              = 5 * time.Second
	NfsFileShareDeletedDelay                                = 5 * time.Second
	SmbFileShareAvailableDelay                              = 5 * time.Second
	SmbFileShareDeletedDelay                                = 5 * time.Second
)

func StorageGatewayGatewayConnected(conn *storagegateway.StorageGateway, gatewayARN string, timeout time.Duration) (*storagegateway.DescribeGatewayInformationOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{storagegateway.ErrorCodeGatewayNotConnected},
		Target:                    []string{StorageGatewayGatewayStatusConnected},
		Refresh:                   StorageGatewayGatewayStatus(conn, gatewayARN),
		Timeout:                   timeout,
		MinTimeout:                StorageGatewayGatewayConnectedMinTimeout,
		ContinuousTargetOccurence: StorageGatewayGatewayConnectedContinuousTargetOccurence, // Gateway activations can take a few seconds and can trigger a reboot of the Gateway
	}

	outputRaw, err := stateConf.WaitForState()

	switch output := outputRaw.(type) {
	case *storagegateway.DescribeGatewayInformationOutput:
		return output, err
	default:
		return nil, err
	}
}

func StorageGatewayGatewayJoinDomainJoined(conn *storagegateway.StorageGateway, volumeARN string) (*storagegateway.DescribeSMBSettingsOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{storagegateway.ActiveDirectoryStatusJoining},
		Target:  []string{storagegateway.ActiveDirectoryStatusJoined},
		Refresh: StorageGatewayGatewayJoinDomainStatus(conn, volumeARN),
		Timeout: StorageGatewayGatewayJoinDomainJoinedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*storagegateway.DescribeSMBSettingsOutput); ok {
		return output, err
	}

	return nil, err
}

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

// NfsFileShareAvailable waits for a NFS File Share to return Available
func NfsFileShareAvailable(conn *storagegateway.StorageGateway, fileShareArn string, timeout time.Duration) (*storagegateway.NFSFileShareInfo, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"BOOTSTRAPPING", "CREATING", "RESTORING", "UPDATING"},
		Target:  []string{"AVAILABLE"},
		Refresh: NfsFileShareStatus(conn, fileShareArn),
		Timeout: timeout,
		Delay:   NfsFileShareAvailableDelay,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*storagegateway.NFSFileShareInfo); ok {
		return output, err
	}

	return nil, err
}

func NfsFileShareDeleted(conn *storagegateway.StorageGateway, fileShareArn string, timeout time.Duration) (*storagegateway.NFSFileShareInfo, error) {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{"AVAILABLE", "DELETING", "FORCE_DELETING"},
		Target:         []string{},
		Refresh:        NfsFileShareStatus(conn, fileShareArn),
		Timeout:        timeout,
		Delay:          NfsFileShareDeletedDelay,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*storagegateway.NFSFileShareInfo); ok {
		return output, err
	}

	return nil, err
}

// SmbFileShareAvailable waits for a SMB File Share to return Available
func SmbFileShareAvailable(conn *storagegateway.StorageGateway, fileShareArn string, timeout time.Duration) (*storagegateway.SMBFileShareInfo, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"CREATING", "UPDATING"},
		Target:  []string{"AVAILABLE"},
		Refresh: SmbFileShareStatus(conn, fileShareArn),
		Timeout: timeout,
		Delay:   SmbFileShareAvailableDelay,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*storagegateway.SMBFileShareInfo); ok {
		return output, err
	}

	return nil, err
}

func SmbFileShareDeleted(conn *storagegateway.StorageGateway, fileShareArn string, timeout time.Duration) (*storagegateway.SMBFileShareInfo, error) {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{"AVAILABLE", "DELETING", "FORCE_DELETING"},
		Target:         []string{},
		Refresh:        SmbFileShareStatus(conn, fileShareArn),
		Timeout:        timeout,
		Delay:          SmbFileShareDeletedDelay,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*storagegateway.SMBFileShareInfo); ok {
		return output, err
	}

	return nil, err
}
