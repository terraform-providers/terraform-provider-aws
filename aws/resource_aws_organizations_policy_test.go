package aws

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func testAccAwsOrganizationsPolicy_basic(t *testing.T) {
	var policy organizations.Policy
	content1 := `{"Version": "2012-10-17", "Statement": { "Effect": "Allow", "Action": "*", "Resource": "*"}}`
	content2 := `{"Version": "2012-10-17", "Statement": { "Effect": "Allow", "Action": "s3:*", "Resource": "*"}}`
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_organizations_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOrganizationsPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsPolicyConfig_Required(rName, content1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsPolicyExists(resourceName, &policy),
					resource.TestMatchResourceAttr(resourceName, "arn", regexp.MustCompile(`^arn:[^:]+:organizations::[^:]+:policy/o-.+/service_control_policy/p-.+$`)),
					resource.TestCheckResourceAttr(resourceName, "content", content1),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "type", organizations.PolicyTypeServiceControlPolicy),
				),
			},
			{
				Config: testAccAwsOrganizationsPolicyConfig_Required(rName, content2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "content", content2),
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

// Reference: https://github.com/terraform-providers/terraform-provider-aws/issues/5073
func testAccAwsOrganizationsPolicy_concurrent(t *testing.T) {
	var policy1, policy2, policy3, policy4, policy5 organizations.Policy
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName1 := "aws_organizations_policy.test1"
	resourceName2 := "aws_organizations_policy.test2"
	resourceName3 := "aws_organizations_policy.test3"
	resourceName4 := "aws_organizations_policy.test4"
	resourceName5 := "aws_organizations_policy.test5"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOrganizationsPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsPolicyConfigConcurrent(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsPolicyExists(resourceName1, &policy1),
					testAccCheckAwsOrganizationsPolicyExists(resourceName2, &policy2),
					testAccCheckAwsOrganizationsPolicyExists(resourceName3, &policy3),
					testAccCheckAwsOrganizationsPolicyExists(resourceName4, &policy4),
					testAccCheckAwsOrganizationsPolicyExists(resourceName5, &policy5),
				),
			},
		},
	})
}

func testAccAwsOrganizationsPolicy_description(t *testing.T) {
	var policy organizations.Policy
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_organizations_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOrganizationsPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsPolicyConfig_Description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				Config: testAccAwsOrganizationsPolicyConfig_Description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
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

func testAccAwsOrganizationsPolicy_type(t *testing.T) {
	var policy organizations.Policy
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_organizations_policy.test"

	serviceControlPolicyContent := `{"Version": "2012-10-17", "Statement": { "Effect": "Allow", "Action": "*", "Resource": "*"}}`
	tagPolicyContent := `{ "tags": { "Product": { "tag_key": { "@@assign": "Product" }, "enforced_for": { "@@assign": [ "ec2:instance" ] } } } }`

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOrganizationsPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsPolicyConfig_Type(rName, serviceControlPolicyContent, organizations.PolicyTypeServiceControlPolicy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "type", organizations.PolicyTypeServiceControlPolicy),
				),
			},
			{
				Config: testAccAwsOrganizationsPolicyConfig_Type(rName, tagPolicyContent, organizations.PolicyTypeTagPolicy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "type", organizations.PolicyTypeTagPolicy),
				),
			},
			{
				Config: testAccAwsOrganizationsPolicyConfig_Required(rName, serviceControlPolicyContent),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "type", organizations.PolicyTypeServiceControlPolicy),
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

func testAccCheckAwsOrganizationsPolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).organizationsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_organizations_policy" {
			continue
		}

		input := &organizations.DescribePolicyInput{
			PolicyId: &rs.Primary.ID,
		}

		resp, err := conn.DescribePolicy(input)

		if isAWSErr(err, organizations.ErrCodeAWSOrganizationsNotInUseException, "") {
			continue
		}

		if isAWSErr(err, organizations.ErrCodePolicyNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if resp != nil && resp.Policy != nil {
			return fmt.Errorf("Policy %q still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccCheckAwsOrganizationsPolicyExists(resourceName string, policy *organizations.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).organizationsconn
		input := &organizations.DescribePolicyInput{
			PolicyId: &rs.Primary.ID,
		}

		resp, err := conn.DescribePolicy(input)

		if err != nil {
			return err
		}

		if resp == nil || resp.Policy == nil {
			return fmt.Errorf("Policy %q does not exist", rs.Primary.ID)
		}

		*policy = *resp.Policy

		return nil
	}
}

func testAccAwsOrganizationsPolicyConfig_Description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_policy" "test" {
  content     = "{\"Version\": \"2012-10-17\", \"Statement\": { \"Effect\": \"Allow\", \"Action\": \"*\", \"Resource\": \"*\"}}"
  description = "%s"
  name        = "%s"

  depends_on = ["aws_organizations_organization.test"]
}
`, description, rName)
}

func testAccAwsOrganizationsPolicyConfig_Required(rName, content string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_policy" "test" {
  content = %s
  name    = "%s"

  depends_on = ["aws_organizations_organization.test"]
}
`, strconv.Quote(content), rName)
}

func testAccAwsOrganizationsPolicyConfigConcurrent(rName string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_policy" "test1" {
  content = "{\"Version\": \"2012-10-17\", \"Statement\": { \"Effect\": \"Deny\", \"Action\": \"cloudtrail:StopLogging\", \"Resource\": \"*\"}}"
  name    = "%[1]s1"

  depends_on = ["aws_organizations_organization.test"]
}

resource "aws_organizations_policy" "test2" {
  content = "{\"Version\": \"2012-10-17\", \"Statement\": { \"Effect\": \"Deny\", \"Action\": \"ec2:DeleteFlowLogs\", \"Resource\": \"*\"}}"
  name    = "%[1]s2"

  depends_on = ["aws_organizations_organization.test"]
}

resource "aws_organizations_policy" "test3" {
  content = "{\"Version\": \"2012-10-17\", \"Statement\": { \"Effect\": \"Deny\", \"Action\": \"logs:DeleteLogGroup\", \"Resource\": \"*\"}}"
  name    = "%[1]s3"

  depends_on = ["aws_organizations_organization.test"]
}

resource "aws_organizations_policy" "test4" {
  content = "{\"Version\": \"2012-10-17\", \"Statement\": { \"Effect\": \"Deny\", \"Action\": \"config:DeleteConfigRule\", \"Resource\": \"*\"}}"
  name    = "%[1]s4"

  depends_on = ["aws_organizations_organization.test"]
}

resource "aws_organizations_policy" "test5" {
  content = "{\"Version\": \"2012-10-17\", \"Statement\": { \"Effect\": \"Deny\", \"Action\": \"iam:DeleteRolePermissionsBoundary\", \"Resource\": \"*\"}}"
  name    = "%[1]s5"

  depends_on = ["aws_organizations_organization.test"]
}
`, rName)
}

func testAccAwsOrganizationsPolicyConfig_Type(rName, content, policyType string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_policy" "test" {
  content     = %s
  name        = "%s"
  type        = "%s"

  depends_on = ["aws_organizations_organization.test"]
}
`, strconv.Quote(content), rName, policyType)
}
