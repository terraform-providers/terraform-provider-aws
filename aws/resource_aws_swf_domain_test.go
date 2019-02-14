package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/swf"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSSwfDomain_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_swf_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSwfDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSwfDomainConfig_Name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSwfDomainExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccAWSSwfDomain_NamePrefix(t *testing.T) {
	resourceName := "aws_swf_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSwfDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSwfDomainConfig_NamePrefix,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSwfDomainExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile(`^tf-acc-test`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"}, // this line is only necessary if the test configuration is using name_prefix
			},
		},
	})
}

func TestAccAWSSwfDomain_GeneratedName(t *testing.T) {
	resourceName := "aws_swf_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSwfDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSwfDomainConfig_GeneratedName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSwfDomainExists(resourceName),
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

func TestAccAWSSwfDomain_Description(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_swf_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSwfDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSwfDomainConfig_Description(rName, "description1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsSwfDomainExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
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

func testAccCheckAwsSwfDomainDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).swfconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_swf_domain" {
			continue
		}

		name := rs.Primary.ID
		input := &swf.DescribeDomainInput{
			Name: aws.String(name),
		}

		resp, err := conn.DescribeDomain(input)
		if err != nil {
			return err
		}

		if *resp.DomainInfo.Status != "DEPRECATED" {
			return fmt.Errorf(`SWF Domain %s status is %s instead of "DEPRECATED". Failing!`, name, *resp.DomainInfo.Status)
		}
	}

	return nil
}

func testAccCheckAwsSwfDomainExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SWF Domain not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SWF Domain name not set")
		}

		name := rs.Primary.ID
		conn := testAccProvider.Meta().(*AWSClient).swfconn

		input := &swf.DescribeDomainInput{
			Name: aws.String(name),
		}

		resp, err := conn.DescribeDomain(input)
		if err != nil {
			return fmt.Errorf("SWF Domain %s not found in AWS", name)
		}

		if *resp.DomainInfo.Status != "REGISTERED" {
			return fmt.Errorf(`SWF Domain %s status is %s instead of "REGISTERED". Failing!`, name, *resp.DomainInfo.Status)
		}
		return nil
	}
}

func testAccAWSSwfDomainConfig_Description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_swf_domain" "test" {
  description                                 = %q
  name                                        = %q
  workflow_execution_retention_period_in_days = 1
}
`, description, rName)
}

const testAccAWSSwfDomainConfig_GeneratedName = `
resource "aws_swf_domain" "test" {
  workflow_execution_retention_period_in_days = 1
}
`

func testAccAWSSwfDomainConfig_Name(rName string) string {
	return fmt.Sprintf(`
resource "aws_swf_domain" "test" {
  name                                        = %q
  workflow_execution_retention_period_in_days = 1
}
`, rName)
}

const testAccAWSSwfDomainConfig_NamePrefix = `
resource "aws_swf_domain" "test" {
  name_prefix                                 = "tf-acc-test"
  workflow_execution_retention_period_in_days = 1
}
`
