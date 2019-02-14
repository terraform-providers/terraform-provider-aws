package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceAwsSecretsManagerSecretVersion_Basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_secretsmanager_secret_version.test"
	datasourceName := "data.aws_secretsmanager_secret_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsSecretsManagerSecretVersionConfig_NonExistent,
				ExpectError: regexp.MustCompile(`not found`),
			},
			{
				Config: testAccDataSourceAwsSecretsManagerSecretVersionConfig_VersionStage_Default(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsSecretsManagerSecretVersionCheck(datasourceName, resourceName),
				),
			},
		},
	})
}

func TestAccDataSourceAwsSecretsManagerSecretVersion_VersionID(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_secretsmanager_secret_version.test"
	datasourceName := "data.aws_secretsmanager_secret_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsSecretsManagerSecretVersionConfig_VersionID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsSecretsManagerSecretVersionCheck(datasourceName, resourceName),
				),
			},
		},
	})
}

func TestAccDataSourceAwsSecretsManagerSecretVersion_VersionStage(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_secretsmanager_secret_version.test"
	datasourceName := "data.aws_secretsmanager_secret_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsSecretsManagerSecretVersionConfig_VersionStage_Custom(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsSecretsManagerSecretVersionCheck(datasourceName, resourceName),
				),
			},
		},
	})
}

func testAccDataSourceAwsSecretsManagerSecretVersionCheck(datasourceName, resourceName string) resource.TestCheckFunc {
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
			"secret_value",
			"version_stages.#",
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

const testAccDataSourceAwsSecretsManagerSecretVersionConfig_NonExistent = `
data "aws_secretsmanager_secret_version" "test" {
  secret_id = "tf-acc-test-does-not-exist"
}
`

func testAccDataSourceAwsSecretsManagerSecretVersionConfig_VersionID(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = "%[1]s"
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = "${aws_secretsmanager_secret.test.id}"
  secret_string = "test-string"
}

data "aws_secretsmanager_secret_version" "test" {
  secret_id  = "${aws_secretsmanager_secret.test.id}"
  version_id = "${aws_secretsmanager_secret_version.test.version_id}"
}
`, rName)
}

func testAccDataSourceAwsSecretsManagerSecretVersionConfig_VersionStage_Custom(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = "%[1]s"
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id      = "${aws_secretsmanager_secret.test.id}"
  secret_string  = "test-string"
  version_stages = ["test-stage", "AWSCURRENT"]
}

data "aws_secretsmanager_secret_version" "test" {
  secret_id     = "${aws_secretsmanager_secret_version.test.secret_id}"
  version_stage = "test-stage"
}
`, rName)
}

func testAccDataSourceAwsSecretsManagerSecretVersionConfig_VersionStage_Default(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = "%[1]s"
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = "${aws_secretsmanager_secret.test.id}"
  secret_string = "test-string"
}

data "aws_secretsmanager_secret_version" "test" {
  secret_id = "${aws_secretsmanager_secret_version.test.secret_id}"
}
`, rName)
}
