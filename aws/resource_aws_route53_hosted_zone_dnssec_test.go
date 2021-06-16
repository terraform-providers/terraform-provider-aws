package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfroute53 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/route53"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/route53/finder"
)

func TestAccAwsRoute53HostedZoneDnssec_basic(t *testing.T) {
	route53ZoneResourceName := "aws_route53_zone.test"
	resourceName := "aws_route53_hosted_zone_dnssec.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckRoute53KeySigningKey(t) },
		ErrorCheck:        testAccErrorCheck(t, route53.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53HostedZoneDnssecDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53HostedZoneDnssecConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsRoute53HostedZoneDnssecExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "hosted_zone_id", route53ZoneResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "signing_status", tfroute53.ServeSignatureSigning),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsRoute53HostedZoneDnssec_disappears(t *testing.T) {
	resourceName := "aws_route53_hosted_zone_dnssec.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckRoute53KeySigningKey(t) },
		ErrorCheck:        testAccErrorCheck(t, route53.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53HostedZoneDnssecDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53HostedZoneDnssecConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsRoute53HostedZoneDnssecExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsRoute53HostedZoneDnssec(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsRoute53HostedZoneDnssec_SigningStatus(t *testing.T) {
	resourceName := "aws_route53_hosted_zone_dnssec.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckRoute53KeySigningKey(t) },
		ErrorCheck:        testAccErrorCheck(t, route53.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53HostedZoneDnssecDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53HostedZoneDnssecConfig_SigningStatus(rName, tfroute53.ServeSignatureNotSigning),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsRoute53HostedZoneDnssecExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "signing_status", tfroute53.ServeSignatureNotSigning),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsRoute53HostedZoneDnssecConfig_SigningStatus(rName, tfroute53.ServeSignatureSigning),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsRoute53HostedZoneDnssecExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "signing_status", tfroute53.ServeSignatureSigning),
				),
			},
			{
				Config: testAccAwsRoute53HostedZoneDnssecConfig_SigningStatus(rName, tfroute53.ServeSignatureNotSigning),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsRoute53HostedZoneDnssecExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "signing_status", tfroute53.ServeSignatureNotSigning),
				),
			},
		},
	})
}

func testAccCheckAwsRoute53HostedZoneDnssecDestroy(s *terraform.State) error {
	conn := testAccProviderRoute53KeySigningKey.Meta().(*AWSClient).r53conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_hosted_zone_dnssec" {
			continue
		}

		hostedZoneDnssec, err := finder.HostedZoneDnssec(conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, route53.ErrCodeDNSSECNotFound) {
			continue
		}

		if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchHostedZone) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading Route 53 Hosted Zone DNSSEC (%s): %w", rs.Primary.ID, err)
		}

		if hostedZoneDnssec != nil && hostedZoneDnssec.Status != nil && aws.StringValue(hostedZoneDnssec.Status.ServeSignature) == tfroute53.ServeSignatureSigning {
			return fmt.Errorf("Route 53 Hosted Zone DNSSEC (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAwsRoute53HostedZoneDnssecExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource %s has not set its id", resourceName)
		}

		conn := testAccProviderRoute53KeySigningKey.Meta().(*AWSClient).r53conn

		hostedZoneDnssec, err := finder.HostedZoneDnssec(conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error reading Route 53 Hosted Zone DNSSEC (%s): %w", rs.Primary.ID, err)
		}

		if hostedZoneDnssec == nil {
			return fmt.Errorf("Route 53 Hosted Zone DNSSEC (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAwsRoute53HostedZoneDnssecConfig_Base(rName string) string {
	return composeConfig(
		testAccRoute53KeySigningKeyRegionProviderConfig(),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  customer_master_key_spec = "ECC_NIST_P256"
  deletion_window_in_days  = 7
  key_usage                = "SIGN_VERIFY"
  policy = jsonencode({
    Statement = [
      {
        Action = [
          "kms:DescribeKey",
          "kms:GetPublicKey",
          "kms:Sign",
        ],
        Effect = "Allow"
        Principal = {
          Service = "api-service.dnssec.route53.aws.internal"
        }
        Sid = "Allow Route 53 DNSSEC Service"
      },
      {
        Action = "kms:*"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Resource = "*"
        Sid      = "Enable IAM User Permissions"
      },
    ]
    Version = "2012-10-17"
  })
}

resource "aws_route53_zone" "test" {
  name = "%[1]s.terraformtest.com"
}

resource "aws_route53_key_signing_key" "test" {
  hosted_zone_id             = aws_route53_zone.test.id
  key_management_service_arn = aws_kms_key.test.arn
  name                       = %[1]q
}
`, rName))
}

func testAccAwsRoute53HostedZoneDnssecConfig(rName string) string {
	return composeConfig(
		testAccAwsRoute53HostedZoneDnssecConfig_Base(rName),
		`
resource "aws_route53_hosted_zone_dnssec" "test" {
  hosted_zone_id = aws_route53_key_signing_key.test.hosted_zone_id
}
`)
}

func testAccAwsRoute53HostedZoneDnssecConfig_SigningStatus(rName string, signingStatus string) string {
	return composeConfig(
		testAccAwsRoute53HostedZoneDnssecConfig_Base(rName),
		fmt.Sprintf(`
resource "aws_route53_hosted_zone_dnssec" "test" {
  hosted_zone_id = aws_route53_key_signing_key.test.hosted_zone_id
  signing_status = %[1]q
}
`, signingStatus))
}
