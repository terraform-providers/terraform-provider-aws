package aws

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceAwsPricingProduct_ec2(t *testing.T) {
	oldRegion := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldRegion)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsPricingProductConfigEc2("test", "c5.large"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_pricing_product.test", "result"),
					testAccPricingCheckValueIsJSON("data.aws_pricing_product.test"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsPricingProduct_redshift(t *testing.T) {
	oldRegion := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldRegion)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
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

func testAccDataSourceAwsPricingProductConfigEc2(dataName string, instanceType string) string {
	return fmt.Sprintf(`data "aws_pricing_product" "%s" {
		service_code = "AmazonEC2"
	  
		filters {
			field = "instanceType"
			value = "%s"
		}

		filters {
			field = "operatingSystem"
			value = "Linux"
		}

		filters {
			field = "location"
			value = "US East (N. Virginia)"
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
`, dataName, instanceType)
}

func testAccDataSourceAwsPricingProductConfigRedshift() string {
	return fmt.Sprintf(`data "aws_pricing_product" "test" {
		service_code = "AmazonRedshift"
	  
		filters {
			field = "instanceType"
			value = "ds1.xlarge"
		}

		filters {
			field = "location"
			value = "US East (N. Virginia)"
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
