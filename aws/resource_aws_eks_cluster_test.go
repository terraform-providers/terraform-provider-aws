package aws

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfawsresource"
)

func init() {
	resource.AddTestSweepers("aws_eks_cluster", &resource.Sweeper{
		Name: "aws_eks_cluster",
		F:    testSweepEksClusters,
		Dependencies: []string{
			"aws_eks_fargate_profile",
			"aws_eks_fargate_node_group",
		},
	})
}

func testSweepEksClusters(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).eksconn

	var errors error
	input := &eks.ListClustersInput{}
	err = conn.ListClustersPages(input, func(page *eks.ListClustersOutput, lastPage bool) bool {
		for _, cluster := range page.Clusters {
			name := aws.StringValue(cluster)

			log.Printf("[INFO] Deleting EKS Cluster: %s", name)
			err := deleteEksCluster(conn, name)
			if err != nil {
				errors = multierror.Append(errors, fmt.Errorf("error deleting EKS Cluster %q: %w", name, err))
				continue
			}
			err = waitForDeleteEksCluster(conn, name, 15*time.Minute)
			if err != nil {
				errors = multierror.Append(errors, fmt.Errorf("error waiting for EKS Cluster %q deletion: %w", name, err))
				continue
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

func TestAccAWSEksCluster_basic(t *testing.T) {
	var cluster eks.Cluster

	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksClusterConfig_Required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksClusterExists(resourceName, &cluster),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "eks", regexp.MustCompile(fmt.Sprintf("cluster/%s$", rName))),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_authority.0.data"),
					resource.TestMatchResourceAttr(resourceName, "endpoint", regexp.MustCompile(`^https://`)),
					resource.TestCheckResourceAttr(resourceName, "identity.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "identity.0.oidc.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "identity.0.oidc.0.issuer", regexp.MustCompile(`^https://`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestMatchResourceAttr(resourceName, "platform_version", regexp.MustCompile(`^eks\.\d+$`)),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", eks.ClusterStatusActive),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestMatchResourceAttr(resourceName, "version", regexp.MustCompile(`^\d+\.\d+$`)),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.endpoint_private_access", "false"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.endpoint_public_access", "true"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", "2"),
					resource.TestMatchResourceAttr(resourceName, "vpc_config.0.vpc_id", regexp.MustCompile(`^vpc-.+`)),
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

func TestAccAWSEksCluster_EncryptionConfig(t *testing.T) {
	var cluster eks.Cluster
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_eks_cluster.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksClusterConfig_EncryptionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "encryption_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_config.0.provider.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_config.0.provider.0.key_arn", kmsKeyResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "encryption_config.0.resources.#", "1"),
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

func TestAccAWSEksCluster_Version(t *testing.T) {
	var cluster1, cluster2 eks.Cluster

	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksClusterConfig_Version(rName, "1.16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "version", "1.16"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEksClusterConfig_Version(rName, "1.17"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksClusterExists(resourceName, &cluster2),
					testAccCheckAWSEksClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "version", "1.17"),
				),
			},
		},
	})
}

func TestAccAWSEksCluster_Logging(t *testing.T) {
	var cluster1, cluster2 eks.Cluster

	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksClusterConfig_Logging(rName, []string{"api"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "enabled_cluster_log_types.#", "1"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "enabled_cluster_log_types.*", "api"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEksClusterConfig_Logging(rName, []string{"api", "audit"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksClusterExists(resourceName, &cluster2),
					testAccCheckAWSEksClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "enabled_cluster_log_types.#", "2"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "enabled_cluster_log_types.*", "api"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "enabled_cluster_log_types.*", "audit"),
				),
			},
			// Disable all log types.
			{
				Config: testAccAWSEksClusterConfig_Required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksClusterExists(resourceName, &cluster2),
					testAccCheckAWSEksClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "enabled_cluster_log_types.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSEksCluster_Tags(t *testing.T) {
	var cluster1, cluster2, cluster3 eks.Cluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksClusterConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksClusterExists(resourceName, &cluster1),
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
				Config: testAccAWSEksClusterConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksClusterExists(resourceName, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSEksClusterConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksClusterExists(resourceName, &cluster3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSEksCluster_VpcConfig_SecurityGroupIds(t *testing.T) {
	var cluster eks.Cluster

	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksClusterConfig_VpcConfig_SecurityGroupIds(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
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

func TestAccAWSEksCluster_VpcConfig_EndpointPrivateAccess(t *testing.T) {
	var cluster1, cluster2, cluster3 eks.Cluster

	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksClusterConfig_VpcConfig_EndpointPrivateAccess(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.endpoint_private_access", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEksClusterConfig_VpcConfig_EndpointPrivateAccess(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksClusterExists(resourceName, &cluster2),
					testAccCheckAWSEksClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.endpoint_private_access", "false"),
				),
			},
			{
				Config: testAccAWSEksClusterConfig_VpcConfig_EndpointPrivateAccess(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksClusterExists(resourceName, &cluster3),
					testAccCheckAWSEksClusterNotRecreated(&cluster2, &cluster3),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.endpoint_private_access", "true"),
				),
			},
		},
	})
}

func TestAccAWSEksCluster_VpcConfig_EndpointPublicAccess(t *testing.T) {
	var cluster1, cluster2, cluster3 eks.Cluster

	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksClusterConfig_VpcConfig_EndpointPublicAccess(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.endpoint_public_access", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEksClusterConfig_VpcConfig_EndpointPublicAccess(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksClusterExists(resourceName, &cluster2),
					testAccCheckAWSEksClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.endpoint_public_access", "true"),
				),
			},
			{
				Config: testAccAWSEksClusterConfig_VpcConfig_EndpointPublicAccess(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksClusterExists(resourceName, &cluster3),
					testAccCheckAWSEksClusterNotRecreated(&cluster2, &cluster3),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.endpoint_public_access", "false"),
				),
			},
		},
	})
}

func TestAccAWSEksCluster_VpcConfig_PublicAccessCidrs(t *testing.T) {
	var cluster1 eks.Cluster

	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksClusterConfig_VpcConfig_PublicAccessCidrs(rName, `["1.2.3.4/32", "5.6.7.8/32"]`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.public_access_cidrs.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEksClusterConfig_VpcConfig_PublicAccessCidrs(rName, `["4.3.2.1/32", "8.7.6.5/32"]`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.public_access_cidrs.#", "2"),
				),
			},
		},
	})
}

func testAccCheckAWSEksClusterExists(resourceName string, cluster *eks.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EKS Cluster ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).eksconn
		output, err := conn.DescribeCluster(&eks.DescribeClusterInput{
			Name: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if output == nil || output.Cluster == nil {
			return fmt.Errorf("EKS Cluster (%s) not found", rs.Primary.ID)
		}

		if aws.StringValue(output.Cluster.Name) != rs.Primary.ID {
			return fmt.Errorf("EKS Cluster (%s) not found", rs.Primary.ID)
		}

		*cluster = *output.Cluster

		return nil
	}
}

func testAccCheckAWSEksClusterDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_eks_cluster" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).eksconn

		// Handle eventual consistency
		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			output, err := conn.DescribeCluster(&eks.DescribeClusterInput{
				Name: aws.String(rs.Primary.ID),
			})

			if err != nil {
				if isAWSErr(err, eks.ErrCodeResourceNotFoundException, "") {
					return nil
				}
				return resource.NonRetryableError(err)
			}

			if output != nil && output.Cluster != nil && aws.StringValue(output.Cluster.Name) == rs.Primary.ID {
				return resource.RetryableError(fmt.Errorf("EKS Cluster %s still exists", rs.Primary.ID))
			}

			return nil
		})

		return err
	}

	return nil
}

func testAccCheckAWSEksClusterNotRecreated(i, j *eks.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CreatedAt) != aws.TimeValue(j.CreatedAt) {
			return errors.New("EKS Cluster was recreated")
		}

		return nil
	}
}

func testAccPreCheckAWSEks(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).eksconn

	input := &eks.ListClustersInput{}

	_, err := conn.ListClusters(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSEksClusterConfig_Base(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "eks.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "test-AmazonEKSClusterPolicy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.test.name
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name                          = "terraform-testacc-eks-cluster-base"
    "kubernetes.io/cluster/%[1]s" = "shared"
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name                          = "terraform-testacc-eks-cluster-base"
    "kubernetes.io/cluster/%[1]s" = "shared"
  }
}
`, rName)
}

func testAccAWSEksClusterConfig_Required(rName string) string {
	return composeConfig(testAccAWSEksClusterConfig_Base(rName), fmt.Sprintf(`
resource "aws_eks_cluster" "test" {
  name     = %q
  role_arn = aws_iam_role.test.arn

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName))
}

func testAccAWSEksClusterConfig_Version(rName, version string) string {
	return composeConfig(testAccAWSEksClusterConfig_Base(rName), fmt.Sprintf(`
resource "aws_eks_cluster" "test" {
  name     = %q
  role_arn = aws_iam_role.test.arn
  version  = %q

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName, version))
}

func testAccAWSEksClusterConfig_Logging(rName string, logTypes []string) string {
	return composeConfig(testAccAWSEksClusterConfig_Base(rName), fmt.Sprintf(`
resource "aws_eks_cluster" "test" {
  name                      = %q
  role_arn                  = aws_iam_role.test.arn
  enabled_cluster_log_types = ["%v"]

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName, strings.Join(logTypes, "\", \"")))
}

func testAccAWSEksClusterConfigTags1(rName, tagKey1, tagValue1 string) string {
	return composeConfig(testAccAWSEksClusterConfig_Base(rName), fmt.Sprintf(`
resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
  }

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName, tagKey1, tagValue1))
}

func testAccAWSEksClusterConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return composeConfig(testAccAWSEksClusterConfig_Base(rName), fmt.Sprintf(`
resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccAWSEksClusterConfig_EncryptionConfig(rName string) string {
	return composeConfig(testAccAWSEksClusterConfig_Base(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  encryption_config {
    resources = ["secrets"]

    provider {
      key_arn = aws_kms_key.test.arn
    }
  }

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName))
}

func testAccAWSEksClusterConfig_VpcConfig_SecurityGroupIds(rName string) string {
	return composeConfig(testAccAWSEksClusterConfig_Base(rName), fmt.Sprintf(`
resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "terraform-testacc-eks-cluster-sg"
  }
}

resource "aws_eks_cluster" "test" {
  name     = %q
  role_arn = aws_iam_role.test.arn

  vpc_config {
    security_group_ids = [aws_security_group.test.id]
    subnet_ids         = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName))
}

func testAccAWSEksClusterConfig_VpcConfig_EndpointPrivateAccess(rName string, endpointPrivateAccess bool) string {
	return composeConfig(testAccAWSEksClusterConfig_Base(rName), fmt.Sprintf(`
resource "aws_eks_cluster" "test" {
  name     = %q
  role_arn = aws_iam_role.test.arn

  vpc_config {
    endpoint_private_access = %t
    endpoint_public_access  = true
    subnet_ids              = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName, endpointPrivateAccess))
}

func testAccAWSEksClusterConfig_VpcConfig_EndpointPublicAccess(rName string, endpointPublicAccess bool) string {
	return composeConfig(testAccAWSEksClusterConfig_Base(rName), fmt.Sprintf(`
resource "aws_eks_cluster" "test" {
  name     = %q
  role_arn = aws_iam_role.test.arn

  vpc_config {
    endpoint_private_access = true
    endpoint_public_access  = %t
    subnet_ids              = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName, endpointPublicAccess))
}

func testAccAWSEksClusterConfig_VpcConfig_PublicAccessCidrs(rName string, publicAccessCidr string) string {
	return composeConfig(testAccAWSEksClusterConfig_Base(rName), fmt.Sprintf(`
resource "aws_eks_cluster" "test" {
  name     = %q
  role_arn = aws_iam_role.test.arn

  vpc_config {
    endpoint_private_access = true
    endpoint_public_access  = true
    public_access_cidrs     = %s
    subnet_ids              = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName, publicAccessCidr))
}
