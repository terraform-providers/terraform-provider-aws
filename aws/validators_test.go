package aws

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
)

func TestValidateTypeStringNullableBoolean(t *testing.T) {
	testCases := []struct {
		val         interface{}
		expectedErr *regexp.Regexp
	}{
		{
			val: "",
		},
		{
			val: "0",
		},
		{
			val: "1",
		},
		{
			val: "true",
		},
		{
			val: "false",
		},
		{
			val:         "invalid",
			expectedErr: regexp.MustCompile(`to be one of \["", false, true\]`),
		},
	}

	matchErr := func(errs []error, r *regexp.Regexp) bool {
		// err must match one provided
		for _, err := range errs {
			if r.MatchString(err.Error()) {
				return true
			}
		}

		return false
	}

	for i, tc := range testCases {
		_, errs := validateTypeStringNullableBoolean(tc.val, "test_property")

		if len(errs) == 0 && tc.expectedErr == nil {
			continue
		}

		if len(errs) != 0 && tc.expectedErr == nil {
			t.Fatalf("expected test case %d to produce no errors, got %v", i, errs)
		}

		if !matchErr(errs, tc.expectedErr) {
			t.Fatalf("expected test case %d to produce error matching \"%s\", got %v", i, tc.expectedErr, errs)
		}
	}
}

func TestValidateTypeStringNullableFloat(t *testing.T) {
	testCases := []struct {
		val         interface{}
		expectedErr *regexp.Regexp
	}{
		{
			val: "",
		},
		{
			val: "0",
		},
		{
			val: "1",
		},
		{
			val: "42.0",
		},
		{
			val:         "threeve",
			expectedErr: regexp.MustCompile(`cannot parse`),
		},
	}

	matchErr := func(errs []error, r *regexp.Regexp) bool {
		// err must match one provided
		for _, err := range errs {
			if r.MatchString(err.Error()) {
				return true
			}
		}

		return false
	}

	for i, tc := range testCases {
		_, errs := validateTypeStringNullableFloat(tc.val, "test_property")

		if len(errs) == 0 && tc.expectedErr == nil {
			continue
		}

		if len(errs) != 0 && tc.expectedErr == nil {
			t.Fatalf("expected test case %d to produce no errors, got %v", i, errs)
		}

		if !matchErr(errs, tc.expectedErr) {
			t.Fatalf("expected test case %d to produce error matching \"%s\", got %v", i, tc.expectedErr, errs)
		}
	}
}

func TestValidateCloudWatchDashboardName(t *testing.T) {
	validNames := []string{
		"HelloWorl_d",
		"hello-world",
		"hello-world-012345",
	}
	for _, v := range validNames {
		_, errors := validateCloudWatchDashboardName(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid CloudWatch dashboard name: %q", v, errors)
		}
	}

	invalidNames := []string{
		"special@character",
		"slash/in-the-middle",
		"dot.in-the-middle",
		strings.Repeat("W", 256), // > 255
	}
	for _, v := range invalidNames {
		_, errors := validateCloudWatchDashboardName(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid CloudWatch dashboard name", v)
		}
	}
}

func TestValidateCloudWatchEventRuleName(t *testing.T) {
	validNames := []string{
		"HelloWorl_d",
		"hello-world",
		"hello.World0125",
	}
	for _, v := range validNames {
		_, errors := validateCloudWatchEventRuleName(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid CW event rule name: %q", v, errors)
		}
	}

	invalidNames := []string{
		"special@character",
		"slash/in-the-middle",
		// Length > 64
		"TooLooooooooooooooooooooooooooooooooooooooooooooooooooooooongName",
	}
	for _, v := range invalidNames {
		_, errors := validateCloudWatchEventRuleName(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid CW event rule name", v)
		}
	}
}

func TestValidateLambdaFunctionName(t *testing.T) {
	validNames := []string{
		"arn:aws:lambda:us-west-2:123456789012:function:ThumbNail",            //lintignore:AWSAT003,AWSAT005
		"arn:aws-us-gov:lambda:us-west-2:123456789012:function:ThumbNail",     //lintignore:AWSAT003,AWSAT005
		"arn:aws-us-gov:lambda:us-gov-west-1:123456789012:function:ThumbNail", //lintignore:AWSAT003,AWSAT005
		"FunctionName",
		"function-name",
	}
	for _, v := range validNames {
		_, errors := validateLambdaFunctionName(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid Lambda function name: %q", v, errors)
		}
	}

	invalidNames := []string{
		"/FunctionNameWithSlash",
		"function.name.with.dots",
		// length > 140
		"arn:aws:lambda:us-west-2:123456789012:function:TooLoooooo" + //lintignore:AWSAT003,AWSAT005
			"ooooooooooooooooooooooooooooooooooooooooooooooooooooooo" +
			"ooooooooooooooooongFunctionName",
	}
	for _, v := range invalidNames {
		_, errors := validateLambdaFunctionName(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid Lambda function name", v)
		}
	}
}

func TestValidateLambdaQualifier(t *testing.T) {
	validNames := []string{
		"123",
		"prod",
		"PROD",
		"MyTestEnv",
		"contains-dashes",
		"contains_underscores",
		"$LATEST",
	}
	for _, v := range validNames {
		_, errors := validateLambdaQualifier(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid Lambda function qualifier: %q", v, errors)
		}
	}

	invalidNames := []string{
		// No ARNs allowed
		"arn:aws:lambda:us-west-2:123456789012:function:prod", //lintignore:AWSAT003,AWSAT005
		// length > 128
		"TooLooooooooooooooooooooooooooooooooooooooooooooooooooo" +
			"ooooooooooooooooooooooooooooooooooooooooooooooooooo" +
			"oooooooooooongQualifier",
	}
	for _, v := range invalidNames {
		_, errors := validateLambdaQualifier(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid Lambda function qualifier", v)
		}
	}
}

func TestValidateLambdaPermissionAction(t *testing.T) {
	validNames := []string{
		"lambda:*",
		"lambda:InvokeFunction",
		"*",
	}
	for _, v := range validNames {
		_, errors := validateLambdaPermissionAction(v, "action")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid Lambda permission action: %q", v, errors)
		}
	}

	invalidNames := []string{
		"yada",
		"lambda:123",
		"*:*",
		"lambda:Invoke*",
	}
	for _, v := range invalidNames {
		_, errors := validateLambdaPermissionAction(v, "action")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid Lambda permission action", v)
		}
	}
}

func TestValidateLambdaPermissionEventSourceToken(t *testing.T) {
	validTokens := []string{
		"amzn1.ask.skill.80c92c86-e6dd-4c4b-8d0d-000000000000",
		"test-event-source-token",
		strings.Repeat(".", 256),
	}
	for _, v := range validTokens {
		_, errors := validateLambdaPermissionEventSourceToken(v, "event_source_token")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid Lambda permission event source token", v)
		}
	}

	invalidTokens := []string{
		"!",
		"test event source token",
		strings.Repeat(".", 257),
	}
	for _, v := range invalidTokens {
		_, errors := validateLambdaPermissionEventSourceToken(v, "event_source_token")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid Lambda permission event source token", v)
		}
	}
}

func TestValidateAwsAccountId(t *testing.T) {
	validNames := []string{
		"123456789012",
		"999999999999",
	}
	for _, v := range validNames {
		_, errors := validateAwsAccountId(v, "account_id")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid AWS Account ID: %q", v, errors)
		}
	}

	invalidNames := []string{
		"12345678901",   // too short
		"1234567890123", // too long
		"invalid",
		"x123456789012",
	}
	for _, v := range invalidNames {
		_, errors := validateAwsAccountId(v, "account_id")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid AWS Account ID", v)
		}
	}
}

func TestValidateArn(t *testing.T) {
	v := ""
	_, errors := validateArn(v, "arn")
	if len(errors) != 0 {
		t.Fatalf("%q should not be validated as an ARN: %q", v, errors)
	}

	validNames := []string{
		"arn:aws:elasticbeanstalk:us-east-1:123456789012:environment/My App/MyEnvironment", // lintignore:AWSAT003,AWSAT005 // Beanstalk
		"arn:aws:iam::123456789012:user/David",                                             // lintignore:AWSAT005          // IAM User
		"arn:aws:iam::aws:policy/CloudWatchReadOnlyAccess",                                 // lintignore:AWSAT005          // Managed IAM policy
		"arn:aws:rds:eu-west-1:123456789012:db:mysql-db",                                   // lintignore:AWSAT003,AWSAT005 // RDS
		"arn:aws:s3:::my_corporate_bucket/exampleobject.png",                               // lintignore:AWSAT005          // S3 object
		"arn:aws:events:us-east-1:319201112229:rule/rule_name",                             // lintignore:AWSAT003,AWSAT005 // CloudWatch Rule
		"arn:aws:lambda:eu-west-1:319201112229:function:myCustomFunction",                  // lintignore:AWSAT003,AWSAT005 // Lambda function
		"arn:aws:lambda:eu-west-1:319201112229:function:myCustomFunction:Qualifier",        // lintignore:AWSAT003,AWSAT005 // Lambda func qualifier
		"arn:aws-cn:ec2:cn-north-1:123456789012:instance/i-12345678",                       // lintignore:AWSAT003,AWSAT005 // China EC2 ARN
		"arn:aws-cn:s3:::bucket/object",                                                    // lintignore:AWSAT005          // China S3 ARN
		"arn:aws-iso:ec2:us-iso-east-1:123456789012:instance/i-12345678",                   // lintignore:AWSAT003,AWSAT005 // C2S EC2 ARN
		"arn:aws-iso:s3:::bucket/object",                                                   // lintignore:AWSAT005          // C2S S3 ARN
		"arn:aws-iso-b:ec2:us-isob-east-1:123456789012:instance/i-12345678",                // lintignore:AWSAT003,AWSAT005 // SC2S EC2 ARN
		"arn:aws-iso-b:s3:::bucket/object",                                                 // lintignore:AWSAT005          // SC2S S3 ARN
		"arn:aws-us-gov:ec2:us-gov-west-1:123456789012:instance/i-12345678",                // lintignore:AWSAT003,AWSAT005 // GovCloud EC2 ARN
		"arn:aws-us-gov:s3:::bucket/object",                                                // lintignore:AWSAT005          // GovCloud S3 ARN
	}
	for _, v := range validNames {
		_, errors := validateArn(v, "arn")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid ARN: %q", v, errors)
		}
	}

	invalidNames := []string{
		"arn",
		"123456789012",
		"arn:aws",
		"arn:aws:logs",            //lintignore:AWSAT005
		"arn:aws:logs:region:*:*", //lintignore:AWSAT005
	}
	for _, v := range invalidNames {
		_, errors := validateArn(v, "arn")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid ARN", v)
		}
	}
}

func TestValidatePrincipal(t *testing.T) {
	v := ""
	_, errors := validatePrincipal(v, "arn")
	if len(errors) == 0 {
		t.Fatalf("%q should not be validated as a principal %d: %q", v, len(errors), errors)
	}

	validNames := []string{
		"IAM_ALLOWED_PRINCIPALS", // Special principal
		"arn:aws-us-gov:iam::357342307427:role/tf-acc-test-3217321001347236965",          // lintignore:AWSAT005          // IAM Role
		"arn:aws:iam::123456789012:user/David",                                           // lintignore:AWSAT005          // IAM User
		"arn:aws-us-gov:iam:us-west-2:357342307427:role/tf-acc-test-3217321001347236965", // lintignore:AWSAT003,AWSAT005 // Non-global IAM Role?
		"arn:aws:iam:us-east-1:123456789012:user/David",                                  // lintignore:AWSAT003,AWSAT005 // Non-global IAM User?
	}
	for _, v := range validNames {
		_, errors := validatePrincipal(v, "arn")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid principal: %q", v, errors)
		}
	}

	invalidNames := []string{
		"IAM_NOT_ALLOWED_PRINCIPALS", // doesn't exist
		"arn",
		"123456789012",
		"arn:aws",
		"arn:aws:logs",            //lintignore:AWSAT005
		"arn:aws:logs:region:*:*", //lintignore:AWSAT005
		"arn:aws:elasticbeanstalk:us-east-1:123456789012:environment/My App/MyEnvironment", // lintignore:AWSAT003,AWSAT005 // not a user or role
		"arn:aws:iam::aws:policy/CloudWatchReadOnlyAccess",                                 // lintignore:AWSAT005          // not a user or role
		"arn:aws:rds:eu-west-1:123456789012:db:mysql-db",                                   // lintignore:AWSAT003,AWSAT005 // not a user or role
		"arn:aws:s3:::my_corporate_bucket/exampleobject.png",                               // lintignore:AWSAT005          // not a user or role
		"arn:aws:events:us-east-1:319201112229:rule/rule_name",                             // lintignore:AWSAT003,AWSAT005 // not a user or role
		"arn:aws-us-gov:ec2:us-gov-west-1:123456789012:instance/i-12345678",                // lintignore:AWSAT003,AWSAT005 // not a user or role
		"arn:aws-us-gov:s3:::bucket/object",                                                // lintignore:AWSAT005          // not a user or role
	}
	for _, v := range invalidNames {
		_, errors := validatePrincipal(v, "arn")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid principal", v)
		}
	}
}

func TestValidateEC2AutomateARN(t *testing.T) {
	validNames := []string{
		"arn:aws:automate:us-east-1:ec2:reboot",    //lintignore:AWSAT003,AWSAT005
		"arn:aws:automate:us-east-1:ec2:recover",   //lintignore:AWSAT003,AWSAT005
		"arn:aws:automate:us-east-1:ec2:stop",      //lintignore:AWSAT003,AWSAT005
		"arn:aws:automate:us-east-1:ec2:terminate", //lintignore:AWSAT003,AWSAT005
	}
	for _, v := range validNames {
		_, errors := validateEC2AutomateARN(v, "test_property")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid ARN: %q", v, errors)
		}
	}

	invalidNames := []string{
		"",
		"arn:aws:elasticbeanstalk:us-east-1:123456789012:environment/My App/MyEnvironment", // lintignore:AWSAT003,AWSAT005 // Beanstalk
		"arn:aws:iam::123456789012:user/David",                                             // lintignore:AWSAT005          // IAM User
		"arn:aws:rds:eu-west-1:123456789012:db:mysql-db",                                   // lintignore:AWSAT003,AWSAT005 // RDS
		"arn:aws:s3:::my_corporate_bucket/exampleobject.png",                               // lintignore:AWSAT005          // S3 object
		"arn:aws:events:us-east-1:319201112229:rule/rule_name",                             // lintignore:AWSAT003,AWSAT005 // CloudWatch Rule
		"arn:aws:lambda:eu-west-1:319201112229:function:myCustomFunction",                  // lintignore:AWSAT003,AWSAT005 // Lambda function
		"arn:aws:lambda:eu-west-1:319201112229:function:myCustomFunction:Qualifier",        // lintignore:AWSAT003,AWSAT005 // Lambda func qualifier
		"arn:aws-us-gov:s3:::corp_bucket/object.png",                                       // lintignore:AWSAT005          // GovCloud ARN
		"arn:aws-us-gov:kms:us-gov-west-1:123456789012:key/some-uuid-abc123",               // lintignore:AWSAT003,AWSAT005 // GovCloud KMS ARN
	}
	for _, v := range invalidNames {
		_, errors := validateEC2AutomateARN(v, "test_property")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid ARN", v)
		}
	}
}

func TestValidatePolicyStatementId(t *testing.T) {
	validNames := []string{
		"YadaHereAndThere",
		"Valid-5tatement_Id",
		"1234",
	}
	for _, v := range validNames {
		_, errors := validatePolicyStatementId(v, "statement_id")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid Statement ID: %q", v, errors)
		}
	}

	invalidNames := []string{
		"Invalid/StatementId/with/slashes",
		"InvalidStatementId.with.dots",
		// length > 100
		"TooooLoooooooooooooooooooooooooooooooooooooooooooo" +
			"ooooooooooooooooooooooooooooooooooooooooStatementId",
	}
	for _, v := range invalidNames {
		_, errors := validatePolicyStatementId(v, "statement_id")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid Statement ID", v)
		}
	}
}

func TestValidateCIDRNetworkAddress(t *testing.T) {
	cases := []struct {
		CIDR              string
		ExpectedErrSubstr string
	}{
		{"notacidr", `is not a valid CIDR block`},
		{"10.0.1.0/16", `is not a valid CIDR block; did you mean`},
		{"10.0.1.0/24", ``},
		{"2001:db8::/122", ``},
		{"2001::/15", `is not a valid CIDR block; did you mean`},
	}

	for i, tc := range cases {
		_, errs := validateCIDRNetworkAddress(tc.CIDR, "foo")
		if tc.ExpectedErrSubstr == "" {
			if len(errs) != 0 {
				t.Fatalf("%d/%d: Expected no error, got errs: %#v",
					i+1, len(cases), errs)
			}
		} else {
			if len(errs) != 1 {
				t.Fatalf("%d/%d: Expected 1 err containing %q, got %d errs",
					i+1, len(cases), tc.ExpectedErrSubstr, len(errs))
			}
			if !strings.Contains(errs[0].Error(), tc.ExpectedErrSubstr) {
				t.Fatalf("%d/%d: Expected err: %q, to include %q",
					i+1, len(cases), errs[0], tc.ExpectedErrSubstr)
			}
		}
	}
}

func TestValidateCIDRBlock(t *testing.T) {
	for _, ts := range []struct {
		cidr  string
		valid bool
	}{
		{"10.2.2.0/24", true},
		{"10.2.2.0/1234", false},
		{"10.2.2.2/24", false},
		{"::/0", true},
		{"::0/0", true},
		{"2000::/15", true},
		{"2001::/15", false},
		{"", false},
	} {
		err := validateCIDRBlock(ts.cidr)
		if !ts.valid && err == nil {
			t.Fatalf("Input '%s' should error but didn't!", ts.cidr)
		}
		if ts.valid && err != nil {
			t.Fatalf("Got unexpected error for '%s' input: %s", ts.cidr, err)
		}
	}
}

func TestValidateIpv4CIDRBlock(t *testing.T) {
	for _, ts := range []struct {
		cidr  string
		valid bool
	}{
		{"10.2.2.0/24", true},
		{"10.2.2.0/1234", false},
		{"10/24", false},
		{"10.2.2.2/24", false},
		{"::/0", false},
		{"2000::/15", false},
		{"", false},
	} {
		err := validateIpv4CIDRBlock(ts.cidr)
		if !ts.valid && err == nil {
			t.Fatalf("Input '%s' should error but didn't!", ts.cidr)
		}
		if ts.valid && err != nil {
			t.Fatalf("Got unexpected error for '%s' input: %s", ts.cidr, err)
		}
	}
}

func TestValidateIpv6CIDRBlock(t *testing.T) {
	for _, ts := range []struct {
		cidr  string
		valid bool
	}{
		{"10.2.2.0/24", false},
		{"10.2.2.0/1234", false},
		{"::/0", true},
		{"::0/0", true},
		{"2000::/15", true},
		{"2001::/15", false},
		{"2001:db8::/122", true},
		{"", false},
	} {
		err := validateIpv6CIDRBlock(ts.cidr)
		if !ts.valid && err == nil {
			t.Fatalf("Input '%s' should error but didn't!", ts.cidr)
		}
		if ts.valid && err != nil {
			t.Fatalf("Got unexpected error for '%s' input: %s", ts.cidr, err)
		}
	}
}

func TestCidrBlocksEqual(t *testing.T) {
	for _, ts := range []struct {
		cidr1 string
		cidr2 string
		equal bool
	}{
		{"10.2.2.0/24", "10.2.2.0/24", true},
		{"10.2.2.0/1234", "10.2.2.0/24", false},
		{"10.2.2.0/24", "10.2.2.0/1234", false},
		{"2001::/15", "2001::/15", true},
		{"::/0", "2001::/15", false},
		{"::/0", "::0/0", true},
		{"", "", false},
	} {
		equal := cidrBlocksEqual(ts.cidr1, ts.cidr2)
		if ts.equal != equal {
			t.Fatalf("cidrBlocksEqual(%q, %q) should be: %t", ts.cidr1, ts.cidr2, ts.equal)
		}
	}
}
func TestCanonicalCidrBlock(t *testing.T) {
	for _, ts := range []struct {
		cidr     string
		expected string
	}{
		{"10.2.2.0/24", "10.2.2.0/24"},
		{"10.2.2.5/24", "10.2.2.0/24"},
		{"::/0", "::/0"},
		{"::0/0", "::/0"},
		{"2001::/15", "2000::/15"},
		{"2001:db8::1/120", "2001:db8::/120"},
		{"", ""},
	} {
		got := canonicalCidrBlock(ts.cidr)
		if ts.expected != got {
			t.Fatalf("canonicalCidrBlock(%q) should be: %q, got: %q", ts.cidr, ts.expected, got)
		}
	}
}

func Test_canonicalCidrBlock(t *testing.T) {
	for _, ts := range []struct {
		cidr     string
		expected string
	}{
		{"10.2.2.0/24", "10.2.2.0/24"},
		{"::/0", "::/0"},
		{"::0/0", "::/0"},
		{"", ""},
	} {
		got := canonicalCidrBlock(ts.cidr)
		if ts.expected != got {
			t.Fatalf("canonicalCidrBlock(%q) should be: %q, got: %q", ts.cidr, ts.expected, got)
		}
	}
}

func TestValidateLogMetricFilterName(t *testing.T) {
	validNames := []string{
		"YadaHereAndThere",
		"Valid-5Metric_Name",
		"This . is also %% valid@!)+(",
		"1234",
		strings.Repeat("W", 512),
	}
	for _, v := range validNames {
		_, errors := validateLogMetricFilterName(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid Log Metric Filter Name: %q", v, errors)
		}
	}

	invalidNames := []string{
		"Here is a name with: colon",
		"and here is another * invalid name",
		"*",
		// length > 512
		strings.Repeat("W", 513),
	}
	for _, v := range invalidNames {
		_, errors := validateLogMetricFilterName(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid Log Metric Filter Name", v)
		}
	}
}

func TestValidateLogMetricTransformationName(t *testing.T) {
	validNames := []string{
		"YadaHereAndThere",
		"Valid-5Metric_Name",
		"This . is also %% valid@!)+(",
		"1234",
		"",
		strings.Repeat("W", 255),
	}
	for _, v := range validNames {
		_, errors := validateLogMetricFilterTransformationName(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid Log Metric Filter Transformation Name: %q", v, errors)
		}
	}

	invalidNames := []string{
		"Here is a name with: colon",
		"and here is another * invalid name",
		"also $ invalid",
		"*",
		// length > 255
		strings.Repeat("W", 256),
	}
	for _, v := range invalidNames {
		_, errors := validateLogMetricFilterTransformationName(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid Log Metric Filter Transformation Name", v)
		}
	}
}

func TestValidateLogGroupName(t *testing.T) {
	validNames := []string{
		"ValidLogGroupName",
		"ValidLogGroup.Name",
		"valid/Log-group",
		"1234",
		"YadaValid#0123",
		"Also_valid-name",
		strings.Repeat("W", 512),
	}
	for _, v := range validNames {
		_, errors := validateLogGroupName(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid Log Group name: %q", v, errors)
		}
	}

	invalidNames := []string{
		"Here is a name with: colon",
		"and here is another * invalid name",
		"also $ invalid",
		"This . is also %% invalid@!)+(",
		"*",
		"",
		// length > 512
		strings.Repeat("W", 513),
	}
	for _, v := range invalidNames {
		_, errors := validateLogGroupName(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid Log Group name", v)
		}
	}
}

func TestValidateLogGroupNamePrefix(t *testing.T) {
	validNames := []string{
		"ValidLogGroupName",
		"ValidLogGroup.Name",
		"valid/Log-group",
		"1234",
		"YadaValid#0123",
		"Also_valid-name",
		strings.Repeat("W", 483),
	}
	for _, v := range validNames {
		_, errors := validateLogGroupNamePrefix(v, "name_prefix")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid Log Group name prefix: %q", v, errors)
		}
	}

	invalidNames := []string{
		"Here is a name with: colon",
		"and here is another * invalid name",
		"also $ invalid",
		"This . is also %% invalid@!)+(",
		"*",
		"",
		// length > 483
		strings.Repeat("W", 484),
	}
	for _, v := range invalidNames {
		_, errors := validateLogGroupNamePrefix(v, "name_prefix")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid Log Group name prefix", v)
		}
	}
}

func TestValidateS3BucketLifecycleTimestamp(t *testing.T) {
	validDates := []string{
		"2016-01-01",
		"2006-01-02",
	}

	for _, v := range validDates {
		_, errors := validateS3BucketLifecycleTimestamp(v, "date")
		if len(errors) != 0 {
			t.Fatalf("%q should be valid date: %q", v, errors)
		}
	}

	invalidDates := []string{
		"Jan 01 2016",
		"20160101",
	}

	for _, v := range invalidDates {
		_, errors := validateS3BucketLifecycleTimestamp(v, "date")
		if len(errors) == 0 {
			t.Fatalf("%q should be invalid date", v)
		}
	}
}

func TestValidateSagemakerName(t *testing.T) {
	validNames := []string{
		"ValidSageMakerName",
		"Valid-5a63Mak3r-Name",
		"123-456-789",
		"1234",
		strings.Repeat("W", 63),
	}
	for _, v := range validNames {
		_, errors := validateSagemakerName(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid SageMaker name with maximum length 63 chars: %q", v, errors)
		}
	}

	invalidNames := []string{
		"Invalid name",          // blanks are not allowed
		"1#{}nook",              // other non-alphanumeric chars
		"-nook",                 // cannot start with hyphen
		strings.Repeat("W", 64), // length > 63
	}
	for _, v := range invalidNames {
		_, errors := validateSagemakerName(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid SageMaker name", v)
		}
	}
}

func TestValidateDbEventSubscriptionName(t *testing.T) {
	validNames := []string{
		"valid-name",
		"valid02-name",
		"Valid-Name1",
	}
	for _, v := range validNames {
		_, errors := validateDbEventSubscriptionName(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid RDS Event Subscription Name: %q", v, errors)
		}
	}

	invalidNames := []string{
		"Here is a name with: colon",
		"and here is another * invalid name",
		"also $ invalid",
		"This . is also %% invalid@!)+(",
		"*",
		"",
		" ",
		"_",
		// length > 255
		strings.Repeat("W", 256),
	}
	for _, v := range invalidNames {
		_, errors := validateDbEventSubscriptionName(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid RDS Event Subscription Name", v)
		}
	}
}

func TestValidateIAMPolicyJsonString(t *testing.T) {
	type testCases struct {
		Value    string
		ErrCount int
	}

	invalidCases := []testCases{
		{
			Value:    `{0:"1"}`,
			ErrCount: 1,
		},
		{
			Value:    `{'abc':1}`,
			ErrCount: 1,
		},
		{
			Value:    `{"def":}`,
			ErrCount: 1,
		},
		{
			Value:    `{"xyz":[}}`,
			ErrCount: 1,
		},
		{
			Value:    ``,
			ErrCount: 1,
		},
		{
			Value:    `    {"xyz": "foo"}`,
			ErrCount: 1,
		},
	}

	for _, tc := range invalidCases {
		_, errors := validateIAMPolicyJson(tc.Value, "json")
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected %q to trigger a validation error.", tc.Value)
		}
	}

	validCases := []testCases{
		{
			Value:    `{}`,
			ErrCount: 0,
		},
		{
			Value:    `{"abc":["1","2"]}`,
			ErrCount: 0,
		},
	}

	for _, tc := range validCases {
		_, errors := validateIAMPolicyJson(tc.Value, "json")
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected %q not to trigger a validation error.", tc.Value)
		}
	}
}

func TestValidateStringIsJsonOrYaml(t *testing.T) {
	type testCases struct {
		Value    string
		ErrCount int
	}

	invalidCases := []testCases{
		{
			Value:    `{"abc":"`,
			ErrCount: 1,
		},
		{
			Value:    "abc: [",
			ErrCount: 1,
		},
	}

	for _, tc := range invalidCases {
		_, errors := validateStringIsJsonOrYaml(tc.Value, "template")
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected %q to trigger a validation error.", tc.Value)
		}
	}

	validCases := []testCases{
		{
			Value:    `{"abc":"1"}`,
			ErrCount: 0,
		},
		{
			Value:    `abc: 1`,
			ErrCount: 0,
		},
	}

	for _, tc := range validCases {
		_, errors := validateStringIsJsonOrYaml(tc.Value, "template")
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected %q not to trigger a validation error.", tc.Value)
		}
	}
}

func TestValidateSQSQueueName(t *testing.T) {
	validNames := []string{
		"valid-name",
		"valid02-name",
		"Valid-Name1",
		"_",
		"-",
		strings.Repeat("W", 80),
	}
	for _, v := range validNames {
		if _, errors := validateSQSQueueName(v, "test_attribute"); len(errors) > 0 {
			t.Fatalf("%q should be a valid SQS queue Name", v)
		}

		if errors := validateSQSNonFifoQueueName(v); len(errors) > 0 {
			t.Fatalf("%q should be a valid SQS non-fifo queue Name", v)
		}
	}

	invalidNames := []string{
		"Here is a name with: colon",
		"another * invalid name",
		"also $ invalid",
		"This . is also %% invalid@!)+(",
		"*",
		"",
		" ",
		".",
		strings.Repeat("W", 81), // length > 80
	}
	for _, v := range invalidNames {
		if _, errors := validateSQSQueueName(v, "test_attribute"); len(errors) == 0 {
			t.Fatalf("%q should be an invalid SQS queue Name", v)
		}

		if errors := validateSQSNonFifoQueueName(v); len(errors) == 0 {
			t.Fatalf("%q should be an invalid SQS non-fifo queue Name", v)
		}
	}
}

func TestValidateSQSFifoQueueName(t *testing.T) {
	validNames := []string{
		"valid-name.fifo",
		"valid02-name.fifo",
		"Valid-Name1.fifo",
		"_.fifo",
		"a.fifo",
		"A.fifo",
		"9.fifo",
		"-.fifo",
		fmt.Sprintf("%s.fifo", strings.Repeat("W", 75)),
	}
	for _, v := range validNames {
		if _, errors := validateSQSQueueName(v, "test_attribute"); len(errors) > 0 {
			t.Fatalf("%q should be a valid SQS queue Name", v)
		}

		if errors := validateSQSFifoQueueName(v); len(errors) > 0 {
			t.Fatalf("%q should be a valid SQS FIFO queue Name: %v", v, errors)
		}
	}

	invalidNames := []string{
		"Here is a name with: colon",
		"another * invalid name",
		"also $ invalid",
		"This . is also %% invalid@!)+(",
		".fifo",
		"*",
		"",
		" ",
		".",
		strings.Repeat("W", 81), // length > 80
	}
	for _, v := range invalidNames {
		if _, errors := validateSQSQueueName(v, "test_attribute"); len(errors) == 0 {
			t.Fatalf("%q should be an invalid SQS queue Name", v)
		}

		if errors := validateSQSFifoQueueName(v); len(errors) == 0 {
			t.Fatalf("%q should be an invalid SQS FIFO queue Name: %v", v, errors)
		}
	}
}

func TestValidateOnceAWeekWindowFormat(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			// once a day window format
			Value:    "04:00-05:00",
			ErrCount: 1,
		},
		{
			// invalid day of week
			Value:    "san:04:00-san:05:00",
			ErrCount: 1,
		},
		{
			// invalid hour
			Value:    "sun:24:00-san:25:00",
			ErrCount: 1,
		},
		{
			// invalid min
			Value:    "sun:04:00-sun:04:60",
			ErrCount: 1,
		},
		{
			// valid format
			Value:    "sun:04:00-sun:05:00",
			ErrCount: 0,
		},
		{
			// "Sun" can also be used
			Value:    "Sun:04:00-Sun:05:00",
			ErrCount: 0,
		},
		{
			// valid format
			Value:    "",
			ErrCount: 0,
		},
	}

	for _, tc := range cases {
		_, errors := validateOnceAWeekWindowFormat(tc.Value, "maintenance_window")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected %d validation errors, But got %d errors for \"%s\"", tc.ErrCount, len(errors), tc.Value)
		}
	}
}

func TestValidateOnceADayWindowFormat(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			// once a week window format
			Value:    "sun:04:00-sun:05:00",
			ErrCount: 1,
		},
		{
			// invalid hour
			Value:    "24:00-25:00",
			ErrCount: 1,
		},
		{
			// invalid min
			Value:    "04:00-04:60",
			ErrCount: 1,
		},
		{
			// valid format
			Value:    "04:00-05:00",
			ErrCount: 0,
		},
		{
			// valid format
			Value:    "",
			ErrCount: 0,
		},
	}

	for _, tc := range cases {
		_, errors := validateOnceADayWindowFormat(tc.Value, "backup_window")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected %d validation errors, But got %d errors for \"%s\"", tc.ErrCount, len(errors), tc.Value)
		}
	}
}

func TestValidateEcsPlacementConstraint(t *testing.T) {
	cases := []struct {
		constType string
		constExpr string
		Err       bool
	}{
		{
			constType: "distinctInstance",
			constExpr: "",
			Err:       false,
		},
		{
			constType: "memberOf",
			constExpr: "",
			Err:       true,
		},
		{
			constType: "distinctInstance",
			constExpr: "expression",
			Err:       false,
		},
		{
			constType: "memberOf",
			constExpr: "expression",
			Err:       false,
		},
	}

	for _, tc := range cases {
		if err := validateAwsEcsPlacementConstraint(tc.constType, tc.constExpr); err != nil && !tc.Err {
			t.Fatalf("Unexpected validation error for \"%s:%s\": %s",
				tc.constType, tc.constExpr, err)
		}

	}
}

func TestValidateEcsPlacementStrategy(t *testing.T) {
	cases := []struct {
		stratType  string
		stratField string
		Err        bool
	}{
		{
			stratType:  "random",
			stratField: "",
			Err:        false,
		},
		{
			stratType:  "spread",
			stratField: "instanceID",
			Err:        false,
		},
		{
			stratType:  "binpack",
			stratField: "cpu",
			Err:        false,
		},
		{
			stratType:  "binpack",
			stratField: "memory",
			Err:        false,
		},
		{
			stratType:  "binpack",
			stratField: "disk",
			Err:        true,
		},
		{
			stratType:  "fakeType",
			stratField: "",
			Err:        true,
		},
	}

	for _, tc := range cases {
		if err := validateAwsEcsPlacementStrategy(tc.stratType, tc.stratField); err != nil && !tc.Err {
			t.Fatalf("Unexpected validation error for \"%s:%s\": %s",
				tc.stratType, tc.stratField, err)
		}
	}
}

func TestValidateStepFunctionStateMachineName(t *testing.T) {
	validTypes := []string{
		"foo",
		"BAR",
		"FooBar123",
		"FooBar123Baz-_",
	}

	invalidTypes := []string{
		"foo bar",
		"foo<bar>",
		"foo{bar}",
		"foo[bar]",
		"foo*bar",
		"foo?bar",
		"foo#bar",
		"foo%bar",
		"foo\bar",
		"foo^bar",
		"foo|bar",
		"foo~bar",
		"foo$bar",
		"foo&bar",
		"foo,bar",
		"foo:bar",
		"foo;bar",
		"foo/bar",
		strings.Repeat("W", 81), // length > 80
	}

	for _, v := range validTypes {
		_, errors := validateSfnStateMachineName(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid Step Function State Machine name: %v", v, errors)
		}
	}

	for _, v := range invalidTypes {
		_, errors := validateSfnStateMachineName(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid Step Function State Machine name", v)
		}
	}
}

func TestValidateEmrCustomAmiId(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "ami-dbcf88b1", //lintignore:AWSAT002
			ErrCount: 0,
		},
		{
			Value:    "vol-as7d65ash",
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validateAwsEmrCustomAmiId(tc.Value, "custom_ami_id")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected %d errors, got %d: %s", tc.ErrCount, len(errors), errors)
		}
	}
}

func TestValidateDmsEndpointId(t *testing.T) {
	validIds := []string{
		"tf-test-endpoint-1",
		"tfTestEndpoint",
	}

	for _, s := range validIds {
		_, errors := validateDmsEndpointId(s, "endpoint_id")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid endpoint id: %v", s, errors)
		}
	}

	invalidIds := []string{
		"tf_test_endpoint_1",
		"tf.test.endpoint.1",
		"tf test endpoint 1",
		"tf-test-endpoint-1!",
		"tf-test-endpoint-1-",
		"tf-test-endpoint--1",
		"tf-test-endpoint-1tf-test-endpoint-1tf-test-endpoint-1tf-test-endpoint-1tf-test-endpoint-1tf-test-endpoint-1tf-test-endpoint-1tf-test-endpoint-1tf-test-endpoint-1tf-test-endpoint-1tf-test-endpoint-1tf-test-endpoint-1tf-test-endpoint-1tf-test-endpoint-1tf-test-endpoint-1",
	}

	for _, s := range invalidIds {
		_, errors := validateDmsEndpointId(s, "endpoint_id")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid endpoint id: %v", s, errors)
		}
	}
}

func TestValidateDmsCertificateId(t *testing.T) {
	validIds := []string{
		"tf-test-certificate-1",
		"tfTestEndpoint",
	}

	for _, s := range validIds {
		_, errors := validateDmsCertificateId(s, "certificate_id")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid certificate id: %v", s, errors)
		}
	}

	invalidIds := []string{
		"tf_test_certificate_1",
		"tf.test.certificate.1",
		"tf test certificate 1",
		"tf-test-certificate-1!",
		"tf-test-certificate-1-",
		"tf-test-certificate--1",
		"tf-test-certificate-1tf-test-certificate-1tf-test-certificate-1tf-test-certificate-1tf-test-certificate-1tf-test-certificate-1tf-test-certificate-1tf-test-certificate-1tf-test-certificate-1tf-test-certificate-1tf-test-certificate-1tf-test-certificate-1tf-test-certificate-1tf-test-certificate-1tf-test-certificate-1",
	}

	for _, s := range invalidIds {
		_, errors := validateDmsEndpointId(s, "certificate_id")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid certificate id: %v", s, errors)
		}
	}
}

func TestValidateDmsReplicationInstanceId(t *testing.T) {
	validIds := []string{
		"tf-test-replication-instance-1",
		"tfTestReplicaitonInstance",
	}

	for _, s := range validIds {
		_, errors := validateDmsReplicationInstanceId(s, "replicaiton_instance_id")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid replication instance id: %v", s, errors)
		}
	}

	invalidIds := []string{
		"tf_test_replication-instance_1",
		"tf.test.replication.instance.1",
		"tf test replication instance 1",
		"tf-test-replication-instance-1!",
		"tf-test-replication-instance-1-",
		"tf-test-replication-instance--1",
		"tf-test-replication-instance-1tf-test-replication-instance-1tf-test-replication-instance-1",
	}

	for _, s := range invalidIds {
		_, errors := validateDmsReplicationInstanceId(s, "replication_instance_id")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid replication instance id: %v", s, errors)
		}
	}
}

func TestValidateDmsReplicationSubnetGroupId(t *testing.T) {
	validIds := []string{
		"tf-test-replication-subnet-group-1",
		"tf_test_replication_subnet_group_1",
		"tf.test.replication.subnet.group.1",
		"tf test replication subnet group 1",
		"tfTestReplicationSubnetGroup",
	}

	for _, s := range validIds {
		_, errors := validateDmsReplicationSubnetGroupId(s, "replication_subnet_group_id")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid replication subnet group id: %v", s, errors)
		}
	}

	invalidIds := []string{
		"default",
		"tf-test-replication-subnet-group-1!",
		"tf-test-replication-subnet-group-1tf-test-replication-subnet-group-1tf-test-replication-subnet-group-1tf-test-replication-subnet-group-1tf-test-replication-subnet-group-1tf-test-replication-subnet-group-1tf-test-replication-subnet-group-1tf-test-replication-subnet-group-1",
	}

	for _, s := range invalidIds {
		_, errors := validateDmsReplicationSubnetGroupId(s, "replication_subnet_group_id")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid replication subnet group id: %v", s, errors)
		}
	}
}

func TestValidateDmsReplicationTaskId(t *testing.T) {
	validIds := []string{
		"tf-test-replication-task-1",
		"tfTestReplicationTask",
	}

	for _, s := range validIds {
		_, errors := validateDmsReplicationTaskId(s, "replication_task_id")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid replication task id: %v", s, errors)
		}
	}

	invalidIds := []string{
		"tf_test_replication_task_1",
		"tf.test.replication.task.1",
		"tf test replication task 1",
		"tf-test-replication-task-1!",
		"tf-test-replication-task-1-",
		"tf-test-replication-task--1",
		"tf-test-replication-task-1tf-test-replication-task-1tf-test-replication-task-1tf-test-replication-task-1tf-test-replication-task-1tf-test-replication-task-1tf-test-replication-task-1tf-test-replication-task-1tf-test-replication-task-1tf-test-replication-task-1",
	}

	for _, s := range invalidIds {
		_, errors := validateDmsReplicationTaskId(s, "replication_task_id")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid replication task id: %v", s, errors)
		}
	}
}

func TestValidateAccountAlias(t *testing.T) {
	validAliases := []string{
		"tf-alias",
		"0tf-alias1",
	}

	for _, s := range validAliases {
		_, errors := validateAccountAlias(s, "account_alias")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid account alias: %v", s, errors)
		}
	}

	invalidAliases := []string{
		"tf",
		"-tf",
		"tf-",
		"TF-Alias",
		"tf-alias-tf-alias-tf-alias-tf-alias-tf-alias-tf-alias-tf-alias-tf-alias",
	}

	for _, s := range invalidAliases {
		_, errors := validateAccountAlias(s, "account_alias")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid account alias: %v", s, errors)
		}
	}
}

func TestValidateIamRoleProfileName(t *testing.T) {
	validNames := []string{
		"tf-test-role-profile-1",
	}

	for _, s := range validNames {
		_, errors := validateIamRolePolicyName(s, "name")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid IAM role policy name: %v", s, errors)
		}
	}

	invalidNames := []string{
		"invalid#name",
		"this-is-a-very-long-role-policy-name-this-is-a-very-long-role-policy-name-this-is-a-very-long-role-policy-name-this-is-a-very-long",
	}

	for _, s := range invalidNames {
		_, errors := validateIamRolePolicyName(s, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid IAM role policy name: %v", s, errors)
		}
	}
}

func TestValidateIamRoleProfileNamePrefix(t *testing.T) {
	validNamePrefixes := []string{
		"tf-test-role-profile-",
	}

	for _, s := range validNamePrefixes {
		_, errors := validateIamRolePolicyNamePrefix(s, "name_prefix")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid IAM role policy name prefix: %v", s, errors)
		}
	}

	invalidNamePrefixes := []string{
		"invalid#name_prefix",
		"this-is-a-very-long-role-policy-name-prefix-this-is-a-very-long-role-policy-name-prefix-this-is-a-very-",
	}

	for _, s := range invalidNamePrefixes {
		_, errors := validateIamRolePolicyNamePrefix(s, "name_prefix")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid IAM role policy name prefix: %v", s, errors)
		}
	}
}

func TestValidateApiGatewayUsagePlanQuotaSettings(t *testing.T) {
	cases := []struct {
		Offset   int
		Period   string
		ErrCount int
	}{
		{
			Offset:   0,
			Period:   "DAY",
			ErrCount: 0,
		},
		{
			Offset:   -1,
			Period:   "DAY",
			ErrCount: 1,
		},
		{
			Offset:   1,
			Period:   "DAY",
			ErrCount: 1,
		},
		{
			Offset:   0,
			Period:   "WEEK",
			ErrCount: 0,
		},
		{
			Offset:   6,
			Period:   "WEEK",
			ErrCount: 0,
		},
		{
			Offset:   -1,
			Period:   "WEEK",
			ErrCount: 1,
		},
		{
			Offset:   7,
			Period:   "WEEK",
			ErrCount: 1,
		},
		{
			Offset:   0,
			Period:   "MONTH",
			ErrCount: 0,
		},
		{
			Offset:   27,
			Period:   "MONTH",
			ErrCount: 0,
		},
		{
			Offset:   -1,
			Period:   "MONTH",
			ErrCount: 1,
		},
		{
			Offset:   28,
			Period:   "MONTH",
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		m := make(map[string]interface{})
		m["offset"] = tc.Offset
		m["period"] = tc.Period

		errors := validateApiGatewayUsagePlanQuotaSettings(m)
		if len(errors) != tc.ErrCount {
			t.Fatalf("API Gateway Usage Plan Quota Settings validation failed: %v", errors)
		}
	}
}

func TestValidateDocDBIdentifier(t *testing.T) {
	validNames := []string{
		"a",
		"hello-world",
		"hello-world-0123456789",
		strings.Repeat("w", 63),
	}
	for _, v := range validNames {
		_, errors := validateDocDBIdentifier(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid DocDB Identifier: %q", v, errors)
		}
	}

	invalidNames := []string{
		"",
		"special@character",
		"slash/in-the-middle",
		"dot.in-the-middle",
		"two-hyphen--in-the-middle",
		"0-first-numeric",
		"-first-hyphen",
		"end-hyphen-",
		strings.Repeat("W", 64),
	}
	for _, v := range invalidNames {
		_, errors := validateDocDBIdentifier(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid DocDB Identifier", v)
		}
	}
}

func TestValidateElbName(t *testing.T) {
	validNames := []string{
		"tf-test-elb",
	}

	for _, s := range validNames {
		_, errors := validateElbName(s, "name")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid ELB name: %v", s, errors)
		}
	}

	invalidNames := []string{
		"tf.test.elb.1",
		"tf-test-elb-tf-test-elb-tf-test-elb",
		"-tf-test-elb",
		"tf-test-elb-",
	}

	for _, s := range invalidNames {
		_, errors := validateElbName(s, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid ELB name: %v", s, errors)
		}
	}
}

func TestValidateElbNamePrefix(t *testing.T) {
	validNamePrefixes := []string{
		"test-",
	}

	for _, s := range validNamePrefixes {
		_, errors := validateElbNamePrefix(s, "name_prefix")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid ELB name prefix: %v", s, errors)
		}
	}

	invalidNamePrefixes := []string{
		"tf.test.elb.",
		"tf-test",
		"-test",
	}

	for _, s := range invalidNamePrefixes {
		_, errors := validateElbNamePrefix(s, "name_prefix")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid ELB name prefix: %v", s, errors)
		}
	}
}

func TestValidateNeptuneEventSubscriptionName(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "testing123!",
			ErrCount: 1,
		},
		{
			Value:    "testing 123",
			ErrCount: 1,
		},
		{
			Value:    "testing_123",
			ErrCount: 1,
		},
		{
			Value:    acctest.RandStringFromCharSet(256, acctest.CharSetAlpha),
			ErrCount: 1,
		},
	}
	for _, tc := range cases {
		_, errors := validateNeptuneEventSubscriptionName(tc.Value, "aws_neptune_event_subscription")
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the Neptune Event Subscription Name to trigger a validation error for %q", tc.Value)
		}
	}
}

func TestValidateNeptuneEventSubscriptionNamePrefix(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "testing123!",
			ErrCount: 1,
		},
		{
			Value:    "testing 123",
			ErrCount: 1,
		},
		{
			Value:    "testing_123",
			ErrCount: 1,
		},
		{
			Value:    acctest.RandStringFromCharSet(254, acctest.CharSetAlpha),
			ErrCount: 1,
		},
	}
	for _, tc := range cases {
		_, errors := validateNeptuneEventSubscriptionNamePrefix(tc.Value, "aws_neptune_event_subscription")
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the Neptune Event Subscription Name Prefix to trigger a validation error for %q", tc.Value)
		}
	}
}

func TestValidateDbSubnetGroupName(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "tEsting",
			ErrCount: 1,
		},
		{
			Value:    "testing?",
			ErrCount: 1,
		},
		{
			Value:    "default",
			ErrCount: 1,
		},
		{
			Value:    acctest.RandStringFromCharSet(300, acctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validateDbSubnetGroupName(tc.Value, "aws_db_subnet_group")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the DB Subnet Group name to trigger a validation error")
		}
	}
}

func TestValidateNeptuneSubnetGroupName(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "tEsting",
			ErrCount: 1,
		},
		{
			Value:    "testing?",
			ErrCount: 1,
		},
		{
			Value:    "default",
			ErrCount: 1,
		},
		{
			Value:    acctest.RandStringFromCharSet(300, acctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validateNeptuneSubnetGroupName(tc.Value, "aws_neptune_subnet_group")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the Neptune Subnet Group name to trigger a validation error")
		}
	}
}

func TestValidateDbSubnetGroupNamePrefix(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "tEsting",
			ErrCount: 1,
		},
		{
			Value:    "testing?",
			ErrCount: 1,
		},
		{
			Value:    acctest.RandStringFromCharSet(230, acctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validateDbSubnetGroupNamePrefix(tc.Value, "aws_db_subnet_group")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the DB Subnet Group name prefix to trigger a validation error")
		}
	}
}

func TestValidateNeptuneSubnetGroupNamePrefix(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "tEsting",
			ErrCount: 1,
		},
		{
			Value:    "testing?",
			ErrCount: 1,
		},
		{
			Value:    acctest.RandStringFromCharSet(230, acctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validateNeptuneSubnetGroupNamePrefix(tc.Value, "aws_neptune_subnet_group")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the Neptune Subnet Group name prefix to trigger a validation error")
		}
	}
}

func TestValidateDbOptionGroupName(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "testing123!",
			ErrCount: 1,
		},
		{
			Value:    "1testing123",
			ErrCount: 1,
		},
		{
			Value:    "testing--123",
			ErrCount: 1,
		},
		{
			Value:    "testing123-",
			ErrCount: 1,
		},
		{
			Value:    acctest.RandStringFromCharSet(256, acctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validateDbOptionGroupName(tc.Value, "aws_db_option_group_name")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the DB Option Group Name to trigger a validation error")
		}
	}
}

func TestValidateDbOptionGroupNamePrefix(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "testing123!",
			ErrCount: 1,
		},
		{
			Value:    "1testing123",
			ErrCount: 1,
		},
		{
			Value:    "testing--123",
			ErrCount: 1,
		},
		{
			Value:    acctest.RandStringFromCharSet(230, acctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validateDbOptionGroupNamePrefix(tc.Value, "aws_db_option_group_name")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the DB Option Group name prefix to trigger a validation error")
		}
	}
}

func TestValidateDbParamGroupName(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "tEsting123",
			ErrCount: 1,
		},
		{
			Value:    "testing123!",
			ErrCount: 1,
		},
		{
			Value:    "1testing123",
			ErrCount: 1,
		},
		{
			Value:    "testing--123",
			ErrCount: 1,
		},
		{
			Value:    "testing_123",
			ErrCount: 1,
		},
		{
			Value:    "testing123-",
			ErrCount: 1,
		},
		{
			Value:    acctest.RandStringFromCharSet(256, acctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validateDbParamGroupName(tc.Value, "aws_db_parameter_group_name")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the DB Parameter Group Name to trigger a validation error")
		}
	}
}

func TestValidateOpenIdURL(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "http://wrong.scheme.com",
			ErrCount: 1,
		},
		{
			Value:    "ftp://wrong.scheme.co.uk",
			ErrCount: 1,
		},
		{
			Value:    "%@invalidUrl",
			ErrCount: 1,
		},
		{
			Value:    "https://example.com/?query=param",
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validateOpenIdURL(tc.Value, "url")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected %d of OpenID URL validation errors, got %d", tc.ErrCount, len(errors))
		}
	}
}

func TestValidateAwsKmsName(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "alias/aws/s3",
			ErrCount: 0,
		},
		{
			Value:    "alias/hashicorp",
			ErrCount: 0,
		},
		{
			Value:    "hashicorp",
			ErrCount: 1,
		},
		{
			Value:    "hashicorp/terraform",
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validateAwsKmsName(tc.Value, "name")
		if len(errors) != tc.ErrCount {
			t.Fatalf("AWS KMS Alias Name validation failed: %v", errors)
		}
	}
}

func TestValidateAwsKmsGrantName(t *testing.T) {
	validValues := []string{
		"123",
		"Abc",
		"grant_1",
		"grant:/-",
	}

	for _, s := range validValues {
		_, errors := validateAwsKmsGrantName(s, "name")
		if len(errors) > 0 {
			t.Fatalf("%q AWS KMS Grant Name should have been valid: %v", s, errors)
		}
	}

	invalidValues := []string{
		strings.Repeat("w", 257),
		"grant.invalid",
		";",
		"white space",
	}

	for _, s := range invalidValues {
		_, errors := validateAwsKmsGrantName(s, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid AWS KMS Grant Name", s)
		}
	}
}

func TestValidateCognitoIdentityPoolName(t *testing.T) {
	validValues := []string{
		"123",
		"1 2 3",
		"foo",
		"foo bar",
		"foo_bar",
		"1foo 2bar 3",
		"foo-bar_123",
		"foo-bar",
	}

	for _, s := range validValues {
		_, errors := validateCognitoIdentityPoolName(s, "identity_pool_name")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid Cognito Identity Pool Name: %v", s, errors)
		}
	}

	invalidValues := []string{
		"foo*",
		"foo:bar",
		"foo&bar",
		"foo1^bar2",
	}

	for _, s := range invalidValues {
		_, errors := validateCognitoIdentityPoolName(s, "identity_pool_name")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid Cognito Identity Pool Name: %v", s, errors)
		}
	}
}

func TestValidateCognitoProviderDeveloperName(t *testing.T) {
	validValues := []string{
		"1",
		"foo",
		"1.2",
		"foo1-bar2-baz3",
		"foo_bar",
	}

	for _, s := range validValues {
		_, errors := validateCognitoProviderDeveloperName(s, "developer_provider_name")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid Cognito Provider Developer Name: %v", s, errors)
		}
	}

	invalidValues := []string{
		"foo!",
		"foo:bar",
		"foo/bar",
		"foo;bar",
	}

	for _, s := range invalidValues {
		_, errors := validateCognitoProviderDeveloperName(s, "developer_provider_name")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid Cognito Provider Developer Name: %v", s, errors)
		}
	}
}

func TestValidateCognitoSupportedLoginProviders(t *testing.T) {
	validValues := []string{
		"foo",
		"7346241598935552",
		"123456789012.apps.googleusercontent.com",
		"foo_bar",
		"foo;bar",
		"foo/bar",
		"foo-bar",
		"xvz1evFS4wEEPTGEFPHBog;kAcSOqF21Fu85e7zjz7ZN2U4ZRhfV3WpwPAoE3Z7kBw",
		strings.Repeat("W", 128),
	}

	for _, s := range validValues {
		_, errors := validateCognitoSupportedLoginProviders(s, "supported_login_providers")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid Cognito Supported Login Providers: %v", s, errors)
		}
	}

	invalidValues := []string{
		"",
		strings.Repeat("W", 129), // > 128
		"foo:bar_baz",
		"foobar,foobaz",
		"foobar=foobaz",
	}

	for _, s := range invalidValues {
		_, errors := validateCognitoSupportedLoginProviders(s, "supported_login_providers")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid Cognito Supported Login Providers: %v", s, errors)
		}
	}
}

func TestValidateCognitoIdentityProvidersClientId(t *testing.T) {
	validValues := []string{
		"7lhlkkfbfb4q5kpp90urffao",
		"12345678",
		"foo_123",
		strings.Repeat("W", 128),
	}

	for _, s := range validValues {
		_, errors := validateCognitoIdentityProvidersClientId(s, "client_id")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid Cognito Identity Provider Client ID: %v", s, errors)
		}
	}

	invalidValues := []string{
		"",
		strings.Repeat("W", 129), // > 128
		"foo-bar",
		"foo:bar",
		"foo;bar",
	}

	for _, s := range invalidValues {
		_, errors := validateCognitoIdentityProvidersClientId(s, "client_id")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid Cognito Identity Provider Client ID: %v", s, errors)
		}
	}
}

func TestValidateCognitoIdentityProvidersProviderName(t *testing.T) {
	validValues := []string{
		"foo",
		"7346241598935552",
		"foo_bar",
		"foo:bar",
		"foo/bar",
		"foo-bar",
		"cognito-idp.us-east-1.amazonaws.com/us-east-1_Zr231apJu", //lintignore:AWSAT003
		strings.Repeat("W", 128),
	}

	for _, s := range validValues {
		_, errors := validateCognitoIdentityProvidersProviderName(s, "provider_name")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid Cognito Identity Provider Name: %v", s, errors)
		}
	}

	invalidValues := []string{
		"",
		strings.Repeat("W", 129), // > 128
		"foo;bar_baz",
		"foobar,foobaz",
		"foobar=foobaz",
	}

	for _, s := range invalidValues {
		_, errors := validateCognitoIdentityProvidersProviderName(s, "provider_name")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid Cognito Identity Provider Name: %v", s, errors)
		}
	}
}

func TestValidateCognitoUserPoolEmailVerificationMessage(t *testing.T) {
	validValues := []string{
		"{####}",
		"Foo {####}",
		"{####} Bar",
		"AZERTYUIOPQSDFGHJKLMWXCVBN?./+%£*¨°0987654321&é\"'(§è!çà)-@^'{####},=ù`$|´”’[å»ÛÁØ]–Ô¥#‰±•",
		"{####}" + strings.Repeat("W", 19994), // = 20000
	}

	for _, s := range validValues {
		_, errors := validateCognitoUserPoolEmailVerificationMessage(s, "email_verification_message")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid Cognito User Pool email verification message: %v", s, errors)
		}
	}

	invalidValues := []string{
		"Foo",
		"{###}",
		"{####}" + strings.Repeat("W", 19995), // > 20000
	}

	for _, s := range invalidValues {
		_, errors := validateCognitoUserPoolEmailVerificationMessage(s, "email_verification_message")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid Cognito User Pool email verification message: %v", s, errors)
		}
	}
}

func TestValidateCognitoUserPoolEmailVerificationSubject(t *testing.T) {
	validValues := []string{
		"FooBar",
		"AZERTYUIOPQSDFGHJKLMWXCVBN?./+%£*¨°0987654321&é\" '(§è!çà)-@^'{####},=ù`$|´”’[å»ÛÁØ]–Ô¥#‰±•",
		"Foo Bar", // special whitespace character
		strings.Repeat("W", 140),
	}

	for _, s := range validValues {
		_, errors := validateCognitoUserPoolEmailVerificationSubject(s, "email_verification_subject")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid Cognito User Pool email verification subject: %v", s, errors)
		}
	}

	invalidValues := []string{
		"Foo",
		strings.Repeat("W", 141),
	}

	for _, s := range invalidValues {
		_, errors := validateCognitoUserPoolEmailVerificationSubject(s, "email_verification_subject")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid Cognito User Pool email verification subject: %v", s, errors)
		}
	}
}

func TestValidateCognitoUserPoolSmsAuthenticationMessage(t *testing.T) {
	validValues := []string{
		"{####}",
		"Foo {####}",
		"{####} Bar",
		"AZERTYUIOPQSDFGHJKLMWXCVBN?./+%£*¨°0987654321&é\"'(§è!çà)-@^'{####},=ù`$|´”’[å»ÛÁØ]–Ô¥#‰±•",
		"{####}" + strings.Repeat("W", 134), // = 140
	}

	for _, s := range validValues {
		_, errors := validateCognitoUserPoolSmsAuthenticationMessage(s, "sms_authentication_message")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid Cognito User Pool sms authentication message: %v", s, errors)
		}
	}

	invalidValues := []string{
		"Foo",
		"{####}" + strings.Repeat("W", 135),
	}

	for _, s := range invalidValues {
		_, errors := validateCognitoUserPoolSmsAuthenticationMessage(s, "sms_authentication_message")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid Cognito User Pool sms authentication message: %v", s, errors)
		}
	}
}

func TestValidateCognitoUserPoolSmsVerificationMessage(t *testing.T) {
	validValues := []string{
		"{####}",
		"Foo {####}",
		"{####} Bar",
		"AZERTYUIOPQSDFGHJKLMWXCVBN?./+%£*¨°0987654321&é\"'(§è!çà)-@^'{####},=ù`$|´”’[å»ÛÁØ]–Ô¥#‰±•",
		"{####}" + strings.Repeat("W", 134), // = 140
	}

	for _, s := range validValues {
		_, errors := validateCognitoUserPoolSmsVerificationMessage(s, "sms_verification_message")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid Cognito User Pool sms authentication message: %v", s, errors)
		}
	}

	invalidValues := []string{
		"Foo",
		"{####}" + strings.Repeat("W", 135),
	}

	for _, s := range invalidValues {
		_, errors := validateCognitoUserPoolSmsVerificationMessage(s, "sms_verification_message")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid Cognito User Pool sms authentication message: %v", s, errors)
		}
	}
}

func TestValidateWafMetricName(t *testing.T) {
	validNames := []string{
		"testrule",
		"testRule",
		"testRule123",
	}
	for _, v := range validNames {
		_, errors := validateWafMetricName(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid WAF metric name: %q", v, errors)
		}
	}

	invalidNames := []string{
		"!",
		"/",
		" ",
		":",
		";",
		"white space",
		"/slash-at-the-beginning",
		"slash-at-the-end/",
	}
	for _, v := range invalidNames {
		_, errors := validateWafMetricName(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid WAF metric name", v)
		}
	}
}

func TestValidateAwsSSMName(t *testing.T) {
	validNames := []string{
		".foo-bar_123",
		strings.Repeat("W", 128),
	}
	for _, v := range validNames {
		_, errors := validateAwsSSMName(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid SSM Name: %q", v, errors)
		}
	}

	invalidNames := []string{
		"foo+bar",
		"tf",
		strings.Repeat("W", 129), // > 128
	}
	for _, v := range invalidNames {
		_, errors := validateAwsSSMName(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid SSM Name: %q", v, errors)
		}
	}
}

func TestValidateBatchName(t *testing.T) {
	validNames := []string{
		strings.Repeat("W", 128), // <= 128
	}
	for _, v := range validNames {
		_, errors := validateBatchName(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid Batch name: %q", v, errors)
		}
	}

	invalidNames := []string{
		"s@mple",
		strings.Repeat("W", 129), // >= 129
	}
	for _, v := range invalidNames {
		_, errors := validateBatchName(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be a invalid Batch name: %q", v, errors)
		}
	}
}

func TestValidateBatchPrefix(t *testing.T) {
	validPrefixes := []string{
		strings.Repeat("W", 102), // <= 102
	}
	for _, v := range validPrefixes {
		_, errors := validateBatchPrefix(v, "prefix")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid Batch prefix: %q", v, errors)
		}
	}

	invalidPrefixes := []string{
		"s@mple",
		strings.Repeat("W", 103), // >= 103
	}
	for _, v := range invalidPrefixes {
		_, errors := validateBatchPrefix(v, "prefix")
		if len(errors) == 0 {
			t.Fatalf("%q should be a invalid Batch prefix: %q", v, errors)
		}
	}
}

func TestValidateCognitoRoleMappingsAmbiguousRoleResolutionAgainstType(t *testing.T) {
	cases := []struct {
		AmbiguousRoleResolution interface{}
		Type                    string
		ErrCount                int
	}{
		{
			AmbiguousRoleResolution: nil,
			Type:                    cognitoidentity.RoleMappingTypeToken,
			ErrCount:                1,
		},
		{
			AmbiguousRoleResolution: "foo",
			Type:                    cognitoidentity.RoleMappingTypeToken,
			ErrCount:                0, // 0 as it should be defined, the value isn't validated here
		},
		{
			AmbiguousRoleResolution: cognitoidentity.AmbiguousRoleResolutionTypeAuthenticatedRole,
			Type:                    cognitoidentity.RoleMappingTypeToken,
			ErrCount:                0,
		},
		{
			AmbiguousRoleResolution: cognitoidentity.AmbiguousRoleResolutionTypeDeny,
			Type:                    cognitoidentity.RoleMappingTypeToken,
			ErrCount:                0,
		},
	}

	for _, tc := range cases {
		m := make(map[string]interface{})
		// Reproducing the undefined ambiguous_role_resolution
		if tc.AmbiguousRoleResolution != nil {
			m["ambiguous_role_resolution"] = tc.AmbiguousRoleResolution
		}
		m["type"] = tc.Type

		errors := validateCognitoRoleMappingsAmbiguousRoleResolutionAgainstType(m)
		if len(errors) != tc.ErrCount {
			t.Fatalf("Cognito Role Mappings validation failed: %v, expected err count %d, got %d, for config %#v", errors, tc.ErrCount, len(errors), m)
		}
	}
}

func TestValidateCognitoRoleMappingsRulesConfiguration(t *testing.T) {
	cases := []struct {
		MappingRule []interface{}
		Type        string
		ErrCount    int
	}{
		{
			MappingRule: nil,
			Type:        cognitoidentity.RoleMappingTypeRules,
			ErrCount:    1,
		},
		{
			MappingRule: []interface{}{
				map[string]interface{}{
					"Claim":     "isAdmin",
					"MatchType": "Equals",
					"RoleARN":   "arn:foo",
					"Value":     "paid",
				},
			},
			Type:     cognitoidentity.RoleMappingTypeRules,
			ErrCount: 0,
		},
		{
			MappingRule: []interface{}{
				map[string]interface{}{
					"Claim":     "isAdmin",
					"MatchType": "Equals",
					"RoleARN":   "arn:foo",
					"Value":     "paid",
				},
			},
			Type:     cognitoidentity.RoleMappingTypeToken,
			ErrCount: 1,
		},
		{
			MappingRule: nil,
			Type:        cognitoidentity.RoleMappingTypeToken,
			ErrCount:    0,
		},
	}

	for _, tc := range cases {
		m := make(map[string]interface{})
		// Reproducing the undefined mapping_rule
		if tc.MappingRule != nil {
			m["mapping_rule"] = tc.MappingRule
		}
		m["type"] = tc.Type

		errors := validateCognitoRoleMappingsRulesConfiguration(m)
		if len(errors) != tc.ErrCount {
			t.Fatalf("Cognito Role Mappings validation failed: %v, expected err count %d, got %d, for config %#v", errors, tc.ErrCount, len(errors), m)
		}
	}
}

func TestValidateSecurityGroupRuleDescription(t *testing.T) {
	validDescriptions := []string{
		"testrule",
		"testRule",
		"testRule 123",
		`testRule 123 ._-:/()#,@[]+=&;{}!$*`,
	}
	for _, v := range validDescriptions {
		_, errors := validateSecurityGroupRuleDescription(v, "description")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid security group rule description: %q", v, errors)
		}
	}

	invalidDescriptions := []string{
		"`",
		"%%",
		`\`,
	}
	for _, v := range invalidDescriptions {
		_, errors := validateSecurityGroupRuleDescription(v, "description")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid security group rule description", v)
		}
	}
}

func TestValidateCognitoRoles(t *testing.T) {
	validValues := []map[string]interface{}{
		{"authenticated": "hoge"},
		{"unauthenticated": "hoge"},
		{"authenticated": "hoge", "unauthenticated": "hoge"},
	}

	for _, s := range validValues {
		errors := validateCognitoRoles(s)
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid Cognito Roles: %v", s, errors)
		}
	}

	invalidValues := []map[string]interface{}{
		{},
		{"invalid": "hoge"},
	}

	for _, s := range invalidValues {
		errors := validateCognitoRoles(s)
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid Cognito Roles: %v", s, errors)
		}
	}
}

func TestValidateKmsKey(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "arbitrary-uuid-1234",
			ErrCount: 0,
		},
		{
			Value:    "arn:aws:kms:us-west-2:111122223333:key/arbitrary-uuid-1234", //lintignore:AWSAT003,AWSAT005
			ErrCount: 0,
		},
		{
			Value:    "alias/arbitrary-key",
			ErrCount: 0,
		},
		{
			Value:    "alias/arbitrary/key",
			ErrCount: 0,
		},
		{
			Value:    "arn:aws:kms:us-west-2:111122223333:alias/arbitrary-key", //lintignore:AWSAT003,AWSAT005
			ErrCount: 0,
		},
		{
			Value:    "arn:aws:kms:us-west-2:111122223333:alias/arbitrary/key", //lintignore:AWSAT003,AWSAT005
			ErrCount: 0,
		},
		{
			Value:    "$%wrongkey",
			ErrCount: 1,
		},
		{
			Value:    "arn:aws:lamda:foo:bar:key/xyz", //lintignore:AWSAT003,AWSAT005
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validateKmsKey(tc.Value, "key_id")
		if len(errors) != tc.ErrCount {
			t.Fatalf("%q validation failed: %v", tc.Value, errors)
		}
	}
}

func TestResourceAWSElastiCacheReplicationGroupAuthTokenValidation(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "this-is-valid!#%()^",
			ErrCount: 0,
		},
		{
			Value:    "this-is-not",
			ErrCount: 1,
		},
		{
			Value:    "this-is-not-valid\"",
			ErrCount: 1,
		},
		{
			Value:    "this-is-not-valid@",
			ErrCount: 1,
		},
		{
			Value:    "this-is-not-valid/",
			ErrCount: 1,
		},
		{
			Value:    acctest.RandStringFromCharSet(129, acctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validateAwsElastiCacheReplicationGroupAuthToken(tc.Value, "aws_elasticache_replication_group_auth_token")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the ElastiCache Replication Group AuthToken to trigger a validation error")
		}
	}
}

func TestValidateCognitoUserGroupName(t *testing.T) {
	validValues := []string{
		"foo",
		"7346241598935552",
		"foo_bar",
		"foo:bar",
		"foo/bar",
		"foo-bar",
		"$foobar",
		strings.Repeat("W", 128),
	}

	for _, s := range validValues {
		_, errors := validateCognitoUserGroupName(s, "name")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid Cognito User Pool Group Name: %v", s, errors)
		}
	}

	invalidValues := []string{
		"",
		strings.Repeat("W", 129), // > 128
	}

	for _, s := range invalidValues {
		_, errors := validateCognitoUserGroupName(s, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid Cognito User Pool Group Name: %v", s, errors)
		}
	}
}

func TestValidateCognitoUserPoolId(t *testing.T) {
	validValues := []string{
		"eu-west-1_Foo123",         //lintignore:AWSAT003
		"ap-southeast-2_BaRBaz987", //lintignore:AWSAT003
	}

	for _, s := range validValues {
		_, errors := validateCognitoUserPoolId(s, "user_pool_id")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid Cognito User Pool Id: %v", s, errors)
		}
	}

	invalidValues := []string{
		"",
		"foo",
		"us-east-1-Foo123",   //lintignore:AWSAT003
		"eu-central-2_Bar+4", //lintignore:AWSAT003
	}

	for _, s := range invalidValues {
		_, errors := validateCognitoUserPoolId(s, "user_pool_id")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid Cognito User Pool Id: %v", s, errors)
		}
	}
}

func TestValidateAmazonSideAsn(t *testing.T) {
	validAsns := []string{
		"7224",
		"9059",
		"10124",
		"17493",
		"64512",
		"64513",
		"65533",
		"65534",
		"4200000000",
		"4200000001",
		"4294967293",
		"4294967294",
	}
	for _, v := range validAsns {
		_, errors := validateAmazonSideAsn(v, "amazon_side_asn")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid ASN: %q", v, errors)
		}
	}

	invalidAsns := []string{
		"1",
		"ABCDEFG",
		"",
		"7225",
		"9058",
		"10125",
		"17492",
		"64511",
		"65535",
		"4199999999",
		"4294967295",
		"9999999999",
	}
	for _, v := range invalidAsns {
		_, errors := validateAmazonSideAsn(v, "amazon_side_asn")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid ASN", v)
		}
	}
}

func TestValidate4ByteAsn(t *testing.T) {
	validAsns := []string{
		"0",
		"1",
		"65534",
		"65535",
		"4294967294",
		"4294967295",
	}
	for _, v := range validAsns {
		_, errors := validate4ByteAsn(v, "bgp_asn")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid ASN: %q", v, errors)
		}
	}

	invalidAsns := []string{
		"-1",
		"ABCDEFG",
		"",
		"4294967296",
		"9999999999",
	}
	for _, v := range invalidAsns {
		_, errors := validate4ByteAsn(v, "bgp_asn")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid ASN", v)
		}
	}
}

func TestValidateLaunchTemplateName(t *testing.T) {
	validNames := []string{
		"fooBAR123",
		"(./_)",
	}
	for _, v := range validNames {
		_, errors := validateLaunchTemplateName(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid Launch Template name: %q", v, errors)
		}
	}

	invalidNames := []string{
		"tf",
		strings.Repeat("W", 126), // > 125
		"invalid*",
		"invalid\name",
		"inavalid&",
		"invalid+",
		"invalid!",
		"invalid:",
		"invalid;",
	}
	for _, v := range invalidNames {
		_, errors := validateLaunchTemplateName(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid Launch Template name: %q", v, errors)
		}
	}

	invalidNamePrefixes := []string{
		strings.Repeat("W", 100), // > 99
	}
	for _, v := range invalidNamePrefixes {
		_, errors := validateLaunchTemplateName(v, "name_prefix")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid Launch Template name prefix: %q", v, errors)
		}
	}
}

func TestValidateLaunchTemplateId(t *testing.T) {
	validIds := []string{
		"lt-foobar123456",
	}
	for _, v := range validIds {
		_, errors := validateLaunchTemplateId(v, "id")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid Launch Template id: %q", v, errors)
		}
	}

	invalidIds := []string{
		strings.Repeat("W", 256),
		"invalid-foobar123456",
		"lt_foobar123456",
	}
	for _, v := range invalidIds {
		_, errors := validateLaunchTemplateId(v, "id")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid Launch Template id: %q", v, errors)
		}
	}
}

func TestValidateNeptuneParamGroupName(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "tEsting123",
			ErrCount: 1,
		},
		{
			Value:    "testing123!",
			ErrCount: 1,
		},
		{
			Value:    "1testing123",
			ErrCount: 1,
		},
		{
			Value:    "testing--123",
			ErrCount: 1,
		},
		{
			Value:    "testing_123",
			ErrCount: 1,
		},
		{
			Value:    "testing123-",
			ErrCount: 1,
		},
		{
			Value:    acctest.RandStringFromCharSet(256, acctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validateNeptuneParamGroupName(tc.Value, "aws_neptune_cluster_parameter_group_name")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the Neptune Parameter Group Name to trigger a validation error for %q", tc.Value)
		}
	}
}

func TestValidateNeptuneParamGroupNamePrefix(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "tEsting123",
			ErrCount: 1,
		},
		{
			Value:    "testing123!",
			ErrCount: 1,
		},
		{
			Value:    "1testing123",
			ErrCount: 1,
		},
		{
			Value:    "testing--123",
			ErrCount: 1,
		},
		{
			Value:    "testing_123",
			ErrCount: 1,
		},
		{
			Value:    acctest.RandStringFromCharSet(256, acctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validateNeptuneParamGroupNamePrefix(tc.Value, "aws_neptune_cluster_parameter_group_name")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the Neptune Parameter Group Name to trigger a validation error for %q", tc.Value)
		}
	}
}

func TestValidateCloudFrontPublicKeyName(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "testing123!",
			ErrCount: 1,
		},
		{
			Value:    "testing 123",
			ErrCount: 1,
		},
		{
			Value:    acctest.RandStringFromCharSet(129, acctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validateCloudFrontPublicKeyName(tc.Value, "aws_cloudfront_public_key")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the CloudFront PublicKey Name to trigger a validation error for %q", tc.Value)
		}
	}
}

func TestValidateCloudFrontPublicKeyNamePrefix(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "testing123!",
			ErrCount: 1,
		},
		{
			Value:    "testing 123",
			ErrCount: 1,
		},
		{
			Value:    acctest.RandStringFromCharSet(128, acctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validateCloudFrontPublicKeyNamePrefix(tc.Value, "aws_cloudfront_public_key")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the CloudFront PublicKey Name to trigger a validation error for %q", tc.Value)
		}
	}
}

func TestValidateDxConnectionBandWidth(t *testing.T) {
	validBandwidths := []string{
		"1Gbps",
		"2Gbps",
		"5Gbps",
		"10Gbps",
		"50Mbps",
		"100Mbps",
		"200Mbps",
		"300Mbps",
		"400Mbps",
		"500Mbps",
	}
	for _, v := range validBandwidths {
		_, errors := validateDxConnectionBandWidth()(v, "bandwidth")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid bandwidth: %q", v, errors)
		}
	}

	invalidBandwidths := []string{
		"1Tbps",
		"100Gbps",
		"10GBpS",
		"42Mbps",
		"0",
		"???",
		"a lot",
	}
	for _, v := range invalidBandwidths {
		_, errors := validateDxConnectionBandWidth()(v, "bandwidth")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid bandwidth", v)
		}
	}
}

func TestValidateLbTargetGroupName(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "tf.test.elb.target.1",
			ErrCount: 1,
		},
		{
			Value:    "-tf-test-target",
			ErrCount: 1,
		},
		{
			Value:    "tf-test-target-",
			ErrCount: 1,
		},
		{
			Value:    acctest.RandStringFromCharSet(33, acctest.CharSetAlpha),
			ErrCount: 1,
		},
	}
	for _, tc := range cases {
		_, errors := validateLbTargetGroupName(tc.Value, "aws_lb_target_group")
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the AWS LB Target Group Name to trigger a validation error for %q", tc.Value)
		}
	}
}

func TestValidateLbTargetGroupNamePrefix(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "tf.lb",
			ErrCount: 1,
		},
		{
			Value:    "-tf-lb",
			ErrCount: 1,
		},
		{
			Value:    acctest.RandStringFromCharSet(32, acctest.CharSetAlpha),
			ErrCount: 1,
		},
	}
	for _, tc := range cases {
		_, errors := validateLbTargetGroupNamePrefix(tc.Value, "aws_lb_target_group")
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the AWS LB Target Group Name to trigger a validation error for %q", tc.Value)
		}
	}
}

func TestValidateSecretManagerSecretName(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "testing123!",
			ErrCount: 1,
		},
		{
			Value:    "testing 123",
			ErrCount: 1,
		},
		{
			Value:    acctest.RandStringFromCharSet(513, acctest.CharSetAlpha),
			ErrCount: 1,
		},
	}
	for _, tc := range cases {
		_, errors := validateSecretManagerSecretName(tc.Value, "aws_secretsmanager_secret")
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the AWS Secretsmanager Secret Name to not trigger a validation error for %q", tc.Value)
		}
	}
}

func TestValidateSecretManagerSecretNamePrefix(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "testing123!",
			ErrCount: 1,
		},
		{
			Value:    "testing 123",
			ErrCount: 1,
		},
		{
			Value:    acctest.RandStringFromCharSet(512, acctest.CharSetAlpha),
			ErrCount: 1,
		},
	}
	for _, tc := range cases {
		_, errors := validateSecretManagerSecretNamePrefix(tc.Value, "aws_secretsmanager_secret")
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the AWS Secretsmanager Secret Name to not trigger a validation error for %q", tc.Value)
		}
	}
}

func TestValidateRoute53ResolverName(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "testing123!",
			ErrCount: 1,
		},
		{
			Value:    "testing - 123__",
			ErrCount: 0,
		},
		{
			Value:    acctest.RandStringFromCharSet(65, acctest.CharSetAlpha),
			ErrCount: 1,
		},
		{
			Value:    "1",
			ErrCount: 1,
		},
		{
			Value:    "10",
			ErrCount: 0,
		},
		{
			Value:    "A",
			ErrCount: 0,
		},
	}
	for _, tc := range cases {
		_, errors := validateRoute53ResolverName(tc.Value, "aws_route53_resolver_endpoint")
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the AWS Route 53 Resolver Endpoint Name to not trigger a validation error for %q", tc.Value)
		}
	}
}

func TestCloudWatchEventCustomEventBusName(t *testing.T) {
	cases := []struct {
		Value   string
		IsValid bool
	}{
		{
			Value:   "",
			IsValid: false,
		},
		{
			Value:   "default",
			IsValid: false,
		},
		{
			Value:   acctest.RandStringFromCharSet(256, acctest.CharSetAlpha),
			IsValid: true,
		},
		{
			Value:   acctest.RandStringFromCharSet(257, acctest.CharSetAlpha),
			IsValid: false,
		},
		{
			Value:   "aws.partner/test/test",
			IsValid: false,
		},
		{
			Value:   "/test0._1-",
			IsValid: false,
		},
		{
			Value:   "test0._1-",
			IsValid: true,
		},
	}
	for _, tc := range cases {
		_, errors := validateCloudWatchEventCustomEventBusName(tc.Value, "aws_cloudwatch_event_bus")
		isValid := len(errors) == 0
		if tc.IsValid && !isValid {
			t.Errorf("expected %q to return valid, but did not", tc.Value)
		} else if !tc.IsValid && isValid {
			t.Errorf("expected %q to not return valid, but did", tc.Value)
		}
	}
}

func TestValidateServiceDiscoveryNamespaceName(t *testing.T) {
	validNames := []string{
		"ValidName",
		"V_-.dN01e",
		"0",
		".",
		"-",
		"_",
		strings.Repeat("x", 1024),
	}
	for _, v := range validNames {
		_, errors := validateServiceDiscoveryNamespaceName(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid namespace name: %q", v, errors)
		}
	}

	invalidNames := []string{
		"Inval:dName",
		"Invalid Name",
		"*",
		"",
		// length > 512
		strings.Repeat("x", 1025),
	}
	for _, v := range invalidNames {
		_, errors := validateServiceDiscoveryNamespaceName(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid namespace name", v)
		}
	}
}

func TestValidateUTCTimestamp(t *testing.T) {
	validT := []string{
		"2006-01-02T15:04:05Z",
	}

	invalidT := []string{
		"2015-03-07 23:45:00",
		"27-03-2019 23:45:00",
		"Mon, 02 Jan 2006 15:04:05 -0700",
	}

	for _, f := range validT {
		_, errors := validateUTCTimestamp(f, "utc_timestamp")
		if len(errors) > 0 {
			t.Fatalf("expected the time %q to be in valid format, got error %q", f, errors)
		}
	}

	for _, f := range invalidT {
		_, errors := validateUTCTimestamp(f, "utc_timestamp")
		if len(errors) == 0 {
			t.Fatalf("expected the time %q to fail validation", f)
		}
	}
}
