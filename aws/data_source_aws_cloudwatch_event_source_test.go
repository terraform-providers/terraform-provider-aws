package aws

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsCloudWatchEventSource_basic(t *testing.T) {
	key := "EVENT_BRIDGE_PARTNER_EVENT_SOURCE_NAME"
	busName := os.Getenv(key)
	if busName == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	parts := strings.Split(busName, "/")
	if len(parts) < 2 {
		t.Errorf("unable to parse partner event bus name %s", busName)
	}
	createdBy := parts[0] + "/" + parts[1]

	dataSourceName := "data.aws_cloudwatch_event_source.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, cloudwatchevents.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDataSourcePartnerEventSourceConfig(busName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", busName),
					resource.TestCheckResourceAttr(dataSourceName, "created_by", createdBy),
					resource.TestCheckResourceAttrSet(dataSourceName, "arn"),
				),
			},
		},
	})
}

func testAccAwsDataSourcePartnerEventSourceConfig(namePrefix string) string {
	return fmt.Sprintf(`
data "aws_cloudwatch_event_source" "test" {
  name_prefix = "%s"
}
`, namePrefix)
}
