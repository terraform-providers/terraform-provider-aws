package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsCloudFrontDistribution() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsCloudFrontDistributionRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"in_progress_validation_batches": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"hosted_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchema(),
		},
	}
}

func dataSourceAwsCloudFrontDistributionRead(d *schema.ResourceData, meta interface{}) error {
	d.SetId(d.Get("id").(string))
	conn := meta.(*AWSClient).cloudfrontconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &cloudfront.GetDistributionInput{
		Id: aws.String(d.Id()),
	}

	output, err := conn.GetDistribution(input)
	if err != nil {
		return fmt.Errorf("error getting CloudFront Distribution (%s): %w", d.Id(), err)
	}
	if output == nil {
		return fmt.Errorf("error getting CloudFront Distribution (%s): empty response", d.Id())
	}
	d.Set("etag", output.ETag)
	if distribution := output.Distribution; distribution != nil {
		d.Set("arn", distribution.ARN)
		d.Set("domain_name", distribution.DomainName)
		d.Set("in_progress_validation_batches", distribution.InProgressInvalidationBatches)
		d.Set("last_modified_time", aws.String(distribution.LastModifiedTime.String()))
		d.Set("status", distribution.Status)
		if distributionConfig := distribution.DistributionConfig; distributionConfig != nil {
			d.Set("enabled", distributionConfig.Enabled)
		}
	}
	tags, err := keyvaluetags.CloudfrontListTags(conn, d.Get("arn").(string))
	if err != nil {
		return fmt.Errorf("error listing tags for CloudFront Distribution (%s): %w", d.Id(), err)
	}
	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("hosted_zone_id", cloudFrontRoute53ZoneID)
	return nil
}
