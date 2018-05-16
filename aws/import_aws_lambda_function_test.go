package aws

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
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

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSLambdaConfigBasic(funcName, policyName, roleName, sgName),
			},

			resource.TestStep{
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

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSLambdaConfigWithVPC(funcName, policyName, roleName, sgName),
			},

			resource.TestStep{
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

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSLambdaConfigS3(bucketName, roleName, funcName),
			},

			resource.TestStep{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"s3_bucket", "s3_key", "publish"},
			},
		},
	})
}
