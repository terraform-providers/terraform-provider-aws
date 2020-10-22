package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSDataSourceRedshiftCluster_basic(t *testing.T) {
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceRedshiftClusterConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_redshift_cluster.test", "allow_version_upgrade"),
					resource.TestCheckResourceAttrSet("data.aws_redshift_cluster.test", "automated_snapshot_retention_period"),
					resource.TestCheckResourceAttrSet("data.aws_redshift_cluster.test", "availability_zone"),
					resource.TestCheckResourceAttrSet("data.aws_redshift_cluster.test", "cluster_identifier"),
					resource.TestCheckResourceAttrSet("data.aws_redshift_cluster.test", "cluster_parameter_group_name"),
					resource.TestCheckResourceAttrSet("data.aws_redshift_cluster.test", "cluster_public_key"),
					resource.TestCheckResourceAttrSet("data.aws_redshift_cluster.test", "cluster_revision_number"),
					resource.TestCheckResourceAttr("data.aws_redshift_cluster.test", "cluster_type", "single-node"),
					resource.TestCheckResourceAttrSet("data.aws_redshift_cluster.test", "cluster_version"),
					resource.TestCheckResourceAttrSet("data.aws_redshift_cluster.test", "database_name"),
					resource.TestCheckResourceAttrSet("data.aws_redshift_cluster.test", "encrypted"),
					resource.TestCheckResourceAttrSet("data.aws_redshift_cluster.test", "endpoint"),
					resource.TestCheckResourceAttrSet("data.aws_redshift_cluster.test", "master_username"),
					resource.TestCheckResourceAttrSet("data.aws_redshift_cluster.test", "node_type"),
					resource.TestCheckResourceAttrSet("data.aws_redshift_cluster.test", "number_of_nodes"),
					resource.TestCheckResourceAttrSet("data.aws_redshift_cluster.test", "port"),
					resource.TestCheckResourceAttrSet("data.aws_redshift_cluster.test", "preferred_maintenance_window"),
					resource.TestCheckResourceAttrSet("data.aws_redshift_cluster.test", "publicly_accessible"),
				),
			},
		},
	})
}

func TestAccAWSDataSourceRedshiftCluster_vpc(t *testing.T) {
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceRedshiftClusterConfigWithVpc(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_redshift_cluster.test", "vpc_id"),
					resource.TestCheckResourceAttr("data.aws_redshift_cluster.test", "vpc_security_group_ids.#", "1"),
					resource.TestCheckResourceAttr("data.aws_redshift_cluster.test", "cluster_type", "multi-node"),
					resource.TestCheckResourceAttr("data.aws_redshift_cluster.test", "cluster_subnet_group_name", fmt.Sprintf("tf-redshift-subnet-group-%d", rInt)),
				),
			},
		},
	})
}

func TestAccAWSDataSourceRedshiftCluster_logging(t *testing.T) {
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceRedshiftClusterConfigWithLogging(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_redshift_cluster.test", "enable_logging", "true"),
					resource.TestCheckResourceAttr("data.aws_redshift_cluster.test", "bucket_name", fmt.Sprintf("tf-test-redshift-logging-%d", rInt)),
					resource.TestCheckResourceAttr("data.aws_redshift_cluster.test", "s3_key_prefix", "cluster-logging/"),
				),
			},
		},
	})
}

func testAccAWSDataSourceRedshiftClusterConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier = "tf-redshift-cluster-%d"

  database_name       = "testdb"
  master_username     = "foo"
  master_password     = "Password1"
  node_type           = "dc1.large"
  cluster_type        = "single-node"
  skip_final_snapshot = true
}

data "aws_redshift_cluster" "test" {
  cluster_identifier = aws_redshift_cluster.test.cluster_identifier
}
`, rInt)
}

func testAccAWSDataSourceRedshiftClusterConfigWithVpc(rInt int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_subnet" "foo" {
  cidr_block        = "10.1.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test.id
}

resource "aws_subnet" "bar" {
  cidr_block        = "10.1.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
  vpc_id            = aws_vpc.test.id
}

resource "aws_redshift_subnet_group" "test" {
  name       = "tf-redshift-subnet-group-%[1]d"
  subnet_ids = [aws_subnet.foo.id, aws_subnet.bar.id]
}

resource "aws_security_group" "test" {
  name   = "tf-redshift-sg-%[1]d"
  vpc_id = aws_vpc.test.id
}

resource "aws_redshift_cluster" "test" {
  cluster_identifier = "tf-redshift-cluster-%[1]d"

  database_name             = "testdb"
  master_username           = "foo"
  master_password           = "Password1"
  node_type                 = "dc1.large"
  cluster_type              = "multi-node"
  number_of_nodes           = 2
  publicly_accessible       = false
  cluster_subnet_group_name = aws_redshift_subnet_group.test.name
  vpc_security_group_ids    = [aws_security_group.test.id]
  skip_final_snapshot       = true
}

data "aws_redshift_cluster" "test" {
  cluster_identifier = aws_redshift_cluster.test.cluster_identifier
}
`, rInt))
}

func testAccAWSDataSourceRedshiftClusterConfigWithLogging(rInt int) string {
	return fmt.Sprintf(`
data "aws_redshift_service_account" "test" {}

resource "aws_s3_bucket" "test" {
  bucket        = "tf-test-redshift-logging-%[1]d"
  force_destroy = true
}

data "aws_iam_policy_document" "test" {
  statement {
    actions   = ["s3:PutObject"]
    resources = ["${aws_s3_bucket.test.arn}/*"]

    principals {
      identifiers = [data.aws_redshift_service_account.test.arn]
      type        = "AWS"
    }
  }

  statement {
    actions   = ["s3:GetBucketAcl"]
    resources = [aws_s3_bucket.test.arn]

    principals {
      identifiers = [data.aws_redshift_service_account.test.arn]
      type        = "AWS"
    }
  }
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.bucket
  policy = data.aws_iam_policy_document.test.json
}

resource "aws_redshift_cluster" "test" {
  depends_on = [aws_s3_bucket_policy.test]

  cluster_identifier  = "tf-redshift-cluster-%[1]d"
  cluster_type        = "single-node"
  database_name       = "testdb"
  master_password     = "Password1"
  master_username     = "foo"
  node_type           = "dc1.large"
  skip_final_snapshot = true

  logging {
    bucket_name   = aws_s3_bucket.test.id
    enable        = true
    s3_key_prefix = "cluster-logging/"
  }
}

data "aws_redshift_cluster" "test" {
  cluster_identifier = aws_redshift_cluster.test.cluster_identifier
}
`, rInt)
}
