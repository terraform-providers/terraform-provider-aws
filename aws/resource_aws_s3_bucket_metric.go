package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	tfs3 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/s3"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/s3/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsS3BucketMetric() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsS3BucketMetricPut,
		Read:   resourceAwsS3BucketMetricRead,
		Update: resourceAwsS3BucketMetricPut,
		Delete: resourceAwsS3BucketMetricDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
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
							Elem:         &schema.Schema{Type: schema.TypeString},
							AtLeastOneOf: filterAtLeastOneOfKeys,
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsS3BucketMetricPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3conn
	bucket := d.Get("bucket").(string)
	name := d.Get("name").(string)

	metricsConfiguration := &s3.MetricsConfiguration{
		Id: aws.String(name),
	}

	if v, ok := d.GetOk("filter"); ok {
		filterList := v.([]interface{})
		if filterMap, ok := filterList[0].(map[string]interface{}); ok {
			metricsConfiguration.Filter = expandS3MetricsFilter(filterMap)
		}
	}

	input := &s3.PutBucketMetricsConfigurationInput{
		Bucket:               aws.String(bucket),
		Id:                   aws.String(name),
		MetricsConfiguration: metricsConfiguration,
	}

	log.Printf("[DEBUG] Putting S3 Bucket Metrics Configuration: %s", input)
	err := resource.Retry(waiter.PropagationTimeout, func() *resource.RetryError {
		_, err := conn.PutBucketMetricsConfiguration(input)

		if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.PutBucketMetricsConfiguration(input)
	}

	if err != nil {
		return fmt.Errorf("error putting S3 Bucket Metrics Configuration: %w", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", bucket, name))

	return resourceAwsS3BucketMetricRead(d, meta)
}

func resourceAwsS3BucketMetricDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3conn

	bucket, name, err := resourceAwsS3BucketMetricParseID(d.Id())
	if err != nil {
		return err
	}

	input := &s3.DeleteBucketMetricsConfigurationInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(name),
	}

	log.Printf("[DEBUG] Deleting S3 Bucket Metrics Configuration: %s", input)
	_, err = conn.DeleteBucketMetricsConfiguration(input)

	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		return nil
	}

	if tfawserr.ErrCodeEquals(err, tfs3.ErrCodeNoSuchConfiguration) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting S3 Bucket Metrics Configuration (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceAwsS3BucketMetricRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3conn

	bucket, name, err := resourceAwsS3BucketMetricParseID(d.Id())
	if err != nil {
		return err
	}

	d.Set("bucket", bucket)
	d.Set("name", name)

	input := &s3.GetBucketMetricsConfigurationInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(name),
	}

	log.Printf("[DEBUG] Reading S3 Bucket Metrics Configuration: %s", input)
	output, err := conn.GetBucketMetricsConfiguration(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket Metrics Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, tfs3.ErrCodeNoSuchConfiguration) {
		log.Printf("[WARN] S3 Bucket Metrics Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading S3 Bucket Metrics Configuration (%s): %w", d.Id(), err)
	}

	if output == nil || output.MetricsConfiguration == nil {
		return fmt.Errorf("error reading S3 Bucket Metrics Configuration (%s): empty response", d.Id())
	}

	if output.MetricsConfiguration.Filter != nil {
		if err := d.Set("filter", []interface{}{flattenS3MetricsFilter(output.MetricsConfiguration.Filter)}); err != nil {
			return err
		}
	}

	return nil
}

func expandS3MetricsFilter(m map[string]interface{}) *s3.MetricsFilter {
	var prefix string
	if v, ok := m["prefix"]; ok {
		prefix = v.(string)
	}

	var tags []*s3.Tag
	if v, ok := m["tags"]; ok {
		tags = keyvaluetags.New(v).IgnoreAws().S3Tags()
	}

	metricsFilter := &s3.MetricsFilter{}
	if prefix != "" && len(tags) > 0 {
		metricsFilter.And = &s3.MetricsAndOperator{
			Prefix: aws.String(prefix),
			Tags:   tags,
		}
	} else if len(tags) > 1 {
		metricsFilter.And = &s3.MetricsAndOperator{
			Tags: tags,
		}
	} else if len(tags) == 1 {
		metricsFilter.Tag = tags[0]
	} else {
		metricsFilter.Prefix = aws.String(prefix)
	}
	return metricsFilter
}

func flattenS3MetricsFilter(metricsFilter *s3.MetricsFilter) map[string]interface{} {
	m := make(map[string]interface{})

	if metricsFilter.And != nil {
		and := *metricsFilter.And
		if and.Prefix != nil {
			m["prefix"] = *and.Prefix
		}
		if and.Tags != nil {
			m["tags"] = keyvaluetags.S3KeyValueTags(and.Tags).IgnoreAws().Map()
		}
	} else if metricsFilter.Prefix != nil {
		m["prefix"] = *metricsFilter.Prefix
	} else if metricsFilter.Tag != nil {
		tags := []*s3.Tag{
			metricsFilter.Tag,
		}
		m["tags"] = keyvaluetags.S3KeyValueTags(tags).IgnoreAws().Map()
	}
	return m
}

func resourceAwsS3BucketMetricParseID(id string) (string, string, error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("please make sure the ID is in the form BUCKET:NAME (i.e. my-bucket:EntireBucket")
	}
	bucket := idParts[0]
	name := idParts[1]
	return bucket, name, nil
}
