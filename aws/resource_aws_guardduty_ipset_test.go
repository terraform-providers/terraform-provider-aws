package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func testAccAwsGuardDutyIpset_basic(t *testing.T) {
	bucketName := fmt.Sprintf("tf-test-%s", acctest.RandString(5))
	keyName1 := fmt.Sprintf("tf-%s", acctest.RandString(5))
	keyName2 := fmt.Sprintf("tf-%s", acctest.RandString(5))
	ipsetName1 := fmt.Sprintf("tf-%s", acctest.RandString(5))
	ipsetName2 := fmt.Sprintf("tf-%s", acctest.RandString(5))
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsGuardDutyIpsetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardDutyIpsetConfig_basic(bucketName, keyName1, ipsetName1, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyIpsetExists("aws_guardduty_ipset.test"),
					resource.TestCheckResourceAttr("aws_guardduty_ipset.test", "name", ipsetName1),
					resource.TestCheckResourceAttr("aws_guardduty_ipset.test", "activate", "true"),
					resource.TestMatchResourceAttr(
						"aws_guardduty_ipset.test", "location", regexp.MustCompile(fmt.Sprintf("%s/%s$", bucketName, keyName1)),
					),
				),
			},
			{
				Config: testAccGuardDutyIpsetConfig_basic(bucketName, keyName2, ipsetName2, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyIpsetExists("aws_guardduty_ipset.test"),
					resource.TestCheckResourceAttr("aws_guardduty_ipset.test", "name", ipsetName2),
					resource.TestCheckResourceAttr("aws_guardduty_ipset.test", "activate", "false"),
					resource.TestMatchResourceAttr(
						"aws_guardduty_ipset.test", "location", regexp.MustCompile(fmt.Sprintf("%s/%s$", bucketName, keyName2)),
					),
				),
			},
		},
	})
}

func testAccAwsGuardDutyIpset_import(t *testing.T) {
	resourceName := "aws_guardduty_ipset.test"
	bucketName := fmt.Sprintf("tf-test-%s", acctest.RandString(5))
	keyName := fmt.Sprintf("tf-%s", acctest.RandString(5))
	ipsetName := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsGuardDutyIpsetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardDutyIpsetConfig_basic(bucketName, keyName, ipsetName, true),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
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
%s

resource "aws_s3_bucket" "test" {
  acl = "private"
  bucket = "%s"
  force_destroy = true
}

resource "aws_s3_bucket_object" "test" {
  acl     = "public-read"
  content = "10.0.0.0/8\n"
  bucket  = "${aws_s3_bucket.test.id}"
  key     = "%s"
}

resource "aws_guardduty_ipset" "test" {
  name = "%s"
  detector_id = "${aws_guardduty_detector.test.id}"
  format = "TXT"
  location = "https://s3.amazonaws.com/${aws_s3_bucket_object.test.bucket}/${aws_s3_bucket_object.test.key}"
  activate = %t
}
`, testAccGuardDutyDetectorConfig_basic1, bucketName, keyName, ipsetName, activate)
}
