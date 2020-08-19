package aws

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSLightsailDomainEntry_basic(t *testing.T) {
	var domainEntry lightsail.DomainEntry
	lightsailDomainName := fmt.Sprintf("tf-test-lightsail-%s.com", acctest.RandString(5))
	lightsailDomainEntryName := fmt.Sprintf("test-%s.%s", acctest.RandString(5), lightsailDomainName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSLightsail(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLightsailDomainEntryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailDomainEntryConfig_basic(lightsailDomainName, lightsailDomainEntryName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailDomainEntryExists("aws_lightsail_domain_entry.entry_test", &domainEntry),
				),
			},
		},
	})
}

func TestAccAWSLightsailDomainEntry_disappears(t *testing.T) {
	var domainEntry lightsail.DomainEntry
	lightsailDomainName := fmt.Sprintf("tf-test-lightsail-%s.com", acctest.RandString(5))
	lightsailDomainEntryName := fmt.Sprintf("test-%s.%s", acctest.RandString(5), lightsailDomainName)

	domainEntryDestroy := func(*terraform.State) error {

		conn := testAccProvider.Meta().(*AWSClient).lightsailconn
		_, err := conn.DeleteDomainEntry(&lightsail.DeleteDomainEntryInput{
			DomainName: aws.String(lightsailDomainName),
			DomainEntry: &lightsail.DomainEntry{
				Name:   aws.String(lightsailDomainEntryName),
				Type:   aws.String("A"),
				Target: aws.String("127.0.0.1"),
			},
		})

		if err != nil {
			return fmt.Errorf("Error deleting Lightsail Domain Entry in disapear test")
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSLightsail(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLightsailDomainEntryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailDomainEntryConfig_basic(lightsailDomainName, lightsailDomainEntryName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailDomainEntryExists("aws_lightsail_domain_entry.entry_test", &domainEntry),
					domainEntryDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSLightsailDomainEntryExists(n string, domainEntry *lightsail.DomainEntry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Lightsail Domain Entry ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).lightsailconn

		resp, err := conn.GetDomain(&lightsail.GetDomainInput{
			DomainName: aws.String(rs.Primary.Attributes["domain_name"]),
		})

		if err != nil {
			return err
		}

		if resp == nil || resp.Domain == nil {
			return fmt.Errorf("Domain (%s) not found", rs.Primary.Attributes["domain_name"])
		}

		entryExists := false
		for _, n := range resp.Domain.DomainEntries {
			if rs.Primary.Attributes["name"] == *n.Name && rs.Primary.Attributes["type"] == *n.Type && rs.Primary.Attributes["target"] == *n.Target {
				*domainEntry = *n
				entryExists = true
				break
			}
		}

		if !entryExists {
			return fmt.Errorf("Domain entry (%s) not found", rs.Primary.Attributes["name"])
		}

		return nil
	}
}

func testAccCheckAWSLightsailDomainEntryDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lightsail_domain_entry" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).lightsailconn

		resp, err := conn.GetDomain(&lightsail.GetDomainInput{
			DomainName: aws.String(rs.Primary.ID),
		})

		if err == nil {
			if resp.Domain != nil {
				return fmt.Errorf("Lightsail Domain Entry %q still exists", rs.Primary.ID)
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

func testAccAWSLightsailDomainEntryConfig_basic(lightsailDomainName string, lightsailDomainEntryName string) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-east-1"
}

resource "aws_lightsail_domain" "domain_test" {
  domain_name = "%s"
}

resource "aws_lightsail_domain_entry" "entry_test" {
        domain_name = "%s"
        name = "%s"
        type = "A"
        target = "127.0.0.1"
        is_alias = false
  		depends_on = [aws_lightsail_domain.domain_test]
}


`, lightsailDomainName, lightsailDomainName, lightsailDomainEntryName)
}
