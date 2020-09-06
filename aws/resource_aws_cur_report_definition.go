package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/costandusagereportservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
				ValidateFunc: validation.StringInSlice(
					costandusagereportservice.TimeUnit_Values(),
					false,
				),
			},
			"format": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice(
					costandusagereportservice.ReportFormat_Values(),
					false,
				),
			},
			"compression": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice(
					costandusagereportservice.CompressionFormat_Values(),
					false,
				),
			},
			"additional_schema_elements": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice(
						costandusagereportservice.SchemaElement_Values(),
						false,
					),
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
					ValidateFunc: validation.StringInSlice(
						costandusagereportservice.AdditionalArtifact_Values(),
						false,
					),
				},
				Set:      schema.HashString,
				Optional: true,
				ForceNew: true,
			},
			"refresh_closed_reports": {
				Type:     schema.TypeBool,
				ForceNew: true,
				Default:  true,
				Optional: true,
			},
			"report_versioning": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Default:  costandusagereportservice.ReportVersioningCreateNewReport,
				ValidateFunc: validation.StringInSlice(
					costandusagereportservice.ReportVersioning_Values(),
					false,
				),
			},
		},
	}
}

func resourceAwsCurReportDefinitionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).costandusagereportconn

	additionalArtifacts := expandStringSet(d.Get("additional_artifacts").(*schema.Set))
	compression := aws.String(d.Get("compression").(string))
	format := aws.String(d.Get("format").(string))
	prefix := aws.String(d.Get("s3_prefix").(string))
	reportVersioning := aws.String(d.Get("report_versioning").(string))

	additionalArtifactsList := make([]string, 0)
	for i := 0; i < len(additionalArtifacts); i++ {
		additionalArtifactsList = append(additionalArtifactsList, *additionalArtifacts[i])
	}

	err := checkAwsCurReportDefinitionPropertyCombination(
		additionalArtifactsList,
		*compression,
		*format,
		*prefix,
		*reportVersioning,
	)

	if err != nil {
		return err
	}

	reportName := d.Get("report_name").(string)

	reportDefinition := &costandusagereportservice.ReportDefinition{
		ReportName:               aws.String(reportName),
		TimeUnit:                 aws.String(d.Get("time_unit").(string)),
		Format:                   format,
		Compression:              compression,
		AdditionalSchemaElements: expandStringSet(d.Get("additional_schema_elements").(*schema.Set)),
		S3Bucket:                 aws.String(d.Get("s3_bucket").(string)),
		S3Prefix:                 prefix,
		S3Region:                 aws.String(d.Get("s3_region").(string)),
		AdditionalArtifacts:      additionalArtifacts,
		RefreshClosedReports:     aws.Bool(d.Get("refresh_closed_reports").(bool)),
		ReportVersioning:         reportVersioning,
	}

	reportDefinitionInput := &costandusagereportservice.PutReportDefinitionInput{
		ReportDefinition: reportDefinition,
	}
	log.Printf("[DEBUG] Creating AWS Cost and Usage Report Definition : %v", reportDefinitionInput)

	_, err = conn.PutReportDefinition(reportDefinitionInput)
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
	d.Set("refresh_closed_reports", matchingReportDefinition.RefreshClosedReports)
	d.Set("report_versioning", matchingReportDefinition.ReportVersioning)
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

func checkAwsCurReportDefinitionPropertyCombination(additionalArtifacts []string, compression string, format string, prefix string, reportVersioning string) error {
	// perform various combination checks, AWS API unhelpfully just returns an empty ValidationException
	// these combinations have been determined from the Create Report AWS Console Web Form

	hasAthena := false

	for _, artifact := range additionalArtifacts {
		if artifact == costandusagereportservice.AdditionalArtifactAthena {
			hasAthena = true
			break
		}
	}

	if hasAthena {
		if len(additionalArtifacts) > 1 {
			return fmt.Errorf(
				"When %s exists within additional_artifacts, no other artifact type can be declared",
				costandusagereportservice.AdditionalArtifactAthena,
			)
		}

		if len(prefix) == 0 {
			return fmt.Errorf(
				"When %s exists within additional_artifacts, prefix cannot be empty",
				costandusagereportservice.AdditionalArtifactAthena,
			)
		}

		if reportVersioning != costandusagereportservice.ReportVersioningOverwriteReport {
			return fmt.Errorf(
				"When %s exists within additional_artifacts, report_versioning must be %s",
				costandusagereportservice.AdditionalArtifactAthena,
				costandusagereportservice.ReportVersioningOverwriteReport,
			)
		}

		if format != costandusagereportservice.ReportFormatParquet {
			return fmt.Errorf(
				"When %s exists within additional_artifacts, both format and compression must be %s",
				costandusagereportservice.AdditionalArtifactAthena,
				costandusagereportservice.ReportFormatParquet,
			)
		}
	} else if len(additionalArtifacts) > 0 && (format == costandusagereportservice.ReportFormatParquet) {
		return fmt.Errorf(
			"When additional_artifacts includes %s and/or %s, format must not be %s",
			costandusagereportservice.AdditionalArtifactQuicksight,
			costandusagereportservice.AdditionalArtifactRedshift,
			costandusagereportservice.ReportFormatParquet,
		)
	}

	if format == costandusagereportservice.ReportFormatParquet {
		if compression != costandusagereportservice.CompressionFormatParquet {
			return fmt.Errorf(
				"When format is %s, compression must also be %s",
				costandusagereportservice.ReportFormatParquet,
				costandusagereportservice.CompressionFormatParquet,
			)
		}
	} else {
		if compression == costandusagereportservice.CompressionFormatParquet {
			return fmt.Errorf(
				"When format is %s, compression must not be %s",
				format,
				costandusagereportservice.CompressionFormatParquet,
			)
		}
	}
	// end checks

	return nil
}
