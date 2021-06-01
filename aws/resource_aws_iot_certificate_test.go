package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSIoTCertificate_csr(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTCertificateDestroy_basic,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTCertificate_csr,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("aws_iot_certificate.foo_cert", "arn"),
					resource.TestCheckResourceAttrSet("aws_iot_certificate.foo_cert", "csr"),
					resource.TestCheckResourceAttrSet("aws_iot_certificate.foo_cert", "certificate_pem"),
					resource.TestCheckNoResourceAttr("aws_iot_certificate.foo_cert", "public_key"),
					resource.TestCheckNoResourceAttr("aws_iot_certificate.foo_cert", "private_key"),
					resource.TestCheckResourceAttr("aws_iot_certificate.foo_cert", "active", "true"),
				),
			},
		},
	})
}

func TestAccAWSIoTCertificate_keys_certificate(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTCertificateDestroy_basic,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTCertificate_keys_certificate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("aws_iot_certificate.foo_cert", "arn"),
					resource.TestCheckNoResourceAttr("aws_iot_certificate.foo_cert", "csr"),
					resource.TestCheckResourceAttrSet("aws_iot_certificate.foo_cert", "certificate_pem"),
					resource.TestCheckResourceAttrSet("aws_iot_certificate.foo_cert", "public_key"),
					resource.TestCheckResourceAttrSet("aws_iot_certificate.foo_cert", "private_key"),
					resource.TestCheckResourceAttr("aws_iot_certificate.foo_cert", "active", "true"),
				),
			},
		},
	})
}

func testAccCheckAWSIoTCertificateDestroy_basic(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iotconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_certificate" {
			continue
		}

		// Try to find the Cert
		DescribeCertOpts := &iot.DescribeCertificateInput{
			CertificateId: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeCertificate(DescribeCertOpts)

		if err == nil {
			if resp.CertificateDescription != nil {
				return fmt.Errorf("Device Certificate still exists")
			}
		}

		// Verify the error is what we want
		if err != nil {
			iotErr, ok := err.(awserr.Error)
			if !ok || iotErr.Code() != "ResourceNotFoundException" {
				return err
			}
		}

	}

	return nil
}

var testAccAWSIoTCertificate_csr = `
resource "aws_iot_certificate" "foo_cert" {
  csr    = file("test-fixtures/iot-csr.pem")
  active = true
}
`

var testAccAWSIoTCertificate_keys_certificate = `
resource "aws_iot_certificate" "foo_cert" {
  active = true
}
`
