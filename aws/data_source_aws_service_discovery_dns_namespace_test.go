package aws

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSServiceDiscoveryDnsNamespaceDataSource_public(t *testing.T) {
	resourceName := "aws_service_discovery_public_dns_namespace.test"
	dataSourceName := "data.aws_service_discovery_dns_namespace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSServiceDiscoveryDnsNamespaceDataSourcePublicConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "type", "public"),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "hosted_zone", resourceName, "hosted_zone"),
				),
			},
		},
	})
}

func TestAccAWSServiceDiscoveryDnsNamespaceDataSource_mostRecentPublic(t *testing.T) {
	resourceName := "aws_service_discovery_public_dns_namespace.recent"
	dataSourceName := "data.aws_service_discovery_dns_namespace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSServiceDiscoveryDnsNamespaceDataSourcePublicMostRecentConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "type", "public"),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "hosted_zone", resourceName, "hosted_zone"),
				),
			},
		},
	})
}

func TestAccAWSServiceDiscoveryDnsNamespaceDataSource_private(t *testing.T) {
	resourceName := "aws_service_discovery_private_dns_namespace.test"
	dataSourceName := "data.aws_service_discovery_dns_namespace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSServiceDiscoveryDnsNamespaceDataSourcePrivateConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "type", "private"),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "hosted_zone", resourceName, "hosted_zone"),
				),
			},
		},
	})
}

const testAccCheckAWSServiceDiscoveryDnsNamespaceDataSourcePublicConfig = `
resource "aws_service_discovery_public_dns_namespace" "test" {
  name = "public.test.acc"
}

data "aws_service_discovery_dns_namespace" "test" {
    name = "${aws_service_discovery_public_dns_namespace.test.name}"
    type = "public"
}
`

const testAccCheckAWSServiceDiscoveryDnsNamespaceDataSourcePublicMostRecentConfig = `
resource "aws_service_discovery_public_dns_namespace" "older" {
  name = "most-recent.test.acc"
}

resource "aws_service_discovery_public_dns_namespace" "recent" {
  name       = "most-recent.test.acc"
  depends_on = ["aws_service_discovery_public_dns_namespace.older"]
}

data "aws_service_discovery_dns_namespace" "test" {
  name        = "${aws_service_discovery_public_dns_namespace.recent.name}"
  type        = "public"
  most_recent = true
}
`

const testAccCheckAWSServiceDiscoveryDnsNamespaceDataSourcePrivateConfig = `
resource "aws_vpc" "test" {
  cidr_block = "172.0.0.0/16"
}

resource "aws_service_discovery_private_dns_namespace" "test" {
  name = "private.test.acc"
  vpc  = "${aws_vpc.test.id}"
}

data "aws_service_discovery_dns_namespace" "test" {
  name = "${aws_service_discovery_private_dns_namespace.test.name}"
  type = "private"
}
`
