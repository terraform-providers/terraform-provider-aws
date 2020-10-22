package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfawsresource"
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
					resource.TestCheckResourceAttr(resourceName, "accounts.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "accounts.0.arn", resourceName, "master_account_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "accounts.0.email", resourceName, "master_account_email"),
					resource.TestCheckResourceAttrPair(resourceName, "accounts.0.id", resourceName, "master_account_id"),
					testAccMatchResourceAttrGlobalARN(resourceName, "arn", "organizations", regexp.MustCompile(`organization/o-.+`)),
					resource.TestCheckResourceAttr(resourceName, "aws_service_access_principals.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "feature_set", organizations.OrganizationFeatureSetAll),
					testAccMatchResourceAttrGlobalARN(resourceName, "master_account_arn", "organizations", regexp.MustCompile(`account/o-.+/.+`)),
					resource.TestMatchResourceAttr(resourceName, "master_account_email", regexp.MustCompile(`.+@.+`)),
					testAccCheckResourceAttrAccountID(resourceName, "master_account_id"),
					resource.TestCheckResourceAttr(resourceName, "non_master_accounts.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "roots.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "roots.0.id", regexp.MustCompile(`r-[a-z0-9]{4,32}`)),
					resource.TestCheckResourceAttrSet(resourceName, "roots.0.name"),
					resource.TestCheckResourceAttrSet(resourceName, "roots.0.arn"),
					resource.TestCheckResourceAttr(resourceName, "roots.0.policy_types.#", "0"),
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
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "aws_service_access_principals.*", "config.amazonaws.com"),
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
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "aws_service_access_principals.*", "config.amazonaws.com"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "aws_service_access_principals.*", "ds.amazonaws.com"),
				),
			},
			{
				Config: testAccAwsOrganizationsOrganizationConfigAwsServiceAccessPrincipals1("fms.amazonaws.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsOrganizationExists(resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "aws_service_access_principals.#", "1"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "aws_service_access_principals.*", "fms.amazonaws.com"),
				),
			},
		},
	})
}

func testAccAwsOrganizationsOrganization_EnabledPolicyTypes(t *testing.T) {
	var organization organizations.Organization
	resourceName := "aws_organizations_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOrganizationsOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsOrganizationConfigEnabledPolicyTypes1(organizations.PolicyTypeServiceControlPolicy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsOrganizationExists(resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.0", organizations.PolicyTypeServiceControlPolicy),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsOrganizationsOrganizationConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsOrganizationExists(resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.#", "0"),
				),
			},
			{
				Config: testAccAwsOrganizationsOrganizationConfigEnabledPolicyTypes1(organizations.PolicyTypeAiservicesOptOutPolicy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsOrganizationExists(resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.0", organizations.PolicyTypeAiservicesOptOutPolicy),
				),
			},
			{
				Config: testAccAwsOrganizationsOrganizationConfigEnabledPolicyTypes1(organizations.PolicyTypeServiceControlPolicy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsOrganizationExists(resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.0", organizations.PolicyTypeServiceControlPolicy),
				),
			},
			{
				Config: testAccAwsOrganizationsOrganizationConfigEnabledPolicyTypes1(organizations.PolicyTypeBackupPolicy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsOrganizationExists(resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.0", organizations.PolicyTypeBackupPolicy),
				),
			},
			{
				Config: testAccAwsOrganizationsOrganizationConfigEnabledPolicyTypes1(organizations.PolicyTypeTagPolicy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsOrganizationExists(resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.0", organizations.PolicyTypeTagPolicy),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsOrganizationsOrganizationConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsOrganizationExists(resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.#", "0"),
				),
			},
			{
				Config: testAccAwsOrganizationsOrganizationConfigEnabledPolicyTypes1(organizations.PolicyTypeTagPolicy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsOrganizationExists(resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.#", "1"),
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

func testAccAwsOrganizationsOrganizationConfigEnabledPolicyTypes1(policyType1 string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {
  enabled_policy_types = [%[1]q]
}
`, policyType1)
}

func testAccAwsOrganizationsOrganizationConfigFeatureSet(featureSet string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {
  feature_set = %q
}
`, featureSet)
}

func TestFlattenOrganizationsRoots(t *testing.T) {
	roots := []*organizations.Root{
		{
			Name: aws.String("Root1"),
			Arn:  aws.String("arn:1"),
			Id:   aws.String("r-1"),
			PolicyTypes: []*organizations.PolicyTypeSummary{
				{
					Status: aws.String("ENABLED"),
					Type:   aws.String("SERVICE_CONTROL_POLICY"),
				},
				{
					Status: aws.String("DISABLED"),
					Type:   aws.String("SERVICE_CONTROL_POLICY"),
				},
			},
		},
	}
	result := flattenOrganizationsRoots(roots)

	if len(result) != len(roots) {
		t.Fatalf("expected result to have %d elements, got %d", len(roots), len(result))
	}

	for i, r := range roots {
		if aws.StringValue(r.Name) != result[i]["name"] {
			t.Fatalf(`expected result[%d]["name"] to equal %q, got %q`, i, aws.StringValue(r.Name), result[i]["name"])
		}
		if aws.StringValue(r.Arn) != result[i]["arn"] {
			t.Fatalf(`expected result[%d]["arn"] to equal %q, got %q`, i, aws.StringValue(r.Arn), result[i]["arn"])
		}
		if aws.StringValue(r.Id) != result[i]["id"] {
			t.Fatalf(`expected result[%d]["id"] to equal %q, got %q`, i, aws.StringValue(r.Id), result[i]["id"])
		}
		if result[i]["policy_types"] == nil {
			continue
		}
		if types, ok := result[i]["policy_types"].([]map[string]interface{}); ok {
			testFlattenOrganizationsRootPolicyTypes(t, i, types, r.PolicyTypes)
			continue
		}
		t.Fatalf(`result[%d]["policy_types"] could not be converted to []map[string]interface{}`, i)
	}
}

func testFlattenOrganizationsRootPolicyTypes(t *testing.T, index int, result []map[string]interface{}, types []*organizations.PolicyTypeSummary) {
	if len(result) != len(types) {
		t.Fatalf(`expected result[%d]["policy_types"] to have %d elements, got %d`, index, len(types), len(result))
	}
	for i, v := range types {
		if aws.StringValue(v.Status) != result[i]["status"] {
			t.Fatalf(`expected result[%d]["policy_types"][%d]["status"] to equal %q, got %q`, index, i, aws.StringValue(v.Status), result[i]["status"])
		}
		if aws.StringValue(v.Type) != result[i]["type"] {
			t.Fatalf(`expected result[%d]["policy_types"][%d]["type"] to equal %q, got %q`, index, i, aws.StringValue(v.Type), result[i]["type"])
		}
	}
}
