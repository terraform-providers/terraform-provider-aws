package aws

import (
	"fmt"

	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSEksNodegroupNamesDataSource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceResourceName := "data.aws_eks_node_group_names.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksNodeGroupNamesConfig(rName),
				Check:  resource.ComposeTestCheckFunc(),
			},
			{
				Config: testAccAWSEksNodeGroupNamesDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceResourceName, "cluster_name", rName),
					resource.TestCheckResourceAttr(dataSourceResourceName, "names.#", "2"),
				),
			},
		},
	})
}

func testAccAWSEksNodeGroupNamesDataSourceConfig(rName string) string {
	return composeConfig(
		testAccAWSEksNodeGroupNamesConfig(rName),
		fmt.Sprintf(`
data "aws_eks_node_group_names" "test" {
  cluster_name = aws_eks_cluster.test.name
}
`))
}

func testAccAWSEksNodeGroupNamesConfig(rName string) string {
	return testAccAWSEksNodeGroupConfigBase(rName) + fmt.Sprintf(`
resource "aws_eks_node_group" "test_a" {
  cluster_name    = aws_eks_cluster.test.name
  node_group_name = "%[1]s-test-a"
  node_role_arn   = aws_iam_role.node.arn
  subnet_ids      = aws_subnet.test[*].id

  scaling_config {
    desired_size = 1
    max_size     = 1
    min_size     = 1
  }

  depends_on = [
    "aws_iam_role_policy_attachment.node-AmazonEKSWorkerNodePolicy",
    "aws_iam_role_policy_attachment.node-AmazonEKS_CNI_Policy",
    "aws_iam_role_policy_attachment.node-AmazonEC2ContainerRegistryReadOnly",
  ]
}

resource "aws_eks_node_group" "test_b" {
	cluster_name    = aws_eks_cluster.test.name
	node_group_name = "%[1]s-test-b"
	node_role_arn   = aws_iam_role.node.arn
	subnet_ids      = aws_subnet.test[*].id

	scaling_config {
	  desired_size = 1
	  max_size     = 1
	  min_size     = 1
	}

	depends_on = [
	  "aws_iam_role_policy_attachment.node-AmazonEKSWorkerNodePolicy",
	  "aws_iam_role_policy_attachment.node-AmazonEKS_CNI_Policy",
	  "aws_iam_role_policy_attachment.node-AmazonEC2ContainerRegistryReadOnly",
	]
  }
`, rName)
}
