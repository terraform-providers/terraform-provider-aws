package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccAwsOrganizationsOrganizationalUnit_basic(t *testing.T) {
	var unit organizations.OrganizationalUnit

	rInt := acctest.RandInt()
	name := fmt.Sprintf("tf_outest_%d", rInt)
	resourceName := "aws_organizations_organizational_unit.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOrganizationsOrganizationalUnitDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsOrganizationalUnitConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsOrganizationalUnitExists(resourceName, &unit),
					resource.TestCheckResourceAttr(resourceName, "accounts.#", "0"),
					testAccMatchResourceAttrGlobalARN(resourceName, "arn", "organizations", regexp.MustCompile(`ou/o-.+/ou-.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", name),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsOrganizationsOrganizationalUnit_Name(t *testing.T) {
	var unit organizations.OrganizationalUnit

	rInt := acctest.RandInt()
	name1 := fmt.Sprintf("tf_outest_%d", rInt)
	name2 := fmt.Sprintf("tf_outest_%d", rInt+1)
	resourceName := "aws_organizations_organizational_unit.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOrganizationsOrganizationalUnitDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsOrganizationalUnitConfig(name1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsOrganizationalUnitExists(resourceName, &unit),
					resource.TestCheckResourceAttr(resourceName, "name", name1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsOrganizationsOrganizationalUnitConfig(name2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsOrganizationalUnitExists(resourceName, &unit),
					resource.TestCheckResourceAttr(resourceName, "name", name2),
				),
			},
		},
	})
}

func testAccCheckAwsOrganizationsOrganizationalUnitDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).organizationsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_organizations_organizational_unit" {
			continue
		}

		params := &organizations.DescribeOrganizationalUnitInput{
			OrganizationalUnitId: &rs.Primary.ID,
		}

		resp, err := conn.DescribeOrganizationalUnit(params)

		if err != nil {
			if isAWSErr(err, organizations.ErrCodeAWSOrganizationsNotInUseException, "") {
				continue
			}
			if isAWSErr(err, organizations.ErrCodeOrganizationalUnitNotFoundException, "") {
				continue
			}
			return err
		}

		if resp != nil && resp.OrganizationalUnit != nil {
			return fmt.Errorf("Bad: Organizational Unit still exists: %q", rs.Primary.ID)
		}
	}

	return nil

}

func testAccCheckAwsOrganizationsOrganizationalUnitExists(n string, ou *organizations.OrganizationalUnit) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).organizationsconn
		params := &organizations.DescribeOrganizationalUnitInput{
			OrganizationalUnitId: &rs.Primary.ID,
		}

		resp, err := conn.DescribeOrganizationalUnit(params)

		if err != nil {
			if isAWSErr(err, organizations.ErrCodeOrganizationalUnitNotFoundException, "") {
				return fmt.Errorf("Organizational Unit %q does not exist", rs.Primary.ID)
			}
			return err
		}

		if resp == nil {
			return fmt.Errorf("failed to DescribeOrganizationalUnit %q, response was nil", rs.Primary.ID)
		}

		ou = resp.OrganizationalUnit

		return nil
	}
}

func testAccAwsOrganizationsOrganizationalUnitConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_organizational_unit" "test" {
  name      = %[1]q
  parent_id = aws_organizations_organization.test.roots[0].id
}
`, name)
}
