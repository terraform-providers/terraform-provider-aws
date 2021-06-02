package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfservicecatalog "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicecatalog"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicecatalog/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

// add sweeper to delete known test servicecat budget resource associations
func init() {
	resource.AddTestSweepers("aws_servicecatalog_budget_resource_association", &resource.Sweeper{
		Name:         "aws_servicecatalog_budget_resource_association",
		Dependencies: []string{},
		F:            testSweepServiceCatalogBudgetResourceAssociations,
	})
}

func testSweepServiceCatalogBudgetResourceAssociations(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).scconn
	sweepResources := make([]*testSweepResource, 0)
	var errs *multierror.Error

	input := &servicecatalog.ListPortfoliosInput{}

	err = conn.ListPortfoliosPages(input, func(page *servicecatalog.ListPortfoliosOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, port := range page.PortfolioDetails {
			if port == nil {
				continue
			}

			resInput := &servicecatalog.ListBudgetsForResourceInput{
				ResourceId: port.Id,
			}

			err = conn.ListBudgetsForResourcePages(resInput, func(page *servicecatalog.ListBudgetsForResourceOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, budget := range page.Budgets {
					if budget == nil {
						continue
					}

					r := resourceAwsServiceCatalogBudgetResourceAssociation()
					d := r.Data(nil)
					d.SetId(tfservicecatalog.BudgetResourceAssociationID(aws.StringValue(budget.BudgetName), aws.StringValue(port.Id)))

					sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
				}

				return !lastPage
			})
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing Service Catalog Budget Resource (Portfolio) Associations for %s: %w", region, err))
	}

	prodInput := &servicecatalog.SearchProductsAsAdminInput{}

	err = conn.SearchProductsAsAdminPages(prodInput, func(page *servicecatalog.SearchProductsAsAdminOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, pvd := range page.ProductViewDetails {
			if pvd == nil || pvd.ProductViewSummary == nil {
				continue
			}

			resInput := &servicecatalog.ListBudgetsForResourceInput{
				ResourceId: pvd.ProductViewSummary.ProductId,
			}

			err = conn.ListBudgetsForResourcePages(resInput, func(page *servicecatalog.ListBudgetsForResourceOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, budget := range page.Budgets {
					if budget == nil {
						continue
					}

					r := resourceAwsServiceCatalogBudgetResourceAssociation()
					d := r.Data(nil)
					d.SetId(tfservicecatalog.BudgetResourceAssociationID(aws.StringValue(budget.BudgetName), aws.StringValue(pvd.ProductViewSummary.ProductId)))

					sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
				}

				return !lastPage
			})
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing Service Catalog Budget Resource (Product) Associations for %s: %w", region, err))
	}

	if err = testSweepResourceOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Service Catalog Budget Resource Associations for %s: %w", region, err))
	}

	if testSweepSkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Service Catalog Budget Resource Associations sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAWSServiceCatalogBudgetResourceAssociation_basic(t *testing.T) {
	resourceName := "aws_servicecatalog_budget_resource_association.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck("budgets", t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID, "budgets"),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogBudgetResourceAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogBudgetResourceAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogBudgetResourceAssociationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "resource_id", "aws_servicecatalog_portfolio.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "budget_name", "aws_budgets_budget.test", "name"),
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

func TestAccAWSServiceCatalogBudgetResourceAssociation_disappears(t *testing.T) {
	resourceName := "aws_servicecatalog_budget_resource_association.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck("budgets", t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID, "budgets"),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogBudgetResourceAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogBudgetResourceAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogBudgetResourceAssociationExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsServiceCatalogBudgetResourceAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsServiceCatalogBudgetResourceAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).scconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_budget_resource_association" {
			continue
		}

		budgetName, resourceID, err := tfservicecatalog.BudgetResourceAssociationParseID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("could not parse ID (%s): %w", rs.Primary.ID, err)
		}

		err = waiter.BudgetResourceAssociationDeleted(conn, budgetName, resourceID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return fmt.Errorf("waiting for Service Catalog Budget Resource Association to be destroyed (%s): %w", rs.Primary.ID, err)
		}
	}

	return nil
}

func testAccCheckAwsServiceCatalogBudgetResourceAssociationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		budgetName, resourceID, err := tfservicecatalog.BudgetResourceAssociationParseID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("could not parse ID (%s): %w", rs.Primary.ID, err)
		}

		conn := testAccProvider.Meta().(*AWSClient).scconn

		_, err = waiter.BudgetResourceAssociationReady(conn, budgetName, resourceID)

		if err != nil {
			return fmt.Errorf("waiting for Service Catalog Budget Resource Association existence (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccAWSServiceCatalogBudgetResourceAssociationConfig_base(rName, budgetType, limitAmount, limitUnit, timePeriodStart, timeUnit string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalog_portfolio" "test" {
  name          = %[1]q
  description   = %[1]q
  provider_name = %[1]q
}

resource "aws_budgets_budget" "test" {
  name              = %[1]q
  budget_type       = %[2]q
  limit_amount      = %[3]q
  limit_unit        = %[4]q
  time_period_start = %[5]q
  time_unit         = %[6]q
}
`, rName, budgetType, limitAmount, limitUnit, timePeriodStart, timeUnit)
}

func testAccAWSServiceCatalogBudgetResourceAssociationConfig_basic(rName string) string {
	return composeConfig(testAccAWSServiceCatalogBudgetResourceAssociationConfig_base(rName, "COST", "100.0", "USD", "2017-01-01_12:00", "MONTHLY"), fmt.Sprintf(`
resource "aws_servicecatalog_budget_resource_association" "test" {
  resource_id = aws_servicecatalog_portfolio.test.id
  budget_name = %[1]q
}
`, rName))
}
