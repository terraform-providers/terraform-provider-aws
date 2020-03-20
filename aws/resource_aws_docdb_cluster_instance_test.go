package aws

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/docdb"
)

func TestAccAWSDocDBClusterInstance_basic(t *testing.T) {
	var v docdb.DBInstance
	resourceName := "aws_docdb_cluster_instance.cluster_instances"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDocDBClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDocDBClusterInstanceConfig(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDocDBClusterInstanceExists(resourceName, &v),
					testAccCheckAWSDocDBClusterInstanceAttributes(&v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "rds", regexp.MustCompile(`db:.+`)),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "preferred_maintenance_window"),
					resource.TestCheckResourceAttrSet(resourceName, "preferred_backup_window"),
					resource.TestCheckResourceAttrSet(resourceName, "dbi_resource_id"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_version"),
					resource.TestCheckResourceAttrSet(resourceName, "ca_cert_identifier"),
					resource.TestCheckResourceAttr(resourceName, "engine", "docdb"),
				),
			},
			{
				Config: testAccAWSDocDBClusterInstanceConfigModified(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDocDBClusterInstanceExists(resourceName, &v),
					testAccCheckAWSDocDBClusterInstanceAttributes(&v),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
				),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
		},
	})
}

func TestAccAWSDocDBClusterInstance_az(t *testing.T) {
	var v docdb.DBInstance
	resourceName := "aws_docdb_cluster_instance.cluster_instances"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDocDBClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDocDBClusterInstanceConfig_az(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDocDBClusterInstanceExists(resourceName, &v),
					testAccCheckAWSDocDBClusterInstanceAttributes(&v),
					resource.TestMatchResourceAttr(resourceName, "availability_zone", regexp.MustCompile("^us-west-2[a-z]{1}$")),
				),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
		},
	})
}

func TestAccAWSDocDBClusterInstance_namePrefix(t *testing.T) {
	var v docdb.DBInstance
	resourceName := "aws_docdb_cluster_instance.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDocDBClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDocDBClusterInstanceConfig_namePrefix(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDocDBClusterInstanceExists(resourceName, &v),
					testAccCheckAWSDocDBClusterInstanceAttributes(&v),
					resource.TestCheckResourceAttr(resourceName, "db_subnet_group_name", fmt.Sprintf("tf-test-%d", rInt)),
					resource.TestMatchResourceAttr(resourceName, "identifier", regexp.MustCompile("^tf-cluster-instance-")),
				),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
		},
	})
}

func TestAccAWSDocDBClusterInstance_generatedName(t *testing.T) {
	var v docdb.DBInstance
	resourceName := "aws_docdb_cluster_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDocDBClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDocDBClusterInstanceConfig_generatedName(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDocDBClusterInstanceExists(resourceName, &v),
					testAccCheckAWSDocDBClusterInstanceAttributes(&v),
					resource.TestMatchResourceAttr(resourceName, "identifier", regexp.MustCompile("^tf-")),
				),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
		},
	})
}

func TestAccAWSDocDBClusterInstance_kmsKey(t *testing.T) {
	var v docdb.DBInstance
	resourceName := "aws_docdb_cluster_instance.cluster_instances"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDocDBClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDocDBClusterInstanceConfigKmsKey(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDocDBClusterInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", "aws_kms_key.foo", "arn"),
				),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
		},
	})
}

// https://github.com/hashicorp/terraform/issues/5350
func TestAccAWSDocDBClusterInstance_disappears(t *testing.T) {
	var v docdb.DBInstance
	resourceName := "aws_docdb_cluster_instance.cluster_instances"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDocDBClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDocDBClusterInstanceConfig(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDocDBClusterInstanceExists(resourceName, &v),
					testAccAWSDocDBClusterInstanceDisappears(&v),
				),
				// A non-empty plan is what we want. A crash is what we don't want. :)
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSDocDBClusterInstanceAttributes(v *docdb.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if *v.Engine != "docdb" {
			return fmt.Errorf("bad engine, expected \"docdb\": %#v", *v.Engine)
		}

		if !strings.HasPrefix(*v.DBClusterIdentifier, "tf-docdb-cluster") {
			return fmt.Errorf("Bad Cluster Identifier prefix:\nexpected: %s\ngot: %s", "tf-docdb-cluster", *v.DBClusterIdentifier)
		}

		return nil
	}
}

func testAccAWSDocDBClusterInstanceDisappears(v *docdb.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).docdbconn
		opts := &docdb.DeleteDBInstanceInput{
			DBInstanceIdentifier: v.DBInstanceIdentifier,
		}
		if _, err := conn.DeleteDBInstance(opts); err != nil {
			return err
		}
		return resource.Retry(40*time.Minute, func() *resource.RetryError {
			opts := &docdb.DescribeDBInstancesInput{
				DBInstanceIdentifier: v.DBInstanceIdentifier,
			}
			_, err := conn.DescribeDBInstances(opts)
			if err != nil {
				dbinstanceerr, ok := err.(awserr.Error)
				if ok && dbinstanceerr.Code() == "DBInstanceNotFound" {
					return nil
				}
				return resource.NonRetryableError(
					fmt.Errorf("Error retrieving DB Instances: %s", err))
			}
			return resource.RetryableError(fmt.Errorf(
				"Waiting for instance to be deleted: %v", v.DBInstanceIdentifier))
		})
	}
}

func testAccCheckAWSDocDBClusterInstanceExists(n string, v *docdb.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DB Instance ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).docdbconn
		resp, err := conn.DescribeDBInstances(&docdb.DescribeDBInstancesInput{
			DBInstanceIdentifier: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		for _, d := range resp.DBInstances {
			if *d.DBInstanceIdentifier == rs.Primary.ID {
				*v = *d
				return nil
			}
		}

		return fmt.Errorf("DB Cluster (%s) not found", rs.Primary.ID)
	}
}

// Add some random to the name, to avoid collision
func testAccAWSDocDBClusterInstanceConfig(n int) string {
	return fmt.Sprintf(`
resource "aws_docdb_cluster" "default" {
  cluster_identifier  = "tf-docdb-cluster-test-%d"
  availability_zones  = ["us-west-2a", "us-west-2b", "us-west-2c"]
  master_username     = "foo"
  master_password     = "mustbeeightcharaters"
  skip_final_snapshot = true
}

resource "aws_docdb_cluster_instance" "cluster_instances" {
  identifier         = "tf-cluster-instance-%d"
  cluster_identifier = "${aws_docdb_cluster.default.id}"
  instance_class     = "db.r4.large"
  promotion_tier     = "3"
}
`, n, n)
}

func testAccAWSDocDBClusterInstanceConfigModified(n int) string {
	return fmt.Sprintf(`
resource "aws_docdb_cluster" "default" {
  cluster_identifier  = "tf-docdb-cluster-test-%d"
  availability_zones  = ["us-west-2a", "us-west-2b", "us-west-2c"]
  master_username     = "foo"
  master_password     = "mustbeeightcharaters"
  skip_final_snapshot = true
}

resource "aws_docdb_cluster_instance" "cluster_instances" {
  identifier                 = "tf-cluster-instance-%d"
  cluster_identifier         = "${aws_docdb_cluster.default.id}"
  instance_class             = "db.r4.large"
  auto_minor_version_upgrade = false
  promotion_tier             = "3"
}
`, n, n)
}

func testAccAWSDocDBClusterInstanceConfig_az(n int) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {}

resource "aws_docdb_cluster" "default" {
  cluster_identifier  = "tf-docdb-cluster-test-%d"
  availability_zones  = ["${data.aws_availability_zones.available.names[0]}", "${data.aws_availability_zones.available.names[1]}", "${data.aws_availability_zones.available.names[2]}"]
  master_username     = "foo"
  master_password     = "mustbeeightcharaters"
  skip_final_snapshot = true
}

resource "aws_docdb_cluster_instance" "cluster_instances" {
  identifier         = "tf-cluster-instance-%d"
  cluster_identifier = "${aws_docdb_cluster.default.id}"
  instance_class     = "db.r4.large"
  promotion_tier     = "3"
  availability_zone  = "${data.aws_availability_zones.available.names[0]}"
}
`, n, n)
}

func testAccAWSDocDBClusterInstanceConfig_namePrefix(n int) string {
	return fmt.Sprintf(`
resource "aws_docdb_cluster_instance" "test" {
  identifier_prefix  = "tf-cluster-instance-"
  cluster_identifier = "${aws_docdb_cluster.test.id}"
  instance_class     = "db.r4.large"
}

resource "aws_docdb_cluster" "test" {
  cluster_identifier   = "tf-docdb-cluster-%d"
  master_username      = "root"
  master_password      = "password"
  db_subnet_group_name = "${aws_docdb_subnet_group.test.name}"
  skip_final_snapshot  = true
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-docdb-cluster-instance-name-prefix"
  }
}

resource "aws_subnet" "a" {
  vpc_id            = "${aws_vpc.test.id}"
  cidr_block        = "10.0.0.0/24"
  availability_zone = "us-west-2a"

  tags = {
    Name = "tf-acc-docdb-cluster-instance-name-prefix-a"
  }
}

resource "aws_subnet" "b" {
  vpc_id            = "${aws_vpc.test.id}"
  cidr_block        = "10.0.1.0/24"
  availability_zone = "us-west-2b"

  tags = {
    Name = "tf-acc-docdb-cluster-instance-name-prefix-b"
  }
}

resource "aws_docdb_subnet_group" "test" {
  name       = "tf-test-%d"
  subnet_ids = ["${aws_subnet.a.id}", "${aws_subnet.b.id}"]
}
`, n, n)
}

func testAccAWSDocDBClusterInstanceConfig_generatedName(n int) string {
	return fmt.Sprintf(`
resource "aws_docdb_cluster_instance" "test" {
  cluster_identifier = "${aws_docdb_cluster.test.id}"
  instance_class     = "db.r4.large"
}

resource "aws_docdb_cluster" "test" {
  cluster_identifier   = "tf-docdb-cluster-%d"
  master_username      = "root"
  master_password      = "password"
  db_subnet_group_name = "${aws_docdb_subnet_group.test.name}"
  skip_final_snapshot  = true
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-docdb-cluster-instance-generated-name"
  }
}

resource "aws_subnet" "a" {
  vpc_id            = "${aws_vpc.test.id}"
  cidr_block        = "10.0.0.0/24"
  availability_zone = "us-west-2a"

  tags = {
    Name = "tf-acc-docdb-cluster-instance-generated-name-a"
  }
}

resource "aws_subnet" "b" {
  vpc_id            = "${aws_vpc.test.id}"
  cidr_block        = "10.0.1.0/24"
  availability_zone = "us-west-2b"

  tags = {
    Name = "tf-acc-docdb-cluster-instance-generated-name-b"
  }
}

resource "aws_docdb_subnet_group" "test" {
  name       = "tf-test-%d"
  subnet_ids = ["${aws_subnet.a.id}", "${aws_subnet.b.id}"]
}
`, n, n)
}

func testAccAWSDocDBClusterInstanceConfigKmsKey(n int) string {
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

resource "aws_docdb_cluster" "default" {
  cluster_identifier  = "tf-docdb-cluster-test-%d"
  availability_zones  = ["us-west-2a", "us-west-2b", "us-west-2c"]
  master_username     = "foo"
  master_password     = "mustbeeightcharaters"
  storage_encrypted   = true
  kms_key_id          = "${aws_kms_key.foo.arn}"
  skip_final_snapshot = true
}

resource "aws_docdb_cluster_instance" "cluster_instances" {
  identifier         = "tf-cluster-instance-%d"
  cluster_identifier = "${aws_docdb_cluster.default.id}"
  instance_class     = "db.r4.large"
}
`, n, n, n)
}
