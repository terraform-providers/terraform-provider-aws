package finder

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// ReplicationGroupByID retrieves an ElastiCache Replication Group by id.
func ReplicationGroupByID(conn *elasticache.ElastiCache, id string) (*elasticache.ReplicationGroup, error) {
	input := &elasticache.DescribeReplicationGroupsInput{
		ReplicationGroupId: aws.String(id),
	}
	output, err := conn.DescribeReplicationGroups(input)
	if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeReplicationGroupNotFoundFault) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if output == nil || len(output.ReplicationGroups) == 0 || output.ReplicationGroups[0] == nil {
		return nil, &resource.NotFoundError{
			Message:     "empty result",
			LastRequest: input,
		}
	}

	return output.ReplicationGroups[0], nil
}

// ReplicationGroupMemberClustersByID retrieves all of an ElastiCache Replication Group's MemberClusters by the id of the Replication Group.
func ReplicationGroupMemberClustersByID(conn *elasticache.ElastiCache, id string) ([]*elasticache.CacheCluster, error) {
	rg, err := ReplicationGroupByID(conn, id)
	if err != nil {
		return nil, err
	}

	clusters, err := CacheClustersByID(conn, aws.StringValueSlice(rg.MemberClusters))
	if err != nil {
		return clusters, err
	}
	if len(clusters) == 0 {
		return clusters, &resource.NotFoundError{
			Message: fmt.Sprintf("No Member Clusters found in Replication Group (%s)", id),
		}
	}

	return clusters, nil
}

// CacheClusterByID retrieves an ElastiCache Cache Cluster by id.
func CacheClusterByID(conn *elasticache.ElastiCache, id string) (*elasticache.CacheCluster, error) {
	input := &elasticache.DescribeCacheClustersInput{
		CacheClusterId: aws.String(id),
	}
	return CacheCluster(conn, input)
}

// CacheClusterWithNodeInfoByID retrieves an ElastiCache Cache Cluster with Node Info by id.
func CacheClusterWithNodeInfoByID(conn *elasticache.ElastiCache, id string) (*elasticache.CacheCluster, error) {
	input := &elasticache.DescribeCacheClustersInput{
		CacheClusterId:    aws.String(id),
		ShowCacheNodeInfo: aws.Bool(true),
	}
	return CacheCluster(conn, input)
}

// CacheCluster retrieves an ElastiCache Cache Cluster using DescribeCacheClustersInput.
func CacheCluster(conn *elasticache.ElastiCache, input *elasticache.DescribeCacheClustersInput) (*elasticache.CacheCluster, error) {
	result, err := conn.DescribeCacheClusters(input)
	if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeCacheClusterNotFoundFault) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if result == nil || len(result.CacheClusters) == 0 || result.CacheClusters[0] == nil {
		return nil, &resource.NotFoundError{
			Message:     "empty result",
			LastRequest: input,
		}
	}

	return result.CacheClusters[0], nil
}

// CacheClustersByID retrieves a list of ElastiCache Cache Clusters by id.
// Order of the clusters is not guaranteed.
func CacheClustersByID(conn *elasticache.ElastiCache, idList []string) ([]*elasticache.CacheCluster, error) {
	var results []*elasticache.CacheCluster
	ids := make(map[string]bool)
	for _, v := range idList {
		ids[v] = true
	}

	input := &elasticache.DescribeCacheClustersInput{}
	err := conn.DescribeCacheClustersPages(input, func(page *elasticache.DescribeCacheClustersOutput, _ bool) bool {
		if page == nil || page.CacheClusters == nil || len(page.CacheClusters) == 0 {
			return true
		}

		for _, v := range page.CacheClusters {
			if ids[aws.StringValue(v.CacheClusterId)] {
				results = append(results, v)
				delete(ids, aws.StringValue(v.CacheClusterId))
				if len(ids) == 0 {
					break
				}
			}
		}

		return len(ids) != 0
	})

	return results, err
}

// GlobalReplicationGroupByID() retrieves an ElastiCache Global Replication Group by id.
func GlobalReplicationGroupByID(conn *elasticache.ElastiCache, id string) (*elasticache.GlobalReplicationGroup, error) {
	input := &elasticache.DescribeGlobalReplicationGroupsInput{
		GlobalReplicationGroupId: aws.String(id),
		ShowMemberInfo:           aws.Bool(true),
	}
	output, err := conn.DescribeGlobalReplicationGroups(input)
	if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeGlobalReplicationGroupNotFoundFault) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if output == nil || len(output.GlobalReplicationGroups) == 0 || output.GlobalReplicationGroups[0] == nil {
		return nil, &resource.NotFoundError{
			Message:     "empty result",
			LastRequest: input,
		}
	}

	return output.GlobalReplicationGroups[0], nil
}

// GlobalReplicationGroupMemberByID retrieves a member Replication Group by id from a Global Replication Group.
func GlobalReplicationGroupMemberByID(conn *elasticache.ElastiCache, globalReplicationGroupID string, id string) (*elasticache.GlobalReplicationGroupMember, error) {
	globalReplicationGroup, err := GlobalReplicationGroupByID(conn, globalReplicationGroupID)
	if err != nil {
		return nil, &resource.NotFoundError{
			Message:   "unable to retrieve enclosing Global Replication Group",
			LastError: err,
		}
	}

	if globalReplicationGroup == nil || len(globalReplicationGroup.Members) == 0 {
		return nil, &resource.NotFoundError{
			Message: "empty result",
		}
	}

	for _, member := range globalReplicationGroup.Members {
		if aws.StringValue(member.ReplicationGroupId) == id {
			return member, nil
		}
	}

	return nil, &resource.NotFoundError{
		Message: fmt.Sprintf("Replication Group (%s) not found in Global Replication Group (%s)", id, globalReplicationGroupID),
	}
}
