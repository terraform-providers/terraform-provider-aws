package aws

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsCurReportDefinition() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsCurReportDefinitionRead,

		Schema: map[string]*schema.Schema{
			"report_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"time_unit": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"format": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"compression": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"additional_schema_elements": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Computed: true,
			},
			"s3_bucket": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"s3_prefix": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"s3_region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"additional_artifacts": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Computed: true,
			},
			"refresh_closed_reports": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"report_versioning": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsCurReportDefinitionRead(d *schema.ResourceData, meta interface{}) error {
	d.SetId(d.Get("report_name").(string))
	return resourceAwsCurReportDefinitionRead(d, meta)
}
