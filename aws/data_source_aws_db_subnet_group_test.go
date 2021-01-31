package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSDbSubnetGroupDataSource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_db_subnet_group.test"
	dataSourceName := "data.aws_db_subnet_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBSubnetGroupDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "description", dataSourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_ids", dataSourceName, "subnet_ids"),
					resource.TestCheckResourceAttrSet(dataSourceName, "status"),
					resource.TestCheckResourceAttrSet(dataSourceName, "vpc_id"),
				),
			},
		},
	})
}

func TestAccAWSDbSubnetGroupDataSource_nonexistent(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSDBSubnetGroupDataSourceConfig_NonExistent,
				ExpectError: regexp.MustCompile(`error reading DB SubnetGroup \(tf-acc-test-does-not-exist\)`),
			},
		},
	})
}

const testAccAWSDBSubnetGroupDataSourceConfig_NonExistent = `
data "aws_db_subnet_group" "test" {
  name = "tf-acc-test-does-not-exist"
}
`

func testAccAWSDBSubnetGroupDataSourceConfig(rName string) string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  count             = 2
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id
}

resource "aws_db_subnet_group" "test" {
  name       = "%s"
  subnet_ids = aws_subnet.test.*.id
}

data "aws_db_subnet_group" "test" {
  name = aws_db_subnet_group.test.name
}
`, rName))
}
