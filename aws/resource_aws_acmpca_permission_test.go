package aws

import (
	"fmt"
	"regexp"
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
				Config: testAccAwsAcmpcaPermissionConfig_Valid(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAcmpcaPermissionExists(resourceName, &permission),
					resource.TestCheckResourceAttr(resourceName, "principal", "acm.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName, "actions.#", "3"),
				),
			},
		},
	})
}

func TestAccAwsAcmpcaPermission_InvalidPrincipal(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAcmpcaPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAwsAcmpcaPermissionConfig_InvalidPrincipal(),
				ExpectError: regexp.MustCompile("config is invalid: expected principal to be one of .*, got .*"),
			},
		},
	})
}

func TestAccAwsAcmpcaPermission_InvalidActionsEntry(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAcmpcaPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAwsAcmpcaPermissionConfig_InvalidActionsEntry(),
				ExpectError: regexp.MustCompile("config is invalid: expected actions.1 to be one of .*, got .*"),
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
			if isAWSErr(err, acmpca.ErrCodeInvalidStateException, "") {
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

func testAccAwsAcmpcaCertificateAuthority() string {
	return `
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "terraformtesting.com"
    }
  }
}
`
}

func testAccAwsAcmpcaPermissionConfig_Valid() string {
	return fmt.Sprintf(`
%s

resource "aws_acmpca_permission" "test" {
	certificate_authority_arn = "${aws_acmpca_certificate_authority.test.arn}"
	principal                 = "acm.amazonaws.com"
	actions                   = ["IssueCertificate", "GetCertificate", "ListPermissions"]
}
`, testAccAwsAcmpcaCertificateAuthority())
}

func testAccAwsAcmpcaPermissionConfig_InvalidPrincipal() string {
	return fmt.Sprintf(`
%s

resource "aws_acmpca_permission" "test" {
	certificate_authority_arn = "${aws_acmpca_certificate_authority.test.arn}"
	principal                 = "notacm.amazonaws.com"
	actions                   = ["IssueCertificate", "GetCertificate", "ListPermissions"]
}
`, testAccAwsAcmpcaCertificateAuthority())
}

func testAccAwsAcmpcaPermissionConfig_InvalidActionsEntry() string {
	return fmt.Sprintf(`
%s

resource "aws_acmpca_permission" "test" {
	certificate_authority_arn = "${aws_acmpca_certificate_authority.test.arn}"
	principal                 = "acm.amazonaws.com"
	actions                   = ["IssueCert", "GetCertificate", "ListPermissions"]
}
`, testAccAwsAcmpcaCertificateAuthority())
}
