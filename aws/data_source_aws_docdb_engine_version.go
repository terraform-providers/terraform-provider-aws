package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsDocdbEngineVersion() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsDocdbEngineVersionRead,
		Schema: map[string]*schema.Schema{
			"engine": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "docdb",
			},

			"engine_description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"exportable_log_types": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Set:      schema.HashString,
			},

			"parameter_group_family": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"preferred_versions": {
				Type:          schema.TypeList,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"version"},
			},

			"supports_log_exports_to_cloudwatch": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"valid_upgrade_targets": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Set:      schema.HashString,
			},

			"version": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"preferred_versions"},
			},

			"version_description": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsDocdbEngineVersionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).docdbconn

	input := &docdb.DescribeDBEngineVersionsInput{}

	if v, ok := d.GetOk("engine"); ok {
		input.Engine = aws.String(v.(string))
	}

	if v, ok := d.GetOk("parameter_group_family"); ok {
		input.DBParameterGroupFamily = aws.String(v.(string))
	}

	if v, ok := d.GetOk("version"); ok {
		input.EngineVersion = aws.String(v.(string))
	}

	if _, ok := d.GetOk("version"); !ok {
		if _, ok := d.GetOk("preferred_versions"); !ok {
			if _, ok := d.GetOk("parameter_group_family"); !ok {
				input.DefaultOnly = aws.Bool(true)
			}
		}
	}

	log.Printf("[DEBUG] Reading DocumentDB engine versions: %v", input)
	var engineVersions []*docdb.DBEngineVersion

	err := conn.DescribeDBEngineVersionsPages(input, func(resp *docdb.DescribeDBEngineVersionsOutput, lastPage bool) bool {
		for _, engineVersion := range resp.DBEngineVersions {
			if engineVersion == nil {
				continue
			}

			engineVersions = append(engineVersions, engineVersion)
		}
		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error reading DocumentDB engine versions: %w", err)
	}

	if len(engineVersions) == 0 {
		return fmt.Errorf("no DocumentDB engine versions found")
	}

	// preferred versions
	var found *docdb.DBEngineVersion
	if l := d.Get("preferred_versions").([]interface{}); len(l) > 0 {
		for _, elem := range l {
			preferredVersion, ok := elem.(string)

			if !ok {
				continue
			}

			for _, engineVersion := range engineVersions {
				if preferredVersion == aws.StringValue(engineVersion.EngineVersion) {
					found = engineVersion
					break
				}
			}

			if found != nil {
				break
			}
		}
	}

	if found == nil && len(engineVersions) > 1 {
		return fmt.Errorf("multiple DocumentDB engine versions (%v) match the criteria", engineVersions)
	}

	if found == nil && len(engineVersions) == 1 {
		found = engineVersions[0]
	}

	if found == nil {
		return fmt.Errorf("no DocumentDB engine versions match the criteria")
	}

	d.SetId(aws.StringValue(found.EngineVersion))

	d.Set("engine", found.Engine)
	d.Set("engine_description", found.DBEngineDescription)
	d.Set("exportable_log_types", found.ExportableLogTypes)
	d.Set("parameter_group_family", found.DBParameterGroupFamily)
	d.Set("supports_log_exports_to_cloudwatch", found.SupportsLogExportsToCloudwatchLogs)

	var upgradeTargets []string
	for _, ut := range found.ValidUpgradeTarget {
		upgradeTargets = append(upgradeTargets, aws.StringValue(ut.EngineVersion))
	}
	d.Set("valid_upgrade_targets", upgradeTargets)

	d.Set("version", found.EngineVersion)
	d.Set("version_description", found.DBEngineVersionDescription)

	return nil
}
