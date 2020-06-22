package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsDocDBClusterInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDocDBClusterInstanceCreate,
		Read:   resourceAwsDocDBClusterInstanceRead,
		Update: resourceAwsDocDBClusterInstanceUpdate,
		Delete: resourceAwsDocDBClusterInstanceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(90 * time.Minute),
			Update: schema.DefaultTimeout(90 * time.Minute),
			Delete: schema.DefaultTimeout(90 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			// apply_immediately is used to determine when the update modifications take place.
			// See http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Overview.DBInstance.Modifying.html
			"apply_immediately": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"auto_minor_version_upgrade": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"cluster_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"dbi_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"db_subnet_group_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"engine": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "docdb",
				ValidateFunc: validateDocDBEngine(),
			},

			"engine_version": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"kms_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"identifier": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"identifier_prefix"},
				ValidateFunc:  validateDocDBIdentifier,
			},

			"identifier_prefix": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateDocDBIdentifierPrefix,
			},

			"instance_class": {
				Type:     schema.TypeString,
				Required: true,
			},

			"port": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"preferred_backup_window": {
				Type:     schema.TypeString,
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

			"promotion_tier": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntBetween(0, 15),
			},

			"publicly_accessible": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"storage_encrypted": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"ca_cert_identifier": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"tags": tagsSchema(),

			"writer": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func resourceAwsDocDBClusterInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).docdbconn
	tags := keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().DocdbTags()

	createOpts := &docdb.CreateDBInstanceInput{
		DBInstanceClass:         aws.String(d.Get("instance_class").(string)),
		DBClusterIdentifier:     aws.String(d.Get("cluster_identifier").(string)),
		Engine:                  aws.String(d.Get("engine").(string)),
		PromotionTier:           aws.Int64(int64(d.Get("promotion_tier").(int))),
		AutoMinorVersionUpgrade: aws.Bool(d.Get("auto_minor_version_upgrade").(bool)),
		Tags:                    tags,
	}

	if attr, ok := d.GetOk("availability_zone"); ok {
		createOpts.AvailabilityZone = aws.String(attr.(string))
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

	if attr, ok := d.GetOk("preferred_maintenance_window"); ok {
		createOpts.PreferredMaintenanceWindow = aws.String(attr.(string))
	}

	log.Printf("[DEBUG] Creating DocDB Instance opts: %s", createOpts)
	var resp *docdb.CreateDBInstanceOutput
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
		return fmt.Errorf("error creating DocDB Instance: %s", err)
	}

	d.SetId(*resp.DBInstance.DBInstanceIdentifier)

	// reuse db_instance refresh func
	stateConf := &resource.StateChangeConf{
		Pending:    resourceAwsDocDBClusterInstanceCreateUpdatePendingStates,
		Target:     []string{"available"},
		Refresh:    resourceAwsDocDBInstanceStateRefreshFunc(d.Id(), conn),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	// Wait, catching any errors
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for DocDB Instance (%s) to become available: %s", d.Id(), err)
	}

	return resourceAwsDocDBClusterInstanceRead(d, meta)
}

func resourceAwsDocDBClusterInstanceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).docdbconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	db, err := resourceAwsDocDBInstanceRetrieve(d.Id(), conn)
	// Errors from this helper are always reportable
	if err != nil {
		return fmt.Errorf("Error on retrieving DocDB Cluster Instance (%s): %s", d.Id(), err)
	}
	// A nil response means "not found"
	if db == nil {
		log.Printf("[WARN] DocDB Cluster Instance (%s): not found, removing from state.", d.Id())
		d.SetId("")
		return nil
	}

	// Retrieve DB Cluster information, to determine if this Instance is a writer
	resp, err := conn.DescribeDBClusters(&docdb.DescribeDBClustersInput{
		DBClusterIdentifier: db.DBClusterIdentifier,
	})

	var dbc *docdb.DBCluster
	for _, c := range resp.DBClusters {
		if *c.DBClusterIdentifier == *db.DBClusterIdentifier {
			dbc = c
		}
	}

	if dbc == nil {
		return fmt.Errorf("Error finding DocDB Cluster (%s) for Cluster Instance (%s): %s",
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
	d.Set("dbi_resource_id", db.DbiResourceId)
	d.Set("engine_version", db.EngineVersion)
	d.Set("engine", db.Engine)
	d.Set("identifier", db.DBInstanceIdentifier)
	d.Set("instance_class", db.DBInstanceClass)
	d.Set("kms_key_id", db.KmsKeyId)
	d.Set("preferred_backup_window", db.PreferredBackupWindow)
	d.Set("preferred_maintenance_window", db.PreferredMaintenanceWindow)
	d.Set("promotion_tier", db.PromotionTier)
	d.Set("publicly_accessible", db.PubliclyAccessible)
	d.Set("storage_encrypted", db.StorageEncrypted)
	d.Set("ca_cert_identifier", db.CACertificateIdentifier)

	tags, err := keyvaluetags.DocdbListTags(conn, d.Get("arn").(string))

	if err != nil {
		return fmt.Errorf("error listing tags for DocumentDB Cluster Instance (%s): %s", d.Get("arn").(string), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsDocDBClusterInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).docdbconn
	requestUpdate := false

	req := &docdb.ModifyDBInstanceInput{
		ApplyImmediately:     aws.Bool(d.Get("apply_immediately").(bool)),
		DBInstanceIdentifier: aws.String(d.Id()),
	}

	if d.HasChange("instance_class") {
		req.DBInstanceClass = aws.String(d.Get("instance_class").(string))
		requestUpdate = true
	}

	if d.HasChange("preferred_maintenance_window") {
		req.PreferredMaintenanceWindow = aws.String(d.Get("preferred_maintenance_window").(string))
		requestUpdate = true
	}

	if d.HasChange("auto_minor_version_upgrade") {
		req.AutoMinorVersionUpgrade = aws.Bool(d.Get("auto_minor_version_upgrade").(bool))
		requestUpdate = true
	}

	if d.HasChange("promotion_tier") {
		req.PromotionTier = aws.Int64(int64(d.Get("promotion_tier").(int)))
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
			Pending:    resourceAwsDocDBClusterInstanceCreateUpdatePendingStates,
			Target:     []string{"available"},
			Refresh:    resourceAwsDocDBInstanceStateRefreshFunc(d.Id(), conn),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			MinTimeout: 10 * time.Second,
			Delay:      30 * time.Second, // Wait 30 secs before starting
		}

		// Wait, catching any errors
		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("error waiting for DocDB Instance (%s) update: %s", d.Id(), err)
		}

	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.DocdbUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating DocumentDB Cluster Instance (%s) tags: %s", d.Get("arn").(string), err)
		}

	}

	return resourceAwsDocDBClusterInstanceRead(d, meta)
}

func resourceAwsDocDBClusterInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).docdbconn

	log.Printf("[DEBUG] DocDB Cluster Instance destroy: %v", d.Id())

	opts := docdb.DeleteDBInstanceInput{DBInstanceIdentifier: aws.String(d.Id())}

	log.Printf("[DEBUG] DocDB Cluster Instance destroy configuration: %s", opts)
	if _, err := conn.DeleteDBInstance(&opts); err != nil {
		return fmt.Errorf("error deleting DocDB Instance (%s): %s", d.Id(), err)
	}

	// re-uses db_instance refresh func
	log.Println("[INFO] Waiting for DocDB Cluster Instance to be destroyed")
	stateConf := &resource.StateChangeConf{
		Pending:    resourceAwsDocDBClusterInstanceDeletePendingStates,
		Target:     []string{},
		Refresh:    resourceAwsDocDBInstanceStateRefreshFunc(d.Id(), conn),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error waiting for DocDB Instance (%s) deletion: %s", d.Id(), err)
	}

	return nil

}

func resourceAwsDocDBInstanceStateRefreshFunc(id string, conn *docdb.DocDB) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		v, err := resourceAwsDocDBInstanceRetrieve(id, conn)

		if err != nil {
			log.Printf("Error on retrieving DocDB Instance when waiting: %s", err)
			return nil, "", err
		}

		if v == nil {
			return nil, "", nil
		}

		if v.DBInstanceStatus != nil {
			log.Printf("[DEBUG] DocDB Instance status for instance %s: %s", id, *v.DBInstanceStatus)
		}

		return v, *v.DBInstanceStatus, nil
	}
}

// resourceAwsDocDBInstanceRetrieve fetches DBInstance information from the AWS
// API. It returns an error if there is a communication problem or unexpected
// error with AWS. When the DBInstance is not found, it returns no error and a
// nil pointer.
func resourceAwsDocDBInstanceRetrieve(id string, conn *docdb.DocDB) (*docdb.DBInstance, error) {
	opts := docdb.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(id),
	}

	log.Printf("[DEBUG] DocDB Instance describe configuration: %#v", opts)

	resp, err := conn.DescribeDBInstances(&opts)
	if err != nil {
		if isAWSErr(err, docdb.ErrCodeDBInstanceNotFoundFault, "") {
			return nil, nil
		}
		return nil, fmt.Errorf("Error retrieving DB Instances: %s", err)
	}

	if len(resp.DBInstances) != 1 || resp.DBInstances[0] == nil || aws.StringValue(resp.DBInstances[0].DBInstanceIdentifier) != id {
		return nil, nil
	}

	return resp.DBInstances[0], nil
}

var resourceAwsDocDBClusterInstanceCreateUpdatePendingStates = []string{
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

var resourceAwsDocDBClusterInstanceDeletePendingStates = []string{
	"configuring-log-exports",
	"modifying",
	"deleting",
}
