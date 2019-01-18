package aws

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
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

	req := &ec2.DescribeInternetGatewaysInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag-value"),
				Values: []*string{
					aws.String("terraform-testacc-*"),
					aws.String("tf-acc-test-*"),
				},
			},
		},
	}
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

	for _, internetGateway := range resp.InternetGateways {
		for _, attachment := range internetGateway.Attachments {
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
				Pending: []string{"detaching"},
				Target:  []string{"detached"},
				Refresh: detachIGStateRefreshFunc(conn, aws.StringValue(internetGateway.InternetGatewayId), aws.StringValue(attachment.VpcId)),
				Timeout: 10 * time.Minute,
				Delay:   10 * time.Second,
			}

			log.Printf("[DEBUG] Waiting for Internet Gateway (%s) to detach from VPC (%s)", aws.StringValue(internetGateway.InternetGatewayId), aws.StringValue(attachment.VpcId))
			if _, err = stateConf.WaitForState(); err != nil {
				return fmt.Errorf("error waiting for VPN Gateway (%s) to detach from VPC (%s): %s", aws.StringValue(internetGateway.InternetGatewayId), aws.StringValue(attachment.VpcId), err)
			}
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

func TestAccAWSInternetGateway_importBasic(t *testing.T) {
	resourceName := "aws_internet_gateway.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInternetGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInternetGatewayConfig,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSInternetGateway_basic(t *testing.T) {
	var v, v2 ec2.InternetGateway

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
		IDRefreshName: "aws_internet_gateway.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInternetGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInternetGatewayConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInternetGatewayExists(
						"aws_internet_gateway.foo", &v),
					testAccCheckResourceAttrAccountID("aws_internet_gateway.foo", "owner_id"),
				),
			},

			{
				Config: testAccInternetGatewayConfigChangeVPC,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInternetGatewayExists(
						"aws_internet_gateway.foo", &v2),
					testNotEqual,
					testAccCheckResourceAttrAccountID("aws_internet_gateway.foo", "owner_id"),
				),
			},
		},
	})
}

func TestAccAWSInternetGateway_delete(t *testing.T) {
	var ig ec2.InternetGateway

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
		IDRefreshName: "aws_internet_gateway.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInternetGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInternetGatewayConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInternetGatewayExists("aws_internet_gateway.foo", &ig)),
			},
			{
				Config: testAccNoInternetGatewayConfig,
				Check:  resource.ComposeTestCheckFunc(testDeleted("aws_internet_gateway.foo")),
			},
		},
	})
}

func TestAccAWSInternetGateway_tags(t *testing.T) {
	var v ec2.InternetGateway

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_internet_gateway.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInternetGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckInternetGatewayConfigTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInternetGatewayExists("aws_internet_gateway.foo", &v),
					testAccCheckTags(&v.Tags, "Name", "terraform-testacc-internet-gateway-tags"),
					testAccCheckTags(&v.Tags, "foo", "bar"),
					testAccCheckResourceAttrAccountID("aws_internet_gateway.foo", "owner_id"),
				),
			},

			{
				Config: testAccCheckInternetGatewayConfigTagsUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInternetGatewayExists("aws_internet_gateway.foo", &v),
					testAccCheckTags(&v.Tags, "Name", "terraform-testacc-internet-gateway-tags"),
					testAccCheckTags(&v.Tags, "foo", ""),
					testAccCheckTags(&v.Tags, "bar", "baz"),
					testAccCheckResourceAttrAccountID("aws_internet_gateway.foo", "owner_id"),
				),
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
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if ec2err.Code() != "InvalidInternetGatewayID.NotFound" {
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
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-no-internet-gateway"
	}
}
`

const testAccInternetGatewayConfig = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-internet-gateway"
	}
}

resource "aws_internet_gateway" "foo" {
	vpc_id = "${aws_vpc.foo.id}"
	tags = {
		Name = "terraform-testacc-internet-gateway"
	}
}
`

const testAccInternetGatewayConfigChangeVPC = `
resource "aws_vpc" "foo" {
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

resource "aws_internet_gateway" "foo" {
	vpc_id = "${aws_vpc.bar.id}"
	tags = {
		Name = "terraform-testacc-internet-gateway-change-vpc-other"
	}
}
`

const testAccCheckInternetGatewayConfigTags = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-internet-gateway-tags"
	}
}

resource "aws_internet_gateway" "foo" {
	vpc_id = "${aws_vpc.foo.id}"
	tags = {
		Name = "terraform-testacc-internet-gateway-tags"
		foo = "bar"
	}
}
`

const testAccCheckInternetGatewayConfigTagsUpdate = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-internet-gateway-tags"
	}
}

resource "aws_internet_gateway" "foo" {
	vpc_id = "${aws_vpc.foo.id}"
	tags = {
		Name = "terraform-testacc-internet-gateway-tags"
		bar = "baz"
	}
}
`
