package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSAppsyncGraphqlApi_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_apikey(acctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists("aws_appsync_graphql_api.test_apikey"),
					resource.TestCheckResourceAttrSet("aws_appsync_graphql_api.test_apikey", "arn"),
				),
			},
		},
	})
}

func TestAccAWSAppsyncGraphqlApi_iam(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_iam(acctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists("aws_appsync_graphql_api.test_iam"),
					resource.TestCheckResourceAttrSet("aws_appsync_graphql_api.test_iam", "arn"),
				),
			},
		},
	})
}

func TestAccAWSAppsyncGraphqlApi_cognito(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncGraphqlApiConfig_cognito(acctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncGraphqlApiExists("aws_appsync_graphql_api.test_cognito"),
					resource.TestCheckResourceAttrSet("aws_appsync_graphql_api.test_cognito", "arn"),
				),
			},
		},
	})
}

func TestAccAWSAppsyncGraphqlApi_import(t *testing.T) {
	resourceName := "aws_appsync_graphql_api.test_apikey"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncGraphqlApiDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAppsyncGraphqlApiConfig_apikey(acctest.RandString(5)),
			},
			resource.TestStep{
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
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccAppsyncGraphqlApiConfig_apikey(rName string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test_apikey" {
  authentication_type = "API_KEY"
  name = "tf_appsync_%s"
}
`, rName)
}

func testAccAppsyncGraphqlApiConfig_iam(rName string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test_iam" {
  authentication_type = "AWS_IAM"
  name = "tf_appsync_%s"
}
`, rName)
}

func testAccAppsyncGraphqlApiConfig_cognito(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "test" {
  current = true
}

resource "aws_cognito_user_pool" "test" {
  name = "tf-%s"
}

resource "aws_appsync_graphql_api" "test_cognito" {
  authentication_type = "AMAZON_COGNITO_USER_POOLS"
  name = "tf_appsync_%s"
  user_pool_config {
    aws_region = "${data.aws_region.test.name}"
    default_action = "ALLOW"
    user_pool_id = "${aws_cognito_user_pool.test.id}"
  }
}
`, rName, rName)
}
