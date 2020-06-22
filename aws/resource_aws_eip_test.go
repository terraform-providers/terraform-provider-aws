package aws

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// Implement a test sweeper for EIPs.
// This will currently skip EIPs with associations,
// although we depend on aws_vpc to potentially have
// the majority of those associations removed.
func init() {
	resource.AddTestSweepers("aws_eip", &resource.Sweeper{
		Name: "aws_eip",
		Dependencies: []string{
			"aws_vpc",
		},
		F: testSweepEc2Eips,
	})
}

func testSweepEc2Eips(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ec2conn

	// There is currently no paginator or Marker/NextToken
	input := &ec2.DescribeAddressesInput{}

	output, err := conn.DescribeAddresses(input)

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 EIP sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing EC2 EIPs: %s", err)
	}

	if output == nil || len(output.Addresses) == 0 {
		log.Print("[DEBUG] No EC2 EIPs to sweep")
		return nil
	}

	for _, address := range output.Addresses {
		publicIP := aws.StringValue(address.PublicIp)

		if address.AssociationId != nil {
			log.Printf("[INFO] Skipping EC2 EIP (%s) with association: %s", publicIP, aws.StringValue(address.AssociationId))
			continue
		}

		input := &ec2.ReleaseAddressInput{}

		// The EC2 API is particular that you only specify one or the other
		// InvalidParameterCombination: You may specify public IP or allocation id, but not both in the same call
		if address.AllocationId != nil {
			input.AllocationId = address.AllocationId
		} else {
			input.PublicIp = address.PublicIp
		}

		log.Printf("[INFO] Releasing EC2 EIP: %s", publicIP)

		_, err := conn.ReleaseAddress(input)

		if err != nil {
			return fmt.Errorf("error releasing EC2 EIP (%s): %s", publicIP, err)
		}
	}

	return nil
}

func TestAccAWSEIP_Ec2Classic(t *testing.T) {
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	resourceName := "aws_eip.test"
	var conf ec2.Address

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPInstanceEc2Classic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists(resourceName, &conf),
					testAccCheckAWSEIPAttributes(&conf),
					testAccCheckAWSEIPPublicDNS(resourceName),
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

func TestAccAWSEIP_basic(t *testing.T) {
	var conf ec2.Address
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists(resourceName, &conf),
					testAccCheckAWSEIPAttributes(&conf),
					testAccCheckAWSEIPPublicDNS(resourceName),
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

func TestAccAWSEIP_instance(t *testing.T) {
	var conf ec2.Address
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPInstanceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists(resourceName, &conf),
					testAccCheckAWSEIPAttributes(&conf),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEIPInstanceConfig2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists(resourceName, &conf),
					testAccCheckAWSEIPAttributes(&conf),
				),
			},
		},
	})
}

func TestAccAWSEIP_networkInterface(t *testing.T) {
	var conf ec2.Address
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPNetworkInterfaceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists(resourceName, &conf),
					testAccCheckAWSEIPAttributes(&conf),
					testAccCheckAWSEIPAssociated(&conf),
					testAccCheckAWSEIPPrivateDNS(resourceName),
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

func TestAccAWSEIP_twoEIPsOneNetworkInterface(t *testing.T) {
	var one, two ec2.Address
	resourceName := "aws_eip.test"
	resourceName2 := "aws_eip.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPMultiNetworkInterfaceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists(resourceName, &one),
					testAccCheckAWSEIPAttributes(&one),
					testAccCheckAWSEIPAssociated(&one),
					testAccCheckAWSEIPExists(resourceName2, &two),
					testAccCheckAWSEIPAttributes(&two),
					testAccCheckAWSEIPAssociated(&two),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"associate_with_private_ip"},
			},
		},
	})
}

// This test is an expansion of TestAccAWSEIP_instance, by testing the
// associated Private EIPs of two instances
func TestAccAWSEIP_associated_user_private_ip(t *testing.T) {
	var one ec2.Address
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPInstanceConfig_associated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists(resourceName, &one),
					testAccCheckAWSEIPAttributes(&one),
					testAccCheckAWSEIPAssociated(&one),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"associate_with_private_ip"},
			},
			{
				Config: testAccAWSEIPInstanceConfig_associated_switch,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists(resourceName, &one),
					testAccCheckAWSEIPAttributes(&one),
					testAccCheckAWSEIPAssociated(&one),
				),
			},
		},
	})
}

// Regression test for https://github.com/hashicorp/terraform/issues/3429 (now
// https://github.com/terraform-providers/terraform-provider-aws/issues/42)
func TestAccAWSEIP_Instance_Reassociate(t *testing.T) {
	instanceResourceName := "aws_instance.test"
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIP_Instance(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "instance", instanceResourceName, "id"),
				),
			},
			{
				Config: testAccAWSEIP_Instance(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "instance", instanceResourceName, "id"),
				),
				Taint: []string{resourceName},
			},
		},
	})
}

func TestAccAWSEIP_disappears(t *testing.T) {
	var conf ec2.Address
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists(resourceName, &conf),
					testAccCheckAWSEIPDisappears(&conf),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSEIPAssociate_notAssociated(t *testing.T) {
	var conf ec2.Address
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPAssociate_not_associated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists(resourceName, &conf),
					testAccCheckAWSEIPAttributes(&conf),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEIPAssociate_associated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists(resourceName, &conf),
					testAccCheckAWSEIPAttributes(&conf),
					testAccCheckAWSEIPAssociated(&conf),
				),
			},
		},
	})
}

func TestAccAWSEIP_tags(t *testing.T) {
	var conf ec2.Address
	resourceName := "aws_eip.test"
	rName1 := fmt.Sprintf("%s-%d", t.Name(), acctest.RandInt())
	rName2 := fmt.Sprintf("%s-%d", t.Name(), acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPConfig_tags(rName1, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists(resourceName, &conf),
					testAccCheckAWSEIPAttributes(&conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.RandomName", rName1),
					resource.TestCheckResourceAttr(resourceName, "tags.TestName", t.Name()),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEIPConfig_tags(rName2, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists(resourceName, &conf),
					testAccCheckAWSEIPAttributes(&conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.RandomName", rName2),
					resource.TestCheckResourceAttr(resourceName, "tags.TestName", t.Name()),
				),
			},
		},
	})
}

func TestAccAWSEIP_PublicIpv4Pool_default(t *testing.T) {
	var conf ec2.Address
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPConfig_PublicIpv4Pool_default,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists(resourceName, &conf),
					testAccCheckAWSEIPAttributes(&conf),
					resource.TestCheckResourceAttr(resourceName, "public_ipv4_pool", "amazon"),
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

func TestAccAWSEIP_PublicIpv4Pool_custom(t *testing.T) {
	if os.Getenv("AWS_EC2_EIP_PUBLIC_IPV4_POOL") == "" {
		t.Skip("Environment variable AWS_EC2_EIP_PUBLIC_IPV4_POOL is not set")
	}

	var conf ec2.Address
	resourceName := "aws_eip.test"

	poolName := os.Getenv("AWS_EC2_EIP_PUBLIC_IPV4_POOL")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPConfig_PublicIpv4Pool_custom(poolName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists(resourceName, &conf),
					testAccCheckAWSEIPAttributes(&conf),
					resource.TestCheckResourceAttr(resourceName, "public_ipv4_pool", poolName),
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

func testAccCheckAWSEIPDisappears(v *ec2.Address) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_eip" {
				continue
			}

			_, err := conn.ReleaseAddress(&ec2.ReleaseAddressInput{
				AllocationId: aws.String(rs.Primary.ID),
			})
			return err
		}
		return nil
	}
}

func testAccCheckAWSEIPDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_eip" {
			continue
		}

		if strings.Contains(rs.Primary.ID, "eipalloc") {
			req := &ec2.DescribeAddressesInput{
				AllocationIds: []*string{aws.String(rs.Primary.ID)},
			}
			describe, err := conn.DescribeAddresses(req)
			if err != nil {
				// Verify the error is what we want
				if ae, ok := err.(awserr.Error); ok && ae.Code() == "InvalidAllocationID.NotFound" || ae.Code() == "InvalidAddress.NotFound" {
					continue
				}
				return err
			}

			if len(describe.Addresses) > 0 {
				return fmt.Errorf("still exists")
			}
		} else {
			req := &ec2.DescribeAddressesInput{
				PublicIps: []*string{aws.String(rs.Primary.ID)},
			}
			describe, err := conn.DescribeAddresses(req)
			if err != nil {
				// Verify the error is what we want
				if ae, ok := err.(awserr.Error); ok && ae.Code() == "InvalidAllocationID.NotFound" || ae.Code() == "InvalidAddress.NotFound" {
					continue
				}
				return err
			}

			if len(describe.Addresses) > 0 {
				return fmt.Errorf("still exists")
			}
		}
	}

	return nil
}

func TestAccAWSEIP_CustomerOwnedIpv4Pool(t *testing.T) {
	var conf ec2.Address
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPConfigCustomerOwnedIpv4Pool(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists(resourceName, &conf),
					resource.TestMatchResourceAttr(resourceName, "customer_owned_ipv4_pool", regexp.MustCompile(`^ipv4pool-coip-.+$`)),
					resource.TestMatchResourceAttr(resourceName, "customer_owned_ip", regexp.MustCompile(`\d+\.\d+\.\d+\.\d+`)),
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

func testAccCheckAWSEIPAttributes(conf *ec2.Address) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *conf.PublicIp == "" {
			return fmt.Errorf("empty public_ip")
		}

		return nil
	}
}

func testAccCheckAWSEIPAssociated(conf *ec2.Address) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if conf.AssociationId == nil || *conf.AssociationId == "" {
			return fmt.Errorf("empty association_id")
		}

		return nil
	}
}

func testAccCheckAWSEIPExists(n string, res *ec2.Address) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EIP ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		if strings.Contains(rs.Primary.ID, "eipalloc") {
			req := &ec2.DescribeAddressesInput{
				AllocationIds: []*string{aws.String(rs.Primary.ID)},
			}
			describe, err := conn.DescribeAddresses(req)
			if err != nil {
				return err
			}

			if len(describe.Addresses) != 1 ||
				*describe.Addresses[0].AllocationId != rs.Primary.ID {
				return fmt.Errorf("EIP not found")
			}
			*res = *describe.Addresses[0]

		} else {
			req := &ec2.DescribeAddressesInput{
				PublicIps: []*string{aws.String(rs.Primary.ID)},
			}
			describe, err := conn.DescribeAddresses(req)
			if err != nil {
				return err
			}

			if len(describe.Addresses) != 1 ||
				*describe.Addresses[0].PublicIp != rs.Primary.ID {
				return fmt.Errorf("EIP not found")
			}
			*res = *describe.Addresses[0]
		}

		return nil
	}
}

func testAccCheckAWSEIPPrivateDNS(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		privateIPDashed := strings.Replace(rs.Primary.Attributes["private_ip"], ".", "-", -1)
		privateDNS := rs.Primary.Attributes["private_dns"]
		expectedPrivateDNS := fmt.Sprintf("ip-%s.%s.compute.internal", privateIPDashed, testAccGetRegion())

		if testAccGetRegion() == "us-east-1" {
			expectedPrivateDNS = fmt.Sprintf("ip-%s.ec2.internal", privateIPDashed)
		}

		if privateDNS != expectedPrivateDNS {
			return fmt.Errorf("expected private_dns value (%s), received: %s", expectedPrivateDNS, privateDNS)
		}

		return nil
	}
}

func testAccCheckAWSEIPPublicDNS(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		publicIPDashed := strings.Replace(rs.Primary.Attributes["public_ip"], ".", "-", -1)
		publicDNS := rs.Primary.Attributes["public_dns"]
		expectedPublicDNS := fmt.Sprintf("ec2-%s.%s.compute.%s", publicIPDashed, testAccGetRegion(), testAccGetPartitionDNSSuffix())

		if testAccGetRegion() == "us-east-1" {
			expectedPublicDNS = fmt.Sprintf("ec2-%s.compute-1.%s", publicIPDashed, testAccGetPartitionDNSSuffix())
		}

		if publicDNS != expectedPublicDNS {
			return fmt.Errorf("expected public_dns value (%s), received: %s", expectedPublicDNS, publicDNS)
		}

		return nil
	}
}

const testAccAWSEIPConfig = `
resource "aws_eip" "test" {
}
`

func testAccAWSEIPConfig_tags(rName, testName string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  tags = {
    RandomName = "%[1]s"
    TestName   = "%[2]s"
  }
}
`, rName, testName)
}

const testAccAWSEIPConfig_PublicIpv4Pool_default = `
resource "aws_eip" "test" {
   vpc = true
}
`

func testAccAWSEIPConfig_PublicIpv4Pool_custom(poolName string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  vpc              = true
  public_ipv4_pool = "%s"
}
`, poolName)
}

const testAccAWSEIPInstanceEc2Classic = `
provider "aws" {
	region = "us-east-1"
}

data "aws_ami" "amzn-ami-minimal-pv" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name = "name"
    values = ["amzn-ami-minimal-pv-*"]
  }
  filter {
    name = "root-device-type"
    values = ["instance-store"]
  }
}

resource "aws_instance" "test" {
	ami = "${data.aws_ami.amzn-ami-minimal-pv.id}"
	instance_type = "m1.small"
	tags = {
		Name = "testAccAWSEIPInstanceEc2Classic"
	}
}

resource "aws_eip" "test" {
	instance = "${aws_instance.test.id}"
}
`

const testAccAWSEIPInstanceConfig = `
data "aws_ami" "amzn-ami-minimal-pv" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name = "name"
    values = ["amzn-ami-minimal-pv-*"]
  }
  filter {
    name = "root-device-type"
    values = ["instance-store"]
  }
}

resource "aws_instance" "test" {
	ami = "${data.aws_ami.amzn-ami-minimal-pv.id}"
	instance_type = "m1.small"
}

resource "aws_eip" "test" {
	instance = "${aws_instance.test.id}"
}
`

const testAccAWSEIPInstanceConfig2 = `
data "aws_ami" "amzn-ami-minimal-pv" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name = "name"
    values = ["amzn-ami-minimal-pv-*"]
  }
  filter {
    name = "root-device-type"
    values = ["instance-store"]
  }
}

resource "aws_instance" "test" {
	ami = "${data.aws_ami.amzn-ami-minimal-pv.id}"
	instance_type = "m1.small"
}

resource "aws_eip" "test" {
	instance = "${aws_instance.test.id}"
}
`

const testAccAWSEIPInstanceConfig_associated = `
data "aws_ami" "amzn-ami-minimal-hvm" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name = "name"
    values = ["amzn-ami-minimal-hvm-*"]
  }
  filter {
    name = "root-device-type"
    values = ["ebs"]
  }
}

resource "aws_vpc" "default" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-eip-instance-associated"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.default.id}"

  tags = {
    Name = "main"
  }
}

resource "aws_subnet" "tf_test_subnet" {
  vpc_id                  = "${aws_vpc.default.id}"
  cidr_block              = "10.0.0.0/24"
  map_public_ip_on_launch = true

  depends_on = ["aws_internet_gateway.gw"]

  tags = {
    Name = "tf-acc-eip-instance-associated"
  }
}

resource "aws_instance" "test" {
  ami           = "${data.aws_ami.amzn-ami-minimal-hvm.id}"
  instance_type = "t2.micro"

  private_ip = "10.0.0.12"
  subnet_id  = "${aws_subnet.tf_test_subnet.id}"

  tags = {
    Name = "test instance"
  }
}

resource "aws_instance" "test2" {
  ami = "${data.aws_ami.amzn-ami-minimal-hvm.id}"

  instance_type = "t2.micro"

  private_ip = "10.0.0.19"
  subnet_id  = "${aws_subnet.tf_test_subnet.id}"

  tags = {
    Name = "test2 instance"
  }
}

resource "aws_eip" "test" {
  vpc = true

  instance                  = "${aws_instance.test2.id}"
  associate_with_private_ip = "10.0.0.19"
}
`
const testAccAWSEIPInstanceConfig_associated_switch = `
data "aws_ami" "amzn-ami-minimal-hvm" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name = "name"
    values = ["amzn-ami-minimal-hvm-*"]
  }
  filter {
    name = "root-device-type"
    values = ["ebs"]
  }
}

resource "aws_vpc" "default" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-eip-instance-associated"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.default.id}"

  tags = {
    Name = "main"
  }
}

resource "aws_subnet" "tf_test_subnet" {
  vpc_id                  = "${aws_vpc.default.id}"
  cidr_block              = "10.0.0.0/24"
  map_public_ip_on_launch = true

  depends_on = ["aws_internet_gateway.gw"]

  tags = {
    Name = "tf-acc-eip-instance-associated"
  }
}

resource "aws_instance" "test" {
  ami           = "${data.aws_ami.amzn-ami-minimal-hvm.id}"
  instance_type = "t2.micro"

  private_ip = "10.0.0.12"
  subnet_id  = "${aws_subnet.tf_test_subnet.id}"

  tags = {
    Name = "test instance"
  }
}

resource "aws_instance" "test2" {
  ami = "${data.aws_ami.amzn-ami-minimal-hvm.id}"

  instance_type = "t2.micro"

  private_ip = "10.0.0.19"
  subnet_id  = "${aws_subnet.tf_test_subnet.id}"

  tags = {
    Name = "test2 instance"
  }
}

resource "aws_eip" "test" {
  vpc = true

  instance                  = "${aws_instance.test.id}"
  associate_with_private_ip = "10.0.0.12"
}
`

const testAccAWSEIPNetworkInterfaceConfig = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
	cidr_block = "10.0.0.0/24"
	tags = {
		Name = "terraform-testacc-eip-network-interface"
	}
}

resource "aws_internet_gateway" "test" {
	vpc_id = "${aws_vpc.test.id}"
}

resource "aws_subnet" "test" {
  vpc_id = "${aws_vpc.test.id}"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  cidr_block = "10.0.0.0/24"
  tags = {
	Name = "tf-acc-eip-network-interface"
  }
}

resource "aws_network_interface" "test" {
  subnet_id = "${aws_subnet.test.id}"
	private_ips = ["10.0.0.10"]
  security_groups = [ "${aws_vpc.test.default_security_group_id}" ]
}

resource "aws_eip" "test" {
	vpc = "true"
	network_interface = "${aws_network_interface.test.id}"
	depends_on = ["aws_internet_gateway.test"]
}
`

const testAccAWSEIPMultiNetworkInterfaceConfig = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}
  
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/24"
	tags = {
		Name = "terraform-testacc-eip-multi-network-interface"
	}
}

resource "aws_internet_gateway" "test" {
  vpc_id = "${aws_vpc.test.id}"
}

resource "aws_subnet" "test" {
  vpc_id            = "${aws_vpc.test.id}"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  cidr_block        = "10.0.0.0/24"
  tags = {
	Name = "tf-acc-eip-multi-network-interface"
  }
}

resource "aws_network_interface" "test" {
  subnet_id       = "${aws_subnet.test.id}"
  private_ips     = ["10.0.0.10", "10.0.0.11"]
  security_groups = ["${aws_vpc.test.default_security_group_id}"]
}

resource "aws_eip" "test" {
  vpc                       = "true"
  network_interface         = "${aws_network_interface.test.id}"
  associate_with_private_ip = "10.0.0.10"
  depends_on                = ["aws_internet_gateway.test"]
}

resource "aws_eip" "test2" {
  vpc                       = "true"
  network_interface         = "${aws_network_interface.test.id}"
  associate_with_private_ip = "10.0.0.11"
  depends_on                = ["aws_internet_gateway.test"]
}
`

func testAccAWSEIP_Instance() string {
	return fmt.Sprintf(`
data "aws_ami" "amzn-ami-minimal-hvm-ebs" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name = "name"
    values = ["amzn-ami-minimal-hvm-*"]
  }

  filter {
    name = "root-device-type"
    values = ["ebs"]
  }
}

data "aws_ec2_instance_type_offering" "available" {
  filter {
    name   = "instance-type"
    values = ["t3.micro", "t2.micro"]
  }

  filter {
    name   = "location"
    values = [aws_subnet.test.availability_zone]
  }

  location_type            = "availability-zone"
  preferred_instance_types = ["t3.micro", "t2.micro"]
}

resource "aws_eip" "test" {
  instance = aws_instance.test.id
  vpc      = true
}

resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  associate_public_ip_address = true
  instance_type               = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id                   = aws_subnet.test.id

  tags = {
    Name = "testAccAWSEIP_Instance"
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-eip-disassociate"
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = "tf-acc-eip-disassociate"
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }
}

resource "aws_route_table_association" "test" {
  subnet_id      = aws_subnet.test.id
  route_table_id = aws_route_table.test.id
}
`)
}

const testAccAWSEIPAssociate_not_associated = `
data "aws_ami" "amzn-ami-minimal-pv" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name = "name"
    values = ["amzn-ami-minimal-pv-*"]
  }
  filter {
    name = "root-device-type"
    values = ["instance-store"]
  }
}

resource "aws_instance" "test" {
	ami = "${data.aws_ami.amzn-ami-minimal-pv.id}"
	instance_type = "m1.small"
}

resource "aws_eip" "test" {
}
`

const testAccAWSEIPAssociate_associated = `
data "aws_ami" "amzn-ami-minimal-pv" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name = "name"
    values = ["amzn-ami-minimal-pv-*"]
  }
  filter {
    name = "root-device-type"
    values = ["instance-store"]
  }
}

resource "aws_instance" "test" {
	ami = "${data.aws_ami.amzn-ami-minimal-pv.id}"
	instance_type = "m1.small"
}

resource "aws_eip" "test" {
	instance = "${aws_instance.test.id}"
}
`

func testAccAWSEIPConfigCustomerOwnedIpv4Pool() string {
	return `
data "aws_ec2_coip_pools" "test" {}

resource "aws_eip" "test" {
  customer_owned_ipv4_pool = tolist(data.aws_ec2_coip_pools.test.pool_ids)[0]
  vpc                      = true
}
`
}
