package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_cloudwatch_event_archive", &resource.Sweeper{
		Name: "aws_cloudwatch_event_archive",
		F:    testSweepCloudWatchEventArchives,
		Dependencies: []string{
			"aws_cloudwatch_event_bus",
		},
	})
}

func testSweepCloudWatchEventArchives(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*AWSClient).cloudwatcheventsconn

	input := &events.ListArchivesInput{}

	for {
		output, err := conn.ListArchives(input)
		if err != nil {
			if testSweepSkipSweepError(err) {
				log.Printf("[WARN] Skipping CloudWatch Events archive sweep for %s: %s", region, err)
				return nil
			}
			return fmt.Errorf("Error retrieving CloudWatch Events archive: %w", err)
		}

		if len(output.Archives) == 0 {
			log.Print("[DEBUG] No CloudWatch Events archives to sweep")
			return nil
		}

		for _, archive := range output.Archives {
			name := aws.StringValue(archive.ArchiveName)
			if name == "default" {
				continue
			}

			log.Printf("[INFO] Deleting CloudWatch Events archive (%s)", name)
			_, err := conn.DeleteArchive(&events.DeleteArchiveInput{
				ArchiveName: aws.String(name),
			})
			if err != nil {
				return fmt.Errorf("Error deleting CloudWatch Events archive (%s): %w", name, err)
			}
		}

		if output.NextToken == nil {
			break
		}
		input.NextToken = output.NextToken
	}

	return nil
}

func TestAccAWSCloudWatchArchive_basic(t *testing.T) {
	var v1 events.DescribeArchiveOutput
	archiveName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudwatch_event_archive.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchArchiveDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchArchiveConfig(archiveName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchArchiveExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", archiveName),
					resource.TestCheckResourceAttr(resourceName, "retention_days", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "events", fmt.Sprintf("archive/%s", archiveName)),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "event_pattern", ""),
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

func TestAccAWSCloudWatchArchive_update(t *testing.T) {
	var v1 events.DescribeArchiveOutput
	resourceName := "aws_cloudwatch_event_archive.test"
	archiveName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchArchiveDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchArchiveConfig(archiveName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchArchiveExists(resourceName, &v1),
				),
			},
			{
				Config: testAccAWSCloudWatchArchiveConfig_updateAttributes(archiveName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchArchiveExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "retention_days", "7"),
					testAccCheckResourceAttrEquivalentJSON(resourceName, "event_pattern", "{\"source\":[\"company.team.service\"]}"),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventArchive_disappears(t *testing.T) {
	var v events.DescribeArchiveOutput
	archiveName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudwatch_event_archive.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventBusDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventBusConfig(archiveName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchArchiveExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudWatchEventArchive(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSCloudWatchArchiveDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudwatcheventsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_event_archive" {
			continue
		}

		params := events.DescribeArchiveInput{
			ArchiveName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeArchive(&params)

		if err == nil {
			return fmt.Errorf("CloudWatch Events event bus (%s) still exists: %s", rs.Primary.ID, resp)
		}
	}

	return nil
}

func testAccCheckCloudWatchArchiveExists(n string, v *events.DescribeArchiveOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudwatcheventsconn
		params := events.DescribeArchiveInput{
			ArchiveName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeArchive(&params)
		if err != nil {
			return err
		}

		if resp == nil {
			return fmt.Errorf("CloudWatch Events archive (%s) not found", n)
		}

		*v = *resp

		return nil
	}
}

func testAccAWSCloudWatchArchiveConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_event_archive" "test" {
  name             = %[1]q
  event_source_arn = aws_cloudwatch_event_bus.test.arn
}
`, name)
}

func testAccAWSCloudWatchArchiveConfig_updateAttributes(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_event_archive" "test" {
  name             = %[1]q
  event_source_arn = aws_cloudwatch_event_bus.test.arn
  retention_days   = 7
  description      = "test"
  event_pattern    = <<PATTERN
{
  "source": ["company.team.service"]
}
PATTERN
}
`, name)
}
