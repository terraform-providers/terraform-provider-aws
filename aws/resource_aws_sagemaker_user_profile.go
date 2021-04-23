package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker/waiter"
)

func resourceAwsSagemakerUserProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSagemakerUserProfileCreate,
		Read:   resourceAwsSagemakerUserProfileRead,
		Update: resourceAwsSagemakerUserProfileUpdate,
		Delete: resourceAwsSagemakerUserProfileDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_profile_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9]){0,62}`), "Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
			"domain_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"single_sign_on_user_identifier": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"single_sign_on_user_value": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"user_settings": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_groups": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 5,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"execution_role": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
						"sharing_settings": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"notebook_output_option": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      sagemaker.NotebookOutputOptionDisabled,
										ValidateFunc: validation.StringInSlice(sagemaker.NotebookOutputOption_Values(), false),
									},
									"s3_kms_key_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateArn,
									},
									"s3_output_path": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"tensor_board_app_settings": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"default_resource_spec": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"instance_type": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(sagemaker.AppInstanceType_Values(), false),
												},
												"sagemaker_image_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateArn,
												},
											},
										},
									},
								},
							},
						},
						"jupyter_server_app_settings": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"default_resource_spec": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"instance_type": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(sagemaker.AppInstanceType_Values(), false),
												},
												"sagemaker_image_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateArn,
												},
											},
										},
									},
								},
							},
						},
						"kernel_gateway_app_settings": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"default_resource_spec": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"instance_type": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(sagemaker.AppInstanceType_Values(), false),
												},
												"sagemaker_image_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateArn,
												},
											},
										},
									},
									"custom_image": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 30,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"app_image_config_name": {
													Type:     schema.TypeString,
													Required: true,
												},
												"image_name": {
													Type:     schema.TypeString,
													Required: true,
												},
												"image_version_number": {
													Type:     schema.TypeInt,
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"tags": tagsSchema(),
			"home_efs_file_system_uid": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsSagemakerUserProfileCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	input := &sagemaker.CreateUserProfileInput{
		UserProfileName: aws.String(d.Get("user_profile_name").(string)),
		DomainId:        aws.String(d.Get("domain_id").(string)),
	}

	if v, ok := d.GetOk("user_settings"); ok {
		input.UserSettings = expandSagemakerDomainDefaultUserSettings(v.([]interface{}))
	}

	if v, ok := d.GetOk("single_sign_on_user_identifier"); ok {
		input.SingleSignOnUserIdentifier = aws.String(v.(string))
	}

	if v, ok := d.GetOk("single_sign_on_user_value"); ok {
		input.SingleSignOnUserValue = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tags"); ok {
		input.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().SagemakerTags()
	}

	log.Printf("[DEBUG] SageMaker User Profile create config: %#v", *input)
	output, err := conn.CreateUserProfile(input)
	if err != nil {
		return fmt.Errorf("error creating SageMaker User Profile: %w", err)
	}

	userProfileArn := aws.StringValue(output.UserProfileArn)
	domainID, userProfileName, err := decodeSagemakerUserProfileName(userProfileArn)
	if err != nil {
		return err
	}

	d.SetId(userProfileArn)

	if _, err := waiter.UserProfileInService(conn, domainID, userProfileName); err != nil {
		return fmt.Errorf("error waiting for SageMaker User Profile (%s) to create: %w", d.Id(), err)
	}

	return resourceAwsSagemakerUserProfileRead(d, meta)
}

func resourceAwsSagemakerUserProfileRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	domainID, userProfileName, err := decodeSagemakerUserProfileName(d.Id())
	if err != nil {
		return err
	}

	UserProfile, err := finder.UserProfileByName(conn, domainID, userProfileName)
	if err != nil {
		if isAWSErr(err, sagemaker.ErrCodeResourceNotFound, "") {
			d.SetId("")
			log.Printf("[WARN] Unable to find SageMaker User Profile (%s), removing from state", d.Id())
			return nil
		}
		return fmt.Errorf("error reading SageMaker User Profile (%s): %w", d.Id(), err)
	}

	arn := aws.StringValue(UserProfile.UserProfileArn)
	d.Set("user_profile_name", UserProfile.UserProfileName)
	d.Set("domain_id", UserProfile.DomainId)
	d.Set("single_sign_on_user_identifier", UserProfile.SingleSignOnUserIdentifier)
	d.Set("single_sign_on_user_value", UserProfile.SingleSignOnUserValue)
	d.Set("arn", arn)
	d.Set("home_efs_file_system_uid", UserProfile.HomeEfsFileSystemUid)

	if err := d.Set("user_settings", flattenSagemakerDomainDefaultUserSettings(UserProfile.UserSettings)); err != nil {
		return fmt.Errorf("error setting user_settings for SageMaker User Profile (%s): %w", d.Id(), err)
	}

	tags, err := keyvaluetags.SagemakerListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for SageMaker User Profile (%s): %w", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceAwsSagemakerUserProfileUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	if d.HasChange("user_settings") {
		domainID := d.Get("domain_id").(string)
		userProfileName := d.Get("user_profile_name").(string)

		input := &sagemaker.UpdateUserProfileInput{
			UserProfileName: aws.String(userProfileName),
			DomainId:        aws.String(domainID),
			UserSettings:    expandSagemakerDomainDefaultUserSettings(d.Get("user_settings").([]interface{})),
		}

		log.Printf("[DEBUG] SageMaker User Profile update config: %#v", *input)
		_, err := conn.UpdateUserProfile(input)
		if err != nil {
			return fmt.Errorf("error updating SageMaker User Profile: %w", err)
		}

		if _, err := waiter.UserProfileInService(conn, domainID, userProfileName); err != nil {
			return fmt.Errorf("error waiting for SageMaker User Profile (%s) to update: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.SagemakerUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating SageMaker UserProfile (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAwsSagemakerUserProfileRead(d, meta)
}

func resourceAwsSagemakerUserProfileDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	userProfileName := d.Get("user_profile_name").(string)
	domainID := d.Get("domain_id").(string)

	input := &sagemaker.DeleteUserProfileInput{
		UserProfileName: aws.String(userProfileName),
		DomainId:        aws.String(domainID),
	}

	if _, err := conn.DeleteUserProfile(input); err != nil {
		if !isAWSErr(err, sagemaker.ErrCodeResourceNotFound, "") {
			return fmt.Errorf("error deleting SageMaker User Profile (%s): %w", d.Id(), err)
		}
	}

	if _, err := waiter.UserProfileDeleted(conn, domainID, userProfileName); err != nil {
		if !isAWSErr(err, sagemaker.ErrCodeResourceNotFound, "") {
			return fmt.Errorf("error waiting for SageMaker User Profile (%s) to delete: %w", d.Id(), err)
		}
	}

	return nil
}

func decodeSagemakerUserProfileName(id string) (string, string, error) {
	userProfileARN, err := arn.Parse(id)
	if err != nil {
		return "", "", err
	}

	userProfileResourceNameName := strings.TrimPrefix(userProfileARN.Resource, "user-profile/")
	parts := strings.Split(userProfileResourceNameName, "/")

	if len(parts) != 2 {
		return "", "", fmt.Errorf("Unexpected format of ID (%q), expected DOMAIN-ID/USER-PROFILE-NAME", userProfileResourceNameName)
	}

	domainID := parts[0]
	userProfileName := parts[1]

	return domainID, userProfileName, nil
}
