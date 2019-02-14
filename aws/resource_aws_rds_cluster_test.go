package aws

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
)

func TestAccAWSRDSCluster_importBasic(t *testing.T) {
	resourceName := "aws_rds_cluster.default"
	ri := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterConfig(ri),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"master_password", "skip_final_snapshot"},
			},
		},
	})
}

func TestAccAWSRDSCluster_basic(t *testing.T) {
	var dbCluster rds.DBCluster
	rInt := acctest.RandInt()
	resourceName := "aws_rds_cluster.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(resourceName, &dbCluster),
					resource.TestMatchResourceAttr(resourceName, "arn", regexp.MustCompile(`^arn:[^:]+:rds:[^:]+:\d{12}:cluster:.+`)),
					resource.TestCheckResourceAttr(resourceName, "backtrack_window", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "db_cluster_parameter_group_name", "default.aurora5.6"),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_resource_id"),
					resource.TestCheckResourceAttr(resourceName, "engine", "aurora"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_version"),
					resource.TestCheckResourceAttr(resourceName, "global_cluster_identifier", ""),
					resource.TestCheckResourceAttrSet(resourceName, "hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName,
						"enabled_cloudwatch_logs_exports.0", "audit"),
					resource.TestCheckResourceAttr(resourceName,
						"enabled_cloudwatch_logs_exports.1", "error"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSRDSCluster_BacktrackWindow(t *testing.T) {
	var dbCluster rds.DBCluster
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterConfig_BacktrackWindow(43200),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "backtrack_window", "43200"),
				),
			},
			{
				Config: testAccAWSClusterConfig_BacktrackWindow(86400),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "backtrack_window", "86400"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"master_password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccAWSRDSCluster_namePrefix(t *testing.T) {
	var v rds.DBCluster

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterConfig_namePrefix(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists("aws_rds_cluster.test", &v),
					resource.TestMatchResourceAttr(
						"aws_rds_cluster.test", "cluster_identifier", regexp.MustCompile("^tf-test-")),
				),
			},
		},
	})
}

func TestAccAWSRDSCluster_s3Restore(t *testing.T) {
	var v rds.DBCluster
	resourceName := "aws_rds_cluster.test"
	bucket := acctest.RandomWithPrefix("tf-acc-test")
	uniqueId := acctest.RandomWithPrefix("tf-acc-s3-import-test")
	bucketPrefix := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterConfig_s3Restore(bucket, bucketPrefix, uniqueId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists("aws_rds_cluster.test", &v),
					resource.TestCheckResourceAttr(resourceName, "engine", "aurora"),
				),
			},
		},
	})
}

func TestAccAWSRDSCluster_generatedName(t *testing.T) {
	var v rds.DBCluster

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterConfig_generatedName(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists("aws_rds_cluster.test", &v),
					resource.TestMatchResourceAttr(
						"aws_rds_cluster.test", "cluster_identifier", regexp.MustCompile("^tf-")),
				),
			},
		},
	})
}

func TestAccAWSRDSCluster_takeFinalSnapshot(t *testing.T) {
	var v rds.DBCluster
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterSnapshot(rInt),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterConfigWithFinalSnapshot(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists("aws_rds_cluster.default", &v),
				),
			},
		},
	})
}

/// This is a regression test to make sure that we always cover the scenario as hightlighted in
/// https://github.com/hashicorp/terraform/issues/11568
func TestAccAWSRDSCluster_missingUserNameCausesError(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSClusterConfigWithoutUserNameAndPassword(acctest.RandInt()),
				ExpectError: regexp.MustCompile(`required field is not set`),
			},
		},
	})
}

func TestAccAWSRDSCluster_updateTags(t *testing.T) {
	var v rds.DBCluster
	ri := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterConfig(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists("aws_rds_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_rds_cluster.default", "tags.%", "1"),
				),
			},
			{
				Config: testAccAWSClusterConfigUpdatedTags(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists("aws_rds_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_rds_cluster.default", "tags.%", "2"),
				),
			},
		},
	})
}

func TestAccAWSRDSCluster_updateCloudwatchLogsExports(t *testing.T) {
	var v rds.DBCluster
	ri := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterConfig(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists("aws_rds_cluster.default", &v),
					resource.TestCheckResourceAttr("aws_rds_cluster.default",
						"enabled_cloudwatch_logs_exports.0", "audit"),
					resource.TestCheckResourceAttr("aws_rds_cluster.default",
						"enabled_cloudwatch_logs_exports.1", "error"),
				),
			},
			{
				Config: testAccAWSClusterConfigUpdatedCloudwatchLogsExports(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists("aws_rds_cluster.default", &v),
					resource.TestCheckResourceAttr("aws_rds_cluster.default",
						"enabled_cloudwatch_logs_exports.0", "error"),
					resource.TestCheckResourceAttr("aws_rds_cluster.default",
						"enabled_cloudwatch_logs_exports.1", "slowquery"),
				),
			},
		},
	})
}

func TestAccAWSRDSCluster_updateIamRoles(t *testing.T) {
	var v rds.DBCluster
	ri := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterConfigIncludingIamRoles(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists("aws_rds_cluster.default", &v),
				),
			},
			{
				Config: testAccAWSClusterConfigAddIamRoles(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists("aws_rds_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_rds_cluster.default", "iam_roles.#", "2"),
				),
			},
			{
				Config: testAccAWSClusterConfigRemoveIamRoles(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists("aws_rds_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_rds_cluster.default", "iam_roles.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSRDSCluster_kmsKey(t *testing.T) {
	var v rds.DBCluster
	keyRegex := regexp.MustCompile("^arn:aws:kms:")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterConfig_kmsKey(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists("aws_rds_cluster.default", &v),
					resource.TestMatchResourceAttr(
						"aws_rds_cluster.default", "kms_key_id", keyRegex),
				),
			},
		},
	})
}

func TestAccAWSRDSCluster_encrypted(t *testing.T) {
	var v rds.DBCluster

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterConfig_encrypted(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists("aws_rds_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_rds_cluster.default", "storage_encrypted", "true"),
					resource.TestCheckResourceAttr(
						"aws_rds_cluster.default", "db_cluster_parameter_group_name", "default.aurora5.6"),
				),
			},
		},
	})
}

func TestAccAWSRDSCluster_EncryptedCrossRegionReplication(t *testing.T) {
	var primaryCluster rds.DBCluster
	var replicaCluster rds.DBCluster

	// record the initialized providers so that we can use them to
	// check for the cluster in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckWithProviders(testAccCheckAWSClusterDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterConfigEncryptedCrossRegionReplica(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExistsWithProvider("aws_rds_cluster.test_primary",
						&primaryCluster, testAccAwsRegionProviderFunc("us-west-2", &providers)),
					testAccCheckAWSClusterExistsWithProvider("aws_rds_cluster.test_replica",
						&replicaCluster, testAccAwsRegionProviderFunc("us-east-1", &providers)),
				),
			},
		},
	})
}

func TestAccAWSRDSCluster_backupsUpdate(t *testing.T) {
	var v rds.DBCluster

	ri := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterConfig_backups(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists("aws_rds_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_rds_cluster.default", "preferred_backup_window", "07:00-09:00"),
					resource.TestCheckResourceAttr(
						"aws_rds_cluster.default", "backup_retention_period", "5"),
					resource.TestCheckResourceAttr(
						"aws_rds_cluster.default", "preferred_maintenance_window", "tue:04:00-tue:04:30"),
				),
			},

			{
				Config: testAccAWSClusterConfig_backupsUpdate(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists("aws_rds_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_rds_cluster.default", "preferred_backup_window", "03:00-09:00"),
					resource.TestCheckResourceAttr(
						"aws_rds_cluster.default", "backup_retention_period", "10"),
					resource.TestCheckResourceAttr(
						"aws_rds_cluster.default", "preferred_maintenance_window", "wed:01:00-wed:01:30"),
				),
			},
		},
	})
}

func TestAccAWSRDSCluster_iamAuth(t *testing.T) {
	var v rds.DBCluster

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterConfig_iamAuth(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists("aws_rds_cluster.default", &v),
					resource.TestCheckResourceAttr(
						"aws_rds_cluster.default", "iam_database_authentication_enabled", "true"),
				),
			},
		},
	})
}

func TestAccAWSRDSCluster_DeletionProtection(t *testing.T) {
	var dbCluster1 rds.DBCluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSClusterConfig_DeletionProtection(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(resourceName, &dbCluster1),
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
					"master_password",
					"skip_final_snapshot",
					"snapshot_identifier",
				},
			},
			{
				Config: testAccAWSRDSClusterConfig_DeletionProtection(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
				),
			},
		},
	})
}

func TestAccAWSRDSCluster_EngineMode(t *testing.T) {
	var dbCluster1, dbCluster2 rds.DBCluster

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSClusterConfig_EngineMode(rName, "serverless"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "engine_mode", "serverless"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"master_password",
					"skip_final_snapshot",
					"snapshot_identifier",
				},
			},
			{
				Config: testAccAWSRDSClusterConfig_EngineMode(rName, "provisioned"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(resourceName, &dbCluster2),
					testAccCheckAWSClusterRecreated(&dbCluster1, &dbCluster2),
					resource.TestCheckResourceAttr(resourceName, "engine_mode", "provisioned"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"master_password",
					"skip_final_snapshot",
					"snapshot_identifier",
				},
			},
		},
	})
}

func TestAccAWSRDSCluster_EngineMode_Global(t *testing.T) {
	var dbCluster1 rds.DBCluster

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSClusterConfig_EngineMode(rName, "global"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "engine_mode", "global"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"master_password",
					"skip_final_snapshot",
					"snapshot_identifier",
				},
			},
		},
	})
}

func TestAccAWSRDSCluster_EngineMode_ParallelQuery(t *testing.T) {
	var dbCluster1 rds.DBCluster

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSClusterConfig_EngineMode(rName, "parallelquery"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "engine_mode", "parallelquery"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"master_password",
					"skip_final_snapshot",
					"snapshot_identifier",
				},
			},
		},
	})
}

func TestAccAWSRDSCluster_EngineVersion(t *testing.T) {
	var dbCluster rds.DBCluster
	rInt := acctest.RandInt()
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterConfig_EngineVersion(rInt, "aurora-postgresql", "9.6.3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "engine", "aurora-postgresql"),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "9.6.3"),
				),
			},
			{
				Config:      testAccAWSClusterConfig_EngineVersion(rInt, "aurora-postgresql", "9.6.6"),
				ExpectError: regexp.MustCompile(`Cannot modify engine version without a primary instance in DB cluster`),
			},
		},
	})
}

func TestAccAWSRDSCluster_EngineVersionWithPrimaryInstance(t *testing.T) {
	var dbCluster rds.DBCluster
	rInt := acctest.RandInt()
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterConfig_EngineVersionWithPrimaryInstance(rInt, "aurora-postgresql", "9.6.3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "engine", "aurora-postgresql"),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "9.6.3"),
				),
			},
			{
				Config: testAccAWSClusterConfig_EngineVersionWithPrimaryInstance(rInt, "aurora-postgresql", "9.6.6"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "engine", "aurora-postgresql"),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "9.6.6"),
				),
			},
		},
	})
}

func TestAccAWSRDSCluster_GlobalClusterIdentifier(t *testing.T) {
	var dbCluster1 rds.DBCluster

	rName := acctest.RandomWithPrefix("tf-acc-test")
	globalClusterResourceName := "aws_rds_global_cluster.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSClusterConfig_GlobalClusterIdentifier(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttrPair(resourceName, "global_cluster_identifier", globalClusterResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"master_password",
					"skip_final_snapshot",
					"snapshot_identifier",
				},
			},
		},
	})
}

func TestAccAWSRDSCluster_GlobalClusterIdentifier_Add(t *testing.T) {
	var dbCluster1 rds.DBCluster

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSClusterConfig_EngineMode(rName, "global"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "global_cluster_identifier", ""),
				),
			},
			{
				Config:      testAccAWSRDSClusterConfig_GlobalClusterIdentifier(rName),
				ExpectError: regexp.MustCompile(`Existing RDS Clusters cannot be added to an existing RDS Global Cluster`),
			},
		},
	})
}

func TestAccAWSRDSCluster_GlobalClusterIdentifier_Remove(t *testing.T) {
	var dbCluster1 rds.DBCluster

	rName := acctest.RandomWithPrefix("tf-acc-test")
	globalClusterResourceName := "aws_rds_global_cluster.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSClusterConfig_GlobalClusterIdentifier(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttrPair(resourceName, "global_cluster_identifier", globalClusterResourceName, "id"),
				),
			},
			{
				Config: testAccAWSRDSClusterConfig_EngineMode(rName, "global"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "global_cluster_identifier", ""),
				),
			},
		},
	})
}

func TestAccAWSRDSCluster_GlobalClusterIdentifier_Update(t *testing.T) {
	var dbCluster1 rds.DBCluster

	rName := acctest.RandomWithPrefix("tf-acc-test")
	globalClusterResourceName1 := "aws_rds_global_cluster.test.0"
	globalClusterResourceName2 := "aws_rds_global_cluster.test.1"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSClusterConfig_GlobalClusterIdentifier_Update(rName, globalClusterResourceName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttrPair(resourceName, "global_cluster_identifier", globalClusterResourceName1, "id"),
				),
			},
			{
				Config:      testAccAWSRDSClusterConfig_GlobalClusterIdentifier_Update(rName, globalClusterResourceName2),
				ExpectError: regexp.MustCompile(`Existing RDS Clusters cannot be migrated between existing RDS Global Clusters`),
			},
		},
	})
}

func TestAccAWSRDSCluster_Port(t *testing.T) {
	var dbCluster1, dbCluster2 rds.DBCluster
	rInt := acctest.RandInt()
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterConfig_Port(rInt, 5432),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(resourceName, &dbCluster1),
					resource.TestCheckResourceAttr(resourceName, "port", "5432"),
				),
			},
			{
				Config: testAccAWSClusterConfig_Port(rInt, 2345),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(resourceName, &dbCluster2),
					testAccCheckAWSClusterRecreated(&dbCluster1, &dbCluster2),
					resource.TestCheckResourceAttr(resourceName, "port", "2345"),
				),
			},
		},
	})
}

func TestAccAWSRDSCluster_ScalingConfiguration(t *testing.T) {
	var dbCluster rds.DBCluster

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSClusterConfig_ScalingConfiguration(rName, false, 128, 4, 301),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.auto_pause", "false"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.max_capacity", "128"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.min_capacity", "4"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.seconds_until_auto_pause", "301"),
				),
			},
			{
				Config: testAccAWSRDSClusterConfig_ScalingConfiguration(rName, true, 256, 8, 86400),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.auto_pause", "true"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.max_capacity", "256"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.min_capacity", "8"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.seconds_until_auto_pause", "86400"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"master_password",
					"skip_final_snapshot",
					"snapshot_identifier",
				},
			},
		},
	})
}

func TestAccAWSRDSCluster_SnapshotIdentifier(t *testing.T) {
	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot rds.DBClusterSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSClusterConfig_SnapshotIdentifier(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(sourceDbResourceName, &sourceDbCluster),
					testAccCheckDbClusterSnapshotExists(snapshotResourceName, &dbClusterSnapshot),
					testAccCheckAWSClusterExists(resourceName, &dbCluster),
				),
			},
		},
	})
}

func TestAccAWSRDSCluster_SnapshotIdentifier_DeletionProtection(t *testing.T) {
	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot rds.DBClusterSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSClusterConfig_SnapshotIdentifier_DeletionProtection(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(sourceDbResourceName, &sourceDbCluster),
					testAccCheckDbClusterSnapshotExists(snapshotResourceName, &dbClusterSnapshot),
					testAccCheckAWSClusterExists(resourceName, &dbCluster),
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
					"master_password",
					"skip_final_snapshot",
					"snapshot_identifier",
				},
			},
			// Ensure we disable deletion protection before attempting to delete :)
			{
				Config: testAccAWSRDSClusterConfig_SnapshotIdentifier_DeletionProtection(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(sourceDbResourceName, &sourceDbCluster),
					testAccCheckDbClusterSnapshotExists(snapshotResourceName, &dbClusterSnapshot),
					testAccCheckAWSClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
				),
			},
		},
	})
}

func TestAccAWSRDSCluster_SnapshotIdentifier_EngineMode_ParallelQuery(t *testing.T) {
	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot rds.DBClusterSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSClusterConfig_SnapshotIdentifier_EngineMode(rName, "parallelquery"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(sourceDbResourceName, &sourceDbCluster),
					testAccCheckDbClusterSnapshotExists(snapshotResourceName, &dbClusterSnapshot),
					testAccCheckAWSClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "engine_mode", "parallelquery"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"master_password",
					"skip_final_snapshot",
					"snapshot_identifier",
				},
			},
		},
	})
}

func TestAccAWSRDSCluster_SnapshotIdentifier_EngineMode_Provisioned(t *testing.T) {
	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot rds.DBClusterSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSClusterConfig_SnapshotIdentifier_EngineMode(rName, "provisioned"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(sourceDbResourceName, &sourceDbCluster),
					testAccCheckDbClusterSnapshotExists(snapshotResourceName, &dbClusterSnapshot),
					testAccCheckAWSClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "engine_mode", "provisioned"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"master_password",
					"skip_final_snapshot",
					"snapshot_identifier",
				},
			},
		},
	})
}

func TestAccAWSRDSCluster_SnapshotIdentifier_EngineMode_Serverless(t *testing.T) {
	// The below is according to AWS Support. This test can be updated in the future
	// to initialize some data.
	t.Skip("serverless does not support snapshot restore on an empty volume")

	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot rds.DBClusterSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSClusterConfig_SnapshotIdentifier_EngineMode(rName, "serverless"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(sourceDbResourceName, &sourceDbCluster),
					testAccCheckDbClusterSnapshotExists(snapshotResourceName, &dbClusterSnapshot),
					testAccCheckAWSClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "engine_mode", "serverless"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"master_password",
					"skip_final_snapshot",
					"snapshot_identifier",
				},
			},
		},
	})
}

// Reference: https://github.com/terraform-providers/terraform-provider-aws/issues/6157
func TestAccAWSRDSCluster_SnapshotIdentifier_EngineVersion_Different(t *testing.T) {
	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot rds.DBClusterSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSClusterConfig_SnapshotIdentifier_EngineVersion(rName, "aurora-postgresql", "9.6.3", "9.6.6"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(sourceDbResourceName, &sourceDbCluster),
					testAccCheckDbClusterSnapshotExists(snapshotResourceName, &dbClusterSnapshot),
					testAccCheckAWSClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "9.6.6"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"master_password",
					"skip_final_snapshot",
					"snapshot_identifier",
				},
			},
		},
	})
}

// Reference: https://github.com/terraform-providers/terraform-provider-aws/issues/6157
func TestAccAWSRDSCluster_SnapshotIdentifier_EngineVersion_Equal(t *testing.T) {
	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot rds.DBClusterSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSClusterConfig_SnapshotIdentifier_EngineVersion(rName, "aurora-postgresql", "9.6.3", "9.6.3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(sourceDbResourceName, &sourceDbCluster),
					testAccCheckDbClusterSnapshotExists(snapshotResourceName, &dbClusterSnapshot),
					testAccCheckAWSClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "9.6.3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"master_password",
					"skip_final_snapshot",
					"snapshot_identifier",
				},
			},
		},
	})
}

func TestAccAWSRDSCluster_SnapshotIdentifier_PreferredBackupWindow(t *testing.T) {
	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot rds.DBClusterSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSClusterConfig_SnapshotIdentifier_PreferredBackupWindow(rName, "00:00-08:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(sourceDbResourceName, &sourceDbCluster),
					testAccCheckDbClusterSnapshotExists(snapshotResourceName, &dbClusterSnapshot),
					testAccCheckAWSClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "preferred_backup_window", "00:00-08:00"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"master_password",
					"skip_final_snapshot",
					"snapshot_identifier",
				},
			},
		},
	})
}

func TestAccAWSRDSCluster_SnapshotIdentifier_PreferredMaintenanceWindow(t *testing.T) {
	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot rds.DBClusterSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSClusterConfig_SnapshotIdentifier_PreferredMaintenanceWindow(rName, "sun:01:00-sun:01:30"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(sourceDbResourceName, &sourceDbCluster),
					testAccCheckDbClusterSnapshotExists(snapshotResourceName, &dbClusterSnapshot),
					testAccCheckAWSClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "preferred_maintenance_window", "sun:01:00-sun:01:30"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"cluster_identifier_prefix",
					"master_password",
					"skip_final_snapshot",
					"snapshot_identifier",
				},
			},
		},
	})
}

func TestAccAWSRDSCluster_SnapshotIdentifier_Tags(t *testing.T) {
	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot rds.DBClusterSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSClusterConfig_SnapshotIdentifier_Tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(sourceDbResourceName, &sourceDbCluster),
					testAccCheckDbClusterSnapshotExists(snapshotResourceName, &dbClusterSnapshot),
					testAccCheckAWSClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func TestAccAWSRDSCluster_SnapshotIdentifier_VpcSecurityGroupIds(t *testing.T) {
	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot rds.DBClusterSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSClusterConfig_SnapshotIdentifier_VpcSecurityGroupIds(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(sourceDbResourceName, &sourceDbCluster),
					testAccCheckDbClusterSnapshotExists(snapshotResourceName, &dbClusterSnapshot),
					testAccCheckAWSClusterExists(resourceName, &dbCluster),
				),
			},
		},
	})
}

// Regression reference: https://github.com/terraform-providers/terraform-provider-aws/issues/5450
// This acceptance test explicitly tests when snapshot_identifier is set,
// vpc_security_group_ids is set (which triggered the resource update function),
// and tags is set which was missing its ARN used for tagging
func TestAccAWSRDSCluster_SnapshotIdentifier_VpcSecurityGroupIds_Tags(t *testing.T) {
	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot rds.DBClusterSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSClusterConfig_SnapshotIdentifier_VpcSecurityGroupIds_Tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(sourceDbResourceName, &sourceDbCluster),
					testAccCheckDbClusterSnapshotExists(snapshotResourceName, &dbClusterSnapshot),
					testAccCheckAWSClusterExists(resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func TestAccAWSRDSCluster_SnapshotIdentifier_EncryptedRestore(t *testing.T) {
	var dbCluster, sourceDbCluster rds.DBCluster
	var dbClusterSnapshot rds.DBClusterSnapshot

	keyRegex := regexp.MustCompile("^arn:aws:kms:")

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_rds_cluster.source"
	snapshotResourceName := "aws_db_cluster_snapshot.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSClusterConfig_SnapshotIdentifier_EncryptedRestore(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterExists(sourceDbResourceName, &sourceDbCluster),
					testAccCheckDbClusterSnapshotExists(snapshotResourceName, &dbClusterSnapshot),
					testAccCheckAWSClusterExists(resourceName, &dbCluster),
					resource.TestMatchResourceAttr(
						"aws_rds_cluster.test", "kms_key_id", keyRegex),
					resource.TestCheckResourceAttr(
						"aws_rds_cluster.test", "storage_encrypted", "true"),
				),
			},
		},
	})
}

func testAccCheckAWSClusterDestroy(s *terraform.State) error {
	return testAccCheckAWSClusterDestroyWithProvider(s, testAccProvider)
}

func testAccCheckAWSClusterDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*AWSClient).rdsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_rds_cluster" {
			continue
		}

		// Try to find the Group
		var err error
		resp, err := conn.DescribeDBClusters(
			&rds.DescribeDBClustersInput{
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

func testAccCheckAWSClusterSnapshot(rInt int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_rds_cluster" {
				continue
			}

			// Try and delete the snapshot before we check for the cluster not found
			snapshot_identifier := fmt.Sprintf("tf-acctest-rdscluster-snapshot-%d", rInt)

			awsClient := testAccProvider.Meta().(*AWSClient)
			conn := awsClient.rdsconn

			log.Printf("[INFO] Deleting the Snapshot %s", snapshot_identifier)
			_, snapDeleteErr := conn.DeleteDBClusterSnapshot(
				&rds.DeleteDBClusterSnapshotInput{
					DBClusterSnapshotIdentifier: aws.String(snapshot_identifier),
				})
			if snapDeleteErr != nil {
				return snapDeleteErr
			}

			// Try to find the Group
			var err error
			resp, err := conn.DescribeDBClusters(
				&rds.DescribeDBClustersInput{
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

func testAccCheckAWSClusterExists(n string, v *rds.DBCluster) resource.TestCheckFunc {
	return testAccCheckAWSClusterExistsWithProvider(n, v, func() *schema.Provider { return testAccProvider })
}

func testAccCheckAWSClusterExistsWithProvider(n string, v *rds.DBCluster, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DB Instance ID is set")
		}

		provider := providerF()
		conn := provider.Meta().(*AWSClient).rdsconn
		resp, err := conn.DescribeDBClusters(&rds.DescribeDBClustersInput{
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

func testAccCheckAWSClusterRecreated(i, j *rds.DBCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.ClusterCreateTime) == aws.TimeValue(j.ClusterCreateTime) {
			return errors.New("RDS Cluster was not recreated")
		}

		return nil
	}
}

func testAccAWSClusterConfig(n int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "default" {
  cluster_identifier = "tf-aurora-cluster-%d"
  availability_zones = ["us-west-2a","us-west-2b","us-west-2c"]
  database_name = "mydb"
  master_username = "foo"
  master_password = "mustbeeightcharaters"
  db_cluster_parameter_group_name = "default.aurora5.6"
  skip_final_snapshot = true
  tags = {
    Environment = "production"
  }
  enabled_cloudwatch_logs_exports = [
	"audit",
	"error",
  ]
}`, n)
}

func testAccAWSClusterConfig_BacktrackWindow(backtrackWindow int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  apply_immediately         = true
  backtrack_window          = %d
  cluster_identifier_prefix = "tf-acc-test-"
  master_password           = "mustbeeightcharaters"
  master_username           = "test"
  skip_final_snapshot       = true
}
`, backtrackWindow)
}

func testAccAWSClusterConfig_namePrefix(n int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier_prefix = "tf-test-"
  master_username = "root"
  master_password = "password"
  db_subnet_group_name = "${aws_db_subnet_group.test.name}"
  skip_final_snapshot = true
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
	tags = {
		Name = "terraform-testacc-rds-cluster-name-prefix"
	}
}

resource "aws_subnet" "a" {
  vpc_id = "${aws_vpc.test.id}"
  cidr_block = "10.0.0.0/24"
  availability_zone = "us-west-2a"
  tags = {
    Name = "tf-acc-rds-cluster-name-prefix-a"
  }
}

resource "aws_subnet" "b" {
  vpc_id = "${aws_vpc.test.id}"
  cidr_block = "10.0.1.0/24"
  availability_zone = "us-west-2b"
  tags = {
    Name = "tf-acc-rds-cluster-name-prefix-b"
  }
}

resource "aws_db_subnet_group" "test" {
  name = "tf-test-%d"
  subnet_ids = ["${aws_subnet.a.id}", "${aws_subnet.b.id}"]
}
`, n)
}

func testAccAWSClusterConfig_s3Restore(bucketName string, bucketPrefix string, uniqueId string) string {
	return fmt.Sprintf(`

data "aws_region" "current" {}

resource "aws_s3_bucket" "xtrabackup" {
  bucket = "%s"
  region = "${data.aws_region.current.name}"
}

resource "aws_s3_bucket_object" "xtrabackup_db" {
  bucket = "${aws_s3_bucket.xtrabackup.id}"
  key    = "%s/mysql-5-6-xtrabackup.tar.gz"
  source = "../files/mysql-5-6-xtrabackup.tar.gz"
  etag   = "${md5(file("../files/mysql-5-6-xtrabackup.tar.gz"))}"
}



resource "aws_iam_role" "rds_s3_access_role" {
    name = "%s-role"
    assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "rds.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "test" {
  name   = "%s-policy"
  policy = <<POLICY
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:*"
            ],
            "Resource": [
                "${aws_s3_bucket.xtrabackup.arn}",
                "${aws_s3_bucket.xtrabackup.arn}/*"
            ]
        }
    ]
}
POLICY
}

resource "aws_iam_policy_attachment" "test-attach" {
    name = "%s-policy-attachment"
    roles = [
        "${aws_iam_role.rds_s3_access_role.name}"
    ]

    policy_arn = "${aws_iam_policy.test.arn}"
}


resource "aws_rds_cluster" "test" {
  cluster_identifier_prefix = "tf-test-"
  master_username = "root"
  master_password = "password"
  db_subnet_group_name = "${aws_db_subnet_group.test.name}"
  skip_final_snapshot = true
  s3_import {
      source_engine = "mysql"
      source_engine_version = "5.6"

      bucket_name = "${aws_s3_bucket.xtrabackup.bucket}"
      bucket_prefix = "%s"
      ingestion_role = "${aws_iam_role.rds_s3_access_role.arn}"
  }

}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
	tags = {
		Name = "%s-vpc"
	}
}

resource "aws_subnet" "a" {
  vpc_id = "${aws_vpc.test.id}"
  cidr_block = "10.0.0.0/24"
  availability_zone = "us-west-2a"
  tags = {
    Name = "%s-subnet-a"
  }
}

resource "aws_subnet" "b" {
  vpc_id = "${aws_vpc.test.id}"
  cidr_block = "10.0.1.0/24"
  availability_zone = "us-west-2b"
  tags = {
    Name = "%s-subnet-b"
  }
}

resource "aws_db_subnet_group" "test" {
  name = "%s-db-subnet-group"
  subnet_ids = ["${aws_subnet.a.id}", "${aws_subnet.b.id}"]
}
`, bucketName, bucketPrefix, uniqueId, uniqueId, uniqueId, bucketPrefix, uniqueId, uniqueId, uniqueId, uniqueId)
}

func testAccAWSClusterConfig_generatedName(n int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  master_username = "root"
  master_password = "password"
  db_subnet_group_name = "${aws_db_subnet_group.test.name}"
  skip_final_snapshot = true
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
	tags = {
		Name = "terraform-testacc-rds-cluster-generated-name"
	}
}

resource "aws_subnet" "a" {
  vpc_id = "${aws_vpc.test.id}"
  cidr_block = "10.0.0.0/24"
  availability_zone = "us-west-2a"
  tags = {
    Name = "tf-acc-rds-cluster-generated-name-a"
  }
}

resource "aws_subnet" "b" {
  vpc_id = "${aws_vpc.test.id}"
  cidr_block = "10.0.1.0/24"
  availability_zone = "us-west-2b"
  tags = {
    Name = "tf-acc-rds-cluster-generated-name-b"
  }
}

resource "aws_db_subnet_group" "test" {
  name = "tf-test-%d"
  subnet_ids = ["${aws_subnet.a.id}", "${aws_subnet.b.id}"]
}
`, n)
}

func testAccAWSClusterConfigWithFinalSnapshot(n int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "default" {
  cluster_identifier = "tf-aurora-cluster-%d"
  availability_zones = ["us-west-2a","us-west-2b","us-west-2c"]
  database_name = "mydb"
  master_username = "foo"
  master_password = "mustbeeightcharaters"
  db_cluster_parameter_group_name = "default.aurora5.6"
  final_snapshot_identifier = "tf-acctest-rdscluster-snapshot-%d"
  tags = {
    Environment = "production"
  }
}`, n, n)
}

func testAccAWSClusterConfigWithoutUserNameAndPassword(n int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "default" {
  cluster_identifier = "tf-aurora-cluster-%d"
  availability_zones = ["us-west-2a","us-west-2b","us-west-2c"]
  database_name = "mydb"
  skip_final_snapshot = true
}`, n)
}

func testAccAWSClusterConfigUpdatedTags(n int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "default" {
  cluster_identifier = "tf-aurora-cluster-%d"
  availability_zones = ["us-west-2a","us-west-2b","us-west-2c"]
  database_name = "mydb"
  master_username = "foo"
  master_password = "mustbeeightcharaters"
  db_cluster_parameter_group_name = "default.aurora5.6"
  skip_final_snapshot = true
  tags = {
    Environment = "production"
    AnotherTag = "test"
  }
}`, n)
}

func testAccAWSClusterConfigUpdatedCloudwatchLogsExports(n int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "default" {
  cluster_identifier = "tf-aurora-cluster-%d"
  availability_zones = ["us-west-2a","us-west-2b","us-west-2c"]
  database_name = "mydb"
  master_username = "foo"
  master_password = "mustbeeightcharaters"
  db_cluster_parameter_group_name = "default.aurora5.6"
  skip_final_snapshot = true
  enabled_cloudwatch_logs_exports = [
    "error",
    "slowquery"
  ]
}`, n)
}

func testAccAWSClusterConfig_kmsKey(n int) string {
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
   cluster_identifier = "tf-aurora-cluster-%d"
   availability_zones = ["us-west-2a","us-west-2b","us-west-2c"]
   database_name = "mydb"
   master_username = "foo"
   master_password = "mustbeeightcharaters"
   db_cluster_parameter_group_name = "default.aurora5.6"
   storage_encrypted = true
   kms_key_id = "${aws_kms_key.foo.arn}"
   skip_final_snapshot = true
 }`, n, n)
}

func testAccAWSClusterConfig_encrypted(n int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "default" {
  cluster_identifier = "tf-aurora-cluster-%d"
  availability_zones = ["us-west-2a","us-west-2b","us-west-2c"]
  database_name = "mydb"
  master_username = "foo"
  master_password = "mustbeeightcharaters"
  storage_encrypted = true
  skip_final_snapshot = true
}`, n)
}

func testAccAWSClusterConfig_backups(n int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "default" {
  cluster_identifier = "tf-aurora-cluster-%d"
  availability_zones = ["us-west-2a","us-west-2b","us-west-2c"]
  database_name = "mydb"
  master_username = "foo"
  master_password = "mustbeeightcharaters"
  backup_retention_period = 5
  preferred_backup_window = "07:00-09:00"
  preferred_maintenance_window = "tue:04:00-tue:04:30"
  skip_final_snapshot = true
}`, n)
}

func testAccAWSClusterConfig_backupsUpdate(n int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "default" {
  cluster_identifier = "tf-aurora-cluster-%d"
  availability_zones = ["us-west-2a","us-west-2b","us-west-2c"]
  database_name = "mydb"
  master_username = "foo"
  master_password = "mustbeeightcharaters"
  backup_retention_period = 10
  preferred_backup_window = "03:00-09:00"
  preferred_maintenance_window = "wed:01:00-wed:01:30"
  apply_immediately = true
  skip_final_snapshot = true
}`, n)
}

func testAccAWSClusterConfig_iamAuth(n int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "default" {
  cluster_identifier = "tf-aurora-cluster-%d"
  availability_zones = ["us-west-2a","us-west-2b","us-west-2c"]
  database_name = "mydb"
  master_username = "foo"
  master_password = "mustbeeightcharaters"
  iam_database_authentication_enabled = true
  skip_final_snapshot = true
}`, n)
}

func testAccAWSClusterConfig_EngineVersion(rInt int, engine, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier              = "tf-acc-test-%d"
  database_name                   = "mydb"
  db_cluster_parameter_group_name = "default.aurora-postgresql9.6"
  engine                          = "%s"
  engine_version                  = "%s"
  master_password                 = "mustbeeightcharaters"
  master_username                 = "foo"
	skip_final_snapshot             = true
	apply_immediately               = true
}`, rInt, engine, engineVersion)
}

func testAccAWSClusterConfig_EngineVersionWithPrimaryInstance(rInt int, engine, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier              = "tf-acc-test-%d"
  database_name                   = "mydb"
  db_cluster_parameter_group_name = "default.aurora-postgresql9.6"
  engine                          = %q
  engine_version                  = %q
  master_password                 = "mustbeeightcharaters"
  master_username                 = "foo"
  skip_final_snapshot             = true
  apply_immediately               = true
}

resource "aws_rds_cluster_instance" "test" {
  identifier         = "tf-acc-test-%d"
  cluster_identifier = "${aws_rds_cluster.test.cluster_identifier}"
  engine             = "${aws_rds_cluster.test.engine}"
  instance_class     = "db.r4.large"
}`, rInt, engine, engineVersion, rInt)
}

func testAccAWSClusterConfig_Port(rInt, port int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier              = "tf-acc-test-%d"
  database_name                   = "mydb"
  db_cluster_parameter_group_name = "default.aurora-postgresql9.6"
  engine                          = "aurora-postgresql"
  master_password                 = "mustbeeightcharaters"
  master_username                 = "foo"
  port                            = %d
  skip_final_snapshot             = true
}`, rInt, port)
}

func testAccAWSClusterConfigIncludingIamRoles(n int) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "rds_sample_role" {
  name = "rds_sample_role_%d"
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
resource "aws_iam_role_policy" "rds_policy" {
	name = "rds_sample_role_policy_%d"
	role = "${aws_iam_role.rds_sample_role.name}"
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
resource "aws_iam_role" "another_rds_sample_role" {
  name = "another_rds_sample_role_%d"
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
resource "aws_iam_role_policy" "another_rds_policy" {
	name = "another_rds_sample_role_policy_%d"
	role = "${aws_iam_role.another_rds_sample_role.name}"
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
resource "aws_rds_cluster" "default" {
  cluster_identifier = "tf-aurora-cluster-%d"
  availability_zones = ["us-west-2a","us-west-2b","us-west-2c"]
  database_name = "mydb"
  master_username = "foo"
  master_password = "mustbeeightcharaters"
  db_cluster_parameter_group_name = "default.aurora5.6"
  skip_final_snapshot = true
  tags = {
    Environment = "production"
  }
  depends_on = ["aws_iam_role.another_rds_sample_role", "aws_iam_role.rds_sample_role"]

}`, n, n, n, n, n)
}

func testAccAWSClusterConfigAddIamRoles(n int) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "rds_sample_role" {
  name = "rds_sample_role_%d"
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
resource "aws_iam_role_policy" "rds_policy" {
	name = "rds_sample_role_policy_%d"
	role = "${aws_iam_role.rds_sample_role.name}"
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
resource "aws_iam_role" "another_rds_sample_role" {
  name = "another_rds_sample_role_%d"
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
resource "aws_iam_role_policy" "another_rds_policy" {
	name = "another_rds_sample_role_policy_%d"
	role = "${aws_iam_role.another_rds_sample_role.name}"
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
resource "aws_rds_cluster" "default" {
  cluster_identifier = "tf-aurora-cluster-%d"
  availability_zones = ["us-west-2a","us-west-2b","us-west-2c"]
  database_name = "mydb"
  master_username = "foo"
  master_password = "mustbeeightcharaters"
  db_cluster_parameter_group_name = "default.aurora5.6"
  skip_final_snapshot = true
  iam_roles = ["${aws_iam_role.rds_sample_role.arn}","${aws_iam_role.another_rds_sample_role.arn}"]
  tags = {
    Environment = "production"
  }
  depends_on = ["aws_iam_role.another_rds_sample_role", "aws_iam_role.rds_sample_role"]

}`, n, n, n, n, n)
}

func testAccAWSClusterConfigRemoveIamRoles(n int) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "another_rds_sample_role" {
  name = "another_rds_sample_role_%d"
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
resource "aws_iam_role_policy" "another_rds_policy" {
	name = "another_rds_sample_role_policy_%d"
	role = "${aws_iam_role.another_rds_sample_role.name}"
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
resource "aws_rds_cluster" "default" {
  cluster_identifier = "tf-aurora-cluster-%d"
  availability_zones = ["us-west-2a","us-west-2b","us-west-2c"]
  database_name = "mydb"
  master_username = "foo"
  master_password = "mustbeeightcharaters"
  db_cluster_parameter_group_name = "default.aurora5.6"
  skip_final_snapshot = true
  iam_roles = ["${aws_iam_role.another_rds_sample_role.arn}"]
  tags = {
    Environment = "production"
  }

  depends_on = ["aws_iam_role.another_rds_sample_role"]
}`, n, n, n)
}

func testAccAWSClusterConfigEncryptedCrossRegionReplica(n int) string {
	return fmt.Sprintf(`
provider "aws" {
  alias  = "useast1"
  region = "us-east-1"
}

provider "aws" {
  alias  = "uswest2"
  region = "us-west-2"
}

data "aws_availability_zones" "us-east-1" {
  provider = "aws.useast1"
}

resource "aws_rds_cluster_instance" "test_instance" {
  provider = "aws.uswest2"
  identifier = "tf-aurora-instance-%[1]d"
  cluster_identifier = "${aws_rds_cluster.test_primary.id}"
  instance_class = "db.t2.small"
}

resource "aws_rds_cluster_parameter_group" "default" {
  provider = "aws.uswest2"
  name        = "tf-aurora-prm-grp-%[1]d"
  family      = "aurora5.6"
  description = "RDS default cluster parameter group"

  parameter {
    name  = "binlog_format"
    value = "STATEMENT"
    apply_method = "pending-reboot"
  }
}

resource "aws_rds_cluster" "test_primary" {
  provider = "aws.uswest2"
  cluster_identifier = "tf-test-primary-%[1]d"
  db_cluster_parameter_group_name = "${aws_rds_cluster_parameter_group.default.name}"
  database_name = "mydb"
  master_username = "foo"
  master_password = "mustbeeightcharaters"
  storage_encrypted = true
  skip_final_snapshot = true
}

data "aws_caller_identity" "current" {}

resource "aws_kms_key" "kms_key_east" {
  provider = "aws.useast1"
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

resource "aws_vpc" "main" {
  provider   = "aws.useast1"
  cidr_block = "10.0.0.0/16"
  tags = {
  	Name = "terraform-acctest-rds-cluster-encrypted-cross-region-replica"
  }
}

resource "aws_subnet" "db" {
  provider          = "aws.useast1"
  count             = 3
  vpc_id            = "${aws_vpc.main.id}"
  availability_zone = "${data.aws_availability_zones.us-east-1.names[count.index]}"
  cidr_block        = "10.0.${count.index}.0/24"
  tags = {
    Name = "tf-acc-rds-cluster-encrypted-cross-region-replica-${count.index}"
  }
}

resource "aws_db_subnet_group" "replica" {
  provider   = "aws.useast1"
  name       = "test_replica-subnet-%[1]d"
  subnet_ids = ["${aws_subnet.db.*.id[0]}", "${aws_subnet.db.*.id[1]}", "${aws_subnet.db.*.id[2]}"]
}

resource "aws_rds_cluster" "test_replica" {
  provider = "aws.useast1"
  cluster_identifier = "tf-test-replica-%[1]d"
  db_subnet_group_name = "${aws_db_subnet_group.replica.name}"
  database_name = "mydb"
  master_username = "foo"
  master_password = "mustbeeightcharaters"
  kms_key_id = "${aws_kms_key.kms_key_east.arn}"
  storage_encrypted = true
  skip_final_snapshot = true
  replication_source_identifier = "arn:aws:rds:us-west-2:${data.aws_caller_identity.current.account_id}:cluster:${aws_rds_cluster.test_primary.cluster_identifier}"
  source_region = "us-west-2"
  depends_on = [
  	"aws_rds_cluster_instance.test_instance"
  ]
}
`, n)
}

func testAccAWSRDSClusterConfig_DeletionProtection(rName string, deletionProtection bool) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier   = %q
  deletion_protection  = %t
  master_password      = "barbarbarbar"
  master_username      = "foo"
  skip_final_snapshot  = true
}
`, rName, deletionProtection)
}

func testAccAWSRDSClusterConfig_EngineMode(rName, engineMode string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier   = %q
  engine_mode          = %q
  master_password      = "barbarbarbar"
  master_username      = "foo"
  skip_final_snapshot  = true
}
`, rName, engineMode)
}

func testAccAWSRDSClusterConfig_GlobalClusterIdentifier(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test" {
  global_cluster_identifier = %q
}

resource "aws_rds_cluster" "test" {
  cluster_identifier        = %q
  global_cluster_identifier = "${aws_rds_global_cluster.test.id}"
  engine_mode               = "global"
  master_password           = "barbarbarbar"
  master_username           = "foo"
  skip_final_snapshot       = true
}
`, rName, rName)
}

func testAccAWSRDSClusterConfig_GlobalClusterIdentifier_Update(rName, globalClusterIdentifierResourceName string) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test" {
  count = 2

  global_cluster_identifier = "%s-${count.index}"
}

resource "aws_rds_cluster" "test" {
  cluster_identifier        = %q
  global_cluster_identifier = "${%s.id}"
  engine_mode               = "global"
  master_password           = "barbarbarbar"
  master_username           = "foo"
  skip_final_snapshot       = true
}
`, rName, rName, globalClusterIdentifierResourceName)
}

func testAccAWSRDSClusterConfig_ScalingConfiguration(rName string, autoPause bool, maxCapacity, minCapacity, secondsUntilAutoPause int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier   = %q
  engine_mode          = "serverless"
  master_password      = "barbarbarbar"
  master_username      = "foo"
  skip_final_snapshot  = true

  scaling_configuration {
    auto_pause               = %t
    max_capacity             = %d
    min_capacity             = %d
    seconds_until_auto_pause = %d
  }
}
`, rName, autoPause, maxCapacity, minCapacity, secondsUntilAutoPause)
}

func testAccAWSRDSClusterConfig_SnapshotIdentifier(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "source" {
  cluster_identifier   = "%s-source"
  master_password      = "barbarbarbar"
  master_username      = "foo"
  skip_final_snapshot  = true
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = "${aws_rds_cluster.source.id}"
  db_cluster_snapshot_identifier = %q
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %q
  skip_final_snapshot = true
  snapshot_identifier = "${aws_db_cluster_snapshot.test.id}"
}
`, rName, rName, rName)
}

func testAccAWSRDSClusterConfig_SnapshotIdentifier_DeletionProtection(rName string, deletionProtection bool) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "source" {
  cluster_identifier  = "%s-source"
  master_password     = "barbarbarbar"
  master_username     = "foo"
  skip_final_snapshot = true
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = "${aws_rds_cluster.source.id}"
  db_cluster_snapshot_identifier = %q
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %q
  deletion_protection = %t
  skip_final_snapshot = true
  snapshot_identifier = "${aws_db_cluster_snapshot.test.id}"
}
`, rName, rName, rName, deletionProtection)
}

func testAccAWSRDSClusterConfig_SnapshotIdentifier_EngineMode(rName, engineMode string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "source" {
  cluster_identifier  = "%s-source"
  engine_mode         = %q
  master_password     = "barbarbarbar"
  master_username     = "foo"
  skip_final_snapshot = true
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = "${aws_rds_cluster.source.id}"
  db_cluster_snapshot_identifier = %q
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %q
  engine_mode         = %q
  skip_final_snapshot = true
  snapshot_identifier = "${aws_db_cluster_snapshot.test.id}"
}
`, rName, engineMode, rName, rName, engineMode)
}

func testAccAWSRDSClusterConfig_SnapshotIdentifier_EngineVersion(rName, engine, engineVersionSource, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "source" {
  cluster_identifier  = "%s-source"
  engine              = %q
  engine_version      = %q
  master_password     = "barbarbarbar"
  master_username     = "foo"
  skip_final_snapshot = true
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = "${aws_rds_cluster.source.id}"
  db_cluster_snapshot_identifier = %q
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %q
  engine              = %q
  engine_version      = %q
  skip_final_snapshot = true
  snapshot_identifier = "${aws_db_cluster_snapshot.test.id}"
}
`, rName, engine, engineVersionSource, rName, rName, engine, engineVersion)
}

func testAccAWSRDSClusterConfig_SnapshotIdentifier_PreferredBackupWindow(rName, preferredBackupWindow string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "source" {
  cluster_identifier  = "%s-source"
  master_password     = "barbarbarbar"
  master_username     = "foo"
  skip_final_snapshot = true
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = "${aws_rds_cluster.source.id}"
  db_cluster_snapshot_identifier = %q
}

resource "aws_rds_cluster" "test" {
  cluster_identifier      = %q
  preferred_backup_window = %q
  skip_final_snapshot     = true
  snapshot_identifier     = "${aws_db_cluster_snapshot.test.id}"
}
`, rName, rName, rName, preferredBackupWindow)
}

func testAccAWSRDSClusterConfig_SnapshotIdentifier_PreferredMaintenanceWindow(rName, preferredMaintenanceWindow string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "source" {
  cluster_identifier  = "%s-source"
  master_password     = "barbarbarbar"
  master_username     = "foo"
  skip_final_snapshot = true
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = "${aws_rds_cluster.source.id}"
  db_cluster_snapshot_identifier = %q
}

resource "aws_rds_cluster" "test" {
  cluster_identifier           = %q
  preferred_maintenance_window = %q
  skip_final_snapshot          = true
  snapshot_identifier          = "${aws_db_cluster_snapshot.test.id}"
}
`, rName, rName, rName, preferredMaintenanceWindow)
}

func testAccAWSRDSClusterConfig_SnapshotIdentifier_Tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "source" {
  cluster_identifier   = "%s-source"
  master_password      = "barbarbarbar"
  master_username      = "foo"
  skip_final_snapshot  = true
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = "${aws_rds_cluster.source.id}"
  db_cluster_snapshot_identifier = %q
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %q
  skip_final_snapshot = true
  snapshot_identifier = "${aws_db_cluster_snapshot.test.id}"

  tags = {
    key1 = "value1"
  }
}
`, rName, rName, rName)
}

func testAccAWSRDSClusterConfig_SnapshotIdentifier_VpcSecurityGroupIds(rName string) string {
	return fmt.Sprintf(`
data "aws_vpc" "default" {
  default = true
}

data "aws_security_group" "default" {
  name   = "default"
  vpc_id = "${data.aws_vpc.default.id}"
}

resource "aws_rds_cluster" "source" {
  cluster_identifier   = "%s-source"
  master_password      = "barbarbarbar"
  master_username      = "foo"
  skip_final_snapshot  = true
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = "${aws_rds_cluster.source.id}"
  db_cluster_snapshot_identifier = %q
}

resource "aws_rds_cluster" "test" {
  cluster_identifier     = %q
  skip_final_snapshot    = true
  snapshot_identifier    = "${aws_db_cluster_snapshot.test.id}"
  vpc_security_group_ids = ["${data.aws_security_group.default.id}"]
}
`, rName, rName, rName)
}

func testAccAWSRDSClusterConfig_SnapshotIdentifier_VpcSecurityGroupIds_Tags(rName string) string {
	return fmt.Sprintf(`
data "aws_vpc" "default" {
  default = true
}

data "aws_security_group" "default" {
  name   = "default"
  vpc_id = "${data.aws_vpc.default.id}"
}

resource "aws_rds_cluster" "source" {
  cluster_identifier   = "%s-source"
  master_password      = "barbarbarbar"
  master_username      = "foo"
  skip_final_snapshot  = true
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = "${aws_rds_cluster.source.id}"
  db_cluster_snapshot_identifier = %q
}

resource "aws_rds_cluster" "test" {
  cluster_identifier     = %q
  skip_final_snapshot    = true
  snapshot_identifier    = "${aws_db_cluster_snapshot.test.id}"
  vpc_security_group_ids = ["${data.aws_security_group.default.id}"]

  tags = {
    key1 = "value1"
  }
}
`, rName, rName, rName)
}

func testAccAWSRDSClusterConfig_SnapshotIdentifier_EncryptedRestore(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {}

resource "aws_rds_cluster" "source" {
  cluster_identifier   = "%s-source"
  master_password      = "barbarbarbar"
  master_username      = "foo"
  skip_final_snapshot  = true
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = "${aws_rds_cluster.source.id}"
  db_cluster_snapshot_identifier = %q
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %q
  skip_final_snapshot = true
  snapshot_identifier = "${aws_db_cluster_snapshot.test.id}"

  storage_encrypted = true
  kms_key_id = "${aws_kms_key.test.arn}"
}
`, rName, rName, rName)
}
