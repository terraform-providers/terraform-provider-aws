package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAssumedRoleRoleSessionName(t *testing.T) {
	testCases := []struct {
		Name                string
		ARN                 string
		ExpectedRoleName    string
		ExpectedSessionName string
		ExpectedError       bool
	}{
		{
			Name:                "not an ARN",
			ARN:                 "abcd",
			ExpectedRoleName:    "",
			ExpectedSessionName: "",
		},
		{
			Name:                "regular role ARN",
			ARN:                 "arn:aws:iam::111122223333:role/role_name", //lintignore:AWSAT005
			ExpectedRoleName:    "",
			ExpectedSessionName: "",
		},
		{
			Name:                "assumed role ARN",
			ARN:                 "arn:aws:sts::444433332222:assumed-role/something_something-admin/sessionIDNotPartOfRoleARN", //lintignore:AWSAT005
			ExpectedRoleName:    "something_something-admin",
			ExpectedSessionName: "sessionIDNotPartOfRoleARN",
		},
		{
			Name:                "'assumed-role' part of ARN resource",
			ARN:                 "arn:aws:iam::444433332222:user/assumed-role-but-not-really", //lintignore:AWSAT005
			ExpectedRoleName:    "",
			ExpectedSessionName: "",
		},
		{
			Name:                "user ARN",
			ARN:                 "arn:aws:iam::123456789012:user/Bob", //lintignore:AWSAT005
			ExpectedRoleName:    "",
			ExpectedSessionName: "",
		},
		{
			Name:                "assumed role from AWS example",
			ARN:                 "arn:aws:sts::123456789012:assumed-role/example-role/AWSCLI-Session", //lintignore:AWSAT005
			ExpectedRoleName:    "example-role",
			ExpectedSessionName: "AWSCLI-Session",
		},
		{
			Name:                "multiple slashes in resource",                                         // not sure this is even valid
			ARN:                 "arn:aws:sts::123456789012:assumed-role/path/role-name/AWSCLI-Session", //lintignore:AWSAT005
			ExpectedRoleName:    "role-name",
			ExpectedSessionName: "AWSCLI-Session",
		},
		{
			Name:                "not an sts ARN",
			ARN:                 "arn:aws:iam::123456789012:assumed-role/example-role/AWSCLI-Session", //lintignore:AWSAT005
			ExpectedRoleName:    "",
			ExpectedSessionName: "",
		},
		{
			Name:                "role with path",
			ARN:                 "arn:aws:iam::123456789012:role/this/is/the/path/role-name", //lintignore:AWSAT005
			ExpectedRoleName:    "",
			ExpectedSessionName: "",
		},
		{
			Name:                "wrong service",
			ARN:                 "arn:aws:ec2::123456789012:role/role-name", //lintignore:AWSAT005
			ExpectedRoleName:    "",
			ExpectedSessionName: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			role, session := roleNameSessionFromARN(testCase.ARN)

			if testCase.ExpectedRoleName != role || testCase.ExpectedSessionName != session {
				t.Errorf("for %s: got role %s, session %s; expected role %s, session %s", testCase.ARN, role, session, testCase.ExpectedRoleName, testCase.ExpectedSessionName)
			}
		})
	}
}

func TestAccAWSDataSourceIAMSessionContext_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_iam_session_context.test"
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, iam.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsIAMSessionContextConfig(rName, "/", "session-id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "issuer_arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "issuer_id", resourceName, "unique_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "issuer_name", resourceName, "name"),
					resource.TestCheckResourceAttr(dataSourceName, "session_name", "session-id"),
				),
			},
		},
	})
}

func TestAccAWSDataSourceIAMSessionContext_withPath(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_iam_session_context.test"
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, iam.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsIAMSessionContextConfig(rName, "/this/is/a/long/path/", "session-id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "issuer_arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "issuer_name", resourceName, "name"),
					resource.TestCheckResourceAttr(dataSourceName, "session_name", "session-id"),
				),
			},
		},
	})
}

func TestAccAWSDataSourceIAMSessionContext_notAssumedRole(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_iam_session_context.test"
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, iam.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsIAMSessionContextNotAssumedConfig(rName, "/"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "issuer_arn", resourceName, "arn"),
					resource.TestCheckResourceAttr(dataSourceName, "issuer_name", ""),
					resource.TestCheckResourceAttr(dataSourceName, "session_name", ""),
				),
			},
		},
	})
}

func TestAccAWSDataSourceIAMSessionContext_notAssumedRoleWithPath(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_iam_session_context.test"
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, iam.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsIAMSessionContextNotAssumedConfig(rName, "/this/is/a/long/path/"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "issuer_arn", resourceName, "arn"),
					resource.TestCheckResourceAttr(dataSourceName, "issuer_name", ""),
					resource.TestCheckResourceAttr(dataSourceName, "session_name", ""),
				),
			},
		},
	})
}

func TestAccAWSDataSourceIAMSessionContext_notAssumedRoleUser(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_iam_session_context.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, iam.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsIAMSessionContextUserConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceAttrGlobalARN(dataSourceName, "arn", "iam", fmt.Sprintf("user/division/extra-division/not-assumed-role/%[1]s", rName)),
					resource.TestCheckResourceAttr(dataSourceName, "issuer_name", ""),
					resource.TestCheckResourceAttr(dataSourceName, "session_name", ""),
				),
			},
		},
	})
}

func testAccAwsIAMSessionContextConfig(rName, path, sessionID string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q
  path = %[2]q

  assume_role_policy = jsonencode({
    "Version" = "2012-10-17"

    "Statement" = [{
      "Action" = "sts:AssumeRole"
      "Principal" = {
        "Service" = "ec2.${data.aws_partition.current.dns_suffix}"
      }
      "Effect" = "Allow"
    }]
  })
}

data "aws_partition" "current" {}
data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "test" {
  arn = "arn:${data.aws_partition.current.partition}:sts::${data.aws_caller_identity.current.account_id}:assumed-role/${aws_iam_role.test.name}/%[3]s"
}
`, rName, path, sessionID)
}

func testAccAwsIAMSessionContextNotAssumedConfig(rName, path string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = %[2]q

  assume_role_policy = jsonencode({
    "Version" = "2012-10-17"

    "Statement" = [{
      "Action" = "sts:AssumeRole"
      "Principal" = {
        "Service" = "ec2.${data.aws_partition.current.dns_suffix}"
      }
      "Effect" = "Allow"
    }]
  })
}

data "aws_iam_session_context" "test" {
  arn = aws_iam_role.test.arn
}
`, rName, path)
}

func testAccAwsIAMSessionContextUserConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "test" {
  arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:user/division/extra-division/not-assumed-role/%[1]s"
}
`, rName)
}
