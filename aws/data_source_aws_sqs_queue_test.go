package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceAwsSqsQueue_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf_acc_test_")
	resourceName := "aws_sqs_queue.test"
	datasourceName := "data.aws_sqs_queue.by_name"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsSqsQueueConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsSqsQueueCheck(datasourceName, resourceName),
					resource.TestCheckResourceAttr(datasourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsSqsQueue_tags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf_acc_test_")
	resourceName := "aws_sqs_queue.test"
	datasourceName := "data.aws_sqs_queue.by_name"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsSqsQueueConfigTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsSqsQueueCheck(datasourceName, resourceName),
					resource.TestCheckResourceAttr(datasourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(datasourceName, "tags.Environment", "Production"),
					resource.TestCheckResourceAttr(datasourceName, "tags.Foo", "Bar"),
					resource.TestCheckResourceAttr(datasourceName, "tags.Empty", ""),
				),
			},
		},
	})
}

func testAccDataSourceAwsSqsQueueCheck(datasourceName, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[datasourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", datasourceName)
		}

		sqsQueueRs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", resourceName)
		}

		attrNames := []string{
			"arn",
			"name",
		}

		for _, attrName := range attrNames {
			if rs.Primary.Attributes[attrName] != sqsQueueRs.Primary.Attributes[attrName] {
				return fmt.Errorf(
					"%s is %s; want %s",
					attrName,
					rs.Primary.Attributes[attrName],
					sqsQueueRs.Primary.Attributes[attrName],
				)
			}
		}

		return nil
	}
}

func testAccDataSourceAwsSqsQueueConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "wrong" {
  name = "%[1]s_wrong"
}

resource "aws_sqs_queue" "test" {
  name = "%[1]s"
}

data "aws_sqs_queue" "by_name" {
  name = aws_sqs_queue.test.name
}
`, rName)
}

func testAccDataSourceAwsSqsQueueConfigTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name = "%[1]s"

  tags = {
    Environment = "Production"
    Foo         = "Bar"
    Empty       = ""
  }
}

data "aws_sqs_queue" "by_name" {
  name = aws_sqs_queue.test.name
}
`, rName)
}
