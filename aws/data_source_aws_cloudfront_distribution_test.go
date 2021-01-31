package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSDataSourceCloudFrontDistribution_basic(t *testing.T) {
	dataSourceName := "data.aws_cloudfront_distribution.test"
	resourceName := "aws_cloudfront_distribution.s3_distribution"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontDistributionData,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "domain_name", resourceName, "domain_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestCheckResourceAttrPair(dataSourceName, "hosted_zone_id", resourceName, "hosted_zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "in_progress_validation_batches", resourceName, "in_progress_validation_batches"),
					resource.TestCheckResourceAttrPair(dataSourceName, "last_modified_time", resourceName, "last_modified_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, "status", resourceName, "status"),
				),
			},
		},
	})
}

var testAccAWSCloudFrontDistributionData = fmt.Sprintf(`
%s

data "aws_cloudfront_distribution" "test" {
  id = aws_cloudfront_distribution.s3_distribution.id
}
`, fmt.Sprintf(testAccAWSCloudFrontDistributionS3ConfigWithTags, acctest.RandInt(), originBucket, logBucket, testAccAWSCloudFrontDistributionRetainConfig()))
