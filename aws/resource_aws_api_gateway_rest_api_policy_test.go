package aws

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSAPIGatewayRestApiPolicy_basic(t *testing.T) {
	var v apigateway.RestApi
	resourceName := "aws_api_gateway_rest_api_policy.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestApiPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestApiPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestApiPolicyExists(resourceName, &v),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`"Action":"execute-api:Invoke".+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGatewayRestApiPolicyConfigUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestApiPolicyExists(resourceName, &v),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`"aws:SourceIp":"123.123.123.123/32".+`))),
			},
		},
	})
}

func TestAccAWSAPIGatewayRestApiPolicy_disappears(t *testing.T) {
	var v apigateway.RestApi
	resourceName := "aws_api_gateway_rest_api_policy.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestApiPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestApiPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestApiPolicyExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsApiGatewayRestApiPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayRestApiPolicy_disappears_restApi(t *testing.T) {
	var v apigateway.RestApi
	resourceName := "aws_api_gateway_rest_api_policy.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayRestApiPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayRestApiPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayRestApiPolicyExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsApiGatewayRestApi(), "aws_api_gateway_rest_api.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSAPIGatewayRestApiPolicyExists(n string, res *apigateway.RestApi) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayconn

		req := &apigateway.GetRestApiInput{
			RestApiId: aws.String(rs.Primary.ID),
		}
		describe, err := conn.GetRestApi(req)
		if err != nil {
			return err
		}

		normalizedPolicy, err := structure.NormalizeJsonString(`"` + aws.StringValue(describe.Policy) + `"`)
		if err != nil {
			return fmt.Errorf("error normalizing API Gateway REST API policy JSON: %w", err)
		}
		policy, err := strconv.Unquote(normalizedPolicy)
		if err != nil {
			return fmt.Errorf("error unescaping API Gateway REST API policy: %w", err)
		}

		if aws.StringValue(describe.Id) != rs.Primary.ID &&
			policy != rs.Primary.Attributes["policy"] {
			return fmt.Errorf("API Gateway REST API Policy not found")
		}

		*res = *describe

		return nil
	}
}

func testAccCheckAWSAPIGatewayRestApiPolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigatewayconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_rest_api_policy" {
			continue
		}

		req := &apigateway.GetRestApisInput{}
		describe, err := conn.GetRestApis(req)

		if err == nil {
			if len(describe.Items) != 0 &&
				aws.StringValue(describe.Items[0].Id) == rs.Primary.ID &&
				aws.StringValue(describe.Items[0].Policy) == "" {
				return fmt.Errorf("API Gateway REST API Policy still exists")
			}
		}

		return err
	}

	return nil
}

func testAccAWSAPIGatewayRestApiPolicyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_rest_api_policy" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
   "Statement": [
       {
           "Effect": "Deny",
           "Principal": {
               "AWS": "*"
           },
           "Action": "execute-api:Invoke",
           "Resource": "${aws_api_gateway_rest_api.test.arn}"
       }
   ]
}
EOF
}
`, rName)
}

func testAccAWSAPIGatewayRestApiPolicyConfigUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_rest_api_policy" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "execute-api:Invoke",
      "Resource": "${aws_api_gateway_rest_api.test.arn}",
      "Condition": {
        "IpAddress": {
          "aws:SourceIp": "123.123.123.123/32"
        }
      }
    }
  ]
}
EOF
}
`, rName)
}
