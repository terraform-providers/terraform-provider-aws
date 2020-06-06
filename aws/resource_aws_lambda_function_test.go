package aws

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	resource.AddTestSweepers("aws_lambda_function", &resource.Sweeper{
		Name: "aws_lambda_function",
		F:    testSweepLambdaFunctions,
	})
}

func testSweepLambdaFunctions(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	lambdaconn := client.(*AWSClient).lambdaconn

	resp, err := lambdaconn.ListFunctions(&lambda.ListFunctionsInput{})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Lambda Function sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Lambda functions: %s", err)
	}

	if len(resp.Functions) == 0 {
		log.Print("[DEBUG] No aws lambda functions to sweep")
		return nil
	}

	for _, f := range resp.Functions {
		_, err := lambdaconn.DeleteFunction(
			&lambda.DeleteFunctionInput{
				FunctionName: f.FunctionName,
			})
		if err != nil {
			return err
		}
	}

	return nil
}

func TestAccAWSLambdaFunction_basic(t *testing.T) {
	var conf lambda.GetFunctionOutput
	resourceName := "aws_lambda_function.test"

	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_basic_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_basic_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_basic_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_basic_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigBasic(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckAwsLambdaFunctionInvokeArn(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "reserved_concurrent_executions", "-1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

func TestAccAWSLambdaFunction_disappears(t *testing.T) {
	var function lambda.GetFunctionOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigBasic(rName, rName, rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, rName, &function),
					testAccCheckAwsLambdaFunctionDisappears(&function),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSLambdaFunction_concurrency(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_concurrency_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_concurrency_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_concurrency_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_concurrency_%s", rString)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigBasicConcurrency(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "reserved_concurrent_executions", "111"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccAWSLambdaConfigConcurrencyUpdate(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "reserved_concurrent_executions", "222"),
				),
			},
		},
	})
}

func TestAccAWSLambdaFunction_concurrencyCycle(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_concurrency_cycle_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_concurrency_cycle_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_concurrency_cycle_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_concurrency_cycle_%s", rString)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigBasic(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "reserved_concurrent_executions", "-1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccAWSLambdaConfigConcurrencyUpdate(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "reserved_concurrent_executions", "222"),
				),
			},
			{
				Config: testAccAWSLambdaConfigBasic(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "reserved_concurrent_executions", "-1"),
				),
			},
		},
	})
}

func TestAccAWSLambdaFunction_updateRuntime(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_update_runtime_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_update_runtime_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_update_runtime_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_update_runtime_%s", rString)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigBasic(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					resource.TestCheckResourceAttr(resourceName, "runtime", "nodejs12.x"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccAWSLambdaConfigBasicUpdateRuntime(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					resource.TestCheckResourceAttr(resourceName, "runtime", "nodejs10.x"),
				),
			},
		},
	})
}

func TestAccAWSLambdaFunction_expectFilenameAndS3Attributes(t *testing.T) {
	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_expect_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_expect_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_expect_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_expect_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSLambdaConfigWithoutFilenameAndS3Attributes(funcName, policyName, roleName, sgName),
				ExpectError: regexp.MustCompile(`filename or s3_\* attributes must be set`),
			},
		},
	})
}

func TestAccAWSLambdaFunction_envVariables(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_env_vars_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_env_vars_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_env_vars_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_env_vars_%s", rString)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigBasic(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckNoResourceAttr(resourceName, "environment"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccAWSLambdaConfigEnvVariables(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "environment.0.variables.foo", "bar"),
				),
			},
			{
				Config: testAccAWSLambdaConfigEnvVariablesModified(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "environment.0.variables.foo", "baz"),
					resource.TestCheckResourceAttr(resourceName, "environment.0.variables.foo1", "bar1"),
				),
			},
			{
				Config: testAccAWSLambdaConfigEnvVariablesModifiedWithoutEnvironment(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckNoResourceAttr(resourceName, "environment"),
				),
			},
		},
	})
}

func TestAccAWSLambdaFunction_encryptedEnvVariables(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)
	keyDesc := fmt.Sprintf("tf_acc_key_lambda_func_encrypted_env_%s", rString)
	funcName := fmt.Sprintf("tf_acc_lambda_func_encrypted_env_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_encrypted_env_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_encrypted_env_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_encrypted_env_%s", rString)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigEncryptedEnvVariables(keyDesc, funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "environment.0.variables.foo", "bar"),
					testAccMatchResourceAttrRegionalARN(resourceName, "kms_key_arn", "kms", regexp.MustCompile(`key/.+`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccAWSLambdaConfigEncryptedEnvVariablesModified(keyDesc, funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "environment.0.variables.foo", "bar"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_arn", ""),
				),
			},
		},
	})
}

func TestAccAWSLambdaFunction_versioned(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_versioned_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_versioned_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_versioned_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_versioned_%s", rString)
	resourceName := "aws_lambda_function.test"

	version := "1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigVersioned("test-fixtures/lambdatest.zip", funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestMatchResourceAttr(resourceName, "version", regexp.MustCompile("^[0-9]+$")),
					resource.TestCheckResourceAttr(resourceName, "version", version),
					testAccMatchResourceAttrRegionalARN(resourceName, "qualified_arn", "lambda", regexp.MustCompile(fmt.Sprintf("function:%s:[0-9]+$", funcName))),
					testAccCheckResourceAttrRegionalARN(resourceName, "qualified_arn", "lambda", fmt.Sprintf("function:%s:%s", funcName, version)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

func TestAccAWSLambdaFunction_versionedUpdate(t *testing.T) {
	var conf lambda.GetFunctionOutput

	path, zipFile, err := createTempFile("lambda_localUpdate")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)

	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_versioned_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_versioned_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_versioned_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_versioned_%s", rString)
	resourceName := "aws_lambda_function.test"

	var timeBeforeUpdate time.Time

	version := "2"
	versionUpdated := "3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigVersioned("test-fixtures/lambdatest.zip", funcName, policyName, roleName, sgName),
			},
			{
				PreConfig: func() {
					if err := testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func_modified.js": "lambda.js"}, zipFile); err != nil {
						t.Fatalf("error creating zip from files: %s", err)
					}
					timeBeforeUpdate = time.Now()
				},
				Config: testAccAWSLambdaConfigVersioned(path, funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "version", "2"),
					testAccCheckResourceAttrRegionalARN(resourceName, "qualified_arn", "lambda", fmt.Sprintf("function:%s:%s", funcName, version)),
					func(s *terraform.State) error {
						return testAccCheckAttributeIsDateAfter(s, resourceName, "last_modified", timeBeforeUpdate)
					},
				),
			},
			{
				PreConfig: func() {
					timeBeforeUpdate = time.Now()
				},
				Config: testAccAWSLambdaConfigVersionedNodeJs10xRuntime(path, funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestMatchResourceAttr(resourceName, "version", regexp.MustCompile("^3$")),
					testAccCheckResourceAttrRegionalARN(resourceName, "qualified_arn", "lambda", fmt.Sprintf("function:%s:%s", funcName, versionUpdated)),
					resource.TestCheckResourceAttr(resourceName, "runtime", lambda.RuntimeNodejs10X),
					func(s *terraform.State) error {
						return testAccCheckAttributeIsDateAfter(s, resourceName, "last_modified", timeBeforeUpdate)
					},
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

func TestAccAWSLambdaFunction_DeadLetterConfig(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_dlconfig_%s", rString)
	topicName := fmt.Sprintf("tf_acc_topic_lambda_func_dlconfig_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_dlconfig_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_dlconfig_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_dlconfig_%s", rString)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigWithDeadLetterConfig(funcName, topicName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					func(s *terraform.State) error {
						if !strings.HasSuffix(*conf.Configuration.DeadLetterConfig.TargetArn, ":"+topicName) {
							return fmt.Errorf(
								"Expected DeadLetterConfig.TargetArn %s to have suffix %s", *conf.Configuration.DeadLetterConfig.TargetArn, ":"+topicName,
							)
						}
						return nil
					},
				),
			},
			// Ensure configuration can be imported
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			// Ensure configuration can be removed
			{
				Config: testAccAWSLambdaConfigBasic(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
				),
			},
		},
	})
}

func TestAccAWSLambdaFunction_DeadLetterConfigUpdated(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_dlcfg_upd_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_dlcfg_upd_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_dlcfg_upd_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_dlcfg_upd_%s", rString)
	topic1Name := fmt.Sprintf("tf_acc_topic_lambda_func_dlcfg_upd_%s", rString)
	topic2Name := fmt.Sprintf("tf_acc_topic_lambda_func_dlcfg_upd_2_%s", rString)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigWithDeadLetterConfig(funcName, topic1Name, policyName, roleName, sgName),
			},
			{
				Config: testAccAWSLambdaConfigWithDeadLetterConfigUpdated(funcName, topic1Name, topic2Name, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					func(s *terraform.State) error {
						if !strings.HasSuffix(*conf.Configuration.DeadLetterConfig.TargetArn, ":"+topic2Name) {
							return fmt.Errorf(
								"Expected DeadLetterConfig.TargetArn %s to have suffix %s", *conf.Configuration.DeadLetterConfig.TargetArn, ":"+topic2Name,
							)
						}
						return nil
					},
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

func TestAccAWSLambdaFunction_nilDeadLetterConfig(t *testing.T) {
	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_nil_dlcfg_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_nil_dlcfg_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_nil_dlcfg_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_nil_dlcfg_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigWithNilDeadLetterConfig(funcName, policyName, roleName, sgName),
				ExpectError: regexp.MustCompile(
					fmt.Sprintf("Nil dead_letter_config supplied for function: %s", funcName)),
			},
		},
	})
}

func TestAccAWSLambdaFunction_tracingConfig(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_tracing_cfg_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_tracing_cfg_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_tracing_cfg_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_tracing_cfg_%s", rString)
	resourceName := "aws_lambda_function.test"

	if testAccGetPartition() == "aws-us-gov" {
		t.Skip("Lambda tracing config is not supported in GovCloud partition")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigWithTracingConfig(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "tracing_config.0.mode", "Active"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccAWSLambdaConfigWithTracingConfigUpdated(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "tracing_config.0.mode", "PassThrough"),
				),
			},
		},
	})
}

// This test is to verify the existing behavior in the Lambda API where the KMS Key ARN
// is not returned if environment variables are not in use. If the API begins saving this
// value and the kms_key_arn check begins failing, the documentation should be updated.
// Reference: https://github.com/terraform-providers/terraform-provider-aws/issues/6366
func TestAccAWSLambdaFunction_KmsKeyArn_NoEnvironmentVariables(t *testing.T) {
	var function1 lambda.GetFunctionOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigKmsKeyArnNoEnvironmentVariables(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, rName, &function1),
					resource.TestCheckResourceAttr(resourceName, "kms_key_arn", ""),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

func TestAccAWSLambdaFunction_Layers(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_layer_%s", rString)
	layerName := fmt.Sprintf("tf_acc_layer_lambda_func_layer_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_layer_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_layer_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_layer_%s", rString)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigWithLayers(funcName, layerName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckAWSLambdaFunctionVersion(&conf, "$LATEST"),
					resource.TestCheckResourceAttr(resourceName, "layers.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

func TestAccAWSLambdaFunction_LayersUpdate(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_layer_%s", rString)
	layerName := fmt.Sprintf("tf_acc_lambda_layer_%s", rString)
	layer2Name := fmt.Sprintf("tf_acc_lambda_layer2_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_vpc_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_vpc_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_vpc_%s", rString)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigWithLayers(funcName, layerName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckAWSLambdaFunctionVersion(&conf, "$LATEST"),
					resource.TestCheckResourceAttr(resourceName, "layers.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccAWSLambdaConfigWithLayersUpdated(funcName, layerName, layer2Name, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckAWSLambdaFunctionVersion(&conf, "$LATEST"),
					resource.TestCheckResourceAttr(resourceName, "layers.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSLambdaFunction_VPC(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_vpc_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_vpc_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_vpc_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_vpc_%s", rString)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigWithVPC(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckAWSLambdaFunctionVersion(&conf, "$LATEST"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "vpc_config.0.vpc_id", regexp.MustCompile("^vpc-")),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

func TestAccAWSLambdaFunction_VPCRemoval(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigWithVPC(rName, rName, rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, rName, &conf),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccAWSLambdaConfigBasic(rName, rName, rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, rName, &conf),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSLambdaFunction_VPCUpdate(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_vpc_upd_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_vpc_upd_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_vpc_upd_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_vpc_upd_%s", rString)
	sgName2 := fmt.Sprintf("tf_acc_sg_lambda_func_2nd_vpc_upd_%s", rString)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigWithVPC(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckAWSLambdaFunctionVersion(&conf, "$LATEST"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccAWSLambdaConfigWithVPCUpdated(funcName, policyName, roleName, sgName, sgName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckAWSLambdaFunctionVersion(&conf, "$LATEST"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "2"),
				),
			},
		},
	})
}

// See https://github.com/hashicorp/terraform/issues/5767
// and https://github.com/hashicorp/terraform/issues/10272
func TestAccAWSLambdaFunction_VPC_withInvocation(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_vpc_w_invc_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_vpc_w_invc_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_vpc_w_invc_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_vpc_w_invc_%s", rString)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigWithVPC(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccAwsInvokeLambdaFunction(&conf),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

// Reference: https://github.com/terraform-providers/terraform-provider-aws/issues/10044
func TestAccAWSLambdaFunction_VpcConfig_ProperIamDependencies(t *testing.T) {
	var function lambda.GetFunctionOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_function.test"
	vpcResourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigVpcConfigProperIamDependencies(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, rName, &function),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.vpc_id", vpcResourceName, "id"),
				),
			},
		},
	})
}

func TestAccAWSLambdaFunction_EmptyVpcConfig(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_empty_vpc_config_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_empty_vpc_config_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_empty_vpc_config_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_empty_vpc_config_%s", rString)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigWithEmptyVpcConfig(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

func TestAccAWSLambdaFunction_s3(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)
	bucketName := fmt.Sprintf("tf-acc-bucket-lambda-func-s3-%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_s3_%s", rString)
	funcName := fmt.Sprintf("tf_acc_lambda_func_s3_%s", rString)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigS3(bucketName, roleName, funcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckAWSLambdaFunctionVersion(&conf, "$LATEST"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"s3_bucket", "s3_key", "publish"},
			},
		},
	})
}

func TestAccAWSLambdaFunction_localUpdate(t *testing.T) {
	var conf lambda.GetFunctionOutput

	path, zipFile, err := createTempFile("lambda_localUpdate")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)

	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_local_upd_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_local_upd_%s", rString)
	resourceName := "aws_lambda_function.test"

	var timeBeforeUpdate time.Time

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func.js": "lambda.js"}, zipFile); err != nil {
						t.Fatalf("error creating zip from files: %s", err)
					}
				},
				Config: genAWSLambdaFunctionConfig_local(path, roleName, funcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckAwsLambdaSourceCodeHash(&conf, "8DPiX+G1l2LQ8hjBkwRchQFf1TSCEvPrYGRKlM9UoyY="),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				PreConfig: func() {
					if err := testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func_modified.js": "lambda.js"}, zipFile); err != nil {
						t.Fatalf("error creating zip from files: %s", err)
					}
					timeBeforeUpdate = time.Now()
				},
				Config: genAWSLambdaFunctionConfig_local(path, roleName, funcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckAwsLambdaSourceCodeHash(&conf, "0tdaP9H9hsk9c2CycSwOG/sa/x5JyAmSYunA/ce99Pg="),
					func(s *terraform.State) error {
						return testAccCheckAttributeIsDateAfter(s, resourceName, "last_modified", timeBeforeUpdate)
					},
				),
			},
		},
	})
}

func TestAccAWSLambdaFunction_localUpdate_nameOnly(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_local_upd_name_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_local_upd_name_%s", rString)
	resourceName := "aws_lambda_function.test"

	path, zipFile, err := createTempFile("lambda_localUpdate")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)

	updatedPath, updatedZipFile, err := createTempFile("lambda_localUpdate_name_change")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(updatedPath)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func.js": "lambda.js"}, zipFile); err != nil {
						t.Fatalf("error creating zip from files: %s", err)
					}
				},
				Config: genAWSLambdaFunctionConfig_local_name_only(path, roleName, funcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckAwsLambdaSourceCodeHash(&conf, "8DPiX+G1l2LQ8hjBkwRchQFf1TSCEvPrYGRKlM9UoyY="),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				PreConfig: func() {
					if err := testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func_modified.js": "lambda.js"}, updatedZipFile); err != nil {
						t.Fatalf("error creating zip from files: %s", err)
					}
				},
				Config: genAWSLambdaFunctionConfig_local_name_only(updatedPath, roleName, funcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckAwsLambdaSourceCodeHash(&conf, "0tdaP9H9hsk9c2CycSwOG/sa/x5JyAmSYunA/ce99Pg="),
				),
			},
		},
	})
}

func TestAccAWSLambdaFunction_s3Update_basic(t *testing.T) {
	var conf lambda.GetFunctionOutput

	path, zipFile, err := createTempFile("lambda_s3Update")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)

	rString := acctest.RandString(8)
	bucketName := fmt.Sprintf("tf-acc-bucket-lambda-func-s3-upd-basic-%s", rString)
	funcName := fmt.Sprintf("tf_acc_lambda_func_s3_upd_basic_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_s3_upd_basic_%s", rString)
	resourceName := "aws_lambda_function.test"

	key := "lambda-func.zip"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					// Upload 1st version
					if err := testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func.js": "lambda.js"}, zipFile); err != nil {
						t.Fatalf("error creating zip from files: %s", err)
					}
				},
				Config: genAWSLambdaFunctionConfig_s3(bucketName, key, path, roleName, funcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckAwsLambdaSourceCodeHash(&conf, "8DPiX+G1l2LQ8hjBkwRchQFf1TSCEvPrYGRKlM9UoyY="),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish", "s3_bucket", "s3_key", "s3_object_version"},
			},
			{
				PreConfig: func() {
					// Upload 2nd version
					if err := testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func_modified.js": "lambda.js"}, zipFile); err != nil {
						t.Fatalf("error creating zip from files: %s", err)
					}
				},
				Config: genAWSLambdaFunctionConfig_s3(bucketName, key, path, roleName, funcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckAwsLambdaSourceCodeHash(&conf, "0tdaP9H9hsk9c2CycSwOG/sa/x5JyAmSYunA/ce99Pg="),
				),
			},
		},
	})
}

func TestAccAWSLambdaFunction_s3Update_unversioned(t *testing.T) {
	var conf lambda.GetFunctionOutput

	path, zipFile, err := createTempFile("lambda_s3Update")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)

	rString := acctest.RandString(8)
	bucketName := fmt.Sprintf("tf-acc-bucket-lambda-func-s3-upd-unver-%s", rString)
	funcName := fmt.Sprintf("tf_acc_lambda_func_s3_upd_unver_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_s3_upd_unver_%s", rString)
	resourceName := "aws_lambda_function.test"
	key := "lambda-func.zip"
	key2 := "lambda-func-modified.zip"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					// Upload 1st version
					if err := testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func.js": "lambda.js"}, zipFile); err != nil {
						t.Fatalf("error creating zip from files: %s", err)
					}
				},
				Config: testAccAWSLambdaFunctionConfig_s3_unversioned_tpl(bucketName, roleName, funcName, key, path),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckAwsLambdaSourceCodeHash(&conf, "8DPiX+G1l2LQ8hjBkwRchQFf1TSCEvPrYGRKlM9UoyY="),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish", "s3_bucket", "s3_key"},
			},
			{
				PreConfig: func() {
					// Upload 2nd version
					if err := testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func_modified.js": "lambda.js"}, zipFile); err != nil {
						t.Fatalf("error creating zip from files: %s", err)
					}
				},
				Config: testAccAWSLambdaFunctionConfig_s3_unversioned_tpl(bucketName, roleName, funcName, key2, path),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckAwsLambdaSourceCodeHash(&conf, "0tdaP9H9hsk9c2CycSwOG/sa/x5JyAmSYunA/ce99Pg="),
				),
			},
		},
	})
}

func TestAccAWSLambdaFunction_runtimeValidation_noRuntime(t *testing.T) {
	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_runtime_valid_no_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_runtime_valid_no_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_runtime_valid_no_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_runtime_valid_no_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSLambdaConfigNoRuntime(funcName, policyName, roleName, sgName),
				ExpectError: regexp.MustCompile(`("runtime": required field is not set|The argument "runtime" is required)`),
			},
		},
	})
}

func TestAccAWSLambdaFunction_runtimeValidation_NodeJs10x(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)
	resourceName := "aws_lambda_function.test"
	funcName := fmt.Sprintf("tf_acc_lambda_func_runtime_valid_nodejs10x_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_runtime_valid_nodejs10x_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_runtime_valid_nodejs10x_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_runtime_valid_nodejs10x_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigNodeJs10xRuntime(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					resource.TestCheckResourceAttr(resourceName, "runtime", lambda.RuntimeNodejs10X),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

func TestAccAWSLambdaFunction_runtimeValidation_NodeJs12x(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)
	resourceName := "aws_lambda_function.test"
	funcName := fmt.Sprintf("tf_acc_policy_lambda_func_runtime_valid_nodejs12x_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_runtime_valid_nodejs12x_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_runtime_valid_nodejs12x_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_runtime_valid_nodejs12x_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigNodeJs12xRuntime(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					resource.TestCheckResourceAttr(resourceName, "runtime", lambda.RuntimeNodejs12X),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

func TestAccAWSLambdaFunction_runtimeValidation_python27(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)
	resourceName := "aws_lambda_function.test"
	funcName := fmt.Sprintf("tf_acc_lambda_func_runtime_valid_p27_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_runtime_valid_p27_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_runtime_valid_p27_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_runtime_valid_p27_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigPython27Runtime(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					resource.TestCheckResourceAttr(resourceName, "runtime", lambda.RuntimePython27),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

func TestAccAWSLambdaFunction_runtimeValidation_java8(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)
	resourceName := "aws_lambda_function.test"
	funcName := fmt.Sprintf("tf_acc_lambda_func_runtime_valid_j8_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_runtime_valid_j8_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_runtime_valid_j8_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_runtime_valid_j8_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigJava8Runtime(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					resource.TestCheckResourceAttr(resourceName, "runtime", lambda.RuntimeJava8),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

func TestAccAWSLambdaFunction_runtimeValidation_java11(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)
	resourceName := "aws_lambda_function.test"
	funcName := fmt.Sprintf("tf_acc_lambda_func_runtime_valid_j11_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_runtime_valid_j11_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_runtime_valid_j11_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_runtime_valid_j11_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigJava11Runtime(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					resource.TestCheckResourceAttr(resourceName, "runtime", lambda.RuntimeJava11),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

func TestAccAWSLambdaFunction_tags(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)
	resourceName := "aws_lambda_function.test"
	funcName := fmt.Sprintf("tf_acc_lambda_func_tags_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_tags_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_tags_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_tags_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigBasic(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckNoResourceAttr(resourceName, "tags"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccAWSLambdaConfigTags(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value One"),
					resource.TestCheckResourceAttr(resourceName, "tags.Description", "Very interesting"),
				),
			},
			{
				Config: testAccAWSLambdaConfigTagsModified(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value One Changed"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value Two"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value Three"),
				),
			},
		},
	})
}

func TestAccAWSLambdaFunction_runtimeValidation_provided(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)
	resourceName := "aws_lambda_function.test"
	funcName := fmt.Sprintf("tf_acc_lambda_func_runtime_valid_provided_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_runtime_valid_provided_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_runtime_valid_provided_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_runtime_valid_provided_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigProvidedRuntime(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					resource.TestCheckResourceAttr(resourceName, "runtime", lambda.RuntimeProvided),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

func TestAccAWSLambdaFunction_runtimeValidation_python36(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)
	resourceName := "aws_lambda_function.test"
	funcName := fmt.Sprintf("tf_acc_lambda_func_runtime_valid_p36_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_runtime_valid_p36_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_runtime_valid_p36_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_runtime_valid_p36_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigPython36Runtime(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					resource.TestCheckResourceAttr(resourceName, "runtime", lambda.RuntimePython36),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}
func TestAccAWSLambdaFunction_runtimeValidation_python37(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)
	resourceName := "aws_lambda_function.test"
	funcName := fmt.Sprintf("tf_acc_lambda_func_runtime_valid_p37_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_runtime_valid_p37_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_runtime_valid_p37_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_runtime_valid_p37_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigPython37Runtime(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					resource.TestCheckResourceAttr(resourceName, "runtime", lambda.RuntimePython37),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

func TestAccAWSLambdaFunction_runtimeValidation_python38(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)
	resourceName := "aws_lambda_function.test"
	funcName := fmt.Sprintf("tf_acc_lambda_func_runtime_valid_p38_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_runtime_valid_p38_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_runtime_valid_p38_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_runtime_valid_p38_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigPython38Runtime(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					resource.TestCheckResourceAttr(resourceName, "runtime", lambda.RuntimePython38),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

func TestAccAWSLambdaFunction_runtimeValidation_ruby25(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)
	resourceName := "aws_lambda_function.test"
	funcName := fmt.Sprintf("tf_acc_lambda_func_runtime_valid_r25_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_runtime_valid_r25_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_runtime_valid_r25_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_runtime_valid_r25_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigRuby25Runtime(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					resource.TestCheckResourceAttr(resourceName, "runtime", lambda.RuntimeRuby25),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

func TestAccAWSLambdaFunction_runtimeValidation_ruby27(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)
	resourceName := "aws_lambda_function.test"
	funcName := fmt.Sprintf("tf_acc_lambda_func_runtime_valid_r27_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_runtime_valid_r27_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_runtime_valid_r27_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_runtime_valid_r27_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigRuby27Runtime(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, funcName, &conf),
					resource.TestCheckResourceAttr(resourceName, "runtime", lambda.RuntimeRuby27),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

func TestAccAWSLambdaFunction_runtimeValidation_dotnetcore31(t *testing.T) {
	var conf lambda.GetFunctionOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigDotNetCore31Runtime(rName, rName, rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(resourceName, rName, &conf),
					resource.TestCheckResourceAttr(resourceName, "runtime", lambda.RuntimeDotnetcore31),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

func testAccCheckLambdaFunctionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).lambdaconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lambda_function" {
			continue
		}

		_, err := conn.GetFunction(&lambda.GetFunctionInput{
			FunctionName: aws.String(rs.Primary.ID),
		})

		if err == nil {
			return fmt.Errorf("Lambda Function still exists")
		}

	}

	return nil

}

func testAccCheckAwsLambdaFunctionDisappears(function *lambda.GetFunctionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).lambdaconn

		input := &lambda.DeleteFunctionInput{
			FunctionName: function.Configuration.FunctionName,
		}

		_, err := conn.DeleteFunction(input)

		return err
	}
}

func testAccCheckAwsLambdaFunctionExists(res, funcName string, function *lambda.GetFunctionOutput) resource.TestCheckFunc {
	// Wait for IAM role
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[res]
		if !ok {
			return fmt.Errorf("Lambda function not found: %s", res)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Lambda function ID not set")
		}

		conn := testAccProvider.Meta().(*AWSClient).lambdaconn

		params := &lambda.GetFunctionInput{
			FunctionName: aws.String(funcName),
		}

		getFunction, err := conn.GetFunction(params)
		if err != nil {
			return err
		}

		*function = *getFunction

		return nil
	}
}

func testAccCheckAwsLambdaFunctionInvokeArn(name string, function *lambda.GetFunctionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		arn := aws.StringValue(function.Configuration.FunctionArn)
		return testAccCheckResourceAttrRegionalARNAccountID(name, "invoke_arn", "apigateway", "lambda", fmt.Sprintf("path/2015-03-31/functions/%s/invocations", arn))(s)
	}
}

func testAccAwsInvokeLambdaFunction(function *lambda.GetFunctionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		f := function.Configuration
		conn := testAccProvider.Meta().(*AWSClient).lambdaconn

		// If the function is VPC-enabled this will create ENI automatically
		_, err := conn.Invoke(&lambda.InvokeInput{
			FunctionName: f.FunctionName,
		})

		return err
	}
}

func testAccCheckAwsLambdaFunctionName(function *lambda.GetFunctionOutput, expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c := function.Configuration
		if *c.FunctionName != expectedName {
			return fmt.Errorf("Expected function name %s, got %s", expectedName, *c.FunctionName)
		}

		return nil
	}
}

func testAccCheckAWSLambdaFunctionVersion(function *lambda.GetFunctionOutput, expectedVersion string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c := function.Configuration
		if *c.Version != expectedVersion {
			return fmt.Errorf("Expected version %s, got %s", expectedVersion, *c.Version)
		}
		return nil
	}
}

func testAccCheckAwsLambdaSourceCodeHash(function *lambda.GetFunctionOutput, expectedHash string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c := function.Configuration
		if *c.CodeSha256 != expectedHash {
			return fmt.Errorf("Expected code hash %s, got %s", expectedHash, *c.CodeSha256)
		}

		return nil
	}
}

func testAccCheckAttributeIsDateAfter(s *terraform.State, name string, key string, before time.Time) error {
	rs, ok := s.RootModule().Resources[name]
	if !ok {
		return fmt.Errorf("Resource %s not found", name)
	}

	v, ok := rs.Primary.Attributes[key]
	if !ok {
		return fmt.Errorf("%s: Attribute '%s' not found", name, key)
	}

	const ISO8601UTC = "2006-01-02T15:04:05Z0700"
	timeValue, err := time.Parse(ISO8601UTC, v)
	if err != nil {
		return err
	}

	if !before.Before(timeValue) {
		return fmt.Errorf("Expected time attribute %s.%s with value %s was not before %s", name, key, v, before.Format(ISO8601UTC))
	}

	return nil
}

func testAccCreateZipFromFiles(files map[string]string, zipFile *os.File) error {
	if err := zipFile.Truncate(0); err != nil {
		return err
	}
	if _, err := zipFile.Seek(0, 0); err != nil {
		return err
	}

	w := zip.NewWriter(zipFile)

	for source, destination := range files {
		f, err := w.Create(destination)
		if err != nil {
			return err
		}

		fileContent, err := ioutil.ReadFile(source)
		if err != nil {
			return err
		}

		_, err = f.Write(fileContent)
		if err != nil {
			return err
		}
	}

	err := w.Close()
	if err != nil {
		return err
	}

	return w.Flush()
}

func createTempFile(prefix string) (string, *os.File, error) {
	f, err := ioutil.TempFile(os.TempDir(), prefix)
	if err != nil {
		return "", nil, err
	}

	pathToFile, err := filepath.Abs(f.Name())
	if err != nil {
		return "", nil, err
	}
	return pathToFile, f, nil
}

func baseAccAWSLambdaConfig(policyName, roleName, sgName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role_policy" "iam_policy_for_lambda" {
  name = "%s"
  role = "${aws_iam_role.iam_for_lambda.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "logs:CreateLogGroup",
                "logs:CreateLogStream",
                "logs:PutLogEvents"
            ],
            "Resource": "arn:${data.aws_partition.current.partition}:logs:*:*:*"
        },
    {
      "Effect": "Allow",
      "Action": [
        "ec2:CreateNetworkInterface",
				"ec2:DescribeNetworkInterfaces",
				"ec2:DeleteNetworkInterface"
      ],
      "Resource": [
        "*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "SNS:Publish"
      ],
      "Resource": [
        "*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "xray:PutTraceSegments"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}

resource "aws_iam_role" "iam_for_lambda" {
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

resource "aws_vpc" "vpc_for_lambda" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lambda-function"
  }
}

resource "aws_subnet" "subnet_for_lambda" {
  vpc_id     = "${aws_vpc.vpc_for_lambda.id}"
  cidr_block = "10.0.1.0/24"

  tags = {
    Name = "tf-acc-lambda-function-1"
  }
}

resource "aws_security_group" "sg_for_lambda" {
  name        = "%s"
  description = "Allow all inbound traffic for lambda test"
  vpc_id      = "${aws_vpc.vpc_for_lambda.id}"

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
`, policyName, roleName, sgName)
}

func testAccAWSLambdaConfigBasic(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs12.x"
}
`, funcName)
}

func testAccAWSLambdaConfigBasicConcurrency(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs12.x"
    reserved_concurrent_executions = 111
}
`, funcName)
}

func testAccAWSLambdaConfigConcurrencyUpdate(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs12.x"
    reserved_concurrent_executions = 222
}
`, funcName)
}

func testAccAWSLambdaConfigBasicUpdateRuntime(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs10.x"
}
`, funcName)
}

func testAccAWSLambdaConfigWithoutFilenameAndS3Attributes(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
		runtime = "nodejs12.x"
}
`, funcName)
}

func testAccAWSLambdaConfigEnvVariables(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs12.x"
    environment {
        variables = {
            foo = "bar"
        }
    }
}
`, funcName)
}

func testAccAWSLambdaConfigEnvVariablesModified(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs12.x"
    environment {
        variables = {
            foo = "baz"
            foo1 = "bar1"
        }
    }
}
`, funcName)
}

func testAccAWSLambdaConfigEnvVariablesModifiedWithoutEnvironment(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs12.x"
}
`, funcName)
}

func testAccAWSLambdaConfigEncryptedEnvVariables(keyDesc, funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_kms_key" "foo" {
    description = "%s"
    policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_lambda_function" "test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    kms_key_arn = "${aws_kms_key.foo.arn}"
    runtime = "nodejs12.x"
    environment {
        variables = {
            foo = "bar"
        }
    }
}
`, keyDesc, funcName)
}

func testAccAWSLambdaConfigEncryptedEnvVariablesModified(keyDesc, funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_kms_key" "foo" {
    description = "%s"
    policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_lambda_function" "test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs12.x"
    environment {
        variables = {
            foo = "bar"
        }
    }
}
`, keyDesc, funcName)
}

func testAccAWSLambdaConfigVersioned(fileName, funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
    filename = "%s"
    function_name = "%s"
    publish = true
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs12.x"
}
`, fileName, funcName)
}

func testAccAWSLambdaConfigVersionedNodeJs10xRuntime(fileName, funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
    filename = "%s"
    function_name = "%s"
    publish = true
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs10.x"
}
`, fileName, funcName)
}

func testAccAWSLambdaConfigVpcConfigProperIamDependencies(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
  role       = "${aws_iam_role.test.id}"
}

resource "aws_iam_role" "test" {
  name = %[1]q

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

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  depends_on = ["aws_iam_role_policy_attachment.test"]

  vpc_id     = "${aws_vpc.test.id}"
  cidr_block = "10.0.0.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  depends_on = ["aws_iam_role_policy_attachment.test"]

  name   = %[1]q
  vpc_id = "${aws_vpc.test.id}"

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = "${aws_iam_role.test.arn}"
  handler       = "exports.example"
  runtime       = "nodejs12.x"

  vpc_config {
    subnet_ids         = ["${aws_subnet.test.id}"]
    security_group_ids = ["${aws_security_group.test.id}"]
  }
}
`, rName)
}

func testAccAWSLambdaConfigWithTracingConfig(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs12.x"

    tracing_config {
        mode = "Active"
    }
}
`, funcName)
}

func testAccAWSLambdaConfigWithTracingConfigUpdated(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs12.x"

    tracing_config {
        mode = "PassThrough"
    }
}
`, funcName)
}

func testAccAWSLambdaConfigWithDeadLetterConfig(funcName, topicName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs12.x"

    dead_letter_config {
        target_arn = "${aws_sns_topic.test.arn}"
    }
}

resource "aws_sns_topic" "test" {
	name = "%s"
}
`, funcName, topicName)
}

func testAccAWSLambdaConfigWithDeadLetterConfigUpdated(funcName, topic1Name, topic2Name, policyName,
	roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs12.x"

    dead_letter_config {
        target_arn = "${aws_sns_topic.test_2.arn}"
    }
}

resource "aws_sns_topic" "test" {
	name = "%s"
}

resource "aws_sns_topic" "test_2" {
	name = "%s"
}
`, funcName, topic1Name, topic2Name)
}

func testAccAWSLambdaConfigWithNilDeadLetterConfig(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs12.x"

    dead_letter_config {
        target_arn = ""
    }
}
`, funcName)
}

func testAccAWSLambdaConfigKmsKeyArnNoEnvironmentVariables(rName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(rName, rName, rName)+`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  kms_key_arn   = "${aws_kms_key.test.arn}"
  role          = "${aws_iam_role.iam_for_lambda.arn}"
  runtime       = "nodejs12.x"
}
`, rName)
}

func testAccAWSLambdaConfigWithLayers(funcName, layerName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_layer_version" "test" {
    filename = "test-fixtures/lambdatest.zip"
    layer_name = "%s"
    compatible_runtimes = ["nodejs12.x"]
}

resource "aws_lambda_function" "test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs12.x"
    layers = ["${aws_lambda_layer_version.test.arn}"]
}
`, layerName, funcName)
}

func testAccAWSLambdaConfigWithLayersUpdated(funcName, layerName, layer2Name, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_layer_version" "test" {
    filename = "test-fixtures/lambdatest.zip"
    layer_name = "%s"
    compatible_runtimes = ["nodejs12.x"]
}

resource "aws_lambda_layer_version" "test_2" {
    filename = "test-fixtures/lambdatest_modified.zip"
    layer_name = "%s"
    compatible_runtimes = ["nodejs12.x"]
}

resource "aws_lambda_function" "test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs12.x"
    layers = [
        "${aws_lambda_layer_version.test.arn}",
        "${aws_lambda_layer_version.test_2.arn}",
    ]
}
`, layerName, layer2Name, funcName)
}

func testAccAWSLambdaConfigWithVPC(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs12.x"

    vpc_config {
        subnet_ids = ["${aws_subnet.subnet_for_lambda.id}"]
        security_group_ids = ["${aws_security_group.sg_for_lambda.id}"]
    }
}
`, funcName)
}

func testAccAWSLambdaConfigWithVPCUpdated(funcName, policyName, roleName, sgName, sgName2 string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs12.x"

    vpc_config {
        subnet_ids = ["${aws_subnet.subnet_for_lambda.id}", "${aws_subnet.subnet_for_lambda_2.id}"]
        security_group_ids = ["${aws_security_group.sg_for_lambda.id}", "${aws_security_group.sg_for_lambda_2.id}"]
    }
}

resource "aws_subnet" "subnet_for_lambda_2" {
    vpc_id = "${aws_vpc.vpc_for_lambda.id}"
    cidr_block = "10.0.2.0/24"

  tags = {
        Name = "tf-acc-lambda-function-2"
    }
}

resource "aws_security_group" "sg_for_lambda_2" {
  name = "sg_for_lambda_%s"
  description = "Allow all inbound traffic for lambda test"
  vpc_id = "${aws_vpc.vpc_for_lambda.id}"

  ingress {
      from_port = 80
      to_port = 80
      protocol = "tcp"
      cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
      from_port = 0
      to_port = 0
      protocol = "-1"
      cidr_blocks = ["0.0.0.0/0"]
  }
}
`, funcName, sgName2)
}

func testAccAWSLambdaConfigWithEmptyVpcConfig(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
    filename      = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role          = "${aws_iam_role.iam_for_lambda.arn}"
    handler       = "exports.example"
    runtime       = "nodejs12.x"

    vpc_config {
        subnet_ids         = []
        security_group_ids = []
    }
}
`, funcName)
}

func testAccAWSLambdaConfigS3(bucketName, roleName, funcName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "lambda_bucket" {
  bucket = "%s"
}

resource "aws_s3_bucket_object" "lambda_code" {
  bucket = "${aws_s3_bucket.lambda_bucket.id}"
  key    = "lambdatest.zip"
  source = "test-fixtures/lambdatest.zip"
}

resource "aws_iam_role" "iam_for_lambda" {
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
  s3_bucket     = "${aws_s3_bucket.lambda_bucket.id}"
  s3_key        = "${aws_s3_bucket_object.lambda_code.id}"
  function_name = "%s"
  role          = "${aws_iam_role.iam_for_lambda.arn}"
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}
`, bucketName, roleName, funcName)
}

func testAccAWSLambdaConfigNoRuntime(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
}
`, funcName)
}

func testAccAWSLambdaConfigNodeJs10xRuntime(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs10.x"
}
`, funcName)
}

func testAccAWSLambdaConfigNodeJs12xRuntime(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs12.x"
}
`, funcName)
}

func testAccAWSLambdaConfigPython27Runtime(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "python2.7"
}
`, funcName)
}

func testAccAWSLambdaConfigJava8Runtime(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "java8"
}
`, funcName)
}

func testAccAWSLambdaConfigJava11Runtime(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "java11"
}
`, funcName)
}

func testAccAWSLambdaConfigTags(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs12.x"
  tags = {
		Key1 = "Value One"
		Description = "Very interesting"
    }
}
`, funcName)
}

func testAccAWSLambdaConfigTagsModified(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs12.x"
  tags = {
		Key1 = "Value One Changed"
		Key2 = "Value Two"
		Key3 = "Value Three"
    }
}
`, funcName)
}

func testAccAWSLambdaConfigProvidedRuntime(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "provided"
}
`, funcName)
}

func testAccAWSLambdaConfigPython36Runtime(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "python3.6"
}
`, funcName)
}

func testAccAWSLambdaConfigPython37Runtime(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "python3.7"
}
`, funcName)
}

func testAccAWSLambdaConfigPython38Runtime(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "python3.8"
}
`, funcName)
}

func testAccAWSLambdaConfigRuby25Runtime(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "ruby2.5"
}
`, funcName)
}

func testAccAWSLambdaConfigRuby27Runtime(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "ruby2.7"
}
`, funcName)
}

func testAccAWSLambdaConfigDotNetCore31Runtime(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "dotnetcore3.1"
}
`, funcName)
}

func genAWSLambdaFunctionConfig_local(filePath, roleName, funcName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "iam_for_lambda" {
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
  filename         = "%s"
  source_code_hash = "${filebase64sha256("%s")}"
  function_name    = "%s"
  role             = "${aws_iam_role.iam_for_lambda.arn}"
  handler          = "exports.example"
  runtime          = "nodejs12.x"
}
`, roleName, filePath, filePath, funcName)
}

func genAWSLambdaFunctionConfig_local_name_only(filePath, roleName, funcName string) string {
	return testAccAWSLambdaFunctionConfig_local_name_only_tpl(filePath, roleName, funcName)
}

func testAccAWSLambdaFunctionConfig_local_name_only_tpl(filePath, roleName, funcName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "iam_for_lambda" {
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
  filename      = "%s"
  function_name = "%s"
  role          = "${aws_iam_role.iam_for_lambda.arn}"
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}
`, roleName, filePath, funcName)
}

func genAWSLambdaFunctionConfig_s3(bucketName, key, path, roleName, funcName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "artifacts" {
  bucket        = "%s"
  acl           = "private"
  force_destroy = true

  versioning {
    enabled = true
  }
}

resource "aws_s3_bucket_object" "o" {
  bucket = "${aws_s3_bucket.artifacts.bucket}"
  key    = "%s"
  source = "%s"
  etag   = "${filemd5("%s")}"
}

resource "aws_iam_role" "iam_for_lambda" {
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
  s3_bucket         = "${aws_s3_bucket_object.o.bucket}"
  s3_key            = "${aws_s3_bucket_object.o.key}"
  s3_object_version = "${aws_s3_bucket_object.o.version_id}"
  function_name     = "%s"
  role              = "${aws_iam_role.iam_for_lambda.arn}"
  handler           = "exports.example"
  runtime           = "nodejs12.x"
}
`, bucketName, key, path, path, roleName, funcName)
}

func testAccAWSLambdaFunctionConfig_s3_unversioned_tpl(bucketName, roleName, funcName, key, path string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "artifacts" {
  bucket        = "%s"
  acl           = "private"
  force_destroy = true
}

resource "aws_s3_bucket_object" "o" {
  bucket = "${aws_s3_bucket.artifacts.bucket}"
  key    = "%s"
  source = "%s"
  etag   = "${filemd5("%s")}"
}

resource "aws_iam_role" "iam_for_lambda" {
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
  s3_bucket     = "${aws_s3_bucket_object.o.bucket}"
  s3_key        = "${aws_s3_bucket_object.o.key}"
  function_name = "%s"
  role          = "${aws_iam_role.iam_for_lambda.arn}"
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}
`, bucketName, key, path, path, roleName, funcName)
}
