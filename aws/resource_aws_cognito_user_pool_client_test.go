package aws

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSCognitoUserPoolClient_basic(t *testing.T) {
	userPoolName := fmt.Sprintf("tf-acc-cognito-user-pool-%s", acctest.RandString(7))
	clientName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolClientConfig_basic(userPoolName, clientName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolClientExists("aws_cognito_user_pool_client.client"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "name", clientName),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "explicit_auth_flows.#", "1"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "explicit_auth_flows.245201344", "ADMIN_NO_SRP_AUTH"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPoolClient_importBasic(t *testing.T) {
	userPoolName := fmt.Sprintf("tf-acc-cognito-user-pool-%s", acctest.RandString(7))
	clientName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resourceName := "aws_cognito_user_pool_client.client"

	getStateId := func(s *terraform.State) (string, error) {

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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolClientConfig_basic(userPoolName, clientName),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: getStateId,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCognitoUserPoolClient_RefreshTokenValidity(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolClientConfig_RefreshTokenValidity(rName, 60),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolClientExists("aws_cognito_user_pool_client.client"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "refresh_token_validity", "60"),
				),
			},
			{
				Config: testAccAWSCognitoUserPoolClientConfig_RefreshTokenValidity(rName, 120),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolClientExists("aws_cognito_user_pool_client.client"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "refresh_token_validity", "120"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPoolClient_allFields(t *testing.T) {
	userPoolName := fmt.Sprintf("tf-acc-cognito-user-pool-%s", acctest.RandString(7))
	clientName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolClientConfig_allFields(userPoolName, clientName, 300),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolClientExists("aws_cognito_user_pool_client.client"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "name", clientName),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "explicit_auth_flows.#", "3"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "explicit_auth_flows.1728632605", "CUSTOM_AUTH_FLOW_ONLY"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "explicit_auth_flows.1860959087", "USER_PASSWORD_AUTH"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "explicit_auth_flows.245201344", "ADMIN_NO_SRP_AUTH"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "generate_secret", "true"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "read_attributes.#", "1"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "read_attributes.881205744", "email"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "write_attributes.#", "1"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "write_attributes.881205744", "email"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "refresh_token_validity", "300"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "allowed_oauth_flows.#", "2"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "allowed_oauth_flows.2645166319", "code"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "allowed_oauth_flows.3465961881", "implicit"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "allowed_oauth_flows_user_pool_client", "true"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "allowed_oauth_scopes.#", "5"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "allowed_oauth_scopes.2517049750", "openid"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "allowed_oauth_scopes.881205744", "email"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "allowed_oauth_scopes.2603607895", "phone"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "allowed_oauth_scopes.380129571", "aws.cognito.signin.user.admin"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "allowed_oauth_scopes.4080487570", "profile"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "callback_urls.#", "2"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "callback_urls.0", "https://www.example.com/callback"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "callback_urls.1", "https://www.example.com/redirect"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "default_redirect_uri", "https://www.example.com/redirect"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "logout_urls.#", "1"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "logout_urls.0", "https://www.example.com/login"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPoolClient_allFieldsUpdatingOneField(t *testing.T) {
	userPoolName := fmt.Sprintf("tf-acc-cognito-user-pool-%s", acctest.RandString(7))
	clientName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolClientConfig_allFields(userPoolName, clientName, 300),
			},
			{
				Config: testAccAWSCognitoUserPoolClientConfig_allFields(userPoolName, clientName, 299),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolClientExists("aws_cognito_user_pool_client.client"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "name", clientName),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "explicit_auth_flows.#", "3"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "explicit_auth_flows.1728632605", "CUSTOM_AUTH_FLOW_ONLY"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "explicit_auth_flows.1860959087", "USER_PASSWORD_AUTH"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "explicit_auth_flows.245201344", "ADMIN_NO_SRP_AUTH"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "generate_secret", "true"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "read_attributes.#", "1"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "read_attributes.881205744", "email"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "write_attributes.#", "1"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "write_attributes.881205744", "email"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "refresh_token_validity", "299"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "allowed_oauth_flows.#", "2"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "allowed_oauth_flows.2645166319", "code"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "allowed_oauth_flows.3465961881", "implicit"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "allowed_oauth_flows_user_pool_client", "true"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "allowed_oauth_scopes.#", "5"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "allowed_oauth_scopes.2517049750", "openid"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "allowed_oauth_scopes.881205744", "email"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "allowed_oauth_scopes.2603607895", "phone"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "allowed_oauth_scopes.380129571", "aws.cognito.signin.user.admin"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "allowed_oauth_scopes.4080487570", "profile"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "callback_urls.#", "2"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "callback_urls.0", "https://www.example.com/callback"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "callback_urls.1", "https://www.example.com/redirect"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "default_redirect_uri", "https://www.example.com/redirect"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "logout_urls.#", "1"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_client.client", "logout_urls.0", "https://www.example.com/login"),
				),
			},
		},
	})
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
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ResourceNotFoundException" {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccCheckAWSCognitoUserPoolClientExists(name string) resource.TestCheckFunc {
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

		_, err := conn.DescribeUserPoolClient(params)

		return err
	}
}

func testAccAWSCognitoUserPoolClientConfig_basic(userPoolName, clientName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "pool" {
  name = "%s"
}

resource "aws_cognito_user_pool_client" "client" {
  name                = "%s"
  user_pool_id        = "${aws_cognito_user_pool.pool.id}"
  explicit_auth_flows = ["ADMIN_NO_SRP_AUTH"]
}
`, userPoolName, clientName)
}

func testAccAWSCognitoUserPoolClientConfig_RefreshTokenValidity(rName string, refreshTokenValidity int) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "pool" {
  name = "%s"
}

resource "aws_cognito_user_pool_client" "client" {
  name                   = "%s"
  refresh_token_validity = %d
  user_pool_id           = "${aws_cognito_user_pool.pool.id}"
}
`, rName, rName, refreshTokenValidity)
}

func testAccAWSCognitoUserPoolClientConfig_allFields(userPoolName, clientName string, refreshTokenValidity int) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "pool" {
  name = "%s"
}

resource "aws_cognito_user_pool_client" "client" {
  name = "%s"

  user_pool_id = "${aws_cognito_user_pool.pool.id}"
  explicit_auth_flows = ["ADMIN_NO_SRP_AUTH", "CUSTOM_AUTH_FLOW_ONLY", "USER_PASSWORD_AUTH"]

  generate_secret = "true"

  read_attributes = ["email"]
  write_attributes = ["email"]

  refresh_token_validity = %d

  allowed_oauth_flows = ["code", "implicit"]
  allowed_oauth_flows_user_pool_client = "true"
  allowed_oauth_scopes = ["phone", "email", "openid", "profile", "aws.cognito.signin.user.admin"]

  callback_urls = ["https://www.example.com/callback", "https://www.example.com/redirect"]
  default_redirect_uri = "https://www.example.com/redirect"
  logout_urls = ["https://www.example.com/login"]
}
`, userPoolName, clientName, refreshTokenValidity)
}
