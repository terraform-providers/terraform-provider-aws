package aws

import (
	//"errors"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSNeptuneCluster_basic(t *testing.T) {
	var dbCluster neptune.DBCluster
	rName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_neptune_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists(resourceName, &dbCluster),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "rds", regexp.MustCompile(`cluster:.+`)),
					resource.TestCheckResourceAttr(resourceName, "storage_encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "neptune_cluster_parameter_group_name", "default.neptune1"),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_resource_id"),
					resource.TestCheckResourceAttr(resourceName, "engine", "neptune"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_version"),
					resource.TestCheckResourceAttrSet(resourceName, "hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "0"),
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
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccAWSNeptuneCluster_namePrefix(t *testing.T) {
	var v neptune.DBCluster
	rName := "tf-test-"
	resourceName := "aws_neptune_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterConfig_namePrefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists(resourceName, &v),
					resource.TestMatchResourceAttr(resourceName, "cluster_identifier", regexp.MustCompile("^tf-test-")),
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
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccAWSNeptuneCluster_takeFinalSnapshot(t *testing.T) {
	var v neptune.DBCluster
	rName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_neptune_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterSnapshot(rName),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterConfigWithFinalSnapshot(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists(resourceName, &v),
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
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccAWSNeptuneCluster_tags(t *testing.T) {
	var v neptune.DBCluster
	rName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_neptune_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
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
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccAWSNeptuneClusterConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSNeptuneClusterConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSNeptuneCluster_updateIamRoles(t *testing.T) {
	var v neptune.DBCluster
	rName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_neptune_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterConfigIncludingIamRoles(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists(resourceName, &v),
				),
			},
			{
				Config: testAccAWSNeptuneClusterConfigAddIamRoles(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "iam_roles.#", "2"),
				),
			},
			{
				Config: testAccAWSNeptuneClusterConfigRemoveIamRoles(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "iam_roles.#", "1"),
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
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccAWSNeptuneCluster_kmsKey(t *testing.T) {
	var v neptune.DBCluster
	resourceName := "aws_neptune_cluster.test"
	keyResourceName := "aws_kms_key.test"
	rName := acctest.RandomWithPrefix("tf-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterConfig_kmsKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_arn", keyResourceName, "arn"),
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
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccAWSNeptuneCluster_encrypted(t *testing.T) {
	var v neptune.DBCluster
	resourceName := "aws_neptune_cluster.test"
	rName := acctest.RandomWithPrefix("tf-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterConfig_encrypted(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "storage_encrypted", "true"),
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
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccAWSNeptuneCluster_backupsUpdate(t *testing.T) {
	var v neptune.DBCluster
	resourceName := "aws_neptune_cluster.test"
	rName := acctest.RandomWithPrefix("tf-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterConfig_backups(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "preferred_backup_window", "07:00-09:00"),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period", "5"),
					resource.TestCheckResourceAttr(resourceName, "preferred_maintenance_window", "tue:04:00-tue:04:30"),
				),
			},
			{
				Config: testAccAWSNeptuneClusterConfig_backupsUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "preferred_backup_window", "03:00-09:00"),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period", "10"),
					resource.TestCheckResourceAttr(resourceName, "preferred_maintenance_window", "wed:01:00-wed:01:30"),
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
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccAWSNeptuneCluster_iamAuth(t *testing.T) {
	var v neptune.DBCluster
	resourceName := "aws_neptune_cluster.test"
	rName := acctest.RandomWithPrefix("tf-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterConfig_iamAuth(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "iam_database_authentication_enabled", "true"),
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
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccAWSNeptuneCluster_updateCloudwatchLogsExports(t *testing.T) {
	var dbCluster neptune.DBCluster
	rName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_neptune_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists(resourceName, &dbCluster),
					resource.TestCheckNoResourceAttr(resourceName, "enable_cloudwatch_logs_exports.#"),
				),
			},
			{
				Config: testAccAWSNeptuneClusterConfig_cloudwatchLogsExports(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "enable_cloudwatch_logs_exports.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enable_cloudwatch_logs_exports.*", "audit"),
				),
			},
			{
				Config: testAccAWSNeptuneClusterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "enable_cloudwatch_logs_exports.#", "0"),
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
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccAWSNeptuneCluster_deleteProtection(t *testing.T) {
	var dbCluster neptune.DBCluster
	rName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_neptune_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists(resourceName, &dbCluster),
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
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccAWSNeptuneClusterConfigDeleteProtection(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "true"),
				),
			},
			{
				Config: testAccAWSNeptuneClusterConfigDeleteProtection(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
				),
			},
		},
	})
}

func testAccCheckAWSNeptuneClusterDestroy(s *terraform.State) error {
	return testAccCheckAWSNeptuneClusterDestroyWithProvider(s, testAccProvider)
}

func testAccCheckAWSNeptuneClusterDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*AWSClient).neptuneconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_neptune_cluster" {
			continue
		}

		// Try to find the Group
		var err error
		resp, err := conn.DescribeDBClusters(
			&neptune.DescribeDBClustersInput{
				DBClusterIdentifier: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if len(resp.DBClusters) != 0 &&
				aws.StringValue(resp.DBClusters[0].DBClusterIdentifier) == rs.Primary.ID {
				return fmt.Errorf("Neptune Cluster %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the cluster is already destroyed
		if err != nil {
			if isAWSErr(err, neptune.ErrCodeDBClusterNotFoundFault, "") {
				return nil
			}
		}

		return err
	}

	return nil
}

func testAccCheckAWSNeptuneClusterExists(n string, v *neptune.DBCluster) resource.TestCheckFunc {
	return testAccCheckAWSNeptuneClusterExistsWithProvider(n, v, func() *schema.Provider { return testAccProvider })
}

func testAccCheckAWSNeptuneClusterExistsWithProvider(n string, v *neptune.DBCluster, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Neptune Instance ID is set")
		}

		provider := providerF()
		conn := provider.Meta().(*AWSClient).neptuneconn
		resp, err := conn.DescribeDBClusters(&neptune.DescribeDBClustersInput{
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

		return fmt.Errorf("Neptune Cluster (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckAWSNeptuneClusterSnapshot(rName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_neptune_cluster" {
				continue
			}

			awsClient := testAccProvider.Meta().(*AWSClient)
			conn := awsClient.neptuneconn

			log.Printf("[INFO] Deleting the Snapshot %s", rName)
			_, snapDeleteErr := conn.DeleteDBClusterSnapshot(
				&neptune.DeleteDBClusterSnapshotInput{
					DBClusterSnapshotIdentifier: aws.String(rName),
				})
			if snapDeleteErr != nil {
				return snapDeleteErr
			}

			// Try to find the Group
			var err error
			resp, err := conn.DescribeDBClusters(
				&neptune.DescribeDBClustersInput{
					DBClusterIdentifier: aws.String(rs.Primary.ID),
				})

			if err == nil {
				if len(resp.DBClusters) != 0 &&
					aws.StringValue(resp.DBClusters[0].DBClusterIdentifier) == rs.Primary.ID {
					return fmt.Errorf("Neptune Cluster %s still exists", rs.Primary.ID)
				}
			}

			// Return nil if the cluster is already destroyed
			if err != nil {
				if isAWSErr(err, neptune.ErrCodeDBClusterNotFoundFault, "") {
					return nil
				}
			}

			return err
		}

		return nil
	}
}

func testAccAWSNeptuneClusterConfigBase() string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), `
locals {
  availability_zone_names = slice(data.aws_availability_zones.available.names, 0, min(3, length(data.aws_availability_zones.available.names)))
}
`)
}

func testAccAWSNeptuneClusterConfig(rName string) string {
	return composeConfig(testAccAWSNeptuneClusterConfigBase(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %q
  availability_zones                   = local.availability_zone_names
  engine                               = "neptune"
  neptune_cluster_parameter_group_name = "default.neptune1"
  skip_final_snapshot                  = true
}
`, rName))
}

func testAccAWSNeptuneClusterConfigDeleteProtection(rName string, isProtected bool) string {
	return composeConfig(testAccAWSNeptuneClusterConfigBase(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %q
  availability_zones                   = local.availability_zone_names
  engine                               = "neptune"
  neptune_cluster_parameter_group_name = "default.neptune1"
  skip_final_snapshot                  = true
  deletion_protection                  = %t
}
`, rName, isProtected))
}

func testAccAWSNeptuneClusterConfig_namePrefix(rName string) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier_prefix            = %q
  engine                               = "neptune"
  neptune_cluster_parameter_group_name = "default.neptune1"
  skip_final_snapshot                  = true
}
`, rName)
}

func testAccAWSNeptuneClusterConfigWithFinalSnapshot(rName string) string {
	return composeConfig(testAccAWSNeptuneClusterConfigBase(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %[1]q
  availability_zones                   = local.availability_zone_names
  neptune_cluster_parameter_group_name = "default.neptune1"
  final_snapshot_identifier            = %[1]q
}
`, rName))
}

func testAccAWSNeptuneClusterConfigTags1(rName, tagKey1, tagValue1 string) string {
	return composeConfig(testAccAWSNeptuneClusterConfigBase(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %[1]q
  availability_zones                   = local.availability_zone_names
  engine                               = "neptune"
  neptune_cluster_parameter_group_name = "default.neptune1"
  skip_final_snapshot                  = true

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccAWSNeptuneClusterConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return composeConfig(testAccAWSNeptuneClusterConfigBase(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %[1]q
  availability_zones                   = local.availability_zone_names
  engine                               = "neptune"
  neptune_cluster_parameter_group_name = "default.neptune1"
  skip_final_snapshot                  = true

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccAWSNeptuneClusterConfigIncludingIamRoles(rName string) string {
	return composeConfig(testAccAWSNeptuneClusterConfigBase(), fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "rds.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF
}

resource "aws_iam_role" "test-2" {
  name = "%[1]s-2"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "rds.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test-2" {
  name = "%[1]s-2"
  role = aws_iam_role.test-2.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF
}

resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %[1]q
  availability_zones                   = local.availability_zone_names
  neptune_cluster_parameter_group_name = "default.neptune1"
  skip_final_snapshot                  = true

  depends_on = [aws_iam_role.test, aws_iam_role.test-2]
}
`, rName))
}

func testAccAWSNeptuneClusterConfigAddIamRoles(rName string) string {
	return composeConfig(testAccAWSNeptuneClusterConfigBase(), fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "rds.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF
}

resource "aws_iam_role" "test-2" {
  name = "%[1]s-2"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "rds.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test-2" {
  name = "%[1]s-2"
  role = aws_iam_role.test-2.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF
}

resource "aws_neptune_cluster" "test" {
  cluster_identifier  = %[1]q
  availability_zones  = local.availability_zone_names
  skip_final_snapshot = true
  iam_roles           = [aws_iam_role.test.arn, aws_iam_role.test-2.arn]

  depends_on = [aws_iam_role.test, aws_iam_role.test-2]
}
`, rName))
}

func testAccAWSNeptuneClusterConfigRemoveIamRoles(rName string) string {
	return composeConfig(testAccAWSNeptuneClusterConfigBase(), fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "rds.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF
}

resource "aws_neptune_cluster" "test" {
  cluster_identifier  = %[1]q
  availability_zones  = local.availability_zone_names
  skip_final_snapshot = true
  iam_roles           = [aws_iam_role.test.arn]

  depends_on = [aws_iam_role.test]
}
`, rName))
}

func testAccAWSNeptuneClusterConfig_kmsKey(rName string) string {
	return composeConfig(testAccAWSNeptuneClusterConfigBase(), fmt.Sprintf(`

resource "aws_kms_key" "test" {
  description = "Terraform acc test"

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

resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %q
  availability_zones                   = local.availability_zone_names
  neptune_cluster_parameter_group_name = "default.neptune1"
  storage_encrypted                    = true
  kms_key_arn                          = aws_kms_key.test.arn
  skip_final_snapshot                  = true
}
`, rName))
}

func testAccAWSNeptuneClusterConfig_encrypted(rName string) string {
	return composeConfig(testAccAWSNeptuneClusterConfigBase(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier  = %q
  availability_zones  = local.availability_zone_names
  storage_encrypted   = true
  skip_final_snapshot = true
}
`, rName))
}

func testAccAWSNeptuneClusterConfig_backups(rName string) string {
	return composeConfig(testAccAWSNeptuneClusterConfigBase(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier           = %q
  availability_zones           = local.availability_zone_names
  backup_retention_period      = 5
  preferred_backup_window      = "07:00-09:00"
  preferred_maintenance_window = "tue:04:00-tue:04:30"
  skip_final_snapshot          = true
}
`, rName))
}

func testAccAWSNeptuneClusterConfig_backupsUpdate(rName string) string {
	return composeConfig(testAccAWSNeptuneClusterConfigBase(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier           = %q
  availability_zones           = local.availability_zone_names
  backup_retention_period      = 10
  preferred_backup_window      = "03:00-09:00"
  preferred_maintenance_window = "wed:01:00-wed:01:30"
  apply_immediately            = true
  skip_final_snapshot          = true
}
`, rName))
}

func testAccAWSNeptuneClusterConfig_iamAuth(rName string) string {
	return composeConfig(testAccAWSNeptuneClusterConfigBase(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier                  = %q
  availability_zones                  = local.availability_zone_names
  iam_database_authentication_enabled = true
  skip_final_snapshot                 = true
}
`, rName))
}

func testAccAWSNeptuneClusterConfig_cloudwatchLogsExports(rName string) string {
	return composeConfig(testAccAWSNeptuneClusterConfigBase(), fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier             = %q
  availability_zones             = local.availability_zone_names
  skip_final_snapshot            = true
  enable_cloudwatch_logs_exports = ["audit"]
}
`, rName))
}
