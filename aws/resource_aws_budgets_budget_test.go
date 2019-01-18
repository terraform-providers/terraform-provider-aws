package aws

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"testing"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/budgets"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSBudgetsBudget_basic(t *testing.T) {
	costFilterKey := "AZ"
	name := fmt.Sprintf("test-budget-%d", acctest.RandInt())
	configBasicDefaults := testAccAWSBudgetsBudgetConfigDefaults(name)
	accountID := "012345678910"
	configBasicUpdate := testAccAWSBudgetsBudgetConfigUpdate(name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccAWSBudgetsBudgetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBudgetsBudgetConfig_BasicDefaults(configBasicDefaults, costFilterKey),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExists("aws_budgets_budget.foo", configBasicDefaults),
					resource.TestMatchResourceAttr("aws_budgets_budget.foo", "name", regexp.MustCompile(*configBasicDefaults.BudgetName)),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "budget_type", *configBasicDefaults.BudgetType),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "limit_amount", *configBasicDefaults.BudgetLimit.Amount),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "limit_unit", *configBasicDefaults.BudgetLimit.Unit),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_period_start", configBasicDefaults.TimePeriod.Start.Format("2006-01-02_15:04")),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_period_end", configBasicDefaults.TimePeriod.End.Format("2006-01-02_15:04")),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_unit", *configBasicDefaults.TimeUnit),
				),
			},
			{
				PlanOnly:    true,
				Config:      testAccAWSBudgetsBudgetConfig_WithAccountID(configBasicDefaults, accountID, costFilterKey),
				ExpectError: regexp.MustCompile("account_id.*" + accountID),
			},
			{
				Config: testAccAWSBudgetsBudgetConfig_Basic(configBasicUpdate, costFilterKey),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExists("aws_budgets_budget.foo", configBasicUpdate),
					resource.TestMatchResourceAttr("aws_budgets_budget.foo", "name", regexp.MustCompile(*configBasicUpdate.BudgetName)),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "budget_type", *configBasicUpdate.BudgetType),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "limit_amount", *configBasicUpdate.BudgetLimit.Amount),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "limit_unit", *configBasicUpdate.BudgetLimit.Unit),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_period_start", configBasicUpdate.TimePeriod.Start.Format("2006-01-02_15:04")),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_period_end", configBasicUpdate.TimePeriod.End.Format("2006-01-02_15:04")),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_unit", *configBasicUpdate.TimeUnit),
				),
			},
			{
				ResourceName:            "aws_budgets_budget.foo",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccAWSBudgetsBudget_prefix(t *testing.T) {
	costFilterKey := "AZ"
	name := "test-budget-"
	configBasicDefaults := testAccAWSBudgetsBudgetConfigDefaults(name)
	configBasicUpdate := testAccAWSBudgetsBudgetConfigUpdate(name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccAWSBudgetsBudgetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBudgetsBudgetConfig_PrefixDefaults(configBasicDefaults, costFilterKey),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExists("aws_budgets_budget.foo", configBasicDefaults),
					resource.TestMatchResourceAttr("aws_budgets_budget.foo", "name_prefix", regexp.MustCompile(*configBasicDefaults.BudgetName)),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "budget_type", *configBasicDefaults.BudgetType),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "limit_amount", *configBasicDefaults.BudgetLimit.Amount),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "limit_unit", *configBasicDefaults.BudgetLimit.Unit),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_period_start", configBasicDefaults.TimePeriod.Start.Format("2006-01-02_15:04")),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_period_end", configBasicDefaults.TimePeriod.End.Format("2006-01-02_15:04")),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_unit", *configBasicDefaults.TimeUnit),
				),
			},

			{
				Config: testAccAWSBudgetsBudgetConfig_Prefix(configBasicUpdate, costFilterKey),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExists("aws_budgets_budget.foo", configBasicUpdate),
					resource.TestMatchResourceAttr("aws_budgets_budget.foo", "name_prefix", regexp.MustCompile(*configBasicUpdate.BudgetName)),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "budget_type", *configBasicUpdate.BudgetType),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "limit_amount", *configBasicUpdate.BudgetLimit.Amount),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "limit_unit", *configBasicUpdate.BudgetLimit.Unit),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_period_start", configBasicUpdate.TimePeriod.Start.Format("2006-01-02_15:04")),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_period_end", configBasicUpdate.TimePeriod.End.Format("2006-01-02_15:04")),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_unit", *configBasicUpdate.TimeUnit),
				),
			},

			{
				ResourceName:            "aws_budgets_budget.foo",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func testAccAWSBudgetsBudgetExists(resourceName string, config budgets.Budget) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		accountID, budgetName, err := decodeBudgetsBudgetID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed decoding ID: %v", err)
		}

		client := testAccProvider.Meta().(*AWSClient).budgetconn
		b, err := client.DescribeBudget(&budgets.DescribeBudgetInput{
			BudgetName: &budgetName,
			AccountId:  &accountID,
		})

		if err != nil {
			return fmt.Errorf("Describebudget error: %v", err)
		}

		if b.Budget == nil {
			return fmt.Errorf("No budget returned %v in %v", b.Budget, b)
		}

		if *b.Budget.BudgetLimit.Amount != *config.BudgetLimit.Amount {
			return fmt.Errorf("budget limit incorrectly set %v != %v", *config.BudgetLimit.Amount, *b.Budget.BudgetLimit.Amount)
		}

		if err := testAccAWSBudgetsBudgetCheckCostTypes(config, *b.Budget.CostTypes); err != nil {
			return err
		}

		if err := testAccAWSBudgetsBudgetCheckTimePeriod(*config.TimePeriod, *b.Budget.TimePeriod); err != nil {
			return err
		}

		if !reflect.DeepEqual(b.Budget.CostFilters, config.CostFilters) {
			return fmt.Errorf("cost filter not set properly: %v != %v", b.Budget.CostFilters, config.CostFilters)
		}

		return nil
	}
}

func testAccAWSBudgetsBudgetCheckTimePeriod(configTimePeriod, timePeriod budgets.TimePeriod) error {
	if configTimePeriod.End.Format("2006-01-02_15:04") != timePeriod.End.Format("2006-01-02_15:04") {
		return fmt.Errorf("TimePeriodEnd not set properly '%v' should be '%v'", *timePeriod.End, *configTimePeriod.End)
	}

	if configTimePeriod.Start.Format("2006-01-02_15:04") != timePeriod.Start.Format("2006-01-02_15:04") {
		return fmt.Errorf("TimePeriodStart not set properly '%v' should be '%v'", *timePeriod.Start, *configTimePeriod.Start)
	}

	return nil
}

func testAccAWSBudgetsBudgetCheckCostTypes(config budgets.Budget, costTypes budgets.CostTypes) error {
	if *costTypes.IncludeCredit != *config.CostTypes.IncludeCredit {
		return fmt.Errorf("IncludeCredit not set properly '%v' should be '%v'", *costTypes.IncludeCredit, *config.CostTypes.IncludeCredit)
	}

	if *costTypes.IncludeOtherSubscription != *config.CostTypes.IncludeOtherSubscription {
		return fmt.Errorf("IncludeOtherSubscription not set properly '%v' should be '%v'", *costTypes.IncludeOtherSubscription, *config.CostTypes.IncludeOtherSubscription)
	}

	if *costTypes.IncludeRecurring != *config.CostTypes.IncludeRecurring {
		return fmt.Errorf("IncludeRecurring not set properly  '%v' should be '%v'", *costTypes.IncludeRecurring, *config.CostTypes.IncludeRecurring)
	}

	if *costTypes.IncludeRefund != *config.CostTypes.IncludeRefund {
		return fmt.Errorf("IncludeRefund not set properly '%v' should be '%v'", *costTypes.IncludeRefund, *config.CostTypes.IncludeRefund)
	}

	if *costTypes.IncludeSubscription != *config.CostTypes.IncludeSubscription {
		return fmt.Errorf("IncludeSubscription not set properly '%v' should be '%v'", *costTypes.IncludeSubscription, *config.CostTypes.IncludeSubscription)
	}

	if *costTypes.IncludeSupport != *config.CostTypes.IncludeSupport {
		return fmt.Errorf("IncludeSupport not set properly '%v' should be '%v'", *costTypes.IncludeSupport, *config.CostTypes.IncludeSupport)
	}

	if *costTypes.IncludeTax != *config.CostTypes.IncludeTax {
		return fmt.Errorf("IncludeTax not set properly '%v' should be '%v'", *costTypes.IncludeTax, *config.CostTypes.IncludeTax)
	}

	if *costTypes.IncludeUpfront != *config.CostTypes.IncludeUpfront {
		return fmt.Errorf("IncludeUpfront not set properly '%v' should be '%v'", *costTypes.IncludeUpfront, *config.CostTypes.IncludeUpfront)
	}

	if *costTypes.UseBlended != *config.CostTypes.UseBlended {
		return fmt.Errorf("UseBlended not set properly '%v' should be '%v'", *costTypes.UseBlended, *config.CostTypes.UseBlended)
	}

	return nil
}

func testAccAWSBudgetsBudgetDestroy(s *terraform.State) error {
	meta := testAccProvider.Meta()
	client := meta.(*AWSClient).budgetconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_budgets_budget" {
			continue
		}

		accountID, budgetName, err := decodeBudgetsBudgetID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Budget '%s': id could not be decoded and could not be deleted properly", rs.Primary.ID)
		}

		_, err = client.DescribeBudget(&budgets.DescribeBudgetInput{
			BudgetName: aws.String(budgetName),
			AccountId:  aws.String(accountID),
		})
		if !isAWSErr(err, budgets.ErrCodeNotFoundException, "") {
			return fmt.Errorf("Budget '%s' was not deleted properly", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAWSBudgetsBudgetConfigUpdate(name string) budgets.Budget {
	dateNow := time.Now().UTC()
	futureDate := dateNow.AddDate(0, 0, 14)
	startDate := dateNow.AddDate(0, 0, -14)
	return budgets.Budget{
		BudgetName: aws.String(name),
		BudgetType: aws.String("COST"),
		BudgetLimit: &budgets.Spend{
			Amount: aws.String("500"),
			Unit:   aws.String("USD"),
		},
		CostFilters: map[string][]*string{
			"AZ": {
				aws.String("us-east-2"),
			},
		},
		CostTypes: &budgets.CostTypes{
			IncludeCredit:            aws.Bool(true),
			IncludeOtherSubscription: aws.Bool(true),
			IncludeRecurring:         aws.Bool(true),
			IncludeRefund:            aws.Bool(true),
			IncludeSubscription:      aws.Bool(false),
			IncludeSupport:           aws.Bool(true),
			IncludeTax:               aws.Bool(false),
			IncludeUpfront:           aws.Bool(true),
			UseBlended:               aws.Bool(false),
		},
		TimeUnit: aws.String("MONTHLY"),
		TimePeriod: &budgets.TimePeriod{
			End:   &futureDate,
			Start: &startDate,
		},
	}
}

func testAccAWSBudgetsBudgetConfigDefaults(name string) budgets.Budget {
	dateNow := time.Now().UTC()
	futureDate := time.Date(2087, 6, 15, 00, 0, 0, 0, time.UTC)
	startDate := dateNow.AddDate(0, 0, -14)
	return budgets.Budget{
		BudgetName: aws.String(name),
		BudgetType: aws.String("COST"),
		BudgetLimit: &budgets.Spend{
			Amount: aws.String("100"),
			Unit:   aws.String("USD"),
		},
		CostFilters: map[string][]*string{
			"AZ": {
				aws.String("us-east-1"),
			},
		},
		CostTypes: &budgets.CostTypes{
			IncludeCredit:            aws.Bool(true),
			IncludeOtherSubscription: aws.Bool(true),
			IncludeRecurring:         aws.Bool(true),
			IncludeRefund:            aws.Bool(true),
			IncludeSubscription:      aws.Bool(true),
			IncludeSupport:           aws.Bool(true),
			IncludeTax:               aws.Bool(true),
			IncludeUpfront:           aws.Bool(true),
			UseBlended:               aws.Bool(false),
		},
		TimeUnit: aws.String("MONTHLY"),
		TimePeriod: &budgets.TimePeriod{
			End:   &futureDate,
			Start: &startDate,
		},
	}
}

func testAccAWSBudgetsBudgetConfig_WithAccountID(budgetConfig budgets.Budget, accountID, costFilterKey string) string {
	t := template.Must(template.New("t1").
		Parse(`
resource "aws_budgets_budget" "foo" {
	account_id = "` + accountID + `"
	name_prefix = "{{.BudgetName}}"
	budget_type = "{{.BudgetType}}"
 	limit_amount = "{{.BudgetLimit.Amount}}"
 	limit_unit = "{{.BudgetLimit.Unit}}"
	time_period_start = "{{.TimePeriod.Start.Format "2006-01-02_15:04"}}" 
 	time_unit = "{{.TimeUnit}}"
	cost_filters = {
		"` + costFilterKey + `" = "` + *budgetConfig.CostFilters[costFilterKey][0] + `"
	}
}
`))
	var doc bytes.Buffer
	t.Execute(&doc, budgetConfig)
	return doc.String()
}

func testAccAWSBudgetsBudgetConfig_PrefixDefaults(budgetConfig budgets.Budget, costFilterKey string) string {
	t := template.Must(template.New("t1").
		Parse(`
resource "aws_budgets_budget" "foo" {
	name_prefix = "{{.BudgetName}}"
	budget_type = "{{.BudgetType}}"
 	limit_amount = "{{.BudgetLimit.Amount}}"
 	limit_unit = "{{.BudgetLimit.Unit}}"
	time_period_start = "{{.TimePeriod.Start.Format "2006-01-02_15:04"}}" 
 	time_unit = "{{.TimeUnit}}"
	cost_filters = {
		"` + costFilterKey + `" = "` + *budgetConfig.CostFilters[costFilterKey][0] + `"
	}
}
`))
	var doc bytes.Buffer
	t.Execute(&doc, budgetConfig)
	return doc.String()
}

func testAccAWSBudgetsBudgetConfig_Prefix(budgetConfig budgets.Budget, costFilterKey string) string {
	t := template.Must(template.New("t1").
		Parse(`
resource "aws_budgets_budget" "foo" {
	name_prefix = "{{.BudgetName}}"
	budget_type = "{{.BudgetType}}"
 	limit_amount = "{{.BudgetLimit.Amount}}"
 	limit_unit = "{{.BudgetLimit.Unit}}"
	cost_types {
		include_tax = "{{.CostTypes.IncludeTax}}"
		include_subscription = "{{.CostTypes.IncludeSubscription}}"
		use_blended = "{{.CostTypes.UseBlended}}"
	}
	time_period_start = "{{.TimePeriod.Start.Format "2006-01-02_15:04"}}" 
	time_period_end = "{{.TimePeriod.End.Format "2006-01-02_15:04"}}"
 	time_unit = "{{.TimeUnit}}"
	cost_filters = {
		"` + costFilterKey + `" = "` + *budgetConfig.CostFilters[costFilterKey][0] + `"
	}
}
`))
	var doc bytes.Buffer
	t.Execute(&doc, budgetConfig)
	return doc.String()
}

func testAccAWSBudgetsBudgetConfig_BasicDefaults(budgetConfig budgets.Budget, costFilterKey string) string {
	t := template.Must(template.New("t1").
		Parse(`
resource "aws_budgets_budget" "foo" {
	name = "{{.BudgetName}}"
	budget_type = "{{.BudgetType}}"
 	limit_amount = "{{.BudgetLimit.Amount}}"
 	limit_unit = "{{.BudgetLimit.Unit}}"
	time_period_start = "{{.TimePeriod.Start.Format "2006-01-02_15:04"}}" 
 	time_unit = "{{.TimeUnit}}"
	cost_filters = {
		"` + costFilterKey + `" = "` + *budgetConfig.CostFilters[costFilterKey][0] + `"
	}
}
`))
	var doc bytes.Buffer
	t.Execute(&doc, budgetConfig)
	return doc.String()
}

func testAccAWSBudgetsBudgetConfig_Basic(budgetConfig budgets.Budget, costFilterKey string) string {
	t := template.Must(template.New("t1").
		Parse(`
resource "aws_budgets_budget" "foo" {
	name = "{{.BudgetName}}"
	budget_type = "{{.BudgetType}}"
 	limit_amount = "{{.BudgetLimit.Amount}}"
 	limit_unit = "{{.BudgetLimit.Unit}}"
	cost_types {
		include_tax = "{{.CostTypes.IncludeTax}}"
		include_subscription = "{{.CostTypes.IncludeSubscription}}"
		use_blended = "{{.CostTypes.UseBlended}}"
	}
	time_period_start = "{{.TimePeriod.Start.Format "2006-01-02_15:04"}}" 
	time_period_end = "{{.TimePeriod.End.Format "2006-01-02_15:04"}}"
 	time_unit = "{{.TimeUnit}}"
	cost_filters = {
		"` + costFilterKey + `" = "` + *budgetConfig.CostFilters[costFilterKey][0] + `"
	}
}
`))
	var doc bytes.Buffer
	t.Execute(&doc, budgetConfig)
	return doc.String()
}
