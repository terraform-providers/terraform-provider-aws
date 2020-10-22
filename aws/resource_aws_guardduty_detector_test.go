package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_guardduty_detector", &resource.Sweeper{
		Name:         "aws_guardduty_detector",
		F:            testSweepGuarddutyDetectors,
		Dependencies: []string{"aws_guardduty_publishing_destination"},
	})
}

func testSweepGuarddutyDetectors(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).guarddutyconn
	input := &guardduty.ListDetectorsInput{}
	var sweeperErrs *multierror.Error

	err = conn.ListDetectorsPages(input, func(page *guardduty.ListDetectorsOutput, lastPage bool) bool {
		for _, detectorID := range page.DetectorIds {
			id := aws.StringValue(detectorID)
			input := &guardduty.DeleteDetectorInput{
				DetectorId: detectorID,
			}

			log.Printf("[INFO] Deleting GuardDuty Detector: %s", id)
			_, err := conn.DeleteDetector(input)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting GuardDuty Detector (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping GuardDuty Detector sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error retrieving GuardDuty Detectors: %s", err)
	}

	return sweeperErrs.ErrorOrNil()
}

func testAccAwsGuardDutyDetector_basic(t *testing.T) {
	resourceName := "aws_guardduty_detector.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsGuardDutyDetectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardDutyDetectorConfig_basic1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyDetectorExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "account_id"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "guardduty", regexp.MustCompile("detector/.+$")),
					resource.TestCheckResourceAttr(resourceName, "enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", "SIX_HOURS"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGuardDutyDetectorConfig_basic2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyDetectorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable", "false"),
				),
			},
			{
				Config: testAccGuardDutyDetectorConfig_basic3,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyDetectorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable", "true"),
				),
			},
			{
				Config: testAccGuardDutyDetectorConfig_basic4,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyDetectorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", "FIFTEEN_MINUTES"),
				),
			},
		},
	})
}

func testAccAwsGuardDutyDetector_tags(t *testing.T) {
	resourceName := "aws_guardduty_detector.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsGuardDutyDetectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardDutyDetectorConfigTags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyDetectorExists(resourceName),
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
				Config: testAccGuardDutyDetectorConfigTags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyDetectorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccGuardDutyDetectorConfigTags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyDetectorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAwsGuardDutyDetectorDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).guarddutyconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_guardduty_detector" {
			continue
		}

		input := &guardduty.GetDetectorInput{
			DetectorId: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetDetector(input)
		if err != nil {
			if isAWSErr(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
				return nil
			}
			return err
		}

		return fmt.Errorf("Expected GuardDuty Detector to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAwsGuardDutyDetectorExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

const testAccGuardDutyDetectorConfig_basic1 = `
resource "aws_guardduty_detector" "test" {}
`

const testAccGuardDutyDetectorConfig_basic2 = `
resource "aws_guardduty_detector" "test" {
  enable = false
}
`

const testAccGuardDutyDetectorConfig_basic3 = `
resource "aws_guardduty_detector" "test" {
  enable = true
}
`

const testAccGuardDutyDetectorConfig_basic4 = `
resource "aws_guardduty_detector" "test" {
  finding_publishing_frequency = "FIFTEEN_MINUTES"
}
`

func testAccGuardDutyDetectorConfigTags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {
  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccGuardDutyDetectorConfigTags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {
  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}
