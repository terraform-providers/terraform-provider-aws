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

	err = conn.DescribeNetworkInterfacesPages(&ec2.DescribeNetworkInterfacesInput{}, func(page *ec2.DescribeNetworkInterfacesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
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

		return !lastPage
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
	subnetResourceName := "aws_subnet.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSENIConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "attachment.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "interface_type", "interface"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "mac_address"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					testAccCheckResourceAttrPrivateDnsName(resourceName, "private_dns_name", &conf.PrivateIpAddress),
					resource.TestCheckResourceAttrSet(resourceName, "private_ip"),
					resource.TestCheckResourceAttr(resourceName, "private_ips.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_dest_check", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAWSENI_IPv6(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSENIConfigIPV6(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
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
				Config: testAccAWSENIConfigIPV6Multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", "2"),
				),
			},
			{
				Config: testAccAWSENIConfigIPV6(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSENI_Tags(t *testing.T) {
	resourceName := "aws_network_interface.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	var conf ec2.NetworkInterface

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSENIConfigTags1(rName, "key1", "value1"),
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
				Config: testAccAWSENIConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSENIConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSENI_IPv6Count(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSENIConfigIPV6Count(rName, 1),
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
				Config: testAccAWSENIConfigIPV6Count(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "2"),
				),
			},
			{
				Config: testAccAWSENIConfigIPV6Count(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "0"),
				),
			},
			{
				Config: testAccAWSENIConfigIPV6Count(rName, 1),
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
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSENIConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &networkInterface),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsNetworkInterface(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSENI_Description(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	subnetResourceName := "aws_subnet.test"
	securityGroupResourceName := "aws_security_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSENIConfigDescription(rName, "description 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "attachment.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", "description 1"),
					resource.TestCheckResourceAttr(resourceName, "interface_type", "interface"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "mac_address"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttrSet(resourceName, "private_dns_name"),
					resource.TestCheckResourceAttr(resourceName, "private_ip", "172.16.10.100"),
					resource.TestCheckResourceAttr(resourceName, "private_ips.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "private_ips.*", "172.16.10.100"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_groups.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "source_dest_check", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSENIConfigDescription(rName, "description 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "attachment.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", "description 2"),
					resource.TestCheckResourceAttr(resourceName, "interface_type", "interface"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "mac_address"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttrSet(resourceName, "private_dns_name"),
					resource.TestCheckResourceAttr(resourceName, "private_ip", "172.16.10.100"),
					resource.TestCheckResourceAttr(resourceName, "private_ips.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "private_ips.*", "172.16.10.100"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_groups.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "source_dest_check", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccAWSENI_Attachment(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSENIConfigAttachment(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "attachment.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attachment.*", map[string]string{
						"device_index": "1",
					}),
					resource.TestCheckResourceAttr(resourceName, "private_ip", "172.16.10.100"),
					resource.TestCheckResourceAttr(resourceName, "private_ips.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "private_ips.*", "172.16.10.100"),
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

func TestAccAWSENI_IgnoreExternalAttachment(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSENIConfigExternalAttachment(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
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

func TestAccAWSENI_SourceDestCheck(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSENIConfigSourceDestCheck(rName, false),
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
				Config: testAccAWSENIConfigSourceDestCheck(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "source_dest_check", "true"),
				),
			},
			{
				Config: testAccAWSENIConfigSourceDestCheck(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "source_dest_check", "false"),
				),
			},
		},
	})
}

func TestAccAWSENI_PrivateIpsCount(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSENIConfigPrivateIpsCount(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "private_ips_count", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSENIConfigPrivateIpsCount(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "private_ips_count", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSENIConfigPrivateIpsCount(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "private_ips_count", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSENIConfigPrivateIpsCount(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
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

func TestAccAWSENI_InterfaceType_efa(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSENIConfigInterfaceType(rName, "efa"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "interface_type", "efa"),
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

		_, err := testAccProvider.Meta().(*AWSClient).ec2conn.AttachNetworkInterface(input)

		if err != nil {
			return fmt.Errorf("error attaching ENI: %w", err)
		}
		return nil
	}
}

func testAccAWSENIIPV4ConfigBase(rName string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
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
  vpc_id = aws_vpc.test.id
  name   = %[1]q

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
`, rName))
}

func testAccAWSENIIPV6ConfigBase(rName string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
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
  vpc_id = aws_vpc.test.id
  name   = %[1]q

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
`, rName))
}

func testAccAWSENIConfig(rName string) string {
	return composeConfig(testAccAWSENIIPV4ConfigBase(rName), `
resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id
}
`)
}

func testAccAWSENIConfigIPV6(rName string) string {
	return composeConfig(testAccAWSENIIPV6ConfigBase(rName), fmt.Sprintf(`
resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["172.16.10.100"]
  ipv6_addresses  = [cidrhost(aws_subnet.test.ipv6_cidr_block, 4)]
  security_groups = [aws_security_group.test.id]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccAWSENIConfigIPV6Multiple(rName string) string {
	return composeConfig(testAccAWSENIIPV6ConfigBase(rName), fmt.Sprintf(`
resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["172.16.10.100"]
  ipv6_addresses  = [cidrhost(aws_subnet.test.ipv6_cidr_block, 4), cidrhost(aws_subnet.test.ipv6_cidr_block, 8)]
  security_groups = [aws_security_group.test.id]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccAWSENIConfigIPV6Count(rName string, ipv6Count int) string {
	return composeConfig(testAccAWSENIIPV6ConfigBase(rName) + fmt.Sprintf(`
resource "aws_network_interface" "test" {
  subnet_id          = aws_subnet.test.id
  private_ips        = ["172.16.10.100"]
  ipv6_address_count = %[2]d
  security_groups    = [aws_security_group.test.id]

  tags = {
    Name = %[1]q
  }
}
`, rName, ipv6Count))
}

func testAccAWSENIConfigDescription(rName, description string) string {
	return composeConfig(testAccAWSENIIPV4ConfigBase(rName), fmt.Sprintf(`
resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["172.16.10.100"]
  security_groups = [aws_security_group.test.id]
  description     = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, description))
}

func testAccAWSENIConfigSourceDestCheck(rName string, sourceDestCheck bool) string {
	return composeConfig(testAccAWSENIIPV6ConfigBase(rName) + fmt.Sprintf(`
resource "aws_network_interface" "test" {
  subnet_id         = aws_subnet.test.id
  source_dest_check = %[2]t
  private_ips       = ["172.16.10.100"]

  tags = {
    Name = %[1]q
  }
}
`, rName, sourceDestCheck))
}

func testAccAWSENIConfigAttachment(rName string) string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableEc2InstanceTypeForRegion("t3.micro", "t2.micro"),
		testAccAWSENIIPV4ConfigBase(rName),
		fmt.Sprintf(`
resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.16.11.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type               = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id                   = aws_subnet.test2.id
  associate_public_ip_address = false
  private_ip                  = "172.16.11.50"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["172.16.10.100"]
  security_groups = [aws_security_group.test.id]

  attachment {
    instance     = aws_instance.test.id
    device_index = 1
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccAWSENIConfigExternalAttachment(rName string) string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableEc2InstanceTypeForRegion("t3.micro", "t2.micro"),
		testAccAWSENIIPV4ConfigBase(rName),
		fmt.Sprintf(`
resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.16.11.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type               = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id                   = aws_subnet.test2.id
  associate_public_ip_address = false
  private_ip                  = "172.16.11.50"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["172.16.10.100"]
  security_groups = [aws_security_group.test.id]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccAWSENIConfigPrivateIpsCount(rName string, privateIpsCount int) string {
	return composeConfig(testAccAWSENIIPV4ConfigBase(rName), fmt.Sprintf(`
resource "aws_network_interface" "test" {
  private_ips_count = %[2]d
  subnet_id         = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, privateIpsCount))
}

func testAccAWSENIConfigTags1(rName, tagKey1, tagValue1 string) string {
	return composeConfig(testAccAWSENIIPV4ConfigBase(rName), fmt.Sprintf(`
resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["172.16.10.100"]
  security_groups = [aws_security_group.test.id]

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccAWSENIConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return composeConfig(testAccAWSENIIPV4ConfigBase(rName), fmt.Sprintf(`
resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["172.16.10.100"]
  security_groups = [aws_security_group.test.id]

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccAWSENIConfigInterfaceType(rName, interfaceType string) string {
	return composeConfig(testAccAWSENIIPV4ConfigBase(rName), fmt.Sprintf(`
resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["172.16.10.100"]
  security_groups = [aws_security_group.test.id]
  interface_type  = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, interfaceType))
}
