package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/costandusagereportservice/finder"
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
	conn := meta.(*AWSClient).costandusagereportconn

	reportName := d.Get("report_name").(string)

	reportDefinition, err := finder.ReportDefinitionByName(conn, reportName)

	if err != nil {
		return fmt.Errorf("error reading Report Definition (%s): %w", reportName, err)
	}

	if reportDefinition == nil {
		return fmt.Errorf("error reading Report Definition (%s): not found", reportName)
	}

	d.SetId(aws.StringValue(reportDefinition.ReportName))
	d.Set("report_name", reportDefinition.ReportName)
	d.Set("time_unit", reportDefinition.TimeUnit)
	d.Set("format", reportDefinition.Format)
	d.Set("compression", reportDefinition.Compression)
	d.Set("additional_schema_elements", aws.StringValueSlice(reportDefinition.AdditionalSchemaElements))
	d.Set("s3_bucket", reportDefinition.S3Bucket)
	d.Set("s3_prefix", reportDefinition.S3Prefix)
	d.Set("s3_region", reportDefinition.S3Region)
	d.Set("additional_artifacts", aws.StringValueSlice(reportDefinition.AdditionalArtifacts))
	d.Set("refresh_closed_reports", reportDefinition.RefreshClosedReports)
	d.Set("report_versioning", reportDefinition.ReportVersioning)

	return nil
}
