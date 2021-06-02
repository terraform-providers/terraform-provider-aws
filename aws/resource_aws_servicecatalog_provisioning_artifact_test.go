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
)

// add sweeper to delete known test servicecat provisioning artifacts
func init() {
	resource.AddTestSweepers("aws_servicecatalog_provisioning_artifact", &resource.Sweeper{
		Name:         "aws_servicecatalog_provisioning_artifact",
		Dependencies: []string{},
		F:            testSweepServiceCatalogProvisioningArtifacts,
	})
}

func testSweepServiceCatalogProvisioningArtifacts(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).scconn
	sweepResources := make([]*testSweepResource, 0)
	var errs *multierror.Error

	input := &servicecatalog.SearchProductsAsAdminInput{}

	err = conn.SearchProductsAsAdminPages(input, func(page *servicecatalog.SearchProductsAsAdminOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, pvd := range page.ProductViewDetails {
			if pvd == nil || pvd.ProductViewSummary == nil || pvd.ProductViewSummary.ProductId == nil {
				continue
			}

			productID := aws.StringValue(pvd.ProductViewSummary.ProductId)

			artInput := &servicecatalog.ListProvisioningArtifactsInput{
				ProductId: aws.String(productID),
			}

			// there's no paginator for ListProvisioningArtifacts
			for {
				output, err := conn.ListProvisioningArtifacts(artInput)

				if err != nil {
					err := fmt.Errorf("error listing Service Catalog Provisioning Artifacts for product (%s): %w", productID, err)
					log.Printf("[ERROR] %s", err)
					errs = multierror.Append(errs, err)
					break
				}

				for _, pad := range output.ProvisioningArtifactDetails {
					r := resourceAwsServiceCatalogProvisioningArtifact()
					d := r.Data(nil)

					d.SetId(aws.StringValue(pad.Id))
					d.Set("product_id", productID)

					sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
				}

				/*
					// Currently, the API has no token field on input (AWS oops)
					if aws.StringValue(output.NextPageToken) == "" {
						break
					}

					artInput.NextPageToken = output.NextPageToken
				*/
				break
			}
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing Service Catalog Provisioning Artifacts for %s: %w", region, err))
	}

	if err = testSweepResourceOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Service Catalog Provisioning Artifacts for %s: %w", region, err))
	}

	if testSweepSkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Service Catalog Provisioning Artifacts sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAWSServiceCatalogProvisioningArtifact_basic(t *testing.T) {
	resourceName := "aws_servicecatalog_provisioning_artifact.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogProvisioningArtifactDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProvisioningArtifactConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProvisioningArtifactExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "accept_language", "en"),
					resource.TestCheckResourceAttr(resourceName, "active", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttr(resourceName, "disable_template_validation", "true"),
					resource.TestCheckResourceAttr(resourceName, "guidance", servicecatalog.ProvisioningArtifactGuidanceDefault),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("%s-2", rName)),
					resource.TestCheckResourceAttrPair(resourceName, "product_id", "aws_servicecatalog_product.test", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "template_url"),
					resource.TestCheckResourceAttr(resourceName, "type", servicecatalog.ProductTypeCloudFormationTemplate),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"accept_language",
					"disable_template_validation",
					"template_url",
				},
			},
		},
	})
}

func TestAccAWSServiceCatalogProvisioningArtifact_disappears(t *testing.T) {
	resourceName := "aws_servicecatalog_provisioning_artifact.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogProvisioningArtifactDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProvisioningArtifactConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProvisioningArtifactExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsServiceCatalogProvisioningArtifact(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSServiceCatalogProvisioningArtifact_update(t *testing.T) {
	resourceName := "aws_servicecatalog_provisioning_artifact.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogProvisioningArtifactDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProvisioningArtifactConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProvisioningArtifactExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "accept_language", "en"),
					resource.TestCheckResourceAttr(resourceName, "active", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttr(resourceName, "guidance", servicecatalog.ProvisioningArtifactGuidanceDefault),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("%s-2", rName)),
				),
			},
			{
				Config: testAccAWSServiceCatalogProvisioningArtifactConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "accept_language", "jp"),
					resource.TestCheckResourceAttr(resourceName, "active", "false"),
					resource.TestCheckResourceAttr(resourceName, "description", fmt.Sprintf("%s-3", rName)),
					resource.TestCheckResourceAttr(resourceName, "guidance", servicecatalog.ProvisioningArtifactGuidanceDeprecated),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("%s-3", rName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"accept_language",
					"disable_template_validation",
					"template_url",
				},
			},
		},
	})
}

func TestAccAWSServiceCatalogProvisioningArtifact_physicalID(t *testing.T) {
	resourceName := "aws_servicecatalog_provisioning_artifact.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogProvisioningArtifactDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProvisioningArtifactConfig_physicalID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProvisioningArtifactExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "accept_language", "en"),
					resource.TestCheckResourceAttr(resourceName, "active", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttr(resourceName, "disable_template_validation", "false"),
					resource.TestCheckResourceAttr(resourceName, "guidance", servicecatalog.ProvisioningArtifactGuidanceDefault),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("%s-2", rName)),
					resource.TestCheckResourceAttrPair(resourceName, "product_id", "aws_servicecatalog_product.test", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "template_physical_id"),
					resource.TestCheckResourceAttr(resourceName, "type", servicecatalog.ProductTypeCloudFormationTemplate),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"accept_language",
					"disable_template_validation",
					"template_physical_id",
				},
			},
		},
	})
}

func testAccCheckAwsServiceCatalogProvisioningArtifactDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).scconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_provisioning_artifact" {
			continue
		}

		artifactID, productID, err := tfservicecatalog.ProvisioningArtifactParseID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error parsing Service Catalog Provisioning Artifact ID (%s): %w", rs.Primary.ID, err)
		}

		input := &servicecatalog.DescribeProvisioningArtifactInput{
			ProductId:              aws.String(productID),
			ProvisioningArtifactId: aws.String(artifactID),
		}

		output, err := conn.DescribeProvisioningArtifact(input)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting Service Catalog Provisioning Artifact (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("Service Catalog Provisioning Artifact (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsServiceCatalogProvisioningArtifactExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).scconn

		artifactID, productID, err := tfservicecatalog.ProvisioningArtifactParseID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error parsing Service Catalog Provisioning Artifact ID (%s): %w", rs.Primary.ID, err)
		}

		input := &servicecatalog.DescribeProvisioningArtifactInput{
			ProductId:              aws.String(productID),
			ProvisioningArtifactId: aws.String(artifactID),
		}

		_, err = conn.DescribeProvisioningArtifact(input)

		if err != nil {
			return fmt.Errorf("error describing Service Catalog Provisioning Artifact (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccAWSServiceCatalogProvisioningArtifactConfigTemplateURLBase(rName string) string {
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
`, rName)
}

func testAccAWSServiceCatalogProvisioningArtifactConfig_basic(rName string) string {
	return composeConfig(testAccAWSServiceCatalogProvisioningArtifactConfigTemplateURLBase(rName), fmt.Sprintf(`
resource "aws_servicecatalog_provisioning_artifact" "test" {
  accept_language             = "en"
  active                      = true
  description                 = %[1]q
  disable_template_validation = true
  guidance                    = "DEFAULT"
  name                        = "%[1]s-2"
  product_id                  = aws_servicecatalog_product.test.id
  template_url                = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/${aws_s3_bucket_object.test.key}"
  type                        = "CLOUD_FORMATION_TEMPLATE"
}
`, rName))
}

func testAccAWSServiceCatalogProvisioningArtifactConfig_update(rName string) string {
	return composeConfig(testAccAWSServiceCatalogProvisioningArtifactConfigTemplateURLBase(rName), fmt.Sprintf(`
resource "aws_servicecatalog_provisioning_artifact" "test" {
  accept_language             = "jp"
  active                      = false
  description                 = "%[1]s-3"
  disable_template_validation = true
  guidance                    = "DEPRECATED"
  name                        = "%[1]s-3"
  product_id                  = aws_servicecatalog_product.test.id
  template_url                = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/${aws_s3_bucket_object.test.key}"
  type                        = "CLOUD_FORMATION_TEMPLATE"
}
`, rName))
}

func testAccAWSServiceCatalogProvisioningArtifactConfigPhysicalIDBase(rName string) string {
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
  description         = %[1]q
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
}
`, rName)
}

func testAccAWSServiceCatalogProvisioningArtifactConfig_physicalID(rName string) string {
	return composeConfig(testAccAWSServiceCatalogProvisioningArtifactConfigPhysicalIDBase(rName), fmt.Sprintf(`
resource "aws_servicecatalog_provisioning_artifact" "test" {
  accept_language             = "en"
  active                      = true
  description                 = %[1]q
  disable_template_validation = false
  guidance                    = "DEFAULT"
  name                        = "%[1]s-2"
  product_id                  = aws_servicecatalog_product.test.id
  template_physical_id        = aws_cloudformation_stack.test.id
  type                        = "CLOUD_FORMATION_TEMPLATE"
}
`, rName))
}
