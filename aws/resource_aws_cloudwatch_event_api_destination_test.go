package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const uuidRegex = "[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$"

func init() {
	resource.AddTestSweepers("aws_cloudwatch_event_api_destination", &resource.Sweeper{
		Name: "aws_cloudwatch_event_api_destination",
		F:    testSweepCloudWatchEventApiDestination,
		Dependencies: []string{
			"aws_cloudwatch_event_connection",
		},
	})
}

func testSweepCloudWatchEventApiDestination(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*AWSClient).cloudwatcheventsconn

	var sweeperErrs *multierror.Error

	input := &events.ListApiDestinationsInput{
		Limit: aws.Int64(100),
	}
	var apiDestinations []*events.ApiDestination
	for {
		output, err := conn.ListApiDestinations(input)
		if err != nil {
			return err
		}
		apiDestinations = append(apiDestinations, output.ApiDestinations...)

		if aws.StringValue(output.NextToken) == "" {
			break
		}
		input.NextToken = output.NextToken
	}

	for _, apiDestination := range apiDestinations {

		input := &events.DeleteApiDestinationInput{
			Name: apiDestination.Name,
		}
		_, err := conn.DeleteApiDestination(input)
		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("Error deleting CloudWatch Event Api Destination (%s): %w", *apiDestination.Name, err))
			continue
		}
	}

	log.Printf("[INFO] Deleted %d CloudWatch Event Api Destinations", len(apiDestinations))

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSCloudWatchEventApiDestination_basic(t *testing.T) {
	var v1, v2, v3 events.DescribeApiDestinationOutput
	name := acctest.RandomWithPrefix("tf-acc-test")
	invocationEndpoint := "https://www.hashicorp.com/"
	httpMethod := "GET"

	nameModified := acctest.RandomWithPrefix("tf-acc-test")
	invocationEndpointModified := "https://www.hashicorp.com/products/terraform"
	httpMethodModified := "POST"

	resourceName := "aws_cloudwatch_event_api_destination.basic"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, events.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventApiDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventApiDestinationConfig(
					name,
					invocationEndpoint,
					httpMethod,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventApiDestinationExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "events", regexp.MustCompile(fmt.Sprintf("api-destination/%s/%s", name, uuidRegex))),
					resource.TestCheckResourceAttr(resourceName, "http_method", httpMethod),
					resource.TestCheckResourceAttr(resourceName, "invocation_endpoint", invocationEndpoint),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudWatchEventApiDestinationConfig(
					nameModified,
					invocationEndpointModified,
					httpMethodModified,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventApiDestinationExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "name", nameModified),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "events", regexp.MustCompile(fmt.Sprintf("api-destination/%s/%s", nameModified, uuidRegex))),
					testAccCheckCloudWatchEventApiDestinationRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "http_method", httpMethodModified),
					resource.TestCheckResourceAttr(resourceName, "invocation_endpoint", invocationEndpointModified),
				),
			},
			{
				Config: testAccAWSCloudWatchEventApiDestinationConfig(
					nameModified,
					invocationEndpointModified,
					httpMethodModified,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventApiDestinationExists(resourceName, &v3),
					testAccCheckCloudWatchEventApiDestinationNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "http_method", httpMethodModified),
					resource.TestCheckResourceAttr(resourceName, "invocation_endpoint", invocationEndpointModified),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventApiDestination_optional(t *testing.T) {
	var v1, v2, v3 events.DescribeApiDestinationOutput
	name := acctest.RandomWithPrefix("tf-acc-test")
	invocationEndpoint := "https://www.hashicorp.com/"
	httpMethod := "GET"
	description := acctest.RandomWithPrefix("tf-acc-test")
	invocationRateLimitPerSecond := 10

	nameModified := acctest.RandomWithPrefix("tf-acc-test")
	invocationEndpointModified := "https://www.hashicorp.com/products/terraform"
	httpMethodModified := "POST"
	descriptionModified := acctest.RandomWithPrefix("tf-acc-test")
	invocationRateLimitPerSecondModified := 12

	resourceName := "aws_cloudwatch_event_api_destination.optional"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, events.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventApiDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventApiDestinationConfig_optional(
					name,
					invocationEndpoint,
					httpMethod,
					description,
					int64(invocationRateLimitPerSecond),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventApiDestinationExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "http_method", httpMethod),
					resource.TestCheckResourceAttr(resourceName, "invocation_endpoint", invocationEndpoint),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "invocation_rate_limit_per_second", fmt.Sprint(invocationRateLimitPerSecond)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudWatchEventApiDestinationConfig_optional(
					nameModified,
					invocationEndpointModified,
					httpMethodModified,
					descriptionModified,
					int64(invocationRateLimitPerSecondModified),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventApiDestinationExists(resourceName, &v2),
					testAccCheckCloudWatchEventApiDestinationRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "name", nameModified),
					resource.TestCheckResourceAttr(resourceName, "http_method", httpMethodModified),
					resource.TestCheckResourceAttr(resourceName, "invocation_endpoint", invocationEndpointModified),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionModified),
					resource.TestCheckResourceAttr(resourceName, "invocation_rate_limit_per_second", fmt.Sprint(invocationRateLimitPerSecondModified)),
				),
			},
			{
				Config: testAccAWSCloudWatchEventApiDestinationConfig_optional(
					nameModified,
					invocationEndpointModified,
					httpMethodModified,
					descriptionModified,
					int64(invocationRateLimitPerSecond),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventApiDestinationExists(resourceName, &v3),
					testAccCheckCloudWatchEventApiDestinationNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "name", nameModified),
					resource.TestCheckResourceAttr(resourceName, "http_method", httpMethodModified),
					resource.TestCheckResourceAttr(resourceName, "invocation_endpoint", invocationEndpointModified),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionModified),
					resource.TestCheckResourceAttr(resourceName, "invocation_rate_limit_per_second", fmt.Sprint(invocationRateLimitPerSecond)),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventApiDestination_disappears(t *testing.T) {
	var v events.DescribeApiDestinationOutput
	name := acctest.RandomWithPrefix("tf-acc-test")
	invocationEndpoint := "https://www.hashicorp.com/"
	httpMethod := "GET"

	resourceName := "aws_cloudwatch_event_api_destination.basic"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, events.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventApiDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventApiDestinationConfig(
					name,
					invocationEndpoint,
					httpMethod,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventApiDestinationExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudWatchEventApiDestination(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSCloudWatchEventApiDestinationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudwatcheventsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_event_api_destination" {
			continue
		}

		params := events.DescribeApiDestinationInput{
			Name: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeApiDestination(&params)

		if err == nil {
			return fmt.Errorf("CloudWatch Events Api Destination (%s) still exists: %s", rs.Primary.ID, resp)
		}
	}

	return nil
}

func testAccCheckCloudWatchEventApiDestinationExists(n string, v *events.DescribeApiDestinationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudwatcheventsconn
		params := events.DescribeApiDestinationInput{
			Name: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeApiDestination(&params)
		if err != nil {
			return err
		}
		if resp == nil {
			return fmt.Errorf("CloudWatch Events Api Destination (%s) not found", n)
		}

		*v = *resp

		return nil
	}
}

func testAccCheckCloudWatchEventApiDestinationRecreated(i, j *events.DescribeApiDestinationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.ApiDestinationArn) == aws.StringValue(j.ApiDestinationArn) {
			return fmt.Errorf("CloudWatch Events Api Destination not recreated")
		}
		return nil
	}
}

func testAccCheckCloudWatchEventApiDestinationNotRecreated(i, j *events.DescribeApiDestinationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.ApiDestinationArn) != aws.StringValue(j.ApiDestinationArn) {
			return fmt.Errorf("CloudWatch Events Api Destination was recreated")
		}
		return nil
	}
}

func testAccAWSCloudWatchEventApiDestinationConfig(name, invocationEndpoint, httpMethod string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_api_destination" "basic" {
  name                = %[1]q
  invocation_endpoint = %[2]q
  http_method         = %[3]q
  connection_arn      = aws_cloudwatch_event_connection.test.arn
}

resource "aws_cloudwatch_event_connection" "test" {
  name               = %[1]q
  authorization_type = "API_KEY"
  auth_parameters {
    api_key {
      key   = "testKey"
      value = "testValue"
    }
  }
}
`, name, invocationEndpoint, httpMethod)
}

func testAccAWSCloudWatchEventApiDestinationConfig_optional(name, invocationEndpoint, httpMethod, description string, invocationRateLimitPerSecond int64) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_api_destination" "optional" {
  name                = %[1]q
  invocation_endpoint = %[2]q
  http_method         = %[3]q
  connection_arn      = aws_cloudwatch_event_connection.test.arn

  description                      = %[4]q
  invocation_rate_limit_per_second = %[5]d
}

resource "aws_cloudwatch_event_connection" "test" {
  name               = %[1]q
  authorization_type = "API_KEY"
  auth_parameters {
    api_key {
      key   = "testKey"
      value = "testValue"
    }
  }
}
`, name, invocationEndpoint, httpMethod, description, invocationRateLimitPerSecond)
}
