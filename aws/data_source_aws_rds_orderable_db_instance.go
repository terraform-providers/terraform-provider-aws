package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsRdsOrderableDbInstance() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsRdsOrderableDbInstanceRead,
		Schema: map[string]*schema.Schema{
			"availability_zone_group": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"availability_zones": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"engine": {
				Type:     schema.TypeString,
				Required: true,
			},

			"engine_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"instance_class": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"license_model": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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

			"outpost_capable": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"preferred_instance_classes": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"preferred_engine_versions": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"read_replica_capable": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"storage_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"supported_engine_modes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"supports_enhanced_monitoring": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"supports_global_databases": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"supports_iam_database_authentication": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"supports_iops": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"supports_kerberos_authentication": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"supports_performance_insights": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"supports_storage_autoscaling": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"supports_storage_encryption": {
				Type:     schema.TypeBool,
				Optional: true,
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

func dataSourceAwsRdsOrderableDbInstanceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	input := &rds.DescribeOrderableDBInstanceOptionsInput{}

	if v, ok := d.GetOk("availability_zone_group"); ok {
		input.AvailabilityZoneGroup = aws.String(v.(string))
	}

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

	log.Printf("[DEBUG] Reading RDS Orderable DB Instance Options: %v", input)
	var instanceClassResults []*rds.OrderableDBInstanceOption

	err := conn.DescribeOrderableDBInstanceOptionsPages(input, func(resp *rds.DescribeOrderableDBInstanceOptionsOutput, lastPage bool) bool {
		for _, instanceOption := range resp.OrderableDBInstanceOptions {
			if instanceOption == nil {
				continue
			}

			if v, ok := d.GetOk("storage_type"); ok {
				if aws.StringValue(instanceOption.StorageType) != v.(string) {
					continue
				}
			}

			if v, ok := d.GetOk("supports_enhanced_monitoring"); ok {
				if aws.BoolValue(instanceOption.SupportsEnhancedMonitoring) != v.(bool) {
					continue
				}
			}

			if v, ok := d.GetOk("supports_global_databases"); ok {
				if aws.BoolValue(instanceOption.SupportsGlobalDatabases) != v.(bool) {
					continue
				}
			}

			if v, ok := d.GetOk("supports_iam_database_authentication"); ok {
				if aws.BoolValue(instanceOption.SupportsIAMDatabaseAuthentication) != v.(bool) {
					continue
				}
			}

			if v, ok := d.GetOk("supports_iops"); ok {
				if aws.BoolValue(instanceOption.SupportsIops) != v.(bool) {
					continue
				}
			}

			if v, ok := d.GetOk("supports_kerberos_authentication"); ok {
				if aws.BoolValue(instanceOption.SupportsKerberosAuthentication) != v.(bool) {
					continue
				}
			}

			if v, ok := d.GetOk("supports_performance_insights"); ok {
				if aws.BoolValue(instanceOption.SupportsPerformanceInsights) != v.(bool) {
					continue
				}
			}

			if v, ok := d.GetOk("supports_storage_autoscaling"); ok {
				if aws.BoolValue(instanceOption.SupportsStorageAutoscaling) != v.(bool) {
					continue
				}
			}

			if v, ok := d.GetOk("supports_storage_encryption"); ok {
				if aws.BoolValue(instanceOption.SupportsStorageEncryption) != v.(bool) {
					continue
				}
			}

			instanceClassResults = append(instanceClassResults, instanceOption)
		}
		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error reading RDS orderable DB instance options: %w", err)
	}

	if len(instanceClassResults) == 0 {
		return fmt.Errorf("no RDS Orderable DB Instance options found matching criteria; try different search")
	}

	// preferred classes/versions
	var found *rds.OrderableDBInstanceOption
	l := d.Get("preferred_instance_classes").([]interface{})
	v := d.Get("preferred_engine_versions").([]interface{})
	if len(l) > 0 && len(v) > 0 {
		for _, elem := range l {
			preferredInstanceClass, ok := elem.(string)

			if !ok {
				continue
			}

			for _, ver := range v {
				preferredEngineVersion, ok := ver.(string)

				if !ok {
					continue
				}

				for _, instanceClassResult := range instanceClassResults {
					if preferredInstanceClass == aws.StringValue(instanceClassResult.DBInstanceClass) &&
						preferredEngineVersion == aws.StringValue(instanceClassResult.EngineVersion) {
						found = instanceClassResult
						break
					}
				}

				if found != nil {
					break
				}
			}

			if found != nil {
				break
			}
		}
	} else if len(l) > 0 {
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
	} else if len(v) > 0 {
		for _, ver := range v {
			preferredEngineVersion, ok := ver.(string)

			if !ok {
				continue
			}

			for _, instanceClassResult := range instanceClassResults {
				if preferredEngineVersion == aws.StringValue(instanceClassResult.EngineVersion) {
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
		return fmt.Errorf("multiple RDS DB Instance Classes (%v) match the criteria; try a different search", instanceClassResults)
	}

	if found == nil && len(instanceClassResults) == 1 {
		found = instanceClassResults[0]
	}

	if found == nil {
		return fmt.Errorf("no RDS DB Instance Classes match the criteria; try a different search")
	}

	d.SetId(aws.StringValue(found.DBInstanceClass))

	d.Set("instance_class", found.DBInstanceClass)
	d.Set("availability_zone_group", found.AvailabilityZoneGroup)

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
	d.Set("outpost_capable", found.OutpostCapable)
	d.Set("read_replica_capable", found.ReadReplicaCapable)
	d.Set("storage_type", found.StorageType)
	d.Set("supported_engine_modes", found.SupportedEngineModes)
	d.Set("supports_enhanced_monitoring", found.SupportsEnhancedMonitoring)
	d.Set("supports_global_databases", found.SupportsGlobalDatabases)
	d.Set("supports_iam_database_authentication", found.SupportsIAMDatabaseAuthentication)
	d.Set("supports_iops", found.SupportsIops)
	d.Set("supports_kerberos_authentication", found.SupportsKerberosAuthentication)
	d.Set("supports_performance_insights", found.SupportsPerformanceInsights)
	d.Set("supports_storage_autoscaling", found.SupportsStorageAutoscaling)
	d.Set("supports_storage_encryption", found.SupportsStorageEncryption)
	d.Set("vpc", found.Vpc)

	return nil
}
