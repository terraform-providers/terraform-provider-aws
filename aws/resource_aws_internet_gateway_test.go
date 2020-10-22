package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_internet_gateway", &resource.Sweeper{
		Name: "aws_internet_gateway",
		Dependencies: []string{
			"aws_subnet",
		},
		F: testSweepInternetGateways,
	})
}

func testSweepInternetGateways(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ec2conn

	req := &ec2.DescribeInternetGatewaysInput{}
	resp, err := conn.DescribeInternetGateways(req)
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Internet Gateway sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error describing Internet Gateways: %s", err)
	}

	if len(resp.InternetGateways) == 0 {
		log.Print("[DEBUG] No AWS Internet Gateways to sweep")
		return nil
	}

	defaultVPCID := ""
	describeVpcsInput := &ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("isDefault"),
				Values: aws.StringSlice([]string{"true"}),
			},
		},
	}

	describeVpcsOutput, err := conn.DescribeVpcs(describeVpcsInput)

	if err != nil {
		return fmt.Errorf("Error describing VPCs: %s", err)
	}

	if describeVpcsOutput != nil && len(describeVpcsOutput.Vpcs) == 1 {
		defaultVPCID = aws.StringValue(describeVpcsOutput.Vpcs[0].VpcId)
	}

	for _, internetGateway := range resp.InternetGateways {
		isDefaultVPCInternetGateway := false

		for _, attachment := range internetGateway.Attachments {
			if aws.StringValue(attachment.VpcId) == defaultVPCID {
				isDefaultVPCInternetGateway = true
				break
			}

			input := &ec2.DetachInternetGatewayInput{
				InternetGatewayId: internetGateway.InternetGatewayId,
				VpcId:             attachment.VpcId,
			}

			log.Printf("[DEBUG] Detaching Internet Gateway: %s", input)
			_, err := conn.DetachInternetGateway(input)
			if err != nil {
				return fmt.Errorf("error detaching Internet Gateway (%s) from VPC (%s): %s", aws.StringValue(internetGateway.InternetGatewayId), aws.StringValue(attachment.VpcId), err)
			}

			stateConf := &resource.StateChangeConf{
				Pending: []string{ec2.AttachmentStatusDetaching},
				Target:  []string{ec2.AttachmentStatusDetached},
				Refresh: detachIGStateRefreshFunc(conn, aws.StringValue(internetGateway.InternetGatewayId), aws.StringValue(attachment.VpcId)),
				Timeout: 10 * time.Minute,
				Delay:   10 * time.Second,
			}

			log.Printf("[DEBUG] Waiting for Internet Gateway (%s) to detach from VPC (%s)", aws.StringValue(internetGateway.InternetGatewayId), aws.StringValue(attachment.VpcId))
			if _, err = stateConf.WaitForState(); err != nil {
				return fmt.Errorf("error waiting for VPN Gateway (%s) to detach from VPC (%s): %s", aws.StringValue(internetGateway.InternetGatewayId), aws.StringValue(attachment.VpcId), err)
			}
		}

		if isDefaultVPCInternetGateway {
			log.Printf("[DEBUG] Skipping Default VPC Internet Gateway: %s", aws.StringValue(internetGateway.InternetGatewayId))
			continue
		}

		input := &ec2.DeleteInternetGatewayInput{
			InternetGatewayId: internetGateway.InternetGatewayId,
		}

		log.Printf("[DEBUG] Deleting Internet Gateway: %s", input)
		_, err := conn.DeleteInternetGateway(input)
		if err != nil {
			return fmt.Errorf("error deleting Internet Gateway (%s): %s", aws.StringValue(internetGateway.InternetGatewayId), err)
		}
	}

	return nil
}

func TestAccAWSInternetGateway_basic(t *testing.T) {
	var v, v2 ec2.InternetGateway
	resourceName := "aws_internet_gateway.test"

	testNotEqual := func(*terraform.State) error {
		if len(v.Attachments) == 0 {
			return fmt.Errorf("IG A is not attached")
		}
		if len(v2.Attachments) == 0 {
			return fmt.Errorf("IG B is not attached")
		}

		id1 := v.Attachments[0].VpcId
		id2 := v2.Attachments[0].VpcId
		if id1 == id2 {
			return fmt.Errorf("Both attachment IDs are the same")
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInternetGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInternetGatewayConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInternetGatewayExists(resourceName, &v),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`internet-gateway/igw-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInternetGatewayConfigChangeVPC,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInternetGatewayExists(resourceName, &v2),
					testNotEqual,
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
				),
			},
		},
	})
}

func TestAccAWSInternetGateway_delete(t *testing.T) {
	var ig ec2.InternetGateway
	resourceName := "aws_internet_gateway.test"

	testDeleted := func(r string) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			_, ok := s.RootModule().Resources[r]
			if ok {
				return fmt.Errorf("Internet Gateway %q should have been deleted", r)
			}
			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInternetGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInternetGatewayConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInternetGatewayExists(resourceName, &ig)),
			},
			{
				Config: testAccNoInternetGatewayConfig,
				Check:  resource.ComposeTestCheckFunc(testDeleted(resourceName)),
			},
		},
	})
}

func TestAccAWSInternetGateway_tags(t *testing.T) {
	var v ec2.InternetGateway
	resourceName := "aws_internet_gateway.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInternetGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSInternetGatewayConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInternetGatewayExists(resourceName, &v),
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
				Config: testAccAWSInternetGatewayConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInternetGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSInternetGatewayConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInternetGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSInternetGateway_disappears(t *testing.T) {
	var v ec2.InternetGateway
	resourceName := "aws_internet_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInternetGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInternetGatewayConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInternetGatewayExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsInternetGateway(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckInternetGatewayDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_internet_gateway" {
			continue
		}

		// Try to find the resource
		resp, err := conn.DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{
			InternetGatewayIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err == nil {
			if len(resp.InternetGateways) > 0 {
				return fmt.Errorf("still exists")
			}

			return nil
		}

		// Verify the error is what we want
		if !isAWSErr(err, "InvalidInternetGatewayID.NotFound", "") {
			return err
		}
	}

	return nil
}

func testAccCheckInternetGatewayExists(n string, ig *ec2.InternetGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		resp, err := conn.DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{
			InternetGatewayIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return err
		}
		if len(resp.InternetGateways) == 0 {
			return fmt.Errorf("InternetGateway not found")
		}

		*ig = *resp.InternetGateways[0]

		return nil
	}
}

const testAccNoInternetGatewayConfig = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-no-internet-gateway"
  }
}
`

const testAccInternetGatewayConfig = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-internet-gateway"
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}
`

const testAccInternetGatewayConfigChangeVPC = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-internet-gateway-change-vpc"
  }
}

resource "aws_vpc" "bar" {
  cidr_block = "10.2.0.0/16"

  tags = {
    Name = "terraform-testacc-internet-gateway-change-vpc-other"
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.bar.id

  tags = {
    Name = "terraform-testacc-internet-gateway-change-vpc-other"
  }
}
`

func testAccAWSInternetGatewayConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSInternetGatewayConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
