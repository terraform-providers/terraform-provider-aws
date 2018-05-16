package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSLambdaAlias_basic(t *testing.T) {
	var conf lambda.AliasConfiguration

	rString := acctest.RandString(8)
	roleName := fmt.Sprintf("tf_acc_role_lambda_alias_basic_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_alias_basic_%s", rString)
	attachmentName := fmt.Sprintf("tf_acc_attachment_%s", rString)
	funcName := fmt.Sprintf("tf_acc_lambda_func_alias_basic_%s", rString)
	aliasName := fmt.Sprintf("tf_acc_lambda_alias_basic_%s", rString)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLambdaAliasDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAwsLambdaAliasConfig(roleName, policyName, attachmentName, funcName, aliasName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaAliasExists("aws_lambda_alias.lambda_alias_test", &conf),
					testAccCheckAwsLambdaAttributes(&conf),
					resource.TestMatchResourceAttr("aws_lambda_alias.lambda_alias_test", "arn",
						regexp.MustCompile(`^arn:aws:lambda:[a-z]+-[a-z]+-[0-9]+:\d{12}:function:`+funcName+`:`+aliasName+`$`)),
				),
			},
		},
	})
}

func TestAccAWSLambdaAlias_nameupdate(t *testing.T) {
	var conf lambda.AliasConfiguration

	rString := acctest.RandString(8)
	roleName := fmt.Sprintf("tf_acc_role_lambda_alias_basic_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_alias_basic_%s", rString)
	attachmentName := fmt.Sprintf("tf_acc_attachment_%s", rString)
	funcName := fmt.Sprintf("tf_acc_lambda_func_alias_basic_%s", rString)
	aliasName := fmt.Sprintf("tf_acc_lambda_alias_basic_%s", rString)
	aliasNameUpdate := fmt.Sprintf("tf_acc_lambda_alias_basic_%s", acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLambdaAliasDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAwsLambdaAliasConfig(roleName, policyName, attachmentName, funcName, aliasName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaAliasExists("aws_lambda_alias.lambda_alias_test", &conf),
					testAccCheckAwsLambdaAttributes(&conf),
					resource.TestMatchResourceAttr("aws_lambda_alias.lambda_alias_test", "arn",
						regexp.MustCompile(`^arn:aws:lambda:[a-z]+-[a-z]+-[0-9]+:\d{12}:function:`+funcName+`:`+aliasName+`$`)),
				),
			},
			resource.TestStep{
				Config: testAccAwsLambdaAliasConfig(roleName, policyName, attachmentName, funcName, aliasNameUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaAliasExists("aws_lambda_alias.lambda_alias_test", &conf),
					testAccCheckAwsLambdaAttributes(&conf),
					resource.TestMatchResourceAttr("aws_lambda_alias.lambda_alias_test", "arn",
						regexp.MustCompile(`^arn:aws:lambda:[a-z]+-[a-z]+-[0-9]+:\d{12}:function:`+funcName+`:`+aliasNameUpdate+`$`)),
				),
			},
		},
	})
}

func testAccCheckAwsLambdaAliasDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).lambdaconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lambda_alias" {
			continue
		}

		_, err := conn.GetAlias(&lambda.GetAliasInput{
			FunctionName: aws.String(rs.Primary.ID),
		})

		if err == nil {
			return fmt.Errorf("Lambda alias was not deleted")
		}

	}

	return nil
}

func testAccCheckAwsLambdaAliasExists(n string, mapping *lambda.AliasConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Lambda alias not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Lambda alias not set")
		}

		conn := testAccProvider.Meta().(*AWSClient).lambdaconn

		params := &lambda.GetAliasInput{
			FunctionName: aws.String(rs.Primary.ID),
			Name:         aws.String(rs.Primary.Attributes["name"]),
		}

		getAliasConfiguration, err := conn.GetAlias(params)
		if err != nil {
			return err
		}

		*mapping = *getAliasConfiguration

		return nil
	}
}

func testAccCheckAwsLambdaAttributes(mapping *lambda.AliasConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		name := *mapping.Name
		arn := *mapping.AliasArn
		if arn == "" {
			return fmt.Errorf("Could not read Lambda alias ARN")
		}
		if name == "" {
			return fmt.Errorf("Could not read Lambda alias name")
		}
		return nil
	}
}

func testAccAwsLambdaAliasConfig(roleName, policyName, attachmentName, funcName, aliasName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "iam_for_lambda" {
  name = "%s"

  assume_role_policy = <<EOF
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
EOF
}

resource "aws_iam_policy" "policy_for_role" {
  name        = "%s"
  path        = "/"
  description = "IAM policy for for Lamda alias testing"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
      {
          "Effect": "Allow",
          "Action": [
            "lambda:*"
          ],
          "Resource": "*"
      }
  ]
}
EOF
}

resource "aws_iam_policy_attachment" "policy_attachment_for_role" {
  name       = "%s"
  roles      = ["${aws_iam_role.iam_for_lambda.name}"]
  policy_arn = "${aws_iam_policy.policy_for_role.arn}"
}

resource "aws_lambda_function" "lambda_function_test_create" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = "${aws_iam_role.iam_for_lambda.arn}"
  handler       = "exports.example"
  runtime       = "nodejs4.3"
}

resource "aws_lambda_alias" "lambda_alias_test" {
  name             = "%s"
  description      = "a sample description"
  function_name    = "${aws_lambda_function.lambda_function_test_create.arn}"
  function_version = "$LATEST"
}`, roleName, policyName, attachmentName, funcName, aliasName)
}
