package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAwsImageBuilderInfrastructureConfigurationDataSource_Arn(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_imagebuilder_infrastructure_configuration.test"
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsImageBuilderInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderInfrastructureConfigurationDataSourceConfigArn(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "date_created", resourceName, "date_created"),
					resource.TestCheckResourceAttrPair(dataSourceName, "date_updated", resourceName, "date_updated"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "instance_profile_name", resourceName, "instance_profile_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "instance_types.#", resourceName, "instance_types.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "key_pair", resourceName, "key_pair"),
					resource.TestCheckResourceAttrPair(dataSourceName, "logging.#", resourceName, "logging.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "resource_tags.%", resourceName, "resource_tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "security_group_ids.#", resourceName, "security_group_ids.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "sns_topic_arn", resourceName, "sns_topic_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnet_id", resourceName, "subnet_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "terminate_instance_on_failure", resourceName, "terminate_instance_on_failure"),
				),
			},
		},
	})
}

func testAccAwsImageBuilderInfrastructureConfigurationDataSourceConfigArn(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_instance_profile" "test" {
  name = aws_iam_role.role.name
  role = aws_iam_role.role.name
}

resource "aws_iam_role" "role" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}"
      }
      Sid = ""
    }]
  })
  name = %[1]q
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
}

data "aws_imagebuilder_infrastructure_configuration" "test" {
  arn = aws_imagebuilder_infrastructure_configuration.test.arn
}
`, rName)
}
