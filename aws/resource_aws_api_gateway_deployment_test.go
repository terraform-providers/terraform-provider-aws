package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSAPIGatewayDeployment_basic(t *testing.T) {
	var deployment apigateway.Deployment
	resourceName := "aws_api_gateway_deployment.test"
	restApiResourceName := "aws_api_gateway_rest_api.test"
	rName := acctest.RandomWithPrefix("tf-acc-test-deployment")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayDeploymentConfigStageName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDeploymentExists(resourceName, &deployment),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccMatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(fmt.Sprintf(".+/%s", rName))),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/%s", testAccGetRegion(), rName))),
					resource.TestCheckResourceAttrPair(resourceName, "rest_api_id", restApiResourceName, "id"),
					resource.TestCheckNoResourceAttr(resourceName, "stage_description"),
					resource.TestCheckResourceAttr(resourceName, "stage_name", rName),
					resource.TestCheckNoResourceAttr(resourceName, "variables.%"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayDeployment_disappears_RestApi(t *testing.T) {
	var deployment apigateway.Deployment
	var restApi apigateway.RestApi
	resourceName := "aws_api_gateway_deployment.test"
	restApiResourceName := "aws_api_gateway_rest_api.test"
	rName := acctest.RandomWithPrefix("tf-acc-test-deployment")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayDeploymentConfigStageName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDeploymentExists(resourceName, &deployment),
					testAccCheckAWSAPIGatewayRestAPIExists(restApiResourceName, &restApi),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsApiGatewayRestApi(), restApiResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayDeployment_Triggers(t *testing.T) {
	var deployment1, deployment2, deployment3, deployment4 apigateway.Deployment
	var stage apigateway.Stage
	resourceName := "aws_api_gateway_deployment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayDeploymentConfigTriggers("description1", "https://example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDeploymentExists(resourceName, &deployment1),
					testAccCheckAWSAPIGatewayDeploymentStageExists(resourceName, &stage),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
					resource.TestCheckResourceAttr(resourceName, "stage_description", "description1"),
				),
				// Due to how the Terraform state is handled for resources during creation,
				// any SHA1 of whole resources will change after first apply, then stabilize.
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccAWSAPIGatewayDeploymentConfigTriggers("description1", "https://example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDeploymentExists(resourceName, &deployment2),
					testAccCheckAWSAPIGatewayDeploymentRecreated(&deployment1, &deployment2),
					testAccCheckAWSAPIGatewayDeploymentStageExists(resourceName, &stage),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
					resource.TestCheckResourceAttr(resourceName, "stage_description", "description1"),
				),
			},
			{
				Config: testAccAWSAPIGatewayDeploymentConfigTriggers("description1", "https://example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDeploymentExists(resourceName, &deployment3),
					testAccCheckAWSAPIGatewayDeploymentNotRecreated(&deployment2, &deployment3),
					testAccCheckAWSAPIGatewayDeploymentStageExists(resourceName, &stage),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
					resource.TestCheckResourceAttr(resourceName, "stage_description", "description1"),
				),
			},
			{
				Config: testAccAWSAPIGatewayDeploymentConfigTriggers("description2", "https://example.org"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDeploymentExists(resourceName, &deployment4),
					testAccCheckAWSAPIGatewayDeploymentRecreated(&deployment3, &deployment4),
					testAccCheckAWSAPIGatewayDeploymentStageExists(resourceName, &stage),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
					resource.TestCheckResourceAttr(resourceName, "stage_description", "description2"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayDeployment_Description(t *testing.T) {
	var deployment apigateway.Deployment
	resourceName := "aws_api_gateway_deployment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayDeploymentConfigDescription("description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDeploymentExists(resourceName, &deployment),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				Config: testAccAWSAPIGatewayDeploymentConfigDescription("description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDeploymentExists(resourceName, &deployment),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayDeployment_StageDescription(t *testing.T) {
	var deployment apigateway.Deployment
	var stage apigateway.Stage
	resourceName := "aws_api_gateway_deployment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayDeploymentConfigStageDescription("description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDeploymentExists(resourceName, &deployment),
					testAccCheckAWSAPIGatewayDeploymentStageExists(resourceName, &stage),
					resource.TestCheckResourceAttr(resourceName, "stage_description", "description1"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayDeployment_StageName(t *testing.T) {
	var deployment apigateway.Deployment
	var stage apigateway.Stage
	resourceName := "aws_api_gateway_deployment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayDeploymentConfigStageName("test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDeploymentExists(resourceName, &deployment),
					testAccCheckAWSAPIGatewayDeploymentStageExists(resourceName, &stage),
					resource.TestCheckResourceAttr(resourceName, "stage_name", "test"),
				),
			},
			{
				Config: testAccAWSAPIGatewayDeploymentConfigRequired(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceName, "stage_name"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayDeployment_StageName_EmptyString(t *testing.T) {
	var deployment apigateway.Deployment
	resourceName := "aws_api_gateway_deployment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayDeploymentConfigStageName(""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDeploymentExists(resourceName, &deployment),
					resource.TestCheckResourceAttr(resourceName, "stage_name", ""),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayDeployment_Variables(t *testing.T) {
	var deployment apigateway.Deployment
	resourceName := "aws_api_gateway_deployment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayDeploymentConfigVariables("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDeploymentExists(resourceName, &deployment),
					resource.TestCheckResourceAttr(resourceName, "variables.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "variables.key1", "value1"),
				),
			},
		},
	})
}

func testAccCheckAWSAPIGatewayDeploymentExists(n string, res *apigateway.Deployment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway Deployment ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayconn

		req := &apigateway.GetDeploymentInput{
			DeploymentId: aws.String(rs.Primary.ID),
			RestApiId:    aws.String(s.RootModule().Resources["aws_api_gateway_rest_api.test"].Primary.ID),
		}
		describe, err := conn.GetDeployment(req)
		if err != nil {
			return err
		}

		if *describe.Id != rs.Primary.ID {
			return fmt.Errorf("APIGateway Deployment not found")
		}

		*res = *describe

		return nil
	}
}

func testAccCheckAWSAPIGatewayDeploymentStageExists(resourceName string, res *apigateway.Stage) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).apigatewayconn

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Deployment not found: %s", resourceName)
		}

		req := &apigateway.GetStageInput{
			StageName: aws.String(rs.Primary.Attributes["stage_name"]),
			RestApiId: aws.String(rs.Primary.Attributes["rest_api_id"]),
		}
		stage, err := conn.GetStage(req)
		if err != nil {
			return err
		}

		*res = *stage

		return nil
	}
}

func testAccCheckAWSAPIGatewayDeploymentDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigatewayconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_deployment" {
			continue
		}

		req := &apigateway.GetDeploymentsInput{
			RestApiId: aws.String(rs.Primary.Attributes["rest_api_id"]),
		}
		describe, err := conn.GetDeployments(req)

		if isAWSErr(err, apigateway.ErrCodeNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if describe != nil {
			for _, deployment := range describe.Items {
				if aws.StringValue(deployment.Id) == rs.Primary.ID {
					return fmt.Errorf("API Gateway Deployment still exists")
				}
			}
		}
	}

	return nil
}

func testAccCheckAWSAPIGatewayDeploymentNotRecreated(i, j *apigateway.Deployment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CreatedDate) != aws.TimeValue(j.CreatedDate) {
			return fmt.Errorf("API Gateway Deployment recreated")
		}

		return nil
	}
}

func testAccCheckAWSAPIGatewayDeploymentRecreated(i, j *apigateway.Deployment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CreatedDate) == aws.TimeValue(j.CreatedDate) {
			return fmt.Errorf("API Gateway Deployment not recreated")
		}

		return nil
	}
}

func testAccAWSAPIGatewayDeploymentConfigBase(uri string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = "tf-acc-test-deployment"
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "NONE"
}

resource "aws_api_gateway_method_response" "error" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method
  status_code = "400"
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method

  type                    = "HTTP"
  uri                     = "%s"
  integration_http_method = "GET"
}

resource "aws_api_gateway_integration_response" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_integration.test.http_method
  status_code = aws_api_gateway_method_response.error.status_code
}
`, uri)
}

func testAccAWSAPIGatewayDeploymentConfigTriggers(description string, url string) string {
	return testAccAWSAPIGatewayDeploymentConfigBase(url) + fmt.Sprintf(`
resource "aws_api_gateway_deployment" "test" {
  description       = %[1]q
  rest_api_id       = aws_api_gateway_rest_api.test.id
  stage_description = %[1]q
  stage_name        = "tf-acc-test"

  triggers = {
    redeployment = sha1(join(",", list(
      jsonencode(aws_api_gateway_integration.test),
    )))
  }

  lifecycle {
    create_before_destroy = true
  }
}
`, description)
}

func testAccAWSAPIGatewayDeploymentConfigDescription(description string) string {
	return testAccAWSAPIGatewayDeploymentConfigBase("http://example.com") + fmt.Sprintf(`
resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  description = %q
  rest_api_id = aws_api_gateway_rest_api.test.id
}
`, description)
}

func testAccAWSAPIGatewayDeploymentConfigRequired() string {
	return testAccAWSAPIGatewayDeploymentConfigBase("http://example.com") + `
resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id = aws_api_gateway_rest_api.test.id
}
`
}

func testAccAWSAPIGatewayDeploymentConfigStageDescription(stageDescription string) string {
	return testAccAWSAPIGatewayDeploymentConfigBase("http://example.com") + fmt.Sprintf(`
resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id       = aws_api_gateway_rest_api.test.id
  stage_description = %q
  stage_name        = "tf-acc-test"
}
`, stageDescription)
}

func testAccAWSAPIGatewayDeploymentConfigStageName(stageName string) string {
	return testAccAWSAPIGatewayDeploymentConfigBase("http://example.com") + fmt.Sprintf(`
resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = %q
}
`, stageName)
}

func testAccAWSAPIGatewayDeploymentConfigVariables(key1, value1 string) string {
	return testAccAWSAPIGatewayDeploymentConfigBase("http://example.com") + fmt.Sprintf(`
resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id = aws_api_gateway_rest_api.test.id

  variables = {
    %q = %q
  }
}
`, key1, value1)
}
