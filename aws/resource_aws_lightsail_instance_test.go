package aws

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"testing"

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
	var instance lightsail.Instance
	resourceName := "aws_lightsail_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lightsail.EndpointsID, t)
			testAccPreCheckAWSLightsail(t)
		},
		ErrorCheck:   testAccErrorCheck(t, lightsail.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLightsailInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailInstanceConfigBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailInstanceExists(resourceName, &instance),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(resourceName, "blueprint_id"),
					resource.TestCheckResourceAttrSet(resourceName, "bundle_id"),
					resource.TestMatchResourceAttr(resourceName, "ipv6_address", regexp.MustCompile(`([a-f0-9]{1,4}:){7}[a-f0-9]{1,4}`)),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "key_pair_name"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestMatchResourceAttr(resourceName, "ram_size", regexp.MustCompile(`\d+(.\d+)?`)),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "dualstack"),
				),
			},
		},
	})
}

func TestAccAWSLightsailInstance_Name(t *testing.T) {
	var instance lightsail.Instance
	resourceName := "aws_lightsail_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rNameWithSpaces := fmt.Sprint(rName, "string with spaces")
	rNameWithStartingDigit := fmt.Sprintf("01-%s", rName)
	rNameWithUnderscore := fmt.Sprintf("%s_123456", rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lightsail.EndpointsID, t)
			testAccPreCheckAWSLightsail(t)
		},
		ErrorCheck:    testAccErrorCheck(t, lightsail.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLightsailInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSLightsailInstanceConfigBasic(rNameWithSpaces),
				ExpectError: regexp.MustCompile(`must contain only alphanumeric characters, underscores, hyphens, and dots`),
			},
			{
				Config:      testAccAWSLightsailInstanceConfigBasic(rNameWithStartingDigit),
				ExpectError: regexp.MustCompile(`must begin with an alphabetic character`),
			},
			{
				Config: testAccAWSLightsailInstanceConfigBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailInstanceExists(resourceName, &instance),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(resourceName, "blueprint_id"),
					resource.TestCheckResourceAttrSet(resourceName, "bundle_id"),
					resource.TestCheckResourceAttrSet(resourceName, "key_pair_name"),
				),
			},
			{
				Config: testAccAWSLightsailInstanceConfigBasic(rNameWithUnderscore),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailInstanceExists(resourceName, &instance),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(resourceName, "blueprint_id"),
					resource.TestCheckResourceAttrSet(resourceName, "bundle_id"),
					resource.TestCheckResourceAttrSet(resourceName, "key_pair_name"),
				),
			},
		},
	})
}

func TestAccAWSLightsailInstance_Tags(t *testing.T) {
	var instance1, instance2, instance3 lightsail.Instance
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lightsail.EndpointsID, t)
			testAccPreCheckAWSLightsail(t)
		},
		ErrorCheck:    testAccErrorCheck(t, lightsail.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLightsailInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailInstanceConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailInstanceExists(resourceName, &instance1),
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
				Config: testAccAWSLightsailInstanceConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailInstanceExists(resourceName, &instance2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSLightsailInstanceConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailInstanceExists(resourceName, &instance3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSLightsailInstance_disappears(t *testing.T) {
	var instance lightsail.Instance
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSLightsail(t) },
		ErrorCheck:   testAccErrorCheck(t, lightsail.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLightsailInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailInstanceConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailInstanceExists(resourceName, &instance),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsLightsailInstance(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSLightsailInstance_IpAddressType(t *testing.T) {
	var instance lightsail.Instance
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSLightsail(t) },
		ErrorCheck:   testAccErrorCheck(t, lightsail.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLightsailInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailInstanceConfigIpAddressType(rName, "ipv4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailInstanceExists(resourceName, &instance),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "ipv4"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLightsailInstanceConfigIpAddressType(rName, "dualstack"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailInstanceExists(resourceName, &instance),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "dualstack"),
				),
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

func testAccAWSLightsailInstanceConfigBasic(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_lightsail_instance" "test" {
  name              = %[1]q
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux"
  bundle_id         = "nano_1_0"
}
`, rName)
}

func testAccAWSLightsailInstanceConfigIpAddressType(rName string, rIpAddressType string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}
resource "aws_lightsail_instance" "test" {
  name              = %[1]q
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux"
  bundle_id         = "nano_1_0"
  ip_address_type   = %[2]q
}
`, rName, rIpAddressType)
}

func testAccAWSLightsailInstanceConfigTags1(rName string, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_lightsail_instance" "test" {
  name              = %[1]q
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux"
  bundle_id         = "nano_1_0"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSLightsailInstanceConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_lightsail_instance" "test" {
  name              = %[1]q
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux"
  bundle_id         = "nano_1_0"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
