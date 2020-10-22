package aws

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfawsresource"
)

func TestAccAWSCognitoUserPoolClient_basic(t *testing.T) {
	var client cognitoidentityprovider.UserPoolClientType
	userPoolName := fmt.Sprintf("tf-acc-cognito-user-pool-%s", acctest.RandString(7))
	clientName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolClientConfig_basic(userPoolName, clientName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolClientExists(resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "name", clientName),
					resource.TestCheckResourceAttr(resourceName, "explicit_auth_flows.#", "1"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "explicit_auth_flows.*", "ADMIN_NO_SRP_AUTH"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSCognitoUserPoolClientImportStateIDFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCognitoUserPoolClient_RefreshTokenValidity(t *testing.T) {
	var client cognitoidentityprovider.UserPoolClientType
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolClientConfig_RefreshTokenValidity(rName, 60),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolClientExists(resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "refresh_token_validity", "60"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSCognitoUserPoolClientImportStateIDFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoUserPoolClientConfig_RefreshTokenValidity(rName, 120),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolClientExists(resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "refresh_token_validity", "120"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPoolClient_Name(t *testing.T) {
	var client cognitoidentityprovider.UserPoolClientType
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolClientConfig_Name(rName, "name1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolClientExists(resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "name", "name1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSCognitoUserPoolClientImportStateIDFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoUserPoolClientConfig_Name(rName, "name2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolClientExists(resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "name", "name2"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPoolClient_allFields(t *testing.T) {
	var client cognitoidentityprovider.UserPoolClientType
	userPoolName := fmt.Sprintf("tf-acc-cognito-user-pool-%s", acctest.RandString(7))
	clientName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolClientConfig_allFields(userPoolName, clientName, 300),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolClientExists(resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "name", clientName),
					resource.TestCheckResourceAttr(resourceName, "explicit_auth_flows.#", "3"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "explicit_auth_flows.*", "CUSTOM_AUTH_FLOW_ONLY"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "explicit_auth_flows.*", "USER_PASSWORD_AUTH"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "explicit_auth_flows.*", "ADMIN_NO_SRP_AUTH"),
					resource.TestCheckResourceAttr(resourceName, "generate_secret", "true"),
					resource.TestCheckResourceAttr(resourceName, "read_attributes.#", "1"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "read_attributes.*", "email"),
					resource.TestCheckResourceAttr(resourceName, "write_attributes.#", "1"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "write_attributes.*", "email"),
					resource.TestCheckResourceAttr(resourceName, "refresh_token_validity", "300"),
					resource.TestCheckResourceAttr(resourceName, "allowed_oauth_flows.#", "2"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "allowed_oauth_flows.*", "code"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "allowed_oauth_flows.*", "implicit"),
					resource.TestCheckResourceAttr(resourceName, "allowed_oauth_flows_user_pool_client", "true"),
					resource.TestCheckResourceAttr(resourceName, "allowed_oauth_scopes.#", "5"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "allowed_oauth_scopes.*", "openid"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "allowed_oauth_scopes.*", "email"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "allowed_oauth_scopes.*", "phone"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "allowed_oauth_scopes.*", "aws.cognito.signin.user.admin"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "allowed_oauth_scopes.*", "profile"),
					resource.TestCheckResourceAttr(resourceName, "callback_urls.#", "2"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "callback_urls.*", "https://www.example.com/callback"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "callback_urls.*", "https://www.example.com/redirect"),
					resource.TestCheckResourceAttr(resourceName, "default_redirect_uri", "https://www.example.com/redirect"),
					resource.TestCheckResourceAttr(resourceName, "logout_urls.#", "1"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "logout_urls.*", "https://www.example.com/login"),
					resource.TestCheckResourceAttr(resourceName, "prevent_user_existence_errors", "LEGACY"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportStateIdFunc:       testAccAWSCognitoUserPoolClientImportStateIDFunc(resourceName),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"generate_secret"},
			},
		},
	})
}

func TestAccAWSCognitoUserPoolClient_allFieldsUpdatingOneField(t *testing.T) {
	var client cognitoidentityprovider.UserPoolClientType
	userPoolName := fmt.Sprintf("tf-acc-cognito-user-pool-%s", acctest.RandString(7))
	clientName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolClientConfig_allFields(userPoolName, clientName, 300),
			},
			{
				Config: testAccAWSCognitoUserPoolClientConfig_allFields(userPoolName, clientName, 299),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolClientExists(resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "name", clientName),
					resource.TestCheckResourceAttr(resourceName, "explicit_auth_flows.#", "3"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "explicit_auth_flows.*", "CUSTOM_AUTH_FLOW_ONLY"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "explicit_auth_flows.*", "USER_PASSWORD_AUTH"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "explicit_auth_flows.*", "ADMIN_NO_SRP_AUTH"),
					resource.TestCheckResourceAttr(resourceName, "generate_secret", "true"),
					resource.TestCheckResourceAttr(resourceName, "read_attributes.#", "1"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "read_attributes.*", "email"),
					resource.TestCheckResourceAttr(resourceName, "write_attributes.#", "1"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "write_attributes.*", "email"),
					resource.TestCheckResourceAttr(resourceName, "refresh_token_validity", "299"),
					resource.TestCheckResourceAttr(resourceName, "allowed_oauth_flows.#", "2"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "allowed_oauth_flows.*", "code"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "allowed_oauth_flows.*", "implicit"),
					resource.TestCheckResourceAttr(resourceName, "allowed_oauth_flows_user_pool_client", "true"),
					resource.TestCheckResourceAttr(resourceName, "allowed_oauth_scopes.#", "5"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "allowed_oauth_scopes.*", "openid"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "allowed_oauth_scopes.*", "email"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "allowed_oauth_scopes.*", "phone"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "allowed_oauth_scopes.*", "aws.cognito.signin.user.admin"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "allowed_oauth_scopes.*", "profile"),
					resource.TestCheckResourceAttr(resourceName, "callback_urls.#", "2"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "callback_urls.*", "https://www.example.com/callback"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "callback_urls.*", "https://www.example.com/redirect"),
					resource.TestCheckResourceAttr(resourceName, "default_redirect_uri", "https://www.example.com/redirect"),
					resource.TestCheckResourceAttr(resourceName, "logout_urls.#", "1"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "logout_urls.*", "https://www.example.com/login"),
					resource.TestCheckResourceAttr(resourceName, "prevent_user_existence_errors", "LEGACY"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportStateIdFunc:       testAccAWSCognitoUserPoolClientImportStateIDFunc(resourceName),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"generate_secret"},
			},
		},
	})
}

func TestAccAWSCognitoUserPoolClient_analyticsConfig(t *testing.T) {
	var client cognitoidentityprovider.UserPoolClientType
	userPoolName := fmt.Sprintf("tf-acc-cognito-user-pool-%s", acctest.RandString(7))
	clientName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSCognitoIdentityProvider(t)
			testAccPreCheckAWSPinpointApp(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolClientConfigAnalyticsConfig(userPoolName, clientName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolClientExists(resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "analytics_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "analytics_configuration.0.external_id", clientName),
					resource.TestCheckResourceAttr(resourceName, "analytics_configuration.0.user_data_shared", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSCognitoUserPoolClientImportStateIDFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoUserPoolClientConfig_basic(userPoolName, clientName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolClientExists(resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "analytics_configuration.#", "0"),
				),
			},
			{
				Config: testAccAWSCognitoUserPoolClientConfigAnalyticsConfigShareUserData(userPoolName, clientName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolClientExists(resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "analytics_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "analytics_configuration.0.external_id", clientName),
					resource.TestCheckResourceAttr(resourceName, "analytics_configuration.0.user_data_shared", "true"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPoolClient_disappears(t *testing.T) {
	var client cognitoidentityprovider.UserPoolClientType
	userPoolName := fmt.Sprintf("tf-acc-cognito-user-pool-%s", acctest.RandString(7))
	clientName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolClientConfig_basic(userPoolName, clientName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolClientExists(resourceName, &client),
					testAccCheckAWSCognitoUserPoolClientDisappears(&client),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAWSCognitoUserPoolClientImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return "", errors.New("No Cognito User Pool Client ID set")
		}

		conn := testAccProvider.Meta().(*AWSClient).cognitoidpconn
		userPoolId := rs.Primary.Attributes["user_pool_id"]
		clientId := rs.Primary.ID

		params := &cognitoidentityprovider.DescribeUserPoolClientInput{
			UserPoolId: aws.String(userPoolId),
			ClientId:   aws.String(clientId),
		}

		_, err := conn.DescribeUserPoolClient(params)

		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%s/%s", userPoolId, clientId), nil
	}
}

func testAccCheckAWSCognitoUserPoolClientDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cognitoidpconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cognito_user_pool_client" {
			continue
		}

		params := &cognitoidentityprovider.DescribeUserPoolClientInput{
			ClientId:   aws.String(rs.Primary.ID),
			UserPoolId: aws.String(rs.Primary.Attributes["user_pool_id"]),
		}

		_, err := conn.DescribeUserPoolClient(params)

		if err != nil {
			if isAWSErr(err, cognitoidentityprovider.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccCheckAWSCognitoUserPoolClientExists(name string, client *cognitoidentityprovider.UserPoolClientType) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Cognito User Pool Client ID set")
		}

		conn := testAccProvider.Meta().(*AWSClient).cognitoidpconn

		params := &cognitoidentityprovider.DescribeUserPoolClientInput{
			ClientId:   aws.String(rs.Primary.ID),
			UserPoolId: aws.String(rs.Primary.Attributes["user_pool_id"]),
		}

		resp, err := conn.DescribeUserPoolClient(params)
		if err != nil {
			return err
		}

		*client = *resp.UserPoolClient

		return nil
	}
}

func testAccCheckAWSCognitoUserPoolClientDisappears(client *cognitoidentityprovider.UserPoolClientType) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).cognitoidpconn

		params := &cognitoidentityprovider.DeleteUserPoolClientInput{
			ClientId:   client.ClientId,
			UserPoolId: client.UserPoolId,
		}

		_, err := conn.DeleteUserPoolClient(params)

		return err
	}
}

func testAccAWSCognitoUserPoolClientConfig_basic(userPoolName, clientName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = "%s"
}

resource "aws_cognito_user_pool_client" "test" {
  name                = "%s"
  user_pool_id        = aws_cognito_user_pool.test.id
  explicit_auth_flows = ["ADMIN_NO_SRP_AUTH"]
}
`, userPoolName, clientName)
}

func testAccAWSCognitoUserPoolClientConfig_RefreshTokenValidity(rName string, refreshTokenValidity int) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = "%s"
}

resource "aws_cognito_user_pool_client" "test" {
  name                   = "%s"
  refresh_token_validity = %d
  user_pool_id           = aws_cognito_user_pool.test.id
}
`, rName, rName, refreshTokenValidity)
}

func testAccAWSCognitoUserPoolClientConfig_Name(rName, name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_user_pool_client" "test" {
  name         = %[2]q
  user_pool_id = aws_cognito_user_pool.test.id
}
`, rName, name)
}

func testAccAWSCognitoUserPoolClientConfig_allFields(userPoolName, clientName string, refreshTokenValidity int) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = "%s"
}

resource "aws_cognito_user_pool_client" "test" {
  name = "%s"

  user_pool_id        = aws_cognito_user_pool.test.id
  explicit_auth_flows = ["ADMIN_NO_SRP_AUTH", "CUSTOM_AUTH_FLOW_ONLY", "USER_PASSWORD_AUTH"]

  generate_secret = "true"

  read_attributes  = ["email"]
  write_attributes = ["email"]

  refresh_token_validity        = %d
  prevent_user_existence_errors = "LEGACY"

  allowed_oauth_flows                  = ["code", "implicit"]
  allowed_oauth_flows_user_pool_client = "true"
  allowed_oauth_scopes                 = ["phone", "email", "openid", "profile", "aws.cognito.signin.user.admin"]

  callback_urls        = ["https://www.example.com/redirect", "https://www.example.com/callback"]
  default_redirect_uri = "https://www.example.com/redirect"
  logout_urls          = ["https://www.example.com/login"]
}
`, userPoolName, clientName, refreshTokenValidity)
}

func testAccAWSCognitoUserPoolClientConfigAnalyticsConfigBase(userPoolName, clientName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_cognito_user_pool" "test" {
  name = "%[1]s"
}

resource "aws_pinpoint_app" "test" {
  name = "%[2]s"
}

resource "aws_iam_role" "test" {
  name = "%[2]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "cognito-idp.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = "%[2]s"
  role = aws_iam_role.test.id

  policy = <<-EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "mobiletargeting:UpdateEndpoint",
        "mobiletargeting:PutItems"
      ],
      "Effect": "Allow",
      "Resource": "arn:${data.aws_partition.current.partition}:mobiletargeting:*:${data.aws_caller_identity.current.account_id}:apps/${aws_pinpoint_app.test.application_id}*"
    }
  ]
}
EOF
}
`, userPoolName, clientName)
}

func testAccAWSCognitoUserPoolClientConfigAnalyticsConfig(userPoolName, clientName string) string {
	return testAccAWSCognitoUserPoolClientConfigAnalyticsConfigBase(userPoolName, clientName) + fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name         = "%[1]s"
  user_pool_id = aws_cognito_user_pool.test.id

  analytics_configuration {
    application_id = aws_pinpoint_app.test.application_id
    external_id    = "%[1]s"
    role_arn       = aws_iam_role.test.arn
  }
}
`, clientName)
}

func testAccAWSCognitoUserPoolClientConfigAnalyticsConfigShareUserData(userPoolName, clientName string) string {
	return testAccAWSCognitoUserPoolClientConfigAnalyticsConfigBase(userPoolName, clientName) + fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name         = "%[1]s"
  user_pool_id = aws_cognito_user_pool.test.id

  analytics_configuration {
    application_id   = aws_pinpoint_app.test.application_id
    external_id      = "%[1]s"
    role_arn         = aws_iam_role.test.arn
    user_data_shared = true
  }
}
`, clientName)
}
