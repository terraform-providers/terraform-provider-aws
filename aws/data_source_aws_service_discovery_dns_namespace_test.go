package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSServiceDiscoveryDnsNamespaceDataSource_private(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_service_discovery_dns_namespace.test"
	resourceName := "aws_service_discovery_private_dns_namespace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t); testAccPreCheckAWSServiceDiscovery(t) },
		ErrorCheck: testAccErrorCheck(t, servicediscovery.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceDiscoveryPrivateDnsNamespaceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "hosted_zone", resourceName, "hosted_zone"),
				),
			},
		},
	})
}

func TestAccAWSServiceDiscoveryDnsNamespaceDataSource_public(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_service_discovery_dns_namespace.test"
	resourceName := "aws_service_discovery_public_dns_namespace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t); testAccPreCheckAWSServiceDiscovery(t) },
		ErrorCheck: testAccErrorCheck(t, servicediscovery.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceDiscoveryPublicDnsNamespaceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "hosted_zone", resourceName, "hosted_zone"),
				),
			},
		},
	})
}

func testAccCheckAwsServiceDiscoveryPrivateDnsNamespaceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_service_discovery_private_dns_namespace" "test" {
  name = "%[1]s.tf"
  vpc  = aws_vpc.test.id
}

data "aws_service_discovery_dns_namespace" "test" {
  name = aws_service_discovery_private_dns_namespace.test.name
  type = "DNS_PRIVATE"
}
`, rName)
}

func testAccCheckAwsServiceDiscoveryPublicDnsNamespaceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_public_dns_namespace" "test" {
  name = "%[1]s.tf"
}

data "aws_service_discovery_dns_namespace" "test" {
  name = aws_service_discovery_public_dns_namespace.test.name
  type = "DNS_PUBLIC"
}
`, rName)
}
