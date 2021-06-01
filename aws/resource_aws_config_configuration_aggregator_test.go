package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_configuration_aggregator.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSConfigConfigurationAggregatorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSConfigConfigurationAggregatorConfig_account(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSConfigConfigurationAggregatorExists(resourceName, &ca),
					testAccCheckAWSConfigConfigurationAggregatorName(resourceName, rName, &ca),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "account_aggregation_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "account_aggregation_source.0.account_ids.#", "1"),
					testAccCheckResourceAttrAccountID(resourceName, "account_aggregation_source.0.account_ids.0"),
					resource.TestCheckResourceAttr(resourceName, "account_aggregation_source.0.regions.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "account_aggregation_source.0.regions.0", "data.aws_region.current", "name"),
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

func TestAccAWSConfigConfigurationAggregator_organization(t *testing.T) {
	var ca configservice.ConfigurationAggregator
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_configuration_aggregator.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSConfigConfigurationAggregatorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSConfigConfigurationAggregatorConfig_organization(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSConfigConfigurationAggregatorExists(resourceName, &ca),
					testAccCheckAWSConfigConfigurationAggregatorName(resourceName, rName, &ca),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "organization_aggregation_source.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "organization_aggregation_source.0.role_arn", "aws_iam_role.r", "arn"),
					resource.TestCheckResourceAttr(resourceName, "organization_aggregation_source.0.all_regions", "true"),
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

func TestAccAWSConfigConfigurationAggregator_switch(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_configuration_aggregator.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSConfigConfigurationAggregatorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSConfigConfigurationAggregatorConfig_account(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "account_aggregation_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "organization_aggregation_source.#", "0"),
				),
			},
			{
				Config: testAccAWSConfigConfigurationAggregatorConfig_organization(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "account_aggregation_source.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "organization_aggregation_source.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSConfigConfigurationAggregator_tags(t *testing.T) {
	var ca configservice.ConfigurationAggregator
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_configuration_aggregator.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSConfigConfigurationAggregatorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSConfigConfigurationAggregatorConfig_tags(rName, "foo", "bar", "fizz", "buzz"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSConfigConfigurationAggregatorExists(resourceName, &ca),
					testAccCheckAWSConfigConfigurationAggregatorName(resourceName, rName, &ca),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.fizz", "buzz"),
				),
			},
			{
				Config: testAccAWSConfigConfigurationAggregatorConfig_tags(rName, "foo", "bar2", "fizz2", "buzz2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSConfigConfigurationAggregatorExists(resourceName, &ca),
					testAccCheckAWSConfigConfigurationAggregatorName(resourceName, rName, &ca),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar2"),
					resource.TestCheckResourceAttr(resourceName, "tags.fizz2", "buzz2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSConfigConfigurationAggregatorConfig_account(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSConfigConfigurationAggregatorExists(resourceName, &ca),
					testAccCheckAWSConfigConfigurationAggregatorName(resourceName, rName, &ca),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func testAccAWSConfigConfigurationAggregatorConfig_account(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_config_configuration_aggregator" "example" {
  name = %q

  account_aggregation_source {
    account_ids = [data.aws_caller_identity.current.account_id]
    regions     = [data.aws_region.current.name]
  }
}

data "aws_caller_identity" "current" {
}
`, rName)
}

func testAccAWSConfigConfigurationAggregatorConfig_organization(rName string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_config_configuration_aggregator" "example" {
  depends_on = [aws_iam_role_policy_attachment.example]

  name = %[1]q

  organization_aggregation_source {
    all_regions = true
    role_arn    = aws_iam_role.example.arn
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role" "example" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "config.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "example" {
  role       = aws_iam_role.example.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSConfigRoleForOrganizations"
}
`, rName)
}

func testAccAWSConfigConfigurationAggregatorConfig_tags(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_config_configuration_aggregator" "example" {
  name = %[1]q

  account_aggregation_source {
    account_ids = [data.aws_caller_identity.current.account_id]
    regions     = [data.aws_region.current.name]
  }

  tags = {
    Name = %[1]q

    %[2]s = %[3]q
    %[4]s = %[5]q
  }
}

data "aws_caller_identity" "current" {}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
