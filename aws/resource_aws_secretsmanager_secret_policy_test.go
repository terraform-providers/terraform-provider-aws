package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/secretsmanager/waiter"
)

func init() {
	resource.AddTestSweepers("aws_secretsmanager_secret_policy", &resource.Sweeper{
		Name: "aws_secretsmanager_secret_policy",
		F:    testSweepSecretsManagerSecretPolicies,
	})
}

func testSweepSecretsManagerSecretPolicies(region string) error {
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

			log.Printf("[INFO] Deleting Secrets Manager Secret Policy: %s", name)
			input := &secretsmanager.DeleteResourcePolicyInput{
				SecretId: aws.String(name),
			}

			_, err := conn.DeleteResourcePolicy(input)
			if err != nil {
				if isAWSErr(err, secretsmanager.ErrCodeResourceNotFoundException, "") {
					continue
				}
				log.Printf("[ERROR] Failed to delete Secrets Manager Secret Policy (%s): %s", name, err)
			}
		}

		return !isLast
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Secrets Manager Secret sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Secrets Manager Secrets: %w", err)
	}
	return nil
}

func TestAccAwsSecretsManagerSecretPolicy_basic(t *testing.T) {
	var policy secretsmanager.GetResourcePolicyOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_secretsmanager_secret_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSecretsManager(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSecretsManagerSecretPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSecretsManagerSecretPolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecretsManagerSecretPolicyExists(resourceName, &policy),
					resource.TestMatchResourceAttr(resourceName, "policy",
						regexp.MustCompile(`{"Action":"secretsmanager:GetSecretValue".+`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"block_public_policy"},
			},
			{
				Config: testAccAwsSecretsManagerSecretPolicyUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecretsManagerSecretPolicyExists(resourceName, &policy),
					resource.TestMatchResourceAttr(resourceName, "policy",
						regexp.MustCompile(`{"Action":"secretsmanager:\*".+`)),
				),
			},
		},
	})
}

func TestAccAwsSecretsManagerSecretPolicy_blockPublicPolicy(t *testing.T) {
	var policy secretsmanager.GetResourcePolicyOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_secretsmanager_secret_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSecretsManager(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSecretsManagerSecretPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSecretsManagerSecretPolicyBlockConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecretsManagerSecretPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "block_public_policy", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"block_public_policy"},
			},
			{
				Config: testAccAwsSecretsManagerSecretPolicyBlockConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecretsManagerSecretPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "block_public_policy", "false"),
				),
			},
			{
				Config: testAccAwsSecretsManagerSecretPolicyBlockConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecretsManagerSecretPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "block_public_policy", "true"),
				),
			},
		},
	})
}

func TestAccAwsSecretsManagerSecretPolicy_disappears(t *testing.T) {
	var policy secretsmanager.GetResourcePolicyOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_secretsmanager_secret_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSecretsManager(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSecretsManagerSecretPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSecretsManagerSecretPolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecretsManagerSecretPolicyExists(resourceName, &policy),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSecretsManagerSecretPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsSecretsManagerSecretPolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).secretsmanagerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_secretsmanager_secret_policy" {
			continue
		}

		secretInput := &secretsmanager.DescribeSecretInput{
			SecretId: aws.String(rs.Primary.ID),
		}

		var output *secretsmanager.DescribeSecretOutput

		err := resource.Retry(waiter.DeletionPropagationTimeout, func() *resource.RetryError {
			var err error
			output, err = conn.DescribeSecret(secretInput)

			if err != nil {
				return resource.NonRetryableError(err)
			}

			if output != nil && output.DeletedDate == nil {
				return resource.RetryableError(fmt.Errorf("Secret %q still exists", rs.Primary.ID))
			}

			return nil
		})

		if isResourceTimeoutError(err) {
			output, err = conn.DescribeSecret(secretInput)
		}

		if isAWSErr(err, secretsmanager.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if output != nil && output.DeletedDate == nil {
			return fmt.Errorf("Secret %q still exists", rs.Primary.ID)
		}

		input := &secretsmanager.GetResourcePolicyInput{
			SecretId: aws.String(rs.Primary.ID),
		}

		_, err = conn.GetResourcePolicy(input)

		if isAWSErr(err, secretsmanager.ErrCodeResourceNotFoundException, "") ||
			isAWSErr(err, secretsmanager.ErrCodeInvalidRequestException,
				"You can't perform this operation on the secret because it was marked for deletion.") {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil

}

func testAccCheckAwsSecretsManagerSecretPolicyExists(resourceName string, policy *secretsmanager.GetResourcePolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).secretsmanagerconn
		input := &secretsmanager.GetResourcePolicyInput{
			SecretId: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetResourcePolicy(input)

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("Secret Policy %q does not exist", rs.Primary.ID)
		}

		*policy = *output

		return nil
	}
}

func testAccAwsSecretsManagerSecretPolicyBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = %[1]q
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
}

resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_policy" "test" {
  secret_arn = aws_secretsmanager_secret.test.arn

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
	{
	  "Sid": "EnableAllPermissions",
	  "Effect": "Allow",
	  "Principal": {
		"AWS": "${aws_iam_role.test.arn}"
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

func testAccAwsSecretsManagerSecretPolicyUpdatedConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_policy" "test" {
  secret_arn = aws_secretsmanager_secret.test.arn

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
	  "Action": "secretsmanager:*",
	  "Resource": "*"
	}
  ]
}
POLICY
}
`, rName)
}

func testAccAwsSecretsManagerSecretPolicyBlockConfig(rName string, block bool) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = %[1]q
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
}

resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_policy" "test" {
  secret_arn          = aws_secretsmanager_secret.test.arn
  block_public_policy = %[2]t

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
	{
	  "Sid": "EnableAllPermissions",
	  "Effect": "Allow",
	  "Principal": {
		"AWS": "${aws_iam_role.test.arn}"
	  },
	  "Action": "secretsmanager:GetSecretValue",
	  "Resource": "*"
	}
  ]
}
POLICY
}
`, rName, block)
}
