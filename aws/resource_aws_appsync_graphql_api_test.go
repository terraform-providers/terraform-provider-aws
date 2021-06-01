package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_appsync_graphql_api", &resource.Sweeper{
		Name: "aws_appsync_graphql_api",
		F:    testSweepAppsyncGraphqlApis,
	})
}

func testSweepAppsyncGraphqlApis(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*AWSClient).appsyncconn

	input := &appsync.ListGraphqlApisInput{}

	for {
		output, err := conn.ListGraphqlApis(input)
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping AppSync GraphQL API sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving AppSync GraphQL APIs: %s", err)
		}

		for _, graphAPI := range output.GraphqlApis {
			id := aws.StringValue(graphAPI.ApiId)
			input := &appsync.DeleteGraphqlApiInput{
				ApiId: graphAPI.ApiId,
			}

			log.Printf("[INFO] Deleting AppSync GraphQL API %s", id)
			_, err := conn.DeleteGraphqlApi(input)

			if err != nil {
				return fmt.Errorf("error deleting AppSync GraphQL API (%s): %s", id, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}

func TestAccAWSAppsyncGraphqlApi_basic(t *testing.T) {
	var api1 appsync.GraphqlApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appsync.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_AuthenticationType(rName, "API_KEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api1),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "appsync", regexp.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "uris.%"),
					resource.TestCheckResourceAttrSet(resourceName, "uris.GRAPHQL"),
					resource.TestCheckNoResourceAttr(resourceName, "tags"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "xray_enabled", "false"),
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

func TestAccAWSAppsyncGraphqlApi_disappears(t *testing.T) {
	var api1 appsync.GraphqlApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appsync.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_AuthenticationType(rName, "API_KEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api1),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsAppsyncGraphqlApi(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSAppsyncGraphqlApi_Schema(t *testing.T) {
	var api1, api2 appsync.GraphqlApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appsync.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_Schema(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api1),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "appsync", regexp.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "xray_enabled", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "schema"),
					resource.TestCheckResourceAttrSet(resourceName, "uris.%"),
					resource.TestCheckResourceAttrSet(resourceName, "uris.GRAPHQL"),
					testAccCheckAwsAppsyncTypeExists(resourceName, "Post"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"schema"},
			},
			{
				Config: testAccAppsyncGraphqlApiConfig_SchemaUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api2),
					testAccCheckAwsAppsyncTypeExists(resourceName, "PostV2"),
				),
			},
		},
	})
}

func TestAccAWSAppsyncGraphqlApi_AuthenticationType(t *testing.T) {
	var api1, api2 appsync.GraphqlApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appsync.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_AuthenticationType(rName, "API_KEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
				),
			},
			{
				Config: testAccAppsyncGraphqlApiConfig_AuthenticationType(rName, "AWS_IAM"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_IAM"),
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

func TestAccAWSAppsyncGraphqlApi_AuthenticationType_APIKey(t *testing.T) {
	var api1 appsync.GraphqlApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appsync.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_AuthenticationType(rName, "API_KEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api1),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "appsync", regexp.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccAWSAppsyncGraphqlApi_AuthenticationType_AWSIAM(t *testing.T) {
	var api1 appsync.GraphqlApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appsync.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_AuthenticationType(rName, "AWS_IAM"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api1),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "appsync", regexp.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_IAM"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccAWSAppsyncGraphqlApi_AuthenticationType_AmazonCognitoUserPools(t *testing.T) {
	var api1 appsync.GraphqlApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	cognitoUserPoolResourceName := "aws_cognito_user_pool.test"
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appsync.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_UserPoolConfig_DefaultAction(rName, "ALLOW"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AMAZON_COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.aws_region", testAccGetRegion()),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.default_action", "ALLOW"),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_config.0.user_pool_id", cognitoUserPoolResourceName, "id"),
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

func TestAccAWSAppsyncGraphqlApi_AuthenticationType_OpenIDConnect(t *testing.T) {
	var api1 appsync.GraphqlApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appsync.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_OpenIDConnectConfig_Issuer(rName, "https://example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.issuer", "https://example.com"),
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

func TestAccAWSAppsyncGraphqlApi_LogConfig(t *testing.T) {
	var api1 appsync.GraphqlApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appsync.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_LogConfig_FieldLogLevel(rName, "ALL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "log_config.0.cloudwatch_logs_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.field_log_level", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.exclude_verbose_content", "false"),
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

func TestAccAWSAppsyncGraphqlApi_LogConfig_FieldLogLevel(t *testing.T) {
	var api1, api2, api3 appsync.GraphqlApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appsync.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_LogConfig_FieldLogLevel(rName, "ALL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "log_config.0.cloudwatch_logs_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.field_log_level", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.exclude_verbose_content", "false"),
				),
			},
			{
				Config: testAccAppsyncGraphqlApiConfig_LogConfig_FieldLogLevel(rName, "ERROR"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "log_config.0.cloudwatch_logs_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.field_log_level", "ERROR"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.exclude_verbose_content", "false"),
				),
			},
			{
				Config: testAccAppsyncGraphqlApiConfig_LogConfig_FieldLogLevel(rName, "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api3),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "log_config.0.cloudwatch_logs_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.field_log_level", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.exclude_verbose_content", "false"),
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

func TestAccAWSAppsyncGraphqlApi_LogConfig_ExcludeVerboseContent(t *testing.T) {
	var api1, api2 appsync.GraphqlApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appsync.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_LogConfig_ExcludeVerboseContent(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "log_config.0.cloudwatch_logs_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.field_log_level", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.exclude_verbose_content", "false"),
				),
			},
			{
				Config: testAccAppsyncGraphqlApiConfig_LogConfig_ExcludeVerboseContent(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "log_config.0.cloudwatch_logs_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.field_log_level", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.exclude_verbose_content", "true"),
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

func TestAccAWSAppsyncGraphqlApi_OpenIDConnectConfig_AuthTTL(t *testing.T) {
	var api1, api2 appsync.GraphqlApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appsync.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_OpenIDConnectConfig_AuthTTL(rName, 1000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.auth_ttl", "1000"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.issuer", "https://example.com"),
				),
			},
			{
				Config: testAccAppsyncGraphqlApiConfig_OpenIDConnectConfig_AuthTTL(rName, 2000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.auth_ttl", "2000"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.issuer", "https://example.com"),
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

func TestAccAWSAppsyncGraphqlApi_OpenIDConnectConfig_ClientID(t *testing.T) {
	var api1, api2 appsync.GraphqlApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appsync.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_OpenIDConnectConfig_ClientID(rName, "ClientID1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.client_id", "ClientID1"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.issuer", "https://example.com"),
				),
			},
			{
				Config: testAccAppsyncGraphqlApiConfig_OpenIDConnectConfig_ClientID(rName, "ClientID2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.client_id", "ClientID2"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.issuer", "https://example.com"),
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

func TestAccAWSAppsyncGraphqlApi_OpenIDConnectConfig_IatTTL(t *testing.T) {
	var api1, api2 appsync.GraphqlApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appsync.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_OpenIDConnectConfig_IatTTL(rName, 1000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.iat_ttl", "1000"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.issuer", "https://example.com"),
				),
			},
			{
				Config: testAccAppsyncGraphqlApiConfig_OpenIDConnectConfig_IatTTL(rName, 2000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.iat_ttl", "2000"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.issuer", "https://example.com"),
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

func TestAccAWSAppsyncGraphqlApi_OpenIDConnectConfig_Issuer(t *testing.T) {
	var api1, api2 appsync.GraphqlApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appsync.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_OpenIDConnectConfig_Issuer(rName, "https://example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.issuer", "https://example.com"),
				),
			},
			{
				Config: testAccAppsyncGraphqlApiConfig_OpenIDConnectConfig_Issuer(rName, "https://example.org"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.issuer", "https://example.org"),
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

func TestAccAWSAppsyncGraphqlApi_Name(t *testing.T) {
	var api1, api2 appsync.GraphqlApi
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appsync.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_AuthenticationType(rName1, "API_KEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
				),
			},
			{
				Config: testAccAppsyncGraphqlApiConfig_AuthenticationType(rName2, "API_KEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
		},
	})
}

func TestAccAWSAppsyncGraphqlApi_UserPoolConfig_AwsRegion(t *testing.T) {
	var api1, api2 appsync.GraphqlApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	cognitoUserPoolResourceName := "aws_cognito_user_pool.test"
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appsync.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_UserPoolConfig_AwsRegion(rName, testAccGetRegion()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AMAZON_COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.aws_region", testAccGetRegion()),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.default_action", "ALLOW"),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_config.0.user_pool_id", cognitoUserPoolResourceName, "id"),
				),
			},
			{
				Config: testAccAppsyncGraphqlApiConfig_UserPoolConfig_DefaultAction(rName, "ALLOW"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AMAZON_COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.aws_region", testAccGetRegion()),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.default_action", "ALLOW"),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_config.0.user_pool_id", cognitoUserPoolResourceName, "id"),
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

func TestAccAWSAppsyncGraphqlApi_UserPoolConfig_DefaultAction(t *testing.T) {
	var api1, api2 appsync.GraphqlApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	cognitoUserPoolResourceName := "aws_cognito_user_pool.test"
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appsync.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_UserPoolConfig_DefaultAction(rName, "ALLOW"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AMAZON_COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.aws_region", testAccGetRegion()),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.default_action", "ALLOW"),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_config.0.user_pool_id", cognitoUserPoolResourceName, "id"),
				),
			},
			{
				Config: testAccAppsyncGraphqlApiConfig_UserPoolConfig_DefaultAction(rName, "DENY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AMAZON_COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.aws_region", testAccGetRegion()),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.0.default_action", "DENY"),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_config.0.user_pool_id", cognitoUserPoolResourceName, "id"),
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

func TestAccAWSAppsyncGraphqlApi_Tags(t *testing.T) {
	var api1 appsync.GraphqlApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appsync.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_Tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value One"),
					resource.TestCheckResourceAttr(resourceName, "tags.Description", "Very interesting"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppsyncGraphqlApiConfig_TagsModified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value One Changed"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value Two"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value Three"),
				),
			},
		},
	})
}

func TestAccAWSAppsyncGraphqlApi_AdditionalAuthentication_APIKey(t *testing.T) {
	var api1 appsync.GraphqlApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appsync.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_AdditionalAuth_AuthType(rName, "AWS_IAM", "API_KEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api1),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "appsync", regexp.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "AWS_IAM"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.authentication_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.openid_connect_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.user_pool_config.#", "0"),
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

func TestAccAWSAppsyncGraphqlApi_AdditionalAuthentication_AWSIAM(t *testing.T) {
	var api1 appsync.GraphqlApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appsync.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_AdditionalAuth_AuthType(rName, "API_KEY", "AWS_IAM"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api1),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "appsync", regexp.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.authentication_type", "AWS_IAM"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.openid_connect_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.user_pool_config.#", "0"),
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

func TestAccAWSAppsyncGraphqlApi_AdditionalAuthentication_CognitoUserPools(t *testing.T) {
	var api1 appsync.GraphqlApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	cognitoUserPoolResourceName := "aws_cognito_user_pool.test"
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appsync.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_AdditionalAuth_UserPoolConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api1),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "appsync", regexp.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.authentication_type", "AMAZON_COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.openid_connect_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.user_pool_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "additional_authentication_provider.0.user_pool_config.0.user_pool_id", cognitoUserPoolResourceName, "id"),
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

func TestAccAWSAppsyncGraphqlApi_AdditionalAuthentication_OpenIDConnect(t *testing.T) {
	var api1 appsync.GraphqlApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appsync.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_AdditionalAuth_OpenIdConnect(rName, "https://example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api1),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "appsync", regexp.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.user_pool_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.openid_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.openid_connect_config.0.issuer", "https://example.com"),
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

func TestAccAWSAppsyncGraphqlApi_AdditionalAuthentication_Multiple(t *testing.T) {
	var api1 appsync.GraphqlApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	cognitoUserPoolResourceName := "aws_cognito_user_pool.test"
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appsync.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_AdditionalAuth_Multiple(rName, "https://example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api1),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "appsync", regexp.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.authentication_type", "AWS_IAM"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.user_pool_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.0.openid_connect_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.1.authentication_type", "AMAZON_COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.1.openid_connect_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.1.user_pool_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "additional_authentication_provider.1.user_pool_config.0.user_pool_id", cognitoUserPoolResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.2.authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.2.user_pool_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.2.openid_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "additional_authentication_provider.2.openid_connect_config.0.issuer", "https://example.com"),
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

func TestAccAWSAppsyncGraphqlApi_XrayEnabled(t *testing.T) {
	var api1, api2 appsync.GraphqlApi
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appsync.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_XrayEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api1),
					resource.TestCheckResourceAttr(resourceName, "xray_enabled", "true"),
				),
			},
			{
				Config: testAccAppsyncGraphqlApiConfig_XrayEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName, &api2),
					resource.TestCheckResourceAttr(resourceName, "xray_enabled", "false"),
				),
			},
		},
	})
}

func testAccCheckAwsAppsyncGraphqlApiDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).appsyncconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appsync_graphql_api" {
			continue
		}

		input := &appsync.GetGraphqlApiInput{
			ApiId: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetGraphqlApi(input)
		if err != nil {
			if isAWSErr(err, appsync.ErrCodeNotFoundException, "") {
				return nil
			}
			return err
		}
	}
	return nil
}

func testAccCheckAwsAppsyncGraphqlApiExists(name string, api *appsync.GraphqlApi) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).appsyncconn

		input := &appsync.GetGraphqlApiInput{
			ApiId: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetGraphqlApi(input)

		if err != nil {
			return err
		}

		*api = *output.GraphqlApi

		return nil
	}
}

func testAccCheckAwsAppsyncTypeExists(name, typeName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).appsyncconn

		input := &appsync.GetTypeInput{
			ApiId:    aws.String(rs.Primary.ID),
			TypeName: aws.String(typeName),
			Format:   aws.String(appsync.OutputTypeSdl),
		}

		_, err := conn.GetType(input)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccAppsyncGraphqlApiConfig_AuthenticationType(rName, authenticationType string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = %q
  name                = %q
}
`, authenticationType, rName)
}

func testAccAppsyncGraphqlApiConfig_LogConfig_FieldLogLevel(rName, fieldLogLevel string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %q

  assume_role_policy = <<POLICY
{
    "Version": "2012-10-17",
    "Statement": [
        {
        "Effect": "Allow",
        "Principal": {
            "Service": "appsync.amazonaws.com"
        },
        "Action": "sts:AssumeRole"
        }
    ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSAppSyncPushToCloudWatchLogs"
  role       = aws_iam_role.test.name
}

resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q

  log_config {
    cloudwatch_logs_role_arn = aws_iam_role.test.arn
    field_log_level          = %q
  }
}
`, rName, rName, fieldLogLevel)
}

func testAccAppsyncGraphqlApiConfig_LogConfig_ExcludeVerboseContent(rName string, excludeVerboseContent bool) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %q

  assume_role_policy = <<POLICY
{
    "Version": "2012-10-17",
    "Statement": [
        {
        "Effect": "Allow",
        "Principal": {
            "Service": "appsync.amazonaws.com"
        },
        "Action": "sts:AssumeRole"
        }
    ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSAppSyncPushToCloudWatchLogs"
  role       = aws_iam_role.test.name
}

resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q

  log_config {
    cloudwatch_logs_role_arn = aws_iam_role.test.arn
    field_log_level          = "ALL"
    exclude_verbose_content  = %t
  }
}
`, rName, rName, excludeVerboseContent)
}

func testAccAppsyncGraphqlApiConfig_OpenIDConnectConfig_AuthTTL(rName string, authTTL int) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "OPENID_CONNECT"
  name                = %q

  openid_connect_config {
    auth_ttl = %d
    issuer   = "https://example.com"
  }
}
`, rName, authTTL)
}

func testAccAppsyncGraphqlApiConfig_OpenIDConnectConfig_ClientID(rName, clientID string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "OPENID_CONNECT"
  name                = %q

  openid_connect_config {
    client_id = %q
    issuer    = "https://example.com"
  }
}
`, rName, clientID)
}

func testAccAppsyncGraphqlApiConfig_OpenIDConnectConfig_IatTTL(rName string, iatTTL int) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "OPENID_CONNECT"
  name                = %q

  openid_connect_config {
    iat_ttl = %d
    issuer  = "https://example.com"
  }
}
`, rName, iatTTL)
}

func testAccAppsyncGraphqlApiConfig_OpenIDConnectConfig_Issuer(rName, issuer string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "OPENID_CONNECT"
  name                = %q

  openid_connect_config {
    issuer = %q
  }
}
`, rName, issuer)
}

func testAccAppsyncGraphqlApiConfig_UserPoolConfig_AwsRegion(rName, awsRegion string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %q
}

resource "aws_appsync_graphql_api" "test" {
  authentication_type = "AMAZON_COGNITO_USER_POOLS"
  name                = %q

  user_pool_config {
    aws_region     = %q
    default_action = "ALLOW"
    user_pool_id   = aws_cognito_user_pool.test.id
  }
}
`, rName, rName, awsRegion)
}

func testAccAppsyncGraphqlApiConfig_UserPoolConfig_DefaultAction(rName, defaultAction string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %q
}

resource "aws_appsync_graphql_api" "test" {
  authentication_type = "AMAZON_COGNITO_USER_POOLS"
  name                = %q

  user_pool_config {
    default_action = %q
    user_pool_id   = aws_cognito_user_pool.test.id
  }
}
`, rName, rName, defaultAction)
}

func testAccAppsyncGraphqlApiConfig_Schema(rName string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q
  schema              = "type Mutation {\n\tputPost(id: ID!, title: String!): Post\n}\n\ntype Post {\n\tid: ID!\n\ttitle: String!\n}\n\ntype Query {\n\tsinglePost(id: ID!): Post\n}\n\nschema {\n\tquery: Query\n\tmutation: Mutation\n\n}\n"
}
`, rName)
}

func testAccAppsyncGraphqlApiConfig_SchemaUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q
  schema              = "type Mutation {\n\tputPostV2(id: ID!, title: String!): PostV2\n}\n\ntype PostV2 {\n\tid: ID!\n\ttitle: String!\n}\n\ntype Query {\n\tsinglePostV2(id: ID!): PostV2\n}\n\nschema {\n\tquery: Query\n\tmutation: Mutation\n\n}\n"
}
`, rName)
}

func testAccAppsyncGraphqlApiConfig_Tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q

  tags = {
    Key1        = "Value One"
    Description = "Very interesting"
  }
}
`, rName)
}

func testAccAppsyncGraphqlApiConfig_TagsModified(rName string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q

  tags = {
    Key1 = "Value One Changed"
    Key2 = "Value Two"
    Key3 = "Value Three"
  }
}
`, rName)
}

func testAccAppsyncGraphqlApiConfig_AdditionalAuth_AuthType(rName, defaultAuthType, additionalAuthType string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = %q
  name                = %q

  additional_authentication_provider {
    authentication_type = %q
  }
}`, defaultAuthType, rName, additionalAuthType)
}

func testAccAppsyncGraphqlApiConfig_AdditionalAuth_UserPoolConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %q
}

resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q

  additional_authentication_provider {
    authentication_type = "AMAZON_COGNITO_USER_POOLS"

    user_pool_config {
      user_pool_id = aws_cognito_user_pool.test.id
    }
  }
}
`, rName, rName)
}

func testAccAppsyncGraphqlApiConfig_AdditionalAuth_OpenIdConnect(rName, issuer string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q

  additional_authentication_provider {
    authentication_type = "OPENID_CONNECT"

    openid_connect_config {
      issuer = %q
    }
  }
}
`, rName, issuer)
}

func testAccAppsyncGraphqlApiConfig_AdditionalAuth_Multiple(rName, issuer string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %q
}

resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q

  additional_authentication_provider {
    authentication_type = "AWS_IAM"
  }

  additional_authentication_provider {
    authentication_type = "AMAZON_COGNITO_USER_POOLS"

    user_pool_config {
      user_pool_id = aws_cognito_user_pool.test.id
    }
  }

  additional_authentication_provider {
    authentication_type = "OPENID_CONNECT"

    openid_connect_config {
      issuer = %q
    }
  }
}
`, rName, rName, issuer)
}

func testAccAppsyncGraphqlApiConfig_XrayEnabled(rName string, xrayEnabled bool) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q
  xray_enabled        = %t
}
`, rName, xrayEnabled)
}
