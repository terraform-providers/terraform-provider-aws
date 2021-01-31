package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/elasticache/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

const (
	ReplicationGroupStatusCreating     = "creating"
	ReplicationGroupStatusAvailable    = "available"
	ReplicationGroupStatusModifying    = "modifying"
	ReplicationGroupStatusDeleting     = "deleting"
	ReplicationGroupStatusCreateFailed = "create-failed"
	ReplicationGroupStatusSnapshotting = "snapshotting"
)

// ReplicationGroupStatus fetches the ReplicationGroup and its Status
func ReplicationGroupStatus(conn *elasticache.ElastiCache, replicationGroupID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		rg, err := finder.ReplicationGroupByID(conn, replicationGroupID)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return rg, aws.StringValue(rg.Status), nil
	}
}

// ReplicationGroupMemberClustersStatus fetches the ReplicationGroup's Member Clusters and either "available" or the first non-"available" status.
// NOTE: This function assumes that the intended end-state is to have all member clusters in "available" status.
func ReplicationGroupMemberClustersStatus(conn *elasticache.ElastiCache, replicationGroupID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		clusters, err := finder.ReplicationGroupMemberClustersByID(conn, replicationGroupID)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		status := CacheClusterStatusAvailable
		for _, v := range clusters {
			clusterStatus := aws.StringValue(v.CacheClusterStatus)
			if clusterStatus != CacheClusterStatusAvailable {
				status = clusterStatus
				break
			}
		}
		return clusters, status, nil
	}
}

const (
	CacheClusterStatusAvailable             = "available"
	CacheClusterStatusCreating              = "creating"
	CacheClusterStatusDeleted               = "deleted"
	CacheClusterStatusDeleting              = "deleting"
	CacheClusterStatusIncompatibleNetwork   = "incompatible-network"
	CacheClusterStatusModifying             = "modifying"
	CacheClusterStatusRebootingClusterNodes = "rebooting cluster nodes"
	CacheClusterStatusRestoreFailed         = "restore-failed"
	CacheClusterStatusSnapshotting          = "snapshotting"
)

// CacheClusterStatus fetches the CacheCluster and its Status
func CacheClusterStatus(conn *elasticache.ElastiCache, cacheClusterID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		c, err := finder.CacheClusterByID(conn, cacheClusterID)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return c, aws.StringValue(c.CacheClusterStatus), nil
	}
}
