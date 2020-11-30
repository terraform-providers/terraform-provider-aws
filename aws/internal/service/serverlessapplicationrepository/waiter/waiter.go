package waiter

import (
	"time"
)

const (
	// Default maximum amount of time to wait for a Stack to be Created
	CloudFormationStackCreatedDefaultTimeout = 30 * time.Minute

	// Default maximum amount of time to wait for a Stack to be Updated
	CloudFormationStackUpdatedDefaultTimeout = 30 * time.Minute

	// Default maximum amount of time to wait for a Stack to be Deleted
	CloudFormationStackDeletedDefaultTimeout = 30 * time.Minute
)
