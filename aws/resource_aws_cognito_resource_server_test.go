package aws

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSCognitoResourceServer_basic(t *testing.T) {
	var resourceServer cognitoidentityprovider.ResourceServerType
	identifier := fmt.Sprintf("tf-acc-test-resource-server-id-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	name1 := fmt.Sprintf("tf-acc-test-resource-server-name-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	name2 := fmt.Sprintf("tf-acc-test-resource-server-name-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	poolName := fmt.Sprintf("tf-acc-test-pool-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	resourceName := "aws_cognito_resource_server.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoResourceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoResourceServerConfig_basic(identifier, name1, poolName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoResourceServerExists(resourceName, &resourceServer),
					resource.TestCheckResourceAttr(resourceName, "identifier", identifier),
					resource.TestCheckResourceAttr(resourceName, "name", name1),
					resource.TestCheckResourceAttr(resourceName, "scope.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope_identifiers.#", "0"),
				),
			},
			{
				Config: testAccAWSCognitoResourceServerConfig_basic(identifier, name2, poolName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoResourceServerExists(resourceName, &resourceServer),
					resource.TestCheckResourceAttr(resourceName, "identifier", identifier),
					resource.TestCheckResourceAttr(resourceName, "name", name2),
					resource.TestCheckResourceAttr(resourceName, "scope.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope_identifiers.#", "0"),
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

func TestAccAWSCognitoResourceServer_scope(t *testing.T) {
	var resourceServer cognitoidentityprovider.ResourceServerType
	identifier := fmt.Sprintf("tf-acc-test-resource-server-id-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	name := fmt.Sprintf("tf-acc-test-resource-server-name-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	poolName := fmt.Sprintf("tf-acc-test-pool-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	resourceName := "aws_cognito_resource_server.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoResourceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoResourceServerConfig_scope(identifier, name, poolName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoResourceServerExists(resourceName, &resourceServer),
					resource.TestCheckResourceAttr(resourceName, "scope.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "scope_identifiers.#", "2"),
				),
			},
			{
				Config: testAccAWSCognitoResourceServerConfig_scope_update(identifier, name, poolName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoResourceServerExists(resourceName, &resourceServer),
					resource.TestCheckResourceAttr(resourceName, "scope.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scope_identifiers.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Ensure we can remove scope completely
			{
				Config: testAccAWSCognitoResourceServerConfig_basic(identifier, name, poolName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoResourceServerExists(resourceName, &resourceServer),
					resource.TestCheckResourceAttr(resourceName, "scope.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope_identifiers.#", "0"),
				),
			},
		},
	})
}

func testAccCheckAWSCognitoResourceServerExists(n string, resourceServer *cognitoidentityprovider.ResourceServerType) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Cognito Resource Server ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).cognitoidpconn

		userPoolID, identifier, err := decodeCognitoResourceServerID(rs.Primary.ID)
		if err != nil {
			return err
		}

		output, err := conn.DescribeResourceServer(&cognitoidentityprovider.DescribeResourceServerInput{
			Identifier: aws.String(identifier),
			UserPoolId: aws.String(userPoolID),
		})

		if err != nil {
			return err
		}

		if output == nil || output.ResourceServer == nil {
			return fmt.Errorf("Cognito Resource Server %q information not found", rs.Primary.ID)
		}

		*resourceServer = *output.ResourceServer

		return nil
	}
}

func testAccCheckAWSCognitoResourceServerDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cognitoidpconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cognito_resource_server" {
			continue
		}

		userPoolID, identifier, err := decodeCognitoResourceServerID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = conn.DescribeResourceServer(&cognitoidentityprovider.DescribeResourceServerInput{
			Identifier: aws.String(identifier),
			UserPoolId: aws.String(userPoolID),
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

func testAccAWSCognitoResourceServerConfig_basic(identifier string, name string, poolName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_resource_server" "main" {
  identifier = "%s"
  name = "%s"
  user_pool_id = "${aws_cognito_user_pool.main.id}"
}

resource "aws_cognito_user_pool" "main" {
  name = "%s"
}
`, identifier, name, poolName)
}

func testAccAWSCognitoResourceServerConfig_scope(identifier string, name string, poolName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_resource_server" "main" {
  identifier = "%s"
  name = "%s"

  scope {
    scope_name = "scope_1_name"
    scope_description = "scope_1_description"
  }

  scope {
    scope_name = "scope_2_name"
    scope_description = "scope_2_description"
  }

  user_pool_id = "${aws_cognito_user_pool.main.id}"
}

resource "aws_cognito_user_pool" "main" {
  name = "%s"
}
`, identifier, name, poolName)
}

func testAccAWSCognitoResourceServerConfig_scope_update(identifier string, name string, poolName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_resource_server" "main" {
  identifier = "%s"
  name = "%s"

  scope {
    scope_name = "scope_1_name_updated"
    scope_description = "scope_1_description"
  }

  user_pool_id = "${aws_cognito_user_pool.main.id}"
}

resource "aws_cognito_user_pool" "main" {
  name = "%s"
}
`, identifier, name, poolName)
}
