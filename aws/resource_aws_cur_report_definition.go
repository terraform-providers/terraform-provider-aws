package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/costandusagereportservice"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"log"
)

func resourceAwsCurReportDefinition() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCurReportDefinitionCreate,
		Read:   resourceAwsCurReportDefinitionRead,
		Delete: resourceAwsCurReportDefinitionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"report_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"time_unit": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					costandusagereportservice.TimeUnitDaily,
					costandusagereportservice.TimeUnitHourly,
				}, false),
			},
			"format": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					costandusagereportservice.ReportFormatTextOrcsv}, false),
			},
			"compression": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					costandusagereportservice.CompressionFormatGzip,
					costandusagereportservice.CompressionFormatZip,
				}, false),
			},
			"additional_schema_elements": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						costandusagereportservice.SchemaElementResources,
					}, false),
				},
				Set:      schema.HashString,
				Required: true,
				ForceNew: true,
			},
			"s3_bucket": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"s3_prefix": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"s3_region": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"additional_artifacts": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						costandusagereportservice.AdditionalArtifactQuicksight,
						costandusagereportservice.AdditionalArtifactRedshift,
					}, false),
				},
				Set:      schema.HashString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsCurReportDefinitionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).costandusagereportconn

	reportName := d.Get("report_name").(string)

	reportDefinition := &costandusagereportservice.ReportDefinition{
		ReportName:               aws.String(reportName),
		TimeUnit:                 aws.String(d.Get("time_unit").(string)),
		Format:                   aws.String(d.Get("format").(string)),
		Compression:              aws.String(d.Get("compression").(string)),
		AdditionalSchemaElements: expandStringSet(d.Get("additional_schema_elements").(*schema.Set)),
		S3Bucket:                 aws.String(d.Get("s3_bucket").(string)),
		S3Prefix:                 aws.String(d.Get("s3_prefix").(string)),
		S3Region:                 aws.String(d.Get("s3_region").(string)),
		AdditionalArtifacts:      expandStringSet(d.Get("additional_artifacts").(*schema.Set)),
	}

	reportDefinitionInput := &costandusagereportservice.PutReportDefinitionInput{
		ReportDefinition: reportDefinition,
	}
	log.Printf("[DEBUG] Creating AWS Cost and Usage Report Definition : %v", reportDefinitionInput)

	_, err := conn.PutReportDefinition(reportDefinitionInput)
	if err != nil {
		return fmt.Errorf("Error creating AWS Cost And Usage Report Definition: %s", err)
	}
	d.SetId(reportName)
	return resourceAwsCurReportDefinitionRead(d, meta)
}

func resourceAwsCurReportDefinitionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).costandusagereportconn

	reportName := *aws.String(d.Id())

	matchingReportDefinition, err := describeCurReportDefinition(conn, reportName)
	if err != nil {
		return err
	}
	if matchingReportDefinition == nil {
		log.Printf("[WARN] Report definition (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	d.SetId(aws.StringValue(matchingReportDefinition.ReportName))
	d.Set("report_name", matchingReportDefinition.ReportName)
	d.Set("time_unit", matchingReportDefinition.TimeUnit)
	d.Set("format", matchingReportDefinition.Format)
	d.Set("compression", matchingReportDefinition.Compression)
	d.Set("additional_schema_elements", aws.StringValueSlice(matchingReportDefinition.AdditionalSchemaElements))
	d.Set("s3_bucket", aws.StringValue(matchingReportDefinition.S3Bucket))
	d.Set("s3_prefix", aws.StringValue(matchingReportDefinition.S3Prefix))
	d.Set("s3_region", aws.StringValue(matchingReportDefinition.S3Region))
	d.Set("additional_artifacts", aws.StringValueSlice(matchingReportDefinition.AdditionalArtifacts))
	return nil
}

func resourceAwsCurReportDefinitionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).costandusagereportconn
	log.Printf("[DEBUG] Deleting AWS Cost and Usage Report Definition : %s", d.Id())
	_, err := conn.DeleteReportDefinition(&costandusagereportservice.DeleteReportDefinitionInput{
		ReportName: aws.String(d.Id()),
	})
	if err != nil {
		return err
	}
	return nil
}

func describeCurReportDefinition(conn *costandusagereportservice.CostandUsageReportService, reportName string) (*costandusagereportservice.ReportDefinition, error) {
	params := &costandusagereportservice.DescribeReportDefinitionsInput{}

	log.Printf("[DEBUG] Reading CurReportDefinition: %s", reportName)

	var matchingReportDefinition *costandusagereportservice.ReportDefinition
	err := conn.DescribeReportDefinitionsPages(params, func(resp *costandusagereportservice.DescribeReportDefinitionsOutput, isLast bool) bool {
		for _, reportDefinition := range resp.ReportDefinitions {
			if aws.StringValue(reportDefinition.ReportName) == reportName {
				matchingReportDefinition = reportDefinition
				return false
			}
		}
		return !isLast
	})
	if err != nil {
		return nil, err
	}
	return matchingReportDefinition, nil
}
