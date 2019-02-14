package aws

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSAPIGatewayDocumentationPart_basic(t *testing.T) {
	var conf apigateway.DocumentationPart

	rString := acctest.RandString(8)
	apiName := fmt.Sprintf("tf_acc_api_doc_part_basic_%s", rString)
	properties := `{"description":"Terraform Acceptance Test"}`
	uProperties := `{"description":"Terraform Acceptance Test Updated"}`

	resourceName := "aws_api_gateway_documentation_part.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayDocumentationPartDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayDocumentationPartConfig(apiName, strconv.Quote(properties)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDocumentationPartExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "location.0.type", "API"),
					resource.TestCheckResourceAttr(resourceName, "properties", properties),
					resource.TestCheckResourceAttrSet(resourceName, "rest_api_id"),
				),
			},
			{
				Config: testAccAWSAPIGatewayDocumentationPartConfig(apiName, strconv.Quote(uProperties)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDocumentationPartExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "location.0.type", "API"),
					resource.TestCheckResourceAttr(resourceName, "properties", uProperties),
					resource.TestCheckResourceAttrSet(resourceName, "rest_api_id"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayDocumentationPart_method(t *testing.T) {
	var conf apigateway.DocumentationPart

	rString := acctest.RandString(8)
	apiName := fmt.Sprintf("tf_acc_api_doc_part_method_%s", rString)
	properties := `{"description":"Terraform Acceptance Test"}`
	uProperties := `{"description":"Terraform Acceptance Test Updated"}`

	resourceName := "aws_api_gateway_documentation_part.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayDocumentationPartDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayDocumentationPartMethodConfig(apiName, strconv.Quote(properties)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDocumentationPartExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "location.0.type", "METHOD"),
					resource.TestCheckResourceAttr(resourceName, "location.0.method", "GET"),
					resource.TestCheckResourceAttr(resourceName, "location.0.path", "/terraform-acc-test"),
					resource.TestCheckResourceAttr(resourceName, "properties", properties),
					resource.TestCheckResourceAttrSet(resourceName, "rest_api_id"),
				),
			},
			{
				Config: testAccAWSAPIGatewayDocumentationPartMethodConfig(apiName, strconv.Quote(uProperties)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDocumentationPartExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "location.0.type", "METHOD"),
					resource.TestCheckResourceAttr(resourceName, "location.0.method", "GET"),
					resource.TestCheckResourceAttr(resourceName, "location.0.path", "/terraform-acc-test"),
					resource.TestCheckResourceAttr(resourceName, "properties", uProperties),
					resource.TestCheckResourceAttrSet(resourceName, "rest_api_id"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayDocumentationPart_responseHeader(t *testing.T) {
	var conf apigateway.DocumentationPart

	rString := acctest.RandString(8)
	apiName := fmt.Sprintf("tf_acc_api_doc_part_resp_header_%s", rString)
	properties := `{"description":"Terraform Acceptance Test"}`
	uProperties := `{"description":"Terraform Acceptance Test Updated"}`

	resourceName := "aws_api_gateway_documentation_part.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayDocumentationPartDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayDocumentationPartResponseHeaderConfig(apiName, strconv.Quote(properties)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDocumentationPartExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "location.0.type", "RESPONSE_HEADER"),
					resource.TestCheckResourceAttr(resourceName, "location.0.method", "GET"),
					resource.TestCheckResourceAttr(resourceName, "location.0.name", "tfacc"),
					resource.TestCheckResourceAttr(resourceName, "location.0.path", "/terraform-acc-test"),
					resource.TestCheckResourceAttr(resourceName, "location.0.status_code", "200"),
					resource.TestCheckResourceAttr(resourceName, "properties", properties),
					resource.TestCheckResourceAttrSet(resourceName, "rest_api_id"),
				),
			},
			{
				Config: testAccAWSAPIGatewayDocumentationPartResponseHeaderConfig(apiName, strconv.Quote(uProperties)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDocumentationPartExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "location.0.type", "RESPONSE_HEADER"),
					resource.TestCheckResourceAttr(resourceName, "location.0.method", "GET"),
					resource.TestCheckResourceAttr(resourceName, "location.0.name", "tfacc"),
					resource.TestCheckResourceAttr(resourceName, "location.0.path", "/terraform-acc-test"),
					resource.TestCheckResourceAttr(resourceName, "location.0.status_code", "200"),
					resource.TestCheckResourceAttr(resourceName, "properties", uProperties),
					resource.TestCheckResourceAttrSet(resourceName, "rest_api_id"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayDocumentationPart_importBasic(t *testing.T) {
	var conf apigateway.DocumentationPart

	rString := acctest.RandString(8)
	apiName := fmt.Sprintf("tf_acc_api_doc_part_import_%s", rString)
	properties := `{"description":"Terraform Acceptance Test"}`

	resourceName := "aws_api_gateway_documentation_part.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayDocumentationPartDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayDocumentationPartConfig(apiName, strconv.Quote(properties)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDocumentationPartExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "location.0.type", "API"),
					resource.TestCheckResourceAttr(resourceName, "properties", properties),
					resource.TestCheckResourceAttrSet(resourceName, "rest_api_id"),
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

func testAccCheckAWSAPIGatewayDocumentationPartExists(n string, res *apigateway.DocumentationPart) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway Documentation Part ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigateway

		apiId, id, err := decodeApiGatewayDocumentationPartId(rs.Primary.ID)
		if err != nil {
			return err
		}

		req := &apigateway.GetDocumentationPartInput{
			DocumentationPartId: aws.String(id),
			RestApiId:           aws.String(apiId),
		}
		docPart, err := conn.GetDocumentationPart(req)
		if err != nil {
			return err
		}

		*res = *docPart

		return nil
	}
}

func testAccCheckAWSAPIGatewayDocumentationPartDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigateway

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_documentation_part" {
			continue
		}

		apiId, id, err := decodeApiGatewayDocumentationPartId(rs.Primary.ID)
		if err != nil {
			return err
		}

		req := &apigateway.GetDocumentationPartInput{
			DocumentationPartId: aws.String(id),
			RestApiId:           aws.String(apiId),
		}
		_, err = conn.GetDocumentationPart(req)
		if err != nil {
			if isAWSErr(err, apigateway.ErrCodeNotFoundException, "") {
				return nil
			}
			return err
		}

		return fmt.Errorf("API Gateway Documentation Part %q still exists.", rs.Primary.ID)
	}
	return nil
}

func testAccAWSAPIGatewayDocumentationPartConfig(apiName, properties string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_documentation_part" "test" {
  location {
    type = "API"
  }
  properties = %v
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
}

resource "aws_api_gateway_rest_api" "test" {
  name = "%s"
}
`, properties, apiName)
}

func testAccAWSAPIGatewayDocumentationPartMethodConfig(apiName, properties string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_documentation_part" "test" {
  location {
    type = "METHOD"
    method = "GET"
    path = "/terraform-acc-test"
  }
  properties = %v
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
}

resource "aws_api_gateway_rest_api" "test" {
  name = "%s"
}
`, properties, apiName)
}

func testAccAWSAPIGatewayDocumentationPartResponseHeaderConfig(apiName, properties string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_documentation_part" "test" {
  location {
    type = "RESPONSE_HEADER"
    method = "GET"
    name = "tfacc"
    path = "/terraform-acc-test"
    status_code = "200"
  }
  properties = %v
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
}

resource "aws_api_gateway_rest_api" "test" {
  name = "%s"
}
`, properties, apiName)
}
