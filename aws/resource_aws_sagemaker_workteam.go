package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker/finder"
)

func resourceAwsSagemakerWorkteam() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSagemakerWorkteamCreate,
		Read:   resourceAwsSagemakerWorkteamRead,
		Update: resourceAwsSagemakerWorkteamUpdate,
		Delete: resourceAwsSagemakerWorkteamDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 200),
			},
			"member_definition": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cognito_member_definition": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"client_id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"user_group": {
										Type:     schema.TypeString,
										Required: true,
									},
									"user_pool": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"oidc_member_definition": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"groups": {
										Type:     schema.TypeSet,
										MaxItems: 10,
										Required: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 63),
										},
									},
								},
							},
						},
					},
				},
			},
			"notification_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"notification_topic_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
						},
					},
				},
				DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
			},
			"subdomain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workforce_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workforce_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"workteam_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9])*$`), "Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
		},
	}
}

func resourceAwsSagemakerWorkteamCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	name := d.Get("workteam_name").(string)
	input := &sagemaker.CreateWorkteamInput{
		WorkteamName:      aws.String(name),
		WorkforceName:     aws.String(d.Get("workforce_name").(string)),
		Description:       aws.String(d.Get("description").(string)),
		MemberDefinitions: expandSagemakerWorkteamMemberDefinition(d.Get("member_definition").([]interface{})),
	}

	if v, ok := d.GetOk("notification_configuration"); ok {
		input.NotificationConfiguration = expandSagemakerWorkteamNotificationConfiguration(v.([]interface{}))
	}

	log.Printf("[DEBUG] Sagemaker Workteam create config: %#v", *input)
	_, err := retryOnAwsCode("ValidationException", func() (interface{}, error) {
		return conn.CreateWorkteam(input)
	})
	if err != nil {
		return fmt.Errorf("error creating SageMaker Workteam: %w", err)
	}

	d.SetId(name)

	return resourceAwsSagemakerWorkteamRead(d, meta)
}

func resourceAwsSagemakerWorkteamRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	workteam, err := finder.WorkteamByName(conn, d.Id())
	if err != nil {
		if isAWSErr(err, "ValidationException", " The work team") {
			d.SetId("")
			log.Printf("[WARN] Unable to find SageMaker workteam (%s); removing from state", d.Id())
			return nil
		}
		return fmt.Errorf("error reading SageMaker workteam (%s): %w", d.Id(), err)

	}

	arn := aws.StringValue(workteam.WorkteamArn)
	d.Set("arn", arn)
	d.Set("subdomain", workteam.SubDomain)
	d.Set("description", workteam.Description)

	// d.Set("workforce_arn", workteam.WorkforceArn)
	d.Set("workteam_name", workteam.WorkteamName)
	// d.Set("workforce_name", workteam.WorkforceName)

	if err := d.Set("member_definition", flattenSagemakerWorkteamMemberDefinition(workteam.MemberDefinitions)); err != nil {
		return fmt.Errorf("error setting member_definition for Sagemaker Workteam (%s): %w", d.Id(), err)
	}

	if err := d.Set("notification_configuration", flattenSagemakerWorkteamNotificationConfiguration(workteam.NotificationConfiguration)); err != nil {
		return fmt.Errorf("error setting notification_configuration for Sagemaker Workteam (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceAwsSagemakerWorkteamUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	input := &sagemaker.UpdateWorkteamInput{
		WorkteamName:      aws.String(d.Id()),
		MemberDefinitions: expandSagemakerWorkteamMemberDefinition(d.Get("member_definition").([]interface{})),
	}

	if d.HasChange("description") {
		input.Description = aws.String(d.Get("description").(string))
	}

	if d.HasChange("notification_configuration") {
		input.NotificationConfiguration = expandSagemakerWorkteamNotificationConfiguration(d.Get("notification_configuration").([]interface{}))
	}

	log.Printf("[DEBUG] Sagemaker Workteam update config: %#v", *input)
	_, err := conn.UpdateWorkteam(input)
	if err != nil {
		return fmt.Errorf("error updating SageMaker Workteam: %w", err)
	}

	return resourceAwsSagemakerWorkteamRead(d, meta)
}

func resourceAwsSagemakerWorkteamDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	input := &sagemaker.DeleteWorkteamInput{
		WorkteamName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteWorkteam(input); err != nil {
		if isAWSErr(err, "ValidationException", "No workteam found for account") {
			return nil
		}
		return fmt.Errorf("error deleting SageMaker workteam (%s): %w", d.Id(), err)
	}

	return nil
}

func expandSagemakerWorkteamMemberDefinition(l []interface{}) []*sagemaker.MemberDefinition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	var members []*sagemaker.MemberDefinition

	for _, mem := range l {
		memRaw := mem.(map[string]interface{})
		member := &sagemaker.MemberDefinition{}

		if v, ok := memRaw["cognito_member_definition"].([]interface{}); ok && len(v) > 0 {
			member.CognitoMemberDefinition = expandSagemakerWorkteamCognitoMemberDefinition(v)
		}

		if v, ok := memRaw["oidc_member_definition"].([]interface{}); ok && len(v) > 0 {
			member.OidcMemberDefinition = expandSagemakerWorkteamOidcMemberDefinition(v)
		}

		members = append(members, member)
	}

	return members
}

func flattenSagemakerWorkteamMemberDefinition(config []*sagemaker.MemberDefinition) []map[string]interface{} {
	members := make([]map[string]interface{}, 0, len(config))

	for _, raw := range config {
		member := make(map[string]interface{})

		if raw.CognitoMemberDefinition != nil {
			member["cognito_member_definition"] = flattenSagemakerWorkteamCognitoMemberDefinition(raw.CognitoMemberDefinition)
		}

		if raw.OidcMemberDefinition != nil {
			member["oidc_member_definition"] = flattenSagemakerWorkteamOidcMemberDefinition(raw.OidcMemberDefinition)
		}

		members = append(members, member)
	}

	return members
}

func expandSagemakerWorkteamCognitoMemberDefinition(l []interface{}) *sagemaker.CognitoMemberDefinition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.CognitoMemberDefinition{
		ClientId:  aws.String(m["client_id"].(string)),
		UserPool:  aws.String(m["user_pool"].(string)),
		UserGroup: aws.String(m["user_group"].(string)),
	}

	return config
}

func flattenSagemakerWorkteamCognitoMemberDefinition(config *sagemaker.CognitoMemberDefinition) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"client_id":  aws.StringValue(config.ClientId),
		"user_pool":  aws.StringValue(config.UserPool),
		"user_group": aws.StringValue(config.UserGroup),
	}

	return []map[string]interface{}{m}
}

func expandSagemakerWorkteamOidcMemberDefinition(l []interface{}) *sagemaker.OidcMemberDefinition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.OidcMemberDefinition{
		Groups: expandStringSet(m["groups"].(*schema.Set)),
	}

	return config
}

func flattenSagemakerWorkteamOidcMemberDefinition(config *sagemaker.OidcMemberDefinition) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"groups": flattenStringSet(config.Groups),
	}

	return []map[string]interface{}{m}
}

func expandSagemakerWorkteamNotificationConfiguration(l []interface{}) *sagemaker.NotificationConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.NotificationConfiguration{
		NotificationTopicArn: aws.String(m["notification_topic_arn"].(string)),
	}

	return config
}

func flattenSagemakerWorkteamNotificationConfiguration(config *sagemaker.NotificationConfiguration) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"notification_topic_arn": aws.StringValue(config.NotificationTopicArn),
	}

	return []map[string]interface{}{m}
}
