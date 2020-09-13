package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSAPIGatewayClientCertificate_basic(t *testing.T) {
	var conf apigateway.ClientCertificate
	resourceName := "aws_api_gateway_client_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayClientCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayClientCertificateConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayClientCertificateExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/clientcertificates/+.`)),
					resource.TestCheckResourceAttr(resourceName, "description", "Hello from TF acceptance test"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGatewayClientCertificateConfig_basic_updated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayClientCertificateExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/clientcertificates/+.`)),
					resource.TestCheckResourceAttr(resourceName, "description", "Hello from TF acceptance test - updated"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayClientCertificate_tags(t *testing.T) {
	var conf apigateway.ClientCertificate
	resourceName := "aws_api_gateway_client_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayClientCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayClientCertificateConfigTags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayClientCertificateExists(resourceName, &conf),
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
				Config: testAccAWSAPIGatewayClientCertificateConfigTags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayClientCertificateExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSAPIGatewayClientCertificateConfigTags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayClientCertificateExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayClientCertificate_disappears(t *testing.T) {
	var conf apigateway.ClientCertificate
	resourceName := "aws_api_gateway_client_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayClientCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayClientCertificateConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayClientCertificateExists(resourceName, &conf),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsApiGatewayClientCertificate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSAPIGatewayClientCertificateExists(n string, res *apigateway.ClientCertificate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway Client Certificate ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayconn

		req := &apigateway.GetClientCertificateInput{
			ClientCertificateId: aws.String(rs.Primary.ID),
		}
		out, err := conn.GetClientCertificate(req)
		if err != nil {
			return err
		}

		*res = *out

		return nil
	}
}

func testAccCheckAWSAPIGatewayClientCertificateDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigatewayconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_client_certificate" {
			continue
		}

		req := &apigateway.GetClientCertificateInput{
			ClientCertificateId: aws.String(rs.Primary.ID),
		}
		out, err := conn.GetClientCertificate(req)
		if err == nil {
			return fmt.Errorf("API Gateway Client Certificate still exists: %s", out)
		}

		awsErr, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if awsErr.Code() != apigateway.ErrCodeNotFoundException {
			return err
		}

		return nil
	}

	return nil
}

const testAccAWSAPIGatewayClientCertificateConfig_basic = `
resource "aws_api_gateway_client_certificate" "test" {
  description = "Hello from TF acceptance test"
}
`

const testAccAWSAPIGatewayClientCertificateConfig_basic_updated = `
resource "aws_api_gateway_client_certificate" "test" {
  description = "Hello from TF acceptance test - updated"
}
`

func testAccAWSAPIGatewayClientCertificateConfigTags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_client_certificate" "test" {
  description = "Hello from TF acceptance test"

  tags = {
    %q = %q
  }
}
`, tagKey1, tagValue1)
}

func testAccAWSAPIGatewayClientCertificateConfigTags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_client_certificate" "test" {
  description = "Hello from TF acceptance test"

  tags = {
    %q = %q
    %q = %q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}
