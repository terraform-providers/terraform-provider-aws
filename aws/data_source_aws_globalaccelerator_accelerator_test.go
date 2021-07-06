package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSDataGlobalAcceleratorAccelerator_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_globalaccelerator_accelerator.test"
	dataSourceName := "data.aws_globalaccelerator_accelerator.test_by_arn"
	dataSourceName2 := "data.aws_globalaccelerator_accelerator.test_by_name"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t); testAccPreCheckGlobalAccelerator(t) },
		ErrorCheck: testAccErrorCheck(t, globalaccelerator.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlobalAcceleratorAcceleratorConfigWithDataSource(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "attributes.#", resourceName, "attributes.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "attributes.0.flow_logs_enabled", resourceName, "attributes.0.flow_logs_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "attributes.0.flow_logs_s3_bucket", resourceName, "attributes.0.flow_logs_s3_bucket"),
					resource.TestCheckResourceAttrPair(dataSourceName, "attributes.0.flow_logs_s3_prefix", resourceName, "attributes.0.flow_logs_s3_prefix"),
					resource.TestCheckResourceAttrPair(dataSourceName, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enabled", resourceName, "enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "hosted_zone_id", resourceName, "hosted_zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ip_address_type", resourceName, "ip_address_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ip_sets.#", resourceName, "ip_sets.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ip_sets.0.ip_addresses.#", resourceName, "ip_sets.0.ip_addresses.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ip_sets.0.ip_addresses.0", resourceName, "ip_sets.0.ip_addresses.0"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ip_sets.0.ip_addresses.1", resourceName, "ip_sets.0.ip_addresses.1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ip_sets.0.ip_family", resourceName, "ip_sets.0.ip_family"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "attributes.#", resourceName, "attributes.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "attributes.0.flow_logs_enabled", resourceName, "attributes.0.flow_logs_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "attributes.0.flow_logs_s3_bucket", resourceName, "attributes.0.flow_logs_s3_bucket"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "attributes.0.flow_logs_s3_prefix", resourceName, "attributes.0.flow_logs_s3_prefix"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "enabled", resourceName, "enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "hosted_zone_id", resourceName, "hosted_zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "ip_address_type", resourceName, "ip_address_type"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "ip_sets.#", resourceName, "ip_sets.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "ip_sets.0.ip_addresses.#", resourceName, "ip_sets.0.ip_addresses.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "ip_sets.0.ip_addresses.0", resourceName, "ip_sets.0.ip_addresses.0"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "ip_sets.0.ip_addresses.1", resourceName, "ip_sets.0.ip_addresses.1"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "ip_sets.0.ip_family", resourceName, "ip_sets.0.ip_family"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "name", resourceName, "name"),
				),
			},
		},
	})
}

func testAccAWSGlobalAcceleratorAcceleratorConfigWithDataSource(rName string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_accelerator" "test" {
  name = %[1]q
  attributes {
    flow_logs_enabled   = false
    flow_logs_s3_bucket = ""
    flow_logs_s3_prefix = "flow-logs/globalaccelerator/"
  }
}

data "aws_globalaccelerator_accelerator" "test_by_arn" {
  arn = aws_globalaccelerator_accelerator.test.id
}

data "aws_globalaccelerator_accelerator" "test_by_name" {
  name = aws_globalaccelerator_accelerator.test.name
}
`, rName)
}
