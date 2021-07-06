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
		PreCheck:   func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(servicequotas.EndpointsID, t) },
		ErrorCheck: testAccErrorCheck(t, servicequotas.EndpointsID),
		Providers:  testAccProviders,
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

func TestAccAwsServiceQuotasServiceQuotaDataSource_PermissionError_QuotaCode(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSServiceQuotas(t); testAccAssumeRoleARNPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicequotas.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config:      testAccAwsServiceQuotasServiceQuotaDataSourceConfig_PermissionError_QuotaCode("elasticloadbalancing", "L-53DA6B97"),
				ExpectError: regexp.MustCompile(`DEPENDENCY_ACCESS_DENIED_ERROR`),
			},
		},
	})
}

func TestAccAwsServiceQuotasServiceQuotaDataSource_QuotaName(t *testing.T) {
	dataSourceName := "data.aws_servicequotas_service_quota.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(servicequotas.EndpointsID, t) },
		ErrorCheck: testAccErrorCheck(t, servicequotas.EndpointsID),
		Providers:  testAccProviders,
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

func TestAccAwsServiceQuotasServiceQuotaDataSource_PermissionError_QuotaName(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSServiceQuotas(t); testAccAssumeRoleARNPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicequotas.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config:      testAccAwsServiceQuotasServiceQuotaDataSourceConfig_PermissionError_QuotaName("elasticloadbalancing", "Application Load Balancers per Region"),
				ExpectError: regexp.MustCompile(`DEPENDENCY_ACCESS_DENIED_ERROR`),
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

func testAccAwsServiceQuotasServiceQuotaDataSourceConfig_PermissionError_QuotaCode(serviceCode, quotaCode string) string {
	policy := `{
  "Version": "2012-10-17",
  "Statement": [
    {
  	  "Effect": "Allow",
  	  "Action": [
  	    "servicequotas:GetServiceQuota"
  	  ],
  	  "Resource": "*"
    },
    {
  	  "Effect": "Deny",
  	  "Action": [
  	    "elasticloadbalancing:*"
  	  ],
  	  "Resource": "*"
    }
  ]
}`

	return composeConfig(
		testAccProviderConfigAssumeRolePolicy(policy),
		fmt.Sprintf(`
data "aws_servicequotas_service_quota" "test" {
  service_code = %[1]q
  quota_code   = %[2]q
}
`, serviceCode, quotaCode))
}

func testAccAwsServiceQuotasServiceQuotaDataSourceConfigQuotaName(serviceCode, quotaName string) string {
	return fmt.Sprintf(`
data "aws_servicequotas_service_quota" "test" {
  quota_name   = %[1]q
  service_code = %[2]q
}
`, quotaName, serviceCode)
}

func testAccAwsServiceQuotasServiceQuotaDataSourceConfig_PermissionError_QuotaName(serviceCode, quotaName string) string {
	policy := `{
  "Version": "2012-10-17",
  "Statement": [
    {
  	  "Effect": "Allow",
  	  "Action": [
  	    "servicequotas:ListServiceQuotas"
  	  ],
  	  "Resource": "*"
    },
    {
  	  "Effect": "Deny",
  	  "Action": [
  	    "elasticloadbalancing:*"
  	  ],
  	  "Resource": "*"
    }
  ]
}`

	return composeConfig(
		testAccProviderConfigAssumeRolePolicy(policy),
		fmt.Sprintf(`
data "aws_servicequotas_service_quota" "test" {
  service_code = %[1]q
  quota_name   = %[2]q
}
`, serviceCode, quotaName))
}
