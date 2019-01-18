package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/jen20/awspolicyequivalence"
)

func init() {
	resource.AddTestSweepers("aws_secretsmanager_secret", &resource.Sweeper{
		Name: "aws_secretsmanager_secret",
		F:    testSweepSecretsManagerSecrets,
	})
}

func testSweepSecretsManagerSecrets(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).secretsmanagerconn

	err = conn.ListSecretsPages(&secretsmanager.ListSecretsInput{}, func(page *secretsmanager.ListSecretsOutput, isLast bool) bool {
		if len(page.SecretList) == 0 {
			log.Print("[DEBUG] No Secrets Manager Secrets to sweep")
			return true
		}

		for _, secret := range page.SecretList {
			name := aws.StringValue(secret.Name)
			if !strings.HasPrefix(name, "tf-acc-test-") {
				log.Printf("[INFO] Skipping Secrets Manager Secret: %s", name)
				continue
			}
			log.Printf("[INFO] Deleting Secrets Manager Secret: %s", name)
			input := &secretsmanager.DeleteSecretInput{
				ForceDeleteWithoutRecovery: aws.Bool(true),
				SecretId:                   aws.String(name),
			}

			_, err := conn.DeleteSecret(input)
			if err != nil {
				if isAWSErr(err, secretsmanager.ErrCodeResourceNotFoundException, "") {
					continue
				}
				log.Printf("[ERROR] Failed to delete Secrets Manager Secret (%s): %s", name, err)
			}
		}

		return !isLast
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Secrets Manager Secret sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Secrets Manager Secrets: %s", err)
	}
	return nil
}

func TestAccAwsSecretsManagerSecret_Basic(t *testing.T) {
	var secret secretsmanager.DescribeSecretOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSecretsManagerSecretDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSecretsManagerSecretConfig_Name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecretsManagerSecretExists(resourceName, &secret),
					resource.TestMatchResourceAttr(resourceName, "arn", regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:secretsmanager:[^:]+:[^:]+:secret:%s-.+$", rName))),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "recovery_window_in_days", "30"),
					resource.TestCheckResourceAttr(resourceName, "rotation_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "rotation_lambda_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recovery_window_in_days"},
			},
		},
	})
}

func TestAccAwsSecretsManagerSecret_withNamePrefix(t *testing.T) {
	var secret secretsmanager.DescribeSecretOutput
	rName := "tf-acc-test-"
	resourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSecretsManagerSecretDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSecretsManagerSecretConfig_withNamePrefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecretsManagerSecretExists(resourceName, &secret),
					resource.TestMatchResourceAttr(resourceName, "arn", regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:secretsmanager:[^:]+:[^:]+:secret:%s.+$", rName))),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile(fmt.Sprintf("^%s", rName))),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recovery_window_in_days", "name_prefix"},
			},
		},
	})
}

func TestAccAwsSecretsManagerSecret_Description(t *testing.T) {
	var secret secretsmanager.DescribeSecretOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSecretsManagerSecretDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSecretsManagerSecretConfig_Description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecretsManagerSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				Config: testAccAwsSecretsManagerSecretConfig_Description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecretsManagerSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recovery_window_in_days"},
			},
		},
	})
}

func TestAccAwsSecretsManagerSecret_KmsKeyID(t *testing.T) {
	var secret secretsmanager.DescribeSecretOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSecretsManagerSecretDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSecretsManagerSecretConfig_KmsKeyID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecretsManagerSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttrSet(resourceName, "kms_key_id"),
				),
			},
			{
				Config: testAccAwsSecretsManagerSecretConfig_KmsKeyID_Updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecretsManagerSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttrSet(resourceName, "kms_key_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recovery_window_in_days"},
			},
		},
	})
}

func TestAccAwsSecretsManagerSecret_RecoveryWindowInDays_Recreate(t *testing.T) {
	var secret secretsmanager.DescribeSecretOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSecretsManagerSecretDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSecretsManagerSecretConfig_RecoveryWindowInDays(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecretsManagerSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "recovery_window_in_days", "0"),
				),
			},
			{
				Config: testAccAwsSecretsManagerSecretConfig_RecoveryWindowInDays(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecretsManagerSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "recovery_window_in_days", "0"),
				),
				Taint: []string{resourceName},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recovery_window_in_days"},
			},
		},
	})
}

func TestAccAwsSecretsManagerSecret_RotationLambdaARN(t *testing.T) {
	var secret secretsmanager.DescribeSecretOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSecretsManagerSecretDestroy,
		Steps: []resource.TestStep{
			// Test enabling rotation on resource creation
			{
				Config: testAccAwsSecretsManagerSecretConfig_RotationLambdaARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecretsManagerSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "rotation_enabled", "true"),
					resource.TestMatchResourceAttr(resourceName, "rotation_lambda_arn", regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:lambda:[^:]+:[^:]+:function:%s-1$", rName))),
				),
			},
			// Test updating rotation
			// We need a valid rotation function for this testing
			// InvalidRequestException: A previous rotation isn’t complete. That rotation will be reattempted.
			/*
				{
					Config: testAccAwsSecretsManagerSecretConfig_RotationLambdaARN_Updated(rName),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAwsSecretsManagerSecretExists(resourceName, &secret),
						resource.TestCheckResourceAttr(resourceName, "rotation_enabled", "true"),
						resource.TestMatchResourceAttr(resourceName, "rotation_lambda_arn", regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:lambda:[^:]+:[^:]+:function:%s-2$", rName))),
					),
				},
			*/
			// Test importing rotation
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recovery_window_in_days"},
			},
			// Test removing rotation on resource update
			{
				Config: testAccAwsSecretsManagerSecretConfig_Name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecretsManagerSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "rotation_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "rotation_lambda_arn", ""),
				),
			},
		},
	})
}

func TestAccAwsSecretsManagerSecret_RotationRules(t *testing.T) {
	var secret secretsmanager.DescribeSecretOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSecretsManagerSecretDestroy,
		Steps: []resource.TestStep{
			// Test creating rotation rules on resource creation
			{
				Config: testAccAwsSecretsManagerSecretConfig_RotationRules(rName, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecretsManagerSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "rotation_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.0.automatically_after_days", "7"),
				),
			},
			// Test updating rotation rules
			// We need a valid rotation function for this testing
			// InvalidRequestException: A previous rotation isn’t complete. That rotation will be reattempted.
			/*
				{
					Config: testAccAwsSecretsManagerSecretConfig_RotationRules(rName, 1),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAwsSecretsManagerSecretExists(resourceName, &secret),
						resource.TestCheckResourceAttr(resourceName, "rotation_enabled", "true"),
						resource.TestCheckResourceAttr(resourceName, "rotation_rules.#", "1"),
						resource.TestCheckResourceAttr(resourceName, "rotation_rules.0.automatically_after_days", "1"),
					),
				},
			*/
			// Test importing rotation rules
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recovery_window_in_days"},
			},
			// Test removing rotation rules on resource update
			{
				Config: testAccAwsSecretsManagerSecretConfig_Name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecretsManagerSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "rotation_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.#", "0"),
				),
			},
		},
	})
}

func TestAccAwsSecretsManagerSecret_Tags(t *testing.T) {
	var secret secretsmanager.DescribeSecretOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSecretsManagerSecretDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSecretsManagerSecretConfig_Tags_Single(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecretsManagerSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "tag1value"),
				),
			},
			{
				Config: testAccAwsSecretsManagerSecretConfig_Tags_SingleUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecretsManagerSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "tag1value-updated"),
				),
			},
			{
				Config: testAccAwsSecretsManagerSecretConfig_Tags_Multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecretsManagerSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "tag1value"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", "tag2value"),
				),
			},
			{
				Config: testAccAwsSecretsManagerSecretConfig_Tags_Single(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecretsManagerSecretExists(resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "tag1value"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recovery_window_in_days"},
			},
		},
	})
}

func TestAccAwsSecretsManagerSecret_policy(t *testing.T) {
	var secret secretsmanager.DescribeSecretOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_secretsmanager_secret.test"
	expectedPolicyText := `{"Version":"2012-10-17","Statement":[{"Sid":"EnableAllPermissions","Effect":"Allow","Principal":{"AWS":"*"},"Action":"secretsmanager:GetSecretValue","Resource":"*"}]}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSecretsManagerSecretDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSecretsManagerSecretConfig_Policy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecretsManagerSecretExists(resourceName, &secret),
					testAccCheckAwsSecretsManagerSecretHasPolicy(resourceName, expectedPolicyText),
				),
			},
		},
	})
}

func testAccCheckAwsSecretsManagerSecretDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).secretsmanagerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_secretsmanager_secret" {
			continue
		}

		input := &secretsmanager.DescribeSecretInput{
			SecretId: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeSecret(input)

		if err != nil {
			if isAWSErr(err, secretsmanager.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return err
		}

		if output != nil && output.DeletedDate == nil {
			return fmt.Errorf("Secret %q still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccCheckAwsSecretsManagerSecretExists(resourceName string, secret *secretsmanager.DescribeSecretOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).secretsmanagerconn
		input := &secretsmanager.DescribeSecretInput{
			SecretId: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeSecret(input)

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("Secret %q does not exist", rs.Primary.ID)
		}

		*secret = *output

		return nil
	}
}

func testAccCheckAwsSecretsManagerSecretHasPolicy(name string, expectedPolicyText string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).secretsmanagerconn
		input := &secretsmanager.GetResourcePolicyInput{
			SecretId: aws.String(rs.Primary.ID),
		}

		out, err := conn.GetResourcePolicy(input)

		if err != nil {
			return err
		}

		actualPolicyText := *out.ResourcePolicy

		equivalent, err := awspolicy.PoliciesAreEquivalent(actualPolicyText, expectedPolicyText)
		if err != nil {
			return fmt.Errorf("Error testing policy equivalence: %s", err)
		}
		if !equivalent {
			return fmt.Errorf("Non-equivalent policy error:\n\nexpected: %s\n\n     got: %s\n",
				expectedPolicyText, actualPolicyText)
		}

		return nil
	}
}

func testAccAwsSecretsManagerSecretConfig_Description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  description = "%s"
  name        = "%s"
}
`, description, rName)
}

func testAccAwsSecretsManagerSecretConfig_Name(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = "%s"
}
`, rName)
}

func testAccAwsSecretsManagerSecretConfig_withNamePrefix(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name_prefix = "%s"
}
`, rName)
}

func testAccAwsSecretsManagerSecretConfig_KmsKeyID(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test1" {
  deletion_window_in_days = 7
}

resource "aws_kms_key" "test2" {
  deletion_window_in_days = 7
}

resource "aws_secretsmanager_secret" "test" {
  kms_key_id = "${aws_kms_key.test1.id}"
  name       = "%s"
}
`, rName)
}

func testAccAwsSecretsManagerSecretConfig_KmsKeyID_Updated(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test1" {
  deletion_window_in_days = 7
}

resource "aws_kms_key" "test2" {
  deletion_window_in_days = 7
}

resource "aws_secretsmanager_secret" "test" {
  kms_key_id = "${aws_kms_key.test2.id}"
  name       = "%s"
}
`, rName)
}

func testAccAwsSecretsManagerSecretConfig_RecoveryWindowInDays(rName string, recoveryWindowInDays int) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name                    = %q
  recovery_window_in_days = %d
}
`, rName, recoveryWindowInDays)
}

func testAccAwsSecretsManagerSecretConfig_RotationLambdaARN(rName string) string {
	return baseAccAWSLambdaConfig(rName, rName, rName) + fmt.Sprintf(`
# Not a real rotation function
resource "aws_lambda_function" "test1" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s-1"
  handler       = "exports.example"
  role          = "${aws_iam_role.iam_for_lambda.arn}"
  runtime       = "nodejs8.10"
}

resource "aws_lambda_permission" "test1" {
  action         = "lambda:InvokeFunction"
  function_name  = "${aws_lambda_function.test1.function_name}"
  principal      = "secretsmanager.amazonaws.com"
  statement_id   = "AllowExecutionFromSecretsManager1"
}

# Not a real rotation function
resource "aws_lambda_function" "test2" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s-2"
  handler       = "exports.example"
  role          = "${aws_iam_role.iam_for_lambda.arn}"
  runtime       = "nodejs8.10"
}

resource "aws_lambda_permission" "test2" {
  action         = "lambda:InvokeFunction"
  function_name  = "${aws_lambda_function.test2.function_name}"
  principal      = "secretsmanager.amazonaws.com"
  statement_id   = "AllowExecutionFromSecretsManager2"
}

resource "aws_secretsmanager_secret" "test" {
  name                = "%[1]s"
  rotation_lambda_arn = "${aws_lambda_function.test1.arn}"

  depends_on = ["aws_lambda_permission.test1"]
}
`, rName)
}

func testAccAwsSecretsManagerSecretConfig_RotationRules(rName string, automaticallyAfterDays int) string {
	return baseAccAWSLambdaConfig(rName, rName, rName) + fmt.Sprintf(`
# Not a real rotation function
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s"
  handler       = "exports.example"
  role          = "${aws_iam_role.iam_for_lambda.arn}"
  runtime       = "nodejs8.10"
}

resource "aws_lambda_permission" "test" {
  action         = "lambda:InvokeFunction"
  function_name  = "${aws_lambda_function.test.function_name}"
  principal      = "secretsmanager.amazonaws.com"
  statement_id   = "AllowExecutionFromSecretsManager1"
}

resource "aws_secretsmanager_secret" "test" {
  name                = "%[1]s"
  rotation_lambda_arn = "${aws_lambda_function.test.arn}"

  rotation_rules {
    automatically_after_days = %[2]d
  }

  depends_on = ["aws_lambda_permission.test"]
}
`, rName, automaticallyAfterDays)
}

func testAccAwsSecretsManagerSecretConfig_Tags_Single(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = "%s"

  tags = {
    tag1 = "tag1value"
  }
}
`, rName)
}

func testAccAwsSecretsManagerSecretConfig_Tags_SingleUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = "%s"

  tags = {
    tag1 = "tag1value-updated"
  }
}
`, rName)
}

func testAccAwsSecretsManagerSecretConfig_Tags_Multiple(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = "%s"

  tags = {
    tag1 = "tag1value"
    tag2 = "tag2value"
  }
}
`, rName)
}

func testAccAwsSecretsManagerSecretConfig_Policy(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = "%s"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
	{
	  "Sid": "EnableAllPermissions",
	  "Effect": "Allow",
	  "Principal": {
		"AWS": "*"
	  },
	  "Action": "secretsmanager:GetSecretValue",
	  "Resource": "*"
	}
  ]
}
POLICY
}
`, rName)
}
