package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSIotRoleAlias_basic(t *testing.T) {
	alias := acctest.RandomWithPrefix("RoleAlias-")
	alias2 := acctest.RandomWithPrefix("RoleAlias2-")

	resourceName := "aws_iot_role_alias.ra"
	resourceName2 := "aws_iot_role_alias.ra2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIotRoleAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIotRoleAliasConfig(alias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIotRoleAliasExists(resourceName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "iot", fmt.Sprintf("rolealias/%s", alias)),
					resource.TestCheckResourceAttr(resourceName, "credential_duration", "3600"),
				),
			},
			{
				Config: testAccAWSIotRoleAliasConfigUpdate1(alias, alias2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIotRoleAliasExists(resourceName),
					testAccCheckAWSIotRoleAliasExists(resourceName2),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "iot", fmt.Sprintf("rolealias/%s", alias)),
					resource.TestCheckResourceAttr(resourceName, "credential_duration", "1800"),
				),
			},
			{
				Config: testAccAWSIotRoleAliasConfigUpdate2(alias2),
				Check:  resource.ComposeTestCheckFunc(testAccCheckAWSIotRoleAliasExists(resourceName2)),
			},
			{
				Config: testAccAWSIotRoleAliasConfigUpdate3(alias2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIotRoleAliasExists(resourceName2),
				),
				ExpectError: regexp.MustCompile("Role alias .+? already exists for this account"),
			},
			{
				Config: testAccAWSIotRoleAliasConfigUpdate4(alias2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIotRoleAliasExists(resourceName2),
				),
			},
			{
				Config: testAccAWSIotRoleAliasConfigUpdate5(alias2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIotRoleAliasExists(resourceName2),
					testAccMatchResourceAttrGlobalARN(resourceName2, "role_arn", "iam", regexp.MustCompile("role/rolebogus")),
				),
			},
			{
				ResourceName:      resourceName2,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})

}

func testAccCheckAWSIotRoleAliasDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iotconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_role_alias" {
			continue
		}

		_, err := getIotRoleAliasDescription(conn, rs.Primary.ID)

		if isAWSErr(err, iot.ErrCodeResourceNotFoundException, "") {
			continue
		}

		return fmt.Errorf("IoT Role Alias (%s) still exists", rs.Primary.ID)
	}
	return nil
}

func testAccCheckAWSIotRoleAliasExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).iotconn
		role_arn := rs.Primary.Attributes["role_arn"]

		roleAliasDescription, err := getIotRoleAliasDescription(conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("Error: Failed to get role alias %s for role %s (%s): %s", rs.Primary.ID, role_arn, n, err)
		}

		if roleAliasDescription == nil {
			return fmt.Errorf("Error: Role alias %s is not attached to role (%s)", rs.Primary.ID, role_arn)
		}

		return nil
	}
}

func testAccAWSIotRoleAliasConfig(alias string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name = "role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Principal": {
      "Service": "credentials.iot.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }
}
EOF

}

resource "aws_iot_role_alias" "ra" {
  alias    = "%s"
  role_arn = aws_iam_role.role.arn
}
`, alias)
}

func testAccAWSIotRoleAliasConfigUpdate1(alias string, alias2 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name = "role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Principal": {
      "Service": "credentials.iot.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }
}
EOF

}

resource "aws_iot_role_alias" "ra" {
  alias               = "%s"
  role_arn            = aws_iam_role.role.arn
  credential_duration = 1800
}

resource "aws_iot_role_alias" "ra2" {
  alias    = "%s"
  role_arn = aws_iam_role.role.arn
}
`, alias, alias2)
}

func testAccAWSIotRoleAliasConfigUpdate2(alias2 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name = "role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Principal": {
      "Service": "credentials.iot.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }
}
EOF

}

resource "aws_iot_role_alias" "ra2" {
  alias    = "%s"
  role_arn = aws_iam_role.role.arn
}
`, alias2)
}

func testAccAWSIotRoleAliasConfigUpdate3(alias2 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name = "role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Principal": {
      "Service": "credentials.iot.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }
}
EOF

}

resource "aws_iot_role_alias" "ra2" {
  alias    = "%s"
  role_arn = aws_iam_role.role.arn
}

resource "aws_iot_role_alias" "ra3" {
  alias    = "%s"
  role_arn = aws_iam_role.role.arn
}
`, alias2, alias2)
}

func testAccAWSIotRoleAliasConfigUpdate4(alias2 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name = "role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Principal": {
      "Service": "credentials.iot.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }
}
EOF

}

resource "aws_iam_role" "role2" {
  name = "role2"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Principal": {
      "Service": "credentials.iot.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }
}
EOF

}

resource "aws_iot_role_alias" "ra2" {
  alias    = "%s"
  role_arn = aws_iam_role.role2.arn
}
`, alias2)
}

func testAccAWSIotRoleAliasConfigUpdate5(alias2 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name = "role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Principal": {
      "Service": "credentials.iot.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }
}
EOF

}

resource "aws_iam_role" "role2" {
  name = "role2"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Principal": {
      "Service": "credentials.iot.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }
}
EOF

}

resource "aws_iot_role_alias" "ra2" {
  alias    = "%s"
  role_arn = "${aws_iam_role.role.arn}bogus"
}
`, alias2)
}
