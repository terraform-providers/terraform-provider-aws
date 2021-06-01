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

func TestAccAWSOpsworksJavaAppLayer_basic(t *testing.T) {
	var opslayer opsworks.Layer
	stackName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_opsworks_java_app_layer.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(opsworks.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOpsworksJavaAppLayerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOpsworksJavaAppLayerConfigVpcCreate(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSOpsworksLayerExists(resourceName, &opslayer),
					resource.TestCheckResourceAttr(resourceName, "name", stackName)),
			},
		},
	})
}

func TestAccAWSOpsworksJavaAppLayer_tags(t *testing.T) {
	var opslayer opsworks.Layer
	stackName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_opsworks_java_app_layer.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(opsworks.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOpsworksJavaAppLayerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOpsworksJavaAppLayerConfigTags1(stackName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSOpsworksLayerExists(resourceName, &opslayer),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccAwsOpsworksJavaAppLayerConfigTags2(stackName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSOpsworksLayerExists(resourceName, &opslayer),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAwsOpsworksJavaAppLayerConfigTags1(stackName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSOpsworksLayerExists(resourceName, &opslayer),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAwsOpsworksJavaAppLayerDestroy(s *terraform.State) error {
	return testAccCheckAwsOpsworksLayerDestroy("aws_opsworks_java_app_layer", s)
}

func testAccAwsOpsworksJavaAppLayerConfigVpcCreate(name string) string {
	return testAccAwsOpsworksStackConfigVpcCreate(name) +
		testAccAwsOpsworksCustomLayerSecurityGroups(name) +
		fmt.Sprintf(`
resource "aws_opsworks_java_app_layer" "test" {
  stack_id = aws_opsworks_stack.tf-acc.id
  name     = "%s"

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-layer1.id,
    aws_security_group.tf-ops-acc-layer2.id,
  ]
}
`, name)
}

func testAccAwsOpsworksJavaAppLayerConfigTags1(name, tagKey1, tagValue1 string) string {
	return testAccAwsOpsworksStackConfigVpcCreate(name) +
		testAccAwsOpsworksCustomLayerSecurityGroups(name) +
		fmt.Sprintf(`
resource "aws_opsworks_java_app_layer" "test" {
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

func testAccAwsOpsworksJavaAppLayerConfigTags2(name, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAwsOpsworksStackConfigVpcCreate(name) +
		testAccAwsOpsworksCustomLayerSecurityGroups(name) +
		fmt.Sprintf(`
resource "aws_opsworks_java_app_layer" "test" {
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
