package aws

import (
	"fmt"
	"sort"

	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSCodeCommitApprovalRuleAssociation_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCodeCommitApprovalRuleAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCodeCommitApprovalRuleAssociation_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCodeCommitApprovalRuleAssociationExists("aws_codecommit_approval_rule_association.test"),
				),
			},
		},
	})
}

func TestAccAWSCodeCommitApprovalRuleAssociation_withChanges(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCodeCommitApprovalRuleAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCodeCommitApprovalRuleAssociation_setup(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCodeCommitApprovalRuleAssociationExists("aws_codecommit_approval_rule_association.test_change"),
					resource.TestCheckResourceAttr(
						"aws_codecommit_approval_rule_association.test_change", "repository_names.#", "1",
					),
				),
			},
			{
				Config: testAccCodeCommitApprovalRuleAssociation_withChanges(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCodeCommitApprovalRuleAssociationExists("aws_codecommit_approval_rule_association.test_change"),
					resource.TestCheckResourceAttr(
						"aws_codecommit_approval_rule_association.test_change", "repository_names.#", "2",
					),
				),
			},
		},
	})
}

func testAccCheckCodeCommitApprovalRuleAssociationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		codecommitconn := testAccProvider.Meta().(*AWSClient).codecommitconn
		out, err := codecommitconn.ListRepositoriesForApprovalRuleTemplate(&codecommit.ListRepositoriesForApprovalRuleTemplateInput{
			ApprovalRuleTemplateName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if strings.Join(aws.StringValueSlice(out.RepositoryNames), "") == "" {
			return fmt.Errorf("No repositories associated with approval rule template: %q", rs.Primary.ID)
		}

		var repoAttributes []string

		for k, v := range rs.Primary.Attributes {
			if strings.Contains(k, "repository_names") && !strings.Contains(k, "repository_names.#") {
				repoAttributes = append(repoAttributes, v)
			}
		}

		repoActual := aws.StringValueSlice(out.RepositoryNames)

		sort.Strings(repoAttributes)
		sort.Strings(repoActual)

		if strings.Join(repoActual, ",") != strings.Join(repoAttributes, ",") {
			return fmt.Errorf("CodeCommit Approval Rule Association mismatch - existing: %q, state: %q",
				strings.Join(aws.StringValueSlice(out.RepositoryNames), ","), repoAttributes)
		}

		return nil
	}
}

func testAccCheckCodeCommitApprovalRuleAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).codecommitconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codecommit_approval_rule_association" {
			continue
		}

		out, err := conn.ListRepositoriesForApprovalRuleTemplate(&codecommit.ListRepositoriesForApprovalRuleTemplateInput{
			ApprovalRuleTemplateName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if len(out.RepositoryNames) == 0 {
			return nil
		} else {
			return fmt.Errorf("Approval rule template associations still exist: %s", rs.Primary.ID)
		}

	}
	return nil
}

func testAccCodeCommitApprovalRuleAssociation_basic() string {
	return fmt.Sprintf(`
resource "aws_codecommit_approval_rule_association" "test" {
  template_name    = "test-rule"
  repository_names = ["test1"]
}
`)
}

func testAccCodeCommitApprovalRuleAssociation_setup() string {
	return fmt.Sprintf(`
resource "aws_codecommit_approval_rule_association" "test_change" {
  template_name    = "test2-rule"
  repository_names = ["test1"]
}
`)
}

func testAccCodeCommitApprovalRuleAssociation_withChanges() string {
	return fmt.Sprintf(`
resource "aws_codecommit_approval_rule_association" "test_change" {
  template_name    = "test2-rule"
  repository_names = ["test1", "test2"]
}
`)
}
