package aws

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSDocDBCluster_basic(t *testing.T) {
	var dbCluster docdb.DBCluster
	rInt := acctest.RandInt()
	resourceName := "aws_docdb_cluster.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDocDBClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDocDBClusterConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocDBClusterExists(resourceName, &dbCluster),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "rds", regexp.MustCompile(`cluster:.+`)),
					resource.TestCheckResourceAttr(resourceName, "storage_encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "db_cluster_parameter_group_name", "default.docdb3.6"),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_resource_id"),
					resource.TestCheckResourceAttr(resourceName, "engine", "docdb"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_version"),
					resource.TestCheckResourceAttrSet(resourceName, "hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName,
						"enabled_cloudwatch_logs_exports.0", "audit"),
					resource.TestCheckResourceAttr(resourceName,
						"enabled_cloudwatch_logs_exports.1", "profiler"),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccAWSDocDBCluster_namePrefix(t *testing.T) {
	var v docdb.DBCluster

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDocDBClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDocDBClusterConfig_namePrefix(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocDBClusterExists("aws_docdb_cluster.test", &v),
					resource.TestMatchResourceAttr(
						"aws_docdb_cluster.test", "cluster_identifier", regexp.MustCompile("^tf-test-")),
				),
			},
			{
				ResourceName:      "aws_docdb_cluster.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccAWSDocDBCluster_generatedName(t *testing.T) {
	var v docdb.DBCluster

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDocDBClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDocDBClusterConfig_generatedName(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocDBClusterExists("aws_docdb_cluster.test", &v),
					resource.TestMatchResourceAttr(
						"aws_docdb_cluster.test", "cluster_identifier", regexp.MustCompile("^tf-")),
				),
			},
			{
				ResourceName:      "aws_docdb_cluster.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccAWSDocDBCluster_takeFinalSnapshot(t *testing.T) {
	var v docdb.DBCluster
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDocDBClusterSnapshot(rInt),
		Steps: []resource.TestStep{
			{
				Config: testAccDocDBClusterConfigWithFinalSnapshot(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocDBClusterExists("aws_docdb_cluster.default", &v),
				),
			},
			{
				ResourceName:      "aws_docdb_cluster.default",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

/// This is a regression test to make sure that we always cover the scenario as hightlighted in
/// https://github.com/hashicorp/terraform/issues/11568
func TestAccAWSDocDBCluster_missingUserNameCausesError(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDocDBClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccDocDBClusterConfigWithoutUserNameAndPassword(acctest.RandInt()),
				ExpectError: regexp.MustCompile(`required field is not set`),
			},
		},
	})
}

func TestAccAWSDocDBCluster_updateTags(t *testing.T) {
	var v docdb.DBCluster
	ri := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDocDBClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDocDBClusterConfig(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocDBClusterExists("aws_docdb_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_docdb_cluster.default", "tags.%", "1"),
				),
			},
			{
				ResourceName:      "aws_docdb_cluster.default",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccDocDBClusterConfigUpdatedTags(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocDBClusterExists("aws_docdb_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_docdb_cluster.default", "tags.%", "2"),
				),
			},
		},
	})
}

func TestAccAWSDocDBCluster_updateCloudwatchLogsExports(t *testing.T) {
	var v docdb.DBCluster
	ri := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDocDBClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDocDBClusterNoCloudwatchLogsConfig(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocDBClusterExists("aws_docdb_cluster.default", &v),
				),
			},
			{
				ResourceName:      "aws_docdb_cluster.default",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccDocDBClusterConfig(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocDBClusterExists("aws_docdb_cluster.default", &v),
					resource.TestCheckResourceAttr("aws_docdb_cluster.default",
						"enabled_cloudwatch_logs_exports.0", "audit"),
				),
			},
		},
	})
}

func TestAccAWSDocDBCluster_kmsKey(t *testing.T) {
	var v docdb.DBCluster

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDocDBClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDocDBClusterConfig_kmsKey(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocDBClusterExists("aws_docdb_cluster.default", &v),
					resource.TestCheckResourceAttrPair("aws_docdb_cluster.default", "kms_key_id", "aws_kms_key.foo", "arn"),
				),
			},
			{
				ResourceName:      "aws_docdb_cluster.default",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccAWSDocDBCluster_encrypted(t *testing.T) {
	var v docdb.DBCluster

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDocDBClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDocDBClusterConfig_encrypted(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocDBClusterExists("aws_docdb_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_docdb_cluster.default", "storage_encrypted", "true"),
					resource.TestCheckResourceAttr(
						"aws_docdb_cluster.default", "db_cluster_parameter_group_name", "default.docdb3.6"),
				),
			},
			{
				ResourceName:      "aws_docdb_cluster.default",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccAWSDocDBCluster_backupsUpdate(t *testing.T) {
	var v docdb.DBCluster

	ri := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDocDBClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDocDBClusterConfig_backups(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocDBClusterExists("aws_docdb_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_docdb_cluster.default", "preferred_backup_window", "07:00-09:00"),
					resource.TestCheckResourceAttr(
						"aws_docdb_cluster.default", "backup_retention_period", "5"),
					resource.TestCheckResourceAttr(
						"aws_docdb_cluster.default", "preferred_maintenance_window", "tue:04:00-tue:04:30"),
				),
			},
			{
				ResourceName:      "aws_docdb_cluster.default",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccDocDBClusterConfig_backupsUpdate(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocDBClusterExists("aws_docdb_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_docdb_cluster.default", "preferred_backup_window", "03:00-09:00"),
					resource.TestCheckResourceAttr(
						"aws_docdb_cluster.default", "backup_retention_period", "10"),
					resource.TestCheckResourceAttr(
						"aws_docdb_cluster.default", "preferred_maintenance_window", "wed:01:00-wed:01:30"),
				),
			},
		},
	})
}

func TestAccAWSDocDBCluster_Port(t *testing.T) {
	var dbCluster1, dbCluster2 docdb.DBCluster
	rInt := acctest.RandInt()
	resourceName := "aws_docdb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDocDBClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDocDBClusterConfig_Port(rInt, 5432),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocDBClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "port", "5432"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccDocDBClusterConfig_Port(rInt, 2345),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocDBClusterExists(resourceName, &dbCluster2),
					testAccCheckDocDBClusterRecreated(&dbCluster1, &dbCluster2),
					resource.TestCheckResourceAttr(resourceName, "port", "2345"),
				),
			},
		},
	})
}

func TestAccAWSDocDBCluster_deleteProtection(t *testing.T) {
	var dbCluster docdb.DBCluster
	resourceName := "aws_docdb_cluster.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDocDBClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDocDBClusterConfigDeleteProtection(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocDBClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"final_snapshot_identifier",
					"master_password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccDocDBClusterConfigDeleteProtection(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocDBClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
				),
			},
			{
				Config: testAccDocDBClusterConfigDeleteProtection(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocDBClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "true"),
				),
			},
			{
				Config: testAccDocDBClusterConfigDeleteProtection(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocDBClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
				),
			},
		},
	})
}

func testAccCheckDocDBClusterDestroy(s *terraform.State) error {
	return testAccCheckDocDBClusterDestroyWithProvider(s, testAccProvider)
}

func testAccCheckDocDBClusterDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*AWSClient).docdbconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_docdb_cluster" {
			continue
		}

		// Try to find the Group
		var err error
		resp, err := conn.DescribeDBClusters(
			&docdb.DescribeDBClustersInput{
				DBClusterIdentifier: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if len(resp.DBClusters) != 0 &&
				*resp.DBClusters[0].DBClusterIdentifier == rs.Primary.ID {
				return fmt.Errorf("DB Cluster %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the cluster is already destroyed
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "DBClusterNotFoundFault" {
				return nil
			}
		}

		return err
	}

	return nil
}

func testAccCheckDocDBClusterExists(n string, v *docdb.DBCluster) resource.TestCheckFunc {
	return testAccCheckDocDBClusterExistsWithProvider(n, v, func() *schema.Provider { return testAccProvider })
}

func testAccCheckDocDBClusterExistsWithProvider(n string, v *docdb.DBCluster, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DB Instance ID is set")
		}

		provider := providerF()
		conn := provider.Meta().(*AWSClient).docdbconn
		resp, err := conn.DescribeDBClusters(&docdb.DescribeDBClustersInput{
			DBClusterIdentifier: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		for _, c := range resp.DBClusters {
			if *c.DBClusterIdentifier == rs.Primary.ID {
				*v = *c
				return nil
			}
		}

		return fmt.Errorf("DB Cluster (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckDocDBClusterRecreated(i, j *docdb.DBCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.ClusterCreateTime) == aws.TimeValue(j.ClusterCreateTime) {
			return errors.New("DocDB Cluster was not recreated")
		}

		return nil
	}
}

func testAccCheckDocDBClusterSnapshot(rInt int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_docdb_cluster" {
				continue
			}

			// Try and delete the snapshot before we check for the cluster not found
			snapshot_identifier := fmt.Sprintf("tf-acctest-docdbcluster-snapshot-%d", rInt)

			awsClient := testAccProvider.Meta().(*AWSClient)
			conn := awsClient.docdbconn

			log.Printf("[INFO] Deleting the Snapshot %s", snapshot_identifier)
			_, snapDeleteErr := conn.DeleteDBClusterSnapshot(
				&docdb.DeleteDBClusterSnapshotInput{
					DBClusterSnapshotIdentifier: aws.String(snapshot_identifier),
				})
			if snapDeleteErr != nil {
				return snapDeleteErr
			}

			// Try to find the Group
			var err error
			resp, err := conn.DescribeDBClusters(
				&docdb.DescribeDBClustersInput{
					DBClusterIdentifier: aws.String(rs.Primary.ID),
				})

			if err == nil {
				if len(resp.DBClusters) != 0 &&
					*resp.DBClusters[0].DBClusterIdentifier == rs.Primary.ID {
					return fmt.Errorf("DB Cluster %s still exists", rs.Primary.ID)
				}
			}

			// Return nil if the cluster is already destroyed
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == "DBClusterNotFoundFault" {
					return nil
				}
			}

			return err
		}

		return nil
	}
}

func testAccDocDBClusterConfig(n int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_docdb_cluster" "default" {
  cluster_identifier = "tf-docdb-cluster-%d"

  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  master_username                 = "foo"
  master_password                 = "mustbeeightcharaters"
  db_cluster_parameter_group_name = "default.docdb3.6"
  skip_final_snapshot             = true

  tags = {
    Environment = "production"
  }

  enabled_cloudwatch_logs_exports = [
    "audit",
    "profiler",
  ]
}
`, n))
}

func testAccDocDBClusterConfig_namePrefix() string {
	return `
resource "aws_docdb_cluster" "test" {
  cluster_identifier_prefix = "tf-test-"
  master_username           = "root"
  master_password           = "password"
  skip_final_snapshot       = true
}
`
}

func testAccDocDBClusterConfig_generatedName() string {
	return `
resource "aws_docdb_cluster" "test" {
  master_username     = "root"
  master_password     = "password"
  skip_final_snapshot = true
}
`
}

func testAccDocDBClusterConfigWithFinalSnapshot(n int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_docdb_cluster" "default" {
  cluster_identifier = "tf-docdb-cluster-%[1]d"

  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  master_username                 = "foo"
  master_password                 = "mustbeeightcharaters"
  db_cluster_parameter_group_name = "default.docdb3.6"
  final_snapshot_identifier       = "tf-acctest-docdbcluster-snapshot-%[1]d"

  tags = {
    Environment = "production"
  }
}
`, n))
}

func testAccDocDBClusterConfigWithoutUserNameAndPassword(n int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_docdb_cluster" "default" {
  cluster_identifier = "tf-docdb-cluster-%d"

  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  skip_final_snapshot = true
}
`, n))
}

func testAccDocDBClusterConfigUpdatedTags(n int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_docdb_cluster" "default" {
  cluster_identifier = "tf-docdb-cluster-%d"

  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  master_username                 = "foo"
  master_password                 = "mustbeeightcharaters"
  db_cluster_parameter_group_name = "default.docdb3.6"
  skip_final_snapshot             = true

  tags = {
    Environment = "production"
    AnotherTag  = "test"
  }
}
`, n))
}

func testAccDocDBClusterNoCloudwatchLogsConfig(n int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_docdb_cluster" "default" {
  cluster_identifier = "tf-docdb-cluster-%d"

  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  master_username                 = "foo"
  master_password                 = "mustbeeightcharaters"
  db_cluster_parameter_group_name = "default.docdb3.6"
  skip_final_snapshot             = true

  tags = {
    Environment = "production"
  }
}
`, n))
}

func testAccDocDBClusterConfig_kmsKey(n int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_kms_key" "foo" {
  description = "Terraform acc test %[1]d"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_docdb_cluster" "default" {
  cluster_identifier = "tf-docdb-cluster-%[1]d"
  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  master_username                 = "foo"
  master_password                 = "mustbeeightcharaters"
  db_cluster_parameter_group_name = "default.docdb3.6"
  storage_encrypted               = true
  kms_key_id                      = aws_kms_key.foo.arn
  skip_final_snapshot             = true
}
`, n))
}

func testAccDocDBClusterConfig_encrypted(n int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_docdb_cluster" "default" {
  cluster_identifier = "tf-docdb-cluster-%d"

  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  master_username     = "foo"
  master_password     = "mustbeeightcharaters"
  storage_encrypted   = true
  skip_final_snapshot = true
}
`, n))
}

func testAccDocDBClusterConfig_backups(n int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_docdb_cluster" "default" {
  cluster_identifier = "tf-docdb-cluster-%d"

  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  master_username              = "foo"
  master_password              = "mustbeeightcharaters"
  backup_retention_period      = 5
  preferred_backup_window      = "07:00-09:00"
  preferred_maintenance_window = "tue:04:00-tue:04:30"
  skip_final_snapshot          = true
}
`, n))
}

func testAccDocDBClusterConfig_backupsUpdate(n int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_docdb_cluster" "default" {
  cluster_identifier = "tf-docdb-cluster-%d"

  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  master_username              = "foo"
  master_password              = "mustbeeightcharaters"
  backup_retention_period      = 10
  preferred_backup_window      = "03:00-09:00"
  preferred_maintenance_window = "wed:01:00-wed:01:30"
  apply_immediately            = true
  skip_final_snapshot          = true
}
`, n))
}

func testAccDocDBClusterConfig_Port(rInt, port int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_docdb_cluster" "test" {
  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  cluster_identifier              = "tf-acc-test-%d"
  db_cluster_parameter_group_name = "default.docdb3.6"
  engine                          = "docdb"
  master_password                 = "mustbeeightcharaters"
  master_username                 = "foo"
  port                            = %d
  skip_final_snapshot             = true
}
`, rInt, port))
}

func testAccDocDBClusterConfigDeleteProtection(isProtected bool) string {
	return fmt.Sprintf(`
resource "aws_docdb_cluster" "default" {
  cluster_identifier_prefix = "tf-test-"
  master_username           = "root"
  master_password           = "password"
  skip_final_snapshot       = true
  deletion_protection       = %t
}
`, isProtected)
}
