package aws

import (
	"context"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// Route 53 Query Logging can only be enabled with CloudWatch Log Groups in specific regions,

// testAccRoute53QueryLogRegion is the chosen Route 53 Query Logging testing region
//
// Cached to prevent issues should multiple regions become available.
var testAccRoute53QueryLogRegion string

// testAccProviderRoute53QueryLog is the Route 53 Query Logging provider instance
//
// This Provider can be used in testing code for API calls without requiring
// the use of saving and referencing specific ProviderFactories instances.
//
// testAccPreCheckRoute53QueryLog(t) must be called before using this provider instance.
var testAccProviderRoute53QueryLog *schema.Provider

// testAccProviderRoute53QueryLogConfigure ensures the provider is only configured once
var testAccProviderRoute53QueryLogConfigure sync.Once

// testAccPreCheckRoute53QueryLog verifies AWS credentials and that Route 53 Query Logging is supported
func testAccPreCheckRoute53QueryLog(t *testing.T) {
	testAccPartitionHasServicePreCheck(route53.EndpointsID, t)

	region := testAccGetRoute53QueryLogRegion()

	if region == "" {
		t.Skip("Route 53 Query Log not available in this AWS Partition")
	}

	// Since we are outside the scope of the Terraform configuration we must
	// call Configure() to properly initialize the provider configuration.
	testAccProviderRoute53QueryLogConfigure.Do(func() {
		testAccProviderRoute53QueryLog = Provider()

		config := map[string]interface{}{
			"region": region,
		}

		diags := testAccProviderRoute53QueryLog.Configure(context.Background(), terraform.NewResourceConfigRaw(config))

		if diags != nil && diags.HasError() {
			for _, d := range diags {
				if d.Severity == diag.Error {
					t.Fatalf("error configuring Route 53 Query Logging provider: %s", d.Summary)
				}
			}
		}
	})
}

// testAccRoute53QueryLogRegionProviderConfig is the Terraform provider configuration for Route 53 Query Logging region testing
//
// Testing Route 53 Query Logging assumes no other provider configurations
// are necessary and overwrites the "aws" provider configuration.
func testAccRoute53QueryLogRegionProviderConfig() string {
	return testAccRegionalProviderConfig(testAccGetRoute53QueryLogRegion())
}

// testAccGetRoute53QueryLogRegion returns the Route 53 Query Logging region for testing
func testAccGetRoute53QueryLogRegion() string {
	if testAccRoute53QueryLogRegion != "" {
		return testAccRoute53QueryLogRegion
	}

	// AWS Commercial: https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/query-logs.html
	// AWS GovCloud (US) - only private DNS: https://docs.aws.amazon.com/govcloud-us/latest/UserGuide/govcloud-r53.html
	// AWS China - not available yet: https://docs.amazonaws.cn/en_us/aws/latest/userguide/route53.html
	switch testAccGetPartition() {
	case endpoints.AwsPartitionID:
		testAccRoute53QueryLogRegion = endpoints.UsEast1RegionID
	}

	return testAccRoute53QueryLogRegion
}
