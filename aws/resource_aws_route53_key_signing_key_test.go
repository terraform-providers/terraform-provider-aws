package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfroute53 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/route53"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/route53/finder"
)

func TestAccAwsRoute53KeySigningKey_basic(t *testing.T) {
	kmsKeyResourceName := "aws_kms_key.test"
	route53ZoneResourceName := "aws_route53_zone.test"
	resourceName := "aws_route53_key_signing_key.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckRoute53KeySigningKey(t) },
		ErrorCheck:        testAccErrorCheckSkipRoute53(t),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53KeySigningKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53KeySigningKeyConfig_Name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsRoute53KeySigningKeyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "digest_algorithm_mnemonic", "SHA-256"),
					resource.TestCheckResourceAttr(resourceName, "digest_algorithm_type", "2"),
					resource.TestMatchResourceAttr(resourceName, "digest_value", regexp.MustCompile(`^[0-9A-F]+$`)),
					resource.TestMatchResourceAttr(resourceName, "dnskey_record", regexp.MustCompile(`^257 [0-9]+ [0-9]+ [a-zA-Z0-9+/]+={0,3}$`)),
					resource.TestMatchResourceAttr(resourceName, "ds_record", regexp.MustCompile(`^[0-9]+ [0-9]+ [0-9]+ [0-9A-F]+$`)),
					resource.TestCheckResourceAttr(resourceName, "flag", "257"),
					resource.TestCheckResourceAttrPair(resourceName, "hosted_zone_id", route53ZoneResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "key_management_service_arn", kmsKeyResourceName, "arn"),
					resource.TestMatchResourceAttr(resourceName, "key_tag", regexp.MustCompile(`^[0-9]+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestMatchResourceAttr(resourceName, "public_key", regexp.MustCompile(`^[a-zA-Z0-9+/]+={0,3}$`)),
					resource.TestCheckResourceAttr(resourceName, "signing_algorithm_mnemonic", "ECDSAP256SHA256"),
					resource.TestCheckResourceAttr(resourceName, "signing_algorithm_type", "13"),
					resource.TestCheckResourceAttr(resourceName, "status", tfroute53.KeySigningKeyStatusActive),
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

func TestAccAwsRoute53KeySigningKey_disappears(t *testing.T) {
	resourceName := "aws_route53_key_signing_key.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckRoute53KeySigningKey(t) },
		ErrorCheck:        testAccErrorCheckSkipRoute53(t),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53KeySigningKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53KeySigningKeyConfig_Name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsRoute53KeySigningKeyExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsRoute53KeySigningKey(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsRoute53KeySigningKey_Status(t *testing.T) {
	resourceName := "aws_route53_key_signing_key.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckRoute53KeySigningKey(t) },
		ErrorCheck:        testAccErrorCheckSkipRoute53(t),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53KeySigningKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53KeySigningKeyConfig_Status(rName, tfroute53.KeySigningKeyStatusInactive),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsRoute53KeySigningKeyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "status", tfroute53.KeySigningKeyStatusInactive),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsRoute53KeySigningKeyConfig_Status(rName, tfroute53.KeySigningKeyStatusActive),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsRoute53KeySigningKeyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "status", tfroute53.KeySigningKeyStatusActive),
				),
			},
			{
				Config: testAccAwsRoute53KeySigningKeyConfig_Status(rName, tfroute53.KeySigningKeyStatusInactive),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsRoute53KeySigningKeyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "status", tfroute53.KeySigningKeyStatusInactive),
				),
			},
		},
	})
}

func testAccCheckAwsRoute53KeySigningKeyDestroy(s *terraform.State) error {
	conn := testAccProviderRoute53KeySigningKey.Meta().(*AWSClient).r53conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_key_signing_key" {
			continue
		}

		keySigningKey, err := finder.KeySigningKeyByResourceID(conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchHostedZone) {
			continue
		}

		if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchKeySigningKey) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading Route 53 Key Signing Key (%s): %w", rs.Primary.ID, err)
		}

		if keySigningKey != nil {
			return fmt.Errorf("Route 53 Key Signing Key (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAwsRoute53KeySigningKeyExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource %s has not set its id", resourceName)
		}

		conn := testAccProviderRoute53KeySigningKey.Meta().(*AWSClient).r53conn

		keySigningKey, err := finder.KeySigningKeyByResourceID(conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error reading Route 53 Key Signing Key (%s): %w", rs.Primary.ID, err)
		}

		if keySigningKey == nil {
			return fmt.Errorf("Route 53 Key Signing Key (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAwsRoute53KeySigningKeyConfig_Base(rName string) string {
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
`, rName))
}

func testAccAwsRoute53KeySigningKeyConfig_Name(rName string) string {
	return composeConfig(
		testAccAwsRoute53KeySigningKeyConfig_Base(rName),
		fmt.Sprintf(`
resource "aws_route53_key_signing_key" "test" {
  hosted_zone_id             = aws_route53_zone.test.id
  key_management_service_arn = aws_kms_key.test.arn
  name                       = %[1]q
}
`, rName))
}

func testAccAwsRoute53KeySigningKeyConfig_Status(rName string, status string) string {
	return composeConfig(
		testAccAwsRoute53KeySigningKeyConfig_Base(rName),
		fmt.Sprintf(`
resource "aws_route53_key_signing_key" "test" {
  hosted_zone_id             = aws_route53_zone.test.id
  key_management_service_arn = aws_kms_key.test.arn
  name                       = %[1]q
  status                     = %[2]q
}
`, rName, status))
}
