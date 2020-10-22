package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAwsServiceQuotasServiceDataSource_ServiceName(t *testing.T) {
	dataSourceName := "data.aws_servicequotas_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsServiceQuotasServiceDataSourceConfigServiceName("Amazon Virtual Private Cloud (Amazon VPC)"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "service_code", "vpc"),
				),
			},
		},
	})
}

func testAccAwsServiceQuotasServiceDataSourceConfigServiceName(serviceName string) string {
	return fmt.Sprintf(`
data "aws_servicequotas_service" "test" {
  service_name = %[1]q
}
`, serviceName)
}
