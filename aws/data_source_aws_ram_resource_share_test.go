package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsRamResourceShare_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ram_resource_share.test"
	datasourceName := "data.aws_ram_resource_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsRamResourceShareConfig_NonExistent,
				ExpectError: regexp.MustCompile(`No matching resource found`),
			},
			{
				Config: testAccDataSourceAwsRamResourceShareConfig_Name(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrSet(datasourceName, "owning_account_id"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsRamResourceShare_Tags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ram_resource_share.test"
	datasourceName := "data.aws_ram_resource_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRamResourceShareConfig_Tags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccDataSourceAwsRamResourceShareConfig_Name(rName string) string {
	return fmt.Sprintf(`
resource "aws_ram_resource_share" "wrong" {
  name = "%s-wrong"
}

resource "aws_ram_resource_share" "test" {
  name = "%s"
}

data "aws_ram_resource_share" "test" {
  name           = aws_ram_resource_share.test.name
  resource_owner = "SELF"
}
`, rName, rName)
}

func testAccDataSourceAwsRamResourceShareConfig_Tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_ram_resource_share" "test" {
  name = "%s"

  tags = {
    Name = "%s-Tags"
  }
}

data "aws_ram_resource_share" "test" {
  name           = aws_ram_resource_share.test.name
  resource_owner = "SELF"

  filter {
    name   = "Name"
    values = ["%s-Tags"]
  }
}
`, rName, rName, rName)
}

const testAccDataSourceAwsRamResourceShareConfig_NonExistent = `
data "aws_ram_resource_share" "test" {
  name           = "tf-acc-test-does-not-exist"
  resource_owner = "SELF"
}
`
