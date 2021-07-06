package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSServiceCatalogPortfolioDataSource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_servicecatalog_portfolio.test"
	resourceName := "aws_servicecatalog_portfolio.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatlaogPortfolioDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceCatalogPortfolioDataSourceConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "created_time", dataSourceName, "created_time"),
					resource.TestCheckResourceAttrPair(resourceName, "description", dataSourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "provider_name", dataSourceName, "provider_name"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.Chicane", dataSourceName, "tags.Chicane"),
				),
			},
		},
	})
}

func testAccCheckAwsServiceCatalogPortfolioDataSourceConfigBasic(rName string) string {
	return composeConfig(testAccCheckAwsServiceCatalogPortfolioResourceConfigTags1(rName, "Chicane", "Nick"), `
data "aws_servicecatalog_portfolio" "test" {
  id = aws_servicecatalog_portfolio.test.id
}
`)
}
