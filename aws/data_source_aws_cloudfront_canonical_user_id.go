package aws

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// See https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/AccessLogs.html
var defaultCloudfrontDeliveryCanonicalId = "c4c1ede66af53448b93c283ce9448c4ba468c9432aa01d700d3878632f77d2d0"

// See https://www.amazonaws.cn/en/about-aws/china/#:~:text=AWS%20China%20(Beijing)%20Region%20and,AWS%20Regions%20located%20within%20China.
var chinaRegions = []string{"cn-north-1", "cn-northwest-1"}

// See https://docs.amazonaws.cn/en_us/aws/latest/userguide/cloudfront.html#feature-diff
var chinaCloudfrontDeliveryCanonicalId = "a52cb28745c0c06e84ec548334e44bfa7fc2a85c54af20cd59e4969344b7af56"

func dataSourceAwsCloudfrontCanonicalId() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsCloudfrontCanonicalIdRead,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceAwsCloudfrontCanonicalIdRead(d *schema.ResourceData, meta interface{}) error {
	canonicalId := defaultCloudfrontDeliveryCanonicalId

	region := meta.(*AWSClient).region
	if v, ok := d.GetOk("region"); ok {
		region = v.(string)
	}
	for _, r := range chinaRegions {
		if r == region {
			canonicalId = chinaCloudfrontDeliveryCanonicalId
			break
		}
	}

	d.SetId(canonicalId)
	return nil
}
