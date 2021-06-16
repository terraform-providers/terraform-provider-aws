package aws

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfservicecatalog "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicecatalog"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicecatalog/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

// add sweeper to delete known test servicecat product portfolio associations
func init() {
	resource.AddTestSweepers("aws_servicecatalog_product_portfolio_association", &resource.Sweeper{
		Name:         "aws_servicecatalog_product_portfolio_association",
		Dependencies: []string{},
		F:            testSweepServiceCatalogProductPortfolioAssociations,
	})
}

func testSweepServiceCatalogProductPortfolioAssociations(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).scconn
	sweepResources := make([]*testSweepResource, 0)
	var errs *multierror.Error

	// no paginator or list operation for associations directly, have to list products and associations of products

	input := &servicecatalog.SearchProductsAsAdminInput{}

	err = conn.SearchProductsAsAdminPages(input, func(page *servicecatalog.SearchProductsAsAdminOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, detail := range page.ProductViewDetails {
			if detail == nil {
				continue
			}

			productARN, err := arn.Parse(aws.StringValue(detail.ProductARN))

			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("error parsing Product ARN for %s: %w", aws.StringValue(detail.ProductARN), err))
				continue
			}

			// arn:aws:catalog:us-west-2:187416307283:product/prod-t5thhvquxw2x2

			resourceParts := strings.SplitN(productARN.Resource, "/", 2)

			if len(resourceParts) != 2 {
				errs = multierror.Append(errs, fmt.Errorf("error parsing Product ARN resource for %s: %w", aws.StringValue(detail.ProductARN), err))
				continue
			}

			productID := resourceParts[1]

			assocInput := &servicecatalog.ListPortfoliosForProductInput{
				ProductId: aws.String(productID),
			}

			err = conn.ListPortfoliosForProductPages(assocInput, func(page *servicecatalog.ListPortfoliosForProductOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, detail := range page.PortfolioDetails {
					if detail == nil {
						continue
					}

					r := resourceAwsServiceCatalogProductPortfolioAssociation()
					d := r.Data(nil)
					d.SetId(tfservicecatalog.ProductPortfolioAssociationCreateID(tfservicecatalog.AcceptLanguageEnglish, aws.StringValue(detail.Id), productID))

					sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
				}

				return !lastPage
			})

			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("error listing Service Catalog Portfolios for Products %s: %w", region, err))
				continue
			}
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing Service Catalog Product Portfolio Associations for %s: %w", region, err))
	}

	if err = testSweepResourceOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Service Catalog Product Portfolio Associations for %s: %w", region, err))
	}

	if testSweepSkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Service Catalog Product Portfolio Associations sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAWSServiceCatalogProductPortfolioAssociation_basic(t *testing.T) {
	resourceName := "aws_servicecatalog_product_portfolio_association.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogProductPortfolioAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProductPortfolioAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductPortfolioAssociationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "portfolio_id", "aws_servicecatalog_portfolio.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "product_id", "aws_servicecatalog_product.test", "id"),
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

func TestAccAWSServiceCatalogProductPortfolioAssociation_disappears(t *testing.T) {
	resourceName := "aws_servicecatalog_product_portfolio_association.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogProductPortfolioAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProductPortfolioAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductPortfolioAssociationExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsServiceCatalogProductPortfolioAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsServiceCatalogProductPortfolioAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).scconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_product_portfolio_association" {
			continue
		}

		acceptLanguage, portfolioID, productID, err := tfservicecatalog.ProductPortfolioAssociationParseID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("could not parse ID (%s): %w", rs.Primary.ID, err)
		}

		err = waiter.ProductPortfolioAssociationDeleted(conn, acceptLanguage, portfolioID, productID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return fmt.Errorf("waiting for Service Catalog Product Portfolio Association to be destroyed (%s): %w", rs.Primary.ID, err)
		}
	}

	return nil
}

func testAccCheckAwsServiceCatalogProductPortfolioAssociationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		acceptLanguage, portfolioID, productID, err := tfservicecatalog.ProductPortfolioAssociationParseID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("could not parse ID (%s): %w", rs.Primary.ID, err)
		}

		conn := testAccProvider.Meta().(*AWSClient).scconn

		_, err = waiter.ProductPortfolioAssociationReady(conn, acceptLanguage, portfolioID, productID)

		if err != nil {
			return fmt.Errorf("waiting for Service Catalog Product Portfolio Association existence (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccAWSServiceCatalogProductPortfolioAssociationConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "test" {
  name = %[1]q

  template_body = jsonencode({
    AWSTemplateFormatVersion = "2010-09-09"

    Resources = {
      MyVPC = {
        Type = "AWS::EC2::VPC"
        Properties = {
          CidrBlock = "10.1.0.0/16"
        }
      }
    }

    Outputs = {
      VpcID = {
        Description = "VPC ID"
        Value = {
          Ref = "MyVPC"
        }
      }
    }
  })
}

resource "aws_servicecatalog_product" "test" {
  description         = "beskrivning"
  distributor         = "distributör"
  name                = %[1]q
  owner               = "ägare"
  type                = "CLOUD_FORMATION_TEMPLATE"
  support_description = "supportbeskrivning"
  support_email       = "support@example.com"
  support_url         = "http://example.com"

  provisioning_artifact_parameters {
    description          = "artefaktbeskrivning"
    name                 = %[1]q
    template_physical_id = aws_cloudformation_stack.test.id
    type                 = "CLOUD_FORMATION_TEMPLATE"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_servicecatalog_portfolio" "test" {
  name          = %[1]q
  provider_name = %[1]q
}
`, rName)
}

func testAccAWSServiceCatalogProductPortfolioAssociationConfig_basic(rName string) string {
	return composeConfig(testAccAWSServiceCatalogProductPortfolioAssociationConfig_base(rName), `
resource "aws_servicecatalog_product_portfolio_association" "test" {
  portfolio_id = aws_servicecatalog_portfolio.test.id
  product_id   = aws_servicecatalog_product.test.id
}
`)
}
