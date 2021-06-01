package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// These tests assume the existence of predefined Opsworks IAM roles named `aws-opsworks-ec2-role`
// and `aws-opsworks-service-role`.

func TestAccAWSOpsworksRailsAppLayer_basic(t *testing.T) {
	var opslayer opsworks.Layer
	stackName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_opsworks_rails_app_layer.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(opsworks.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOpsworksRailsAppLayerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOpsworksRailsAppLayerConfigVpcCreate(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSOpsworksLayerExists(resourceName, &opslayer),
					resource.TestCheckResourceAttr(resourceName, "name", stackName),
					resource.TestCheckResourceAttr(resourceName, "manage_bundler", "true"),
				),
			},
			{
				Config: testAccAwsOpsworksRailsAppLayerNoManageBundlerConfigVpcCreate(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSOpsworksLayerExists(resourceName, &opslayer),
					resource.TestCheckResourceAttr(resourceName, "name", stackName),
					resource.TestCheckResourceAttr(resourceName, "manage_bundler", "false"),
				),
			},
		},
	})
}

func TestAccAWSOpsworksRailsAppLayer_tags(t *testing.T) {
	var opslayer opsworks.Layer
	stackName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_opsworks_rails_app_layer.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(opsworks.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOpsworksRailsAppLayerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOpsworksRailsAppLayerConfigTags1(stackName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSOpsworksLayerExists(resourceName, &opslayer),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccAwsOpsworksRailsAppLayerConfigTags2(stackName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSOpsworksLayerExists(resourceName, &opslayer),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAwsOpsworksRailsAppLayerConfigTags1(stackName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSOpsworksLayerExists(resourceName, &opslayer),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAwsOpsworksRailsAppLayerDestroy(s *terraform.State) error {
	return testAccCheckAwsOpsworksLayerDestroy("aws_opsworks_rails_app_layer", s)
}

func testAccAwsOpsworksRailsAppLayerConfigVpcCreate(name string) string {
	return testAccAwsOpsworksStackConfigVpcCreate(name) +
		testAccAwsOpsworksCustomLayerSecurityGroups(name) +
		fmt.Sprintf(`
resource "aws_opsworks_rails_app_layer" "test" {
  stack_id = aws_opsworks_stack.tf-acc.id
  name     = "%s"

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-layer1.id,
    aws_security_group.tf-ops-acc-layer2.id,
  ]
}
`, name)
}

func testAccAwsOpsworksRailsAppLayerNoManageBundlerConfigVpcCreate(name string) string {
	return testAccAwsOpsworksStackConfigVpcCreate(name) +
		testAccAwsOpsworksCustomLayerSecurityGroups(name) +
		fmt.Sprintf(`
resource "aws_opsworks_rails_app_layer" "test" {
  stack_id = aws_opsworks_stack.tf-acc.id
  name     = "%s"

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-layer1.id,
    aws_security_group.tf-ops-acc-layer2.id,
  ]

  manage_bundler = false
}
`, name)
}

func testAccAwsOpsworksRailsAppLayerConfigTags1(name, tagKey1, tagValue1 string) string {
	return testAccAwsOpsworksStackConfigVpcCreate(name) +
		testAccAwsOpsworksCustomLayerSecurityGroups(name) +
		fmt.Sprintf(`
resource "aws_opsworks_rails_app_layer" "test" {
  stack_id = aws_opsworks_stack.tf-acc.id
  name     = "%s"

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-layer1.id,
    aws_security_group.tf-ops-acc-layer2.id,
  ]

  tags = {
    %[2]q = %[3]q
  }
}
`, name, tagKey1, tagValue1)
}

func testAccAwsOpsworksRailsAppLayerConfigTags2(name, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAwsOpsworksStackConfigVpcCreate(name) +
		testAccAwsOpsworksCustomLayerSecurityGroups(name) +
		fmt.Sprintf(`
resource "aws_opsworks_rails_app_layer" "test" {
  stack_id = aws_opsworks_stack.tf-acc.id
  name     = "%s"

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-layer1.id,
    aws_security_group.tf-ops-acc-layer2.id,
  ]

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, name, tagKey1, tagValue1, tagKey2, tagValue2)
}
