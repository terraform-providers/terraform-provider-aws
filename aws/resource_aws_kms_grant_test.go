package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfawsresource"
)

func TestAccAWSKmsGrant_basic(t *testing.T) {
	resourceName := "aws_kms_grant.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsGrantDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsGrant_Basic(rName, "\"Encrypt\", \"Decrypt\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsGrantExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "operations.#", "2"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "operations.*", "Encrypt"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "operations.*", "Decrypt"),
					resource.TestCheckResourceAttrSet(resourceName, "grantee_principal"),
					resource.TestCheckResourceAttrSet(resourceName, "key_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"grant_token", "retire_on_delete"},
			},
		},
	})
}

func TestAccAWSKmsGrant_withConstraints(t *testing.T) {
	resourceName := "aws_kms_grant.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsGrantDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsGrant_withConstraints(rName, "encryption_context_equals", `foo = "bar"
                        baz = "kaz"`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsGrantExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "constraints.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "constraints.*", map[string]string{
						"encryption_context_equals.%":   "2",
						"encryption_context_equals.baz": "kaz",
						"encryption_context_equals.foo": "bar",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"grant_token", "retire_on_delete"},
			},
			{
				Config: testAccAWSKmsGrant_withConstraints(rName, "encryption_context_subset", `foo = "bar"
			            baz = "kaz"`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsGrantExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "constraints.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "constraints.*", map[string]string{
						"encryption_context_subset.%":   "2",
						"encryption_context_subset.baz": "kaz",
						"encryption_context_subset.foo": "bar",
					}),
				),
			},
		},
	})
}

func TestAccAWSKmsGrant_withRetiringPrincipal(t *testing.T) {
	resourceName := "aws_kms_grant.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsGrantDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsGrant_withRetiringPrincipal(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsGrantExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "retiring_principal"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"grant_token", "retire_on_delete"},
			},
		},
	})
}

func TestAccAWSKmsGrant_bare(t *testing.T) {
	resourceName := "aws_kms_grant.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsGrantDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsGrant_bare(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsGrantExists(resourceName),
					resource.TestCheckNoResourceAttr(resourceName, "name"),
					resource.TestCheckNoResourceAttr(resourceName, "constraints.#"),
					resource.TestCheckNoResourceAttr(resourceName, "retiring_principal"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"grant_token", "retire_on_delete"},
			},
		},
	})
}

func TestAccAWSKmsGrant_ARN(t *testing.T) {
	resourceName := "aws_kms_grant.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsGrantDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsGrant_ARN(rName, "\"Encrypt\", \"Decrypt\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsGrantExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "operations.#", "2"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "operations.*", "Encrypt"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "operations.*", "Decrypt"),
					resource.TestCheckResourceAttrSet(resourceName, "grantee_principal"),
					resource.TestCheckResourceAttrSet(resourceName, "key_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"grant_token", "retire_on_delete"},
			},
		},
	})
}

func TestAccAWSKmsGrant_disappears(t *testing.T) {
	resourceName := "aws_kms_grant.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsGrantDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsGrant_Basic(rName, "\"Encrypt\", \"Decrypt\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsGrantExists(resourceName),
					testAccCheckAWSKmsGrantDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSKmsGrantDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).kmsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_kms_grant" {
			continue
		}

		err := waitForKmsGrantToBeRevoked(conn, rs.Primary.Attributes["key_id"], rs.Primary.ID)
		return err
	}

	return nil
}

func testAccCheckAWSKmsGrantExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		return nil
	}
}

func testAccCheckAWSKmsGrantDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).kmsconn

		revokeInput := kms.RevokeGrantInput{
			GrantId: aws.String(rs.Primary.Attributes["grant_id"]),
			KeyId:   aws.String(rs.Primary.Attributes["key_id"]),
		}

		_, err := conn.RevokeGrant(&revokeInput)
		return err
	}
}

func testAccAWSKmsGrantConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "Terraform acc test key %[1]s"
  deletion_window_in_days = 7
}

data "aws_iam_policy_document" "test" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/service-role/"
  assume_role_policy = data.aws_iam_policy_document.test.json
}
`, rName)
}

func testAccAWSKmsGrant_Basic(rName string, operations string) string {
	return testAccAWSKmsGrantConfigBase(rName) + fmt.Sprintf(`
resource "aws_kms_grant" "test" {
  name              = %[1]q
  key_id            = aws_kms_key.test.key_id
  grantee_principal = aws_iam_role.test.arn
  operations        = [%[2]s]
}
`, rName, operations)
}

func testAccAWSKmsGrant_withConstraints(rName string, constraintName string, encryptionContext string) string {
	return testAccAWSKmsGrantConfigBase(rName) + fmt.Sprintf(`
resource "aws_kms_grant" "test" {
  name              = "%[1]s"
  key_id            = aws_kms_key.test.key_id
  grantee_principal = aws_iam_role.test.arn
  operations        = ["RetireGrant", "DescribeKey"]

  constraints {
    %[2]s = {
      %[3]s
    }
  }
}
`, rName, constraintName, encryptionContext)
}

func testAccAWSKmsGrant_withRetiringPrincipal(rName string) string {
	return testAccAWSKmsGrantConfigBase(rName) + fmt.Sprintf(`
resource "aws_kms_grant" "test" {
  name               = "%[1]s"
  key_id             = aws_kms_key.test.key_id
  grantee_principal  = aws_iam_role.test.arn
  operations         = ["ReEncryptTo", "CreateGrant"]
  retiring_principal = aws_iam_role.test.arn
}
`, rName)
}

func testAccAWSKmsGrant_bare(rName string) string {
	return testAccAWSKmsGrantConfigBase(rName) + `
resource "aws_kms_grant" "test" {
  key_id            = aws_kms_key.test.key_id
  grantee_principal = aws_iam_role.test.arn
  operations        = ["ReEncryptTo", "CreateGrant"]
}
`
}

func testAccAWSKmsGrant_ARN(rName string, operations string) string {
	return testAccAWSKmsGrantConfigBase(rName) + fmt.Sprintf(`
resource "aws_kms_grant" "test" {
  name              = "%[1]s"
  key_id            = aws_kms_key.test.arn
  grantee_principal = aws_iam_role.test.arn
  operations        = [%[2]s]
}
`, rName, operations)
}
