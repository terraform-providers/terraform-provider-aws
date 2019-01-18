package aws

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_cognito_user_pool", &resource.Sweeper{
		Name: "aws_cognito_user_pool",
		F:    testSweepCognitoUserPools,
	})
}

func testSweepCognitoUserPools(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*AWSClient).cognitoidpconn

	input := &cognitoidentityprovider.ListUserPoolsInput{
		MaxResults: aws.Int64(int64(50)),
	}

	for {
		output, err := conn.ListUserPools(input)
		if err != nil {
			if testSweepSkipSweepError(err) {
				log.Printf("[WARN] Skipping Cognito User Pool sweep for %s: %s", region, err)
				return nil
			}
			return fmt.Errorf("Error retrieving Cognito User Pools: %s", err)
		}

		if len(output.UserPools) == 0 {
			log.Print("[DEBUG] No Cognito User Pools to sweep")
			return nil
		}

		for _, userPool := range output.UserPools {
			name := aws.StringValue(userPool.Name)

			if !strings.HasPrefix(name, "tf_acc_") {
				continue
			}

			log.Printf("[INFO] Deleting Cognito User Pool %s", name)
			_, err := conn.DeleteUserPool(&cognitoidentityprovider.DeleteUserPoolInput{
				UserPoolId: userPool.Id,
			})
			if err != nil {
				return fmt.Errorf("Error deleting Cognito User Pool %s: %s", name, err)
			}
		}

		if output.NextToken == nil {
			break
		}
		input.NextToken = output.NextToken
	}

	return nil
}

func TestAccAWSCognitoUserPool_importBasic(t *testing.T) {
	resourceName := "aws_cognito_user_pool.pool"
	name := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchDashboardDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_basic(name),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCognitoUserPool_basic(t *testing.T) {
	name := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists("aws_cognito_user_pool.pool"),
					resource.TestMatchResourceAttr("aws_cognito_user_pool.pool", "arn",
						regexp.MustCompile(`^arn:aws:cognito-idp:[^:]+:[0-9]{12}:userpool/[\w-]+_[0-9a-zA-Z]+$`)),
					resource.TestMatchResourceAttr("aws_cognito_user_pool.pool", "endpoint",
						regexp.MustCompile(`^cognito-idp\.[^.]+\.amazonaws.com/[\w-]+_[0-9a-zA-Z]+$`)),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "name", "terraform-test-pool-"+name),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool.pool", "creation_date"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool.pool", "last_modified_date"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPool_withAdminCreateUserConfiguration(t *testing.T) {
	name := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_withAdminCreateUserConfiguration(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists("aws_cognito_user_pool.pool"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "admin_create_user_config.0.unused_account_validity_days", "6"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "admin_create_user_config.0.allow_admin_create_user_only", "true"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "admin_create_user_config.0.invite_message_template.0.email_message", "Your username is {username} and temporary password is {####}. "),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "admin_create_user_config.0.invite_message_template.0.email_subject", "FooBar {####}"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "admin_create_user_config.0.invite_message_template.0.sms_message", "Your username is {username} and temporary password is {####}."),
				),
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_withAdminCreateUserConfigurationUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "admin_create_user_config.0.unused_account_validity_days", "7"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "admin_create_user_config.0.allow_admin_create_user_only", "false"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "admin_create_user_config.0.invite_message_template.0.email_message", "Your username is {username} and constant password is {####}. "),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "admin_create_user_config.0.invite_message_template.0.email_subject", "Foo{####}BaBaz"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "admin_create_user_config.0.invite_message_template.0.sms_message", "Your username is {username} and constant password is {####}."),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPool_withDeviceConfiguration(t *testing.T) {
	name := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_withDeviceConfiguration(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists("aws_cognito_user_pool.pool"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "device_configuration.0.challenge_required_on_new_device", "true"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "device_configuration.0.device_only_remembered_on_user_prompt", "false"),
				),
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_withDeviceConfigurationUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "device_configuration.0.challenge_required_on_new_device", "false"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "device_configuration.0.device_only_remembered_on_user_prompt", "true"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPool_withEmailVerificationMessage(t *testing.T) {
	name := acctest.RandString(5)
	subject := acctest.RandString(10)
	updatedSubject := acctest.RandString(10)
	message := fmt.Sprintf("%s {####}", acctest.RandString(10))
	upatedMessage := fmt.Sprintf("%s {####}", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_withEmailVerificationMessage(name, subject, message),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists("aws_cognito_user_pool.pool"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "email_verification_subject", subject),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "email_verification_message", message),
				),
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_withEmailVerificationMessage(name, updatedSubject, upatedMessage),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "email_verification_subject", updatedSubject),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "email_verification_message", upatedMessage),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPool_withSmsVerificationMessage(t *testing.T) {
	name := acctest.RandString(5)
	authenticationMessage := fmt.Sprintf("%s {####}", acctest.RandString(10))
	updatedAuthenticationMessage := fmt.Sprintf("%s {####}", acctest.RandString(10))
	verificationMessage := fmt.Sprintf("%s {####}", acctest.RandString(10))
	upatedVerificationMessage := fmt.Sprintf("%s {####}", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_withSmsVerificationMessage(name, authenticationMessage, verificationMessage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists("aws_cognito_user_pool.pool"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "sms_authentication_message", authenticationMessage),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "sms_verification_message", verificationMessage),
				),
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_withSmsVerificationMessage(name, updatedAuthenticationMessage, upatedVerificationMessage),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "sms_authentication_message", updatedAuthenticationMessage),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "sms_verification_message", upatedVerificationMessage),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPool_withEmailConfiguration(t *testing.T) {
	name := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists("aws_cognito_user_pool.pool"),
				),
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_withEmailConfiguration(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "email_configuration.#", "1"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "email_configuration.0.reply_to_email_address", "foo.bar@baz"),
				),
			},
		},
	})
}

// Ensure we can create a User Pool, handling IAM role propagation,
// taking some time.
func TestAccAWSCognitoUserPool_withSmsConfiguration(t *testing.T) {
	name := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_withSmsConfiguration(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "sms_configuration.#", "1"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool.pool", "sms_configuration.0.external_id"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool.pool", "sms_configuration.0.sns_caller_arn"),
				),
			},
		},
	})
}

// Ensure we can update a User Pool, handling IAM role propagation.
func TestAccAWSCognitoUserPool_withSmsConfigurationUpdated(t *testing.T) {
	name := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists("aws_cognito_user_pool.pool"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "sms_configuration.#", "0"),
				),
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_withSmsConfiguration(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "sms_configuration.#", "1"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool.pool", "sms_configuration.0.external_id"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool.pool", "sms_configuration.0.sns_caller_arn"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPool_withTags(t *testing.T) {
	name := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_withTags(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists("aws_cognito_user_pool.pool"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "tags.Name", "Foo"),
				),
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_withTagsUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "tags.Name", "FooBar"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "tags.Project", "Terraform"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPool_withAliasAttributes(t *testing.T) {
	name := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_withAliasAttributes(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists("aws_cognito_user_pool.pool"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "alias_attributes.#", "1"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "alias_attributes.1888159429", "preferred_username"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "auto_verified_attributes.#", "0"),
				),
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_withAliasAttributesUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "alias_attributes.#", "2"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "alias_attributes.881205744", "email"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "alias_attributes.1888159429", "preferred_username"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "auto_verified_attributes.#", "1"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "auto_verified_attributes.881205744", "email"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPool_withPasswordPolicy(t *testing.T) {
	name := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_withPasswordPolicy(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists("aws_cognito_user_pool.pool"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "password_policy.#", "1"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "password_policy.0.minimum_length", "7"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "password_policy.0.require_lowercase", "true"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "password_policy.0.require_numbers", "false"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "password_policy.0.require_symbols", "true"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "password_policy.0.require_uppercase", "false"),
				),
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_withPasswordPolicyUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "password_policy.#", "1"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "password_policy.0.minimum_length", "9"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "password_policy.0.require_lowercase", "false"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "password_policy.0.require_numbers", "true"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "password_policy.0.require_symbols", "false"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "password_policy.0.require_uppercase", "true"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPool_withLambdaConfig(t *testing.T) {
	name := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_withLambdaConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists("aws_cognito_user_pool.main"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "lambda_config.#", "1"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool.main", "lambda_config.0.create_auth_challenge"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool.main", "lambda_config.0.custom_message"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool.main", "lambda_config.0.define_auth_challenge"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool.main", "lambda_config.0.post_authentication"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool.main", "lambda_config.0.post_confirmation"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool.main", "lambda_config.0.pre_authentication"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool.main", "lambda_config.0.pre_sign_up"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool.main", "lambda_config.0.pre_token_generation"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool.main", "lambda_config.0.user_migration"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool.main", "lambda_config.0.verify_auth_challenge_response"),
				),
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_withLambdaConfigUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "lambda_config.#", "1"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool.main", "lambda_config.0.create_auth_challenge"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool.main", "lambda_config.0.custom_message"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool.main", "lambda_config.0.define_auth_challenge"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool.main", "lambda_config.0.post_authentication"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool.main", "lambda_config.0.post_confirmation"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool.main", "lambda_config.0.pre_authentication"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool.main", "lambda_config.0.pre_sign_up"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool.main", "lambda_config.0.pre_token_generation"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool.main", "lambda_config.0.user_migration"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool.main", "lambda_config.0.verify_auth_challenge_response"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPool_withSchemaAttributes(t *testing.T) {
	name := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_withSchemaAttributes(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists("aws_cognito_user_pool.main"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.#", "2"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.145451252.attribute_data_type", "String"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.145451252.developer_only_attribute", "false"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.145451252.mutable", "false"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.145451252.name", "email"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.145451252.number_attribute_constraints.#", "0"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.145451252.required", "true"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.145451252.string_attribute_constraints.#", "1"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.145451252.string_attribute_constraints.0.min_length", "5"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.145451252.string_attribute_constraints.0.max_length", "10"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.770828826.attribute_data_type", "Boolean"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.770828826.developer_only_attribute", "true"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.770828826.mutable", "false"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.770828826.name", "mybool"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.770828826.number_attribute_constraints.#", "0"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.770828826.required", "false"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.770828826.string_attribute_constraints.#", "0"),
				),
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_withSchemaAttributesUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.#", "3"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.2078884933.attribute_data_type", "String"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.2078884933.developer_only_attribute", "false"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.2078884933.mutable", "false"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.2078884933.name", "email"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.2078884933.number_attribute_constraints.#", "0"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.2078884933.required", "true"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.2078884933.string_attribute_constraints.#", "1"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.2078884933.string_attribute_constraints.0.min_length", "7"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.2078884933.string_attribute_constraints.0.max_length", "15"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.2718111653.attribute_data_type", "Number"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.2718111653.developer_only_attribute", "true"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.2718111653.mutable", "true"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.2718111653.name", "mynumber"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.2718111653.number_attribute_constraints.#", "1"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.2718111653.number_attribute_constraints.0.min_value", "2"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.2718111653.number_attribute_constraints.0.max_value", "6"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.2718111653.required", "false"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.2718111653.string_attribute_constraints.#", "0"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.2753746449.attribute_data_type", "Number"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.2753746449.developer_only_attribute", "false"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.2753746449.mutable", "true"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.2753746449.name", "mynondevnumber"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.2753746449.number_attribute_constraints.#", "1"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.2753746449.number_attribute_constraints.0.min_value", "2"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.2753746449.number_attribute_constraints.0.max_value", "6"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.2753746449.required", "false"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "schema.2753746449.string_attribute_constraints.#", "0"),
				),
			},
			{
				ResourceName:      "aws_cognito_user_pool.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCognitoUserPool_withVerificationMessageTemplate(t *testing.T) {
	name := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_withVerificationMessageTemplate(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists("aws_cognito_user_pool.pool"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "verification_message_template.0.default_email_option", "CONFIRM_WITH_LINK"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "verification_message_template.0.email_message", "Foo {####} Bar"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "verification_message_template.0.email_message_by_link", "{##foobar##}"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "verification_message_template.0.email_subject", "FooBar {####}"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "verification_message_template.0.email_subject_by_link", "foobar"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "verification_message_template.0.sms_message", "{####} Baz"),
				),
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_withVerificationMessageTemplateUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "verification_message_template.0.default_email_option", "CONFIRM_WITH_CODE"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPool_update(t *testing.T) {
	name := acctest.RandString(5)
	optionalMfa := "OPTIONAL"
	offMfa := "OFF"
	authenticationMessage := fmt.Sprintf("%s {####}", acctest.RandString(10))
	updatedAuthenticationMessage := fmt.Sprintf("%s {####}", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_update(name, optionalMfa, authenticationMessage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists("aws_cognito_user_pool.pool"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "auto_verified_attributes.#", "1"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "mfa_configuration", optionalMfa),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "email_verification_message", "Foo {####} Bar"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "email_verification_subject", "FooBar {####}"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "sms_verification_message", "{####} Baz"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "sms_authentication_message", authenticationMessage),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "admin_create_user_config.0.unused_account_validity_days", "6"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "admin_create_user_config.0.allow_admin_create_user_only", "true"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "admin_create_user_config.0.invite_message_template.0.email_message", "Your username is {username} and temporary password is {####}. "),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "admin_create_user_config.0.invite_message_template.0.email_subject", "FooBar {####}"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "admin_create_user_config.0.invite_message_template.0.sms_message", "Your username is {username} and temporary password is {####}."),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "device_configuration.0.challenge_required_on_new_device", "true"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "device_configuration.0.device_only_remembered_on_user_prompt", "false"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "verification_message_template.0.default_email_option", "CONFIRM_WITH_CODE"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "sms_configuration.#", "1"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool.pool", "sms_configuration.0.external_id"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool.pool", "sms_configuration.0.sns_caller_arn"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "tags.Name", "Foo"),
				),
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_update(name, optionalMfa, updatedAuthenticationMessage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists("aws_cognito_user_pool.pool"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "auto_verified_attributes.#", "1"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "mfa_configuration", optionalMfa),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "email_verification_message", "Foo {####} Bar"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "email_verification_subject", "FooBar {####}"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "sms_verification_message", "{####} Baz"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "sms_authentication_message", updatedAuthenticationMessage),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "admin_create_user_config.0.unused_account_validity_days", "6"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "admin_create_user_config.0.allow_admin_create_user_only", "true"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "admin_create_user_config.0.invite_message_template.0.email_message", "Your username is {username} and temporary password is {####}. "),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "admin_create_user_config.0.invite_message_template.0.email_subject", "FooBar {####}"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "admin_create_user_config.0.invite_message_template.0.sms_message", "Your username is {username} and temporary password is {####}."),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "device_configuration.0.challenge_required_on_new_device", "true"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "device_configuration.0.device_only_remembered_on_user_prompt", "false"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "verification_message_template.0.default_email_option", "CONFIRM_WITH_CODE"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "sms_configuration.#", "1"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool.pool", "sms_configuration.0.external_id"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool.pool", "sms_configuration.0.sns_caller_arn"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "tags.Name", "Foo"),
				),
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_update(name, offMfa, updatedAuthenticationMessage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists("aws_cognito_user_pool.pool"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "auto_verified_attributes.#", "1"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "mfa_configuration", offMfa),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "email_verification_message", "Foo {####} Bar"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "email_verification_subject", "FooBar {####}"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "sms_verification_message", "{####} Baz"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "sms_authentication_message", updatedAuthenticationMessage),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "admin_create_user_config.0.unused_account_validity_days", "6"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "admin_create_user_config.0.allow_admin_create_user_only", "true"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "admin_create_user_config.0.invite_message_template.0.email_message", "Your username is {username} and temporary password is {####}. "),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "admin_create_user_config.0.invite_message_template.0.email_subject", "FooBar {####}"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "admin_create_user_config.0.invite_message_template.0.sms_message", "Your username is {username} and temporary password is {####}."),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "device_configuration.0.challenge_required_on_new_device", "true"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "device_configuration.0.device_only_remembered_on_user_prompt", "false"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "verification_message_template.0.default_email_option", "CONFIRM_WITH_CODE"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "sms_configuration.#", "1"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool.pool", "sms_configuration.0.external_id"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool.pool", "sms_configuration.0.sns_caller_arn"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "tags.Name", "Foo"),
				),
			},
		},
	})
}

func testAccCheckAWSCognitoUserPoolDestroy(s *terraform.State) error {
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

func testAccCheckAWSCognitoUserPoolExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Cognito User Pool ID set")
		}

		conn := testAccProvider.Meta().(*AWSClient).cognitoidpconn

		params := &cognitoidentityprovider.DescribeUserPoolInput{
			UserPoolId: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeUserPool(params)

		return err
	}
}

func testAccAWSCognitoUserPoolConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "pool" {
  name = "terraform-test-pool-%s"
}`, name)
}

func testAccAWSCognitoUserPoolConfig_withAdminCreateUserConfiguration(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "pool" {
  name = "terraform-test-pool-%s"

  admin_create_user_config {
    allow_admin_create_user_only = true
    unused_account_validity_days = 6

    invite_message_template {
      email_message = "Your username is {username} and temporary password is {####}. "
      email_subject = "FooBar {####}"
      sms_message   = "Your username is {username} and temporary password is {####}."
    }
  }
}`, name)
}

func testAccAWSCognitoUserPoolConfig_withAdminCreateUserConfigurationUpdated(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "pool" {
  name = "terraform-test-pool-%s"

  admin_create_user_config {
    allow_admin_create_user_only = false
    unused_account_validity_days = 7

    invite_message_template {
      email_message = "Your username is {username} and constant password is {####}. "
      email_subject = "Foo{####}BaBaz"
      sms_message   = "Your username is {username} and constant password is {####}."
    }
  }
}`, name)
}

func testAccAWSCognitoUserPoolConfig_withDeviceConfiguration(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "pool" {
  name = "terraform-test-pool-%s"

  device_configuration {
    challenge_required_on_new_device      = true
    device_only_remembered_on_user_prompt = false
  }
}`, name)
}

func testAccAWSCognitoUserPoolConfig_withDeviceConfigurationUpdated(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "pool" {
  name = "terraform-test-pool-%s"

  device_configuration {
    challenge_required_on_new_device      = false
    device_only_remembered_on_user_prompt = true
  }
}`, name)
}

func testAccAWSCognitoUserPoolConfig_withEmailVerificationMessage(name, subject, message string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "pool" {
  name = "terraform-test-pool-%s"
  email_verification_subject = "%s"
  email_verification_message = "%s"

  verification_message_template {
    default_email_option  = "CONFIRM_WITH_CODE"
  }
}`, name, subject, message)
}

func testAccAWSCognitoUserPoolConfig_withSmsVerificationMessage(name, authenticationMessage, verificationMessage string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "pool" {
  name = "terraform-test-pool-%s"
  sms_authentication_message = "%s"
  sms_verification_message   = "%s"
}`, name, authenticationMessage, verificationMessage)
}

func testAccAWSCognitoUserPoolConfig_withTags(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "pool" {
  name = "terraform-test-pool-%s"

  tags = {
    "Name" = "Foo"
  }
}`, name)
}

func testAccAWSCognitoUserPoolConfig_withEmailConfiguration(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "pool" {
  name = "terraform-test-pool-%s"


  email_configuration {
    reply_to_email_address = "foo.bar@baz"
  }
}`, name)
}

func testAccAWSCognitoUserPoolConfig_withSmsConfiguration(name string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_iam_role" "main" {
  name = "test-role-%[1]s"
  path = "/service-role/"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "cognito-idp.amazonaws.com"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "${data.aws_caller_identity.current.account_id}"
        }
      }
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "main" {
  name = "test-role-policy-%[1]s"
  role = "${aws_iam_role.main.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "sns:publish"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}

resource "aws_cognito_user_pool" "pool" {
  name = "terraform-test-pool-%[1]s"

  sms_configuration {
    external_id    = "${data.aws_caller_identity.current.account_id}"
    sns_caller_arn = "${aws_iam_role.main.arn}"
  }
}`, name)
}

func testAccAWSCognitoUserPoolConfig_withTagsUpdated(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "pool" {
  name = "terraform-test-pool-%s"

  tags = {
    "Name"    = "FooBar"
    "Project" = "Terraform"
  }
}`, name)
}

func testAccAWSCognitoUserPoolConfig_withAliasAttributes(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "pool" {
  name = "terraform-test-pool-%s"

  alias_attributes = ["preferred_username"]
}`, name)
}

func testAccAWSCognitoUserPoolConfig_withAliasAttributesUpdated(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "pool" {
  name = "terraform-test-pool-%s"

  alias_attributes         = ["email", "preferred_username"]
  auto_verified_attributes = ["email"]
}`, name)
}

func testAccAWSCognitoUserPoolConfig_withPasswordPolicy(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "pool" {
  name = "terraform-test-pool-%s"

  password_policy {
    minimum_length    = 7
    require_lowercase = true
    require_numbers   = false
    require_symbols   = true
    require_uppercase = false
  }
}`, name)
}

func testAccAWSCognitoUserPoolConfig_withPasswordPolicyUpdated(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "pool" {
  name = "terraform-test-pool-%s"

  password_policy {
    minimum_length    = 9
    require_lowercase = false
    require_numbers   = true
    require_symbols   = false
    require_uppercase = true
  }
}`, name)
}

func testAccAWSCognitoUserPoolConfig_withLambdaConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "main" {
  name = "%s"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_lambda_function" "main" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s"
  role          = "${aws_iam_role.main.arn}"
  handler       = "exports.example"
  runtime       = "nodejs8.10"
}

resource "aws_cognito_user_pool" "main" {
  name = "%[1]s"

  lambda_config {
    create_auth_challenge          = "${aws_lambda_function.main.arn}"
    custom_message                 = "${aws_lambda_function.main.arn}"
    define_auth_challenge          = "${aws_lambda_function.main.arn}"
    post_authentication            = "${aws_lambda_function.main.arn}"
    post_confirmation              = "${aws_lambda_function.main.arn}"
    pre_authentication             = "${aws_lambda_function.main.arn}"
    pre_sign_up                    = "${aws_lambda_function.main.arn}"
    pre_token_generation           = "${aws_lambda_function.main.arn}"
    user_migration                 = "${aws_lambda_function.main.arn}"
    verify_auth_challenge_response = "${aws_lambda_function.main.arn}"
  }
}`, name)
}

func testAccAWSCognitoUserPoolConfig_withLambdaConfigUpdated(name string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "main" {
  name = "%s"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_lambda_function" "main" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s"
  role          = "${aws_iam_role.main.arn}"
  handler       = "exports.example"
  runtime       = "nodejs8.10"
}

resource "aws_lambda_function" "second" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s_second"
  role          = "${aws_iam_role.main.arn}"
  handler       = "exports.example"
  runtime       = "nodejs8.10"
}

resource "aws_cognito_user_pool" "main" {
  name = "%[1]s"

  lambda_config {
    create_auth_challenge          = "${aws_lambda_function.second.arn}"
    custom_message                 = "${aws_lambda_function.second.arn}"
    define_auth_challenge          = "${aws_lambda_function.second.arn}"
    post_authentication            = "${aws_lambda_function.second.arn}"
    post_confirmation              = "${aws_lambda_function.second.arn}"
    pre_authentication             = "${aws_lambda_function.second.arn}"
    pre_sign_up                    = "${aws_lambda_function.second.arn}"
    pre_token_generation           = "${aws_lambda_function.second.arn}"
    user_migration                 = "${aws_lambda_function.second.arn}"
    verify_auth_challenge_response = "${aws_lambda_function.second.arn}"
  }
}`, name)
}

func testAccAWSCognitoUserPoolConfig_withSchemaAttributes(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "main" {
  name = "%[1]s"

  schema {
    attribute_data_type      = "String"
    developer_only_attribute = false
    mutable                  = false
    name                     = "email"
    required                 = true

    string_attribute_constraints {
      min_length = 5
      max_length = 10
    }
  }

  schema {
    attribute_data_type      = "Boolean"
    developer_only_attribute = true
    mutable                  = false
    name                     = "mybool"
    required                 = false
  }
}`, name)
}

func testAccAWSCognitoUserPoolConfig_withSchemaAttributesUpdated(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "main" {
  name = "%[1]s"

  schema {
    attribute_data_type      = "String"
    developer_only_attribute = false
    mutable                  = false
    name                     = "email"
    required                 = true

    string_attribute_constraints {
      min_length = 7
      max_length = 15
    }
  }

  schema {
    attribute_data_type      = "Number"
    developer_only_attribute = true
    mutable                  = true
    name                     = "mynumber"
    required                 = false

    number_attribute_constraints {
      min_value = 2
      max_value = 6
    }
  }

  schema {
    attribute_data_type      = "Number"
    developer_only_attribute = false
    mutable                  = true
    name                     = "mynondevnumber"
    required                 = false

    number_attribute_constraints {
      min_value = 2
      max_value = 6
    }
  }
}`, name)
}

func testAccAWSCognitoUserPoolConfig_withVerificationMessageTemplate(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "pool" {
  name = "terraform-test-pool-%s"

  email_verification_message = "Foo {####} Bar"
  email_verification_subject = "FooBar {####}"
  sms_verification_message   = "{####} Baz"

  # Setting Verification template attributes like EmailMessage, EmailSubject or SmsMessage
  # will implicitly set EmailVerificationMessage, EmailVerificationSubject and SmsVerificationMessage
  # attributes.
  verification_message_template {
    default_email_option  = "CONFIRM_WITH_LINK"
    email_message         = "Foo {####} Bar"
    email_message_by_link = "{##foobar##}"
    email_subject         = "FooBar {####}"
    email_subject_by_link = "foobar"
    sms_message           = "{####} Baz"
  }
}`, name)
}

func testAccAWSCognitoUserPoolConfig_withVerificationMessageTemplateUpdated(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "pool" {
  name = "terraform-test-pool-%s"

  email_verification_message = "Foo {####} Bar"
  email_verification_subject = "FooBar {####}"
  sms_verification_message   = "{####} Baz"

  verification_message_template {
    default_email_option = "CONFIRM_WITH_CODE"
  }
}`, name)
}

func testAccAWSCognitoUserPoolConfig_update(name string, mfaconfig, smsAuthMsg string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_iam_role" "main" {
  name = "test-role-%s"
  path = "/service-role/"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "cognito-idp.amazonaws.com"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "${data.aws_caller_identity.current.account_id}"
        }
      }
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "main" {
  name = "test-role-policy-%s"
  role = "${aws_iam_role.main.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "sns:publish"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}

resource "aws_cognito_user_pool" "pool" {
  name = "terraform-test-pool-%s"
  auto_verified_attributes = ["email"]
  mfa_configuration = "%s"

  email_verification_message = "Foo {####} Bar"
  email_verification_subject = "FooBar {####}"
  sms_verification_message   = "{####} Baz"
  sms_authentication_message = "%s"

  admin_create_user_config {
    allow_admin_create_user_only = true
    unused_account_validity_days = 6

    invite_message_template {
      email_message = "Your username is {username} and temporary password is {####}. "
      email_subject = "FooBar {####}"
      sms_message   = "Your username is {username} and temporary password is {####}."
    }
  }

  device_configuration {
    challenge_required_on_new_device      = true
    device_only_remembered_on_user_prompt = false
  }

  verification_message_template {
    default_email_option = "CONFIRM_WITH_CODE"
  }

  sms_configuration {
    external_id    = "${data.aws_caller_identity.current.account_id}"
    sns_caller_arn = "${aws_iam_role.main.arn}"
  }

  tags = {
    "Name" = "Foo"
  }
}`, name, name, name, mfaconfig, smsAuthMsg)
}
