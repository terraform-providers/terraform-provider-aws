package waiter

import (
	"time"
)

const (
	// Maximum amount of time to wait for IAM changes to propagate
	// This timeout should not be increased without strong consideration
	// as this will negatively impact user experience when configurations
	// have incorrect references or permissions.
	// Reference: https://docs.aws.amazon.com/IAM/latest/UserGuide/troubleshoot_general.html#troubleshoot_general_eventual-consistency
	PropagationTimeout = 2 * time.Minute
)
