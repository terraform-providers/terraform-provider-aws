package aws

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceAwsKmsCiphertext_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAwsKmsCiphertextConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"aws_kms_ciphertext.foo", "ciphertext_blob"),
				),
			},
		},
	})
}

func TestAccResourceAwsKmsCiphertext_validate(t *testing.T) {
	kmsSecretsDataSource := "data.aws_kms_secrets.foo"
	resourceName := "aws_kms_ciphertext.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAwsKmsCiphertextConfig_validate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "ciphertext_blob"),
					resource.TestCheckResourceAttrPair(resourceName, "plaintext", kmsSecretsDataSource, "plaintext.plaintext"),
				),
			},
		},
	})
}

func TestAccResourceAwsKmsCiphertext_validate_withContext(t *testing.T) {
	kmsSecretsDataSource := "data.aws_kms_secrets.foo"
	resourceName := "aws_kms_ciphertext.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAwsKmsCiphertextConfig_validate_withContext,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "ciphertext_blob"),
					resource.TestCheckResourceAttrPair(resourceName, "plaintext", kmsSecretsDataSource, "plaintext.plaintext"),
				),
			},
		},
	})
}

const testAccResourceAwsKmsCiphertextConfig_basic = `
resource "aws_kms_key" "foo" {
  description = "tf-test-acc-data-source-aws-kms-ciphertext-basic"
  is_enabled  = true
}

resource "aws_kms_ciphertext" "foo" {
  key_id = aws_kms_key.foo.key_id

  plaintext = "Super secret data"
}
`

const testAccResourceAwsKmsCiphertextConfig_validate = `
resource "aws_kms_key" "foo" {
  description = "tf-test-acc-data-source-aws-kms-ciphertext-validate"
  is_enabled  = true
}

resource "aws_kms_ciphertext" "foo" {
  key_id = aws_kms_key.foo.key_id

  plaintext = "Super secret data"
}

data "aws_kms_secrets" "foo" {
  secret {
    name    = "plaintext"
    payload = aws_kms_ciphertext.foo.ciphertext_blob
  }
}
`

const testAccResourceAwsKmsCiphertextConfig_validate_withContext = `
resource "aws_kms_key" "foo" {
  description = "tf-test-acc-data-source-aws-kms-ciphertext-validate-with-context"
  is_enabled  = true
}

resource "aws_kms_ciphertext" "foo" {
  key_id = aws_kms_key.foo.key_id

  plaintext = "Super secret data"

  context = {
    name = "value"
  }
}

data "aws_kms_secrets" "foo" {
  secret {
    name    = "plaintext"
    payload = aws_kms_ciphertext.foo.ciphertext_blob

    context = {
      name = "value"
    }
  }
}
`
