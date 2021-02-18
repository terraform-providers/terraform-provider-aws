package aws

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudsearch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSCloudSearchDomain_basic(t *testing.T) {
	var domains cloudsearch.DescribeDomainsOutput
	resourceName := "aws_cloudsearch_domain.test"
	rString := acctest.RandString(8)
	domainName := fmt.Sprintf("tf-acc-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudSearchDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudSearchDomainConfig_basic(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudSearchDomainExists(resourceName, &domains),
					resource.TestCheckResourceAttr(resourceName, "name", domainName),
				),
			},
			// {
			// 	ResourceName:      resourceName,
			// 	ImportState:       false,
			// 	ImportStateVerify: false,
			// },
		},
	})
}

func TestAccAWSCloudSearchDomain_badName(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSCloudSearchDomainConfig_basic("-this-is-a-bad-name"),
				ExpectError: regexp.MustCompile(`.*"name" must begin with a.*`),
			},
		},
	})
}

func TestAccAWSCloudSearchDomain_badInstanceType(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSCloudSearchDomainConfig_withInstanceType("bad-instance-type", "nope.small"),
				ExpectError: regexp.MustCompile(`.*failed to satisfy constraint.*`),
			},
		},
	})
}

func TestAccAWSCloudSearchDomain_badIndexFieldNames(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSCloudSearchDomainConfig_withIndex("bad-index-name", "HELLO", "text"),
				ExpectError: regexp.MustCompile(`.*must begin with a letter and be at least 3 and no more than 64 characters long.*`),
			},
			{
				Config:      testAccAWSCloudSearchDomainConfig_withIndex("bad-index-name", "w-a", "text"),
				ExpectError: regexp.MustCompile(`.*must begin with a letter and be at least 3 and no more than 64 characters long.*`),
			},
			{
				Config:      testAccAWSCloudSearchDomainConfig_withIndex("bad-index-name", "jfjdbfjdhsjakhfdhsajkfhdjksahfdsbfkjchndsjkhafbjdkshafjkdshjfhdsjkahfjkdsha", "text"),
				ExpectError: regexp.MustCompile(`.*must begin with a letter and be at least 3 and no more than 64 characters long.*`),
			},
			{
				Config:      testAccAWSCloudSearchDomainConfig_withIndex("bad-index-name", "w", "text"),
				ExpectError: regexp.MustCompile(`.*must begin with a letter and be at least 3 and no more than 64 characters long.*`),
			},
		},
	})
}

func TestAccAWSCloudSearchDomain_badIndexFieldType(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		// CheckDestroy: testAccCheckAWSCloudSearchDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSCloudSearchDomainConfig_withIndex("bad-index-type", "name", "not-a-type"),
				ExpectError: regexp.MustCompile(`.*is not a valid index type.*`),
			},
		},
	})
}

func testAccCheckAWSCloudSearchDomainExists(n string, domains *cloudsearch.DescribeDomainsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudsearchconn

		domainList := cloudsearch.DescribeDomainsInput{
			DomainNames: []*string{
				aws.String(rs.Primary.Attributes["name"]),
			},
		}

		resp, err := conn.DescribeDomains(&domainList)
		if err != nil {
			return err
		}

		*domains = *resp

		return nil
	}
}

func testAccCheckAWSCloudSearchDomainDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudsearch_domain" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudsearchconn
		// Wait for the resource to start being deleted, which is marked as "Deleted" from the API.
		stateConf := &resource.StateChangeConf{
			Pending:        []string{"false"},
			Target:         []string{"true"},
			Timeout:        20 * time.Minute,
			NotFoundChecks: 100,
			Refresh: func() (interface{}, string, error) {
				domainlist := cloudsearch.DescribeDomainsInput{
					DomainNames: []*string{
						aws.String(rs.Primary.Attributes["name"]),
					},
				}

				resp, err := conn.DescribeDomains(&domainlist)
				if err != nil {
					return nil, "false", err
				}

				domain := resp.DomainStatusList[0]

				// If we see that the domain has been deleted, go ahead and return true.
				if *domain.Deleted {
					return domain, "true", nil
				}

				if domain.Deleted != nil {
					return nil, strconv.FormatBool(*domain.Deleted), nil
				}

				return nil, "false", nil
			},
		}
		_, err := stateConf.WaitForState()
		return err
	}

	return nil
}

func testAccAWSCloudSearchDomainConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudsearch_domain" "test" {
	name = "%s"

	index {
		name            = "headline"
		type            = "text"
		search          = true
		return          = true
		sort            = true
		highlight       = false
		analysis_scheme = "_en_default_"
	}

	wait_for_endpoints = false

	access_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [{
		"Effect": "Allow",
		"Principal": {
			"AWS": ["*"]
		},
		"Action": ["cloudsearch:*"]
	}]
}
	EOF
}
`, name)
}

func testAccAWSCloudSearchDomainConfig_withInstanceType(name string, instance_type string) string {
	return fmt.Sprintf(`
resource "aws_cloudsearch_domain" "test" {
	name = "%s"

	instance_type = "%s"

	#wait_for_endpoints = "false"
	
	access_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [{
		"Effect": "Allow",
		"Principal": {
			"AWS": ["*"]
		},
		"Action": ["cloudsearch:*"]
	}]
}
	EOF
}
`, name, instance_type)
}

func testAccAWSCloudSearchDomainConfig_withIndex(name string, index_name string, index_type string) string {
	return fmt.Sprintf(`
resource "aws_cloudsearch_domain" "test" {
	name = "%s"

	index {
		name            = "%s"
		type            = "%s"
		search          = true
		return          = true
		sort            = true
		highlight       = false
		analysis_scheme = "_en_default_"
	}

	#wait_for_endpoints = "false"
	
	access_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [{
		"Effect": "Allow",
		"Principal": {
			"AWS": ["*"]
		},
		"Action": ["cloudsearch:*"]
	}]
}
	EOF
}
`, name, index_name, index_type)
}
