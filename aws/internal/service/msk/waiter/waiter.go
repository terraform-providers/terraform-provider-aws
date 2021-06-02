package waiter

import (
	"time"
)

const (
	ClusterCreateTimeout = 120 * time.Minute
	ClusterUpdateTimeout = 120 * time.Minute
	ClusterDeleteTimeout = 120 * time.Minute
)
