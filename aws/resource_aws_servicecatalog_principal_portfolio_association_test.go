package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfservicecatalog "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicecatalog"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicecatalog/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

// add sweeper to delete known test servicecat principal portfolio associations
func init() {
	resource.AddTestSweepers("aws_servicecatalog_principal_portfolio_association", &resource.Sweeper{
		Name:         "aws_servicecatalog_principal_portfolio_association",
		Dependencies: []string{},
		F:            testSweepServiceCatalogPrincipalPortfolioAssociations,
	})
}

func testSweepServiceCatalogPrincipalPortfolioAssociations(region string) error {
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

		for _, detail := range page.PortfolioDetails {
			if detail == nil {
				continue
			}

			pInput := &servicecatalog.ListPrincipalsForPortfolioInput{
				PortfolioId: detail.Id,
			}

			err = conn.ListPrincipalsForPortfolioPages(pInput, func(page *servicecatalog.ListPrincipalsForPortfolioOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, principal := range page.Principals {
					if principal == nil {
						continue
					}

					r := resourceAwsServiceCatalogPrincipalPortfolioAssociation()
					d := r.Data(nil)
					d.SetId(tfservicecatalog.PrincipalPortfolioAssociationID(tfservicecatalog.AcceptLanguageEnglish, aws.StringValue(principal.PrincipalARN), aws.StringValue(detail.Id)))

					sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
				}

				return !lastPage
			})

			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("error listing Service Catalog Portfolios for Principals %s: %w", region, err))
				continue
			}
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing Service Catalog Principal Portfolio Associations for %s: %w", region, err))
	}

	if err = testSweepResourceOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Service Catalog Principal Portfolio Associations for %s: %w", region, err))
	}

	if testSweepSkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Service Catalog Principal Portfolio Associations sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAWSServiceCatalogPrincipalPortfolioAssociation_basic(t *testing.T) {
	resourceName := "aws_servicecatalog_principal_portfolio_association.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogPrincipalPortfolioAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogPrincipalPortfolioAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogPrincipalPortfolioAssociationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "portfolio_id", "aws_servicecatalog_portfolio.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "principal_arn", "aws_iam_role.test", "arn"),
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

func TestAccAWSServiceCatalogPrincipalPortfolioAssociation_disappears(t *testing.T) {
	resourceName := "aws_servicecatalog_principal_portfolio_association.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogPrincipalPortfolioAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogPrincipalPortfolioAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogPrincipalPortfolioAssociationExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsServiceCatalogPrincipalPortfolioAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsServiceCatalogPrincipalPortfolioAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).scconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_principal_portfolio_association" {
			continue
		}

		acceptLanguage, principalARN, portfolioID, err := tfservicecatalog.PrincipalPortfolioAssociationParseID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("could not parse ID (%s): %w", rs.Primary.ID, err)
		}

		err = waiter.PrincipalPortfolioAssociationDeleted(conn, acceptLanguage, principalARN, portfolioID)

		if tfresource.NotFound(err) || tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("waiting for Service Catalog Principal Portfolio Association to be destroyed (%s): %w", rs.Primary.ID, err)
		}
	}

	return nil
}

func testAccCheckAwsServiceCatalogPrincipalPortfolioAssociationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		acceptLanguage, principalARN, portfolioID, err := tfservicecatalog.PrincipalPortfolioAssociationParseID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("could not parse ID (%s): %w", rs.Primary.ID, err)
		}

		conn := testAccProvider.Meta().(*AWSClient).scconn

		_, err = waiter.PrincipalPortfolioAssociationReady(conn, acceptLanguage, principalARN, portfolioID)

		if err != nil {
			return fmt.Errorf("waiting for Service Catalog Principal Portfolio Association existence (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccAWSServiceCatalogPrincipalPortfolioAssociationConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "servicecatalog.${data.aws_partition.current.dns_suffix}"
      }
      Sid = ""
    }]
  })
}

resource "aws_servicecatalog_portfolio" "test" {
  name          = %[1]q
  provider_name = %[1]q
}
`, rName)
}

func testAccAWSServiceCatalogPrincipalPortfolioAssociationConfig_basic(rName string) string {
	return composeConfig(testAccAWSServiceCatalogPrincipalPortfolioAssociationConfig_base(rName), `
resource "aws_servicecatalog_principal_portfolio_association" "test" {
  portfolio_id  = aws_servicecatalog_portfolio.test.id
  principal_arn = aws_iam_role.test.arn
}
`)
}
