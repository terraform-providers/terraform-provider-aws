package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_msk_configuration", &resource.Sweeper{
		Name: "aws_msk_configuration",
		F:    testSweepMskConfigurations,
		Dependencies: []string{
			"aws_msk_cluster",
		},
	})
}

func testSweepMskConfigurations(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).kafkaconn
	var sweeperErrs *multierror.Error

	input := &kafka.ListConfigurationsInput{}

	err = conn.ListConfigurationsPages(input, func(page *kafka.ListConfigurationsOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, configuration := range page.Configurations {
			if configuration == nil {
				continue
			}

			arn := aws.StringValue(configuration.Arn)
			log.Printf("[INFO] Deleting MSK Configuration: %s", arn)

			r := resourceAwsMskConfiguration()
			d := r.Data(nil)
			d.SetId(arn)
			err := r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !isLast
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping MSK Configurations sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving MSK Configurations: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSMskConfiguration_basic(t *testing.T) {
	var configuration1 kafka.DescribeConfigurationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_msk_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMsk(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskConfigurationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMskConfigurationExists(resourceName, &configuration1),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "kafka", regexp.MustCompile(`configuration/.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "kafka_versions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest_revision", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestMatchResourceAttr(resourceName, "server_properties", regexp.MustCompile(`auto.create.topics.enable = true`)),
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

func TestAccAWSMskConfiguration_disappears(t *testing.T) {
	var configuration1 kafka.DescribeConfigurationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_msk_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMsk(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskConfigurationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMskConfigurationExists(resourceName, &configuration1),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsMskConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSMskConfiguration_Description(t *testing.T) {
	var configuration1, configuration2 kafka.DescribeConfigurationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_msk_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMsk(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskConfigurationConfigDescription(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMskConfigurationExists(resourceName, &configuration1),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMskConfigurationConfigDescription(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMskConfigurationExists(resourceName, &configuration2),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
					resource.TestCheckResourceAttr(resourceName, "latest_revision", "2"),
				),
			},
		},
	})
}

func TestAccAWSMskConfiguration_KafkaVersions(t *testing.T) {
	var configuration1 kafka.DescribeConfigurationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_msk_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMsk(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskConfigurationConfigKafkaVersions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMskConfigurationExists(resourceName, &configuration1),
					resource.TestCheckResourceAttr(resourceName, "kafka_versions.#", "2"),
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

func TestAccAWSMskConfiguration_ServerProperties(t *testing.T) {
	var configuration1, configuration2 kafka.DescribeConfigurationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_msk_configuration.test"
	serverProperty1 := "auto.create.topics.enable = false"
	serverProperty2 := "auto.create.topics.enable = true"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMsk(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskConfigurationConfigServerProperties(rName, serverProperty1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMskConfigurationExists(resourceName, &configuration1),
					resource.TestMatchResourceAttr(resourceName, "server_properties", regexp.MustCompile(serverProperty1)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMskConfigurationConfigServerProperties(rName, serverProperty2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMskConfigurationExists(resourceName, &configuration2),
					resource.TestCheckResourceAttr(resourceName, "latest_revision", "2"),
					resource.TestMatchResourceAttr(resourceName, "server_properties", regexp.MustCompile(serverProperty2)),
				),
			},
		},
	})
}

func testAccCheckMskConfigurationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).kafkaconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_msk_configuration" {
			continue
		}

		input := &kafka.DescribeConfigurationInput{
			Arn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeConfiguration(input)

		if tfawserr.ErrMessageContains(err, kafka.ErrCodeBadRequestException, "Configuration ARN does not exist") {
			continue
		}

		if err != nil {
			return err
		}

		if output != nil {
			return fmt.Errorf("MSK Configuration (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckMskConfigurationExists(resourceName string, configuration *kafka.DescribeConfigurationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource ID not set: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).kafkaconn

		input := &kafka.DescribeConfigurationInput{
			Arn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeConfiguration(input)

		if err != nil {
			return fmt.Errorf("error describing MSK Cluster (%s): %s", rs.Primary.ID, err)
		}

		*configuration = *output

		return nil
	}
}

func testAccMskConfigurationConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_msk_configuration" "test" {
  kafka_versions = ["2.1.0"]
  name           = %[1]q

  server_properties = <<PROPERTIES
auto.create.topics.enable = true
delete.topic.enable = true
PROPERTIES
}
`, rName)
}

func testAccMskConfigurationConfigDescription(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_msk_configuration" "test" {
  description    = %[2]q
  kafka_versions = ["2.1.0"]
  name           = %[1]q

  server_properties = <<PROPERTIES
auto.create.topics.enable = true
PROPERTIES
}
`, rName, description)
}

func testAccMskConfigurationConfigKafkaVersions(rName string) string {
	return fmt.Sprintf(`
resource "aws_msk_configuration" "test" {
  kafka_versions = ["1.1.1", "2.1.0"]
  name           = %[1]q

  server_properties = <<PROPERTIES
auto.create.topics.enable = true
PROPERTIES
}
`, rName)
}

func testAccMskConfigurationConfigServerProperties(rName string, serverProperty string) string {
	return fmt.Sprintf(`
resource "aws_msk_configuration" "test" {
  kafka_versions = ["2.1.0"]
  name           = %[1]q

  server_properties = <<PROPERTIES
%[2]s
PROPERTIES
}
`, rName, serverProperty)
}
