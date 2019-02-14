package aws

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsOrganizationsPolicy_basic(t *testing.T) {
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

func TestAccAwsOrganizationsPolicy_description(t *testing.T) {
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

		if err != nil {
			if isAWSErr(err, organizations.ErrCodePolicyNotFoundException, "") {
				return nil
			}
			return err
		}

		if resp == nil && resp.Policy != nil {
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
resource "aws_organizations_policy" "test" {
  content     = "{\"Version\": \"2012-10-17\", \"Statement\": { \"Effect\": \"Allow\", \"Action\": \"*\", \"Resource\": \"*\"}}"
  description = "%s"
  name        = "%s"
}
`, description, rName)
}

func testAccAwsOrganizationsPolicyConfig_Required(rName, content string) string {
	return fmt.Sprintf(`
resource "aws_organizations_policy" "test" {
  content = %s
  name    = "%s"
}
`, strconv.Quote(content), rName)
}
