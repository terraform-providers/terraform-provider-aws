package aws

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/amplify"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	tfamplify "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/amplify"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/amplify/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsAmplifyApp() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAmplifyAppCreate,
		Read:   resourceAwsAmplifyAppRead,
		Update: resourceAwsAmplifyAppUpdate,
		Delete: resourceAwsAmplifyAppDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: customdiff.Sequence(
			SetTagsDiff,
			customdiff.ForceNewIfChange("description", func(_ context.Context, old, new, meta interface{}) bool {
				// Any existing value cannot be cleared.
				return new.(string) == ""
			}),
			customdiff.ForceNewIfChange("iam_service_role_arn", func(_ context.Context, old, new, meta interface{}) bool {
				// Any existing value cannot be cleared.
				return new.(string) == ""
			}),
		),

		Schema: map[string]*schema.Schema{
			"access_token": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"auto_branch_creation_config": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"basic_auth_credentials": {
							Type:         schema.TypeString,
							Optional:     true,
							Sensitive:    true,
							ValidateFunc: validation.StringLenBetween(1, 2000),
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// These credentials are ignored if basic auth is not enabled.
								if d.Get("auto_branch_creation_config.0.enable_basic_auth").(bool) {
									return old == new
								}

								return true
							},
						},

						"build_spec": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 25000),
						},

						"enable_auto_build": {
							Type:     schema.TypeBool,
							Optional: true,
						},

						"enable_basic_auth": {
							Type:     schema.TypeBool,
							Optional: true,
						},

						"enable_performance_mode": {
							Type:     schema.TypeBool,
							Optional: true,
						},

						"enable_pull_request_preview": {
							Type:     schema.TypeBool,
							Optional: true,
						},

						"environment_variables": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},

						"framework": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},

						"pull_request_environment_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},

						"stage": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(amplify.Stage_Values(), false),
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// API returns "NONE" by default.
								if old == tfamplify.StageNone && new == "" {
									return true
								}

								return old == new
							},
						},
					},
				},
			},

			"auto_branch_creation_patterns": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// These patterns are ignored if branch auto-creation is not enabled.
					if d.Get("enable_auto_branch_creation").(bool) {
						return old == new
					}

					return true
				},
			},

			"basic_auth_credentials": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(1, 2000),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// These credentials are ignored if basic auth is not enabled.
					if d.Get("enable_basic_auth").(bool) {
						return old == new
					}

					return true
				},
			},

			"build_spec": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 25000),
			},

			"custom_rule": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"condition": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 2048),
						},

						"source": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 2048),
						},

						"status": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"200",
								"301",
								"302",
								"404",
								"404-200",
							}, false),
						},

						"target": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 2048),
						},
					},
				},
			},

			"default_domain": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},

			"enable_auto_branch_creation": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"enable_basic_auth": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"enable_branch_auto_build": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"enable_branch_auto_deletion": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"environment_variables": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"iam_service_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},

			"oauth_token": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},

			"platform": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      amplify.PlatformWeb,
				ValidateFunc: validation.StringInSlice(amplify.Platform_Values(), false),
			},

			"production_branch": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"branch_name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"last_deploy_time": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"thumbnail_url": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"repository": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},

			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},
	}
}

func resourceAwsAmplifyAppCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).amplifyconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)

	input := &amplify.CreateAppInput{
		Name: aws.String(name),
	}

	if v, ok := d.GetOk("access_token"); ok {
		input.AccessToken = aws.String(v.(string))
	}

	if v, ok := d.GetOk("auto_branch_creation_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.AutoBranchCreationConfig = expandAmplifyAutoBranchCreationConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("auto_branch_creation_patterns"); ok && v.(*schema.Set).Len() > 0 {
		input.AutoBranchCreationPatterns = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("basic_auth_credentials"); ok {
		input.BasicAuthCredentials = aws.String(v.(string))
	}

	if v, ok := d.GetOk("build_spec"); ok {
		input.BuildSpec = aws.String(v.(string))
	}

	if v, ok := d.GetOk("custom_rule"); ok && len(v.([]interface{})) > 0 {
		input.CustomRules = expandAmplifyCustomRules(v.([]interface{}))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("enable_auto_branch_creation"); ok {
		input.EnableAutoBranchCreation = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("enable_basic_auth"); ok {
		input.EnableBasicAuth = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("enable_branch_auto_build"); ok {
		input.EnableBranchAutoBuild = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("enable_branch_auto_deletion"); ok {
		input.EnableBranchAutoDeletion = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("environment_variables"); ok && len(v.(map[string]interface{})) > 0 {
		input.EnvironmentVariables = expandStringMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("iam_service_role_arn"); ok {
		input.IamServiceRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("oauth_token"); ok {
		input.OauthToken = aws.String(v.(string))
	}

	if v, ok := d.GetOk("platform"); ok {
		input.Platform = aws.String(v.(string))
	}

	if v, ok := d.GetOk("repository"); ok {
		input.Repository = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().AmplifyTags()
	}

	log.Printf("[DEBUG] Creating Amplify App: %s", input)
	output, err := conn.CreateApp(input)

	if err != nil {
		return fmt.Errorf("error creating Amplify App (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.App.AppId))

	return resourceAwsAmplifyAppRead(d, meta)
}

func resourceAwsAmplifyAppRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).amplifyconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	app, err := finder.AppByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Amplify App (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Amplify App (%s): %w", d.Id(), err)
	}

	d.Set("arn", app.AppArn)
	if app.AutoBranchCreationConfig != nil {
		if err := d.Set("auto_branch_creation_config", []interface{}{flattenAmplifyAutoBranchCreationConfig(app.AutoBranchCreationConfig)}); err != nil {
			return fmt.Errorf("error setting auto_branch_creation_config: %w", err)
		}
	} else {
		d.Set("auto_branch_creation_config", nil)
	}
	d.Set("auto_branch_creation_patterns", aws.StringValueSlice(app.AutoBranchCreationPatterns))
	d.Set("basic_auth_credentials", app.BasicAuthCredentials)
	d.Set("build_spec", app.BuildSpec)
	if err := d.Set("custom_rule", flattenAmplifyCustomRules(app.CustomRules)); err != nil {
		return fmt.Errorf("error setting custom_rule: %w", err)
	}
	d.Set("default_domain", app.DefaultDomain)
	d.Set("description", app.Description)
	d.Set("enable_auto_branch_creation", app.EnableAutoBranchCreation)
	d.Set("enable_basic_auth", app.EnableBasicAuth)
	d.Set("enable_branch_auto_build", app.EnableBranchAutoBuild)
	d.Set("enable_branch_auto_deletion", app.EnableBranchAutoDeletion)
	d.Set("environment_variables", aws.StringValueMap(app.EnvironmentVariables))
	d.Set("iam_service_role_arn", app.IamServiceRoleArn)
	d.Set("name", app.Name)
	d.Set("platform", app.Platform)
	if app.ProductionBranch != nil {
		if err := d.Set("production_branch", []interface{}{flattenAmplifyProductionBranch(app.ProductionBranch)}); err != nil {
			return fmt.Errorf("error setting production_branch: %w", err)
		}
	} else {
		d.Set("production_branch", nil)
	}
	d.Set("repository", app.Repository)

	tags := keyvaluetags.AmplifyKeyValueTags(app.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsAmplifyAppUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).amplifyconn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &amplify.UpdateAppInput{
			AppId: aws.String(d.Id()),
		}

		if d.HasChange("access_token") {
			input.AccessToken = aws.String(d.Get("access_token").(string))
		}

		if d.HasChange("auto_branch_creation_config") {
			input.AutoBranchCreationConfig = expandAmplifyAutoBranchCreationConfig(d.Get("auto_branch_creation_config").([]interface{})[0].(map[string]interface{}))

			if d.HasChange("auto_branch_creation_config.0.environment_variables") {
				if v := d.Get("auto_branch_creation_config.0.environment_variables").(map[string]interface{}); len(v) == 0 {
					input.AutoBranchCreationConfig.EnvironmentVariables = aws.StringMap(map[string]string{"": ""})
				}
			}
		}

		if d.HasChange("auto_branch_creation_patterns") {
			input.AutoBranchCreationPatterns = expandStringSet(d.Get("auto_branch_creation_patterns").(*schema.Set))
		}

		if d.HasChange("basic_auth_credentials") {
			input.BasicAuthCredentials = aws.String(d.Get("basic_auth_credentials").(string))
		}

		if d.HasChange("build_spec") {
			input.BuildSpec = aws.String(d.Get("build_spec").(string))
		}

		if d.HasChange("custom_rule") {
			if v := d.Get("custom_rule").([]interface{}); len(v) > 0 {
				input.CustomRules = expandAmplifyCustomRules(v)
			} else {
				input.CustomRules = []*amplify.CustomRule{}
			}
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChange("enable_auto_branch_creation") {
			input.EnableAutoBranchCreation = aws.Bool(d.Get("enable_auto_branch_creation").(bool))
		}

		if d.HasChange("enable_basic_auth") {
			input.EnableBasicAuth = aws.Bool(d.Get("enable_basic_auth").(bool))
		}

		if d.HasChange("enable_branch_auto_build") {
			input.EnableBranchAutoBuild = aws.Bool(d.Get("enable_branch_auto_build").(bool))
		}

		if d.HasChange("enable_branch_auto_deletion") {
			input.EnableBranchAutoDeletion = aws.Bool(d.Get("enable_branch_auto_deletion").(bool))
		}

		if d.HasChange("environment_variables") {
			if v := d.Get("environment_variables").(map[string]interface{}); len(v) > 0 {
				input.EnvironmentVariables = expandStringMap(v)
			} else {
				input.EnvironmentVariables = aws.StringMap(map[string]string{"": ""})
			}
		}

		if d.HasChange("iam_service_role_arn") {
			input.IamServiceRoleArn = aws.String(d.Get("iam_service_role_arn").(string))
		}

		if d.HasChange("name") {
			input.Name = aws.String(d.Get("name").(string))
		}

		if d.HasChange("oauth_token") {
			input.OauthToken = aws.String(d.Get("oauth_token").(string))
		}

		if d.HasChange("platform") {
			input.Platform = aws.String(d.Get("platform").(string))
		}

		if d.HasChange("repository") {
			input.Repository = aws.String(d.Get("repository").(string))
		}

		_, err := conn.UpdateApp(input)

		if err != nil {
			return fmt.Errorf("error updating Amplify App (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := keyvaluetags.AmplifyUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceAwsAmplifyAppRead(d, meta)
}

func resourceAwsAmplifyAppDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).amplifyconn

	log.Printf("[DEBUG] Deleting Amplify App (%s)", d.Id())
	_, err := conn.DeleteApp(&amplify.DeleteAppInput{
		AppId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, amplify.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Amplify App (%s): %w", d.Id(), err)
	}

	return nil
}

func expandAmplifyAutoBranchCreationConfig(tfMap map[string]interface{}) *amplify.AutoBranchCreationConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &amplify.AutoBranchCreationConfig{}

	if v, ok := tfMap["basic_auth_credentials"].(string); ok && v != "" {
		apiObject.BasicAuthCredentials = aws.String(v)
	}

	if v, ok := tfMap["build_spec"].(string); ok && v != "" {
		apiObject.BuildSpec = aws.String(v)
	}

	if v, ok := tfMap["enable_auto_build"].(bool); ok {
		apiObject.EnableAutoBuild = aws.Bool(v)
	}

	if v, ok := tfMap["enable_basic_auth"].(bool); ok {
		apiObject.EnableBasicAuth = aws.Bool(v)
	}

	if v, ok := tfMap["enable_performance_mode"].(bool); ok {
		apiObject.EnablePerformanceMode = aws.Bool(v)
	}

	if v, ok := tfMap["enable_pull_request_preview"].(bool); ok {
		apiObject.EnablePullRequestPreview = aws.Bool(v)
	}

	if v, ok := tfMap["environment_variables"].(map[string]interface{}); ok && len(v) > 0 {
		apiObject.EnvironmentVariables = expandStringMap(v)
	}

	if v, ok := tfMap["framework"].(string); ok && v != "" {
		apiObject.Framework = aws.String(v)
	}

	if v, ok := tfMap["pull_request_environment_name"].(string); ok && v != "" {
		apiObject.PullRequestEnvironmentName = aws.String(v)
	}

	if v, ok := tfMap["stage"].(string); ok && v != "" && v != tfamplify.StageNone {
		apiObject.Stage = aws.String(v)
	}

	return apiObject
}

func flattenAmplifyAutoBranchCreationConfig(apiObject *amplify.AutoBranchCreationConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.BasicAuthCredentials; v != nil {
		tfMap["basic_auth_credentials"] = aws.StringValue(v)
	}

	if v := apiObject.BuildSpec; v != nil {
		tfMap["build_spec"] = aws.StringValue(v)
	}

	if v := apiObject.EnableAutoBuild; v != nil {
		tfMap["enable_auto_build"] = aws.BoolValue(v)
	}

	if v := apiObject.EnableBasicAuth; v != nil {
		tfMap["enable_basic_auth"] = aws.BoolValue(v)
	}

	if v := apiObject.EnablePerformanceMode; v != nil {
		tfMap["enable_performance_mode"] = aws.BoolValue(v)
	}

	if v := apiObject.EnablePullRequestPreview; v != nil {
		tfMap["enable_pull_request_preview"] = aws.BoolValue(v)
	}

	if v := apiObject.EnvironmentVariables; v != nil {
		tfMap["environment_variables"] = aws.StringValueMap(v)
	}

	if v := apiObject.Framework; v != nil {
		tfMap["framework"] = aws.StringValue(v)
	}

	if v := apiObject.PullRequestEnvironmentName; v != nil {
		tfMap["pull_request_environment_name"] = aws.StringValue(v)
	}

	if v := apiObject.Stage; v != nil {
		tfMap["stage"] = aws.StringValue(v)
	}

	return tfMap
}

func expandAmplifyCustomRule(tfMap map[string]interface{}) *amplify.CustomRule {
	if tfMap == nil {
		return nil
	}

	apiObject := &amplify.CustomRule{}

	if v, ok := tfMap["condition"].(string); ok && v != "" {
		apiObject.Condition = aws.String(v)
	}

	if v, ok := tfMap["source"].(string); ok && v != "" {
		apiObject.Source = aws.String(v)
	}

	if v, ok := tfMap["status"].(string); ok && v != "" {
		apiObject.Status = aws.String(v)
	}

	if v, ok := tfMap["target"].(string); ok && v != "" {
		apiObject.Target = aws.String(v)
	}

	return apiObject
}

func expandAmplifyCustomRules(tfList []interface{}) []*amplify.CustomRule {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*amplify.CustomRule

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandAmplifyCustomRule(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenAmplifyCustomRule(apiObject *amplify.CustomRule) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Condition; v != nil {
		tfMap["condition"] = aws.StringValue(v)
	}

	if v := apiObject.Source; v != nil {
		tfMap["source"] = aws.StringValue(v)
	}

	if v := apiObject.Status; v != nil {
		tfMap["status"] = aws.StringValue(v)
	}

	if v := apiObject.Target; v != nil {
		tfMap["target"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenAmplifyCustomRules(apiObjects []*amplify.CustomRule) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenAmplifyCustomRule(apiObject))
	}

	return tfList
}

func flattenAmplifyProductionBranch(apiObject *amplify.ProductionBranch) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.BranchName; v != nil {
		tfMap["branch_name"] = aws.StringValue(v)
	}

	if v := apiObject.LastDeployTime; v != nil {
		tfMap["last_deploy_time"] = aws.TimeValue(v).Format(time.RFC3339)
	}

	if v := apiObject.Status; v != nil {
		tfMap["status"] = aws.StringValue(v)
	}

	if v := apiObject.ThumbnailUrl; v != nil {
		tfMap["thumbnail_url"] = aws.StringValue(v)
	}

	return tfMap
}
