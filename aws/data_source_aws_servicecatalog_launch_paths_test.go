package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSServiceCatalogLaunchPathsDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_servicecatalog_launch_paths.test"
	resourceNameProduct := "aws_servicecatalog_product.test"
	resourceNamePortfolio := "aws_servicecatalog_portfolio.test"

	rName := acctest.RandomWithPrefix("tf-acc-test")
	domain := fmt.Sprintf("http://%s", testAccRandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogLaunchPathsDataSourceConfig_basic(rName, domain, testAccDefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "accept_language", "en"),
					resource.TestCheckResourceAttrPair(dataSourceName, "product_id", resourceNameProduct, "id"),
					resource.TestCheckResourceAttr(dataSourceName, "summaries.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "summaries.0.name", resourceNamePortfolio, "name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "summaries.0.path_id"),
				),
			},
		},
	})
}

func testAccAWSServiceCatalogLaunchPathsDataSourceConfig_base(rName, domain, email string) string {
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
  support_email       = %[3]q
  support_url         = %[2]q

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

resource "aws_servicecatalog_product_portfolio_association" "test" {
  portfolio_id = aws_servicecatalog_portfolio.test.id
  product_id   = aws_servicecatalog_product.test.id
}

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

data "aws_caller_identity" "current" {}

resource "aws_servicecatalog_principal_portfolio_association" "test" {
  portfolio_id  = aws_servicecatalog_portfolio.test.id
  principal_arn = data.aws_caller_identity.current.arn
}
`, rName, domain, email)
}

func testAccAWSServiceCatalogLaunchPathsDataSourceConfig_basic(rName, domain, email string) string {
	return composeConfig(testAccAWSServiceCatalogLaunchPathsDataSourceConfig_base(rName, domain, email), `
data "aws_servicecatalog_launch_paths" "test" {
  product_id = aws_servicecatalog_product_portfolio_association.test.product_id

  depends_on = [aws_servicecatalog_principal_portfolio_association.test]
}
`)
}
