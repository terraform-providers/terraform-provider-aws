package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

var validAppsyncAuthTypes = []string{
	appsync.AuthenticationTypeApiKey,
	appsync.AuthenticationTypeAwsIam,
	appsync.AuthenticationTypeAmazonCognitoUserPools,
	appsync.AuthenticationTypeOpenidConnect,
}

func resourceAwsAppsyncGraphqlApi() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAppsyncGraphqlApiCreate,
		Read:   resourceAwsAppsyncGraphqlApiRead,
		Update: resourceAwsAppsyncGraphqlApiUpdate,
		Delete: resourceAwsAppsyncGraphqlApiDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"additional_authentication_provider": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"authentication_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(validAppsyncAuthTypes, false),
						},
						"openid_connect_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"auth_ttl": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"client_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"iat_ttl": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"issuer": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"user_pool_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"app_id_client_regex": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"aws_region": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"user_pool_id": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"authentication_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(validAppsyncAuthTypes, false),
			},
			"schema": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if !regexp.MustCompile(`[_A-Za-z][_0-9A-Za-z]*`).MatchString(value) {
						errors = append(errors, fmt.Errorf("%q must match [_A-Za-z][_0-9A-Za-z]*", k))
					}
					return
				},
			},
			"log_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cloudwatch_logs_role_arn": {
							Type:     schema.TypeString,
							Required: true,
						},
						"field_log_level": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								appsync.FieldLogLevelAll,
								appsync.FieldLogLevelError,
								appsync.FieldLogLevelNone,
							}, false),
						},
						"exclude_verbose_content": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"openid_connect_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auth_ttl": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"client_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"iat_ttl": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"issuer": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"user_pool_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"app_id_client_regex": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"aws_region": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"default_action": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								appsync.DefaultActionAllow,
								appsync.DefaultActionDeny,
							}, false),
						},
						"user_pool_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"uris": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tagsSchema(),
			"xray_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceAwsAppsyncGraphqlApiCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn

	input := &appsync.CreateGraphqlApiInput{
		AuthenticationType: aws.String(d.Get("authentication_type").(string)),
		Name:               aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("log_config"); ok {
		input.LogConfig = expandAppsyncGraphqlApiLogConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("openid_connect_config"); ok {
		input.OpenIDConnectConfig = expandAppsyncGraphqlApiOpenIDConnectConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("user_pool_config"); ok {
		input.UserPoolConfig = expandAppsyncGraphqlApiUserPoolConfig(v.([]interface{}), meta.(*AWSClient).region)
	}

	if v, ok := d.GetOk("additional_authentication_provider"); ok {
		input.AdditionalAuthenticationProviders = expandAppsyncGraphqlApiAdditionalAuthProviders(v.([]interface{}), meta.(*AWSClient).region)
	}

	if v, ok := d.GetOk("tags"); ok {
		input.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().AppsyncTags()
	}

	if v, ok := d.GetOk("xray_enabled"); ok {
		input.XrayEnabled = aws.Bool(v.(bool))
	}

	resp, err := conn.CreateGraphqlApi(input)
	if err != nil {
		return fmt.Errorf("error creating AppSync GraphQL API: %s", err)
	}

	d.SetId(*resp.GraphqlApi.ApiId)

	if err := resourceAwsAppsyncSchemaPut(d, meta); err != nil {
		return fmt.Errorf("error creating AppSync GraphQL API (%s) Schema: %s", d.Id(), err)
	}

	return resourceAwsAppsyncGraphqlApiRead(d, meta)
}

func resourceAwsAppsyncGraphqlApiRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &appsync.GetGraphqlApiInput{
		ApiId: aws.String(d.Id()),
	}

	resp, err := conn.GetGraphqlApi(input)

	if isAWSErr(err, appsync.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] No such entity found for Appsync Graphql API (%s)", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting AppSync GraphQL API (%s): %s", d.Id(), err)
	}

	d.Set("arn", resp.GraphqlApi.Arn)
	d.Set("authentication_type", resp.GraphqlApi.AuthenticationType)
	d.Set("name", resp.GraphqlApi.Name)

	if err := d.Set("log_config", flattenAppsyncGraphqlApiLogConfig(resp.GraphqlApi.LogConfig)); err != nil {
		return fmt.Errorf("error setting log_config: %s", err)
	}

	if err := d.Set("openid_connect_config", flattenAppsyncGraphqlApiOpenIDConnectConfig(resp.GraphqlApi.OpenIDConnectConfig)); err != nil {
		return fmt.Errorf("error setting openid_connect_config: %s", err)
	}

	if err := d.Set("user_pool_config", flattenAppsyncGraphqlApiUserPoolConfig(resp.GraphqlApi.UserPoolConfig)); err != nil {
		return fmt.Errorf("error setting user_pool_config: %s", err)
	}

	if err := d.Set("additional_authentication_provider", flattenAppsyncGraphqlApiAdditionalAuthenticationProviders(resp.GraphqlApi.AdditionalAuthenticationProviders)); err != nil {
		return fmt.Errorf("error setting additional_authentication_provider: %s", err)
	}

	if err := d.Set("uris", aws.StringValueMap(resp.GraphqlApi.Uris)); err != nil {
		return fmt.Errorf("error setting uris: %s", err)
	}

	if err := d.Set("tags", keyvaluetags.AppsyncKeyValueTags(resp.GraphqlApi.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	if err := d.Set("xray_enabled", aws.BoolValue(resp.GraphqlApi.XrayEnabled)); err != nil {
		return fmt.Errorf("error setting xray_enabled: %s", err)
	}

	return nil
}

func resourceAwsAppsyncGraphqlApiUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.AppsyncUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating AppSync GraphQL API (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	input := &appsync.UpdateGraphqlApiInput{
		ApiId:              aws.String(d.Id()),
		AuthenticationType: aws.String(d.Get("authentication_type").(string)),
		Name:               aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("log_config"); ok {
		input.LogConfig = expandAppsyncGraphqlApiLogConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("openid_connect_config"); ok {
		input.OpenIDConnectConfig = expandAppsyncGraphqlApiOpenIDConnectConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("user_pool_config"); ok {
		input.UserPoolConfig = expandAppsyncGraphqlApiUserPoolConfig(v.([]interface{}), meta.(*AWSClient).region)
	}

	if v, ok := d.GetOk("additional_authentication_provider"); ok {
		input.AdditionalAuthenticationProviders = expandAppsyncGraphqlApiAdditionalAuthProviders(v.([]interface{}), meta.(*AWSClient).region)
	}

	if v, ok := d.GetOk("xray_enabled"); ok {
		input.XrayEnabled = aws.Bool(v.(bool))
	}

	_, err := conn.UpdateGraphqlApi(input)
	if err != nil {
		return fmt.Errorf("error updating AppSync GraphQL API (%s): %s", d.Id(), err)
	}

	if d.HasChange("schema") {
		if err := resourceAwsAppsyncSchemaPut(d, meta); err != nil {
			return fmt.Errorf("error updating AppSync GraphQL API (%s) Schema: %s", d.Id(), err)
		}
	}

	return resourceAwsAppsyncGraphqlApiRead(d, meta)
}

func resourceAwsAppsyncGraphqlApiDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn

	input := &appsync.DeleteGraphqlApiInput{
		ApiId: aws.String(d.Id()),
	}
	_, err := conn.DeleteGraphqlApi(input)

	if isAWSErr(err, appsync.ErrCodeNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting AppSync GraphQL API (%s): %s", d.Id(), err)
	}

	return nil
}

func expandAppsyncGraphqlApiLogConfig(l []interface{}) *appsync.LogConfig {
	if len(l) < 1 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	logConfig := &appsync.LogConfig{
		CloudWatchLogsRoleArn: aws.String(m["cloudwatch_logs_role_arn"].(string)),
		FieldLogLevel:         aws.String(m["field_log_level"].(string)),
		ExcludeVerboseContent: aws.Bool(m["exclude_verbose_content"].(bool)),
	}

	return logConfig
}

func expandAppsyncGraphqlApiOpenIDConnectConfig(l []interface{}) *appsync.OpenIDConnectConfig {
	if len(l) < 1 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	openIDConnectConfig := &appsync.OpenIDConnectConfig{
		Issuer: aws.String(m["issuer"].(string)),
	}

	if v, ok := m["auth_ttl"].(int); ok && v != 0 {
		openIDConnectConfig.AuthTTL = aws.Int64(int64(v))
	}

	if v, ok := m["client_id"].(string); ok && v != "" {
		openIDConnectConfig.ClientId = aws.String(v)
	}

	if v, ok := m["iat_ttl"].(int); ok && v != 0 {
		openIDConnectConfig.IatTTL = aws.Int64(int64(v))
	}

	return openIDConnectConfig
}

func expandAppsyncGraphqlApiUserPoolConfig(l []interface{}, currentRegion string) *appsync.UserPoolConfig {
	if len(l) < 1 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	userPoolConfig := &appsync.UserPoolConfig{
		AwsRegion:     aws.String(currentRegion),
		DefaultAction: aws.String(m["default_action"].(string)),
		UserPoolId:    aws.String(m["user_pool_id"].(string)),
	}

	if v, ok := m["app_id_client_regex"].(string); ok && v != "" {
		userPoolConfig.AppIdClientRegex = aws.String(v)
	}

	if v, ok := m["aws_region"].(string); ok && v != "" {
		userPoolConfig.AwsRegion = aws.String(v)
	}

	return userPoolConfig
}

func expandAppsyncGraphqlApiAdditionalAuthProviders(items []interface{}, currentRegion string) []*appsync.AdditionalAuthenticationProvider {
	if len(items) < 1 {
		return nil
	}

	additionalAuthProviders := make([]*appsync.AdditionalAuthenticationProvider, 0, len(items))
	for _, l := range items {
		if l == nil {
			continue
		}

		m := l.(map[string]interface{})
		additionalAuthProvider := &appsync.AdditionalAuthenticationProvider{
			AuthenticationType: aws.String(m["authentication_type"].(string)),
		}

		if v, ok := m["openid_connect_config"]; ok {
			additionalAuthProvider.OpenIDConnectConfig = expandAppsyncGraphqlApiOpenIDConnectConfig(v.([]interface{}))
		}

		if v, ok := m["user_pool_config"]; ok {
			additionalAuthProvider.UserPoolConfig = expandAppsyncGraphqlApiCognitoUserPoolConfig(v.([]interface{}), currentRegion)
		}

		additionalAuthProviders = append(additionalAuthProviders, additionalAuthProvider)
	}

	return additionalAuthProviders
}

func expandAppsyncGraphqlApiCognitoUserPoolConfig(l []interface{}, currentRegion string) *appsync.CognitoUserPoolConfig {
	if len(l) < 1 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	userPoolConfig := &appsync.CognitoUserPoolConfig{
		AwsRegion:  aws.String(currentRegion),
		UserPoolId: aws.String(m["user_pool_id"].(string)),
	}

	if v, ok := m["app_id_client_regex"].(string); ok && v != "" {
		userPoolConfig.AppIdClientRegex = aws.String(v)
	}

	if v, ok := m["aws_region"].(string); ok && v != "" {
		userPoolConfig.AwsRegion = aws.String(v)
	}

	return userPoolConfig
}

func flattenAppsyncGraphqlApiLogConfig(logConfig *appsync.LogConfig) []interface{} {
	if logConfig == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"cloudwatch_logs_role_arn": aws.StringValue(logConfig.CloudWatchLogsRoleArn),
		"field_log_level":          aws.StringValue(logConfig.FieldLogLevel),
		"exclude_verbose_content":  aws.BoolValue(logConfig.ExcludeVerboseContent),
	}

	return []interface{}{m}
}

func flattenAppsyncGraphqlApiOpenIDConnectConfig(openIDConnectConfig *appsync.OpenIDConnectConfig) []interface{} {
	if openIDConnectConfig == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"auth_ttl":  aws.Int64Value(openIDConnectConfig.AuthTTL),
		"client_id": aws.StringValue(openIDConnectConfig.ClientId),
		"iat_ttl":   aws.Int64Value(openIDConnectConfig.IatTTL),
		"issuer":    aws.StringValue(openIDConnectConfig.Issuer),
	}

	return []interface{}{m}
}

func flattenAppsyncGraphqlApiUserPoolConfig(userPoolConfig *appsync.UserPoolConfig) []interface{} {
	if userPoolConfig == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"aws_region":     aws.StringValue(userPoolConfig.AwsRegion),
		"default_action": aws.StringValue(userPoolConfig.DefaultAction),
		"user_pool_id":   aws.StringValue(userPoolConfig.UserPoolId),
	}

	if userPoolConfig.AppIdClientRegex != nil {
		m["app_id_client_regex"] = aws.StringValue(userPoolConfig.AppIdClientRegex)
	}

	return []interface{}{m}
}

func flattenAppsyncGraphqlApiAdditionalAuthenticationProviders(additionalAuthenticationProviders []*appsync.AdditionalAuthenticationProvider) []interface{} {
	if len(additionalAuthenticationProviders) == 0 {
		return []interface{}{}
	}

	result := make([]interface{}, len(additionalAuthenticationProviders))
	for i, provider := range additionalAuthenticationProviders {
		result[i] = map[string]interface{}{
			"authentication_type":   aws.StringValue(provider.AuthenticationType),
			"openid_connect_config": flattenAppsyncGraphqlApiOpenIDConnectConfig(provider.OpenIDConnectConfig),
			"user_pool_config":      flattenAppsyncGraphqlApiCognitoUserPoolConfig(provider.UserPoolConfig),
		}
	}

	return result
}

func flattenAppsyncGraphqlApiCognitoUserPoolConfig(userPoolConfig *appsync.CognitoUserPoolConfig) []interface{} {
	if userPoolConfig == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"aws_region":   aws.StringValue(userPoolConfig.AwsRegion),
		"user_pool_id": aws.StringValue(userPoolConfig.UserPoolId),
	}

	if userPoolConfig.AppIdClientRegex != nil {
		m["app_id_client_regex"] = aws.StringValue(userPoolConfig.AppIdClientRegex)
	}

	return []interface{}{m}
}

func resourceAwsAppsyncSchemaPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn

	if v, ok := d.GetOk("schema"); ok {
		input := &appsync.StartSchemaCreationInput{
			ApiId:      aws.String(d.Id()),
			Definition: ([]byte)(v.(string)),
		}
		if _, err := conn.StartSchemaCreation(input); err != nil {
			return err
		}

		activeSchemaConfig := &resource.StateChangeConf{
			Pending: []string{appsync.SchemaStatusProcessing},
			Target:  []string{"SUCCESS", appsync.SchemaStatusActive}, // should be only appsync.SchemaStatusActive . I think this is a problem in documentation: https://docs.aws.amazon.com/appsync/latest/APIReference/API_GetSchemaCreationStatus.html
			Refresh: func() (interface{}, string, error) {
				result, err := conn.GetSchemaCreationStatus(&appsync.GetSchemaCreationStatusInput{
					ApiId: aws.String(d.Id()),
				})
				if err != nil {
					return 0, "", err
				}
				return result, *result.Status, nil
			},
			Timeout: d.Timeout(schema.TimeoutCreate),
		}

		if _, err := activeSchemaConfig.WaitForState(); err != nil {
			return fmt.Errorf("Error waiting for schema creation status on AppSync API %s: %s", d.Id(), err)
		}
	}

	return nil
}
