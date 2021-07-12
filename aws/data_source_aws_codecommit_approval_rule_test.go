package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSCodeCommitApprovalRuleTemplateDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acctest-%d", acctest.RandInt())
	resourceName := "aws_codecommit_approval_rule.default"
	datasourceName := "data.aws_codecommit_approval_rule.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsCodeCommitApprovalRuleTemplateDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "content", resourceName, "content"),
				),
			},
		},
	})
}

func testAccCheckAwsCodeCommitApprovalRuleTemplateDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_codecommit_approval_rule" "default" {
	name = "%s"
	content = <<EOF
{
	"DestinationReferences": ["refs/heads/master"],
	"Statements": [{
			"Type": "Approvers",
			"NumberOfApprovalsNeeded": 2,
			"ApprovalPoolMembers": ["arn:aws:sts::444194214726:assumed-role/CodeCommitReview/*"]}],
	"Version": "2018-11-08"
}
EOF
}

data "aws_codecommit_approval_rule" "default" {
  name = "${aws_codecommit_approval_rule.default.name}"
}
`, rName)
}
