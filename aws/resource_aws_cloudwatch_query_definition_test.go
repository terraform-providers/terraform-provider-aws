package aws

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudwatchlogs/finder"
)

func init() {
	resource.AddTestSweepers("aws_cloudwatch_query_definition", &resource.Sweeper{
		Name: "aws_cloudwatch_query_definition",
		F:    testSweepCloudwatchlogQueryDefinitions,
	})
}

func testSweepCloudwatchlogQueryDefinitions(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*AWSClient).cloudwatchlogsconn
	sweepResources := make([]*testSweepResource, 0)
	var errs *multierror.Error

	input := &cloudwatchlogs.DescribeQueryDefinitionsInput{}

	// AWS SDK Go does not currently provide paginator
	for {
		output, err := conn.DescribeQueryDefinitions(input)

		if err != nil {
			err := fmt.Errorf("error reading CloudWatch Log Query Definition: %w", err)
			log.Printf("[ERROR] %s", err)
			errs = multierror.Append(errs, err)
			break
		}

		for _, queryDefinition := range output.QueryDefinitions {
			r := resourceAwsCloudWatchQueryDefinition()
			d := r.Data(nil)

			d.SetId(aws.StringValue(queryDefinition.QueryDefinitionId))

			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	if err := testSweepResourceOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping CloudWatch Log Query Definition for %s: %w", region, err))
	}

	if testSweepSkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping CloudWatch Log Query Definition sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAWSCloudWatchQueryDefinition_basic(t *testing.T) {
	var v cloudwatchlogs.QueryDefinition
	resourceName := "aws_cloudwatch_query_definition.test"
	queryName := acctest.RandomWithPrefix("tf-acc-test")

	expectedQueryString := `fields @timestamp, @message
| sort @timestamp desc
| limit 20
`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSCloudWatchQueryDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchQueryDefinitionConfig_Basic(queryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudWatchQueryDefinitionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", queryName),
					resource.TestCheckResourceAttr(resourceName, "query_string", expectedQueryString),
					resource.TestCheckResourceAttr(resourceName, "log_group_names.#", "0"),
					resource.TestMatchResourceAttr(resourceName, "query_definition_id", regexp.MustCompile(uuidRegexPattern)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAWSCloudWatchQueryDefinitionImportStateId(&v),
			},
		},
	})
}

func testAccAWSCloudWatchQueryDefinitionImportStateId(v *cloudwatchlogs.QueryDefinition) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		id := arn.ARN{
			AccountID: testAccGetAccountID(),
			Partition: testAccGetPartition(),
			Region:    testAccGetRegion(),
			Service:   cloudwatchlogs.ServiceName,
			Resource:  fmt.Sprintf("query-definition:%s", aws.StringValue(v.QueryDefinitionId)),
		}

		return id.String(), nil
	}
}

func TestAccAWSCloudWatchQueryDefinition_disappears(t *testing.T) {
	var v cloudwatchlogs.QueryDefinition
	resourceName := "aws_cloudwatch_query_definition.test"
	queryName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, cloudwatchlogs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchQueryDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchQueryDefinitionConfig_Basic(queryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudWatchQueryDefinitionExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudWatchQueryDefinition(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSCloudWatchQueryDefinition_Rename(t *testing.T) {
	var v1, v2 cloudwatchlogs.QueryDefinition
	resourceName := "aws_cloudwatch_query_definition.test"
	queryName := acctest.RandomWithPrefix("tf-acc-test")
	updatedQueryName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSCloudWatchQueryDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchQueryDefinitionConfig_Basic(queryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudWatchQueryDefinitionExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", queryName),
				),
			},
			{
				Config: testAccAWSCloudWatchQueryDefinitionConfig_Basic(updatedQueryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudWatchQueryDefinitionExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "name", updatedQueryName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAWSCloudWatchQueryDefinitionImportStateId(&v2),
			},
		},
	})
}

func TestAccAWSCloudWatchQueryDefinition_LogGroups(t *testing.T) {
	var v1, v2 cloudwatchlogs.QueryDefinition
	resourceName := "aws_cloudwatch_query_definition.test"
	queryName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSCloudWatchQueryDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchQueryDefinitionConfig_LogGroups(queryName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudWatchQueryDefinitionExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", queryName),
					resource.TestCheckResourceAttr(resourceName, "log_group_names.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "log_group_names.0", "aws_cloudwatch_log_group.test.0", "name"),
				),
			},
			{
				Config: testAccAWSCloudWatchQueryDefinitionConfig_LogGroups(queryName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudWatchQueryDefinitionExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "name", queryName),
					resource.TestCheckResourceAttr(resourceName, "log_group_names.#", "5"),
					resource.TestCheckResourceAttrPair(resourceName, "log_group_names.0", "aws_cloudwatch_log_group.test.0", "name"),
					resource.TestCheckResourceAttrPair(resourceName, "log_group_names.1", "aws_cloudwatch_log_group.test.1", "name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAWSCloudWatchQueryDefinitionImportStateId(&v2),
			},
		},
	})
}

func testAccCheckAWSCloudWatchQueryDefinitionExists(rName string, v *cloudwatchlogs.QueryDefinition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rName]
		if !ok {
			return fmt.Errorf("Not found: %s", rName)
		}
		conn := testAccProvider.Meta().(*AWSClient).cloudwatchlogsconn

		result, err := finder.QueryDefinition(context.Background(), conn, "", rs.Primary.ID)

		if err != nil {
			return err
		}

		if result == nil {
			return fmt.Errorf("CloudWatch query definition (%s) not found", rs.Primary.ID)
		}

		*v = *result

		return nil
	}
}

func testAccCheckAWSCloudWatchQueryDefinitionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudwatchlogsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_query_definition" {
			continue
		}

		result, err := finder.QueryDefinition(context.Background(), conn, "", rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error reading CloudWatch query definition (%s): %w", rs.Primary.ID, err)
		}

		if result != nil {
			return fmt.Errorf("CloudWatch query definition (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAWSCloudWatchQueryDefinitionConfig_Basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_query_definition" "test" {
  name = %[1]q

  query_string = <<EOF
fields @timestamp, @message
| sort @timestamp desc
| limit 20
EOF
}
`, rName)
}

func testAccAWSCloudWatchQueryDefinitionConfig_LogGroups(rName string, count int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_query_definition" "test" {
  name = %[1]q

  log_group_names = aws_cloudwatch_log_group.test[*].name

  query_string = <<EOF
fields @timestamp, @message
| sort @timestamp desc
| limit 20
EOF
}

resource "aws_cloudwatch_log_group" "test" {
  count = %[2]d

  name = "%[1]s-${count.index}"
}
`, rName, count)
}
