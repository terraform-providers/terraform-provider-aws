package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// See https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/AccessLogs.html
var awslogsdeliveryCanonicalId = "c4c1ede66af53448b93c283ce9448c4ba468c9432aa01d700d3878632f77d2d0"

func dataSourceAwsAwslogsdeliveryCanonicalId() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsAwslogsdeliveryCanonicalIdRead,

		Schema: map[string]*schema.Schema{},
	}
}

func dataSourceAwsAwslogsdeliveryCanonicalIdRead(d *schema.ResourceData, meta interface{}) error {
	d.SetId(aws.StringValue(&awslogsdeliveryCanonicalId))
	return nil
}
