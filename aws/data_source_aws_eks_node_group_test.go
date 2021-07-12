package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSEksNodegroupDataSource_basic(t *testing.T) {
	var nodeGroup eks.Nodegroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceResourceName := "data.aws_eks_node_group.test"
	resourceName := "aws_eks_node_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksNodeGroupConfigNodeGroupName(rName),
				Check:  resource.ComposeTestCheckFunc(),
			},
			{
				Config: testAccAWSEksNodeGroupDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksNodeGroupExists(resourceName, &nodeGroup),
					resource.TestCheckResourceAttrPair(resourceName, "ami_type", dataSourceResourceName, "ami_type"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_name", dataSourceResourceName, "cluster_name"),
					resource.TestCheckResourceAttrPair(resourceName, "disk_size", dataSourceResourceName, "disk_size"),
					resource.TestCheckResourceAttr(dataSourceResourceName, "instance_types.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_type", dataSourceResourceName, "instance_type"),
					resource.TestCheckResourceAttrPair(resourceName, "labels.%", dataSourceResourceName, "labels.%"),
					resource.TestCheckResourceAttrPair(resourceName, "node_group_name", dataSourceResourceName, "node_group_name"),
					resource.TestCheckResourceAttrPair(resourceName, "node_role_arn", dataSourceResourceName, "node_role_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "release_version", dataSourceResourceName, "release_version"),
					resource.TestCheckResourceAttr(dataSourceResourceName, "remote_access.#", "0"),
					resource.TestCheckResourceAttr(dataSourceResourceName, "resources.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "resources", dataSourceResourceName, "resources"),
					resource.TestCheckResourceAttr(dataSourceResourceName, "scaling_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "scaling_config", dataSourceResourceName, "scaling_config"),
					resource.TestCheckResourceAttrPair(resourceName, "status", dataSourceResourceName, "status"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_ids.#", dataSourceResourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_ids", dataSourceResourceName, "subnet_ids"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceResourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "version", dataSourceResourceName, "version"),
				),
			},
		},
	})
}

func testAccAWSEksNodeGroupDataSourceConfig(rName string) string {
	return composeConfig(
		testAccAWSEksNodeGroupConfigNodeGroupName(rName),
		fmt.Sprintf(`
data "aws_eks_node_group" "test" {
  cluster_name    = aws_eks_cluster.test.name
  node_group_name = %[1]q
}
`, rName))
}
