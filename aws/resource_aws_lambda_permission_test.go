package aws

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestLambdaPermissionUnmarshalling(t *testing.T) {
	v := LambdaPolicy{}
	err := json.Unmarshal(testLambdaPolicy, &v)
	if err != nil {
		t.Fatalf("Expected no error when unmarshalling: %s", err)
	}

	expectedSid := "36fe77d9-a4ae-13fb-8beb-5dc6821d5291"
	if v.Statement[0].Sid != expectedSid {
		t.Fatalf("Expected Sid to match (%q != %q)", v.Statement[0].Sid, expectedSid)
	}

	expectedFunctionName := "arn:aws:lambda:eu-west-1:319201112229:function:myCustomFunction" // lintignore:AWSAT003,AWSAT005 // unit test
	if v.Statement[0].Resource != expectedFunctionName {
		t.Fatalf("Expected function name to match (%q != %q)", v.Statement[0].Resource, expectedFunctionName)
	}

	expectedAction := "lambda:InvokeFunction"
	if v.Statement[0].Action != expectedAction {
		t.Fatalf("Expected Action to match (%q != %q)", v.Statement[0].Action, expectedAction)
	}

	expectedPrincipal := "events.amazonaws.com"
	if v.Statement[0].Principal["Service"] != expectedPrincipal {
		t.Fatalf("Expected Principal to match (%q != %q)", v.Statement[0].Principal["Service"], expectedPrincipal)
	}

	expectedSourceAccount := "319201112229"
	if v.Statement[0].Condition["StringEquals"]["AWS:SourceAccount"] != expectedSourceAccount {
		t.Fatalf("Expected Source Account to match (%q != %q)",
			v.Statement[0].Condition["StringEquals"]["AWS:SourceAccount"],
			expectedSourceAccount)
	}

	expectedEventSourceToken := "test-event-source-token"
	if v.Statement[0].Condition["StringEquals"]["lambda:EventSourceToken"] != expectedEventSourceToken {
		t.Fatalf("Expected Event Source Token to match (%q != %q)",
			v.Statement[0].Condition["StringEquals"]["lambda:EventSourceToken"],
			expectedEventSourceToken)
	}
}

func TestLambdaPermissionGetQualifierFromLambdaAliasOrVersionArn_alias(t *testing.T) {
	arnWithAlias := "arn:aws:lambda:us-west-2:187636751137:function:lambda_function_name:testalias" // lintignore:AWSAT003,AWSAT005 // unit test
	expectedQualifier := "testalias"
	qualifier, err := getQualifierFromLambdaAliasOrVersionArn(arnWithAlias)
	if err != nil {
		t.Fatalf("Expected no error when getting qualifier: %s", err)
	}
	if qualifier != expectedQualifier {
		t.Fatalf("Expected qualifier to match (%q != %q)", qualifier, expectedQualifier)
	}
}
func TestLambdaPermissionGetQualifierFromLambdaAliasOrVersionArn_govcloud(t *testing.T) {
	arnWithAlias := "arn:aws-us-gov:lambda:us-gov-west-1:187636751137:function:lambda_function_name:testalias" // lintignore:AWSAT003,AWSAT005 // unit test
	expectedQualifier := "testalias"
	qualifier, err := getQualifierFromLambdaAliasOrVersionArn(arnWithAlias)
	if err != nil {
		t.Fatalf("Expected no error when getting qualifier: %s", err)
	}
	if qualifier != expectedQualifier {
		t.Fatalf("Expected qualifier to match (%q != %q)", qualifier, expectedQualifier)
	}
}

func TestLambdaPermissionGetQualifierFromLambdaAliasOrVersionArn_version(t *testing.T) {
	arnWithVersion := "arn:aws:lambda:us-west-2:187636751137:function:lambda_function_name:223" // lintignore:AWSAT003,AWSAT005 // unit test
	expectedQualifier := "223"
	qualifier, err := getQualifierFromLambdaAliasOrVersionArn(arnWithVersion)
	if err != nil {
		t.Fatalf("Expected no error when getting qualifier: %s", err)
	}
	if qualifier != expectedQualifier {
		t.Fatalf("Expected qualifier to match (%q != %q)", qualifier, expectedQualifier)
	}
}

func TestLambdaPermissionGetQualifierFromLambdaAliasOrVersionArn_invalid(t *testing.T) {
	invalidArn := "arn:aws:lambda:us-west-2:187636751137:function:lambda_function_name" // lintignore:AWSAT003,AWSAT005 // unit test
	qualifier, err := getQualifierFromLambdaAliasOrVersionArn(invalidArn)
	if err == nil {
		t.Fatalf("Expected error when getting qualifier")
	}
	if qualifier != "" {
		t.Fatalf("Expected qualifier to be empty (%q)", qualifier)
	}

	// with trailing colon
	invalidArn = "arn:aws:lambda:us-west-2:187636751137:function:lambda_function_name:" // lintignore:AWSAT003,AWSAT005 // unit test
	qualifier, err = getQualifierFromLambdaAliasOrVersionArn(invalidArn)
	if err == nil {
		t.Fatalf("Expected error when getting qualifier")
	}
	if qualifier != "" {
		t.Fatalf("Expected qualifier to be empty (%q)", qualifier)
	}
}

func TestLambdaPermissionGetFunctionNameFromLambdaArn_invalid(t *testing.T) {
	invalidArn := "arn:aws:lambda:us-west-2:187636751137:function:" // lintignore:AWSAT003,AWSAT005 // unit test
	fn, err := getFunctionNameFromLambdaArn(invalidArn)
	if err == nil {
		t.Fatalf("Expected error when parsing invalid ARN (%q)", invalidArn)
	}
	if fn != "" {
		t.Fatalf("Expected empty string when parsing invalid ARN (%q)", invalidArn)
	}
}

func TestLambdaPermissionGetFunctionNameFromLambdaArn_valid(t *testing.T) {
	validArn := "arn:aws:lambda:us-west-2:187636751137:function:lambda_function_name" // lintignore:AWSAT003,AWSAT005 // unit test
	fn, err := getFunctionNameFromLambdaArn(validArn)
	if err != nil {
		t.Fatalf("Expected no error (%q): %q", validArn, err)
	}
	expectedFunctionname := "lambda_function_name"
	if fn != expectedFunctionname {
		t.Fatalf("Expected Lambda function name to match (%q != %q)",
			validArn, expectedFunctionname)
	}

	// With qualifier
	validArn = "arn:aws:lambda:us-west-2:187636751137:function:lambda_function_name:12" // lintignore:AWSAT003,AWSAT005 // unit test
	fn, err = getFunctionNameFromLambdaArn(validArn)
	if err != nil {
		t.Fatalf("Expected no error (%q): %q", validArn, err)
	}
	expectedFunctionname = "lambda_function_name"
	if fn != expectedFunctionname {
		t.Fatalf("Expected Lambda function name to match (%q != %q)",
			validArn, expectedFunctionname)
	}
}

func TestLambdaPermissionGetFunctionNameFromGovCloudLambdaArn(t *testing.T) {
	validArn := "arn:aws-us-gov:lambda:us-gov-west-1:187636751137:function:lambda_function_name" // lintignore:AWSAT003,AWSAT005 // unit test
	fn, err := getFunctionNameFromLambdaArn(validArn)
	if err != nil {
		t.Fatalf("Expected no error (%q): %q", validArn, err)
	}
	expectedFunctionname := "lambda_function_name"
	if fn != expectedFunctionname {
		t.Fatalf("Expected Lambda function name to match (%q != %q)",
			validArn, expectedFunctionname)
	}
}

func TestAccAWSLambdaPermission_basic(t *testing.T) {
	var statement LambdaPolicyStatement

	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_perm_basic_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_perm_basic_%s", rString)

	resourceName := "aws_lambda_permission.allow_cloudwatch"
	functionResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLambdaPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaPermissionConfig(funcName, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLambdaPermissionExists(resourceName, &statement),
					resource.TestCheckResourceAttr(resourceName, "action", "lambda:InvokeFunction"),
					resource.TestCheckResourceAttr(resourceName, "principal", "events.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "AllowExecutionFromCloudWatch"),
					resource.TestCheckResourceAttr(resourceName, "qualifier", ""),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "event_source_token", "test-event-source-token"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCLambdaPermissionImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSLambdaPermission_StatementId_Duplicate(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLambdaPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSLambdaPermissionConfigStatementIdDuplicate(rName),
				ExpectError: regexp.MustCompile(`ResourceConflictException`),
			},
		},
	})
}

func TestAccAWSLambdaPermission_withRawFunctionName(t *testing.T) {
	var statement LambdaPolicyStatement

	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_perm_w_raw_fname_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_perm_w_raw_fname_%s", rString)

	resourceName := "aws_lambda_permission.with_raw_func_name"
	functionResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLambdaPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaPermissionConfig_withRawFunctionName(funcName, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLambdaPermissionExists(resourceName, &statement),
					resource.TestCheckResourceAttr(resourceName, "action", "lambda:InvokeFunction"),
					resource.TestCheckResourceAttr(resourceName, "principal", "events.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "AllowExecutionWithRawFuncName"),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCLambdaPermissionImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSLambdaPermission_withStatementIdPrefix(t *testing.T) {
	var statement LambdaPolicyStatement

	rName := acctest.RandomWithPrefix("tf-acc-test")
	startsWithPrefix := regexp.MustCompile("^AllowExecutionWithStatementIdPrefix-")

	resourceName := "aws_lambda_permission.with_statement_id_prefix"
	functionResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLambdaPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaPermissionConfig_withStatementIdPrefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLambdaPermissionExists(resourceName, &statement),
					resource.TestCheckResourceAttr(resourceName, "action", "lambda:InvokeFunction"),
					resource.TestCheckResourceAttr(resourceName, "principal", "events.amazonaws.com"),
					resource.TestMatchResourceAttr(resourceName, "statement_id", startsWithPrefix),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccAWSCLambdaPermissionImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"statement_id_prefix"},
			},
		},
	})
}

func TestAccAWSLambdaPermission_withQualifier(t *testing.T) {
	var statement LambdaPolicyStatement

	rString := acctest.RandString(8)
	aliasName := fmt.Sprintf("tf_acc_lambda_perm_alias_w_qualifier_%s", rString)
	funcName := fmt.Sprintf("tf_acc_lambda_perm_w_qualifier_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_perm_w_qualifier_%s", rString)

	resourceName := "aws_lambda_permission.with_qualifier"
	functionResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLambdaPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaPermissionConfig_withQualifier(aliasName, funcName, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLambdaPermissionExists(resourceName, &statement),
					resource.TestCheckResourceAttr(resourceName, "action", "lambda:InvokeFunction"),
					resource.TestCheckResourceAttr(resourceName, "principal", "events.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "AllowExecutionWithQualifier"),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "qualifier", aliasName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCLambdaPermissionImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSLambdaPermission_disappears(t *testing.T) {
	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_perm_multi_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_perm_multi_%s", rString)

	resourceName := "aws_lambda_permission.first"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLambdaPermissionDestroy,
		Steps: []resource.TestStep{
			// Here we delete the Lambda permission to verify the follow-on refresh after this step
			// should not error.
			{
				Config: testAccAWSLambdaPermissionConfig_multiplePerms(funcName, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSLambdaPermissionDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSLambdaPermission_multiplePerms(t *testing.T) {
	var firstStatement LambdaPolicyStatement
	var firstStatementModified LambdaPolicyStatement
	var secondStatement LambdaPolicyStatement
	var secondStatementModified LambdaPolicyStatement
	var thirdStatement LambdaPolicyStatement

	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_perm_multi_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_perm_multi_%s", rString)

	resourceNameFirst := "aws_lambda_permission.first"
	resourceNameSecond := "aws_lambda_permission.second"
	resourceNameSecondModified := "aws_lambda_permission.sec0nd"
	resourceNameThird := "aws_lambda_permission.third"
	functionResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLambdaPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaPermissionConfig_multiplePerms(funcName, roleName),
				Check: resource.ComposeTestCheckFunc(
					// 1st
					testAccCheckLambdaPermissionExists(resourceNameFirst, &firstStatement),
					resource.TestCheckResourceAttr(resourceNameFirst, "action", "lambda:InvokeFunction"),
					resource.TestCheckResourceAttr(resourceNameFirst, "principal", "events.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceNameFirst, "statement_id", "AllowExecutionFirst"),
					resource.TestCheckResourceAttrPair(resourceNameFirst, "function_name", functionResourceName, "arn"),
					// 2nd
					testAccCheckLambdaPermissionExists(resourceNameSecond, &firstStatementModified),
					resource.TestCheckResourceAttr(resourceNameSecond, "action", "lambda:*"),
					resource.TestCheckResourceAttr(resourceNameSecond, "principal", "events.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceNameSecond, "statement_id", "AllowExecutionSecond"),
					resource.TestCheckResourceAttrPair(resourceNameSecond, "function_name", functionResourceName, "arn"),
				),
			},
			{
				Config: testAccAWSLambdaPermissionConfig_multiplePermsModified(funcName, roleName),
				Check: resource.ComposeTestCheckFunc(
					// 1st
					testAccCheckLambdaPermissionExists(resourceNameFirst, &secondStatement),
					resource.TestCheckResourceAttr(resourceNameFirst, "action", "lambda:InvokeFunction"),
					resource.TestCheckResourceAttr(resourceNameFirst, "principal", "events.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceNameFirst, "statement_id", "AllowExecutionFirst"),
					resource.TestCheckResourceAttrPair(resourceNameFirst, "function_name", functionResourceName, "arn"),
					// 2nd
					testAccCheckLambdaPermissionExists(resourceNameSecondModified, &secondStatementModified),
					resource.TestCheckResourceAttr(resourceNameSecondModified, "action", "lambda:*"),
					resource.TestCheckResourceAttr(resourceNameSecondModified, "principal", "events.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceNameSecondModified, "statement_id", "AllowExecutionSec0nd"),
					resource.TestCheckResourceAttrPair(resourceNameSecondModified, "function_name", functionResourceName, "arn"),
					// 3rd
					testAccCheckLambdaPermissionExists(resourceNameThird, &thirdStatement),
					resource.TestCheckResourceAttr(resourceNameThird, "action", "lambda:*"),
					resource.TestCheckResourceAttr(resourceNameThird, "principal", "events.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceNameThird, "statement_id", "AllowExecutionThird"),
					resource.TestCheckResourceAttrPair(resourceNameThird, "function_name", functionResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceNameFirst,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCLambdaPermissionImportStateIdFunc(resourceNameFirst),
				ImportStateVerify: true,
			},
			{
				ResourceName:      resourceNameSecondModified,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCLambdaPermissionImportStateIdFunc(resourceNameSecondModified),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSLambdaPermission_withS3(t *testing.T) {
	var statement LambdaPolicyStatement

	rString := acctest.RandString(8)
	bucketName := fmt.Sprintf("tf-acc-bucket-lambda-perm-w-s3-%s", rString)
	funcName := fmt.Sprintf("tf_acc_lambda_perm_w_s3_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_perm_w_s3_%s", rString)

	resourceName := "aws_lambda_permission.with_s3"
	functionResourceName := "aws_lambda_function.test"
	bucketResourceName := "aws_s3_bucket.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLambdaPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaPermissionConfig_withS3(bucketName, funcName, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLambdaPermissionExists(resourceName, &statement),
					resource.TestCheckResourceAttr(resourceName, "action", "lambda:InvokeFunction"),
					resource.TestCheckResourceAttr(resourceName, "principal", "s3.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "AllowExecutionFromS3"),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "source_arn", bucketResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCLambdaPermissionImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSLambdaPermission_withSNS(t *testing.T) {
	var statement LambdaPolicyStatement

	rString := acctest.RandString(8)
	topicName := fmt.Sprintf("tf_acc_topic_lambda_perm_w_sns_%s", rString)
	funcName := fmt.Sprintf("tf_acc_lambda_perm_w_sns_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_perm_w_sns_%s", rString)

	resourceName := "aws_lambda_permission.with_sns"
	functionResourceName := "aws_lambda_function.test"
	snsTopicResourceName := "aws_sns_topic.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLambdaPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaPermissionConfig_withSNS(topicName, funcName, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLambdaPermissionExists(resourceName, &statement),
					resource.TestCheckResourceAttr(resourceName, "action", "lambda:InvokeFunction"),
					resource.TestCheckResourceAttr(resourceName, "principal", "sns.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "AllowExecutionFromSNS"),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "source_arn", snsTopicResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCLambdaPermissionImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSLambdaPermission_withIAMRole(t *testing.T) {
	var statement LambdaPolicyStatement

	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_perm_w_iam_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_perm_w_iam_%s", rString)

	resourceName := "aws_lambda_permission.iam_role"
	iamRoleResourceName := "aws_iam_role.test"
	functionResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLambdaPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaPermissionConfig_withIAMRole(funcName, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLambdaPermissionExists(resourceName, &statement),
					resource.TestCheckResourceAttr(resourceName, "action", "lambda:InvokeFunction"),
					resource.TestCheckResourceAttrPair(resourceName, "principal", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "AllowExecutionFromIAMRole"),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCLambdaPermissionImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAWSLambdaPermissionDisappears(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).lambdaconn

		if rs.Primary.ID == "" {
			return fmt.Errorf("No resource ID is set")
		}

		// Delete permission out of band
		input := lambda.RemovePermissionInput{
			FunctionName: aws.String(rs.Primary.Attributes["function_name"]),
			StatementId:  aws.String(rs.Primary.ID),
		}

		_, err := conn.RemovePermission(&input)
		if err != nil {
			// Missing whole policy or Lambda function (API error)
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == "ResourceNotFoundException" {
					log.Printf("[WARN] No Lambda Permission Policy found: %v", input)
					return nil
				}
			}
			return err
		}

		return err
	}
}

func testAccCheckLambdaPermissionExists(n string, statement *LambdaPolicyStatement) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).lambdaconn

		// IAM is eventually consistent
		var foundStatement *LambdaPolicyStatement
		err := resource.Retry(5*time.Minute, func() *resource.RetryError {
			var err error
			foundStatement, err = lambdaPermissionExists(rs, conn)
			if err != nil {
				if strings.HasPrefix(err.Error(), "ResourceNotFoundException") {
					return resource.RetryableError(err)
				}
				if strings.HasPrefix(err.Error(), "Lambda policy not found") {
					return resource.RetryableError(err)
				}
				if strings.HasPrefix(err.Error(), "Failed to find statement") {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if err != nil {
			return err
		}

		*statement = *foundStatement

		return nil
	}
}

func testAccCheckAWSLambdaPermissionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).lambdaconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lambda_permission" {
			continue
		}

		// IAM is eventually consistent
		err := resource.Retry(5*time.Minute, func() *resource.RetryError {
			err := isLambdaPermissionGone(rs, conn)
			if err != nil {
				if !strings.HasPrefix(err.Error(), "Error unmarshalling Lambda policy") {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func isLambdaPermissionGone(rs *terraform.ResourceState, conn *lambda.Lambda) error {
	params := &lambda.GetPolicyInput{
		FunctionName: aws.String(rs.Primary.Attributes["function_name"]),
	}
	if v, ok := rs.Primary.Attributes["qualifier"]; ok && v != "" {
		params.Qualifier = aws.String(v)
	}

	resp, err := conn.GetPolicy(params)
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == "ResourceNotFoundException" {
			// no policy found => all statements deleted
			return nil
		}
	}
	if err != nil {
		return fmt.Errorf("Unexpected error when checking existence of Lambda permission: %s\n%s",
			rs.Primary.ID, err)
	}

	policyInBytes := []byte(*resp.Policy)
	policy := LambdaPolicy{}
	err = json.Unmarshal(policyInBytes, &policy)
	if err != nil {
		return fmt.Errorf("Error unmarshalling Lambda policy (%s): %s", *resp.Policy, err)
	}

	state, err := findLambdaPolicyStatementById(&policy, rs.Primary.ID)
	if err != nil {
		// statement not found => deleted
		return nil
	}

	return fmt.Errorf("Policy statement expected to be gone (%s):\n%s",
		rs.Primary.ID, *state)
}

func lambdaPermissionExists(rs *terraform.ResourceState, conn *lambda.Lambda) (*LambdaPolicyStatement, error) {
	params := &lambda.GetPolicyInput{
		FunctionName: aws.String(rs.Primary.Attributes["function_name"]),
	}
	if v, ok := rs.Primary.Attributes["qualifier"]; ok && v != "" {
		params.Qualifier = aws.String(v)
	}

	resp, err := conn.GetPolicy(params)
	if err != nil {
		return nil, fmt.Errorf("Lambda policy not found: %q", err)
	}

	if resp.Policy == nil {
		return nil, fmt.Errorf("Received Lambda policy is empty")
	}

	policyInBytes := []byte(*resp.Policy)
	policy := LambdaPolicy{}
	err = json.Unmarshal(policyInBytes, &policy)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling Lambda policy: %s", err)
	}

	return findLambdaPolicyStatementById(&policy, rs.Primary.ID)
}

func testAccAWSCLambdaPermissionImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		if v, ok := rs.Primary.Attributes["qualifier"]; ok && v != "" {
			return fmt.Sprintf("%s:%s/%s", rs.Primary.Attributes["function_name"], v, rs.Primary.Attributes["statement_id"]), nil
		}
		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["function_name"], rs.Primary.Attributes["statement_id"]), nil
	}
}

func testAccAWSLambdaPermissionConfig(funcName, roleName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_permission" "allow_cloudwatch" {
  statement_id       = "AllowExecutionFromCloudWatch"
  action             = "lambda:InvokeFunction"
  function_name      = aws_lambda_function.test.arn
  principal          = "events.amazonaws.com"
  event_source_token = "test-event-source-token"
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.handler"
  runtime       = "nodejs12.x"
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
`, funcName, roleName)
}

func testAccAWSLambdaPermissionConfigStatementIdDuplicate(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_permission" "test1" {
  action             = "lambda:InvokeFunction"
  event_source_token = "test-event-source-token"
  function_name      = aws_lambda_function.test.arn
  principal          = "events.amazonaws.com"
  statement_id       = "AllowExecutionFromCloudWatch"
}

resource "aws_lambda_permission" "test2" {
  action             = "lambda:InvokeFunction"
  event_source_token = "test-event-source-token"
  function_name      = aws_lambda_function.test.arn
  principal          = "events.amazonaws.com"
  statement_id       = "AllowExecutionFromCloudWatch"
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %q
  handler       = "exports.handler"
  role          = aws_iam_role.test.arn
  runtime       = "nodejs12.x"
}

resource "aws_iam_role" "test" {
  name = %q

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
`, rName, rName)
}

func testAccAWSLambdaPermissionConfig_withRawFunctionName(funcName, roleName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_permission" "with_raw_func_name" {
  statement_id  = "AllowExecutionWithRawFuncName"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.arn
  principal     = "events.amazonaws.com"
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.handler"
  runtime       = "nodejs12.x"
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
`, funcName, roleName)
}

func testAccAWSLambdaPermissionConfig_withStatementIdPrefix(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_permission" "with_statement_id_prefix" {
  statement_id_prefix = "AllowExecutionWithStatementIdPrefix-"
  action              = "lambda:InvokeFunction"
  function_name       = aws_lambda_function.test.arn
  principal           = "events.amazonaws.com"
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "lambda_function_name_perm"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.handler"
  runtime       = "nodejs12.x"
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
`, rName)
}

func testAccAWSLambdaPermissionConfig_withQualifier(aliasName, funcName, roleName string) string {
	// lintignore:AWSAT003,AWSAT005 // ARN, region not actually used
	return fmt.Sprintf(`
resource "aws_lambda_permission" "with_qualifier" {
  statement_id   = "AllowExecutionWithQualifier"
  action         = "lambda:InvokeFunction"
  function_name  = aws_lambda_function.test.arn
  principal      = "events.amazonaws.com"
  source_account = "111122223333"
  source_arn     = "arn:aws:events:eu-west-1:111122223333:rule/RunDaily"
  qualifier      = aws_lambda_alias.test_alias.name
}

resource "aws_lambda_alias" "test_alias" {
  name             = "%s"
  description      = "a sample description"
  function_name    = aws_lambda_function.test.arn
  function_version = "$LATEST"
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.handler"
  runtime       = "nodejs12.x"
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
`, aliasName, funcName, roleName)
}

var testAccAWSLambdaPermissionConfig_multiplePerms_tpl = `
resource "aws_lambda_permission" "first" {
  statement_id  = "AllowExecutionFirst"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.arn
  principal     = "events.amazonaws.com"
}

resource "aws_lambda_permission" "%s" {
  statement_id  = "%s"
  action        = "lambda:*"
  function_name = aws_lambda_function.test.arn
  principal     = "events.amazonaws.com"
}
%s

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.handler"
  runtime       = "nodejs12.x"
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
`

func testAccAWSLambdaPermissionConfig_multiplePerms(funcName, roleName string) string {
	return fmt.Sprintf(testAccAWSLambdaPermissionConfig_multiplePerms_tpl,
		"second", "AllowExecutionSecond", "", funcName, roleName)
}

func testAccAWSLambdaPermissionConfig_multiplePermsModified(funcName, roleName string) string {
	return fmt.Sprintf(testAccAWSLambdaPermissionConfig_multiplePerms_tpl,
		"sec0nd", "AllowExecutionSec0nd", `
resource "aws_lambda_permission" "third" {
  statement_id  = "AllowExecutionThird"
  action        = "lambda:*"
  function_name = aws_lambda_function.test.arn
  principal     = "events.amazonaws.com"
}
`, funcName, roleName)
}

func testAccAWSLambdaPermissionConfig_withS3(bucketName, funcName, roleName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_permission" "with_s3" {
  statement_id  = "AllowExecutionFromS3"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.arn
  principal     = "s3.amazonaws.com"
  source_arn    = aws_s3_bucket.default.arn
}

resource "aws_s3_bucket" "default" {
  bucket = "%s"
  acl    = "private"
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.test.arn
  handler       = "exports.handler"
  runtime       = "nodejs12.x"
}

resource "aws_iam_role" "test" {
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
`, bucketName, funcName, roleName)
}

func testAccAWSLambdaPermissionConfig_withSNS(topicName, funcName, roleName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_permission" "with_sns" {
  statement_id  = "AllowExecutionFromSNS"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.arn
  principal     = "sns.amazonaws.com"
  source_arn    = aws_sns_topic.default.arn
}

resource "aws_sns_topic" "default" {
  name = "%s"
}

resource "aws_sns_topic_subscription" "lambda" {
  topic_arn = aws_sns_topic.default.arn
  protocol  = "lambda"
  endpoint  = aws_lambda_function.test.arn
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.test.arn
  handler       = "exports.handler"
  runtime       = "nodejs12.x"
}

resource "aws_iam_role" "test" {
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
`, topicName, funcName, roleName)
}

func testAccAWSLambdaPermissionConfig_withIAMRole(funcName, roleName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_permission" "iam_role" {
  statement_id  = "AllowExecutionFromIAMRole"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.arn
  principal     = aws_iam_role.test.arn
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.test.arn
  handler       = "exports.handler"
  runtime       = "nodejs12.x"
}

resource "aws_iam_role" "test" {
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
`, funcName, roleName)
}

// lintignore:AWSAT003,AWSAT005 // unit test
var testLambdaPolicy = []byte(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Condition": {
        "StringEquals": {
          "AWS:SourceAccount": "319201112229",
          "lambda:EventSourceToken": "test-event-source-token"
        },
        "ArnLike": {
          "AWS:SourceArn": "arn:aws:events:eu-west-1:319201112229:rule/RunDaily"
        }
      },
      "Action": "lambda:InvokeFunction",
      "Resource": "arn:aws:lambda:eu-west-1:319201112229:function:myCustomFunction",
      "Effect": "Allow",
      "Principal": {
        "Service": "events.amazonaws.com"
      },
      "Sid": "36fe77d9-a4ae-13fb-8beb-5dc6821d5291"
    }
  ],
  "Id": "default"
}`)
