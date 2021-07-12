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
	resource.AddTestSweepers("aws_lightsail_load_balancer", &resource.Sweeper{
		Name: "aws_lightsail_load_balancer",
		F:    testSweepLightsailLoadBalancers,
	})
}

func testSweepLightsailLoadBalancers(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*AWSClient).lightsailconn

	input := &lightsail.GetLoadBalancersInput{}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.GetLoadBalancers(input)

		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Lightsail Load Balancer sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving Lightsail Load Balancers: %s", err)
		}

		for _, loadBalancer := range output.LoadBalancers {
			name := aws.StringValue(loadBalancer.Name)
			input := &lightsail.DeleteLoadBalancerInput{
				LoadBalancerName: loadBalancer.Name,
			}

			log.Printf("[INFO] Deleting Lightsail Load Balancer: %s", name)
			_, err := conn.DeleteLoadBalancer(input)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Lightsail Load Balancer (%s): %s", name, err)
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

func TestAccAWSLightsailLoadBalancer_basic(t *testing.T) {
	var lb lightsail.LoadBalancer
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_load_balancer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t); testAccPreCheckAWSLightsail(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLightsailLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailLoadBalancerConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailLoadBalancerExists(resourceName, &lb),
					resource.TestCheckResourceAttr(resourceName, "health_check_path", "/"),
					resource.TestCheckResourceAttr(resourceName, "instance_port", "80"),
					resource.TestCheckResourceAttrSet(resourceName, "dns_name"),
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

func TestAccAWSLightsailLoadBalancer_Name(t *testing.T) {
	var lb lightsail.LoadBalancer
	rName := acctest.RandomWithPrefix("tf-acc-test")
	lightsailNameWithSpaces := fmt.Sprint(rName, "string with spaces")
	lightsailNameWithStartingDigit := fmt.Sprintf("01-%s", rName)
	lightsailNameWithUnderscore := fmt.Sprintf("%s_123456", rName)
	resourceName := "aws_lightsail_load_balancer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLightsailLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSLightsailLoadBalancerConfigBasic(lightsailNameWithSpaces),
				ExpectError: regexp.MustCompile(`must contain only alphanumeric characters, underscores, hyphens, and dots`),
			},
			{
				Config:      testAccAWSLightsailLoadBalancerConfigBasic(lightsailNameWithStartingDigit),
				ExpectError: regexp.MustCompile(`must begin with an alphabetic character`),
			},
			{
				Config: testAccAWSLightsailLoadBalancerConfigBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailLoadBalancerExists(resourceName, &lb),
					resource.TestCheckResourceAttrSet(resourceName, "health_check_path"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_port"),
				),
			},
			{
				Config: testAccAWSLightsailLoadBalancerConfigBasic(lightsailNameWithUnderscore),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailLoadBalancerExists(resourceName, &lb),
					resource.TestCheckResourceAttrSet(resourceName, "health_check_path"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_port"),
				),
			},
		},
	})
}

func TestAccAWSLightsailLoadBalancer_HealthCheckPath(t *testing.T) {
	var lb lightsail.LoadBalancer
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_load_balancer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSLightsail(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLightsailLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailLoadBalancerConfigHealthCheckPath(rName, "/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailLoadBalancerExists(resourceName, &lb),
					resource.TestCheckResourceAttr(resourceName, "health_check_path", "/"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLightsailLoadBalancerConfigHealthCheckPath(rName, "/healthcheck"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailLoadBalancerExists(resourceName, &lb),
					resource.TestCheckResourceAttr(resourceName, "health_check_path", "/healthcheck"),
				),
			},
		},
	})
}

func TestAccAWSLightsailLoadBalancer_Tags(t *testing.T) {
	var lb1, lb2, lb3 lightsail.LoadBalancer
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_load_balancer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t); testAccPreCheckAWSLightsail(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLightsailLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailLoadBalancerConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailLoadBalancerExists(resourceName, &lb1),
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
				Config: testAccAWSLightsailLoadBalancerConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailLoadBalancerExists(resourceName, &lb2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSLightsailLoadBalancerConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailLoadBalancerExists(resourceName, &lb3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAWSLightsailLoadBalancerExists(n string, res *lightsail.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No LightsailLoadBalancer ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).lightsailconn

		respLoadBalancer, err := conn.GetLoadBalancer(&lightsail.GetLoadBalancerInput{
			LoadBalancerName: aws.String(rs.Primary.Attributes["name"]),
		})

		if err != nil {
			return err
		}

		if respLoadBalancer == nil || respLoadBalancer.LoadBalancer == nil {
			return fmt.Errorf("Load Balancer (%s) not found", rs.Primary.Attributes["name"])
		}
		*res = *respLoadBalancer.LoadBalancer
		return nil
	}
}

func TestAccAWSLightsailLoadBalancer_disappears(t *testing.T) {
	var lb lightsail.LoadBalancer
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_load_balancer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSLightsail(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLightsailLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailLoadBalancerConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailLoadBalancerExists(resourceName, &lb),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsLightsailLoadBalancer(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSLightsailLoadBalancerDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lightsail_load_balancer" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).lightsailconn

		resp, err := conn.GetLoadBalancer(&lightsail.GetLoadBalancerInput{
			LoadBalancerName: aws.String(rs.Primary.ID),
		})

		if err == nil {
			if resp.LoadBalancer != nil {
				return fmt.Errorf("Lightsail Load Balancer %q still exists", rs.Primary.ID)
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

func testAccAWSLightsailLoadBalancerConfigBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_load_balancer" "test" {
  name              = %[1]q
  health_check_path = "/"
  instance_port     = "80"
}
`, rName)
}

func testAccAWSLightsailLoadBalancerConfigHealthCheckPath(rName string, rPath string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_load_balancer" "test" {
  name              = %[1]q
  health_check_path = %[2]q
  instance_port     = "80"
}
`, rName, rPath)
}

func testAccAWSLightsailLoadBalancerConfigTags1(rName string, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_load_balancer" "test" {
  name              = %[1]q
  health_check_path = "/"
  instance_port     = "80"
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSLightsailLoadBalancerConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_load_balancer" "test" {
  name              = %[1]q
  health_check_path = "/"
  instance_port     = "80"
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
