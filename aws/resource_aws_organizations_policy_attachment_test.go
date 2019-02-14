package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsOrganizationsPolicyAttachment_account(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_organizations_policy_attachment.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOrganizationsPolicyAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsPolicyAttachmentConfig_Account(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsPolicyAttachmentExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "policy_id"),
					resource.TestCheckResourceAttrSet(resourceName, "target_id"),
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

func testAccCheckAwsOrganizationsPolicyAttachmentDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).organizationsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_organizations_policy_attachment" {
			continue
		}

		targetID, policyID, err := decodeAwsOrganizationsPolicyAttachmentID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &organizations.ListPoliciesForTargetInput{
			Filter:   aws.String(organizations.PolicyTypeServiceControlPolicy),
			TargetId: aws.String(targetID),
		}

		log.Printf("[DEBUG] Listing Organizations Policies for Target: %s", input)
		var output *organizations.PolicySummary
		err = conn.ListPoliciesForTargetPages(input, func(page *organizations.ListPoliciesForTargetOutput, lastPage bool) bool {
			for _, policySummary := range page.Policies {
				if aws.StringValue(policySummary.Id) == policyID {
					output = policySummary
					return true
				}
			}
			return !lastPage
		})

		if err != nil {
			if isAWSErr(err, organizations.ErrCodeTargetNotFoundException, "") {
				return nil
			}
			return err
		}

		if output == nil {
			return nil
		}

		return fmt.Errorf("Policy attachment %q still exists", rs.Primary.ID)
	}

	return nil

}

func testAccCheckAwsOrganizationsPolicyAttachmentExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).organizationsconn

		targetID, policyID, err := decodeAwsOrganizationsPolicyAttachmentID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &organizations.ListPoliciesForTargetInput{
			Filter:   aws.String(organizations.PolicyTypeServiceControlPolicy),
			TargetId: aws.String(targetID),
		}

		log.Printf("[DEBUG] Listing Organizations Policies for Target: %s", input)
		var output *organizations.PolicySummary
		err = conn.ListPoliciesForTargetPages(input, func(page *organizations.ListPoliciesForTargetOutput, lastPage bool) bool {
			for _, policySummary := range page.Policies {
				if aws.StringValue(policySummary.Id) == policyID {
					output = policySummary
					return true
				}
			}
			return !lastPage
		})

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("Policy attachment %q does not exist", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAwsOrganizationsPolicyAttachmentConfig_Account(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_organizations_policy" "test" {
  content     = "{\"Version\": \"2012-10-17\", \"Statement\": { \"Effect\": \"Allow\", \"Action\": \"*\", \"Resource\": \"*\"}}"
  name        = "%s"
}

resource "aws_organizations_policy_attachment" "test" {
  policy_id = "${aws_organizations_policy.test.id}"
  target_id = "${data.aws_caller_identity.current.account_id}"
}
`, rName)
}
