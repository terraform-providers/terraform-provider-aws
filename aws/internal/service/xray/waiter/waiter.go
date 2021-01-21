package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/xray"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	EncryptionConfigAvailableTimeout = 15 * time.Minute
)

// EncryptionConfigAvailable waits for a EncryptionConfig to return Available
func EncryptionConfigAvailable(conn *xray.XRay) (*xray.EncryptionConfig, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{xray.EncryptionStatusUpdating},
		Target:  []string{xray.EncryptionStatusActive},
		Refresh: EncryptionConfigStatus(conn),
		Timeout: EncryptionConfigAvailableTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*xray.EncryptionConfig); ok {
		return v, err
	}

	return nil, err
}
