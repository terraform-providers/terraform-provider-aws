package aws

import (
	"fmt"
	"log"
	"regexp"
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
)

// add sweeper to delete known test servicecat provisioned products
func init() {
	resource.AddTestSweepers("aws_servicecatalog_provisioned_product", &resource.Sweeper{
		Name:         "aws_servicecatalog_provisioned_product",
		Dependencies: []string{},
		F:            testSweepServiceCatalogProvisionedProducts,
	})
}

func testSweepServiceCatalogProvisionedProducts(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).scconn
	sweepResources := make([]*testSweepResource, 0)
	var errs *multierror.Error

	input := &servicecatalog.SearchProvisionedProductsInput{
		AccessLevelFilter: &servicecatalog.AccessLevelFilter{
			Key:   aws.String(servicecatalog.AccessLevelFilterKeyAccount),
			Value: aws.String(client.(*AWSClient).accountid),
		},
	}

	err = conn.SearchProvisionedProductsPages(input, func(page *servicecatalog.SearchProvisionedProductsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, detail := range page.ProvisionedProducts {
			if detail == nil {
				continue
			}

			r := resourceAwsServiceCatalogProvisionedProduct()
			d := r.Data(nil)
			d.SetId(aws.StringValue(detail.Id))

			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing Service Catalog Provisioned Products for %s: %w", region, err))
	}

	if err = testSweepResourceOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Service Catalog Provisioned Products for %s: %w", region, err))
	}

	if testSweepSkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Service Catalog Provisioned Products sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAWSServiceCatalogProvisionedProduct_basic(t *testing.T) {
	resourceName := "aws_servicecatalog_provisioned_product.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogProvisionedProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProvisionedProductConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProvisionedProductExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "accept_language", tfservicecatalog.AcceptLanguageEnglish),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", servicecatalog.ServiceName, regexp.MustCompile(fmt.Sprintf(`stack/%s/pp-.*`, rName))),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
					resource.TestCheckResourceAttrSet(resourceName, "last_provisioning_record_id"),
					resource.TestCheckResourceAttrSet(resourceName, "last_record_id"),
					resource.TestCheckResourceAttrSet(resourceName, "last_successful_provisioning_record_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "path_id", "data.aws_servicecatalog_launch_paths.test", "summaries.0.path_id"),
					resource.TestCheckResourceAttrPair(resourceName, "product_id", "aws_servicecatalog_product.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "provisioning_artifact_name", "aws_servicecatalog_product.test", "provisioning_artifact_parameters.0.name"),
					resource.TestCheckResourceAttr(resourceName, "status", servicecatalog.StatusAvailable),
					resource.TestCheckResourceAttr(resourceName, "type", "CFN_STACK"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"accept_language",
					"ignore_errors",
					"provisioning_artifact_name",
					"retain_physical_resources",
				},
			},
		},
	})
}

func TestAccAWSServiceCatalogProvisionedProduct_disappears(t *testing.T) {
	resourceName := "aws_servicecatalog_provisioned_product.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogProvisionedProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProvisionedProductConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProvisionedProductExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsServiceCatalogProvisionedProduct(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSServiceCatalogProvisionedProduct_tags(t *testing.T) {
	resourceName := "aws_servicecatalog_provisioned_product.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogProvisionedProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProvisionedProductConfig_tags(rName, "Name", rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProvisionedProductExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccAWSServiceCatalogProvisionedProductConfig_tags(rName, "NotName", rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProvisionedProductExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.NotName", rName),
				),
			},
		},
	})
}

func testAccCheckAwsServiceCatalogProvisionedProductDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).scconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_provisioned_product" {
			continue
		}

		err := waiter.ProvisionedProductTerminated(conn, tfservicecatalog.AcceptLanguageEnglish, rs.Primary.ID, "")

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting Service Catalog Provisioned Product (%s): %w", rs.Primary.ID, err)
		}
	}

	return nil
}

func testAccCheckAwsServiceCatalogProvisionedProductExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).scconn

		_, err := waiter.ProvisionedProductReady(conn, tfservicecatalog.AcceptLanguageEnglish, rs.Primary.ID, "")

		if err != nil {
			return fmt.Errorf("error describing Service Catalog Provisioned Product (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccAWSServiceCatalogProvisionedProductConfigTemplateURLBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  acl           = "private"
  force_destroy = true
}

resource "aws_s3_bucket_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = "%[1]s.json"

  content = jsonencode({
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
  description         = %[1]q
  distributor         = "distributör"
  name                = %[1]q
  owner               = "ägare"
  type                = "CLOUD_FORMATION_TEMPLATE"
  support_description = %[1]q
  support_email       = "support@example.com"
  support_url         = "http://example.com"

  provisioning_artifact_parameters {
    description                 = "artefaktbeskrivning"
    disable_template_validation = true
    name                        = %[1]q
    template_url                = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/${aws_s3_bucket_object.test.key}"
    type                        = "CLOUD_FORMATION_TEMPLATE"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_servicecatalog_portfolio" "test" {
  name          = %[1]q
  description   = %[1]q
  provider_name = %[1]q
}

resource "aws_servicecatalog_constraint" "test" {
  description  = %[1]q
  portfolio_id = aws_servicecatalog_product_portfolio_association.test.portfolio_id
  product_id   = aws_servicecatalog_product_portfolio_association.test.product_id
  type         = "RESOURCE_UPDATE"

  parameters = jsonencode({
    Version = "2.0"
    Properties = {
      TagUpdateOnProvisionedProduct = "ALLOWED"
    }
  })
}

resource "aws_servicecatalog_product_portfolio_association" "test" {
  portfolio_id = aws_servicecatalog_principal_portfolio_association.test.portfolio_id # avoid depends_on
  product_id   = aws_servicecatalog_product.test.id
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_servicecatalog_principal_portfolio_association" "test" {
  portfolio_id  = aws_servicecatalog_portfolio.test.id
  principal_arn = data.aws_iam_session_context.current.issuer_arn # unfortunately, you cannot get launch_path for arbitrary role - only caller
}

data "aws_servicecatalog_launch_paths" "test" {
  product_id = aws_servicecatalog_product_portfolio_association.test.product_id # avoid depends_on
}
`, rName)
}

func testAccAWSServiceCatalogProvisionedProductConfig_basic(rName string) string {
	return composeConfig(testAccAWSServiceCatalogProvisionedProductConfigTemplateURLBase(rName), fmt.Sprintf(`
resource "aws_servicecatalog_provisioned_product" "test" {
  name                       = %[1]q
  product_id                 = aws_servicecatalog_product.test.id
  provisioning_artifact_name = %[1]q
  path_id                    = data.aws_servicecatalog_launch_paths.test.summaries[0].path_id
}
`, rName))
}

func testAccAWSServiceCatalogProvisionedProductConfig_tags(rName, tagKey, tagValue string) string {
	return composeConfig(testAccAWSServiceCatalogProvisionedProductConfigTemplateURLBase(rName), fmt.Sprintf(`
resource "aws_servicecatalog_provisioned_product" "test" {
  name                       = %[1]q
  product_id                 = aws_servicecatalog_constraint.test.product_id
  provisioning_artifact_name = %[1]q
  path_id                    = data.aws_servicecatalog_launch_paths.test.summaries[0].path_id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey, tagValue))
}
