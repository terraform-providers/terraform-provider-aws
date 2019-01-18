package aws

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSNeptuneClusterInstance_basic(t *testing.T) {
	var v neptune.DBInstance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterInstanceConfig(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterInstanceExists("aws_neptune_cluster_instance.cluster_instances", &v),
					testAccCheckAWSNeptuneClusterInstanceAttributes(&v),
					resource.TestCheckResourceAttrSet("aws_neptune_cluster_instance.cluster_instances", "address"),
					resource.TestMatchResourceAttr("aws_neptune_cluster_instance.cluster_instances", "arn", regexp.MustCompile(`^arn:[^:]+:rds:[^:]+:[^:]+:db:.+`)),
					resource.TestCheckResourceAttr("aws_neptune_cluster_instance.cluster_instances", "auto_minor_version_upgrade", "true"),
					resource.TestMatchResourceAttr("aws_neptune_cluster_instance.cluster_instances", "availability_zone", regexp.MustCompile(fmt.Sprintf("^%s", testAccGetRegion()))),
					resource.TestCheckResourceAttrSet("aws_neptune_cluster_instance.cluster_instances", "cluster_identifier"),
					resource.TestCheckResourceAttrSet("aws_neptune_cluster_instance.cluster_instances", "dbi_resource_id"),
					resource.TestMatchResourceAttr("aws_neptune_cluster_instance.cluster_instances", "endpoint", regexp.MustCompile(`:8182$`)),
					resource.TestCheckResourceAttr("aws_neptune_cluster_instance.cluster_instances", "engine", "neptune"),
					resource.TestCheckResourceAttrSet("aws_neptune_cluster_instance.cluster_instances", "engine_version"),
					resource.TestCheckResourceAttrSet("aws_neptune_cluster_instance.cluster_instances", "identifier"),
					resource.TestCheckResourceAttr("aws_neptune_cluster_instance.cluster_instances", "instance_class", "db.r4.large"),
					resource.TestCheckResourceAttr("aws_neptune_cluster_instance.cluster_instances", "kms_key_arn", ""),
					resource.TestMatchResourceAttr("aws_neptune_cluster_instance.cluster_instances", "neptune_parameter_group_name", regexp.MustCompile(`^tf-cluster-test-group-`)),
					resource.TestCheckResourceAttr("aws_neptune_cluster_instance.cluster_instances", "neptune_subnet_group_name", "default"),
					resource.TestCheckResourceAttr("aws_neptune_cluster_instance.cluster_instances", "port", "8182"),
					resource.TestCheckResourceAttrSet("aws_neptune_cluster_instance.cluster_instances", "preferred_backup_window"),
					resource.TestCheckResourceAttrSet("aws_neptune_cluster_instance.cluster_instances", "preferred_maintenance_window"),
					resource.TestCheckResourceAttr("aws_neptune_cluster_instance.cluster_instances", "promotion_tier", "3"),
					resource.TestCheckResourceAttr("aws_neptune_cluster_instance.cluster_instances", "publicly_accessible", "false"),
					resource.TestCheckResourceAttr("aws_neptune_cluster_instance.cluster_instances", "storage_encrypted", "false"),
					resource.TestCheckResourceAttr("aws_neptune_cluster_instance.cluster_instances", "tags.%", "0"),
					resource.TestCheckResourceAttr("aws_neptune_cluster_instance.cluster_instances", "writer", "true"),
				),
			},
			{
				Config: testAccAWSNeptuneClusterInstanceConfigModified(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterInstanceExists("aws_neptune_cluster_instance.cluster_instances", &v),
					testAccCheckAWSNeptuneClusterInstanceAttributes(&v),
					resource.TestCheckResourceAttr("aws_neptune_cluster_instance.cluster_instances", "auto_minor_version_upgrade", "false"),
				),
			},
		},
	})
}

func TestAccAWSNeptuneClusterInstance_withaz(t *testing.T) {
	var v neptune.DBInstance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterInstanceConfig_az(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterInstanceExists("aws_neptune_cluster_instance.cluster_instances", &v),
					testAccCheckAWSNeptuneClusterInstanceAttributes(&v),
					resource.TestMatchResourceAttr("aws_neptune_cluster_instance.cluster_instances", "availability_zone", regexp.MustCompile("^us-west-2[a-z]{1}$")),
				),
			},
		},
	})
}

func TestAccAWSNeptuneClusterInstance_namePrefix(t *testing.T) {
	var v neptune.DBInstance
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterInstanceConfig_namePrefix(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterInstanceExists("aws_neptune_cluster_instance.test", &v),
					testAccCheckAWSNeptuneClusterInstanceAttributes(&v),
					resource.TestMatchResourceAttr(
						"aws_neptune_cluster_instance.test", "identifier", regexp.MustCompile("^tf-cluster-instance-")),
				),
			},
		},
	})
}

func TestAccAWSNeptuneClusterInstance_withSubnetGroup(t *testing.T) {
	var v neptune.DBInstance
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterInstanceConfig_withSubnetGroup(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterInstanceExists("aws_neptune_cluster_instance.test", &v),
					testAccCheckAWSNeptuneClusterInstanceAttributes(&v),
					resource.TestCheckResourceAttr(
						"aws_neptune_cluster_instance.test", "neptune_subnet_group_name", fmt.Sprintf("tf-test-%d", rInt)),
				),
			},
		},
	})
}

func TestAccAWSNeptuneClusterInstance_generatedName(t *testing.T) {
	var v neptune.DBInstance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterInstanceConfig_generatedName(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterInstanceExists("aws_neptune_cluster_instance.test", &v),
					testAccCheckAWSNeptuneClusterInstanceAttributes(&v),
					resource.TestMatchResourceAttr(
						"aws_neptune_cluster_instance.test", "identifier", regexp.MustCompile("^tf-")),
				),
			},
		},
	})
}

func TestAccAWSNeptuneClusterInstance_kmsKey(t *testing.T) {
	var v neptune.DBInstance
	keyRegex := regexp.MustCompile("^arn:aws:kms:")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterInstanceConfigKmsKey(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterInstanceExists("aws_neptune_cluster_instance.cluster_instances", &v),
					resource.TestMatchResourceAttr(
						"aws_neptune_cluster_instance.cluster_instances", "kms_key_arn", keyRegex),
				),
			},
		},
	})
}

func testAccCheckAWSNeptuneClusterInstanceExists(n string, v *neptune.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Instance not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Neptune Instance ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).neptuneconn
		resp, err := conn.DescribeDBInstances(&neptune.DescribeDBInstancesInput{
			DBInstanceIdentifier: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		for _, d := range resp.DBInstances {
			if aws.StringValue(d.DBInstanceIdentifier) == rs.Primary.ID {
				*v = *d
				return nil
			}
		}

		return fmt.Errorf("Neptune Cluster (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckAWSNeptuneClusterInstanceAttributes(v *neptune.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if aws.StringValue(v.Engine) != "neptune" {
			return fmt.Errorf("Incorrect engine, expected \"neptune\": %#v", aws.StringValue(v.Engine))
		}

		if !strings.HasPrefix(aws.StringValue(v.DBClusterIdentifier), "tf-neptune-cluster") {
			return fmt.Errorf("Incorrect Cluster Identifier prefix:\nexpected: %s\ngot: %s", "tf-neptune-cluster", aws.StringValue(v.DBClusterIdentifier))
		}

		return nil
	}
}

func testAccAWSNeptuneClusterInstanceConfig(n int) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster" "default" {
  cluster_identifier 	= "tf-neptune-cluster-test-%d"
  availability_zones 	= ["us-west-2a", "us-west-2b", "us-west-2c"]
  skip_final_snapshot 	= true
}

resource "aws_neptune_cluster_instance" "cluster_instances" {
  identifier              		= "tf-cluster-instance-%d"
  cluster_identifier      		= "${aws_neptune_cluster.default.id}"
  instance_class          		= "db.r4.large"
  neptune_parameter_group_name 	= "${aws_neptune_parameter_group.bar.name}"
  promotion_tier          		= "3"
}

resource "aws_neptune_parameter_group" "bar" {
  name   = "tf-cluster-test-group-%d"
  family = "neptune1"

  parameter {
    name         = "neptune_query_timeout"
    value        = "25"
  }

  tags = {
    foo = "bar"
  }
}
`, n, n, n)
}

func testAccAWSNeptuneClusterInstanceConfigModified(n int) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster" "default" {
  cluster_identifier 	= "tf-neptune-cluster-test-%d"
  availability_zones 	= ["us-west-2a", "us-west-2b", "us-west-2c"]
  skip_final_snapshot 	= true
}

resource "aws_neptune_cluster_instance" "cluster_instances" {
  identifier              		= "tf-cluster-instance-%d"
  cluster_identifier      		= "${aws_neptune_cluster.default.id}"
  instance_class          		= "db.r4.large"
  neptune_parameter_group_name 	= "${aws_neptune_parameter_group.bar.name}"
  auto_minor_version_upgrade 	= false
  promotion_tier          		= "3"
}

resource "aws_neptune_parameter_group" "bar" {
  name   = "tf-cluster-test-group-%d"
  family = "neptune1"

  parameter {
    name         = "neptune_query_timeout"
    value        = "25"
  }

  tags = {
    foo = "bar"
  }
}
`, n, n, n)
}

func testAccAWSNeptuneClusterInstanceConfig_az(n int) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {}

resource "aws_neptune_cluster" "default" {
  cluster_identifier 	= "tf-neptune-cluster-test-%d"
  availability_zones 	= ["${data.aws_availability_zones.available.names[0]}", "${data.aws_availability_zones.available.names[1]}"]
  skip_final_snapshot 	= true
}

resource "aws_neptune_cluster_instance" "cluster_instances" {
  identifier              		= "tf-cluster-instance-%d"
  cluster_identifier      		= "${aws_neptune_cluster.default.id}"
  instance_class          		= "db.r4.large"
  neptune_parameter_group_name 	= "${aws_neptune_parameter_group.bar.name}"
  promotion_tier          		= "3"
  availability_zone       		= "${data.aws_availability_zones.available.names[0]}"
}

resource "aws_neptune_parameter_group" "bar" {
  name   = "tf-cluster-test-group-%d"
  family = "neptune1"

  parameter {
    name         = "neptune_query_timeout"
    value        = "25"
  }

  tags = {
    foo = "bar"
  }
}
`, n, n, n)
}

func testAccAWSNeptuneClusterInstanceConfig_withSubnetGroup(n int) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster_instance" "test" {
  identifier = "tf-cluster-instance-%d"
  cluster_identifier = "${aws_neptune_cluster.test.id}"
  instance_class = "db.r4.large"
}

resource "aws_neptune_cluster" "test" {
  cluster_identifier = "tf-neptune-cluster-%d"
  neptune_subnet_group_name = "${aws_neptune_subnet_group.test.name}"
  skip_final_snapshot = true
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
	tags = {
		Name = "terraform-testacc-neptune-cluster-instance-name-prefix"
	}
}

resource "aws_subnet" "a" {
  vpc_id = "${aws_vpc.test.id}"
  cidr_block = "10.0.0.0/24"
  availability_zone = "us-west-2a"
  tags = {
    Name = "tf-acc-neptune-cluster-instance-name-prefix-a"
  }
}

resource "aws_subnet" "b" {
  vpc_id = "${aws_vpc.test.id}"
  cidr_block = "10.0.1.0/24"
  availability_zone = "us-west-2b"
  tags = {
    Name = "tf-acc-neptune-cluster-instance-name-prefix-b"
  }
}

resource "aws_neptune_subnet_group" "test" {
  name = "tf-test-%d"
  subnet_ids = ["${aws_subnet.a.id}", "${aws_subnet.b.id}"]
}
`, n, n, n)
}

func testAccAWSNeptuneClusterInstanceConfig_namePrefix(n int) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster_instance" "test" {
  identifier_prefix = "tf-cluster-instance-"
  cluster_identifier = "${aws_neptune_cluster.test.id}"
  instance_class = "db.r4.large"
}

resource "aws_neptune_cluster" "test" {
  cluster_identifier = "tf-neptune-cluster-%d"
  neptune_subnet_group_name = "${aws_neptune_subnet_group.test.name}"
  skip_final_snapshot = true
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
	tags = {
		Name = "terraform-testacc-neptune-cluster-instance-name-prefix"
	}
}

resource "aws_subnet" "a" {
  vpc_id = "${aws_vpc.test.id}"
  cidr_block = "10.0.0.0/24"
  availability_zone = "us-west-2a"
  tags = {
    Name = "tf-acc-neptune-cluster-instance-name-prefix-a"
  }
}

resource "aws_subnet" "b" {
  vpc_id = "${aws_vpc.test.id}"
  cidr_block = "10.0.1.0/24"
  availability_zone = "us-west-2b"
  tags = {
    Name = "tf-acc-neptune-cluster-instance-name-prefix-b"
  }
}

resource "aws_neptune_subnet_group" "test" {
  name = "tf-test-%d"
  subnet_ids = ["${aws_subnet.a.id}", "${aws_subnet.b.id}"]
}
`, n, n)
}

func testAccAWSNeptuneClusterInstanceConfig_generatedName(n int) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster_instance" "test" {
  cluster_identifier = "${aws_neptune_cluster.test.id}"
  instance_class = "db.r4.large"
}

resource "aws_neptune_cluster" "test" {
  cluster_identifier = "tf-neptune-cluster-%d"
  neptune_subnet_group_name = "${aws_neptune_subnet_group.test.name}"
  skip_final_snapshot = true
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
	tags = {
		Name = "terraform-testacc-neptune-cluster-instance-name-prefix"
	}
}

resource "aws_subnet" "a" {
  vpc_id = "${aws_vpc.test.id}"
  cidr_block = "10.0.0.0/24"
  availability_zone = "us-west-2a"
  tags = {
    Name = "tf-acc-neptune-cluster-instance-name-prefix-a"
  }
}

resource "aws_subnet" "b" {
  vpc_id = "${aws_vpc.test.id}"
  cidr_block = "10.0.1.0/24"
  availability_zone = "us-west-2b"
  tags = {
    Name = "tf-acc-neptune-cluster-instance-name-prefix-b"
  }
}

resource "aws_neptune_subnet_group" "test" {
  name = "tf-test-%d"
  subnet_ids = ["${aws_subnet.a.id}", "${aws_subnet.b.id}"]
}
`, n, n)
}

func testAccAWSNeptuneClusterInstanceConfigKmsKey(n int) string {
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
  cluster_identifier 	= "tf-neptune-cluster-test-%d"
  availability_zones 	= ["us-west-2a", "us-west-2b", "us-west-2c"]
  skip_final_snapshot 	= true
  storage_encrypted 	= true
  kms_key_arn			= "${aws_kms_key.foo.arn}"
}

resource "aws_neptune_cluster_instance" "cluster_instances" {
  identifier              		= "tf-cluster-instance-%d"
  cluster_identifier      		= "${aws_neptune_cluster.default.id}"
  instance_class          		= "db.r4.large"
  neptune_parameter_group_name 	= "${aws_neptune_parameter_group.bar.name}"
}

resource "aws_neptune_parameter_group" "bar" {
  name   = "tf-cluster-test-group-%d"
  family = "neptune1"

  parameter {
    name         = "neptune_query_timeout"
    value        = "25"
  }

  tags = {
    foo = "bar"
  }
}
`, n, n, n, n)
}
