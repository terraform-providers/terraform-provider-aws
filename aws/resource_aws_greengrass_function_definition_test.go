package aws

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/greengrass"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSGreengrassFunctionDefinition_basic(t *testing.T) {
	rString := acctest.RandString(8)
	resourceName := "aws_greengrass_function_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGreengrassFunctionDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGreengrassFunctionDefinitionConfig_basic(rString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("function_definition_%s", rString)),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr("aws_greengrass_function_definition.test", "tags.tagKey", "tagValue"),
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

func TestAccAWSGreengrassFunctionDefinition_DefinitionVersion(t *testing.T) {
	rString := acctest.RandString(8)
	resourceName := "aws_greengrass_function_definition.test"

	runAs := map[string]interface{}{
		"gid": 2,
		"uid": 2,
	}

	execution := map[string]interface{}{
		"isolation_mode": "GreengrassContainer",
		"run_as":         runAs,
	}

	variables := map[string]string{
		"var": "val",
	}

	accessPolicy := map[string]interface{}{
		"permission":  "rw",
		"resource_id": "1",
	}
	environment := map[string]interface{}{
		"access_sysfs":  false,
		"variables":     variables,
		"execution":     execution,
		"access_policy": accessPolicy,
	}

	functionConfiguration := map[string]interface{}{
		"encoding_type": "json",
		"exec_args":     "arg",
		"executable":    "exec_func",
		"memory_size":   1,
		"pinned":        false,
		"timeout":       2,
		"environment":   environment,
	}

	function := map[string]interface{}{
		"id":                     "test_id",
		"function_configuration": functionConfiguration,
	}

	defaultRunAs := map[string]interface{}{
		"gid": 1,
		"uid": 1,
	}
	defaultConfig := map[string]interface{}{
		"isolation_mode": "GreengrassContainer",
		"run_as":         defaultRunAs,
	}

	functionDefinition := map[string]interface{}{
		"default_config": defaultConfig,
		"function":       function,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGreengrassFunctionDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGreengrassFunctionDefinitionConfig_definitionVersion(rString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("function_definition_%s", rString)),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					testAccCheckGreengrassFunction_checkFunctionDefinition(resourceName, functionDefinition),
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

func checkRunAsConfig(config *greengrass.FunctionRunAsConfig, expectedConfig map[string]interface{}) error {
	expectedGid := int64(expectedConfig["gid"].(int))
	if *config.Gid != expectedGid {
		return fmt.Errorf("Gid %d is not equal expected %d", *config.Gid, expectedGid)
	}

	expectedUid := int64(expectedConfig["uid"].(int))
	if *config.Uid != expectedUid {
		return fmt.Errorf("Uid %d is not equal expected %d", *config.Uid, expectedUid)
	}

	return nil
}

func checkDefaultConfig(config *greengrass.FunctionDefaultConfig, expectedConfig map[string]interface{}) error {
	execConfig := config.Execution

	expectedIsolationMode := expectedConfig["isolation_mode"].(string)
	if *execConfig.IsolationMode != expectedIsolationMode {
		return fmt.Errorf("Isolation Mode %s is not equal expected %s", *execConfig.IsolationMode, expectedIsolationMode)
	}

	expectedRunAs := expectedConfig["run_as"].(map[string]interface{})
	if err := checkRunAsConfig(execConfig.RunAs, expectedRunAs); err != nil {
		return err
	}

	return nil
}

func checkResourceAccessPolicy(policy *greengrass.ResourceAccessPolicy, expectedPolicy map[string]interface{}) error {
	expectedPermission := expectedPolicy["permission"].(string)
	if *policy.Permission != expectedPermission {
		return fmt.Errorf("Permission %s is not equal expected %s", *policy.Permission, expectedPermission)
	}

	expectedResourceId := expectedPolicy["resource_id"].(string)
	if *policy.ResourceId != expectedResourceId {
		return fmt.Errorf("Resource Id %s is not equal expected %s", *policy.ResourceId, expectedResourceId)
	}

	return nil

}

func checkExecuctionConfig(config *greengrass.FunctionExecutionConfig, expectedConfig map[string]interface{}) error {
	expectedIsolationMode := expectedConfig["isolation_mode"].(string)
	if *config.IsolationMode != expectedIsolationMode {
		return fmt.Errorf("Isolation Mode %s is not equal expected %s", *config.IsolationMode, expectedIsolationMode)
	}

	expectedRunAs := expectedConfig["run_as"].(map[string]interface{})
	if err := checkRunAsConfig(config.RunAs, expectedRunAs); err != nil {
		return err
	}

	return nil

}

func checkConfigurationEnvironment(config *greengrass.FunctionConfigurationEnvironment, expectedConfig map[string]interface{}) error {
	expectedAccessSysfs := expectedConfig["access_sysfs"].(bool)
	if *config.AccessSysfs != expectedAccessSysfs {
		return fmt.Errorf("AccessSysfs %t is not equal expected %t", *config.AccessSysfs, expectedAccessSysfs)
	}

	if config.Execution == nil {
		return fmt.Errorf("Execution is nil")
	}
	expectedExecution := expectedConfig["execution"].(map[string]interface{})
	if err := checkExecuctionConfig(config.Execution, expectedExecution); err != nil {
		return err
	}

	if len(config.ResourceAccessPolicies) != 1 {
		return fmt.Errorf("Resource Access Policies len %d is not equal expected %d", len(config.ResourceAccessPolicies), 1)
	}
	policy := config.ResourceAccessPolicies[0]
	expectedPolicy := expectedConfig["access_policy"].(map[string]interface{})
	if err := checkResourceAccessPolicy(policy, expectedPolicy); err != nil {
		return err
	}

	variables := make(map[string]string)
	for k, v := range config.Variables {
		variables[k] = *v
	}

	expectedVariables := expectedConfig["variables"].(map[string]string)
	if !reflect.DeepEqual(variables, expectedVariables) {
		return fmt.Errorf("Variables data %v is not equal to expected %v", variables, expectedVariables)
	}

	return nil

}

func checkFunctionConfiguration(config *greengrass.FunctionConfiguration, expectedConfig map[string]interface{}) error {
	expectedEncodingType := expectedConfig["encoding_type"].(string)
	if *config.EncodingType != expectedEncodingType {
		return fmt.Errorf("Encoding Type %s is not equal expected %s", *config.EncodingType, expectedEncodingType)
	}

	expectedExecArgs := expectedConfig["exec_args"].(string)
	if *config.ExecArgs != expectedExecArgs {
		return fmt.Errorf("ExecArgs %s is not equal expected %s", *config.ExecArgs, expectedExecArgs)
	}

	expectedExecutable := expectedConfig["executable"].(string)
	if *config.Executable != expectedExecutable {
		return fmt.Errorf("Executable %s is not equal expected %s", *config.Executable, expectedExecutable)
	}

	expectedPinned := expectedConfig["pinned"].(bool)
	if *config.Pinned != expectedPinned {
		return fmt.Errorf("Pinned %t is not equal expected %t", *config.Pinned, expectedPinned)
	}

	expectedMemorySize := int64(expectedConfig["memory_size"].(int))
	if *config.MemorySize != expectedMemorySize {
		return fmt.Errorf("MemorySize %d is not equal expected %d", *config.MemorySize, expectedMemorySize)
	}

	expectedTimeout := int64(expectedConfig["timeout"].(int))
	if *config.Timeout != expectedTimeout {
		return fmt.Errorf("Timeout %d is not equal expected %d", *config.Timeout, expectedTimeout)
	}

	if config.Environment == nil {
		return fmt.Errorf("Environment is nil")
	}
	expectedEnvironment := expectedConfig["environment"].(map[string]interface{})
	if err := checkConfigurationEnvironment(config.Environment, expectedEnvironment); err != nil {
		return err
	}

	return nil
}

func checkFunction(function *greengrass.Function, expectedFunction map[string]interface{}) error {
	client := testAccProvider.Meta().(*AWSClient).stsconn
	res, _ := client.GetCallerIdentity(&sts.GetCallerIdentityInput{})

	expectedFunctionArn := fmt.Sprintf("arn:aws:lambda:us-west-2:%s:function:test_lambda_wv8l0glb:test", *res.Account)
	if *function.FunctionArn != expectedFunctionArn {
		return fmt.Errorf("FunctionArn %s is not equal expected %s", *function.FunctionArn, expectedFunctionArn)
	}

	expectedId := expectedFunction["id"].(string)
	if *function.Id != expectedId {
		return fmt.Errorf("Id %s is not equal expected %s", *function.Id, expectedId)
	}

	if function.FunctionConfiguration == nil {
		return fmt.Errorf("FunctionConfiguration is nil")
	}
	expectedFunctionConfiguration := expectedFunction["function_configuration"].(map[string]interface{})
	if err := checkFunctionConfiguration(function.FunctionConfiguration, expectedFunctionConfiguration); err != nil {
		return err
	}

	return nil

}

func testAccCheckGreengrassFunction_checkFunctionDefinition(n string, expectedFunctionDefinition map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Greengrass Function Definition ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).greengrassconn

		getFunctionInput := &greengrass.GetFunctionDefinitionInput{
			FunctionDefinitionId: aws.String(rs.Primary.ID),
		}
		definitionOut, err := conn.GetFunctionDefinition(getFunctionInput)

		if err != nil {
			return err
		}

		getVersionInput := &greengrass.GetFunctionDefinitionVersionInput{
			FunctionDefinitionId:        aws.String(rs.Primary.ID),
			FunctionDefinitionVersionId: definitionOut.LatestVersion,
		}
		versionOut, err := conn.GetFunctionDefinitionVersion(getVersionInput)
		if err != nil {
			return err
		}

		function := versionOut.Definition.Functions[0]
		expectedFunction := expectedFunctionDefinition["function"].(map[string]interface{})
		if err := checkFunction(function, expectedFunction); err != nil {
			return err
		}

		defaultConfig := versionOut.Definition.DefaultConfig
		expectedDefaultConfig := expectedFunctionDefinition["default_config"].(map[string]interface{})
		if err := checkDefaultConfig(defaultConfig, expectedDefaultConfig); err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckAWSGreengrassFunctionDefinitionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).greengrassconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_greengrass_function_definition" {
			continue
		}

		params := &greengrass.ListFunctionDefinitionsInput{
			MaxResults: aws.String("20"),
		}

		out, err := conn.ListFunctionDefinitions(params)
		if err != nil {
			return err
		}
		for _, definition := range out.Definitions {
			if *definition.Id == rs.Primary.ID {
				return fmt.Errorf("Expected Greengrass Function Definition to be destroyed, %s found", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccAWSGreengrassFunctionDefinitionConfig_basic(rString string) string {
	return fmt.Sprintf(`
resource "aws_greengrass_function_definition" "test" {
  name = "function_definition_%s"

  tags = {
	"tagKey" = "tagValue"
  } 
}
`, rString)
}

func testAccAWSGreengrassFunctionDefinitionConfig_definitionVersion(rString string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_greengrass_function_definition" "test" {
	name = "function_definition_%[1]s"
	function_definition_version {
		default_config {
			isolation_mode = "GreengrassContainer"
			run_as {
				gid = 1
				uid = 1
			}
		}
		function {
			function_arn = "arn:aws:lambda:us-west-2:${data.aws_caller_identity.current.account_id}:function:test_lambda_wv8l0glb:test"
			id = "test_id"

			function_configuration {
				encoding_type = "json"
				exec_args = "arg"
				executable = "exec_func"
				memory_size = 1
				pinned = false
				timeout = 2

				environment {
					access_sysfs = false
					variables = {
						"var" = "val",
					}

					execution {
						isolation_mode = "GreengrassContainer"
						run_as {
							gid = 2
							uid = 2
						}
					}

					resource_access_policy {
						permission = "rw"
						resource_id = "1"
					}
				}
			}
		}
	}
}
`, rString)
}
