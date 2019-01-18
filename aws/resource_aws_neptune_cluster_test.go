package aws

import (
	//"errors"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
)

func TestAccAWSNeptuneCluster_basic(t *testing.T) {
	var dbCluster neptune.DBCluster
	rInt := acctest.RandInt()
	resourceName := "aws_neptune_cluster.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "storage_encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "neptune_cluster_parameter_group_name", "default.neptune1"),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_resource_id"),
					resource.TestCheckResourceAttr(resourceName, "engine", "neptune"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_version"),
					resource.TestCheckResourceAttrSet(resourceName, "hosted_zone_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_final_snapshot"},
			},
		},
	})
}

func TestAccAWSNeptuneCluster_namePrefix(t *testing.T) {
	var v neptune.DBCluster

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterConfig_namePrefix(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists("aws_neptune_cluster.test", &v),
					resource.TestMatchResourceAttr(
						"aws_neptune_cluster.test", "cluster_identifier", regexp.MustCompile("^tf-test-")),
				),
			},
			{
				ResourceName:            "aws_neptune_cluster.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"cluster_identifier_prefix", "skip_final_snapshot"},
			},
		},
	})
}

func TestAccAWSNeptuneCluster_takeFinalSnapshot(t *testing.T) {
	var v neptune.DBCluster
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterSnapshot(rInt),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterConfigWithFinalSnapshot(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists("aws_neptune_cluster.default", &v),
				),
			},
			{
				ResourceName:            "aws_neptune_cluster.default",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"final_snapshot_identifier", "skip_final_snapshot"},
			},
		},
	})
}

func TestAccAWSNeptuneCluster_updateTags(t *testing.T) {
	var v neptune.DBCluster
	ri := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterConfig(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists("aws_neptune_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_neptune_cluster.default", "tags.%", "1"),
				),
			},
			{
				Config: testAccAWSNeptuneClusterConfigUpdatedTags(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists("aws_neptune_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_neptune_cluster.default", "tags.%", "2"),
				),
			},
			{
				ResourceName:            "aws_neptune_cluster.default",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_final_snapshot"},
			},
		},
	})
}

func TestAccAWSNeptuneCluster_updateIamRoles(t *testing.T) {
	var v neptune.DBCluster
	ri := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterConfigIncludingIamRoles(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists("aws_neptune_cluster.default", &v),
				),
			},
			{
				Config: testAccAWSNeptuneClusterConfigAddIamRoles(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists("aws_neptune_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_neptune_cluster.default", "iam_roles.#", "2"),
				),
			},
			{
				Config: testAccAWSNeptuneClusterConfigRemoveIamRoles(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists("aws_neptune_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_neptune_cluster.default", "iam_roles.#", "1"),
				),
			},
			{
				ResourceName:            "aws_neptune_cluster.default",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_final_snapshot"},
			},
		},
	})
}

func TestAccAWSNeptuneCluster_kmsKey(t *testing.T) {
	var v neptune.DBCluster
	keyRegex := regexp.MustCompile("^arn:aws:kms:")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterConfig_kmsKey(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists("aws_neptune_cluster.default", &v),
					resource.TestMatchResourceAttr(
						"aws_neptune_cluster.default", "kms_key_arn", keyRegex),
				),
			},
			{
				ResourceName:            "aws_neptune_cluster.default",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_final_snapshot"},
			},
		},
	})
}

func TestAccAWSNeptuneCluster_encrypted(t *testing.T) {
	var v neptune.DBCluster

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterConfig_encrypted(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists("aws_neptune_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_neptune_cluster.default", "storage_encrypted", "true"),
				),
			},
			{
				ResourceName:            "aws_neptune_cluster.default",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_final_snapshot"},
			},
		},
	})
}

func TestAccAWSNeptuneCluster_backupsUpdate(t *testing.T) {
	var v neptune.DBCluster

	ri := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterConfig_backups(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists("aws_neptune_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_neptune_cluster.default", "preferred_backup_window", "07:00-09:00"),
					resource.TestCheckResourceAttr(
						"aws_neptune_cluster.default", "backup_retention_period", "5"),
					resource.TestCheckResourceAttr(
						"aws_neptune_cluster.default", "preferred_maintenance_window", "tue:04:00-tue:04:30"),
				),
			},
			{
				Config: testAccAWSNeptuneClusterConfig_backupsUpdate(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists("aws_neptune_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_neptune_cluster.default", "preferred_backup_window", "03:00-09:00"),
					resource.TestCheckResourceAttr(
						"aws_neptune_cluster.default", "backup_retention_period", "10"),
					resource.TestCheckResourceAttr(
						"aws_neptune_cluster.default", "preferred_maintenance_window", "wed:01:00-wed:01:30"),
				),
			},
			{
				ResourceName:            "aws_neptune_cluster.default",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "skip_final_snapshot"},
			},
		},
	})
}

func TestAccAWSNeptuneCluster_iamAuth(t *testing.T) {
	var v neptune.DBCluster

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterConfig_iamAuth(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterExists("aws_neptune_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_neptune_cluster.default", "iam_database_authentication_enabled", "true"),
				),
			},
			{
				ResourceName:            "aws_neptune_cluster.default",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_final_snapshot"},
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

func testAccCheckAWSNeptuneClusterSnapshot(rInt int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_neptune_cluster" {
				continue
			}

			// Try and delete the snapshot before we check for the cluster not found
			snapshot_identifier := fmt.Sprintf("tf-acctest-neptunecluster-snapshot-%d", rInt)

			awsClient := testAccProvider.Meta().(*AWSClient)
			conn := awsClient.neptuneconn

			log.Printf("[INFO] Deleting the Snapshot %s", snapshot_identifier)
			_, snapDeleteErr := conn.DeleteDBClusterSnapshot(
				&neptune.DeleteDBClusterSnapshotInput{
					DBClusterSnapshotIdentifier: aws.String(snapshot_identifier),
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

func testAccAWSNeptuneClusterConfig(n int) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster" "default" {
  cluster_identifier = "tf-neptune-cluster-%d"
  availability_zones = ["us-west-2a","us-west-2b","us-west-2c"]
  engine = "neptune"
  neptune_cluster_parameter_group_name = "default.neptune1"
  skip_final_snapshot = true
  tags = {
    Environment = "production"
  }
}`, n)
}

func testAccAWSNeptuneClusterConfig_namePrefix() string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier_prefix = "tf-test-"
  engine = "neptune"
  neptune_cluster_parameter_group_name = "default.neptune1"
  skip_final_snapshot = true
}
`)
}

func testAccAWSNeptuneClusterConfigWithFinalSnapshot(n int) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster" "default" {
  cluster_identifier = "tf-neptune-cluster-%d"
  availability_zones = ["us-west-2a","us-west-2b","us-west-2c"]
  neptune_cluster_parameter_group_name = "default.neptune1"
  final_snapshot_identifier = "tf-acctest-neptunecluster-snapshot-%d"
  tags = {
    Environment = "production"
  }
}`, n, n)
}

func testAccAWSNeptuneClusterConfigUpdatedTags(n int) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster" "default" {
  cluster_identifier = "tf-neptune-cluster-%d"
  availability_zones = ["us-west-2a","us-west-2b","us-west-2c"]
  neptune_cluster_parameter_group_name = "default.neptune1"
  skip_final_snapshot = true
  tags = {
    Environment = "production"
    AnotherTag = "test"
  }
}`, n)
}

func testAccAWSNeptuneClusterConfigIncludingIamRoles(n int) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "neptune_sample_role" {
  name = "neptune_sample_role_%d"
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
resource "aws_iam_role_policy" "neptune_policy" {
	name = "neptune_sample_role_policy_%d"
	role = "${aws_iam_role.neptune_sample_role.name}"
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
resource "aws_iam_role" "another_neptune_sample_role" {
  name = "another_neptune_sample_role_%d"
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
resource "aws_iam_role_policy" "another_neptune_policy" {
	name = "another_neptune_sample_role_policy_%d"
	role = "${aws_iam_role.another_neptune_sample_role.name}"
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
resource "aws_neptune_cluster" "default" {
  cluster_identifier = "tf-neptune-cluster-%d"
  availability_zones = ["us-west-2a","us-west-2b","us-west-2c"]
  neptune_cluster_parameter_group_name = "default.neptune1"
  skip_final_snapshot = true
  tags = {
    Environment = "production"
  }
  depends_on = ["aws_iam_role.another_neptune_sample_role", "aws_iam_role.neptune_sample_role"]

}`, n, n, n, n, n)
}

func testAccAWSNeptuneClusterConfigAddIamRoles(n int) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "neptune_sample_role" {
  name = "neptune_sample_role_%d"
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
resource "aws_iam_role_policy" "neptune_policy" {
	name = "neptune_sample_role_policy_%d"
	role = "${aws_iam_role.neptune_sample_role.name}"
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
resource "aws_iam_role" "another_neptune_sample_role" {
  name = "another_neptune_sample_role_%d"
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
resource "aws_iam_role_policy" "another_neptune_policy" {
	name = "another_neptune_sample_role_policy_%d"
	role = "${aws_iam_role.another_neptune_sample_role.name}"
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
resource "aws_neptune_cluster" "default" {
  cluster_identifier = "tf-neptune-cluster-%d"
  availability_zones = ["us-west-2a","us-west-2b","us-west-2c"]
  skip_final_snapshot = true
  iam_roles = ["${aws_iam_role.neptune_sample_role.arn}","${aws_iam_role.another_neptune_sample_role.arn}"]
  tags = {
    Environment = "production"
  }
  depends_on = ["aws_iam_role.another_neptune_sample_role", "aws_iam_role.neptune_sample_role"]

}`, n, n, n, n, n)
}

func testAccAWSNeptuneClusterConfigRemoveIamRoles(n int) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "another_neptune_sample_role" {
  name = "another_neptune_sample_role_%d"
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
resource "aws_iam_role_policy" "another_neptune_policy" {
	name = "another_neptune_sample_role_policy_%d"
	role = "${aws_iam_role.another_neptune_sample_role.name}"
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
resource "aws_neptune_cluster" "default" {
  cluster_identifier = "tf-neptune-cluster-%d"
  availability_zones = ["us-west-2a","us-west-2b","us-west-2c"]
  skip_final_snapshot = true
  iam_roles = ["${aws_iam_role.another_neptune_sample_role.arn}"]
  tags = {
    Environment = "production"
  }

  depends_on = ["aws_iam_role.another_neptune_sample_role"]
}`, n, n, n)
}

func testAccAWSNeptuneClusterConfig_kmsKey(n int) string {
	return fmt.Sprintf(`

 resource "aws_kms_key" "foo" {
     description = "Terraform acc test %d"
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

 resource "aws_neptune_cluster" "default" {
   cluster_identifier = "tf-neptune-cluster-%d"
   availability_zones = ["us-west-2a","us-west-2b","us-west-2c"]
   neptune_cluster_parameter_group_name = "default.neptune1"
   storage_encrypted = true
   kms_key_arn = "${aws_kms_key.foo.arn}"
   skip_final_snapshot = true
 }`, n, n)
}

func testAccAWSNeptuneClusterConfig_encrypted(n int) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster" "default" {
  cluster_identifier = "tf-neptune-cluster-%d"
  availability_zones = ["us-west-2a","us-west-2b","us-west-2c"]
  storage_encrypted = true
  skip_final_snapshot = true
}`, n)
}

func testAccAWSNeptuneClusterConfig_backups(n int) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster" "default" {
  cluster_identifier = "tf-neptune-cluster-%d"
  availability_zones = ["us-west-2a","us-west-2b","us-west-2c"]
  backup_retention_period = 5
  preferred_backup_window = "07:00-09:00"
  preferred_maintenance_window = "tue:04:00-tue:04:30"
  skip_final_snapshot = true
}`, n)
}

func testAccAWSNeptuneClusterConfig_backupsUpdate(n int) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster" "default" {
  cluster_identifier = "tf-neptune-cluster-%d"
  availability_zones = ["us-west-2a","us-west-2b","us-west-2c"]
  backup_retention_period = 10
  preferred_backup_window = "03:00-09:00"
  preferred_maintenance_window = "wed:01:00-wed:01:30"
  apply_immediately = true
  skip_final_snapshot = true
}`, n)
}

func testAccAWSNeptuneClusterConfig_iamAuth(n int) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster" "default" {
  cluster_identifier = "tf-neptune-cluster-%d"
  availability_zones = ["us-west-2a","us-west-2b","us-west-2c"]
  iam_database_authentication_enabled = true
  skip_final_snapshot = true
}`, n)
}
