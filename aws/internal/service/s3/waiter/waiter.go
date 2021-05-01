package waiter

import (
	"time"
)

const (
	// Maximum amount of time to wait for S3 changes to propagate
	PropagationTimeout = 1 * time.Minute
)
