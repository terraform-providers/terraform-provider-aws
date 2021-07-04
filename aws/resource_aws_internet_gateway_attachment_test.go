package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
)

func init() {
	resource.AddTestSweepers("aws_internet_gateway_attachment", &resource.Sweeper{
		Name: "aws_internet_gateway_attachment",
		F:    testSweepInternetGatewayAttachments,
	})
}

func testSweepInternetGatewayAttachments(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ec2conn
	var sweeperErrs *multierror.Error

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
		for _, attachment := range internetGateway.Attachments {
			if aws.StringValue(attachment.VpcId) == defaultVPCID {
				break
			}

			igwID := aws.StringValue(internetGateway.InternetGatewayId)
			vpcID := aws.StringValue(attachment.VpcId)

			log.Printf("[DEBUG] Detaching Internet Gateway %s from VPC %s", igwID, vpcID)
			r := resourceAwsInternetGatewayAttachment()
			d := r.Data(nil)
			d.SetId(fmt.Sprintf("%s:%s", vpcID, igwID))
			err := r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}
	}

	return nil
}

func TestAccAWSInternetGatewayAttachment_basic(t *testing.T) {
	var v ec2.InternetGatewayAttachment
	resourceName := "aws_internet_gateway_attachment.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheck(t, ec2.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInternetGatewayAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSInternetGatewayAttachmentBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInternetGatewayAttachmentExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", "aws_vpc.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "internet_gateway_id", "aws_internet_gateway.test", "id"),
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

func TestAccAWSInternetGatewayAttachment_disappears(t *testing.T) {
	var v ec2.InternetGatewayAttachment
	resourceName := "aws_internet_gateway_attachment.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInternetGatewayAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSInternetGatewayAttachmentBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInternetGatewayAttachmentExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsInternetGatewayAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckInternetGatewayAttachmentDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_internet_gateway_attachment" {
			continue
		}

		vpcID, igwID, err := tfec2.InternetGatewayAttachmentParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = finder.InternetGatewayAttachmentByID(conn, igwID, vpcID)
		if !tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidInternetGatewayIDNotFound) {
			return err
		}
	}

	return nil
}

func testAccCheckInternetGatewayAttachmentExists(n string, ig *ec2.InternetGatewayAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		vpcID, igwID, err := tfec2.InternetGatewayAttachmentParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		resp, err := finder.InternetGatewayAttachmentByID(conn, igwID, vpcID)
		if err != nil {
			return err
		}

		*ig = *resp

		return nil
	}
}

func testAccAWSInternetGatewayAttachmentBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  tags = {
    Name = %[1]q
  }

  lifecycle {
    ignore_changes = ["vpc_id"]
  }
}

resource "aws_internet_gateway_attachment" "test" {
  vpc_id              = aws_vpc.test.id
  internet_gateway_id = aws_internet_gateway.test.id
}
`, rName)
}
