package aws

import (
	"context"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/fms"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// Firewall Management Service admin APIs are only enabled in specific regions, otherwise:
// InvalidOperationException: This operation is not supported in the 'us-west-2' region.

// testAccFmsAdminRegion is the chosen Firewall Management Service testing region
//
// Cached to prevent issues should multiple regions become available.
var testAccFmsAdminRegion string

// testAccProviderFmsAdmin is the Firewall Management Service provider instance
//
// This Provider can be used in testing code for API calls without requiring
// the use of saving and referencing specific ProviderFactories instances.
//
// testAccPreCheckFmsAdmin(t) must be called before using this provider instance.
var testAccProviderFmsAdmin *schema.Provider

// testAccProviderFmsAdminConfigure ensures the provider is only configured once
var testAccProviderFmsAdminConfigure sync.Once

// testAccPreCheckFmsAdmin verifies AWS credentials and that Firewall Management Service is supported
func testAccPreCheckFmsAdmin(t *testing.T) {
	testAccPartitionHasServicePreCheck(fms.EndpointsID, t)

	// Since we are outside the scope of the Terraform configuration we must
	// call Configure() to properly initialize the provider configuration.
	testAccProviderFmsAdminConfigure.Do(func() {
		testAccProviderFmsAdmin = Provider()

		config := map[string]interface{}{
			"region": testAccGetFmsAdminRegion(),
		}

		diags := testAccProviderFmsAdmin.Configure(context.Background(), terraform.NewResourceConfigRaw(config))

		if diags != nil && diags.HasError() {
			for _, d := range diags {
				if d.Severity == diag.Error {
					t.Fatalf("error configuring Firewall Management Service provider: %s", d.Summary)
				}
			}
		}
	})
}

// testAccFmsAdminRegionProviderConfig is the Terraform provider configuration for Firewall Management Service region testing
//
// Testing Firewall Management Service assumes no other provider configurations
// are necessary and overwrites the "aws" provider configuration.
func testAccFmsAdminRegionProviderConfig() string {
	return testAccRegionalProviderConfig(testAccGetFmsAdminRegion())
}

// testAccGetFmsAdminRegion returns the Firewall Management Service region for testing
func testAccGetFmsAdminRegion() string {
	if testAccFmsAdminRegion != "" {
		return testAccFmsAdminRegion
	}

	testAccFmsAdminRegion = endpoints.UsEast1RegionID

	return testAccFmsAdminRegion
}
