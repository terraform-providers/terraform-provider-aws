package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/apigatewayv2/waiter"
)

func init() {
	resource.AddTestSweepers("aws_apigatewayv2_vpc_link", &resource.Sweeper{
		Name: "aws_apigatewayv2_vpc_link",
		F:    testSweepAPIGatewayV2VpcLinks,
	})
}

func testSweepAPIGatewayV2VpcLinks(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).apigatewayv2conn
	input := &apigatewayv2.GetVpcLinksInput{}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.GetVpcLinks(input)
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping API Gateway v2 VPC Link sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving API Gateway v2 VPC Links: %s", err)
		}

		for _, link := range output.Items {
			log.Printf("[INFO] Deleting API Gateway v2 VPC Link: %s", aws.StringValue(link.VpcLinkId))
			_, err := conn.DeleteVpcLink(&apigatewayv2.DeleteVpcLinkInput{
				VpcLinkId: link.VpcLinkId,
			})
			if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
				continue
			}
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting API Gateway v2 VPC Link (%s): %w", aws.StringValue(link.VpcLinkId), err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			_, err = waiter.VpcLinkDeleted(conn, aws.StringValue(link.VpcLinkId))
			if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
				continue
			}
			if err != nil {
				sweeperErr := fmt.Errorf("error waiting for API Gateway v2 VPC Link (%s) deletion: %w", aws.StringValue(link.VpcLinkId), err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}
		input.NextToken = output.NextToken
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSAPIGatewayV2VpcLink_basic(t *testing.T) {
	var v apigatewayv2.GetVpcLinkOutput
	resourceName := "aws_apigatewayv2_vpc_link.test"
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2VpcLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2VpcLinkConfig_basic(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2VpcLinkExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/vpclinks/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccAWSAPIGatewayV2VpcLinkConfig_basic(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2VpcLinkExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/vpclinks/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
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

func TestAccAWSAPIGatewayV2VpcLink_disappears(t *testing.T) {
	var v apigatewayv2.GetVpcLinkOutput
	resourceName := "aws_apigatewayv2_vpc_link.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2VpcLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2VpcLinkConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2VpcLinkExists(resourceName, &v),
					testAccCheckAWSAPIGatewayV2VpcLinkDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayV2VpcLink_Tags(t *testing.T) {
	var v apigatewayv2.GetVpcLinkOutput
	resourceName := "aws_apigatewayv2_vpc_link.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2VpcLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2VpcLinkConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2VpcLinkExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/vpclinks/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGatewayV2VpcLinkConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2VpcLinkExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccCheckAWSAPIGatewayV2VpcLinkDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_apigatewayv2_vpc_link" {
			continue
		}

		_, err := conn.GetVpcLink(&apigatewayv2.GetVpcLinkInput{
			VpcLinkId: aws.String(rs.Primary.ID),
		})
		if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("API Gateway v2 VPC Link %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSAPIGatewayV2VpcLinkDisappears(v *apigatewayv2.GetVpcLinkOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

		if _, err := conn.DeleteVpcLink(&apigatewayv2.DeleteVpcLinkInput{
			VpcLinkId: v.VpcLinkId,
		}); err != nil {
			return err
		}

		_, err := waiter.VpcLinkDeleted(conn, aws.StringValue(v.VpcLinkId))
		if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
			return nil
		}
		if err != nil {
			return fmt.Errorf("error waiting for API Gateway v2 VPC Link (%s) deletion: %s", aws.StringValue(v.VpcLinkId), err)
		}

		return nil
	}
}

func testAccCheckAWSAPIGatewayV2VpcLinkExists(n string, v *apigatewayv2.GetVpcLinkOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 VPC Link ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

		resp, err := conn.GetVpcLink(&apigatewayv2.GetVpcLinkInput{
			VpcLinkId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccAWSAPIGatewayV2VpcLinkConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_subnet" "test" {
  count = 2

  vpc_id            = aws_vpc.test.id
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 2, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSAPIGatewayV2VpcLinkConfig_basic(rName string) string {
	return testAccAWSAPIGatewayV2VpcLinkConfig_base(rName) + fmt.Sprintf(`
resource "aws_apigatewayv2_vpc_link" "test" {
  name               = %[1]q
  security_group_ids = [aws_security_group.test.id]
  subnet_ids         = aws_subnet.test.*.id
}
`, rName)
}

func testAccAWSAPIGatewayV2VpcLinkConfig_tags(rName string) string {
	return testAccAWSAPIGatewayV2VpcLinkConfig_base(rName) + fmt.Sprintf(`
resource "aws_apigatewayv2_vpc_link" "test" {
  name               = %[1]q
  security_group_ids = [aws_security_group.test.id]
  subnet_ids         = aws_subnet.test.*.id

  tags = {
    Key1 = "Value1"
    Key2 = "Value2"
  }
}
`, rName)
}
