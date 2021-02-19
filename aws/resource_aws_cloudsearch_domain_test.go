package aws

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudsearch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_cloudsearch_domain", &resource.Sweeper{
		Name: "aws_cloudsearch_domain",
		F: func(region string) error {
			client, err := sharedClientForRegion(region)
			if err != nil {
				return fmt.Errorf("error getting client: %s", err)
			}
			conn := client.(*AWSClient).cloudsearchconn

			domains, err := conn.DescribeDomains(&cloudsearch.DescribeDomainsInput{})
			if err != nil {
				return fmt.Errorf("error describing CloudSearch domains: %s", err)
			}

			for _, domain := range domains.DomainStatusList {
				if !strings.HasPrefix(*domain.DomainName, "tf-acc-") {
					continue
				}
				_, err := conn.DeleteDomain(&cloudsearch.DeleteDomainInput{
					DomainName: domain.DomainName,
				})
				if err != nil {
					return fmt.Errorf("error deleting CloudSearch domain: %s", err)
				}
			}
			return nil
		},
	})
}

func TestAccAWSCloudSearchDomain_basic(t *testing.T) {
	var domains cloudsearch.DescribeDomainsOutput
	resourceName := "aws_cloudsearch_domain.test"
	domainName := fmt.Sprintf("tf-acc-%s", acctest.RandString(8))

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
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// {
			// 	Config: testAccAWSCloudSearchDomainConfig_basicIndexMix(domainName),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		testAccCheckAWSCloudSearchDomainExists(resourceName, &domains),
			// 		resource.TestCheckResourceAttr(resourceName, "name", domainName),
			// 	),
			// },
			// {
			// 	ResourceName:      resourceName,
			// 	ImportState:       true,
			// 	ImportStateVerify: true,
			// },
		},
	})
}

func TestAccAWSCloudSearchDomain_textAnalysisScheme(t *testing.T) {
	var domains cloudsearch.DescribeDomainsOutput
	resourceName := "aws_cloudsearch_domain.test"
	domainName := fmt.Sprintf("tf-acc-%s", acctest.RandString(8))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudSearchDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudSearchDomainConfig_textAnalysisScheme(domainName, "_en_default_"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudSearchDomainExists(resourceName, &domains),
					resource.TestCheckResourceAttr(resourceName, "name", domainName),
				),
			},
			{
				Config: testAccAWSCloudSearchDomainConfig_textAnalysisScheme(domainName, "_fr_default_"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudSearchDomainExists(resourceName, &domains),
					resource.TestCheckResourceAttr(resourceName, "name", domainName),
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
	domainName := fmt.Sprintf("tf-acc-%s", acctest.RandString(8))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSCloudSearchDomainConfig_withInstanceType(domainName, "nope.small"),
				ExpectError: regexp.MustCompile(`.*is not a valid instance type.*`),
			},
		},
	})
}

func TestAccAWSCloudSearchDomain_badIndexFieldNames(t *testing.T) {
	domainName := fmt.Sprintf("tf-acc-%s", acctest.RandString(8))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSCloudSearchDomainConfig_withIndex(domainName, "HELLO", "text"),
				ExpectError: regexp.MustCompile(`.*must begin with a letter and be at least 3 and no more than 64 characters long.*`),
			},
			{
				Config:      testAccAWSCloudSearchDomainConfig_withIndex(domainName, "w-a", "text"),
				ExpectError: regexp.MustCompile(`.*must begin with a letter and be at least 3 and no more than 64 characters long.*`),
			},
			{
				Config:      testAccAWSCloudSearchDomainConfig_withIndex(domainName, "jfjdbfjdhsjakhfdhsajkfhdjksahfdsbfkjchndsjkhafbjdkshafjkdshjfhdsjkahfjkdsha", "text"),
				ExpectError: regexp.MustCompile(`.*must begin with a letter and be at least 3 and no more than 64 characters long.*`),
			},
			{
				Config:      testAccAWSCloudSearchDomainConfig_withIndex(domainName, "w", "text"),
				ExpectError: regexp.MustCompile(`.*must begin with a letter and be at least 3 and no more than 64 characters long.*`),
			},
		},
	})
}

func TestAccAWSCloudSearchDomain_badIndexFieldType(t *testing.T) {
	domainName := fmt.Sprintf("tf-acc-%s", acctest.RandString(8))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSCloudSearchDomainConfig_withIndex(domainName, "directory", "not-a-type"),
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
		name   = "date_test"
		type   = "date"
		facet  = true
		search = true
		return = true
		sort   = true
	}

	index {
		name   = "date_array_test"
		type   = "date-array"
		facet  = true
		search = true
		return = true
	}

	index {
		name   = "double_test"
		type   = "double"
		facet  = true
		search = true
		return = true
		sort   = true
	}

	index {
		name   = "double_array_test"
		type   = "double-array"
		facet  = true
		search = true
		return = true
	}
	
	index {
		name   = "int_test"
		type   = "int"
		facet  = true
		search = true
		return = true
		sort   = true
	}

	index {
		name   = "int_array_test"
		type   = "int-array"
		facet  = true
		search = true
		return = true
	}

	index {
		name   = "latlon_test"
		type   = "latlon"
		facet  = true
		search = true
		return = true
		sort   = true
	}

	index {
		name   = "literal_test"
		type   = "literal"
		facet  = true
		search = true
		return = true
		sort   = true
	}

	index {
		name   = "literal_array_test"
		type   = "literal-array"
		facet  = true
		search = true
		return = true
	}

	index {
		name            = "text_test"
		type            = "text"
		analysis_scheme = "_en_default_"
		highlight       = true
		return          = true
		sort            = true
	}

	index {
		name            = "text_array_test"
		type            = "text-array"
		analysis_scheme = "_en_default_"
		highlight       = true
		return          = true
	}

	wait_for_endpoints = false
	service_access_policies = <<EOF
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

func testAccAWSCloudSearchDomainConfig_basicIndexMix(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudsearch_domain" "test" {
	name = "%s"

	index {
		name            = "how_about_one_up_here"
		type            = "text"
		analysis_scheme = "_en_default_"
	}

	index {
		name   = "date_test"
		type   = "date"
		facet  = true
		search = true
		return = true
		sort   = true
	}

	index {
		name   = "double_test_2"
		type   = "double"
		facet  = true
		search = true
		return = true
		sort   = true
	}

	index {
		name   = "double_array_test"
		type   = "double-array"
		facet  = true
		search = true
		return = true
	}

	index {
		name   = "just_another_index_name"
		type   = "literal-array"
		facet  = true
		search = true
		return = true
	}

	index {
		name            = "text_test"
		type            = "text"
		analysis_scheme = "_en_default_"
		highlight       = true
		return          = true
		sort            = true
	}

	index {
		name = "captain_janeway_is_pretty_cool"
		type = "double"
	}

	wait_for_endpoints = false
	service_access_policies = <<EOF
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

// NOTE: I'd like to get text and text arrays field to work properly without having to explicitly set the
// `analysis_scheme` field, but I cannot find a way to suppress the diff Terraform ends up generating as a result.
func testAccAWSCloudSearchDomainConfig_textAnalysisScheme(name string, scheme string) string {
	return fmt.Sprintf(`
resource "aws_cloudsearch_domain" "test" {
	name = "%s"

	#index {
	#	name = "use_default_scheme"
	#	type = "text"
	#}

	index {
		name            = "specify_scheme"
		type            = "text"
		analysis_scheme = "%s"
	}

	wait_for_endpoints = false
	service_access_policies = <<EOF
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
`, name, scheme)
}

func testAccAWSCloudSearchDomainConfig_withInstanceType(name string, instance_type string) string {
	return fmt.Sprintf(`
resource "aws_cloudsearch_domain" "test" {
	name = "%s"

	instance_type = "%s"

	wait_for_endpoints = false
	
	service_access_policies = <<EOF
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
		facet           = false
		search          = false
		return          = true
		sort            = true
		highlight       = false
		analysis_scheme = "_en_default_"
	}

	wait_for_endpoints = false
	
	service_access_policies = <<EOF
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
