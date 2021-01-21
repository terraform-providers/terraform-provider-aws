package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_guardduty_publishing_destination", &resource.Sweeper{
		Name: "aws_guardduty_publishing_destination",
		F:    testSweepGuarddutyPublishingDestinations,
	})
}

func testSweepGuarddutyPublishingDestinations(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).guarddutyconn
	var sweeperErrs *multierror.Error

	detect_input := &guardduty.ListDetectorsInput{}

	err = conn.ListDetectorsPages(detect_input, func(page *guardduty.ListDetectorsOutput, lastPage bool) bool {
		for _, detectorID := range page.DetectorIds {
			list_input := &guardduty.ListPublishingDestinationsInput{
				DetectorId: detectorID,
			}

			err = conn.ListPublishingDestinationsPages(list_input, func(page *guardduty.ListPublishingDestinationsOutput, lastPage bool) bool {
				for _, destination_element := range page.Destinations {
					input := &guardduty.DeletePublishingDestinationInput{
						DestinationId: destination_element.DestinationId,
						DetectorId:    detectorID,
					}

					log.Printf("[INFO] Deleting GuardDuty Publishing Destination: %s", *destination_element.DestinationId)
					_, err := conn.DeletePublishingDestination(input)

					if err != nil {
						sweeperErr := fmt.Errorf("error deleting GuardDuty Publishing Destination (%s): %w", *destination_element.DestinationId, err)
						log.Printf("[ERROR] %s", sweeperErr)
						sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					}
				}
				return !lastPage
			})
		}
		return !lastPage
	})

	if err != nil {
		sweeperErr := fmt.Errorf("Error receiving Guardduty detectors for publishing sweep : %w", err)
		log.Printf("[ERROR] %s", sweeperErr)
		sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
	}

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping GuardDuty Publishing Destination sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error retrieving GuardDuty Publishing Destinations: %s", err)
	}

	return sweeperErrs.ErrorOrNil()
}

func testAccAwsGuardDutyPublishingDestination_basic(t *testing.T) {
	resourceName := "aws_guardduty_publishing_destination.test"
	bucketName := acctest.RandomWithPrefix("tf-acc-test")
	detectorResourceName := "aws_guardduty_detector.test_gd"
	bucketResourceName := "aws_s3_bucket.gd_bucket"
	kmsKeyResourceName := "aws_kms_key.gd_key"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsGuardDutyPublishingDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsGuardDutyPublishingDestinationConfig_basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyPublishingDestinationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", detectorResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "destination_arn", bucketResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_arn", kmsKeyResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "destination_type", "S3")),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsGuardDutyPublishingDestination_disappears(t *testing.T) {
	resourceName := "aws_guardduty_publishing_destination.test"
	bucketName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsGuardDutyPublishingDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsGuardDutyPublishingDestinationConfig_basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyPublishingDestinationExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsGuardDutyPublishingDestination(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAwsGuardDutyPublishingDestinationConfig_basic(bucketName string) string {
	return fmt.Sprintf(`

data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

data "aws_partition" "current" {}

data "aws_iam_policy_document" "bucket_pol" {
  statement {
    sid = "Allow PutObject"
    actions = [
      "s3:PutObject"
    ]

    resources = [
      "${aws_s3_bucket.gd_bucket.arn}/*"
    ]

    principals {
      type        = "Service"
      identifiers = ["guardduty.${data.aws_partition.current.dns_suffix}"]
    }
  }

  statement {
    sid = "Allow GetBucketLocation"
    actions = [
      "s3:GetBucketLocation"
    ]

    resources = [
      aws_s3_bucket.gd_bucket.arn
    ]

    principals {
      type        = "Service"
      identifiers = ["guardduty.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

data "aws_iam_policy_document" "kms_pol" {

  statement {
    sid = "Allow GuardDuty to encrypt findings"
    actions = [
      "kms:GenerateDataKey"
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:kms:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:key/*"
    ]

    principals {
      type        = "Service"
      identifiers = ["guardduty.${data.aws_partition.current.dns_suffix}"]
    }
  }

  statement {
    sid = "Allow all users to modify/delete key (test only)"
    actions = [
      "kms:*"
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:kms:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:key/*"
    ]

    principals {
      type        = "AWS"
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
  }

}

resource "aws_guardduty_detector" "test_gd" {
  enable = true
}

resource "aws_s3_bucket" "gd_bucket" {
  bucket        = %[1]q
  acl           = "private"
  force_destroy = true
}

resource "aws_s3_bucket_policy" "gd_bucket_policy" {
  bucket = aws_s3_bucket.gd_bucket.id
  policy = data.aws_iam_policy_document.bucket_pol.json
}

resource "aws_kms_key" "gd_key" {
  description             = "Temporary key for AccTest of TF"
  deletion_window_in_days = 7
  policy                  = data.aws_iam_policy_document.kms_pol.json
}

resource "aws_guardduty_publishing_destination" "test" {
  detector_id     = aws_guardduty_detector.test_gd.id
  destination_arn = aws_s3_bucket.gd_bucket.arn
  kms_key_arn     = aws_kms_key.gd_key.arn

  depends_on = [
    aws_s3_bucket_policy.gd_bucket_policy,
  ]
}`, bucketName)
}

func testAccCheckAwsGuardDutyPublishingDestinationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		destination_id, detector_id, err_state_read := decodeGuardDutyPublishDestinationID(rs.Primary.ID)

		if err_state_read != nil {
			return err_state_read
		}

		input := &guardduty.DescribePublishingDestinationInput{
			DetectorId:    aws.String(detector_id),
			DestinationId: aws.String(destination_id),
		}

		conn := testAccProvider.Meta().(*AWSClient).guarddutyconn
		_, err := conn.DescribePublishingDestination(input)
		return err
	}
}

func testAccCheckAwsGuardDutyPublishingDestinationDestroy(s *terraform.State) error {

	conn := testAccProvider.Meta().(*AWSClient).guarddutyconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_guardduty_publishing_destination" {
			continue
		}

		destination_id, detector_id, err_state_read := decodeGuardDutyPublishDestinationID(rs.Primary.ID)

		if err_state_read != nil {
			return err_state_read
		}

		input := &guardduty.DescribePublishingDestinationInput{
			DetectorId:    aws.String(detector_id),
			DestinationId: aws.String(destination_id),
		}

		_, err := conn.DescribePublishingDestination(input)
		// Catch expected error.
		if err == nil {
			return fmt.Errorf("Resource still exists.")
		}
	}
	return nil
}
