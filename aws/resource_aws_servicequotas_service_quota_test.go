package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/servicequotas"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// This resource is different than many since quotas are pre-existing
// and the resource is only designed to help with increases.
// In the basic case, we test that the resource can match the existing quota
// without unexpected changes.
func TestAccAwsServiceQuotasServiceQuota_basic(t *testing.T) {
	dataSourceName := "data.aws_servicequotas_service_quota.test"
	resourceName := "aws_servicequotas_service_quota.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSServiceQuotas(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsServiceQuotasServiceQuotaConfigSameValue("L-F678F1CE", "vpc"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "adjustable", dataSourceName, "adjustable"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "default_value", dataSourceName, "default_value"),
					resource.TestCheckResourceAttrPair(resourceName, "quota_code", dataSourceName, "quota_code"),
					resource.TestCheckResourceAttrPair(resourceName, "quota_name", dataSourceName, "quota_name"),
					resource.TestCheckResourceAttrPair(resourceName, "service_code", dataSourceName, "service_code"),
					resource.TestCheckResourceAttrPair(resourceName, "service_name", dataSourceName, "service_name"),
					resource.TestCheckResourceAttrPair(resourceName, "value", dataSourceName, "value"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsServiceQuotasServiceQuota_Value_IncreaseOnCreate(t *testing.T) {
	quotaCode := os.Getenv("SERVICEQUOTAS_INCREASE_ON_CREATE_QUOTA_CODE")
	if quotaCode == "" {
		t.Skip(
			"Environment variable SERVICEQUOTAS_INCREASE_ON_CREATE_QUOTA_CODE is not set. " +
				"WARNING: This test will submit a real service quota increase!")
	}

	serviceCode := os.Getenv("SERVICEQUOTAS_INCREASE_ON_CREATE_SERVICE_CODE")
	if serviceCode == "" {
		t.Skip(
			"Environment variable SERVICEQUOTAS_INCREASE_ON_CREATE_SERVICE_CODE is not set. " +
				"WARNING: This test will submit a real service quota increase!")
	}

	value := os.Getenv("SERVICEQUOTAS_INCREASE_ON_CREATE_VALUE")
	if value == "" {
		t.Skip(
			"Environment variable SERVICEQUOTAS_INCREASE_ON_CREATE_VALUE is not set. " +
				"WARNING: This test will submit a real service quota increase!")
	}

	resourceName := "aws_servicequotas_service_quota.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSServiceQuotas(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsServiceQuotasServiceQuotaConfigValue(quotaCode, serviceCode, value),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "quota_code", quotaCode),
					resource.TestCheckResourceAttr(resourceName, "service_code", serviceCode),
					resource.TestCheckResourceAttr(resourceName, "value", value),
				),
			},
		},
	})
}

func TestAccAwsServiceQuotasServiceQuota_Value_IncreaseOnUpdate(t *testing.T) {
	quotaCode := os.Getenv("SERVICEQUOTAS_INCREASE_ON_UPDATE_QUOTA_CODE")
	if quotaCode == "" {
		t.Skip(
			"Environment variable SERVICEQUOTAS_INCREASE_ON_UPDATE_QUOTA_CODE is not set. " +
				"WARNING: This test will submit a real service quota increase!")
	}

	serviceCode := os.Getenv("SERVICEQUOTAS_INCREASE_ON_UPDATE_SERVICE_CODE")
	if serviceCode == "" {
		t.Skip(
			"Environment variable SERVICEQUOTAS_INCREASE_ON_UPDATE_SERVICE_CODE is not set. " +
				"WARNING: This test will submit a real service quota increase!")
	}

	value := os.Getenv("SERVICEQUOTAS_INCREASE_ON_UPDATE_VALUE")
	if value == "" {
		t.Skip(
			"Environment variable SERVICEQUOTAS_INCREASE_ON_UPDATE_VALUE is not set. " +
				"WARNING: This test will submit a real service quota increase!")
	}

	dataSourceName := "aws_servicequotas_service_quota.test"
	resourceName := "aws_servicequotas_service_quota.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSServiceQuotas(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsServiceQuotasServiceQuotaConfigSameValue(quotaCode, serviceCode),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "quota_code", quotaCode),
					resource.TestCheckResourceAttr(resourceName, "service_code", serviceCode),
					resource.TestCheckResourceAttrPair(resourceName, "value", dataSourceName, "value"),
				),
			},
			{
				Config: testAccAwsServiceQuotasServiceQuotaConfigValue(quotaCode, serviceCode, value),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "quota_code", quotaCode),
					resource.TestCheckResourceAttr(resourceName, "service_code", serviceCode),
					resource.TestCheckResourceAttr(resourceName, "value", value),
				),
			},
		},
	})
}

func testAccPreCheckAWSServiceQuotas(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).servicequotasconn

	input := &servicequotas.ListServicesInput{}

	_, err := conn.ListServices(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAwsServiceQuotasServiceQuotaConfigSameValue(quotaCode, serviceCode string) string {
	return fmt.Sprintf(`
data "aws_servicequotas_service_quota" "test" {
  quota_code   = %[1]q
  service_code = %[2]q
}

resource "aws_servicequotas_service_quota" "test" {
  quota_code   = data.aws_servicequotas_service_quota.test.quota_code
  service_code = data.aws_servicequotas_service_quota.test.service_code
  value        = data.aws_servicequotas_service_quota.test.value
}
`, quotaCode, serviceCode)
}

func testAccAwsServiceQuotasServiceQuotaConfigValue(quotaCode, serviceCode, value string) string {
	return fmt.Sprintf(`
resource "aws_servicequotas_service_quota" "test" {
  quota_code   = %[1]q
  service_code = %[2]q
  value        = %[3]s
}
`, quotaCode, serviceCode, value)
}
