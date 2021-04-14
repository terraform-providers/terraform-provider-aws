package aws

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// API Gateway Edge-Optimized Domain Name can only be created with ACM Certificates in specific regions.

// testAccApigatewayEdgeDomainNameRegion is the chosen API Gateway Domain Name testing region
//
// Cached to prevent issues should multiple regions become available.
var testAccApigatewayEdgeDomainNameRegion string

// testAccProviderApigatewayEdgeDomainName is the API Gateway Domain Name provider instance
//
// This Provider can be used in testing code for API calls without requiring
// the use of saving and referencing specific ProviderFactories instances.
//
// testAccPreCheckApigatewayEdgeDomainName(t) must be called before using this provider instance.
var testAccProviderApigatewayEdgeDomainName *schema.Provider

// testAccProviderApigatewayEdgeDomainNameConfigure ensures the provider is only configured once
var testAccProviderApigatewayEdgeDomainNameConfigure sync.Once

// testAccPreCheckApigatewayEdgeDomainName verifies AWS credentials and that API Gateway Domain Name is supported
func testAccPreCheckApigatewayEdgeDomainName(t *testing.T) {
	testAccPartitionHasServicePreCheck(apigateway.EndpointsID, t)

	region := testAccGetApigatewayEdgeDomainNameRegion()

	if region == "" {
		t.Skip("API Gateway Domain Name not available in this AWS Partition")
	}

	// Since we are outside the scope of the Terraform configuration we must
	// call Configure() to properly initialize the provider configuration.
	testAccProviderApigatewayEdgeDomainNameConfigure.Do(func() {
		testAccProviderApigatewayEdgeDomainName = Provider()

		config := map[string]interface{}{
			"region": region,
		}

		diags := testAccProviderApigatewayEdgeDomainName.Configure(context.Background(), terraform.NewResourceConfigRaw(config))

		if diags != nil && diags.HasError() {
			for _, d := range diags {
				if d.Severity == diag.Error {
					t.Fatalf("error configuring API Gateway Domain Name provider: %s", d.Summary)
				}
			}
		}
	})
}

// testAccApigatewayEdgeDomainNameRegionProviderConfig is the Terraform provider configuration for API Gateway Domain Name region testing
//
// Testing API Gateway Domain Name assumes no other provider configurations
// are necessary and overwrites the "aws" provider configuration.
func testAccApigatewayEdgeDomainNameRegionProviderConfig() string {
	return testAccRegionalProviderConfig(testAccGetApigatewayEdgeDomainNameRegion())
}

// testAccGetApigatewayEdgeDomainNameRegion returns the API Gateway Domain Name region for testing
func testAccGetApigatewayEdgeDomainNameRegion() string {
	if testAccApigatewayEdgeDomainNameRegion != "" {
		return testAccApigatewayEdgeDomainNameRegion
	}

	// AWS Commercial: https://docs.aws.amazon.com/apigateway/latest/developerguide/how-to-custom-domains.html
	// AWS GovCloud (US) - edge custom domain names not supported: https://docs.aws.amazon.com/govcloud-us/latest/UserGuide/govcloud-abp.html
	// AWS China - edge custom domain names not supported: https://docs.amazonaws.cn/en_us/aws/latest/userguide/api-gateway.html
	switch testAccGetPartition() {
	case endpoints.AwsPartitionID:
		testAccApigatewayEdgeDomainNameRegion = endpoints.UsEast1RegionID
	}

	return testAccApigatewayEdgeDomainNameRegion
}

// testAccCheckResourceAttrRegionalARNApigatewayEdgeDomainName ensures the Terraform state exactly matches the expected API Gateway Edge Domain Name format
func testAccCheckResourceAttrRegionalARNApigatewayEdgeDomainName(resourceName, attributeName, arnService string, domain string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := arn.ARN{
			Partition: testAccGetPartition(),
			Region:    testAccGetApigatewayEdgeDomainNameRegion(),
			Resource:  fmt.Sprintf("/domainnames/%s", domain),
			Service:   arnService,
		}.String()

		return resource.TestCheckResourceAttr(resourceName, attributeName, attributeValue)(s)
	}
}

// testAccCheckResourceAttrRegionalARNApigatewayRegionalDomainName ensures the Terraform state exactly matches the expected API Gateway Regional Domain Name format
func testAccCheckResourceAttrRegionalARNApigatewayRegionalDomainName(resourceName, attributeName, arnService string, domain string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := arn.ARN{
			Partition: testAccGetPartition(),
			Region:    testAccGetRegion(),
			Resource:  fmt.Sprintf("/domainnames/%s", domain),
			Service:   arnService,
		}.String()

		return resource.TestCheckResourceAttr(resourceName, attributeName, attributeValue)(s)
	}
}
