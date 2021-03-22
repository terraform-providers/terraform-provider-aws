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

// Route 53 Key Signing Key can only be enabled with KMS Keys in specific regions,

// testAccRoute53KeySigningKeyRegion is the chosen Route 53 Key Signing Key testing region
//
// Cached to prevent issues should multiple regions become available.
var testAccRoute53KeySigningKeyRegion string

// testAccProviderRoute53KeySigningKey is the Route 53 Key Signing Key provider instance
//
// This Provider can be used in testing code for API calls without requiring
// the use of saving and referencing specific ProviderFactories instances.
//
// testAccPreCheckRoute53KeySigningKey(t) must be called before using this provider instance.
var testAccProviderRoute53KeySigningKey *schema.Provider

// testAccProviderRoute53KeySigningKeyConfigure ensures the provider is only configured once
var testAccProviderRoute53KeySigningKeyConfigure sync.Once

// testAccPreCheckRoute53KeySigningKey verifies AWS credentials and that Route 53 Key Signing Key is supported
func testAccPreCheckRoute53KeySigningKey(t *testing.T) {
	testAccPartitionHasServicePreCheck(route53.EndpointsID, t)

	region := testAccGetRoute53KeySigningKeyRegion()

	if region == "" {
		t.Skip("Route 53 Key Signing Key not available in this AWS Partition")
	}

	// Since we are outside the scope of the Terraform configuration we must
	// call Configure() to properly initialize the provider configuration.
	testAccProviderRoute53KeySigningKeyConfigure.Do(func() {
		testAccProviderRoute53KeySigningKey = Provider()

		config := map[string]interface{}{
			"region": region,
		}

		diags := testAccProviderRoute53KeySigningKey.Configure(context.Background(), terraform.NewResourceConfigRaw(config))

		if diags != nil && diags.HasError() {
			for _, d := range diags {
				if d.Severity == diag.Error {
					t.Fatalf("error configuring Route 53 Key Signing Key provider: %s", d.Summary)
				}
			}
		}
	})
}

// testAccRoute53KeySigningKeyRegionProviderConfig is the Terraform provider configuration for Route 53 Key Signing Key region testing
//
// Testing Route 53 Key Signing Key assumes no other provider configurations
// are necessary and overwrites the "aws" provider configuration.
func testAccRoute53KeySigningKeyRegionProviderConfig() string {
	return testAccRegionalProviderConfig(testAccGetRoute53KeySigningKeyRegion())
}

// testAccGetRoute53KeySigningKeyRegion returns the Route 53 Key Signing Key region for testing
func testAccGetRoute53KeySigningKeyRegion() string {
	if testAccRoute53KeySigningKeyRegion != "" {
		return testAccRoute53KeySigningKeyRegion
	}

	// AWS Commercial: https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/dns-configuring-dnssec-cmk-requirements.html
	// AWS GovCloud (US) - not available yet: https://docs.aws.amazon.com/govcloud-us/latest/UserGuide/govcloud-r53.html
	// AWS China - not available yet: https://docs.amazonaws.cn/en_us/aws/latest/userguide/route53.html
	switch testAccGetPartition() {
	case endpoints.AwsPartitionID:
		testAccRoute53KeySigningKeyRegion = endpoints.UsEast1RegionID
	}

	return testAccRoute53KeySigningKeyRegion
}
