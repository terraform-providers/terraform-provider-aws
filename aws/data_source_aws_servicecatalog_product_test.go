package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSServiceCatalogProductDataSource_basic(t *testing.T) {
	resourceName := "aws_servicecatalog_product.test"
	dataSourceName := "data.aws_servicecatalog_product.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProductDataSourceConfig_basic(rName, "beskrivning", "supportbeskrivning"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "created_time", dataSourceName, "created_time"),
					resource.TestCheckResourceAttrPair(resourceName, "description", dataSourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "distributor", dataSourceName, "distributor"),
					resource.TestCheckResourceAttrPair(resourceName, "has_default_path", dataSourceName, "has_default_path"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "owner", dataSourceName, "owner"),
					resource.TestCheckResourceAttrPair(resourceName, "status", dataSourceName, "status"),
					resource.TestCheckResourceAttrPair(resourceName, "support_description", dataSourceName, "support_description"),
					resource.TestCheckResourceAttrPair(resourceName, "support_email", dataSourceName, "support_email"),
					resource.TestCheckResourceAttrPair(resourceName, "support_url", dataSourceName, "support_url"),
					resource.TestCheckResourceAttrPair(resourceName, "type", dataSourceName, "type"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.Name", dataSourceName, "tags.Name"),
				),
			},
		},
	})
}

func TestAccAWSServiceCatalogProductDataSource_physicalID(t *testing.T) {
	resourceName := "aws_servicecatalog_product.test"
	dataSourceName := "data.aws_servicecatalog_product.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProductDataSourceConfig_physicalID(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "created_time", dataSourceName, "created_time"),
					resource.TestCheckResourceAttrPair(resourceName, "description", dataSourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "distributor", dataSourceName, "distributor"),
					resource.TestCheckResourceAttrPair(resourceName, "has_default_path", dataSourceName, "has_default_path"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "owner", dataSourceName, "owner"),
					resource.TestCheckResourceAttrPair(resourceName, "status", dataSourceName, "status"),
					resource.TestCheckResourceAttrPair(resourceName, "support_description", dataSourceName, "support_description"),
					resource.TestCheckResourceAttrPair(resourceName, "support_email", dataSourceName, "support_email"),
					resource.TestCheckResourceAttrPair(resourceName, "support_url", dataSourceName, "support_url"),
					resource.TestCheckResourceAttrPair(resourceName, "type", dataSourceName, "type"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.Name", dataSourceName, "tags.Name"),
				),
			},
		},
	})
}

func testAccAWSServiceCatalogProductDataSourceConfig_basic(rName, description, supportDescription string) string {
	return composeConfig(testAccAWSServiceCatalogProductConfig_basic(rName, description, supportDescription), `
data "aws_servicecatalog_product" "test" {
  id = aws_servicecatalog_product.test.id
}
`)
}

func testAccAWSServiceCatalogProductDataSourceConfig_physicalID(rName string) string {
	return composeConfig(testAccAWSServiceCatalogProductConfig_physicalID(rName), `
data "aws_servicecatalog_product" "test" {
  id = aws_servicecatalog_product.test.id
}
`)
}
