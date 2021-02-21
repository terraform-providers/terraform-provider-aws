package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsNeptuneOrderableDbInstance() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsNeptuneOrderableDbInstanceRead,
		Schema: map[string]*schema.Schema{
			"availability_zones": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"engine": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "neptune",
			},

			"engine_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"instance_class": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"preferred_instance_classes"},
			},

			"license_model": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "amazon-license",
			},

			"max_iops_per_db_instance": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"max_iops_per_gib": {
				Type:     schema.TypeFloat,
				Computed: true,
			},

			"max_storage_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"min_iops_per_db_instance": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"min_iops_per_gib": {
				Type:     schema.TypeFloat,
				Computed: true,
			},

			"min_storage_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"multi_az_capable": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"preferred_instance_classes": {
				Type:          schema.TypeList,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"instance_class"},
			},

			"read_replica_capable": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"storage_type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"supports_enhanced_monitoring": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"supports_iam_database_authentication": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"supports_iops": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"supports_performance_insights": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"supports_storage_encryption": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"vpc": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsNeptuneOrderableDbInstanceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).neptuneconn

	input := &neptune.DescribeOrderableDBInstanceOptionsInput{}

	if v, ok := d.GetOk("instance_class"); ok {
		input.DBInstanceClass = aws.String(v.(string))
	}

	if v, ok := d.GetOk("engine"); ok {
		input.Engine = aws.String(v.(string))
	}

	if v, ok := d.GetOk("engine_version"); ok {
		input.EngineVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("license_model"); ok {
		input.LicenseModel = aws.String(v.(string))
	}

	if v, ok := d.GetOk("vpc"); ok {
		input.Vpc = aws.Bool(v.(bool))
	}

	log.Printf("[DEBUG] Reading Neptune Orderable DB Instance Options: %v", input)

	var instanceClassResults []*neptune.OrderableDBInstanceOption
	err := conn.DescribeOrderableDBInstanceOptionsPages(input, func(resp *neptune.DescribeOrderableDBInstanceOptionsOutput, lastPage bool) bool {
		for _, instanceOption := range resp.OrderableDBInstanceOptions {
			if instanceOption == nil {
				continue
			}

			instanceClassResults = append(instanceClassResults, instanceOption)
		}
		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error reading Neptune orderable DB instance options: %w", err)
	}

	if len(instanceClassResults) == 0 {
		return fmt.Errorf("no Neptune Orderable DB Instance options found matching criteria; try different search")
	}

	// preferred classes
	var found *neptune.OrderableDBInstanceOption
	if l := d.Get("preferred_instance_classes").([]interface{}); len(l) > 0 {
		for _, elem := range l {
			preferredInstanceClass, ok := elem.(string)

			if !ok {
				continue
			}

			for _, instanceClassResult := range instanceClassResults {
				if preferredInstanceClass == aws.StringValue(instanceClassResult.DBInstanceClass) {
					found = instanceClassResult
					break
				}
			}

			if found != nil {
				break
			}
		}
	}

	if found == nil && len(instanceClassResults) > 1 {
		return fmt.Errorf("multiple Neptune DB Instance Classes (%v) match the criteria; try a different search", instanceClassResults)
	}

	if found == nil && len(instanceClassResults) == 1 {
		found = instanceClassResults[0]
	}

	if found == nil {
		return fmt.Errorf("no Neptune DB Instance Classes match the criteria; try a different search")
	}

	d.SetId(aws.StringValue(found.DBInstanceClass))

	d.Set("instance_class", found.DBInstanceClass)

	var availabilityZones []string
	for _, az := range found.AvailabilityZones {
		availabilityZones = append(availabilityZones, aws.StringValue(az.Name))
	}
	d.Set("availability_zones", availabilityZones)

	d.Set("engine", found.Engine)
	d.Set("engine_version", found.EngineVersion)
	d.Set("license_model", found.LicenseModel)
	d.Set("max_iops_per_db_instance", found.MaxIopsPerDbInstance)
	d.Set("max_iops_per_gib", found.MaxIopsPerGib)
	d.Set("max_storage_size", found.MaxStorageSize)
	d.Set("min_iops_per_db_instance", found.MinIopsPerDbInstance)
	d.Set("min_iops_per_gib", found.MinIopsPerGib)
	d.Set("min_storage_size", found.MinStorageSize)
	d.Set("multi_az_capable", found.MultiAZCapable)
	d.Set("read_replica_capable", found.ReadReplicaCapable)
	d.Set("storage_type", found.StorageType)
	d.Set("supports_enhanced_monitoring", found.SupportsEnhancedMonitoring)
	d.Set("supports_iam_database_authentication", found.SupportsIAMDatabaseAuthentication)
	d.Set("supports_iops", found.SupportsIops)
	d.Set("supports_performance_insights", found.SupportsPerformanceInsights)
	d.Set("supports_storage_encryption", found.SupportsStorageEncryption)
	d.Set("vpc", found.Vpc)

	return nil
}
