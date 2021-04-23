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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

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
	resourceName := "aws_eip.test"
	var conf ec2.Address

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPInstanceEc2Classic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPEc2ClassicExists(resourceName, &conf),
					testAccCheckAWSEIPAttributes(&conf),
					testAccCheckAWSEIPPublicDNSEc2Classic(resourceName),
					resource.TestCheckResourceAttr(resourceName, "domain", ec2.DomainTypeStandard),
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
					resource.TestCheckResourceAttr(resourceName, "domain", ec2.DomainTypeVpc),
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
				Config: testAccAWSEIPInstanceConfig(),
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
					resource.TestCheckResourceAttr(resourceName, "domain", ec2.DomainTypeVpc),
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
					resource.TestCheckResourceAttr(resourceName, "domain", ec2.DomainTypeVpc),

					testAccCheckAWSEIPExists(resourceName2, &two),
					testAccCheckAWSEIPAttributes(&two),
					testAccCheckAWSEIPAssociated(&two),
					resource.TestCheckResourceAttr(resourceName2, "domain", ec2.DomainTypeVpc),
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
				Config: testAccAWSEIPInstanceConfig_associated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists(resourceName, &one),
					testAccCheckAWSEIPAttributes(&one),
					testAccCheckAWSEIPAssociated(&one),
					resource.TestCheckResourceAttr(resourceName, "domain", ec2.DomainTypeVpc),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"associate_with_private_ip"},
			},
			{
				Config: testAccAWSEIPInstanceConfig_associated_switch(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists(resourceName, &one),
					testAccCheckAWSEIPAttributes(&one),
					testAccCheckAWSEIPAssociated(&one),
					resource.TestCheckResourceAttr(resourceName, "domain", ec2.DomainTypeVpc),
				),
			},
		},
	})
}

// Regression test for https://github.com/hashicorp/terraform/issues/3429 (now
// https://github.com/hashicorp/terraform-provider-aws/issues/42)
func TestAccAWSEIP_Instance_Reassociate(t *testing.T) {
	instanceResourceName := "aws_instance.test"
	resourceName := "aws_eip.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIP_Instance(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "instance", instanceResourceName, "id"),
				),
			},
			{
				Config: testAccAWSEIP_Instance(rName),
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
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEip(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSEIP_notAssociated(t *testing.T) {
	var conf ec2.Address
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPAssociate_not_associated(),
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
				Config: testAccAWSEIPAssociate_associated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists(resourceName, &conf),
					testAccCheckAWSEIPAttributes(&conf),
					testAccCheckAWSEIPAssociated(&conf),
				),
			},
		},
	})
}

func TestAccAWSEIP_tags_Vpc(t *testing.T) {
	var conf ec2.Address
	resourceName := "aws_eip.test"
	rName1 := fmt.Sprintf("%s-%d", t.Name(), acctest.RandInt())
	rName2 := fmt.Sprintf("%s-%d", t.Name(), acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPConfig_tags(rName1, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists(resourceName, &conf),
					testAccCheckAWSEIPAttributes(&conf),
					resource.TestCheckResourceAttr(resourceName, "domain", ec2.DomainTypeVpc),
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

func TestAccAWSEIP_tags_Ec2Classic(t *testing.T) {
	rName1 := fmt.Sprintf("%s-%d", t.Name(), acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSEIPConfig_tags_Ec2Classic(rName1, t.Name()),
				ExpectError: regexp.MustCompile(`tags can not be set for an EIP in EC2 Classic`),
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
					resource.TestCheckResourceAttr(resourceName, "domain", ec2.DomainTypeVpc),
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

func TestAccAWSEIP_NetworkBorderGroup(t *testing.T) {
	var conf ec2.Address
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPConfigNetworkBorderGroup,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists(resourceName, &conf),
					testAccCheckAWSEIPAttributes(&conf),
					resource.TestCheckResourceAttr(resourceName, "public_ipv4_pool", "amazon"),
					resource.TestCheckResourceAttr(resourceName, "network_border_group", testAccGetRegion()),
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

func TestAccAWSEIP_CarrierIP(t *testing.T) {
	var conf ec2.Address
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t); testAccPreCheckAWSWavelengthZoneAvailable(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPConfigCarrierIP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "carrier_ip"),
					resource.TestCheckResourceAttrSet(resourceName, "network_border_group"),
					resource.TestCheckResourceAttr(resourceName, "public_ip", ""),
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
					resource.TestCheckResourceAttr(resourceName, "domain", ec2.DomainTypeStandard),
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

func testAccCheckAWSEIPEc2ClassicExists(n string, res *ec2.Address) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EIP ID is set")
		}

		conn := testAccProviderEc2Classic.Meta().(*AWSClient).ec2conn

		input := &ec2.DescribeAddressesInput{
			PublicIps: []*string{aws.String(rs.Primary.ID)},
		}

		describe, err := conn.DescribeAddresses(input)

		if err != nil {
			return fmt.Errorf("error describing EC2 Address (%s): %w", rs.Primary.ID, err)
		}

		if len(describe.Addresses) != 1 || aws.StringValue(describe.Addresses[0].PublicIp) != rs.Primary.ID {
			return fmt.Errorf("EC2 Address (%s) not found", rs.Primary.ID)
		}

		*res = *describe.Addresses[0]

		return nil
	}
}

func testAccCheckAWSEIPPrivateDNS(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		privateDNS := rs.Primary.Attributes["private_dns"]
		expectedPrivateDNS := fmt.Sprintf(
			"ip-%s.%s",
			resourceAwsEc2DashIP(rs.Primary.Attributes["private_ip"]),
			resourceAwsEc2RegionalPrivateDnsSuffix(testAccGetRegion()),
		)

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

		publicDNS := rs.Primary.Attributes["public_dns"]
		expectedPublicDNS := fmt.Sprintf(
			"ec2-%s.%s.%s",
			resourceAwsEc2DashIP(rs.Primary.Attributes["public_ip"]),
			resourceAwsEc2RegionalPublicDnsSuffix(testAccGetRegion()),
			testAccGetPartitionDNSSuffix(),
		)

		if publicDNS != expectedPublicDNS {
			return fmt.Errorf("expected public_dns value (%s), received: %s", expectedPublicDNS, publicDNS)
		}

		return nil
	}
}

func testAccCheckAWSEIPPublicDNSEc2Classic(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		publicDNS := rs.Primary.Attributes["public_dns"]
		expectedPublicDNS := fmt.Sprintf(
			"ec2-%s.%s.%s",
			resourceAwsEc2DashIP(rs.Primary.Attributes["public_ip"]),
			resourceAwsEc2RegionalPublicDnsSuffix(testAccGetEc2ClassicRegion()),
			testAccGetPartitionDNSSuffix(),
		)

		if publicDNS != expectedPublicDNS {
			return fmt.Errorf("expected public_dns value (%s), received: %s", expectedPublicDNS, publicDNS)
		}

		return nil
	}
}

const testAccAWSEIPConfig = `
resource "aws_eip" "test" {
  vpc = true
}
`

func testAccAWSEIPConfig_tags(rName, testName string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  vpc = true

  tags = {
    RandomName = "%[1]s"
    TestName   = "%[2]s"
  }
}
`, rName, testName)
}

func testAccAWSEIPConfig_tags_Ec2Classic(rName, testName string) string {
	return composeConfig(
		testAccEc2ClassicRegionProviderConfig(),
		fmt.Sprintf(`
resource "aws_eip" "test" {
  tags = {
    RandomName = "%[1]s"
    TestName   = "%[2]s"
  }
}
`, rName, testName))
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

func testAccAWSEIPInstanceEc2Classic() string {
	return composeConfig(
		testAccEc2ClassicRegionProviderConfig(),
		testAccLatestAmazonLinuxPvEbsAmiConfig(),
		testAccAvailableEc2InstanceTypeForRegion("t1.micro", "m3.medium", "m3.large", "c3.large", "r3.large"),
		`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-pv-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
}

resource "aws_eip" "test" {
  instance = aws_instance.test.id
}
`)
}

func testAccAWSEIPInstanceConfig() string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableEc2InstanceTypeForAvailabilityZone("aws_subnet.test.availability_zone", "t3.micro", "t2.micro"),
		testAccAvailableAZsNoOptInConfig(),
		`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  vpc_id            = aws_vpc.test.id
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id
}

resource "aws_eip" "test" {
  instance = aws_instance.test.id
  vpc      = true
}
`)
}

func testAccAWSEIPInstanceConfig_associated() string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableEc2InstanceTypeForAvailabilityZone("aws_subnet.test.availability_zone", "t3.micro", "t2.micro"), `
resource "aws_vpc" "default" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-eip-instance-associated"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.default.id

  tags = {
    Name = "main"
  }
}

resource "aws_subnet" "test" {
  vpc_id                  = aws_vpc.default.id
  availability_zone       = data.aws_availability_zones.available.names[0]
  cidr_block              = "10.0.0.0/24"
  map_public_ip_on_launch = true

  depends_on = [aws_internet_gateway.gw]

  tags = {
    Name = "tf-acc-eip-instance-associated"
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  private_ip = "10.0.0.12"
  subnet_id  = aws_subnet.test.id

  tags = {
    Name = "test instance"
  }
}

resource "aws_instance" "test2" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  private_ip = "10.0.0.19"
  subnet_id  = aws_subnet.test.id

  tags = {
    Name = "test2 instance"
  }
}

resource "aws_eip" "test" {
  vpc = true

  instance                  = aws_instance.test2.id
  associate_with_private_ip = "10.0.0.19"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}
`)
}

func testAccAWSEIPInstanceConfig_associated_switch() string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(), `
resource "aws_vpc" "default" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-eip-instance-associated"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.default.id

  tags = {
    Name = "main"
  }
}

resource "aws_subnet" "test" {
  vpc_id                  = aws_vpc.default.id
  cidr_block              = "10.0.0.0/24"
  map_public_ip_on_launch = true

  depends_on = [aws_internet_gateway.gw]

  tags = {
    Name = "tf-acc-eip-instance-associated"
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"

  private_ip = "10.0.0.12"
  subnet_id  = aws_subnet.test.id

  tags = {
    Name = "test instance"
  }
}

resource "aws_instance" "test2" {
  ami = data.aws_ami.amzn-ami-minimal-hvm-ebs.id

  instance_type = "t2.micro"

  private_ip = "10.0.0.19"
  subnet_id  = aws_subnet.test.id

  tags = {
    Name = "test2 instance"
  }
}

resource "aws_eip" "test" {
  vpc = true

  instance                  = aws_instance.test.id
  associate_with_private_ip = "10.0.0.12"
}
`)
}

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
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  tags = {
    Name = "tf-acc-eip-network-interface"
  }
}

resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["10.0.0.10"]
  security_groups = [aws_vpc.test.default_security_group_id]
}

resource "aws_eip" "test" {
  vpc               = "true"
  network_interface = aws_network_interface.test.id
  depends_on        = [aws_internet_gateway.test]
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
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  tags = {
    Name = "tf-acc-eip-multi-network-interface"
  }
}

resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["10.0.0.10", "10.0.0.11"]
  security_groups = [aws_vpc.test.default_security_group_id]
}

resource "aws_eip" "test" {
  vpc                       = "true"
  network_interface         = aws_network_interface.test.id
  associate_with_private_ip = "10.0.0.10"
  depends_on                = [aws_internet_gateway.test]
}

resource "aws_eip" "test2" {
  vpc                       = "true"
  network_interface         = aws_network_interface.test.id
  associate_with_private_ip = "10.0.0.11"
  depends_on                = [aws_internet_gateway.test]
}
`

func testAccAWSEIP_Instance(rName string) string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableEc2InstanceTypeForAvailabilityZone("aws_subnet.test.availability_zone", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_eip" "test" {
  instance = aws_instance.test.id
  vpc      = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  associate_public_ip_address = true
  instance_type               = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id                   = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  subnet_id      = aws_subnet.test.id
  route_table_id = aws_route_table.test.id
}
`, rName))
}

func testAccAWSEIPAssociate_not_associated() string {
	return composeConfig(
		testAccAvailableEc2InstanceTypeForAvailabilityZone("aws_subnet.test.availability_zone", "t3.micro", "t2.micro"),
		testAccAvailableAZsNoOptInConfig(),
		testAccLatestAmazonLinuxHvmEbsAmiConfig(), `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  vpc_id            = aws_vpc.test.id
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id
}

resource "aws_eip" "test" {
}
`)
}

func testAccAWSEIPAssociate_associated() string {
	return composeConfig(
		testAccAvailableEc2InstanceTypeForAvailabilityZone("aws_subnet.test.availability_zone", "t3.micro", "t2.micro"),
		testAccAvailableAZsNoOptInConfig(),
		testAccLatestAmazonLinuxHvmEbsAmiConfig(), `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  vpc_id            = aws_vpc.test.id
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id
}

resource "aws_eip" "test" {
  instance = aws_instance.test.id
  vpc      = true
}
`)
}

func testAccAWSEIPConfigCustomerOwnedIpv4Pool() string {
	return `
data "aws_ec2_coip_pools" "test" {}

resource "aws_eip" "test" {
  customer_owned_ipv4_pool = tolist(data.aws_ec2_coip_pools.test.pool_ids)[0]
  vpc                      = true
}
`
}

const testAccAWSEIPConfigNetworkBorderGroup = `
data "aws_region" current {}

resource "aws_eip" "test" {
  vpc                  = true
  network_border_group = data.aws_region.current.name
}
`

func testAccAWSEIPConfigCarrierIP(rName string) string {
	return composeConfig(
		testAccAvailableAZsWavelengthZonesDefaultExcludeConfig(),
		fmt.Sprintf(`
data "aws_availability_zone" "available" {
  name = data.aws_availability_zones.available.names[0]
}

resource "aws_eip" "test" {
  vpc                  = true
  network_border_group = data.aws_availability_zone.available.network_border_group

  tags = {
    Name = %[1]q
  }
}
`, rName))
}
