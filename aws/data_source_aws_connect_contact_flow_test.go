package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsConnectContactFlow_basic(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_connect_contact_flow.foo"
	datasourceName := "data.aws_connect_contact_flow.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsConnectContactFlowDataSourceConfig_basic(rInt, resourceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "contact_flow_id", resourceName, "contact_flow_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "instance_id", resourceName, "instance_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "content", resourceName, "content"),
					resource.TestCheckResourceAttrPair(datasourceName, "type", resourceName, "type"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsConnectContactFlow_byname(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_connect_contact_flow.foo"
	datasourceName := "data.aws_connect_contact_flow.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsConnectContactFlowDataSourceConfig_byname(rInt, resourceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "contact_flow_id", resourceName, "contact_flow_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "instance_id", resourceName, "instance_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "content", resourceName, "content"),
					resource.TestCheckResourceAttrPair(datasourceName, "type", resourceName, "type"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccAwsConnectContactFlowDataSourceBaseConfig(rInt int, contactFlowName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "foo" {
  instance_alias = "resource-test-terraform-%d"
}

resource "aws_connect_contact_flow" "foo" {
  instance_id = aws_connect_instance.foo.id
  name        = "%[2]s"
  description = "Test Contact Flow Description"
  type        = "CONTACT_FLOW"
  content     = <<JSON
	{
		"Version": "2019-10-30",
		"StartAction": "12345678-1234-1234-1234-123456789012",
		"Actions": [ 
			{
				"Identifier": "12345678-1234-1234-1234-123456789012",
				"Type": "MessageParticipant",
				"Transitions": {
					"NextAction": "abcdef-abcd-abcd-abcd-abcdefghijkl",
					"Errors": [],
					"Conditions": []
				},
				"Parameters": {
					"Text": "Thanks for calling the sample flow!"
				}
			},
			{
				"Identifier": "abcdef-abcd-abcd-abcd-abcdefghijkl",
				"Type": "DisconnectParticipant",
				"Transitions": {},
				"Parameters": {}
			}
		]
	}
	JSON
  tags = map(
    "Name", "Test Contact Flow",
    "Application", "Terraform",
    "Method", "Create"
  )
}
	`, rInt, contactFlowName)
}

func testAccAwsConnectContactFlowDataSourceConfig_basic(rInt int, contactFlowName string) string {
	return fmt.Sprintf(testAccAwsConnectContactFlowDataSourceBaseConfig(rInt, contactFlowName) + `
data "aws_connect_contact_flow" "foo" {
  instance_id     = aws_connect_instance.foo.id
  contact_flow_id = aws_connect_contact_flow.foo.contact_flow_id
}
`)
}

func testAccAwsConnectContactFlowDataSourceConfig_byname(rInt int, contactFlowName string) string {
	return fmt.Sprintf(testAccAwsConnectContactFlowDataSourceBaseConfig(rInt, contactFlowName) + `
data "aws_connect_contact_flow" "foo" {
  instance_id = aws_connect_instance.foo.id
  name        = aws_connect_contact_flow.foo.name
}
`)
}
