package aws

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSCognitoIdentityPool_basic(t *testing.T) {
	name := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	updatedName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "aws_cognito_identity_pool.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentity(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoIdentityPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoIdentityPoolConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoIdentityPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "cognito-identity", regexp.MustCompile(`identitypool/.+`)),
					resource.TestCheckResourceAttr(resourceName, "allow_unauthenticated_identities", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoIdentityPoolConfig_basic(updatedName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoIdentityPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", updatedName)),
				),
			},
		},
	})
}

func TestAccAWSCognitoIdentityPool_supportedLoginProviders(t *testing.T) {
	name := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "aws_cognito_identity_pool.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentity(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoIdentityPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoIdentityPoolConfig_supportedLoginProviders(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoIdentityPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
					resource.TestCheckResourceAttr(resourceName, "supported_login_providers.graph.facebook.com", "7346241598935555"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoIdentityPoolConfig_supportedLoginProvidersModified(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoIdentityPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
					resource.TestCheckResourceAttr(resourceName, "supported_login_providers.graph.facebook.com", "7346241598935552"),
					resource.TestCheckResourceAttr(resourceName, "supported_login_providers.accounts.google.com", "123456789012.apps.googleusercontent.com"),
				),
			},
			{
				Config: testAccAWSCognitoIdentityPoolConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoIdentityPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
				),
			},
		},
	})
}

func TestAccAWSCognitoIdentityPool_openidConnectProviderArns(t *testing.T) {
	name := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "aws_cognito_identity_pool.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentity(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoIdentityPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoIdentityPoolConfig_openidConnectProviderArns(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoIdentityPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_provider_arns.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoIdentityPoolConfig_openidConnectProviderArnsModified(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoIdentityPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_provider_arns.#", "2"),
				),
			},
			{
				Config: testAccAWSCognitoIdentityPoolConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoIdentityPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
				),
			},
		},
	})
}

func TestAccAWSCognitoIdentityPool_samlProviderArns(t *testing.T) {
	name := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "aws_cognito_identity_pool.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentity(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoIdentityPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoIdentityPoolConfig_samlProviderArns(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoIdentityPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
					resource.TestCheckResourceAttr(resourceName, "saml_provider_arns.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoIdentityPoolConfig_samlProviderArnsModified(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoIdentityPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
					resource.TestCheckResourceAttr(resourceName, "saml_provider_arns.#", "1"),
				),
			},
			{
				Config: testAccAWSCognitoIdentityPoolConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoIdentityPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
					resource.TestCheckResourceAttr(resourceName, "saml_provider_arns.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSCognitoIdentityPool_cognitoIdentityProviders(t *testing.T) {
	name := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "aws_cognito_identity_pool.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentity(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoIdentityPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoIdentityPoolConfig_cognitoIdentityProviders(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoIdentityPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
					resource.TestCheckResourceAttr(resourceName, "cognito_identity_providers.66456389.client_id", "7lhlkkfbfb4q5kpp90urffao"),
					resource.TestCheckResourceAttr(resourceName, "cognito_identity_providers.66456389.provider_name", "cognito-idp.us-east-1.amazonaws.com/us-east-1_Zr231apJu"),
					resource.TestCheckResourceAttr(resourceName, "cognito_identity_providers.66456389.server_side_token_check", "false"),
					resource.TestCheckResourceAttr(resourceName, "cognito_identity_providers.3571192419.client_id", "7lhlkkfbfb4q5kpp90urffao"),
					resource.TestCheckResourceAttr(resourceName, "cognito_identity_providers.3571192419.provider_name", "cognito-idp.us-east-1.amazonaws.com/us-east-1_Ab129faBb"),
					resource.TestCheckResourceAttr(resourceName, "cognito_identity_providers.3571192419.server_side_token_check", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoIdentityPoolConfig_cognitoIdentityProvidersModified(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoIdentityPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
					resource.TestCheckResourceAttr(resourceName, "cognito_identity_providers.3661724441.client_id", "6lhlkkfbfb4q5kpp90urffae"),
					resource.TestCheckResourceAttr(resourceName, "cognito_identity_providers.3661724441.provider_name", "cognito-idp.us-east-1.amazonaws.com/us-east-1_Zr231apJu"),
					resource.TestCheckResourceAttr(resourceName, "cognito_identity_providers.3661724441.server_side_token_check", "false"),
				),
			},
			{
				Config: testAccAWSCognitoIdentityPoolConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoIdentityPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
				),
			},
		},
	})
}

func TestAccAWSCognitoIdentityPool_addingNewProviderKeepsOldProvider(t *testing.T) {
	name := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "aws_cognito_identity_pool.main"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentity(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoIdentityPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoIdentityPoolConfig_cognitoIdentityProviders(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoIdentityPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
					resource.TestCheckResourceAttr(resourceName, "cognito_identity_providers.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoIdentityPoolConfig_cognitoIdentityProvidersAndOpenidConnectProviderArns(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoIdentityPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
					resource.TestCheckResourceAttr(resourceName, "cognito_identity_providers.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_provider_arns.#", "1"),
				),
			},
			{
				Config: testAccAWSCognitoIdentityPoolConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoIdentityPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
					resource.TestCheckResourceAttr(resourceName, "cognito_identity_providers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_provider_arns.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSCognitoIdentityPool_tags(t *testing.T) {
	name := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "aws_cognito_identity_pool.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentity(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoIdentityPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoIdentityPoolConfig_Tags1(name, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoIdentityPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoIdentityPoolConfig_Tags2(name, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoIdentityPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSCognitoIdentityPoolConfig_Tags1(name, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoIdentityPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAWSCognitoIdentityPoolExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Cognito Identity Pool ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).cognitoconn

		_, err := conn.DescribeIdentityPool(&cognitoidentity.DescribeIdentityPoolInput{
			IdentityPoolId: aws.String(rs.Primary.ID),
		})

		return err
	}
}

func testAccCheckAWSCognitoIdentityPoolDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cognitoconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cognito_identity_pool" {
			continue
		}

		_, err := conn.DescribeIdentityPool(&cognitoidentity.DescribeIdentityPoolInput{
			IdentityPoolId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if wserr, ok := err.(awserr.Error); ok && wserr.Code() == cognitoidentity.ErrCodeResourceNotFoundException {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccPreCheckAWSCognitoIdentity(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).cognitoconn

	input := &cognitoidentity.ListIdentityPoolsInput{
		MaxResults: aws.Int64(int64(1)),
	}

	_, err := conn.ListIdentityPools(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSCognitoIdentityPoolConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_identity_pool" "main" {
  identity_pool_name               = "identity pool %s"
  allow_unauthenticated_identities = false
  developer_provider_name          = "my.developer"
}
`, name)
}

func testAccAWSCognitoIdentityPoolConfig_supportedLoginProviders(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_identity_pool" "main" {
  identity_pool_name               = "identity pool %s"
  allow_unauthenticated_identities = false

  supported_login_providers = {
    "graph.facebook.com" = "7346241598935555"
  }
}
`, name)
}

func testAccAWSCognitoIdentityPoolConfig_supportedLoginProvidersModified(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_identity_pool" "main" {
  identity_pool_name               = "identity pool %s"
  allow_unauthenticated_identities = false

  supported_login_providers = {
    "graph.facebook.com"  = "7346241598935552"
    "accounts.google.com" = "123456789012.apps.googleusercontent.com"
  }
}
`, name)
}

func testAccAWSCognitoIdentityPoolConfig_openidConnectProviderArns(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_identity_pool" "main" {
  identity_pool_name               = "identity pool %s"
  allow_unauthenticated_identities = false

  openid_connect_provider_arns = ["arn:aws:iam::123456789012:oidc-provider/server.example.com"]
}
`, name)
}

func testAccAWSCognitoIdentityPoolConfig_openidConnectProviderArnsModified(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_identity_pool" "main" {
  identity_pool_name               = "identity pool %s"
  allow_unauthenticated_identities = false

  openid_connect_provider_arns = ["arn:aws:iam::123456789012:oidc-provider/foo.example.com", "arn:aws:iam::123456789012:oidc-provider/bar.example.com"]
}
`, name)
}

func testAccAWSCognitoIdentityPoolConfig_samlProviderArns(name string) string {
	return fmt.Sprintf(`
resource "aws_iam_saml_provider" "default" {
  name                   = "myprovider-%s"
  saml_metadata_document = "${file("./test-fixtures/saml-metadata.xml")}"
}

resource "aws_cognito_identity_pool" "main" {
  identity_pool_name               = "identity pool %s"
  allow_unauthenticated_identities = false

  saml_provider_arns = ["${aws_iam_saml_provider.default.arn}"]
}
`, name, name)
}

func testAccAWSCognitoIdentityPoolConfig_samlProviderArnsModified(name string) string {
	return fmt.Sprintf(`
resource "aws_iam_saml_provider" "default" {
  name                   = "default-%s"
  saml_metadata_document = "${file("./test-fixtures/saml-metadata.xml")}"
}

resource "aws_iam_saml_provider" "secondary" {
  name                   = "secondary-%s"
  saml_metadata_document = "${file("./test-fixtures/saml-metadata.xml")}"
}

resource "aws_cognito_identity_pool" "main" {
  identity_pool_name               = "identity pool %s"
  allow_unauthenticated_identities = false

  saml_provider_arns = ["${aws_iam_saml_provider.secondary.arn}"]
}
`, name, name, name)
}

func testAccAWSCognitoIdentityPoolConfig_cognitoIdentityProviders(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_identity_pool" "main" {
  identity_pool_name               = "identity pool %s"
  allow_unauthenticated_identities = false

  cognito_identity_providers {
    client_id               = "7lhlkkfbfb4q5kpp90urffao"
    provider_name           = "cognito-idp.us-east-1.amazonaws.com/us-east-1_Ab129faBb"
    server_side_token_check = false
  }

  cognito_identity_providers {
    client_id               = "7lhlkkfbfb4q5kpp90urffao"
    provider_name           = "cognito-idp.us-east-1.amazonaws.com/us-east-1_Zr231apJu"
    server_side_token_check = false
  }
}
`, name)
}

func testAccAWSCognitoIdentityPoolConfig_cognitoIdentityProvidersModified(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_identity_pool" "main" {
  identity_pool_name               = "identity pool %s"
  allow_unauthenticated_identities = false

  cognito_identity_providers {
    client_id               = "6lhlkkfbfb4q5kpp90urffae"
    provider_name           = "cognito-idp.us-east-1.amazonaws.com/us-east-1_Zr231apJu"
    server_side_token_check = false
  }
}
`, name)
}

func testAccAWSCognitoIdentityPoolConfig_cognitoIdentityProvidersAndOpenidConnectProviderArns(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_identity_pool" "main" {
  identity_pool_name               = "identity pool %s"
  allow_unauthenticated_identities = false

  cognito_identity_providers {
    client_id               = "7lhlkkfbfb4q5kpp90urffao"
    provider_name           = "cognito-idp.us-east-1.amazonaws.com/us-east-1_Ab129faBb"
    server_side_token_check = false
  }

  cognito_identity_providers {
    client_id               = "7lhlkkfbfb4q5kpp90urffao"
    provider_name           = "cognito-idp.us-east-1.amazonaws.com/us-east-1_Zr231apJu"
    server_side_token_check = false
  }

  openid_connect_provider_arns = ["arn:aws:iam::123456789012:oidc-provider/server.example.com"]
}
`, name)
}

func testAccAWSCognitoIdentityPoolConfig_Tags1(name, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_cognito_identity_pool" "main" {
  identity_pool_name               = %[1]q
  allow_unauthenticated_identities = false

  tags = {
    %[2]q = %[3]q
  }
}
`, name, tagKey1, tagValue1)
}

func testAccAWSCognitoIdentityPoolConfig_Tags2(name, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_cognito_identity_pool" "main" {
  identity_pool_name               = %[1]q
  allow_unauthenticated_identities = false

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, name, tagKey1, tagValue1, tagKey2, tagValue2)
}
