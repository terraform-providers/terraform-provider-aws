package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_db_instance", &resource.Sweeper{
		Name: "aws_db_instance",
		F:    testSweepDbInstances,
	})
}

func testSweepDbInstances(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).rdsconn

	err = conn.DescribeDBInstancesPages(&rds.DescribeDBInstancesInput{}, func(out *rds.DescribeDBInstancesOutput, lastPage bool) bool {
		for _, dbi := range out.DBInstances {
			log.Printf("[INFO] Deleting DB instance: %s", *dbi.DBInstanceIdentifier)

			_, err := conn.DeleteDBInstance(&rds.DeleteDBInstanceInput{
				DBInstanceIdentifier: dbi.DBInstanceIdentifier,
				SkipFinalSnapshot:    aws.Bool(true),
			})
			if err != nil {
				log.Printf("[ERROR] Failed to delete DB instance %s: %s",
					*dbi.DBInstanceIdentifier, err)
				continue
			}

			err = waitUntilAwsDbInstanceIsDeleted(*dbi.DBInstanceIdentifier, conn, 40*time.Minute)
			if err != nil {
				log.Printf("[ERROR] Failure while waiting for DB instance %s to be deleted: %s",
					*dbi.DBInstanceIdentifier, err)
			}
		}
		return !lastPage
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping RDS DB Instance sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving DB instances: %s", err)
	}

	return nil
}

func TestAccAWSDBInstance_basic(t *testing.T) {
	var dbInstance1 rds.DBInstance
	resourceName := "aws_db_instance.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance1),
					testAccCheckAWSDBInstanceAttributes(&dbInstance1),
					resource.TestCheckResourceAttr(resourceName, "allocated_storage", "10"),
					resource.TestCheckNoResourceAttr(resourceName, "allow_major_version_upgrade"),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "true"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "rds", regexp.MustCompile(`db:.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "backup_window"),
					resource.TestCheckResourceAttrSet(resourceName, "ca_cert_identifier"),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", "false"),
					resource.TestCheckResourceAttr(resourceName, "db_subnet_group_name", "default"),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint"),
					resource.TestCheckResourceAttr(resourceName, "engine", "mysql"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_version"),
					resource.TestCheckResourceAttrSet(resourceName, "hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "iam_database_authentication_enabled", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_class", "data.aws_rds_orderable_db_instance.test", "instance_class"),
					resource.TestCheckResourceAttr(resourceName, "license_model", "general-public-license"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window"),
					resource.TestCheckResourceAttr(resourceName, "max_allocated_storage", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", "baz"),
					resource.TestCheckResourceAttr(resourceName, "option_group_name", "default:mysql-5-6"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.mysql5.6"),
					resource.TestCheckResourceAttr(resourceName, "port", "3306"),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_id"),
					resource.TestCheckResourceAttr(resourceName, "status", "available"),
					resource.TestCheckResourceAttr(resourceName, "storage_encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "username", "foo"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"final_snapshot_identifier",
					"password",
					"skip_final_snapshot",
					"delete_automated_backups",
				},
			},
		},
	})
}

func TestAccAWSDBInstance_namePrefix(t *testing.T) {
	var v rds.DBInstance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_namePrefix(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists("aws_db_instance.test", &v),
					testAccCheckAWSDBInstanceAttributes(&v),
					resource.TestMatchResourceAttr(
						"aws_db_instance.test", "identifier", regexp.MustCompile("^tf-test-")),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_generatedName(t *testing.T) {
	var v rds.DBInstance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_generatedName(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists("aws_db_instance.test", &v),
					testAccCheckAWSDBInstanceAttributes(&v),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_kmsKey(t *testing.T) {
	var v rds.DBInstance
	kmsKeyResourceName := "aws_kms_key.foo"
	resourceName := "aws_db_instance.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_KmsKeyId(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &v),
					testAccCheckAWSDBInstanceAttributes(&v),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsKeyResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"delete_automated_backups",
					"final_snapshot_identifier",
					"password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccAWSDBInstance_subnetGroup(t *testing.T) {
	var v rds.DBInstance
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_WithSubnetGroup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists("aws_db_instance.bar", &v),
					resource.TestCheckResourceAttr(
						"aws_db_instance.bar", "db_subnet_group_name", "foo-"+rName),
				),
			},
			{
				Config: testAccAWSDBInstanceConfig_WithSubnetGroupUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists("aws_db_instance.bar", &v),
					resource.TestCheckResourceAttr(
						"aws_db_instance.bar", "db_subnet_group_name", "bar-"+rName),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_optionGroup(t *testing.T) {
	var v rds.DBInstance

	rName := fmt.Sprintf("tf-option-test-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_WithOptionGroup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists("aws_db_instance.bar", &v),
					testAccCheckAWSDBInstanceAttributes(&v),
					resource.TestCheckResourceAttr(
						"aws_db_instance.bar", "option_group_name", rName),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_iamAuth(t *testing.T) {
	var v rds.DBInstance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_IAMAuth(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists("aws_db_instance.bar", &v),
					testAccCheckAWSDBInstanceAttributes(&v),
					resource.TestCheckResourceAttr(
						"aws_db_instance.bar", "iam_database_authentication_enabled", "true"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_AllowMajorVersionUpgrade(t *testing.T) {
	var dbInstance1 rds.DBInstance

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_AllowMajorVersionUpgrade(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance1),
					resource.TestCheckResourceAttr(resourceName, "allow_major_version_upgrade", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"allow_major_version_upgrade",
					"apply_immediately",
					"final_snapshot_identifier",
					"password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccAWSDBInstanceConfig_AllowMajorVersionUpgrade(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance1),
					resource.TestCheckResourceAttr(resourceName, "allow_major_version_upgrade", "false"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_DbSubnetGroupName(t *testing.T) {
	var dbInstance rds.DBInstance
	var dbSubnetGroup rds.DBSubnetGroup

	rName := acctest.RandomWithPrefix("tf-acc-test")
	dbSubnetGroupResourceName := "aws_db_subnet_group.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_DbSubnetGroupName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					testAccCheckDBSubnetGroupExists(resourceName, &dbSubnetGroup),
					resource.TestCheckResourceAttrPair(resourceName, "db_subnet_group_name", dbSubnetGroupResourceName, "name"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_DbSubnetGroupName_RamShared(t *testing.T) {
	var dbInstance rds.DBInstance
	var dbSubnetGroup rds.DBSubnetGroup
	var providers []*schema.Provider

	rName := acctest.RandomWithPrefix("tf-acc-test")
	dbSubnetGroupResourceName := "aws_db_subnet_group.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
			testAccOrganizationsEnabledPreCheck(t)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_DbSubnetGroupName_RamShared(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					testAccCheckDBSubnetGroupExists(dbSubnetGroupResourceName, &dbSubnetGroup),
					resource.TestCheckResourceAttrPair(resourceName, "db_subnet_group_name", dbSubnetGroupResourceName, "name"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_DbSubnetGroupName_VpcSecurityGroupIds(t *testing.T) {
	var dbInstance rds.DBInstance
	var dbSubnetGroup rds.DBSubnetGroup

	rName := acctest.RandomWithPrefix("tf-acc-test")
	dbSubnetGroupResourceName := "aws_db_subnet_group.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_DbSubnetGroupName_VpcSecurityGroupIds(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					testAccCheckDBSubnetGroupExists(resourceName, &dbSubnetGroup),
					resource.TestCheckResourceAttrPair(resourceName, "db_subnet_group_name", dbSubnetGroupResourceName, "name"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_DeletionProtection(t *testing.T) {
	var dbInstance rds.DBInstance

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_DeletionProtection(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"final_snapshot_identifier",
					"password",
					"skip_final_snapshot",
					"delete_automated_backups",
				},
			},
			{
				Config: testAccAWSDBInstanceConfig_DeletionProtection(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_FinalSnapshotIdentifier(t *testing.T) {
	var snap rds.DBInstance
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		// testAccCheckAWSDBInstanceSnapshot verifies a database snapshot is
		// created, and subsequently deletes it
		CheckDestroy: testAccCheckAWSDBInstanceSnapshot,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_FinalSnapshotIdentifier(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists("aws_db_instance.snapshot", &snap),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_FinalSnapshotIdentifier_SkipFinalSnapshot(t *testing.T) {
	var snap rds.DBInstance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceNoSnapshot,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_FinalSnapshotIdentifier_SkipFinalSnapshot(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists("aws_db_instance.snapshot", &snap),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_IsAlreadyBeingDeleted(t *testing.T) {
	var dbInstance rds.DBInstance

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_MariaDB(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
				),
			},
			{
				PreConfig: func() {
					// Get Database Instance into deleting state
					conn := testAccProvider.Meta().(*AWSClient).rdsconn
					input := &rds.DeleteDBInstanceInput{
						DBInstanceIdentifier: aws.String(rName),
						SkipFinalSnapshot:    aws.Bool(true),
					}
					_, err := conn.DeleteDBInstance(input)
					if err != nil {
						t.Fatalf("error deleting Database Instance: %s", err)
					}
				},
				Config:  testAccAWSDBInstanceConfig_MariaDB(rName),
				Destroy: true,
			},
		},
	})
}

func TestAccAWSDBInstance_MaxAllocatedStorage(t *testing.T) {
	var dbInstance rds.DBInstance

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_MaxAllocatedStorage(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "max_allocated_storage", "10"),
				),
			},
			{
				Config: testAccAWSDBInstanceConfig_MaxAllocatedStorage(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "max_allocated_storage", "0"),
				),
			},
			{
				Config: testAccAWSDBInstanceConfig_MaxAllocatedStorage(rName, 15),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "max_allocated_storage", "15"),
				),
			},
			{
				Config: testAccAWSDBInstanceConfig_MaxAllocatedStorage(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "max_allocated_storage", "0"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_Password(t *testing.T) {
	var dbInstance rds.DBInstance

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			// Password should not be shown in error message
			{
				Config:      testAccAWSDBInstanceConfig_Password(rName, "invalid"),
				ExpectError: regexp.MustCompile(`MasterUserPassword: "\*{8}",`),
			},
			{
				Config: testAccAWSDBInstanceConfig_Password(rName, "valid-password"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "password", "valid-password"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"final_snapshot_identifier",
					"password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccAWSDBInstance_ReplicateSourceDb(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_ReplicateSourceDb(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					testAccCheckAWSDBInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_ReplicateSourceDb_AllocatedStorage(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_ReplicateSourceDb_AllocatedStorage(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					testAccCheckAWSDBInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "allocated_storage", "10"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_ReplicateSourceDb_AllowMajorVersionUpgrade(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_ReplicateSourceDb_AllowMajorVersionUpgrade(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					testAccCheckAWSDBInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "allow_major_version_upgrade", "true"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_ReplicateSourceDb_AutoMinorVersionUpgrade(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_ReplicateSourceDb_AutoMinorVersionUpgrade(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					testAccCheckAWSDBInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_ReplicateSourceDb_AvailabilityZone(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_ReplicateSourceDb_AvailabilityZone(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					testAccCheckAWSDBInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_ReplicateSourceDb_BackupRetentionPeriod(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_ReplicateSourceDb_BackupRetentionPeriod(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					testAccCheckAWSDBInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period", "1"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_ReplicateSourceDb_BackupWindow(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_ReplicateSourceDb_BackupWindow(rName, "00:00-08:00", "sun:23:00-sun:23:30"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					testAccCheckAWSDBInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "backup_window", "00:00-08:00"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_ReplicateSourceDb_DbSubnetGroupName(t *testing.T) {
	var dbInstance rds.DBInstance
	var dbSubnetGroup rds.DBSubnetGroup
	var providers []*schema.Provider

	rName := acctest.RandomWithPrefix("tf-acc-test")
	dbSubnetGroupResourceName := "aws_db_subnet_group.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionPreCheck(t, 2)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_ReplicateSourceDb_DbSubnetGroupName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					testAccCheckDBSubnetGroupExists(resourceName, &dbSubnetGroup),
					resource.TestCheckResourceAttrPair(resourceName, "db_subnet_group_name", dbSubnetGroupResourceName, "name"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_ReplicateSourceDb_DbSubnetGroupName_RamShared(t *testing.T) {
	var dbInstance rds.DBInstance
	var dbSubnetGroup rds.DBSubnetGroup
	var providers []*schema.Provider

	rName := acctest.RandomWithPrefix("tf-acc-test")
	dbSubnetGroupResourceName := "aws_db_subnet_group.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionPreCheck(t, 2)
			testAccAlternateAccountPreCheck(t)
			testAccOrganizationsEnabledPreCheck(t)
		},
		ProviderFactories: testAccProviderFactoriesAlternateAccountAndAlternateRegion(&providers),
		CheckDestroy:      testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_ReplicateSourceDb_DbSubnetGroupName_RamShared(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					testAccCheckDBSubnetGroupExists(dbSubnetGroupResourceName, &dbSubnetGroup),
					resource.TestCheckResourceAttrPair(resourceName, "db_subnet_group_name", dbSubnetGroupResourceName, "name"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_ReplicateSourceDb_DbSubnetGroupName_VpcSecurityGroupIds(t *testing.T) {
	var dbInstance rds.DBInstance
	var dbSubnetGroup rds.DBSubnetGroup
	var providers []*schema.Provider

	rName := acctest.RandomWithPrefix("tf-acc-test")
	dbSubnetGroupResourceName := "aws_db_subnet_group.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionPreCheck(t, 2)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_ReplicateSourceDb_DbSubnetGroupName_VpcSecurityGroupIds(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					testAccCheckDBSubnetGroupExists(resourceName, &dbSubnetGroup),
					resource.TestCheckResourceAttrPair(resourceName, "db_subnet_group_name", dbSubnetGroupResourceName, "name"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_ReplicateSourceDb_DeletionProtection(t *testing.T) {
	TestAccSkip(t, "CreateDBInstanceReadReplica API currently ignores DeletionProtection=true with SourceDBInstanceIdentifier set")
	// --- FAIL: TestAccAWSDBInstance_ReplicateSourceDb_DeletionProtection (1624.88s)
	//     testing.go:527: Step 0 error: Check failed: Check 4/4 error: aws_db_instance.test: Attribute 'deletion_protection' expected "true", got "false"
	//
	// Action=CreateDBInstanceReadReplica&AutoMinorVersionUpgrade=true&CopyTagsToSnapshot=false&DBInstanceClass=db.t2.micro&DBInstanceIdentifier=tf-acc-test-6591588621809891413&DeletionProtection=true&PubliclyAccessible=false&SourceDBInstanceIdentifier=tf-acc-test-6591588621809891413-source&Tags=&Version=2014-10-31
	// <RestoreDBInstanceFromDBSnapshotResponse xmlns="http://rds.amazonaws.com/doc/2014-10-31/">
	//   <RestoreDBInstanceFromDBSnapshotResult>
	//     <DBInstance>
	//       <DeletionProtection>false</DeletionProtection>
	//
	// AWS Support has confirmed this issue and noted that it will be fixed in the future.

	var dbInstance, sourceDbInstance rds.DBInstance

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_ReplicateSourceDb_DeletionProtection(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					testAccCheckAWSDBInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "true"),
				),
			},
			// Ensure we disable deletion protection before attempting to delete :)
			{
				Config: testAccAWSDBInstanceConfig_ReplicateSourceDb_DeletionProtection(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					testAccCheckAWSDBInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_ReplicateSourceDb_IamDatabaseAuthenticationEnabled(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_ReplicateSourceDb_IamDatabaseAuthenticationEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					testAccCheckAWSDBInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "iam_database_authentication_enabled", "true"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_ReplicateSourceDb_MaintenanceWindow(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_ReplicateSourceDb_MaintenanceWindow(rName, "00:00-08:00", "sun:23:00-sun:23:30"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					testAccCheckAWSDBInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window", "sun:23:00-sun:23:30"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_ReplicateSourceDb_MaxAllocatedStorage(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_ReplicateSourceDb_MaxAllocatedStorage(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					testAccCheckAWSDBInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "max_allocated_storage", "10"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_ReplicateSourceDb_Monitoring(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_ReplicateSourceDb_Monitoring(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					testAccCheckAWSDBInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "5"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_ReplicateSourceDb_MultiAZ(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_ReplicateSourceDb_MultiAZ(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					testAccCheckAWSDBInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "multi_az", "true"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_ReplicateSourceDb_ParameterGroupName(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_ReplicateSourceDb_ParameterGroupName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					testAccCheckAWSDBInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", rName),
					testAccCheckAWSDBInstanceParameterApplyStatusInSync(&dbInstance),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_ReplicateSourceDb_Port(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_ReplicateSourceDb_Port(rName, 9999),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					testAccCheckAWSDBInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "port", "9999"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_ReplicateSourceDb_VpcSecurityGroupIds(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_ReplicateSourceDb_VpcSecurityGroupIds(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					testAccCheckAWSDBInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_ReplicateSourceDb_CACertificateIdentifier(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"
	dataSourceName := "data.aws_rds_certificate.latest"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_ReplicateSourceDb_CACertificateIdentifier(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					testAccCheckAWSDBInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttrPair(sourceResourceName, "ca_cert_identifier", dataSourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "ca_cert_identifier", dataSourceName, "id"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_S3Import(t *testing.T) {
	var snap rds.DBInstance
	bucket := acctest.RandomWithPrefix("tf-acc-test")
	uniqueId := acctest.RandomWithPrefix("tf-acc-s3-import-test")
	bucketPrefix := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_S3Import(bucket, bucketPrefix, uniqueId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists("aws_db_instance.s3", &snap),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_SnapshotIdentifier(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_SnapshotIdentifier(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDbSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_SnapshotIdentifier_AllocatedStorage(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_SnapshotIdentifier_AllocatedStorage(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDbSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "allocated_storage", "10"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_SnapshotIdentifier_Io1Storage(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_SnapshotIdentifier_Io1Storage(rName, 1000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDbSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "iops", "1000"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_SnapshotIdentifier_AllowMajorVersionUpgrade(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_SnapshotIdentifier_AllowMajorVersionUpgrade(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDbSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "allow_major_version_upgrade", "true"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_SnapshotIdentifier_AutoMinorVersionUpgrade(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_SnapshotIdentifier_AutoMinorVersionUpgrade(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDbSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_SnapshotIdentifier_AvailabilityZone(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_SnapshotIdentifier_AvailabilityZone(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDbSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_SnapshotIdentifier_BackupRetentionPeriod(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_SnapshotIdentifier_BackupRetentionPeriod(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDbSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period", "1"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_SnapshotIdentifier_BackupRetentionPeriod_Unset(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_SnapshotIdentifier_BackupRetentionPeriod_Unset(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDbSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_period", "0"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_SnapshotIdentifier_BackupWindow(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_SnapshotIdentifier_BackupWindow(rName, "00:00-08:00", "sun:23:00-sun:23:30"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDbSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "backup_window", "00:00-08:00"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_SnapshotIdentifier_DbSubnetGroupName(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot
	var dbSubnetGroup rds.DBSubnetGroup

	rName := acctest.RandomWithPrefix("tf-acc-test")
	dbSubnetGroupResourceName := "aws_db_subnet_group.test"
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_SnapshotIdentifier_DbSubnetGroupName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDbSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					testAccCheckDBSubnetGroupExists(resourceName, &dbSubnetGroup),
					resource.TestCheckResourceAttrPair(resourceName, "db_subnet_group_name", dbSubnetGroupResourceName, "name"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_SnapshotIdentifier_DbSubnetGroupName_RamShared(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot
	var dbSubnetGroup rds.DBSubnetGroup
	var providers []*schema.Provider

	rName := acctest.RandomWithPrefix("tf-acc-test")
	dbSubnetGroupResourceName := "aws_db_subnet_group.test"
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
			testAccOrganizationsEnabledPreCheck(t)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_SnapshotIdentifier_DbSubnetGroupName_RamShared(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDbSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					testAccCheckDBSubnetGroupExists(dbSubnetGroupResourceName, &dbSubnetGroup),
					resource.TestCheckResourceAttrPair(resourceName, "db_subnet_group_name", dbSubnetGroupResourceName, "name"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_SnapshotIdentifier_DbSubnetGroupName_VpcSecurityGroupIds(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot
	var dbSubnetGroup rds.DBSubnetGroup

	rName := acctest.RandomWithPrefix("tf-acc-test")
	dbSubnetGroupResourceName := "aws_db_subnet_group.test"
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_SnapshotIdentifier_DbSubnetGroupName_VpcSecurityGroupIds(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDbSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					testAccCheckDBSubnetGroupExists(resourceName, &dbSubnetGroup),
					resource.TestCheckResourceAttrPair(resourceName, "db_subnet_group_name", dbSubnetGroupResourceName, "name"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_SnapshotIdentifier_DeletionProtection(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_SnapshotIdentifier_DeletionProtection(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDbSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "true"),
				),
			},
			// Ensure we disable deletion protection before attempting to delete :)
			{
				Config: testAccAWSDBInstanceConfig_SnapshotIdentifier_DeletionProtection(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDbSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_SnapshotIdentifier_IamDatabaseAuthenticationEnabled(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_SnapshotIdentifier_IamDatabaseAuthenticationEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDbSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "iam_database_authentication_enabled", "true"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_SnapshotIdentifier_MaintenanceWindow(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_SnapshotIdentifier_MaintenanceWindow(rName, "00:00-08:00", "sun:23:00-sun:23:30"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDbSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window", "sun:23:00-sun:23:30"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_SnapshotIdentifier_MaxAllocatedStorage(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_SnapshotIdentifier_MaxAllocatedStorage(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDbSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "max_allocated_storage", "10"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_SnapshotIdentifier_Monitoring(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_SnapshotIdentifier_Monitoring(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDbSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "5"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_SnapshotIdentifier_MultiAZ(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_SnapshotIdentifier_MultiAZ(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDbSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "multi_az", "true"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_SnapshotIdentifier_MultiAZ_SQLServer(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_SnapshotIdentifier_MultiAZ_SQLServer(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDbSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "multi_az", "true"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_SnapshotIdentifier_ParameterGroupName(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_SnapshotIdentifier_ParameterGroupName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDbSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", rName),
					testAccCheckAWSDBInstanceParameterApplyStatusInSync(&dbInstance),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_SnapshotIdentifier_Port(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_SnapshotIdentifier_Port(rName, 9999),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDbSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "port", "9999"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_SnapshotIdentifier_Tags(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_SnapshotIdentifier_Tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDbSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_SnapshotIdentifier_Tags_Unset(t *testing.T) {
	TestAccSkip(t, "To be fixed: https://github.com/hashicorp/terraform-provider-aws/issues/5959")
	// --- FAIL: TestAccAWSDBInstance_SnapshotIdentifier_Tags_Unset (1086.15s)
	//     testing.go:527: Step 0 error: Check failed: Check 4/4 error: aws_db_instance.test: Attribute 'tags.%' expected "0", got "1"

	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_SnapshotIdentifier_Tags_Unset(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDbSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_SnapshotIdentifier_VpcSecurityGroupIds(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_SnapshotIdentifier_VpcSecurityGroupIds(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDbSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
				),
			},
		},
	})
}

// Regression reference: https://github.com/hashicorp/terraform-provider-aws/issues/5360
// This acceptance test explicitly tests when snapshot_identifier is set,
// vpc_security_group_ids is set (which triggered the resource update function),
// and tags is set which was missing its ARN used for tagging
func TestAccAWSDBInstance_SnapshotIdentifier_VpcSecurityGroupIds_Tags(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_SnapshotIdentifier_VpcSecurityGroupIds_Tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDbSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_MonitoringInterval(t *testing.T) {
	var dbInstance rds.DBInstance
	resourceName := "aws_db_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDbInstanceConfig_MonitoringInterval(rName, 30),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "30"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"final_snapshot_identifier",
					"password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccAWSDbInstanceConfig_MonitoringInterval(rName, 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "60"),
				),
			},
			{
				Config: testAccAWSDbInstanceConfig_MonitoringInterval(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "0"),
				),
			},
			{
				Config: testAccAWSDbInstanceConfig_MonitoringInterval(rName, 30),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "30"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_MonitoringRoleArn_EnabledToDisabled(t *testing.T) {
	var dbInstance rds.DBInstance
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_db_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDbInstanceConfig_MonitoringRoleArn(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "monitoring_role_arn", iamRoleResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"final_snapshot_identifier",
					"password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccAWSDbInstanceConfig_MonitoringInterval(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "0"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_MonitoringRoleArn_EnabledToRemoved(t *testing.T) {
	var dbInstance rds.DBInstance
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_db_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDbInstanceConfig_MonitoringRoleArn(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "monitoring_role_arn", iamRoleResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"final_snapshot_identifier",
					"password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccAWSDbInstanceConfig_MonitoringRoleArnRemoved(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_MonitoringRoleArn_RemovedToEnabled(t *testing.T) {
	var dbInstance rds.DBInstance
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_db_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDbInstanceConfig_MonitoringRoleArnRemoved(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"final_snapshot_identifier",
					"password",
					"skip_final_snapshot",
				},
			},
			{
				Config: testAccAWSDbInstanceConfig_MonitoringRoleArn(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "monitoring_role_arn", iamRoleResourceName, "arn"),
				),
			},
		},
	})
}

// Regression test for https://github.com/hashicorp/terraform/issues/3760 .
// We apply a plan, then change just the iops. If the apply succeeds, we
// consider this a pass, as before in 3760 the request would fail
func TestAccAWSDBInstance_separateIopsUpdate(t *testing.T) {
	var v rds.DBInstance

	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_SnapshotInstanceConfig_iopsUpdate(rName, 1000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists("aws_db_instance.bar", &v),
					testAccCheckAWSDBInstanceAttributes(&v),
				),
			},

			{
				Config: testAccAWSDBInstanceConfig_SnapshotInstanceConfig_iopsUpdate(rName, 2000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists("aws_db_instance.bar", &v),
					testAccCheckAWSDBInstanceAttributes(&v),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_portUpdate(t *testing.T) {
	var v rds.DBInstance

	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_SnapshotInstanceConfig_mysqlPort(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists("aws_db_instance.bar", &v),
					resource.TestCheckResourceAttr(
						"aws_db_instance.bar", "port", "3306"),
				),
			},

			{
				Config: testAccAWSDBInstanceConfig_SnapshotInstanceConfig_updateMysqlPort(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists("aws_db_instance.bar", &v),
					resource.TestCheckResourceAttr(
						"aws_db_instance.bar", "port", "3305"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_MSSQL_TZ(t *testing.T) {
	var v rds.DBInstance
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_MSSQL_timezone(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists("aws_db_instance.mssql", &v),
					testAccCheckAWSDBInstanceAttributes_MSSQL(&v, ""),
					resource.TestCheckResourceAttr(
						"aws_db_instance.mssql", "allocated_storage", "20"),
					resource.TestCheckResourceAttr(
						"aws_db_instance.mssql", "engine", "sqlserver-ex"),
				),
			},

			{
				Config: testAccAWSDBInstanceConfig_MSSQL_timezone_AKST(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists("aws_db_instance.mssql", &v),
					testAccCheckAWSDBInstanceAttributes_MSSQL(&v, "Alaskan Standard Time"),
					resource.TestCheckResourceAttr(
						"aws_db_instance.mssql", "allocated_storage", "20"),
					resource.TestCheckResourceAttr(
						"aws_db_instance.mssql", "engine", "sqlserver-ex"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_MSSQL_Domain(t *testing.T) {
	var vBefore, vAfter rds.DBInstance
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_MSSQLDomain(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists("aws_db_instance.mssql", &vBefore),
					testAccCheckAWSDBInstanceDomainAttributes("terraformtesting.com", &vBefore),
					resource.TestCheckResourceAttrSet(
						"aws_db_instance.mssql", "domain"),
					resource.TestCheckResourceAttrSet(
						"aws_db_instance.mssql", "domain_iam_role_name"),
				),
			},
			{
				Config: testAccAWSDBInstanceConfig_MSSQLUpdateDomain(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists("aws_db_instance.mssql", &vAfter),
					testAccCheckAWSDBInstanceDomainAttributes("corp.notexample.com", &vAfter),
					resource.TestCheckResourceAttrSet(
						"aws_db_instance.mssql", "domain"),
					resource.TestCheckResourceAttrSet(
						"aws_db_instance.mssql", "domain_iam_role_name"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_MSSQL_DomainSnapshotRestore(t *testing.T) {
	var v, vRestoredInstance rds.DBInstance
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_MSSQLDomainSnapshotRestore(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists("aws_db_instance.mssql_restore", &vRestoredInstance),
					testAccCheckAWSDBInstanceExists("aws_db_instance.mssql", &v),
					testAccCheckAWSDBInstanceDomainAttributes("terraformtesting.com", &vRestoredInstance),
					resource.TestCheckResourceAttrSet(
						"aws_db_instance.mssql_restore", "domain"),
					resource.TestCheckResourceAttrSet(
						"aws_db_instance.mssql_restore", "domain_iam_role_name"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_MySQL_SnapshotRestoreWithEngineVersion(t *testing.T) {
	var v, vRestoredInstance rds.DBInstance
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_MySQLSnapshotRestoreWithEngineVersion(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists("aws_db_instance.mysql_restore", &vRestoredInstance),
					testAccCheckAWSDBInstanceExists("aws_db_instance.mysql", &v),
					resource.TestCheckResourceAttr(
						"aws_db_instance.mysql", "engine_version", "5.6.35"),
					resource.TestCheckResourceAttr(
						"aws_db_instance.mysql_restore", "engine_version", "5.6.41"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_MinorVersion(t *testing.T) {
	var v rds.DBInstance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_AutoMinorVersion(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists("aws_db_instance.bar", &v),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_ec2Classic(t *testing.T) {
	var v rds.DBInstance
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_Ec2Classic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists("aws_db_instance.bar", &v),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_cloudwatchLogsExportConfiguration(t *testing.T) {
	var v rds.DBInstance

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_CloudwatchLogsExportConfiguration(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists("aws_db_instance.bar", &v),
				),
			},
			{
				ResourceName:      "aws_db_instance.bar",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"final_snapshot_identifier",
					"password",
					"skip_final_snapshot",
					"delete_automated_backups",
				},
			},
		},
	})
}

func TestAccAWSDBInstance_EnabledCloudwatchLogsExports_MySQL(t *testing.T) {
	var v rds.DBInstance
	resourceName := "aws_db_instance.bar"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_CloudwatchLogsExportConfiguration(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "audit"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "error"),
				),
			},
			{
				Config: testAccAWSDBInstanceConfig_CloudwatchLogsExportConfigurationAdd(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "audit"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "error"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "general"),
				),
			},
			{
				Config: testAccAWSDBInstanceConfig_CloudwatchLogsExportConfigurationModify(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "audit"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "general"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cloudwatch_logs_exports.*", "slowquery"),
				),
			},
			{
				Config: testAccAWSDBInstanceConfig_CloudwatchLogsExportConfigurationDelete(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_EnabledCloudwatchLogsExports_MSSQL(t *testing.T) {
	var dbInstance rds.DBInstance

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_EnabledCloudwatchLogsExports_MSSQL(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"final_snapshot_identifier",
					"password",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccAWSDBInstance_EnabledCloudwatchLogsExports_Oracle(t *testing.T) {
	var dbInstance rds.DBInstance

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_EnabledCloudwatchLogsExports_Oracle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", "3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"final_snapshot_identifier",
					"password",
					"skip_final_snapshot",
					"delete_automated_backups",
				},
			},
		},
	})
}

func TestAccAWSDBInstance_EnabledCloudwatchLogsExports_Postgresql(t *testing.T) {
	var dbInstance rds.DBInstance

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_EnabledCloudwatchLogsExports_Postgresql(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "enabled_cloudwatch_logs_exports.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"final_snapshot_identifier",
					"password",
					"skip_final_snapshot",
					"delete_automated_backups",
				},
			},
		},
	})
}

func TestAccAWSDBInstance_NoDeleteAutomatedBackups(t *testing.T) {
	var dbInstance rds.DBInstance

	rName := acctest.RandomWithPrefix("tf-testacc-nodelautobak")
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceAutomatedBackups,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_NoDeleteAutomatedBackups(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
				),
			},
		},
	})
}

func testAccCheckAWSDBInstanceAutomatedBackups(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).rdsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_instance" {
			continue
		}

		log.Printf("[INFO] Trying to locate the DBInstance Automated Backup")
		describeOutput, err := conn.DescribeDBInstanceAutomatedBackups(
			&rds.DescribeDBInstanceAutomatedBackupsInput{
				DBInstanceIdentifier: aws.String(rs.Primary.ID),
			})
		if err != nil {
			return err
		}

		if describeOutput == nil || len(describeOutput.DBInstanceAutomatedBackups) == 0 {
			return fmt.Errorf("Automated backup for %s not found", rs.Primary.ID)
		}

		log.Printf("[INFO] Deleting automated backup for %s", rs.Primary.ID)
		_, err = conn.DeleteDBInstanceAutomatedBackup(
			&rds.DeleteDBInstanceAutomatedBackupInput{
				DbiResourceId: describeOutput.DBInstanceAutomatedBackups[0].DbiResourceId,
			})
		if err != nil {
			return err
		}
	}

	return testAccCheckAWSDBInstanceDestroy(s)
}

func testAccCheckAWSDBInstanceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).rdsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_instance" {
			continue
		}

		// Try to find the Group
		var err error
		resp, err := conn.DescribeDBInstances(
			&rds.DescribeDBInstancesInput{
				DBInstanceIdentifier: aws.String(rs.Primary.ID),
			})

		if err != nil {
			if isAWSErr(err, rds.ErrCodeDBInstanceNotFoundFault, "") {
				continue
			}
			return err
		}

		if len(resp.DBInstances) != 0 &&
			*resp.DBInstances[0].DBInstanceIdentifier == rs.Primary.ID {
			return fmt.Errorf("DB Instance still exists")
		}
	}

	return nil
}

func testAccCheckAWSDBInstanceAttributes(v *rds.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *v.Engine != "mysql" {
			return fmt.Errorf("bad engine: %#v", *v.Engine)
		}

		if *v.EngineVersion == "" {
			return fmt.Errorf("bad engine_version: %#v", *v.EngineVersion)
		}

		if *v.BackupRetentionPeriod != 0 {
			return fmt.Errorf("bad backup_retention_period: %#v", *v.BackupRetentionPeriod)
		}

		return nil
	}
}

func testAccCheckAWSDBInstanceAttributes_MSSQL(v *rds.DBInstance, tz string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *v.Engine != "sqlserver-ex" {
			return fmt.Errorf("bad engine: %#v", *v.Engine)
		}

		rtz := ""
		if v.Timezone != nil {
			rtz = *v.Timezone
		}

		if tz != rtz {
			return fmt.Errorf("Expected (%s) Timezone for MSSQL test, got (%s)", tz, rtz)
		}

		return nil
	}
}

func testAccCheckAWSDBInstanceDomainAttributes(domain string, v *rds.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, dm := range v.DomainMemberships {
			if *dm.FQDN != domain {
				continue
			}

			return nil
		}

		return fmt.Errorf("Domain %s not found in domain memberships", domain)
	}
}

func testAccCheckAWSDBInstanceParameterApplyStatusInSync(dbInstance *rds.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, dbParameterGroup := range dbInstance.DBParameterGroups {
			parameterApplyStatus := aws.StringValue(dbParameterGroup.ParameterApplyStatus)
			if parameterApplyStatus != "in-sync" {
				id := aws.StringValue(dbInstance.DBInstanceIdentifier)
				parameterGroupName := aws.StringValue(dbParameterGroup.DBParameterGroupName)
				return fmt.Errorf("expected DB Instance (%s) Parameter Group (%s) apply status to be: \"in-sync\", got: %q", id, parameterGroupName, parameterApplyStatus)
			}
		}

		return nil
	}
}

func testAccCheckAWSDBInstanceReplicaAttributes(source, replica *rds.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if replica.ReadReplicaSourceDBInstanceIdentifier != nil && *replica.ReadReplicaSourceDBInstanceIdentifier != *source.DBInstanceIdentifier {
			return fmt.Errorf("bad source identifier for replica, expected: '%s', got: '%s'", *source.DBInstanceIdentifier, *replica.ReadReplicaSourceDBInstanceIdentifier)
		}

		return nil
	}
}

func testAccCheckAWSDBInstanceSnapshot(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_instance" {
			continue
		}

		awsClient := testAccProvider.Meta().(*AWSClient)
		conn := awsClient.rdsconn

		log.Printf("[INFO] Trying to locate the DBInstance Final Snapshot")
		snapOutput, err := conn.DescribeDBSnapshots(
			&rds.DescribeDBSnapshotsInput{
				DBSnapshotIdentifier: aws.String(rs.Primary.Attributes["final_snapshot_identifier"]),
			})

		if err != nil {
			return err
		}

		if snapOutput == nil || len(snapOutput.DBSnapshots) == 0 {
			return fmt.Errorf("Snapshot %s not found", rs.Primary.Attributes["final_snapshot_identifier"])
		}

		// verify we have the tags copied to the snapshot
		tagsARN := aws.StringValue(snapOutput.DBSnapshots[0].DBSnapshotArn)
		listTagsOutput, err := conn.ListTagsForResource(&rds.ListTagsForResourceInput{
			ResourceName: aws.String(tagsARN),
		})
		if err != nil {
			return fmt.Errorf("Error retrieving tags for ARN (%s): %s", tagsARN, err)
		}

		if listTagsOutput.TagList == nil || len(listTagsOutput.TagList) == 0 {
			return fmt.Errorf("Tag list is nil or zero: %s", listTagsOutput.TagList)
		}

		var found bool
		for _, t := range listTagsOutput.TagList {
			if *t.Key == "Name" && *t.Value == "tf-tags-db" {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("Expected to find tag Name (%s), but wasn't found. Tags: %s", "tf-tags-db", listTagsOutput.TagList)
		}
		// end tag search

		log.Printf("[INFO] Deleting the Snapshot %s", rs.Primary.Attributes["final_snapshot_identifier"])
		_, err = conn.DeleteDBSnapshot(
			&rds.DeleteDBSnapshotInput{
				DBSnapshotIdentifier: aws.String(rs.Primary.Attributes["final_snapshot_identifier"]),
			})
		if err != nil {
			return err
		}

		resp, err := conn.DescribeDBInstances(
			&rds.DescribeDBInstancesInput{
				DBInstanceIdentifier: aws.String(rs.Primary.ID),
			})

		if err != nil {
			if isAWSErr(err, rds.ErrCodeDBInstanceNotFoundFault, "") {
				continue
			}
			return err

		}

		if len(resp.DBInstances) != 0 && aws.StringValue(resp.DBInstances[0].DBInstanceIdentifier) == rs.Primary.ID {
			return fmt.Errorf("DB Instance still exists")
		}
	}

	return nil
}

func testAccCheckAWSDBInstanceNoSnapshot(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).rdsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_instance" {
			continue
		}

		resp, err := conn.DescribeDBInstances(
			&rds.DescribeDBInstancesInput{
				DBInstanceIdentifier: aws.String(rs.Primary.ID),
			})

		if err != nil && !isAWSErr(err, rds.ErrCodeDBInstanceNotFoundFault, "") {
			return err
		}

		if len(resp.DBInstances) != 0 && aws.StringValue(resp.DBInstances[0].DBInstanceIdentifier) == rs.Primary.ID {
			return fmt.Errorf("DB Instance still exists")
		}

		_, err = conn.DescribeDBSnapshots(
			&rds.DescribeDBSnapshotsInput{
				DBSnapshotIdentifier: aws.String(rs.Primary.Attributes["final_snapshot_identifier"]),
			})

		if err != nil && !isAWSErr(err, rds.ErrCodeDBSnapshotNotFoundFault, "") {
			return err
		}
	}

	return nil
}

func testAccCheckAWSDBInstanceExists(n string, v *rds.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DB Instance ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).rdsconn

		opts := rds.DescribeDBInstancesInput{
			DBInstanceIdentifier: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeDBInstances(&opts)

		if err != nil {
			return err
		}

		if len(resp.DBInstances) != 1 ||
			*resp.DBInstances[0].DBInstanceIdentifier != rs.Primary.ID {
			return fmt.Errorf("DB Instance not found")
		}

		*v = *resp.DBInstances[0]

		return nil
	}
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/8792
func TestAccAWSDBInstance_PerformanceInsightsEnabled_DisabledToEnabled(t *testing.T) {
	var dbInstance rds.DBInstance
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccRDSPerformanceInsightsDefaultVersionPreCheck(t, "mysql") },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_PerformanceInsightsDisabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
					"skip_final_snapshot",
					"final_snapshot_identifier",
				},
			},
			{
				Config: testAccAWSDBInstanceConfig_PerformanceInsightsEnabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_PerformanceInsightsEnabled_EnabledToDisabled(t *testing.T) {
	var dbInstance rds.DBInstance
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccRDSPerformanceInsightsDefaultVersionPreCheck(t, "mysql") },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_PerformanceInsightsEnabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
					"skip_final_snapshot",
					"final_snapshot_identifier",
				},
			},
			{
				Config: testAccAWSDBInstanceConfig_PerformanceInsightsDisabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "false"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_PerformanceInsightsKmsKeyId(t *testing.T) {
	var dbInstance rds.DBInstance
	rName := acctest.RandomWithPrefix("tf-acc-test")
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccRDSPerformanceInsightsDefaultVersionPreCheck(t, "mysql") },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_PerformanceInsightsKmsKeyId(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "performance_insights_kms_key_id", kmsKeyResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
					"skip_final_snapshot",
					"final_snapshot_identifier",
				},
			},
			{
				Config: testAccAWSDBInstanceConfig_PerformanceInsightsKmsKeyIdDisabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "performance_insights_kms_key_id", kmsKeyResourceName, "arn"),
				),
			},
			{
				Config: testAccAWSDBInstanceConfig_PerformanceInsightsKmsKeyId(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "performance_insights_kms_key_id", kmsKeyResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_PerformanceInsightsRetentionPeriod(t *testing.T) {
	var dbInstance rds.DBInstance
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccRDSPerformanceInsightsDefaultVersionPreCheck(t, "mysql") },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_PerformanceInsightsRetentionPeriod(rName, 731),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_retention_period", "731"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
					"skip_final_snapshot",
					"final_snapshot_identifier",
				},
			},
			{
				Config: testAccAWSDBInstanceConfig_PerformanceInsightsRetentionPeriod(rName, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_retention_period", "7"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_ReplicateSourceDb_PerformanceInsightsEnabled(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance

	rName := acctest.RandomWithPrefix("tf-acc-test")
	kmsKeyResourceName := "aws_kms_key.test"
	sourceResourceName := "aws_db_instance.source"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccRDSPerformanceInsightsDefaultVersionPreCheck(t, "mysql") },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_ReplicateSourceDb_PerformanceInsightsEnabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceResourceName, &sourceDbInstance),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					testAccCheckAWSDBInstanceReplicaAttributes(&sourceDbInstance, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "performance_insights_kms_key_id", kmsKeyResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_retention_period", "7"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_SnapshotIdentifier_PerformanceInsightsEnabled(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance
	var dbSnapshot rds.DBSnapshot

	rName := acctest.RandomWithPrefix("tf-acc-test")
	kmsKeyResourceName := "aws_kms_key.test"
	sourceDbResourceName := "aws_db_instance.source"
	snapshotResourceName := "aws_db_snapshot.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccRDSPerformanceInsightsDefaultVersionPreCheck(t, "mysql") },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_SnapshotIdentifier_PerformanceInsightsEnabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceDbResourceName, &sourceDbInstance),
					testAccCheckDbSnapshotExists(snapshotResourceName, &dbSnapshot),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "performance_insights_kms_key_id", kmsKeyResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_retention_period", "7"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_CACertificateIdentifier(t *testing.T) {
	var dbInstance rds.DBInstance

	resourceName := "aws_db_instance.bar"
	dataSourceName := "data.aws_rds_certificate.latest"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_WithCACertificateIdentifier(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "ca_cert_identifier", dataSourceName, "id"),
				),
			},
		},
	})
}

func TestAccAWSDBInstance_RestoreToPointInTime_SourceIdentifier(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance
	sourceName := "aws_db_instance.test"
	resourceName := "aws_db_instance.restore"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_RestoreToPointInTime_SourceIdentifier(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceName, &sourceDbInstance),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"delete_automated_backups",
					"final_snapshot_identifier",
					"latest_restorable_time", // dynamic value of a DBInstance
					"password",
					"restore_to_point_in_time",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func TestAccAWSDBInstance_RestoreToPointInTime_SourceResourceID(t *testing.T) {
	var dbInstance, sourceDbInstance rds.DBInstance
	sourceName := "aws_db_instance.test"
	resourceName := "aws_db_instance.restore"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceConfig_RestoreToPointInTime_SourceResourceID(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(sourceName, &sourceDbInstance),
					testAccCheckAWSDBInstanceExists(resourceName, &dbInstance),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"delete_automated_backups",
					"final_snapshot_identifier",
					"latest_restorable_time", // dynamic value of a DBInstance
					"password",
					"restore_to_point_in_time",
					"skip_final_snapshot",
				},
			},
		},
	})
}

func testAccAWSDBInstanceConfig_orderableClass(engine, version, license string) string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine         = %q
  engine_version = %q
  license_model  = %q
  storage_type   = "standard"

  preferred_instance_classes = ["db.t3.micro", "db.t2.micro", "db.t2.medium"]
}
`, engine, version, license)
}

func testAccAWSDBInstanceConfig_orderableClassMysql() string {
	return testAccAWSDBInstanceConfig_orderableClass("mysql", "5.6.35", "general-public-license")
}

func testAccAWSDBInstanceConfig_orderableClassMariadb() string {
	return testAccAWSDBInstanceConfig_orderableClass("mariadb", "10.2.15", "general-public-license")
}

func testAccAWSDBInstanceConfig_orderableClassSQLServerEx() string {
	return testAccAWSDBInstanceConfig_orderableClass("sqlserver-ex", "14.00.1000.169.v1", "license-included")
}

func testAccAWSDBInstanceConfig_basic() string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMysql(), `
resource "aws_db_instance" "bar" {
  allocated_storage       = 10
  backup_retention_period = 0
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  name                    = "baz"
  parameter_group_name    = "default.mysql5.6"
  password                = "barbarbarbar"
  skip_final_snapshot     = true
  username                = "foo"

  # Maintenance Window is stored in lower case in the API, though not strictly
  # documented. Terraform will downcase this to match (as opposed to throw a
  # validation error).
  maintenance_window = "Fri:09:00-Fri:09:30"
}
`)
}

func testAccAWSDBInstanceConfig_namePrefix() string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMysql(), `
resource "aws_db_instance" "test" {
  allocated_storage   = 10
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier_prefix   = "tf-test-"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "password"
  publicly_accessible = true
  skip_final_snapshot = true
  username            = "root"
}
`)
}

func testAccAWSDBInstanceConfig_generatedName() string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMysql(), `
resource "aws_db_instance" "test" {
  allocated_storage   = 10
  engine              = data.aws_rds_orderable_db_instance.test.engine
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "password"
  publicly_accessible = true
  skip_final_snapshot = true
  username            = "root"
}
`)
}

func testAccAWSDBInstanceConfig_KmsKeyId(rInt int) string {
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

data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  license_model              = "general-public-license"
  storage_type               = "standard"
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t2.medium"]

  supports_storage_encryption = true
}

resource "aws_db_instance" "bar" {
  allocated_storage       = 10
  backup_retention_period = 0
  engine                  = data.aws_rds_engine_version.default.engine
  engine_version          = data.aws_rds_engine_version.default.version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  kms_key_id              = aws_kms_key.foo.arn
  name                    = "baz"
  parameter_group_name    = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  password                = "barbarbarbar"
  skip_final_snapshot     = true
  storage_encrypted       = true
  username                = "foo"

  # Maintenance Window is stored in lower case in the API, though not strictly
  # documented. Terraform will downcase this to match (as opposed to throw a
  # validation error).
  maintenance_window = "Fri:09:00-Fri:09:30"
}
`, rInt)
}

func testAccAWSDBInstanceConfig_WithCACertificateIdentifier() string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMysql(), `
data "aws_rds_certificate" "latest" {
  latest_valid_till = true
}

resource "aws_db_instance" "bar" {
  allocated_storage   = 10
  apply_immediately   = true
  ca_cert_identifier  = data.aws_rds_certificate.latest.id
  engine              = data.aws_rds_orderable_db_instance.test.engine
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  name                = "baz"
  password            = "barbarbarbar"
  skip_final_snapshot = true
  username            = "foo"
}
`)
}

func testAccAWSDBInstanceConfig_WithOptionGroup(rName string) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMysql(), fmt.Sprintf(`
resource "aws_db_option_group" "test" {
  engine_name              = data.aws_rds_orderable_db_instance.test.engine
  major_engine_version     = "5.6"
  name                     = %[1]q
  option_group_description = "Test option group for terraform"
}

resource "aws_db_instance" "bar" {
  allocated_storage   = 10
  engine              = aws_db_option_group.test.engine_name
  engine_version      = aws_db_option_group.test.major_engine_version
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  name                = "baz"
  option_group_name   = aws_db_option_group.test.name
  password            = "barbarbarbar"
  skip_final_snapshot = true
  username            = "foo"
}
`, rName))
}

func testAccAWSDBInstanceConfig_IAMAuth(n int) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  license_model              = "general-public-license"
  storage_type               = "standard"
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t2.medium"]

  supports_iam_database_authentication = true
}

resource "aws_db_instance" "bar" {
  identifier                          = "foobarbaz-test-terraform-%d"
  allocated_storage                   = 10
  engine                              = data.aws_rds_engine_version.default.engine
  engine_version                      = data.aws_rds_engine_version.default.version
  instance_class                      = data.aws_rds_orderable_db_instance.test.instance_class
  name                                = "baz"
  password                            = "barbarbarbar"
  username                            = "foo"
  backup_retention_period             = 0
  skip_final_snapshot                 = true
  parameter_group_name                = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  iam_database_authentication_enabled = true
}
`, n)
}

func testAccAWSDBInstanceConfig_FinalSnapshotIdentifier_SkipFinalSnapshot() string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMysql(), fmt.Sprintf(`
resource "aws_db_instance" "snapshot" {
  identifier = "tf-acc-test-%[1]d"

  allocated_storage       = 5
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  name                    = "baz"
  password                = "barbarbarbar"
  username                = "foo"
  backup_retention_period = 1

  publicly_accessible = true

  parameter_group_name = "default.mysql5.6"

  skip_final_snapshot       = true
  final_snapshot_identifier = "tf-acc-test-%[1]d"
}
`, acctest.RandInt()))
}

func testAccAWSDBInstanceConfig_S3Import(bucketName string, bucketPrefix string, uniqueId string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_s3_bucket" "xtrabackup" {
  bucket = %[1]q
}

resource "aws_s3_bucket_object" "xtrabackup_db" {
  bucket = aws_s3_bucket.xtrabackup.id
  key    = "%[2]s/mysql-5-6-xtrabackup.tar.gz"
  source = "./testdata/mysql-5-6-xtrabackup.tar.gz"
  etag   = filemd5("./testdata/mysql-5-6-xtrabackup.tar.gz")
}

data "aws_partition" "current" {}

resource "aws_iam_role" "rds_s3_access_role" {
  name = "%[3]s-role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "rds.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "test" {
  name = "%[3]s-policy"

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
  name = "%[3]s-policy-attachment"

  roles = [
    aws_iam_role.rds_s3_access_role.name,
  ]

  policy_arn = aws_iam_policy.test.arn
}

#  Make sure EVERYTHING required is here...
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-db-instance-with-subnet-group"
  }
}

resource "aws_subnet" "foo" {
  cidr_block        = "10.1.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-db-instance-with-subnet-group-1"
  }
}

resource "aws_subnet" "bar" {
  cidr_block        = "10.1.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
  vpc_id            = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-db-instance-with-subnet-group-2"
  }
}

resource "aws_db_subnet_group" "foo" {
  name       = "%[3]s-subnet-group"
  subnet_ids = [aws_subnet.foo.id, aws_subnet.bar.id]

  tags = {
    Name = "tf-dbsubnet-group-test"
  }
}

data "aws_rds_orderable_db_instance" "test" {
  engine         = "mysql"
  engine_version = "5.6.41"
  license_model  = "general-public-license"
  storage_type   = "standard"

  # instance class db.t2.micro is not supported for restoring from S3
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t2.medium"]
}

resource "aws_db_instance" "s3" {
  identifier = "%[3]s-db"

  allocated_storage          = 5
  engine                     = data.aws_rds_orderable_db_instance.test.engine
  engine_version             = "5.6"
  auto_minor_version_upgrade = true
  instance_class             = data.aws_rds_orderable_db_instance.test.instance_class
  name                       = "baz"
  password                   = "barbarbarbar"
  publicly_accessible        = false
  username                   = "foo"
  backup_retention_period    = 0

  parameter_group_name = "default.mysql5.6"
  skip_final_snapshot  = true
  multi_az             = false
  db_subnet_group_name = aws_db_subnet_group.foo.id

  s3_import {
    source_engine         = data.aws_rds_orderable_db_instance.test.engine
    source_engine_version = "5.6"

    bucket_name    = aws_s3_bucket.xtrabackup.bucket
    bucket_prefix  = %[2]q
    ingestion_role = aws_iam_role.rds_s3_access_role.arn
  }
}
`, bucketName, bucketPrefix, uniqueId))
}

func testAccAWSDBInstanceConfig_FinalSnapshotIdentifier(rInt int) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMysql(), fmt.Sprintf(`
resource "aws_db_instance" "snapshot" {
  identifier = "tf-snapshot-%[1]d"

  allocated_storage       = 5
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  name                    = "baz"
  password                = "barbarbarbar"
  publicly_accessible     = true
  username                = "foo"
  backup_retention_period = 1

  parameter_group_name = "default.mysql5.6"

  copy_tags_to_snapshot     = true
  final_snapshot_identifier = "foobarbaz-test-terraform-final-snapshot-%[1]d"

  tags = {
    Name = "tf-tags-db"
  }
}
`, rInt))
}

func testAccAWSDbInstanceConfig_MonitoringInterval(rName string, monitoringInterval int) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "monitoring.rds.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonRDSEnhancedMonitoringRole"
  role       = aws_iam_role.test.name
}

data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  license_model              = "general-public-license"
  storage_type               = "standard"
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t2.medium"]

  supports_enhanced_monitoring = true
}

resource "aws_db_instance" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]

  allocated_storage   = 5
  engine              = data.aws_rds_engine_version.default.engine
  engine_version      = data.aws_rds_engine_version.default.version
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  monitoring_interval = %[2]d
  monitoring_role_arn = aws_iam_role.test.arn
  name                = "baz"
  password            = "barbarbarbar"
  skip_final_snapshot = true
  username            = "foo"
}
`, rName, monitoringInterval)
}

func testAccAWSDbInstanceConfig_MonitoringRoleArnRemoved(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  license_model              = "general-public-license"
  storage_type               = "standard"
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t2.medium"]

  supports_enhanced_monitoring = true
}

resource "aws_db_instance" "test" {
  allocated_storage   = 5
  engine              = data.aws_rds_engine_version.default.engine
  engine_version      = data.aws_rds_engine_version.default.version
  identifier          = %q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  name                = "baz"
  password            = "barbarbarbar"
  skip_final_snapshot = true
  username            = "foo"
}
`, rName)
}

func testAccAWSDbInstanceConfig_MonitoringRoleArn(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "monitoring.rds.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonRDSEnhancedMonitoringRole"
  role       = aws_iam_role.test.name
}

data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  license_model              = "general-public-license"
  storage_type               = "standard"
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t2.medium"]

  supports_enhanced_monitoring = true
}

resource "aws_db_instance" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]

  allocated_storage   = 5
  engine              = data.aws_rds_engine_version.default.engine
  engine_version      = data.aws_rds_engine_version.default.version
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  monitoring_interval = 5
  monitoring_role_arn = aws_iam_role.test.arn
  name                = "baz"
  password            = "barbarbarbar"
  skip_final_snapshot = true
  username            = "foo"
}
`, rName)
}

const testAccAWSDBInstanceBaseConfig = `
data "aws_rds_engine_version" "test" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.test.engine
  engine_version             = data.aws_rds_engine_version.test.version
  preferred_instance_classes = ["db.t3.micro", "db.t2.micro", "db.t3.small"]
}

resource "aws_db_instance" "test" {
  allocated_storage       = 10
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  name                    = "baz"
  parameter_group_name    = "default.${data.aws_rds_engine_version.test.parameter_group_family}"
  password                = "barbarbarbar"
  skip_final_snapshot     = true
  username                = "foo"
}
`

func testAccAWSDBInstanceConfig_RestoreToPointInTime_SourceIdentifier() string {
	return composeConfig(
		testAccAWSDBInstanceBaseConfig,
		fmt.Sprintf(`
resource "aws_db_instance" "restore" {
  identifier     = "tf-acc-test-point-in-time-restore-%d"
  instance_class = aws_db_instance.test.instance_class
  restore_to_point_in_time {
    source_db_instance_identifier = aws_db_instance.test.identifier
    use_latest_restorable_time    = true
  }
  skip_final_snapshot = true
}
`, acctest.RandInt()))
}

func testAccAWSDBInstanceConfig_RestoreToPointInTime_SourceResourceID() string {
	return composeConfig(
		testAccAWSDBInstanceBaseConfig,
		fmt.Sprintf(`
resource "aws_db_instance" "restore" {
  identifier     = "tf-acc-test-point-in-time-restore-%d"
  instance_class = aws_db_instance.test.instance_class
  restore_to_point_in_time {
    source_dbi_resource_id     = aws_db_instance.test.resource_id
    use_latest_restorable_time = true
  }
  skip_final_snapshot = true
}
`, acctest.RandInt()))
}

func testAccAWSDBInstanceConfig_SnapshotInstanceConfig_iopsUpdate(rName string, iops int) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  license_model              = "general-public-license"
  preferred_instance_classes = ["db.t3.micro", "db.t2.micro", "db.t2.medium"]

  storage_type  = "io1"
  supports_iops = true
}

resource "aws_db_instance" "bar" {
  identifier           = "mydb-rds-%s"
  engine               = data.aws_rds_engine_version.default.engine
  engine_version       = data.aws_rds_engine_version.default.version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  name                 = "mydb"
  username             = "foo"
  password             = "barbarbar"
  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  skip_final_snapshot  = true

  apply_immediately = true

  storage_type      = data.aws_rds_orderable_db_instance.test.storage_type
  allocated_storage = 200
  iops              = %d
}
`, rName, iops)
}

func testAccAWSDBInstanceConfig_SnapshotInstanceConfig_mysqlPort(rName string) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMysql(), fmt.Sprintf(`
resource "aws_db_instance" "bar" {
  identifier           = "mydb-rds-%s"
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  name                 = "mydb"
  username             = "foo"
  password             = "barbarbar"
  parameter_group_name = "default.mysql5.6"
  port                 = 3306
  allocated_storage    = 10
  skip_final_snapshot  = true

  apply_immediately = true
}
`, rName))
}

func testAccAWSDBInstanceConfig_SnapshotInstanceConfig_updateMysqlPort(rName string) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMysql(), fmt.Sprintf(`
resource "aws_db_instance" "bar" {
  identifier           = "mydb-rds-%s"
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  name                 = "mydb"
  username             = "foo"
  password             = "barbarbar"
  parameter_group_name = "default.mysql5.6"
  port                 = 3305
  allocated_storage    = 10
  skip_final_snapshot  = true

  apply_immediately = true
}
`, rName))
}

func testAccAWSDBInstanceConfig_WithSubnetGroup(rName string) string {
	return composeConfig(
		testAccAWSDBInstanceConfig_orderableClassMysql(),
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-db-instance-with-subnet-group"
  }
}

resource "aws_subnet" "foo" {
  cidr_block        = "10.1.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-db-instance-with-subnet-group-1"
  }
}

resource "aws_subnet" "bar" {
  cidr_block        = "10.1.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
  vpc_id            = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-db-instance-with-subnet-group-2"
  }
}

resource "aws_db_subnet_group" "foo" {
  name       = "foo-%[1]s"
  subnet_ids = [aws_subnet.foo.id, aws_subnet.bar.id]

  tags = {
    Name = "tf-dbsubnet-group-test"
  }
}

resource "aws_db_instance" "bar" {
  identifier           = "mydb-rds-%[1]s"
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  name                 = "mydb"
  username             = "foo"
  password             = "barbarbar"
  parameter_group_name = "default.mysql5.6"
  db_subnet_group_name = aws_db_subnet_group.foo.name
  port                 = 3305
  allocated_storage    = 10
  skip_final_snapshot  = true

  backup_retention_period = 0
  apply_immediately       = true
}
`, rName))
}

func testAccAWSDBInstanceConfig_WithSubnetGroupUpdated(rName string) string {
	return composeConfig(
		testAccAWSDBInstanceConfig_orderableClassMysql(),
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-db-instance-with-subnet-group-updated-foo"
  }
}

resource "aws_vpc" "bar" {
  cidr_block = "10.10.0.0/16"

  tags = {
    Name = "terraform-testacc-db-instance-with-subnet-group-updated-bar"
  }
}

resource "aws_subnet" "foo" {
  cidr_block        = "10.1.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-db-instance-with-subnet-group-1"
  }
}

resource "aws_subnet" "bar" {
  cidr_block        = "10.1.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
  vpc_id            = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-db-instance-with-subnet-group-2"
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.10.3.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
  vpc_id            = aws_vpc.bar.id

  tags = {
    Name = "tf-acc-db-instance-with-subnet-group-3"
  }
}

resource "aws_subnet" "another_test" {
  cidr_block        = "10.10.4.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.bar.id

  tags = {
    Name = "tf-acc-db-instance-with-subnet-group-4"
  }
}

resource "aws_db_subnet_group" "foo" {
  name       = "foo-%[1]s"
  subnet_ids = [aws_subnet.foo.id, aws_subnet.bar.id]

  tags = {
    Name = "tf-dbsubnet-group-test"
  }
}

resource "aws_db_subnet_group" "bar" {
  name       = "bar-%[1]s"
  subnet_ids = [aws_subnet.test.id, aws_subnet.another_test.id]

  tags = {
    Name = "tf-dbsubnet-group-test-updated"
  }
}

resource "aws_db_instance" "bar" {
  identifier           = "mydb-rds-%[1]s"
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  name                 = "mydb"
  username             = "foo"
  password             = "barbarbar"
  parameter_group_name = "default.mysql5.6"
  db_subnet_group_name = aws_db_subnet_group.bar.name
  port                 = 3305
  allocated_storage    = 10
  skip_final_snapshot  = true

  backup_retention_period = 0

  apply_immediately = true
}
`, rName))
}

func testAccAWSDBInstanceConfig_MSSQL_timezone(rInt int) string {
	return composeConfig(
		testAccAWSDBInstanceConfig_orderableClassSQLServerEx(),
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-db-instance-mssql-timezone"
  }
}

resource "aws_db_subnet_group" "rds_one" {
  name        = "tf_acc_test_%[1]d"
  description = "db subnets for rds_one"

  subnet_ids = [aws_subnet.main.id, aws_subnet.other.id]
}

resource "aws_subnet" "main" {
  vpc_id            = aws_vpc.foo.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.1.1.0/24"

  tags = {
    Name = "tf-acc-db-instance-mssql-timezone-main"
  }
}

resource "aws_subnet" "other" {
  vpc_id            = aws_vpc.foo.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "10.1.2.0/24"

  tags = {
    Name = "tf-acc-db-instance-mssql-timezone-other"
  }
}

resource "aws_db_instance" "mssql" {
  identifier = "tf-test-mssql-%[1]d"

  db_subnet_group_name = aws_db_subnet_group.rds_one.name

  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  allocated_storage       = 20
  username                = "somecrazyusername"
  password                = "somecrazypassword"
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  backup_retention_period = 0
  skip_final_snapshot     = true

  #publicly_accessible = true

  vpc_security_group_ids = [aws_security_group.rds-mssql.id]
}

resource "aws_security_group" "rds-mssql" {
  name = "tf-rds-mssql-test-%[1]d"

  description = "TF Testing"
  vpc_id      = aws_vpc.foo.id
}

resource "aws_security_group_rule" "rds-mssql-1" {
  type        = "egress"
  from_port   = 0
  to_port     = 0
  protocol    = "-1"
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = aws_security_group.rds-mssql.id
}
`, rInt))
}

func testAccAWSDBInstanceConfig_MSSQL_timezone_AKST(rInt int) string {
	return composeConfig(
		testAccAWSDBInstanceConfig_orderableClassSQLServerEx(),
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-db-instance-mssql-timezone-akst"
  }
}

resource "aws_db_subnet_group" "rds_one" {
  name        = "tf_acc_test_%[1]d"
  description = "db subnets for rds_one"

  subnet_ids = [aws_subnet.main.id, aws_subnet.other.id]
}

resource "aws_subnet" "main" {
  vpc_id            = aws_vpc.foo.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.1.1.0/24"

  tags = {
    Name = "tf-acc-db-instance-mssql-timezone-akst-main"
  }
}

resource "aws_subnet" "other" {
  vpc_id            = aws_vpc.foo.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "10.1.2.0/24"

  tags = {
    Name = "tf-acc-db-instance-mssql-timezone-akst-other"
  }
}

resource "aws_db_instance" "mssql" {
  identifier = "tf-test-mssql-%[1]d"

  db_subnet_group_name = aws_db_subnet_group.rds_one.name

  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  allocated_storage       = 20
  username                = "somecrazyusername"
  password                = "somecrazypassword"
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  backup_retention_period = 0
  skip_final_snapshot     = true

  #publicly_accessible = true

  vpc_security_group_ids = [aws_security_group.rds-mssql.id]
  timezone               = "Alaskan Standard Time"
}

resource "aws_security_group" "rds-mssql" {
  name = "tf-rds-mssql-test-%[1]d"

  description = "TF Testing"
  vpc_id      = aws_vpc.foo.id
}

resource "aws_security_group_rule" "rds-mssql-1" {
  type        = "egress"
  from_port   = 0
  to_port     = 0
  protocol    = "-1"
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = aws_security_group.rds-mssql.id
}
`, rInt))
}

func testAccAWSDBInstanceConfig_MSSQLDomain(rInt int) string {
	return composeConfig(
		testAccAWSDBInstanceConfig_orderableClassSQLServerEx(),
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-db-instance-mssql-domain"
  }
}

resource "aws_db_subnet_group" "rds_one" {
  name        = "tf_acc_test_%[1]d"
  description = "db subnets for rds_one"

  subnet_ids = [aws_subnet.main.id, aws_subnet.other.id]
}

resource "aws_subnet" "main" {
  vpc_id            = aws_vpc.foo.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.1.1.0/24"

  tags = {
    Name = "tf-acc-db-instance-mssql-domain-main"
  }
}

resource "aws_subnet" "other" {
  vpc_id            = aws_vpc.foo.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "10.1.2.0/24"

  tags = {
    Name = "tf-acc-db-instance-mssql-domain-other"
  }
}

resource "aws_db_instance" "mssql" {
  identifier = "tf-test-mssql-%[1]d"

  db_subnet_group_name = aws_db_subnet_group.rds_one.name

  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  allocated_storage       = 20
  username                = "somecrazyusername"
  password                = "somecrazypassword"
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  backup_retention_period = 0
  skip_final_snapshot     = true

  domain               = aws_directory_service_directory.foo.id
  domain_iam_role_name = aws_iam_role.role.name

  vpc_security_group_ids = [aws_security_group.rds-mssql.id]
}

resource "aws_security_group" "rds-mssql" {
  name = "tf-rds-mssql-test-%[1]d"

  description = "TF Testing"
  vpc_id      = aws_vpc.foo.id
}

resource "aws_security_group_rule" "rds-mssql-1" {
  type        = "egress"
  from_port   = 0
  to_port     = 0
  protocol    = "-1"
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = aws_security_group.rds-mssql.id
}

resource "aws_directory_service_directory" "foo" {
  name     = "terraformtesting.com"
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.foo.id
    subnet_ids = [aws_subnet.main.id, aws_subnet.other.id]
  }
}

resource "aws_directory_service_directory" "bar" {
  name     = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.foo.id
    subnet_ids = [aws_subnet.main.id, aws_subnet.other.id]
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role" "role" {
  name = "tf-acc-db-instance-mssql-domain-role-%[1]d"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "rds.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "attatch-policy" {
  role       = aws_iam_role.role.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonRDSDirectoryServiceAccess"
}
`, rInt))
}

func testAccAWSDBInstanceConfig_MSSQLUpdateDomain(rInt int) string {
	return composeConfig(
		testAccAWSDBInstanceConfig_orderableClassSQLServerEx(),
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-db-instance-mssql-domain"
  }
}

resource "aws_db_subnet_group" "rds_one" {
  name        = "tf_acc_test_%[1]d"
  description = "db subnets for rds_one"

  subnet_ids = [aws_subnet.main.id, aws_subnet.other.id]
}

resource "aws_subnet" "main" {
  vpc_id            = aws_vpc.foo.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.1.1.0/24"

  tags = {
    Name = "tf-acc-db-instance-mssql-domain-main"
  }
}

resource "aws_subnet" "other" {
  vpc_id            = aws_vpc.foo.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "10.1.2.0/24"

  tags = {
    Name = "tf-acc-db-instance-mssql-domain-other"
  }
}

resource "aws_db_instance" "mssql" {
  identifier = "tf-test-mssql-%[1]d"

  db_subnet_group_name = aws_db_subnet_group.rds_one.name

  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  allocated_storage       = 20
  username                = "somecrazyusername"
  password                = "somecrazypassword"
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  backup_retention_period = 0
  skip_final_snapshot     = true
  apply_immediately       = true

  domain               = aws_directory_service_directory.bar.id
  domain_iam_role_name = aws_iam_role.role.name

  vpc_security_group_ids = [aws_security_group.rds-mssql.id]
}

resource "aws_security_group" "rds-mssql" {
  name = "tf-rds-mssql-test-%[1]d"

  description = "TF Testing"
  vpc_id      = aws_vpc.foo.id
}

resource "aws_security_group_rule" "rds-mssql-1" {
  type        = "egress"
  from_port   = 0
  to_port     = 0
  protocol    = "-1"
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = aws_security_group.rds-mssql.id
}

resource "aws_directory_service_directory" "foo" {
  name     = "terraformtesting.com"
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.foo.id
    subnet_ids = [aws_subnet.main.id, aws_subnet.other.id]
  }
}

resource "aws_directory_service_directory" "bar" {
  name     = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.foo.id
    subnet_ids = [aws_subnet.main.id, aws_subnet.other.id]
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role" "role" {
  name = "tf-acc-db-instance-mssql-domain-role-%[1]d"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "rds.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "attatch-policy" {
  role       = aws_iam_role.role.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonRDSDirectoryServiceAccess"
}
`, rInt))
}

func testAccAWSDBInstanceConfig_MSSQLDomainSnapshotRestore(rInt int) string {
	return composeConfig(
		testAccAWSDBInstanceConfig_orderableClassSQLServerEx(),
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-db-instance-mssql-domain"
  }
}

resource "aws_db_subnet_group" "rds_one" {
  name        = "tf_acc_test_%[1]d"
  description = "db subnets for rds_one"

  subnet_ids = [aws_subnet.main.id, aws_subnet.other.id]
}

resource "aws_subnet" "main" {
  vpc_id            = aws_vpc.foo.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.1.1.0/24"

  tags = {
    Name = "tf-acc-db-instance-mssql-domain-main"
  }
}

resource "aws_subnet" "other" {
  vpc_id            = aws_vpc.foo.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "10.1.2.0/24"

  tags = {
    Name = "tf-acc-db-instance-mssql-domain-other"
  }
}

resource "aws_db_instance" "mssql" {
  allocated_storage   = 20
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "tf-test-mssql-%[1]d"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "somecrazypassword"
  skip_final_snapshot = true
  username            = "somecrazyusername"
}

resource "aws_db_snapshot" "mssql-snap" {
  db_instance_identifier = aws_db_instance.mssql.id
  db_snapshot_identifier = "tf-acc-test-%[1]d"
}

resource "aws_db_instance" "mssql_restore" {
  identifier = "tf-test-mssql-%[1]d-restore"

  db_subnet_group_name = aws_db_subnet_group.rds_one.name

  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  allocated_storage       = 20
  username                = "somecrazyusername"
  password                = "somecrazypassword"
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  backup_retention_period = 0
  skip_final_snapshot     = true
  snapshot_identifier     = aws_db_snapshot.mssql-snap.id

  domain               = aws_directory_service_directory.foo.id
  domain_iam_role_name = aws_iam_role.role.name

  apply_immediately      = true
  vpc_security_group_ids = [aws_security_group.rds-mssql.id]
}

resource "aws_security_group" "rds-mssql" {
  name = "tf-rds-mssql-test-%[1]d"

  description = "TF Testing"
  vpc_id      = aws_vpc.foo.id
}

resource "aws_security_group_rule" "rds-mssql-1" {
  type        = "egress"
  from_port   = 0
  to_port     = 0
  protocol    = "-1"
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = aws_security_group.rds-mssql.id
}

resource "aws_directory_service_directory" "foo" {
  name     = "terraformtesting.com"
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.foo.id
    subnet_ids = [aws_subnet.main.id, aws_subnet.other.id]
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role" "role" {
  name = "tf-acc-db-instance-mssql-domain-role-%[1]d"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "rds.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "attatch-policy" {
  role       = aws_iam_role.role.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonRDSDirectoryServiceAccess"
}
`, rInt))
}

func testAccAWSDBInstanceConfig_MySQLSnapshotRestoreWithEngineVersion(rInt int) string {
	return composeConfig(
		testAccAWSDBInstanceConfig_orderableClassMysql(),
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-db-instance-mysql-domain"
  }
}

resource "aws_db_subnet_group" "rds_one" {
  name        = "tf_acc_test_%[1]d"
  description = "db subnets for rds_one"

  subnet_ids = [aws_subnet.main.id, aws_subnet.other.id]
}

resource "aws_subnet" "main" {
  vpc_id            = aws_vpc.foo.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.1.1.0/24"

  tags = {
    Name = "tf-acc-db-instance-mysql-domain-main"
  }
}

resource "aws_subnet" "other" {
  vpc_id            = aws_vpc.foo.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "10.1.2.0/24"

  tags = {
    Name = "tf-acc-db-instance-mysql-domain-other"
  }
}

resource "aws_db_instance" "mysql" {
  allocated_storage   = 20
  engine              = data.aws_rds_orderable_db_instance.test.engine
  engine_version      = data.aws_rds_orderable_db_instance.test.engine_version
  identifier          = "tf-test-mysql-%[1]d"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "password"
  skip_final_snapshot = true
  username            = "root"
}

resource "aws_db_snapshot" "mysql-snap" {
  db_instance_identifier = aws_db_instance.mysql.id
  db_snapshot_identifier = "tf-acc-test-%[1]d"
}

resource "aws_db_instance" "mysql_restore" {
  identifier = "tf-test-mysql-%[1]d-restore"

  db_subnet_group_name = aws_db_subnet_group.rds_one.name

  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  allocated_storage       = 20
  username                = "root"
  password                = "password"
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = "5.6.41"
  backup_retention_period = 0
  skip_final_snapshot     = true
  snapshot_identifier     = aws_db_snapshot.mysql-snap.id

  apply_immediately      = true
  vpc_security_group_ids = [aws_security_group.rds-mysql.id]
}

resource "aws_security_group" "rds-mysql" {
  name = "tf-rds-mysql-test-%[1]d"

  description = "TF Testing"
  vpc_id      = aws_vpc.foo.id
}

resource "aws_security_group_rule" "rds-mysql-1" {
  type        = "egress"
  from_port   = 0
  to_port     = 0
  protocol    = "-1"
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = aws_security_group.rds-mysql.id
}
`, rInt))
}

func testAccAWSDBInstanceConfig_AllowMajorVersionUpgrade(rName string, allowMajorVersionUpgrade bool) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMysql(), fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage           = 10
  allow_major_version_upgrade = %t
  engine                      = data.aws_rds_orderable_db_instance.test.engine
  engine_version              = data.aws_rds_orderable_db_instance.test.engine_version
  identifier                  = %q
  instance_class              = data.aws_rds_orderable_db_instance.test.instance_class
  name                        = "baz"
  password                    = "barbarbarbar"
  skip_final_snapshot         = true
  username                    = "foo"
}
`, allowMajorVersionUpgrade, rName))
}

func testAccAWSDBInstanceConfig_AutoMinorVersion() string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMysql(), fmt.Sprintf(`
resource "aws_db_instance" "bar" {
  identifier          = "foobarbaz-test-terraform-%d"
  allocated_storage   = 10
  engine              = data.aws_rds_orderable_db_instance.test.engine
  engine_version      = "5.6"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  name                = "baz"
  password            = "barbarbarbar"
  username            = "foo"
  skip_final_snapshot = true
}
`, acctest.RandInt()))
}

func testAccAWSDBInstanceConfig_CloudwatchLogsExportConfiguration(rInt int) string {
	return composeConfig(
		testAccAWSDBInstanceConfig_orderableClassMysql(),
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-db-instance-enable-cloudwatch"
  }
}

resource "aws_db_subnet_group" "rds_one" {
  name        = "tf_acc_test_%[1]d"
  description = "db subnets for rds_one"

  subnet_ids = [aws_subnet.main.id, aws_subnet.other.id]
}

resource "aws_subnet" "main" {
  vpc_id            = aws_vpc.foo.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.1.1.0/24"

  tags = {
    Name = "tf-acc-db-instance-enable-cloudwatch-main"
  }
}

resource "aws_subnet" "other" {
  vpc_id            = aws_vpc.foo.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "10.1.2.0/24"

  tags = {
    Name = "tf-acc-db-instance-enable-cloudwatch-other"
  }
}

resource "aws_db_instance" "bar" {
  identifier = "foobarbaz-test-terraform-%[1]d"

  db_subnet_group_name = aws_db_subnet_group.rds_one.name
  allocated_storage    = 10
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  name                 = "baz"
  password             = "barbarbarbar"
  username             = "foo"
  skip_final_snapshot  = true

  enabled_cloudwatch_logs_exports = [
    "audit",
    "error",
  ]
}
`, rInt))
}

func testAccAWSDBInstanceConfig_CloudwatchLogsExportConfigurationAdd(rInt int) string {
	return composeConfig(
		testAccAWSDBInstanceConfig_orderableClassMysql(),
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-db-instance-enable-cloudwatch"
  }
}

resource "aws_db_subnet_group" "rds_one" {
  name        = "tf_acc_test_%[1]d"
  description = "db subnets for rds_one"

  subnet_ids = [aws_subnet.main.id, aws_subnet.other.id]
}

resource "aws_subnet" "main" {
  vpc_id            = aws_vpc.foo.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.1.1.0/24"

  tags = {
    Name = "tf-acc-db-instance-enable-cloudwatch-main"
  }
}

resource "aws_subnet" "other" {
  vpc_id            = aws_vpc.foo.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "10.1.2.0/24"

  tags = {
    Name = "tf-acc-db-instance-enable-cloudwatch-other"
  }
}

resource "aws_db_instance" "bar" {
  identifier = "foobarbaz-test-terraform-%[1]d"

  db_subnet_group_name = aws_db_subnet_group.rds_one.name
  allocated_storage    = 10
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  name                 = "baz"
  password             = "barbarbarbar"
  username             = "foo"
  skip_final_snapshot  = true

  apply_immediately = true

  enabled_cloudwatch_logs_exports = [
    "audit",
    "error",
    "general",
  ]
}
`, rInt))
}

func testAccAWSDBInstanceConfig_CloudwatchLogsExportConfigurationModify(rInt int) string {
	return composeConfig(
		testAccAWSDBInstanceConfig_orderableClassMysql(),
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-db-instance-enable-cloudwatch"
  }
}

resource "aws_db_subnet_group" "rds_one" {
  name        = "tf_acc_test_%[1]d"
  description = "db subnets for rds_one"

  subnet_ids = [aws_subnet.main.id, aws_subnet.other.id]
}

resource "aws_subnet" "main" {
  vpc_id            = aws_vpc.foo.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.1.1.0/24"

  tags = {
    Name = "tf-acc-db-instance-enable-cloudwatch-main"
  }
}

resource "aws_subnet" "other" {
  vpc_id            = aws_vpc.foo.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "10.1.2.0/24"

  tags = {
    Name = "tf-acc-db-instance-enable-cloudwatch-other"
  }
}

resource "aws_db_instance" "bar" {
  identifier = "foobarbaz-test-terraform-%[1]d"

  db_subnet_group_name = aws_db_subnet_group.rds_one.name
  allocated_storage    = 10
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  name                 = "baz"
  password             = "barbarbarbar"
  username             = "foo"
  skip_final_snapshot  = true

  apply_immediately = true

  enabled_cloudwatch_logs_exports = [
    "audit",
    "general",
    "slowquery",
  ]
}
`, rInt))
}

func testAccAWSDBInstanceConfig_CloudwatchLogsExportConfigurationDelete(rInt int) string {
	return composeConfig(
		testAccAWSDBInstanceConfig_orderableClassMysql(),
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-db-instance-enable-cloudwatch"
  }
}

resource "aws_db_subnet_group" "rds_one" {
  name        = "tf_acc_test_%[1]d"
  description = "db subnets for rds_one"

  subnet_ids = [aws_subnet.main.id, aws_subnet.other.id]
}

resource "aws_subnet" "main" {
  vpc_id            = aws_vpc.foo.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.1.1.0/24"

  tags = {
    Name = "tf-acc-db-instance-enable-cloudwatch-main"
  }
}

resource "aws_subnet" "other" {
  vpc_id            = aws_vpc.foo.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "10.1.2.0/24"

  tags = {
    Name = "tf-acc-db-instance-enable-cloudwatch-other"
  }
}

resource "aws_db_instance" "bar" {
  identifier = "foobarbaz-test-terraform-%[1]d"

  db_subnet_group_name = aws_db_subnet_group.rds_one.name
  allocated_storage    = 10
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  name                 = "baz"
  password             = "barbarbarbar"
  username             = "foo"
  skip_final_snapshot  = true

  apply_immediately = true
}
`, rInt))
}

func testAccAWSDBInstanceConfig_Ec2Classic(rInt int) string {
	return composeConfig(
		testAccEc2ClassicRegionProviderConfig(),
		fmt.Sprintf(`
# EC2-Classic specific
data "aws_rds_orderable_db_instance" "test" {
  engine                     = "mysql"
  engine_version             = "5.6.41"
  preferred_instance_classes = ["db.m3.medium", "db.m3.large", "db.r3.large"]
}

resource "aws_db_instance" "bar" {
  identifier           = "foobarbaz-test-terraform-%[1]d"
  allocated_storage    = 10
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  storage_type         = data.aws_rds_orderable_db_instance.test.storage_type
  name                 = "baz"
  password             = "barbarbarbar"
  username             = "foo"
  publicly_accessible  = true
  security_group_names = ["default"]
  parameter_group_name = "default.mysql5.6"
  skip_final_snapshot  = true
}
`, rInt))
}

func testAccAWSDBInstanceConfig_MariaDB(rName string) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = %q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}
`, rName))
}

func testAccAWSDBInstanceConfig_DbSubnetGroupName(rName string) string {
	return composeConfig(
		testAccAWSDBInstanceConfig_orderableClassMysql(),
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_db_instance" "test" {
  allocated_storage    = 5
  db_subnet_group_name = aws_db_subnet_group.test.name
  engine               = data.aws_rds_orderable_db_instance.test.engine
  identifier           = %[1]q
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  password             = "avoid-plaintext-passwords"
  username             = "tfacctest"
  skip_final_snapshot  = true
}
`, rName))
}

func testAccAWSDBInstanceConfig_DbSubnetGroupName_RamShared(rName string) string {
	return composeConfig(
		testAccAWSDBInstanceConfig_orderableClassMysql(),
		testAccAlternateAccountProviderConfig(),
		fmt.Sprintf(`
data "aws_availability_zones" "alternate" {
  provider = "awsalternate"

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_organizations_organization" "test" {}

resource "aws_vpc" "test" {
  provider = "awsalternate"

  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count    = 2
  provider = "awsalternate"

  availability_zone = data.aws_availability_zones.alternate.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ram_resource_share" "test" {
  provider = "awsalternate"

  name = %[1]q
}

resource "aws_ram_principal_association" "test" {
  provider = "awsalternate"

  principal          = data.aws_organizations_organization.test.arn
  resource_share_arn = aws_ram_resource_share.test.arn
}

resource "aws_ram_resource_association" "test" {
  count    = 2
  provider = "awsalternate"

  resource_arn       = aws_subnet.test[count.index].arn
  resource_share_arn = aws_ram_resource_share.test.id
}

resource "aws_db_subnet_group" "test" {
  depends_on = [aws_ram_principal_association.test, aws_ram_resource_association.test]

  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_security_group" "test" {
  depends_on = [aws_ram_principal_association.test, aws_ram_resource_association.test]

  name   = %[1]q
  vpc_id = aws_vpc.test.id
}

resource "aws_db_instance" "test" {
  allocated_storage      = 5
  db_subnet_group_name   = aws_db_subnet_group.test.name
  engine                 = data.aws_rds_orderable_db_instance.test.engine
  identifier             = %[1]q
  instance_class         = data.aws_rds_orderable_db_instance.test.instance_class
  password               = "avoid-plaintext-passwords"
  username               = "tfacctest"
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]
}
`, rName))
}

func testAccAWSDBInstanceConfig_DbSubnetGroupName_VpcSecurityGroupIds(rName string) string {
	return composeConfig(
		testAccAWSDBInstanceConfig_orderableClassMysql(),
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_db_instance" "test" {
  allocated_storage      = 5
  db_subnet_group_name   = aws_db_subnet_group.test.name
  engine                 = data.aws_rds_orderable_db_instance.test.engine
  identifier             = %[1]q
  instance_class         = data.aws_rds_orderable_db_instance.test.instance_class
  password               = "avoid-plaintext-passwords"
  username               = "tfacctest"
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]
}
`, rName))
}

func testAccAWSDBInstanceConfig_DeletionProtection(rName string, deletionProtection bool) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMysql(), fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage   = 5
  deletion_protection = %t
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = %q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}
`, deletionProtection, rName))
}

func testAccAWSDBInstanceConfig_EnabledCloudwatchLogsExports_Oracle(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine         = "oracle-se"
  engine_version = "11.2.0.4.v25"
  license_model  = "bring-your-own-license"
  storage_type   = "standard"

  preferred_instance_classes = ["db.m5.large", "db.m4.large", "db.r4.large"]
}

resource "aws_db_instance" "test" {
  allocated_storage               = 10
  enabled_cloudwatch_logs_exports = ["alert", "listener", "trace"]
  engine                          = data.aws_rds_orderable_db_instance.test.engine
  identifier                      = %q
  instance_class                  = data.aws_rds_orderable_db_instance.test.instance_class
  password                        = "avoid-plaintext-passwords"
  username                        = "tfacctest"
  skip_final_snapshot             = true
}
`, rName)
}

func testAccAWSDBInstanceConfig_EnabledCloudwatchLogsExports_MSSQL(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine         = "sqlserver-se"
  engine_version = "14.00.1000.169.v1"
  license_model  = "license-included"
  storage_type   = "standard"

  preferred_instance_classes = ["db.m5.large", "db.m4.large", "db.r4.large"]
}

resource "aws_db_instance" "test" {
  allocated_storage               = 20
  enabled_cloudwatch_logs_exports = ["agent", "error"]
  engine                          = data.aws_rds_orderable_db_instance.test.engine
  identifier                      = %q
  instance_class                  = data.aws_rds_orderable_db_instance.test.instance_class
  license_model                   = data.aws_rds_orderable_db_instance.test.license_model
  password                        = "avoid-plaintext-passwords"
  username                        = "tfacctest"
  skip_final_snapshot             = true
}
`, rName)
}

func testAccAWSDBInstanceConfig_EnabledCloudwatchLogsExports_Postgresql(rName string) string {
	return composeConfig(
		testAccAWSDBInstanceConfig_orderableClass("postgres", "12.2", "postgresql-license"),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage               = 10
  enabled_cloudwatch_logs_exports = ["postgresql", "upgrade"]
  engine                          = data.aws_rds_orderable_db_instance.test.engine
  identifier                      = %q
  instance_class                  = data.aws_rds_orderable_db_instance.test.instance_class
  password                        = "avoid-plaintext-passwords"
  username                        = "tfacctest"
  skip_final_snapshot             = true
}
`, rName))
}

func testAccAWSDBInstanceConfig_MaxAllocatedStorage(rName string, maxAllocatedStorage int) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMysql(), fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage     = 5
  engine                = data.aws_rds_orderable_db_instance.test.engine
  identifier            = %q
  instance_class        = data.aws_rds_orderable_db_instance.test.instance_class
  max_allocated_storage = %d
  password              = "avoid-plaintext-passwords"
  username              = "tfacctest"
  skip_final_snapshot   = true
}
`, rName, maxAllocatedStorage))
}

func testAccAWSDBInstanceConfig_Password(rName, password string) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMysql(), fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = %q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = %q
  username            = "tfacctest"
  skip_final_snapshot = true
}
`, rName, password))
}

func testAccAWSDBInstanceConfig_ReplicateSourceDb(rName string) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMysql(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  replicate_source_db = aws_db_instance.source.id
  skip_final_snapshot = true
}
`, rName))
}

func testAccAWSDBInstanceConfig_ReplicateSourceDb_AllocatedStorage(rName string, allocatedStorage int) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMysql(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  allocated_storage   = %[2]d
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  replicate_source_db = aws_db_instance.source.id
  skip_final_snapshot = true
}
`, rName, allocatedStorage))
}

func testAccAWSDBInstanceConfig_ReplicateSourceDb_AllowMajorVersionUpgrade(rName string, allowMajorVersionUpgrade bool) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMysql(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  allow_major_version_upgrade = %[2]t
  identifier                  = %[1]q
  instance_class              = aws_db_instance.source.instance_class
  replicate_source_db         = aws_db_instance.source.id
  skip_final_snapshot         = true
}
`, rName, allowMajorVersionUpgrade))
}

func testAccAWSDBInstanceConfig_ReplicateSourceDb_AutoMinorVersionUpgrade(rName string, autoMinorVersionUpgrade bool) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMysql(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  auto_minor_version_upgrade = %[2]t
  identifier                 = %[1]q
  instance_class             = aws_db_instance.source.instance_class
  replicate_source_db        = aws_db_instance.source.id
  skip_final_snapshot        = true
}
`, rName, autoMinorVersionUpgrade))
}

func testAccAWSDBInstanceConfig_ReplicateSourceDb_AvailabilityZone(rName string) string {
	return composeConfig(
		testAccAWSDBInstanceConfig_orderableClassMysql(),
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  availability_zone   = data.aws_availability_zones.available.names[0]
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  replicate_source_db = aws_db_instance.source.id
  skip_final_snapshot = true
}
`, rName))
}

func testAccAWSDBInstanceConfig_ReplicateSourceDb_BackupRetentionPeriod(rName string, backupRetentionPeriod int) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMysql(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  backup_retention_period = %[2]d
  identifier              = %[1]q
  instance_class          = aws_db_instance.source.instance_class
  replicate_source_db     = aws_db_instance.source.id
  skip_final_snapshot     = true
}
`, rName, backupRetentionPeriod))
}

// We provide maintenance_window to prevent the following error from a randomly selected window:
// InvalidParameterValue: The backup window and maintenance window must not overlap.
func testAccAWSDBInstanceConfig_ReplicateSourceDb_BackupWindow(rName, backupWindow, maintenanceWindow string) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMysql(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  backup_window       = %[2]q
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  maintenance_window  = %[3]q
  replicate_source_db = aws_db_instance.source.id
  skip_final_snapshot = true
}
`, rName, backupWindow, maintenanceWindow))
}

func testAccAWSDBInstanceConfig_ReplicateSourceDb_DbSubnetGroupName(rName string) string {
	return composeConfig(
		testAccAlternateRegionProviderConfig(),
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
data "aws_availability_zones" "alternate" {
  provider = "awsalternate"

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "alternate" {
  provider = "awsalternate"

  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "alternate" {
  count    = 2
  provider = "awsalternate"

  availability_zone = data.aws_availability_zones.alternate.names[count.index]
  cidr_block        = "10.1.${count.index}.0/24"
  vpc_id            = aws_vpc.alternate.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_subnet_group" "alternate" {
  provider = "awsalternate"

  name       = %[1]q
  subnet_ids = aws_subnet.alternate[*].id
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

data "aws_rds_orderable_db_instance" "test" {
  provider = "awsalternate"

  engine         = "mysql"
  engine_version = "5.6.35"
  license_model  = "general-public-license"
  storage_type   = "standard"

  preferred_instance_classes = ["db.t3.micro", "db.t2.micro", "db.t2.medium"]
}

resource "aws_db_instance" "source" {
  provider = "awsalternate"

  allocated_storage       = 5
  backup_retention_period = 1
  db_subnet_group_name    = aws_db_subnet_group.alternate.name
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  db_subnet_group_name = aws_db_subnet_group.test.name
  identifier           = %[1]q
  instance_class       = aws_db_instance.source.instance_class
  replicate_source_db  = aws_db_instance.source.arn
  skip_final_snapshot  = true
}
`, rName))
}

func testAccAWSDBInstanceConfig_ReplicateSourceDb_DbSubnetGroupName_RamShared(rName string) string {
	return composeConfig(testAccAlternateAccountAndAlternateRegionProviderConfig() + fmt.Sprintf(`
data "aws_availability_zones" "alternateaccountsameregion" {
  provider = "awsalternateaccountsameregion"

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_availability_zones" "sameaccountalternateregion" {
  provider = "awssameaccountalternateregion"

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_organizations_organization" "test" {}

resource "aws_vpc" "sameaccountalternateregion" {
  provider = "awssameaccountalternateregion"

  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "alternateaccountsameregion" {
  provider = "awsalternateaccountsameregion"

  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "sameaccountalternateregion" {
  count    = 2
  provider = "awssameaccountalternateregion"

  availability_zone = data.aws_availability_zones.sameaccountalternateregion.names[count.index]
  cidr_block        = "10.1.${count.index}.0/24"
  vpc_id            = aws_vpc.sameaccountalternateregion.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "alternateaccountsameregion" {
  count    = 2
  provider = "awsalternateaccountsameregion"

  availability_zone = data.aws_availability_zones.alternateaccountsameregion.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.alternateaccountsameregion.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ram_resource_share" "alternateaccountsameregion" {
  provider = "awsalternateaccountsameregion"

  name = %[1]q
}

resource "aws_ram_principal_association" "alternateaccountsameregion" {
  provider = "awsalternateaccountsameregion"

  principal          = data.aws_organizations_organization.test.arn
  resource_share_arn = aws_ram_resource_share.alternateaccountsameregion.arn
}

resource "aws_ram_resource_association" "alternateaccountsameregion" {
  count    = 2
  provider = "awsalternateaccountsameregion"

  resource_arn       = aws_subnet.alternateaccountsameregion[count.index].arn
  resource_share_arn = aws_ram_resource_share.alternateaccountsameregion.id
}

resource "aws_db_subnet_group" "sameaccountalternateregion" {
  provider = "awssameaccountalternateregion"

  name       = %[1]q
  subnet_ids = aws_subnet.sameaccountalternateregion[*].id
}

resource "aws_db_subnet_group" "test" {
  depends_on = [aws_ram_principal_association.alternateaccountsameregion, aws_ram_resource_association.alternateaccountsameregion]

  name       = %[1]q
  subnet_ids = aws_subnet.alternateaccountsameregion[*].id
}

resource "aws_security_group" "test" {
  depends_on = [aws_ram_principal_association.alternateaccountsameregion, aws_ram_resource_association.alternateaccountsameregion]

  name   = %[1]q
  vpc_id = aws_vpc.alternateaccountsameregion.id
}

data "aws_rds_orderable_db_instance" "test" {
  provider = "awssameaccountalternateregion"

  engine         = "mysql"
  engine_version = "5.6.35"
  license_model  = "general-public-license"
  storage_type   = "standard"

  preferred_instance_classes = ["db.t3.micro", "db.t2.micro", "db.t2.medium"]
}

resource "aws_db_instance" "source" {
  provider = "awssameaccountalternateregion"

  allocated_storage       = 5
  backup_retention_period = 1
  db_subnet_group_name    = aws_db_subnet_group.sameaccountalternateregion.name
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  db_subnet_group_name   = aws_db_subnet_group.test.name
  identifier             = %[1]q
  instance_class         = aws_db_instance.source.instance_class
  replicate_source_db    = aws_db_instance.source.arn
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]
}
`, rName))
}

func testAccAWSDBInstanceConfig_ReplicateSourceDb_DbSubnetGroupName_VpcSecurityGroupIds(rName string) string {
	return composeConfig(
		testAccAlternateRegionProviderConfig(),
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
data "aws_availability_zones" "alternate" {
  provider = "awsalternate"

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "alternate" {
  provider = "awsalternate"

  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "alternate" {
  count    = 2
  provider = "awsalternate"

  availability_zone = data.aws_availability_zones.alternate.names[count.index]
  cidr_block        = "10.1.${count.index}.0/24"
  vpc_id            = aws_vpc.alternate.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_subnet_group" "alternate" {
  provider = "awsalternate"

  name       = %[1]q
  subnet_ids = aws_subnet.alternate[*].id
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

data "aws_rds_orderable_db_instance" "test" {
  provider = "awsalternate"

  engine         = "mysql"
  engine_version = "5.6.35"
  license_model  = "general-public-license"
  storage_type   = "standard"

  preferred_instance_classes = ["db.t3.micro", "db.t2.micro", "db.t2.medium"]
}

resource "aws_db_instance" "source" {
  provider = "awsalternate"

  allocated_storage       = 5
  backup_retention_period = 1
  db_subnet_group_name    = aws_db_subnet_group.alternate.name
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  db_subnet_group_name   = aws_db_subnet_group.test.name
  identifier             = %[1]q
  instance_class         = aws_db_instance.source.instance_class
  replicate_source_db    = aws_db_instance.source.arn
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]
}
`, rName))
}

func testAccAWSDBInstanceConfig_ReplicateSourceDb_DeletionProtection(rName string, deletionProtection bool) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMysql(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  deletion_protection = %[2]t
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  replicate_source_db = aws_db_instance.source.id
  skip_final_snapshot = true
}
`, rName, deletionProtection))
}

func testAccAWSDBInstanceConfig_ReplicateSourceDb_IamDatabaseAuthenticationEnabled(rName string, iamDatabaseAuthenticationEnabled bool) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMysql(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  iam_database_authentication_enabled = %[2]t
  identifier                          = %[1]q
  instance_class                      = aws_db_instance.source.instance_class
  replicate_source_db                 = aws_db_instance.source.id
  skip_final_snapshot                 = true
}
`, rName, iamDatabaseAuthenticationEnabled))
}

// We provide backup_window to prevent the following error from a randomly selected window:
// InvalidParameterValue: The backup window and maintenance window must not overlap.
func testAccAWSDBInstanceConfig_ReplicateSourceDb_MaintenanceWindow(rName, backupWindow, maintenanceWindow string) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMysql(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  backup_window       = %[2]q
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  maintenance_window  = %[3]q
  replicate_source_db = aws_db_instance.source.id
  skip_final_snapshot = true
}
`, rName, backupWindow, maintenanceWindow))
}

func testAccAWSDBInstanceConfig_ReplicateSourceDb_MaxAllocatedStorage(rName string, maxAllocatedStorage int) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMysql(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  allocated_storage     = aws_db_instance.source.allocated_storage
  identifier            = %[1]q
  instance_class        = aws_db_instance.source.instance_class
  max_allocated_storage = %[2]d
  replicate_source_db   = aws_db_instance.source.id
  skip_final_snapshot   = true
}
`, rName, maxAllocatedStorage))
}

func testAccAWSDBInstanceConfig_ReplicateSourceDb_Monitoring(rName string, monitoringInterval int) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMysql(), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "monitoring.rds.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonRDSEnhancedMonitoringRole"
  role       = aws_iam_role.test.id
}

resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  monitoring_interval = %[2]d
  monitoring_role_arn = aws_iam_role.test.arn
  replicate_source_db = aws_db_instance.source.id
  skip_final_snapshot = true
}
`, rName, monitoringInterval))
}

func testAccAWSDBInstanceConfig_ReplicateSourceDb_MultiAZ(rName string, multiAz bool) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMysql(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  multi_az            = %[2]t
  replicate_source_db = aws_db_instance.source.id
  skip_final_snapshot = true
}
`, rName, multiAz))
}

func testAccAWSDBInstanceConfig_ReplicateSourceDb_ParameterGroupName(rName string) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMysql(), fmt.Sprintf(`
resource "aws_db_parameter_group" "test" {
  family = "mysql5.6"
  name   = %[1]q

  parameter {
    name  = "sync_binlog"
    value = 0
  }
}

resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier           = %[1]q
  instance_class       = aws_db_instance.source.instance_class
  parameter_group_name = aws_db_parameter_group.test.id
  replicate_source_db  = aws_db_instance.source.id
  skip_final_snapshot  = true
}
`, rName))
}

func testAccAWSDBInstanceConfig_ReplicateSourceDb_Port(rName string, port int) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMysql(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  port                = %[2]d
  replicate_source_db = aws_db_instance.source.id
  skip_final_snapshot = true
}
`, rName, port))
}

func testAccAWSDBInstanceConfig_ReplicateSourceDb_VpcSecurityGroupIds(rName string) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMysql(), fmt.Sprintf(`
data "aws_vpc" "default" {
  default = true
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = data.aws_vpc.default.id
}

resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier             = %[1]q
  instance_class         = aws_db_instance.source.instance_class
  replicate_source_db    = aws_db_instance.source.id
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]
}
`, rName))
}

func testAccAWSDBInstanceConfig_ReplicateSourceDb_CACertificateIdentifier(rName string) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMysql(), fmt.Sprintf(`
data "aws_rds_certificate" "latest" {
  latest_valid_till = true
}

resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  ca_cert_identifier      = data.aws_rds_certificate.latest.id
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  replicate_source_db = aws_db_instance.source.id
  ca_cert_identifier  = data.aws_rds_certificate.latest.id
  skip_final_snapshot = true
}
`, rName))
}

func testAccAWSDBInstanceConfig_SnapshotIdentifier(rName string) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
}
`, rName))
}

func testAccAWSDBInstanceConfig_SnapshotIdentifier_AllocatedStorage(rName string, allocatedStorage int) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  allocated_storage   = %[2]d
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
}
`, rName, allocatedStorage))
}

func testAccAWSDBInstanceConfig_SnapshotIdentifier_Io1Storage(rName string, iops int) string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine         = "mariadb"
  engine_version = "10.2.15"
  license_model  = "general-public-license"
  storage_type   = "io1"

  preferred_instance_classes = ["db.t3.micro", "db.t2.micro", "db.t2.medium"]
}

resource "aws_db_instance" "source" {
  allocated_storage   = 200
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
  allocated_storage   = 200
  iops                = %[2]d
  storage_type        = data.aws_rds_orderable_db_instance.test.storage_type
}
`, rName, iops)
}

func testAccAWSDBInstanceConfig_SnapshotIdentifier_AllowMajorVersionUpgrade(rName string, allowMajorVersionUpgrade bool) string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "postgres10" {
  engine         = "postgres"
  engine_version = "10.1"
  license_model  = "postgresql-license"
  storage_type   = "standard"

  preferred_instance_classes = ["db.t3.micro", "db.t2.micro", "db.t2.medium"]
}

resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.postgres10.engine
  engine_version      = data.aws_rds_orderable_db_instance.postgres10.engine_version
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.postgres10.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

data "aws_rds_orderable_db_instance" "postgres11" {
  engine         = "postgres"
  engine_version = "11.1"
  license_model  = "postgresql-license"
  storage_type   = "standard"

  preferred_instance_classes = ["db.t3.micro", "db.t2.micro", "db.t2.medium"]
}

resource "aws_db_instance" "test" {
  allow_major_version_upgrade = %[2]t
  engine                      = data.aws_rds_orderable_db_instance.postgres11.engine
  engine_version              = data.aws_rds_orderable_db_instance.postgres11.engine_version
  identifier                  = %[1]q
  instance_class              = aws_db_instance.source.instance_class
  snapshot_identifier         = aws_db_snapshot.test.id
  skip_final_snapshot         = true
}
`, rName, allowMajorVersionUpgrade)
}

func testAccAWSDBInstanceConfig_SnapshotIdentifier_AutoMinorVersionUpgrade(rName string, autoMinorVersionUpgrade bool) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  auto_minor_version_upgrade = %[2]t
  identifier                 = %[1]q
  instance_class             = aws_db_instance.source.instance_class
  snapshot_identifier        = aws_db_snapshot.test.id
  skip_final_snapshot        = true
}
`, rName, autoMinorVersionUpgrade))
}

func testAccAWSDBInstanceConfig_SnapshotIdentifier_AvailabilityZone(rName string) string {
	return composeConfig(
		testAccAWSDBInstanceConfig_orderableClassMariadb(),
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  availability_zone   = data.aws_availability_zones.available.names[0]
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
}
`, rName))
}

func testAccAWSDBInstanceConfig_SnapshotIdentifier_BackupRetentionPeriod(rName string, backupRetentionPeriod int) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  backup_retention_period = %[2]d
  identifier              = %[1]q
  instance_class          = aws_db_instance.source.instance_class
  snapshot_identifier     = aws_db_snapshot.test.id
  skip_final_snapshot     = true
}
`, rName, backupRetentionPeriod))
}

func testAccAWSDBInstanceConfig_SnapshotIdentifier_BackupRetentionPeriod_Unset(rName string) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  backup_retention_period = 0
  identifier              = %[1]q
  instance_class          = aws_db_instance.source.instance_class
  snapshot_identifier     = aws_db_snapshot.test.id
  skip_final_snapshot     = true
}
`, rName))
}

// We provide maintenance_window to prevent the following error from a randomly selected window:
// InvalidParameterValue: The backup window and maintenance window must not overlap.
func testAccAWSDBInstanceConfig_SnapshotIdentifier_BackupWindow(rName, backupWindow, maintenanceWindow string) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  backup_window       = %[2]q
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  maintenance_window  = %[3]q
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
}
`, rName, backupWindow, maintenanceWindow))
}

func testAccAWSDBInstanceConfig_SnapshotIdentifier_DbSubnetGroupName(rName string) string {
	return composeConfig(
		testAccAWSDBInstanceConfig_orderableClassMariadb(),
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  db_subnet_group_name = aws_db_subnet_group.test.name
  identifier           = %[1]q
  instance_class       = aws_db_instance.source.instance_class
  snapshot_identifier  = aws_db_snapshot.test.id
  skip_final_snapshot  = true
}
`, rName))
}

func testAccAWSDBInstanceConfig_SnapshotIdentifier_DbSubnetGroupName_RamShared(rName string) string {
	return composeConfig(
		testAccAWSDBInstanceConfig_orderableClassMariadb(),
		testAccAlternateAccountProviderConfig(),
		fmt.Sprintf(`
data "aws_availability_zones" "alternate" {
  provider = "awsalternate"

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_organizations_organization" "test" {}

resource "aws_vpc" "test" {
  provider = "awsalternate"

  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count    = 2
  provider = "awsalternate"

  availability_zone = data.aws_availability_zones.alternate.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ram_resource_share" "test" {
  provider = "awsalternate"

  name = %[1]q
}

resource "aws_ram_principal_association" "test" {
  provider = "awsalternate"

  principal          = data.aws_organizations_organization.test.arn
  resource_share_arn = aws_ram_resource_share.test.arn
}

resource "aws_ram_resource_association" "test" {
  count    = 2
  provider = "awsalternate"

  resource_arn       = aws_subnet.test[count.index].arn
  resource_share_arn = aws_ram_resource_share.test.id
}

resource "aws_db_subnet_group" "test" {
  depends_on = [aws_ram_principal_association.test, aws_ram_resource_association.test]

  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_security_group" "test" {
  depends_on = [aws_ram_principal_association.test, aws_ram_resource_association.test]

  name   = %[1]q
  vpc_id = aws_vpc.test.id
}

resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  db_subnet_group_name   = aws_db_subnet_group.test.name
  identifier             = %[1]q
  instance_class         = aws_db_instance.source.instance_class
  snapshot_identifier    = aws_db_snapshot.test.id
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]
}
`, rName))
}

func testAccAWSDBInstanceConfig_SnapshotIdentifier_DbSubnetGroupName_VpcSecurityGroupIds(rName string) string {
	return composeConfig(
		testAccAWSDBInstanceConfig_orderableClassMariadb(),
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  db_subnet_group_name   = aws_db_subnet_group.test.name
  identifier             = %[1]q
  instance_class         = aws_db_instance.source.instance_class
  snapshot_identifier    = aws_db_snapshot.test.id
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]
}
`, rName))
}

func testAccAWSDBInstanceConfig_SnapshotIdentifier_DeletionProtection(rName string, deletionProtection bool) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMysql(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  deletion_protection = %[2]t
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
}
`, rName, deletionProtection))
}

func testAccAWSDBInstanceConfig_SnapshotIdentifier_IamDatabaseAuthenticationEnabled(rName string, iamDatabaseAuthenticationEnabled bool) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMysql(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  iam_database_authentication_enabled = %[2]t
  identifier                          = %[1]q
  instance_class                      = aws_db_instance.source.instance_class
  snapshot_identifier                 = aws_db_snapshot.test.id
  skip_final_snapshot                 = true
}
`, rName, iamDatabaseAuthenticationEnabled))
}

// We provide backup_window to prevent the following error from a randomly selected window:
// InvalidParameterValue: The backup window and maintenance window must not overlap.
func testAccAWSDBInstanceConfig_SnapshotIdentifier_MaintenanceWindow(rName, backupWindow, maintenanceWindow string) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  backup_window       = %[2]q
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  maintenance_window  = %[3]q
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
}
`, rName, backupWindow, maintenanceWindow))
}

func testAccAWSDBInstanceConfig_SnapshotIdentifier_MaxAllocatedStorage(rName string, maxAllocatedStorage int) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  allocated_storage     = aws_db_instance.source.allocated_storage
  identifier            = %[1]q
  instance_class        = aws_db_instance.source.instance_class
  max_allocated_storage = %[2]d
  snapshot_identifier   = aws_db_snapshot.test.id
  skip_final_snapshot   = true
}
`, rName, maxAllocatedStorage))
}

func testAccAWSDBInstanceConfig_SnapshotIdentifier_Monitoring(rName string, monitoringInterval int) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "monitoring.rds.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonRDSEnhancedMonitoringRole"
  role       = aws_iam_role.test.id
}

resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  monitoring_interval = %[2]d
  monitoring_role_arn = aws_iam_role.test.arn
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
}
`, rName, monitoringInterval))
}

func testAccAWSDBInstanceConfig_SnapshotIdentifier_MultiAZ(rName string, multiAz bool) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  multi_az            = %[2]t
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
}
`, rName, multiAz))
}

func testAccAWSDBInstanceConfig_SnapshotIdentifier_MultiAZ_SQLServer(rName string, multiAz bool) string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine         = "sqlserver-se"
  engine_version = "14.00.1000.169.v1"
  license_model  = "license-included"
  storage_type   = "standard"

  preferred_instance_classes = ["db.m5.large", "db.m4.large", "db.r4.large"]
}

resource "aws_db_instance" "source" {
  allocated_storage   = 20
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  license_model       = data.aws_rds_orderable_db_instance.test.license_model
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  # InvalidParameterValue: Mirroring cannot be applied to instances with backup retention set to zero.
  backup_retention_period = 1
  identifier              = %[1]q
  instance_class          = aws_db_instance.source.instance_class
  multi_az                = %[2]t
  snapshot_identifier     = aws_db_snapshot.test.id
  skip_final_snapshot     = true
}
`, rName, multiAz)
}

func testAccAWSDBInstanceConfig_SnapshotIdentifier_ParameterGroupName(rName string) string {
	return composeConfig(
		testAccAWSDBInstanceConfig_orderableClassMariadb(),
		fmt.Sprintf(`
resource "aws_db_parameter_group" "test" {
  family = "mariadb10.2"
  name   = %[1]q

  parameter {
    name  = "sync_binlog"
    value = 0
  }
}

resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  engine_version      = data.aws_rds_orderable_db_instance.test.engine_version
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier           = %[1]q
  instance_class       = aws_db_instance.source.instance_class
  parameter_group_name = aws_db_parameter_group.test.id
  snapshot_identifier  = aws_db_snapshot.test.id
  skip_final_snapshot  = true
}
`, rName))
}

func testAccAWSDBInstanceConfig_SnapshotIdentifier_Port(rName string, port int) string {
	return composeConfig(
		testAccAWSDBInstanceConfig_orderableClassMariadb(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  port                = %[2]d
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true
}
`, rName, port))
}

func testAccAWSDBInstanceConfig_SnapshotIdentifier_Tags(rName string) string {
	return composeConfig(
		testAccAWSDBInstanceConfig_orderableClassMariadb(),
		fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true

  tags = {
    key1 = "value1"
  }
}
`, rName))
}

func testAccAWSDBInstanceConfig_SnapshotIdentifier_Tags_Unset(rName string) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true

  tags = {
    key1 = "value1"
  }
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  instance_class      = aws_db_instance.source.instance_class
  snapshot_identifier = aws_db_snapshot.test.id
  skip_final_snapshot = true

  tags = {}
}
`, rName))
}

func testAccAWSDBInstanceConfig_SnapshotIdentifier_VpcSecurityGroupIds(rName string) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
data "aws_vpc" "default" {
  default = true
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = data.aws_vpc.default.id
}

resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier             = %[1]q
  instance_class         = aws_db_instance.source.instance_class
  snapshot_identifier    = aws_db_snapshot.test.id
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]
}
`, rName))
}

func testAccAWSDBInstanceConfig_SnapshotIdentifier_VpcSecurityGroupIds_Tags(rName string) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
data "aws_vpc" "default" {
  default = true
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = data.aws_vpc.default.id
}

resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier             = %[1]q
  instance_class         = aws_db_instance.source.instance_class
  snapshot_identifier    = aws_db_snapshot.test.id
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]

  tags = {
    key1 = "value1"
  }
}
`, rName))
}

func testAccAWSDBInstanceConfig_PerformanceInsightsDisabled(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = data.aws_rds_engine_version.default.engine
  engine_version                = data.aws_rds_engine_version.default.version
  license_model                 = "general-public-license"
  storage_type                  = "standard"
  supports_performance_insights = true
  preferred_instance_classes    = ["db.m3.medium", "db.m3.large", "db.m4.large"]
}

resource "aws_db_instance" "test" {
  allocated_storage       = 5
  backup_retention_period = 0
  engine                  = data.aws_rds_engine_version.default.engine
  engine_version          = data.aws_rds_engine_version.default.version
  identifier              = %q
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  name                    = "mydb"
  password                = "mustbeeightcharaters"
  skip_final_snapshot     = true
  username                = "foo"
}
`, rName)
}

func testAccAWSDBInstanceConfig_PerformanceInsightsEnabled(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = data.aws_rds_engine_version.default.engine
  engine_version                = data.aws_rds_engine_version.default.version
  license_model                 = "general-public-license"
  storage_type                  = "standard"
  supports_performance_insights = true
  preferred_instance_classes    = ["db.m3.medium", "db.m3.large", "db.m4.large"]
}

resource "aws_db_instance" "test" {
  allocated_storage                     = 5
  backup_retention_period               = 0
  engine                                = data.aws_rds_engine_version.default.engine
  engine_version                        = data.aws_rds_engine_version.default.version
  identifier                            = %q
  instance_class                        = data.aws_rds_orderable_db_instance.test.instance_class
  name                                  = "mydb"
  password                              = "mustbeeightcharaters"
  performance_insights_enabled          = true
  performance_insights_retention_period = 7
  skip_final_snapshot                   = true
  username                              = "foo"
}
`, rName)
}

func testAccAWSDBInstanceConfig_PerformanceInsightsKmsKeyIdDisabled(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = data.aws_rds_engine_version.default.engine
  engine_version                = data.aws_rds_engine_version.default.version
  license_model                 = "general-public-license"
  storage_type                  = "standard"
  supports_performance_insights = true
  preferred_instance_classes    = ["db.m3.medium", "db.m3.large", "db.m4.large"]
}

resource "aws_db_instance" "test" {
  engine                  = data.aws_rds_engine_version.default.engine
  identifier              = %q
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  allocated_storage       = 5
  backup_retention_period = 0
  name                    = "mydb"
  username                = "foo"
  password                = "mustbeeightcharaters"
  skip_final_snapshot     = true
}
`, rName)
}

func testAccAWSDBInstanceConfig_PerformanceInsightsKmsKeyId(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = data.aws_rds_engine_version.default.engine
  engine_version                = data.aws_rds_engine_version.default.version
  license_model                 = "general-public-license"
  storage_type                  = "standard"
  supports_performance_insights = true
  preferred_instance_classes    = ["db.m3.medium", "db.m3.large", "db.m4.large"]
}

resource "aws_db_instance" "test" {
  allocated_storage                     = 5
  backup_retention_period               = 0
  engine                                = data.aws_rds_engine_version.default.engine
  engine_version                        = data.aws_rds_engine_version.default.version
  identifier                            = %q
  instance_class                        = data.aws_rds_orderable_db_instance.test.instance_class
  name                                  = "mydb"
  password                              = "mustbeeightcharaters"
  performance_insights_enabled          = true
  performance_insights_kms_key_id       = aws_kms_key.test.arn
  performance_insights_retention_period = 7
  skip_final_snapshot                   = true
  username                              = "foo"
}
`, rName)
}

func testAccAWSDBInstanceConfig_PerformanceInsightsRetentionPeriod(rName string, performanceInsightsRetentionPeriod int) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = data.aws_rds_engine_version.default.engine
  engine_version                = data.aws_rds_engine_version.default.version
  license_model                 = "general-public-license"
  storage_type                  = "standard"
  supports_performance_insights = true
  preferred_instance_classes    = ["db.m3.medium", "db.m3.large", "db.m4.large"]
}

resource "aws_db_instance" "test" {
  allocated_storage                     = 5
  backup_retention_period               = 0
  engine                                = data.aws_rds_engine_version.default.engine
  engine_version                        = data.aws_rds_engine_version.default.version
  identifier                            = %q
  instance_class                        = data.aws_rds_orderable_db_instance.test.instance_class
  name                                  = "mydb"
  password                              = "mustbeeightcharaters"
  performance_insights_enabled          = true
  performance_insights_retention_period = %d
  skip_final_snapshot                   = true
  username                              = "foo"
}
`, rName, performanceInsightsRetentionPeriod)
}

func testAccAWSDBInstanceConfig_ReplicateSourceDb_PerformanceInsightsEnabled(rName string) string {
	return fmt.Sprintf(`
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

data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = data.aws_rds_engine_version.default.engine
  engine_version                = data.aws_rds_engine_version.default.version
  license_model                 = "general-public-license"
  storage_type                  = "standard"
  supports_performance_insights = true
  preferred_instance_classes    = ["db.m3.medium", "db.m3.large", "db.m4.large"]
}

resource "aws_db_instance" "source" {
  allocated_storage       = 5
  backup_retention_period = 1
  engine                  = data.aws_rds_engine_version.default.engine
  engine_version          = data.aws_rds_engine_version.default.version
  identifier              = "%[1]s-source"
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  password                = "mustbeeightcharaters"
  username                = "tfacctest"
  skip_final_snapshot     = true
}

resource "aws_db_instance" "test" {
  identifier                            = %[1]q
  instance_class                        = aws_db_instance.source.instance_class
  performance_insights_enabled          = true
  performance_insights_kms_key_id       = aws_kms_key.test.arn
  performance_insights_retention_period = 7
  replicate_source_db                   = aws_db_instance.source.id
  skip_final_snapshot                   = true
}
`, rName)
}

func testAccAWSDBInstanceConfig_SnapshotIdentifier_PerformanceInsightsEnabled(rName string) string {
	return fmt.Sprintf(`
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

data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = data.aws_rds_engine_version.default.engine
  engine_version                = data.aws_rds_engine_version.default.version
  license_model                 = "general-public-license"
  storage_type                  = "standard"
  supports_performance_insights = true
  preferred_instance_classes    = ["db.m3.medium", "db.m3.large", "db.m4.large"]
}

resource "aws_db_instance" "source" {
  allocated_storage   = 5
  engine              = data.aws_rds_engine_version.default.engine
  engine_version      = data.aws_rds_engine_version.default.version
  identifier          = "%[1]s-source"
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.source.id
  db_snapshot_identifier = %[1]q
}

resource "aws_db_instance" "test" {
  identifier                            = %[1]q
  instance_class                        = aws_db_instance.source.instance_class
  performance_insights_enabled          = true
  performance_insights_kms_key_id       = aws_kms_key.test.arn
  performance_insights_retention_period = 7
  snapshot_identifier                   = aws_db_snapshot.test.id
  skip_final_snapshot                   = true
}
`, rName)
}

func testAccAWSDBInstanceConfig_NoDeleteAutomatedBackups(rName string) string {
	return composeConfig(testAccAWSDBInstanceConfig_orderableClassMariadb(), fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage   = 10
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = %q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true

  backup_retention_period  = 1
  delete_automated_backups = false
}
`, rName))
}
