package aws

import (
	"context"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// testAccPricingRegion is the chosen Pricing testing region
//
// Cached to prevent issues should multiple regions become available.
var testAccPricingRegion string

// testAccProviderPricing is the Pricing provider instance
//
// This Provider can be used in testing code for API calls without requiring
// the use of saving and referencing specific ProviderFactories instances.
//
// testAccPreCheckPricing(t) must be called before using this provider instance.
var testAccProviderPricing *schema.Provider

// testAccProviderPricingConfigure ensures the provider is only configured once
var testAccProviderPricingConfigure sync.Once

// testAccPreCheckPricing verifies AWS credentials and that Pricing is supported
func testAccPreCheckPricing(t *testing.T) {
	testAccPartitionHasServicePreCheck(pricing.EndpointsID, t)

	// Since we are outside the scope of the Terraform configuration we must
	// call Configure() to properly initialize the provider configuration.
	testAccProviderPricingConfigure.Do(func() {
		testAccProviderPricing = Provider()

		config := map[string]interface{}{
			"region": testAccGetPricingRegion(),
		}

		diags := testAccProviderPricing.Configure(context.Background(), terraform.NewResourceConfigRaw(config))

		if diags != nil && diags.HasError() {
			for _, d := range diags {
				if d.Severity == diag.Error {
					t.Fatalf("error configuring Pricing provider: %s", d.Summary)
				}
			}
		}
	})
}

// testAccPricingRegionProviderConfig is the Terraform provider configuration for Pricing region testing
//
// Testing Pricing assumes no other provider configurations
// are necessary and overwrites the "aws" provider configuration.
func testAccPricingRegionProviderConfig() string {
	return testAccRegionalProviderConfig(testAccGetPricingRegion())
}

// testAccGetPricingRegion returns the Pricing region for testing
func testAccGetPricingRegion() string {
	if testAccPricingRegion != "" {
		return testAccPricingRegion
	}

	if rs, ok := endpoints.RegionsForService(endpoints.DefaultPartitions(), testAccGetPartition(), pricing.ServiceName); ok {
		// return available region (random if multiple)
		for regionID := range rs {
			testAccPricingRegion = regionID
			return testAccPricingRegion
		}
	}

	testAccPricingRegion = testAccGetRegion()

	return testAccPricingRegion
}
