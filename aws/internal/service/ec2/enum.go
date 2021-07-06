package ec2

const (
	// https://docs.aws.amazon.com/vpc/latest/privatelink/vpce-interface.html#vpce-interface-lifecycle
	VpcEndpointStateAvailable         = "available"
	VpcEndpointStateDeleted           = "deleted"
	VpcEndpointStateDeleting          = "deleting"
	VpcEndpointStateFailed            = "failed"
	VpcEndpointStatePending           = "pending"
	VpcEndpointStatePendingAcceptance = "pendingAcceptance"
	VpcEndpointStateRejected          = "rejected"
)
