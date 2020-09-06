package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccConfigOrganizationManagedRule_basic(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_organization_managed_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigOrganizationManagedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationManagedRuleConfigRuleIdentifier(rName, "IAM_PASSWORD_POLICY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationManagedRuleExists(resourceName, &rule),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("organization-config-rule/%s-.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "excluded_accounts.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "input_parameters", ""),
					resource.TestCheckResourceAttr(resourceName, "maximum_execution_frequency", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "resource_id_scope", ""),
					resource.TestCheckResourceAttr(resourceName, "resource_types_scope.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule_identifier", "IAM_PASSWORD_POLICY"),
					resource.TestCheckResourceAttr(resourceName, "tag_key_scope", ""),
					resource.TestCheckResourceAttr(resourceName, "tag_value_scope", ""),
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

func testAccConfigOrganizationManagedRule_disappears(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_organization_managed_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigOrganizationManagedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationManagedRuleConfigRuleIdentifier(rName, "IAM_PASSWORD_POLICY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationManagedRuleExists(resourceName, &rule),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsConfigOrganizationManagedRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccConfigOrganizationManagedRule_errorHandling(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigOrganizationManagedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccConfigOrganizationManagedRuleConfigErrorHandling(rName),
				ExpectError: regexp.MustCompile(`NoAvailableConfigurationRecorder`),
			},
		},
	})
}

func testAccConfigOrganizationManagedRule_Description(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_organization_managed_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigOrganizationManagedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationManagedRuleConfigDescription(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationManagedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigOrganizationManagedRuleConfigDescription(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationManagedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func testAccConfigOrganizationManagedRule_ExcludedAccounts(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_organization_managed_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigOrganizationManagedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationManagedRuleConfigExcludedAccounts1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationManagedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "excluded_accounts.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigOrganizationManagedRuleConfigExcludedAccounts2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationManagedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "excluded_accounts.#", "2"),
				),
			},
		},
	})
}

func testAccConfigOrganizationManagedRule_InputParameters(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_organization_managed_rule.test"

	inputParameters1 := `{"tag1Key":"CostCenter", "tag2Key":"Owner"}`
	inputParameters2 := `{"tag1Key":"Department", "tag2Key":"Owner"}`

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigOrganizationManagedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationManagedRuleConfigInputParameters(rName, inputParameters1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationManagedRuleExists(resourceName, &rule),
					resource.TestMatchResourceAttr(resourceName, "input_parameters", regexp.MustCompile(`CostCenter`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigOrganizationManagedRuleConfigInputParameters(rName, inputParameters2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationManagedRuleExists(resourceName, &rule),
					resource.TestMatchResourceAttr(resourceName, "input_parameters", regexp.MustCompile(`Department`)),
				),
			},
		},
	})
}

func testAccConfigOrganizationManagedRule_MaximumExecutionFrequency(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_organization_managed_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigOrganizationManagedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationManagedRuleConfigMaximumExecutionFrequency(rName, "One_Hour"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationManagedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "maximum_execution_frequency", "One_Hour"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigOrganizationManagedRuleConfigMaximumExecutionFrequency(rName, "Three_Hours"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationManagedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "maximum_execution_frequency", "Three_Hours"),
				),
			},
		},
	})
}

func testAccConfigOrganizationManagedRule_ResourceIdScope(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_organization_managed_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigOrganizationManagedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationManagedRuleConfigResourceIdScope(rName, "i-12345678"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationManagedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "resource_id_scope", "i-12345678"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigOrganizationManagedRuleConfigResourceIdScope(rName, "i-87654321"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationManagedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "resource_id_scope", "i-87654321"),
				),
			},
		},
	})
}

func testAccConfigOrganizationManagedRule_ResourceTypesScope(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_organization_managed_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigOrganizationManagedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationManagedRuleConfigResourceTypesScope1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationManagedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "resource_types_scope.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigOrganizationManagedRuleConfigResourceTypesScope2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationManagedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "resource_types_scope.#", "2"),
				),
			},
		},
	})
}

func testAccConfigOrganizationManagedRule_RuleIdentifier(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_organization_managed_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigOrganizationManagedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationManagedRuleConfigRuleIdentifier(rName, "EC2_INSTANCE_DETAILED_MONITORING_ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationManagedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "rule_identifier", "EC2_INSTANCE_DETAILED_MONITORING_ENABLED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigOrganizationManagedRuleConfigRuleIdentifier(rName, "EC2_INSTANCE_NO_PUBLIC_IP"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationManagedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "rule_identifier", "EC2_INSTANCE_NO_PUBLIC_IP"),
				),
			},
		},
	})
}

func testAccConfigOrganizationManagedRule_TagKeyScope(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_organization_managed_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigOrganizationManagedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationManagedRuleConfigTagKeyScope(rName, "key1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationManagedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "tag_key_scope", "key1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigOrganizationManagedRuleConfigTagKeyScope(rName, "key2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationManagedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "tag_key_scope", "key2"),
				),
			},
		},
	})
}

func testAccConfigOrganizationManagedRule_TagValueScope(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_organization_managed_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigOrganizationManagedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationManagedRuleConfigTagValueScope(rName, "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationManagedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "tag_value_scope", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigOrganizationManagedRuleConfigTagValueScope(rName, "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationManagedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "tag_value_scope", "value2"),
				),
			},
		},
	})
}

func testAccCheckConfigOrganizationManagedRuleExists(resourceName string, ocr *configservice.OrganizationConfigRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not Found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).configconn

		rule, err := configDescribeOrganizationConfigRule(conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error describing Config Organization Managed Rule (%s): %s", rs.Primary.ID, err)
		}

		if rule == nil {
			return fmt.Errorf("Config Organization Managed Rule (%s) not found", rs.Primary.ID)
		}

		*ocr = *rule

		return nil
	}
}

func testAccCheckConfigOrganizationManagedRuleDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).configconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_config_organization_managed_rule" {
			continue
		}

		rule, err := configDescribeOrganizationConfigRule(conn, rs.Primary.ID)

		if isAWSErr(err, configservice.ErrCodeNoSuchOrganizationConfigRuleException, "") {
			continue
		}

		if err != nil {
			return fmt.Errorf("error describing Config Organization Managed Rule (%s): %s", rs.Primary.ID, err)
		}

		if rule != nil {
			return fmt.Errorf("Config Organization Managed Rule (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccConfigOrganizationManagedRuleConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {
}

resource "aws_config_configuration_recorder" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "config.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSConfigRole"
  role       = aws_iam_role.test.name
}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["config-multiaccountsetup.amazonaws.com"]
  feature_set                   = "ALL"
}
`, rName)
}

func testAccConfigOrganizationManagedRuleConfigDescription(rName, description string) string {
	return testAccConfigOrganizationManagedRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_managed_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_organizations_organization.test]

  description     = %[2]q
  name            = %[1]q
  rule_identifier = "IAM_PASSWORD_POLICY"
}
`, rName, description)
}

func testAccConfigOrganizationManagedRuleConfigErrorHandling(rName string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["config-multiaccountsetup.amazonaws.com"]
  feature_set                   = "ALL"
}

resource "aws_config_organization_managed_rule" "test" {
  depends_on = [aws_organizations_organization.test]

  name            = %[1]q
  rule_identifier = "IAM_PASSWORD_POLICY"
}
`, rName)
}

func testAccConfigOrganizationManagedRuleConfigExcludedAccounts1(rName string) string {
	return testAccConfigOrganizationManagedRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_managed_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_organizations_organization.test]

  excluded_accounts = ["111111111111"]
  name              = %[1]q
  rule_identifier   = "IAM_PASSWORD_POLICY"
}
`, rName)
}

func testAccConfigOrganizationManagedRuleConfigExcludedAccounts2(rName string) string {
	return testAccConfigOrganizationManagedRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_managed_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_organizations_organization.test]

  excluded_accounts = ["111111111111", "222222222222"]
  name              = %[1]q
  rule_identifier   = "IAM_PASSWORD_POLICY"
}
`, rName)
}

func testAccConfigOrganizationManagedRuleConfigInputParameters(rName, inputParameters string) string {
	return testAccConfigOrganizationManagedRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_managed_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_organizations_organization.test]

  input_parameters = <<PARAMS
%[2]s
PARAMS

  name            = %[1]q
  rule_identifier = "REQUIRED_TAGS"
}
`, rName, inputParameters)
}

func testAccConfigOrganizationManagedRuleConfigMaximumExecutionFrequency(rName, maximumExecutionFrequency string) string {
	return testAccConfigOrganizationManagedRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_managed_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_organizations_organization.test]

  maximum_execution_frequency = %[2]q
  name                        = %[1]q
  rule_identifier             = "IAM_PASSWORD_POLICY"
}
`, rName, maximumExecutionFrequency)
}

func testAccConfigOrganizationManagedRuleConfigResourceIdScope(rName, resourceIdScope string) string {
	return testAccConfigOrganizationManagedRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_managed_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_organizations_organization.test]

  name                 = %[1]q
  resource_id_scope    = %[2]q
  resource_types_scope = ["AWS::EC2::Instance"]
  rule_identifier      = "EC2_INSTANCE_DETAILED_MONITORING_ENABLED"
}
`, rName, resourceIdScope)
}

func testAccConfigOrganizationManagedRuleConfigResourceTypesScope1(rName string) string {
	return testAccConfigOrganizationManagedRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_managed_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_organizations_organization.test]

  input_parameters = <<EOF
{
  "tag1Key": "CostCenter",
  "tag2Key": "Owner"
}
EOF

  name                 = %[1]q
  resource_types_scope = ["AWS::EC2::Instance"]
  rule_identifier      = "REQUIRED_TAGS"
}
`, rName)
}

func testAccConfigOrganizationManagedRuleConfigResourceTypesScope2(rName string) string {
	return testAccConfigOrganizationManagedRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_managed_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_organizations_organization.test]

  input_parameters = <<EOF
{
  "tag1Key": "CostCenter",
  "tag2Key": "Owner"
}
EOF

  name                 = %[1]q
  resource_types_scope = ["AWS::EC2::Instance", "AWS::EC2::VPC"]
  rule_identifier      = "REQUIRED_TAGS"
}
`, rName)
}

func testAccConfigOrganizationManagedRuleConfigRuleIdentifier(rName, ruleIdentifier string) string {
	return testAccConfigOrganizationManagedRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_managed_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_organizations_organization.test]

  name            = %[1]q
  rule_identifier = %[2]q
}
`, rName, ruleIdentifier)
}

func testAccConfigOrganizationManagedRuleConfigTagKeyScope(rName, tagKeyScope string) string {
	return testAccConfigOrganizationManagedRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_managed_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_organizations_organization.test]

  name            = %[1]q
  rule_identifier = "EC2_INSTANCE_DETAILED_MONITORING_ENABLED"
  tag_key_scope   = %[2]q
}
`, rName, tagKeyScope)
}

func testAccConfigOrganizationManagedRuleConfigTagValueScope(rName, tagValueScope string) string {
	return testAccConfigOrganizationManagedRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_managed_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_organizations_organization.test]

  name            = %[1]q
  rule_identifier = "EC2_INSTANCE_DETAILED_MONITORING_ENABLED"
  tag_key_scope   = "key1"
  tag_value_scope = %[2]q
}
`, rName, tagValueScope)
}
