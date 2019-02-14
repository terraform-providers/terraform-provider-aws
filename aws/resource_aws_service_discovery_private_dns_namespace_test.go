package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSServiceDiscoveryPrivateDnsNamespace_basic(t *testing.T) {
	rName := acctest.RandString(5) + ".example.com"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceDiscoveryPrivateDnsNamespaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDiscoveryPrivateDnsNamespaceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryPrivateDnsNamespaceExists("aws_service_discovery_private_dns_namespace.test"),
					resource.TestCheckResourceAttrSet("aws_service_discovery_private_dns_namespace.test", "arn"),
					resource.TestCheckResourceAttrSet("aws_service_discovery_private_dns_namespace.test", "hosted_zone"),
				),
			},
		},
	})
}

func TestAccAWSServiceDiscoveryPrivateDnsNamespace_longname(t *testing.T) {
	rName := acctest.RandString(64-len("example.com")) + ".example.com"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceDiscoveryPrivateDnsNamespaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDiscoveryPrivateDnsNamespaceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryPrivateDnsNamespaceExists("aws_service_discovery_private_dns_namespace.test"),
					resource.TestCheckResourceAttrSet("aws_service_discovery_private_dns_namespace.test", "arn"),
					resource.TestCheckResourceAttrSet("aws_service_discovery_private_dns_namespace.test", "hosted_zone"),
				),
			},
		},
	})
}

// This acceptance test ensures we properly send back error messaging. References:
//  * https://github.com/terraform-providers/terraform-provider-aws/issues/2830
//  * https://github.com/terraform-providers/terraform-provider-aws/issues/5532
func TestAccAWSServiceDiscoveryPrivateDnsNamespace_error_Overlap(t *testing.T) {
	rName := acctest.RandString(5) + ".example.com"
	subDomain := acctest.RandString(5) + "." + rName

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceDiscoveryPrivateDnsNamespaceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccServiceDiscoveryPrivateDnsNamespaceConfigOverlapping(rName, subDomain),
				ExpectError: regexp.MustCompile(`overlapping name space`),
			},
		},
	})
}

func testAccCheckAwsServiceDiscoveryPrivateDnsNamespaceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sdconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_service_discovery_private_dns_namespace" {
			continue
		}

		input := &servicediscovery.GetNamespaceInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetNamespace(input)
		if err != nil {
			if isAWSErr(err, servicediscovery.ErrCodeNamespaceNotFound, "") {
				return nil
			}
			return err
		}
	}
	return nil
}

func testAccCheckAwsServiceDiscoveryPrivateDnsNamespaceExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccServiceDiscoveryPrivateDnsNamespaceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-service-discovery-private-dns-ns"
  }
}

resource "aws_service_discovery_private_dns_namespace" "test" {
  name = "%s"
  description = "test"
  vpc = "${aws_vpc.test.id}"
}
`, rName)
}

func testAccServiceDiscoveryPrivateDnsNamespaceConfigOverlapping(topDomain, subDomain string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-service-discovery-private-dns-ns"
  }
}

resource "aws_service_discovery_private_dns_namespace" "top" {
  name = %q
  vpc  = "${aws_vpc.test.id}"
}

# Ensure ordering after first namespace
resource "aws_service_discovery_private_dns_namespace" "subdomain" {
  name = %q
  vpc  = "${aws_service_discovery_private_dns_namespace.top.vpc}"
}
`, topDomain, subDomain)
}
