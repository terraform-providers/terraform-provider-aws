package aws

import "github.com/aws/aws-sdk-go/aws/endpoints"

// testAccCloudfrontRegionProviderConfig is the Terraform provider configuration for CloudFront region testing
//
// Testing CloudFront assumes no other provider configurations
// are necessary and overwrites the "aws" provider configuration.
func testAccCloudfrontRegionProviderConfig() string {
	switch testAccGetPartition() {
	case endpoints.AwsPartitionID:
		return testAccRegionalProviderConfig(endpoints.UsEast1RegionID)
	case endpoints.AwsCnPartitionID:
		return testAccRegionalProviderConfig(endpoints.CnNorthwest1RegionID)
	default:
		return testAccRegionalProviderConfig(testAccGetRegion())
	}
}
