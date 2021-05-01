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
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicecatalog/waiter"
)

// add sweeper to delete known test servicecat products
func init() {
	resource.AddTestSweepers("aws_servicecatalog_product", &resource.Sweeper{
		Name:         "aws_servicecatalog_product",
		Dependencies: []string{},
		F:            testSweepServiceCatalogProducts,
	})
}

func testSweepServiceCatalogProducts(region string) error {
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
			if pvd == nil || pvd.ProductViewSummary == nil {
				continue
			}

			id := aws.StringValue(pvd.ProductViewSummary.ProductId)

			r := resourceAwsServiceCatalogProduct()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing Service Catalog Products for %s: %w", region, err))
	}

	if err = testSweepResourceOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Service Catalog Products for %s: %w", region, err))
	}

	if testSweepSkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Service Catalog Products sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAWSServiceCatalogProduct_basic(t *testing.T) {
	resourceName := "aws_servicecatalog_product.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProductConfig_basic(rName, "beskrivning", "supportbeskrivning"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "catalog", regexp.MustCompile(`product/prod-.*`)),
					resource.TestCheckResourceAttr(resourceName, "accept_language", "en"),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "description", "beskrivning"),
					resource.TestCheckResourceAttr(resourceName, "distributor", "distributör"),
					resource.TestCheckResourceAttr(resourceName, "has_default_path", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "owner", "ägare"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.0.description", "artefaktbeskrivning"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.0.disable_template_validation", "true"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.0.name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "provisioning_artifact_parameters.0.template_url"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.0.type", servicecatalog.ProvisioningArtifactTypeCloudFormationTemplate),
					resource.TestCheckResourceAttr(resourceName, "status", waiter.ProductStatusCreated),
					resource.TestCheckResourceAttr(resourceName, "support_description", "supportbeskrivning"),
					resource.TestCheckResourceAttr(resourceName, "support_email", "support@example.com"),
					resource.TestCheckResourceAttr(resourceName, "support_url", "http://example.com"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "type", servicecatalog.ProductTypeCloudFormationTemplate),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"accept_language",
					"provisioning_artifact_parameters",
				},
			},
		},
	})
}

func TestAccAWSServiceCatalogProduct_disappears(t *testing.T) {
	resourceName := "aws_servicecatalog_product.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProductConfig_basic(rName, rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsServiceCatalogProduct(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSServiceCatalogProduct_update(t *testing.T) {
	resourceName := "aws_servicecatalog_product.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProductConfig_basic(rName, "beskrivning", "supportbeskrivning"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "beskrivning"),
					resource.TestCheckResourceAttr(resourceName, "support_description", "supportbeskrivning"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccAWSServiceCatalogProductConfig_basic(rName, "ny beskrivning", "ny supportbeskrivning"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "ny beskrivning"),
					resource.TestCheckResourceAttr(resourceName, "support_description", "ny supportbeskrivning"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccAWSServiceCatalogProduct_updateTags(t *testing.T) {
	resourceName := "aws_servicecatalog_product.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProductConfig_basic(rName, "beskrivning", "supportbeskrivning"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccAWSServiceCatalogProductConfig_updateTags(rName, "beskrivning", "supportbeskrivning"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Yak", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "natural"),
				),
			},
		},
	})
}

func TestAccAWSServiceCatalogProduct_physicalID(t *testing.T) {
	resourceName := "aws_servicecatalog_product.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProductConfig_physicalID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "catalog", regexp.MustCompile(`product/prod-.*`)),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.0.description", "artefaktbeskrivning"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.0.name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "provisioning_artifact_parameters.0.template_physical_id"),
					testAccMatchResourceAttrRegionalARN(
						resourceName,
						"provisioning_artifact_parameters.0.template_physical_id",
						"cloudformation",
						regexp.MustCompile(fmt.Sprintf(`stack/%s/.*`, rName)),
					),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.0.type", servicecatalog.ProvisioningArtifactTypeCloudFormationTemplate),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"accept_language",
					"provisioning_artifact_parameters",
				},
			},
		},
	})
}

func testAccCheckAwsServiceCatalogProductDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).scconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_product" {
			continue
		}

		input := &servicecatalog.DescribeProductAsAdminInput{
			Id: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeProductAsAdmin(input)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting Service Catalog Product (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("Service Catalog Product (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsServiceCatalogProductExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).scconn

		input := &servicecatalog.DescribeProductAsAdminInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeProductAsAdmin(input)

		if err != nil {
			return fmt.Errorf("error describing Service Catalog Product (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccAWSServiceCatalogProductConfigTemplateURLBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  acl           = "private"
  force_destroy = true
}

resource "aws_s3_bucket_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = "%[1]s.json"

  content = <<EOF
{
  "Resources" : {
    "MyVPC": {
      "Type" : "AWS::EC2::VPC",
      "Properties" : {
        "CidrBlock" : "10.0.0.0/16",
        "Tags" : [
          {"Key": "Name", "Value": "Primary_CF_VPC"}
        ]
      }
    }
  },
  "Outputs" : {
    "DefaultSgId" : {
      "Description": "The ID of default security group",
      "Value" : { "Fn::GetAtt" : [ "MyVPC", "DefaultSecurityGroup" ]}
    },
    "VpcID" : {
      "Description": "The VPC ID",
      "Value" : { "Ref" : "MyVPC" }
    }
  }
}
EOF
}
`, rName)
}

func testAccAWSServiceCatalogProductConfig_basic(rName, description, supportDescription string) string {
	return composeConfig(testAccAWSServiceCatalogProductConfigTemplateURLBase(rName), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_servicecatalog_product" "test" {
  description         = %[2]q
  distributor         = "distributör"
  name                = %[1]q
  owner               = "ägare"
  type                = "CLOUD_FORMATION_TEMPLATE"
  support_description = %[3]q
  support_email       = "support@example.com"
  support_url         = "http://example.com"

  provisioning_artifact_parameters {
    description                 = "artefaktbeskrivning"
    disable_template_validation = true
    name                        = %[1]q
    template_url                = "https://s3.${data.aws_partition.current.dns_suffix}/${aws_s3_bucket.test.id}/${aws_s3_bucket_object.test.key}"
    type                        = "CLOUD_FORMATION_TEMPLATE"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, description, supportDescription))
}

func testAccAWSServiceCatalogProductConfig_updateTags(rName, description, supportDescription string) string {
	return composeConfig(testAccAWSServiceCatalogProductConfigTemplateURLBase(rName), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_servicecatalog_product" "test" {
  description         = %[2]q
  distributor         = "distributör"
  name                = %[1]q
  owner               = "ägare"
  type                = "CLOUD_FORMATION_TEMPLATE"
  support_description = %[3]q
  support_email       = "support@example.com"
  support_url         = "http://example.com"

  provisioning_artifact_parameters {
    description                 = "artefaktbeskrivning"
    disable_template_validation = true
    name                        = %[1]q
    template_url                = "https://s3.${data.aws_partition.current.dns_suffix}/${aws_s3_bucket.test.id}/${aws_s3_bucket_object.test.key}"
    type                        = "CLOUD_FORMATION_TEMPLATE"
  }

  tags = {
    Yak         = %[1]q
    Environment = "natural"
  }
}
`, rName, description, supportDescription))
}

func testAccAWSServiceCatalogProductConfig_physicalID(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "test" {
  name = %[1]q

  template_body = <<STACK
{
  "Resources" : {
    "MyVPC": {
      "Type" : "AWS::EC2::VPC",
      "Properties" : {
        "CidrBlock" : "10.0.0.0/16",
        "Tags" : [
          {"Key": "Name", "Value": "Primary_CF_VPC"}
        ]
      }
    }
  },
  "Outputs" : {
    "DefaultSgId" : {
      "Description": "The ID of default security group",
      "Value" : { "Fn::GetAtt" : [ "MyVPC", "DefaultSecurityGroup" ]}
    },
    "VpcID" : {
      "Description": "The VPC ID",
      "Value" : { "Ref" : "MyVPC" }
    }
  }
}
STACK
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
`, rName)
}
