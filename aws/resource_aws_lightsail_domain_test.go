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

func TestAccAWSLightsailDomain_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Domain": {
			"basic":      testAccAWSLightsailDomain_basic,
			"disappears": testAccAWSLightsailDomain_disappears,
			"DomainName": testAccAWSLightsailDomain_DomainName,
			"Tags":       testAccAWSLightsailDomain_Tags,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}

func testAccAWSLightsailDomain_basic(t *testing.T) {
	var domain lightsail.Domain
	rName := acctest.RandomWithPrefix("tf-acc-test.com")
	resourceName := "aws_lightsail_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckLightsailDomain(t) },
		CheckDestroy: testAccCheckAWSLightsailDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailDomainConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "domain_name", rName),
				),
			},
		},
	})
}

func testAccAWSLightsailDomain_disappears(t *testing.T) {
	var domain lightsail.Domain
	rName := acctest.RandomWithPrefix("tf-acc-test.com")
	resourceName := "aws_lightsail_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckLightsailDomain(t) },
		CheckDestroy: testAccCheckAWSLightsailDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailDomainConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
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

func testAccAWSLightsailDomain_DomainName(t *testing.T) {
	var domain1, domain2 lightsail.Domain
	rName := acctest.RandomWithPrefix("tf-acc-test.com")
	rName2 := acctest.RandomWithPrefix("tf-acc-test.com")
	resourceName := "aws_lightsail_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckLightsailDomain(t) },
		CheckDestroy: testAccCheckAWSLightsailDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailDomainConfigBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailDomainExists(resourceName, &domain1),
				),
			},
			{
				Config: testAccAWSLightsailDomainConfigBasic(rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailDomainExists(resourceName, &domain2),
				),
			},
		},
	})
}

func testAccAWSLightsailDomain_Tags(t *testing.T) {
	var domain1, domain2, domain3 lightsail.Domain
	rName := acctest.RandomWithPrefix("tf-acc-test.com")
	resourceName := "aws_lightsail_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckLightsailDomain(t) },
		CheckDestroy: testAccCheckAWSLightsailDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailDomainConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailDomainExists(resourceName, &domain1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLightsailDomainConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailDomainExists(resourceName, &domain2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSLightsailDomainConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailDomainExists(resourceName, &domain3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
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

func testAccAWSLightsailDomainConfigBasic(rName string) string {
	return composeConfig(
		testAccLightsailDomainRegionProviderConfig(),
		fmt.Sprintf(`
resource "aws_lightsail_domain" "test" {
  domain_name = %[1]q
}
`, rName))
}

func testAccAWSLightsailDomainConfigTags1(rName string, tagKey1, tagValue1 string) string {
	return composeConfig(
		testAccLightsailDomainRegionProviderConfig(),
		fmt.Sprintf(`
resource "aws_lightsail_domain" "test" {
  domain_name = %[1]q
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccAWSLightsailDomainConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return composeConfig(
		testAccLightsailDomainRegionProviderConfig(),
		fmt.Sprintf(`
resource "aws_lightsail_domain" "test" {
  domain_name = %[1]q
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
