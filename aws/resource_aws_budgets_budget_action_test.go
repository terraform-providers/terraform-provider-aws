package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/budgets"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/budgets/finder"
)

func init() {
	resource.AddTestSweepers("aws_budgets_budget_action", &resource.Sweeper{
		Name: "aws_budgets_budget_action",
		F:    testSweepBudgetsBudgetActionss,
	})
}

func testSweepBudgetsBudgetActionss(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).budgetconn
	accountID := client.(*AWSClient).accountid
	input := &budgets.DescribeBudgetActionsForAccountInput{
		AccountId: aws.String(accountID),
	}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.DescribeBudgetActionsForAccount(input)
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Budgets sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
		}
		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Budgets: %w", err))
			return sweeperErrs
		}

		for _, action := range output.Actions {
			name := aws.StringValue(action.BudgetName)
			log.Printf("[INFO] Deleting Budget Action: %s", name)
			id := fmt.Sprintf("%s:%s:%s", accountID, aws.StringValue(action.ActionId), name)

			r := resourceAwsBudgetsBudgetAction()
			d := r.Data(nil)
			d.SetId(id)

			err := r.Delete(d, client)
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Budget Action (%s): %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}
		input.NextToken = output.NextToken
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSBudgetsBudgetAction_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_budgets_budget_action.test"
	var conf budgets.Action

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(budgets.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, budgets.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccAWSBudgetsBudgetActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBudgetsBudgetActionConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetActionExists(resourceName, &conf),
					testAccMatchResourceAttrGlobalARN(resourceName, "arn", "budgetservice", regexp.MustCompile(fmt.Sprintf(`budget/%s/action/.+`, rName))),
					resource.TestCheckResourceAttrPair(resourceName, "budget_name", "aws_budgets_budget.test", "name"),
					resource.TestCheckResourceAttrPair(resourceName, "execution_role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "action_type", "APPLY_IAM_POLICY"),
					resource.TestCheckResourceAttr(resourceName, "approval_model", "AUTOMATIC"),
					resource.TestCheckResourceAttr(resourceName, "notification_type", "ACTUAL"),
					resource.TestCheckResourceAttr(resourceName, "action_threshold.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action_threshold.0.action_threshold_type", "ABSOLUTE_VALUE"),
					resource.TestCheckResourceAttr(resourceName, "action_threshold.0.action_threshold_value", "100"),
					resource.TestCheckResourceAttr(resourceName, "definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.iam_action_definition.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "definition.0.iam_action_definition.0.policy_arn", "aws_iam_policy.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.iam_action_definition.0.roles.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "subscriber.#", "1"),
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

func TestAccAWSBudgetsBudgetAction_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_budgets_budget_action.test"
	var conf budgets.Action

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(budgets.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, budgets.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccAWSBudgetsBudgetActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBudgetsBudgetActionConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetActionExists(resourceName, &conf),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsBudgetsBudgetAction(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAWSBudgetsBudgetActionExists(resourceName string, config *budgets.Action) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).budgetconn
		out, err := finder.ActionById(conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("Describe budget action error: %v", err)
		}

		if out.Action == nil {
			return fmt.Errorf("No budget Action returned %v in %v", out.Action, out)
		}

		*out.Action = *config

		return nil
	}
}

func testAccAWSBudgetsBudgetActionDestroy(s *terraform.State) error {
	meta := testAccProvider.Meta()
	conn := meta.(*AWSClient).budgetconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_budgets_budget_action" {
			continue
		}

		_, err := finder.ActionById(conn, rs.Primary.ID)
		if !isAWSErr(err, budgets.ErrCodeNotFoundException, "") {
			return fmt.Errorf("Budget Action '%s' was not deleted properly", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAWSBudgetsBudgetActionConfigBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test" {
  name        = %[1]q
  description = "My test policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "budgets.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_budgets_budget" "test" {
  name              = %[1]q
  budget_type       = "USAGE"
  limit_amount      = "10.0"
  limit_unit        = "dollars"
  time_period_start = "2006-01-02_15:04"
  time_unit         = "MONTHLY"
}

resource "aws_budgets_budget_action" "test" {
  budget_name        = aws_budgets_budget.test.name
  action_type        = "APPLY_IAM_POLICY"
  approval_model     = "AUTOMATIC"
  notification_type  = "ACTUAL"
  execution_role_arn = aws_iam_role.test.arn

  action_threshold {
    action_threshold_type  = "ABSOLUTE_VALUE"
    action_threshold_value = 100
  }

  definition {
    iam_action_definition {
      policy_arn = aws_iam_policy.test.arn
      roles      = [aws_iam_role.test.name]
    }
  }

  subscriber {
    address           = "test@test.test"
    subscription_type = "EMAIL"
  }
}
`, rName)
}
