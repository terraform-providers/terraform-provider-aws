package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emrcontainers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_emrcontainers_virtual_cluster", &resource.Sweeper{
		Name: "aws_emrcontainers_virtual_cluster",
		F:    testSweepEMRContainersVirtualCluster,
	})
}

func testSweepEMRContainersVirtualCluster(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).emrcontainersconn

	input := &emrcontainers.ListVirtualClustersInput{}
	err = conn.ListVirtualClustersPages(input, func(page *emrcontainers.ListVirtualClustersOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, vc := range page.VirtualClusters {
			log.Printf("[INFO] EMR containers virtual cluster: %s", aws.StringValue(vc.Id))
			_, err = conn.DeleteVirtualCluster(&emrcontainers.DeleteVirtualClusterInput{
				Id: vc.Id,
			})

			if err != nil {
				log.Printf("[ERROR] Error deleting containers virtual cluster (%s): %s", aws.StringValue(vc.Id), err)
			}
		}

		return !isLast
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping EMR containers virtual cluster sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error retrieving EMR containers virtual cluster: %s", err)
	}

	return nil
}

func TestAccAwsEMRContainersVirtualCluster_basic(t *testing.T) {
	var vc emrcontainers.DescribeVirtualClusterOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emrcontainers_virtual_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEMRContainersVirtualClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEMRContainersVirtualClusterBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEMRContainersVirtualClusterExists(resourceName, &vc),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "airflow", "environment/"+rName),
					resource.TestCheckResourceAttr(resourceName, "container_provider.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container_provider.0.id.#", rName),
					resource.TestCheckResourceAttr(resourceName, "container_provider.0.info.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "container_provider.0.type", "EKS"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "state", rName),
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

func TestAccAwsEMRContainersVirtualCluster_disappears(t *testing.T) {
	var vc emrcontainers.DescribeVirtualClusterOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emrcontainers_virtual_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEMRContainersVirtualClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEMRContainersVirtualClusterBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEMRContainersVirtualClusterExists(resourceName, &vc),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEMRContainersVirtualCluster(), resourceName),
				),
			},
		},
	})
}

func testAccCheckAwsEMRContainersVirtualClusterExists(resourceName string, vc *emrcontainers.DescribeVirtualClusterOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EMR containers virtual cluster ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).emrcontainersconn
		resp, err := conn.DescribeVirtualCluster(&emrcontainers.DescribeVirtualClusterInput{
			Id: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return fmt.Errorf("Error getting EMR containers virtual cluster: %s", err.Error())
		}

		*vc = *resp

		return nil
	}
}

func testAccCheckAwsEMRContainersVirtualClusterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).emrcontainersconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_emrcontainers_virtual_cluster" {
			continue
		}

		_, err := conn.DescribeVirtualCluster(&emrcontainers.DescribeVirtualClusterInput{
			Id: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if isAWSErr(err, emrcontainers.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return err
		}

		return fmt.Errorf("Expected EMR containers virtual cluster to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAcctestAccAwsEMRContainersVirtualClusterBase(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_partition" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_iam_role" "cluster" {
  name = "%[1]s-cluster"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = [
          "eks.${data.aws_partition.current.dns_suffix}",
          "eks-nodegroup.${data.aws_partition.current.dns_suffix}",
        ]
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role_policy_attachment" "cluster-AmazonEKSClusterPolicy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.cluster.name
}

resource "aws_iam_role" "node" {
  name = "%[1]s-node"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role_policy_attachment" "node-AmazonEKSWorkerNodePolicy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEKSWorkerNodePolicy"
  role       = aws_iam_role.node.name
}

resource "aws_iam_role_policy_attachment" "node-AmazonEKS_CNI_Policy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEKS_CNI_Policy"
  role       = aws_iam_role.node.name
}

resource "aws_iam_role_policy_attachment" "node-AmazonEC2ContainerRegistryReadOnly" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  role       = aws_iam_role.node.name
}

resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name                          = "tf-acc-test-eks-node-group"
    "kubernetes.io/cluster/%[1]s" = "shared"
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }
}

resource "aws_main_route_table_association" "test" {
  route_table_id = aws_route_table.test.id
  vpc_id         = aws_vpc.test.id
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = -1
    to_port     = 0
  }

  ingress {
    cidr_blocks = [aws_vpc.test.cidr_block]
    from_port   = 0
    protocol    = -1
    to_port     = 0
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone       = data.aws_availability_zones.available.names[count.index]
  cidr_block              = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  map_public_ip_on_launch = true
  vpc_id                  = aws_vpc.test.id

  tags = {
    Name                          = "tf-acc-test-eks-node-group"
    "kubernetes.io/cluster/%[1]s" = "shared"
  }
}

resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.cluster.arn

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [
    "aws_iam_role_policy_attachment.cluster-AmazonEKSClusterPolicy",
    "aws_main_route_table_association.test",
  ]
}

resource "aws_eks_node_group" "test" {
  cluster_name    = aws_eks_cluster.test.name
  node_group_name = %[1]q
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

data "aws_eks_cluster_auth" "cluster" {
  name = aws_eks_cluster.test.id
}

provider "kubernetes" {
  host                   =  aws_eks_cluster.test.endpoint
  cluster_ca_certificate = base64decode(aws_eks_cluster.test.certificate_authority.0.data)
  exec {
    api_version = "client.authentication.k8s.io/v1alpha1"
    args        = ["eks", "get-token", "--cluster-name", aws_eks_cluster.test.id]
    command     = "aws"
  }
}

resource "kubernetes_role" "emrcontainers_role" {
  metadata {
    name = "emr-containers"
    namespace = "default"
  }

  rule {
    api_groups     = [""]
    resources      = ["namespaces"]
    verbs          = ["get"]
  }

  rule {
    api_groups = [""]
    resources = ["serviceaccounts", "services", "configmaps", "events", "pods", "pods/log"]
    verbs = ["get", "list", "watch", "describe", "create", "edit", "delete", "deletecollection", "annotate", "patch", "label"]
  }

  rule {
    api_groups = [""]
    resources = ["secrets"]
    verbs = ["create", "patch", "delete", "watch"]
  }

  rule {
    api_groups = ["apps"]
    resources = ["statefulsets", "deployments"]
    verbs = ["get", "list", "watch", "describe", "create", "edit", "delete", "annotate", "patch", "label"]
  }

  rule {
    api_groups = ["batch"]
    resources = ["jobs"]
    verbs = ["get", "list", "watch", "describe", "create", "edit", "delete", "annotate", "patch", "label"]
  }

  rule {
    api_groups = ["extensions"]
    resources = ["ingresses"]
    verbs = ["get", "list", "watch", "describe", "create", "edit", "delete", "annotate", "patch", "label"]
  }

  rule {
    api_groups = ["rbac.authorization.k8s.io"]
    resources = ["roles", "rolebindings"]
    verbs = ["get", "list", "watch", "describe", "create", "edit", "delete", "deletecollection", "annotate", "patch", "label"]
  }
}

resource "kubernetes_role_binding" "emrcontainers_rolemapping" {
  metadata {
    name = "emr-containers"
    namespace = "default"
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "Role"
    name      = kubernetes_role.emrcontainers_role.metadata[0].name
  }
  subject {
    kind      = "User"
    name      = "emr-containers"
    api_group = "rbac.authorization.k8s.io"
  }
}

resource "kubernetes_config_map" "aws_auth" {
  metadata {
    name = "aws-cm-auth"
    namespace = "kube-system"
  }

  data = {
mapRoles = <<EOF
- rolearn: ${aws_iam_role.node.arn}
  username: system:node:{{EC2PrivateDNSName}}
  groups:
  - system:bootstrappers
  - system:nodes
- rolearn: arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/AWSServiceRoleForAmazonEMRContainers
  username: emr-containers
EOF
	}
}

`, rName)
}

func testAccAwsEMRContainersVirtualClusterBasicConfig(rName string) string {
	return testAcctestAccAwsEMRContainersVirtualClusterBase(rName) + fmt.Sprintf(`
resource "aws_emrcontainers_virtual_cluster" "test" {
  container_provider {
    id   = aws_eks_cluster.test.name
    type = "EKS"

    info {
      eks_info {
        namespace = "default"
      }
    }
  }

  name = %[1]q
}
`, rName)
}
