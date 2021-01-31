package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_vpc_dhcp_options", &resource.Sweeper{
		Name: "aws_vpc_dhcp_options",
		F:    testSweepVpcDhcpOptions,
	})
}

func testSweepVpcDhcpOptions(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ec2conn

	input := &ec2.DescribeDhcpOptionsInput{}

	err = conn.DescribeDhcpOptionsPages(input, func(page *ec2.DescribeDhcpOptionsOutput, lastPage bool) bool {
		for _, dhcpOption := range page.DhcpOptions {
			var defaultDomainNameFound, defaultDomainNameServersFound bool

			// This skips the default dhcp configurations so they don't get deleted
			for _, dhcpConfiguration := range dhcpOption.DhcpConfigurations {
				if aws.StringValue(dhcpConfiguration.Key) == "domain-name" {
					if len(dhcpConfiguration.Values) != 1 || dhcpConfiguration.Values[0] == nil {
						continue
					}

					if aws.StringValue(dhcpConfiguration.Values[0].Value) == resourceAwsEc2RegionalPrivateDnsSuffix(region) {
						defaultDomainNameFound = true
					}
				} else if aws.StringValue(dhcpConfiguration.Key) == "domain-name-servers" {
					if len(dhcpConfiguration.Values) != 1 || dhcpConfiguration.Values[0] == nil {
						continue
					}

					if aws.StringValue(dhcpConfiguration.Values[0].Value) == "AmazonProvidedDNS" {
						defaultDomainNameServersFound = true
					}
				}
			}

			if defaultDomainNameFound && defaultDomainNameServersFound {
				continue
			}

			input := &ec2.DeleteDhcpOptionsInput{
				DhcpOptionsId: dhcpOption.DhcpOptionsId,
			}

			_, err := conn.DeleteDhcpOptions(input)

			if err != nil {
				log.Printf("[ERROR] Error deleting EC2 DHCP Option (%s): %s", aws.StringValue(dhcpOption.DhcpOptionsId), err)
			}
		}
		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 DHCP Option sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing DHCP Options: %s", err)
	}

	return nil
}

func TestAccAWSDHCPOptions_basic(t *testing.T) {
	var d ec2.DhcpOptions
	resourceName := "aws_vpc_dhcp_options.test"
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDHCPOptionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPOptionsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDHCPOptionsExists(resourceName, &d),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`dhcp-options/dopt-.+`)),
					resource.TestCheckResourceAttr(resourceName, "domain_name", fmt.Sprintf("service.%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "domain_name_servers.0", "127.0.0.1"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_servers.1", "10.0.0.2"),
					resource.TestCheckResourceAttr(resourceName, "ntp_servers.0", "127.0.0.1"),
					resource.TestCheckResourceAttr(resourceName, "netbios_name_servers.0", "127.0.0.1"),
					resource.TestCheckResourceAttr(resourceName, "netbios_node_type", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
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

func TestAccAWSDHCPOptions_deleteOptions(t *testing.T) {
	var d ec2.DhcpOptions
	resourceName := "aws_vpc_dhcp_options.test"
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDHCPOptionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPOptionsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDHCPOptionsExists(resourceName, &d),
					testAccCheckDHCPOptionsDelete(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSDHCPOptions_tags(t *testing.T) {
	var d ec2.DhcpOptions
	resourceName := "aws_vpc_dhcp_options.test"
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDHCPOptionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPOptionsConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDHCPOptionsExists(resourceName, &d),
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
				Config: testAccDHCPOptionsConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDHCPOptionsExists(resourceName, &d),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDHCPOptionsConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDHCPOptionsExists(resourceName, &d),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSDHCPOptions_disappears(t *testing.T) {
	var d ec2.DhcpOptions
	resourceName := "aws_vpc_dhcp_options.test"
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDHCPOptionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPOptionsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDHCPOptionsExists(resourceName, &d),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsVpcDhcpOptions(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDHCPOptionsDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_dhcp_options" {
			continue
		}

		// Try to find the resource
		resp, err := conn.DescribeDhcpOptions(&ec2.DescribeDhcpOptionsInput{
			DhcpOptionsIds: []*string{
				aws.String(rs.Primary.ID),
			},
		})
		if isAWSErr(err, "InvalidDhcpOptionID.NotFound", "") {
			continue
		}
		if err == nil {
			if len(resp.DhcpOptions) > 0 {
				return fmt.Errorf("still exists")
			}

			return nil
		}

		if !isAWSErr(err, "InvalidDhcpOptionID.NotFound", "") {
			return err
		}
	}

	return nil
}

func testAccCheckDHCPOptionsExists(n string, d *ec2.DhcpOptions) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		resp, err := conn.DescribeDhcpOptions(&ec2.DescribeDhcpOptionsInput{
			DhcpOptionsIds: []*string{
				aws.String(rs.Primary.ID),
			},
		})
		if err != nil {
			return err
		}
		if len(resp.DhcpOptions) == 0 {
			return fmt.Errorf("DHCP Options not found")
		}

		*d = *resp.DhcpOptions[0]

		return nil
	}
}

func testAccCheckDHCPOptionsDelete(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		_, err := conn.DeleteDhcpOptions(&ec2.DeleteDhcpOptionsInput{
			DhcpOptionsId: aws.String(rs.Primary.ID),
		})

		return err
	}
}

func testAccDHCPOptionsConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc_dhcp_options" "test" {
  domain_name          = "service.%s"
  domain_name_servers  = ["127.0.0.1", "10.0.0.2"]
  ntp_servers          = ["127.0.0.1"]
  netbios_name_servers = ["127.0.0.1"]
  netbios_node_type    = 2
}
`, rName)
}

func testAccDHCPOptionsConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpc_dhcp_options" "test" {
  domain_name          = "service.%[1]s"
  domain_name_servers  = ["127.0.0.1", "10.0.0.2"]
  ntp_servers          = ["127.0.0.1"]
  netbios_name_servers = ["127.0.0.1"]
  netbios_node_type    = 2

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccDHCPOptionsConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc_dhcp_options" "test" {
  domain_name          = "service.%[1]s"
  domain_name_servers  = ["127.0.0.1", "10.0.0.2"]
  ntp_servers          = ["127.0.0.1"]
  netbios_name_servers = ["127.0.0.1"]
  netbios_node_type    = 2

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
