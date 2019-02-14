package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsSecretsManagerSecretVersion_BasicString(t *testing.T) {
	var version secretsmanager.GetSecretValueOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_secretsmanager_secret_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSecretsManagerSecretVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSecretsManagerSecretVersionConfig_SecretString(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecretsManagerSecretVersionExists(resourceName, &version),
					resource.TestCheckResourceAttr(resourceName, "secret_string", "test-string"),
					resource.TestCheckResourceAttrSet(resourceName, "version_id"),
					resource.TestCheckResourceAttr(resourceName, "version_stages.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "version_stages.3070137", "AWSCURRENT"),
					resource.TestMatchResourceAttr(resourceName, "arn",
						regexp.MustCompile(`^arn:[\w-]+:secretsmanager:[^:]+:\d{12}:secret:.+$`)),
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

func TestAccAwsSecretsManagerSecretVersion_Base64Binary(t *testing.T) {
	var version secretsmanager.GetSecretValueOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_secretsmanager_secret_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSecretsManagerSecretVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSecretsManagerSecretVersionConfig_SecretBinary(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecretsManagerSecretVersionExists(resourceName, &version),
					resource.TestCheckResourceAttr(resourceName, "secret_binary", base64Encode([]byte("test-binary"))),
					resource.TestCheckResourceAttrSet(resourceName, "version_id"),
					resource.TestCheckResourceAttr(resourceName, "version_stages.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "version_stages.3070137", "AWSCURRENT"),
					resource.TestMatchResourceAttr(resourceName, "arn",
						regexp.MustCompile(`^arn:[\w-]+:secretsmanager:[^:]+:\d{12}:secret:.+$`)),
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

func TestAccAwsSecretsManagerSecretVersion_VersionStages(t *testing.T) {
	var version secretsmanager.GetSecretValueOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_secretsmanager_secret_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSecretsManagerSecretVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSecretsManagerSecretVersionConfig_VersionStages_Single(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecretsManagerSecretVersionExists(resourceName, &version),
					resource.TestCheckResourceAttr(resourceName, "secret_string", "test-string"),
					resource.TestCheckResourceAttr(resourceName, "version_stages.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "version_stages.3070137", "AWSCURRENT"),
					resource.TestCheckResourceAttr(resourceName, "version_stages.2848565413", "one"),
				),
			},
			{
				Config: testAccAwsSecretsManagerSecretVersionConfig_VersionStages_SingleUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecretsManagerSecretVersionExists(resourceName, &version),
					resource.TestCheckResourceAttr(resourceName, "secret_string", "test-string"),
					resource.TestCheckResourceAttr(resourceName, "version_stages.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "version_stages.3070137", "AWSCURRENT"),
					resource.TestCheckResourceAttr(resourceName, "version_stages.3351840846", "two"),
				),
			},
			{
				Config: testAccAwsSecretsManagerSecretVersionConfig_VersionStages_Multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecretsManagerSecretVersionExists(resourceName, &version),
					resource.TestCheckResourceAttr(resourceName, "secret_string", "test-string"),
					resource.TestCheckResourceAttr(resourceName, "version_stages.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "version_stages.3070137", "AWSCURRENT"),
					resource.TestCheckResourceAttr(resourceName, "version_stages.2848565413", "one"),
					resource.TestCheckResourceAttr(resourceName, "version_stages.3351840846", "two"),
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

func testAccCheckAwsSecretsManagerSecretVersionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).secretsmanagerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_secretsmanager_secret_version" {
			continue
		}

		secretID, versionID, err := decodeSecretsManagerSecretVersionID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &secretsmanager.GetSecretValueInput{
			SecretId:  aws.String(secretID),
			VersionId: aws.String(versionID),
		}

		output, err := conn.GetSecretValue(input)

		if err != nil {
			if isAWSErr(err, secretsmanager.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			if isAWSErr(err, secretsmanager.ErrCodeInvalidRequestException, "You can’t perform this operation on the secret because it was deleted") {
				return nil
			}
			return err
		}

		if output == nil {
			return nil
		}

		if len(output.VersionStages) == 0 {
			return nil
		}

		if len(output.VersionStages) == 1 && aws.StringValue(output.VersionStages[0]) == "AWSCURRENT" {
			return nil
		}

		return fmt.Errorf("Secret Version %q still exists", rs.Primary.ID)
	}

	return nil

}

func testAccCheckAwsSecretsManagerSecretVersionExists(resourceName string, version *secretsmanager.GetSecretValueOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		secretID, versionID, err := decodeSecretsManagerSecretVersionID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).secretsmanagerconn

		input := &secretsmanager.GetSecretValueInput{
			SecretId:  aws.String(secretID),
			VersionId: aws.String(versionID),
		}

		output, err := conn.GetSecretValue(input)

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("Secret Version %q does not exist", rs.Primary.ID)
		}

		*version = *output

		return nil
	}
}

func testAccAwsSecretsManagerSecretVersionConfig_SecretString(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = "%s"
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = "${aws_secretsmanager_secret.test.id}"
  secret_string = "test-string"
}
`, rName)
}

func testAccAwsSecretsManagerSecretVersionConfig_SecretBinary(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = "%s"
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = "${aws_secretsmanager_secret.test.id}"
  secret_binary = "${base64encode("test-binary")}"
}
`, rName)
}

func testAccAwsSecretsManagerSecretVersionConfig_VersionStages_Single(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = "%s"
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = "${aws_secretsmanager_secret.test.id}"
  secret_string = "test-string"

  version_stages = ["one", "AWSCURRENT"]
}
`, rName)
}

func testAccAwsSecretsManagerSecretVersionConfig_VersionStages_SingleUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = "%s"
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = "${aws_secretsmanager_secret.test.id}"
  secret_string = "test-string"

  version_stages = ["two", "AWSCURRENT"]
}
`, rName)
}

func testAccAwsSecretsManagerSecretVersionConfig_VersionStages_Multiple(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = "%s"
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = "${aws_secretsmanager_secret.test.id}"
  secret_string = "test-string"

  version_stages = ["one", "two", "AWSCURRENT"]
}
`, rName)
}
