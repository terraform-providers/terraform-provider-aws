package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_cloudwatch_event_bus", &resource.Sweeper{
		Name: "aws_cloudwatch_event_bus",
		F:    testSweepCloudWatchEventBuses,
		Dependencies: []string{
			"aws_cloudwatch_event_rule",
			"aws_cloudwatch_event_target",
		},
	})
}

func testSweepCloudWatchEventBuses(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*AWSClient).cloudwatcheventsconn

	input := &events.ListEventBusesInput{}

	for {
		output, err := conn.ListEventBuses(input)
		if err != nil {
			if testSweepSkipSweepError(err) {
				log.Printf("[WARN] Skipping CloudWatch Events event bus sweep for %s: %s", region, err)
				return nil
			}
			return fmt.Errorf("Error retrieving CloudWatch Events event bus: %w", err)
		}

		if len(output.EventBuses) == 0 {
			log.Print("[DEBUG] No CloudWatch Events event buses to sweep")
			return nil
		}

		for _, eventBus := range output.EventBuses {
			name := aws.StringValue(eventBus.Name)
			if name == "default" {
				continue
			}

			log.Printf("[INFO] Deleting CloudWatch Events event bus (%s)", name)
			_, err := conn.DeleteEventBus(&events.DeleteEventBusInput{
				Name: aws.String(name),
			})
			if err != nil {
				return fmt.Errorf("Error deleting CloudWatch Events event bus (%s): %w", name, err)
			}
		}

		if output.NextToken == nil {
			break
		}
		input.NextToken = output.NextToken
	}

	return nil
}

func TestAccAWSCloudWatchEventBus_basic(t *testing.T) {
	var v1, v2, v3 events.DescribeEventBusOutput
	busName := acctest.RandomWithPrefix("tf-acc-test")
	busNameModified := acctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_cloudwatch_event_bus.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventBusDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventBusConfig(busName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventBusExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", busName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "events", fmt.Sprintf("event-bus/%s", busName)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudWatchEventBusConfig(busNameModified),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventBusExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "name", busNameModified),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "events", fmt.Sprintf("event-bus/%s", busNameModified)),
					testAccCheckCloudWatchEventBusRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccAWSCloudWatchEventBusConfig_Tags1(busNameModified, "key", "value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventBusExists(resourceName, &v3),
					testAccCheckCloudWatchEventBusNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key", "value"),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventBus_tags(t *testing.T) {
	var v1, v2, v3, v4 events.DescribeEventBusOutput
	busName := acctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_cloudwatch_event_bus.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventBusDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventBusConfig_Tags1(busName, "key1", "value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventBusExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudWatchEventBusConfig_Tags2(busName, "key1", "updated", "key2", "added"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventBusExists(resourceName, &v2),
					testAccCheckCloudWatchEventBusNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "added"),
				),
			},
			{
				Config: testAccAWSCloudWatchEventBusConfig_Tags1(busName, "key2", "added"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventBusExists(resourceName, &v3),
					testAccCheckCloudWatchEventBusNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "added"),
				),
			},
			{
				Config: testAccAWSCloudWatchEventBusConfig(busName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventBusExists(resourceName, &v4),
					testAccCheckCloudWatchEventBusNotRecreated(&v3, &v4),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventBus_default(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventBusDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSCloudWatchEventBusConfig("default"),
				ExpectError: regexp.MustCompile(`cannot be 'default'`),
			},
		},
	})
}

func TestAccAWSCloudWatchEventBus_disappears(t *testing.T) {
	var v events.DescribeEventBusOutput
	busName := acctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_cloudwatch_event_bus.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventBusDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventBusConfig(busName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventBusExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudWatchEventBus(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSCloudWatchEventBusDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudwatcheventsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_event_bus" {
			continue
		}

		params := events.DescribeEventBusInput{
			Name: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeEventBus(&params)

		if err == nil {
			return fmt.Errorf("CloudWatch Events event bus (%s) still exists: %s", rs.Primary.ID, resp)
		}
	}

	return nil
}

func testAccCheckCloudWatchEventBusExists(n string, v *events.DescribeEventBusOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudwatcheventsconn
		params := events.DescribeEventBusInput{
			Name: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeEventBus(&params)
		if err != nil {
			return err
		}
		if resp == nil {
			return fmt.Errorf("CloudWatch Events event bus (%s) not found", n)
		}

		*v = *resp

		return nil
	}
}

func testAccCheckCloudWatchEventBusRecreated(i, j *events.DescribeEventBusOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.Arn) == aws.StringValue(j.Arn) {
			return fmt.Errorf("CloudWatch Events event bus not recreated")
		}
		return nil
	}
}

func testAccCheckCloudWatchEventBusNotRecreated(i, j *events.DescribeEventBusOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.Arn) != aws.StringValue(j.Arn) {
			return fmt.Errorf("CloudWatch Events event bus was recreated")
		}
		return nil
	}
}

func testAccAWSCloudWatchEventBusConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}
`, name)
}

func testAccAWSCloudWatchEventBusConfig_Tags1(name, key, value string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, name, key, value)
}

func testAccAWSCloudWatchEventBusConfig_Tags2(name, key1, value1, key2, value2 string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, name, key1, value1, key2, value2)
}
