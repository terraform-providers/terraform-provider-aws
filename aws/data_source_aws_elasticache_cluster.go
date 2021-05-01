package aws

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/elasticache/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func dataSourceAwsElastiCacheCluster() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsElastiCacheClusterRead,

		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				StateFunc: func(v interface{}) string {
					value := v.(string)
					return strings.ToLower(value)
				},
			},

			"node_type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"num_cache_nodes": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"subnet_group_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"engine": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"engine_version": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"parameter_group_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"replication_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"security_group_names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"security_group_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"maintenance_window": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"snapshot_window": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"snapshot_retention_limit": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"availability_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"notification_topic_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"port": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"configuration_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"cluster_address": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"cache_nodes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"port": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"availability_zone": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsElastiCacheClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	clusterID := d.Get("cluster_id").(string)
	cluster, err := finder.CacheClusterWithNodeInfoByID(conn, clusterID)
	if tfresource.NotFound(err) {
		return fmt.Errorf("Your query returned no results. Please change your search criteria and try again")
	}
	if err != nil {
		return fmt.Errorf("error reading ElastiCache Cache Cluster (%s): %w", clusterID, err)
	}

	d.SetId(aws.StringValue(cluster.CacheClusterId))

	d.Set("cluster_id", cluster.CacheClusterId)
	d.Set("node_type", cluster.CacheNodeType)
	d.Set("num_cache_nodes", cluster.NumCacheNodes)
	d.Set("subnet_group_name", cluster.CacheSubnetGroupName)
	d.Set("engine", cluster.Engine)
	d.Set("engine_version", cluster.EngineVersion)
	d.Set("security_group_names", flattenElastiCacheSecurityGroupNames(cluster.CacheSecurityGroups))
	d.Set("security_group_ids", flattenElastiCacheSecurityGroupIds(cluster.SecurityGroups))

	if cluster.CacheParameterGroup != nil {
		d.Set("parameter_group_name", cluster.CacheParameterGroup.CacheParameterGroupName)
	}

	if cluster.ReplicationGroupId != nil {
		d.Set("replication_group_id", cluster.ReplicationGroupId)
	}

	d.Set("maintenance_window", cluster.PreferredMaintenanceWindow)
	d.Set("snapshot_window", cluster.SnapshotWindow)
	d.Set("snapshot_retention_limit", cluster.SnapshotRetentionLimit)
	d.Set("availability_zone", cluster.PreferredAvailabilityZone)

	if cluster.NotificationConfiguration != nil {
		if aws.StringValue(cluster.NotificationConfiguration.TopicStatus) == "active" {
			d.Set("notification_topic_arn", cluster.NotificationConfiguration.TopicArn)
		}
	}

	if cluster.ConfigurationEndpoint != nil {
		d.Set("port", cluster.ConfigurationEndpoint.Port)
		d.Set("configuration_endpoint", aws.String(fmt.Sprintf("%s:%d", *cluster.ConfigurationEndpoint.Address, *cluster.ConfigurationEndpoint.Port)))
		d.Set("cluster_address", aws.String(*cluster.ConfigurationEndpoint.Address))
	}

	if err := setCacheNodeData(d, cluster); err != nil {
		return err
	}

	d.Set("arn", cluster.ARN)

	tags, err := keyvaluetags.ElasticacheListTags(conn, aws.StringValue(cluster.ARN))

	if err != nil {
		return fmt.Errorf("error listing tags for Elasticache Cluster (%s): %w", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil

}
