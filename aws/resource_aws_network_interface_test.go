package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_network_interface", &resource.Sweeper{
		Name: "aws_network_interface",
		F:    testSweepEc2NetworkInterfaces,
		Dependencies: []string{
			"aws_instance",
		},
	})
}

func testSweepEc2NetworkInterfaces(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ec2conn

	err = conn.DescribeNetworkInterfacesPages(&ec2.DescribeNetworkInterfacesInput{}, func(page *ec2.DescribeNetworkInterfacesOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, networkInterface := range page.NetworkInterfaces {
			id := aws.StringValue(networkInterface.NetworkInterfaceId)

			if aws.StringValue(networkInterface.Status) != ec2.NetworkInterfaceStatusAvailable {
				log.Printf("[INFO] Skipping EC2 Network Interface in unavailable (%s) status: %s", aws.StringValue(networkInterface.Status), id)
				continue
			}

			input := &ec2.DeleteNetworkInterfaceInput{
				NetworkInterfaceId: aws.String(id),
			}

			log.Printf("[INFO] Deleting EC2 Network Interface: %s", id)
			_, err := conn.DeleteNetworkInterface(input)

			if err != nil {
				log.Printf("[ERROR] Error deleting EC2 Network Interface (%s): %s", id, err)
			}
		}

		return !isLast
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Network Interface sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error retrieving EC2 Network Interfaces: %s", err)
	}

	return nil
}

func TestAccAWSENI_basic(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSENIConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					testAccCheckAWSENIAttributes(&conf),
					resource.TestCheckResourceAttr(resourceName, "private_ips.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "private_dns_name"),
					resource.TestCheckResourceAttrSet(resourceName, "mac_address"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
					testAccCheckAWSENIAvailabilityZone("data.aws_availability_zones.available", "names.0", &conf),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
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

func TestAccAWSENI_ipv6(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSENIIPV6Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					testAccCheckAWSENIAttributes(&conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSENIIPV6MultipleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					testAccCheckAWSENIAttributes(&conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", "2"),
				),
			},
			{
				Config: testAccAWSENIIPV6Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					testAccCheckAWSENIAttributes(&conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSENI_tags(t *testing.T) {
	resourceName := "aws_network_interface.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	var conf ec2.NetworkInterface

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSENITagsConfig1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
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
				Config: testAccAWSENITagsConfig2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSENITagsConfig1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSENI_ipv6_count(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSENIIPV6CountConfig(1, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSENIIPV6CountConfig(2, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "2"),
				),
			},
			{
				Config: testAccAWSENIIPV6CountConfig(0, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "0"),
				),
			},
			{
				Config: testAccAWSENIIPV6CountConfig(1, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "1"),
				),
			},
		},
	})
}

func TestAccAWSENI_disappears(t *testing.T) {
	var networkInterface ec2.NetworkInterface
	resourceName := "aws_network_interface.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSENIConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &networkInterface),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsNetworkInterface(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSENI_updatedDescription(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSENIConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSENIConfigUpdatedDescription(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated ENI Description"),
				),
			},
		},
	})
}

func TestAccAWSENI_attached(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSENIConfigWithAttachment(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					testAccCheckAWSENIAttributesWithAttachment(&conf),
					testAccCheckAWSENIAvailabilityZone("data.aws_availability_zones.available", "names.0", &conf),
					resource.TestCheckResourceAttr(resourceName, "private_ips.#", "1"),
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

func TestAccAWSENI_ignoreExternalAttachment(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSENIConfigExternalAttachment(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					testAccCheckAWSENIAttributes(&conf),
					testAccCheckAWSENIAvailabilityZone("data.aws_availability_zones.available", "names.0", &conf),
					testAccCheckAWSENIMakeExternalAttachment("aws_instance.test", &conf),
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

func TestAccAWSENI_sourceDestCheck(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSENIConfigWithSourceDestCheck(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "source_dest_check", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSENIConfigWithSourceDestCheck(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "source_dest_check", "true"),
				),
			},
			{
				Config: testAccAWSENIConfigWithSourceDestCheck(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "source_dest_check", "false"),
				),
			},
		},
	})
}

func TestAccAWSENI_computedIPs(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSENIConfigWithNoPrivateIPs(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "private_ips.#", "1"),
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

func TestAccAWSENI_PrivateIpsCount(t *testing.T) {
	var networkInterface1, networkInterface2, networkInterface3, networkInterface4 ec2.NetworkInterface
	resourceName := "aws_network_interface.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSENIConfigPrivateIpsCount(1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &networkInterface1),
					resource.TestCheckResourceAttr(resourceName, "private_ips_count", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSENIConfigPrivateIpsCount(2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &networkInterface2),
					resource.TestCheckResourceAttr(resourceName, "private_ips_count", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSENIConfigPrivateIpsCount(0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &networkInterface3),
					resource.TestCheckResourceAttr(resourceName, "private_ips_count", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSENIConfigPrivateIpsCount(1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &networkInterface4),
					resource.TestCheckResourceAttr(resourceName, "private_ips_count", "1"),
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

func testAccCheckAWSENIExists(n string, res *ec2.NetworkInterface) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ENI ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		input := &ec2.DescribeNetworkInterfacesInput{
			NetworkInterfaceIds: []*string{aws.String(rs.Primary.ID)},
		}
		describeResp, err := conn.DescribeNetworkInterfaces(input)

		if err != nil {
			return err
		}

		if len(describeResp.NetworkInterfaces) != 1 ||
			*describeResp.NetworkInterfaces[0].NetworkInterfaceId != rs.Primary.ID {
			return fmt.Errorf("ENI not found")
		}

		*res = *describeResp.NetworkInterfaces[0]

		return nil
	}
}

func testAccCheckAWSENIAttributes(conf *ec2.NetworkInterface) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if conf.Attachment != nil {
			return fmt.Errorf("expected attachment to be nil")
		}

		if len(conf.Groups) != 1 && *conf.Groups[0].GroupName != "foo" {
			return fmt.Errorf("expected security group to be foo, but was %#v", conf.Groups)
		}

		if *conf.PrivateIpAddress != "172.16.10.100" {
			return fmt.Errorf("expected private ip to be 172.16.10.100, but was %s", *conf.PrivateIpAddress)
		}

		expectedPrivateDnsName := fmt.Sprintf("ip-%s.%s", resourceAwsEc2DashIP(*conf.PrivateIpAddress), resourceAwsEc2RegionalPrivateDnsSuffix(testAccGetRegion()))
		if *conf.PrivateDnsName != expectedPrivateDnsName {
			return fmt.Errorf("expected private dns name to be %s, but was %s", expectedPrivateDnsName, *conf.PrivateDnsName)
		}

		if len(*conf.MacAddress) == 0 {
			return fmt.Errorf("expected mac_address to be set")
		}

		if !*conf.SourceDestCheck {
			return fmt.Errorf("expected source_dest_check to be true, but was %t", *conf.SourceDestCheck)
		}

		return nil
	}
}

func testAccCheckAWSENIAvailabilityZone(name, attr string, conf *ec2.NetworkInterface) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok || rs.Primary.ID == "" {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.Attributes[attr] != *conf.AvailabilityZone {
			return fmt.Errorf("%s", fmt.Sprintf("expected %s, found %s", rs.Primary.Attributes[attr], *conf.AvailabilityZone))
		}

		return nil
	}
}

func testAccCheckAWSENIAttributesWithAttachment(conf *ec2.NetworkInterface) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if conf.Attachment == nil {
			return fmt.Errorf("expected attachment to be set, but was nil")
		}

		if *conf.Attachment.DeviceIndex != 1 {
			return fmt.Errorf("expected attachment device index to be 1, but was %d", *conf.Attachment.DeviceIndex)
		}

		if len(conf.Groups) != 1 && *conf.Groups[0].GroupName != "foo" {
			return fmt.Errorf("expected security group to be foo, but was %#v", conf.Groups)
		}

		if *conf.PrivateIpAddress != "172.16.10.100" {
			return fmt.Errorf("expected private ip to be 172.16.10.100, but was %s", *conf.PrivateIpAddress)
		}

		expectedPrivateDnsName := fmt.Sprintf("ip-%s.%s", resourceAwsEc2DashIP(*conf.PrivateIpAddress), resourceAwsEc2RegionalPrivateDnsSuffix(testAccGetRegion()))
		if *conf.PrivateDnsName != expectedPrivateDnsName {
			return fmt.Errorf("expected private dns name to be %s, but was %s", expectedPrivateDnsName, *conf.PrivateDnsName)
		}

		return nil
	}
}

func testAccCheckAWSENIDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_network_interface" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		input := &ec2.DescribeNetworkInterfacesInput{
			NetworkInterfaceIds: []*string{aws.String(rs.Primary.ID)},
		}
		_, err := conn.DescribeNetworkInterfaces(input)

		if err != nil {
			if isAWSErr(err, "InvalidNetworkInterfaceID.NotFound", "") {
				return nil
			}

			return err
		}
	}

	return nil
}

func testAccCheckAWSENIMakeExternalAttachment(n string, conf *ec2.NetworkInterface) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok || rs.Primary.ID == "" {
			return fmt.Errorf("Not found: %s", n)
		}
		input := &ec2.AttachNetworkInterfaceInput{
			DeviceIndex:        aws.Int64(1),
			InstanceId:         aws.String(rs.Primary.ID),
			NetworkInterfaceId: conf.NetworkInterfaceId,
		}
		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		_, err := conn.AttachNetworkInterface(input)
		if err != nil {
			return fmt.Errorf("Error attaching ENI: %s", err)
		}
		return nil
	}
}

func testAccAWSENIConfig() string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), `
resource "aws_vpc" "test" {
  cidr_block           = "172.16.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "tf-acc-network-interface"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.16.10.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  tags = {
    Name = "tf-acc-network-interface"
  }
}

resource "aws_security_group" "test" {
  vpc_id      = aws_vpc.test.id
  description = "test"
  name        = "tf-acc-network-interface"

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  tags = {
    Name = "tf-acc-network-interface"
  }
}

resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["172.16.10.100"]
  security_groups = [aws_security_group.test.id]
  description     = "Managed by Terraform"
}
`)
}

func testAccAWSENIIPV6ConfigBase(rName string) string {
	return testAccAvailableAZsNoOptInConfig() + fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "172.16.0.0/16"
  assign_generated_ipv6_cidr_block = true
  enable_dns_hostnames             = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.16.10.0/24"
  ipv6_cidr_block   = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 16)
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  vpc_id      = aws_vpc.test.id
  description = "test"
  name        = %[1]q

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSENIIPV6Config(rName string) string {
	return composeConfig(testAccAWSENIIPV6ConfigBase(rName), `
resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["172.16.10.100"]
  ipv6_addresses  = [cidrhost(aws_subnet.test.ipv6_cidr_block, 4)]
  security_groups = [aws_security_group.test.id]
  description     = "Managed by Terraform"
}
`)
}

func testAccAWSENIIPV6MultipleConfig(rName string) string {
	return composeConfig(testAccAWSENIIPV6ConfigBase(rName), `
resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["172.16.10.100"]
  ipv6_addresses  = [cidrhost(aws_subnet.test.ipv6_cidr_block, 4), cidrhost(aws_subnet.test.ipv6_cidr_block, 8)]
  security_groups = [aws_security_group.test.id]
  description     = "Managed by Terraform"
}
`)
}

func testAccAWSENIIPV6CountConfig(ipCount int, rName string) string {
	return testAccAWSENIIPV6ConfigBase(rName) + fmt.Sprintf(`
resource "aws_network_interface" "test" {
  subnet_id          = aws_subnet.test.id
  private_ips        = ["172.16.10.100"]
  ipv6_address_count = %[1]d
  security_groups    = [aws_security_group.test.id]
  description        = "Managed by Terraform"
}
`, ipCount)
}

func testAccAWSENIConfigUpdatedDescription() string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), `
resource "aws_vpc" "test" {
  cidr_block           = "172.16.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-network-interface"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.16.10.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  tags = {
    Name = "tf-acc-network-interface"
  }
}

resource "aws_security_group" "test" {
  vpc_id      = aws_vpc.test.id
  description = "test"
  name        = "tf-acc-network-interface"

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }
}

resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["172.16.10.100"]
  security_groups = [aws_security_group.test.id]
  description     = "Updated ENI Description"
}
`)
}

func testAccAWSENIConfigWithSourceDestCheck(enabled bool) string {
	return testAccAvailableAZsNoOptInConfig() + fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "172.16.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-network-interface-w-source-dest-check"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.16.10.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-network-interface-w-source-dest-check"
  }
}

resource "aws_network_interface" "test" {
  subnet_id         = aws_subnet.test.id
  source_dest_check = %[1]t
  private_ips       = ["172.16.10.100"]

  tags = {
    Name = "tf-acc-network-interface-w-source-dest-check"
  }
}
`, enabled)
}

func testAccAWSENIConfigWithNoPrivateIPs() string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), `
resource "aws_vpc" "test" {
  cidr_block           = "172.16.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-network-interface-w-no-private-ips"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.16.10.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-network-interface-w-no-private-ips"
  }
}

resource "aws_network_interface" "test" {
  subnet_id         = aws_subnet.test.id
  source_dest_check = false
}
`)
}

func testAccAWSENIConfigWithAttachment() string {
	return composeConfig(testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableEc2InstanceTypeForRegion("t3.micro", "t2.micro"),
		testAccAvailableAZsNoOptInConfig(), `
resource "aws_vpc" "test" {
  cidr_block           = "172.16.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-network-interface-w-attachment"
  }
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.16.10.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-network-interface-w-attachment-test"
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.16.11.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-network-interface-w-attachment-test"
  }
}

resource "aws_security_group" "test" {
  vpc_id      = aws_vpc.test.id
  description = "test"
  name        = "tf-acc-network-interface-w-attachment-test"
}

resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type               = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id                   = aws_subnet.test2.id
  associate_public_ip_address = false
  private_ip                  = "172.16.11.50"

  tags = {
    Name = "test-tf-eni-test"
  }
}

resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test1.id
  private_ips     = ["172.16.10.100"]
  security_groups = [aws_security_group.test.id]

  attachment {
    instance     = aws_instance.test.id
    device_index = 1
  }

  tags = {
    Name = "test_interface"
  }
}
`)
}

func testAccAWSENIConfigExternalAttachment() string {
	return composeConfig(testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableEc2InstanceTypeForRegion("t3.micro", "t2.micro"),
		testAccAvailableAZsNoOptInConfig(), `
resource "aws_vpc" "test" {
  cidr_block           = "172.16.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-network-interface-external-attachment"
  }
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.16.10.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-network-interface-external-attachment-test"
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.16.11.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-network-interface-external-attachment-test"
  }
}

resource "aws_security_group" "test" {
  vpc_id      = aws_vpc.test.id
  description = "test"
  name        = "tf-acc-network-interface-external-attachment-test"
}

resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type               = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id                   = aws_subnet.test2.id
  associate_public_ip_address = false
  private_ip                  = "172.16.11.50"

  tags = {
    Name = "tf-eni-test"
  }
}

resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test1.id
  private_ips     = ["172.16.10.100"]
  security_groups = [aws_security_group.test.id]

  tags = {
    Name = "test_interface"
  }
}
`)
}

func testAccAWSENIConfigPrivateIpsCount(privateIpsCount int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-network-interface-private-ips-count"
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-network-interface-private-ips-count"
  }
}

resource "aws_network_interface" "test" {
  private_ips_count = %[1]d
  subnet_id         = aws_subnet.test.id
}
`, privateIpsCount)
}

func testAccAWSENITagsConfig1(rName, tagKey1, tagValue1 string) string {
	return testAccAvailableAZsNoOptInConfig() + fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "172.16.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.16.10.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  vpc_id      = aws_vpc.test.id
  description = %[1]q
  name        = %[1]q

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["172.16.10.100"]
  security_groups = [aws_security_group.test.id]
  description     = "Managed by Terraform"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSENITagsConfig2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAvailableAZsNoOptInConfig() + fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "172.16.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.16.10.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  vpc_id      = aws_vpc.test.id
  description = %[1]q
  name        = %[1]q

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["172.16.10.100"]
  security_groups = [aws_security_group.test.id]
  description     = "Managed by Terraform"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
