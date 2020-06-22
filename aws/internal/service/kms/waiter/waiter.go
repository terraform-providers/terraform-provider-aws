package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const (
	// Maximum amount of time to wait for KeyState to return PendingDeletion
	KeyStatePendingDeletionTimeout = 20 * time.Minute
)

// KeyStatePendingDeletion waits for KeyState to return PendingDeletion
func KeyStatePendingDeletion(conn *kms.KMS, keyID string) (*kms.DescribeKeyOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			kms.KeyStateDisabled,
			kms.KeyStateEnabled,
		},
		Target:  []string{kms.KeyStatePendingDeletion},
		Refresh: KeyState(conn, keyID),
		Timeout: KeyStatePendingDeletionTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*kms.DescribeKeyOutput); ok {
		return output, err
	}

	return nil, err
}
