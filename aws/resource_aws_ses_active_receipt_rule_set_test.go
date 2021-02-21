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

// Only one SES Receipt RuleSet can be active at a time, so run serially
// locally and in TeamCity.
func TestAccAWSSESActiveReceiptRuleSet_serial(t *testing.T) {
	testFuncs := map[string]func(t *testing.T){
		"basic":      testAccAWSSESActiveReceiptRuleSet_basic,
		"disappears": testAccAWSSESActiveReceiptRuleSet_disappears,
	}

	for name, testFunc := range testFuncs {
		testFunc := testFunc

		t.Run(name, func(t *testing.T) {
			testFunc(t)
		})
	}
}

func testAccAWSSESActiveReceiptRuleSet_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ses_active_receipt_rule_set.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSSES(t)
			testAccPreCheckSESReceiptRule(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSESActiveReceiptRuleSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSESActiveReceiptRuleSetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESActiveReceiptRuleSetExists(resourceName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "ses", fmt.Sprintf("receipt-rule-set/%s", rName)),
				),
			},
		},
	})
}

func testAccAWSSESActiveReceiptRuleSet_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ses_active_receipt_rule_set.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSSES(t)
			testAccPreCheckSESReceiptRule(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSESActiveReceiptRuleSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSESActiveReceiptRuleSetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESActiveReceiptRuleSetExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSesActiveReceiptRuleSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSESActiveReceiptRuleSetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sesconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ses_active_receipt_rule_set" {
			continue
		}

		response, err := conn.DescribeActiveReceiptRuleSet(&ses.DescribeActiveReceiptRuleSetInput{})
		if err != nil {
			return err
		}

		if response.Metadata != nil && (aws.StringValue(response.Metadata.Name) == rs.Primary.ID) {
			return fmt.Errorf("Active receipt rule set still exists")
		}

	}

	return nil

}

func testAccCheckAwsSESActiveReceiptRuleSetExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES Active Receipt Rule Set not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES Active Receipt Rule Set name not set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sesconn

		response, err := conn.DescribeActiveReceiptRuleSet(&ses.DescribeActiveReceiptRuleSetInput{})
		if err != nil {
			return err
		}

		if response.Metadata != nil && (aws.StringValue(response.Metadata.Name) != rs.Primary.ID) {
			return fmt.Errorf("The active receipt rule set (%s) was not set to %s", aws.StringValue(response.Metadata.Name), rs.Primary.ID)
		}

		return nil
	}
}

func testAccAWSSESActiveReceiptRuleSetConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_ses_receipt_rule_set" "test" {
  rule_set_name = %[1]q
}

resource "aws_ses_active_receipt_rule_set" "test" {
  rule_set_name = aws_ses_receipt_rule_set.test.rule_set_name
}
`, name)
}
