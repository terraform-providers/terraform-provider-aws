package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_api_gateway_vpc_link", &resource.Sweeper{
		Name: "aws_api_gateway_vpc_link",
		F:    testSweepAPIGatewayVpcLinks,
	})
}

func testSweepAPIGatewayVpcLinks(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).apigatewayconn

	err = conn.GetVpcLinksPages(&apigateway.GetVpcLinksInput{}, func(page *apigateway.GetVpcLinksOutput, lastPage bool) bool {
		for _, item := range page.Items {
			input := &apigateway.DeleteVpcLinkInput{
				VpcLinkId: item.Id,
			}
			id := aws.StringValue(item.Id)

			log.Printf("[INFO] Deleting API Gateway VPC Link: %s", id)
			_, err := conn.DeleteVpcLink(input)

			if err != nil {
				log.Printf("[ERROR] Failed to delete API Gateway VPC Link %s: %s", id, err)
				continue
			}

			if err := waitForApiGatewayVpcLinkDeletion(conn, id); err != nil {
				log.Printf("[ERROR] Error waiting for API Gateway VPC Link (%s) deletion: %s", id, err)
			}
		}
		return !lastPage
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping API Gateway VPC Link sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving API Gateway VPC Links: %s", err)
	}

	return nil
}

func TestAccAWSAPIGatewayVpcLink_basic(t *testing.T) {
	rName := acctest.RandString(5)
	resourceName := "aws_api_gateway_vpc_link.test"
	vpcLinkName := fmt.Sprintf("tf-apigateway-%s", rName)
	vpcLinkNameUpdated := fmt.Sprintf("tf-apigateway-update-%s", rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAPIGatewayVpcLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIGatewayVpcLinkConfig(rName, "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAPIGatewayVpcLinkExists(resourceName),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/vpclinks/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", vpcLinkName),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "target_arns.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAPIGatewayVpcLinkConfig_Update(rName, "test update"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAPIGatewayVpcLinkExists(resourceName),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/vpclinks/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", vpcLinkNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "description", "test update"),
					resource.TestCheckResourceAttr(resourceName, "target_arns.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayVpcLink_tags(t *testing.T) {
	rName := acctest.RandString(5)
	resourceName := "aws_api_gateway_vpc_link.test"
	vpcLinkName := fmt.Sprintf("tf-apigateway-%s", rName)
	description := "test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAPIGatewayVpcLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIGatewayVpcLinkConfigTags1(rName, description, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAPIGatewayVpcLinkExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", vpcLinkName),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "target_arns.#", "1"),
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
				Config: testAccAPIGatewayVpcLinkConfigTags2(rName, description, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAPIGatewayVpcLinkExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", vpcLinkName),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "target_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAPIGatewayVpcLinkConfigTags1(rName, description, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAPIGatewayVpcLinkExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", vpcLinkName),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "target_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayVpcLink_disappears(t *testing.T) {
	rName := acctest.RandString(5)
	resourceName := "aws_api_gateway_vpc_link.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAPIGatewayVpcLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIGatewayVpcLinkConfig(rName, "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAPIGatewayVpcLinkExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsApiGatewayVpcLink(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsAPIGatewayVpcLinkDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigatewayconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_vpc_link" {
			continue
		}

		input := &apigateway.GetVpcLinkInput{
			VpcLinkId: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetVpcLink(input)
		if err != nil {
			if isAWSErr(err, apigateway.ErrCodeNotFoundException, "") {
				return nil
			}
			return err
		}

		return fmt.Errorf("Expected VPC Link to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAwsAPIGatewayVpcLinkExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayconn

		input := &apigateway.GetVpcLinkInput{
			VpcLinkId: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetVpcLink(input)
		return err
	}
}

func testAccAPIGatewayVpcLinkConfig_basis(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb" "test_a" {
  name               = "tf-lb-%s"
  internal           = true
  load_balancer_type = "network"
  subnets            = [aws_subnet.test.id]
}

resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"
  tags = {
    Name = "tf-acc-api-gateway-vpc-link-%s"
  }
}

data "aws_availability_zones" "test" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.10.0.0/21"
  availability_zone = data.aws_availability_zones.test.names[0]

  tags = {
    Name = "tf-acc-api-gateway-vpc-link"
  }
}
`, rName, rName)
}

func testAccAPIGatewayVpcLinkConfig(rName, description string) string {
	return testAccAPIGatewayVpcLinkConfig_basis(rName) + fmt.Sprintf(`
resource "aws_api_gateway_vpc_link" "test" {
  name        = "tf-apigateway-%s"
  description = %q
  target_arns = [aws_lb.test_a.arn]
}
`, rName, description)
}

func testAccAPIGatewayVpcLinkConfigTags1(rName, description, tagKey1, tagValue1 string) string {
	return testAccAPIGatewayVpcLinkConfig_basis(rName) + fmt.Sprintf(`
resource "aws_api_gateway_vpc_link" "test" {
  name        = "tf-apigateway-%s"
  description = %q
  target_arns = [aws_lb.test_a.arn]

  tags = {
    %q = %q
  }
}
`, rName, description, tagKey1, tagValue1)
}

func testAccAPIGatewayVpcLinkConfigTags2(rName, description, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAPIGatewayVpcLinkConfig_basis(rName) + fmt.Sprintf(`
resource "aws_api_gateway_vpc_link" "test" {
  name        = "tf-apigateway-%s"
  description = %q
  target_arns = [aws_lb.test_a.arn]

  tags = {
    %q = %q
    %q = %q
  }
}
`, rName, description, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAPIGatewayVpcLinkConfig_Update(rName, description string) string {
	return testAccAPIGatewayVpcLinkConfig_basis(rName) + fmt.Sprintf(`
resource "aws_api_gateway_vpc_link" "test" {
  name        = "tf-apigateway-update-%s"
  description = %q
  target_arns = [aws_lb.test_a.arn]
}
`, rName, description)
}
