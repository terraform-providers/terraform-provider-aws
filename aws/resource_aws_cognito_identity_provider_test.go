package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSCognitoIdentityProvider_basic(t *testing.T) {
	var identityProvider cognitoidentityprovider.IdentityProviderType
	resourceName := "aws_cognito_identity_provider.test"
	userPoolName := fmt.Sprintf("tf-acc-cognito-user-pool-%s", acctest.RandString(7))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoIdentityProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoIdentityProviderConfig_basic(userPoolName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoIdentityProviderExists(resourceName, &identityProvider),
					resource.TestCheckResourceAttr(resourceName, "attribute_mapping.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "attribute_mapping.username", "sub"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.%", "9"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.authorize_scopes", "email"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.authorize_url", "https://accounts.google.com/o/oauth2/v2/auth"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.client_id", "test-url.apps.googleusercontent.com"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.client_secret", "client_secret"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.attributes_url", "https://people.googleapis.com/v1/people/me?personFields="),
					resource.TestCheckResourceAttr(resourceName, "provider_details.attributes_url_add_attributes", "true"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.token_request_method", "POST"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.token_url", "https://www.googleapis.com/oauth2/v4/token"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.oidc_issuer", "https://accounts.google.com"),
					resource.TestCheckResourceAttr(resourceName, "provider_name", "Google"),
					resource.TestCheckResourceAttr(resourceName, "provider_type", "Google"),
				),
			},
			{
				Config: testAccAWSCognitoIdentityProviderConfig_basicUpdated(userPoolName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoIdentityProviderExists(resourceName, &identityProvider),
					resource.TestCheckResourceAttr(resourceName, "attribute_mapping.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "attribute_mapping.username", "sub"),
					resource.TestCheckResourceAttr(resourceName, "attribute_mapping.email", "email"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.%", "9"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.authorize_scopes", "email"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.authorize_url", "https://accounts.google.com/o/oauth2/v2/auth"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.client_id", "new-client-id-url.apps.googleusercontent.com"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.client_secret", "updated_client_secret"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.attributes_url", "https://people.googleapis.com/v1/people/me?personFields="),
					resource.TestCheckResourceAttr(resourceName, "provider_details.attributes_url_add_attributes", "true"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.token_request_method", "POST"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.token_url", "https://www.googleapis.com/oauth2/v4/token"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.oidc_issuer", "https://accounts.google.com"),
					resource.TestCheckResourceAttr(resourceName, "provider_name", "Google"),
					resource.TestCheckResourceAttr(resourceName, "provider_type", "Google"),
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

func testAccCheckAWSCognitoIdentityProviderDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cognitoidpconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cognito_identity_provider" {
			continue
		}

		userPoolID, providerName, err := decodeCognitoIdentityProviderID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = conn.DescribeIdentityProvider(&cognitoidentityprovider.DescribeIdentityProviderInput{
			ProviderName: aws.String(providerName),
			UserPoolId:   aws.String(userPoolID),
		})

		if err != nil {
			if isAWSErr(err, cognitoidentityprovider.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccCheckAWSCognitoIdentityProviderExists(resourceName string, identityProvider *cognitoidentityprovider.IdentityProviderType) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		userPoolID, providerName, err := decodeCognitoIdentityProviderID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).cognitoidpconn

		input := &cognitoidentityprovider.DescribeIdentityProviderInput{
			ProviderName: aws.String(providerName),
			UserPoolId:   aws.String(userPoolID),
		}

		output, err := conn.DescribeIdentityProvider(input)

		if err != nil {
			return err
		}

		if output == nil || output.IdentityProvider == nil {
			return fmt.Errorf("Cognito Identity Provider %q does not exist", rs.Primary.ID)
		}

		*identityProvider = *output.IdentityProvider

		return nil
	}
}

func testAccAWSCognitoIdentityProviderConfig_basic(userPoolName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name                     = "%s"
  auto_verified_attributes = ["email"]
}

resource "aws_cognito_identity_provider" "test" {
  user_pool_id  = aws_cognito_user_pool.test.id
  provider_name = "Google"
  provider_type = "Google"

  provider_details = {
    attributes_url                = "https://people.googleapis.com/v1/people/me?personFields="
    attributes_url_add_attributes = "true"
    authorize_scopes              = "email"
    authorize_url                 = "https://accounts.google.com/o/oauth2/v2/auth"
    client_id                     = "test-url.apps.googleusercontent.com"
    client_secret                 = "client_secret"
    oidc_issuer                   = "https://accounts.google.com"
    token_request_method          = "POST"
    token_url                     = "https://www.googleapis.com/oauth2/v4/token"
  }
}
`, userPoolName)
}

func testAccAWSCognitoIdentityProviderConfig_basicUpdated(userPoolName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name                     = "%s"
  auto_verified_attributes = ["email"]
}

resource "aws_cognito_identity_provider" "test" {
  user_pool_id  = aws_cognito_user_pool.test.id
  provider_name = "Google"
  provider_type = "Google"

  provider_details = {
    attributes_url                = "https://people.googleapis.com/v1/people/me?personFields="
    attributes_url_add_attributes = "true"
    authorize_scopes              = "email"
    authorize_url                 = "https://accounts.google.com/o/oauth2/v2/auth"
    client_id                     = "new-client-id-url.apps.googleusercontent.com"
    client_secret                 = "updated_client_secret"
    oidc_issuer                   = "https://accounts.google.com"
    token_request_method          = "POST"
    token_url                     = "https://www.googleapis.com/oauth2/v4/token"
  }

  attribute_mapping = {
    email    = "email"
    username = "sub"
  }
}
`, userPoolName)
}
