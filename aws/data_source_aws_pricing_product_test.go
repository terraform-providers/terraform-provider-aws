package aws

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceAwsPricingProduct_ec2(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckPricing(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsPricingProductConfigEc2(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_pricing_product.test", "result"),
					testAccPricingCheckValueIsJSON("data.aws_pricing_product.test"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsPricingProduct_redshift(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckPricing(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsPricingProductConfigRedshift(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_pricing_product.test", "result"),
					testAccPricingCheckValueIsJSON("data.aws_pricing_product.test"),
				),
			},
		},
	})
}

func testAccDataSourceAwsPricingProductConfigEc2() string {
	return composeConfig(
		testAccPricingRegionProviderConfig(),
		`
data "aws_ec2_instance_type_offering" "available" {
  preferred_instance_types = ["c5.large", "c4.large"]
}

data "aws_region" "current" {}

data "aws_pricing_product" "test" {
  service_code = "AmazonEC2"

  filters {
    field = "instanceType"
    value = data.aws_ec2_instance_type_offering.available.instance_type
  }

  filters {
    field = "operatingSystem"
    value = "Linux"
  }

  filters {
    field = "location"
    value = data.aws_region.current.description
  }

  filters {
    field = "preInstalledSw"
    value = "NA"
  }

  filters {
    field = "licenseModel"
    value = "No License required"
  }

  filters {
    field = "tenancy"
    value = "Shared"
  }

  filters {
    field = "capacitystatus"
    value = "Used"
  }
}
`)
}

func testAccDataSourceAwsPricingProductConfigRedshift() string {
	return composeConfig(
		testAccPricingRegionProviderConfig(),
		`
data "aws_redshift_orderable_cluster" "test" {
  preferred_node_types = ["dc2.8xlarge", "ds2.8xlarge"]
}

data "aws_region" "current" {}

data "aws_pricing_product" "test" {
  service_code = "AmazonRedshift"

  filters {
    field = "instanceType"
    value = data.aws_redshift_orderable_cluster.test.node_type
  }

  filters {
    field = "location"
    value = data.aws_region.current.description
  }
}
`)
}

func testAccPricingCheckValueIsJSON(data string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[data]

		if !ok {
			return fmt.Errorf("Can't find resource: %s", data)
		}

		result := rs.Primary.Attributes["result"]
		var objmap map[string]*json.RawMessage

		if err := json.Unmarshal([]byte(result), &objmap); err != nil {
			return fmt.Errorf("%s result value (%s) is not JSON: %s", data, result, err)
		}

		if len(objmap) == 0 {
			return fmt.Errorf("%s result value (%s) unmarshalling resulted in an empty map", data, result)
		}

		return nil
	}
}
