package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSSESReceiptRuleSet_basic(t *testing.T) {
	resourceName := "aws_ses_receipt_rule_set.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSESReceiptRuleSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSESReceiptRuleSetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESReceiptRuleSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_set_name", rName),
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

func testAccCheckSESReceiptRuleSetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sesConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ses_receipt_rule_set" {
			continue
		}

		params := &ses.DescribeReceiptRuleSetInput{
			RuleSetName: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeReceiptRuleSet(params)

		if isAWSErr(err, ses.ErrCodeRuleSetDoesNotExistException, "") {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("SES Receipt Rule Set (%s) still exists", rs.Primary.ID)
	}

	return nil

}

func testAccCheckAwsSESReceiptRuleSetExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES Receipt Rule Set not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES Receipt Rule Set name not set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sesConn

		params := &ses.DescribeReceiptRuleSetInput{
			RuleSetName: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeReceiptRuleSet(params)
		return err
	}
}

func testAccAWSSESReceiptRuleSetConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ses_receipt_rule_set" "test" {
  rule_set_name = %q
}
`, rName)
}
