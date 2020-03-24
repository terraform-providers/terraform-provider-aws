package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAwsAcmpcaPermission_Valid(t *testing.T) {
	var permission acmpca.Permission
	resourceName := "aws_acmpca_permission.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAcmpcaPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAcmpcaPermissionConfig_Valid,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAcmpcaPermissionExists(resourceName, &permission),
					resource.TestCheckResourceAttr(resourceName, "principal", "acm.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName, "actions.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "actions.0", "IssueCertificate"),
					resource.TestCheckResourceAttr(resourceName, "actions.1", "GetCertificate"),
					resource.TestCheckResourceAttr(resourceName, "actions.2", "ListPermissions"),
				),
			},
			{
				ResourceName: resourceName,
			},
		},
	})
}

func TestAccAwsAcmpcaPermission_InvalidPrincipal(t *testing.T) {
	var permission acmpca.Permission
	resourceName := "aws_acmpca_permission.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAcmpcaPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAcmpcaPermissionConfig_InvalidPrincipal,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAcmpcaPermissionExists(resourceName, &permission),
					resource.TestCheckResourceAttr(resourceName, "principal", "acm.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName, "actions.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "actions.0", "IssueCertificate"),
					resource.TestCheckResourceAttr(resourceName, "actions.1", "GetCertificate"),
					resource.TestCheckResourceAttr(resourceName, "actions.2", "ListPermissions"),
				),
			},
			{
				ResourceName: resourceName,
			},
		},
	})
}

func TestAccAwsAcmpcaPermission_InvalidActionsCount(t *testing.T) {
	var permission acmpca.Permission
	resourceName := "aws_acmpca_permission.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAcmpcaPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAcmpcaPermissionConfig_InvalidActionsCount,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAcmpcaPermissionExists(resourceName, &permission),
					resource.TestCheckResourceAttr(resourceName, "principal", "acm.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName, "actions.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "actions.0", "IssueCertificate"),
					resource.TestCheckResourceAttr(resourceName, "actions.1", "GetCertificate"),
					resource.TestCheckResourceAttr(resourceName, "actions.2", "ListPermissions"),
				),
			},
			{
				ResourceName: resourceName,
			},
		},
	})
}

func TestAccAwsAcmpcaPermission_InvalidActionsEntry(t *testing.T) {
	var permission acmpca.Permission
	resourceName := "aws_acmpca_permission.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAcmpcaPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAcmpcaPermissionConfig_InvalidActionsEntry,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAcmpcaPermissionExists(resourceName, &permission),
					resource.TestCheckResourceAttr(resourceName, "principal", "acm.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName, "actions.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "actions.0", "IssueCertificate"),
					resource.TestCheckResourceAttr(resourceName, "actions.1", "GetCertificate"),
					resource.TestCheckResourceAttr(resourceName, "actions.2", "ListPermissions"),
				),
			},
			{
				ResourceName: resourceName,
			},
		},
	})
}

func testAccCheckAwsAcmpcaPermissionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).acmpcaconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_acmpca_permission" {
			continue
		}

		input := &acmpca.ListPermissionsInput{
			CertificateAuthorityArn: aws.String(rs.Primary.Attributes["certificate_authority_arn"]),
		}

		output, err := conn.ListPermissions(input)

		if err != nil {
			if isAWSErr(err, acmpca.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return err
		}

		if output != nil {
			return fmt.Errorf("ACMPCA Permission %q still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccCheckAwsAcmpcaPermissionExists(resourceName string, permission *acmpca.Permission) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).acmpcaconn
		input := &acmpca.ListPermissionsInput{
			CertificateAuthorityArn: aws.String(rs.Primary.Attributes["certificate_authority_arn"]),
		}

		output, err := conn.ListPermissions(input)

		if err != nil {
			return err
		}

		if output == nil || output.Permissions == nil {
			return fmt.Errorf("ACMPCA Permission %q does not exist", rs.Primary.ID)
		}

		*permission = *output.Permissions[0]

		return nil
	}
}

const testAccAwsAcmpcaPermissionConfig_Valid = `
resource "aws_acmpca_certificate_authority" "test" {
  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "terraformtesting.com"
    }
  }
}

resource "aws_acmpca_permission" "test" {
	certificate_authority_arn = "${aws_acmpca_certificate_authority.test.arn}"
	principal                 = "acm.amazonaws.com"
	actions                   = ["IssueCertificate", "GetCertificate", "ListPermissions"]
}
`
const testAccAwsAcmpcaPermissionConfig_InvalidPrincipal = `
resource "aws_acmpca_certificate_authority" "test" {
  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "terraformtesting.com"
    }
  }
}

resource "aws_acmpca_permission" "test" {
	certificate_authority_arn = "${aws_acmpca_certificate_authority.test.arn}"
	principal                 = "notacm.amazonaws.com"
	actions                   = ["IssueCertificate", "GetCertificate", "ListPermissions"]
}
`

const testAccAwsAcmpcaPermissionConfig_InvalidActionsCount = `
resource "aws_acmpca_certificate_authority" "test" {
  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "terraformtesting.com"
    }
  }
}

resource "aws_acmpca_permission" "test" {
	certificate_authority_arn = "${aws_acmpca_certificate_authority.test.arn}"
	principal                 = "acm.amazonaws.com"
	actions                   = ["GetCertificate", "ListPermissions"]
}
`

const testAccAwsAcmpcaPermissionConfig_InvalidActionsEntry = `
resource "aws_acmpca_certificate_authority" "test" {
  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "terraformtesting.com"
    }
  }
}

resource "aws_acmpca_permission" "test" {
	certificate_authority_arn = "${aws_acmpca_certificate_authority.test.arn}"
	principal                 = "acm.amazonaws.com"
	actions                   = ["IssueCert", "GetCertificate", "ListPermissions"]
}
`
