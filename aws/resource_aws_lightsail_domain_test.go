package aws

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSLightsailDomain_basic(t *testing.T) {
	var domain lightsail.Domain
	lightsailDomainName := fmt.Sprintf("tf-test-lightsail-%s.com", acctest.RandString(5))
	resourceName := "aws_lightsail_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckLightsailDomain(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSLightsailDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailDomainConfig_basic(lightsailDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailDomainExists(resourceName, &domain),
				),
			},
		},
	})
}

func TestAccAWSLightsailDomain_disappears(t *testing.T) {
	var domain lightsail.Domain
	lightsailDomainName := fmt.Sprintf("tf-test-lightsail-%s.com", acctest.RandString(5))
	resourceName := "aws_lightsail_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckLightsailDomain(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSLightsailDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailDomainConfig_basic(lightsailDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailDomainExists(resourceName, &domain),
					testAccCheckResourceDisappears(testAccProviderLightsailDomain, resourceAwsLightsailDomain(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSLightsailDomainExists(n string, domain *lightsail.Domain) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Lightsail Domain ID is set")
		}

		conn := testAccProviderLightsailDomain.Meta().(*AWSClient).lightsailconn

		resp, err := conn.GetDomain(&lightsail.GetDomainInput{
			DomainName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if resp == nil || resp.Domain == nil {
			return fmt.Errorf("Domain (%s) not found", rs.Primary.ID)
		}
		*domain = *resp.Domain
		return nil
	}
}

func testAccCheckAWSLightsailDomainDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lightsail_domain" {
			continue
		}

		conn := testAccProviderLightsailDomain.Meta().(*AWSClient).lightsailconn

		resp, err := conn.GetDomain(&lightsail.GetDomainInput{
			DomainName: aws.String(rs.Primary.ID),
		})

		if err == nil {
			if resp.Domain != nil {
				return fmt.Errorf("Lightsail Domain %q still exists", rs.Primary.ID)
			}
		}

		// Verify the error
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "NotFoundException" {
				return nil
			}
		}
		return err
	}

	return nil
}

func testAccAWSLightsailDomainConfig_basic(lightsailDomainName string) string {
	return composeConfig(
		testAccLightsailDomainRegionProviderConfig(),
		fmt.Sprintf(`
resource "aws_lightsail_domain" "test" {
  domain_name = "%s"
}
`, lightsailDomainName))
}
