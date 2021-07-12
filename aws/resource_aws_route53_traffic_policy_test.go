package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSRoute53TrafficPolicy_basic(t *testing.T) {
	policyName := fmt.Sprintf("policy-%s", acctest.RandString(8))
	resourceName := "test"
	fullResourceName := fmt.Sprintf("aws_route53_traffic_policy.%s", resourceName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53TrafficPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53TrafficPolicyConfig(resourceName, policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceName, "latest_version", "1"),
				),
			},
			{
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAWSRoute53TrafficPolicyImportStateIdFunc(fullResourceName),
				ResourceName:      fullResourceName,
			},
		},
	})
}

func testAccRoute53TrafficPolicyConfig(resourceName, policyName string) string {
	return fmt.Sprintf(`
resource "aws_route53_traffic_policy" "%s" {
	name     = "%s"
	comment  = "comment"
	document = "{\"AWSPolicyFormatVersion\":\"2015-10-01\",\"RecordType\":\"A\",\"Endpoints\":{\"endpoint-start-NkPh\":{\"Type\":\"value\",\"Value\":\"10.0.0.1\"}},\"StartEndpoint\":\"endpoint-start-NkPh\"}"
}
`, resourceName, policyName)
}

func testAccCheckRoute53TrafficPolicyDestroy(s *terraform.State) error {
	return testAccCheckRoute53TrafficPolicyDestroyWithProvider(s, testAccProvider)
}

func testAccCheckRoute53TrafficPolicyDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*AWSClient).r53conn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_traffic_policy" {
			continue
		}
		tp, err := getTrafficPolicyById(rs.Primary.ID, conn)
		if err != nil {
			return fmt.Errorf("Error during check if traffic policy still exists, %#v", err)
		}
		if tp != nil {
			return fmt.Errorf("Traffic Policy still exists")
		}
	}
	return nil
}

func testAccAWSRoute53TrafficPolicyImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["id"], rs.Primary.Attributes["latest_version"]), nil
	}
}
