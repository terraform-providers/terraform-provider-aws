package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsConnectContactFlow_basic(t *testing.T) {
	var v connect.DescribeContactFlowOutput
	rInt := acctest.RandInt()
	resourceName := "aws_connect_contact_flow.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsConnectContactFlowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsConnectContactFlowConfigBasic(rInt, resourceName, "Created"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsConnectContactFlowExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "contact_flow_id"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrSet(resourceName, "content"),
					resource.TestCheckResourceAttr(resourceName, "type", connect.ContactFlowTypeContactFlow),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsConnectContactFlowConfigBasic(rInt, resourceName, "Updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsConnectContactFlowExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "contact_flow_id"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated"),
					resource.TestCheckResourceAttrSet(resourceName, "content"),
					resource.TestCheckResourceAttr(resourceName, "type", connect.ContactFlowTypeContactFlow),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
				),
			},
		},
	})
}

func TestAccAwsConnectContactFlow_filename(t *testing.T) {
	var v connect.DescribeContactFlowOutput
	rInt := acctest.RandInt()
	resourceName := "aws_connect_contact_flow.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsConnectContactFlowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsConnectContactFlowConfig_filename(rInt, resourceName, "Created", "test-fixtures/connect_contact_flow.json"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsConnectContactFlowExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "contact_flow_id"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttr(resourceName, "type", connect.ContactFlowTypeContactFlow),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"content_hash",
					"filename",
				},
			},
			{
				Config: testAccAwsConnectContactFlowConfig_filename(rInt, resourceName, "Updated", "test-fixtures/connect_contact_flow_updated.json"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsConnectContactFlowExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "contact_flow_id"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated"),
					resource.TestCheckResourceAttr(resourceName, "type", connect.ContactFlowTypeContactFlow),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
				),
			},
		},
	})
}

func TestAccAwsConnectContactFlow_disappears(t *testing.T) {
	var v connect.DescribeContactFlowOutput
	rInt := acctest.RandInt()
	resourceName := "aws_connect_contact_flow.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsConnectContactFlowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsConnectContactFlowConfigBasic(rInt, resourceName, "Disappear"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsConnectContactFlowExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsConnectContactFlow(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsConnectContactFlowExists(resourceName string, function *connect.DescribeContactFlowOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Connect flow not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Connect flow ID not set")
		}
		instanceID, contactFlowID, err := resourceAwsConnectContactFlowParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).connectconn

		params := &connect.DescribeContactFlowInput{
			ContactFlowId: aws.String(contactFlowID),
			InstanceId:    aws.String(instanceID),
		}

		getFunction, err := conn.DescribeContactFlow(params)
		if err != nil {
			return err
		}

		*function = *getFunction

		return nil
	}
}

func testAccCheckAwsConnectContactFlowDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_connect_contact_flow" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).connectconn

		instanceID, contactFlowID, err := resourceAwsConnectContactFlowParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		params := &connect.DescribeContactFlowInput{
			ContactFlowId: aws.String(contactFlowID),
			InstanceId:    aws.String(instanceID),
		}

		_, experr := conn.DescribeContactFlow(params)
		// Verify the error is what we want
		if experr != nil {
			if awsErr, ok := experr.(awserr.Error); ok && awsErr.Code() == "ResourceNotFoundException" {
				continue
			}
			return experr
		}
	}
	return nil
}

func testAccAwsConnectContactFlowConfigBasic(rInt int, contactFlowName string, label string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "foo" {
  instance_alias = "resource-test-terraform-%[1]d"
}

resource "aws_connect_contact_flow" "foo" {
  instance_id = aws_connect_instance.foo.id
  name        = "%[2]s"
  description = "%[3]s"
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
					"Text": "%[3]s"
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
    "Method", "%[3]s"
  )
}
`, rInt, contactFlowName, label)
}

func testAccAwsConnectContactFlowConfig_filename(rInt int, contactFlowName string, label string, filepath string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "foo" {
  instance_alias = "resource-test-terraform-%[1]d"
}

resource "aws_connect_contact_flow" "foo" {
  instance_id  = aws_connect_instance.foo.id
  name         = "%[2]s"
  description  = "%[3]s"
  type         = "CONTACT_FLOW"
  filename     = "%[4]s"
  content_hash = filebase64sha256("%[4]s")
  tags = map(
    "Name", "Test Contact Flow",
    "Method", "%[3]s"
  )
}
`, rInt, contactFlowName, label, filepath)
}
