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
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
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
		var testOptGroup bool
		for _, testName := range []string{"tf_test", "tf_acc_"} {
			if strings.HasPrefix(*f.FunctionName, testName) {
				testOptGroup = true
			}
		}

		if !testOptGroup {
			continue
		}

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

func TestAccAWSLambdaFunction_importLocalFile(t *testing.T) {
	resourceName := "aws_lambda_function.lambda_function_test"

	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_import_local_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_import_local_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_import_local_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_import_local_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigBasic(funcName, policyName, roleName, sgName),
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

func TestAccAWSLambdaFunction_importLocalFile_VPC(t *testing.T) {
	resourceName := "aws_lambda_function.lambda_function_test"

	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_import_vpc_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_import_vpc_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_import_vpc_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_import_vpc_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigWithVPC(funcName, policyName, roleName, sgName),
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

func TestAccAWSLambdaFunction_importS3(t *testing.T) {
	resourceName := "aws_lambda_function.lambda_function_s3test"

	rString := acctest.RandString(8)
	bucketName := fmt.Sprintf("tf-acc-bucket-lambda-func-import-s3-%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_import_s3_%s", rString)
	funcName := fmt.Sprintf("tf_acc_lambda_func_import_s3_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigS3(bucketName, roleName, funcName),
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

func TestAccAWSLambdaFunction_basic(t *testing.T) {
	var conf lambda.GetFunctionOutput
	resourceName := "aws_lambda_function.lambda_function_test"

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
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, ":"+funcName),
					resource.TestMatchResourceAttr(resourceName, "invoke_arn", regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:apigateway:[^:]+:lambda:path/2015-03-31/functions/arn:[^:]+:lambda:[^:]+:[^:]+:function:%s/invocations$", funcName))),
					resource.TestCheckResourceAttr(resourceName, "reserved_concurrent_executions", "0"),
				),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigBasicConcurrency(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, ":"+funcName),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "reserved_concurrent_executions", "111"),
				),
			},
			{
				Config: testAccAWSLambdaConfigConcurrencyUpdate(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, ":"+funcName),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "reserved_concurrent_executions", "222"),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigBasic(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, ":"+funcName),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "reserved_concurrent_executions", "0"),
				),
			},
			{
				Config: testAccAWSLambdaConfigConcurrencyUpdate(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, ":"+funcName),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "reserved_concurrent_executions", "222"),
				),
			},
			{
				Config: testAccAWSLambdaConfigBasic(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, ":"+funcName),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "reserved_concurrent_executions", "0"),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigBasic(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "runtime", "nodejs8.10"),
				),
			},
			{
				Config: testAccAWSLambdaConfigBasicUpdateRuntime(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "runtime", "nodejs6.10"),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigBasic(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, ":"+funcName),
					resource.TestCheckNoResourceAttr("aws_lambda_function.lambda_function_test", "environment"),
				),
			},
			{
				Config: testAccAWSLambdaConfigEnvVariables(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, ":"+funcName),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "environment.0.variables.foo", "bar"),
				),
			},
			{
				Config: testAccAWSLambdaConfigEnvVariablesModified(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, ":"+funcName),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "environment.0.variables.foo", "baz"),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "environment.0.variables.foo1", "bar1"),
				),
			},
			{
				Config: testAccAWSLambdaConfigEnvVariablesModifiedWithoutEnvironment(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, ":"+funcName),
					resource.TestCheckNoResourceAttr("aws_lambda_function.lambda_function_test", "environment"),
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
	keyRegex := regexp.MustCompile(`^arn:aws[\w-]*:kms:`)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigEncryptedEnvVariables(keyDesc, funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, ":"+funcName),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "environment.0.variables.foo", "bar"),
					resource.TestMatchResourceAttr("aws_lambda_function.lambda_function_test", "kms_key_arn", keyRegex),
				),
			},
			{
				Config: testAccAWSLambdaConfigEncryptedEnvVariablesModified(keyDesc, funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, ":"+funcName),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "environment.0.variables.foo", "bar"),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "kms_key_arn", ""),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigVersioned("test-fixtures/lambdatest.zip", funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, ":"+funcName),
					resource.TestMatchResourceAttr("aws_lambda_function.lambda_function_test", "version",
						regexp.MustCompile("^[0-9]+$")),
					resource.TestMatchResourceAttr("aws_lambda_function.lambda_function_test", "qualified_arn",
						regexp.MustCompile(":"+funcName+":[0-9]+$")),
				),
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

	var timeBeforeUpdate time.Time

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
					testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func_modified.js": "lambda.js"}, zipFile)
					timeBeforeUpdate = time.Now()
				},
				Config: testAccAWSLambdaConfigVersioned(path, funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, ":"+funcName),
					resource.TestMatchResourceAttr("aws_lambda_function.lambda_function_test", "version",
						regexp.MustCompile("^2$")),
					resource.TestMatchResourceAttr("data.template_file.function_version", "rendered",
						regexp.MustCompile("^2$")),
					resource.TestMatchResourceAttr("aws_lambda_function.lambda_function_test", "qualified_arn",
						regexp.MustCompile(":"+funcName+":[0-9]+$")),
					resource.TestMatchResourceAttr("data.template_file.qualified_arn", "rendered",
						regexp.MustCompile(fmt.Sprintf(":function:%s:2$", funcName))),
					func(s *terraform.State) error {
						return testAccCheckAttributeIsDateAfter(s, "data.template_file.last_modified", "rendered", timeBeforeUpdate)
					},
				),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigWithDeadLetterConfig(funcName, topicName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, ":"+funcName),
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
				ResourceName:            "aws_lambda_function.lambda_function_test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			// Ensure configuration can be removed
			{
				Config: testAccAWSLambdaConfigBasic(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
				),
			},
		},
	})
}

func TestAccAWSLambdaFunction_DeadLetterConfigUpdated(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_dlcfg_upd_%s", rString)
	uFuncName := fmt.Sprintf("tf_acc_lambda_func_dlcfg_upd_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_dlcfg_upd_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_dlcfg_upd_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_dlcfg_upd_%s", rString)
	topic1Name := fmt.Sprintf("tf_acc_topic_lambda_func_dlcfg_upd_%s", rString)
	topic2Name := fmt.Sprintf("tf_acc_topic_lambda_func_dlcfg_upd_2_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigWithDeadLetterConfig(funcName, topic1Name, policyName, roleName, sgName),
			},
			{
				Config: testAccAWSLambdaConfigWithDeadLetterConfigUpdated(funcName, topic1Name, topic2Name, policyName, roleName, sgName, uFuncName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
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
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, ":"+funcName),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "tracing_config.0.mode", "Active"),
				),
			},
			{
				Config: testAccAWSLambdaConfigWithTracingConfigUpdated(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, ":"+funcName),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "tracing_config.0.mode", "PassThrough"),
				),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigWithLayers(funcName, layerName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, ":"+funcName),
					testAccCheckAWSLambdaFunctionVersion(&conf, "$LATEST"),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "layers.#", "1"),
				),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigWithLayers(funcName, layerName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, ":"+funcName),
					testAccCheckAWSLambdaFunctionVersion(&conf, "$LATEST"),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "layers.#", "1"),
				),
			},
			{
				Config: testAccAWSLambdaConfigWithLayersUpdated(funcName, layerName, layer2Name, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, ":"+funcName),
					testAccCheckAWSLambdaFunctionVersion(&conf, "$LATEST"),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "layers.#", "2"),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigWithVPC(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, ":"+funcName),
					testAccCheckAWSLambdaFunctionVersion(&conf, "$LATEST"),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "vpc_config.#", "1"),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "vpc_config.0.subnet_ids.#", "1"),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "vpc_config.0.security_group_ids.#", "1"),
					resource.TestMatchResourceAttr("aws_lambda_function.lambda_function_test", "vpc_config.0.vpc_id", regexp.MustCompile("^vpc-")),
				),
			},
		},
	})
}

func TestAccAWSLambdaFunction_VPCRemoval(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_function.lambda_function_test"

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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigWithVPC(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, ":"+funcName),
					testAccCheckAWSLambdaFunctionVersion(&conf, "$LATEST"),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "vpc_config.#", "1"),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "vpc_config.0.subnet_ids.#", "1"),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "vpc_config.0.security_group_ids.#", "1"),
				),
			},
			{
				Config: testAccAWSLambdaConfigWithVPCUpdated(funcName, policyName, roleName, sgName, sgName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, ":"+funcName),
					testAccCheckAWSLambdaFunctionVersion(&conf, "$LATEST"),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "vpc_config.#", "1"),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "vpc_config.0.subnet_ids.#", "2"),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "vpc_config.0.security_group_ids.#", "2"),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigWithVPC(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					testAccAwsInvokeLambdaFunction(&conf),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigWithEmptyVpcConfig(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.test", funcName, &conf),
					resource.TestCheckResourceAttr("aws_lambda_function.test", "vpc_config.#", "0"),
				),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigS3(bucketName, roleName, funcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_s3test", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, ":"+funcName),
					testAccCheckAWSLambdaFunctionVersion(&conf, "$LATEST"),
				),
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

	var timeBeforeUpdate time.Time

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func.js": "lambda.js"}, zipFile)
				},
				Config: genAWSLambdaFunctionConfig_local(path, roleName, funcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_local", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, funcName),
					testAccCheckAwsLambdaSourceCodeHash(&conf, "8DPiX+G1l2LQ8hjBkwRchQFf1TSCEvPrYGRKlM9UoyY="),
				),
			},
			{
				PreConfig: func() {
					testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func_modified.js": "lambda.js"}, zipFile)
					timeBeforeUpdate = time.Now()
				},
				Config: genAWSLambdaFunctionConfig_local(path, roleName, funcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_local", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, funcName),
					testAccCheckAwsLambdaSourceCodeHash(&conf, "0tdaP9H9hsk9c2CycSwOG/sa/x5JyAmSYunA/ce99Pg="),
					func(s *terraform.State) error {
						return testAccCheckAttributeIsDateAfter(s, "data.template_file.last_modified", "rendered", timeBeforeUpdate)
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
					testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func.js": "lambda.js"}, zipFile)
				},
				Config: genAWSLambdaFunctionConfig_local_name_only(path, roleName, funcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_local", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, funcName),
					testAccCheckAwsLambdaSourceCodeHash(&conf, "8DPiX+G1l2LQ8hjBkwRchQFf1TSCEvPrYGRKlM9UoyY="),
				),
			},
			{
				PreConfig: func() {
					testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func_modified.js": "lambda.js"}, updatedZipFile)
				},
				Config: genAWSLambdaFunctionConfig_local_name_only(updatedPath, roleName, funcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_local", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, funcName),
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

	key := "lambda-func.zip"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					// Upload 1st version
					testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func.js": "lambda.js"}, zipFile)
				},
				Config: genAWSLambdaFunctionConfig_s3(bucketName, key, path, roleName, funcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_s3", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, funcName),
					testAccCheckAwsLambdaSourceCodeHash(&conf, "8DPiX+G1l2LQ8hjBkwRchQFf1TSCEvPrYGRKlM9UoyY="),
				),
			},
			{
				PreConfig: func() {
					// Upload 2nd version
					testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func_modified.js": "lambda.js"}, zipFile)
				},
				Config: genAWSLambdaFunctionConfig_s3(bucketName, key, path, roleName, funcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_s3", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, funcName),
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
					testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func.js": "lambda.js"}, zipFile)
				},
				Config: testAccAWSLambdaFunctionConfig_s3_unversioned_tpl(bucketName, roleName, funcName, key, path),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_s3", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, funcName),
					testAccCheckAwsLambdaSourceCodeHash(&conf, "8DPiX+G1l2LQ8hjBkwRchQFf1TSCEvPrYGRKlM9UoyY="),
				),
			},
			{
				PreConfig: func() {
					// Upload 2nd version
					testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func_modified.js": "lambda.js"}, zipFile)
				},
				Config: testAccAWSLambdaFunctionConfig_s3_unversioned_tpl(bucketName, roleName, funcName, key2, path),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_s3", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, funcName),
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
				ExpectError: regexp.MustCompile(`"runtime": required field is not set`),
			},
		},
	})
}

func TestAccAWSLambdaFunction_runtimeValidation_NodeJs810(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)

	funcName := fmt.Sprintf("tf_acc_lambda_func_runtime_valid_nodejs810_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_runtime_valid_nodejs810_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_runtime_valid_nodejs810_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_runtime_valid_nodejs810_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaConfigNodeJs810Runtime(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "runtime", lambda.RuntimeNodejs810),
				),
			},
		},
	})
}

func TestAccAWSLambdaFunction_runtimeValidation_python27(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)

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
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "runtime", lambda.RuntimePython27),
				),
			},
		},
	})
}

func TestAccAWSLambdaFunction_runtimeValidation_java8(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)

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
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "runtime", lambda.RuntimeJava8),
				),
			},
		},
	})
}

func TestAccAWSLambdaFunction_tags(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)

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
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, ":"+funcName),
					resource.TestCheckNoResourceAttr("aws_lambda_function.lambda_function_test", "tags"),
				),
			},
			{
				Config: testAccAWSLambdaConfigTags(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, ":"+funcName),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "tags.%", "2"),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "tags.Key1", "Value One"),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "tags.Description", "Very interesting"),
				),
			},
			{
				Config: testAccAWSLambdaConfigTagsModified(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					testAccCheckAwsLambdaFunctionName(&conf, funcName),
					testAccCheckAwsLambdaFunctionArnHasSuffix(&conf, ":"+funcName),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "tags.%", "3"),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "tags.Key1", "Value One Changed"),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "tags.Key2", "Value Two"),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "tags.Key3", "Value Three"),
				),
			},
		},
	})
}

func TestAccAWSLambdaFunction_runtimeValidation_provided(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)

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
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "runtime", lambda.RuntimeProvided),
				),
			},
		},
	})
}

func TestAccAWSLambdaFunction_runtimeValidation_python36(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)

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
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "runtime", lambda.RuntimePython36),
				),
			},
		},
	})
}
func TestAccAWSLambdaFunction_runtimeValidation_python37(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)

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
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "runtime", lambda.RuntimePython37),
				),
			},
		},
	})
}

func TestAccAWSLambdaFunction_runtimeValidation_ruby25(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := acctest.RandString(8)

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
					testAccCheckAwsLambdaFunctionExists("aws_lambda_function.lambda_function_test", funcName, &conf),
					resource.TestCheckResourceAttr("aws_lambda_function.lambda_function_test", "runtime", lambda.RuntimeRuby25),
				),
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

func testAccCheckAwsLambdaFunctionArnHasSuffix(function *lambda.GetFunctionOutput, arnSuffix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c := function.Configuration
		if !strings.HasSuffix(*c.FunctionArn, arnSuffix) {
			return fmt.Errorf("Expected function ARN %s to have suffix %s", *c.FunctionArn, arnSuffix)
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
	zipFile.Truncate(0)
	zipFile.Seek(0, 0)

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
    vpc_id = "${aws_vpc.vpc_for_lambda.id}"
    cidr_block = "10.0.1.0/24"

  tags = {
        Name = "tf-acc-lambda-function-1"
    }
}

resource "aws_security_group" "sg_for_lambda" {
  name = "%s"
  description = "Allow all inbound traffic for lambda test"
  vpc_id = "${aws_vpc.vpc_for_lambda.id}"

  ingress {
      from_port = 0
      to_port = 0
      protocol = "-1"
      cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
      from_port = 0
      to_port = 0
      protocol = "-1"
      cidr_blocks = ["0.0.0.0/0"]
  }
}`, policyName, roleName, sgName)
}

func testAccAWSLambdaConfigBasic(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "lambda_function_test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs8.10"
}
`, funcName)
}

func testAccAWSLambdaConfigBasicConcurrency(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "lambda_function_test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs8.10"
    reserved_concurrent_executions = 111
}
`, funcName)
}

func testAccAWSLambdaConfigConcurrencyUpdate(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "lambda_function_test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs8.10"
    reserved_concurrent_executions = 222
}
`, funcName)
}

func testAccAWSLambdaConfigBasicUpdateRuntime(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "lambda_function_test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs6.10"
}
`, funcName)
}

func testAccAWSLambdaConfigWithoutFilenameAndS3Attributes(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "lambda_function_test" {
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
		runtime = "nodejs8.10"
}
`, funcName)
}

func testAccAWSLambdaConfigEnvVariables(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "lambda_function_test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs8.10"
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
resource "aws_lambda_function" "lambda_function_test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs8.10"
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
resource "aws_lambda_function" "lambda_function_test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs8.10"
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

resource "aws_lambda_function" "lambda_function_test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    kms_key_arn = "${aws_kms_key.foo.arn}"
    runtime = "nodejs8.10"
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

resource "aws_lambda_function" "lambda_function_test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs8.10"
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
resource "aws_lambda_function" "lambda_function_test" {
    filename = "%s"
    function_name = "%s"
    publish = true
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs8.10"
}

data "template_file" "function_version" {
  template = "$${function_version}"

  vars {
    function_version = "${aws_lambda_function.lambda_function_test.version}"
  }
}

data "template_file" "last_modified" {
  template = "$${last_modified}"

  vars {
    last_modified = "${aws_lambda_function.lambda_function_test.last_modified}"
  }
}

data "template_file" "qualified_arn" {
  template = "$${qualified_arn}"

  vars {
    qualified_arn = "${aws_lambda_function.lambda_function_test.qualified_arn}"
  }
}
`, fileName, funcName)
}

func testAccAWSLambdaConfigWithTracingConfig(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "lambda_function_test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs8.10"

    tracing_config {
        mode = "Active"
    }
}

`, funcName)
}

func testAccAWSLambdaConfigWithTracingConfigUpdated(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "lambda_function_test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs8.10"

    tracing_config {
        mode = "PassThrough"
    }
}

`, funcName)
}

func testAccAWSLambdaConfigWithDeadLetterConfig(funcName, topicName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "lambda_function_test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs8.10"

    dead_letter_config {
        target_arn = "${aws_sns_topic.lambda_function_test.arn}"
    }
}

resource "aws_sns_topic" "lambda_function_test" {
	name = "%s"
}

`, funcName, topicName)
}

func testAccAWSLambdaConfigWithDeadLetterConfigUpdated(funcName, topic1Name, topic2Name, policyName,
	roleName, sgName, uFuncName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "lambda_function_test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs8.10"

    dead_letter_config {
        target_arn = "${aws_sns_topic.lambda_function_test_2.arn}"
    }
}

resource "aws_sns_topic" "lambda_function_test" {
	name = "%s"
}

resource "aws_sns_topic" "lambda_function_test_2" {
	name = "%s"
}

`, uFuncName, topic1Name, topic2Name)
}

func testAccAWSLambdaConfigWithNilDeadLetterConfig(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "lambda_function_test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs8.10"

    dead_letter_config {
        target_arn = ""
    }
}
`, funcName)
}

func testAccAWSLambdaConfigWithLayers(funcName, layerName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_layer_version" "lambda_function_test" {
    filename = "test-fixtures/lambdatest.zip"
    layer_name = "%s"
    compatible_runtimes = ["nodejs8.10"]
}

resource "aws_lambda_function" "lambda_function_test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs8.10"
    layers = ["${aws_lambda_layer_version.lambda_function_test.layer_arn}"]
}
`, layerName, funcName)
}

func testAccAWSLambdaConfigWithLayersUpdated(funcName, layerName, layer2Name, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_layer_version" "lambda_function_test" {
    filename = "test-fixtures/lambdatest.zip"
    layer_name = "%s"
    compatible_runtimes = ["nodejs8.10"]
}

resource "aws_lambda_layer_version" "lambda_function_test_2" {
    filename = "test-fixtures/lambdatest_modified.zip"
    layer_name = "%s"
    compatible_runtimes = ["nodejs8.10"]
}

resource "aws_lambda_function" "lambda_function_test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs8.10"
    layers = [
        "${aws_lambda_layer_version.lambda_function_test.layer_arn}",
        "${aws_lambda_layer_version.lambda_function_test_2.layer_arn}",
    ]
}
`, layerName, layer2Name, funcName)
}

func testAccAWSLambdaConfigWithVPC(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "lambda_function_test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs8.10"

    vpc_config = {
        subnet_ids = ["${aws_subnet.subnet_for_lambda.id}"]
        security_group_ids = ["${aws_security_group.sg_for_lambda.id}"]
    }
}`, funcName)
}

func testAccAWSLambdaConfigWithVPCUpdated(funcName, policyName, roleName, sgName, sgName2 string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "lambda_function_test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs8.10"

    vpc_config = {
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
    runtime       = "nodejs8.10"

    vpc_config {
        subnet_ids         = []
        security_group_ids = []
    }
}`, funcName)
}

func testAccAWSLambdaConfigS3(bucketName, roleName, funcName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "lambda_bucket" {
  bucket = "%s"
}

resource "aws_s3_bucket_object" "lambda_code" {
  bucket = "${aws_s3_bucket.lambda_bucket.id}"
  key = "lambdatest.zip"
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

resource "aws_lambda_function" "lambda_function_s3test" {
    s3_bucket = "${aws_s3_bucket.lambda_bucket.id}"
    s3_key = "${aws_s3_bucket_object.lambda_code.id}"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs8.10"
}
`, bucketName, roleName, funcName)
}

func testAccAWSLambdaConfigNoRuntime(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "lambda_function_test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
}
`, funcName)
}

func testAccAWSLambdaConfigNodeJs810Runtime(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "lambda_function_test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs8.10"
}
`, funcName)
}

func testAccAWSLambdaConfigPython27Runtime(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "lambda_function_test" {
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
resource "aws_lambda_function" "lambda_function_test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "java8"
}
`, funcName)
}

func testAccAWSLambdaConfigTags(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "lambda_function_test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs8.10"
  tags = {
		Key1 = "Value One"
		Description = "Very interesting"
    }
}
`, funcName)
}

func testAccAWSLambdaConfigTagsModified(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "lambda_function_test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs8.10"
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
resource "aws_lambda_function" "lambda_function_test" {
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
resource "aws_lambda_function" "lambda_function_test" {
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
resource "aws_lambda_function" "lambda_function_test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "python3.7"
}
`, funcName)
}

func testAccAWSLambdaConfigRuby25Runtime(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(baseAccAWSLambdaConfig(policyName, roleName, sgName)+`
resource "aws_lambda_function" "lambda_function_test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "ruby2.5"
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
resource "aws_lambda_function" "lambda_function_local" {
    filename = "%s"
    source_code_hash = "${base64sha256(file("%s"))}"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs8.10"
}

data "template_file" "last_modified" {
  template = "$${last_modified}"

  vars {
    last_modified = "${aws_lambda_function.lambda_function_local.last_modified}"
  }
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
resource "aws_lambda_function" "lambda_function_local" {
    filename = "%s"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs8.10"
}`, roleName, filePath, funcName)
}

func genAWSLambdaFunctionConfig_s3(bucketName, key, path, roleName, funcName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "artifacts" {
    bucket = "%s"
    acl = "private"
    force_destroy = true
    versioning {
        enabled = true
    }
}
resource "aws_s3_bucket_object" "o" {
    bucket = "${aws_s3_bucket.artifacts.bucket}"
    key = "%s"
    source = "%s"
    etag = "${md5(file("%s"))}"
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
resource "aws_lambda_function" "lambda_function_s3" {
    s3_bucket = "${aws_s3_bucket_object.o.bucket}"
    s3_key = "${aws_s3_bucket_object.o.key}"
    s3_object_version = "${aws_s3_bucket_object.o.version_id}"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs8.10"
}
`, bucketName, key, path, path, roleName, funcName)
}

func testAccAWSLambdaFunctionConfig_s3_unversioned_tpl(bucketName, roleName, funcName, key, path string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "artifacts" {
    bucket = "%s"
    acl = "private"
    force_destroy = true
}
resource "aws_s3_bucket_object" "o" {
    bucket = "${aws_s3_bucket.artifacts.bucket}"
    key = "%s"
    source = "%s"
    etag = "${md5(file("%s"))}"
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
resource "aws_lambda_function" "lambda_function_s3" {
    s3_bucket = "${aws_s3_bucket_object.o.bucket}"
    s3_key = "${aws_s3_bucket_object.o.key}"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs8.10"
}`, bucketName, key, path, path, roleName, funcName)
}
