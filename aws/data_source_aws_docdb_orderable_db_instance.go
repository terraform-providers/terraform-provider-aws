package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsDocdbOrderableDbInstance() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsDocdbOrderableDbInstanceRead,
		Schema: map[string]*schema.Schema{
			"availability_zones": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"engine": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "docdb",
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
				Default:  "na",
			},

			"preferred_instance_classes": {
				Type:          schema.TypeList,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"instance_class"},
			},

			"vpc": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsDocdbOrderableDbInstanceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).docdbconn

	input := &docdb.DescribeOrderableDBInstanceOptionsInput{}

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

	log.Printf("[DEBUG] Reading DocDB Orderable DB Instance Classes: %v", input)
	var instanceClassResults []*docdb.OrderableDBInstanceOption

	err := conn.DescribeOrderableDBInstanceOptionsPages(input, func(resp *docdb.DescribeOrderableDBInstanceOptionsOutput, lastPage bool) bool {
		for _, instanceOption := range resp.OrderableDBInstanceOptions {
			if instanceOption == nil {
				continue
			}

			instanceClassResults = append(instanceClassResults, instanceOption)
		}
		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error reading DocDB orderable DB instance options: %w", err)
	}

	if len(instanceClassResults) == 0 {
		return fmt.Errorf("no DocDB Orderable DB Instance options found matching criteria; try different search")
	}

	// preferred classes
	var found *docdb.OrderableDBInstanceOption
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
		return fmt.Errorf("multiple DocDB DB Instance Classes (%v) match the criteria; try a different search", instanceClassResults)
	}

	if found == nil && len(instanceClassResults) == 1 {
		found = instanceClassResults[0]
	}

	if found == nil {
		return fmt.Errorf("no DocDB DB Instance Classes match the criteria; try a different search")
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
	d.Set("vpc", found.Vpc)

	return nil
}
