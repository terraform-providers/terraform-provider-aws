package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccAwsGuardDutyIpset_basic(t *testing.T) {
	bucketName := fmt.Sprintf("tf-test-%s", acctest.RandString(5))
	keyName1 := fmt.Sprintf("tf-%s", acctest.RandString(5))
	keyName2 := fmt.Sprintf("tf-%s", acctest.RandString(5))
	ipsetName1 := fmt.Sprintf("tf-%s", acctest.RandString(5))
	ipsetName2 := fmt.Sprintf("tf-%s", acctest.RandString(5))
	resourceName := "aws_guardduty_ipset.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsGuardDutyIpsetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardDutyIpsetConfig_basic(bucketName, keyName1, ipsetName1, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyIpsetExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "guardduty", regexp.MustCompile("detector/.+/ipset/.+$")),
					resource.TestCheckResourceAttr(resourceName, "name", ipsetName1),
					resource.TestCheckResourceAttr(resourceName, "activate", "true"),
					resource.TestMatchResourceAttr(resourceName, "location", regexp.MustCompile(fmt.Sprintf("%s/%s$", bucketName, keyName1))),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGuardDutyIpsetConfig_basic(bucketName, keyName2, ipsetName2, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyIpsetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", ipsetName2),
					resource.TestCheckResourceAttr(resourceName, "activate", "false"),
					resource.TestMatchResourceAttr(resourceName, "location", regexp.MustCompile(fmt.Sprintf("%s/%s$", bucketName, keyName2))),
				),
			},
		},
	})
}

func testAccAwsGuardDutyIpset_tags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_guardduty_ipset.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsGuardDutyIpsetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardDutyIpsetConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyIpsetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGuardDutyIpsetConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyIpsetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccGuardDutyIpsetConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyIpsetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAwsGuardDutyIpsetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).guarddutyconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_guardduty_ipset" {
			continue
		}

		ipSetId, detectorId, err := decodeGuardDutyIpsetID(rs.Primary.ID)
		if err != nil {
			return err
		}
		input := &guardduty.GetIPSetInput{
			IpSetId:    aws.String(ipSetId),
			DetectorId: aws.String(detectorId),
		}

		resp, err := conn.GetIPSet(input)
		if err != nil {
			if isAWSErr(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
				return nil
			}
			return err
		}

		if *resp.Status == guardduty.IpSetStatusDeletePending || *resp.Status == guardduty.IpSetStatusDeleted {
			return nil
		}

		return fmt.Errorf("Expected GuardDuty Ipset to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAwsGuardDutyIpsetExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		ipSetId, detectorId, err := decodeGuardDutyIpsetID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &guardduty.GetIPSetInput{
			DetectorId: aws.String(detectorId),
			IpSetId:    aws.String(ipSetId),
		}

		conn := testAccProvider.Meta().(*AWSClient).guarddutyconn
		_, err = conn.GetIPSet(input)
		return err
	}
}

func testAccGuardDutyIpsetConfig_basic(bucketName, keyName, ipsetName string, activate bool) string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {}

resource "aws_s3_bucket" "test" {
  acl           = "private"
  bucket        = "%s"
  force_destroy = true
}

resource "aws_s3_bucket_object" "test" {
  acl     = "public-read"
  content = "10.0.0.0/8\n"
  bucket  = aws_s3_bucket.test.id
  key     = "%s"
}

resource "aws_guardduty_ipset" "test" {
  name        = "%s"
  detector_id = aws_guardduty_detector.test.id
  format      = "TXT"
  location    = "https://s3.amazonaws.com/${aws_s3_bucket_object.test.bucket}/${aws_s3_bucket_object.test.key}"
  activate    = %t
}
`, bucketName, keyName, ipsetName, activate)
}

func testAccGuardDutyIpsetConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {}

resource "aws_s3_bucket" "test" {
  acl           = "private"
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_object" "test" {
  acl     = "public-read"
  content = "10.0.0.0/8\n"
  bucket  = aws_s3_bucket.test.id
  key     = %[1]q
}

resource "aws_guardduty_ipset" "test" {
  activate    = true
  detector_id = aws_guardduty_detector.test.id
  format      = "TXT"
  location    = "https://s3.amazonaws.com/${aws_s3_bucket_object.test.bucket}/${aws_s3_bucket_object.test.key}"
  name        = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccGuardDutyIpsetConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {}

resource "aws_s3_bucket" "test" {
  acl           = "private"
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_object" "test" {
  acl     = "public-read"
  content = "10.0.0.0/8\n"
  bucket  = aws_s3_bucket.test.id
  key     = %[1]q
}

resource "aws_guardduty_ipset" "test" {
  activate    = true
  detector_id = aws_guardduty_detector.test.id
  format      = "TXT"
  location    = "https://s3.amazonaws.com/${aws_s3_bucket_object.test.bucket}/${aws_s3_bucket_object.test.key}"
  name        = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
