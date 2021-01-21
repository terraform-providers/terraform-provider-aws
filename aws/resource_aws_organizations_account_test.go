package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccAwsOrganizationsAccount_basic(t *testing.T) {
	TestAccSkip(t, "AWS Organizations Account testing is not currently automated due to manual account deletion steps.")

	var account organizations.Account

	orgsEmailDomain, ok := os.LookupEnv("TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN")

	if !ok {
		TestAccSkip(t, "'TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN' not set, skipping test.")
	}

	rInt := acctest.RandInt()
	name := fmt.Sprintf("tf_acctest_%d", rInt)
	email := fmt.Sprintf("tf-acctest+%d@%s", rInt, orgsEmailDomain)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOrganizationsAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsAccountConfig(name, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsAccountExists("aws_organizations_account.test", &account),
					resource.TestCheckResourceAttrSet("aws_organizations_account.test", "arn"),
					resource.TestCheckResourceAttrSet("aws_organizations_account.test", "joined_method"),
					testAccCheckResourceAttrRfc3339("aws_organizations_account.test", "joined_timestamp"),
					resource.TestCheckResourceAttrSet("aws_organizations_account.test", "parent_id"),
					resource.TestCheckResourceAttr("aws_organizations_account.test", "name", name),
					resource.TestCheckResourceAttr("aws_organizations_account.test", "email", email),
					resource.TestCheckResourceAttrSet("aws_organizations_account.test", "status"),
					resource.TestCheckResourceAttr("aws_organizations_account.test", "tags.%", "0"),
				),
			},
			{
				ResourceName:      "aws_organizations_account.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsOrganizationsAccount_ParentId(t *testing.T) {
	TestAccSkip(t, "AWS Organizations Account testing is not currently automated due to manual account deletion steps.")

	var account organizations.Account

	orgsEmailDomain, ok := os.LookupEnv("TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN")

	if !ok {
		TestAccSkip(t, "'TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN' not set, skipping test.")
	}

	rInt := acctest.RandInt()
	name := fmt.Sprintf("tf_acctest_%d", rInt)
	email := fmt.Sprintf("tf-acctest+%d@%s", rInt, orgsEmailDomain)
	resourceName := "aws_organizations_account.test"
	parentIdResourceName1 := "aws_organizations_organizational_unit.test1"
	parentIdResourceName2 := "aws_organizations_organizational_unit.test2"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOrganizationsAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsAccountConfigParentId1(name, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsAccountExists(resourceName, &account),
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", parentIdResourceName1, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsOrganizationsAccountConfigParentId2(name, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsAccountExists(resourceName, &account),
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", parentIdResourceName2, "id"),
				),
			},
		},
	})
}

func testAccAwsOrganizationsAccount_Tags(t *testing.T) {
	TestAccSkip(t, "AWS Organizations Account testing is not currently automated due to manual account deletion steps.")

	var account organizations.Account

	orgsEmailDomain, ok := os.LookupEnv("TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN")

	if !ok {
		TestAccSkip(t, "'TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN' not set, skipping test.")
	}

	rInt := acctest.RandInt()
	name := fmt.Sprintf("tf_acctest_%d", rInt)
	email := fmt.Sprintf("tf-acctest+%d@%s", rInt, orgsEmailDomain)
	resourceName := "aws_organizations_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOrganizationsAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsAccountConfigTags1(name, email, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsAccountExists(resourceName, &account),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsOrganizationsAccountConfigTags2(name, email, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsAccountExists(resourceName, &account),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAwsOrganizationsAccountConfig(name, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsAccountExists("aws_organizations_account.test", &account),
					resource.TestCheckResourceAttr("aws_organizations_account.test", "tags.%", "0"),
				),
			},
		},
	})
}

func testAccCheckAwsOrganizationsAccountDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).organizationsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_organizations_account" {
			continue
		}

		params := &organizations.DescribeAccountInput{
			AccountId: &rs.Primary.ID,
		}

		resp, err := conn.DescribeAccount(params)

		if isAWSErr(err, organizations.ErrCodeAccountNotFoundException, "") {
			return nil
		}

		if err != nil {
			return err
		}

		if resp != nil && resp.Account != nil {
			return fmt.Errorf("Bad: Account still exists: %q", rs.Primary.ID)
		}
	}

	return nil

}

func testAccCheckAwsOrganizationsAccountExists(n string, a *organizations.Account) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).organizationsconn
		params := &organizations.DescribeAccountInput{
			AccountId: &rs.Primary.ID,
		}

		resp, err := conn.DescribeAccount(params)

		if err != nil {
			return err
		}

		if resp == nil || resp.Account == nil {
			return fmt.Errorf("Account %q does not exist", rs.Primary.ID)
		}

		a = resp.Account

		return nil
	}
}

func testAccAwsOrganizationsAccountConfig(name, email string) string {
	return fmt.Sprintf(`
resource "aws_organizations_account" "test" {
  name  = "%s"
  email = "%s"
}
`, name, email)
}

func testAccAwsOrganizationsAccountConfigParentId1(name, email string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_organizational_unit" "test1" {
  name      = "test1"
  parent_id = aws_organizations_organization.test.roots[0].id
}

resource "aws_organizations_organizational_unit" "test2" {
  name      = "test2"
  parent_id = aws_organizations_organization.test.roots[0].id
}

resource "aws_organizations_account" "test" {
  name      = %[1]q
  email     = %[2]q
  parent_id = aws_organizations_organizational_unit.test1.id
}
`, name, email)
}

func testAccAwsOrganizationsAccountConfigParentId2(name, email string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_organizational_unit" "test1" {
  name      = "test1"
  parent_id = aws_organizations_organization.test.roots[0].id
}

resource "aws_organizations_organizational_unit" "test2" {
  name      = "test2"
  parent_id = aws_organizations_organization.test.roots[0].id
}

resource "aws_organizations_account" "test" {
  name      = %[1]q
  email     = %[2]q
  parent_id = aws_organizations_organizational_unit.test2.id
}
`, name, email)
}

func testAccAwsOrganizationsAccountConfigTags1(name, email, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_account" "test" {
  name  = %[1]q
  email = %[2]q

  tags = {
    %[3]q = %[4]q
  }
}
`, name, email, tagKey1, tagValue1)
}

func testAccAwsOrganizationsAccountConfigTags2(name, email, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_account" "test" {
  name  = %[1]q
  email = %[2]q

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, name, email, tagKey1, tagValue1, tagKey2, tagValue2)
}
