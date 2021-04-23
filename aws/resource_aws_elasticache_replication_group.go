package aws

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/elasticache/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/elasticache/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsElasticacheReplicationGroup() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		Create: resourceAwsElasticacheReplicationGroupCreate,
		Read:   resourceAwsElasticacheReplicationGroupRead,
		Update: resourceAwsElasticacheReplicationGroupUpdate,
		Delete: resourceAwsElasticacheReplicationGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"apply_immediately": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"at_rest_encryption_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"auth_token": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ForceNew:     true,
				ValidateFunc: validateAwsElastiCacheReplicationGroupAuthToken,
			},
			"auto_minor_version_upgrade": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"automatic_failover_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"availability_zones": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"cluster_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"cluster_mode": {
				Type:         schema.TypeList,
				Optional:     true,
				Computed:     true,
				MaxItems:     1,
				ExactlyOneOf: []string{"cluster_mode", "number_cache_clusters"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"replicas_per_node_group": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"num_node_groups": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"configuration_endpoint_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "redis",
				ValidateFunc: validation.StringInSlice([]string{"redis"}, true),
			},
			"engine_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"global_replication_group_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
				ConflictsWith: []string{
					"automatic_failover_enabled",
					"cluster_mode", // should/will be "num_node_groups"
					"parameter_group_name",
					"engine",
					"engine_version",
					"node_type",
					"security_group_names",
					"transit_encryption_enabled",
					"at_rest_encryption_enabled",
					"snapshot_arns",
					"snapshot_name",
				},
			},
			"maintenance_window": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				StateFunc: func(val interface{}) string {
					// Elasticache always changes the maintenance to lowercase
					return strings.ToLower(val.(string))
				},
				ValidateFunc: validateOnceAWeekWindowFormat,
			},
			"member_clusters": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"multi_az_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"node_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"notification_topic_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},
			"number_cache_clusters": {
				Type:         schema.TypeInt,
				Computed:     true,
				Optional:     true,
				ExactlyOneOf: []string{"cluster_mode", "number_cache_clusters"},
			},
			"parameter_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// Suppress default Redis ports when not defined
					if !d.IsNewResource() && new == "0" && old == elasticacheDefaultRedisPort {
						return true
					}
					return false
				},
			},
			"primary_endpoint_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"reader_endpoint_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"replication_group_description": {
				Type:     schema.TypeString,
				Required: true,
			},
			"replication_group_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateReplicationGroupID,
				StateFunc: func(val interface{}) string {
					return strings.ToLower(val.(string))
				},
			},
			"security_group_names": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"snapshot_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				// Note: Unlike aws_elasticache_cluster, this does not have a limit of 1 item.
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.All(
						validateArn,
						validation.StringDoesNotContainAny(","),
					),
				},
				Set: schema.HashString,
			},
			"snapshot_retention_limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtMost(35),
			},
			"snapshot_window": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateOnceADayWindowFormat,
			},
			"snapshot_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"subnet_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"tags": tagsSchema(),
			"transit_encryption_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"final_snapshot_identifier": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
		SchemaVersion: 1,

		// SchemaVersion: 1 did not include any state changes via MigrateState.
		// Perform a no-operation state upgrade for Terraform 0.12 compatibility.
		// Future state migrations should be performed with StateUpgraders.
		MigrateState: func(v int, inst *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
			return inst, nil
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(waiter.ReplicationGroupDefaultCreatedTimeout),
			Delete: schema.DefaultTimeout(waiter.ReplicationGroupDefaultDeletedTimeout),
			Update: schema.DefaultTimeout(waiter.ReplicationGroupDefaultUpdatedTimeout),
		},

		CustomizeDiff: customdiff.Sequence(
			func(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
				if v := diff.Get("multi_az_enabled").(bool); !v {
					return nil
				}
				if v := diff.Get("automatic_failover_enabled").(bool); !v {
					return errors.New(`automatic_failover_enabled must be true if multi_az_enabled is true`)
				}
				return nil
			},
			func(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
				if v := diff.Get("automatic_failover_enabled").(bool); !v {
					return nil
				}

				if v, ok := diff.GetOkExists("number_cache_clusters"); ok {
					if v.(int) > 1 {
						return nil
					}
					return errors.New(`if automatic_failover_enabled is true, number_cache_clusters must be greater than 1`)
				}

				if v, ok := diff.GetOkExists("cluster_mode.0.replicas_per_node_group"); ok {
					if v.(int) > 0 {
						return nil
					}
					return errors.New(`if automatic_failover_enabled is true, cluster_mode[0].replicas_per_node_group must be greater than 0`)
				}

				return nil
			},
			customdiff.ComputedIf("member_clusters", func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				return diff.HasChange("number_cache_clusters") ||
					diff.HasChange("cluster_mode.0.num_node_groups") ||
					diff.HasChange("cluster_mode.0.replicas_per_node_group")
			}),
		),
	}
}

func resourceAwsElasticacheReplicationGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn

	tags := keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().ElasticacheTags()
	params := &elasticache.CreateReplicationGroupInput{
		ReplicationGroupId:          aws.String(d.Get("replication_group_id").(string)),
		ReplicationGroupDescription: aws.String(d.Get("replication_group_description").(string)),
		AutoMinorVersionUpgrade:     aws.Bool(d.Get("auto_minor_version_upgrade").(bool)),
		Tags:                        tags,
	}

	if v, ok := d.GetOk("global_replication_group_id"); ok {
		params.GlobalReplicationGroupId = aws.String(v.(string))
	} else {
		// This cannot be handled at plan-time
		nodeType := d.Get("node_type").(string)
		if nodeType == "" {
			return errors.New(`"node_type" is required unless "global_replication_group_id" is set.`)
		}
		params.AutomaticFailoverEnabled = aws.Bool(d.Get("automatic_failover_enabled").(bool))
		params.CacheNodeType = aws.String(nodeType)
		params.Engine = aws.String(d.Get("engine").(string))
	}

	if v, ok := d.GetOk("engine_version"); ok {
		params.EngineVersion = aws.String(v.(string))
	}

	if preferredAzs := d.Get("availability_zones").(*schema.Set); preferredAzs.Len() > 0 {
		params.PreferredCacheClusterAZs = expandStringSet(preferredAzs)
	}

	if v, ok := d.GetOk("parameter_group_name"); ok {
		params.CacheParameterGroupName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("port"); ok {
		params.Port = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("subnet_group_name"); ok {
		params.CacheSubnetGroupName = aws.String(v.(string))
	}

	if SGNames := d.Get("security_group_names").(*schema.Set); SGNames.Len() > 0 {
		params.CacheSecurityGroupNames = expandStringSet(SGNames)
	}

	if SGIds := d.Get("security_group_ids").(*schema.Set); SGIds.Len() > 0 {
		params.SecurityGroupIds = expandStringSet(SGIds)
	}

	if snaps := d.Get("snapshot_arns").(*schema.Set); snaps.Len() > 0 {
		params.SnapshotArns = expandStringSet(snaps)
	}

	if v, ok := d.GetOk("maintenance_window"); ok {
		params.PreferredMaintenanceWindow = aws.String(v.(string))
	}

	if _, ok := d.GetOk("multi_az_enabled"); ok {
		params.MultiAZEnabled = aws.Bool(d.Get("multi_az_enabled").(bool))
	}

	if v, ok := d.GetOk("notification_topic_arn"); ok {
		params.NotificationTopicArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		params.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("snapshot_retention_limit"); ok {
		params.SnapshotRetentionLimit = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("snapshot_window"); ok {
		params.SnapshotWindow = aws.String(v.(string))
	}

	if v, ok := d.GetOk("snapshot_name"); ok {
		params.SnapshotName = aws.String(v.(string))
	}

	if _, ok := d.GetOk("transit_encryption_enabled"); ok {
		params.TransitEncryptionEnabled = aws.Bool(d.Get("transit_encryption_enabled").(bool))
	}

	if _, ok := d.GetOk("at_rest_encryption_enabled"); ok {
		params.AtRestEncryptionEnabled = aws.Bool(d.Get("at_rest_encryption_enabled").(bool))
	}

	if v, ok := d.GetOk("auth_token"); ok {
		params.AuthToken = aws.String(v.(string))
	}

	if clusterMode, ok := d.GetOk("cluster_mode"); ok {
		clusterModeList := clusterMode.([]interface{})
		attributes := clusterModeList[0].(map[string]interface{})

		if v, ok := attributes["num_node_groups"]; ok {
			params.NumNodeGroups = aws.Int64(int64(v.(int)))
		}

		if v, ok := attributes["replicas_per_node_group"]; ok {
			params.ReplicasPerNodeGroup = aws.Int64(int64(v.(int)))
		}
	}

	if cacheClusters, ok := d.GetOk("number_cache_clusters"); ok {
		params.NumCacheClusters = aws.Int64(int64(cacheClusters.(int)))
	}
	resp, err := conn.CreateReplicationGroup(params)
	if err != nil {
		return fmt.Errorf("Error creating ElastiCache Replication Group (%s): %w", d.Get("replication_group_id").(string), err)
	}

	d.SetId(aws.StringValue(resp.ReplicationGroup.ReplicationGroupId))

	_, err = waiter.ReplicationGroupAvailable(conn, d.Id(), d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return fmt.Errorf("error creating ElastiCache Replication Group (%s): waiting for completion: %w", d.Id(), err)
	}

	return resourceAwsElasticacheReplicationGroupRead(d, meta)
}

func resourceAwsElasticacheReplicationGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	rgp, err := finder.ReplicationGroupByID(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ElastiCache Replication Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return err
	}

	if aws.StringValue(rgp.Status) == waiter.ReplicationGroupStatusDeleting {
		log.Printf("[WARN] ElastiCache Replication Group (%s) is currently in the `deleting` status, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if rgp.GlobalReplicationGroupInfo != nil && rgp.GlobalReplicationGroupInfo.GlobalReplicationGroupId != nil {
		d.Set("global_replication_group_id", rgp.GlobalReplicationGroupInfo.GlobalReplicationGroupId)
	}

	if rgp.AutomaticFailover != nil {
		switch strings.ToLower(aws.StringValue(rgp.AutomaticFailover)) {
		case elasticache.AutomaticFailoverStatusDisabled, elasticache.AutomaticFailoverStatusDisabling:
			d.Set("automatic_failover_enabled", false)
		case elasticache.AutomaticFailoverStatusEnabled, elasticache.AutomaticFailoverStatusEnabling:
			d.Set("automatic_failover_enabled", true)
		default:
			log.Printf("Unknown AutomaticFailover state %q", aws.StringValue(rgp.AutomaticFailover))
		}
	}

	if rgp.MultiAZ != nil {
		switch strings.ToLower(aws.StringValue(rgp.MultiAZ)) {
		case elasticache.MultiAZStatusEnabled:
			d.Set("multi_az_enabled", true)
		case elasticache.MultiAZStatusDisabled:
			d.Set("multi_az_enabled", false)
		default:
			log.Printf("Unknown MultiAZ state %q", aws.StringValue(rgp.MultiAZ))
		}
	}

	d.Set("kms_key_id", rgp.KmsKeyId)
	d.Set("replication_group_description", rgp.Description)
	d.Set("number_cache_clusters", len(rgp.MemberClusters))
	if err := d.Set("member_clusters", flattenStringSet(rgp.MemberClusters)); err != nil {
		return fmt.Errorf("error setting member_clusters: %w", err)
	}
	if err := d.Set("cluster_mode", flattenElasticacheNodeGroupsToClusterMode(rgp.NodeGroups)); err != nil {
		return fmt.Errorf("error setting cluster_mode attribute: %w", err)
	}
	d.Set("cluster_enabled", rgp.ClusterEnabled)
	d.Set("replication_group_id", rgp.ReplicationGroupId)
	d.Set("arn", rgp.ARN)

	if rgp.NodeGroups != nil {
		if len(rgp.NodeGroups[0].NodeGroupMembers) == 0 {
			return nil
		}

		cacheCluster := *rgp.NodeGroups[0].NodeGroupMembers[0]

		res, err := conn.DescribeCacheClusters(&elasticache.DescribeCacheClustersInput{
			CacheClusterId:    cacheCluster.CacheClusterId,
			ShowCacheNodeInfo: aws.Bool(true),
		})
		if err != nil {
			return err
		}

		if len(res.CacheClusters) == 0 {
			return nil
		}

		c := res.CacheClusters[0]
		d.Set("node_type", c.CacheNodeType)
		d.Set("engine", c.Engine)
		d.Set("engine_version", c.EngineVersion)
		d.Set("subnet_group_name", c.CacheSubnetGroupName)
		d.Set("security_group_names", flattenElastiCacheSecurityGroupNames(c.CacheSecurityGroups))
		d.Set("security_group_ids", flattenElastiCacheSecurityGroupIds(c.SecurityGroups))

		if c.CacheParameterGroup != nil {
			d.Set("parameter_group_name", c.CacheParameterGroup.CacheParameterGroupName)
		}

		d.Set("maintenance_window", c.PreferredMaintenanceWindow)
		d.Set("snapshot_window", rgp.SnapshotWindow)
		d.Set("snapshot_retention_limit", rgp.SnapshotRetentionLimit)

		if rgp.ConfigurationEndpoint != nil {
			d.Set("port", rgp.ConfigurationEndpoint.Port)
			d.Set("configuration_endpoint_address", rgp.ConfigurationEndpoint.Address)
		} else {
			d.Set("port", rgp.NodeGroups[0].PrimaryEndpoint.Port)
			d.Set("primary_endpoint_address", rgp.NodeGroups[0].PrimaryEndpoint.Address)
			d.Set("reader_endpoint_address", rgp.NodeGroups[0].ReaderEndpoint.Address)
		}

		d.Set("auto_minor_version_upgrade", c.AutoMinorVersionUpgrade)
		d.Set("at_rest_encryption_enabled", c.AtRestEncryptionEnabled)
		d.Set("transit_encryption_enabled", c.TransitEncryptionEnabled)

		if c.AuthTokenEnabled != nil && !aws.BoolValue(c.AuthTokenEnabled) {
			d.Set("auth_token", nil)
		}

		arn := arn.ARN{
			Partition: meta.(*AWSClient).partition,
			Service:   "elasticache",
			Region:    meta.(*AWSClient).region,
			AccountID: meta.(*AWSClient).accountid,
			Resource:  fmt.Sprintf("cluster:%s", aws.StringValue(c.CacheClusterId)),
		}.String()

		tags, err := keyvaluetags.ElasticacheListTags(conn, arn)

		if err != nil {
			return fmt.Errorf("error listing tags for resource (%s): %w", arn, err)
		}

		if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
			return fmt.Errorf("error setting tags: %w", err)
		}
	}

	return nil
}

func resourceAwsElasticacheReplicationGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn

	if d.HasChanges("cluster_mode.0.num_node_groups", "cluster_mode.0.replicas_per_node_group") {
		err := elasticacheReplicationGroupModifyShardConfiguration(conn, d)
		if err != nil {
			return fmt.Errorf("error modifying ElastiCache Replication Group (%s) shard configuration: %w", d.Id(), err)
		}
	} else if d.HasChange("number_cache_clusters") {
		err := elasticacheReplicationGroupModifyNumCacheClusters(conn, d)
		if err != nil {
			return fmt.Errorf("error modifying ElastiCache Replication Group (%s) clusters: %w", d.Id(), err)
		}
	}

	requestUpdate := false
	params := &elasticache.ModifyReplicationGroupInput{
		ApplyImmediately:   aws.Bool(d.Get("apply_immediately").(bool)),
		ReplicationGroupId: aws.String(d.Id()),
	}

	if d.HasChange("replication_group_description") {
		params.ReplicationGroupDescription = aws.String(d.Get("replication_group_description").(string))
		requestUpdate = true
	}

	if d.HasChange("automatic_failover_enabled") {
		params.AutomaticFailoverEnabled = aws.Bool(d.Get("automatic_failover_enabled").(bool))
		requestUpdate = true
	}

	if d.HasChange("auto_minor_version_upgrade") {
		params.AutoMinorVersionUpgrade = aws.Bool(d.Get("auto_minor_version_upgrade").(bool))
		requestUpdate = true
	}

	if d.HasChange("security_group_ids") {
		if attr := d.Get("security_group_ids").(*schema.Set); attr.Len() > 0 {
			params.SecurityGroupIds = expandStringSet(attr)
			requestUpdate = true
		}
	}

	if d.HasChange("security_group_names") {
		if attr := d.Get("security_group_names").(*schema.Set); attr.Len() > 0 {
			params.CacheSecurityGroupNames = expandStringSet(attr)
			requestUpdate = true
		}
	}

	if d.HasChange("maintenance_window") {
		params.PreferredMaintenanceWindow = aws.String(d.Get("maintenance_window").(string))
		requestUpdate = true
	}

	if d.HasChange("multi_az_enabled") {
		params.MultiAZEnabled = aws.Bool(d.Get("multi_az_enabled").(bool))
		requestUpdate = true
	}

	if d.HasChange("notification_topic_arn") {
		params.NotificationTopicArn = aws.String(d.Get("notification_topic_arn").(string))
		requestUpdate = true
	}

	if d.HasChange("parameter_group_name") {
		params.CacheParameterGroupName = aws.String(d.Get("parameter_group_name").(string))
		requestUpdate = true
	}

	if d.HasChange("engine_version") {
		params.EngineVersion = aws.String(d.Get("engine_version").(string))
		requestUpdate = true
	}

	if d.HasChange("snapshot_retention_limit") {
		// This is a real hack to set the Snapshotting Cluster ID to be the first Cluster in the RG
		o, _ := d.GetChange("snapshot_retention_limit")
		if o.(int) == 0 {
			params.SnapshottingClusterId = aws.String(fmt.Sprintf("%s-001", d.Id()))
		}

		params.SnapshotRetentionLimit = aws.Int64(int64(d.Get("snapshot_retention_limit").(int)))
		requestUpdate = true
	}

	if d.HasChange("snapshot_window") {
		params.SnapshotWindow = aws.String(d.Get("snapshot_window").(string))
		requestUpdate = true
	}

	if d.HasChange("node_type") {
		params.CacheNodeType = aws.String(d.Get("node_type").(string))
		requestUpdate = true
	}

	if requestUpdate {
		err := resourceAwsElasticacheReplicationGroupModify(conn, d.Timeout(schema.TimeoutUpdate), params)
		if err != nil {
			return fmt.Errorf("error updating ElastiCache Replication Group (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		clusters := d.Get("member_clusters").(*schema.Set).List()

		for _, cluster := range clusters {

			arn := arn.ARN{
				Partition: meta.(*AWSClient).partition,
				Service:   "elasticache",
				Region:    meta.(*AWSClient).region,
				AccountID: meta.(*AWSClient).accountid,
				Resource:  fmt.Sprintf("cluster:%s", cluster),
			}.String()

			o, n := d.GetChange("tags")
			if err := keyvaluetags.ElasticacheUpdateTags(conn, arn, o, n); err != nil {
				return fmt.Errorf("error updating tags: %w", err)
			}
		}
	}

	return resourceAwsElasticacheReplicationGroupRead(d, meta)
}

func resourceAwsElasticacheReplicationGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn

	if globalReplicationGroupID, ok := d.GetOk("global_replication_group_id"); ok {
		err := disassociateElasticacheReplicationGroup(conn, globalReplicationGroupID.(string), d.Id(), meta.(*AWSClient).region)
		if err != nil {
			return fmt.Errorf("error disassociating ElastiCache Replication Group (%s) from Global Replication Group (%s): %w", d.Id(), globalReplicationGroupID, err)
		}
	}

	var finalSnapshotID = d.Get("final_snapshot_identifier").(string)
	err := deleteElasticacheReplicationGroup(d.Id(), conn, finalSnapshotID, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return fmt.Errorf("error deleting ElastiCache Replication Group (%s): %w", d.Id(), err)
	}

	return nil
}

func disassociateElasticacheReplicationGroup(conn *elasticache.ElastiCache, globalReplicationGroupID, id, region string) error {
	input := &elasticache.DisassociateGlobalReplicationGroupInput{
		GlobalReplicationGroupId: aws.String(globalReplicationGroupID),
		ReplicationGroupId:       aws.String(id),
		ReplicationGroupRegion:   aws.String(region),
	}
	err := resource.Retry(waiter.GlobalReplicationGroupDisassociationRetryTimeout, func() *resource.RetryError {
		_, err := conn.DisassociateGlobalReplicationGroup(input)
		if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeGlobalReplicationGroupNotFoundFault) {
			return nil
		}
		if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeInvalidGlobalReplicationGroupStateFault) {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = conn.DisassociateGlobalReplicationGroup(input)
	}
	if tfawserr.ErrMessageContains(err, elasticache.ErrCodeInvalidParameterValueException, "is not associated with Global Replication Group") {
		return nil
	}
	if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeInvalidGlobalReplicationGroupStateFault) {
		return fmt.Errorf("tried for %s: %w", waiter.GlobalReplicationGroupDisassociationRetryTimeout.String(), err)
	}

	if err != nil {
		return err
	}

	_, err = waiter.GlobalReplicationGroupMemberDetached(conn, globalReplicationGroupID, id)
	if err != nil {
		return fmt.Errorf("waiting for completion: %w", err)
	}

	return nil

}

func deleteElasticacheReplicationGroup(replicationGroupID string, conn *elasticache.ElastiCache, finalSnapshotID string, timeout time.Duration) error {
	input := &elasticache.DeleteReplicationGroupInput{
		ReplicationGroupId: aws.String(replicationGroupID),
	}
	if finalSnapshotID != "" {
		input.FinalSnapshotIdentifier = aws.String(finalSnapshotID)
	}

	// 10 minutes should give any creating/deleting cache clusters or snapshots time to complete
	err := resource.Retry(10*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteReplicationGroup(input)
		if isAWSErr(err, elasticache.ErrCodeReplicationGroupNotFoundFault, "") {
			return nil
		}
		// Cache Cluster is creating/deleting or Replication Group is snapshotting
		// InvalidReplicationGroupState: Cache cluster tf-acc-test-uqhe-003 is not in a valid state to be deleted
		if isAWSErr(err, elasticache.ErrCodeInvalidReplicationGroupStateFault, "") {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = conn.DeleteReplicationGroup(input)
	}

	if isAWSErr(err, elasticache.ErrCodeReplicationGroupNotFoundFault, "") {
		return nil
	}
	if err != nil {
		return err
	}

	_, err = waiter.ReplicationGroupDeleted(conn, replicationGroupID, timeout)
	if err != nil {
		return err
	}

	return nil
}

func flattenElasticacheNodeGroupsToClusterMode(nodeGroups []*elasticache.NodeGroup) []map[string]interface{} {
	if len(nodeGroups) == 0 {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"num_node_groups":         len(nodeGroups),
		"replicas_per_node_group": (len(nodeGroups[0].NodeGroupMembers) - 1),
	}
	return []map[string]interface{}{m}
}

func elasticacheReplicationGroupModifyShardConfiguration(conn *elasticache.ElastiCache, d *schema.ResourceData) error {
	if d.HasChange("cluster_mode.0.num_node_groups") {
		err := elasticacheReplicationGroupModifyShardConfigurationNumNodeGroups(conn, d)
		if err != nil {
			return err
		}
	}

	if d.HasChange("cluster_mode.0.replicas_per_node_group") {
		err := elasticacheReplicationGroupModifyShardConfigurationReplicasPerNodeGroup(conn, d)
		if err != nil {
			return err
		}
	}

	return nil
}

func elasticacheReplicationGroupModifyShardConfigurationNumNodeGroups(conn *elasticache.ElastiCache, d *schema.ResourceData) error {
	o, n := d.GetChange("cluster_mode.0.num_node_groups")
	oldNumNodeGroups := o.(int)
	newNumNodeGroups := n.(int)

	input := &elasticache.ModifyReplicationGroupShardConfigurationInput{
		ApplyImmediately:   aws.Bool(true),
		NodeGroupCount:     aws.Int64(int64(newNumNodeGroups)),
		ReplicationGroupId: aws.String(d.Id()),
	}

	if oldNumNodeGroups > newNumNodeGroups {
		// Node Group IDs are 1 indexed: 0001 through 0015
		// Loop from highest old ID until we reach highest new ID
		nodeGroupsToRemove := []string{}
		for i := oldNumNodeGroups; i > newNumNodeGroups; i-- {
			nodeGroupID := fmt.Sprintf("%04d", i)
			nodeGroupsToRemove = append(nodeGroupsToRemove, nodeGroupID)
		}
		input.NodeGroupsToRemove = aws.StringSlice(nodeGroupsToRemove)
	}

	log.Printf("[DEBUG] Modifying ElastiCache Replication Group (%s) shard configuration: %s", d.Id(), input)
	_, err := conn.ModifyReplicationGroupShardConfiguration(input)
	if err != nil {
		return fmt.Errorf("error modifying ElastiCache Replication Group shard configuration: %w", err)
	}

	_, err = waiter.ReplicationGroupAvailable(conn, d.Id(), d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return fmt.Errorf("error waiting for ElastiCache Replication Group (%s) shard reconfiguration completion: %w", d.Id(), err)
	}

	return nil
}

func elasticacheReplicationGroupModifyShardConfigurationReplicasPerNodeGroup(conn *elasticache.ElastiCache, d *schema.ResourceData) error {
	o, n := d.GetChange("cluster_mode.0.replicas_per_node_group")
	oldReplicas := o.(int)
	newReplicas := n.(int)

	if newReplicas > oldReplicas {
		input := &elasticache.IncreaseReplicaCountInput{
			ApplyImmediately:   aws.Bool(true),
			NewReplicaCount:    aws.Int64(int64(newReplicas)),
			ReplicationGroupId: aws.String(d.Id()),
		}
		_, err := conn.IncreaseReplicaCount(input)
		if err != nil {
			return fmt.Errorf("error adding ElastiCache Replication Group (%s) replicas: %w", d.Id(), err)
		}
		_, err = waiter.ReplicationGroupAvailable(conn, d.Id(), d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return fmt.Errorf("error waiting for ElastiCache Replication Group (%s) replica addition: %w", d.Id(), err)
		}
	} else {
		input := &elasticache.DecreaseReplicaCountInput{
			ApplyImmediately:   aws.Bool(true),
			NewReplicaCount:    aws.Int64(int64(newReplicas)),
			ReplicationGroupId: aws.String(d.Id()),
		}
		_, err := conn.DecreaseReplicaCount(input)
		if err != nil {
			return fmt.Errorf("error removing ElastiCache Replication Group (%s) replicas: %w", d.Id(), err)
		}
		_, err = waiter.ReplicationGroupAvailable(conn, d.Id(), d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return fmt.Errorf("error waiting for ElastiCache Replication Group (%s) replica removal: %w", d.Id(), err)
		}
	}

	return nil
}

func elasticacheReplicationGroupModifyNumCacheClusters(conn *elasticache.ElastiCache, d *schema.ResourceData) error {
	o, n := d.GetChange("number_cache_clusters")
	oldNumberCacheClusters := o.(int)
	newNumberCacheClusters := n.(int)

	var err error
	if newNumberCacheClusters > oldNumberCacheClusters {
		err = elasticacheReplicationGroupIncreaseNumCacheClusters(conn, d.Id(), newNumberCacheClusters, d.Timeout(schema.TimeoutUpdate))
	} else if newNumberCacheClusters < oldNumberCacheClusters {
		err = elasticacheReplicationGroupDecreaseNumCacheClusters(conn, d.Id(), newNumberCacheClusters, d.Timeout(schema.TimeoutUpdate))
	}
	return err
}

func elasticacheReplicationGroupIncreaseNumCacheClusters(conn *elasticache.ElastiCache, replicationGroupID string, newNumberCacheClusters int, timeout time.Duration) error {
	input := &elasticache.IncreaseReplicaCountInput{
		ApplyImmediately:   aws.Bool(true),
		NewReplicaCount:    aws.Int64(int64(newNumberCacheClusters - 1)),
		ReplicationGroupId: aws.String(replicationGroupID),
	}
	_, err := conn.IncreaseReplicaCount(input)
	if err != nil {
		return fmt.Errorf("error adding ElastiCache Replication Group (%s) replicas: %w", replicationGroupID, err)
	}

	_, err = waiter.ReplicationGroupMemberClustersAvailable(conn, replicationGroupID, timeout)
	if err != nil {
		return fmt.Errorf("error waiting for ElastiCache Replication Group (%s) replica addition: %w", replicationGroupID, err)
	}

	return nil
}

func elasticacheReplicationGroupDecreaseNumCacheClusters(conn *elasticache.ElastiCache, replicationGroupID string, newNumberCacheClusters int, timeout time.Duration) error {
	input := &elasticache.DecreaseReplicaCountInput{
		ApplyImmediately:   aws.Bool(true),
		NewReplicaCount:    aws.Int64(int64(newNumberCacheClusters - 1)),
		ReplicationGroupId: aws.String(replicationGroupID),
	}
	_, err := conn.DecreaseReplicaCount(input)
	if err != nil {
		return fmt.Errorf("error removing ElastiCache Replication Group (%s) replicas: %w", replicationGroupID, err)
	}

	_, err = waiter.ReplicationGroupMemberClustersAvailable(conn, replicationGroupID, timeout)
	if err != nil {
		return fmt.Errorf("error waiting for ElastiCache Replication Group (%s) replica removal: %w", replicationGroupID, err)
	}

	return nil
}

func resourceAwsElasticacheReplicationGroupModify(conn *elasticache.ElastiCache, timeout time.Duration, input *elasticache.ModifyReplicationGroupInput) error {
	_, err := conn.ModifyReplicationGroup(input)
	if err != nil {
		return fmt.Errorf("error requesting modification: %w", err)
	}

	_, err = waiter.ReplicationGroupAvailable(conn, aws.StringValue(input.ReplicationGroupId), timeout)
	if err != nil {
		return fmt.Errorf("error waiting for modification: %w", err)
	}
	return nil
}

var validateReplicationGroupID schema.SchemaValidateFunc = validation.All(
	validation.StringLenBetween(1, 40),
	validation.StringMatch(regexp.MustCompile(`^[0-9a-zA-Z-]+$`), "must contain only alphanumeric characters and hyphens"),
	validation.StringMatch(regexp.MustCompile(`^[a-zA-Z]`), "must begin with a letter"),
	validation.StringDoesNotMatch(regexp.MustCompile(`--`), "cannot contain two consecutive hyphens"),
	validation.StringDoesNotMatch(regexp.MustCompile(`-$`), "cannot end with a hyphen"),
)
