package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func testAccAwsOrganizationsOrganization_basic(t *testing.T) {
	var organization organizations.Organization
	resourceName := "aws_organizations_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOrganizationsOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsOrganizationConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsOrganizationExists(resourceName, &organization),
					testAccMatchResourceAttrGlobalARN(resourceName, "arn", "organizations", regexp.MustCompile(`organization/o-.+`)),
					resource.TestCheckResourceAttr(resourceName, "aws_service_access_principals.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "feature_set", organizations.OrganizationFeatureSetAll),
					testAccMatchResourceAttrGlobalARN(resourceName, "master_account_arn", "organizations", regexp.MustCompile(`account/o-.+/.+`)),
					resource.TestMatchResourceAttr(resourceName, "master_account_email", regexp.MustCompile(`.+@.+`)),
					testAccCheckResourceAttrAccountID(resourceName, "master_account_id"),
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

func testAccAwsOrganizationsOrganization_AwsServiceAccessPrincipals(t *testing.T) {
	var organization organizations.Organization
	resourceName := "aws_organizations_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOrganizationsOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsOrganizationConfigAwsServiceAccessPrincipals1("config.amazonaws.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsOrganizationExists(resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "aws_service_access_principals.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "aws_service_access_principals.553690328", "config.amazonaws.com"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsOrganizationsOrganizationConfigAwsServiceAccessPrincipals2("config.amazonaws.com", "ds.amazonaws.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsOrganizationExists(resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "aws_service_access_principals.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "aws_service_access_principals.553690328", "config.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName, "aws_service_access_principals.3567899500", "ds.amazonaws.com"),
				),
			},
			{
				Config: testAccAwsOrganizationsOrganizationConfigAwsServiceAccessPrincipals1("fms.amazonaws.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsOrganizationExists(resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "aws_service_access_principals.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "aws_service_access_principals.4066123156", "fms.amazonaws.com"),
				),
			},
		},
	})
}

func testAccAwsOrganizationsOrganization_FeatureSet(t *testing.T) {
	var organization organizations.Organization
	resourceName := "aws_organizations_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOrganizationsOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsOrganizationConfigFeatureSet(organizations.OrganizationFeatureSetConsolidatedBilling),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsOrganizationExists(resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "feature_set", organizations.OrganizationFeatureSetConsolidatedBilling),
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

func testAccCheckAwsOrganizationsOrganizationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).organizationsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_organizations_organization" {
			continue
		}

		params := &organizations.DescribeOrganizationInput{}

		resp, err := conn.DescribeOrganization(params)

		if isAWSErr(err, organizations.ErrCodeAWSOrganizationsNotInUseException, "") {
			return nil
		}

		if err != nil {
			return err
		}

		if resp != nil && resp.Organization != nil {
			return fmt.Errorf("Bad: Organization still exists: %q", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsOrganizationsOrganizationExists(n string, a *organizations.Organization) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Organization ID not set")
		}

		conn := testAccProvider.Meta().(*AWSClient).organizationsconn
		params := &organizations.DescribeOrganizationInput{}

		resp, err := conn.DescribeOrganization(params)

		if err != nil {
			return err
		}

		if resp == nil || resp.Organization == nil {
			return fmt.Errorf("Organization %q does not exist", rs.Primary.ID)
		}

		a = resp.Organization

		return nil
	}
}

const testAccAwsOrganizationsOrganizationConfig = "resource \"aws_organizations_organization\" \"test\" {}"

func testAccAwsOrganizationsOrganizationConfigAwsServiceAccessPrincipals1(principal1 string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {
  aws_service_access_principals = [%q]
}
`, principal1)
}

func testAccAwsOrganizationsOrganizationConfigAwsServiceAccessPrincipals2(principal1, principal2 string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {
  aws_service_access_principals = [%q, %q]
}
`, principal1, principal2)
}

func testAccAwsOrganizationsOrganizationConfigFeatureSet(featureSet string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {
  feature_set = %q
}
`, featureSet)
}
