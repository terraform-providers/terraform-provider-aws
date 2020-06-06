package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsS3BucketAnalyticsConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsS3BucketAnalyticsConfigurationPut,
		Read:   resourceAwsS3BucketAnalyticsConfigurationRead,
		Update: resourceAwsS3BucketAnalyticsConfigurationPut,
		Delete: resourceAwsS3BucketAnalyticsConfigurationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"filter": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"prefix": {
							Type:         schema.TypeString,
							Optional:     true,
							AtLeastOneOf: filterAtLeastOneOfKeys,
						},
						"tags": {
							Type:         schema.TypeMap,
							Optional:     true,
							AtLeastOneOf: filterAtLeastOneOfKeys,
							Elem:         &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"storage_class_analysis": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data_export": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"output_schema_version": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      s3.StorageClassAnalysisSchemaVersionV1,
										ValidateFunc: validation.StringInSlice([]string{s3.StorageClassAnalysisSchemaVersionV1}, false),
									},
									"destination": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"s3_bucket_destination": {
													Type:     schema.TypeList,
													Required: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"bucket_arn": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validateArn,
															},
															"bucket_account_id": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validateAwsAccountId,
															},
															"format": {
																Type:         schema.TypeString,
																Optional:     true,
																Default:      s3.AnalyticsS3ExportFileFormatCsv,
																ValidateFunc: validation.StringInSlice([]string{s3.AnalyticsS3ExportFileFormatCsv}, false),
															},
															"prefix": {
																Type:     schema.TypeString,
																Optional: true,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

var filterAtLeastOneOfKeys = []string{"filter.0.prefix", "filter.0.tags"}

func resourceAwsS3BucketAnalyticsConfigurationPut(d *schema.ResourceData, meta interface{}) error {
	s3conn := meta.(*AWSClient).s3conn

	bucket := d.Get("bucket").(string)
	name := d.Get("name").(string)

	log.Printf("[DEBUG] S3 bucket %q, add analytics configuration %q", bucket, name)

	analyticsConfiguration := &s3.AnalyticsConfiguration{
		Id:                   aws.String(name),
		Filter:               expandS3AnalyticsFilter(d.Get("filter").([]interface{})),
		StorageClassAnalysis: expandS3StorageClassAnalysis(d.Get("storage_class_analysis").([]interface{})),
	}

	input := &s3.PutBucketAnalyticsConfigurationInput{
		Bucket:                 aws.String(bucket),
		Id:                     aws.String(name),
		AnalyticsConfiguration: analyticsConfiguration,
	}

	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err := s3conn.PutBucketAnalyticsConfiguration(input)
		if err != nil {
			if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = s3conn.PutBucketAnalyticsConfiguration(input)
	}
	if err != nil {
		return fmt.Errorf("Error adding S3 analytics configuration: %w", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", bucket, name))

	return resourceAwsS3BucketAnalyticsConfigurationRead(d, meta)
}

func resourceAwsS3BucketAnalyticsConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3conn

	bucket, name, err := resourceAwsS3BucketAnalyticsConfigurationParseID(d.Id())
	if err != nil {
		return err
	}

	d.Set("bucket", bucket)
	d.Set("name", name)

	input := &s3.GetBucketAnalyticsConfigurationInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(name),
	}

	log.Printf("[DEBUG] Reading S3 bucket analytics configuration: %s", input)
	output, err := conn.GetBucketAnalyticsConfiguration(input)
	if err != nil {
		if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") || isAWSErr(err, "NoSuchConfiguration", "The specified configuration does not exist.") {
			log.Printf("[WARN] %s S3 bucket analytics configuration not found, removing from state.", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error getting S3 Bucket Analytics Configuration %q: %w", d.Id(), err)
	}

	if err := d.Set("filter", flattenS3AnalyticsFilter(output.AnalyticsConfiguration.Filter)); err != nil {
		return fmt.Errorf("error setting filter: %w", err)
	}

	if err = d.Set("storage_class_analysis", flattenS3StorageClassAnalysis(output.AnalyticsConfiguration.StorageClassAnalysis)); err != nil {
		return fmt.Errorf("error setting storage class anyalytics: %w", err)
	}

	return nil
}

func resourceAwsS3BucketAnalyticsConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3conn

	bucket, name, err := resourceAwsS3BucketAnalyticsConfigurationParseID(d.Id())
	if err != nil {
		return err
	}

	input := &s3.DeleteBucketAnalyticsConfigurationInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(name),
	}

	log.Printf("[DEBUG] Deleting S3 bucket analytics configuration: %s", input)
	_, err = conn.DeleteBucketAnalyticsConfiguration(input)
	if err != nil {
		if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") || isAWSErr(err, "NoSuchConfiguration", "The specified configuration does not exist.") {
			return nil
		}
		return fmt.Errorf("Error deleting S3 analytics configuration: %w", err)
	}

	return waitForDeleteS3BucketAnalyticsConfiguration(conn, bucket, name, 1*time.Minute)
}

func resourceAwsS3BucketAnalyticsConfigurationParseID(id string) (string, string, error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("please make sure the ID is in the form BUCKET:NAME (i.e. my-bucket:EntireBucket")
	}
	bucket := idParts[0]
	name := idParts[1]
	return bucket, name, nil
}

func expandS3AnalyticsFilter(l []interface{}) *s3.AnalyticsFilter {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	var prefix string
	if v, ok := m["prefix"]; ok {
		prefix = v.(string)
	}

	var tags []*s3.Tag
	if v, ok := m["tags"]; ok {
		tags = keyvaluetags.New(v).IgnoreAws().S3Tags()
	}

	if prefix == "" && len(tags) == 0 {
		return nil
	}
	analyticsFilter := &s3.AnalyticsFilter{}
	if prefix != "" && len(tags) > 0 {
		analyticsFilter.And = &s3.AnalyticsAndOperator{
			Prefix: aws.String(prefix),
			Tags:   tags,
		}
	} else if len(tags) > 1 {
		analyticsFilter.And = &s3.AnalyticsAndOperator{
			Tags: tags,
		}
	} else if len(tags) == 1 {
		analyticsFilter.Tag = tags[0]
	} else {
		analyticsFilter.Prefix = aws.String(prefix)
	}
	return analyticsFilter
}

func expandS3StorageClassAnalysis(l []interface{}) *s3.StorageClassAnalysis {
	result := &s3.StorageClassAnalysis{}

	if len(l) == 0 || l[0] == nil {
		return result
	}

	m := l[0].(map[string]interface{})
	if v, ok := m["data_export"]; ok {
		dataExport := &s3.StorageClassAnalysisDataExport{}
		result.DataExport = dataExport

		foo := v.([]interface{})
		if len(foo) != 0 && foo[0] != nil {
			bar := foo[0].(map[string]interface{})
			if v, ok := bar["output_schema_version"]; ok {
				dataExport.OutputSchemaVersion = aws.String(v.(string))
			}

			dataExport.Destination = expandS3AnalyticsExportDestination(bar["destination"].([]interface{}))
		}
	}

	return result
}

func expandS3AnalyticsExportDestination(edl []interface{}) *s3.AnalyticsExportDestination {
	result := &s3.AnalyticsExportDestination{}

	if len(edl) != 0 && edl[0] != nil {
		edm := edl[0].(map[string]interface{})
		result.S3BucketDestination = expandS3AnalyticsS3BucketDestination(edm["s3_bucket_destination"].([]interface{}))
	}
	return result
}

func expandS3AnalyticsS3BucketDestination(bdl []interface{}) *s3.AnalyticsS3BucketDestination {
	result := &s3.AnalyticsS3BucketDestination{}

	if len(bdl) != 0 && bdl[0] != nil {
		bdm := bdl[0].(map[string]interface{})
		result.Bucket = aws.String(bdm["bucket_arn"].(string))
		result.Format = aws.String(bdm["format"].(string))

		if v, ok := bdm["bucket_account_id"]; ok && v != "" {
			result.BucketAccountId = aws.String(v.(string))
		}

		if v, ok := bdm["prefix"]; ok && v != "" {
			result.Prefix = aws.String(v.(string))
		}
	}

	return result
}

func flattenS3AnalyticsFilter(analyticsFilter *s3.AnalyticsFilter) []map[string]interface{} {
	if analyticsFilter == nil {
		return nil
	}

	result := make(map[string]interface{})
	if analyticsFilter.And != nil {
		and := *analyticsFilter.And
		if and.Prefix != nil {
			result["prefix"] = *and.Prefix
		}
		if and.Tags != nil {
			result["tags"] = keyvaluetags.S3KeyValueTags(and.Tags).IgnoreAws().Map()
		}
	} else if analyticsFilter.Prefix != nil {
		result["prefix"] = *analyticsFilter.Prefix
	} else if analyticsFilter.Tag != nil {
		tags := []*s3.Tag{
			analyticsFilter.Tag,
		}
		result["tags"] = keyvaluetags.S3KeyValueTags(tags).IgnoreAws().Map()
	} else {
		return nil
	}
	return []map[string]interface{}{result}
}

func flattenS3StorageClassAnalysis(storageClassAnalysis *s3.StorageClassAnalysis) []map[string]interface{} {
	if storageClassAnalysis == nil || storageClassAnalysis.DataExport == nil {
		return []map[string]interface{}{}
	}

	dataExport := storageClassAnalysis.DataExport
	de := make(map[string]interface{})
	if dataExport.OutputSchemaVersion != nil {
		de["output_schema_version"] = aws.StringValue(dataExport.OutputSchemaVersion)
	}
	if dataExport.Destination != nil {
		de["destination"] = flattenS3AnalyticsExportDestination(dataExport.Destination)
	}
	result := map[string]interface{}{
		"data_export": []interface{}{de},
	}

	return []map[string]interface{}{result}
}

func flattenS3AnalyticsExportDestination(destination *s3.AnalyticsExportDestination) []interface{} {
	if destination == nil || destination.S3BucketDestination == nil {
		return []interface{}{}
	}

	return []interface{}{
		map[string]interface{}{
			"s3_bucket_destination": flattenS3AnalyticsS3BucketDestination(destination.S3BucketDestination),
		},
	}
}

func flattenS3AnalyticsS3BucketDestination(bucketDestination *s3.AnalyticsS3BucketDestination) []interface{} {
	if bucketDestination == nil {
		return nil
	}

	result := map[string]interface{}{
		"bucket_arn": aws.StringValue(bucketDestination.Bucket),
		"format":     aws.StringValue(bucketDestination.Format),
	}
	if bucketDestination.BucketAccountId != nil {
		result["bucket_account_id"] = aws.StringValue(bucketDestination.BucketAccountId)
	}
	if bucketDestination.Prefix != nil {
		result["prefix"] = aws.StringValue(bucketDestination.Prefix)
	}

	return []interface{}{result}
}

func waitForDeleteS3BucketAnalyticsConfiguration(conn *s3.S3, bucket, name string, timeout time.Duration) error {
	err := resource.Retry(timeout, func() *resource.RetryError {
		input := &s3.GetBucketAnalyticsConfigurationInput{
			Bucket: aws.String(bucket),
			Id:     aws.String(name),
		}
		log.Printf("[DEBUG] Reading S3 bucket analytics configuration: %s", input)
		output, err := conn.GetBucketAnalyticsConfiguration(input)
		if err != nil {
			if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") || isAWSErr(err, "NoSuchConfiguration", "The specified configuration does not exist.") {
				return nil
			}
			return resource.NonRetryableError(err)
		}
		if output.AnalyticsConfiguration != nil {
			return resource.RetryableError(fmt.Errorf("S3 bucket analytics configuration exists: %v", output))
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error deleting S3 Bucket Analytics Configuration \"%s:%s\": %w", bucket, name, err)
	}
	return nil
}
