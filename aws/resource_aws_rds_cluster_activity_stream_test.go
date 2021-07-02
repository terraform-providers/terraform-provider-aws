package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_rds_cluster_activity_stream", &resource.Sweeper{
		Name: "aws_rds_cluster_activity_stream",
		F:    func(region string) error { return nil },
		Dependencies: []string{
			"aws_kms_key",
			"aws_kinesis_stream",
			"aws_rds_cluster",
		},
	})
}

func TestAccAWSRDSClusterActivityStream_basic(t *testing.T) {
	var dbCluster rds.DBCluster
	clusterName := acctest.RandomWithPrefix("tf-testacc-aurora-cluster")
	instanceName := acctest.RandomWithPrefix("tf-testacc-aurora-instance")
	resourceName := "aws_rds_cluster_activity_stream.test"
	rdsClusterResourceName := "aws_rds_cluster.test"
	kmsKeyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterActivityStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterActivityStreamConfig(clusterName, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRDSClusterActivityStreamExists(resourceName, &dbCluster),
					testAccCheckAWSRDSClusterActivityStreamAttributes(&dbCluster),
					testAccMatchResourceAttrRegionalARN(resourceName, "resource_arn", "rds", regexp.MustCompile("cluster:"+clusterName)),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", rdsClusterResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsKeyResourceName, "key_id"),
					resource.TestCheckResourceAttrSet(resourceName, "kinesis_stream_name"),
					resource.TestCheckResourceAttr(resourceName, "mode", rds.ActivityStreamModeAsync),
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

func TestAccAWSRDSClusterActivityStream_disappears(t *testing.T) {
	var dbCluster rds.DBCluster
	clusterName := acctest.RandomWithPrefix("tf-testacc-aurora-cluster")
	instanceName := acctest.RandomWithPrefix("tf-testacc-aurora-instance")
	resourceName := "aws_rds_cluster_activity_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterActivityStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterActivityStreamConfig(clusterName, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRDSClusterActivityStreamExists(resourceName, &dbCluster),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsRDSClusterActivityStream(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSRDSClusterActivityStream_kmsKeyId(t *testing.T) {
	var dbCluster rds.DBCluster
	clusterName := acctest.RandomWithPrefix("tf-testacc-aurora-cluster")
	instanceName := acctest.RandomWithPrefix("tf-testacc-aurora-instance")
	resourceName := "aws_rds_cluster_activity_stream.test"
	kmsKeyResourceName := "aws_kms_key.test"
	newKmsKeyResourceName := "aws_kms_key.new_kms_key"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterActivityStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterActivityStreamConfig(clusterName, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRDSClusterActivityStreamExists(resourceName, &dbCluster),
					testAccCheckAWSRDSClusterActivityStreamAttributes(&dbCluster),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsKeyResourceName, "key_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSClusterActivityStreamConfig_kmsKeyId(clusterName, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRDSClusterActivityStreamExists(resourceName, &dbCluster),
					testAccCheckAWSRDSClusterActivityStreamAttributes(&dbCluster),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", newKmsKeyResourceName, "key_id"),
				),
			},
		},
	})
}

func TestAccAWSRDSClusterActivityStream_mode(t *testing.T) {
	var dbCluster rds.DBCluster
	clusterName := acctest.RandomWithPrefix("tf-testacc-aurora-cluster")
	instanceName := acctest.RandomWithPrefix("tf-testacc-aurora-instance")
	resourceName := "aws_rds_cluster_activity_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterActivityStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterActivityStreamConfig(clusterName, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRDSClusterActivityStreamExists(resourceName, &dbCluster),
					testAccCheckAWSRDSClusterActivityStreamAttributes(&dbCluster),
					resource.TestCheckResourceAttr(resourceName, "mode", rds.ActivityStreamModeAsync),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSClusterActivityStreamConfig_modeSync(clusterName, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRDSClusterActivityStreamExists(resourceName, &dbCluster),
					testAccCheckAWSRDSClusterActivityStreamAttributes(&dbCluster),
					resource.TestCheckResourceAttr(resourceName, "mode", rds.ActivityStreamModeSync),
				),
			},
		},
	})
}

func TestAccAWSRDSClusterActivityStream_resourceArn(t *testing.T) {
	var dbCluster rds.DBCluster
	clusterName := acctest.RandomWithPrefix("tf-testacc-aurora-cluster")
	instanceName := acctest.RandomWithPrefix("tf-testacc-aurora-instance")
	newClusterName := acctest.RandomWithPrefix("tf-testacc-new-aurora-cluster")
	newInstanceName := acctest.RandomWithPrefix("tf-testacc-new-aurora-instance")

	resourceName := "aws_rds_cluster_activity_stream.test"
	rdsClusterResourceName := "aws_rds_cluster.test"
	newRdsClusterResourceName := "aws_rds_cluster.new_rds_cluster_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterActivityStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterActivityStreamConfig(clusterName, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRDSClusterActivityStreamExists(resourceName, &dbCluster),
					testAccCheckAWSRDSClusterActivityStreamAttributes(&dbCluster),
					testAccMatchResourceAttrRegionalARN(resourceName, "resource_arn", "rds", regexp.MustCompile("cluster:"+clusterName)),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", rdsClusterResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSClusterActivityStreamConfig_resourceArn(newClusterName, newInstanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRDSClusterActivityStreamExists(resourceName, &dbCluster),
					testAccCheckAWSRDSClusterActivityStreamAttributes(&dbCluster),
					testAccMatchResourceAttrRegionalARN(resourceName, "resource_arn", "rds", regexp.MustCompile("cluster:"+newClusterName)),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", newRdsClusterResourceName, "arn"),
				),
			},
		},
	})
}

func testAccCheckAWSRDSClusterActivityStreamExists(resourceName string, dbCluster *rds.DBCluster) resource.TestCheckFunc {
	return testAccCheckAWSRDSClusterActivityStreamExistsWithProvider(resourceName, dbCluster, testAccProvider)
}

func testAccCheckAWSRDSClusterActivityStreamExistsWithProvider(resourceName string, dbCluster *rds.DBCluster, provider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("DBCluster ID is not set")
		}

		conn := provider.Meta().(*AWSClient).rdsconn

		response, err := conn.DescribeDBClusters(&rds.DescribeDBClustersInput{
			DBClusterIdentifier: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if len(response.DBClusters) != 1 || *response.DBClusters[0].DBClusterArn != rs.Primary.ID {
			return fmt.Errorf("DBCluster not found")
		}

		*dbCluster = *response.DBClusters[0]
		return nil
	}
}

func testAccCheckAWSRDSClusterActivityStreamAttributes(v *rds.DBCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if aws.StringValue(v.DBClusterArn) == "" {
			return fmt.Errorf("empty RDS Cluster arn")
		}

		if aws.StringValue(v.ActivityStreamKmsKeyId) == "" {
			return fmt.Errorf("empty RDS Cluster activity stream kms key id")
		}

		if aws.StringValue(v.ActivityStreamKinesisStreamName) == "" {
			return fmt.Errorf("empty RDS Cluster activity stream kinesis stream name")
		}

		if aws.StringValue(v.ActivityStreamStatus) != rds.ActivityStreamStatusStarted {
			return fmt.Errorf("incorrect activity stream status: expected: %s, got: %s", rds.ActivityStreamStatusStarted, aws.StringValue(v.ActivityStreamStatus))
		}

		if aws.StringValue(v.ActivityStreamMode) != rds.ActivityStreamModeSync && aws.StringValue(v.ActivityStreamMode) != rds.ActivityStreamModeAsync {
			return fmt.Errorf("incorrect activity stream mode: expected: sync or async, got: %s", aws.StringValue(v.ActivityStreamMode))
		}

		return nil
	}
}

func testAccCheckAWSClusterActivityStreamDestroy(s *terraform.State) error {
	return testAccCheckAWSClusterActivityStreamDestroyWithProvider(s, testAccProvider)
}

func testAccCheckAWSClusterActivityStreamDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*AWSClient).rdsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_rds_cluster_activity_stream" {
			continue
		}

		var err error
		resp, err := conn.DescribeDBClusters(
			&rds.DescribeDBClustersInput{
				DBClusterIdentifier: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if len(resp.DBClusters) != 0 &&
				*resp.DBClusters[0].ActivityStreamStatus != rds.ActivityStreamStatusStopped {
				return fmt.Errorf("DB Cluster %s Activity Stream still exists", rs.Primary.ID)
			}
		}

		// Return nil if the cluster is already destroyed
		if isAWSErr(err, rds.ErrCodeDBClusterNotFoundFault, "") {
			return nil
		}

		return err
	}

	return nil
}

func testAccAWSClusterActivityStreamConfig(clusterName, instanceName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_kms_key" "test" {
  description             = "Testing for AWS RDS Cluster Activity Stream"
  deletion_window_in_days = 7
}

resource "aws_rds_cluster" "test" {
  cluster_identifier              = "%[1]s"
  engine                          = "aurora-postgresql"
  engine_version                  = "10.11"
  availability_zones              = ["${data.aws_availability_zones.available.names[0]}", "${data.aws_availability_zones.available.names[1]}", "${data.aws_availability_zones.available.names[2]}"]
  database_name                   = "mydb"
  master_username                 = "foo"
  master_password                 = "mustbeeightcharaters"
  db_cluster_parameter_group_name = "default.aurora-postgresql10"
  skip_final_snapshot             = true
  deletion_protection             = false
}

resource "aws_rds_cluster_instance" "test" {
  identifier         = "%[2]s"
  cluster_identifier = aws_rds_cluster.test.cluster_identifier
  engine             = aws_rds_cluster.test.engine
  instance_class     = "db.r5.large"
}

resource "aws_rds_cluster_activity_stream" "test" {
  resource_arn = aws_rds_cluster.test.arn
  kms_key_id   = aws_kms_key.test.key_id
  mode         = "async"

  depends_on = [aws_rds_cluster_instance.test]
}
`, clusterName, instanceName)
}

func testAccAWSClusterActivityStreamConfig_kmsKeyId(clusterName, instanceName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_kms_key" "new_kms_key" {
  description             = "Testing for AWS RDS Cluster Activity Stream"
  deletion_window_in_days = 7
}

resource "aws_rds_cluster" "test" {
  cluster_identifier              = "%[1]s"
  engine                          = "aurora-postgresql"
  engine_version                  = "10.11"
  availability_zones              = ["${data.aws_availability_zones.available.names[0]}", "${data.aws_availability_zones.available.names[1]}", "${data.aws_availability_zones.available.names[2]}"]
  database_name                   = "mydb"
  master_username                 = "foo"
  master_password                 = "mustbeeightcharaters"
  db_cluster_parameter_group_name = "default.aurora-postgresql10"
  skip_final_snapshot             = true
  deletion_protection             = false
}

resource "aws_rds_cluster_instance" "test" {
  identifier         = "%[2]s"
  cluster_identifier = aws_rds_cluster.test.cluster_identifier
  engine             = aws_rds_cluster.test.engine
  instance_class     = "db.r5.large"
}

resource "aws_rds_cluster_activity_stream" "test" {
  resource_arn = aws_rds_cluster.test.arn
  kms_key_id   = aws_kms_key.new_kms_key.key_id
  mode         = "async"

  depends_on = [aws_rds_cluster_instance.test]
}
`, clusterName, instanceName)
}

func testAccAWSClusterActivityStreamConfig_modeSync(clusterName, instanceName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_kms_key" "test" {
  description             = "Testing for AWS RDS Cluster Activity Stream"
  deletion_window_in_days = 7
}

resource "aws_rds_cluster" "test" {
  cluster_identifier              = "%[1]s"
  engine                          = "aurora-postgresql"
  engine_version                  = "10.11"
  availability_zones              = ["${data.aws_availability_zones.available.names[0]}", "${data.aws_availability_zones.available.names[1]}", "${data.aws_availability_zones.available.names[2]}"]
  database_name                   = "mydb"
  master_username                 = "foo"
  master_password                 = "mustbeeightcharaters"
  db_cluster_parameter_group_name = "default.aurora-postgresql10"
  skip_final_snapshot             = true
  deletion_protection             = false
}

resource "aws_rds_cluster_instance" "test" {
  identifier         = "%[2]s"
  cluster_identifier = aws_rds_cluster.test.cluster_identifier
  engine             = aws_rds_cluster.test.engine
  instance_class     = "db.r5.large"
}

resource "aws_rds_cluster_activity_stream" "test" {
  resource_arn = aws_rds_cluster.test.arn
  kms_key_id   = aws_kms_key.test.key_id
  mode         = "sync"

  depends_on = [aws_rds_cluster_instance.test]
}
`, clusterName, instanceName)
}

func testAccAWSClusterActivityStreamConfig_resourceArn(clusterName, instanceName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_kms_key" "test" {
  description             = "Testing for AWS RDS Cluster Activity Stream"
  deletion_window_in_days = 7
}

resource "aws_rds_cluster" "new_rds_cluster_test" {
  cluster_identifier              = "%[1]s"
  engine                          = "aurora-postgresql"
  engine_version                  = "10.11"
  availability_zones              = ["${data.aws_availability_zones.available.names[0]}", "${data.aws_availability_zones.available.names[1]}", "${data.aws_availability_zones.available.names[2]}"]
  database_name                   = "mydb"
  master_username                 = "foo"
  master_password                 = "mustbeeightcharaters"
  db_cluster_parameter_group_name = "default.aurora-postgresql10"
  skip_final_snapshot             = true
  deletion_protection             = false
}

resource "aws_rds_cluster_instance" "new_rds_instance_test" {
  identifier         = "%[2]s"
  cluster_identifier = aws_rds_cluster.new_rds_cluster_test.cluster_identifier
  engine             = aws_rds_cluster.new_rds_cluster_test.engine
  instance_class     = "db.r5.large"
}

resource "aws_rds_cluster_activity_stream" "test" {
  resource_arn = aws_rds_cluster.new_rds_cluster_test.arn
  kms_key_id   = aws_kms_key.test.key_id
  mode         = "async"

  depends_on = [aws_rds_cluster_instance.new_rds_instance_test]
}
`, clusterName, instanceName)
}
