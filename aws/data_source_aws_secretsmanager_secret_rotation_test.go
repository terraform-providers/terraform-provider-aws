package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsSecretsManagerSecretRotation_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_secretsmanager_secret_rotation.test"
	datasourceName := "data.aws_secretsmanager_secret_rotation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSSecretsManager(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsSecretsManagerSecretRotationConfig_NonExistent,
				ExpectError: regexp.MustCompile(`ResourceNotFoundException`),
			},
			{
				Config: testAccDataSourceAwsSecretsManagerSecretRotationConfig_Default(rName, 7),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "rotation_enabled", resourceName, "rotation_enabled"),
					resource.TestCheckResourceAttrPair(datasourceName, "rotation_lambda_arn", resourceName, "rotation_lambda_arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "rotation_rules.#", resourceName, "rotation_rules.#"),
				),
			},
		},
	})
}

const testAccDataSourceAwsSecretsManagerSecretRotationConfig_NonExistent = `
data "aws_secretsmanager_secret_rotation" "test" {
  secret_id = "tf-acc-test-does-not-exist"
}
`

func testAccDataSourceAwsSecretsManagerSecretRotationConfig_Default(rName string, automaticallyAfterDays int) string {
	return baseAccAWSLambdaConfig(rName, rName, rName) + fmt.Sprintf(`
# Not a real rotation function
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s-1"
  handler       = "exports.example"
  role          = aws_iam_role.iam_for_lambda.arn
  runtime       = "nodejs12.x"
}

resource "aws_lambda_permission" "test" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.function_name
  principal     = "secretsmanager.amazonaws.com"
  statement_id  = "AllowExecutionFromSecretsManager"
}

resource "aws_secretsmanager_secret" "test" {
  name = "%[1]s"
}

resource "aws_secretsmanager_secret_rotation" "test" {
  secret_id           = aws_secretsmanager_secret.test.id
  rotation_lambda_arn = aws_lambda_function.test.arn

  rotation_rules {
    automatically_after_days = %[2]d
  }
}

data "aws_secretsmanager_secret_rotation" "test" {
  secret_id = aws_secretsmanager_secret_rotation.test.secret_id
}
`, rName, automaticallyAfterDays)
}
