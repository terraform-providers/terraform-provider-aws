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

func resourceAwsSagemakerWorkforce() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSagemakerWorkforceCreate,
		Read:   resourceAwsSagemakerWorkforceRead,
		Update: resourceAwsSagemakerWorkforceUpdate,
		Delete: resourceAwsSagemakerWorkforceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cognito_config": {
				Type:         schema.TypeList,
				Optional:     true,
				ForceNew:     true,
				MaxItems:     1,
				ExactlyOneOf: []string{"oidc_config", "cognito_config"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"client_id": {
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
			"oidc_config": {
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				ExactlyOneOf: []string{"oidc_config", "cognito_config"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"authorization_endpoint": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 500),
								validation.IsURLWithHTTPS,
							),
						},
						"client_id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
						"client_secret": {
							Type:         schema.TypeString,
							Required:     true,
							Sensitive:    true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
						"issuer": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 500),
								validation.IsURLWithHTTPS,
							)},
						"jwks_uri": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 500),
								validation.IsURLWithHTTPS,
							)},
						"logout_endpoint": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 500),
								validation.IsURLWithHTTPS,
							)},
						"token_endpoint": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 500),
								validation.IsURLWithHTTPS,
							)},
						"user_info_endpoint": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 500),
								validation.IsURLWithHTTPS,
							),
						},
					},
				},
			},
			"source_ip_config": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cidrs": {
							Type:     schema.TypeSet,
							Required: true,
							MaxItems: 10,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.IsCIDR,
							},
						},
					},
				},
			},
			"subdomain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workforce_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-])*$`), "Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
		},
	}
}

func resourceAwsSagemakerWorkforceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	name := d.Get("workforce_name").(string)
	input := &sagemaker.CreateWorkforceInput{
		WorkforceName: aws.String(name),
	}

	if v, ok := d.GetOk("cognito_config"); ok {
		input.CognitoConfig = expandSagemakerWorkforceCognitoConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("oidc_config"); ok {
		input.OidcConfig = expandSagemakerWorkforceOidcConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("source_ip_config"); ok {
		input.SourceIpConfig = expandSagemakerWorkforceSourceIpConfig(v.([]interface{}))
	}

	log.Printf("[DEBUG] Sagemaker Workforce create config: %#v", *input)
	_, err := conn.CreateWorkforce(input)
	if err != nil {
		return fmt.Errorf("error creating SageMaker Workforce: %w", err)
	}

	d.SetId(name)

	return resourceAwsSagemakerWorkforceRead(d, meta)
}

func resourceAwsSagemakerWorkforceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	workforce, err := finder.WorkforceByName(conn, d.Id())
	if err != nil {
		if isAWSErr(err, "ValidationException", "No workforce") {
			d.SetId("")
			log.Printf("[WARN] Unable to find SageMaker workforce (%s); removing from state", d.Id())
			return nil
		}
		return fmt.Errorf("error reading SageMaker workforce (%s): %w", d.Id(), err)

	}

	arn := aws.StringValue(workforce.WorkforceArn)
	d.Set("arn", arn)
	d.Set("subdomain", workforce.SubDomain)
	d.Set("workforce_name", workforce.WorkforceName)

	if err := d.Set("cognito_config", flattenSagemakerWorkforceCognitoConfig(workforce.CognitoConfig)); err != nil {
		return fmt.Errorf("error setting cognito_config for Sagemaker Workforce (%s): %w", d.Id(), err)
	}

	if workforce.OidcConfig != nil {
		if err := d.Set("oidc_config", flattenSagemakerWorkforceOidcConfig(workforce.OidcConfig, d.Get("oidc_config.0.client_secret").(string))); err != nil {
			return fmt.Errorf("error setting oidc_config for Sagemaker Workforce (%s): %w", d.Id(), err)
		}
	}

	if err := d.Set("source_ip_config", flattenSagemakerWorkforceSourceIpConfig(workforce.SourceIpConfig)); err != nil {
		return fmt.Errorf("error setting source_ip_config for Sagemaker Workforce (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceAwsSagemakerWorkforceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	input := &sagemaker.UpdateWorkforceInput{
		WorkforceName: aws.String(d.Id()),
	}

	if d.HasChange("source_ip_config") {
		input.SourceIpConfig = expandSagemakerWorkforceSourceIpConfig(d.Get("source_ip_config").([]interface{}))
	}

	if d.HasChange("oidc_config") {
		input.OidcConfig = expandSagemakerWorkforceOidcConfig(d.Get("oidc_config").([]interface{}))
	}

	log.Printf("[DEBUG] Sagemaker Workforce update config: %#v", *input)
	_, err := conn.UpdateWorkforce(input)
	if err != nil {
		return fmt.Errorf("error updating SageMaker Workforce: %w", err)
	}

	return resourceAwsSagemakerWorkforceRead(d, meta)
}

func resourceAwsSagemakerWorkforceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	input := &sagemaker.DeleteWorkforceInput{
		WorkforceName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteWorkforce(input); err != nil {
		if isAWSErr(err, "ValidationException", "No workforce found for account") {
			return nil
		}
		return fmt.Errorf("error deleting SageMaker workforce (%s): %w", d.Id(), err)
	}

	return nil
}

func expandSagemakerWorkforceSourceIpConfig(l []interface{}) *sagemaker.SourceIpConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.SourceIpConfig{
		Cidrs: expandStringSet(m["cidrs"].(*schema.Set)),
	}

	return config
}

func flattenSagemakerWorkforceSourceIpConfig(config *sagemaker.SourceIpConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"cidrs": flattenStringSet(config.Cidrs),
	}

	return []map[string]interface{}{m}
}

func expandSagemakerWorkforceCognitoConfig(l []interface{}) *sagemaker.CognitoConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.CognitoConfig{
		ClientId: aws.String(m["client_id"].(string)),
		UserPool: aws.String(m["user_pool"].(string)),
	}

	return config
}

func flattenSagemakerWorkforceCognitoConfig(config *sagemaker.CognitoConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"client_id": aws.StringValue(config.ClientId),
		"user_pool": aws.StringValue(config.UserPool),
	}

	return []map[string]interface{}{m}
}

func expandSagemakerWorkforceOidcConfig(l []interface{}) *sagemaker.OidcConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.OidcConfig{
		AuthorizationEndpoint: aws.String(m["authorization_endpoint"].(string)),
		ClientId:              aws.String(m["client_id"].(string)),
		ClientSecret:          aws.String(m["client_secret"].(string)),
		Issuer:                aws.String(m["issuer"].(string)),
		JwksUri:               aws.String(m["jwks_uri"].(string)),
		LogoutEndpoint:        aws.String(m["logout_endpoint"].(string)),
		TokenEndpoint:         aws.String(m["token_endpoint"].(string)),
		UserInfoEndpoint:      aws.String(m["user_info_endpoint"].(string)),
	}

	return config
}

func flattenSagemakerWorkforceOidcConfig(config *sagemaker.OidcConfigForResponse, clientSecret string) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"authorization_endpoint": aws.StringValue(config.AuthorizationEndpoint),
		"client_id":              aws.StringValue(config.ClientId),
		"client_secret":          clientSecret,
		"issuer":                 aws.StringValue(config.Issuer),
		"jwks_uri":               aws.StringValue(config.JwksUri),
		"logout_endpoint":        aws.StringValue(config.LogoutEndpoint),
		"token_endpoint":         aws.StringValue(config.TokenEndpoint),
		"user_info_endpoint":     aws.StringValue(config.UserInfoEndpoint),
	}

	return []map[string]interface{}{m}
}
