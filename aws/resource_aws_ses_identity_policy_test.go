package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSSESIdentityPolicy_basic(t *testing.T) {
	domain := fmt.Sprintf(
		"%s.terraformtesting.com",
		acctest.RandString(10))
	resourceName := "aws_ses_identity_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSESIdentityPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSESIdentityPolicyConfigIdentityDomain(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESIdentityPolicyExists(resourceName),
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

func TestAccAWSSESIdentityPolicy_Identity_Email(t *testing.T) {
	email := fmt.Sprintf(
		"%s@terraformtesting.com",
		acctest.RandString(10))
	resourceName := "aws_ses_identity_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSESIdentityPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSESIdentityPolicyConfigIdentityEmail(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESIdentityPolicyExists(resourceName),
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

func TestAccAWSSESIdentityPolicy_Policy(t *testing.T) {
	domain := fmt.Sprintf(
		"%s.terraformtesting.com",
		acctest.RandString(10))
	resourceName := "aws_ses_identity_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSESIdentityPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSESIdentityPolicyConfigPolicy1(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESIdentityPolicyExists(resourceName),
				),
			},
			{
				Config: testAccAWSSESIdentityPolicyConfigPolicy2(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESIdentityPolicyExists(resourceName),
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

func testAccCheckAwsSESIdentityPolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sesconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ses_identity_policy" {
			continue
		}

		identityARN, policyName, err := resourceAwsSesIdentityPolicyParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &ses.GetIdentityPoliciesInput{
			Identity:    aws.String(identityARN),
			PolicyNames: aws.StringSlice([]string{policyName}),
		}

		output, err := conn.GetIdentityPolicies(input)

		if err != nil {
			return err
		}

		if output != nil && len(output.Policies) > 0 && aws.StringValue(output.Policies[policyName]) != "" {
			return fmt.Errorf("SES Identity (%s) Policy (%s) still exists", identityARN, policyName)
		}
	}

	return nil
}

func testAccCheckAwsSESIdentityPolicyExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("SES Identity Policy not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES Identity Policy ID not set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sesconn

		identityARN, policyName, err := resourceAwsSesIdentityPolicyParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &ses.GetIdentityPoliciesInput{
			Identity:    aws.String(identityARN),
			PolicyNames: aws.StringSlice([]string{policyName}),
		}

		output, err := conn.GetIdentityPolicies(input)

		if err != nil {
			return err
		}

		if output == nil || len(output.Policies) == 0 {
			return fmt.Errorf("SES Identity (%s) Policy (%s) not found", identityARN, policyName)
		}

		return nil
	}
}

func testAccAWSSESIdentityPolicyConfigIdentityDomain(domain string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "test" {
  statement {
    actions   = ["SES:SendEmail", "SES:SendRawEmail"]
    resources = [aws_ses_domain_identity.test.arn]

    principals {
      identifiers = ["*"]
      type        = "AWS"
    }
  }
}

resource "aws_ses_domain_identity" "test" {
  domain = %[1]q
}

resource "aws_ses_identity_policy" "test" {
  identity = aws_ses_domain_identity.test.arn
  name     = "test"
  policy   = data.aws_iam_policy_document.test.json
}
`, domain)
}

func testAccAWSSESIdentityPolicyConfigIdentityEmail(email string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "test" {
  statement {
    actions   = ["SES:SendEmail", "SES:SendRawEmail"]
    resources = [aws_ses_email_identity.test.arn]

    principals {
      identifiers = ["*"]
      type        = "AWS"
    }
  }
}

resource "aws_ses_email_identity" "test" {
  email = %[1]q
}

resource "aws_ses_identity_policy" "test" {
  identity = aws_ses_email_identity.test.email
  name     = "test"
  policy   = data.aws_iam_policy_document.test.json
}
`, email)
}

func testAccAWSSESIdentityPolicyConfigPolicy1(domain string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "test" {
  statement {
    actions   = ["SES:SendEmail", "SES:SendRawEmail"]
    resources = [aws_ses_domain_identity.test.arn]

    principals {
      identifiers = ["*"]
      type        = "AWS"
    }
  }
}

resource "aws_ses_domain_identity" "test" {
  domain = %[1]q
}

resource "aws_ses_identity_policy" "test" {
  identity = aws_ses_domain_identity.test.arn
  name     = "test"
  policy   = data.aws_iam_policy_document.test.json
}
`, domain)
}

func testAccAWSSESIdentityPolicyConfigPolicy2(domain string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    actions   = ["SES:SendEmail", "SES:SendRawEmail"]
    resources = [aws_ses_domain_identity.test.arn]

    principals {
      identifiers = [data.aws_caller_identity.current.account_id]
      type        = "AWS"
    }
  }
}

resource "aws_ses_domain_identity" "test" {
  domain = %[1]q
}

resource "aws_ses_identity_policy" "test" {
  identity = aws_ses_domain_identity.test.arn
  name     = "test"
  policy   = data.aws_iam_policy_document.test.json
}
`, domain)
}
