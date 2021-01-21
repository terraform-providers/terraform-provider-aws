package aws

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_rds_global_cluster", &resource.Sweeper{
		Name: "aws_rds_global_cluster",
		F:    testSweepRdsGlobalClusters,
		Dependencies: []string{
			"aws_rds_cluster",
		},
	})
}

func testSweepRdsGlobalClusters(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).rdsconn
	input := &rds.DescribeGlobalClustersInput{}

	err = conn.DescribeGlobalClustersPages(input, func(out *rds.DescribeGlobalClustersOutput, lastPage bool) bool {
		for _, globalCluster := range out.GlobalClusters {
			id := aws.StringValue(globalCluster.GlobalClusterIdentifier)
			input := &rds.DeleteGlobalClusterInput{
				GlobalClusterIdentifier: globalCluster.GlobalClusterIdentifier,
			}

			log.Printf("[INFO] Deleting RDS Global Cluster: %s", id)

			_, err := conn.DeleteGlobalCluster(input)

			if err != nil {
				log.Printf("[ERROR] Failed to delete RDS Global Cluster (%s): %s", id, err)
				continue
			}

			if err := waitForRdsGlobalClusterDeletion(conn, id); err != nil {
				log.Printf("[ERROR] Failure while waiting for RDS Global Cluster (%s) to be deleted: %s", id, err)
			}
		}
		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping RDS Global Cluster sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error retrieving RDS Global Clusters: %s", err)
	}

	return nil
}

func TestAccAWSRdsGlobalCluster_basic(t *testing.T) {
	var globalCluster1 rds.GlobalCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSRdsGlobalCluster(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRdsGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRdsGlobalClusterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRdsGlobalClusterExists(resourceName, &globalCluster1),
					testAccCheckResourceAttrGlobalARN(resourceName, "arn", "rds", fmt.Sprintf("global-cluster:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "database_name", ""),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "engine"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_version"),
					resource.TestCheckResourceAttr(resourceName, "global_cluster_identifier", rName),
					resource.TestMatchResourceAttr(resourceName, "global_cluster_resource_id", regexp.MustCompile(`cluster-.+`)),
					resource.TestCheckResourceAttr(resourceName, "storage_encrypted", "false"),
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

func TestAccAWSRdsGlobalCluster_disappears(t *testing.T) {
	var globalCluster1 rds.GlobalCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSRdsGlobalCluster(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRdsGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRdsGlobalClusterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRdsGlobalClusterExists(resourceName, &globalCluster1),
					testAccCheckAWSRdsGlobalClusterDisappears(&globalCluster1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSRdsGlobalCluster_DatabaseName(t *testing.T) {
	var globalCluster1, globalCluster2 rds.GlobalCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSRdsGlobalCluster(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRdsGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRdsGlobalClusterConfigDatabaseName(rName, "database1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRdsGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttr(resourceName, "database_name", "database1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSRdsGlobalClusterConfigDatabaseName(rName, "database2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRdsGlobalClusterExists(resourceName, &globalCluster2),
					testAccCheckAWSRdsGlobalClusterRecreated(&globalCluster1, &globalCluster2),
					resource.TestCheckResourceAttr(resourceName, "database_name", "database2"),
				),
			},
		},
	})
}

func TestAccAWSRdsGlobalCluster_DeletionProtection(t *testing.T) {
	var globalCluster1, globalCluster2 rds.GlobalCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSRdsGlobalCluster(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRdsGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRdsGlobalClusterConfigDeletionProtection(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRdsGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSRdsGlobalClusterConfigDeletionProtection(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRdsGlobalClusterExists(resourceName, &globalCluster2),
					testAccCheckAWSRdsGlobalClusterNotRecreated(&globalCluster1, &globalCluster2),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
				),
			},
		},
	})
}

func TestAccAWSRdsGlobalCluster_Engine_Aurora(t *testing.T) {
	var globalCluster1 rds.GlobalCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSRdsGlobalCluster(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRdsGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRdsGlobalClusterConfigEngine(rName, "aurora"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRdsGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttr(resourceName, "engine", "aurora"),
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

func TestAccAWSRdsGlobalCluster_EngineVersion_Aurora(t *testing.T) {
	var globalCluster1 rds.GlobalCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSRdsGlobalCluster(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRdsGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRdsGlobalClusterConfigEngineVersion(rName, "aurora", "5.6.10a"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRdsGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "5.6.10a"),
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

func TestAccAWSRdsGlobalCluster_EngineVersion_AuroraMySQL(t *testing.T) {
	var globalCluster1 rds.GlobalCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSRdsGlobalCluster(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRdsGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRdsGlobalClusterConfigEngineVersion(rName, "aurora-mysql", "5.7.mysql_aurora.2.07.1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRdsGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "5.7.mysql_aurora.2.07.1"),
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

func TestAccAWSRdsGlobalCluster_EngineVersion_AuroraPostgresql(t *testing.T) {
	var globalCluster1 rds.GlobalCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSRdsGlobalCluster(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRdsGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRdsGlobalClusterConfigEngineVersion(rName, "aurora-postgresql", "10.11"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRdsGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "10.11"),
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

func TestAccAWSRdsGlobalCluster_SourceDbClusterIdentifier(t *testing.T) {
	var globalCluster1 rds.GlobalCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	clusterResourceName := "aws_rds_cluster.test"
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSRdsGlobalCluster(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRdsGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRdsGlobalClusterConfigSourceDbClusterIdentifier(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRdsGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttrPair(resourceName, "source_db_cluster_identifier", clusterResourceName, "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "source_db_cluster_identifier"},
			},
		},
	})
}

func TestAccAWSRdsGlobalCluster_SourceDbClusterIdentifier_StorageEncrypted(t *testing.T) {
	var globalCluster1 rds.GlobalCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	clusterResourceName := "aws_rds_cluster.test"
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSRdsGlobalCluster(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRdsGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRdsGlobalClusterConfigSourceDbClusterIdentifierStorageEncrypted(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRdsGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttrPair(resourceName, "source_db_cluster_identifier", clusterResourceName, "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "source_db_cluster_identifier"},
			},
		},
	})
}

func TestAccAWSRdsGlobalCluster_StorageEncrypted(t *testing.T) {
	var globalCluster1, globalCluster2 rds.GlobalCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSRdsGlobalCluster(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRdsGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRdsGlobalClusterConfigStorageEncrypted(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRdsGlobalClusterExists(resourceName, &globalCluster1),
					resource.TestCheckResourceAttr(resourceName, "storage_encrypted", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSRdsGlobalClusterConfigStorageEncrypted(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRdsGlobalClusterExists(resourceName, &globalCluster2),
					testAccCheckAWSRdsGlobalClusterRecreated(&globalCluster1, &globalCluster2),
					resource.TestCheckResourceAttr(resourceName, "storage_encrypted", "false"),
				),
			},
		},
	})
}

func testAccCheckAWSRdsGlobalClusterExists(resourceName string, globalCluster *rds.GlobalCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No RDS Global Cluster ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).rdsconn

		cluster, err := rdsDescribeGlobalCluster(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if cluster == nil {
			return fmt.Errorf("RDS Global Cluster not found")
		}

		if aws.StringValue(cluster.Status) != "available" {
			return fmt.Errorf("RDS Global Cluster (%s) exists in non-available (%s) state", rs.Primary.ID, aws.StringValue(cluster.Status))
		}

		*globalCluster = *cluster

		return nil
	}
}

func testAccCheckAWSRdsGlobalClusterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).rdsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_rds_global_cluster" {
			continue
		}

		globalCluster, err := rdsDescribeGlobalCluster(conn, rs.Primary.ID)

		if isAWSErr(err, rds.ErrCodeGlobalClusterNotFoundFault, "") {
			continue
		}

		if err != nil {
			return err
		}

		if globalCluster == nil {
			continue
		}

		return fmt.Errorf("RDS Global Cluster (%s) still exists in non-deleted (%s) state", rs.Primary.ID, aws.StringValue(globalCluster.Status))
	}

	return nil
}

func testAccCheckAWSRdsGlobalClusterDisappears(globalCluster *rds.GlobalCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).rdsconn

		input := &rds.DeleteGlobalClusterInput{
			GlobalClusterIdentifier: globalCluster.GlobalClusterIdentifier,
		}

		_, err := conn.DeleteGlobalCluster(input)

		if err != nil {
			return err
		}

		return waitForRdsGlobalClusterDeletion(conn, aws.StringValue(globalCluster.GlobalClusterIdentifier))
	}
}

func testAccCheckAWSRdsGlobalClusterNotRecreated(i, j *rds.GlobalCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.GlobalClusterResourceId) != aws.StringValue(j.GlobalClusterResourceId) {
			return errors.New("RDS Global Cluster was recreated")
		}

		return nil
	}
}

func testAccCheckAWSRdsGlobalClusterRecreated(i, j *rds.GlobalCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.GlobalClusterResourceId) == aws.StringValue(j.GlobalClusterResourceId) {
			return errors.New("RDS Global Cluster was not recreated")
		}

		return nil
	}
}

func testAccPreCheckAWSRdsGlobalCluster(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).rdsconn

	input := &rds.DescribeGlobalClustersInput{}

	_, err := conn.DescribeGlobalClusters(input)

	if testAccPreCheckSkipError(err) || isAWSErr(err, "InvalidParameterValue", "Access Denied to API Version: APIGlobalDatabases") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSRdsGlobalClusterConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test" {
  global_cluster_identifier = %q
}
`, rName)
}

func testAccAWSRdsGlobalClusterConfigDatabaseName(rName, databaseName string) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test" {
  database_name             = %q
  global_cluster_identifier = %q
}
`, databaseName, rName)
}

func testAccAWSRdsGlobalClusterConfigDeletionProtection(rName string, deletionProtection bool) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test" {
  deletion_protection       = %t
  global_cluster_identifier = %q
}
`, deletionProtection, rName)
}

func testAccAWSRdsGlobalClusterConfigEngine(rName, engine string) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test" {
  engine                    = %q
  global_cluster_identifier = %q
}
`, engine, rName)
}

func testAccAWSRdsGlobalClusterConfigEngineVersion(rName, engine, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test" {
  engine                    = %q
  engine_version            = %q
  global_cluster_identifier = %q
}
`, engine, engineVersion, rName)
}

func testAccAWSRdsGlobalClusterConfigSourceDbClusterIdentifier(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  engine              = "aurora-postgresql"
  engine_version      = "10.11" # Minimum supported version for Global Clusters
  master_password     = "mustbeeightcharacters"
  master_username     = "test"
  skip_final_snapshot = true

  # global_cluster_identifier cannot be Computed

  lifecycle {
    ignore_changes = [global_cluster_identifier]
  }
}

resource "aws_rds_global_cluster" "test" {
  force_destroy                = true
  global_cluster_identifier    = %[1]q
  source_db_cluster_identifier = aws_rds_cluster.test.arn
}
`, rName)
}

func testAccAWSRdsGlobalClusterConfigSourceDbClusterIdentifierStorageEncrypted(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  engine              = "aurora-postgresql"
  engine_version      = "10.11" # Minimum supported version for Global Clusters
  master_password     = "mustbeeightcharacters"
  master_username     = "test"
  skip_final_snapshot = true
  storage_encrypted   = true

  # global_cluster_identifier cannot be Computed

  lifecycle {
    ignore_changes = [global_cluster_identifier]
  }
}

resource "aws_rds_global_cluster" "test" {
  force_destroy                = true
  global_cluster_identifier    = %[1]q
  source_db_cluster_identifier = aws_rds_cluster.test.arn
}
`, rName)
}

func testAccAWSRdsGlobalClusterConfigStorageEncrypted(rName string, storageEncrypted bool) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test" {
  global_cluster_identifier = %q
  storage_encrypted         = %t
}
`, rName, storageEncrypted)
}
