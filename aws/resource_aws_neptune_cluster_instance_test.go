package aws

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSNeptuneClusterInstance_basic(t *testing.T) {
	var v neptune.DBInstance
	rInt := acctest.RandInt()

	resourceName := "aws_neptune_cluster_instance.cluster_instances"
	clusterResourceName := "aws_neptune_cluster.default"
	parameterGroupResourceName := "aws_neptune_parameter_group.test"

	clusterInstanceName := fmt.Sprintf("tf-cluster-instance-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterInstanceConfig(clusterInstanceName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterInstanceExists(resourceName, &v),
					testAccCheckAWSNeptuneClusterInstanceAttributes(&v),
					testAccCheckNeptuneClusterAddress(&v, resourceName, neptuneDefaultPort),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "rds", fmt.Sprintf("db:%s", clusterInstanceName)),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "true"),
					resource.TestMatchResourceAttr(resourceName, "availability_zone", regexp.MustCompile(fmt.Sprintf("^%s[a-z]{1}$", testAccGetRegion()))),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_identifier", clusterResourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "dbi_resource_id"),
					resource.TestCheckResourceAttr(resourceName, "engine", "neptune"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_version"),
					resource.TestCheckResourceAttr(resourceName, "identifier", clusterInstanceName),
					resource.TestCheckResourceAttrPair(resourceName, "instance_class", "data.aws_neptune_orderable_db_instance.test", "instance_class"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_arn", ""),
					resource.TestCheckResourceAttrPair(resourceName, "neptune_parameter_group_name", parameterGroupResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "neptune_subnet_group_name", "default"),
					resource.TestCheckResourceAttrSet(resourceName, "preferred_backup_window"),
					resource.TestCheckResourceAttrSet(resourceName, "preferred_maintenance_window"),
					resource.TestCheckResourceAttr(resourceName, "promotion_tier", "3"),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "false"),
					resource.TestCheckResourceAttr(resourceName, "storage_encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "writer", "true"),
				),
			},
			{
				Config: testAccAWSNeptuneClusterInstanceConfigModified(clusterInstanceName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterInstanceExists(resourceName, &v),
					testAccCheckAWSNeptuneClusterInstanceAttributes(&v),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
				),
			},
		},
	})
}

func TestAccAWSNeptuneClusterInstance_withaz(t *testing.T) {
	var v neptune.DBInstance
	rInt := acctest.RandInt()

	resourceName := "aws_neptune_cluster_instance.cluster_instances"
	availabiltyZonesDataSourceName := "data.aws_availability_zones.available"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterInstanceConfig_az(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterInstanceExists(resourceName, &v),
					testAccCheckAWSNeptuneClusterInstanceAttributes(&v),
					resource.TestMatchResourceAttr(resourceName, "availability_zone", regexp.MustCompile(fmt.Sprintf("^%s[a-z]{1}$", testAccGetRegion()))), // NOPE
					resource.TestCheckResourceAttrPair(resourceName, "availability_zone", availabiltyZonesDataSourceName, "names.0"),
				),
			},
		},
	})
}

func TestAccAWSNeptuneClusterInstance_namePrefix(t *testing.T) {
	var v neptune.DBInstance
	rInt := acctest.RandInt()

	resourceName := "aws_neptune_cluster_instance.test"

	namePrefix := "tf-cluster-instance-"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterInstanceConfig_namePrefix(namePrefix, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterInstanceExists(resourceName, &v),
					testAccCheckAWSNeptuneClusterInstanceAttributes(&v),
					resource.TestMatchResourceAttr(resourceName, "identifier", regexp.MustCompile(fmt.Sprintf("^%s", namePrefix))),
				),
			},
		},
	})
}

func TestAccAWSNeptuneClusterInstance_withSubnetGroup(t *testing.T) {
	var v neptune.DBInstance
	rInt := acctest.RandInt()

	resourceName := "aws_neptune_cluster_instance.test"
	subnetGroupResourceName := "aws_neptune_subnet_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterInstanceConfig_withSubnetGroup(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterInstanceExists(resourceName, &v),
					testAccCheckAWSNeptuneClusterInstanceAttributes(&v),
					resource.TestCheckResourceAttrPair(resourceName, "neptune_subnet_group_name", subnetGroupResourceName, "name"),
				),
			},
		},
	})
}

func TestAccAWSNeptuneClusterInstance_generatedName(t *testing.T) {
	var v neptune.DBInstance
	rInt := acctest.RandInt()

	resourceName := "aws_neptune_cluster_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterInstanceConfig_generatedName(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterInstanceExists(resourceName, &v),
					testAccCheckAWSNeptuneClusterInstanceAttributes(&v),
					resource.TestMatchResourceAttr(resourceName, "identifier", regexp.MustCompile("^tf-")),
				),
			},
		},
	})
}

func TestAccAWSNeptuneClusterInstance_kmsKey(t *testing.T) {
	var v neptune.DBInstance
	rInt := acctest.RandInt()

	resourceName := "aws_neptune_cluster_instance.cluster_instances"
	kmsKeyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterInstanceConfigKmsKey(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_arn", kmsKeyResourceName, "arn"),
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

func testAccCheckNeptuneClusterAddress(v *neptune.DBInstance, resourceName string, portNumber int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		address := aws.StringValue(v.Endpoint.Address)
		if err := resource.TestCheckResourceAttr(resourceName, "address", address)(s); err != nil {
			return err
		}

		port := strconv.Itoa(portNumber)
		if err := resource.TestCheckResourceAttr(resourceName, "port", port)(s); err != nil {
			return err
		}

		if err := resource.TestCheckResourceAttr(resourceName, "endpoint", fmt.Sprintf("%s:%s", address, port))(s); err != nil {
			return err
		}

		return nil
	}
}

func testAccAWSNeptuneClusterInstanceConfig(instanceName string, n int) string {
	return composeConfig(
		testAccAWSNeptuneClusterConfigBase(),
		fmt.Sprintf(`
data "aws_neptune_orderable_db_instance" "test" {
  engine        = "neptune"
  license_model = "amazon-license"

  preferred_instance_classes = ["db.t3.medium", "db.r5.large", "db.r4.large"]
}

resource "aws_neptune_cluster_instance" "cluster_instances" {
  identifier                   = %[1]q
  cluster_identifier           = aws_neptune_cluster.default.id
  instance_class               = data.aws_neptune_orderable_db_instance.test.instance_class
  engine_version               = data.aws_neptune_orderable_db_instance.test.engine_version
  neptune_parameter_group_name = aws_neptune_parameter_group.test.name
  promotion_tier               = "3"
}

resource "aws_neptune_cluster" "default" {
  cluster_identifier  = "tf-neptune-cluster-test-%[2]d"
  availability_zones  = local.availability_zone_names
  skip_final_snapshot = true
  engine_version      = data.aws_neptune_orderable_db_instance.test.engine_version
}

resource "aws_neptune_parameter_group" "test" {
  name   = "tf-cluster-test-group-%[2]d"
  family = "neptune1"

  parameter {
    name  = "neptune_query_timeout"
    value = "25"
  }

  tags = {
    Name = "test"
  }
}
`, instanceName, n))
}

func testAccAWSNeptuneClusterInstanceConfigModified(instanceName string, n int) string {
	return composeConfig(
		testAccAWSNeptuneClusterConfigBase(),
		fmt.Sprintf(`
data "aws_neptune_orderable_db_instance" "test" {
  engine        = "neptune"
  license_model = "amazon-license"

  preferred_instance_classes = ["db.t3.medium", "db.r5.large", "db.r4.large"]
}

resource "aws_neptune_cluster_instance" "cluster_instances" {
  identifier                   = %[1]q
  cluster_identifier           = aws_neptune_cluster.default.id
  instance_class               = data.aws_neptune_orderable_db_instance.test.instance_class
  engine_version               = data.aws_neptune_orderable_db_instance.test.engine_version
  neptune_parameter_group_name = aws_neptune_parameter_group.test.name
  auto_minor_version_upgrade   = false
  promotion_tier               = "3"
}

resource "aws_neptune_cluster" "default" {
  cluster_identifier  = "tf-neptune-cluster-test-%[2]d"
  availability_zones  = local.availability_zone_names
  skip_final_snapshot = true
  engine_version      = data.aws_neptune_orderable_db_instance.test.engine_version
}

resource "aws_neptune_parameter_group" "test" {
  name   = "tf-cluster-test-group-%[2]d"
  family = "neptune1"

  parameter {
    name  = "neptune_query_timeout"
    value = "25"
  }

  tags = {
    Name = "test"
  }
}
`, instanceName, n))
}

func testAccAWSNeptuneClusterInstanceConfig_az(n int) string {
	return composeConfig(
		testAccAWSNeptuneClusterConfigBase(),
		fmt.Sprintf(`
data "aws_neptune_orderable_db_instance" "test" {
  engine        = "neptune"
  license_model = "amazon-license"

  preferred_instance_classes = ["db.t3.medium", "db.r5.large", "db.r4.large"]
}

resource "aws_neptune_cluster_instance" "cluster_instances" {
  identifier                   = "tf-cluster-instance-%[1]d"
  cluster_identifier           = aws_neptune_cluster.default.id
  instance_class               = data.aws_neptune_orderable_db_instance.test.instance_class
  engine_version               = data.aws_neptune_orderable_db_instance.test.engine_version
  neptune_parameter_group_name = aws_neptune_parameter_group.test.name
  promotion_tier               = "3"
  availability_zone            = data.aws_availability_zones.available.names[0]
}

resource "aws_neptune_cluster" "default" {
  cluster_identifier  = "tf-neptune-cluster-test-%[1]d"
  availability_zones  = local.availability_zone_names
  skip_final_snapshot = true
  engine_version      = data.aws_neptune_orderable_db_instance.test.engine_version
}

resource "aws_neptune_parameter_group" "test" {
  name   = "tf-cluster-test-group-%[1]d"
  family = "neptune1"

  parameter {
    name  = "neptune_query_timeout"
    value = "25"
  }

  tags = {
    Name = "test"
  }
}
`, n))
}

func testAccAWSNeptuneClusterInstanceConfig_withSubnetGroup(n int) string {
	return composeConfig(
		testAccAWSNeptuneClusterConfigBase(),
		fmt.Sprintf(`
data "aws_neptune_orderable_db_instance" "test" {
  engine        = "neptune"
  license_model = "amazon-license"

  preferred_instance_classes = ["db.t3.medium", "db.r5.large", "db.r4.large"]
}

resource "aws_neptune_cluster_instance" "test" {
  identifier         = "tf-cluster-instance-%[1]d"
  cluster_identifier = aws_neptune_cluster.test.id
  instance_class     = data.aws_neptune_orderable_db_instance.test.instance_class
  engine_version     = data.aws_neptune_orderable_db_instance.test.engine_version
}

resource "aws_neptune_cluster" "test" {
  cluster_identifier        = "tf-neptune-cluster-%[1]d"
  neptune_subnet_group_name = aws_neptune_subnet_group.test.name
  skip_final_snapshot       = true
  engine_version            = data.aws_neptune_orderable_db_instance.test.engine_version
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-neptune-cluster-instance-name-prefix"
  }
}

resource "aws_subnet" "a" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.0.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-neptune-cluster-instance-name-prefix-a"
  }
}

resource "aws_subnet" "b" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-neptune-cluster-instance-name-prefix-b"
  }
}

resource "aws_neptune_subnet_group" "test" {
  name       = "tf-test-%[1]d"
  subnet_ids = [aws_subnet.a.id, aws_subnet.b.id]
}
`, n))
}

func testAccAWSNeptuneClusterInstanceConfig_namePrefix(namePrefix string, n int) string {
	return composeConfig(
		testAccAWSNeptuneClusterConfigBase(),
		fmt.Sprintf(`
data "aws_neptune_orderable_db_instance" "test" {
  engine        = "neptune"
  license_model = "amazon-license"

  preferred_instance_classes = ["db.t3.medium", "db.r5.large", "db.r4.large"]
}

resource "aws_neptune_cluster_instance" "test" {
  identifier_prefix  = %[1]q
  cluster_identifier = aws_neptune_cluster.test.id
  instance_class     = data.aws_neptune_orderable_db_instance.test.instance_class
  engine_version     = data.aws_neptune_orderable_db_instance.test.engine_version
}

resource "aws_neptune_cluster" "test" {
  cluster_identifier        = "tf-neptune-cluster-%[2]d"
  neptune_subnet_group_name = aws_neptune_subnet_group.test.name
  skip_final_snapshot       = true
  engine_version            = data.aws_neptune_orderable_db_instance.test.engine_version
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-neptune-cluster-instance-name-prefix"
  }
}

resource "aws_subnet" "a" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.0.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-neptune-cluster-instance-name-prefix-a"
  }
}

resource "aws_subnet" "b" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-neptune-cluster-instance-name-prefix-b"
  }
}

resource "aws_neptune_subnet_group" "test" {
  name       = "tf-test-%[2]d"
  subnet_ids = [aws_subnet.a.id, aws_subnet.b.id]
}
`, namePrefix, n))
}

func testAccAWSNeptuneClusterInstanceConfig_generatedName(n int) string {
	return composeConfig(
		testAccAWSNeptuneClusterConfigBase(),
		fmt.Sprintf(`
data "aws_neptune_orderable_db_instance" "test" {
  engine        = "neptune"
  license_model = "amazon-license"

  preferred_instance_classes = ["db.t3.medium", "db.r5.large", "db.r4.large"]
}

resource "aws_neptune_cluster_instance" "test" {
  cluster_identifier = aws_neptune_cluster.test.id
  instance_class     = data.aws_neptune_orderable_db_instance.test.instance_class
  engine_version     = data.aws_neptune_orderable_db_instance.test.engine_version
}

resource "aws_neptune_cluster" "test" {
  cluster_identifier        = "tf-neptune-cluster-%[1]d"
  neptune_subnet_group_name = aws_neptune_subnet_group.test.name
  skip_final_snapshot       = true
  engine_version            = data.aws_neptune_orderable_db_instance.test.engine_version
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-neptune-cluster-instance-name-prefix"
  }
}

resource "aws_subnet" "a" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.0.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-neptune-cluster-instance-name-prefix-a"
  }
}

resource "aws_subnet" "b" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-neptune-cluster-instance-name-prefix-b"
  }
}

resource "aws_neptune_subnet_group" "test" {
  name       = "tf-test-%[1]d"
  subnet_ids = [aws_subnet.a.id, aws_subnet.b.id]
}
`, n))
}

func testAccAWSNeptuneClusterInstanceConfigKmsKey(n int) string {
	return composeConfig(
		testAccAWSNeptuneClusterConfigBase(),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
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

data "aws_neptune_orderable_db_instance" "test" {
  engine        = "neptune"
  license_model = "amazon-license"

  preferred_instance_classes = ["db.t3.medium", "db.r5.large", "db.r4.large"]
}

resource "aws_neptune_cluster" "default" {
  cluster_identifier  = "tf-neptune-cluster-test-%[1]d"
  availability_zones  = local.availability_zone_names
  skip_final_snapshot = true
  storage_encrypted   = true
  kms_key_arn         = aws_kms_key.test.arn
  engine_version      = data.aws_neptune_orderable_db_instance.test.engine_version
}

resource "aws_neptune_cluster_instance" "cluster_instances" {
  identifier                   = "tf-cluster-instance-%[1]d"
  cluster_identifier           = aws_neptune_cluster.default.id
  instance_class               = data.aws_neptune_orderable_db_instance.test.instance_class
  engine_version               = data.aws_neptune_orderable_db_instance.test.engine_version
  neptune_parameter_group_name = aws_neptune_parameter_group.test.name
}

resource "aws_neptune_parameter_group" "test" {
  name   = "tf-cluster-test-group-%[1]d"
  family = "neptune1"

  parameter {
    name  = "neptune_query_timeout"
    value = "25"
  }

  tags = {
    Name = "test"
  }
}
`, n))
}
