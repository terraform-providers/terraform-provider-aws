package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsSfnStateMachine_basic(t *testing.T) {
	rName := acctest.RandString(5)
	dataSourceName := "data.aws_sfn_state_machine.test"
	resourceName := "aws_sfn_state_machine.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsSfnStateMachineConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "creation_date", dataSourceName, "creation_date"),
					resource.TestCheckResourceAttrPair(resourceName, "definition", dataSourceName, "definition"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", dataSourceName, "role_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "status", dataSourceName, "status"),
				),
			},
		},
	})
}

func testAccDataSourceAwsSfnStateMachineConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_iam_role" "iam_for_sfn" {
  name = "iam_for_sfn_%s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "states.${data.aws_region.current.name}.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_sfn_state_machine" "test" {
  name     = "test_sfn_%s"
  role_arn = aws_iam_role.iam_for_sfn.arn

  definition = <<EOF
{
  "StartAt": "HelloWorld",
  "States": {
    "HelloWorld": {
      "Type": "Succeed"
    }
  }
}
EOF
}

data "aws_sfn_state_machine" "test" {
  name = aws_sfn_state_machine.test.name
}
`, rName, rName)
}
