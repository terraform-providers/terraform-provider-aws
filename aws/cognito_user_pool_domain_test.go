package aws

import (
	"context"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// Cognito User Pool Custom Domains can only be created with ACM Certificates in specific regions.

// testAccCognitoUserPoolCustomDomainRegion is the chosen Cognito User Pool Custom Domains testing region
//
// Cached to prevent issues should multiple regions become available.
var testAccCognitoUserPoolCustomDomainRegion string

// testAccProviderCognitoUserPoolCustomDomain is the Cognito User Pool Custom Domains provider instance
//
// This Provider can be used in testing code for API calls without requiring
// the use of saving and referencing specific ProviderFactories instances.
//
// testAccPreCheckCognitoUserPoolCustomDomain(t) must be called before using this provider instance.
var testAccProviderCognitoUserPoolCustomDomain *schema.Provider

// testAccProviderCognitoUserPoolCustomDomainConfigure ensures the provider is only configured once
var testAccProviderCognitoUserPoolCustomDomainConfigure sync.Once

// testAccPreCheckCognitoUserPoolCustomDomain verifies AWS credentials and that Cognito User Pool Custom Domains is supported
func testAccPreCheckCognitoUserPoolCustomDomain(t *testing.T) {
	testAccPartitionHasServicePreCheck(cognitoidentityprovider.EndpointsID, t)

	region := testAccGetCognitoUserPoolCustomDomainRegion()

	if region == "" {
		t.Skip("Cognito User Pool Custom Domains not available in this AWS Partition")
	}

	// Since we are outside the scope of the Terraform configuration we must
	// call Configure() to properly initialize the provider configuration.
	testAccProviderCognitoUserPoolCustomDomainConfigure.Do(func() {
		testAccProviderCognitoUserPoolCustomDomain = Provider()

		config := map[string]interface{}{
			"region": region,
		}

		diags := testAccProviderCognitoUserPoolCustomDomain.Configure(context.Background(), terraform.NewResourceConfigRaw(config))

		if diags != nil && diags.HasError() {
			for _, d := range diags {
				if d.Severity == diag.Error {
					t.Fatalf("error configuring Cognito User Pool Custom Domains provider: %s", d.Summary)
				}
			}
		}
	})
}

// testAccCognitoUserPoolCustomDomainRegionProviderConfig is the Terraform provider configuration for Cognito User Pool Custom Domains region testing
//
// Testing Cognito User Pool Custom Domains assumes no other provider configurations
// are necessary and overwrites the "aws" provider configuration.
func testAccCognitoUserPoolCustomDomainRegionProviderConfig() string {
	return testAccRegionalProviderConfig(testAccGetCognitoUserPoolCustomDomainRegion())
}

// testAccGetCognitoUserPoolCustomDomainRegion returns the Cognito User Pool Custom Domains region for testing
func testAccGetCognitoUserPoolCustomDomainRegion() string {
	if testAccCognitoUserPoolCustomDomainRegion != "" {
		return testAccCognitoUserPoolCustomDomainRegion
	}

	// AWS Commercial: https://docs.aws.amazon.com/cognito/latest/developerguide/cognito-user-pools-add-custom-domain.html
	// AWS GovCloud (US) - not supported: https://docs.aws.amazon.com/govcloud-us/latest/UserGuide/govcloud-cog.html
	// AWS China - not supported: https://docs.amazonaws.cn/en_us/aws/latest/userguide/cognito.html
	switch testAccGetPartition() {
	case endpoints.AwsPartitionID:
		testAccCognitoUserPoolCustomDomainRegion = endpoints.UsEast1RegionID
	}

	return testAccCognitoUserPoolCustomDomainRegion
}
