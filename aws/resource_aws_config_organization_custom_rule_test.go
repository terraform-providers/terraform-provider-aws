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

func testAccConfigOrganizationCustomRule_basic(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := acctest.RandomWithPrefix("tf-acc-test")
	lambdaFunctionResourceName := "aws_lambda_function.test"
	resourceName := "aws_config_organization_custom_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigOrganizationCustomRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationCustomRuleConfigTriggerTypes1(rName, "ConfigurationItemChangeNotification"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationCustomRuleExists(resourceName, &rule),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("organization-config-rule/%s-.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "excluded_accounts.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "input_parameters", ""),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_function_arn", lambdaFunctionResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "maximum_execution_frequency", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "resource_id_scope", ""),
					resource.TestCheckResourceAttr(resourceName, "resource_types_scope.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tag_key_scope", ""),
					resource.TestCheckResourceAttr(resourceName, "tag_value_scope", ""),
					resource.TestCheckResourceAttr(resourceName, "trigger_types.#", "1"),
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

func testAccConfigOrganizationCustomRule_disappears(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_organization_custom_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigOrganizationCustomRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationCustomRuleConfigTriggerTypes1(rName, "ConfigurationItemChangeNotification"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationCustomRuleExists(resourceName, &rule),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsConfigOrganizationCustomRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccConfigOrganizationCustomRule_errorHandling(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigOrganizationCustomRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccConfigOrganizationCustomRuleConfigErrorHandling(rName),
				ExpectError: regexp.MustCompile(`InsufficientPermission`),
			},
		},
	})
}

func testAccConfigOrganizationCustomRule_Description(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_organization_custom_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigOrganizationCustomRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationCustomRuleConfigDescription(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigOrganizationCustomRuleConfigDescription(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func testAccConfigOrganizationCustomRule_ExcludedAccounts(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_organization_custom_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigOrganizationCustomRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationCustomRuleConfigExcludedAccounts1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "excluded_accounts.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigOrganizationCustomRuleConfigExcludedAccounts2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "excluded_accounts.#", "2"),
				),
			},
		},
	})
}

func testAccConfigOrganizationCustomRule_InputParameters(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_organization_custom_rule.test"

	inputParameters1 := `{"tag1Key":"CostCenter", "tag2Key":"Owner"}`
	inputParameters2 := `{"tag1Key":"Department", "tag2Key":"Owner"}`

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigOrganizationCustomRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationCustomRuleConfigInputParameters(rName, inputParameters1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestMatchResourceAttr(resourceName, "input_parameters", regexp.MustCompile(`CostCenter`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigOrganizationCustomRuleConfigInputParameters(rName, inputParameters2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestMatchResourceAttr(resourceName, "input_parameters", regexp.MustCompile(`Department`)),
				),
			},
		},
	})
}

func testAccConfigOrganizationCustomRule_LambdaFunctionArn(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := acctest.RandomWithPrefix("tf-acc-test")
	lambdaFunctionResourceName1 := "aws_lambda_function.test"
	lambdaFunctionResourceName2 := "aws_lambda_function.test2"
	resourceName := "aws_config_organization_custom_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigOrganizationCustomRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationCustomRuleConfigLambdaFunctionArn1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_function_arn", lambdaFunctionResourceName1, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigOrganizationCustomRuleConfigLambdaFunctionArn2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_function_arn", lambdaFunctionResourceName2, "arn"),
				),
			},
		},
	})
}

func testAccConfigOrganizationCustomRule_MaximumExecutionFrequency(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_organization_custom_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigOrganizationCustomRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationCustomRuleConfigMaximumExecutionFrequency(rName, "One_Hour"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "maximum_execution_frequency", "One_Hour"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigOrganizationCustomRuleConfigMaximumExecutionFrequency(rName, "Three_Hours"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "maximum_execution_frequency", "Three_Hours"),
				),
			},
		},
	})
}

func testAccConfigOrganizationCustomRule_ResourceIdScope(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_organization_custom_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigOrganizationCustomRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationCustomRuleConfigResourceIdScope(rName, "i-12345678"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "resource_id_scope", "i-12345678"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigOrganizationCustomRuleConfigResourceIdScope(rName, "i-87654321"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "resource_id_scope", "i-87654321"),
				),
			},
		},
	})
}

func testAccConfigOrganizationCustomRule_ResourceTypesScope(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_organization_custom_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigOrganizationCustomRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationCustomRuleConfigResourceTypesScope1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "resource_types_scope.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigOrganizationCustomRuleConfigResourceTypesScope2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "resource_types_scope.#", "2"),
				),
			},
		},
	})
}

func testAccConfigOrganizationCustomRule_TagKeyScope(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_organization_custom_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigOrganizationCustomRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationCustomRuleConfigTagKeyScope(rName, "key1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "tag_key_scope", "key1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigOrganizationCustomRuleConfigTagKeyScope(rName, "key2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "tag_key_scope", "key2"),
				),
			},
		},
	})
}

func testAccConfigOrganizationCustomRule_TagValueScope(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_organization_custom_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigOrganizationCustomRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationCustomRuleConfigTagValueScope(rName, "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "tag_value_scope", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigOrganizationCustomRuleConfigTagValueScope(rName, "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "tag_value_scope", "value2"),
				),
			},
		},
	})
}

func testAccConfigOrganizationCustomRule_TriggerTypes(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_organization_custom_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigOrganizationCustomRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationCustomRuleConfigTriggerTypes1(rName, "ScheduledNotification"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "trigger_types.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigOrganizationCustomRuleConfigTriggerTypes2(rName, "ConfigurationItemChangeNotification", "OversizedConfigurationItemChangeNotification"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "trigger_types.#", "2"),
				),
			},
		},
	})
}

func testAccCheckConfigOrganizationCustomRuleExists(resourceName string, ocr *configservice.OrganizationConfigRule) resource.TestCheckFunc {
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

func testAccCheckConfigOrganizationCustomRuleDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).configconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_config_organization_custom_rule" {
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

func testAccConfigOrganizationCustomRuleConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {
}

resource "aws_config_configuration_recorder" "test" {
  depends_on = [aws_iam_role_policy_attachment.config]

  name     = %[1]q
  role_arn = aws_iam_role.config.arn
}

resource "aws_iam_role" "config" {
  name = "%[1]s-config"

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

resource "aws_iam_role_policy_attachment" "config" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSConfigRole"
  role       = aws_iam_role.config.name
}

resource "aws_iam_role" "lambda" {
  name = "%[1]s-lambda"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "lambda" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSConfigRulesExecutionRole"
  role       = aws_iam_role.lambda.name
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}

resource "aws_lambda_permission" "test" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.arn
  principal     = "config.amazonaws.com"
  statement_id  = "AllowExecutionFromConfig"
}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["config-multiaccountsetup.amazonaws.com"]
  feature_set                   = "ALL"
}
`, rName)
}

func testAccConfigOrganizationCustomRuleConfigDescription(rName, description string) string {
	return testAccConfigOrganizationCustomRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_custom_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_lambda_permission.test, aws_organizations_organization.test]

  description         = %[2]q
  lambda_function_arn = aws_lambda_function.test.arn
  name                = %[1]q
  trigger_types       = ["ScheduledNotification"]
}
`, rName, description)
}

func testAccConfigOrganizationCustomRuleConfigErrorHandling(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {
}

resource "aws_iam_role" "lambda" {
  name = "%[1]s-lambda"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "lambda" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSConfigRulesExecutionRole"
  role       = aws_iam_role.lambda.name
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["config-multiaccountsetup.amazonaws.com"]
  feature_set                   = "ALL"
}

resource "aws_config_organization_custom_rule" "test" {
  depends_on = [aws_organizations_organization.test]

  lambda_function_arn = aws_lambda_function.test.arn
  name                = %[1]q
  trigger_types       = ["ScheduledNotification"]
}
`, rName)
}

func testAccConfigOrganizationCustomRuleConfigExcludedAccounts1(rName string) string {
	return testAccConfigOrganizationCustomRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_custom_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_lambda_permission.test, aws_organizations_organization.test]

  excluded_accounts   = ["111111111111"]
  lambda_function_arn = aws_lambda_function.test.arn
  name                = %[1]q
  trigger_types       = ["ScheduledNotification"]
}
`, rName)
}

func testAccConfigOrganizationCustomRuleConfigExcludedAccounts2(rName string) string {
	return testAccConfigOrganizationCustomRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_custom_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_lambda_permission.test, aws_organizations_organization.test]

  excluded_accounts   = ["111111111111", "222222222222"]
  lambda_function_arn = aws_lambda_function.test.arn
  name                = %[1]q
  trigger_types       = ["ScheduledNotification"]
}
`, rName)
}

func testAccConfigOrganizationCustomRuleConfigInputParameters(rName, inputParameters string) string {
	return testAccConfigOrganizationCustomRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_custom_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_lambda_permission.test, aws_organizations_organization.test]

  input_parameters = <<PARAMS
%[2]s
PARAMS

  lambda_function_arn = aws_lambda_function.test.arn
  name                = %[1]q
  trigger_types       = ["ScheduledNotification"]
}
`, rName, inputParameters)
}

func testAccConfigOrganizationCustomRuleConfigLambdaFunctionArn1(rName string) string {
	return testAccConfigOrganizationCustomRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_custom_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_lambda_permission.test, aws_organizations_organization.test]

  lambda_function_arn = aws_lambda_function.test.arn
  name                = %[1]q
  trigger_types       = ["ScheduledNotification"]
}
`, rName)
}

func testAccConfigOrganizationCustomRuleConfigLambdaFunctionArn2(rName string) string {
	return testAccConfigOrganizationCustomRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_lambda_function" "test2" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s2"
  role          = aws_iam_role.lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}

resource "aws_lambda_permission" "test2" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test2.arn
  principal     = "config.amazonaws.com"
  statement_id  = "AllowExecutionFromConfig"
}

resource "aws_config_organization_custom_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_lambda_permission.test2, aws_organizations_organization.test]

  lambda_function_arn = aws_lambda_function.test2.arn
  name                = %[1]q
  trigger_types       = ["ScheduledNotification"]
}
`, rName)
}

func testAccConfigOrganizationCustomRuleConfigMaximumExecutionFrequency(rName, maximumExecutionFrequency string) string {
	return testAccConfigOrganizationCustomRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_custom_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_lambda_permission.test, aws_organizations_organization.test]

  lambda_function_arn         = aws_lambda_function.test.arn
  maximum_execution_frequency = %[2]q
  name                        = %[1]q
  trigger_types               = ["ScheduledNotification"]
}
`, rName, maximumExecutionFrequency)
}

func testAccConfigOrganizationCustomRuleConfigResourceIdScope(rName, resourceIdScope string) string {
	return testAccConfigOrganizationCustomRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_custom_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_lambda_permission.test, aws_organizations_organization.test]

  lambda_function_arn  = aws_lambda_function.test.arn
  name                 = %[1]q
  resource_id_scope    = %[2]q
  resource_types_scope = ["AWS::EC2::Instance"]
  trigger_types        = ["ScheduledNotification"]
}
`, rName, resourceIdScope)
}

func testAccConfigOrganizationCustomRuleConfigResourceTypesScope1(rName string) string {
	return testAccConfigOrganizationCustomRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_custom_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_lambda_permission.test, aws_organizations_organization.test]

  lambda_function_arn  = aws_lambda_function.test.arn
  name                 = %[1]q
  resource_types_scope = ["AWS::EC2::Instance"]
  trigger_types        = ["ScheduledNotification"]
}
`, rName)
}

func testAccConfigOrganizationCustomRuleConfigResourceTypesScope2(rName string) string {
	return testAccConfigOrganizationCustomRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_custom_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_lambda_permission.test, aws_organizations_organization.test]

  lambda_function_arn  = aws_lambda_function.test.arn
  name                 = %[1]q
  resource_types_scope = ["AWS::EC2::Instance", "AWS::EC2::VPC"]
  trigger_types        = ["ScheduledNotification"]
}
`, rName)
}

func testAccConfigOrganizationCustomRuleConfigTagKeyScope(rName, tagKeyScope string) string {
	return testAccConfigOrganizationCustomRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_custom_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_lambda_permission.test, aws_organizations_organization.test]

  lambda_function_arn = aws_lambda_function.test.arn
  name                = %[1]q
  tag_key_scope       = %[2]q
  trigger_types       = ["ScheduledNotification"]
}
`, rName, tagKeyScope)
}

func testAccConfigOrganizationCustomRuleConfigTagValueScope(rName, tagValueScope string) string {
	return testAccConfigOrganizationCustomRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_custom_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_lambda_permission.test, aws_organizations_organization.test]

  lambda_function_arn = aws_lambda_function.test.arn
  name                = %[1]q
  tag_key_scope       = "key1"
  tag_value_scope     = %[2]q
  trigger_types       = ["ScheduledNotification"]
}
`, rName, tagValueScope)
}

func testAccConfigOrganizationCustomRuleConfigTriggerTypes1(rName, triggerType1 string) string {
	return testAccConfigOrganizationCustomRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_custom_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_lambda_permission.test, aws_organizations_organization.test]

  lambda_function_arn = aws_lambda_function.test.arn
  name                = %[1]q
  trigger_types       = [%[2]q]
}
`, rName, triggerType1)
}

func testAccConfigOrganizationCustomRuleConfigTriggerTypes2(rName, triggerType1, triggerType2 string) string {
	return testAccConfigOrganizationCustomRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_custom_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_lambda_permission.test, aws_organizations_organization.test]

  lambda_function_arn = aws_lambda_function.test.arn
  name                = %[1]q
  trigger_types       = [%[2]q, %[3]q]
}
`, rName, triggerType1, triggerType2)
}
