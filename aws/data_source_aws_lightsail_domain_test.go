package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSLightsailDomainDataSource_DomainName(t *testing.T) {
	var domain lightsail.Domain
	rName := acctest.RandomWithPrefix("tf-acc-test.com")
	resourceName := "aws_lightsail_domain.test"
	dataSourceName := "data.aws_lightsail_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckLightsailDomain(t) },
		CheckDestroy: testAccCheckAWSLightsailDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailDomainDataSourceConfigDomainName(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name", dataSourceName, "domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "tags", dataSourceName, "tags"),
				),
			},
		},
	})
}

func testAccAWSLightsailDomainDataSourceConfigDomainName(rName string) string {
	return composeConfig(
		testAccLightsailDomainRegionProviderConfig(),
		fmt.Sprintf(`
resource "aws_lightsail_domain" "test" {
  domain_name = "%s"
  tags = {
    "key1" = "value1"
    "key2" = "value2"
  }
}
data "aws_lightsail_domain" "test" {
  domain_name = aws_lightsail_domain.test.domain_name
}
`, rName))
}
