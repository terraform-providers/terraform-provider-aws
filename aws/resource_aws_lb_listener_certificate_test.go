package aws

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsLbListenerCertificate_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
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

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
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
			if wserr, ok := err.(awserr.Error); ok && wserr.Code() == "ListenerNotFound" {
				return nil
			}
			return err
		}

		for _, cert := range resp.Certificates {
			// We only care about additional certificates.
			if *cert.IsDefault {
				continue
			}

			if *cert.CertificateArn == rs.Primary.Attributes["certificate_arn"] {
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
  subnets = ["${aws_subnet.test.*.id}"]
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
  ssl_policy        = "ELBSecurityPolicy-2015-05"
  certificate_arn   = "${aws_iam_server_certificate.default.arn}"

  default_action {
    target_group_arn = "${aws_lb_target_group.test.arn}"
    type             = "forward"
  }
}

resource "aws_iam_server_certificate" "default" {
  name             = "terraform-default-cert-%s"
  certificate_body = <<EOF
-----BEGIN CERTIFICATE-----
MIICpDCCAYwCCQC8EdACDsZ33jANBgkqhkiG9w0BAQsFADAUMRIwEAYDVQQDDAls
b2NhbGhvc3QwHhcNMTcxMjE1MjE0MDAxWhcNMTgxMjE1MjE0MDAxWjAUMRIwEAYD
VQQDDAlsb2NhbGhvc3QwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCv
f6dnvrQrLtjFt5NHaKoNO7ZzM4IfplqpQw82msjBiGDR2O3NQWOPTnZHtrm/Xumi
ClY9p2sKOTn9Bwz/+7SUK2/OF6XGrWt/Sqz2Hh+qVNg4fbvU2FbSwMWfyB4QJgMZ
M4l3oHcTyJ2I8vysvX/AJvGnjd2m7ADehTqB8VRR6hkkFWZo5hpg2PQP89ZfbJDq
ZxXyxg6UYck7bxiIdbQPRRWYjYPAwHpcELlCo2pAoTp75XB7NmKtNwege10Ck/CV
ttjOXtHSu1vflUL+3Kue8epoja5E7KelIUVu9KwIerlWgYddkYEi4hRgAsc17OJr
J0Jf6k6VdN9sAw5JCTU3AgMBAAEwDQYJKoZIhvcNAQELBQADggEBAG4+p3+192l6
5pGs9FD2QJA/QUI1lc+DX5lzJIoozfjzM8fhrKPNBe4hOK4RoDdaEc2AKzNwbYjV
ceU0EKx9es6r5lohDpuzuYW4T8RlC3jOrHZ5XrNgi4htDG7/KMEDiHJJSUll4fX5
z/qFSfJHdm8Imms0XvxZXKFdepI8caKEugD1/y0KPXO3qiDc57tZFDi27Nmvi/tj
YicZTuYAgr9cSisbN3+ZkBvc/m3P77GSSRYOKeVQjSJ7ATRvZPQ3Wsx2Dr7ZlEGy
eDtW8C9S3T4sZU3w32uJo+1GP4fGzf7CPQ+vnUH3S2+oQ7y32ufcAY3fYFBd4gIu
mDP3OUufD7c=
-----END CERTIFICATE-----
EOF
  private_key      = <<EOF
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAr3+nZ760Ky7YxbeTR2iqDTu2czOCH6ZaqUMPNprIwYhg0djt
zUFjj052R7a5v17pogpWPadrCjk5/QcM//u0lCtvzhelxq1rf0qs9h4fqlTYOH27
1NhW0sDFn8geECYDGTOJd6B3E8idiPL8rL1/wCbxp43dpuwA3oU6gfFUUeoZJBVm
aOYaYNj0D/PWX2yQ6mcV8sYOlGHJO28YiHW0D0UVmI2DwMB6XBC5QqNqQKE6e+Vw
ezZirTcHoHtdApPwlbbYzl7R0rtb35VC/tyrnvHqaI2uROynpSFFbvSsCHq5VoGH
XZGBIuIUYALHNeziaydCX+pOlXTfbAMOSQk1NwIDAQABAoIBAQCWaZkn0ImITUFa
y8h9tlWwu9HWkHng+GnRkfjy+tw/Csy4bez6MyXKSBwVwKUYQJeK2sMpWljiTUPG
+gkJSEhviX7sqtXZHv73/R+aXR0Ull0upYybksNvI+r80734ZyvWqJYUIkKMgS+L
lX476roYDQima29iRflEvfj0L8rt3IMlIpWvJwRMscgkdJsRktAabpjf5UuH6KTr
j3J2ciimUGF/I6rnZUCl8Yu+tyHFhyS9kNHilX91a/zcgQDmqsXu8AvW+KO+tL1M
ivr892hLQpVOqkenBS9AafInXAOdrBqDNjwD0V5wq3D0qKFmH1wF3fCkYeMAm/PX
WqgKWJNZAoGBANrZEZE57evkGPsG8+1Jm+PFMo/oW0zBBKTdeMtwo/pU53+/STMr
T0hTP/pwgZZqcfhVshFormzP+8/O4ycoJmAMcw3eXodpXvqd+ry7xVb1LzPH1dqI
70WQvHS1REMzx99DgcybagQnGClyFn0KFpIsNL5ImHPkgEy/oHF6zqUtAoGBAM1K
qk6jOAcqPfBD3iK/gvXaUJjz6j4RIO5wkJsMSRYDoUpbicFbypuSL/hlkHjP9Y6Y
rq1cvfD5YvgCURTaRRuiFy7ny44cFCeIM2HCiRt4a3aXYsZer5GXN2VBBh8Iz/cm
lsdn3W8XcfvPgJ51GlYJpBSlN0kr6sRv7hECDkpzAoGBAKZylFblPVy9RnaeSiX+
Zy9sS1GCgvY0k8iknXv1tvHtY4kYvp7JYOp8Ttu2eAkj+nzLCL0O5iLiaP4bt06P
zegdb+BrcXACJ3frccnb8nJ51qXGZpNotLsvIvaM61dFac4YNP+ecJqp9UmIeSwu
4Q3Zy1+yLSlv8FjvIiNNKSAdAoGAOj3hnVe/EIFSezS188PDgr6COTKSFTPE1QDI
dcSBg6ZZ/v+DUIEbNRG/XEhsOWo+b0sv221BUfles5/sou7dxl4xF5SZcmLS8Pg6
I5UOUuXSDx4Z3s+EHdj51VciRnG4lpSzGDWGY/sR0m/nPI1agGhRza2lxrOX8k0T
DG454bMCgYBpsawRAUyqzv7uLNlrTToHIwq52kA5oPjV+N2+hQokbw2bXdTcJBtu
o82dP3VvmdziDBfTa+ti95Hd3jFZ3lZbLb82iapkIro4fC0Zl7k70SD73Za+CO5s
ObJC+peQTdmNKPZqgkB3flK3K7EZflzPEFrtlDUfo8vNSYXtELdjPA==
-----END RSA PRIVATE KEY-----
EOF

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_iam_server_certificate" "additional_1" {
  name             = "terraform-additional-cert-1-%s"
  certificate_body = <<EOF
-----BEGIN CERTIFICATE-----
MIICpDCCAYwCCQD3BjmOb0++dDANBgkqhkiG9w0BAQsFADAUMRIwEAYDVQQDDAls
b2NhbGhvc3QwHhcNMTcxMjE1MjE0MDMzWhcNMTgxMjE1MjE0MDMzWjAUMRIwEAYD
VQQDDAlsb2NhbGhvc3QwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDJ
Kfhl1MOfPG6Usr8djY7dHyVxyuKiQVGQrEszRjVSLSucyuZdpGULDMbYQkK9zp4O
iWukaE2jzaRL4MxrMEMjdhuGVciHyG+rLyW8K2SnBk0BjNibDBCPvXWFg5C4v4N5
phPyr0t6DlxcqvBfGV0/nQWesHjvwyR2L0/mXY+2zCfChJNOzmpfXBErrBz6yurF
Ssydl8qaxRyUYmBSYSbdU4WobagdrdAAeoZnf3lCGlDxFlhXn0mH83bfEarbM/YI
PhDAmedGbosqsxwTNCV12CQg2d+2CwyByqgubXDYE/befNO/RscXTzx6TZYl+Kg/
b06Loxl7oU9/Bh8So7vhAgMBAAEwDQYJKoZIhvcNAQELBQADggEBAE0dqTzU7sGq
zTgfcS+kYO/q8EoKv9owyKb5wgiHhW3SUzbYhDF182g8F9qF51eoNmxgVKxS++ej
5kbsLpy/yNc7e4BBhleP8bye15/rEYzkxUjqMkuTA3K1WKghbWvuonaIeaQP26x4
l4vkAkxbYHI+japny8I9Y4ZtbzypIOCbku7lnC5r9FEQy1jh1DH8NRobP/4vO+6d
3T4RlkBgHeJy1p0LEkmsTZd9L0BeYEQHorj1/Hd9V3o5a0dKy1hfFxUgr2XR6HwD
tVtG7eKD+NJsJWBngxqEJsJAAgvC+QPO7Lr/YpdM9breumTNM9EINojZpwCKpMzG
JmznQJoWXI8=
-----END CERTIFICATE-----
EOF
  private_key      = <<EOF
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAySn4ZdTDnzxulLK/HY2O3R8lccriokFRkKxLM0Y1Ui0rnMrm
XaRlCwzG2EJCvc6eDolrpGhNo82kS+DMazBDI3YbhlXIh8hvqy8lvCtkpwZNAYzY
mwwQj711hYOQuL+DeaYT8q9Leg5cXKrwXxldP50FnrB478Mkdi9P5l2PtswnwoST
Ts5qX1wRK6wc+srqxUrMnZfKmsUclGJgUmEm3VOFqG2oHa3QAHqGZ395QhpQ8RZY
V59Jh/N23xGq2zP2CD4QwJnnRm6LKrMcEzQlddgkINnftgsMgcqoLm1w2BP23nzT
v0bHF088ek2WJfioP29Oi6MZe6FPfwYfEqO74QIDAQABAoIBAHLR2O1OvwXBuaB4
UiutI/FEmNvVoPxZ6hN4tEek+ERacf1BtrGWZvIirdC8KVord/32JBGoU0B+3RtO
SX+ZAYlZHabUiewu1MZR1+kKn53SM9wBp5UAEufojQ7TJKS+821ZBSPNOHiHf+KI
00CEurvXhWCpe63mpYYrxSBJIQGFjM/f1/wFJMI8KScHNIyxcGzwdZzpzPGFsCXH
l1ZgrCLBPlrB7V3rAwhk5qc2UOikytS1V1L5tBOsOhNInxAT7dlmVR8HlH3aR63K
LvIOxCNMmOvwM4SiDU+5ocOT2uOppKtsCxQj23S1ek1d43kQ3ay4g6fa179tJdFg
10FYye0CgYEA5F8e7wOXRQuKGUvyvt4IyTi/zWkd6JC6LBf8C+iHJKalBs5wKSgo
Kd5iaCxwzFhqTqsLdJn4SzmK3a/5dfR5V9Kqmi3Csz7VSQvhPstLD6b1XgG8zhES
oGHdoKlMBkpyQDFxqR93yXBhfkyMbU7NDDz/OpxYjjfESSPm6qa0E0MCgYEA4YA0
SxMAOiSnhXevRPEvWQ4fByJbsJd7lMOoULfXdZSraIVOGzbOy0qWplmSxUGFVeZB
/o8TBtZ1QsXOp9sfSdaxz6eetlfiCWgb+Be2dGFEKEo1BL7LXT3HOoY1JKtX11AU
40xwlemNTWBHoZxKzfNMYr5ZUnfHqgZaIYnq+AsCgYEAvC712pbm6+paXgYLfeSQ
8N9mjel0z9OS1grdkyKFWlpH2pf6LK/+iKHMDXOxb3HcB/9CbU8DH1nHaG497kkK
RKhAFNRogDPipVK4xXnX3IoD3vcnkdbXtnlum5lmIDjwdJ1Jv8dCeie07tI9VUV+
Cfupha6X+nlRscN34RyFfukCgYEAmrE2LmIkf75xZS/LfoHttyvmwSAiwivIzS7D
okvbdH++bn80K5sXlYHfgtJjywm9jEXe89/2b3RjEKFduOyqtB6h8A/O4su69jUD
KtgphADNns35PP4dyCL/LviuMC+SnNQE4ECp401Kb9Aik40CC/JhbkOiRt6Ai/S7
k9Jm7C0CgYAr7C2UTXjX/a8GJjfThm7AxKHAbwFSXY4pJn0kHzNsjPf+cvhlvpyv
j37PW1IgYNcjS4yefPa1M603m8fqwM8kk+Fd2GNZKACPLv5jERW8LHH3gw7a9Qc9
RRQw28RZVtgssJ81L2ygkishP2P3atpStG9sKhBb0HRWZrUIW5hRRw==
-----END RSA PRIVATE KEY-----
EOF

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_iam_server_certificate" "additional_2" {
  name             = "terraform-additional-cert-2-%s"
  certificate_body = <<EOF
-----BEGIN CERTIFICATE-----
MIICpDCCAYwCCQDZ2oRa1sGckDANBgkqhkiG9w0BAQsFADAUMRIwEAYDVQQDDAls
b2NhbGhvc3QwHhcNMTcxMjE1MjE0MDU5WhcNMTgxMjE1MjE0MDU5WjAUMRIwEAYD
VQQDDAlsb2NhbGhvc3QwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDJ
F2VR5tsSSnsKQ3FQvYzSl4CwhElQB9RLFeaV2ss7z/xKcZ8QZOYTbSOz4viNT1f7
eIqG/rhOsi6lecIITP40YeXFy9rJI1g2rhn7lTsU9v01v2RNVL1ST7i/0r43/zcO
XXC1qC+iT4Y3v7MMYobLJUAehUxNxavIFVJXn13dlNZYWQGKN66vidsdWRH/hdKT
qbSaIrCd3UwD5YxJE+gryRX6nMyZo9p6cdAtAv6ph2N4vF/bJYcm9OzsihoT8Uf0
X6Y9qisXjQvUUbzzejDw3eosR1r5T3Xz1IaNrvDDV2Pwe8b0BN8M2ZbfjBMoU5LL
KLJTkV15emzm2rTYyEgzAgMBAAEwDQYJKoZIhvcNAQELBQADggEBAMHCMOTV5qT6
3eU/BN05fXdzQcrN2d2BOhV6swCWopcOToXQtaQS2pmtRLs6WrlifvAzv+SvPOPH
hJCBmpK1P6YSCxj2e9TYBkbCvs0CFTImFkaVQ8XZqBCZ2YV6eXouYvJLj7M9G/pH
s5Uca+40RIbTEwOgxsGmFRF7aP/5cNCJzBM+9u4tu9jECH6Vd6EA0C0t7Ekfl+qW
M3x8po9CFSbgugb2CzuFw1o9LV3NDXtjC+qKTTsy3Ql4n27FaaOedTAuflp2EFHM
bQ99YIot0P5hbhxYeHE5II8dw7MMHFLzAWl9euQ9se+90R4TQMNrgD6tXDlSmilo
TKl9pmk1gUw=
-----END CERTIFICATE-----
EOF
  private_key      = <<EOF
-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEAyRdlUebbEkp7CkNxUL2M0peAsIRJUAfUSxXmldrLO8/8SnGf
EGTmE20js+L4jU9X+3iKhv64TrIupXnCCEz+NGHlxcvaySNYNq4Z+5U7FPb9Nb9k
TVS9Uk+4v9K+N/83Dl1wtagvok+GN7+zDGKGyyVAHoVMTcWryBVSV59d3ZTWWFkB
ijeur4nbHVkR/4XSk6m0miKwnd1MA+WMSRPoK8kV+pzMmaPaenHQLQL+qYdjeLxf
2yWHJvTs7IoaE/FH9F+mPaorF40L1FG883ow8N3qLEda+U9189SGja7ww1dj8HvG
9ATfDNmW34wTKFOSyyiyU5FdeXps5tq02MhIMwIDAQABAoIBADBhyKbj/GFyOhhG
EcVzVaZ1fSj7KwhuWc2W/1uewLcrW3At1i+DlzelUqm9OkAFjw2Z+vpv3rhQdpip
qt2EaMUBqN7mJUWvk1HUobu+M/DfXBzKZ7+TW9mqBLFiaxHd/ckfAjcyuAM9TvWq
0dFxAy1tUPgG6kzr+mCxgJZEabkd4R+t/7VC44Z4/RahRF57B5NpuJlCVTh4pFY2
5dalu2DkPIqh8j8LoVxPUJsq84lUNZ8G87qIJYzVlKjkT7MHHyMYTTTZ4Aw43ts3
HboHOZN8SbFrslPb0uRDoOCgKBBcqNttdVjRBDBTPdABsFFvfJERch48onGnotH1
Fz2kycECgYEA5+XzTWKW09dJE234MpVc/jZgiDlDrXNW7Es9ipq4RXNcOuyIV3SX
JygCieeo2QY7wahVzOLFHstqwn2KgKFUW4GChnu18aF1Sm11KPdvmu2kd1FsL1lL
DczbBzsfhEOsBV+BNByvwBRQ8JxDOxAGmu8I6fRpEejD8Cewuf+psL8CgYEA3f3G
vItWKMCrvv00iKgfo672HN+4M7Jq4/YJ326G+9e7vI0AHdjXs5UvCEa5HN2/Pspa
D4EUrzFoHLDVYLMN4w+kXDMY5uLri5BTy6+W7mtLrZlv9LEP21t4sxOuEKmsx0A3
/u6bsjyaFYUkuzXJjwj3VlbdeKmd3d+i4bx60Y0CgYEAhgZzwNTrIRI67NzQ5sNG
lLHuxqx5/eQ8Z6LwtYvIVnNe6btM7Wa3+Wx5UySthIjCvqFAvYKOtMaSNEgEZWVY
cO5/9qPHOxi6xkJOxVeEjEEunbtUUGVGKHquWBaGl5XY9N4GuYye0t+rC/T0Mk2H
08G1ICofE7e4jrMRw94MP9cCgYEAkbeCm8LuxINseUrmEAoj8qLnZJ6p4C1uosKf
Sm3X5zp+Pk9j0zPCq1vy6oDaBAu+/K2BHopBlJpe1+5vsjD2naRn5CmaX/x2Vz9e
8eYJsej2XTbJ2ZncacYKXao+aSungYcy+oGP7BiBoDyslsLA0sY07RTJ/emA+eJR
ndHF4QUCgYEA0pB7gvCEjGv5fRqwm+NuhRfZRLgEkelfxH+27wGJx+WJPTv82C0d
MV22T3MKUgVhbRyOAKqPSoKaTwfa7UQbe3kG7Mha4dZhFYWd+heghdGP7ZK+f3N+
3CldC6E+JbKRk5CGMnYgL/ULnczwp50bGFAq9HaqmpU4cXeJQfRgnxc=
-----END RSA PRIVATE KEY-----
EOF

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_iam_server_certificate" "additional_3" {
  name             = "terraform-additional-cert-3-%s"
  certificate_body = <<EOF
-----BEGIN CERTIFICATE-----
MIICpDCCAYwCCQC5bnxXukDHoTANBgkqhkiG9w0BAQsFADAUMRIwEAYDVQQDDAls
b2NhbGhvc3QwHhcNMTcxMjE2MTEzNjMzWhcNMTgxMjE2MTEzNjMzWjAUMRIwEAYD
VQQDDAlsb2NhbGhvc3QwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDD
ZWZ5jVrp6vQ01crpVlrUDyfKTHdlviFagVLa0VBbFdwhQL08fZGKEIURA3s+i41B
NfLGZsdMtu/TjvOsQ67HdT4sMdaO5TvwCdNcXfvQP4IAWBORWTSTs23gKdGJr37i
f4Ptg6vJEP6QCho26XYcNvuFG9OMLumrXxIrFQWiozxcjNl6yCvA4ycCPVzLrOKV
l2gseZsvgwIyppU7uA4ziC3hfOmeuIZ3epJR8VICZFDximS0Fxw6cZQ09UWESBkL
iaxBIKNmp246GsxV0D1KRA0mBVNAwNLxfp3ZxUhULhqEcXEv0YKSYJlaYyvxKmeL
7colOYh1ZZUyuCayxLEvAgMBAAEwDQYJKoZIhvcNAQELBQADggEBAI4lf2BQ08WT
DJ7k4ufdZpzrQtoftaCMPZg1lWHzXJdy9GML0AAWHQ9gq/HgdNeLMxp06aO7od41
AMN+Xy3A1qPrAsJwR9B/Db2Bo+BBE5xt+SL7+HsMaDW7IiLNIRPCN/MABfJQTwLQ
pp97lhBkhR2f25Sz7t+ZrR/q8OOrC1WT6oer+wndwRVJ2Rza58XGtCvPSZpz62SG
VD+LI9F3mAEBY6yzTe0zxuJV1Fo4MQJFrPk4dVbkF29Cuc2sEFQxh22B4zTUhaEW
OkdfahzOdUm4La95V+RlJqsyazrlUvu245+vL5sp5dqUkNcIO3SSo65ZM0p/xcu3
24IKYRPOa8g=
-----END CERTIFICATE-----
EOF
  private_key      = <<EOF
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAw2VmeY1a6er0NNXK6VZa1A8nykx3Zb4hWoFS2tFQWxXcIUC9
PH2RihCFEQN7PouNQTXyxmbHTLbv047zrEOux3U+LDHWjuU78AnTXF370D+CAFgT
kVk0k7Nt4CnRia9+4n+D7YOryRD+kAoaNul2HDb7hRvTjC7pq18SKxUFoqM8XIzZ
esgrwOMnAj1cy6zilZdoLHmbL4MCMqaVO7gOM4gt4XzpnriGd3qSUfFSAmRQ8Ypk
tBccOnGUNPVFhEgZC4msQSCjZqduOhrMVdA9SkQNJgVTQMDS8X6d2cVIVC4ahHFx
L9GCkmCZWmMr8Spni+3KJTmIdWWVMrgmssSxLwIDAQABAoIBAADFrtD+KQTRo+Nt
kN65M6Sw5qRbBwHE7ZbQ+gXZW+rwPC15dwX9LQ0RoaT+kYcewBEo6Gu2TkVUV8BL
SVU98zvgj71W+RUZfNInB8nOqUcaaSYdMv6ZDqcix2ViZOyZX/P/MwNGSPCDe64Q
DCh5ZbkY0oelI4HjUZMWzhiTfbE12CrPZd/nW4JEvxtx9iAKNYG2bpANN39/lYL2
FhFP6XC93D0vAR4HUCrx7kIFEgYN8dLXwfFlsfTilSbh2aXP0+8E5eGnhVFZJ68R
LqPM4r3WGQTAoZNDkdLQZ8sTyM+KAI0Bh5HI7HPFmwHv7xhtuPpBflnyB4LBgakl
/dLzcDECgYEA7eayS8nszhP4Ppcz6GMPPJZG8ktHrun9T3ZOrI+sfH6+kbuSlpeS
GwU50XqAKuyMzLqammDxOhGa5mgPKgpIgJbjMMqhEpIuPia5CRkeGNL76x6UZBmv
0lqHZCDkwG2PgPpF99/2Mu6mzhr1nrL9hJy7XU/jP7uR3oMB7mIHQlkCgYEA0kLi
xbuABhN15z2AfU2ZJ46yCqcYq1FNWT08NGLsLzT8CdGoypgFm5+fcNQm+nuuH5su
RblQCR9o7ZaIrB6YEDSR0uLTm9T+l6PSu1jXKqvTpKXvIK1ZEgCAA97atglmK9FB
o4DKRg5q8EVqjd93ms0gCJG/mALz62ngv8kRTscCgYEA3pyJ6GtZ4HhFSeRY2TKg
llQUrTMOL7mapBmTgtuqTpCXKG05vRq1x/z63m7fscrJ8eUHOEBQWcMRjFqBHhij
QVhv3T8uu9730IaRjNbpF9eNjbR+rLBwmsjFekdhZkLfDpSillEG4x/4DFKj1c2t
dsLmfGl9vyx4UZuLWhJ2snkCgYEArcUvjAGJLfxZXfIbRfOi1ul8xYcRwUyhI1aT
ciwrTFx6zFalLEJ1qAdFC0eaqzsaTe6/UEp2FgQKgQj/DVj/ja0Us3hZMJnYi1SO
bd1fflmhwZqNxbDeQx18rzY4BWhBM4duuwVOppV11ftYs8XzIFHU3qAt+yC9nFrV
r0sqbbMCgYBVPDZSQihTJGPkhX6iByf2PzeaHjhrW+YBpFPXdbkG1mheaFrcPoCq
RhecogLX06+XvjPxOZhw3E0FBdSZTj7PYSn5/bodIP1/rbWTWeFcM/I77plcmh40
cwnHydfTRl8f/DF2k0j3RErUxSx5vAQPV5RT6t5lGEAmoskXXgtgyQ==
-----END RSA PRIVATE KEY-----
EOF

  lifecycle {
    create_before_destroy = true
  }
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags {
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
  tags {
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
