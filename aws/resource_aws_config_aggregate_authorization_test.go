package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_config_aggregate_authorization", &resource.Sweeper{
		Name: "aws_config_aggregate_authorization",
		F:    testSweepConfigAggregateAuthorizations,
	})
}

func testSweepConfigAggregateAuthorizations(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*AWSClient).configconn

	aggregateAuthorizations, err := describeConfigAggregateAuthorizations(conn)
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Config Aggregate Authorizations sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving config aggregate authorizations: %s", err)
	}

	if len(aggregateAuthorizations) == 0 {
		log.Print("[DEBUG] No config aggregate authorizations to sweep")
		return nil
	}

	log.Printf("[INFO] Found %d config aggregate authorizations", len(aggregateAuthorizations))

	for _, auth := range aggregateAuthorizations {
		log.Printf("[INFO] Deleting config authorization %s", *auth.AggregationAuthorizationArn)
		_, err := conn.DeleteAggregationAuthorization(&configservice.DeleteAggregationAuthorizationInput{
			AuthorizedAccountId: auth.AuthorizedAccountId,
			AuthorizedAwsRegion: auth.AuthorizedAwsRegion,
		})
		if err != nil {
			return fmt.Errorf("Error deleting config aggregate authorization %s: %s", *auth.AggregationAuthorizationArn, err)
		}
	}

	return nil
}

func TestAccAWSConfigAggregateAuthorization_basic(t *testing.T) {
	rString := acctest.RandStringFromCharSet(12, "0123456789")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSConfigAggregateAuthorizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSConfigAggregateAuthorizationConfig_basic(rString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_config_aggregate_authorization.example", "account_id", rString),
					resource.TestCheckResourceAttr("aws_config_aggregate_authorization.example", "region", "eu-west-1"),
					resource.TestMatchResourceAttr("aws_config_aggregate_authorization.example", "arn", regexp.MustCompile(`^arn:aws:config:[\w-]+:\d{12}:aggregation-authorization/\d{12}/[\w-]+$`)),
				),
			},
			{
				ResourceName:      "aws_config_aggregate_authorization.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSConfigAggregateAuthorizationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).configconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_config_aggregate_authorization" {
			continue
		}

		accountId, region, err := resourceAwsConfigAggregateAuthorizationParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		aggregateAuthorizations, err := describeConfigAggregateAuthorizations(conn)

		if err != nil {
			return err
		}

		for _, auth := range aggregateAuthorizations {
			if accountId == aws.StringValue(auth.AuthorizedAccountId) && region == aws.StringValue(auth.AuthorizedAwsRegion) {
				return fmt.Errorf("Config aggregate authorization still exists: %s", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccAWSConfigAggregateAuthorizationConfig_basic(rString string) string {
	return fmt.Sprintf(`
resource "aws_config_aggregate_authorization" "example" {
  account_id = "%s"
  region = "eu-west-1"
}`, rString)
}
