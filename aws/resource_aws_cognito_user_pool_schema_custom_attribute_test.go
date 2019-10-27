package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"testing"
)

func TestAccAWSCognitoUserPoolAddCustomAttribute_basic(t *testing.T) {
	userPoolName := fmt.Sprintf("tf-acc-cognito-user-pool-%s", acctest.RandString(7))
	attributeName := "test_attribute"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolAttributeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolAddCustomAttributeConfig_basic(userPoolName, attributeName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("aws_cognito_user_pool_schema_custom_attributes.custom_attribute_1", "schema.#", "1"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_schema_custom_attributes.custom_attribute_1", "schema.3072888811.attribute_data_type", "String"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_schema_custom_attributes.custom_attribute_1", "schema.3072888811.name", attributeName),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_schema_custom_attributes.custom_attribute_1", "schema.3072888811.mutable", "true"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_schema_custom_attributes.custom_attribute_1", "schema.3072888811.required", "false"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_schema_custom_attributes.custom_attribute_1", "schema.3072888811.developer_only_attribute", "false"),
				),
			},
		},
	})
}

func testAccAWSCognitoUserPoolAddCustomAttributeConfig_basic(userPoolName, attributeName string) string {
	return fmt.Sprintf(`
	resource "aws_cognito_user_pool" "pool" {
		name = "%s"
	}
	
	resource "aws_cognito_user_pool_schema_custom_attributes" "custom_attribute_1" {
		user_pool_id        = "${aws_cognito_user_pool.pool.id}"
		schema {
			attribute_data_type      = "String"
			developer_only_attribute = false
			mutable                  = true
			name                     = "%s"
			required                 = false
		  }
	}
	`, userPoolName, attributeName)
}

// If user pool is destroyed the custom attributes will also get automatically destroyed
func testAccCheckAWSCognitoUserPoolAttributeDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cognitoidpconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cognito_user_pool" {
			continue
		}

		params := &cognitoidentityprovider.DescribeUserPoolInput{
			UserPoolId: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeUserPool(params)

		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ResourceNotFoundException" {
				return nil
			}
			return err
		}
	}

	return nil
}
