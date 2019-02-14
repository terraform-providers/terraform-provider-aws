package aws

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
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

	expectedFunctionName := "arn:aws:lambda:eu-west-1:319201112229:function:myCustomFunction"
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
	arnWithAlias := "arn:aws:lambda:us-west-2:187636751137:function:lambda_function_name:testalias"
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
	arnWithAlias := "arn:aws-us-gov:lambda:us-gov-west-1:187636751137:function:lambda_function_name:testalias"
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
	arnWithVersion := "arn:aws:lambda:us-west-2:187636751137:function:lambda_function_name:223"
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
	invalidArn := "arn:aws:lambda:us-west-2:187636751137:function:lambda_function_name"
	qualifier, err := getQualifierFromLambdaAliasOrVersionArn(invalidArn)
	if err == nil {
		t.Fatalf("Expected error when getting qualifier")
	}
	if qualifier != "" {
		t.Fatalf("Expected qualifier to be empty (%q)", qualifier)
	}

	// with trailing colon
	invalidArn = "arn:aws:lambda:us-west-2:187636751137:function:lambda_function_name:"
	qualifier, err = getQualifierFromLambdaAliasOrVersionArn(invalidArn)
	if err == nil {
		t.Fatalf("Expected error when getting qualifier")
	}
	if qualifier != "" {
		t.Fatalf("Expected qualifier to be empty (%q)", qualifier)
	}
}

func TestLambdaPermissionGetFunctionNameFromLambdaArn_invalid(t *testing.T) {
	invalidArn := "arn:aws:lambda:us-west-2:187636751137:function:"
	fn, err := getFunctionNameFromLambdaArn(invalidArn)
	if err == nil {
		t.Fatalf("Expected error when parsing invalid ARN (%q)", invalidArn)
	}
	if fn != "" {
		t.Fatalf("Expected empty string when parsing invalid ARN (%q)", invalidArn)
	}
}

func TestLambdaPermissionGetFunctionNameFromLambdaArn_valid(t *testing.T) {
	validArn := "arn:aws:lambda:us-west-2:187636751137:function:lambda_function_name"
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
	validArn = "arn:aws:lambda:us-west-2:187636751137:function:lambda_function_name:12"
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
	validArn := "arn:aws-us-gov:lambda:us-gov-west-1:187636751137:function:lambda_function_name"
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
	funcArnRe := regexp.MustCompile(":function:" + funcName + "$")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLambdaPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaPermissionConfig(funcName, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLambdaPermissionExists("aws_lambda_permission.allow_cloudwatch", &statement),
					resource.TestCheckResourceAttr("aws_lambda_permission.allow_cloudwatch", "action", "lambda:InvokeFunction"),
					resource.TestCheckResourceAttr("aws_lambda_permission.allow_cloudwatch", "principal", "events.amazonaws.com"),
					resource.TestCheckResourceAttr("aws_lambda_permission.allow_cloudwatch", "statement_id", "AllowExecutionFromCloudWatch"),
					resource.TestCheckResourceAttr("aws_lambda_permission.allow_cloudwatch", "qualifier", ""),
					resource.TestMatchResourceAttr("aws_lambda_permission.allow_cloudwatch", "function_name", funcArnRe),
					resource.TestCheckResourceAttr("aws_lambda_permission.allow_cloudwatch", "event_source_token", "test-event-source-token"),
				),
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
	funcArnRe := regexp.MustCompile(":function:" + funcName + "$")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLambdaPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaPermissionConfig_withRawFunctionName(funcName, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLambdaPermissionExists("aws_lambda_permission.with_raw_func_name", &statement),
					resource.TestCheckResourceAttr("aws_lambda_permission.with_raw_func_name", "action", "lambda:InvokeFunction"),
					resource.TestCheckResourceAttr("aws_lambda_permission.with_raw_func_name", "principal", "events.amazonaws.com"),
					resource.TestCheckResourceAttr("aws_lambda_permission.with_raw_func_name", "statement_id", "AllowExecutionWithRawFuncName"),
					resource.TestMatchResourceAttr("aws_lambda_permission.with_raw_func_name", "function_name", funcArnRe),
				),
			},
		},
	})
}

func TestAccAWSLambdaPermission_withStatementIdPrefix(t *testing.T) {
	var statement LambdaPolicyStatement
	rName := acctest.RandomWithPrefix("tf-acc-test")
	endsWithFuncName := regexp.MustCompile(":function:lambda_function_name_perm$")
	startsWithPrefix := regexp.MustCompile("^AllowExecutionWithStatementIdPrefix-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLambdaPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaPermissionConfig_withStatementIdPrefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLambdaPermissionExists("aws_lambda_permission.with_statement_id_prefix", &statement),
					resource.TestCheckResourceAttr("aws_lambda_permission.with_statement_id_prefix", "action", "lambda:InvokeFunction"),
					resource.TestCheckResourceAttr("aws_lambda_permission.with_statement_id_prefix", "principal", "events.amazonaws.com"),
					resource.TestMatchResourceAttr("aws_lambda_permission.with_statement_id_prefix", "statement_id", startsWithPrefix),
					resource.TestMatchResourceAttr("aws_lambda_permission.with_statement_id_prefix", "function_name", endsWithFuncName),
				),
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
	funcArnRe := regexp.MustCompile(":function:" + funcName + "$")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLambdaPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaPermissionConfig_withQualifier(aliasName, funcName, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLambdaPermissionExists("aws_lambda_permission.with_qualifier", &statement),
					resource.TestCheckResourceAttr("aws_lambda_permission.with_qualifier", "action", "lambda:InvokeFunction"),
					resource.TestCheckResourceAttr("aws_lambda_permission.with_qualifier", "principal", "events.amazonaws.com"),
					resource.TestCheckResourceAttr("aws_lambda_permission.with_qualifier", "statement_id", "AllowExecutionWithQualifier"),
					resource.TestMatchResourceAttr("aws_lambda_permission.with_qualifier", "function_name", funcArnRe),
					resource.TestCheckResourceAttr("aws_lambda_permission.with_qualifier", "qualifier", aliasName),
				),
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
	funcArnRe := regexp.MustCompile(":function:" + funcName + "$")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLambdaPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaPermissionConfig_multiplePerms(funcName, roleName),
				Check: resource.ComposeTestCheckFunc(
					// 1st
					testAccCheckLambdaPermissionExists("aws_lambda_permission.first", &firstStatement),
					resource.TestCheckResourceAttr("aws_lambda_permission.first", "action", "lambda:InvokeFunction"),
					resource.TestCheckResourceAttr("aws_lambda_permission.first", "principal", "events.amazonaws.com"),
					resource.TestCheckResourceAttr("aws_lambda_permission.first", "statement_id", "AllowExecutionFirst"),
					resource.TestMatchResourceAttr("aws_lambda_permission.first", "function_name", funcArnRe),
					// 2nd
					testAccCheckLambdaPermissionExists("aws_lambda_permission.second", &firstStatementModified),
					resource.TestCheckResourceAttr("aws_lambda_permission.second", "action", "lambda:*"),
					resource.TestCheckResourceAttr("aws_lambda_permission.second", "principal", "events.amazonaws.com"),
					resource.TestCheckResourceAttr("aws_lambda_permission.second", "statement_id", "AllowExecutionSecond"),
					resource.TestMatchResourceAttr("aws_lambda_permission.second", "function_name", funcArnRe),
				),
			},
			{
				Config: testAccAWSLambdaPermissionConfig_multiplePermsModified(funcName, roleName),
				Check: resource.ComposeTestCheckFunc(
					// 1st
					testAccCheckLambdaPermissionExists("aws_lambda_permission.first", &secondStatement),
					resource.TestCheckResourceAttr("aws_lambda_permission.first", "action", "lambda:InvokeFunction"),
					resource.TestCheckResourceAttr("aws_lambda_permission.first", "principal", "events.amazonaws.com"),
					resource.TestCheckResourceAttr("aws_lambda_permission.first", "statement_id", "AllowExecutionFirst"),
					resource.TestMatchResourceAttr("aws_lambda_permission.first", "function_name", funcArnRe),
					// 2nd
					testAccCheckLambdaPermissionExists("aws_lambda_permission.sec0nd", &secondStatementModified),
					resource.TestCheckResourceAttr("aws_lambda_permission.sec0nd", "action", "lambda:*"),
					resource.TestCheckResourceAttr("aws_lambda_permission.sec0nd", "principal", "events.amazonaws.com"),
					resource.TestCheckResourceAttr("aws_lambda_permission.sec0nd", "statement_id", "AllowExecutionSec0nd"),
					resource.TestMatchResourceAttr("aws_lambda_permission.sec0nd", "function_name", funcArnRe),
					// 3rd
					testAccCheckLambdaPermissionExists("aws_lambda_permission.third", &thirdStatement),
					resource.TestCheckResourceAttr("aws_lambda_permission.third", "action", "lambda:*"),
					resource.TestCheckResourceAttr("aws_lambda_permission.third", "principal", "events.amazonaws.com"),
					resource.TestCheckResourceAttr("aws_lambda_permission.third", "statement_id", "AllowExecutionThird"),
					resource.TestMatchResourceAttr("aws_lambda_permission.third", "function_name", funcArnRe),
				),
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
	funcArnRe := regexp.MustCompile(":function:" + funcName + "$")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLambdaPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaPermissionConfig_withS3(bucketName, funcName, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLambdaPermissionExists("aws_lambda_permission.with_s3", &statement),
					resource.TestCheckResourceAttr("aws_lambda_permission.with_s3", "action", "lambda:InvokeFunction"),
					resource.TestCheckResourceAttr("aws_lambda_permission.with_s3", "principal", "s3.amazonaws.com"),
					resource.TestCheckResourceAttr("aws_lambda_permission.with_s3", "statement_id", "AllowExecutionFromS3"),
					resource.TestMatchResourceAttr("aws_lambda_permission.with_s3", "function_name", funcArnRe),
					resource.TestCheckResourceAttr("aws_lambda_permission.with_s3", "source_arn",
						fmt.Sprintf("arn:aws:s3:::%s", bucketName)),
				),
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

	funcArnRe := regexp.MustCompile(":function:" + funcName + "$")
	topicArnRe := regexp.MustCompile(":" + topicName + "$")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLambdaPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaPermissionConfig_withSNS(topicName, funcName, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLambdaPermissionExists("aws_lambda_permission.with_sns", &statement),
					resource.TestCheckResourceAttr("aws_lambda_permission.with_sns", "action", "lambda:InvokeFunction"),
					resource.TestCheckResourceAttr("aws_lambda_permission.with_sns", "principal", "sns.amazonaws.com"),
					resource.TestCheckResourceAttr("aws_lambda_permission.with_sns", "statement_id", "AllowExecutionFromSNS"),
					resource.TestMatchResourceAttr("aws_lambda_permission.with_sns", "function_name", funcArnRe),
					resource.TestMatchResourceAttr("aws_lambda_permission.with_sns", "source_arn", topicArnRe),
				),
			},
		},
	})
}

func TestAccAWSLambdaPermission_withIAMRole(t *testing.T) {
	var statement LambdaPolicyStatement

	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_perm_w_iam_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_perm_w_iam_%s", rString)
	funcArnRe := regexp.MustCompile(":function:" + funcName + "$")
	roleArnRe := regexp.MustCompile("/" + roleName + "$")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLambdaPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaPermissionConfig_withIAMRole(funcName, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLambdaPermissionExists("aws_lambda_permission.iam_role", &statement),
					resource.TestCheckResourceAttr("aws_lambda_permission.iam_role", "action", "lambda:InvokeFunction"),
					resource.TestMatchResourceAttr("aws_lambda_permission.iam_role", "principal", roleArnRe),
					resource.TestCheckResourceAttr("aws_lambda_permission.iam_role", "statement_id", "AllowExecutionFromIAMRole"),
					resource.TestMatchResourceAttr("aws_lambda_permission.iam_role", "function_name", funcArnRe),
				),
			},
		},
	})
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

func testAccAWSLambdaPermissionConfig(funcName, roleName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_permission" "allow_cloudwatch" {
    statement_id = "AllowExecutionFromCloudWatch"
    action = "lambda:InvokeFunction"
    function_name = "${aws_lambda_function.test_lambda.arn}"
    principal = "events.amazonaws.com"
    event_source_token = "test-event-source-token"
}

resource "aws_lambda_function" "test_lambda" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.handler"
    runtime = "nodejs8.10"
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
}`, funcName, roleName)
}

func testAccAWSLambdaPermissionConfigStatementIdDuplicate(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_permission" "test1" {
  action             = "lambda:InvokeFunction"
  event_source_token = "test-event-source-token"
  function_name      = "${aws_lambda_function.test.arn}"
  principal          = "events.amazonaws.com"
  statement_id       = "AllowExecutionFromCloudWatch"
}

resource "aws_lambda_permission" "test2" {
  action             = "lambda:InvokeFunction"
  event_source_token = "test-event-source-token"
  function_name      = "${aws_lambda_function.test.arn}"
  principal          = "events.amazonaws.com"
  statement_id       = "AllowExecutionFromCloudWatch"
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %q
  handler       = "exports.handler"
  role          = "${aws_iam_role.test.arn}"
  runtime       = "nodejs8.10"
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
}`, rName, rName)
}

func testAccAWSLambdaPermissionConfig_withRawFunctionName(funcName, roleName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_permission" "with_raw_func_name" {
    statement_id = "AllowExecutionWithRawFuncName"
    action = "lambda:InvokeFunction"
    function_name = "${aws_lambda_function.test_lambda.arn}"
    principal = "events.amazonaws.com"
}

resource "aws_lambda_function" "test_lambda" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.handler"
    runtime = "nodejs8.10"
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
    action = "lambda:InvokeFunction"
    function_name = "${aws_lambda_function.test_lambda.arn}"
    principal = "events.amazonaws.com"
}

resource "aws_lambda_function" "test_lambda" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "lambda_function_name_perm"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.handler"
    runtime = "nodejs8.10"
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
	return fmt.Sprintf(`
resource "aws_lambda_permission" "with_qualifier" {
    statement_id = "AllowExecutionWithQualifier"
    action = "lambda:InvokeFunction"
    function_name = "${aws_lambda_function.test_lambda.arn}"
    principal = "events.amazonaws.com"
    source_account = "111122223333"
    source_arn = "arn:aws:events:eu-west-1:111122223333:rule/RunDaily"
    qualifier = "${aws_lambda_alias.test_alias.name}"
}

resource "aws_lambda_alias" "test_alias" {
    name = "%s"
    description = "a sample description"
    function_name = "${aws_lambda_function.test_lambda.arn}"
    function_version = "$LATEST"
}

resource "aws_lambda_function" "test_lambda" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.handler"
    runtime = "nodejs8.10"
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
    statement_id = "AllowExecutionFirst"
    action = "lambda:InvokeFunction"
    function_name = "${aws_lambda_function.test_lambda.arn}"
    principal = "events.amazonaws.com"
}

resource "aws_lambda_permission" "%s" {
    statement_id = "%s"
    action = "lambda:*"
    function_name = "${aws_lambda_function.test_lambda.arn}"
    principal = "events.amazonaws.com"
}
%s

resource "aws_lambda_function" "test_lambda" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.handler"
    runtime = "nodejs8.10"
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
    statement_id = "AllowExecutionThird"
    action = "lambda:*"
    function_name = "${aws_lambda_function.test_lambda.arn}"
    principal = "events.amazonaws.com"
}`, funcName, roleName)
}

func testAccAWSLambdaPermissionConfig_withS3(bucketName, funcName, roleName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_permission" "with_s3" {
    statement_id = "AllowExecutionFromS3"
    action = "lambda:InvokeFunction"
    function_name = "${aws_lambda_function.my-func.arn}"
    principal = "s3.amazonaws.com"
    source_arn = "${aws_s3_bucket.default.arn}"
}

resource "aws_s3_bucket" "default" {
	bucket = "%s"
    acl = "private"
}

resource "aws_lambda_function" "my-func" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.police.arn}"
    handler = "exports.handler"
    runtime = "nodejs8.10"
}

resource "aws_iam_role" "police" {
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
    statement_id = "AllowExecutionFromSNS"
    action = "lambda:InvokeFunction"
    function_name = "${aws_lambda_function.my-func.arn}"
    principal = "sns.amazonaws.com"
    source_arn = "${aws_sns_topic.default.arn}"
}

resource "aws_sns_topic" "default" {
	name = "%s"
}

resource "aws_sns_topic_subscription" "lambda" {
    topic_arn = "${aws_sns_topic.default.arn}"
    protocol  = "lambda"
    endpoint  = "${aws_lambda_function.my-func.arn}"
}

resource "aws_lambda_function" "my-func" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.police.arn}"
    handler = "exports.handler"
    runtime = "nodejs8.10"
}

resource "aws_iam_role" "police" {
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
    statement_id = "AllowExecutionFromIAMRole"
    action = "lambda:InvokeFunction"
    function_name = "${aws_lambda_function.my-func.arn}"
    principal = "${aws_iam_role.police.arn}"
}

resource "aws_lambda_function" "my-func" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.police.arn}"
    handler = "exports.handler"
    runtime = "nodejs8.10"
}

resource "aws_iam_role" "police" {
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

var testLambdaPolicy = []byte(`{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Condition": {
				"StringEquals": {"AWS:SourceAccount": "319201112229", "lambda:EventSourceToken": "test-event-source-token"},
				"ArnLike":{"AWS:SourceArn":"arn:aws:events:eu-west-1:319201112229:rule/RunDaily"}
			},
			"Action": "lambda:InvokeFunction",
			"Resource": "arn:aws:lambda:eu-west-1:319201112229:function:myCustomFunction",
			"Effect": "Allow",
			"Principal": {"Service":"events.amazonaws.com"},
			"Sid": "36fe77d9-a4ae-13fb-8beb-5dc6821d5291"
		}
	],
	"Id":"default"
}`)
