package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSLambdaCodeSigningConfig_basic(t *testing.T) {
	resourceName := "aws_lambda_code_signing_config.code_signing_config"
	signingProfile1 := "aws_signer_signing_profile.test1"
	signingProfile2 := "aws_signer_signing_profile.test2"
	var conf lambda.GetCodeSigningConfigOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCodeSigningConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaCodeSigningConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCodeSigningConfigExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", "Code Signing Config for test account"),
					resource.TestCheckResourceAttr(resourceName, "allowed_publishers.0.signing_profile_version_arns.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "allowed_publishers.0.signing_profile_version_arns.*", signingProfile1, "version_arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "allowed_publishers.0.signing_profile_version_arns.*", signingProfile2, "version_arn"),
					resource.TestCheckResourceAttr(resourceName, "policies.0.untrusted_artifact_on_deployment", "Warn"),
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

func TestAccAWSLambdaCodeSigningConfig_UpdatePolicy(t *testing.T) {
	resourceName := "aws_lambda_code_signing_config.code_signing_config"
	var conf lambda.GetCodeSigningConfigOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCodeSigningConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaCodeSigningConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCodeSigningConfigExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", "Code Signing Config for test account"),
					resource.TestCheckResourceAttr(resourceName, "policies.0.untrusted_artifact_on_deployment", "Warn"),
				),
			},
			{
				Config: testAccAWSLambdaCodeSigningConfigUpdatePolicy(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCodeSigningConfigExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "policies.0.untrusted_artifact_on_deployment", "Enforce"),
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

func TestAccAWSLambdaCodeSigningConfig_UpdatePublishers(t *testing.T) {
	resourceName := "aws_lambda_code_signing_config.code_signing_config"
	signingProfile1 := "aws_signer_signing_profile.test1"
	signingProfile2 := "aws_signer_signing_profile.test2"
	var conf lambda.GetCodeSigningConfigOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCodeSigningConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaCodeSigningConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCodeSigningConfigExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", "Code Signing Config for test account"),
					resource.TestCheckResourceAttr(resourceName, "allowed_publishers.0.signing_profile_version_arns.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "allowed_publishers.0.signing_profile_version_arns.*", signingProfile1, "version_arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "allowed_publishers.0.signing_profile_version_arns.*", signingProfile2, "version_arn"),
				),
			},
			{
				Config: testAccAWSLambdaCodeSigningConfigUpdatePublishers(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCodeSigningConfigExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "allowed_publishers.0.signing_profile_version_arns.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "allowed_publishers.0.signing_profile_version_arns.*", signingProfile1, "version_arn"),
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

func testAccAWSLambdaCodeSigningConfigUpdatePublishers() string {
	return `
resource "aws_signer_signing_profile" "test1" {
  platform_id = "AWSLambda-SHA384-ECDSA"
}

resource "aws_signer_signing_profile" "test2" {
  platform_id = "AWSLambda-SHA384-ECDSA"
}

resource "aws_lambda_code_signing_config" "code_signing_config" {
  allowed_publishers {
    signing_profile_version_arns = [
      aws_signer_signing_profile.test1.version_arn
    ]
  }
}`
}

func testAccAWSLambdaCodeSigningConfigUpdatePolicy() string {
	return `
resource "aws_signer_signing_profile" "test1" {
  platform_id = "AWSLambda-SHA384-ECDSA"
}

resource "aws_signer_signing_profile" "test2" {
  platform_id = "AWSLambda-SHA384-ECDSA"
}

resource "aws_lambda_code_signing_config" "code_signing_config" {
  allowed_publishers {
    signing_profile_version_arns = [
      aws_signer_signing_profile.test1.version_arn,
      aws_signer_signing_profile.test2.version_arn
    ]
  }

  policies {
    untrusted_artifact_on_deployment = "Enforce"
  }
}`
}

func testAccAWSLambdaCodeSigningConfigBasic() string {
	return `
resource "aws_signer_signing_profile" "test1" {
  platform_id = "AWSLambda-SHA384-ECDSA"
}

resource "aws_signer_signing_profile" "test2" {
  platform_id = "AWSLambda-SHA384-ECDSA"
}

resource "aws_lambda_code_signing_config" "code_signing_config" {
  allowed_publishers {
    signing_profile_version_arns = [
      aws_signer_signing_profile.test1.version_arn,
      aws_signer_signing_profile.test2.version_arn
    ]
  }

  policies {
    untrusted_artifact_on_deployment = "Warn"
  }

  description = "Code Signing Config for test account"
}`
}

func testAccCheckAwsCodeSigningConfigExists(n string, mapping *lambda.GetCodeSigningConfigOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Code Signing Config not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Code Signing Config ID not set")
		}

		conn := testAccProvider.Meta().(*AWSClient).lambdaconn

		params := &lambda.GetCodeSigningConfigInput{
			CodeSigningConfigArn: aws.String(rs.Primary.ID),
		}

		getCodeSigningConfig, err := conn.GetCodeSigningConfig(params)
		if err != nil {
			return err
		}

		*mapping = *getCodeSigningConfig

		return nil
	}
}

func testAccCheckCodeSigningConfigDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).lambdaconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lambda_code_signing_config" {
			continue
		}

		_, err := conn.GetCodeSigningConfig(&lambda.GetCodeSigningConfigInput{
			CodeSigningConfigArn: aws.String(rs.Primary.ID),
		})

		if tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Code Signing Config still exists")

	}

	return nil

}
