package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSLightsailInstanceDataSource_Name(t *testing.T) {
	var instance lightsail.Instance
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_instance.test"
	dataSourceName := "data.aws_lightsail_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lightsail.EndpointsID, t)
			testAccPreCheckAWSLightsail(t)
		},
		ErrorCheck:   testAccErrorCheck(t, lightsail.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLightsailInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailInstanceDataSourceConfigName(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailInstanceExists(resourceName, &instance),
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zone", dataSourceName, "availability_zone"),
					resource.TestCheckResourceAttrPair(resourceName, "blueprint_id", dataSourceName, "blueprint_id"),
					resource.TestCheckResourceAttrPair(resourceName, "bundle_id", dataSourceName, "bundle_id"),
					resource.TestCheckResourceAttrPair(resourceName, "tags", dataSourceName, "tags"),
					resource.TestCheckResourceAttrPair(resourceName, "ip_address_type", dataSourceName, "ip_address_type"),
				),
			},
		},
	})
}

func testAccAWSLightsailInstanceDataSourceConfigName(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_lightsail_instance" "lightsail_instance_test" {
  name              = "%s"
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux"
  bundle_id         = "nano_1_0"
  ip_address_type   = "dualstack"

  tags = {
    Name       = "tf-test"
    KeyOnlyTag = ""
    ExtraName  = "tf-test"
  }
}

data "aws_lightsail_instance" "lightsail_instance_test" {
  name = aws_lightsail_instance.lightsail_instance_test.id
}
`, rName)
}
