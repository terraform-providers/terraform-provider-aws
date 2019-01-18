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

func testAccAwsGuardDutyThreatintelset_basic(t *testing.T) {
	bucketName := fmt.Sprintf("tf-test-%s", acctest.RandString(5))
	keyName1 := fmt.Sprintf("tf-%s", acctest.RandString(5))
	keyName2 := fmt.Sprintf("tf-%s", acctest.RandString(5))
	threatintelsetName1 := fmt.Sprintf("tf-%s", acctest.RandString(5))
	threatintelsetName2 := fmt.Sprintf("tf-%s", acctest.RandString(5))
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsGuardDutyThreatintelsetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardDutyThreatintelsetConfig_basic(bucketName, keyName1, threatintelsetName1, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyThreatintelsetExists("aws_guardduty_threatintelset.test"),
					resource.TestCheckResourceAttr("aws_guardduty_threatintelset.test", "name", threatintelsetName1),
					resource.TestCheckResourceAttr("aws_guardduty_threatintelset.test", "activate", "true"),
					resource.TestMatchResourceAttr(
						"aws_guardduty_threatintelset.test", "location", regexp.MustCompile(fmt.Sprintf("%s/%s$", bucketName, keyName1)),
					),
				),
			},
			{
				Config: testAccGuardDutyThreatintelsetConfig_basic(bucketName, keyName2, threatintelsetName2, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyThreatintelsetExists("aws_guardduty_threatintelset.test"),
					resource.TestCheckResourceAttr("aws_guardduty_threatintelset.test", "name", threatintelsetName2),
					resource.TestCheckResourceAttr("aws_guardduty_threatintelset.test", "activate", "false"),
					resource.TestMatchResourceAttr(
						"aws_guardduty_threatintelset.test", "location", regexp.MustCompile(fmt.Sprintf("%s/%s$", bucketName, keyName2)),
					),
				),
			},
		},
	})
}

func testAccAwsGuardDutyThreatintelset_import(t *testing.T) {
	resourceName := "aws_guardduty_threatintelset.test"
	bucketName := fmt.Sprintf("tf-test-%s", acctest.RandString(5))
	keyName := fmt.Sprintf("tf-%s", acctest.RandString(5))
	threatintelsetName := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsGuardDutyThreatintelsetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardDutyThreatintelsetConfig_basic(bucketName, keyName, threatintelsetName, true),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsGuardDutyThreatintelsetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).guarddutyconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_guardduty_threatintelset" {
			continue
		}

		threatIntelSetId, detectorId, err := decodeGuardDutyThreatintelsetID(rs.Primary.ID)
		if err != nil {
			return err
		}
		input := &guardduty.GetThreatIntelSetInput{
			ThreatIntelSetId: aws.String(threatIntelSetId),
			DetectorId:       aws.String(detectorId),
		}

		resp, err := conn.GetThreatIntelSet(input)
		if err != nil {
			if isAWSErr(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
				return nil
			}
			return err
		}

		if *resp.Status == guardduty.ThreatIntelSetStatusDeletePending || *resp.Status == guardduty.ThreatIntelSetStatusDeleted {
			return nil
		}

		return fmt.Errorf("Expected GuardDuty ThreatIntelSet to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAwsGuardDutyThreatintelsetExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		threatIntelSetId, detectorId, err := decodeGuardDutyThreatintelsetID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &guardduty.GetThreatIntelSetInput{
			DetectorId:       aws.String(detectorId),
			ThreatIntelSetId: aws.String(threatIntelSetId),
		}

		conn := testAccProvider.Meta().(*AWSClient).guarddutyconn
		_, err = conn.GetThreatIntelSet(input)
		return err
	}
}

func testAccGuardDutyThreatintelsetConfig_basic(bucketName, keyName, threatintelsetName string, activate bool) string {
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

resource "aws_guardduty_threatintelset" "test" {
  name = "%s"
  detector_id = "${aws_guardduty_detector.test.id}"
  format = "TXT"
  location = "https://s3.amazonaws.com/${aws_s3_bucket_object.test.bucket}/${aws_s3_bucket_object.test.key}"
  activate = %t
}
`, testAccGuardDutyDetectorConfig_basic1, bucketName, keyName, threatintelsetName, activate)
}
