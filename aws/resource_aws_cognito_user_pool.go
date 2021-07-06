package aws

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	iamwaiter "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
)

func resourceAwsCognitoUserPool() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCognitoUserPoolCreate,
		Read:   resourceAwsCognitoUserPoolRead,
		Update: resourceAwsCognitoUserPoolUpdate,
		Delete: resourceAwsCognitoUserPoolDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		// https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_CreateUserPool.html
		Schema: map[string]*schema.Schema{
			"account_recovery_setting": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"recovery_mechanism": {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(cognitoidentityprovider.RecoveryOptionNameType_Values(), false),
									},
									"priority": {
										Type:     schema.TypeInt,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"admin_create_user_config": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow_admin_create_user_only": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"invite_message_template": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"email_message": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateCognitoUserPoolInviteTemplateEmailMessage,
									},
									"email_subject": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateCognitoUserPoolTemplateEmailSubject,
									},
									"sms_message": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateCognitoUserPoolInviteTemplateSmsMessage,
									},
								},
							},
						},
					},
				},
			},
			"alias_attributes": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(cognitoidentityprovider.AliasAttributeType_Values(), false),
				},
				ConflictsWith: []string{"username_attributes"},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"custom_domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"estimated_number_of_users": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"auto_verified_attributes": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(cognitoidentityprovider.VerifiedAttributeType_Values(), false),
				},
			},
			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"device_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"challenge_required_on_new_device": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"device_only_remembered_on_user_prompt": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"email_configuration": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"configuration_set": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"email_sending_account": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      cognitoidentityprovider.EmailSendingAccountTypeCognitoDefault,
							ValidateFunc: validation.StringInSlice(cognitoidentityprovider.EmailSendingAccountType_Values(), false),
						},
						"from_email_address": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"reply_to_email_address": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.Any(
								validation.StringInSlice([]string{""}, false),
								validation.StringMatch(regexp.MustCompile(`[\p{L}\p{M}\p{S}\p{N}\p{P}]+@[\p{L}\p{M}\p{S}\p{N}\p{P}]+`),
									`must satisfy regular expression pattern: [\p{L}\p{M}\p{S}\p{N}\p{P}]+@[\p{L}\p{M}\p{S}\p{N}\p{P}]+`),
							),
						},
						"source_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},
			"email_verification_subject": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ValidateFunc:  validateCognitoUserPoolEmailVerificationSubject,
				ConflictsWith: []string{"verification_message_template.0.email_subject"},
			},
			"email_verification_message": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ValidateFunc:  validateCognitoUserPoolEmailVerificationMessage,
				ConflictsWith: []string{"verification_message_template.0.email_message"},
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"lambda_config": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"create_auth_challenge": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
						},
						"custom_message": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
						},
						"define_auth_challenge": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
						},
						"post_authentication": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
						},
						"post_confirmation": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
						},
						"pre_authentication": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
						},
						"pre_sign_up": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
						},
						"pre_token_generation": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
						},
						"user_migration": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
						},
						"verify_auth_challenge_response": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
						},
						"kms_key_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
						},
						"custom_email_sender": {
							Type:         schema.TypeList,
							Optional:     true,
							Computed:     true,
							MaxItems:     1,
							RequiredWith: []string{"lambda_config.0.kms_key_id"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"lambda_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validateArn,
									},
									"lambda_version": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(cognitoidentityprovider.CustomEmailSenderLambdaVersionType_Values(), false),
									},
								},
							},
						},
						"custom_sms_sender": {
							Type:         schema.TypeList,
							Optional:     true,
							Computed:     true,
							MaxItems:     1,
							RequiredWith: []string{"lambda_config.0.kms_key_id"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"lambda_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validateArn,
									},
									"lambda_version": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(cognitoidentityprovider.CustomSMSSenderLambdaVersionType_Values(), false),
									},
								},
							},
						},
					},
				},
			},
			"last_modified_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"mfa_configuration": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      cognitoidentityprovider.UserPoolMfaTypeOff,
				ValidateFunc: validation.StringInSlice(cognitoidentityprovider.UserPoolMfaType_Values(), false),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.Any(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexp.MustCompile(`[\w\s+=,.@-]+`),
						`must satisfy regular expression pattern: [\w\s+=,.@-]+`),
				),
			},
			"password_policy": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"minimum_length": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(6, 99),
						},
						"require_lowercase": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"require_numbers": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"require_symbols": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"require_uppercase": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"temporary_password_validity_days": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 365),
						},
					},
				},
			},
			"schema": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 1,
				MaxItems: 50,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attribute_data_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(cognitoidentityprovider.AttributeDataType_Values(), false),
						},
						"developer_only_attribute": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"mutable": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateCognitoUserPoolSchemaName,
						},
						"number_attribute_constraints": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"min_value": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"max_value": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"required": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"string_attribute_constraints": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"min_length": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"max_length": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"sms_authentication_message": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateCognitoUserPoolSmsAuthenticationMessage,
			},
			"sms_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"external_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"sns_caller_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},
			"sms_verification_message": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ValidateFunc:  validateCognitoUserPoolSmsVerificationMessage,
				ConflictsWith: []string{"verification_message_template.0.sms_message"},
			},
			"software_token_mfa_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
			"username_attributes": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(cognitoidentityprovider.UsernameAttributeType_Values(), false),
				},
				ConflictsWith: []string{"alias_attributes"},
			},
			"username_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"case_sensitive": {
							Type:     schema.TypeBool,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			"user_pool_add_ons": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"advanced_security_mode": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(cognitoidentityprovider.AdvancedSecurityModeType_Values(), false),
						},
					},
				},
			},
			"verification_message_template": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"default_email_option": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      cognitoidentityprovider.DefaultEmailOptionTypeConfirmWithCode,
							ValidateFunc: validation.StringInSlice(cognitoidentityprovider.DefaultEmailOptionType_Values(), false),
						},
						"email_message": {
							Type:          schema.TypeString,
							Optional:      true,
							Computed:      true,
							ValidateFunc:  validateCognitoUserPoolTemplateEmailMessage,
							ConflictsWith: []string{"email_verification_message"},
						},
						"email_message_by_link": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validateCognitoUserPoolTemplateEmailMessageByLink,
						},
						"email_subject": {
							Type:          schema.TypeString,
							Optional:      true,
							Computed:      true,
							ValidateFunc:  validateCognitoUserPoolTemplateEmailSubject,
							ConflictsWith: []string{"email_verification_subject"},
						},
						"email_subject_by_link": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validateCognitoUserPoolTemplateEmailSubjectByLink,
						},
						"sms_message": {
							Type:          schema.TypeString,
							Optional:      true,
							Computed:      true,
							ValidateFunc:  validateCognitoUserPoolTemplateSmsMessage,
							ConflictsWith: []string{"sms_verification_message"},
						},
					},
				},
			},
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsCognitoUserPoolCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	params := &cognitoidentityprovider.CreateUserPoolInput{
		PoolName: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("admin_create_user_config"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if ok && config != nil {
			params.AdminCreateUserConfig = expandCognitoUserPoolAdminCreateUserConfig(config)
		}
	}

	if v, ok := d.GetOk("account_recovery_setting"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if ok && config != nil {
			params.AccountRecoverySetting = expandCognitoUserPoolAccountRecoverySettingConfig(config)
		}
	}

	if v, ok := d.GetOk("alias_attributes"); ok {
		params.AliasAttributes = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("auto_verified_attributes"); ok {
		params.AutoVerifiedAttributes = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("email_configuration"); ok && len(v.([]interface{})) > 0 {
		params.EmailConfiguration = expandCognitoUserPoolEmailConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("admin_create_user_config"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if ok && config != nil {
			params.AdminCreateUserConfig = expandCognitoUserPoolAdminCreateUserConfig(config)
		}
	}

	if v, ok := d.GetOk("device_configuration"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if ok && config != nil {
			params.DeviceConfiguration = expandCognitoUserPoolDeviceConfiguration(config)
		}
	}

	if v, ok := d.GetOk("email_verification_subject"); ok {
		params.EmailVerificationSubject = aws.String(v.(string))
	}

	if v, ok := d.GetOk("email_verification_message"); ok {
		params.EmailVerificationMessage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("lambda_config"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if ok && config != nil {
			params.LambdaConfig = expandCognitoUserPoolLambdaConfig(config)
		}
	}

	if v, ok := d.GetOk("password_policy"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if ok && config != nil {
			policies := &cognitoidentityprovider.UserPoolPolicyType{}
			policies.PasswordPolicy = expandCognitoUserPoolPasswordPolicy(config)
			params.Policies = policies
		}
	}

	if v, ok := d.GetOk("schema"); ok {
		params.Schema = expandCognitoUserPoolSchema(v.(*schema.Set).List())
	}

	// For backwards compatibility, include this outside of MFA configuration
	// since its configuration is allowed by the API even without SMS MFA.
	if v, ok := d.GetOk("sms_authentication_message"); ok {
		params.SmsAuthenticationMessage = aws.String(v.(string))
	}

	// Include the SMS configuration outside of MFA configuration since it
	// can be used for user verification.
	if v, ok := d.GetOk("sms_configuration"); ok {
		params.SmsConfiguration = expandCognitoSmsConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("username_attributes"); ok {
		params.UsernameAttributes = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("username_configuration"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if ok && config != nil {
			params.UsernameConfiguration = expandCognitoUserPoolUsernameConfiguration(config)
		}
	}

	if v, ok := d.GetOk("user_pool_add_ons"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if ok {
			userPoolAddons := &cognitoidentityprovider.UserPoolAddOnsType{}

			if v, ok := config["advanced_security_mode"]; ok && v.(string) != "" {
				userPoolAddons.AdvancedSecurityMode = aws.String(v.(string))
			}
			params.UserPoolAddOns = userPoolAddons
		}
	}

	if v, ok := d.GetOk("verification_message_template"); ok {
		configs := v.([]interface{})
		config, ok := configs[0].(map[string]interface{})

		if ok && config != nil {
			params.VerificationMessageTemplate = expandCognitoUserPoolVerificationMessageTemplate(config)
		}
	}

	if v, ok := d.GetOk("sms_verification_message"); ok {
		params.SmsVerificationMessage = aws.String(v.(string))
	}

	if len(tags) > 0 {
		params.UserPoolTags = tags.IgnoreAws().CognitoidentityproviderTags()
	}
	log.Printf("[DEBUG] Creating Cognito User Pool: %s", params)

	// IAM roles & policies can take some time to propagate and be attached
	// to the User Pool
	var resp *cognitoidentityprovider.CreateUserPoolOutput
	err := resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		var err error
		resp, err = conn.CreateUserPool(params)
		if isAWSErr(err, cognitoidentityprovider.ErrCodeInvalidSmsRoleTrustRelationshipException, "Role does not have a trust relationship allowing Cognito to assume the role") {
			log.Printf("[DEBUG] Received %s, retrying CreateUserPool", err)
			return resource.RetryableError(err)
		}
		if isAWSErr(err, cognitoidentityprovider.ErrCodeInvalidSmsRoleAccessPolicyException, "Role does not have permission to publish with SNS") {
			log.Printf("[DEBUG] Received %s, retrying CreateUserPool", err)
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		resp, err = conn.CreateUserPool(params)
	}
	if err != nil {
		return fmt.Errorf("error creating Cognito User Pool: %w", err)
	}

	d.SetId(aws.StringValue(resp.UserPool.Id))

	if v := d.Get("mfa_configuration").(string); v != cognitoidentityprovider.UserPoolMfaTypeOff {
		input := &cognitoidentityprovider.SetUserPoolMfaConfigInput{
			MfaConfiguration:              aws.String(v),
			SoftwareTokenMfaConfiguration: expandCognitoSoftwareTokenMfaConfiguration(d.Get("software_token_mfa_configuration").([]interface{})),
			UserPoolId:                    aws.String(d.Id()),
		}

		if v := d.Get("sms_configuration").([]interface{}); len(v) > 0 && v[0] != nil {
			input.SmsMfaConfiguration = &cognitoidentityprovider.SmsMfaConfigType{
				SmsConfiguration: expandCognitoSmsConfiguration(v),
			}

			if v, ok := d.GetOk("sms_authentication_message"); ok {
				input.SmsMfaConfiguration.SmsAuthenticationMessage = aws.String(v.(string))
			}
		}

		// IAM Roles and Policies can take some time to propagate
		err := resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
			_, err := conn.SetUserPoolMfaConfig(input)

			if isAWSErr(err, cognitoidentityprovider.ErrCodeInvalidSmsRoleTrustRelationshipException, "Role does not have a trust relationship allowing Cognito to assume the role") {
				return resource.RetryableError(err)
			}

			if isAWSErr(err, cognitoidentityprovider.ErrCodeInvalidSmsRoleAccessPolicyException, "Role does not have permission to publish with SNS") {
				return resource.RetryableError(err)
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return nil
		})

		if isResourceTimeoutError(err) {
			_, err = conn.SetUserPoolMfaConfig(input)
		}

		if err != nil {
			return fmt.Errorf("error setting Cognito User Pool (%s) MFA Configuration: %w", d.Id(), err)
		}
	}

	return resourceAwsCognitoUserPoolRead(d, meta)
}

func resourceAwsCognitoUserPoolRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	params := &cognitoidentityprovider.DescribeUserPoolInput{
		UserPoolId: aws.String(d.Id()),
	}

	resp, err := conn.DescribeUserPool(params)

	if isAWSErr(err, cognitoidentityprovider.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Cognito User Pool (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing Cognito User Pool (%s): %w", d.Id(), err)
	}

	userPool := resp.UserPool

	if err := d.Set("admin_create_user_config", flattenCognitoUserPoolAdminCreateUserConfig(userPool.AdminCreateUserConfig)); err != nil {
		return fmt.Errorf("failed setting admin_create_user_config: %w", err)
	}
	if userPool.AliasAttributes != nil {
		d.Set("alias_attributes", flattenStringSet(userPool.AliasAttributes))
	}

	d.Set("arn", userPool.Arn)
	d.Set("custom_domain", userPool.CustomDomain)
	d.Set("domain", userPool.Domain)
	d.Set("estimated_number_of_users", userPool.EstimatedNumberOfUsers)
	d.Set("endpoint", fmt.Sprintf("%s/%s", meta.(*AWSClient).RegionalHostname("cognito-idp"), d.Id()))
	d.Set("auto_verified_attributes", flattenStringSet(userPool.AutoVerifiedAttributes))

	if userPool.EmailVerificationSubject != nil {
		d.Set("email_verification_subject", userPool.EmailVerificationSubject)
	}
	if userPool.EmailVerificationMessage != nil {
		d.Set("email_verification_message", userPool.EmailVerificationMessage)
	}
	if err := d.Set("lambda_config", flattenCognitoUserPoolLambdaConfig(userPool.LambdaConfig)); err != nil {
		return fmt.Errorf("failed setting lambda_config: %w", err)
	}
	if userPool.SmsVerificationMessage != nil {
		d.Set("sms_verification_message", userPool.SmsVerificationMessage)
	}
	if userPool.SmsAuthenticationMessage != nil {
		d.Set("sms_authentication_message", userPool.SmsAuthenticationMessage)
	}

	if err := d.Set("device_configuration", flattenCognitoUserPoolDeviceConfiguration(userPool.DeviceConfiguration)); err != nil {
		return fmt.Errorf("failed setting device_configuration: %w", err)
	}

	if err := d.Set("account_recovery_setting", flattenCognitoUserPoolAccountRecoverySettingConfig(userPool.AccountRecoverySetting)); err != nil {
		return fmt.Errorf("failed setting account_recovery_setting: %w", err)
	}

	if userPool.EmailConfiguration != nil {
		if err := d.Set("email_configuration", flattenCognitoUserPoolEmailConfiguration(userPool.EmailConfiguration)); err != nil {
			return fmt.Errorf("failed setting email_configuration: %w", err)
		}
	}

	if userPool.Policies != nil && userPool.Policies.PasswordPolicy != nil {
		if err := d.Set("password_policy", flattenCognitoUserPoolPasswordPolicy(userPool.Policies.PasswordPolicy)); err != nil {
			return fmt.Errorf("failed setting password_policy: %w", err)
		}
	}

	var configuredSchema []interface{}
	if v, ok := d.GetOk("schema"); ok {
		configuredSchema = v.(*schema.Set).List()
	}
	if err := d.Set("schema", flattenCognitoUserPoolSchema(expandCognitoUserPoolSchema(configuredSchema), userPool.SchemaAttributes)); err != nil {
		return fmt.Errorf("failed setting schema: %w", err)
	}

	if err := d.Set("sms_configuration", flattenCognitoSmsConfiguration(userPool.SmsConfiguration)); err != nil {
		return fmt.Errorf("failed setting sms_configuration: %w", err)
	}

	if userPool.UsernameAttributes != nil {
		d.Set("username_attributes", flattenStringSet(userPool.UsernameAttributes))
	}

	if err := d.Set("username_configuration", flattenCognitoUserPoolUsernameConfiguration(userPool.UsernameConfiguration)); err != nil {
		return fmt.Errorf("failed setting username_configuration: %w", err)
	}

	if err := d.Set("user_pool_add_ons", flattenCognitoUserPoolUserPoolAddOns(userPool.UserPoolAddOns)); err != nil {
		return fmt.Errorf("failed setting user_pool_add_ons: %w", err)
	}

	if err := d.Set("verification_message_template", flattenCognitoUserPoolVerificationMessageTemplate(userPool.VerificationMessageTemplate)); err != nil {
		return fmt.Errorf("failed setting verification_message_template: %w", err)
	}

	d.Set("creation_date", userPool.CreationDate.Format(time.RFC3339))
	d.Set("last_modified_date", userPool.LastModifiedDate.Format(time.RFC3339))
	d.Set("name", userPool.Name)
	tags := keyvaluetags.CognitoidentityKeyValueTags(userPool.UserPoolTags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	input := &cognitoidentityprovider.GetUserPoolMfaConfigInput{
		UserPoolId: aws.String(d.Id()),
	}

	output, err := conn.GetUserPoolMfaConfig(input)

	if isAWSErr(err, cognitoidentityprovider.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Cognito User Pool (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting Cognito User Pool (%s) MFA Configuration: %w", d.Id(), err)
	}

	d.Set("mfa_configuration", output.MfaConfiguration)

	if err := d.Set("software_token_mfa_configuration", flattenCognitoSoftwareTokenMfaConfiguration(output.SoftwareTokenMfaConfiguration)); err != nil {
		return fmt.Errorf("error setting software_token_mfa_configuration: %w", err)
	}

	return nil
}

func resourceAwsCognitoUserPoolUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	// Multi-Factor Authentication updates
	if d.HasChanges(
		"mfa_configuration",
		"sms_authentication_message",
		"sms_configuration",
		"software_token_mfa_configuration",
	) {
		mfaConfiguration := d.Get("mfa_configuration").(string)
		input := &cognitoidentityprovider.SetUserPoolMfaConfigInput{
			MfaConfiguration:              aws.String(mfaConfiguration),
			SoftwareTokenMfaConfiguration: expandCognitoSoftwareTokenMfaConfiguration(d.Get("software_token_mfa_configuration").([]interface{})),
			UserPoolId:                    aws.String(d.Id()),
		}

		// Since SMS configuration applies to both verification and MFA, only include if MFA is enabled.
		// Otherwise, the API will return the following error:
		// InvalidParameterException: Invalid MFA configuration given, can't turn off MFA and configure an MFA together.
		if v := d.Get("sms_configuration").([]interface{}); len(v) > 0 && v[0] != nil && mfaConfiguration != cognitoidentityprovider.UserPoolMfaTypeOff {
			input.SmsMfaConfiguration = &cognitoidentityprovider.SmsMfaConfigType{
				SmsConfiguration: expandCognitoSmsConfiguration(v),
			}

			if v, ok := d.GetOk("sms_authentication_message"); ok {
				input.SmsMfaConfiguration.SmsAuthenticationMessage = aws.String(v.(string))
			}
		}

		// IAM Roles and Policies can take some time to propagate
		err := resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
			_, err := conn.SetUserPoolMfaConfig(input)

			if isAWSErr(err, cognitoidentityprovider.ErrCodeInvalidSmsRoleTrustRelationshipException, "Role does not have a trust relationship allowing Cognito to assume the role") {
				return resource.RetryableError(err)
			}

			if isAWSErr(err, cognitoidentityprovider.ErrCodeInvalidSmsRoleAccessPolicyException, "Role does not have permission to publish with SNS") {
				return resource.RetryableError(err)
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return nil
		})

		if isResourceTimeoutError(err) {
			_, err = conn.SetUserPoolMfaConfig(input)
		}

		if err != nil {
			return fmt.Errorf("error setting Cognito User Pool (%s) MFA Configuration: %w", d.Id(), err)
		}
	}

	// Non Multi-Factor Authentication updates
	// NOTES:
	//  * Include SMS configuration changes since settings are shared between verification and MFA.
	//  * For backwards compatibility, include SMS authentication message changes without SMS MFA since the API allows it.
	if d.HasChanges(
		"admin_create_user_config",
		"auto_verified_attributes",
		"device_configuration",
		"email_configuration",
		"email_verification_message",
		"email_verification_subject",
		"lambda_config",
		"password_policy",
		"sms_authentication_message",
		"sms_configuration",
		"sms_verification_message",
		"tags",
		"tags_all",
		"user_pool_add_ons",
		"verification_message_template",
		"account_recovery_setting",
	) {
		params := &cognitoidentityprovider.UpdateUserPoolInput{
			UserPoolId: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("admin_create_user_config"); ok {
			configs := v.([]interface{})
			config, ok := configs[0].(map[string]interface{})

			if ok && config != nil {
				params.AdminCreateUserConfig = expandCognitoUserPoolAdminCreateUserConfig(config)
			}
		}

		if v, ok := d.GetOk("auto_verified_attributes"); ok {
			params.AutoVerifiedAttributes = expandStringSet(v.(*schema.Set))
		}

		if v, ok := d.GetOk("account_recovery_setting"); ok {
			configs := v.([]interface{})
			config, ok := configs[0].(map[string]interface{})

			if ok && config != nil {
				params.AccountRecoverySetting = expandCognitoUserPoolAccountRecoverySettingConfig(config)
			}
		}

		if v, ok := d.GetOk("device_configuration"); ok {
			configs := v.([]interface{})
			config, ok := configs[0].(map[string]interface{})

			if ok && config != nil {
				params.DeviceConfiguration = expandCognitoUserPoolDeviceConfiguration(config)
			}
		}

		if v, ok := d.GetOk("email_configuration"); ok && len(v.([]interface{})) > 0 {
			params.EmailConfiguration = expandCognitoUserPoolEmailConfig(v.([]interface{}))
		}

		if v, ok := d.GetOk("email_verification_subject"); ok {
			params.EmailVerificationSubject = aws.String(v.(string))
		}

		if v, ok := d.GetOk("email_verification_message"); ok {
			params.EmailVerificationMessage = aws.String(v.(string))
		}

		if v, ok := d.GetOk("lambda_config"); ok {
			configs := v.([]interface{})
			config, ok := configs[0].(map[string]interface{})

			if ok && config != nil {
				params.LambdaConfig = expandCognitoUserPoolLambdaConfig(config)
			}
		}

		if v, ok := d.GetOk("mfa_configuration"); ok {
			params.MfaConfiguration = aws.String(v.(string))
		}

		if v, ok := d.GetOk("password_policy"); ok {
			configs := v.([]interface{})
			config, ok := configs[0].(map[string]interface{})

			if ok && config != nil {
				policies := &cognitoidentityprovider.UserPoolPolicyType{}
				policies.PasswordPolicy = expandCognitoUserPoolPasswordPolicy(config)
				params.Policies = policies
			}
		}

		if v, ok := d.GetOk("sms_authentication_message"); ok {
			params.SmsAuthenticationMessage = aws.String(v.(string))
		}

		if v, ok := d.GetOk("sms_configuration"); ok {
			params.SmsConfiguration = expandCognitoSmsConfiguration(v.([]interface{}))
		}

		if v, ok := d.GetOk("user_pool_add_ons"); ok {
			configs := v.([]interface{})
			config, ok := configs[0].(map[string]interface{})

			if ok && config != nil {
				userPoolAddons := &cognitoidentityprovider.UserPoolAddOnsType{}

				if v, ok := config["advanced_security_mode"]; ok && v.(string) != "" {
					userPoolAddons.AdvancedSecurityMode = aws.String(v.(string))
				}
				params.UserPoolAddOns = userPoolAddons
			}
		}

		if v, ok := d.GetOk("verification_message_template"); ok {
			configs := v.([]interface{})
			config, ok := configs[0].(map[string]interface{})

			if d.HasChange("email_verification_message") {
				config["email_message"] = d.Get("email_verification_message")
			}
			if d.HasChange("email_verification_subject") {
				config["email_subject"] = d.Get("email_verification_subject")
			}
			if d.HasChange("sms_verification_message") {
				config["sms_message"] = d.Get("sms_verification_message")
			}

			if ok && config != nil {
				params.VerificationMessageTemplate = expandCognitoUserPoolVerificationMessageTemplate(config)
			}
		}

		if v, ok := d.GetOk("sms_verification_message"); ok {
			params.SmsVerificationMessage = aws.String(v.(string))
		}

		if len(tags) > 0 {
			params.UserPoolTags = tags.IgnoreAws().CognitoidentityproviderTags()
		}

		log.Printf("[DEBUG] Updating Cognito User Pool: %s", params)

		// IAM roles & policies can take some time to propagate and be attached
		// to the User Pool.
		err := resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
			var err error
			_, err = conn.UpdateUserPool(params)
			if isAWSErr(err, cognitoidentityprovider.ErrCodeInvalidSmsRoleTrustRelationshipException, "Role does not have a trust relationship allowing Cognito to assume the role") {
				log.Printf("[DEBUG] Received %s, retrying UpdateUserPool", err)
				return resource.RetryableError(err)
			}
			if isAWSErr(err, cognitoidentityprovider.ErrCodeInvalidSmsRoleAccessPolicyException, "Role does not have permission to publish with SNS") {
				log.Printf("[DEBUG] Received %s, retrying UpdateUserPool", err)
				return resource.RetryableError(err)
			}
			if isAWSErr(err, cognitoidentityprovider.ErrCodeInvalidParameterException, "Please use TemporaryPasswordValidityDays in PasswordPolicy instead of UnusedAccountValidityDays") {
				log.Printf("[DEBUG] Received %s, retrying UpdateUserPool without UnusedAccountValidityDays", err)
				params.AdminCreateUserConfig.UnusedAccountValidityDays = nil
				return resource.RetryableError(err)
			}
			if err != nil {
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if isResourceTimeoutError(err) {
			_, err = conn.UpdateUserPool(params)
		}
		if err != nil {
			return fmt.Errorf("error updating Cognito User pool (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("schema") {
		oldSchema, newSchema := d.GetChange("schema")
		if oldSchema.(*schema.Set).Difference(newSchema.(*schema.Set)).Len() == 0 {
			params := &cognitoidentityprovider.AddCustomAttributesInput{
				UserPoolId:       aws.String(d.Id()),
				CustomAttributes: expandCognitoUserPoolSchema(newSchema.(*schema.Set).Difference(oldSchema.(*schema.Set)).List()),
			}
			_, err := conn.AddCustomAttributes(params)
			if err != nil {
				return fmt.Errorf("error updating Cognito User Pool (%s): unable to add custom attributes from schema: %w", d.Id(), err)
			}
		} else {
			return fmt.Errorf("error updating Cognito User Pool (%s): cannot modify or remove schema items", d.Id())
		}
	}

	return resourceAwsCognitoUserPoolRead(d, meta)
}

func resourceAwsCognitoUserPoolDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.DeleteUserPoolInput{
		UserPoolId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Cognito User Pool: %s", params)

	_, err := conn.DeleteUserPool(params)

	if tfawserr.ErrCodeEquals(err, cognitoidentityprovider.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Cognito user pool (%s): %w", d.Id(), err)
	}

	return nil
}

func expandCognitoSmsConfiguration(tfList []interface{}) *cognitoidentityprovider.SmsConfigurationType {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	apiObject := &cognitoidentityprovider.SmsConfigurationType{}

	if v, ok := tfMap["external_id"].(string); ok && v != "" {
		apiObject.ExternalId = aws.String(v)
	}

	if v, ok := tfMap["sns_caller_arn"].(string); ok && v != "" {
		apiObject.SnsCallerArn = aws.String(v)
	}

	return apiObject
}

func expandCognitoSoftwareTokenMfaConfiguration(tfList []interface{}) *cognitoidentityprovider.SoftwareTokenMfaConfigType {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	apiObject := &cognitoidentityprovider.SoftwareTokenMfaConfigType{}

	if v, ok := tfMap["enabled"].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	return apiObject
}

func flattenCognitoSmsConfiguration(apiObject *cognitoidentityprovider.SmsConfigurationType) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ExternalId; v != nil {
		tfMap["external_id"] = aws.StringValue(v)
	}

	if v := apiObject.SnsCallerArn; v != nil {
		tfMap["sns_caller_arn"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

func flattenCognitoSoftwareTokenMfaConfiguration(apiObject *cognitoidentityprovider.SoftwareTokenMfaConfigType) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Enabled; v != nil {
		tfMap["enabled"] = aws.BoolValue(v)
	}

	return []interface{}{tfMap}
}

func expandCognitoUserPoolAccountRecoverySettingConfig(config map[string]interface{}) *cognitoidentityprovider.AccountRecoverySettingType {
	configs := &cognitoidentityprovider.AccountRecoverySettingType{}

	mechs := make([]*cognitoidentityprovider.RecoveryOptionType, 0)

	if v, ok := config["recovery_mechanism"]; ok {
		data := v.(*schema.Set).List()

		for _, m := range data {
			param := m.(map[string]interface{})
			opt := &cognitoidentityprovider.RecoveryOptionType{}

			if v, ok := param["name"]; ok {
				opt.Name = aws.String(v.(string))
			}

			if v, ok := param["priority"]; ok {
				opt.Priority = aws.Int64(int64(v.(int)))
			}

			mechs = append(mechs, opt)
		}
	}

	configs.RecoveryMechanisms = mechs

	return configs
}

func flattenCognitoUserPoolAccountRecoverySettingConfig(config *cognitoidentityprovider.AccountRecoverySettingType) []interface{} {
	if config == nil {
		return nil
	}

	settings := map[string]interface{}{}

	mechanisms := make([]map[string]interface{}, 0)

	for _, conf := range config.RecoveryMechanisms {
		mech := map[string]interface{}{
			"name":     aws.StringValue(conf.Name),
			"priority": aws.Int64Value(conf.Priority),
		}
		mechanisms = append(mechanisms, mech)
	}

	settings["recovery_mechanism"] = mechanisms

	return []interface{}{settings}
}

func flattenCognitoUserPoolEmailConfiguration(s *cognitoidentityprovider.EmailConfigurationType) []map[string]interface{} {
	m := make(map[string]interface{})

	if s == nil {
		return nil
	}

	if s.ReplyToEmailAddress != nil {
		m["reply_to_email_address"] = aws.StringValue(s.ReplyToEmailAddress)
	}

	if s.From != nil {
		m["from_email_address"] = aws.StringValue(s.From)
	}

	if s.SourceArn != nil {
		m["source_arn"] = aws.StringValue(s.SourceArn)
	}

	if s.EmailSendingAccount != nil {
		m["email_sending_account"] = aws.StringValue(s.EmailSendingAccount)
	}

	if len(m) > 0 {
		return []map[string]interface{}{m}
	}

	return []map[string]interface{}{}
}

func expandCognitoUserPoolAdminCreateUserConfig(config map[string]interface{}) *cognitoidentityprovider.AdminCreateUserConfigType {
	configs := &cognitoidentityprovider.AdminCreateUserConfigType{}

	if v, ok := config["allow_admin_create_user_only"]; ok {
		configs.AllowAdminCreateUserOnly = aws.Bool(v.(bool))
	}

	if v, ok := config["invite_message_template"]; ok {
		data := v.([]interface{})

		if len(data) > 0 {
			m, ok := data[0].(map[string]interface{})

			if ok {
				imt := &cognitoidentityprovider.MessageTemplateType{}

				if v, ok := m["email_message"]; ok {
					imt.EmailMessage = aws.String(v.(string))
				}

				if v, ok := m["email_subject"]; ok {
					imt.EmailSubject = aws.String(v.(string))
				}

				if v, ok := m["sms_message"]; ok {
					imt.SMSMessage = aws.String(v.(string))
				}

				configs.InviteMessageTemplate = imt
			}
		}
	}

	return configs
}

func flattenCognitoUserPoolAdminCreateUserConfig(s *cognitoidentityprovider.AdminCreateUserConfigType) []map[string]interface{} {
	config := map[string]interface{}{}

	if s == nil {
		return nil
	}

	if s.AllowAdminCreateUserOnly != nil {
		config["allow_admin_create_user_only"] = aws.BoolValue(s.AllowAdminCreateUserOnly)
	}

	if s.InviteMessageTemplate != nil {
		subconfig := map[string]interface{}{}

		if s.InviteMessageTemplate.EmailMessage != nil {
			subconfig["email_message"] = aws.StringValue(s.InviteMessageTemplate.EmailMessage)
		}

		if s.InviteMessageTemplate.EmailSubject != nil {
			subconfig["email_subject"] = aws.StringValue(s.InviteMessageTemplate.EmailSubject)
		}

		if s.InviteMessageTemplate.SMSMessage != nil {
			subconfig["sms_message"] = aws.StringValue(s.InviteMessageTemplate.SMSMessage)
		}

		if len(subconfig) > 0 {
			config["invite_message_template"] = []map[string]interface{}{subconfig}
		}
	}

	return []map[string]interface{}{config}
}

func expandCognitoUserPoolDeviceConfiguration(config map[string]interface{}) *cognitoidentityprovider.DeviceConfigurationType {
	configs := &cognitoidentityprovider.DeviceConfigurationType{}

	if v, ok := config["challenge_required_on_new_device"]; ok {
		configs.ChallengeRequiredOnNewDevice = aws.Bool(v.(bool))
	}

	if v, ok := config["device_only_remembered_on_user_prompt"]; ok {
		configs.DeviceOnlyRememberedOnUserPrompt = aws.Bool(v.(bool))
	}

	return configs
}

func expandCognitoUserPoolLambdaConfig(config map[string]interface{}) *cognitoidentityprovider.LambdaConfigType {
	configs := &cognitoidentityprovider.LambdaConfigType{}

	if v, ok := config["create_auth_challenge"]; ok && v.(string) != "" {
		configs.CreateAuthChallenge = aws.String(v.(string))
	}

	if v, ok := config["custom_message"]; ok && v.(string) != "" {
		configs.CustomMessage = aws.String(v.(string))
	}

	if v, ok := config["define_auth_challenge"]; ok && v.(string) != "" {
		configs.DefineAuthChallenge = aws.String(v.(string))
	}

	if v, ok := config["post_authentication"]; ok && v.(string) != "" {
		configs.PostAuthentication = aws.String(v.(string))
	}

	if v, ok := config["post_confirmation"]; ok && v.(string) != "" {
		configs.PostConfirmation = aws.String(v.(string))
	}

	if v, ok := config["pre_authentication"]; ok && v.(string) != "" {
		configs.PreAuthentication = aws.String(v.(string))
	}

	if v, ok := config["pre_sign_up"]; ok && v.(string) != "" {
		configs.PreSignUp = aws.String(v.(string))
	}

	if v, ok := config["pre_token_generation"]; ok && v.(string) != "" {
		configs.PreTokenGeneration = aws.String(v.(string))
	}

	if v, ok := config["user_migration"]; ok && v.(string) != "" {
		configs.UserMigration = aws.String(v.(string))
	}

	if v, ok := config["verify_auth_challenge_response"]; ok && v.(string) != "" {
		configs.VerifyAuthChallengeResponse = aws.String(v.(string))
	}

	if v, ok := config["kms_key_id"]; ok && v.(string) != "" {
		configs.KMSKeyID = aws.String(v.(string))
	}

	if v, ok := config["custom_sms_sender"].([]interface{}); ok && len(v) > 0 {
		s, sok := v[0].(map[string]interface{})
		if sok && s != nil {
			configs.CustomSMSSender = expandCognitoUserPoolCustomSMSSender(s)
		}
	}

	if v, ok := config["custom_email_sender"].([]interface{}); ok && len(v) > 0 {
		s, sok := v[0].(map[string]interface{})
		if sok && s != nil {
			configs.CustomEmailSender = expandCognitoUserPoolCustomEmailSender(s)
		}
	}

	return configs
}

func flattenCognitoUserPoolLambdaConfig(s *cognitoidentityprovider.LambdaConfigType) []map[string]interface{} {
	m := map[string]interface{}{}

	if s == nil {
		return nil
	}

	if s.CreateAuthChallenge != nil {
		m["create_auth_challenge"] = aws.StringValue(s.CreateAuthChallenge)
	}

	if s.CustomMessage != nil {
		m["custom_message"] = aws.StringValue(s.CustomMessage)
	}

	if s.DefineAuthChallenge != nil {
		m["define_auth_challenge"] = aws.StringValue(s.DefineAuthChallenge)
	}

	if s.PostAuthentication != nil {
		m["post_authentication"] = aws.StringValue(s.PostAuthentication)
	}

	if s.PostConfirmation != nil {
		m["post_confirmation"] = aws.StringValue(s.PostConfirmation)
	}

	if s.PreAuthentication != nil {
		m["pre_authentication"] = aws.StringValue(s.PreAuthentication)
	}

	if s.PreSignUp != nil {
		m["pre_sign_up"] = aws.StringValue(s.PreSignUp)
	}

	if s.PreTokenGeneration != nil {
		m["pre_token_generation"] = aws.StringValue(s.PreTokenGeneration)
	}

	if s.UserMigration != nil {
		m["user_migration"] = aws.StringValue(s.UserMigration)
	}

	if s.VerifyAuthChallengeResponse != nil {
		m["verify_auth_challenge_response"] = aws.StringValue(s.VerifyAuthChallengeResponse)
	}

	if s.KMSKeyID != nil {
		m["kms_key_id"] = aws.StringValue(s.KMSKeyID)
	}

	if s.CustomSMSSender != nil {
		m["custom_sms_sender"] = flattenCognitoUserPoolCustomSMSSender(s.CustomSMSSender)
	}

	if s.CustomEmailSender != nil {
		m["custom_email_sender"] = flattenCognitoUserPoolCustomEmailSender(s.CustomEmailSender)
	}

	if len(m) > 0 {
		return []map[string]interface{}{m}
	}

	return []map[string]interface{}{}
}

func expandCognitoUserPoolPasswordPolicy(config map[string]interface{}) *cognitoidentityprovider.PasswordPolicyType {
	configs := &cognitoidentityprovider.PasswordPolicyType{}

	if v, ok := config["minimum_length"]; ok {
		configs.MinimumLength = aws.Int64(int64(v.(int)))
	}

	if v, ok := config["require_lowercase"]; ok {
		configs.RequireLowercase = aws.Bool(v.(bool))
	}

	if v, ok := config["require_numbers"]; ok {
		configs.RequireNumbers = aws.Bool(v.(bool))
	}

	if v, ok := config["require_symbols"]; ok {
		configs.RequireSymbols = aws.Bool(v.(bool))
	}

	if v, ok := config["require_uppercase"]; ok {
		configs.RequireUppercase = aws.Bool(v.(bool))
	}

	if v, ok := config["temporary_password_validity_days"]; ok {
		configs.TemporaryPasswordValidityDays = aws.Int64(int64(v.(int)))
	}

	return configs
}

func flattenCognitoUserPoolUserPoolAddOns(s *cognitoidentityprovider.UserPoolAddOnsType) []map[string]interface{} {
	config := make(map[string]interface{})

	if s == nil {
		return []map[string]interface{}{}
	}

	if s.AdvancedSecurityMode != nil {
		config["advanced_security_mode"] = aws.StringValue(s.AdvancedSecurityMode)
	}

	return []map[string]interface{}{config}
}

func expandCognitoUserPoolSchema(inputs []interface{}) []*cognitoidentityprovider.SchemaAttributeType {
	configs := make([]*cognitoidentityprovider.SchemaAttributeType, len(inputs))

	for i, input := range inputs {
		param := input.(map[string]interface{})
		config := &cognitoidentityprovider.SchemaAttributeType{}

		if v, ok := param["attribute_data_type"]; ok {
			config.AttributeDataType = aws.String(v.(string))
		}

		if v, ok := param["developer_only_attribute"]; ok {
			config.DeveloperOnlyAttribute = aws.Bool(v.(bool))
		}

		if v, ok := param["mutable"]; ok {
			config.Mutable = aws.Bool(v.(bool))
		}

		if v, ok := param["name"]; ok {
			config.Name = aws.String(v.(string))
		}

		if v, ok := param["required"]; ok {
			config.Required = aws.Bool(v.(bool))
		}

		if v, ok := param["number_attribute_constraints"]; ok {
			data := v.([]interface{})

			if len(data) > 0 {
				m, ok := data[0].(map[string]interface{})
				if ok {
					numberAttributeConstraintsType := &cognitoidentityprovider.NumberAttributeConstraintsType{}

					if v, ok := m["min_value"]; ok && v.(string) != "" {
						numberAttributeConstraintsType.MinValue = aws.String(v.(string))
					}

					if v, ok := m["max_value"]; ok && v.(string) != "" {
						numberAttributeConstraintsType.MaxValue = aws.String(v.(string))
					}

					config.NumberAttributeConstraints = numberAttributeConstraintsType
				}
			}
		}

		if v, ok := param["string_attribute_constraints"]; ok {
			data := v.([]interface{})

			if len(data) > 0 {
				m, _ := data[0].(map[string]interface{})
				if ok {
					stringAttributeConstraintsType := &cognitoidentityprovider.StringAttributeConstraintsType{}

					if l, ok := m["min_length"]; ok && l.(string) != "" {
						stringAttributeConstraintsType.MinLength = aws.String(l.(string))
					}

					if l, ok := m["max_length"]; ok && l.(string) != "" {
						stringAttributeConstraintsType.MaxLength = aws.String(l.(string))
					}

					config.StringAttributeConstraints = stringAttributeConstraintsType
				}
			}
		}

		configs[i] = config
	}

	return configs
}

func flattenCognitoUserPoolSchema(configuredAttributes, inputs []*cognitoidentityprovider.SchemaAttributeType) []map[string]interface{} {
	values := make([]map[string]interface{}, 0)

	for _, input := range inputs {
		if input == nil {
			continue
		}

		// The API returns all standard attributes
		// https://docs.aws.amazon.com/cognito/latest/developerguide/user-pool-settings-attributes.html#cognito-user-pools-standard-attributes
		// Ignore setting them in state if they are unconfigured to prevent a huge and unexpected diff
		configured := false

		for _, configuredAttribute := range configuredAttributes {
			if reflect.DeepEqual(input, configuredAttribute) {
				configured = true
			}
		}

		if !configured {
			if cognitoUserPoolSchemaAttributeMatchesStandardAttribute(input) {
				continue
			}
			// When adding a Cognito Identity Provider, the API will automatically add an "identities" attribute
			identitiesAttribute := cognitoidentityprovider.SchemaAttributeType{
				AttributeDataType:          aws.String(cognitoidentityprovider.AttributeDataTypeString),
				DeveloperOnlyAttribute:     aws.Bool(false),
				Mutable:                    aws.Bool(true),
				Name:                       aws.String("identities"),
				Required:                   aws.Bool(false),
				StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{},
			}
			if reflect.DeepEqual(*input, identitiesAttribute) {
				continue
			}
		}

		var value = map[string]interface{}{
			"attribute_data_type":      aws.StringValue(input.AttributeDataType),
			"developer_only_attribute": aws.BoolValue(input.DeveloperOnlyAttribute),
			"mutable":                  aws.BoolValue(input.Mutable),
			"name":                     strings.TrimPrefix(strings.TrimPrefix(aws.StringValue(input.Name), "dev:"), "custom:"),
			"required":                 aws.BoolValue(input.Required),
		}

		if input.NumberAttributeConstraints != nil {
			subvalue := make(map[string]interface{})

			if input.NumberAttributeConstraints.MinValue != nil {
				subvalue["min_value"] = aws.StringValue(input.NumberAttributeConstraints.MinValue)
			}

			if input.NumberAttributeConstraints.MaxValue != nil {
				subvalue["max_value"] = aws.StringValue(input.NumberAttributeConstraints.MaxValue)
			}

			value["number_attribute_constraints"] = []map[string]interface{}{subvalue}
		}

		if input.StringAttributeConstraints != nil {
			subvalue := make(map[string]interface{})

			if input.StringAttributeConstraints.MinLength != nil {
				subvalue["min_length"] = aws.StringValue(input.StringAttributeConstraints.MinLength)
			}

			if input.StringAttributeConstraints.MaxLength != nil {
				subvalue["max_length"] = aws.StringValue(input.StringAttributeConstraints.MaxLength)
			}

			value["string_attribute_constraints"] = []map[string]interface{}{subvalue}
		}

		values = append(values, value)
	}

	return values
}

func expandCognitoUserPoolUsernameConfiguration(config map[string]interface{}) *cognitoidentityprovider.UsernameConfigurationType {
	usernameConfigurationType := &cognitoidentityprovider.UsernameConfigurationType{
		CaseSensitive: aws.Bool(config["case_sensitive"].(bool)),
	}

	return usernameConfigurationType
}

func flattenCognitoUserPoolUsernameConfiguration(u *cognitoidentityprovider.UsernameConfigurationType) []map[string]interface{} {
	m := map[string]interface{}{}

	if u == nil {
		return nil
	}

	m["case_sensitive"] = aws.BoolValue(u.CaseSensitive)

	return []map[string]interface{}{m}
}

func expandCognitoUserPoolVerificationMessageTemplate(config map[string]interface{}) *cognitoidentityprovider.VerificationMessageTemplateType {
	verificationMessageTemplateType := &cognitoidentityprovider.VerificationMessageTemplateType{}

	if v, ok := config["default_email_option"]; ok && v.(string) != "" {
		verificationMessageTemplateType.DefaultEmailOption = aws.String(v.(string))
	}

	if v, ok := config["email_message"]; ok && v.(string) != "" {
		verificationMessageTemplateType.EmailMessage = aws.String(v.(string))
	}

	if v, ok := config["email_message_by_link"]; ok && v.(string) != "" {
		verificationMessageTemplateType.EmailMessageByLink = aws.String(v.(string))
	}

	if v, ok := config["email_subject"]; ok && v.(string) != "" {
		verificationMessageTemplateType.EmailSubject = aws.String(v.(string))
	}

	if v, ok := config["email_subject_by_link"]; ok && v.(string) != "" {
		verificationMessageTemplateType.EmailSubjectByLink = aws.String(v.(string))
	}

	if v, ok := config["sms_message"]; ok && v.(string) != "" {
		verificationMessageTemplateType.SmsMessage = aws.String(v.(string))
	}

	return verificationMessageTemplateType
}

func flattenCognitoUserPoolVerificationMessageTemplate(s *cognitoidentityprovider.VerificationMessageTemplateType) []map[string]interface{} {
	m := map[string]interface{}{}

	if s == nil {
		return nil
	}

	if s.DefaultEmailOption != nil {
		m["default_email_option"] = aws.StringValue(s.DefaultEmailOption)
	}

	if s.EmailMessage != nil {
		m["email_message"] = aws.StringValue(s.EmailMessage)
	}

	if s.EmailMessageByLink != nil {
		m["email_message_by_link"] = aws.StringValue(s.EmailMessageByLink)
	}

	if s.EmailSubject != nil {
		m["email_subject"] = aws.StringValue(s.EmailSubject)
	}

	if s.EmailSubjectByLink != nil {
		m["email_subject_by_link"] = aws.StringValue(s.EmailSubjectByLink)
	}

	if s.SmsMessage != nil {
		m["sms_message"] = aws.StringValue(s.SmsMessage)
	}

	if len(m) > 0 {
		return []map[string]interface{}{m}
	}

	return []map[string]interface{}{}
}

func flattenCognitoUserPoolDeviceConfiguration(s *cognitoidentityprovider.DeviceConfigurationType) []map[string]interface{} {
	config := map[string]interface{}{}

	if s == nil {
		return nil
	}

	if s.ChallengeRequiredOnNewDevice != nil {
		config["challenge_required_on_new_device"] = aws.BoolValue(s.ChallengeRequiredOnNewDevice)
	}

	if s.DeviceOnlyRememberedOnUserPrompt != nil {
		config["device_only_remembered_on_user_prompt"] = aws.BoolValue(s.DeviceOnlyRememberedOnUserPrompt)
	}

	return []map[string]interface{}{config}
}

func flattenCognitoUserPoolPasswordPolicy(s *cognitoidentityprovider.PasswordPolicyType) []map[string]interface{} {
	m := map[string]interface{}{}

	if s == nil {
		return nil
	}

	if s.MinimumLength != nil {
		m["minimum_length"] = aws.Int64Value(s.MinimumLength)
	}

	if s.RequireLowercase != nil {
		m["require_lowercase"] = aws.BoolValue(s.RequireLowercase)
	}

	if s.RequireNumbers != nil {
		m["require_numbers"] = aws.BoolValue(s.RequireNumbers)
	}

	if s.RequireSymbols != nil {
		m["require_symbols"] = aws.BoolValue(s.RequireSymbols)
	}

	if s.RequireUppercase != nil {
		m["require_uppercase"] = aws.BoolValue(s.RequireUppercase)
	}

	if s.TemporaryPasswordValidityDays != nil {
		m["temporary_password_validity_days"] = aws.Int64Value(s.TemporaryPasswordValidityDays)
	}

	if len(m) > 0 {
		return []map[string]interface{}{m}
	}

	return []map[string]interface{}{}
}

func cognitoUserPoolSchemaAttributeMatchesStandardAttribute(input *cognitoidentityprovider.SchemaAttributeType) bool {
	if input == nil {
		return false
	}

	// All standard attributes always returned by API
	// https://docs.aws.amazon.com/cognito/latest/developerguide/user-pool-settings-attributes.html#cognito-user-pools-standard-attributes
	var standardAttributes = []cognitoidentityprovider.SchemaAttributeType{
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("address"),
			Required:               aws.Bool(false),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("2048"),
				MinLength: aws.String("0"),
			},
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("birthdate"),
			Required:               aws.Bool(false),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("10"),
				MinLength: aws.String("10"),
			},
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("email"),
			Required:               aws.Bool(false),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("2048"),
				MinLength: aws.String("0"),
			},
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeBoolean),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("email_verified"),
			Required:               aws.Bool(false),
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("gender"),
			Required:               aws.Bool(false),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("2048"),
				MinLength: aws.String("0"),
			},
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("given_name"),
			Required:               aws.Bool(false),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("2048"),
				MinLength: aws.String("0"),
			},
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("family_name"),
			Required:               aws.Bool(false),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("2048"),
				MinLength: aws.String("0"),
			},
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("locale"),
			Required:               aws.Bool(false),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("2048"),
				MinLength: aws.String("0"),
			},
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("middle_name"),
			Required:               aws.Bool(false),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("2048"),
				MinLength: aws.String("0"),
			},
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("name"),
			Required:               aws.Bool(false),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("2048"),
				MinLength: aws.String("0"),
			},
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("nickname"),
			Required:               aws.Bool(false),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("2048"),
				MinLength: aws.String("0"),
			},
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("phone_number"),
			Required:               aws.Bool(false),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("2048"),
				MinLength: aws.String("0"),
			},
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeBoolean),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("phone_number_verified"),
			Required:               aws.Bool(false),
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("picture"),
			Required:               aws.Bool(false),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("2048"),
				MinLength: aws.String("0"),
			},
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("preferred_username"),
			Required:               aws.Bool(false),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("2048"),
				MinLength: aws.String("0"),
			},
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("profile"),
			Required:               aws.Bool(false),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("2048"),
				MinLength: aws.String("0"),
			},
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(false),
			Name:                   aws.String("sub"),
			Required:               aws.Bool(true),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("2048"),
				MinLength: aws.String("1"),
			},
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeNumber),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("updated_at"),
			NumberAttributeConstraints: &cognitoidentityprovider.NumberAttributeConstraintsType{
				MinValue: aws.String("0"),
			},
			Required: aws.Bool(false),
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("website"),
			Required:               aws.Bool(false),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("2048"),
				MinLength: aws.String("0"),
			},
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("zoneinfo"),
			Required:               aws.Bool(false),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("2048"),
				MinLength: aws.String("0"),
			},
		},
	}
	for _, standardAttribute := range standardAttributes {
		if reflect.DeepEqual(*input, standardAttribute) {
			return true
		}
	}
	return false
}

func expandCognitoUserPoolCustomSMSSender(config map[string]interface{}) *cognitoidentityprovider.CustomSMSLambdaVersionConfigType {
	usernameConfigurationType := &cognitoidentityprovider.CustomSMSLambdaVersionConfigType{
		LambdaArn:     aws.String(config["lambda_arn"].(string)),
		LambdaVersion: aws.String(config["lambda_version"].(string)),
	}

	return usernameConfigurationType
}

func flattenCognitoUserPoolCustomSMSSender(u *cognitoidentityprovider.CustomSMSLambdaVersionConfigType) []map[string]interface{} {
	m := map[string]interface{}{}

	if u == nil {
		return nil
	}

	m["lambda_arn"] = aws.StringValue(u.LambdaArn)
	m["lambda_version"] = aws.StringValue(u.LambdaVersion)

	return []map[string]interface{}{m}
}

func expandCognitoUserPoolCustomEmailSender(config map[string]interface{}) *cognitoidentityprovider.CustomEmailLambdaVersionConfigType {
	usernameConfigurationType := &cognitoidentityprovider.CustomEmailLambdaVersionConfigType{
		LambdaArn:     aws.String(config["lambda_arn"].(string)),
		LambdaVersion: aws.String(config["lambda_version"].(string)),
	}

	return usernameConfigurationType
}

func flattenCognitoUserPoolCustomEmailSender(u *cognitoidentityprovider.CustomEmailLambdaVersionConfigType) []map[string]interface{} {
	m := map[string]interface{}{}

	if u == nil {
		return nil
	}

	m["lambda_arn"] = aws.StringValue(u.LambdaArn)
	m["lambda_version"] = aws.StringValue(u.LambdaVersion)

	return []map[string]interface{}{m}
}

func expandCognitoUserPoolEmailConfig(emailConfig []interface{}) *cognitoidentityprovider.EmailConfigurationType {
	config := emailConfig[0].(map[string]interface{})

	emailConfigurationType := &cognitoidentityprovider.EmailConfigurationType{}

	if v, ok := config["reply_to_email_address"]; ok && v.(string) != "" {
		emailConfigurationType.ReplyToEmailAddress = aws.String(v.(string))
	}

	if v, ok := config["source_arn"]; ok && v.(string) != "" {
		emailConfigurationType.SourceArn = aws.String(v.(string))
	}

	if v, ok := config["from_email_address"]; ok && v.(string) != "" {
		emailConfigurationType.From = aws.String(v.(string))
	}

	if v, ok := config["email_sending_account"]; ok && v.(string) != "" {
		emailConfigurationType.EmailSendingAccount = aws.String(v.(string))
	}

	if v, ok := config["configuration_set"]; ok && v.(string) != "" {
		emailConfigurationType.ConfigurationSet = aws.String(v.(string))
	}

	return emailConfigurationType
}
