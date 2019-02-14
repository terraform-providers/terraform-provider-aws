package aws

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsLbListenerCertificate_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAwsLbListenerCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLbListenerCertificateConfig(acctest.RandString(5), acctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLbListenerCertificateExists("aws_lb_listener_certificate.default"),
					testAccCheckAwsLbListenerCertificateExists("aws_lb_listener_certificate.additional_1"),
					testAccCheckAwsLbListenerCertificateExists("aws_lb_listener_certificate.additional_2"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.default", "listener_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.default", "certificate_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.additional_1", "listener_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.additional_1", "certificate_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.additional_2", "listener_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.additional_2", "certificate_arn"),
				),
			},
		},
	})
}

func TestAccAwsLbListenerCertificate_cycle(t *testing.T) {
	rName := acctest.RandString(5)
	suffix := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAwsLbListenerCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLbListenerCertificateConfig(rName, suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLbListenerCertificateExists("aws_lb_listener_certificate.default"),
					testAccCheckAwsLbListenerCertificateExists("aws_lb_listener_certificate.additional_1"),
					testAccCheckAwsLbListenerCertificateExists("aws_lb_listener_certificate.additional_2"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.default", "listener_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.default", "certificate_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.additional_1", "listener_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.additional_1", "certificate_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.additional_2", "listener_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.additional_2", "certificate_arn"),
				),
			},
			{
				Config: testAccLbListenerCertificateAddNew(rName, suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLbListenerCertificateExists("aws_lb_listener_certificate.default"),
					testAccCheckAwsLbListenerCertificateExists("aws_lb_listener_certificate.additional_1"),
					testAccCheckAwsLbListenerCertificateExists("aws_lb_listener_certificate.additional_2"),
					testAccCheckAwsLbListenerCertificateExists("aws_lb_listener_certificate.additional_3"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.default", "listener_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.default", "certificate_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.additional_1", "listener_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.additional_1", "certificate_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.additional_2", "listener_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.additional_2", "certificate_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.additional_3", "listener_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.additional_3", "certificate_arn"),
				),
			},
			{
				Config: testAccLbListenerCertificateConfig(rName, suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLbListenerCertificateExists("aws_lb_listener_certificate.default"),
					testAccCheckAwsLbListenerCertificateExists("aws_lb_listener_certificate.additional_1"),
					testAccCheckAwsLbListenerCertificateExists("aws_lb_listener_certificate.additional_2"),
					testAccCheckAwsLbListenerCertificateNotExists("aws_lb_listener_certificate.additional_3"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.default", "listener_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.default", "certificate_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.additional_1", "listener_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.additional_1", "certificate_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.additional_2", "listener_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.additional_2", "certificate_arn"),
				),
			},
		},
	})
}

func testAccCheckAwsLbListenerCertificateDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).elbv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lb_listener_certificate" {
			continue
		}

		input := &elbv2.DescribeListenerCertificatesInput{
			ListenerArn: aws.String(rs.Primary.Attributes["listener_arn"]),
			PageSize:    aws.Int64(400),
		}

		resp, err := conn.DescribeListenerCertificates(input)
		if err != nil {
			if isAWSErr(err, elbv2.ErrCodeListenerNotFoundException, "") {
				return nil
			}
			return err
		}

		for _, cert := range resp.Certificates {
			// We only care about additional certificates.
			if aws.BoolValue(cert.IsDefault) {
				continue
			}

			if aws.StringValue(cert.CertificateArn) == rs.Primary.Attributes["certificate_arn"] {
				return errors.New("LB listener certificate not destroyed")
			}
		}
	}

	return nil
}

func testAccCheckAwsLbListenerCertificateExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccCheckAwsLbListenerCertificateNotExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return nil
		}

		return fmt.Errorf("Not expecting but found: %s", name)
	}
}

func testAccLbListenerCertificateConfig(rName, suffix string) string {
	return fmt.Sprintf(`
resource "tls_private_key" "test" {
  count = 4

  algorithm = "RSA"
}

resource "tls_self_signed_cert" "test" {
  count = 4

  key_algorithm         = "RSA"
  private_key_pem       = "${tls_private_key.test.*.private_key_pem[count.index]}"
  validity_period_hours = 12

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]

  subject {
    common_name  = "example.com"
    organization = "ACME Examples, Inc"
  }
}

resource "aws_lb_listener_certificate" "default" {
  listener_arn    = "${aws_lb_listener.test.arn}"
  certificate_arn = "${aws_iam_server_certificate.default.arn}"
}

resource "aws_lb_listener_certificate" "additional_1" {
  listener_arn    = "${aws_lb_listener.test.arn}"
  certificate_arn = "${aws_iam_server_certificate.additional_1.arn}"
}

resource "aws_lb_listener_certificate" "additional_2" {
  listener_arn    = "${aws_lb_listener.test.arn}"
  certificate_arn = "${aws_iam_server_certificate.additional_2.arn}"
}

resource "aws_lb" "test" {
  name_prefix    = "%s"
  subnets = ["${aws_subnet.test.*.id[0]}", "${aws_subnet.test.*.id[1]}"]
  internal = true
}

resource "aws_lb_target_group" "test" {
  port     = "443"
  protocol = "HTTP"
  vpc_id            = "${aws_vpc.test.id}"
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = "${aws_lb.test.arn}"
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = "${aws_iam_server_certificate.default.arn}"

  default_action {
    target_group_arn = "${aws_lb_target_group.test.arn}"
    type             = "forward"
  }
}

resource "aws_iam_server_certificate" "default" {
  name             = "terraform-default-cert-%s"
  certificate_body = "${tls_self_signed_cert.test.*.cert_pem[0]}"
  private_key      = "${tls_private_key.test.*.private_key_pem[0]}"
}

resource "aws_iam_server_certificate" "additional_1" {
  name             = "terraform-additional-cert-1-%s"
  certificate_body = "${tls_self_signed_cert.test.*.cert_pem[1]}"
  private_key      = "${tls_private_key.test.*.private_key_pem[1]}"
}

resource "aws_iam_server_certificate" "additional_2" {
  name             = "terraform-additional-cert-2-%s"
  certificate_body = "${tls_self_signed_cert.test.*.cert_pem[2]}"
  private_key      = "${tls_private_key.test.*.private_key_pem[2]}"
}

resource "aws_iam_server_certificate" "additional_3" {
  name             = "terraform-additional-cert-3-%s"
  certificate_body = "${tls_self_signed_cert.test.*.cert_pem[3]}"
  private_key      = "${tls_private_key.test.*.private_key_pem[3]}"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags = {
  	Name = "terraform-testacc-lb-listener-certificate"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
}

resource "aws_subnet" "test" {
  count             = "${length(var.subnets)}"
  vpc_id            = "${aws_vpc.test.id}"
  cidr_block        = "${element(var.subnets, count.index)}"
  availability_zone = "${element(data.aws_availability_zones.available.names, count.index)}"
  tags = {
    Name = "tf-acc-lb-listener-certificate-${count.index}"
  }
}`, rName, suffix, suffix, suffix, suffix)
}

func testAccLbListenerCertificateAddNew(rName, prefix string) string {
	return fmt.Sprintf(testAccLbListenerCertificateConfig(rName, prefix) + `
resource "aws_lb_listener_certificate" "additional_3" {
  listener_arn    = "${aws_lb_listener.test.arn}"
  certificate_arn = "${aws_iam_server_certificate.additional_3.arn}"
}
`)
}
