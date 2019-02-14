package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_config_configuration_aggregator", &resource.Sweeper{
		Name: "aws_config_configuration_aggregator",
		F:    testSweepConfigConfigurationAggregators,
	})
}

func testSweepConfigConfigurationAggregators(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*AWSClient).configconn

	resp, err := conn.DescribeConfigurationAggregators(&configservice.DescribeConfigurationAggregatorsInput{})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Config Configuration Aggregators sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving config configuration aggregators: %s", err)
	}

	if len(resp.ConfigurationAggregators) == 0 {
		log.Print("[DEBUG] No config configuration aggregators to sweep")
		return nil
	}

	log.Printf("[INFO] Found %d config configuration aggregators", len(resp.ConfigurationAggregators))

	for _, agg := range resp.ConfigurationAggregators {
		if !strings.HasPrefix(*agg.ConfigurationAggregatorName, "tf-") {
			continue
		}

		log.Printf("[INFO] Deleting config configuration aggregator %s", *agg.ConfigurationAggregatorName)
		_, err := conn.DeleteConfigurationAggregator(&configservice.DeleteConfigurationAggregatorInput{
			ConfigurationAggregatorName: agg.ConfigurationAggregatorName,
		})

		if err != nil {
			return fmt.Errorf("Error deleting config configuration aggregator %s: %s", *agg.ConfigurationAggregatorName, err)
		}
	}

	return nil
}

func TestAccAWSConfigConfigurationAggregator_account(t *testing.T) {
	var ca configservice.ConfigurationAggregator
	rString := acctest.RandString(10)
	expectedName := fmt.Sprintf("tf-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSConfigConfigurationAggregatorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSConfigConfigurationAggregatorConfig_account(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSConfigConfigurationAggregatorExists("aws_config_configuration_aggregator.example", &ca),
					testAccCheckAWSConfigConfigurationAggregatorName("aws_config_configuration_aggregator.example", expectedName, &ca),
					resource.TestCheckResourceAttr("aws_config_configuration_aggregator.example", "name", expectedName),
					resource.TestCheckResourceAttr("aws_config_configuration_aggregator.example", "account_aggregation_source.#", "1"),
					resource.TestCheckResourceAttr("aws_config_configuration_aggregator.example", "account_aggregation_source.0.account_ids.#", "1"),
					resource.TestMatchResourceAttr("aws_config_configuration_aggregator.example", "account_aggregation_source.0.account_ids.0", regexp.MustCompile(`^\d{12}$`)),
					resource.TestCheckResourceAttr("aws_config_configuration_aggregator.example", "account_aggregation_source.0.regions.#", "1"),
					resource.TestCheckResourceAttr("aws_config_configuration_aggregator.example", "account_aggregation_source.0.regions.0", "us-west-2"),
				),
			},
			{
				ResourceName:      "aws_config_configuration_aggregator.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSConfigConfigurationAggregator_organization(t *testing.T) {
	var ca configservice.ConfigurationAggregator
	rString := acctest.RandString(10)
	expectedName := fmt.Sprintf("tf-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSConfigConfigurationAggregatorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSConfigConfigurationAggregatorConfig_organization(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSConfigConfigurationAggregatorExists("aws_config_configuration_aggregator.example", &ca),
					testAccCheckAWSConfigConfigurationAggregatorName("aws_config_configuration_aggregator.example", expectedName, &ca),
					resource.TestCheckResourceAttr("aws_config_configuration_aggregator.example", "name", expectedName),
					resource.TestCheckResourceAttr("aws_config_configuration_aggregator.example", "organization_aggregation_source.#", "1"),
					resource.TestMatchResourceAttr("aws_config_configuration_aggregator.example", "organization_aggregation_source.0.role_arn", regexp.MustCompile(`^arn:aws:iam::\d+:role/`)),
					resource.TestCheckResourceAttr("aws_config_configuration_aggregator.example", "organization_aggregation_source.0.all_regions", "true"),
				),
			},
			{
				ResourceName:      "aws_config_configuration_aggregator.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSConfigConfigurationAggregator_switch(t *testing.T) {
	rString := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSConfigConfigurationAggregatorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSConfigConfigurationAggregatorConfig_account(rString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_config_configuration_aggregator.example", "account_aggregation_source.#", "1"),
					resource.TestCheckResourceAttr("aws_config_configuration_aggregator.example", "organization_aggregation_source.#", "0"),
				),
			},
			{
				Config: testAccAWSConfigConfigurationAggregatorConfig_organization(rString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_config_configuration_aggregator.example", "account_aggregation_source.#", "0"),
					resource.TestCheckResourceAttr("aws_config_configuration_aggregator.example", "organization_aggregation_source.#", "1"),
				),
			},
		},
	})
}

func testAccCheckAWSConfigConfigurationAggregatorName(n, desired string, obj *configservice.ConfigurationAggregator) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.Attributes["name"] != *obj.ConfigurationAggregatorName {
			return fmt.Errorf("Expected name: %q, given: %q", desired, *obj.ConfigurationAggregatorName)
		}
		return nil
	}
}

func testAccCheckAWSConfigConfigurationAggregatorExists(n string, obj *configservice.ConfigurationAggregator) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No config configuration aggregator ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).configconn
		out, err := conn.DescribeConfigurationAggregators(&configservice.DescribeConfigurationAggregatorsInput{
			ConfigurationAggregatorNames: []*string{aws.String(rs.Primary.Attributes["name"])},
		})
		if err != nil {
			return fmt.Errorf("Failed to describe config configuration aggregator: %s", err)
		}
		if len(out.ConfigurationAggregators) < 1 {
			return fmt.Errorf("No config configuration aggregator found when describing %q", rs.Primary.Attributes["name"])
		}

		ca := out.ConfigurationAggregators[0]
		*obj = *ca

		return nil
	}
}

func testAccCheckAWSConfigConfigurationAggregatorDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).configconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_config_configuration_aggregator" {
			continue
		}

		resp, err := conn.DescribeConfigurationAggregators(&configservice.DescribeConfigurationAggregatorsInput{
			ConfigurationAggregatorNames: []*string{aws.String(rs.Primary.Attributes["name"])},
		})

		if err == nil {
			if len(resp.ConfigurationAggregators) != 0 &&
				*resp.ConfigurationAggregators[0].ConfigurationAggregatorName == rs.Primary.Attributes["name"] {
				return fmt.Errorf("config configuration aggregator still exists: %s", rs.Primary.Attributes["name"])
			}
		}
	}

	return nil
}

func testAccAWSConfigConfigurationAggregatorConfig_account(rString string) string {
	return fmt.Sprintf(`
resource "aws_config_configuration_aggregator" "example" {
  name = "tf-%s"

  account_aggregation_source {
    account_ids = ["${data.aws_caller_identity.current.account_id}"]
    regions     = ["us-west-2"]
  }
}

data "aws_caller_identity" "current" {}
`, rString)
}

func testAccAWSConfigConfigurationAggregatorConfig_organization(rString string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_config_configuration_aggregator" "example" {
  depends_on = ["aws_iam_role_policy_attachment.example"]

  name = "tf-%s"

  organization_aggregation_source {
    all_regions = true
    role_arn    = "${aws_iam_role.example.arn}"
  }
}

resource "aws_iam_role" "example" {
  name = "tf-%s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "config.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "example" {
  role       = "${aws_iam_role.example.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSConfigRoleForOrganizations"
}
`, rString, rString)
}
