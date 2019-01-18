package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSAppsyncGraphqlApi_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_AuthenticationType(rName, "API_KEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "appsync", regexp.MustCompile(`apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "user_pool_config.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "uris.%"),
					resource.TestCheckResourceAttrSet(resourceName, "uris.GRAPHQL"),
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

func TestAccAWSAppsyncGraphqlApi_AuthenticationType(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_AuthenticationType(rName, "API_KEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "API_KEY"),
				),
			},
			{
				Config: testAccAppsyncGraphqlApiConfig_AuthenticationType(rName, "AWS_IAM"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName),
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
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_AuthenticationType(rName, "API_KEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName),
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
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_AuthenticationType(rName, "AWS_IAM"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName),
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
	rName := acctest.RandomWithPrefix("tf-acc-test")
	cognitoUserPoolResourceName := "aws_cognito_user_pool.test"
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_UserPoolConfig_DefaultAction(rName, "ALLOW"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName),
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
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_OpenIDConnectConfig_Issuer(rName, "https://example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName),
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
	rName := acctest.RandomWithPrefix("tf-acc-test")
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_LogConfig_FieldLogLevel(rName, "ALL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "log_config.0.cloudwatch_logs_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.field_log_level", "ALL"),
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
	rName := acctest.RandomWithPrefix("tf-acc-test")
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_LogConfig_FieldLogLevel(rName, "ALL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "log_config.0.cloudwatch_logs_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.field_log_level", "ALL"),
				),
			},
			{
				Config: testAccAppsyncGraphqlApiConfig_LogConfig_FieldLogLevel(rName, "ERROR"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "log_config.0.cloudwatch_logs_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.field_log_level", "ERROR"),
				),
			},
			{
				Config: testAccAppsyncGraphqlApiConfig_LogConfig_FieldLogLevel(rName, "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "log_config.0.cloudwatch_logs_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.field_log_level", "NONE"),
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
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_OpenIDConnectConfig_AuthTTL(rName, 1000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.auth_ttl", "1000"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.issuer", "https://example.com"),
				),
			},
			{
				Config: testAccAppsyncGraphqlApiConfig_OpenIDConnectConfig_AuthTTL(rName, 2000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName),
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
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_OpenIDConnectConfig_ClientID(rName, "ClientID1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.client_id", "ClientID1"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.issuer", "https://example.com"),
				),
			},
			{
				Config: testAccAppsyncGraphqlApiConfig_OpenIDConnectConfig_ClientID(rName, "ClientID2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName),
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
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_OpenIDConnectConfig_IatTTL(rName, 1000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.iat_ttl", "1000"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.issuer", "https://example.com"),
				),
			},
			{
				Config: testAccAppsyncGraphqlApiConfig_OpenIDConnectConfig_IatTTL(rName, 2000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName),
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
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_OpenIDConnectConfig_Issuer(rName, "https://example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_config.0.issuer", "https://example.com"),
				),
			},
			{
				Config: testAccAppsyncGraphqlApiConfig_OpenIDConnectConfig_Issuer(rName, "https://example.org"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName),
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
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_AuthenticationType(rName1, "API_KEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
				),
			},
			{
				Config: testAccAppsyncGraphqlApiConfig_AuthenticationType(rName2, "API_KEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
		},
	})
}

func TestAccAWSAppsyncGraphqlApi_UserPoolConfig_AwsRegion(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	cognitoUserPoolResourceName := "aws_cognito_user_pool.test"
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_UserPoolConfig_AwsRegion(rName, testAccGetRegion()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName),
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
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName),
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
	rName := acctest.RandomWithPrefix("tf-acc-test")
	cognitoUserPoolResourceName := "aws_cognito_user_pool.test"
	resourceName := "aws_appsync_graphql_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_UserPoolConfig_DefaultAction(rName, "ALLOW"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName),
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
					testAccCheckAwsAppsyncGraphqlApiExists(resourceName),
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

func testAccCheckAwsAppsyncGraphqlApiExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).appsyncconn

		input := &appsync.GetGraphqlApiInput{
			ApiId: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetGraphqlApi(input)
		return err
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
  role       = "${aws_iam_role.test.name}"
}

resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q

  log_config {
    cloudwatch_logs_role_arn = "${aws_iam_role.test.arn}"
    field_log_level          = %q
  }
}
`, rName, rName, fieldLogLevel)
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
    user_pool_id   = "${aws_cognito_user_pool.test.id}"
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
    user_pool_id   = "${aws_cognito_user_pool.test.id}"
  }
}
`, rName, rName, defaultAction)
}
