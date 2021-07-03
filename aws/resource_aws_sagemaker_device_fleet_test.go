package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker/finder"
)

func init() {
	resource.AddTestSweepers("aws_sagemaker_device_fleet", &resource.Sweeper{
		Name: "aws_sagemaker_device_fleet",
		F:    testSweepSagemakerDeviceFleets,
	})
}

func testSweepSagemakerDeviceFleets(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).sagemakerconn
	var sweeperErrs *multierror.Error

	err = conn.ListDeviceFleetsPages(&sagemaker.ListDeviceFleetsInput{}, func(page *sagemaker.ListDeviceFleetsOutput, lastPage bool) bool {
		for _, deviceFleet := range page.DeviceFleetSummaries {
			name := aws.StringValue(deviceFleet.DeviceFleetName)

			r := resourceAwsSagemakerDeviceFleet()
			d := r.Data(nil)
			d.SetId(name)
			err := r.Delete(d, client)
			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker Device Fleet sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Sagemaker Device Fleets: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSSagemakerDeviceFleet_basic(t *testing.T) {
	var deviceFleet sagemaker.DescribeDeviceFleetOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_device_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerDeviceFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerDeviceFleetBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerDeviceFleetExists(resourceName, &deviceFleet),
					resource.TestCheckResourceAttr(resourceName, "device_fleet_name", rName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("device-fleet/%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "output_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "output_config.0.s3_output_location", fmt.Sprintf("s3://%s/prefix/", rName)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAWSSagemakerDeviceFleet_description(t *testing.T) {
	var deviceFleet sagemaker.DescribeDeviceFleetOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_device_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerDeviceFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerDeviceFleetDescription(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerDeviceFleetExists(resourceName, &deviceFleet),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSagemakerDeviceFleetDescription(rName, "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerDeviceFleetExists(resourceName, &deviceFleet),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
				),
			},
		},
	})
}

func TestAccAWSSagemakerDeviceFleet_tags(t *testing.T) {
	var deviceFleet sagemaker.DescribeDeviceFleetOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_device_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerDeviceFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerDeviceFleetConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerDeviceFleetExists(resourceName, &deviceFleet),
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
				Config: testAccAWSSagemakerDeviceFleetConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerDeviceFleetExists(resourceName, &deviceFleet),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSSagemakerDeviceFleetConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerDeviceFleetExists(resourceName, &deviceFleet),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSSagemakerDeviceFleet_disappears(t *testing.T) {
	var deviceFleet sagemaker.DescribeDeviceFleetOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_device_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerDeviceFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerDeviceFleetBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerDeviceFleetExists(resourceName, &deviceFleet),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSagemakerDeviceFleet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSSagemakerDeviceFleetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_device_fleet" {
			continue
		}

		deviceFleet, err := finder.DeviceFleetByName(conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error reading Sagemaker Device Fleet (%s): %w", rs.Primary.ID, err)
		}

		if aws.StringValue(deviceFleet.DeviceFleetName) == rs.Primary.ID {
			return fmt.Errorf("sagemaker Device Fleet %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSSagemakerDeviceFleetExists(n string, device_fleet *sagemaker.DescribeDeviceFleetOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker Device Fleet ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn
		resp, err := finder.DeviceFleetByName(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*device_fleet = *resp

		return nil
	}
}

func testAccAWSSagemakerDeviceFleetConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  acl           = "private"
  force_destroy = true
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_iam_role_policy_attachment" "test2" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AWSIoTFullAccess"
}
`, rName)
}

func testAccAWSSagemakerDeviceFleetBasicConfig(rName string) string {
	return testAccAWSSagemakerDeviceFleetConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_device_fleet" "test" {
  device_fleet_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  output_config {
    s3_output_location = "s3://${aws_s3_bucket.test.bucket}/prefix/"
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy_attachment.test2]
}
`, rName)
}

func testAccAWSSagemakerDeviceFleetDescription(rName, desc string) string {
	return testAccAWSSagemakerDeviceFleetConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_device_fleet" "test" {
  device_fleet_name = %[1]q
  role_arn          = aws_iam_role.test.arn
  description       = %[2]q

  output_config {
    s3_output_location = "s3://${aws_s3_bucket.test.bucket}/prefix/"
  }
}
`, rName, desc)
}

func testAccAWSSagemakerDeviceFleetConfigTags1(rName, tagKey1, tagValue1 string) string {
	return testAccAWSSagemakerDeviceFleetConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_device_fleet" "test" {
  device_fleet_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  output_config {
    s3_output_location = "s3://${aws_s3_bucket.test.bucket}/prefix/"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSSagemakerDeviceFleetConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAWSSagemakerDeviceFleetConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_device_fleet" "test" {
  device_fleet_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  output_config {
    s3_output_location = "s3://${aws_s3_bucket.test.bucket}/prefix/"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
