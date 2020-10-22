package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_ses_configuration_set", &resource.Sweeper{
		Name: "aws_ses_configuration_set",
		F:    testSweepSesConfigurationSets,
	})
}

func testSweepSesConfigurationSets(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).sesconn
	input := &ses.ListConfigurationSetsInput{}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.ListConfigurationSets(input)
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping SES Configuration Sets sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
		}
		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving SES Configuration Sets: %w", err))
			return sweeperErrs
		}

		for _, configurationSet := range output.ConfigurationSets {
			name := aws.StringValue(configurationSet.Name)

			log.Printf("[INFO] Deleting SES Configuration Set: %s", name)
			_, err := conn.DeleteConfigurationSet(&ses.DeleteConfigurationSetInput{
				ConfigurationSetName: aws.String(name),
			})
			if isAWSErr(err, ses.ErrCodeConfigurationSetDoesNotExistException, "") {
				continue
			}
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting SES Configuration Set (%s): %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}
		input.NextToken = output.NextToken
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSSESConfigurationSet_basic(t *testing.T) {
	var escRandomInteger = acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSSES(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSESConfigurationSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSESConfigurationSetConfig(escRandomInteger),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESConfigurationSetExists("aws_ses_configuration_set.test"),
				),
			},
			{
				ResourceName:      "aws_ses_configuration_set.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsSESConfigurationSetExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES configuration set not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES configuration set ID not set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sesconn

		response, err := conn.DescribeConfigurationSet(&ses.DescribeConfigurationSetInput{
			ConfigurationSetName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if aws.StringValue(response.ConfigurationSet.Name) != rs.Primary.ID {
			return fmt.Errorf("The configuration set was not created")
		}
		return nil

	}
}

func testAccCheckSESConfigurationSetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sesconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ses_configuration_set" {
			continue
		}

		_, err := conn.DescribeConfigurationSet(&ses.DescribeConfigurationSetInput{
			ConfigurationSetName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if isAWSErr(err, ses.ErrCodeConfigurationSetDoesNotExistException, "") {
				return nil
			}
			return err
		}
	}
	return nil
}

func testAccAWSSESConfigurationSetConfig(escRandomInteger int) string {
	return fmt.Sprintf(`
resource "aws_ses_configuration_set" "test" {
  name = "some-configuration-set-%d"
}
`, escRandomInteger)
}
