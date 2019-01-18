package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSDirectoryServiceConditionForwarder_basic(t *testing.T) {
	resourceName := "aws_directory_service_conditional_forwarder.fwd"

	ip1, ip2, ip3 := "8.8.8.8", "1.1.1.1", "8.8.4.4"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDirectoryServiceConditionalForwarderDestroy,
		Steps: []resource.TestStep{
			// test create
			{
				Config: testAccDirectoryServiceConditionalForwarderConfig(ip1, ip2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDirectoryServiceConditionalForwarderExists(
						resourceName,
						[]string{ip1, ip2},
					),
				),
			},
			// test update
			{
				Config: testAccDirectoryServiceConditionalForwarderConfig(ip1, ip3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDirectoryServiceConditionalForwarderExists(
						resourceName,
						[]string{ip1, ip3},
					),
				),
			},
			// test import
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsDirectoryServiceConditionalForwarderDestroy(s *terraform.State) error {
	dsconn := testAccProvider.Meta().(*AWSClient).dsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_directory_service_conditional_forwarder" {
			continue
		}

		directoryId, domainName, err := parseDSConditionalForwarderId(rs.Primary.ID)
		if err != nil {
			return err
		}

		res, err := dsconn.DescribeConditionalForwarders(&directoryservice.DescribeConditionalForwardersInput{
			DirectoryId:       aws.String(directoryId),
			RemoteDomainNames: []*string{aws.String(domainName)},
		})

		if err != nil {
			if isAWSErr(err, directoryservice.ErrCodeEntityDoesNotExistException, "") {
				return nil
			}
			return err
		}

		if len(res.ConditionalForwarders) > 0 {
			return fmt.Errorf("Expected AWS Directory Service Conditional Forwarder to be gone, but was still found")
		}

		return nil
	}

	return fmt.Errorf("Default error in Service Directory Test")
}

func testAccCheckAwsDirectoryServiceConditionalForwarderExists(name string, dnsIps []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		directoryId, domainName, err := parseDSConditionalForwarderId(rs.Primary.ID)
		if err != nil {
			return err
		}

		dsconn := testAccProvider.Meta().(*AWSClient).dsconn

		res, err := dsconn.DescribeConditionalForwarders(&directoryservice.DescribeConditionalForwardersInput{
			DirectoryId:       aws.String(directoryId),
			RemoteDomainNames: []*string{aws.String(domainName)},
		})

		if err != nil {
			return err
		}

		if len(res.ConditionalForwarders) == 0 {
			return fmt.Errorf("No Conditional Fowrwarder found")
		}

		cfd := res.ConditionalForwarders[0]

		if dnsIps != nil {
			if len(dnsIps) != len(cfd.DnsIpAddrs) {
				return fmt.Errorf("DnsIpAddrs length mismatch")
			}

			for k, v := range cfd.DnsIpAddrs {
				if *v != dnsIps[k] {
					return fmt.Errorf("DnsIp mismatch, '%s' != '%s' at index '%d'", *v, dnsIps[k], k)
				}
			}
		}

		return nil
	}
}

func testAccDirectoryServiceConditionalForwarderConfig(ip1, ip2 string) string {
	return fmt.Sprintf(`
resource "aws_directory_service_directory" "bar" {
  name = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  type = "MicrosoftAD"
  edition = "Standard"

  vpc_settings {
    vpc_id = "${aws_vpc.main.id}"
    subnet_ids = ["${aws_subnet.foo.id}", "${aws_subnet.bar.id}"]
  }

  tags = {
    Name = "terraform-testacc-directory-service-conditional-forwarder"
  }
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-directory-service-conditional-forwarder"
  }
}

resource "aws_subnet" "foo" {
  vpc_id = "${aws_vpc.main.id}"
  availability_zone = "us-west-2a"
  cidr_block = "10.0.1.0/24"
  tags = {
    Name = "terraform-testacc-directory-service-conditional-forwarder"
  }
}

resource "aws_subnet" "bar" {
  vpc_id = "${aws_vpc.main.id}"
  availability_zone = "us-west-2b"
  cidr_block = "10.0.2.0/24"
  tags = {
    Name = "terraform-testacc-directory-service-conditional-forwarder"
  }
}

resource "aws_directory_service_conditional_forwarder" "fwd" {
  directory_id = "${aws_directory_service_directory.bar.id}"

  remote_domain_name = "test.example.com"

  dns_ips = [
    "%s",
    "%s",
  ]
}
`, ip1, ip2)
}
