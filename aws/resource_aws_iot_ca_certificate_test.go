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

func TestAccAWSIoTCACertificate_certificate(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTCACertificateDestroy_basic,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTCACertificate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("aws_iot_ca_certificate.foo", "arn"),
					resource.TestCheckResourceAttrSet("aws_iot_ca_certificate.foo", "id"),
					resource.TestCheckResourceAttr("aws_iot_ca_certificate.foo", "active", "true"),
					resource.TestCheckResourceAttr("aws_iot_ca_certificate.foo", "auto_registration", "true"),
				),
			},
		},
	})
}

func testAccCheckAWSIoTCACertificateDestroy_basic(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iotconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_ca_certificate" {
			continue
		}

		DescribeCACertOpts := &iot.DescribeCACertificateInput{
			CertificateId: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeCACertificate(DescribeCACertOpts)

		if err == nil {
			if resp.CertificateDescription != nil {
				return fmt.Errorf("CA Certificate still exists")
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

var testAccAWSIoTCACertificate = `
resource "aws_iot_ca_certificate" "foo" {
  ca_certificate    	   = "test"
  verification_certificate = "test"
  active 			       = true
  auto_registration        = true
}
`
