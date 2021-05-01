package waiter

import (
	"time"
)

const (
	// Maximum amount of time to wait for Appmesh changes to propagate
	PropagationTimeout = 2 * time.Minute
)
