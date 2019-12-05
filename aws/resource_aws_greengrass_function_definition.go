package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/greengrass"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func generateFunctionExecutionConfigSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"isolation_mode": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"run_as": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"gid": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"uid": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func generateFunctionEnviornmentConfigurationSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"access_sysfs": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"variables": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"execution": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem:     generateFunctionExecutionConfigSchema(),
			},
			"resource_access_policy": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"permission": {
							Type:     schema.TypeString,
							Required: true,
						},
						"resource_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func generateFunctionConfigurationSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"encoding_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "json",
			},
			"exec_args": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"executable": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"memory_size": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"pinned": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"timeout": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"environment": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem:     generateFunctionEnviornmentConfigurationSchema(),
			},
		},
	}
}

func resourceAwsGreengrassFunctionDefinition() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGreengrassFunctionDefinitionCreate,
		Read:   resourceAwsGreengrassFunctionDefinitionRead,
		Update: resourceAwsGreengrassFunctionDefinitionUpdate,
		Delete: resourceAwsGreengrassFunctionDefinitionDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"amzn_client_token": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"tags": tagsSchema(),
			"latest_definition_version_arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"function_definition_version": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"default_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem:     generateFunctionExecutionConfigSchema(),
						},
						"function": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"function_arn": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateArn,
									},
									"id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"function_configuration": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem:     generateFunctionConfigurationSchema(),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func parseRunAs(rawRunAsConfig map[string]interface{}) *greengrass.FunctionRunAsConfig {
	runAsConfig := &greengrass.FunctionRunAsConfig{}

	if v, ok := rawRunAsConfig["gid"]; ok {
		runAsConfig.Gid = aws.Int64(int64(v.(int)))
	}

	if v, ok := rawRunAsConfig["uid"]; ok {
		runAsConfig.Uid = aws.Int64(int64(v.(int)))
	}

	return runAsConfig
}

func parseFunctionExecutionConfiguration(rawExecConfig map[string]interface{}) *greengrass.FunctionExecutionConfig {
	execConfig := &greengrass.FunctionExecutionConfig{}

	if v, ok := rawExecConfig["isolation_mode"]; ok {
		execConfig.IsolationMode = aws.String(v.(string))
	}

	if v := rawExecConfig["run_as"].([]interface{}); len(v) > 0 {
		execConfig.RunAs = parseRunAs(v[0].(map[string]interface{}))
	}

	return execConfig
}

func parseFunctionDefaultExecutionConfiguration(rawExecConfig map[string]interface{}) *greengrass.FunctionDefaultExecutionConfig {
	execConfig := &greengrass.FunctionDefaultExecutionConfig{}

	if v, ok := rawExecConfig["isolation_mode"]; ok {
		execConfig.IsolationMode = aws.String(v.(string))
	}

	if v := rawExecConfig["run_as"].([]interface{}); len(v) > 0 {
		execConfig.RunAs = parseRunAs(v[0].(map[string]interface{}))
	}

	return execConfig
}

func parseResourseAccessPolicy(rawAccessPolicy map[string]interface{}) *greengrass.ResourceAccessPolicy {
	accessPolicy := &greengrass.ResourceAccessPolicy{
		Permission: aws.String(rawAccessPolicy["permission"].(string)),
		ResourceId: aws.String(rawAccessPolicy["resource_id"].(string)),
	}
	return accessPolicy
}

func parseFunctionEnvironmentConfigurationSchema(rawEnvConfig map[string]interface{}) *greengrass.FunctionConfigurationEnvironment {
	envConfig := &greengrass.FunctionConfigurationEnvironment{
		AccessSysfs: aws.Bool(rawEnvConfig["access_sysfs"].(bool)),
	}

	if rawVars, ok := rawEnvConfig["variables"]; ok {
		variableList := make(map[string]*string)
		for key, value := range rawVars.(map[string]interface{}) {
			variableList[key] = aws.String(value.(string))
		}
		envConfig.Variables = variableList
	}

	if v := rawEnvConfig["execution"].([]interface{}); len(v) > 0 {
		envConfig.Execution = parseFunctionExecutionConfiguration(v[0].(map[string]interface{}))
	}

	accessPolicyList := make([]*greengrass.ResourceAccessPolicy, 0)
	for _, rawAccessPolicy := range rawEnvConfig["resource_access_policy"].([]interface{}) {
		accessPolicy := parseResourseAccessPolicy(rawAccessPolicy.(map[string]interface{}))
		accessPolicyList = append(accessPolicyList, accessPolicy)
	}
	envConfig.ResourceAccessPolicies = accessPolicyList

	return envConfig
}

func parseFunctionConfigurationSchema(rawFunctionConfiguration map[string]interface{}) *greengrass.FunctionConfiguration {
	functionConfiguration := &greengrass.FunctionConfiguration{
		Pinned: aws.Bool(rawFunctionConfiguration["pinned"].(bool)),
	}

	if v, ok := rawFunctionConfiguration["encoding_type"]; ok {
		functionConfiguration.EncodingType = aws.String(v.(string))
	}

	if v, ok := rawFunctionConfiguration["exec_args"]; ok {
		functionConfiguration.ExecArgs = aws.String(v.(string))
	}

	if v, ok := rawFunctionConfiguration["executable"]; ok {
		functionConfiguration.Executable = aws.String(v.(string))
	}

	if v, ok := rawFunctionConfiguration["memory_size"]; ok {
		functionConfiguration.MemorySize = aws.Int64(int64(v.(int)))
	}

	if v, ok := rawFunctionConfiguration["timeout"]; ok {
		functionConfiguration.Timeout = aws.Int64(int64(v.(int)))
	}

	if v := rawFunctionConfiguration["environment"].([]interface{}); len(v) > 0 {
		functionConfiguration.Environment = parseFunctionEnvironmentConfigurationSchema(v[0].(map[string]interface{}))
	}

	return functionConfiguration
}

func createFunctionDefinitionVersion(d *schema.ResourceData, conn *greengrass.Greengrass) error {
	var rawData map[string]interface{}
	if v := d.Get("function_definition_version").(*schema.Set).List(); len(v) == 0 {
		return nil
	} else {
		rawData = v[0].(map[string]interface{})
	}

	params := &greengrass.CreateFunctionDefinitionVersionInput{
		FunctionDefinitionId: aws.String(d.Id()),
	}

	if v := d.Get("amzn_client_token").(string); v != "" {
		params.AmznClientToken = aws.String(v)
	}

	if v := rawData["default_config"].([]interface{}); len(v) > 0 {
		params.DefaultConfig = &greengrass.FunctionDefaultConfig{
			Execution: parseFunctionDefaultExecutionConfiguration(v[0].(map[string]interface{})),
		}
	}

	functions := make([]*greengrass.Function, 0)
	for _, functionToCast := range rawData["function"].([]interface{}) {
		rawFunction := functionToCast.(map[string]interface{})
		function := &greengrass.Function{
			Id:          aws.String(rawFunction["id"].(string)),
			FunctionArn: aws.String(rawFunction["function_arn"].(string)),
		}
		if v := rawFunction["function_configuration"].([]interface{}); len(v) > 0 {
			function.FunctionConfiguration = parseFunctionConfigurationSchema(v[0].(map[string]interface{}))
		}
		functions = append(functions, function)
	}
	params.Functions = functions

	log.Printf("[DEBUG] Creating Greengrass Function Definition Version: %s", params)
	_, err := conn.CreateFunctionDefinitionVersion(params)

	if err != nil {
		return err
	}

	return nil
}

func resourceAwsGreengrassFunctionDefinitionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.CreateFunctionDefinitionInput{
		Name: aws.String(d.Get("name").(string)),
	}

	if tags := d.Get("tags").(map[string]interface{}); len(tags) > 0 {
		params.Tags = keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().GreengrassTags()
	}

	log.Printf("[DEBUG] Creating Greengrass Function Definition: %s", params)
	out, err := conn.CreateFunctionDefinition(params)
	if err != nil {
		return err
	}

	d.SetId(*out.Id)

	err = createFunctionDefinitionVersion(d, conn)

	if err != nil {
		return err
	}

	return resourceAwsGreengrassFunctionDefinitionRead(d, meta)
}

func flattenRunAs(runAsConfig *greengrass.FunctionRunAsConfig) map[string]interface{} {
	rawRunAsConfig := make(map[string]interface{})

	if runAsConfig.Gid != nil {
		rawRunAsConfig["gid"] = aws.Int64Value(runAsConfig.Gid)
	}

	if runAsConfig.Uid != nil {
		rawRunAsConfig["uid"] = aws.Int64Value(runAsConfig.Uid)
	}

	return rawRunAsConfig
}

func flattenFunctionExecutionConfiguration(execConfig *greengrass.FunctionExecutionConfig) map[string]interface{} {
	rawExecConfig := make(map[string]interface{})

	if execConfig.IsolationMode != nil {
		rawExecConfig["isolation_mode"] = aws.StringValue(execConfig.IsolationMode)
	}

	if execConfig.RunAs != nil {
		rawExecConfig["run_as"] = []map[string]interface{}{flattenRunAs(execConfig.RunAs)}
	}

	return rawExecConfig
}

func flattenFunctionDefaultExecutionConfiguration(execConfig *greengrass.FunctionDefaultExecutionConfig) map[string]interface{} {
	rawExecConfig := make(map[string]interface{})

	if execConfig.IsolationMode != nil {
		rawExecConfig["isolation_mode"] = aws.StringValue(execConfig.IsolationMode)
	}

	if execConfig.RunAs != nil {
		rawExecConfig["run_as"] = []map[string]interface{}{flattenRunAs(execConfig.RunAs)}
	}

	return rawExecConfig

}

func flattenResourseAccessPolicy(accessPolicy *greengrass.ResourceAccessPolicy) map[string]interface{} {
	rawAccessPolicy := map[string]interface{}{
		"permission":  aws.StringValue(accessPolicy.Permission),
		"resource_id": aws.StringValue(accessPolicy.ResourceId),
	}
	return rawAccessPolicy
}

func flattenFunctionEnvironmentConfigurationSchema(envConfig *greengrass.FunctionConfigurationEnvironment) map[string]interface{} {
	rawEnvConfig := make(map[string]interface{})
	rawEnvConfig["access_sysfs"] = aws.BoolValue(envConfig.AccessSysfs)

	if envConfig.Variables != nil {
		rawVariables := make(map[string]interface{})
		for key, value := range envConfig.Variables {
			rawVariables[key] = *value
		}
		rawEnvConfig["variables"] = rawVariables
	}

	if envConfig.Execution != nil {
		rawEnvConfig["execution"] = []map[string]interface{}{flattenFunctionExecutionConfiguration(envConfig.Execution)}
	}

	rawAccessPolicies := make([]map[string]interface{}, 0)
	for _, accessPolicy := range envConfig.ResourceAccessPolicies {
		rawAccessPolicies = append(rawAccessPolicies, flattenResourseAccessPolicy(accessPolicy))
	}
	rawEnvConfig["resource_access_policy"] = rawAccessPolicies

	return rawEnvConfig
}

func flattenFunctionConfigurationSchema(functionConfiguration *greengrass.FunctionConfiguration) map[string]interface{} {
	rawFunctionConfiguration := make(map[string]interface{})
	rawFunctionConfiguration["pinned"] = aws.BoolValue(functionConfiguration.Pinned)

	if functionConfiguration.EncodingType != nil {
		rawFunctionConfiguration["encoding_type"] = aws.StringValue(functionConfiguration.EncodingType)
	}

	if functionConfiguration.ExecArgs != nil {
		rawFunctionConfiguration["exec_args"] = aws.StringValue(functionConfiguration.ExecArgs)
	}

	if functionConfiguration.Executable != nil {
		rawFunctionConfiguration["executable"] = aws.StringValue(functionConfiguration.Executable)
	}

	if functionConfiguration.MemorySize != nil {
		rawFunctionConfiguration["memory_size"] = aws.Int64Value(functionConfiguration.MemorySize)
	}

	if functionConfiguration.Timeout != nil {
		rawFunctionConfiguration["timeout"] = aws.Int64Value(functionConfiguration.Timeout)
	}

	if functionConfiguration.Environment != nil {
		rawFunctionConfiguration["environment"] = []map[string]interface{}{flattenFunctionEnvironmentConfigurationSchema(functionConfiguration.Environment)}
	}

	return rawFunctionConfiguration
}

func setFunctionDefinitionVersion(latestVersion string, d *schema.ResourceData, conn *greengrass.Greengrass) error {
	params := &greengrass.GetFunctionDefinitionVersionInput{
		FunctionDefinitionId:        aws.String(d.Id()),
		FunctionDefinitionVersionId: aws.String(latestVersion),
	}

	out, err := conn.GetFunctionDefinitionVersion(params)

	if err != nil {
		return err
	}

	rawVersion := make(map[string]interface{})
	d.Set("latest_definition_version_arn", *out.Arn)

	if out.Definition.DefaultConfig != nil && out.Definition.DefaultConfig.Execution != nil {
		rawVersion["default_config"] = []map[string]interface{}{flattenFunctionDefaultExecutionConfiguration(out.Definition.DefaultConfig.Execution)}
	}

	rawFunctionList := make([]map[string]interface{}, 0)
	for _, function := range out.Definition.Functions {
		rawFunction := make(map[string]interface{})
		rawFunction["id"] = *function.Id
		rawFunction["function_arn"] = *function.FunctionArn

		if function.FunctionConfiguration != nil {
			rawFunction["function_configuration"] = []map[string]interface{}{flattenFunctionConfigurationSchema(function.FunctionConfiguration)}
		}

		rawFunctionList = append(rawFunctionList, rawFunction)
	}

	rawVersion["function"] = rawFunctionList

	d.Set("function_definition_version", []map[string]interface{}{rawVersion})

	return nil
}

func resourceAwsGreengrassFunctionDefinitionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.GetFunctionDefinitionInput{
		FunctionDefinitionId: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Reading Greengrass Function Definition: %s", params)
	out, err := conn.GetFunctionDefinition(params)

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Received Greengrass Function Definition: %s", out)

	d.Set("arn", out.Arn)
	d.Set("name", out.Name)

	arn := *out.Arn
	tags, err := keyvaluetags.GreengrassListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for resource (%s): %s", arn, err)
	}
	if err := d.Set("tags", tags.IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	if out.LatestVersion != nil {
		err = setFunctionDefinitionVersion(*out.LatestVersion, d, conn)

		if err != nil {
			return err
		}
	}

	return nil
}

func resourceAwsGreengrassFunctionDefinitionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.UpdateFunctionDefinitionInput{
		Name:                 aws.String(d.Get("name").(string)),
		FunctionDefinitionId: aws.String(d.Id()),
	}

	_, err := conn.UpdateFunctionDefinition(params)
	if err != nil {
		return err
	}

	if d.HasChange("function_definition_version") {
		err = createFunctionDefinitionVersion(d, conn)
		if err != nil {
			return err
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.GreengrassUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}
	return resourceAwsGreengrassFunctionDefinitionRead(d, meta)
}

func resourceAwsGreengrassFunctionDefinitionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.DeleteFunctionDefinitionInput{
		FunctionDefinitionId: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Deleting Greengrass Function Definition: %s", params)

	_, err := conn.DeleteFunctionDefinition(params)

	if err != nil {
		return err
	}

	return nil
}
