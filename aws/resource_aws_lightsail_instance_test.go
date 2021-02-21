package aws

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_lightsail_instance", &resource.Sweeper{
		Name: "aws_lightsail_instance",
		F:    testSweepLightsailInstances,
	})
}

func testSweepLightsailInstances(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*AWSClient).lightsailconn

	input := &lightsail.GetInstancesInput{}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.GetInstances(input)

		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Lightsail Instance sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving Lightsail Instances: %s", err)
		}

		for _, instance := range output.Instances {
			name := aws.StringValue(instance.Name)
			input := &lightsail.DeleteInstanceInput{
				InstanceName: instance.Name,
			}

			log.Printf("[INFO] Deleting Lightsail Instance: %s", name)
			_, err := conn.DeleteInstance(input)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Lightsail Instance (%s): %s", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			}
		}

		if aws.StringValue(output.NextPageToken) == "" {
			break
		}

		input.PageToken = output.NextPageToken
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSLightsailInstance_basic(t *testing.T) {
	var conf lightsail.Instance
	lightsailName := fmt.Sprintf("tf-test-lightsail-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lightsail.EndpointsID, t)
			testAccPreCheckAWSLightsail(t)
		},
		IDRefreshName: "aws_lightsail_instance.lightsail_instance_test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLightsailInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailInstanceConfig_basic(lightsailName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailInstanceExists("aws_lightsail_instance.lightsail_instance_test", &conf),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "availability_zone"),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "blueprint_id"),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "bundle_id"),
					resource.TestMatchResourceAttr("aws_lightsail_instance.lightsail_instance_test", "ipv6_address", regexp.MustCompile(`([a-f0-9]{1,4}:){7}[a-f0-9]{1,4}`)),
					resource.TestCheckResourceAttr("aws_lightsail_instance.lightsail_instance_test", "ipv6_addresses.#", "1"),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "key_pair_name"),
					resource.TestCheckResourceAttr("aws_lightsail_instance.lightsail_instance_test", "tags.%", "0"),
					resource.TestMatchResourceAttr("aws_lightsail_instance.lightsail_instance_test", "ram_size", regexp.MustCompile(`\d+(.\d+)?`)),
				),
			},
		},
	})
}

func TestAccAWSLightsailInstance_Name(t *testing.T) {
	var conf lightsail.Instance
	lightsailName := fmt.Sprintf("tf-test-lightsail-%d", acctest.RandInt())
	lightsailNameWithSpaces := fmt.Sprint(lightsailName, "string with spaces")
	lightsailNameWithStartingDigit := fmt.Sprintf("01-%s", lightsailName)
	lightsailNameWithUnderscore := fmt.Sprintf("%s_123456", lightsailName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lightsail.EndpointsID, t)
			testAccPreCheckAWSLightsail(t)
		},
		IDRefreshName: "aws_lightsail_instance.lightsail_instance_test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLightsailInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSLightsailInstanceConfig_basic(lightsailNameWithSpaces),
				ExpectError: regexp.MustCompile(`must contain only alphanumeric characters, underscores, hyphens, and dots`),
			},
			{
				Config:      testAccAWSLightsailInstanceConfig_basic(lightsailNameWithStartingDigit),
				ExpectError: regexp.MustCompile(`must begin with an alphabetic character`),
			},
			{
				Config: testAccAWSLightsailInstanceConfig_basic(lightsailName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailInstanceExists("aws_lightsail_instance.lightsail_instance_test", &conf),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "availability_zone"),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "blueprint_id"),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "bundle_id"),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "key_pair_name"),
				),
			},
			{
				Config: testAccAWSLightsailInstanceConfig_basic(lightsailNameWithUnderscore),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailInstanceExists("aws_lightsail_instance.lightsail_instance_test", &conf),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "availability_zone"),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "blueprint_id"),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "bundle_id"),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "key_pair_name"),
				),
			},
		},
	})
}

func TestAccAWSLightsailInstance_Tags(t *testing.T) {
	var conf lightsail.Instance
	lightsailName := fmt.Sprintf("tf-test-lightsail-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lightsail.EndpointsID, t)
			testAccPreCheckAWSLightsail(t)
		},
		IDRefreshName: "aws_lightsail_instance.lightsail_instance_test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLightsailInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailInstanceConfig_tags1(lightsailName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailInstanceExists("aws_lightsail_instance.lightsail_instance_test", &conf),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "availability_zone"),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "blueprint_id"),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "bundle_id"),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "key_pair_name"),
					resource.TestCheckResourceAttr("aws_lightsail_instance.lightsail_instance_test", "tags.%", "2"),
				),
			},
			{
				Config: testAccAWSLightsailInstanceConfig_tags2(lightsailName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailInstanceExists("aws_lightsail_instance.lightsail_instance_test", &conf),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "availability_zone"),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "blueprint_id"),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "bundle_id"),
					resource.TestCheckResourceAttrSet("aws_lightsail_instance.lightsail_instance_test", "key_pair_name"),
					resource.TestCheckResourceAttr("aws_lightsail_instance.lightsail_instance_test", "tags.%", "3"),
				),
			},
		},
	})
}

func TestAccAWSLightsailInstance_disapear(t *testing.T) {
	var conf lightsail.Instance
	lightsailName := fmt.Sprintf("tf-test-lightsail-%d", acctest.RandInt())

	testDestroy := func(*terraform.State) error {
		// reach out and DELETE the Instance
		conn := testAccProvider.Meta().(*AWSClient).lightsailconn
		_, err := conn.DeleteInstance(&lightsail.DeleteInstanceInput{
			InstanceName: aws.String(lightsailName),
		})

		if err != nil {
			return fmt.Errorf("error deleting Lightsail Instance in disappear test")
		}

		// sleep 7 seconds to give it time, so we don't have to poll
		time.Sleep(7 * time.Second)

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lightsail.EndpointsID, t)
			testAccPreCheckAWSLightsail(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLightsailInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailInstanceConfig_basic(lightsailName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailInstanceExists("aws_lightsail_instance.lightsail_instance_test", &conf),
					testDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSLightsailInstanceExists(n string, res *lightsail.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No LightsailInstance ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).lightsailconn

		respInstance, err := conn.GetInstance(&lightsail.GetInstanceInput{
			InstanceName: aws.String(rs.Primary.Attributes["name"]),
		})

		if err != nil {
			return err
		}

		if respInstance == nil || respInstance.Instance == nil {
			return fmt.Errorf("Instance (%s) not found", rs.Primary.Attributes["name"])
		}
		*res = *respInstance.Instance
		return nil
	}
}

func testAccCheckAWSLightsailInstanceDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lightsail_instance" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).lightsailconn

		respInstance, err := conn.GetInstance(&lightsail.GetInstanceInput{
			InstanceName: aws.String(rs.Primary.Attributes["name"]),
		})

		if err == nil {
			if respInstance.Instance != nil {
				return fmt.Errorf("LightsailInstance %q still exists", rs.Primary.ID)
			}
		}

		// Verify the error
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "NotFoundException" {
				return nil
			}
		}
		return err
	}

	return nil
}

func testAccPreCheckAWSLightsail(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).lightsailconn

	input := &lightsail.GetInstancesInput{}

	_, err := conn.GetInstances(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSLightsailInstanceConfig_basic(lightsailName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_lightsail_instance" "lightsail_instance_test" {
  name              = "%s"
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux"
  bundle_id         = "nano_1_0"
}
`, lightsailName)
}

func testAccAWSLightsailInstanceConfig_tags1(lightsailName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_lightsail_instance" "lightsail_instance_test" {
  name              = "%s"
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux"
  bundle_id         = "nano_1_0"

  tags = {
    Name       = "tf-test"
    KeyOnlyTag = ""
  }
}
`, lightsailName)
}

func testAccAWSLightsailInstanceConfig_tags2(lightsailName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_lightsail_instance" "lightsail_instance_test" {
  name              = "%s"
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux"
  bundle_id         = "nano_1_0"

  tags = {
    Name       = "tf-test",
    KeyOnlyTag = ""
    ExtraName  = "tf-test"
  }
}
`, lightsailName)
}
