package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSService_basic(t *testing.T) {
	dataSourceName := "data.aws_service.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "dns_name", fmt.Sprintf("%s.%s.%s", ec2.EndpointsID, testAccGetRegion(), "amazonaws.com")),
					resource.TestCheckResourceAttr(dataSourceName, "partition", testAccGetPartition()),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_prefix", "com.amazonaws"),
					resource.TestCheckResourceAttr(dataSourceName, "region", testAccGetRegion()),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_name", fmt.Sprintf("%s.%s.%s", "com.amazonaws", testAccGetRegion(), ec2.EndpointsID)),
					resource.TestCheckResourceAttr(dataSourceName, "service_id", ec2.EndpointsID),
					resource.TestCheckResourceAttr(dataSourceName, "supported", "true"),
				),
			},
		},
	})
}

func TestAccAWSService_byReverseDNSName(t *testing.T) {
	dataSourceName := "data.aws_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceConfig_byReverseDNSName(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "region", endpoints.CnNorth1RegionID),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_name", fmt.Sprintf("%s.%s.%s", "cn.com.amazonaws", endpoints.CnNorth1RegionID, s3.EndpointsID)),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_prefix", "cn.com.amazonaws"),
					resource.TestCheckResourceAttr(dataSourceName, "service_id", s3.EndpointsID),
					resource.TestCheckResourceAttr(dataSourceName, "supported", "true"),
				),
			},
		},
	})
}

func TestAccAWSService_byDNSName(t *testing.T) {
	dataSourceName := "data.aws_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceConfig_byDNSName(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "region", endpoints.UsEast1RegionID),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_name", fmt.Sprintf("%s.%s.%s", "com.amazonaws", endpoints.UsEast1RegionID, rds.EndpointsID)),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_prefix", "com.amazonaws"),
					resource.TestCheckResourceAttr(dataSourceName, "service_id", rds.EndpointsID),
					resource.TestCheckResourceAttr(dataSourceName, "supported", "true"),
				),
			},
		},
	})
}

func TestAccAWSService_byParts(t *testing.T) {
	dataSourceName := "data.aws_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceConfig_byPart(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "dns_name", fmt.Sprintf("%s.%s.%s", s3.EndpointsID, testAccGetRegion(), "amazonaws.com")),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_name", fmt.Sprintf("%s.%s.%s", "com.amazonaws", testAccGetRegion(), s3.EndpointsID)),
					resource.TestCheckResourceAttr(dataSourceName, "supported", "true"),
				),
			},
		},
	})
}

func TestAccAWSService_unsupported(t *testing.T) {
	dataSourceName := "data.aws_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceConfig_unsupported(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "dns_name", fmt.Sprintf("%s.%s.%s", waf.EndpointsID, endpoints.UsGovWest1RegionID, "amazonaws.com")),
					resource.TestCheckResourceAttr(dataSourceName, "partition", endpoints.AwsUsGovPartitionID),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_prefix", "com.amazonaws"),
					resource.TestCheckResourceAttr(dataSourceName, "region", endpoints.UsGovWest1RegionID),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_name", fmt.Sprintf("%s.%s.%s", "com.amazonaws", endpoints.UsGovWest1RegionID, waf.EndpointsID)),
					resource.TestCheckResourceAttr(dataSourceName, "service_id", waf.EndpointsID),
					resource.TestCheckResourceAttr(dataSourceName, "supported", "false"),
				),
			},
		},
	})
}

func testAccCheckAwsServiceConfig_basic() string {
	return `
data "aws_service" "default" {}
`
}

func testAccCheckAwsServiceConfig_byReverseDNSName() string {
	// lintignore:AWSAT003
	return `
data "aws_service" "test" {
  reverse_dns_name = "cn.com.amazonaws.cn-north-1.s3"
}
`
}

func testAccCheckAwsServiceConfig_byDNSName() string {
	// lintignore:AWSAT003
	return `
data "aws_service" "test" {
  dns_name = "rds.us-east-1.amazonaws.com"
}
`
}

func testAccCheckAwsServiceConfig_byPart() string {
	return `
data "aws_region" "current" {}

data "aws_service" "test" {
  reverse_dns_prefix = "com.amazonaws"
  region             = data.aws_region.current.name
  service_id         = "s3"
}
`
}

func testAccCheckAwsServiceConfig_unsupported() string {
	// lintignore:AWSAT003
	return `
data "aws_service" "test" {
  reverse_dns_name = "com.amazonaws.us-gov-west-1.waf"
}
`
}
