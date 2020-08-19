package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudsearch"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"strconv"
	"testing"
	"time"
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
					resource.TestCheckResourceAttr(resourceName, "domain_name", domainName),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "search.m1.small"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       false,
				ImportStateVerify: false,
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

		domainlist := cloudsearch.DescribeDomainsInput{
			DomainNames: []*string{
				aws.String(rs.Primary.Attributes["domain_name"]),
			},
		}

		resp, err := conn.DescribeDomains(&domainlist)
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
		// wait for the resource to go into deleted state.
		stateConf := &resource.StateChangeConf{
			Pending: []string{"false"},
			Target:  []string{"true"},
			Timeout: 20 * time.Minute,
			Refresh: func() (interface{}, string, error) {
				domainlist := cloudsearch.DescribeDomainsInput{
					DomainNames: []*string{
						aws.String(rs.Primary.Attributes["domain_name"]),
					},
				}

				resp, err := conn.DescribeDomains(&domainlist)
				if err != nil {
					return nil, "false", err
				}
				domain := resp.DomainStatusList[0]
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

func testAccAWSCloudSearchDomainConfig_basic(domainName string) string {
	return fmt.Sprintf(`
resource "aws_cloudsearch_domain" "test" {
	domain_name   = "%s"
	wait_for_endpoints = "false"
	indexes {
		name            = "headline"
		type            = "text"
		search          = true
		return          = true
		sort            = true
		highlight       = false
		analysis_scheme = "_en_default_"
		}
	
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
`, domainName)
}
