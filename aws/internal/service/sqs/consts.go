package sqs

const (
	ErrCodeInvalidAction = "InvalidAction"
)

const (
	FifoQueueNameSuffix = ".fifo"
)

const (
	DefaultQueueDelaySeconds                  = 0
	DefaultQueueKmsDataKeyReusePeriodSeconds  = 300
	DefaultQueueMaximumMessageSize            = 262_144 // 256 KiB.
	DefaultQueueMessageRetentionPeriod        = 345_600 // 4 days.
	DefaultQueueReceiveMessageWaitTimeSeconds = 0
	DefaultQueueVisibilityTimeout             = 30
)

const (
	DeduplicationScopeMessageGroup = "messageGroup"
	DeduplicationScopeQueue        = "queue"
)

func DeduplicationScope_Values() []string {
	return []string{
		DeduplicationScopeMessageGroup,
		DeduplicationScopeQueue,
	}
}

const (
	FifoThroughputLimitPerMessageGroupId = "perMessageGroupId"
	FifoThroughputLimitPerQueue          = "perQueue"
)

func FifoThroughputLimit_Values() []string {
	return []string{
		FifoThroughputLimitPerMessageGroupId,
		FifoThroughputLimitPerQueue,
	}
}
