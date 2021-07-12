package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSCodeCommitApprovalRule_basic(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_codecommit_approval_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCodeCommitApprovalRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCodeCommitApprovalRule_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCodeCommitApprovalRuleExists("aws_codecommit_approval_rule.test"),
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

func TestAccAWSCodeCommitApprovalRule_withChanges(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_codecommit_approval_rule.test"

	testContent, _ := structure.NormalizeJsonString(fmt.Sprintf(`
	{
		"Version": "2018-11-08",
		"DestinationReferences": ["refs/heads/master"],
		"Statements": [{
				"Type": "Approvers",
				"NumberOfApprovalsNeeded": 1,
				"ApprovalPoolMembers": ["arn:aws:sts::444194214726:assumed-role/CodeCommitReview/*"]}]
	}`))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCodeCommitApprovalRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCodeCommitApprovalRule_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCodeCommitApprovalRuleExists("aws_codecommit_approval_rule.test"),
					resource.TestCheckResourceAttr(
						"aws_codecommit_approval_rule.test", "description", "This is a test description"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCodeCommitApprovalRule_withChanges(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCodeCommitApprovalRuleExists("aws_codecommit_approval_rule.test"),
					resource.TestCheckResourceAttr(
						"aws_codecommit_approval_rule.test", "description", "This is a test description - with changes"),
					resource.TestCheckResourceAttr(
						"aws_codecommit_approval_rule.test", "content", testContent),
				),
			},
		},
	})
}

func testAccCheckCodeCommitApprovalRuleExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		codecommitconn := testAccProvider.Meta().(*AWSClient).codecommitconn
		out, err := codecommitconn.GetApprovalRuleTemplate(&codecommit.GetApprovalRuleTemplateInput{
			ApprovalRuleTemplateName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if out.ApprovalRuleTemplate.ApprovalRuleTemplateId == nil {
			return fmt.Errorf("No CodeCommit Approval Rule Template Found")
		}

		if *out.ApprovalRuleTemplate.ApprovalRuleTemplateName != rs.Primary.ID {
			return fmt.Errorf("CodeCommit Approval Rule Template mismatch - existing: %q, state: %q",
				*out.ApprovalRuleTemplate.ApprovalRuleTemplateName, rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckCodeCommitApprovalRuleDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).codecommitconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codecommit_approval_rule" {
			continue
		}

		_, err := conn.GetApprovalRuleTemplate(&codecommit.GetApprovalRuleTemplateInput{
			ApprovalRuleTemplateName: aws.String(rs.Primary.ID),
		})

		if ae, ok := err.(awserr.Error); ok && ae.Code() == "ApprovalRuleTemplateDoesNotExistException" {
			continue
		}
		if err == nil {
			return fmt.Errorf("Approval rule template still exists: %s", rs.Primary.ID)
		}
		return err
	}

	return nil
}

func testAccCodeCommitApprovalRule_basic(rInt int) string {
	return fmt.Sprintf(`
resource "aws_codecommit_approval_rule" "test" {
  name = "test_rule_%d"

  description = "This is a test description"

  content = <<EOF
  {
	  "Version": "2018-11-08",
	  "DestinationReferences": ["refs/heads/master"],
	  "Statements": [{
			  "Type": "Approvers",
			  "NumberOfApprovalsNeeded": 2,
			  "ApprovalPoolMembers": ["arn:aws:sts::444194214726:assumed-role/CodeCommitReview/*"]}]
  }
  EOF 
}
`, rInt)
}

func testAccCodeCommitApprovalRule_withChanges(rInt int) string {
	return fmt.Sprintf(`
resource "aws_codecommit_approval_rule" "test" {
  name        = "test_rule_%d"
  description = "This is a test description - with changes"

  content = <<EOF
  {
	  "Version": "2018-11-08",
	  "DestinationReferences": ["refs/heads/master"],
	  "Statements": [{
			  "Type": "Approvers",
			  "NumberOfApprovalsNeeded": 1,
			  "ApprovalPoolMembers": ["arn:aws:sts::444194214726:assumed-role/CodeCommitReview/*"]}]
  }
  EOF 
}
`, rInt)
}
