package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

// These tests need to be serialized, else resources get orphaned after "TooManyRequests" errors.
func TestAccAWSAPIGatewayV2ApiMapping_basic(t *testing.T) {
	var certificateArn string
	rName := acctest.RandomWithPrefix("tf-acc-test")

	// Create an ACM certificate to be used by all the tests.
	// It is created outside the Terraform configurations because deletion
	// of CloudFront distribution backing the API Gateway domain name is asynchronous
	// and can take up to 60 minutes and the distribution keeps the certificate alive.
	t.Run("createCertificate", func(t *testing.T) {
		testAccAWSAPIGatewayV2ApiMapping_createCertificate(t, rName, &certificateArn)
	})

	testCases := map[string]func(t *testing.T, rName string, certificateArn *string){
		"basic":         testAccAWSAPIGatewayV2ApiMapping_basic,
		"disappears":    testAccAWSAPIGatewayV2ApiMapping_disappears,
		"ApiMappingKey": testAccAWSAPIGatewayV2ApiMapping_ApiMappingKey,
	}
	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t, rName, &certificateArn)
		})
	}
}

func testAccAWSAPIGatewayV2ApiMapping_createCertificate(t *testing.T, rName string, certificateArn *string) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: "# Dummy config.",
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2ApiMappingCreateCertificate(rName, certificateArn),
				),
			},
		},
	})

	log.Printf("[INFO] Created ACM certificate %s", *certificateArn)
}

func testAccAWSAPIGatewayV2ApiMapping_basic(t *testing.T, rName string, certificateArn *string) {
	var domainName string
	var v apigatewayv2.GetApiMappingOutput
	resourceName := "aws_apigatewayv2_api_mapping.test"
	domainNameResourceName := "aws_apigatewayv2_domain_name.test"
	stageResourceName := "aws_apigatewayv2_stage.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2ApiMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2ApiMappingConfig_basic(rName, *certificateArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2ApiMappingExists(resourceName, &domainName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name", domainNameResourceName, "domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "stage", stageResourceName, "name")),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGatewayV2ApiMappingImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAWSAPIGatewayV2ApiMapping_disappears(t *testing.T, rName string, certificateArn *string) {
	var domainName string
	var v apigatewayv2.GetApiMappingOutput
	resourceName := "aws_apigatewayv2_api_mapping.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2ApiMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2ApiMappingConfig_basic(rName, *certificateArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2ApiMappingExists(resourceName, &domainName, &v),
					testAccCheckAWSAPIGatewayV2ApiMappingDisappears(&domainName, &v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAWSAPIGatewayV2ApiMapping_ApiMappingKey(t *testing.T, rName string, certificateArn *string) {
	var domainName string
	var v apigatewayv2.GetApiMappingOutput
	resourceName := "aws_apigatewayv2_api_mapping.test"
	domainNameResourceName := "aws_apigatewayv2_domain_name.test"
	stageResourceName := "aws_apigatewayv2_stage.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2ApiMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2ApiMappingConfig_apiMappingKey(rName, *certificateArn, "$context.domainName"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2ApiMappingExists(resourceName, &domainName, &v),
					resource.TestCheckResourceAttr(resourceName, "api_mapping_key", "$context.domainName"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name", domainNameResourceName, "domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "stage", stageResourceName, "name")),
			},
			{
				Config: testAccAWSAPIGatewayV2ApiMappingConfig_apiMappingKey(rName, *certificateArn, "$context.apiId"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2ApiMappingExists(resourceName, &domainName, &v),
					resource.TestCheckResourceAttr(resourceName, "api_mapping_key", "$context.apiId"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name", domainNameResourceName, "domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "stage", stageResourceName, "name")),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGatewayV2ApiMappingImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSAPIGatewayV2ApiMappingCreateCertificate(rName string, certificateArn *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		privateKey := tlsRsaPrivateKeyPem(2048)
		certificate := tlsRsaX509SelfSignedCertificatePem(privateKey, fmt.Sprintf("%s.example.com", rName))

		conn := testAccProvider.Meta().(*AWSClient).acmconn

		output, err := conn.ImportCertificate(&acm.ImportCertificateInput{
			Certificate: []byte(certificate),
			PrivateKey:  []byte(privateKey),
			Tags: keyvaluetags.New(map[string]interface{}{
				"Name": rName,
			}).IgnoreAws().AcmTags(),
		})
		if err != nil {
			return err
		}

		*certificateArn = *output.CertificateArn

		return nil
	}
}

func testAccCheckAWSAPIGatewayV2ApiMappingDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_apigatewayv2_api_mapping" {
			continue
		}

		_, err := conn.GetApiMapping(&apigatewayv2.GetApiMappingInput{
			ApiMappingId: aws.String(rs.Primary.ID),
			DomainName:   aws.String(rs.Primary.Attributes["domain_name"]),
		})
		if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("API Gateway v2 API mapping %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSAPIGatewayV2ApiMappingDisappears(domainName *string, v *apigatewayv2.GetApiMappingOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

		_, err := conn.DeleteApiMapping(&apigatewayv2.DeleteApiMappingInput{
			ApiMappingId: v.ApiMappingId,
			DomainName:   domainName,
		})

		return err
	}
}

func testAccCheckAWSAPIGatewayV2ApiMappingExists(n string, vDomainName *string, v *apigatewayv2.GetApiMappingOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 API mapping ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

		domainName := aws.String(rs.Primary.Attributes["domain_name"])
		resp, err := conn.GetApiMapping(&apigatewayv2.GetApiMappingInput{
			ApiMappingId: aws.String(rs.Primary.ID),
			DomainName:   domainName,
		})
		if err != nil {
			return err
		}

		*vDomainName = *domainName
		*v = *resp

		return nil
	}
}

func testAccAWSAPIGatewayV2ApiMappingImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.ID, rs.Primary.Attributes["domain_name"]), nil
	}
}

func testAccAWSAPIGatewayV2ApiMappingConfig_base(rName, certificateArn string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_domain_name" "test" {
  domain_name = "%[1]s.example.com"

  domain_name_configuration {
    certificate_arn = %[2]q
    endpoint_type   = "REGIONAL"
    security_policy = "TLS_1_2"
  }
}
`, rName, certificateArn)
}

func testAccAWSAPIGatewayV2ApiMappingConfig_basic(rName, certificateArn string) string {
	return testAccAWSAPIGatewayV2ApiMappingConfig_base(rName, certificateArn) + testAccAWSAPIGatewayV2StageConfig_basicWebSocket(rName) + `
resource "aws_apigatewayv2_api_mapping" "test" {
  api_id      = aws_apigatewayv2_api.test.id
  domain_name = aws_apigatewayv2_domain_name.test.id
  stage       = aws_apigatewayv2_stage.test.id
}
`
}

func testAccAWSAPIGatewayV2ApiMappingConfig_apiMappingKey(rName, certificateArn, apiMappingKey string) string {
	return testAccAWSAPIGatewayV2ApiMappingConfig_base(rName, certificateArn) + testAccAWSAPIGatewayV2StageConfig_basicWebSocket(rName) + fmt.Sprintf(`
resource "aws_apigatewayv2_api_mapping" "test" {
  api_id      = aws_apigatewayv2_api.test.id
  domain_name = aws_apigatewayv2_domain_name.test.id
  stage       = aws_apigatewayv2_stage.test.id

  api_mapping_key = %[1]q
}
`, apiMappingKey)
}
