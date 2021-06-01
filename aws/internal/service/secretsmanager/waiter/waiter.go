package waiter

import (
	"time"
)

const (
	// Maximum amount of time to wait for Secrets Manager deletions to propagate
	DeletionPropagationTimeout = 2 * time.Minute
)
