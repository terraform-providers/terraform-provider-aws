package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsRDSClusterInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRDSClusterInstanceCreate,
		Read:   resourceAwsRDSClusterInstanceRead,
		Update: resourceAwsRDSClusterInstanceUpdate,
		Delete: resourceAwsRDSClusterInstanceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(90 * time.Minute),
			Update: schema.DefaultTimeout(90 * time.Minute),
			Delete: schema.DefaultTimeout(90 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"identifier": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"identifier_prefix"},
				ValidateFunc:  validateRdsIdentifier,
			},
			"identifier_prefix": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateRdsIdentifierPrefix,
			},

			"db_subnet_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"writer": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"cluster_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"port": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"publicly_accessible": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"instance_class": {
				Type:     schema.TypeString,
				Required: true,
			},

			"engine": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "aurora",
				ValidateFunc: validateRdsEngine(),
			},

			"engine_version": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"db_parameter_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			// apply_immediately is used to determine when the update modifications
			// take place.
			// See http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Overview.DBInstance.Modifying.html
			"apply_immediately": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"kms_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"storage_encrypted": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"dbi_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"auto_minor_version_upgrade": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"monitoring_role_arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"preferred_maintenance_window": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				StateFunc: func(v interface{}) string {
					if v != nil {
						value := v.(string)
						return strings.ToLower(value)
					}
					return ""
				},
				ValidateFunc: validateOnceAWeekWindowFormat,
			},

			"preferred_backup_window": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateOnceADayWindowFormat,
			},

			"monitoring_interval": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},

			"promotion_tier": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},

			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"performance_insights_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"performance_insights_kms_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateArn,
			},

			"copy_tags_to_snapshot": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"ca_cert_identifier": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsRDSClusterInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	createOpts := &rds.CreateDBInstanceInput{
		DBInstanceClass:         aws.String(d.Get("instance_class").(string)),
		CopyTagsToSnapshot:      aws.Bool(d.Get("copy_tags_to_snapshot").(bool)),
		DBClusterIdentifier:     aws.String(d.Get("cluster_identifier").(string)),
		Engine:                  aws.String(d.Get("engine").(string)),
		PubliclyAccessible:      aws.Bool(d.Get("publicly_accessible").(bool)),
		PromotionTier:           aws.Int64(int64(d.Get("promotion_tier").(int))),
		AutoMinorVersionUpgrade: aws.Bool(d.Get("auto_minor_version_upgrade").(bool)),
		Tags:                    keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().RdsTags(),
	}

	if attr, ok := d.GetOk("availability_zone"); ok {
		createOpts.AvailabilityZone = aws.String(attr.(string))
	}

	if attr, ok := d.GetOk("db_parameter_group_name"); ok {
		createOpts.DBParameterGroupName = aws.String(attr.(string))
	}

	if v, ok := d.GetOk("identifier"); ok {
		createOpts.DBInstanceIdentifier = aws.String(v.(string))
	} else {
		if v, ok := d.GetOk("identifier_prefix"); ok {
			createOpts.DBInstanceIdentifier = aws.String(resource.PrefixedUniqueId(v.(string)))
		} else {
			createOpts.DBInstanceIdentifier = aws.String(resource.PrefixedUniqueId("tf-"))
		}
	}

	if attr, ok := d.GetOk("db_subnet_group_name"); ok {
		createOpts.DBSubnetGroupName = aws.String(attr.(string))
	}

	if attr, ok := d.GetOk("engine_version"); ok {
		createOpts.EngineVersion = aws.String(attr.(string))
	}

	if attr, ok := d.GetOk("monitoring_role_arn"); ok {
		createOpts.MonitoringRoleArn = aws.String(attr.(string))
	}

	if attr, ok := d.GetOk("performance_insights_enabled"); ok {
		createOpts.EnablePerformanceInsights = aws.Bool(attr.(bool))
	}

	if attr, ok := d.GetOk("performance_insights_kms_key_id"); ok {
		createOpts.PerformanceInsightsKMSKeyId = aws.String(attr.(string))
	}

	if attr, ok := d.GetOk("preferred_backup_window"); ok {
		createOpts.PreferredBackupWindow = aws.String(attr.(string))
	}

	if attr, ok := d.GetOk("preferred_maintenance_window"); ok {
		createOpts.PreferredMaintenanceWindow = aws.String(attr.(string))
	}

	if attr, ok := d.GetOk("monitoring_interval"); ok {
		createOpts.MonitoringInterval = aws.Int64(int64(attr.(int)))
	}

	log.Printf("[DEBUG] Creating RDS DB Instance opts: %s", createOpts)
	var resp *rds.CreateDBInstanceOutput
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error
		resp, err = conn.CreateDBInstance(createOpts)
		if err != nil {
			if isAWSErr(err, "InvalidParameterValue", "IAM role ARN value is invalid or does not include the required permissions") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		resp, err = conn.CreateDBInstance(createOpts)
	}
	if err != nil {
		return fmt.Errorf("error creating RDS DB Instance: %s", err)
	}

	d.SetId(*resp.DBInstance.DBInstanceIdentifier)

	// reuse db_instance refresh func
	stateConf := &resource.StateChangeConf{
		Pending:    resourceAwsRdsClusterInstanceCreateUpdatePendingStates,
		Target:     []string{"available"},
		Refresh:    resourceAwsDbInstanceStateRefreshFunc(d.Id(), conn),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	// Wait, catching any errors
	_, err = stateConf.WaitForState()
	if err != nil {
		return err
	}

	// See also: resource_aws_db_instance.go
	// Some API calls (e.g. CreateDBInstanceReadReplica and
	// RestoreDBInstanceFromDBSnapshot do not support all parameters to
	// correctly apply all settings in one pass. For missing parameters or
	// unsupported configurations, we may need to call ModifyDBInstance
	// afterwards to prevent Terraform operators from API errors or needing
	// to double apply.
	var requiresModifyDbInstance bool
	modifyDbInstanceInput := &rds.ModifyDBInstanceInput{
		ApplyImmediately: aws.Bool(true),
	}

	// Some ModifyDBInstance parameters (e.g. DBParameterGroupName) require
	// a database instance reboot to take affect. During resource creation,
	// we expect everything to be in sync before returning completion.
	var requiresRebootDbInstance bool

	if attr, ok := d.GetOk("ca_cert_identifier"); ok && attr.(string) != aws.StringValue(resp.DBInstance.CACertificateIdentifier) {
		modifyDbInstanceInput.CACertificateIdentifier = aws.String(attr.(string))
		requiresModifyDbInstance = true
		requiresRebootDbInstance = true
	}

	if requiresModifyDbInstance {
		modifyDbInstanceInput.DBInstanceIdentifier = aws.String(d.Id())

		log.Printf("[INFO] DB Instance (%s) configuration requires ModifyDBInstance: %s", d.Id(), modifyDbInstanceInput)
		_, err := conn.ModifyDBInstance(modifyDbInstanceInput)
		if err != nil {
			return fmt.Errorf("error modifying DB Instance (%s): %s", d.Id(), err)
		}

		log.Printf("[INFO] Waiting for DB Instance (%s) to be available", d.Id())
		err = waitUntilAwsDbInstanceIsAvailableAfterUpdate(d.Id(), conn, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return fmt.Errorf("error waiting for DB Instance (%s) to be available: %s", d.Id(), err)
		}
	}

	if requiresRebootDbInstance {
		rebootDbInstanceInput := &rds.RebootDBInstanceInput{
			DBInstanceIdentifier: aws.String(d.Id()),
		}

		log.Printf("[INFO] DB Instance (%s) configuration requires RebootDBInstance: %s", d.Id(), rebootDbInstanceInput)
		_, err := conn.RebootDBInstance(rebootDbInstanceInput)
		if err != nil {
			return fmt.Errorf("error rebooting DB Instance (%s): %s", d.Id(), err)
		}

		log.Printf("[INFO] Waiting for DB Instance (%s) to be available", d.Id())
		err = waitUntilAwsDbInstanceIsAvailableAfterUpdate(d.Id(), conn, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return fmt.Errorf("error waiting for DB Instance (%s) to be available: %s", d.Id(), err)
		}
	}

	return resourceAwsRDSClusterInstanceRead(d, meta)
}

func resourceAwsRDSClusterInstanceRead(d *schema.ResourceData, meta interface{}) error {
	db, err := resourceAwsDbInstanceRetrieve(d.Id(), meta.(*AWSClient).rdsconn)
	// Errors from this helper are always reportable
	if err != nil {
		return fmt.Errorf("Error on retrieving RDS Cluster Instance (%s): %s", d.Id(), err)
	}
	// A nil response means "not found"
	if db == nil {
		log.Printf("[WARN] RDS Cluster Instance (%s): not found, removing from state.", d.Id())
		d.SetId("")
		return nil
	}
	// Database instance is not in RDS Cluster
	if db.DBClusterIdentifier == nil {
		return fmt.Errorf("Cluster identifier is missing from instance (%s). The aws_db_instance resource should be used for non-Aurora instances", d.Id())
	}

	// Retrieve DB Cluster information, to determine if this Instance is a writer
	conn := meta.(*AWSClient).rdsconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := conn.DescribeDBClusters(&rds.DescribeDBClustersInput{
		DBClusterIdentifier: db.DBClusterIdentifier,
	})

	var dbc *rds.DBCluster
	for _, c := range resp.DBClusters {
		if *c.DBClusterIdentifier == *db.DBClusterIdentifier {
			dbc = c
		}
	}

	if dbc == nil {
		return fmt.Errorf("Error finding RDS Cluster (%s) for Cluster Instance (%s): %s",
			*db.DBClusterIdentifier, *db.DBInstanceIdentifier, err)
	}

	for _, m := range dbc.DBClusterMembers {
		if *db.DBInstanceIdentifier == *m.DBInstanceIdentifier {
			if *m.IsClusterWriter {
				d.Set("writer", true)
			} else {
				d.Set("writer", false)
			}
		}
	}

	if db.Endpoint != nil {
		d.Set("endpoint", db.Endpoint.Address)
		d.Set("port", db.Endpoint.Port)
	}

	if db.DBSubnetGroup != nil {
		d.Set("db_subnet_group_name", db.DBSubnetGroup.DBSubnetGroupName)
	}

	d.Set("arn", db.DBInstanceArn)
	d.Set("auto_minor_version_upgrade", db.AutoMinorVersionUpgrade)
	d.Set("availability_zone", db.AvailabilityZone)
	d.Set("cluster_identifier", db.DBClusterIdentifier)
	d.Set("copy_tags_to_snapshot", db.CopyTagsToSnapshot)
	d.Set("dbi_resource_id", db.DbiResourceId)
	d.Set("engine_version", db.EngineVersion)
	d.Set("engine", db.Engine)
	d.Set("identifier", db.DBInstanceIdentifier)
	d.Set("instance_class", db.DBInstanceClass)
	d.Set("kms_key_id", db.KmsKeyId)
	d.Set("monitoring_interval", db.MonitoringInterval)
	d.Set("monitoring_role_arn", db.MonitoringRoleArn)
	d.Set("performance_insights_enabled", db.PerformanceInsightsEnabled)
	d.Set("performance_insights_kms_key_id", db.PerformanceInsightsKMSKeyId)
	d.Set("preferred_backup_window", db.PreferredBackupWindow)
	d.Set("preferred_maintenance_window", db.PreferredMaintenanceWindow)
	d.Set("promotion_tier", db.PromotionTier)
	d.Set("publicly_accessible", db.PubliclyAccessible)
	d.Set("storage_encrypted", db.StorageEncrypted)
	d.Set("ca_cert_identifier", db.CACertificateIdentifier)

	if len(db.DBParameterGroups) > 0 {
		d.Set("db_parameter_group_name", db.DBParameterGroups[0].DBParameterGroupName)
	}

	tags, err := keyvaluetags.RdsListTags(conn, aws.StringValue(db.DBInstanceArn))
	if err != nil {
		return fmt.Errorf("error listing tags for RDS Cluster Instance (%s): %s", d.Id(), err)
	}
	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsRDSClusterInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn
	requestUpdate := false

	req := &rds.ModifyDBInstanceInput{
		ApplyImmediately:     aws.Bool(d.Get("apply_immediately").(bool)),
		DBInstanceIdentifier: aws.String(d.Id()),
	}

	if d.HasChange("db_parameter_group_name") {
		req.DBParameterGroupName = aws.String(d.Get("db_parameter_group_name").(string))
		requestUpdate = true
	}

	if d.HasChange("instance_class") {
		req.DBInstanceClass = aws.String(d.Get("instance_class").(string))
		requestUpdate = true
	}

	if d.HasChange("monitoring_role_arn") {
		req.MonitoringRoleArn = aws.String(d.Get("monitoring_role_arn").(string))
		requestUpdate = true
	}

	if d.HasChanges("performance_insights_enabled", "performance_insights_kms_key_id") {
		req.EnablePerformanceInsights = aws.Bool(d.Get("performance_insights_enabled").(bool))

		if v, ok := d.GetOk("performance_insights_kms_key_id"); ok {
			req.PerformanceInsightsKMSKeyId = aws.String(v.(string))
		}

		requestUpdate = true
	}

	if d.HasChange("preferred_backup_window") {
		req.PreferredBackupWindow = aws.String(d.Get("preferred_backup_window").(string))
		requestUpdate = true
	}

	if d.HasChange("preferred_maintenance_window") {
		req.PreferredMaintenanceWindow = aws.String(d.Get("preferred_maintenance_window").(string))
		requestUpdate = true
	}

	if d.HasChange("monitoring_interval") {
		req.MonitoringInterval = aws.Int64(int64(d.Get("monitoring_interval").(int)))
		requestUpdate = true
	}

	if d.HasChange("auto_minor_version_upgrade") {
		req.AutoMinorVersionUpgrade = aws.Bool(d.Get("auto_minor_version_upgrade").(bool))
		requestUpdate = true
	}

	if d.HasChange("copy_tags_to_snapshot") {
		req.CopyTagsToSnapshot = aws.Bool(d.Get("copy_tags_to_snapshot").(bool))
		requestUpdate = true
	}

	if d.HasChange("promotion_tier") {
		req.PromotionTier = aws.Int64(int64(d.Get("promotion_tier").(int)))
		requestUpdate = true
	}

	if d.HasChange("publicly_accessible") {
		req.PubliclyAccessible = aws.Bool(d.Get("publicly_accessible").(bool))
		requestUpdate = true
	}

	if d.HasChange("ca_cert_identifier") {
		req.CACertificateIdentifier = aws.String(d.Get("ca_cert_identifier").(string))
		requestUpdate = true
	}

	log.Printf("[DEBUG] Send DB Instance Modification request: %#v", requestUpdate)
	if requestUpdate {
		log.Printf("[DEBUG] DB Instance Modification request: %#v", req)
		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			_, err := conn.ModifyDBInstance(req)
			if err != nil {
				if isAWSErr(err, "InvalidParameterValue", "IAM role ARN value is invalid or does not include the required permissions") {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if isResourceTimeoutError(err) {
			_, err = conn.ModifyDBInstance(req)
		}
		if err != nil {
			return fmt.Errorf("Error modifying DB Instance %s: %s", d.Id(), err)
		}

		// reuse db_instance refresh func
		stateConf := &resource.StateChangeConf{
			Pending:    resourceAwsRdsClusterInstanceCreateUpdatePendingStates,
			Target:     []string{"available"},
			Refresh:    resourceAwsDbInstanceStateRefreshFunc(d.Id(), conn),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			MinTimeout: 10 * time.Second,
			Delay:      30 * time.Second, // Wait 30 secs before starting
		}

		// Wait, catching any errors
		_, err = stateConf.WaitForState()
		if err != nil {
			return err
		}

	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.RdsUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating RDS Cluster Instance (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsRDSClusterInstanceRead(d, meta)
}

func resourceAwsRDSClusterInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	log.Printf("[DEBUG] RDS Cluster Instance destroy: %v", d.Id())

	opts := rds.DeleteDBInstanceInput{DBInstanceIdentifier: aws.String(d.Id())}

	log.Printf("[DEBUG] RDS Cluster Instance destroy configuration: %s", opts)
	_, err := conn.DeleteDBInstance(&opts)

	if err != nil && !isAWSErr(err, rds.ErrCodeInvalidDBInstanceStateFault, "is already being deleted") {
		return fmt.Errorf("error deleting Database Instance %q: %s", d.Id(), err)
	}

	// re-uses db_instance refresh func
	log.Println("[INFO] Waiting for RDS Cluster Instance to be destroyed")
	stateConf := &resource.StateChangeConf{
		Pending:    resourceAwsRdsClusterInstanceDeletePendingStates,
		Target:     []string{},
		Refresh:    resourceAwsDbInstanceStateRefreshFunc(d.Id(), conn),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}

	_, err = stateConf.WaitForState()
	return err

}

var resourceAwsRdsClusterInstanceCreateUpdatePendingStates = []string{
	"backing-up",
	"configuring-enhanced-monitoring",
	"configuring-log-exports",
	"creating",
	"maintenance",
	"modifying",
	"rebooting",
	"renaming",
	"resetting-master-credentials",
	"starting",
	"upgrading",
}

var resourceAwsRdsClusterInstanceDeletePendingStates = []string{
	"configuring-log-exports",
	"modifying",
	"deleting",
}
