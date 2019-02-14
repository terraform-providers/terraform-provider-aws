package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceAwsSecretsManagerSecret_Basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsSecretsManagerSecretConfig_MissingRequired,
				ExpectError: regexp.MustCompile(`must specify either arn or name`),
			},
			{
				Config:      testAccDataSourceAwsSecretsManagerSecretConfig_MultipleSpecified,
				ExpectError: regexp.MustCompile(`specify only arn or name`),
			},
			{
				Config:      testAccDataSourceAwsSecretsManagerSecretConfig_NonExistent,
				ExpectError: regexp.MustCompile(`not found`),
			},
		},
	})
}

func TestAccDataSourceAwsSecretsManagerSecret_ARN(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_secretsmanager_secret.test"
	datasourceName := "data.aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsSecretsManagerSecretConfig_ARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsSecretsManagerSecretCheck(datasourceName, resourceName),
				),
			},
		},
	})
}

func TestAccDataSourceAwsSecretsManagerSecret_Name(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_secretsmanager_secret.test"
	datasourceName := "data.aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsSecretsManagerSecretConfig_Name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsSecretsManagerSecretCheck(datasourceName, resourceName),
				),
			},
		},
	})
}

func TestAccDataSourceAwsSecretsManagerSecret_Policy(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_secretsmanager_secret.test"
	datasourceName := "data.aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsSecretsManagerSecretConfig_Policy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsSecretsManagerSecretCheck(datasourceName, resourceName),
				),
			},
		},
	})
}

func testAccDataSourceAwsSecretsManagerSecretCheck(datasourceName, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resource, ok := s.RootModule().Resources[datasourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", datasourceName)
		}

		dataSource, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", resourceName)
		}

		attrNames := []string{
			"arn",
			"description",
			"kms_key_id",
			"name",
			"policy",
			"rotation_enabled",
			"rotation_lambda_arn",
			"rotation_rules.#",
			"tags.#",
		}

		for _, attrName := range attrNames {
			if resource.Primary.Attributes[attrName] != dataSource.Primary.Attributes[attrName] {
				return fmt.Errorf(
					"%s is %s; want %s",
					attrName,
					resource.Primary.Attributes[attrName],
					dataSource.Primary.Attributes[attrName],
				)
			}
		}

		return nil
	}
}

func testAccDataSourceAwsSecretsManagerSecretConfig_ARN(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "wrong" {
  name = "%[1]s-wrong"
}
resource "aws_secretsmanager_secret" "test" {
  name = "%[1]s"
}

data "aws_secretsmanager_secret" "test" {
  arn = "${aws_secretsmanager_secret.test.arn}"
}
`, rName)
}

const testAccDataSourceAwsSecretsManagerSecretConfig_MissingRequired = `
data "aws_secretsmanager_secret" "test" {}
`

const testAccDataSourceAwsSecretsManagerSecretConfig_MultipleSpecified = `
data "aws_secretsmanager_secret" "test" {
  arn  = "arn:aws:secretsmanager:us-east-1:123456789012:secret:tf-acc-test-does-not-exist"
  name = "tf-acc-test-does-not-exist"
}
`

func testAccDataSourceAwsSecretsManagerSecretConfig_Name(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "wrong" {
  name = "%[1]s-wrong"
}
resource "aws_secretsmanager_secret" "test" {
  name = "%[1]s"
}

data "aws_secretsmanager_secret" "test" {
  name = "${aws_secretsmanager_secret.test.name}"
}
`, rName)
}

func testAccDataSourceAwsSecretsManagerSecretConfig_Policy(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name   = "%[1]s"

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

data "aws_secretsmanager_secret" "test" {
  name   = "${aws_secretsmanager_secret.test.name}"
}
`, rName)
}

const testAccDataSourceAwsSecretsManagerSecretConfig_NonExistent = `
data "aws_secretsmanager_secret" "test" {
  name = "tf-acc-test-does-not-exist"
}
`
