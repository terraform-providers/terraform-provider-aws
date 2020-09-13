package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestDecodeApiGatewayBasePathMappingId(t *testing.T) {
	var testCases = []struct {
		Input      string
		DomainName string
		BasePath   string
		ErrCount   int
	}{
		{
			Input:    "no-slash",
			ErrCount: 1,
		},
		{
			Input:    "/missing-domain-name",
			ErrCount: 1,
		},
		{
			Input:      "domain-name/base-path",
			DomainName: "domain-name",
			BasePath:   "base-path",
			ErrCount:   0,
		},
		{
			Input:      "domain-name/base/path",
			DomainName: "domain-name",
			BasePath:   "base/path",
			ErrCount:   0,
		},
		{
			Input:      "domain-name/",
			DomainName: "domain-name",
			BasePath:   emptyBasePathMappingValue,
			ErrCount:   0,
		},
	}

	for _, tc := range testCases {
		domainName, basePath, err := decodeApiGatewayBasePathMappingId(tc.Input)
		if tc.ErrCount == 0 && err != nil {
			t.Fatalf("expected %q not to trigger an error, received: %s", tc.Input, err)
		}
		if tc.ErrCount > 0 && err == nil {
			t.Fatalf("expected %q to trigger an error", tc.Input)
		}
		if domainName != tc.DomainName {
			t.Fatalf("expected domain name %q to be %q", domainName, tc.DomainName)
		}
		if basePath != tc.BasePath {
			t.Fatalf("expected base path %q to be %q", basePath, tc.BasePath)
		}
	}
}

func TestAccAWSAPIGatewayBasePathMapping_basic(t *testing.T) {
	var conf apigateway.BasePathMapping

	name := fmt.Sprintf("tf-acc-%s.terraformtest.com", acctest.RandString(8))

	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayBasePathDestroy(name),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayBasePathConfigBasePath(name, key, certificate, "tf-acc-test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayBasePathExists("aws_api_gateway_base_path_mapping.test", &conf),
				),
			},
			{
				ResourceName:      "aws_api_gateway_base_path_mapping.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// https://github.com/hashicorp/terraform/issues/9212
func TestAccAWSAPIGatewayBasePathMapping_BasePath_Empty(t *testing.T) {
	var conf apigateway.BasePathMapping

	name := fmt.Sprintf("tf-acc-%s.terraformtest.com", acctest.RandString(8))

	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayBasePathDestroy(name),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayBasePathConfigBasePath(name, key, certificate, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayBasePathExists("aws_api_gateway_base_path_mapping.test", &conf),
				),
			},
			{
				ResourceName:      "aws_api_gateway_base_path_mapping.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayBasePathMapping_disappears(t *testing.T) {
	var conf apigateway.BasePathMapping

	name := fmt.Sprintf("tf-acc-%s.terraformtest.com", acctest.RandString(8))
	resourceName := "aws_api_gateway_base_path_mapping.test"

	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayBasePathDestroy(name),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayBasePathConfigBasePath(name, key, certificate, "tf-acc-test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayBasePathExists(resourceName, &conf),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsApiGatewayBasePathMapping(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSAPIGatewayBasePathExists(n string, res *apigateway.BasePathMapping) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayconn

		domainName, basePath, err := decodeApiGatewayBasePathMappingId(rs.Primary.ID)
		if err != nil {
			return err
		}

		req := &apigateway.GetBasePathMappingInput{
			DomainName: aws.String(domainName),
			BasePath:   aws.String(basePath),
		}
		describe, err := conn.GetBasePathMapping(req)
		if err != nil {
			return err
		}

		*res = *describe

		return nil
	}
}

func testAccCheckAWSAPIGatewayBasePathDestroy(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).apigatewayconn

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_api_gateway_base_path_mapping" {
				continue
			}

			domainName, basePath, err := decodeApiGatewayBasePathMappingId(rs.Primary.ID)
			if err != nil {
				return err
			}

			req := &apigateway.GetBasePathMappingInput{
				DomainName: aws.String(domainName),
				BasePath:   aws.String(basePath),
			}
			_, err = conn.GetBasePathMapping(req)

			if err != nil {
				if isAWSErr(err, apigateway.ErrCodeNotFoundException, "") {
					return nil
				}
				return err
			}

			return fmt.Errorf("expected error reading deleted base path, but got success")
		}

		return nil
	}
}

func testAccAWSAPIGatewayBasePathConfigBase(domainName, key, certificate string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_api_gateway_domain_name" "test" {
  domain_name              = %[1]q
  regional_certificate_arn = aws_acm_certificate.test.arn

  endpoint_configuration {
    types = ["REGIONAL"]
  }
}

resource "aws_api_gateway_rest_api" "test" {
  name        = "tf-acc-apigateway-base-path-mapping"
  description = "Terraform Acceptance Tests"

  endpoint_configuration {
    types = ["REGIONAL"]
  }
}

# API gateway won't let us deploy an empty API
resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "tf-acc"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method
  type        = "MOCK"
}

resource "aws_api_gateway_deployment" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = "test"
  depends_on  = [aws_api_gateway_integration.test]
}
`, domainName, tlsPemEscapeNewlines(certificate), tlsPemEscapeNewlines(key))
}

func testAccAWSAPIGatewayBasePathConfigBasePath(domainName, key, certificate, basePath string) string {
	return testAccAWSAPIGatewayBasePathConfigBase(domainName, key, certificate) + fmt.Sprintf(`
resource "aws_api_gateway_base_path_mapping" "test" {
  api_id      = aws_api_gateway_rest_api.test.id
  base_path   = %[1]q
  stage_name  = aws_api_gateway_deployment.test.stage_name
  domain_name = aws_api_gateway_domain_name.test.domain_name
}
`, basePath)
}
