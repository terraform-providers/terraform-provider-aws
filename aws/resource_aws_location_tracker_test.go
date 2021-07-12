package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/locationservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsLocationTracker_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_location_tracker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckLocationService(t) },
		ErrorCheck:   testAccErrorCheck(t, locationservice.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLocationTrackerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsLocationTrackerConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLocationTrackerExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "geo", regexp.MustCompile(`tracker/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "id", rName),
					resource.TestCheckResourceAttr(resourceName, "pricing_plan", "RequestBasedUsage"),
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

func TestAccAwsLocationTracker_disappers(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_location_tracker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckLocationService(t) },
		ErrorCheck:   testAccErrorCheck(t, locationservice.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLocationTrackerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsLocationTrackerConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLocationTrackerExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsLocationTracker(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsLocationTracker_Description(t *testing.T) {
	var v1, v2 locationservice.DescribeTrackerOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_location_tracker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckLocationService(t) },
		ErrorCheck:   testAccErrorCheck(t, locationservice.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLocationTrackerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsLocationTrackerConfigDescription(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLocationTrackerExistsWithDescribeTrackerOutput(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				Config: testAccAwsLocationTrackerConfigDescription(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLocationTrackerExistsWithDescribeTrackerOutput(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
					testAccCheckAwsLocationTrackerRecreated("description", &v1, &v2),
				),
			},
		},
	})
}

func TestAccAwsLocationTracker_Tags(t *testing.T) {
	var v1, v2, v3 locationservice.DescribeTrackerOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_location_tracker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckLocationService(t) },
		ErrorCheck:   testAccErrorCheck(t, locationservice.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLocationTrackerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsLocationTrackerConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLocationTrackerExistsWithDescribeTrackerOutput(resourceName, &v1),
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
				Config: testAccAwsLocationTrackerConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLocationTrackerExistsWithDescribeTrackerOutput(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAwsLocationTrackerConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLocationTrackerExistsWithDescribeTrackerOutput(resourceName, &v3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAwsLocationTrackerExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).locationconn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		_, err := conn.DescribeTracker(&locationservice.DescribeTrackerInput{
			TrackerName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckAwsLocationTrackerExistsWithDescribeTrackerOutput(name string, res *locationservice.DescribeTrackerOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).locationconn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		output, err := conn.DescribeTracker(&locationservice.DescribeTrackerInput{
			TrackerName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*res = *output

		return nil
	}
}

func testAccCheckAwsLocationTrackerRecreated(attributeName string, v1, v2 *locationservice.DescribeTrackerOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if v2.CreateTime.Equal(*v1.CreateTime) {
			return fmt.Errorf("Location tracker not recreated when changing %s", attributeName)
		}

		return nil
	}
}

func testAccCheckLocationTrackerDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).locationconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_acm_certificate" {
			continue
		}
		_, err := conn.DescribeTracker(&locationservice.DescribeTrackerInput{
			TrackerName: aws.String(rs.Primary.ID),
		})

		if err == nil {
			return fmt.Errorf("Tracker still exists.")
		}

		// Verify the error is what we want
		if !isAWSErr(err, locationservice.ErrCodeResourceNotFoundException, "") {
			return err
		}
	}

	return nil
}

func testAccPreCheckLocationService(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).locationconn
	input := &locationservice.ListTrackersInput{}
	_, err := conn.ListTrackers(input)
	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAwsLocationTrackerConfigName(rName string) string {
	return fmt.Sprintf(`
resource "aws_location_tracker" "test" {
  name         = %[1]q
  pricing_plan = "RequestBasedUsage"
}
`, rName)
}

func testAccAwsLocationTrackerConfigDescription(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_location_tracker" "test" {
  name         = %[1]q
  description  = %[2]q
  pricing_plan = "RequestBasedUsage"
}
`, rName, description)
}

func testAccAwsLocationTrackerConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_location_tracker" "test" {
  name         = %[1]q
  pricing_plan = "RequestBasedUsage"
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAwsLocationTrackerConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_location_tracker" "test" {
  name         = %[1]q
  pricing_plan = "RequestBasedUsage"
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
