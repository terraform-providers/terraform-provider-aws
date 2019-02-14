package aws

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSKmsGrant_Basic(t *testing.T) {
	timestamp := time.Now().Format(time.RFC1123)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsGrantDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsGrant_Basic("basic", timestamp, "\"Encrypt\", \"Decrypt\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsGrantExists("aws_kms_grant.basic"),
					resource.TestCheckResourceAttr("aws_kms_grant.basic", "name", "basic"),
					resource.TestCheckResourceAttr("aws_kms_grant.basic", "operations.#", "2"),
					resource.TestCheckResourceAttr("aws_kms_grant.basic", "operations.2238845196", "Encrypt"),
					resource.TestCheckResourceAttr("aws_kms_grant.basic", "operations.1237510779", "Decrypt"),
					resource.TestCheckResourceAttrSet("aws_kms_grant.basic", "grantee_principal"),
					resource.TestCheckResourceAttrSet("aws_kms_grant.basic", "key_id"),
				),
			},
		},
	})
}

func TestAccAWSKmsGrant_withConstraints(t *testing.T) {
	timestamp := time.Now().Format(time.RFC1123)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsGrantDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsGrant_withConstraints("withConstraintsEq", timestamp, "encryption_context_equals", `foo = "bar"
                        baz = "kaz"`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsGrantExists("aws_kms_grant.withConstraintsEq"),
					resource.TestCheckResourceAttr("aws_kms_grant.withConstraintsEq", "name", "withConstraintsEq"),
					resource.TestCheckResourceAttr("aws_kms_grant.withConstraintsEq", "constraints.#", "1"),
					resource.TestCheckResourceAttr("aws_kms_grant.withConstraintsEq", "constraints.449762259.encryption_context_equals.%", "2"),
					resource.TestCheckResourceAttr("aws_kms_grant.withConstraintsEq", "constraints.449762259.encryption_context_equals.baz", "kaz"),
					resource.TestCheckResourceAttr("aws_kms_grant.withConstraintsEq", "constraints.449762259.encryption_context_equals.foo", "bar"),
				),
			},
			{
				Config: testAccAWSKmsGrant_withConstraints("withConstraintsSub", timestamp, "encryption_context_subset", `foo = "bar"
			            baz = "kaz"`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsGrantExists("aws_kms_grant.withConstraintsSub"),
					resource.TestCheckResourceAttr("aws_kms_grant.withConstraintsSub", "name", "withConstraintsSub"),
					resource.TestCheckResourceAttr("aws_kms_grant.withConstraintsSub", "constraints.#", "1"),
					resource.TestCheckResourceAttr("aws_kms_grant.withConstraintsSub", "constraints.2645649985.encryption_context_subset.%", "2"),
					resource.TestCheckResourceAttr("aws_kms_grant.withConstraintsSub", "constraints.2645649985.encryption_context_subset.baz", "kaz"),
					resource.TestCheckResourceAttr("aws_kms_grant.withConstraintsSub", "constraints.2645649985.encryption_context_subset.foo", "bar"),
				),
			},
		},
	})
}

func TestAccAWSKmsGrant_withRetiringPrincipal(t *testing.T) {
	timestamp := time.Now().Format(time.RFC1123)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsGrantDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsGrant_withRetiringPrincipal("withRetiringPrincipal", timestamp),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsGrantExists("aws_kms_grant.withRetiringPrincipal"),
					resource.TestCheckResourceAttrSet("aws_kms_grant.withRetiringPrincipal", "retiring_principal"),
				),
			},
		},
	})
}

func TestAccAWSKmsGrant_bare(t *testing.T) {
	timestamp := time.Now().Format(time.RFC1123)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsGrantDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsGrant_bare("bare", timestamp),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsGrantExists("aws_kms_grant.bare"),
					resource.TestCheckNoResourceAttr("aws_kms_grant.bare", "name"),
					resource.TestCheckNoResourceAttr("aws_kms_grant.bare", "constraints.#"),
					resource.TestCheckNoResourceAttr("aws_kms_grant.bare", "retiring_principal"),
				),
			},
		},
	})
}

func TestAccAWSKmsGrant_ARN(t *testing.T) {
	timestamp := time.Now().Format(time.RFC1123)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsGrantDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsGrant_ARN("arn", timestamp, "\"Encrypt\", \"Decrypt\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsGrantExists("aws_kms_grant.arn"),
					resource.TestCheckResourceAttr("aws_kms_grant.arn", "name", "arn"),
					resource.TestCheckResourceAttr("aws_kms_grant.arn", "operations.#", "2"),
					resource.TestCheckResourceAttr("aws_kms_grant.arn", "operations.2238845196", "Encrypt"),
					resource.TestCheckResourceAttr("aws_kms_grant.arn", "operations.1237510779", "Decrypt"),
					resource.TestCheckResourceAttrSet("aws_kms_grant.arn", "grantee_principal"),
					resource.TestCheckResourceAttrSet("aws_kms_grant.arn", "key_id"),
				),
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

func testAccAWSKmsGrant_Basic(rName string, timestamp string, operations string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "tf-acc-test-key" {
    description = "Terraform acc test key %s"
    deletion_window_in_days = 7
}

%s

resource "aws_iam_role" "tf-acc-test-role" {
  name               = "tf-acc-test-kms-grant-role-%s"
  path               = "/service-role/"
  assume_role_policy = "${data.aws_iam_policy_document.assumerole-policy-template.json}"
}

resource "aws_kms_grant" "%s" {
	name = "%s"
	key_id = "${aws_kms_key.tf-acc-test-key.key_id}"
	grantee_principal = "${aws_iam_role.tf-acc-test-role.arn}"
	operations = [ %s ]
}
`, timestamp, staticAssumeRolePolicyString, rName, rName, rName, operations)
}

func testAccAWSKmsGrant_withConstraints(rName string, timestamp string, constraintName string, encryptionContext string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "tf-acc-test-key" {
    description = "Terraform acc test key %s"
    deletion_window_in_days = 7
}

%s

resource "aws_iam_role" "tf-acc-test-role" {
  name               = "tf-acc-test-kms-grant-role-%s"
  path               = "/service-role/"
  assume_role_policy = "${data.aws_iam_policy_document.assumerole-policy-template.json}"
}

resource "aws_kms_grant" "%s" {
	name = "%s"
	key_id = "${aws_kms_key.tf-acc-test-key.key_id}"
	grantee_principal = "${aws_iam_role.tf-acc-test-role.arn}"
	operations = [ "RetireGrant", "DescribeKey" ]
	constraints {
		%s = {
			%s
		}
	}
}
`, timestamp, staticAssumeRolePolicyString, rName, rName, rName, constraintName, encryptionContext)
}

func testAccAWSKmsGrant_withRetiringPrincipal(rName string, timestamp string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "tf-acc-test-key" {
    description = "Terraform acc test key %s"
    deletion_window_in_days = 7
}

%s

resource "aws_iam_role" "tf-acc-test-role" {
  name               = "tf-acc-test-kms-grant-role-%s"
  path               = "/service-role/"
  assume_role_policy = "${data.aws_iam_policy_document.assumerole-policy-template.json}"
}

resource "aws_kms_grant" "%s" {
	name = "%s"
	key_id = "${aws_kms_key.tf-acc-test-key.key_id}"
	grantee_principal = "${aws_iam_role.tf-acc-test-role.arn}"
	operations = [ "ReEncryptTo", "CreateGrant" ]
	retiring_principal = "${aws_iam_role.tf-acc-test-role.arn}"
}
`, timestamp, staticAssumeRolePolicyString, rName, rName, rName)
}

func testAccAWSKmsGrant_bare(rName string, timestamp string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "tf-acc-test-key" {
    description = "Terraform acc test key %s"
    deletion_window_in_days = 7
}

%s

resource "aws_iam_role" "tf-acc-test-role" {
  name               = "tf-acc-test-kms-grant-role-%s"
  path               = "/service-role/"
  assume_role_policy = "${data.aws_iam_policy_document.assumerole-policy-template.json}"
}

resource "aws_kms_grant" "%s" {
	key_id = "${aws_kms_key.tf-acc-test-key.key_id}"
	grantee_principal = "${aws_iam_role.tf-acc-test-role.arn}"
	operations = [ "ReEncryptTo", "CreateGrant" ]
}
`, timestamp, staticAssumeRolePolicyString, rName, rName)
}

const staticAssumeRolePolicyString = `
data "aws_iam_policy_document" "assumerole-policy-template" {
  statement {
    effect  = "Allow"
    actions = [ "sts:AssumeRole" ]
    principals {
      type        = "Service"
      identifiers = [ "ec2.amazonaws.com" ]
    }
  }
}
`

func testAccAWSKmsGrant_ARN(rName string, timestamp string, operations string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "tf-acc-test-key" {
    description = "Terraform acc test key %s"
    deletion_window_in_days = 7
}

%s

resource "aws_iam_role" "tf-acc-test-role" {
  name               = "tf-acc-test-kms-grant-role-%s"
  path               = "/service-role/"
  assume_role_policy = "${data.aws_iam_policy_document.assumerole-policy-template.json}"
}

resource "aws_kms_grant" "%s" {
	name = "%s"
	key_id = "${aws_kms_key.tf-acc-test-key.arn}"
	grantee_principal = "${aws_iam_role.tf-acc-test-role.arn}"
	operations = [ %s ]
}
`, timestamp, staticAssumeRolePolicyString, rName, rName, rName, operations)
}
