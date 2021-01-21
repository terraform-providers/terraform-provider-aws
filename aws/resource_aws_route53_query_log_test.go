package aws

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_route53_query_log", &resource.Sweeper{
		Name: "aws_route53_query_log",
		F:    testSweepRoute53QueryLogs,
	})
}

func testSweepRoute53QueryLogs(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).r53conn
	var sweeperErrs *multierror.Error

	err = conn.ListQueryLoggingConfigsPages(&route53.ListQueryLoggingConfigsInput{}, func(page *route53.ListQueryLoggingConfigsOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, queryLoggingConfig := range page.QueryLoggingConfigs {
			id := aws.StringValue(queryLoggingConfig.Id)

			log.Printf("[INFO] Deleting Route53 query logging configuration: %s", id)
			_, err := conn.DeleteQueryLoggingConfig(&route53.DeleteQueryLoggingConfigInput{
				Id: aws.String(id),
			})
			if isAWSErr(err, route53.ErrCodeNoSuchQueryLoggingConfig, "") {
				continue
			}
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Route53 query logging configuration (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !isLast
	})
	// In unsupported AWS partitions, the API may return an error even the SDK cannot handle.
	// Reference: https://github.com/aws/aws-sdk-go/issues/3313
	if testSweepSkipSweepError(err) || isAWSErr(err, "SerializationError", "failed to unmarshal error message") || isAWSErr(err, "AccessDeniedException", "Unable to determine service/operation name to be authorized") {
		log.Printf("[WARN] Skipping Route53 query logging configurations sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Route53 query logging configurations: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSRoute53QueryLog_basic(t *testing.T) {
	cloudwatchLogGroupResourceName := "aws_cloudwatch_log_group.test"
	resourceName := "aws_route53_query_log.test"
	route53ZoneResourceName := "aws_route53_zone.test"
	rName := strings.ToLower(fmt.Sprintf("%s-%s", t.Name(), acctest.RandString(5)))

	var queryLoggingConfig route53.QueryLoggingConfig
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckRoute53QueryLog(t) },
		ErrorCheck:        testAccErrorCheckSkipRoute53(t),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRoute53QueryLogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSRoute53QueryLogResourceConfigBasic1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53QueryLogExists(resourceName, &queryLoggingConfig),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_log_group_arn", cloudwatchLogGroupResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "zone_id", route53ZoneResourceName, "zone_id"),
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

func testAccCheckRoute53QueryLogExists(pr string, queryLoggingConfig *route53.QueryLoggingConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProviderRoute53QueryLog.Meta().(*AWSClient).r53conn
		rs, ok := s.RootModule().Resources[pr]
		if !ok {
			return fmt.Errorf("Not found: %s", pr)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		out, err := conn.GetQueryLoggingConfig(&route53.GetQueryLoggingConfigInput{
			Id: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}
		if out.QueryLoggingConfig == nil {
			return fmt.Errorf("Route53 query logging configuration does not exist: %q", rs.Primary.ID)
		}

		*queryLoggingConfig = *out.QueryLoggingConfig

		return nil
	}
}

func testAccCheckRoute53QueryLogDestroy(s *terraform.State) error {
	conn := testAccProviderRoute53QueryLog.Meta().(*AWSClient).r53conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_query_log" {
			continue
		}

		out, err := conn.GetQueryLoggingConfig(&route53.GetQueryLoggingConfigInput{
			Id: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return nil
		}

		if out.QueryLoggingConfig != nil {
			return fmt.Errorf("Route53 query logging configuration exists: %q", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSRoute53QueryLogResourceConfigBasic1(rName string) string {
	return composeConfig(
		testAccRoute53QueryLogRegionProviderConfig(),
		fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name              = "/aws/route53/${aws_route53_zone.test.name}"
  retention_in_days = 1
}

data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = ["arn:${data.aws_partition.current.partition}:logs:*:*:log-group:/aws/route53/*"]

    principals {
      identifiers = ["route53.${data.aws_partition.current.dns_suffix}"]
      type        = "Service"
    }
  }
}

resource "aws_cloudwatch_log_resource_policy" "test" {
  policy_name     = "%[1]s"
  policy_document = data.aws_iam_policy_document.test.json
}

resource "aws_route53_zone" "test" {
  name = "%[1]s.com"
}

resource "aws_route53_query_log" "test" {
  depends_on = [aws_cloudwatch_log_resource_policy.test]

  cloudwatch_log_group_arn = aws_cloudwatch_log_group.test.arn
  zone_id                  = aws_route53_zone.test.zone_id
}
`, rName))
}
