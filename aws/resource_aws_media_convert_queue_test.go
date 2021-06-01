package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediaconvert"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSMediaConvertQueue_basic(t *testing.T) {
	var queue mediaconvert.Queue
	resourceName := "aws_media_convert_queue.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMediaConvert(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMediaConvertQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMediaConvertQueueConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaConvertQueueExists(resourceName, &queue),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "mediaconvert", regexp.MustCompile(`queues/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "pricing_plan", mediaconvert.PricingPlanOnDemand),
					resource.TestCheckResourceAttr(resourceName, "status", mediaconvert.QueueStatusActive),
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

func TestAccAWSMediaConvertQueue_ReservationPlanSettings(t *testing.T) {
	TestAccSkip(t, "MediaConvert Reserved Queues are $400/month and cannot be deleted for 1 year.")

	var queue mediaconvert.Queue
	resourceName := "aws_media_convert_queue.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMediaConvert(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMediaConvertQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMediaConvertQueueConfig_ReservedQueue(rName, mediaconvert.CommitmentOneYear, mediaconvert.RenewalTypeAutoRenew, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaConvertQueueExists(resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, "pricing_plan", mediaconvert.PricingPlanReserved),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.0.commitment", mediaconvert.CommitmentOneYear),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.0.renewal_type", mediaconvert.RenewalTypeAutoRenew),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.0.reserved_slots", "1"),
				),
			},
			{
				Config: testAccMediaConvertQueueConfig_ReservedQueue(rName, mediaconvert.CommitmentOneYear, mediaconvert.RenewalTypeExpire, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaConvertQueueExists(resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, "pricing_plan", mediaconvert.PricingPlanReserved),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.0.commitment", mediaconvert.CommitmentOneYear),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.0.renewal_type", mediaconvert.RenewalTypeExpire),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.0.reserved_slots", "1"),
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

func TestAccAWSMediaConvertQueue_withStatus(t *testing.T) {
	var queue mediaconvert.Queue
	resourceName := "aws_media_convert_queue.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMediaConvert(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMediaConvertQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMediaConvertQueueConfig_withStatus(rName, mediaconvert.QueueStatusPaused),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaConvertQueueExists(resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, "status", mediaconvert.QueueStatusPaused),
				),
			},
			{
				Config: testAccMediaConvertQueueConfig_withStatus(rName, mediaconvert.QueueStatusActive),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaConvertQueueExists(resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, "status", mediaconvert.QueueStatusActive),
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

func TestAccAWSMediaConvertQueue_withTags(t *testing.T) {
	var queue mediaconvert.Queue
	resourceName := "aws_media_convert_queue.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMediaConvert(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMediaConvertQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMediaConvertQueueConfig_withTags(rName, "foo", "bar", "fizz", "buzz"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaConvertQueueExists(resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.fizz", "buzz"),
				),
			},
			{
				Config: testAccMediaConvertQueueConfig_withTags(rName, "foo", "bar2", "fizz2", "buzz2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaConvertQueueExists(resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar2"),
					resource.TestCheckResourceAttr(resourceName, "tags.fizz2", "buzz2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMediaConvertQueueConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaConvertQueueExists(resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSMediaConvertQueue_disappears(t *testing.T) {
	var queue mediaconvert.Queue
	resourceName := "aws_media_convert_queue.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMediaConvert(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMediaConvertQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMediaConvertQueueConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaConvertQueueExists(resourceName, &queue),
					testAccCheckAwsMediaConvertQueueDisappears(&queue),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSMediaConvertQueue_withDescription(t *testing.T) {
	var queue mediaconvert.Queue
	resourceName := "aws_media_convert_queue.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	description1 := acctest.RandomWithPrefix("Description: ")
	description2 := acctest.RandomWithPrefix("Description: ")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMediaConvert(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMediaConvertQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMediaConvertQueueConfig_withDescription(rName, description1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaConvertQueueExists(resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, "description", description1),
				),
			},
			{
				Config: testAccMediaConvertQueueConfig_withDescription(rName, description2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaConvertQueueExists(resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, "description", description2),
				),
			},
		},
	})
}

func testAccCheckAwsMediaConvertQueueDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_media_convert_queue" {
			continue
		}
		conn, err := getAwsMediaConvertAccountClient(testAccProvider.Meta().(*AWSClient))
		if err != nil {
			return fmt.Errorf("Error getting Media Convert Account Client: %s", err)
		}

		_, err = conn.GetQueue(&mediaconvert.GetQueueInput{
			Name: aws.String(rs.Primary.ID),
		})
		if err != nil {
			if isAWSErr(err, mediaconvert.ErrCodeNotFoundException, "") {
				continue
			}
			return err
		}
	}

	return nil
}

func testAccCheckAwsMediaConvertQueueDisappears(queue *mediaconvert.Queue) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn, err := getAwsMediaConvertAccountClient(testAccProvider.Meta().(*AWSClient))
		if err != nil {
			return fmt.Errorf("Error getting Media Convert Account Client: %s", err)
		}

		_, err = conn.DeleteQueue(&mediaconvert.DeleteQueueInput{
			Name: queue.Name,
		})
		if err != nil {
			return fmt.Errorf("Deleting Media Convert Queue: %s", err)
		}
		return nil
	}
}

func testAccCheckAwsMediaConvertQueueExists(n string, queue *mediaconvert.Queue) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Queue id is set")
		}

		conn, err := getAwsMediaConvertAccountClient(testAccProvider.Meta().(*AWSClient))
		if err != nil {
			return fmt.Errorf("Error getting Media Convert Account Client: %s", err)
		}

		resp, err := conn.GetQueue(&mediaconvert.GetQueueInput{
			Name: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return fmt.Errorf("Error getting queue: %s", err)
		}

		*queue = *resp.Queue
		return nil
	}
}

func testAccPreCheckAWSMediaConvert(t *testing.T) {
	_, err := getAwsMediaConvertAccountClient(testAccProvider.Meta().(*AWSClient))

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccMediaConvertQueueConfig_Basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_media_convert_queue" "test" {
  name = %[1]q
}
`, rName)
}

func testAccMediaConvertQueueConfig_withStatus(rName, status string) string {
	return fmt.Sprintf(`
resource "aws_media_convert_queue" "test" {
  name   = %[1]q
  status = %[2]q
}
`, rName, status)
}

func testAccMediaConvertQueueConfig_withDescription(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_media_convert_queue" "test" {
  name        = %[1]q
  description = %[2]q
}
`, rName, description)
}

func testAccMediaConvertQueueConfig_withTags(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_media_convert_queue" "test" {
  name = %[1]q

  tags = {
    %[2]s = %[3]q
    %[4]s = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccMediaConvertQueueConfig_ReservedQueue(rName, commitment, renewalType string, reservedSlots int) string {
	return fmt.Sprintf(`
resource "aws_media_convert_queue" "test" {
  name         = %[1]q
  pricing_plan = %[2]q

  reservation_plan_settings {
    commitment     = %[3]q
    renewal_type   = %[4]q
    reserved_slots = %[5]d
  }
}
`, rName, mediaconvert.PricingPlanReserved, commitment, renewalType, reservedSlots)
}
