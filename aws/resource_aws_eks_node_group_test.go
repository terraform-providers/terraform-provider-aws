package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	resource.AddTestSweepers("aws_eks_fargate_node_group", &resource.Sweeper{
		Name: "aws_eks_fargate_node_group",
		F:    testSweepEksFargateNodegroups,
	})
}

func testSweepEksFargateNodegroups(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).eksconn

	var errors error
	input := &eks.ListClustersInput{}
	err = conn.ListClustersPages(input, func(page *eks.ListClustersOutput, lastPage bool) bool {
		for _, cluster := range page.Clusters {
			clusterName := aws.StringValue(cluster)
			input := &eks.ListNodegroupsInput{
				ClusterName: cluster,
			}
			err := conn.ListNodegroupsPages(input, func(page *eks.ListNodegroupsOutput, lastPage bool) bool {
				for _, nodegroup := range page.Nodegroups {
					nodegroupName := aws.StringValue(nodegroup)
					log.Printf("[INFO] Deleting EKS Node Group %q", nodegroupName)
					input := &eks.DeleteNodegroupInput{
						ClusterName:   cluster,
						NodegroupName: nodegroup,
					}
					_, err := conn.DeleteNodegroup(input)

					if err != nil && !isAWSErr(err, eks.ErrCodeResourceNotFoundException, "") {
						errors = multierror.Append(errors, fmt.Errorf("error deleting EKS Node Group %q: %w", nodegroupName, err))
						continue
					}

					if err := waitForEksNodeGroupDeletion(conn, clusterName, nodegroupName, 10*time.Minute); err != nil {
						errors = multierror.Append(errors, fmt.Errorf("error waiting for EKS Node Group %q deletion: %w", nodegroupName, err))
						continue
					}
				}
				return true
			})
			if err != nil {
				errors = multierror.Append(errors, fmt.Errorf("error listing Node Groups for EKS Cluster %s: %w", clusterName, err))
			}
		}

		return true
	})
	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping EKS Clusters sweep for %s: %s", region, err)
		return errors // In case we have completed some pages, but had errors
	}
	if err != nil {
		errors = multierror.Append(errors, fmt.Errorf("error retrieving EKS Clusters: %w", err))
	}

	return errors
}

func TestAccAWSEksNodeGroup_basic(t *testing.T) {
	var nodeGroup eks.Nodegroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	eksClusterResourceName := "aws_eks_cluster.test"
	iamRoleResourceName := "aws_iam_role.node"
	resourceName := "aws_eks_node_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksNodeGroupConfigNodeGroupName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksNodeGroupExists(resourceName, &nodeGroup),
					resource.TestCheckResourceAttr(resourceName, "ami_type", eks.AMITypesAl2X8664),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "eks", regexp.MustCompile(fmt.Sprintf("nodegroup/%[1]s/%[1]s/.+", rName))),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_name", eksClusterResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "disk_size", "20"),
					resource.TestCheckResourceAttr(resourceName, "instance_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "labels.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "node_group_name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "node_role_arn", iamRoleResourceName, "arn"),
					resource.TestMatchResourceAttr(resourceName, "release_version", regexp.MustCompile(`^\d+\.\d+\.\d+-\d{8}$`)),
					resource.TestCheckResourceAttr(resourceName, "remote_access.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resources.0.autoscaling_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.0.desired_size", "1"),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.0.max_size", "1"),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.0.min_size", "1"),
					resource.TestCheckResourceAttr(resourceName, "status", eks.NodegroupStatusActive),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "version", eksClusterResourceName, "version"),
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

func TestAccAWSEksNodeGroup_disappears(t *testing.T) {
	var nodeGroup eks.Nodegroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_node_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksNodeGroupConfigNodeGroupName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksNodeGroupExists(resourceName, &nodeGroup),
					testAccCheckAWSEksNodeGroupDisappears(&nodeGroup),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSEksNodeGroup_AmiType(t *testing.T) {
	var nodeGroup1 eks.Nodegroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_node_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksNodeGroupConfigAmiType(rName, eks.AMITypesAl2X8664Gpu),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksNodeGroupExists(resourceName, &nodeGroup1),
					resource.TestCheckResourceAttr(resourceName, "ami_type", eks.AMITypesAl2X8664Gpu),
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

func TestAccAWSEksNodeGroup_DiskSize(t *testing.T) {
	var nodeGroup1 eks.Nodegroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_node_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksNodeGroupConfigDiskSize(rName, 21),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksNodeGroupExists(resourceName, &nodeGroup1),
					resource.TestCheckResourceAttr(resourceName, "disk_size", "21"),
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

func TestAccAWSEksNodeGroup_InstanceTypes(t *testing.T) {
	var nodeGroup1 eks.Nodegroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_node_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksNodeGroupConfigInstanceTypes1(rName, "t3.large"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksNodeGroupExists(resourceName, &nodeGroup1),
					resource.TestCheckResourceAttr(resourceName, "instance_types.#", "1"),
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

func TestAccAWSEksNodeGroup_Labels(t *testing.T) {
	var nodeGroup1, nodeGroup2, nodeGroup3 eks.Nodegroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_node_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksNodeGroupConfigLabels1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksNodeGroupExists(resourceName, &nodeGroup1),
					resource.TestCheckResourceAttr(resourceName, "labels.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "labels.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEksNodeGroupConfigLabels2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksNodeGroupExists(resourceName, &nodeGroup2),
					resource.TestCheckResourceAttr(resourceName, "labels.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "labels.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "labels.key2", "value2"),
				),
			},
			{
				Config: testAccAWSEksNodeGroupConfigLabels1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksNodeGroupExists(resourceName, &nodeGroup3),
					resource.TestCheckResourceAttr(resourceName, "labels.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "labels.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSEksNodeGroup_ReleaseVersion(t *testing.T) {
	var nodeGroup1 eks.Nodegroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_node_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksNodeGroupConfigReleaseVersion(rName, "1.14.8-20191213"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksNodeGroupExists(resourceName, &nodeGroup1),
					resource.TestCheckResourceAttr(resourceName, "release_version", "1.14.8-20191213"),
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

func TestAccAWSEksNodeGroup_RemoteAccess_Ec2SshKey(t *testing.T) {
	var nodeGroup1 eks.Nodegroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_node_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksNodeGroupConfigRemoteAccessEc2SshKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksNodeGroupExists(resourceName, &nodeGroup1),
					resource.TestCheckResourceAttr(resourceName, "remote_access.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "remote_access.0.ec2_ssh_key", rName),
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

func TestAccAWSEksNodeGroup_RemoteAccess_SourceSecurityGroupIds(t *testing.T) {
	var nodeGroup1 eks.Nodegroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_node_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksNodeGroupConfigRemoteAccessSourceSecurityGroupIds1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksNodeGroupExists(resourceName, &nodeGroup1),
					resource.TestCheckResourceAttr(resourceName, "remote_access.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "remote_access.0.source_security_group_ids.#", "1"),
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

func TestAccAWSEksNodeGroup_ScalingConfig_DesiredSize(t *testing.T) {
	var nodeGroup1, nodeGroup2 eks.Nodegroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_node_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksNodeGroupConfigScalingConfigSizes(rName, 2, 2, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksNodeGroupExists(resourceName, &nodeGroup1),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.0.desired_size", "2"),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.0.max_size", "2"),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.0.min_size", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEksNodeGroupConfigScalingConfigSizes(rName, 1, 2, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksNodeGroupExists(resourceName, &nodeGroup2),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.0.desired_size", "1"),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.0.max_size", "2"),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.0.min_size", "1"),
				),
			},
		},
	})
}

func TestAccAWSEksNodeGroup_ScalingConfig_MaxSize(t *testing.T) {
	var nodeGroup1, nodeGroup2 eks.Nodegroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_node_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksNodeGroupConfigScalingConfigSizes(rName, 1, 2, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksNodeGroupExists(resourceName, &nodeGroup1),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.0.desired_size", "1"),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.0.max_size", "2"),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.0.min_size", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEksNodeGroupConfigScalingConfigSizes(rName, 1, 1, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksNodeGroupExists(resourceName, &nodeGroup2),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.0.desired_size", "1"),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.0.max_size", "1"),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.0.min_size", "1"),
				),
			},
		},
	})
}

func TestAccAWSEksNodeGroup_ScalingConfig_MinSize(t *testing.T) {
	var nodeGroup1, nodeGroup2 eks.Nodegroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_node_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksNodeGroupConfigScalingConfigSizes(rName, 2, 2, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksNodeGroupExists(resourceName, &nodeGroup1),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.0.desired_size", "2"),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.0.max_size", "2"),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.0.min_size", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEksNodeGroupConfigScalingConfigSizes(rName, 2, 2, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksNodeGroupExists(resourceName, &nodeGroup2),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.0.desired_size", "2"),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.0.max_size", "2"),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.0.min_size", "1"),
				),
			},
		},
	})
}

func TestAccAWSEksNodeGroup_Tags(t *testing.T) {
	var nodeGroup1, nodeGroup2, nodeGroup3 eks.Nodegroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_node_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksNodeGroupConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksNodeGroupExists(resourceName, &nodeGroup1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEksNodeGroupConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksNodeGroupExists(resourceName, &nodeGroup2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSEksNodeGroupConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksNodeGroupExists(resourceName, &nodeGroup3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSEksNodeGroup_Version(t *testing.T) {
	var nodeGroup1 eks.Nodegroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_node_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksNodeGroupConfigVersion(rName, "1.14"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksNodeGroupExists(resourceName, &nodeGroup1),
					resource.TestCheckResourceAttr(resourceName, "version", "1.14"),
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

func testAccCheckAWSEksNodeGroupExists(resourceName string, nodeGroup *eks.Nodegroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EKS Node Group ID is set")
		}

		clusterName, nodeGroupName, err := resourceAwsEksNodeGroupParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).eksconn

		input := &eks.DescribeNodegroupInput{
			ClusterName:   aws.String(clusterName),
			NodegroupName: aws.String(nodeGroupName),
		}

		output, err := conn.DescribeNodegroup(input)

		if err != nil {
			return err
		}

		if output == nil || output.Nodegroup == nil {
			return fmt.Errorf("EKS Node Group (%s) not found", rs.Primary.ID)
		}

		if aws.StringValue(output.Nodegroup.NodegroupName) != nodeGroupName {
			return fmt.Errorf("EKS Node Group (%s) not found", rs.Primary.ID)
		}

		if got, want := aws.StringValue(output.Nodegroup.Status), eks.NodegroupStatusActive; got != want {
			return fmt.Errorf("EKS Node Group (%s) not in %s status, got: %s", rs.Primary.ID, want, got)
		}

		*nodeGroup = *output.Nodegroup

		return nil
	}
}

func testAccCheckAWSEksNodeGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).eksconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_eks_node_group" {
			continue
		}

		clusterName, nodeGroupName, err := resourceAwsEksNodeGroupParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &eks.DescribeNodegroupInput{
			ClusterName:   aws.String(clusterName),
			NodegroupName: aws.String(nodeGroupName),
		}

		output, err := conn.DescribeNodegroup(input)

		if isAWSErr(err, eks.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if output != nil && output.Nodegroup != nil && aws.StringValue(output.Nodegroup.NodegroupName) == nodeGroupName {
			return fmt.Errorf("EKS Node Group (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSEksNodeGroupDisappears(nodeGroup *eks.Nodegroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).eksconn

		input := &eks.DeleteNodegroupInput{
			ClusterName:   nodeGroup.ClusterName,
			NodegroupName: nodeGroup.NodegroupName,
		}

		_, err := conn.DeleteNodegroup(input)

		if isAWSErr(err, eks.ErrCodeResourceNotFoundException, "") {
			return nil
		}

		if err != nil {
			return err
		}

		return waitForEksNodeGroupDeletion(conn, aws.StringValue(nodeGroup.ClusterName), aws.StringValue(nodeGroup.NodegroupName), 60*time.Minute)
	}
}

func testAccAWSEksNodeGroupConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role" "cluster" {
  name = "%[1]s-cluster"

  assume_role_policy = jsonencode({
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
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

resource "aws_iam_role_policy_attachment" "cluster-AmazonEKSServicePolicy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEKSServicePolicy"
  role       = aws_iam_role.cluster.name
}

resource "aws_iam_role" "node" {
  name = "%[1]s-node"
  
  assume_role_policy = jsonencode({
    Statement = [{
    Action    = "sts:AssumeRole"
    Effect    = "Allow"
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

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id            = aws_vpc.test.id

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
    "aws_iam_role_policy_attachment.cluster-AmazonEKSServicePolicy",
    "aws_main_route_table_association.test",
  ]
}
`, rName)
}

func testAccAWSEksNodeGroupConfigNodeGroupName(rName string) string {
	return testAccAWSEksNodeGroupConfigBase(rName) + fmt.Sprintf(`
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
`, rName)
}

func testAccAWSEksNodeGroupConfigAmiType(rName, amiType string) string {
	return testAccAWSEksNodeGroupConfigBase(rName) + fmt.Sprintf(`
resource "aws_eks_node_group" "test" {
  ami_type        = %[2]q
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
`, rName, amiType)
}

func testAccAWSEksNodeGroupConfigDiskSize(rName string, diskSize int) string {
	return testAccAWSEksNodeGroupConfigBase(rName) + fmt.Sprintf(`
resource "aws_eks_node_group" "test" {
  cluster_name    = aws_eks_cluster.test.name
  disk_size       = %[2]d
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
`, rName, diskSize)
}

func testAccAWSEksNodeGroupConfigInstanceTypes1(rName, instanceType1 string) string {
	return testAccAWSEksNodeGroupConfigBase(rName) + fmt.Sprintf(`
resource "aws_eks_node_group" "test" {
  cluster_name    = aws_eks_cluster.test.name
  instance_types  = [%[2]q]
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
`, rName, instanceType1)
}

func testAccAWSEksNodeGroupConfigLabels1(rName, labelKey1, labelValue1 string) string {
	return testAccAWSEksNodeGroupConfigBase(rName) + fmt.Sprintf(`
resource "aws_eks_node_group" "test" {
  cluster_name    = aws_eks_cluster.test.name
  node_group_name = %[1]q
  node_role_arn   = aws_iam_role.node.arn
  subnet_ids      = aws_subnet.test[*].id

  labels = {
    %[2]q = %[3]q
  }

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
`, rName, labelKey1, labelValue1)
}

func testAccAWSEksNodeGroupConfigLabels2(rName, labelKey1, labelValue1, labelKey2, labelValue2 string) string {
	return testAccAWSEksNodeGroupConfigBase(rName) + fmt.Sprintf(`
resource "aws_eks_node_group" "test" {
  cluster_name    = aws_eks_cluster.test.name
  node_group_name = %[1]q
  node_role_arn   = aws_iam_role.node.arn
  subnet_ids      = aws_subnet.test[*].id

  labels = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

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
`, rName, labelKey1, labelValue1, labelKey2, labelValue2)
}

func testAccAWSEksNodeGroupConfigReleaseVersion(rName, releaseVersion string) string {
	return testAccAWSEksNodeGroupConfigBase(rName) + fmt.Sprintf(`
resource "aws_eks_node_group" "test" {
  cluster_name    = aws_eks_cluster.test.name
  node_group_name = %[1]q
  node_role_arn   = aws_iam_role.node.arn
  release_version = %[2]q
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
`, rName, releaseVersion)
}

func testAccAWSEksNodeGroupConfigRemoteAccessEc2SshKey(rName string) string {
	return testAccAWSEksNodeGroupConfigBase(rName) + fmt.Sprintf(`
resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 example@example.com"
}

resource "aws_eks_node_group" "test" {
  cluster_name    = aws_eks_cluster.test.name
  node_group_name = %[1]q
  node_role_arn   = aws_iam_role.node.arn
  subnet_ids      = aws_subnet.test[*].id

  remote_access {
    ec2_ssh_key = aws_key_pair.test.key_name
  }

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

func testAccAWSEksNodeGroupConfigRemoteAccessSourceSecurityGroupIds1(rName string) string {
	return testAccAWSEksNodeGroupConfigBase(rName) + fmt.Sprintf(`
resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 example@example.com"
}

resource "aws_eks_node_group" "test" {
  cluster_name    = aws_eks_cluster.test.name
  node_group_name = %[1]q
  node_role_arn   = aws_iam_role.node.arn
  subnet_ids      = aws_subnet.test[*].id

  remote_access {
    ec2_ssh_key               = aws_key_pair.test.key_name
    source_security_group_ids = [aws_security_group.test.id]
  }

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

func testAccAWSEksNodeGroupConfigScalingConfigSizes(rName string, desiredSize, maxSize, minSize int) string {
	return testAccAWSEksNodeGroupConfigBase(rName) + fmt.Sprintf(`
resource "aws_eks_node_group" "test" {
  cluster_name    = aws_eks_cluster.test.name
  node_group_name = %[1]q
  node_role_arn   = aws_iam_role.node.arn
  subnet_ids      = aws_subnet.test[*].id

  scaling_config {
    desired_size = %[2]d
    max_size     = %[3]d
    min_size     = %[4]d
  }

  depends_on = [
    "aws_iam_role_policy_attachment.node-AmazonEKSWorkerNodePolicy",
    "aws_iam_role_policy_attachment.node-AmazonEKS_CNI_Policy",
    "aws_iam_role_policy_attachment.node-AmazonEC2ContainerRegistryReadOnly",
  ]
}
`, rName, desiredSize, maxSize, minSize)
}

func testAccAWSEksNodeGroupConfigTags1(rName, tagKey1, tagValue1 string) string {
	return testAccAWSEksNodeGroupConfigBase(rName) + fmt.Sprintf(`
resource "aws_eks_node_group" "test" {
  cluster_name    = aws_eks_cluster.test.name
  node_group_name = %[1]q
  node_role_arn   = aws_iam_role.node.arn
  subnet_ids      = aws_subnet.test[*].id

  tags = {
    %[2]q = %[3]q
  }

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
`, rName, tagKey1, tagValue1)
}

func testAccAWSEksNodeGroupConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAWSEksNodeGroupConfigBase(rName) + fmt.Sprintf(`
resource "aws_eks_node_group" "test" {
  cluster_name    = aws_eks_cluster.test.name
  node_group_name = %[1]q
  node_role_arn   = aws_iam_role.node.arn
  subnet_ids      = aws_subnet.test[*].id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

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
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSEksNodeGroupConfigVersion(rName, version string) string {
	return testAccAWSEksNodeGroupConfigBase(rName) + fmt.Sprintf(`
resource "aws_eks_node_group" "test" {
  cluster_name    = aws_eks_cluster.test.name
  node_group_name = %[1]q
  node_role_arn   = aws_iam_role.node.arn
  subnet_ids      = aws_subnet.test[*].id
  version         = %[2]q

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
`, rName, version)
}
