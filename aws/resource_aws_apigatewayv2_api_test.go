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
)

func init() {
	resource.AddTestSweepers("aws_apigatewayv2_api", &resource.Sweeper{
		Name: "aws_apigatewayv2_api",
		F:    testSweepAPIGatewayV2Apis,
		Dependencies: []string{
			"aws_apigatewayv2_domain_name",
		},
	})
}

func testSweepAPIGatewayV2Apis(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).apigatewayv2conn
	input := &apigatewayv2.GetApisInput{}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.GetApis(input)
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping API Gateway v2 API sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving API Gateway v2 APIs: %s", err)
		}

		for _, api := range output.Items {
			log.Printf("[INFO] Deleting API Gateway v2 API: %s", aws.StringValue(api.ApiId))
			_, err := conn.DeleteApi(&apigatewayv2.DeleteApiInput{
				ApiId: api.ApiId,
			})
			if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
				continue
			}
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting API Gateway v2 API (%s): %w", aws.StringValue(api.ApiId), err)
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

func TestAccAWSAPIGatewayV2Api_basicWebSocket(t *testing.T) {
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2ApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2ApiConfig_basicWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2ApiExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "false"),
					testAccMatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(`.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeWebsocket),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.body.action"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", ""),
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

func TestAccAWSAPIGatewayV2Api_basicHttp(t *testing.T) {
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2ApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2ApiConfig_basicHttp(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2ApiExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "false"),
					testAccMatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(`.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.method $request.path"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", ""),
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

func TestAccAWSAPIGatewayV2Api_disappears(t *testing.T) {
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2ApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2ApiConfig_basicWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2ApiExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsApiGatewayV2Api(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayV2Api_AllAttributesWebSocket(t *testing.T) {
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2ApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2ApiConfig_allAttributesWebSocket(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2ApiExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$context.authorizer.usageIdentifierKey"),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", "test description"),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "true"),
					testAccMatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(`.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeWebsocket),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.body.service"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "v1"),
				),
			},
			{
				Config: testAccAWSAPIGatewayV2ApiConfig_basicWebSocket(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2ApiExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeWebsocket),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.body.action"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", ""),
				),
			},
			{
				Config: testAccAWSAPIGatewayV2ApiConfig_allAttributesWebSocket(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2ApiExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$context.authorizer.usageIdentifierKey"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", "test description"),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "true"),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.body.service"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "v1"),
				),
			},
			{
				Config: testAccAWSAPIGatewayV2ApiConfig_allAttributesWebSocket(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2ApiExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$context.authorizer.usageIdentifierKey"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", "test description"),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "true"),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.body.service"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "v1"),
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

func TestAccAWSAPIGatewayV2Api_AllAttributesHttp(t *testing.T) {
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2ApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2ApiConfig_allAttributesHttp(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2ApiExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", "test description"),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "true"),
					testAccMatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(`.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.method $request.path"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "v1"),
				),
			},
			{
				Config: testAccAWSAPIGatewayV2ApiConfig_basicHttp(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2ApiExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.method $request.path"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", ""),
				),
			},
			{
				Config: testAccAWSAPIGatewayV2ApiConfig_allAttributesHttp(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2ApiExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", "test description"),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "true"),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.method $request.path"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "v1"),
				),
			},
			{
				Config: testAccAWSAPIGatewayV2ApiConfig_allAttributesHttp(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2ApiExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", "test description"),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "true"),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.method $request.path"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "v1"),
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

func TestAccAWSAPIGatewayV2Api_Openapi(t *testing.T) {
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2ApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2ApiConfig_OpenAPI(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2ApiExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "version", ""),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					testAccCheckAWSAPIGatewayV2ApiRoutes(&v, []string{"GET /test"}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			{
				Config: testAccAWSAPIGatewayV2ApiConfig_UpdatedOpenAPIYaml(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "version", ""),
					testAccCheckAWSAPIGatewayV2ApiExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					testAccCheckAWSAPIGatewayV2ApiRoutes(&v, []string{"GET /update"}),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayV2Api_Openapi_WithTags(t *testing.T) {
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2ApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2ApiConfig_OpenAPIYaml_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2ApiExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					testAccCheckAWSAPIGatewayV2ApiRoutes(&v, []string{"GET /test"}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			{
				Config: testAccAWSAPIGatewayV2ApiConfig_OpenAPIYaml_tagsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2ApiExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					testAccCheckAWSAPIGatewayV2ApiRoutes(&v, []string{"GET /update"}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1U"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2U"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayV2Api_Openapi_WithCorsConfiguration(t *testing.T) {
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2ApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2ApiConfig_OpenAPIYaml_corsConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2ApiExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_methods.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_methods.*", "delete"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_origins.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_origins.*", "https://www.google.de"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			{
				Config: testAccAWSAPIGatewayV2ApiConfig_OpenAPIYaml_corsConfigurationUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2ApiExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					testAccCheckAWSAPIGatewayV2ApiRoutes(&v, []string{"GET /update"}),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "0"),
				),
			},
			{
				Config: testAccAWSAPIGatewayV2ApiConfig_OpenAPIYaml_corsConfigurationUpdated2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2ApiExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					testAccCheckAWSAPIGatewayV2ApiRoutes(&v, []string{"GET /update"}),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_methods.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_methods.*", "get"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_methods.*", "put"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_origins.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_origins.*", "https://www.example.com"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_origins.*", "https://www.google.de"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayV2Api_OpenapiWithMoreFields(t *testing.T) {
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2ApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2ApiConfig_OpenAPIYaml(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2ApiExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "version", ""),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					testAccCheckAWSAPIGatewayV2ApiRoutes(&v, []string{"GET /test"}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			{
				Config: testAccAWSAPIGatewayV2ApiConfig_UpdatedOpenAPI2(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "description test"),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "version", "2017-04-21T04:08:08Z"),
					testAccCheckAWSAPIGatewayV2ApiExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					testAccCheckAWSAPIGatewayV2ApiRoutes(&v, []string{"GET /update"}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
		},
	})
}

func testAccCheckAWSAPIGatewayV2ApiRoutes(v *apigatewayv2.GetApiOutput, routes []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

		resp, err := conn.GetRoutes(&apigatewayv2.GetRoutesInput{
			ApiId: v.ApiId,
		})
		if err != nil {
			return err
		}

		actualRoutePaths := map[string]bool{}
		for _, route := range resp.Items {
			actualRoutePaths[*route.RouteKey] = true
		}

		for _, route := range routes {
			if _, ok := actualRoutePaths[route]; !ok {
				return fmt.Errorf("Expected path %v but did not find it in %v", route, actualRoutePaths)
			}
			delete(actualRoutePaths, route)
		}

		if len(actualRoutePaths) > 0 {
			return fmt.Errorf("Found unexpected paths %v", actualRoutePaths)
		}

		return nil
	}
}

func TestAccAWSAPIGatewayV2Api_Tags(t *testing.T) {
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2ApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2ApiConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2ApiExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccMatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(`.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeWebsocket),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.body.action"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2"),
					resource.TestCheckResourceAttr(resourceName, "version", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGatewayV2ApiConfig_basicWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2ApiExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeWebsocket),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.body.action"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", ""),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayV2Api_CorsConfiguration(t *testing.T) {
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2ApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2ApiConfig_corsConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2ApiExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_credentials", "false"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_headers.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_headers.*", "Authorization"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_methods.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_methods.*", "GET"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_methods.*", "put"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_origins.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_origins.*", "https://www.example.com"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.expose_headers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.max_age", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccMatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(`.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.method $request.path"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGatewayV2ApiConfig_corsConfigurationUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2ApiExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_credentials", "true"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_headers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_methods.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_methods.*", "*"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_origins.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_origins.*", "HTTP://WWW.EXAMPLE.ORG"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_origins.*", "https://example.io"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.expose_headers.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.expose_headers.*", "X-Api-Id"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.max_age", "500"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccMatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(`.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.method $request.path"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", ""),
				),
			},
			{
				Config: testAccAWSAPIGatewayV2ApiConfig_basicHttp(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2ApiExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccMatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(`.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.method $request.path"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", ""),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayV2Api_QuickCreate(t *testing.T) {
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2ApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2ApiConfig_quickCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2ApiExists(resourceName, &v),
					testAccCheckAWSAPIGatewayV2ApiQuickCreateIntegration(resourceName, "HTTP_PROXY", "http://www.example.com/"),
					testAccCheckAWSAPIGatewayV2ApiQuickCreateRoute(resourceName, "GET /pets"),
					testAccCheckAWSAPIGatewayV2ApiQuickCreateStage(resourceName, "$default"),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccMatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(`.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					resource.TestCheckResourceAttr(resourceName, "route_key", "GET /pets"),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.method $request.path"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "target", "http://www.example.com/"),
					resource.TestCheckResourceAttr(resourceName, "version", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"route_key",
					"target",
				},
			},
		},
	})
}

func testAccCheckAWSAPIGatewayV2ApiDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_apigatewayv2_api" {
			continue
		}

		_, err := conn.GetApi(&apigatewayv2.GetApiInput{
			ApiId: aws.String(rs.Primary.ID),
		})
		if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("API Gateway v2 API %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSAPIGatewayV2ApiExists(n string, v *apigatewayv2.GetApiOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 API ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

		resp, err := conn.GetApi(&apigatewayv2.GetApiInput{
			ApiId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccCheckAWSAPIGatewayV2ApiQuickCreateIntegration(n, expectedType, expectedUri string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 API ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

		resp, err := conn.GetIntegrations(&apigatewayv2.GetIntegrationsInput{
			ApiId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if got := len(resp.Items); got != 1 {
			return fmt.Errorf("Incorrect number of integrations: %d", got)
		}

		if got := aws.StringValue(resp.Items[0].IntegrationType); got != expectedType {
			return fmt.Errorf("Incorrect integration type. Expected: %s, got: %s", expectedType, got)
		}
		if got := aws.StringValue(resp.Items[0].IntegrationUri); got != expectedUri {
			return fmt.Errorf("Incorrect integration URI. Expected: %s, got: %s", expectedUri, got)
		}

		return nil
	}
}

func testAccCheckAWSAPIGatewayV2ApiQuickCreateRoute(n, expectedRouteKey string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 API ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

		resp, err := conn.GetRoutes(&apigatewayv2.GetRoutesInput{
			ApiId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if got := len(resp.Items); got != 1 {
			return fmt.Errorf("Incorrect number of routes: %d", got)
		}

		if got := aws.StringValue(resp.Items[0].RouteKey); got != expectedRouteKey {
			return fmt.Errorf("Incorrect route key. Expected: %s, got: %s", expectedRouteKey, got)
		}

		return nil
	}
}

func testAccCheckAWSAPIGatewayV2ApiQuickCreateStage(n, expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 API ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

		resp, err := conn.GetStages(&apigatewayv2.GetStagesInput{
			ApiId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if got := len(resp.Items); got != 1 {
			return fmt.Errorf("Incorrect number of stages: %d", got)
		}

		if got := aws.StringValue(resp.Items[0].StageName); got != expectedName {
			return fmt.Errorf("Incorrect stage name. Expected: %s, got: %s", expectedName, got)
		}

		return nil
	}
}

func testAccAWSAPIGatewayV2ApiConfig_basicWebSocket(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name                       = %[1]q
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.action"
}
`, rName)
}

func testAccAWSAPIGatewayV2ApiConfig_basicHttp(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = %[1]q
  protocol_type = "HTTP"
}
`, rName)
}

func testAccAWSAPIGatewayV2ApiConfig_allAttributesWebSocket(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  api_key_selection_expression = "$context.authorizer.usageIdentifierKey"
  description                  = "test description"
  disable_execute_api_endpoint = true
  name                         = %[1]q
  protocol_type                = "WEBSOCKET"
  route_selection_expression   = "$request.body.service"
  version                      = "v1"
}
`, rName)
}

func testAccAWSAPIGatewayV2ApiConfig_allAttributesHttp(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  description                  = "test description"
  disable_execute_api_endpoint = true
  name                         = %[1]q
  protocol_type                = "HTTP"
  version                      = "v1"
}
`, rName)
}

func testAccAWSAPIGatewayV2ApiConfig_tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name                       = %[1]q
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.action"

  tags = {
    Key1 = "Value1"
    Key2 = "Value2"
  }
}
`, rName)
}

func testAccAWSAPIGatewayV2ApiConfig_corsConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = %[1]q
  protocol_type = "HTTP"

  cors_configuration {
    allow_headers = ["Authorization"]
    allow_methods = ["GET", "put"]
    allow_origins = ["https://www.example.com"]
  }
}
`, rName)
}

func testAccAWSAPIGatewayV2ApiConfig_corsConfigurationUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = %[1]q
  protocol_type = "HTTP"

  cors_configuration {
    allow_credentials = true
    allow_methods     = ["*"]
    allow_origins     = ["HTTP://WWW.EXAMPLE.ORG", "https://example.io"]
    expose_headers    = ["X-Api-Id"]
    max_age           = 500
  }
}
`, rName)
}

func testAccAWSAPIGatewayV2ApiConfig_quickCreate(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = %[1]q
  protocol_type = "HTTP"
  target        = "http://www.example.com/"
  route_key     = "GET /pets"
}
`, rName)
}

func testAccAWSAPIGatewayV2ApiConfig_OpenAPI(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = "%s"
  protocol_type = "HTTP"
  body          = <<EOF
{
  "openapi": "3.0.1",
  "info": {
    "title": "%s_DIFFERENT",
    "version": "1.0"
  },
  "paths": {
    "/test": {
      "get": {
        "x-amazon-apigateway-integration": {
          "type": "HTTP_PROXY",
          "httpMethod": "GET",
          "payloadFormatVersion": "1.0",
          "uri": "https://www.google.de"
        }
      }
    }
  }
}
EOF
}
`, rName, rName)
}

func testAccAWSAPIGatewayV2ApiConfig_OpenAPIYaml(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = "%s"
  protocol_type = "HTTP"
  body          = <<EOF
---
openapi: 3.0.1
info:
  title: %s_DIFFERENT
  version: 1.0
paths:
  "/test":
    get:
      x-amazon-apigateway-integration:
        type: HTTP_PROXY
        httpMethod: GET
        payloadFormatVersion: '1.0'
        uri: https://www.google.de
EOF
}
`, rName, rName)
}

func testAccAWSAPIGatewayV2ApiConfig_OpenAPIYaml_corsConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = "%s"
  protocol_type = "HTTP"
  cors_configuration {
    allow_methods = ["delete"]
    allow_origins = ["https://www.google.de"]
  }
  body = <<EOF
---
openapi: 3.0.1
info:
  title: %s_DIFFERENT
  version: 2.0
x-amazon-apigateway-cors:
  allow_methods:
    - delete
  allow_origins:
    - https://www.google.de
paths:
  "/test":
    get:
      x-amazon-apigateway-integration:
        type: HTTP_PROXY
        httpMethod: GET
        payloadFormatVersion: '1.0'
        uri: https://www.google.de
EOF
}
`, rName, rName)
}

func testAccAWSAPIGatewayV2ApiConfig_OpenAPIYaml_corsConfigurationUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = "%s"
  protocol_type = "HTTP"
  body          = <<EOF
---
openapi: 3.0.1
info:
  title: %s_DIFFERENT
  version: 2.0
x-amazon-apigateway-cors:
  allow_methods:
    - delete
  allow_origins:
    - https://www.google.de
paths:
  "/update":
    get:
      x-amazon-apigateway-integration:
        type: HTTP_PROXY
        httpMethod: GET
        payloadFormatVersion: 1.0
        uri: https://www.google.de
EOF
}
`, rName, rName)
}

func testAccAWSAPIGatewayV2ApiConfig_OpenAPIYaml_corsConfigurationUpdated2(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = "%s"
  protocol_type = "HTTP"
  cors_configuration {
    allow_methods = ["put", "get"]
    allow_origins = ["https://www.google.de", "https://www.example.com"]
  }
  body = <<EOF
---
openapi: 3.0.1
info:
  title: %s_DIFFERENT
  version: 2.0
x-amazon-apigateway-cors:
  allow_methods:
    - delete
  allow_origins:
    - https://www.google.de
paths:
  "/update":
    get:
      x-amazon-apigateway-integration:
        type: HTTP_PROXY
        httpMethod: GET
        payloadFormatVersion: 1.0
        uri: https://www.google.de
EOF
}
`, rName, rName)
}

func testAccAWSAPIGatewayV2ApiConfig_OpenAPIYaml_tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = "%s"
  protocol_type = "HTTP"
  tags = {
    Key1 = "Value1"
    Key2 = "Value2"
  }
  body = <<EOF
---
openapi: 3.0.1
info:
  title: %s_DIFFERENT
  version: 2.0
tags:
  - name: Key1
    x-amazon-apigateway-tag-value: Value3
paths:
  "/test":
    get:
      x-amazon-apigateway-integration:
        type: HTTP_PROXY
        httpMethod: GET
        payloadFormatVersion: '1.0'
        uri: https://www.google.de
EOF
}
`, rName, rName)
}

func testAccAWSAPIGatewayV2ApiConfig_OpenAPIYaml_tagsUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = "%s"
  protocol_type = "HTTP"
  tags = {
    Key1 = "Value1U"
    Key2 = "Value2U"
  }
  body = <<EOF
---
openapi: 3.0.1
info:
  title: %s_DIFFERENT
  version: 2.0
tags:
  - name: Key3
    x-amazon-apigateway-tag-value: Value3
  - name: Key4
    x-amazon-apigateway-tag-value: Value3
paths:
  "/update":
    get:
      x-amazon-apigateway-integration:
        type: HTTP_PROXY
        httpMethod: GET
        payloadFormatVersion: 1.0
        uri: https://www.google.de
EOF
}
`, rName, rName)
}

func testAccAWSAPIGatewayV2ApiConfig_UpdatedOpenAPIYaml(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = "%s"
  protocol_type = "HTTP"
  body          = <<EOF
---
openapi: 3.0.1
info:
  title: %s_DIFFERENT
  version: 2.0
paths:
  "/update":
    get:
      x-amazon-apigateway-integration:
        type: HTTP_PROXY
        httpMethod: GET
        payloadFormatVersion: 1.0
        uri: https://www.google.de
EOF
}
`, rName, rName)
}

func testAccAWSAPIGatewayV2ApiConfig_UpdatedOpenAPI2(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = "%s"
  protocol_type = "HTTP"
  version       = "2017-04-21T04:08:08Z"
  description   = "description test"
  body          = <<EOF
{
  "openapi": "3.0.1",
  "info": {
    "title": "%s_DIFFERENT",
    "version": "2.0",
    "description": "description different"
  },
  "paths": {
    "/update": {
      "get": {
        "x-amazon-apigateway-integration": {
          "type": "HTTP_PROXY",
          "httpMethod": "GET",
          "payloadFormatVersion": "1.0",
          "uri": "https://www.google.de"
        }
      }
    }
  }
}
EOF
}
`, rName, rName)
}
