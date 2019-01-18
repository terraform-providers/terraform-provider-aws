package aws

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
)

func TestAccAWSRDSClusterInstance_importBasic(t *testing.T) {
	resourceName := "aws_rds_cluster_instance.cluster_instances"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfig(acctest.RandInt()),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_basic(t *testing.T) {
	var v rds.DBInstance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfig(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists("aws_rds_cluster_instance.cluster_instances", &v),
					testAccCheckAWSDBClusterInstanceAttributes(&v),
					resource.TestMatchResourceAttr("aws_rds_cluster_instance.cluster_instances", "arn", regexp.MustCompile(`^arn:[^:]+:rds:[^:]+:[^:]+:db:.+`)),
					resource.TestCheckResourceAttr("aws_rds_cluster_instance.cluster_instances", "auto_minor_version_upgrade", "true"),
					resource.TestCheckResourceAttr("aws_rds_cluster_instance.cluster_instances", "copy_tags_to_snapshot", "false"),
					resource.TestCheckResourceAttrSet("aws_rds_cluster_instance.cluster_instances", "preferred_maintenance_window"),
					resource.TestCheckResourceAttrSet("aws_rds_cluster_instance.cluster_instances", "preferred_backup_window"),
					resource.TestCheckResourceAttrSet("aws_rds_cluster_instance.cluster_instances", "dbi_resource_id"),
					resource.TestCheckResourceAttrSet("aws_rds_cluster_instance.cluster_instances", "availability_zone"),
					resource.TestCheckResourceAttrSet("aws_rds_cluster_instance.cluster_instances", "engine_version"),
					resource.TestCheckResourceAttr("aws_rds_cluster_instance.cluster_instances", "engine", "aurora"),
				),
			},
			{
				Config: testAccAWSClusterInstanceConfigModified(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists("aws_rds_cluster_instance.cluster_instances", &v),
					testAccCheckAWSDBClusterInstanceAttributes(&v),
					resource.TestCheckResourceAttr("aws_rds_cluster_instance.cluster_instances", "auto_minor_version_upgrade", "false"),
				),
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_az(t *testing.T) {
	var v rds.DBInstance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfig_az(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists("aws_rds_cluster_instance.cluster_instances", &v),
					testAccCheckAWSDBClusterInstanceAttributes(&v),
					resource.TestMatchResourceAttr("aws_rds_cluster_instance.cluster_instances", "availability_zone", regexp.MustCompile("^us-west-2[a-z]{1}$")),
				),
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_namePrefix(t *testing.T) {
	var v rds.DBInstance
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfig_namePrefix(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists("aws_rds_cluster_instance.test", &v),
					testAccCheckAWSDBClusterInstanceAttributes(&v),
					resource.TestCheckResourceAttr(
						"aws_rds_cluster_instance.test", "db_subnet_group_name", fmt.Sprintf("tf-test-%d", rInt)),
					resource.TestMatchResourceAttr(
						"aws_rds_cluster_instance.test", "identifier", regexp.MustCompile("^tf-cluster-instance-")),
				),
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_generatedName(t *testing.T) {
	var v rds.DBInstance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfig_generatedName(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists("aws_rds_cluster_instance.test", &v),
					testAccCheckAWSDBClusterInstanceAttributes(&v),
					resource.TestMatchResourceAttr(
						"aws_rds_cluster_instance.test", "identifier", regexp.MustCompile("^tf-")),
				),
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_kmsKey(t *testing.T) {
	var v rds.DBInstance
	keyRegex := regexp.MustCompile("^arn:aws:kms:")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfigKmsKey(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists("aws_rds_cluster_instance.cluster_instances", &v),
					resource.TestMatchResourceAttr(
						"aws_rds_cluster_instance.cluster_instances", "kms_key_id", keyRegex),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform/issues/5350
func TestAccAWSRDSClusterInstance_disappears(t *testing.T) {
	var v rds.DBInstance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfig(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists("aws_rds_cluster_instance.cluster_instances", &v),
					testAccAWSClusterInstanceDisappears(&v),
				),
				// A non-empty plan is what we want. A crash is what we don't want. :)
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_PubliclyAccessible(t *testing.T) {
	var dbInstance rds.DBInstance
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_rds_cluster_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSClusterInstanceConfig_PubliclyAccessible(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				Config: testAccAWSRDSClusterInstanceConfig_PubliclyAccessible(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "false"),
				),
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_CopyTagsToSnapshot(t *testing.T) {
	var dbInstance rds.DBInstance
	rNameSuffix := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfig_CopyTagsToSnapshot(rNameSuffix, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists("aws_rds_cluster_instance.cluster_instances", &dbInstance),
					resource.TestCheckResourceAttr(
						"aws_rds_cluster_instance.cluster_instances", "copy_tags_to_snapshot", "true"),
				),
			},
			{
				Config: testAccAWSClusterInstanceConfig_CopyTagsToSnapshot(rNameSuffix, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists("aws_rds_cluster_instance.cluster_instances", &dbInstance),
					resource.TestCheckResourceAttr(
						"aws_rds_cluster_instance.cluster_instances", "copy_tags_to_snapshot", "false"),
				),
			},
		},
	})
}

func testAccCheckAWSDBClusterInstanceAttributes(v *rds.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if *v.Engine != "aurora" && *v.Engine != "aurora-postgresql" && *v.Engine != "aurora-mysql" {
			return fmt.Errorf("bad engine, expected \"aurora\", \"aurora-mysql\" or \"aurora-postgresql\": %#v", *v.Engine)
		}

		if !strings.HasPrefix(*v.DBClusterIdentifier, "tf-aurora-cluster") {
			return fmt.Errorf("Bad Cluster Identifier prefix:\nexpected: %s\ngot: %s", "tf-aurora-cluster", *v.DBClusterIdentifier)
		}

		return nil
	}
}

func testAccAWSClusterInstanceDisappears(v *rds.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).rdsconn
		opts := &rds.DeleteDBInstanceInput{
			DBInstanceIdentifier: v.DBInstanceIdentifier,
		}
		if _, err := conn.DeleteDBInstance(opts); err != nil {
			return err
		}
		return resource.Retry(40*time.Minute, func() *resource.RetryError {
			opts := &rds.DescribeDBInstancesInput{
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

func testAccCheckAWSClusterInstanceExists(n string, v *rds.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DB Instance ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).rdsconn
		resp, err := conn.DescribeDBInstances(&rds.DescribeDBInstancesInput{
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

func TestAccAWSRDSClusterInstance_withInstanceEnhancedMonitor(t *testing.T) {
	var v rds.DBInstance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceEnhancedMonitor(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists("aws_rds_cluster_instance.cluster_instances", &v),
					testAccCheckAWSDBClusterInstanceAttributes(&v),
				),
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_withInstancePerformanceInsights(t *testing.T) {
	var v rds.DBInstance
	keyRegex := regexp.MustCompile("^arn:aws:kms:")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstancePerformanceInsights(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists("aws_rds_cluster_instance.cluster_instances", &v),
					testAccCheckAWSDBClusterInstanceAttributes(&v),
					resource.TestCheckResourceAttr(
						"aws_rds_cluster_instance.cluster_instances", "performance_insights_enabled", "true"),
					resource.TestMatchResourceAttr(
						"aws_rds_cluster_instance.cluster_instances", "performance_insights_kms_key_id", keyRegex),
				),
			},
		},
	})
}

// Add some random to the name, to avoid collision
func testAccAWSClusterInstanceConfig(n int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "default" {
  cluster_identifier = "tf-aurora-cluster-test-%d"
  availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]
  database_name      = "mydb"
  master_username    = "foo"
  master_password    = "mustbeeightcharaters"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "cluster_instances" {
  identifier              = "tf-cluster-instance-%d"
  cluster_identifier      = "${aws_rds_cluster.default.id}"
  instance_class          = "db.t2.small"
  db_parameter_group_name = "${aws_db_parameter_group.bar.name}"
  promotion_tier          = "3"
}

resource "aws_db_parameter_group" "bar" {
  name   = "tfcluster-test-group-%d"
  family = "aurora5.6"

  parameter {
    name         = "back_log"
    value        = "32767"
    apply_method = "pending-reboot"
  }

  tags = {
    foo = "bar"
  }
}
`, n, n, n)
}

func testAccAWSClusterInstanceConfigModified(n int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "default" {
  cluster_identifier = "tf-aurora-cluster-test-%d"
  availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]
  database_name      = "mydb"
  master_username    = "foo"
  master_password    = "mustbeeightcharaters"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "cluster_instances" {
  identifier                 = "tf-cluster-instance-%d"
  cluster_identifier         = "${aws_rds_cluster.default.id}"
  instance_class             = "db.t2.small"
  db_parameter_group_name    = "${aws_db_parameter_group.bar.name}"
  auto_minor_version_upgrade = false
  promotion_tier             = "3"
}

resource "aws_db_parameter_group" "bar" {
  name   = "tfcluster-test-group-%d"
  family = "aurora5.6"

  parameter {
    name         = "back_log"
    value        = "32767"
    apply_method = "pending-reboot"
  }

  tags = {
    foo = "bar"
  }
}
`, n, n, n)
}

func testAccAWSClusterInstanceConfig_az(n int) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {}

resource "aws_rds_cluster" "default" {
  cluster_identifier = "tf-aurora-cluster-test-%d"
  availability_zones = ["${data.aws_availability_zones.available.names[0]}", "${data.aws_availability_zones.available.names[1]}", "${data.aws_availability_zones.available.names[2]}"]
  database_name      = "mydb"
  master_username    = "foo"
  master_password    = "mustbeeightcharaters"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "cluster_instances" {
  identifier              = "tf-cluster-instance-%d"
  cluster_identifier      = "${aws_rds_cluster.default.id}"
  instance_class          = "db.t2.small"
  db_parameter_group_name = "${aws_db_parameter_group.bar.name}"
  promotion_tier          = "3"
  availability_zone       = "${data.aws_availability_zones.available.names[0]}"
}

resource "aws_db_parameter_group" "bar" {
  name   = "tfcluster-test-group-%d"
  family = "aurora5.6"

  parameter {
    name         = "back_log"
    value        = "32767"
    apply_method = "pending-reboot"
  }

  tags = {
    foo = "bar"
  }
}
`, n, n, n)
}

func testAccAWSClusterInstanceConfig_namePrefix(n int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster_instance" "test" {
  identifier_prefix = "tf-cluster-instance-"
  cluster_identifier = "${aws_rds_cluster.test.id}"
  instance_class = "db.t2.small"
}

resource "aws_rds_cluster" "test" {
  cluster_identifier = "tf-aurora-cluster-%d"
  master_username = "root"
  master_password = "password"
  db_subnet_group_name = "${aws_db_subnet_group.test.name}"
  skip_final_snapshot = true
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
	tags = {
		Name = "terraform-testacc-rds-cluster-instance-name-prefix"
	}
}

resource "aws_subnet" "a" {
  vpc_id = "${aws_vpc.test.id}"
  cidr_block = "10.0.0.0/24"
  availability_zone = "us-west-2a"
  tags = {
    Name = "tf-acc-rds-cluster-instance-name-prefix-a"
  }
}

resource "aws_subnet" "b" {
  vpc_id = "${aws_vpc.test.id}"
  cidr_block = "10.0.1.0/24"
  availability_zone = "us-west-2b"
  tags = {
    Name = "tf-acc-rds-cluster-instance-name-prefix-b"
  }
}

resource "aws_db_subnet_group" "test" {
  name = "tf-test-%d"
  subnet_ids = ["${aws_subnet.a.id}", "${aws_subnet.b.id}"]
}
`, n, n)
}

func testAccAWSClusterInstanceConfig_generatedName(n int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster_instance" "test" {
  cluster_identifier = "${aws_rds_cluster.test.id}"
  instance_class = "db.t2.small"
}

resource "aws_rds_cluster" "test" {
  cluster_identifier = "tf-aurora-cluster-%d"
  master_username = "root"
  master_password = "password"
  db_subnet_group_name = "${aws_db_subnet_group.test.name}"
  skip_final_snapshot = true
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
	tags = {
		Name = "terraform-testacc-rds-cluster-instance-generated-name"
	}
}

resource "aws_subnet" "a" {
  vpc_id = "${aws_vpc.test.id}"
  cidr_block = "10.0.0.0/24"
  availability_zone = "us-west-2a"
  tags = {
    Name = "tf-acc-rds-cluster-instance-generated-name-a"
  }
}

resource "aws_subnet" "b" {
  vpc_id = "${aws_vpc.test.id}"
  cidr_block = "10.0.1.0/24"
  availability_zone = "us-west-2b"
  tags = {
    Name = "tf-acc-rds-cluster-instance-generated-name-b"
  }
}

resource "aws_db_subnet_group" "test" {
  name = "tf-test-%d"
  subnet_ids = ["${aws_subnet.a.id}", "${aws_subnet.b.id}"]
}
`, n, n)
}

func testAccAWSClusterInstanceConfigKmsKey(n int) string {
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

resource "aws_rds_cluster" "default" {
  cluster_identifier = "tf-aurora-cluster-test-%d"
  availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]
  database_name      = "mydb"
  master_username    = "foo"
  master_password    = "mustbeeightcharaters"
  storage_encrypted = true
  kms_key_id = "${aws_kms_key.foo.arn}"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "cluster_instances" {
  identifier              = "tf-cluster-instance-%d"
  cluster_identifier      = "${aws_rds_cluster.default.id}"
  instance_class          = "db.t2.small"
  db_parameter_group_name = "${aws_db_parameter_group.bar.name}"
}

resource "aws_db_parameter_group" "bar" {
  name   = "tfcluster-test-group-%d"
  family = "aurora5.6"

  parameter {
    name         = "back_log"
    value        = "32767"
    apply_method = "pending-reboot"
  }

  tags = {
    foo = "bar"
  }
}
`, n, n, n, n)
}

func testAccAWSClusterInstanceEnhancedMonitor(n int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "default" {
  cluster_identifier = "tf-aurora-cluster-test-%d"
  availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]
  database_name      = "mydb"
  master_username    = "foo"
  master_password    = "mustbeeightcharaters"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "cluster_instances" {
  identifier              = "tf-cluster-instance-%d"
  cluster_identifier      = "${aws_rds_cluster.default.id}"
  instance_class          = "db.t2.small"
  db_parameter_group_name = "${aws_db_parameter_group.bar.name}"
  monitoring_interval     = "60"
  monitoring_role_arn     = "${aws_iam_role.tf_enhanced_monitor_role.arn}"
}

resource "aws_iam_role" "tf_enhanced_monitor_role" {
    name = "tf_enhanced_monitor_role-%d"
    assume_role_policy = <<EOF
{
            "Version": "2012-10-17",
            "Statement": [
                {
                    "Action": "sts:AssumeRole",
                    "Principal": {
                        "Service": "monitoring.rds.amazonaws.com"
                    },
                    "Effect": "Allow",
                    "Sid": ""
                }
            ]
   }
EOF
}

resource "aws_iam_policy_attachment" "rds_m_attach" {
    name = "tf-enhanced-monitoring-attachment-%d"
    roles = ["${aws_iam_role.tf_enhanced_monitor_role.name}"]
    policy_arn = "${aws_iam_policy.test.arn}"
}

resource "aws_iam_policy" "test" {
  name   = "tf-enhanced-monitoring-policy-%d"
  policy = <<POLICY
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "EnableCreationAndManagementOfRDSCloudwatchLogGroups",
            "Effect": "Allow",
            "Action": [
                "logs:CreateLogGroup",
                "logs:PutRetentionPolicy"
            ],
            "Resource": [
                "arn:aws:logs:*:*:log-group:RDS*"
            ]
        },
        {
            "Sid": "EnableCreationAndManagementOfRDSCloudwatchLogStreams",
            "Effect": "Allow",
            "Action": [
                "logs:CreateLogStream",
                "logs:PutLogEvents",
                "logs:DescribeLogStreams",
                "logs:GetLogEvents"
            ],
            "Resource": [
                "arn:aws:logs:*:*:log-group:RDS*:log-stream:*"
            ]
        }
    ]
}
POLICY
}

resource "aws_db_parameter_group" "bar" {
  name   = "tfcluster-test-group-%d"
  family = "aurora5.6"

  parameter {
    name         = "back_log"
    value        = "32767"
    apply_method = "pending-reboot"
  }

  tags = {
    foo = "bar"
  }
}
`, n, n, n, n, n, n)
}

func testAccAWSClusterInstancePerformanceInsights(n int) string {
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

resource "aws_rds_cluster" "default" {
  engine             = "aurora-postgresql"
  cluster_identifier = "tf-aurora-cluster-test-%d"
  availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]
  database_name      = "mydb"
  master_username    = "foo"
  master_password    = "mustbeeightcharaters"
  storage_encrypted = true
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "cluster_instances" {
  engine                  = "aurora-postgresql"
  identifier              = "tf-cluster-instance-%d"
  cluster_identifier      = "${aws_rds_cluster.default.id}"
  instance_class          = "db.r4.large"
  db_parameter_group_name = "${aws_db_parameter_group.bar.name}"
  performance_insights_enabled = true
  performance_insights_kms_key_id = "${aws_kms_key.foo.arn}"
}

resource "aws_db_parameter_group" "bar" {
  name   = "tfcluster-test-group-%d"
  family = "aurora-postgresql9.6"

  parameter {
    name  = "authentication_timeout"
    value = "10"
  }

  tags = {
    foo = "bar"
  }
}
`, n, n, n, n)
}

func testAccAWSRDSClusterInstanceConfig_PubliclyAccessible(rName string, publiclyAccessible bool) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %q
  master_username     = "foo"
  master_password     = "mustbeeightcharaters"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "test" {
  apply_immediately       = true
  cluster_identifier      = "${aws_rds_cluster.test.id}"
  identifier              = %q
  instance_class          = "db.t2.small"
  publicly_accessible     = %t
}
`, rName, rName, publiclyAccessible)
}

func testAccAWSClusterInstanceConfig_CopyTagsToSnapshot(n int, f bool) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "default" {
  cluster_identifier = "tf-aurora-cluster-test-%d"
  availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]
  database_name      = "mydb"
  master_username    = "foo"
  master_password    = "mustbeeightcharaters"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "cluster_instances" {
  identifier              = "tf-cluster-instance-%d"
  cluster_identifier      = "${aws_rds_cluster.default.id}"
  instance_class          = "db.t2.small"
  promotion_tier          = "3"
  copy_tags_to_snapshot 	= %t
}
`, n, n, f)
}
