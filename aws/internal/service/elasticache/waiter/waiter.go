package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	ReplicationGroupDefaultCreatedTimeout = 60 * time.Minute
	ReplicationGroupDefaultUpdatedTimeout = 40 * time.Minute
	ReplicationGroupDefaultDeletedTimeout = 40 * time.Minute

	replicationGroupAvailableMinTimeout = 10 * time.Second
	replicationGroupAvailableDelay      = 30 * time.Second

	replicationGroupDeletedMinTimeout = 10 * time.Second
	replicationGroupDeletedDelay      = 30 * time.Second
)

// ReplicationGroupAvailable waits for a ReplicationGroup to return Available
func ReplicationGroupAvailable(conn *elasticache.ElastiCache, replicationGroupID string, timeout time.Duration) (*elasticache.ReplicationGroup, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			ReplicationGroupStatusCreating,
			ReplicationGroupStatusModifying,
			ReplicationGroupStatusSnapshotting,
		},
		Target:     []string{ReplicationGroupStatusAvailable},
		Refresh:    ReplicationGroupStatus(conn, replicationGroupID),
		Timeout:    timeout,
		MinTimeout: replicationGroupAvailableMinTimeout,
		Delay:      replicationGroupAvailableDelay,
	}

	outputRaw, err := stateConf.WaitForState()
	if v, ok := outputRaw.(*elasticache.ReplicationGroup); ok {
		return v, err
	}
	return nil, err
}

// ReplicationGroupDeleted waits for a ReplicationGroup to be deleted
func ReplicationGroupDeleted(conn *elasticache.ElastiCache, replicationGroupID string, timeout time.Duration) (*elasticache.ReplicationGroup, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			ReplicationGroupStatusCreating,
			ReplicationGroupStatusAvailable,
			ReplicationGroupStatusDeleting,
		},
		Target:     []string{},
		Refresh:    ReplicationGroupStatus(conn, replicationGroupID),
		Timeout:    timeout,
		MinTimeout: replicationGroupDeletedMinTimeout,
		Delay:      replicationGroupDeletedDelay,
	}

	outputRaw, err := stateConf.WaitForState()
	if v, ok := outputRaw.(*elasticache.ReplicationGroup); ok {
		return v, err
	}
	return nil, err
}

// ReplicationGroupMemberClustersAvailable waits for all of a ReplicationGroup's Member Clusters to return Available
func ReplicationGroupMemberClustersAvailable(conn *elasticache.ElastiCache, replicationGroupID string, timeout time.Duration) ([]*elasticache.CacheCluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			CacheClusterStatusCreating,
			CacheClusterStatusDeleting,
			CacheClusterStatusModifying,
		},
		Target:     []string{CacheClusterStatusAvailable},
		Refresh:    ReplicationGroupMemberClustersStatus(conn, replicationGroupID),
		Timeout:    timeout,
		MinTimeout: cacheClusterAvailableMinTimeout,
		Delay:      cacheClusterAvailableDelay,
	}

	outputRaw, err := stateConf.WaitForState()
	if v, ok := outputRaw.([]*elasticache.CacheCluster); ok {
		return v, err
	}
	return nil, err
}

const (
	CacheClusterCreatedTimeout = 40 * time.Minute
	CacheClusterUpdatedTimeout = 80 * time.Minute
	CacheClusterDeletedTimeout = 40 * time.Minute

	cacheClusterAvailableMinTimeout = 10 * time.Second
	cacheClusterAvailableDelay      = 30 * time.Second

	cacheClusterDeletedMinTimeout = 10 * time.Second
	cacheClusterDeletedDelay      = 30 * time.Second
)

// CacheClusterAvailable waits for a Cache Cluster to return Available
func CacheClusterAvailable(conn *elasticache.ElastiCache, cacheClusterID string, timeout time.Duration) (*elasticache.CacheCluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			CacheClusterStatusCreating,
			CacheClusterStatusModifying,
			CacheClusterStatusSnapshotting,
			CacheClusterStatusRebootingClusterNodes,
		},
		Target:     []string{CacheClusterStatusAvailable},
		Refresh:    CacheClusterStatus(conn, cacheClusterID),
		Timeout:    timeout,
		MinTimeout: cacheClusterAvailableMinTimeout,
		Delay:      cacheClusterAvailableDelay,
	}

	outputRaw, err := stateConf.WaitForState()
	if v, ok := outputRaw.(*elasticache.CacheCluster); ok {
		return v, err
	}
	return nil, err
}

// CacheClusterDeleted waits for a Cache Cluster to be deleted
func CacheClusterDeleted(conn *elasticache.ElastiCache, cacheClusterID string, timeout time.Duration) (*elasticache.CacheCluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			CacheClusterStatusCreating,
			CacheClusterStatusAvailable,
			CacheClusterStatusModifying,
			CacheClusterStatusDeleting,
			CacheClusterStatusIncompatibleNetwork,
			CacheClusterStatusRestoreFailed,
			CacheClusterStatusSnapshotting,
		},
		Target:     []string{},
		Refresh:    CacheClusterStatus(conn, cacheClusterID),
		Timeout:    timeout,
		MinTimeout: cacheClusterDeletedMinTimeout,
		Delay:      cacheClusterDeletedDelay,
	}

	outputRaw, err := stateConf.WaitForState()
	if v, ok := outputRaw.(*elasticache.CacheCluster); ok {
		return v, err
	}
	return nil, err
}

const (
	GlobalReplicationGroupDefaultCreatedTimeout = 20 * time.Minute
	GlobalReplicationGroupDefaultUpdatedTimeout = 40 * time.Minute
	GlobalReplicationGroupDefaultDeletedTimeout = 20 * time.Minute

	globalReplicationGroupAvailableMinTimeout = 10 * time.Second
	globalReplicationGroupAvailableDelay      = 30 * time.Second

	globalReplicationGroupDeletedMinTimeout = 10 * time.Second
	globalReplicationGroupDeletedDelay      = 30 * time.Second
)

// GlobalReplicationGroupAvailable waits for a Global Replication Group to be available,
// with status either "available" or "primary-only"
func GlobalReplicationGroupAvailable(conn *elasticache.ElastiCache, globalReplicationGroupID string, timeout time.Duration) (*elasticache.GlobalReplicationGroup, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{GlobalReplicationGroupStatusCreating, GlobalReplicationGroupStatusModifying},
		Target:     []string{GlobalReplicationGroupStatusAvailable, GlobalReplicationGroupStatusPrimaryOnly},
		Refresh:    GlobalReplicationGroupStatus(conn, globalReplicationGroupID),
		Timeout:    timeout,
		MinTimeout: globalReplicationGroupAvailableMinTimeout,
		Delay:      globalReplicationGroupAvailableDelay,
	}

	outputRaw, err := stateConf.WaitForState()
	if v, ok := outputRaw.(*elasticache.GlobalReplicationGroup); ok {
		return v, err
	}
	return nil, err
}

// GlobalReplicationGroupDeleted waits for a Global Replication Group to be deleted
func GlobalReplicationGroupDeleted(conn *elasticache.ElastiCache, globalReplicationGroupID string) (*elasticache.GlobalReplicationGroup, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			GlobalReplicationGroupStatusAvailable,
			GlobalReplicationGroupStatusPrimaryOnly,
			GlobalReplicationGroupStatusModifying,
			GlobalReplicationGroupStatusDeleting,
		},
		Target:     []string{},
		Refresh:    GlobalReplicationGroupStatus(conn, globalReplicationGroupID),
		Timeout:    GlobalReplicationGroupDefaultDeletedTimeout,
		MinTimeout: globalReplicationGroupDeletedMinTimeout,
		Delay:      globalReplicationGroupDeletedDelay,
	}

	outputRaw, err := stateConf.WaitForState()
	if v, ok := outputRaw.(*elasticache.GlobalReplicationGroup); ok {
		return v, err
	}
	return nil, err
}

const (
	GlobalReplicationGroupDisassociationRetryTimeout = 45 * time.Minute

	globalReplicationGroupDisassociationTimeout = 20 * time.Minute

	globalReplicationGroupDisassociationMinTimeout = 10 * time.Second
	globalReplicationGroupDisassociationDelay      = 30 * time.Second
)

func GlobalReplicationGroupMemberDetached(conn *elasticache.ElastiCache, globalReplicationGroupID, id string) (*elasticache.GlobalReplicationGroupMember, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			GlobalReplicationGroupMemberStatusAssociated,
		},
		Target:     []string{},
		Refresh:    GlobalReplicationGroupMemberStatus(conn, globalReplicationGroupID, id),
		Timeout:    globalReplicationGroupDisassociationTimeout,
		MinTimeout: globalReplicationGroupDisassociationMinTimeout,
		Delay:      globalReplicationGroupDisassociationDelay,
	}

	outputRaw, err := stateConf.WaitForState()
	if v, ok := outputRaw.(*elasticache.GlobalReplicationGroupMember); ok {
		return v, err
	}
	return nil, err
}
