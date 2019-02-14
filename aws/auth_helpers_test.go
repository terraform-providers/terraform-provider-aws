package aws

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
)

func TestGetAccountIDAndPartition(t *testing.T) {
	var testCases = []struct {
		Description          string
		AuthProviderName     string
		EC2MetadataEndpoints []*endpoint
		IAMEndpoints         []*awsMockEndpoint
		STSEndpoints         []*awsMockEndpoint
		ErrCount             int
		ExpectedAccountID    string
		ExpectedPartition    string
	}{
		{
			Description:          "EC2 Metadata over iam:GetUser when using EC2 Instance Profile",
			AuthProviderName:     ec2rolecreds.ProviderName,
			EC2MetadataEndpoints: append(ec2metadata_securityCredentialsEndpoints, ec2metadata_instanceIdEndpoint, ec2metadata_iamInfoEndpoint),
			IAMEndpoints: []*awsMockEndpoint{
				{
					Request:  &awsMockRequest{"POST", "/", "Action=GetUser&Version=2010-05-08"},
					Response: &awsMockResponse{200, iamResponse_GetUser_valid, "text/xml"},
				},
			},
			ExpectedAccountID: ec2metadata_iamInfoEndpoint_expectedAccountID,
			ExpectedPartition: ec2metadata_iamInfoEndpoint_expectedPartition,
		},
		{
			Description:          "Mimic the metadata service mocked by Hologram (https://github.com/AdRoll/hologram)",
			AuthProviderName:     ec2rolecreds.ProviderName,
			EC2MetadataEndpoints: ec2metadata_securityCredentialsEndpoints,
			IAMEndpoints: []*awsMockEndpoint{
				{
					Request:  &awsMockRequest{"POST", "/", "Action=GetUser&Version=2010-05-08"},
					Response: &awsMockResponse{403, iamResponse_GetUser_unauthorized, "text/xml"},
				},
			},
			STSEndpoints: []*awsMockEndpoint{
				{
					Request:  &awsMockRequest{"POST", "/", "Action=GetCallerIdentity&Version=2011-06-15"},
					Response: &awsMockResponse{200, stsResponse_GetCallerIdentity_valid, "text/xml"},
				},
			},
			ExpectedAccountID: stsResponse_GetCallerIdentity_valid_expectedAccountID,
			ExpectedPartition: stsResponse_GetCallerIdentity_valid_expectedPartition,
		},
		{
			Description: "iam:ListRoles if iam:GetUser AccessDenied and sts:GetCallerIdentity fails",
			IAMEndpoints: []*awsMockEndpoint{
				{
					Request:  &awsMockRequest{"POST", "/", "Action=GetUser&Version=2010-05-08"},
					Response: &awsMockResponse{403, iamResponse_GetUser_unauthorized, "text/xml"},
				},
				{
					Request:  &awsMockRequest{"POST", "/", "Action=ListRoles&MaxItems=1&Version=2010-05-08"},
					Response: &awsMockResponse{200, iamResponse_ListRoles_valid, "text/xml"},
				},
			},
			STSEndpoints: []*awsMockEndpoint{
				{
					Request:  &awsMockRequest{"POST", "/", "Action=GetCallerIdentity&Version=2011-06-15"},
					Response: &awsMockResponse{403, stsResponse_GetCallerIdentity_unauthorized, "text/xml"},
				},
			},
			ExpectedAccountID: iamResponse_ListRoles_valid_expectedAccountID,
			ExpectedPartition: iamResponse_ListRoles_valid_expectedPartition,
		},
		{
			Description: "iam:ListRoles if iam:GetUser ValidationError and sts:GetCallerIdentity fails",
			IAMEndpoints: []*awsMockEndpoint{
				{
					Request:  &awsMockRequest{"POST", "/", "Action=GetUser&Version=2010-05-08"},
					Response: &awsMockResponse{400, iamResponse_GetUser_federatedFailure, "text/xml"},
				},
				{
					Request:  &awsMockRequest{"POST", "/", "Action=ListRoles&MaxItems=1&Version=2010-05-08"},
					Response: &awsMockResponse{200, iamResponse_ListRoles_valid, "text/xml"},
				},
			},
			STSEndpoints: []*awsMockEndpoint{
				{
					Request:  &awsMockRequest{"POST", "/", "Action=GetCallerIdentity&Version=2011-06-15"},
					Response: &awsMockResponse{403, stsResponse_GetCallerIdentity_unauthorized, "text/xml"},
				},
			},
			ExpectedAccountID: iamResponse_ListRoles_valid_expectedAccountID,
			ExpectedPartition: iamResponse_ListRoles_valid_expectedPartition,
		},
		{
			Description: "Error when all endpoints fail",
			IAMEndpoints: []*awsMockEndpoint{
				{
					Request:  &awsMockRequest{"POST", "/", "Action=GetUser&Version=2010-05-08"},
					Response: &awsMockResponse{400, iamResponse_GetUser_federatedFailure, "text/xml"},
				},
				{
					Request:  &awsMockRequest{"POST", "/", "Action=ListRoles&MaxItems=1&Version=2010-05-08"},
					Response: &awsMockResponse{403, iamResponse_ListRoles_unauthorized, "text/xml"},
				},
			},
			STSEndpoints: []*awsMockEndpoint{
				{
					Request:  &awsMockRequest{"POST", "/", "Action=GetCallerIdentity&Version=2011-06-15"},
					Response: &awsMockResponse{403, stsResponse_GetCallerIdentity_unauthorized, "text/xml"},
				},
			},
			ErrCount: 1,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Description, func(t *testing.T) {
			resetEnv := unsetEnv(t)
			defer resetEnv()
			// capture the test server's close method, to call after the test returns
			awsTs := awsMetadataApiMock(testCase.EC2MetadataEndpoints)
			defer awsTs()

			closeIam, iamSess, err := getMockedAwsApiSession("IAM", testCase.IAMEndpoints)
			defer closeIam()
			if err != nil {
				t.Fatal(err)
			}

			closeSts, stsSess, err := getMockedAwsApiSession("STS", testCase.STSEndpoints)
			defer closeSts()
			if err != nil {
				t.Fatal(err)
			}

			iamConn := iam.New(iamSess)
			stsConn := sts.New(stsSess)

			accountID, partition, err := GetAccountIDAndPartition(iamConn, stsConn, testCase.AuthProviderName)
			if err != nil && testCase.ErrCount == 0 {
				t.Fatalf("Expected no error, received error: %s", err)
			}
			if err == nil && testCase.ErrCount > 0 {
				t.Fatalf("Expected %d error(s), received none", testCase.ErrCount)
			}
			if accountID != testCase.ExpectedAccountID {
				t.Fatalf("Parsed account ID doesn't match with expected (%q != %q)", accountID, testCase.ExpectedAccountID)
			}
			if partition != testCase.ExpectedPartition {
				t.Fatalf("Parsed partition doesn't match with expected (%q != %q)", partition, testCase.ExpectedPartition)
			}
		})
	}
}

func TestGetAccountIDAndPartitionFromEC2Metadata(t *testing.T) {
	t.Run("EC2 metadata success", func(t *testing.T) {
		resetEnv := unsetEnv(t)
		defer resetEnv()
		// capture the test server's close method, to call after the test returns
		awsTs := awsMetadataApiMock(append(ec2metadata_securityCredentialsEndpoints, ec2metadata_instanceIdEndpoint, ec2metadata_iamInfoEndpoint))
		defer awsTs()

		id, partition, err := GetAccountIDAndPartitionFromEC2Metadata()
		if err != nil {
			t.Fatalf("Getting account ID from EC2 metadata API failed: %s", err)
		}

		if id != ec2metadata_iamInfoEndpoint_expectedAccountID {
			t.Fatalf("Expected account ID: %s, given: %s", ec2metadata_iamInfoEndpoint_expectedAccountID, id)
		}
		if partition != ec2metadata_iamInfoEndpoint_expectedPartition {
			t.Fatalf("Expected partition: %s, given: %s", ec2metadata_iamInfoEndpoint_expectedPartition, partition)
		}
	})
}

func TestGetAccountIDAndPartitionFromIAMGetUser(t *testing.T) {
	var testCases = []struct {
		Description       string
		MockEndpoints     []*awsMockEndpoint
		ErrCount          int
		ExpectedAccountID string
		ExpectedPartition string
	}{
		{
			Description: "Ignore iam:GetUser failure with federated user",
			MockEndpoints: []*awsMockEndpoint{
				{
					Request:  &awsMockRequest{"POST", "/", "Action=GetUser&Version=2010-05-08"},
					Response: &awsMockResponse{400, iamResponse_GetUser_federatedFailure, "text/xml"},
				},
			},
			ErrCount: 0,
		},
		{
			Description: "Ignore iam:GetUser failure with unauthorized user",
			MockEndpoints: []*awsMockEndpoint{
				{
					Request:  &awsMockRequest{"POST", "/", "Action=GetUser&Version=2010-05-08"},
					Response: &awsMockResponse{403, iamResponse_GetUser_unauthorized, "text/xml"},
				},
			},
			ErrCount: 0,
		},
		{
			Description: "iam:GetUser success",
			MockEndpoints: []*awsMockEndpoint{
				{
					Request:  &awsMockRequest{"POST", "/", "Action=GetUser&Version=2010-05-08"},
					Response: &awsMockResponse{200, iamResponse_GetUser_valid, "text/xml"},
				},
			},
			ExpectedAccountID: iamResponse_GetUser_valid_expectedAccountID,
			ExpectedPartition: iamResponse_GetUser_valid_expectedPartition,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Description, func(t *testing.T) {
			closeIam, iamSess, err := getMockedAwsApiSession("IAM", testCase.MockEndpoints)
			defer closeIam()
			if err != nil {
				t.Fatal(err)
			}

			iamConn := iam.New(iamSess)

			accountID, partition, err := GetAccountIDAndPartitionFromIAMGetUser(iamConn)
			if err != nil && testCase.ErrCount == 0 {
				t.Fatalf("Expected no error, received error: %s", err)
			}
			if err == nil && testCase.ErrCount > 0 {
				t.Fatalf("Expected %d error(s), received none", testCase.ErrCount)
			}
			if accountID != testCase.ExpectedAccountID {
				t.Fatalf("Parsed account ID doesn't match with expected (%q != %q)", accountID, testCase.ExpectedAccountID)
			}
			if partition != testCase.ExpectedPartition {
				t.Fatalf("Parsed partition doesn't match with expected (%q != %q)", partition, testCase.ExpectedPartition)
			}
		})
	}
}

func TestGetAccountIDAndPartitionFromIAMListRoles(t *testing.T) {
	var testCases = []struct {
		Description       string
		MockEndpoints     []*awsMockEndpoint
		ErrCount          int
		ExpectedAccountID string
		ExpectedPartition string
	}{
		{
			Description: "iam:ListRoles unauthorized",
			MockEndpoints: []*awsMockEndpoint{
				{
					Request:  &awsMockRequest{"POST", "/", "Action=ListRoles&MaxItems=1&Version=2010-05-08"},
					Response: &awsMockResponse{403, iamResponse_ListRoles_unauthorized, "text/xml"},
				},
			},
			ErrCount: 1,
		},
		{
			Description: "iam:ListRoles success",
			MockEndpoints: []*awsMockEndpoint{
				{
					Request:  &awsMockRequest{"POST", "/", "Action=ListRoles&MaxItems=1&Version=2010-05-08"},
					Response: &awsMockResponse{200, iamResponse_ListRoles_valid, "text/xml"},
				},
			},
			ExpectedAccountID: iamResponse_ListRoles_valid_expectedAccountID,
			ExpectedPartition: iamResponse_ListRoles_valid_expectedPartition,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Description, func(t *testing.T) {
			closeIam, iamSess, err := getMockedAwsApiSession("IAM", testCase.MockEndpoints)
			defer closeIam()
			if err != nil {
				t.Fatal(err)
			}

			iamConn := iam.New(iamSess)

			accountID, partition, err := GetAccountIDAndPartitionFromIAMListRoles(iamConn)
			if err != nil && testCase.ErrCount == 0 {
				t.Fatalf("Expected no error, received error: %s", err)
			}
			if err == nil && testCase.ErrCount > 0 {
				t.Fatalf("Expected %d error(s), received none", testCase.ErrCount)
			}
			if accountID != testCase.ExpectedAccountID {
				t.Fatalf("Parsed account ID doesn't match with expected (%q != %q)", accountID, testCase.ExpectedAccountID)
			}
			if partition != testCase.ExpectedPartition {
				t.Fatalf("Parsed partition doesn't match with expected (%q != %q)", partition, testCase.ExpectedPartition)
			}
		})
	}
}

func TestGetAccountIDAndPartitionFromSTSGetCallerIdentity(t *testing.T) {
	var testCases = []struct {
		Description       string
		MockEndpoints     []*awsMockEndpoint
		ErrCount          int
		ExpectedAccountID string
		ExpectedPartition string
	}{
		{
			Description: "sts:GetCallerIdentity unauthorized",
			MockEndpoints: []*awsMockEndpoint{
				{
					Request:  &awsMockRequest{"POST", "/", "Action=GetCallerIdentity&Version=2011-06-15"},
					Response: &awsMockResponse{403, stsResponse_GetCallerIdentity_unauthorized, "text/xml"},
				},
			},
			ErrCount: 1,
		},
		{
			Description: "sts:GetCallerIdentity success",
			MockEndpoints: []*awsMockEndpoint{
				{
					Request:  &awsMockRequest{"POST", "/", "Action=GetCallerIdentity&Version=2011-06-15"},
					Response: &awsMockResponse{200, stsResponse_GetCallerIdentity_valid, "text/xml"},
				},
			},
			ExpectedAccountID: stsResponse_GetCallerIdentity_valid_expectedAccountID,
			ExpectedPartition: stsResponse_GetCallerIdentity_valid_expectedPartition,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Description, func(t *testing.T) {
			closeSts, stsSess, err := getMockedAwsApiSession("STS", testCase.MockEndpoints)
			defer closeSts()
			if err != nil {
				t.Fatal(err)
			}

			stsConn := sts.New(stsSess)

			accountID, partition, err := GetAccountIDAndPartitionFromSTSGetCallerIdentity(stsConn)
			if err != nil && testCase.ErrCount == 0 {
				t.Fatalf("Expected no error, received error: %s", err)
			}
			if err == nil && testCase.ErrCount > 0 {
				t.Fatalf("Expected %d error(s), received none", testCase.ErrCount)
			}
			if accountID != testCase.ExpectedAccountID {
				t.Fatalf("Parsed account ID doesn't match with expected (%q != %q)", accountID, testCase.ExpectedAccountID)
			}
			if partition != testCase.ExpectedPartition {
				t.Fatalf("Parsed partition doesn't match with expected (%q != %q)", partition, testCase.ExpectedPartition)
			}
		})
	}
}

func TestAWSParseAccountIDAndPartitionFromARN(t *testing.T) {
	var testCases = []struct {
		InputARN          string
		ErrCount          int
		ExpectedAccountID string
		ExpectedPartition string
	}{
		{
			InputARN: "invalid-arn",
			ErrCount: 1,
		},
		{
			InputARN:          "arn:aws:iam::123456789012:instance-profile/name",
			ExpectedAccountID: "123456789012",
			ExpectedPartition: "aws",
		},
		{
			InputARN:          "arn:aws:iam::123456789012:user/name",
			ExpectedAccountID: "123456789012",
			ExpectedPartition: "aws",
		},
		{
			InputARN:          "arn:aws:sts::123456789012:assumed-role/name",
			ExpectedAccountID: "123456789012",
			ExpectedPartition: "aws",
		},
		{
			InputARN:          "arn:aws-us-gov:sts::123456789012:assumed-role/name",
			ExpectedAccountID: "123456789012",
			ExpectedPartition: "aws-us-gov",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.InputARN, func(t *testing.T) {
			accountID, partition, err := parseAccountIDAndPartitionFromARN(testCase.InputARN)
			if err != nil && testCase.ErrCount == 0 {
				t.Fatalf("Expected no error when parsing ARN, received error: %s", err)
			}
			if err == nil && testCase.ErrCount > 0 {
				t.Fatalf("Expected %d error(s) when parsing ARN, received none", testCase.ErrCount)
			}
			if accountID != testCase.ExpectedAccountID {
				t.Fatalf("Parsed account ID doesn't match with expected (%q != %q)", accountID, testCase.ExpectedAccountID)
			}
			if partition != testCase.ExpectedPartition {
				t.Fatalf("Parsed partition doesn't match with expected (%q != %q)", partition, testCase.ExpectedPartition)
			}
		})
	}
}

func TestAWSGetCredentials_shouldError(t *testing.T) {
	resetEnv := unsetEnv(t)
	defer resetEnv()
	cfg := Config{}

	c, err := GetCredentials(&cfg)
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() != "NoCredentialProviders" {
			t.Fatal("Expected NoCredentialProviders error")
		}
	}
	_, err = c.Get()
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() != "NoCredentialProviders" {
			t.Fatal("Expected NoCredentialProviders error")
		}
	}
	if err == nil {
		t.Fatal("Expected an error with empty env, keys, and IAM in AWS Config")
	}
}

func TestAWSGetCredentials_shouldBeStatic(t *testing.T) {
	simple := []struct {
		Key, Secret, Token string
	}{
		{
			Key:    "test",
			Secret: "secret",
		}, {
			Key:    "test",
			Secret: "test",
			Token:  "test",
		},
	}

	for _, c := range simple {
		cfg := Config{
			AccessKey: c.Key,
			SecretKey: c.Secret,
			Token:     c.Token,
		}

		creds, err := GetCredentials(&cfg)
		if err != nil {
			t.Fatalf("Error gettings creds: %s", err)
		}
		if creds == nil {
			t.Fatal("Expected a static creds provider to be returned")
		}

		v, err := creds.Get()
		if err != nil {
			t.Fatalf("Error gettings creds: %s", err)
		}

		if v.AccessKeyID != c.Key {
			t.Fatalf("AccessKeyID mismatch, expected: (%s), got (%s)", c.Key, v.AccessKeyID)
		}
		if v.SecretAccessKey != c.Secret {
			t.Fatalf("SecretAccessKey mismatch, expected: (%s), got (%s)", c.Secret, v.SecretAccessKey)
		}
		if v.SessionToken != c.Token {
			t.Fatalf("SessionToken mismatch, expected: (%s), got (%s)", c.Token, v.SessionToken)
		}
	}
}

// TestAWSGetCredentials_shouldIAM is designed to test the scenario of running Terraform
// from an EC2 instance, without environment variables or manually supplied
// credentials.
func TestAWSGetCredentials_shouldIAM(t *testing.T) {
	// clear AWS_* environment variables
	resetEnv := unsetEnv(t)
	defer resetEnv()

	// capture the test server's close method, to call after the test returns
	ts := awsMetadataApiMock(append(ec2metadata_securityCredentialsEndpoints, ec2metadata_instanceIdEndpoint, ec2metadata_iamInfoEndpoint))
	defer ts()

	// An empty config, no key supplied
	cfg := Config{}

	creds, err := GetCredentials(&cfg)
	if err != nil {
		t.Fatalf("Error gettings creds: %s", err)
	}
	if creds == nil {
		t.Fatal("Expected a static creds provider to be returned")
	}

	v, err := creds.Get()
	if err != nil {
		t.Fatalf("Error gettings creds: %s", err)
	}
	if v.AccessKeyID != "somekey" {
		t.Fatalf("AccessKeyID mismatch, expected: (somekey), got (%s)", v.AccessKeyID)
	}
	if v.SecretAccessKey != "somesecret" {
		t.Fatalf("SecretAccessKey mismatch, expected: (somesecret), got (%s)", v.SecretAccessKey)
	}
	if v.SessionToken != "sometoken" {
		t.Fatalf("SessionToken mismatch, expected: (sometoken), got (%s)", v.SessionToken)
	}
}

// TestAWSGetCredentials_shouldIAM is designed to test the scenario of running Terraform
// from an EC2 instance, without environment variables or manually supplied
// credentials.
func TestAWSGetCredentials_shouldIgnoreIAM(t *testing.T) {
	resetEnv := unsetEnv(t)
	defer resetEnv()
	// capture the test server's close method, to call after the test returns
	ts := awsMetadataApiMock(append(ec2metadata_securityCredentialsEndpoints, ec2metadata_instanceIdEndpoint, ec2metadata_iamInfoEndpoint))
	defer ts()
	simple := []struct {
		Key, Secret, Token string
	}{
		{
			Key:    "test",
			Secret: "secret",
		}, {
			Key:    "test",
			Secret: "test",
			Token:  "test",
		},
	}

	for _, c := range simple {
		cfg := Config{
			AccessKey: c.Key,
			SecretKey: c.Secret,
			Token:     c.Token,
		}

		creds, err := GetCredentials(&cfg)
		if err != nil {
			t.Fatalf("Error gettings creds: %s", err)
		}
		if creds == nil {
			t.Fatal("Expected a static creds provider to be returned")
		}

		v, err := creds.Get()
		if err != nil {
			t.Fatalf("Error gettings creds: %s", err)
		}
		if v.AccessKeyID != c.Key {
			t.Fatalf("AccessKeyID mismatch, expected: (%s), got (%s)", c.Key, v.AccessKeyID)
		}
		if v.SecretAccessKey != c.Secret {
			t.Fatalf("SecretAccessKey mismatch, expected: (%s), got (%s)", c.Secret, v.SecretAccessKey)
		}
		if v.SessionToken != c.Token {
			t.Fatalf("SessionToken mismatch, expected: (%s), got (%s)", c.Token, v.SessionToken)
		}
	}
}

func TestAWSGetCredentials_shouldErrorWithInvalidEndpoint(t *testing.T) {
	resetEnv := unsetEnv(t)
	defer resetEnv()
	// capture the test server's close method, to call after the test returns
	ts := invalidAwsEnv(t)
	defer ts()

	creds, err := GetCredentials(&Config{})
	if err != nil {
		t.Fatalf("Error gettings creds: %s", err)
	}
	if creds == nil {
		t.Fatal("Expected a static creds provider to be returned")
	}

	v, err := creds.Get()
	if err == nil {
		t.Fatal("Expected error returned when getting creds w/ invalid EC2 endpoint")
	}

	if v.ProviderName != "" {
		t.Fatalf("Expected provider name to be empty, %q given", v.ProviderName)
	}
}

func TestAWSGetCredentials_shouldIgnoreInvalidEndpoint(t *testing.T) {
	resetEnv := unsetEnv(t)
	defer resetEnv()
	// capture the test server's close method, to call after the test returns
	ts := invalidAwsEnv(t)
	defer ts()

	creds, err := GetCredentials(&Config{AccessKey: "accessKey", SecretKey: "secretKey"})
	if err != nil {
		t.Fatalf("Error gettings creds: %s", err)
	}
	v, err := creds.Get()
	if err != nil {
		t.Fatalf("Getting static credentials w/ invalid EC2 endpoint failed: %s", err)
	}
	if creds == nil {
		t.Fatal("Expected a static creds provider to be returned")
	}

	if v.ProviderName != "StaticProvider" {
		t.Fatalf("Expected provider name to be %q, %q given", "StaticProvider", v.ProviderName)
	}

	if v.AccessKeyID != "accessKey" {
		t.Fatalf("Static Access Key %q doesn't match: %s", "accessKey", v.AccessKeyID)
	}

	if v.SecretAccessKey != "secretKey" {
		t.Fatalf("Static Secret Key %q doesn't match: %s", "secretKey", v.SecretAccessKey)
	}
}

func TestAWSGetCredentials_shouldCatchEC2RoleProvider(t *testing.T) {
	resetEnv := unsetEnv(t)
	defer resetEnv()
	// capture the test server's close method, to call after the test returns
	ts := awsMetadataApiMock(append(ec2metadata_securityCredentialsEndpoints, ec2metadata_instanceIdEndpoint, ec2metadata_iamInfoEndpoint))
	defer ts()

	creds, err := GetCredentials(&Config{})
	if err != nil {
		t.Fatalf("Error gettings creds: %s", err)
	}
	if creds == nil {
		t.Fatal("Expected an EC2Role creds provider to be returned")
	}

	v, err := creds.Get()
	if err != nil {
		t.Fatalf("Expected no error when getting creds: %s", err)
	}
	expectedProvider := "EC2RoleProvider"
	if v.ProviderName != expectedProvider {
		t.Fatalf("Expected provider name to be %q, %q given",
			expectedProvider, v.ProviderName)
	}
}

var credentialsFileContents = `[myprofile]
aws_access_key_id = accesskey
aws_secret_access_key = secretkey
`

func TestAWSGetCredentials_shouldBeShared(t *testing.T) {
	file, err := ioutil.TempFile(os.TempDir(), "terraform_aws_cred")
	if err != nil {
		t.Fatalf("Error writing temporary credentials file: %s", err)
	}
	_, err = file.WriteString(credentialsFileContents)
	if err != nil {
		t.Fatalf("Error writing temporary credentials to file: %s", err)
	}
	err = file.Close()
	if err != nil {
		t.Fatalf("Error closing temporary credentials file: %s", err)
	}

	defer os.Remove(file.Name())

	resetEnv := unsetEnv(t)
	defer resetEnv()

	if err := os.Setenv("AWS_PROFILE", "myprofile"); err != nil {
		t.Fatalf("Error resetting env var AWS_PROFILE: %s", err)
	}
	if err := os.Setenv("AWS_SHARED_CREDENTIALS_FILE", file.Name()); err != nil {
		t.Fatalf("Error resetting env var AWS_SHARED_CREDENTIALS_FILE: %s", err)
	}

	creds, err := GetCredentials(&Config{Profile: "myprofile", CredsFilename: file.Name()})
	if err != nil {
		t.Fatalf("Error gettings creds: %s", err)
	}
	if creds == nil {
		t.Fatal("Expected a provider chain to be returned")
	}

	v, err := creds.Get()
	if err != nil {
		t.Fatalf("Error gettings creds: %s", err)
	}

	if v.AccessKeyID != "accesskey" {
		t.Fatalf("AccessKeyID mismatch, expected (%s), got (%s)", "accesskey", v.AccessKeyID)
	}

	if v.SecretAccessKey != "secretkey" {
		t.Fatalf("SecretAccessKey mismatch, expected (%s), got (%s)", "accesskey", v.AccessKeyID)
	}
}

func TestAWSGetCredentials_shouldBeENV(t *testing.T) {
	// need to set the environment variables to a dummy string, as we don't know
	// what they may be at runtime without hardcoding here
	s := "some_env"
	resetEnv := setEnv(s, t)

	defer resetEnv()

	cfg := Config{}
	creds, err := GetCredentials(&cfg)
	if err != nil {
		t.Fatalf("Error gettings creds: %s", err)
	}
	if creds == nil {
		t.Fatalf("Expected a static creds provider to be returned")
	}

	v, err := creds.Get()
	if err != nil {
		t.Fatalf("Error gettings creds: %s", err)
	}
	if v.AccessKeyID != s {
		t.Fatalf("AccessKeyID mismatch, expected: (%s), got (%s)", s, v.AccessKeyID)
	}
	if v.SecretAccessKey != s {
		t.Fatalf("SecretAccessKey mismatch, expected: (%s), got (%s)", s, v.SecretAccessKey)
	}
	if v.SessionToken != s {
		t.Fatalf("SessionToken mismatch, expected: (%s), got (%s)", s, v.SessionToken)
	}
}

// unsetEnv unsets environment variables for testing a "clean slate" with no
// credentials in the environment
func unsetEnv(t *testing.T) func() {
	// Grab any existing AWS keys and preserve. In some tests we'll unset these, so
	// we need to have them and restore them after
	e := getEnv()
	if err := os.Unsetenv("AWS_ACCESS_KEY_ID"); err != nil {
		t.Fatalf("Error unsetting env var AWS_ACCESS_KEY_ID: %s", err)
	}
	if err := os.Unsetenv("AWS_SECRET_ACCESS_KEY"); err != nil {
		t.Fatalf("Error unsetting env var AWS_SECRET_ACCESS_KEY: %s", err)
	}
	if err := os.Unsetenv("AWS_SESSION_TOKEN"); err != nil {
		t.Fatalf("Error unsetting env var AWS_SESSION_TOKEN: %s", err)
	}
	if err := os.Unsetenv("AWS_PROFILE"); err != nil {
		t.Fatalf("Error unsetting env var AWS_PROFILE: %s", err)
	}
	if err := os.Unsetenv("AWS_SHARED_CREDENTIALS_FILE"); err != nil {
		t.Fatalf("Error unsetting env var AWS_SHARED_CREDENTIALS_FILE: %s", err)
	}

	return func() {
		// re-set all the envs we unset above
		if err := os.Setenv("AWS_ACCESS_KEY_ID", e.Key); err != nil {
			t.Fatalf("Error resetting env var AWS_ACCESS_KEY_ID: %s", err)
		}
		if err := os.Setenv("AWS_SECRET_ACCESS_KEY", e.Secret); err != nil {
			t.Fatalf("Error resetting env var AWS_SECRET_ACCESS_KEY: %s", err)
		}
		if err := os.Setenv("AWS_SESSION_TOKEN", e.Token); err != nil {
			t.Fatalf("Error resetting env var AWS_SESSION_TOKEN: %s", err)
		}
		if err := os.Setenv("AWS_PROFILE", e.Profile); err != nil {
			t.Fatalf("Error resetting env var AWS_PROFILE: %s", err)
		}
		if err := os.Setenv("AWS_SHARED_CREDENTIALS_FILE", e.CredsFilename); err != nil {
			t.Fatalf("Error resetting env var AWS_SHARED_CREDENTIALS_FILE: %s", err)
		}
	}
}

func setEnv(s string, t *testing.T) func() {
	e := getEnv()
	// Set all the envs to a dummy value
	if err := os.Setenv("AWS_ACCESS_KEY_ID", s); err != nil {
		t.Fatalf("Error setting env var AWS_ACCESS_KEY_ID: %s", err)
	}
	if err := os.Setenv("AWS_SECRET_ACCESS_KEY", s); err != nil {
		t.Fatalf("Error setting env var AWS_SECRET_ACCESS_KEY: %s", err)
	}
	if err := os.Setenv("AWS_SESSION_TOKEN", s); err != nil {
		t.Fatalf("Error setting env var AWS_SESSION_TOKEN: %s", err)
	}
	if err := os.Setenv("AWS_PROFILE", s); err != nil {
		t.Fatalf("Error setting env var AWS_PROFILE: %s", err)
	}
	if err := os.Setenv("AWS_SHARED_CREDENTIALS_FILE", s); err != nil {
		t.Fatalf("Error setting env var AWS_SHARED_CREDENTIALS_FLE: %s", err)
	}

	return func() {
		// re-set all the envs we unset above
		if err := os.Setenv("AWS_ACCESS_KEY_ID", e.Key); err != nil {
			t.Fatalf("Error resetting env var AWS_ACCESS_KEY_ID: %s", err)
		}
		if err := os.Setenv("AWS_SECRET_ACCESS_KEY", e.Secret); err != nil {
			t.Fatalf("Error resetting env var AWS_SECRET_ACCESS_KEY: %s", err)
		}
		if err := os.Setenv("AWS_SESSION_TOKEN", e.Token); err != nil {
			t.Fatalf("Error resetting env var AWS_SESSION_TOKEN: %s", err)
		}
		if err := os.Setenv("AWS_PROFILE", e.Profile); err != nil {
			t.Fatalf("Error setting env var AWS_PROFILE: %s", err)
		}
		if err := os.Setenv("AWS_SHARED_CREDENTIALS_FILE", s); err != nil {
			t.Fatalf("Error setting env var AWS_SHARED_CREDENTIALS_FLE: %s", err)
		}
	}
}

// awsMetadataApiMock establishes a httptest server to mock out the internal AWS Metadata
// service. IAM Credentials are retrieved by the EC2RoleProvider, which makes
// API calls to this internal URL. By replacing the server with a test server,
// we can simulate an AWS environment
func awsMetadataApiMock(endpoints []*endpoint) func() {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Add("Server", "MockEC2")
		log.Printf("[DEBUG] Mocker server received request to %q", r.RequestURI)
		for _, e := range endpoints {
			if r.RequestURI == e.Uri {
				fmt.Fprintln(w, e.Body)
				w.WriteHeader(200)
				return
			}
		}
		w.WriteHeader(400)
	}))

	os.Setenv("AWS_METADATA_URL", ts.URL+"/latest")
	return ts.Close
}

// invalidAwsEnv establishes a httptest server to simulate behaviour
// when endpoint doesn't respond as expected
func invalidAwsEnv(t *testing.T) func() {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
	}))

	os.Setenv("AWS_METADATA_URL", ts.URL+"/latest")
	return ts.Close
}

func getEnv() *currentEnv {
	// Grab any existing AWS keys and preserve. In some tests we'll unset these, so
	// we need to have them and restore them after
	return &currentEnv{
		Key:           os.Getenv("AWS_ACCESS_KEY_ID"),
		Secret:        os.Getenv("AWS_SECRET_ACCESS_KEY"),
		Token:         os.Getenv("AWS_SESSION_TOKEN"),
		Profile:       os.Getenv("AWS_PROFILE"),
		CredsFilename: os.Getenv("AWS_SHARED_CREDENTIALS_FILE"),
	}
}

// struct to preserve the current environment
type currentEnv struct {
	Key, Secret, Token, Profile, CredsFilename string
}

type endpoint struct {
	Uri  string `json:"uri"`
	Body string `json:"body"`
}

var ec2metadata_instanceIdEndpoint = &endpoint{
	Uri:  "/latest/meta-data/instance-id",
	Body: "mock-instance-id",
}

var ec2metadata_securityCredentialsEndpoints = []*endpoint{
	{
		Uri:  "/latest/meta-data/iam/security-credentials/",
		Body: "test_role",
	},
	{
		Uri:  "/latest/meta-data/iam/security-credentials/test_role",
		Body: "{\"Code\":\"Success\",\"LastUpdated\":\"2015-12-11T17:17:25Z\",\"Type\":\"AWS-HMAC\",\"AccessKeyId\":\"somekey\",\"SecretAccessKey\":\"somesecret\",\"Token\":\"sometoken\"}",
	},
}

var ec2metadata_iamInfoEndpoint = &endpoint{
	Uri:  "/latest/meta-data/iam/info",
	Body: "{\"Code\": \"Success\",\"LastUpdated\": \"2016-03-17T12:27:32Z\",\"InstanceProfileArn\": \"arn:aws:iam::000000000000:instance-profile/my-instance-profile\",\"InstanceProfileId\": \"AIPAABCDEFGHIJKLMN123\"}",
}

const ec2metadata_iamInfoEndpoint_expectedAccountID = `000000000000`
const ec2metadata_iamInfoEndpoint_expectedPartition = `aws`

const iamResponse_GetUser_valid = `<GetUserResponse xmlns="https://iam.amazonaws.com/doc/2010-05-08/">
  <GetUserResult>
    <User>
      <UserId>AIDACKCEVSQ6C2EXAMPLE</UserId>
      <Path>/division_abc/subdivision_xyz/</Path>
      <UserName>Bob</UserName>
      <Arn>arn:aws:iam::111111111111:user/division_abc/subdivision_xyz/Bob</Arn>
      <CreateDate>2013-10-02T17:01:44Z</CreateDate>
      <PasswordLastUsed>2014-10-10T14:37:51Z</PasswordLastUsed>
    </User>
  </GetUserResult>
  <ResponseMetadata>
    <RequestId>7a62c49f-347e-4fc4-9331-6e8eEXAMPLE</RequestId>
  </ResponseMetadata>
</GetUserResponse>`

const iamResponse_GetUser_valid_expectedAccountID = `111111111111`
const iamResponse_GetUser_valid_expectedPartition = `aws`

const iamResponse_GetUser_unauthorized = `<ErrorResponse xmlns="https://iam.amazonaws.com/doc/2010-05-08/">
  <Error>
    <Type>Sender</Type>
    <Code>AccessDenied</Code>
    <Message>User: arn:aws:iam::123456789012:user/Bob is not authorized to perform: iam:GetUser on resource: arn:aws:iam::123456789012:user/Bob</Message>
  </Error>
  <RequestId>7a62c49f-347e-4fc4-9331-6e8eEXAMPLE</RequestId>
</ErrorResponse>`

const stsResponse_GetCallerIdentity_valid = `<GetCallerIdentityResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
  <GetCallerIdentityResult>
   <Arn>arn:aws:iam::222222222222:user/Alice</Arn>
    <UserId>AKIAI44QH8DHBEXAMPLE</UserId>
    <Account>222222222222</Account>
  </GetCallerIdentityResult>
  <ResponseMetadata>
    <RequestId>01234567-89ab-cdef-0123-456789abcdef</RequestId>
  </ResponseMetadata>
</GetCallerIdentityResponse>`

const stsResponse_GetCallerIdentity_valid_expectedAccountID = `222222222222`
const stsResponse_GetCallerIdentity_valid_expectedPartition = `aws`

const stsResponse_GetCallerIdentity_unauthorized = `<ErrorResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
  <Error>
    <Type>Sender</Type>
    <Code>AccessDenied</Code>
    <Message>User: arn:aws:iam::123456789012:user/Bob is not authorized to perform: sts:GetCallerIdentity</Message>
  </Error>
  <RequestId>01234567-89ab-cdef-0123-456789abcdef</RequestId>
</ErrorResponse>`

const iamResponse_GetUser_federatedFailure = `<ErrorResponse xmlns="https://iam.amazonaws.com/doc/2010-05-08/">
  <Error>
    <Type>Sender</Type>
    <Code>ValidationError</Code>
    <Message>Must specify userName when calling with non-User credentials</Message>
  </Error>
  <RequestId>7a62c49f-347e-4fc4-9331-6e8eEXAMPLE</RequestId>
</ErrorResponse>`

const iamResponse_ListRoles_valid = `<ListRolesResponse xmlns="https://iam.amazonaws.com/doc/2010-05-08/">
  <ListRolesResult>
    <IsTruncated>true</IsTruncated>
    <Marker>AWceSSsKsazQ4IEplT9o4hURCzBs00iavlEvEXAMPLE</Marker>
    <Roles>
      <member>
        <Path>/</Path>
        <AssumeRolePolicyDocument>%7B%22Version%22%3A%222008-10-17%22%2C%22Statement%22%3A%5B%7B%22Sid%22%3A%22%22%2C%22Effect%22%3A%22Allow%22%2C%22Principal%22%3A%7B%22Service%22%3A%22ec2.amazonaws.com%22%7D%2C%22Action%22%3A%22sts%3AAssumeRole%22%7D%5D%7D</AssumeRolePolicyDocument>
        <RoleId>AROACKCEVSQ6C2EXAMPLE</RoleId>
        <RoleName>elasticbeanstalk-role</RoleName>
        <Arn>arn:aws:iam::444444444444:role/elasticbeanstalk-role</Arn>
        <CreateDate>2013-10-02T17:01:44Z</CreateDate>
      </member>
    </Roles>
  </ListRolesResult>
  <ResponseMetadata>
    <RequestId>7a62c49f-347e-4fc4-9331-6e8eEXAMPLE</RequestId>
  </ResponseMetadata>
</ListRolesResponse>`

const iamResponse_ListRoles_valid_expectedAccountID = `444444444444`
const iamResponse_ListRoles_valid_expectedPartition = `aws`

const iamResponse_ListRoles_unauthorized = `<ErrorResponse xmlns="https://iam.amazonaws.com/doc/2010-05-08/">
  <Error>
    <Type>Sender</Type>
    <Code>AccessDenied</Code>
    <Message>User: arn:aws:iam::123456789012:user/Bob is not authorized to perform: iam:ListRoles on resource: arn:aws:iam::123456789012:role/</Message>
  </Error>
  <RequestId>7a62c49f-347e-4fc4-9331-6e8eEXAMPLE</RequestId>
</ErrorResponse>`
