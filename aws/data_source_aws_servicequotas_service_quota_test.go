package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/servicequotas"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAwsServiceQuotasServiceQuotaDataSource_QuotaCode(t *testing.T) {
	dataSourceName := "data.aws_servicequotas_service_quota.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(servicequotas.EndpointsID, t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsServiceQuotasServiceQuotaDataSourceConfigQuotaCode("vpc", "L-F678F1CE"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "adjustable", "true"),
					testAccCheckResourceAttrRegionalARN(dataSourceName, "arn", "servicequotas", "vpc/L-F678F1CE"),
					resource.TestCheckResourceAttr(dataSourceName, "default_value", "5"),
					resource.TestCheckResourceAttr(dataSourceName, "global_quota", "false"),
					resource.TestCheckResourceAttr(dataSourceName, "quota_code", "L-F678F1CE"),
					resource.TestCheckResourceAttr(dataSourceName, "quota_name", "VPCs per Region"),
					resource.TestCheckResourceAttr(dataSourceName, "service_code", "vpc"),
					resource.TestCheckResourceAttr(dataSourceName, "service_name", "Amazon Virtual Private Cloud (Amazon VPC)"),
					resource.TestMatchResourceAttr(dataSourceName, "value", regexp.MustCompile(`^\d+$`)),
				),
			},
		},
	})
}

func TestAccAwsServiceQuotasServiceQuotaDataSource_QuotaName(t *testing.T) {
	dataSourceName := "data.aws_servicequotas_service_quota.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(servicequotas.EndpointsID, t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsServiceQuotasServiceQuotaDataSourceConfigQuotaName("vpc", "VPCs per Region"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "adjustable", "true"),
					testAccCheckResourceAttrRegionalARN(dataSourceName, "arn", "servicequotas", "vpc/L-F678F1CE"),
					resource.TestCheckResourceAttr(dataSourceName, "default_value", "5"),
					resource.TestCheckResourceAttr(dataSourceName, "global_quota", "false"),
					resource.TestCheckResourceAttr(dataSourceName, "quota_code", "L-F678F1CE"),
					resource.TestCheckResourceAttr(dataSourceName, "quota_name", "VPCs per Region"),
					resource.TestCheckResourceAttr(dataSourceName, "service_code", "vpc"),
					resource.TestCheckResourceAttr(dataSourceName, "service_name", "Amazon Virtual Private Cloud (Amazon VPC)"),
					resource.TestMatchResourceAttr(dataSourceName, "value", regexp.MustCompile(`^\d+$`)),
				),
			},
		},
	})
}

func testAccAwsServiceQuotasServiceQuotaDataSourceConfigQuotaCode(serviceCode, quotaCode string) string {
	return fmt.Sprintf(`
data "aws_servicequotas_service_quota" "test" {
  quota_code   = %[1]q
  service_code = %[2]q
}
`, quotaCode, serviceCode)
}

func testAccAwsServiceQuotasServiceQuotaDataSourceConfigQuotaName(serviceCode, quotaName string) string {
	return fmt.Sprintf(`
data "aws_servicequotas_service_quota" "test" {
  quota_name   = %[1]q
  service_code = %[2]q
}
`, quotaName, serviceCode)
}
