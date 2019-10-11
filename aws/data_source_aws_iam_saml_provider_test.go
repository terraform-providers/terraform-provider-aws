package aws

import (
"fmt"
"testing"

"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSDataSourceSAMLProvider_basic(t *testing.T) {
	providerName := acctest.RandomWithPrefix("tf-saml-provider-test")
	dataSourceName := "data.aws_iam_saml_provider.test"
	resourceName := "aws_iam_saml_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsIAMRoleConfig(roleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "assume_role_policy", resourceName, "assume_role_policy"),
					resource.TestCheckResourceAttrPair(dataSourceName, "create_date", resourceName, "create_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "max_session_duration", resourceName, "max_session_duration"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "path", resourceName, "path"),
					resource.TestCheckResourceAttrPair(dataSourceName, "unique_id", resourceName, "unique_id"),
				),
			},
		},
	})
}

func testAccAwsIAMRoleConfig(roleName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF

  path = "/testpath/"
}

data "aws_iam_role" "test" {
  name = "${aws_iam_role.test.name}"
}
`, roleName)
}

