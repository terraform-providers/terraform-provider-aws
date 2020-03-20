package aws

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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

func TestAccAWSCognitoUserPool_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_Name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "cognito-idp", regexp.MustCompile(`userpool/.+`)),
					resource.TestMatchResourceAttr(resourceName, "endpoint", regexp.MustCompile(`^cognito-idp\.[^.]+\.amazonaws.com/[\w-]+_[0-9a-zA-Z]+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", "0"),
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

func TestAccAWSCognitoUserPool_withAdminCreateUserConfiguration(t *testing.T) {
	name := acctest.RandString(5)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_withAdminCreateUserConfiguration(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.unused_account_validity_days", "6"),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.allow_admin_create_user_only", "true"),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.email_message", "Your username is {username} and temporary password is {####}. "),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.email_subject", "FooBar {####}"),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.sms_message", "Your username is {username} and temporary password is {####}."),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_withAdminCreateUserConfigurationUpdatedError(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.unused_account_validity_days", "6"),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.allow_admin_create_user_only", "false"),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.email_message", "Your username is {username} and constant password is {####}. "),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.email_subject", "Foo{####}BaBaz"),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.sms_message", "Your username is {username} and constant password is {####}."),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_withAdminCreateUserConfigurationUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.unused_account_validity_days", "6"),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.allow_admin_create_user_only", "false"),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.email_message", "Your username is {username} and constant password is {####}. "),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.email_subject", "Foo{####}BaBaz"),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.sms_message", "Your username is {username} and constant password is {####}."),
				),
			},
		},
	})
}

// Reference: https://github.com/terraform-providers/terraform-provider-aws/issues/11858
func TestAccAWSCognitoUserPool_withAdminCreateUserConfigurationAndPasswordPolicy(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_withAdminCreateUserConfigAndPasswordPolicy(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.allow_admin_create_user_only", "true"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.temporary_password_validity_days", "7"),
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

func TestAccAWSCognitoUserPool_withAdvancedSecurityMode(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_AdvancedSecurityMode(rName, "OFF"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "user_pool_add_ons.0.advanced_security_mode", "OFF"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_AdvancedSecurityMode(rName, "ENFORCED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "user_pool_add_ons.0.advanced_security_mode", "ENFORCED"),
				),
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_Name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "user_pool_add_ons.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPool_withDeviceConfiguration(t *testing.T) {
	name := acctest.RandString(5)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_withDeviceConfiguration(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "device_configuration.0.challenge_required_on_new_device", "true"),
					resource.TestCheckResourceAttr(resourceName, "device_configuration.0.device_only_remembered_on_user_prompt", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_withDeviceConfigurationUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "device_configuration.0.challenge_required_on_new_device", "false"),
					resource.TestCheckResourceAttr(resourceName, "device_configuration.0.device_only_remembered_on_user_prompt", "true"),
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
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_withEmailVerificationMessage(name, subject, message),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "email_verification_subject", subject),
					resource.TestCheckResourceAttr(resourceName, "email_verification_message", message),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_withEmailVerificationMessage(name, updatedSubject, upatedMessage),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "email_verification_subject", updatedSubject),
					resource.TestCheckResourceAttr(resourceName, "email_verification_message", upatedMessage),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPool_MfaConfiguration_SmsConfiguration(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_MfaConfiguration_SmsConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "ON"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_MfaConfiguration(rName, cognitoidentityprovider.UserPoolMfaTypeOff),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", "0"),
				),
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_MfaConfiguration_SmsConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "ON"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPool_MfaConfiguration_SmsConfigurationAndSoftwareTokenMfaConfiguration(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_MfaConfiguration_SmsConfigurationAndSoftwareTokenMfaConfigurationEnabled(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "ON"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.0.enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_MfaConfiguration_SmsConfigurationAndSoftwareTokenMfaConfigurationEnabled(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "ON"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.0.enabled", "true"),
				),
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_MfaConfiguration(rName, cognitoidentityprovider.UserPoolMfaTypeOff),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPool_MfaConfiguration_SmsConfigurationToSoftwareTokenMfaConfiguration(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_MfaConfiguration_SmsConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "ON"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_MfaConfiguration_SoftwareTokenMfaConfigurationEnabled(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "ON"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.0.enabled", "true"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPool_MfaConfiguration_SoftwareTokenMfaConfiguration(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_MfaConfiguration_SoftwareTokenMfaConfigurationEnabled(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "ON"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.0.enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_MfaConfiguration(rName, cognitoidentityprovider.UserPoolMfaTypeOff),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", "0"),
				),
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_MfaConfiguration_SoftwareTokenMfaConfigurationEnabled(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "ON"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.0.enabled", "true"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPool_MfaConfiguration_SoftwareTokenMfaConfigurationToSmsConfiguration(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_MfaConfiguration_SoftwareTokenMfaConfigurationEnabled(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "ON"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.0.enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_MfaConfiguration_SmsConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "ON"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "software_token_mfa_configuration.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPool_SmsAuthenticationMessage(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	smsAuthenticationMessage1 := "test authentication message {####}"
	smsAuthenticationMessage2 := "test authentication message updated {####}"
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_SmsAuthenticationMessage(rName, smsAuthenticationMessage1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "sms_authentication_message", smsAuthenticationMessage1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_SmsAuthenticationMessage(rName, smsAuthenticationMessage2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "sms_authentication_message", smsAuthenticationMessage2),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPool_SmsConfiguration(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_SmsConfiguration_ExternalId(rName, "test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_Name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
				),
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_SmsConfiguration_ExternalId(rName, "test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPool_SmsConfiguration_ExternalId(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_SmsConfiguration_ExternalId(rName, "test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_SmsConfiguration_ExternalId(rName, "test2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test2"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPool_SmsConfiguration_SnsCallerArn(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_SmsConfiguration_ExternalId(rName, "test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_SmsConfiguration_SnsCallerArn2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.0.external_id", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "sms_configuration.0.sns_caller_arn", iamRoleResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPool_SmsVerificationMessage(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	smsVerificationMessage1 := "test verification message {####}"
	smsVerificationMessage2 := "test verification message updated {####}"
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_SmsVerificationMessage(rName, smsVerificationMessage1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "sms_verification_message", smsVerificationMessage1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_SmsVerificationMessage(rName, smsVerificationMessage2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "sms_verification_message", smsVerificationMessage2),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPool_withEmailConfiguration(t *testing.T) {
	name := acctest.RandString(5)
	replyTo := fmt.Sprintf("tf-acc-reply-%s@terraformtesting.com", name)
	resourceName := "aws_cognito_user_pool.test"

	sourceARN, ok := os.LookupEnv("TEST_AWS_SES_VERIFIED_EMAIL_ARN")
	if !ok {
		t.Skip("'TEST_AWS_SES_VERIFIED_EMAIL_ARN' not set, skipping test.")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_withEmailConfiguration(name, "", "", "", "COGNITO_DEFAULT"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "email_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "email_configuration.0.reply_to_email_address", ""),
					resource.TestCheckResourceAttr(resourceName, "email_configuration.0.email_sending_account", "COGNITO_DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "email_configuration.0.from_email_address", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_withEmailConfiguration(name, replyTo, sourceARN, "John Smith <john@smith.com>", "DEVELOPER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "email_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "email_configuration.0.reply_to_email_address", replyTo),
					resource.TestCheckResourceAttr(resourceName, "email_configuration.0.email_sending_account", "DEVELOPER"),
					resource.TestCheckResourceAttr(resourceName, "email_configuration.0.source_arn", sourceARN),
					resource.TestCheckResourceAttr(resourceName, "email_configuration.0.from_email_address", "John Smith <john@smith.com>"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPool_withTags(t *testing.T) {
	name := acctest.RandString(5)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_Tags1(name, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists(resourceName),
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
				Config: testAccAWSCognitoUserPoolConfig_Tags2(name, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_Tags1(name, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPool_withAliasAttributes(t *testing.T) {
	name := acctest.RandString(5)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_withAliasAttributes(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "alias_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "alias_attributes.1888159429", "preferred_username"),
					resource.TestCheckResourceAttr(resourceName, "auto_verified_attributes.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_withAliasAttributesUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "alias_attributes.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "alias_attributes.881205744", "email"),
					resource.TestCheckResourceAttr(resourceName, "alias_attributes.1888159429", "preferred_username"),
					resource.TestCheckResourceAttr(resourceName, "auto_verified_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_verified_attributes.881205744", "email"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPool_withPasswordPolicy(t *testing.T) {
	name := acctest.RandString(5)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_withPasswordPolicy(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "password_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.minimum_length", "7"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.require_lowercase", "true"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.require_numbers", "false"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.require_symbols", "true"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.require_uppercase", "false"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.temporary_password_validity_days", "7"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_withPasswordPolicyUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "password_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.minimum_length", "9"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.require_lowercase", "false"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.require_numbers", "true"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.require_symbols", "false"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.require_uppercase", "true"),
					resource.TestCheckResourceAttr(resourceName, "password_policy.0.temporary_password_validity_days", "14"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPool_withUsernameConfiguration(t *testing.T) {
	name := acctest.RandString(5)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_withUsernameConfiguration(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "username_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "username_configuration.0.case_sensitive", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_withUsernameConfigurationUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "username_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "username_configuration.0.case_sensitive", "false"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPool_withLambdaConfig(t *testing.T) {
	name := acctest.RandString(5)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_withLambdaConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "lambda_config.0.create_auth_challenge"),
					resource.TestCheckResourceAttrSet(resourceName, "lambda_config.0.custom_message"),
					resource.TestCheckResourceAttrSet(resourceName, "lambda_config.0.define_auth_challenge"),
					resource.TestCheckResourceAttrSet(resourceName, "lambda_config.0.post_authentication"),
					resource.TestCheckResourceAttrSet(resourceName, "lambda_config.0.post_confirmation"),
					resource.TestCheckResourceAttrSet(resourceName, "lambda_config.0.pre_authentication"),
					resource.TestCheckResourceAttrSet(resourceName, "lambda_config.0.pre_sign_up"),
					resource.TestCheckResourceAttrSet(resourceName, "lambda_config.0.pre_token_generation"),
					resource.TestCheckResourceAttrSet(resourceName, "lambda_config.0.user_migration"),
					resource.TestCheckResourceAttrSet(resourceName, "lambda_config.0.verify_auth_challenge_response"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_withLambdaConfigUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "lambda_config.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "lambda_config.0.create_auth_challenge"),
					resource.TestCheckResourceAttrSet(resourceName, "lambda_config.0.custom_message"),
					resource.TestCheckResourceAttrSet(resourceName, "lambda_config.0.define_auth_challenge"),
					resource.TestCheckResourceAttrSet(resourceName, "lambda_config.0.post_authentication"),
					resource.TestCheckResourceAttrSet(resourceName, "lambda_config.0.post_confirmation"),
					resource.TestCheckResourceAttrSet(resourceName, "lambda_config.0.pre_authentication"),
					resource.TestCheckResourceAttrSet(resourceName, "lambda_config.0.pre_sign_up"),
					resource.TestCheckResourceAttrSet(resourceName, "lambda_config.0.pre_token_generation"),
					resource.TestCheckResourceAttrSet(resourceName, "lambda_config.0.user_migration"),
					resource.TestCheckResourceAttrSet(resourceName, "lambda_config.0.verify_auth_challenge_response"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPool_withSchemaAttributes(t *testing.T) {
	name := acctest.RandString(5)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_withSchemaAttributes(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "schema.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "schema.145451252.attribute_data_type", "String"),
					resource.TestCheckResourceAttr(resourceName, "schema.145451252.developer_only_attribute", "false"),
					resource.TestCheckResourceAttr(resourceName, "schema.145451252.mutable", "false"),
					resource.TestCheckResourceAttr(resourceName, "schema.145451252.name", "email"),
					resource.TestCheckResourceAttr(resourceName, "schema.145451252.number_attribute_constraints.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "schema.145451252.required", "true"),
					resource.TestCheckResourceAttr(resourceName, "schema.145451252.string_attribute_constraints.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema.145451252.string_attribute_constraints.0.min_length", "5"),
					resource.TestCheckResourceAttr(resourceName, "schema.145451252.string_attribute_constraints.0.max_length", "10"),
					resource.TestCheckResourceAttr(resourceName, "schema.770828826.attribute_data_type", "Boolean"),
					resource.TestCheckResourceAttr(resourceName, "schema.770828826.developer_only_attribute", "true"),
					resource.TestCheckResourceAttr(resourceName, "schema.770828826.mutable", "false"),
					resource.TestCheckResourceAttr(resourceName, "schema.770828826.name", "mybool"),
					resource.TestCheckResourceAttr(resourceName, "schema.770828826.number_attribute_constraints.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "schema.770828826.required", "false"),
					resource.TestCheckResourceAttr(resourceName, "schema.770828826.string_attribute_constraints.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_withSchemaAttributesUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "schema.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "schema.2078884933.attribute_data_type", "String"),
					resource.TestCheckResourceAttr(resourceName, "schema.2078884933.developer_only_attribute", "false"),
					resource.TestCheckResourceAttr(resourceName, "schema.2078884933.mutable", "false"),
					resource.TestCheckResourceAttr(resourceName, "schema.2078884933.name", "email"),
					resource.TestCheckResourceAttr(resourceName, "schema.2078884933.number_attribute_constraints.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "schema.2078884933.required", "true"),
					resource.TestCheckResourceAttr(resourceName, "schema.2078884933.string_attribute_constraints.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema.2078884933.string_attribute_constraints.0.min_length", "7"),
					resource.TestCheckResourceAttr(resourceName, "schema.2078884933.string_attribute_constraints.0.max_length", "15"),
					resource.TestCheckResourceAttr(resourceName, "schema.2718111653.attribute_data_type", "Number"),
					resource.TestCheckResourceAttr(resourceName, "schema.2718111653.developer_only_attribute", "true"),
					resource.TestCheckResourceAttr(resourceName, "schema.2718111653.mutable", "true"),
					resource.TestCheckResourceAttr(resourceName, "schema.2718111653.name", "mynumber"),
					resource.TestCheckResourceAttr(resourceName, "schema.2718111653.number_attribute_constraints.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema.2718111653.number_attribute_constraints.0.min_value", "2"),
					resource.TestCheckResourceAttr(resourceName, "schema.2718111653.number_attribute_constraints.0.max_value", "6"),
					resource.TestCheckResourceAttr(resourceName, "schema.2718111653.required", "false"),
					resource.TestCheckResourceAttr(resourceName, "schema.2718111653.string_attribute_constraints.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "schema.2753746449.attribute_data_type", "Number"),
					resource.TestCheckResourceAttr(resourceName, "schema.2753746449.developer_only_attribute", "false"),
					resource.TestCheckResourceAttr(resourceName, "schema.2753746449.mutable", "true"),
					resource.TestCheckResourceAttr(resourceName, "schema.2753746449.name", "mynondevnumber"),
					resource.TestCheckResourceAttr(resourceName, "schema.2753746449.number_attribute_constraints.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema.2753746449.number_attribute_constraints.0.min_value", "2"),
					resource.TestCheckResourceAttr(resourceName, "schema.2753746449.number_attribute_constraints.0.max_value", "6"),
					resource.TestCheckResourceAttr(resourceName, "schema.2753746449.required", "false"),
					resource.TestCheckResourceAttr(resourceName, "schema.2753746449.string_attribute_constraints.#", "0"),
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

func TestAccAWSCognitoUserPool_withVerificationMessageTemplate(t *testing.T) {
	name := acctest.RandString(5)
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_withVerificationMessageTemplate(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.default_email_option", "CONFIRM_WITH_LINK"),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.email_message", "foo {####} bar"),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.email_message_by_link", "{##foobar##}"),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.email_subject", "foobar {####}"),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.email_subject_by_link", "foobar"),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.sms_message", "{####} baz"),

					/* Setting Verification template attributes like EmailMessage, EmailSubject or SmsMessage
					will implicitly set EmailVerificationMessage, EmailVerificationSubject and SmsVerificationMessage attributes.
					*/
					resource.TestCheckResourceAttr(resourceName, "email_verification_message", "foo {####} bar"),
					resource.TestCheckResourceAttr(resourceName, "email_verification_subject", "foobar {####}"),
					resource.TestCheckResourceAttr(resourceName, "sms_verification_message", "{####} baz"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_withVerificationMessageTemplate_DefaultEmailOption(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.default_email_option", "CONFIRM_WITH_CODE"),
					resource.TestCheckResourceAttr(resourceName, "email_verification_message", "{####} Baz"),
					resource.TestCheckResourceAttr(resourceName, "email_verification_subject", "BazBaz {####}"),
					resource.TestCheckResourceAttr(resourceName, "sms_verification_message", "{####} BazBazBar?"),

					/* Setting EmailVerificationMessage, EmailVerificationSubject and SmsVerificationMessage attributes
					will implicitly set verification template attributes like EmailMessage, EmailSubject or SmsMessage.
					*/
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.email_message", "{####} Baz"),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.email_subject", "BazBaz {####}"),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.sms_message", "{####} BazBazBar?"),
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
	resourceName := "aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolConfig_update(name, optionalMfa, authenticationMessage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_verified_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", optionalMfa),
					resource.TestCheckResourceAttr(resourceName, "email_verification_message", "Foo {####} Bar"),
					resource.TestCheckResourceAttr(resourceName, "email_verification_subject", "FooBar {####}"),
					resource.TestCheckResourceAttr(resourceName, "sms_verification_message", "{####} Baz"),
					resource.TestCheckResourceAttr(resourceName, "sms_authentication_message", authenticationMessage),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.unused_account_validity_days", "6"),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.allow_admin_create_user_only", "true"),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.email_message", "Your username is {username} and temporary password is {####}. "),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.email_subject", "FooBar {####}"),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.sms_message", "Your username is {username} and temporary password is {####}."),
					resource.TestCheckResourceAttr(resourceName, "device_configuration.0.challenge_required_on_new_device", "true"),
					resource.TestCheckResourceAttr(resourceName, "device_configuration.0.device_only_remembered_on_user_prompt", "false"),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.default_email_option", "CONFIRM_WITH_CODE"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "sms_configuration.0.external_id"),
					resource.TestCheckResourceAttrSet(resourceName, "sms_configuration.0.sns_caller_arn"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Foo"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_update(name, optionalMfa, updatedAuthenticationMessage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_verified_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", optionalMfa),
					resource.TestCheckResourceAttr(resourceName, "email_verification_message", "Foo {####} Bar"),
					resource.TestCheckResourceAttr(resourceName, "email_verification_subject", "FooBar {####}"),
					resource.TestCheckResourceAttr(resourceName, "sms_verification_message", "{####} Baz"),
					resource.TestCheckResourceAttr(resourceName, "sms_authentication_message", updatedAuthenticationMessage),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.unused_account_validity_days", "6"),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.allow_admin_create_user_only", "true"),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.email_message", "Your username is {username} and temporary password is {####}. "),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.email_subject", "FooBar {####}"),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.sms_message", "Your username is {username} and temporary password is {####}."),
					resource.TestCheckResourceAttr(resourceName, "device_configuration.0.challenge_required_on_new_device", "true"),
					resource.TestCheckResourceAttr(resourceName, "device_configuration.0.device_only_remembered_on_user_prompt", "false"),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.default_email_option", "CONFIRM_WITH_CODE"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "sms_configuration.0.external_id"),
					resource.TestCheckResourceAttrSet(resourceName, "sms_configuration.0.sns_caller_arn"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Foo"),
				),
			},
			{
				Config: testAccAWSCognitoUserPoolConfig_update(name, offMfa, updatedAuthenticationMessage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_verified_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mfa_configuration", offMfa),
					resource.TestCheckResourceAttr(resourceName, "email_verification_message", "Foo {####} Bar"),
					resource.TestCheckResourceAttr(resourceName, "email_verification_subject", "FooBar {####}"),
					resource.TestCheckResourceAttr(resourceName, "sms_verification_message", "{####} Baz"),
					resource.TestCheckResourceAttr(resourceName, "sms_authentication_message", updatedAuthenticationMessage),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.unused_account_validity_days", "6"),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.allow_admin_create_user_only", "true"),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.email_message", "Your username is {username} and temporary password is {####}. "),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.email_subject", "FooBar {####}"),
					resource.TestCheckResourceAttr(resourceName, "admin_create_user_config.0.invite_message_template.0.sms_message", "Your username is {username} and temporary password is {####}."),
					resource.TestCheckResourceAttr(resourceName, "device_configuration.0.challenge_required_on_new_device", "true"),
					resource.TestCheckResourceAttr(resourceName, "device_configuration.0.device_only_remembered_on_user_prompt", "false"),
					resource.TestCheckResourceAttr(resourceName, "verification_message_template.0.default_email_option", "CONFIRM_WITH_CODE"),
					resource.TestCheckResourceAttr(resourceName, "sms_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "sms_configuration.0.external_id"),
					resource.TestCheckResourceAttrSet(resourceName, "sms_configuration.0.sns_caller_arn"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Foo"),
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
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == cognitoidentityprovider.ErrCodeResourceNotFoundException {
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

func testAccPreCheckAWSCognitoIdentityProvider(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).cognitoidpconn

	input := &cognitoidentityprovider.ListUserPoolsInput{
		MaxResults: aws.Int64(int64(1)),
	}

	_, err := conn.ListUserPools(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSCognitoUserPoolConfigSmsConfigurationBase(rName string, externalID string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Statement = [{
      Action    = "sts:AssumeRole"
      Condition = {
        "StringEquals" = {
          "sts:ExternalId" = %[2]q
        }
      }
      Effect    = "Allow"
      Principal = {
        Service = "cognito-idp.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.id

  policy = jsonencode({
    Statement = [{
      Action   = "sns:publish"
      Effect   = "Allow"
      Resource = "*"
    }]
    Version = "2012-10-17"
  })
}
`, rName, externalID)
}

func testAccAWSCognitoUserPoolConfig_Name(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}
`, rName)
}

func testAccAWSCognitoUserPoolConfig_withAdminCreateUserConfiguration(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
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
}
`, name)
}

func testAccAWSCognitoUserPoolConfig_withAdminCreateUserConfigurationUpdatedError(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
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
}
`, name)
}

func testAccAWSCognitoUserPoolConfig_withAdminCreateUserConfigurationUpdated(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = "terraform-test-pool-%s"

  admin_create_user_config {
    allow_admin_create_user_only = false
    unused_account_validity_days = 6

    invite_message_template {
      email_message = "Your username is {username} and constant password is {####}. "
      email_subject = "Foo{####}BaBaz"
      sms_message   = "Your username is {username} and constant password is {####}."
    }
  }
}
`, name)
}

func testAccAWSCognitoUserPoolConfig_AdvancedSecurityMode(rName string, advancedSecurityMode string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  user_pool_add_ons {
    advanced_security_mode = %[2]q
  }
}
`, rName, advancedSecurityMode)
}

func testAccAWSCognitoUserPoolConfig_withDeviceConfiguration(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = "terraform-test-pool-%s"

  device_configuration {
    challenge_required_on_new_device      = true
    device_only_remembered_on_user_prompt = false
  }
}
`, name)
}

func testAccAWSCognitoUserPoolConfig_withDeviceConfigurationUpdated(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = "terraform-test-pool-%s"

  device_configuration {
    challenge_required_on_new_device      = false
    device_only_remembered_on_user_prompt = true
  }
}
`, name)
}

func testAccAWSCognitoUserPoolConfig_withEmailVerificationMessage(name, subject, message string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name                       = "terraform-test-pool-%s"
  email_verification_subject = "%s"
  email_verification_message = "%s"

  verification_message_template {
    default_email_option = "CONFIRM_WITH_CODE"
  }
}
`, name, subject, message)
}

func testAccAWSCognitoUserPoolConfig_MfaConfiguration(rName string, mfaConfiguration string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  mfa_configuration = %[2]q
  name              = %[1]q
}
`, rName, mfaConfiguration)
}

func testAccAWSCognitoUserPoolConfig_MfaConfiguration_SmsConfiguration(rName string) string {
	return testAccAWSCognitoUserPoolConfigSmsConfigurationBase(rName, "test") + fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  mfa_configuration = "ON"
  name              = %[1]q

  sms_configuration {
    external_id    = "test"
    sns_caller_arn = aws_iam_role.test.arn
  }
}
`, rName)
}

func testAccAWSCognitoUserPoolConfig_MfaConfiguration_SmsConfigurationAndSoftwareTokenMfaConfigurationEnabled(rName string, enabled bool) string {
	return testAccAWSCognitoUserPoolConfigSmsConfigurationBase(rName, "test") + fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  mfa_configuration = "ON"
  name              = %[1]q

  sms_configuration {
    external_id    = "test"
    sns_caller_arn = aws_iam_role.test.arn
  }

  software_token_mfa_configuration {
    enabled = %[2]t
  }
}
`, rName, enabled)
}

func testAccAWSCognitoUserPoolConfig_MfaConfiguration_SoftwareTokenMfaConfigurationEnabled(rName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  mfa_configuration = "ON"
  name              = %[1]q

  software_token_mfa_configuration {
    enabled = %[2]t
  }
}
`, rName, enabled)
}

func testAccAWSCognitoUserPoolConfig_SmsAuthenticationMessage(rName, smsAuthenticationMessage string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name                       = %[1]q
  sms_authentication_message = %[2]q
}
`, rName, smsAuthenticationMessage)
}

func testAccAWSCognitoUserPoolConfig_SmsConfiguration_ExternalId(rName string, externalID string) string {
	return testAccAWSCognitoUserPoolConfigSmsConfigurationBase(rName, externalID) + fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  sms_configuration {
    external_id    = %[2]q
    sns_caller_arn = aws_iam_role.test.arn
  }
}
`, rName, externalID)
}

func testAccAWSCognitoUserPoolConfig_SmsConfiguration_SnsCallerArn2(rName string) string {
	return testAccAWSCognitoUserPoolConfigSmsConfigurationBase(rName+"-2", "test") + fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  sms_configuration {
    external_id    = "test"
    sns_caller_arn = aws_iam_role.test.arn
  }
}
`, rName)
}

func testAccAWSCognitoUserPoolConfig_SmsVerificationMessage(rName, smsVerificationMessage string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name                     = %[1]q
  sms_verification_message = %[2]q
}
`, rName, smsVerificationMessage)
}

func testAccAWSCognitoUserPoolConfig_Tags1(name, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, name, tagKey1, tagValue1)
}

func testAccAWSCognitoUserPoolConfig_Tags2(name, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, name, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSCognitoUserPoolConfig_withEmailConfiguration(name, email, arn, from, account string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
    name = "terraform-test-pool-%[1]s"


    email_configuration {
      reply_to_email_address = %[2]q
      source_arn = %[3]q
      from_email_address = %[4]q
      email_sending_account = %[5]q
    }
  }`, name, email, arn, from, account)
}

func testAccAWSCognitoUserPoolConfig_withAliasAttributes(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = "terraform-test-pool-%s"

  alias_attributes = ["preferred_username"]
}
`, name)
}

func testAccAWSCognitoUserPoolConfig_withAliasAttributesUpdated(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = "terraform-test-pool-%s"

  alias_attributes         = ["email", "preferred_username"]
  auto_verified_attributes = ["email"]
}
`, name)
}

func testAccAWSCognitoUserPoolConfig_withAdminCreateUserConfigAndPasswordPolicy(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  admin_create_user_config {
    allow_admin_create_user_only = true
  }

  password_policy {
    minimum_length                   = 7
    require_lowercase                = true
    require_numbers                  = false
    require_symbols                  = true
    require_uppercase                = false
    temporary_password_validity_days = 7
  }
}
`, rName)
}

func testAccAWSCognitoUserPoolConfig_withPasswordPolicy(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = "terraform-test-pool-%s"

  password_policy {
    minimum_length                   = 7
    require_lowercase                = true
    require_numbers                  = false
    require_symbols                  = true
    require_uppercase                = false
    temporary_password_validity_days = 7
  }
}
`, name)
}

func testAccAWSCognitoUserPoolConfig_withPasswordPolicyUpdated(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = "terraform-test-pool-%s"

  password_policy {
    minimum_length                   = 9
    require_lowercase                = false
    require_numbers                  = true
    require_symbols                  = false
    require_uppercase                = true
    temporary_password_validity_days = 14
  }
}
`, name)
}

func testAccAWSCognitoUserPoolConfig_withUsernameConfiguration(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = "terraform-test-pool-%s"

  username_configuration {
    case_sensitive = true
  }
}
`, name)
}

func testAccAWSCognitoUserPoolConfig_withUsernameConfigurationUpdated(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = "terraform-test-pool-%s"

  username_configuration {
    case_sensitive = false
  }
}
`, name)
}

func testAccAWSCognitoUserPoolConfig_withLambdaConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
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

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s"
  role          = "${aws_iam_role.test.arn}"
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}

resource "aws_cognito_user_pool" "test" {
  name = "%[1]s"

  lambda_config {
    create_auth_challenge          = "${aws_lambda_function.test.arn}"
    custom_message                 = "${aws_lambda_function.test.arn}"
    define_auth_challenge          = "${aws_lambda_function.test.arn}"
    post_authentication            = "${aws_lambda_function.test.arn}"
    post_confirmation              = "${aws_lambda_function.test.arn}"
    pre_authentication             = "${aws_lambda_function.test.arn}"
    pre_sign_up                    = "${aws_lambda_function.test.arn}"
    pre_token_generation           = "${aws_lambda_function.test.arn}"
    user_migration                 = "${aws_lambda_function.test.arn}"
    verify_auth_challenge_response = "${aws_lambda_function.test.arn}"
  }
}
`, name)
}

func testAccAWSCognitoUserPoolConfig_withLambdaConfigUpdated(name string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
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

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s"
  role          = "${aws_iam_role.test.arn}"
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}

resource "aws_lambda_function" "second" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s_second"
  role          = "${aws_iam_role.test.arn}"
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}

resource "aws_cognito_user_pool" "test" {
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
}
`, name)
}

func testAccAWSCognitoUserPoolConfig_withSchemaAttributes(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
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
}
`, name)
}

func testAccAWSCognitoUserPoolConfig_withSchemaAttributesUpdated(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
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
}
`, name)
}

func testAccAWSCognitoUserPoolConfig_withVerificationMessageTemplate(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = "terraform-test-pool-%s"

  # Setting Verification template attributes like EmailMessage, EmailSubject or SmsMessage
  # will implicitly set EmailVerificationMessage, EmailVerificationSubject and SmsVerificationMessage
  # attributes.
  verification_message_template {
    default_email_option  = "CONFIRM_WITH_LINK"
    email_message = "foo {####} bar"
    email_message_by_link = "{##foobar##}"
    email_subject = "foobar {####}"
    email_subject_by_link = "foobar"
    sms_message           = "{####} baz"
  }
}
`, name)
}

func testAccAWSCognitoUserPoolConfig_withVerificationMessageTemplate_DefaultEmailOption(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = "terraform-test-pool-%s"

  email_verification_message = "{####} Baz"
  email_verification_subject = "BazBaz {####}"
  sms_verification_message   = "{####} BazBazBar?"

  verification_message_template {
    default_email_option = "CONFIRM_WITH_CODE"
  }
}
`, name)
}

func testAccAWSCognitoUserPoolConfig_update(name string, mfaconfig, smsAuthMsg string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_iam_role" "test" {
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

resource "aws_iam_role_policy" "test" {
  name = "test-role-policy-%s"
  role = "${aws_iam_role.test.id}"

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

resource "aws_cognito_user_pool" "test" {
  name                     = "terraform-test-pool-%s"
  auto_verified_attributes = ["email"]
  mfa_configuration        = "%s"

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
    sns_caller_arn = "${aws_iam_role.test.arn}"
  }

  tags = {
    "Name" = "Foo"
  }
}
`, name, name, name, mfaconfig, smsAuthMsg)
}
