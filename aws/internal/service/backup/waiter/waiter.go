package waiter

import (
	"time"
)

const (
	// Maximum amount of time to wait for Backup changes to propagate
	PropagationTimeout = 2 * time.Minute
)
