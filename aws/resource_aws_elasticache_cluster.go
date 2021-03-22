package aws

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	gversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/elasticache/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/elasticache/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

const (
	elasticacheDefaultRedisPort     = "6379"
	elasticacheDefaultMemcachedPort = "11211"
)

func resourceAwsElasticacheCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsElasticacheClusterCreate,
		Read:   resourceAwsElasticacheClusterRead,
		Update: resourceAwsElasticacheClusterUpdate,
		Delete: resourceAwsElasticacheClusterDelete,
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
			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"az_mode": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					elasticache.AZModeCrossAz,
					elasticache.AZModeSingleAz,
				}, false),
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
			"cluster_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				StateFunc: func(val interface{}) string {
					// ElastiCache normalizes cluster ids to lowercase,
					// so we have to do this too or else we can end up
					// with non-converging diffs.
					return strings.ToLower(val.(string))
				},
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 50),
					validation.StringMatch(regexp.MustCompile(`^[0-9a-z-]+$`), "must contain only lowercase alphanumeric characters and hyphens"),
					validation.StringMatch(regexp.MustCompile(`^[a-z]`), "must begin with a lowercase letter"),
					validation.StringDoesNotMatch(regexp.MustCompile(`--`), "cannot contain two consecutive hyphens"),
					validation.StringDoesNotMatch(regexp.MustCompile(`-$`), "cannot end with a hyphen"),
				),
			},
			"configuration_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"engine_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"maintenance_window": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				StateFunc: func(val interface{}) string {
					// ElastiCache always changes the maintenance
					// to lowercase
					return strings.ToLower(val.(string))
				},
				ValidateFunc: validateOnceAWeekWindowFormat,
			},
			"node_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"notification_topic_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"num_cache_nodes": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"parameter_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// Suppress default memcached/redis ports when not defined
					if !d.IsNewResource() && new == "0" && (old == elasticacheDefaultRedisPort || old == elasticacheDefaultMemcachedPort) {
						return true
					}
					return false
				},
			},
			"preferred_availability_zones": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"replication_group_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateReplicationGroupID,
				ConflictsWith: []string{
					"az_mode",
					"engine_version",
					"engine",
					"maintenance_window",
					"node_type",
					"notification_topic_arn",
					"num_cache_nodes",
					"parameter_group_name",
					"port",
					"security_group_ids",
					"security_group_names",
					"snapshot_arns",
					"snapshot_name",
					"snapshot_retention_limit",
					"snapshot_window",
					"subnet_group_name",
				},
				Computed: true,
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
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.All(
						validateArn,
						validation.StringDoesNotContainAny(","),
					),
				},
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
			"final_snapshot_identifier": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": tagsSchema(),
		},

		CustomizeDiff: customdiff.Sequence(
			func(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
				// Plan time validation for az_mode
				// InvalidParameterCombination: Must specify at least two cache nodes in order to specify AZ Mode of 'cross-az'.
				if v, ok := diff.GetOk("az_mode"); !ok || v.(string) != elasticache.AZModeCrossAz {
					return nil
				}
				if v, ok := diff.GetOk("num_cache_nodes"); !ok || v.(int) != 1 {
					return nil
				}
				return errors.New(`az_mode "cross-az" is not supported with num_cache_nodes = 1`)
			},
			func(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
				// Plan time validation for engine_version
				// InvalidParameterCombination: Cannot modify memcached from 1.4.33 to 1.4.24
				// InvalidParameterCombination: Cannot modify redis from 3.2.6 to 3.2.4
				if diff.Id() == "" || !diff.HasChange("engine_version") {
					return nil
				}
				o, n := diff.GetChange("engine_version")
				oVersion, err := gversion.NewVersion(o.(string))
				if err != nil {
					return err
				}
				nVersion, err := gversion.NewVersion(n.(string))
				if err != nil {
					return err
				}
				if nVersion.GreaterThan(oVersion) {
					return nil
				}
				return diff.ForceNew("engine_version")
			},
			func(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
				// Plan time validation for num_cache_nodes
				// InvalidParameterValue: Cannot create a Redis cluster with a NumCacheNodes parameter greater than 1.
				if v, ok := diff.GetOk("engine"); !ok || v.(string) == "memcached" {
					return nil
				}
				if v, ok := diff.GetOk("num_cache_nodes"); !ok || v.(int) == 1 {
					return nil
				}
				return errors.New(`engine "redis" does not support num_cache_nodes > 1`)
			},
			func(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
				// Engine memcached does not currently support vertical scaling
				// InvalidParameterCombination: Scaling is not supported for engine memcached
				// https://docs.aws.amazon.com/AmazonElastiCache/latest/mem-ug/Scaling.html#Scaling.Memcached.Vertically
				if diff.Id() == "" || !diff.HasChange("node_type") {
					return nil
				}
				if v, ok := diff.GetOk("engine"); !ok || v.(string) == "redis" {
					return nil
				}
				return diff.ForceNew("node_type")
			},
			func(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
				if v, ok := diff.GetOk("engine"); !ok || v.(string) == "redis" {
					return nil
				}
				if _, ok := diff.GetOk("final_snapshot_identifier"); !ok {
					return nil
				}
				return errors.New(`engine "memcached" does not support final_snapshot_identifier`)
			},
		),
	}
}

func resourceAwsElasticacheClusterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn

	req := &elasticache.CreateCacheClusterInput{}

	if v, ok := d.GetOk("replication_group_id"); ok {
		req.ReplicationGroupId = aws.String(v.(string))
	} else {
		req.CacheSecurityGroupNames = expandStringSet(d.Get("security_group_names").(*schema.Set))
		req.SecurityGroupIds = expandStringSet(d.Get("security_group_ids").(*schema.Set))
		req.Tags = keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().ElasticacheTags()
	}

	if v, ok := d.GetOk("cluster_id"); ok {
		req.CacheClusterId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("node_type"); ok {
		req.CacheNodeType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("num_cache_nodes"); ok {
		req.NumCacheNodes = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("engine"); ok {
		req.Engine = aws.String(v.(string))
	}

	if v, ok := d.GetOk("engine_version"); ok {
		req.EngineVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("port"); ok {
		req.Port = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("subnet_group_name"); ok {
		req.CacheSubnetGroupName = aws.String(v.(string))
	}

	// parameter groups are optional and can be defaulted by AWS
	if v, ok := d.GetOk("parameter_group_name"); ok {
		req.CacheParameterGroupName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("snapshot_retention_limit"); ok {
		req.SnapshotRetentionLimit = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("snapshot_window"); ok {
		req.SnapshotWindow = aws.String(v.(string))
	}

	if v, ok := d.GetOk("maintenance_window"); ok {
		req.PreferredMaintenanceWindow = aws.String(v.(string))
	}

	if v, ok := d.GetOk("notification_topic_arn"); ok {
		req.NotificationTopicArn = aws.String(v.(string))
	}

	snaps := d.Get("snapshot_arns").([]interface{})
	if len(snaps) > 0 {
		req.SnapshotArns = expandStringList(snaps)
		log.Printf("[DEBUG] Restoring Redis cluster from S3 snapshot: %#v", snaps)
	}

	if v, ok := d.GetOk("snapshot_name"); ok {
		req.SnapshotName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("az_mode"); ok {
		req.AZMode = aws.String(v.(string))
	}

	if v, ok := d.GetOk("availability_zone"); ok {
		req.PreferredAvailabilityZone = aws.String(v.(string))
	}

	if v, ok := d.GetOk("preferred_availability_zones"); ok && len(v.([]interface{})) > 0 {
		req.PreferredAvailabilityZones = expandStringList(v.([]interface{}))
	}

	id, err := createElasticacheCacheCluster(conn, req)
	if err != nil {
		return fmt.Errorf("error creating ElastiCache Cache Cluster: %w", err)
	}

	d.SetId(id)

	_, err = waiter.CacheClusterAvailable(conn, d.Id(), 40*time.Minute)
	if err != nil {
		return fmt.Errorf("error waiting for ElastiCache Cache Cluster (%s) to be created: %w", d.Id(), err)
	}

	return resourceAwsElasticacheClusterRead(d, meta)
}

func resourceAwsElasticacheClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	c, err := finder.CacheClusterWithNodeInfoByID(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ElastiCache Cache Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading ElastiCache Cache Cluster (%s): %w", d.Id(), err)
	}

	d.Set("cluster_id", c.CacheClusterId)
	d.Set("node_type", c.CacheNodeType)
	d.Set("num_cache_nodes", c.NumCacheNodes)
	d.Set("engine", c.Engine)
	d.Set("engine_version", c.EngineVersion)
	if c.ConfigurationEndpoint != nil {
		d.Set("port", c.ConfigurationEndpoint.Port)
		d.Set("configuration_endpoint", aws.String(fmt.Sprintf("%s:%d", aws.StringValue(c.ConfigurationEndpoint.Address), aws.Int64Value(c.ConfigurationEndpoint.Port))))
		d.Set("cluster_address", c.ConfigurationEndpoint.Address)
	} else if len(c.CacheNodes) > 0 {
		d.Set("port", int(aws.Int64Value(c.CacheNodes[0].Endpoint.Port)))
	}

	if c.ReplicationGroupId != nil {
		d.Set("replication_group_id", c.ReplicationGroupId)
	}

	d.Set("subnet_group_name", c.CacheSubnetGroupName)
	d.Set("security_group_names", flattenElastiCacheSecurityGroupNames(c.CacheSecurityGroups))
	d.Set("security_group_ids", flattenElastiCacheSecurityGroupIds(c.SecurityGroups))
	if c.CacheParameterGroup != nil {
		d.Set("parameter_group_name", c.CacheParameterGroup.CacheParameterGroupName)
	}
	d.Set("maintenance_window", c.PreferredMaintenanceWindow)
	d.Set("snapshot_window", c.SnapshotWindow)
	d.Set("snapshot_retention_limit", c.SnapshotRetentionLimit)
	if c.NotificationConfiguration != nil {
		if *c.NotificationConfiguration.TopicStatus == "active" {
			d.Set("notification_topic_arn", c.NotificationConfiguration.TopicArn)
		}
	}
	d.Set("availability_zone", c.PreferredAvailabilityZone)
	if aws.StringValue(c.PreferredAvailabilityZone) == "Multiple" {
		d.Set("az_mode", "cross-az")
	} else {
		d.Set("az_mode", "single-az")
	}

	if err := setCacheNodeData(d, c); err != nil {
		return err
	}

	d.Set("arn", c.ARN)

	tags, err := keyvaluetags.ElasticacheListTags(conn, aws.StringValue(c.ARN))

	if err != nil {
		return fmt.Errorf("error listing tags for ElastiCache Cluster (%s): %w", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceAwsElasticacheClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.ElasticacheUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating ElastiCache Cluster (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	req := &elasticache.ModifyCacheClusterInput{
		CacheClusterId:   aws.String(d.Id()),
		ApplyImmediately: aws.Bool(d.Get("apply_immediately").(bool)),
	}

	requestUpdate := false
	if d.HasChange("security_group_ids") {
		if attr := d.Get("security_group_ids").(*schema.Set); attr.Len() > 0 {
			req.SecurityGroupIds = expandStringSet(attr)
			requestUpdate = true
		}
	}

	if d.HasChange("parameter_group_name") {
		req.CacheParameterGroupName = aws.String(d.Get("parameter_group_name").(string))
		requestUpdate = true
	}

	if d.HasChange("maintenance_window") {
		req.PreferredMaintenanceWindow = aws.String(d.Get("maintenance_window").(string))
		requestUpdate = true
	}

	if d.HasChange("notification_topic_arn") {
		v := d.Get("notification_topic_arn").(string)
		req.NotificationTopicArn = aws.String(v)
		if v == "" {
			inactive := "inactive"
			req.NotificationTopicStatus = &inactive
		}
		requestUpdate = true
	}

	if d.HasChange("engine_version") {
		req.EngineVersion = aws.String(d.Get("engine_version").(string))
		requestUpdate = true
	}

	if d.HasChange("snapshot_window") {
		req.SnapshotWindow = aws.String(d.Get("snapshot_window").(string))
		requestUpdate = true
	}

	if d.HasChange("node_type") {
		req.CacheNodeType = aws.String(d.Get("node_type").(string))
		requestUpdate = true
	}

	if d.HasChange("snapshot_retention_limit") {
		req.SnapshotRetentionLimit = aws.Int64(int64(d.Get("snapshot_retention_limit").(int)))
		requestUpdate = true
	}

	if d.HasChange("az_mode") {
		req.AZMode = aws.String(d.Get("az_mode").(string))
		requestUpdate = true
	}

	if d.HasChange("num_cache_nodes") {
		oraw, nraw := d.GetChange("num_cache_nodes")
		o := oraw.(int)
		n := nraw.(int)
		if n < o {
			log.Printf("[INFO] Cluster %s is marked for Decreasing cache nodes from %d to %d", d.Id(), o, n)
			nodesToRemove := getCacheNodesToRemove(o, o-n)
			req.CacheNodeIdsToRemove = nodesToRemove
		} else {
			log.Printf("[INFO] Cluster %s is marked for increasing cache nodes from %d to %d", d.Id(), o, n)
			// SDK documentation for NewAvailabilityZones states:
			// The list of Availability Zones where the new Memcached cache nodes are created.
			//
			// This parameter is only valid when NumCacheNodes in the request is greater
			// than the sum of the number of active cache nodes and the number of cache
			// nodes pending creation (which may be zero). The number of Availability Zones
			// supplied in this list must match the cache nodes being added in this request.
			if v, ok := d.GetOk("preferred_availability_zones"); ok && len(v.([]interface{})) > 0 {
				// Here we check the list length to prevent a potential panic :)
				if len(v.([]interface{})) != n {
					return fmt.Errorf("length of preferred_availability_zones (%d) must match num_cache_nodes (%d)", len(v.([]interface{})), n)
				}
				req.NewAvailabilityZones = expandStringList(v.([]interface{})[o:])
			}
		}

		req.NumCacheNodes = aws.Int64(int64(d.Get("num_cache_nodes").(int)))
		requestUpdate = true

	}

	if requestUpdate {
		log.Printf("[DEBUG] Modifying ElastiCache Cluster (%s), opts:\n%s", d.Id(), req)
		_, err := conn.ModifyCacheCluster(req)
		if err != nil {
			return fmt.Errorf("Error updating ElastiCache cluster (%s), error: %w", d.Id(), err)
		}

		_, err = waiter.CacheClusterAvailable(conn, d.Id(), waiter.CacheClusterUpdatedTimeout)
		if err != nil {
			return fmt.Errorf("error waiting for ElastiCache Cache Cluster (%s) to update: %w", d.Id(), err)
		}
	}

	return resourceAwsElasticacheClusterRead(d, meta)
}

func getCacheNodesToRemove(oldNumberOfNodes int, cacheNodesToRemove int) []*string {
	nodesIdsToRemove := []*string{}
	for i := oldNumberOfNodes; i > oldNumberOfNodes-cacheNodesToRemove && i > 0; i-- {
		s := fmt.Sprintf("%04d", i)
		nodesIdsToRemove = append(nodesIdsToRemove, &s)
	}

	return nodesIdsToRemove
}

func setCacheNodeData(d *schema.ResourceData, c *elasticache.CacheCluster) error {
	sortedCacheNodes := make([]*elasticache.CacheNode, len(c.CacheNodes))
	copy(sortedCacheNodes, c.CacheNodes)
	sort.Sort(byCacheNodeId(sortedCacheNodes))

	cacheNodeData := make([]map[string]interface{}, 0, len(sortedCacheNodes))

	for _, node := range sortedCacheNodes {
		if node.CacheNodeId == nil || node.Endpoint == nil || node.Endpoint.Address == nil || node.Endpoint.Port == nil || node.CustomerAvailabilityZone == nil {
			return fmt.Errorf("Unexpected nil pointer in: %s", node)
		}
		cacheNodeData = append(cacheNodeData, map[string]interface{}{
			"id":                *node.CacheNodeId,
			"address":           *node.Endpoint.Address,
			"port":              int(*node.Endpoint.Port),
			"availability_zone": *node.CustomerAvailabilityZone,
		})
	}

	return d.Set("cache_nodes", cacheNodeData)
}

type byCacheNodeId []*elasticache.CacheNode

func (b byCacheNodeId) Len() int      { return len(b) }
func (b byCacheNodeId) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b byCacheNodeId) Less(i, j int) bool {
	return b[i].CacheNodeId != nil && b[j].CacheNodeId != nil &&
		*b[i].CacheNodeId < *b[j].CacheNodeId
}

func resourceAwsElasticacheClusterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn

	var finalSnapshotID = d.Get("final_snapshot_identifier").(string)
	err := deleteElasticacheCacheCluster(conn, d.Id(), finalSnapshotID)
	if err != nil {
		if isAWSErr(err, elasticache.ErrCodeCacheClusterNotFoundFault, "") {
			return nil
		}
		return fmt.Errorf("error deleting ElastiCache Cache Cluster (%s): %w", d.Id(), err)
	}
	_, err = waiter.CacheClusterDeleted(conn, d.Id(), waiter.CacheClusterDeletedTimeout)
	if err != nil {
		return fmt.Errorf("error waiting for ElastiCache Cache Cluster (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}

func createElasticacheCacheCluster(conn *elasticache.ElastiCache, input *elasticache.CreateCacheClusterInput) (string, error) {
	log.Printf("[DEBUG] Creating ElastiCache Cache Cluster: %s", input)
	output, err := conn.CreateCacheCluster(input)
	if err != nil {
		return "", err
	}
	if output == nil || output.CacheCluster == nil {
		return "", errors.New("missing cluster ID after creation")
	}
	// Elasticache always retains the id in lower case, so we have to
	// mimic that or else we won't be able to refresh a resource whose
	// name contained uppercase characters.
	return strings.ToLower(aws.StringValue(output.CacheCluster.CacheClusterId)), nil
}

func deleteElasticacheCacheCluster(conn *elasticache.ElastiCache, cacheClusterID string, finalSnapshotID string) error {
	input := &elasticache.DeleteCacheClusterInput{
		CacheClusterId: aws.String(cacheClusterID),
	}
	if finalSnapshotID != "" {
		input.FinalSnapshotIdentifier = aws.String(finalSnapshotID)
	}

	log.Printf("[DEBUG] Deleting ElastiCache Cache Cluster: %s", input)
	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteCacheCluster(input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, elasticache.ErrCodeInvalidCacheClusterStateFault, "serving as primary") {
				return resource.NonRetryableError(err)
			}
			if tfawserr.ErrMessageContains(err, elasticache.ErrCodeInvalidCacheClusterStateFault, "only member of a replication group") {
				return resource.NonRetryableError(err)
			}
			// The cluster may be just snapshotting, so we retry until it's ready for deletion
			if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeInvalidCacheClusterStateFault) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = conn.DeleteCacheCluster(input)
	}

	return err
}
