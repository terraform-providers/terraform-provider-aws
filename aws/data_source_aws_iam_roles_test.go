package aws

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSDataSourceIAMRoles_basic(t *testing.T) {
	dataSourceName := "data.aws_iam_roles.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, iam.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceIAMRolesConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					// testCheckResourceAttrIsNot(dataSourceName, "names.#", "0"),
					resource.TestMatchResourceAttr(dataSourceName, "names.#", regexp.MustCompile("[^0].*$")),
				),
			},
		},
	})
}

func TestAccAWSDataSourceIAMRoles_filterRoleName(t *testing.T) {
	rCount := strconv.Itoa(acctest.RandIntRange(1, 4))
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_iam_roles.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, iam.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceIAMRolesConfig_filterByName(rCount, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "names.#", rCount),
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", rCount),
				),
			},
		},
	})
}

func TestAccAWSDataSourceIAMRoles_pathPrefix(t *testing.T) {
	rCount := strconv.Itoa(acctest.RandIntRange(1, 4))
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rPathPrefix := acctest.RandomWithPrefix("tf-acc-path")
	dataSourceName := "data.aws_iam_roles.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, iam.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceIAMRolesConfig_pathPrefix(rCount, rName, rPathPrefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "names.#", rCount),
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", rCount),
				),
			},
		},
	})
}

func TestAccAWSDataSourceIAMRoles_pathPrefixAndFilterRoleName(t *testing.T) {
	rCount := strconv.Itoa(acctest.RandIntRange(1, 4))
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rPathPrefix := acctest.RandomWithPrefix("tf-acc-path")
	dataSourceName := "data.aws_iam_roles.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, iam.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceIAMRolesConfig_pathPrefixAndFilterRoleName(rCount, rName, rPathPrefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "names.#", rCount),
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", rCount),
				),
			},
		},
	})
}

func testAccAWSDataSourceIAMRolesConfig_basic() string {
	return `
data "aws_iam_roles" "test" {}
`
}

func testAccAWSDataSourceIAMRolesConfig_filterByName(rCount string, rName string) string {
	return fmt.Sprintf(`

resource "aws_iam_role" "test" {
  count = %[1]s
  name  = "%[2]s-${count.index}-role"

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
  tags = {
    Seed = %[2]q
  }
}

data "aws_iam_roles" "test" {
  filter {
    name   = "role-name"
    values = ["${aws_iam_role.test[0].tags["Seed"]}-*-role"]
  }
}
`, rCount, rName)
}

func testAccAWSDataSourceIAMRolesConfig_pathPrefix(rCount string, rName string, rPathPrefix string) string {
	return fmt.Sprintf(`

resource "aws_iam_role" "test" {
  count = %[1]s
  name  = "%[2]s-${count.index}-role"

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

  path = "/%[3]s/"
}

data "aws_iam_roles" "test" {
  path_prefix = aws_iam_role.test[0].path
}
`, rCount, rName, rPathPrefix)
}

func testAccAWSDataSourceIAMRolesConfig_pathPrefixAndFilterRoleName(rCount string, rName string, rPathPrefix string) string {
	return fmt.Sprintf(`

resource "aws_iam_role" "test" {
  count = %[1]s
  name  = "%[2]s-${count.index}-role"

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

  path = "/%[3]s/"
  tags = {
    Seed = %[2]q
  }
}

data "aws_iam_roles" "test" {
  filter {
    name   = "role-name"
    values = ["${aws_iam_role.test[0].tags["Seed"]}-*-role", "*-role"]
  }
  path_prefix = aws_iam_role.test[0].path
}
`, rCount, rName, rPathPrefix)
}
