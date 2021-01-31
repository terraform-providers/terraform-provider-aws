package aws

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSRDSClusterInstance_basic(t *testing.T) {
	var v rds.DBInstance
	resourceName := "aws_rds_cluster_instance.cluster_instances"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfig(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &v),
					testAccCheckAWSDBClusterInstanceAttributes(&v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "rds", regexp.MustCompile(`db:.+`)),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "true"),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "preferred_maintenance_window"),
					resource.TestCheckResourceAttrSet(resourceName, "preferred_backup_window"),
					resource.TestCheckResourceAttrSet(resourceName, "dbi_resource_id"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_version"),
					resource.TestCheckResourceAttr(resourceName, "engine", "aurora"),
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
			{
				Config: testAccAWSClusterInstanceConfigModified(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &v),
					testAccCheckAWSDBClusterInstanceAttributes(&v),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
				),
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_isAlreadyBeingDeleted(t *testing.T) {
	var v rds.DBInstance
	resourceName := "aws_rds_cluster_instance.cluster_instances"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &v),
				),
			},
			{
				PreConfig: func() {
					// Get Database Instance into deleting state
					conn := testAccProvider.Meta().(*AWSClient).rdsconn
					input := &rds.DeleteDBInstanceInput{
						DBInstanceIdentifier: aws.String(fmt.Sprintf("tf-cluster-instance-%d", rInt)),
						SkipFinalSnapshot:    aws.Bool(true),
					}
					_, err := conn.DeleteDBInstance(input)
					if err != nil {
						t.Fatalf("error deleting Database Instance: %s", err)
					}
				},
				Config:  testAccAWSClusterInstanceConfig(rInt),
				Destroy: true,
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_az(t *testing.T) {
	var v rds.DBInstance
	resourceName := "aws_rds_cluster_instance.cluster_instances"
	availabilityZonesDataSourceName := "data.aws_availability_zones.available"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfig_az(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &v),
					testAccCheckAWSDBClusterInstanceAttributes(&v),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zone", availabilityZonesDataSourceName, "names.0"),
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

func TestAccAWSRDSClusterInstance_namePrefix(t *testing.T) {
	var v rds.DBInstance
	rInt := acctest.RandInt()
	resourceName := "aws_rds_cluster_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfig_namePrefix(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &v),
					testAccCheckAWSDBClusterInstanceAttributes(&v),
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

func TestAccAWSRDSClusterInstance_generatedName(t *testing.T) {
	var v rds.DBInstance
	resourceName := "aws_rds_cluster_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfig_generatedName(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &v),
					testAccCheckAWSDBClusterInstanceAttributes(&v),
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

func TestAccAWSRDSClusterInstance_kmsKey(t *testing.T) {
	var v rds.DBInstance
	kmsKeyResourceName := "aws_kms_key.foo"
	resourceName := "aws_rds_cluster_instance.cluster_instances"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfigKmsKey(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsKeyResourceName, "arn"),
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
func TestAccAWSRDSClusterInstance_disappears(t *testing.T) {
	var v rds.DBInstance
	resourceName := "aws_rds_cluster_instance.cluster_instances"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfig(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &v),
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
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
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
	resourceName := "aws_rds_cluster_instance.cluster_instances"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfig_CopyTagsToSnapshot(rNameSuffix, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", "true"),
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
			{
				Config: testAccAWSClusterInstanceConfig_CopyTagsToSnapshot(rNameSuffix, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", "false"),
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
				if ok && dbinstanceerr.Code() == rds.ErrCodeDBInstanceNotFoundFault {
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

func TestAccAWSRDSClusterInstance_MonitoringInterval(t *testing.T) {
	var dbInstance rds.DBInstance
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfigMonitoringInterval(rName, 30),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "30"),
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
			{
				Config: testAccAWSClusterInstanceConfigMonitoringInterval(rName, 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "60"),
				),
			},
			{
				Config: testAccAWSClusterInstanceConfigMonitoringInterval(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "0"),
				),
			},
			{
				Config: testAccAWSClusterInstanceConfigMonitoringInterval(rName, 30),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "30"),
				),
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_MonitoringRoleArn_EnabledToDisabled(t *testing.T) {
	var dbInstance rds.DBInstance
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfigMonitoringRoleArn(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "monitoring_role_arn", iamRoleResourceName, "arn"),
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
			{
				Config: testAccAWSClusterInstanceConfigMonitoringInterval(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "0"),
				),
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_MonitoringRoleArn_EnabledToRemoved(t *testing.T) {
	var dbInstance rds.DBInstance
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfigMonitoringRoleArn(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "monitoring_role_arn", iamRoleResourceName, "arn"),
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
			{
				Config: testAccAWSClusterInstanceConfigMonitoringRoleArnRemoved(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
				),
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_MonitoringRoleArn_RemovedToEnabled(t *testing.T) {
	var dbInstance rds.DBInstance
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfigMonitoringRoleArnRemoved(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
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
			{
				Config: testAccAWSClusterInstanceConfigMonitoringRoleArn(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "monitoring_role_arn", iamRoleResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_PerformanceInsightsEnabled_AuroraMysql1(t *testing.T) {
	var dbInstance rds.DBInstance
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	engine := "aurora"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccRDSPerformanceInsightsDefaultVersionPreCheck(t, engine) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfigPerformanceInsightsEnabledAuroraMysql1(rName, engine),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
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

func TestAccAWSRDSClusterInstance_PerformanceInsightsEnabled_AuroraMysql2(t *testing.T) {
	var dbInstance rds.DBInstance
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	engine := "aurora-mysql"
	engineVersion := "5.7.mysql_aurora.2.04.2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccRDSPerformanceInsightsPreCheck(t, engine, engineVersion) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfigPerformanceInsightsEnabledAuroraMysql2(rName, engine, engineVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
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

func TestAccAWSRDSClusterInstance_PerformanceInsightsEnabled_AuroraPostgresql(t *testing.T) {
	var dbInstance rds.DBInstance
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	engine := "aurora-postgresql"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccRDSPerformanceInsightsDefaultVersionPreCheck(t, engine) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfigPerformanceInsightsEnabledAuroraPostgresql(rName, engine),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
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

func TestAccAWSRDSClusterInstance_PerformanceInsightsKmsKeyId_AuroraMysql1(t *testing.T) {
	var dbInstance rds.DBInstance
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	engine := "aurora"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccRDSPerformanceInsightsDefaultVersionPreCheck(t, engine) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfigPerformanceInsightsKmsKeyIdAuroraMysql1(rName, engine),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "performance_insights_kms_key_id", kmsKeyResourceName, "arn"),
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

func TestAccAWSRDSClusterInstance_PerformanceInsightsKmsKeyId_AuroraMysql1_DefaultKeyToCustomKey(t *testing.T) {
	var dbInstance rds.DBInstance
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	engine := "aurora"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccRDSPerformanceInsightsDefaultVersionPreCheck(t, engine) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfigPerformanceInsightsEnabledAuroraMysql1(rName, engine),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
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
			{
				Config:      testAccAWSClusterInstanceConfigPerformanceInsightsKmsKeyIdAuroraMysql1(rName, engine),
				ExpectError: regexp.MustCompile(`InvalidParameterCombination: You cannot change your Performance Insights KMS key`),
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_PerformanceInsightsKmsKeyId_AuroraMysql2(t *testing.T) {
	var dbInstance rds.DBInstance
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	engine := "aurora-mysql"
	engineVersion := "5.7.mysql_aurora.2.04.2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccRDSPerformanceInsightsPreCheck(t, engine, engineVersion) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfigPerformanceInsightsKmsKeyIdAuroraMysql2(rName, engine, engineVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "performance_insights_kms_key_id", kmsKeyResourceName, "arn"),
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

func TestAccAWSRDSClusterInstance_PerformanceInsightsKmsKeyId_AuroraMysql2_DefaultKeyToCustomKey(t *testing.T) {
	var dbInstance rds.DBInstance
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	engine := "aurora-mysql"
	engineVersion := "5.7.mysql_aurora.2.04.2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccRDSPerformanceInsightsPreCheck(t, engine, engineVersion) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfigPerformanceInsightsEnabledAuroraMysql2(rName, engine, engineVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
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
			{
				Config:      testAccAWSClusterInstanceConfigPerformanceInsightsKmsKeyIdAuroraMysql2(rName, engine, engineVersion),
				ExpectError: regexp.MustCompile(`InvalidParameterCombination: You cannot change your Performance Insights KMS key`),
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_PerformanceInsightsKmsKeyId_AuroraPostgresql(t *testing.T) {
	var dbInstance rds.DBInstance
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	engine := "aurora-postgresql"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccRDSPerformanceInsightsDefaultVersionPreCheck(t, engine) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfigPerformanceInsightsKmsKeyIdAuroraPostgresql(rName, engine),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "performance_insights_kms_key_id", kmsKeyResourceName, "arn"),
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

func TestAccAWSRDSClusterInstance_PerformanceInsightsKmsKeyId_AuroraPostgresql_DefaultKeyToCustomKey(t *testing.T) {
	var dbInstance rds.DBInstance
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	engine := "aurora-postgresql"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccRDSPerformanceInsightsDefaultVersionPreCheck(t, engine) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfigPerformanceInsightsEnabledAuroraPostgresql(rName, engine),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
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
			{
				Config:      testAccAWSClusterInstanceConfigPerformanceInsightsKmsKeyIdAuroraPostgresql(rName, engine),
				ExpectError: regexp.MustCompile(`InvalidParameterCombination: You cannot change your Performance Insights KMS key`),
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_CACertificateIdentifier(t *testing.T) {
	var dbInstance rds.DBInstance
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_rds_cluster_instance.test"
	dataSourceName := "data.aws_rds_certificate.latest"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSClusterInstanceConfig_CACertificateIdentifier(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "ca_cert_identifier", dataSourceName, "id"),
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

func testAccRDSPerformanceInsightsDefaultVersionPreCheck(t *testing.T, engine string) {
	conn := testAccProvider.Meta().(*AWSClient).rdsconn

	input := &rds.DescribeDBEngineVersionsInput{
		DefaultOnly: aws.Bool(true),
		Engine:      aws.String(engine),
	}

	result, err := conn.DescribeDBEngineVersions(input)
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	if len(result.DBEngineVersions) < 1 {
		t.Fatalf("unexpected PreCheck error, no default version for engine: %s", engine)
	}

	testAccRDSPerformanceInsightsPreCheck(t, engine, aws.StringValue(result.DBEngineVersions[0].EngineVersion))
}

func testAccRDSPerformanceInsightsPreCheck(t *testing.T, engine string, engineVersion string) {
	conn := testAccProvider.Meta().(*AWSClient).rdsconn

	input := &rds.DescribeOrderableDBInstanceOptionsInput{
		Engine:        aws.String(engine),
		EngineVersion: aws.String(engineVersion), // version cuts response time
	}

	supportsPerformanceInsights := false
	err := conn.DescribeOrderableDBInstanceOptionsPages(input, func(resp *rds.DescribeOrderableDBInstanceOptionsOutput, lastPage bool) bool {
		for _, instanceOption := range resp.OrderableDBInstanceOptions {
			if instanceOption == nil {
				continue
			}

			if aws.BoolValue(instanceOption.SupportsPerformanceInsights) {
				supportsPerformanceInsights = true
				return false // stop processing pages
			}
		}
		return !lastPage
	})

	if isAWSErr(err, "InvalidParameterCombination", "") {
		t.Skipf("RDS Performance Insights not supported, skipping acceptance test")
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	if !supportsPerformanceInsights {
		t.Skipf("RDS Performance Insights not supported, skipping acceptance test")
	}
}

// Add some random to the name, to avoid collision
func testAccAWSClusterInstanceConfig(n int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine                     = aws_rds_cluster.default.engine
  engine_version             = aws_rds_cluster.default.engine_version
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t3.medium"]
}

resource "aws_rds_cluster" "default" {
  cluster_identifier = "tf-aurora-cluster-test-%[1]d"
  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]
  database_name       = "mydb"
  master_username     = "foo"
  master_password     = "mustbeeightcharacters"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "cluster_instances" {
  identifier              = "tf-cluster-instance-%[1]d"
  cluster_identifier      = aws_rds_cluster.default.id
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_parameter_group_name = aws_db_parameter_group.bar.name
  promotion_tier          = "3"
}

resource "aws_db_parameter_group" "bar" {
  name   = "tfcluster-test-group-%[1]d"
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
`, n))
}

func testAccAWSClusterInstanceConfigModified(n int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_rds_cluster" "default" {
  cluster_identifier = "tf-aurora-cluster-test-%[1]d"
  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]
  database_name       = "mydb"
  master_username     = "foo"
  master_password     = "mustbeeightcharacters"
  skip_final_snapshot = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = aws_rds_cluster.default.engine
  engine_version             = aws_rds_cluster.default.engine_version
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t3.medium"]
}

resource "aws_rds_cluster_instance" "cluster_instances" {
  identifier                 = "tf-cluster-instance-%[1]d"
  cluster_identifier         = aws_rds_cluster.default.id
  instance_class             = data.aws_rds_orderable_db_instance.test.instance_class
  db_parameter_group_name    = aws_db_parameter_group.bar.name
  auto_minor_version_upgrade = false
  promotion_tier             = "3"
}

resource "aws_db_parameter_group" "bar" {
  name   = "tfcluster-test-group-%[1]d"
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
`, n))
}

func testAccAWSClusterInstanceConfig_az(n int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_rds_cluster" "default" {
  cluster_identifier = "tf-aurora-cluster-test-%[1]d"
  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]
  database_name       = "mydb"
  master_username     = "foo"
  master_password     = "mustbeeightcharacters"
  skip_final_snapshot = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = aws_rds_cluster.default.engine
  engine_version             = aws_rds_cluster.default.engine_version
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t3.medium"]
}

resource "aws_rds_cluster_instance" "cluster_instances" {
  identifier              = "tf-cluster-instance-%[1]d"
  cluster_identifier      = aws_rds_cluster.default.id
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_parameter_group_name = aws_db_parameter_group.bar.name
  promotion_tier          = "3"
  availability_zone       = data.aws_availability_zones.available.names[0]
}

resource "aws_db_parameter_group" "bar" {
  name   = "tfcluster-test-group-%[1]d"
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
`, n))
}

func testAccAWSClusterInstanceConfig_namePrefix(n int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine                     = aws_rds_cluster.test.engine
  engine_version             = aws_rds_cluster.test.engine_version
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t3.medium"]
}

resource "aws_rds_cluster_instance" "test" {
  identifier_prefix  = "tf-cluster-instance-"
  cluster_identifier = aws_rds_cluster.test.id
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class
}

resource "aws_rds_cluster" "test" {
  cluster_identifier   = "tf-aurora-cluster-%[1]d"
  master_username      = "root"
  master_password      = "password"
  db_subnet_group_name = aws_db_subnet_group.test.name
  skip_final_snapshot  = true
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-rds-cluster-instance-name-prefix"
  }
}

resource "aws_subnet" "a" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.0.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-rds-cluster-instance-name-prefix-a"
  }
}

resource "aws_subnet" "b" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-rds-cluster-instance-name-prefix-b"
  }
}

resource "aws_db_subnet_group" "test" {
  name       = "tf-test-%[1]d"
  subnet_ids = [aws_subnet.a.id, aws_subnet.b.id]
}
`, n))
}

func testAccAWSClusterInstanceConfig_generatedName(n int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine                     = aws_rds_cluster.test.engine
  engine_version             = aws_rds_cluster.test.engine_version
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t3.medium"]
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier = aws_rds_cluster.test.id
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class
}

resource "aws_rds_cluster" "test" {
  cluster_identifier   = "tf-aurora-cluster-%[1]d"
  master_username      = "root"
  master_password      = "password"
  db_subnet_group_name = aws_db_subnet_group.test.name
  skip_final_snapshot  = true
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-rds-cluster-instance-generated-name"
  }
}

resource "aws_subnet" "a" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.0.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-rds-cluster-instance-generated-name-a"
  }
}

resource "aws_subnet" "b" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-rds-cluster-instance-generated-name-b"
  }
}

resource "aws_db_subnet_group" "test" {
  name       = "tf-test-%[1]d"
  subnet_ids = [aws_subnet.a.id, aws_subnet.b.id]
}
`, n))
}

func testAccAWSClusterInstanceConfigKmsKey(n int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_kms_key" "foo" {
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

resource "aws_rds_cluster" "default" {
  cluster_identifier = "tf-aurora-cluster-test-%[1]d"
  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]
  database_name       = "mydb"
  master_username     = "foo"
  master_password     = "mustbeeightcharacters"
  storage_encrypted   = true
  kms_key_id          = aws_kms_key.foo.arn
  skip_final_snapshot = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = aws_rds_cluster.default.engine
  engine_version             = aws_rds_cluster.default.engine_version
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t3.medium"]
}

resource "aws_rds_cluster_instance" "cluster_instances" {
  identifier              = "tf-cluster-instance-%[1]d"
  cluster_identifier      = aws_rds_cluster.default.id
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_parameter_group_name = aws_db_parameter_group.bar.name
}

resource "aws_db_parameter_group" "bar" {
  name   = "tfcluster-test-group-%[1]d"
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
`, n))
}

func testAccAWSClusterInstanceConfigMonitoringInterval(rName string, monitoringInterval int) string {
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
        "Service": "monitoring.rds.amazonaws.com"
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

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  master_username     = "foo"
  master_password     = "mustbeeightcharacters"
  skip_final_snapshot = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = aws_rds_cluster.test.engine
  engine_version             = aws_rds_cluster.test.engine_version
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t3.medium"]
}

resource "aws_rds_cluster_instance" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]

  cluster_identifier  = aws_rds_cluster.test.id
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  monitoring_interval = %[2]d
  monitoring_role_arn = aws_iam_role.test.arn
}
`, rName, monitoringInterval)
}

func testAccAWSClusterInstanceConfigMonitoringRoleArnRemoved(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  master_username     = "foo"
  master_password     = "mustbeeightcharaters"
  skip_final_snapshot = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = aws_rds_cluster.test.engine
  engine_version             = aws_rds_cluster.test.engine_version
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t3.medium"]
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier = aws_rds_cluster.test.id
  identifier         = %[1]q
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class
}
`, rName)
}

func testAccAWSClusterInstanceConfigMonitoringRoleArn(rName string) string {
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
        "Service": "monitoring.rds.amazonaws.com"
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

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  master_username     = "foo"
  master_password     = "mustbeeightcharaters"
  skip_final_snapshot = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = aws_rds_cluster.test.engine
  engine_version             = aws_rds_cluster.test.engine_version
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t3.medium"]
}

resource "aws_rds_cluster_instance" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]

  cluster_identifier  = aws_rds_cluster.test.id
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  monitoring_interval = 5
  monitoring_role_arn = aws_iam_role.test.arn
}
`, rName)
}

func testAccAWSClusterInstanceConfigPerformanceInsightsEnabledAuroraMysql1(rName, engine string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  engine              = %[2]q
  master_password     = "mustbeeightcharacters"
  master_username     = "foo"
  skip_final_snapshot = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = aws_rds_cluster.test.engine
  engine_version                = aws_rds_cluster.test.engine_version
  supports_performance_insights = true
  preferred_instance_classes    = ["db.t3.medium", "db.r5.large", "db.r4.large"]
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier           = aws_rds_cluster.test.id
  engine                       = aws_rds_cluster.test.engine
  identifier                   = %[1]q
  instance_class               = data.aws_rds_orderable_db_instance.test.instance_class
  performance_insights_enabled = true
}
`, rName, engine)
}

func testAccAWSClusterInstanceConfigPerformanceInsightsEnabledAuroraMysql2(rName, engine, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  engine              = %[2]q
  engine_version      = %[3]q
  master_password     = "mustbeeightcharacters"
  master_username     = "foo"
  skip_final_snapshot = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = aws_rds_cluster.test.engine
  engine_version                = aws_rds_cluster.test.engine_version
  supports_performance_insights = true
  preferred_instance_classes    = ["db.t3.medium", "db.r5.large", "db.r4.large"]
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier           = aws_rds_cluster.test.id
  engine                       = aws_rds_cluster.test.engine
  engine_version               = aws_rds_cluster.test.engine_version
  identifier                   = %[1]q
  instance_class               = data.aws_rds_orderable_db_instance.test.instance_class
  performance_insights_enabled = true
}
`, rName, engine, engineVersion)
}

func testAccAWSClusterInstanceConfigPerformanceInsightsEnabledAuroraPostgresql(rName, engine string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  engine              = %[2]q
  master_password     = "mustbeeightcharacters"
  master_username     = "foo"
  skip_final_snapshot = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = aws_rds_cluster.test.engine
  engine_version                = aws_rds_cluster.test.engine_version
  supports_performance_insights = true
  preferred_instance_classes    = ["db.t3.medium", "db.r5.large", "db.r4.large"]
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier           = aws_rds_cluster.test.id
  engine                       = aws_rds_cluster.test.engine
  identifier                   = %[1]q
  instance_class               = data.aws_rds_orderable_db_instance.test.instance_class
  performance_insights_enabled = true
}
`, rName, engine)
}

func testAccAWSClusterInstanceConfigPerformanceInsightsKmsKeyIdAuroraMysql1(rName, engine string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  engine              = %[2]q
  master_password     = "mustbeeightcharacters"
  master_username     = "foo"
  skip_final_snapshot = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = aws_rds_cluster.test.engine
  engine_version                = aws_rds_cluster.test.engine_version
  supports_performance_insights = true
  preferred_instance_classes    = ["db.t3.medium", "db.r5.large", "db.r4.large"]
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier              = aws_rds_cluster.test.id
  engine                          = aws_rds_cluster.test.engine
  identifier                      = %[1]q
  instance_class                  = data.aws_rds_orderable_db_instance.test.instance_class
  performance_insights_enabled    = true
  performance_insights_kms_key_id = aws_kms_key.test.arn
}
`, rName, engine)
}

func testAccAWSClusterInstanceConfigPerformanceInsightsKmsKeyIdAuroraMysql2(rName, engine, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  engine              = %[2]q
  engine_version      = %[3]q
  master_password     = "mustbeeightcharacters"
  master_username     = "foo"
  skip_final_snapshot = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = aws_rds_cluster.test.engine
  engine_version                = aws_rds_cluster.test.engine_version
  supports_performance_insights = true
  preferred_instance_classes    = ["db.t3.medium", "db.r5.large", "db.r4.large"]
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier              = aws_rds_cluster.test.id
  engine                          = aws_rds_cluster.test.engine
  engine_version                  = aws_rds_cluster.test.engine_version
  identifier                      = %[1]q
  instance_class                  = data.aws_rds_orderable_db_instance.test.instance_class
  performance_insights_enabled    = true
  performance_insights_kms_key_id = aws_kms_key.test.arn
}
`, rName, engine, engineVersion)
}

func testAccAWSClusterInstanceConfigPerformanceInsightsKmsKeyIdAuroraPostgresql(rName, engine string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  engine              = %[2]q
  master_password     = "mustbeeightcharacters"
  master_username     = "foo"
  skip_final_snapshot = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = aws_rds_cluster.test.engine
  engine_version                = aws_rds_cluster.test.engine_version
  supports_performance_insights = true
  preferred_instance_classes    = ["db.t3.medium", "db.r5.large", "db.r4.large"]
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier              = aws_rds_cluster.test.id
  engine                          = aws_rds_cluster.test.engine
  identifier                      = %[1]q
  instance_class                  = data.aws_rds_orderable_db_instance.test.instance_class
  performance_insights_enabled    = true
  performance_insights_kms_key_id = aws_kms_key.test.arn
}
`, rName, engine)
}

func testAccAWSRDSClusterInstanceConfig_PubliclyAccessible(rName string, publiclyAccessible bool) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  master_username     = "foo"
  master_password     = "mustbeeightcharacters"
  skip_final_snapshot = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = aws_rds_cluster.test.engine
  engine_version             = aws_rds_cluster.test.engine_version
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t3.medium"]
}

resource "aws_rds_cluster_instance" "test" {
  apply_immediately   = true
  cluster_identifier  = aws_rds_cluster.test.id
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  publicly_accessible = %[2]t
}
`, rName, publiclyAccessible)
}

func testAccAWSClusterInstanceConfig_CopyTagsToSnapshot(n int, f bool) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_rds_cluster" "default" {
  cluster_identifier = "tf-aurora-cluster-test-%[1]d"
  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]
  database_name       = "mydb"
  master_username     = "foo"
  master_password     = "mustbeeightcharacters"
  skip_final_snapshot = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = aws_rds_cluster.default.engine
  engine_version             = aws_rds_cluster.default.engine_version
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t3.medium"]
}

resource "aws_rds_cluster_instance" "cluster_instances" {
  identifier            = "tf-cluster-instance-%[1]d"
  cluster_identifier    = aws_rds_cluster.default.id
  instance_class        = data.aws_rds_orderable_db_instance.test.instance_class
  promotion_tier        = "3"
  copy_tags_to_snapshot = %t
}
`, n, f))
}

func testAccAWSRDSClusterInstanceConfig_CACertificateIdentifier(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  master_username     = "foo"
  master_password     = "mustbeeightcharacters"
  skip_final_snapshot = true
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = aws_rds_cluster.test.engine
  engine_version             = aws_rds_cluster.test.engine_version
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t3.medium"]
}

data "aws_rds_certificate" "latest" {
  latest_valid_till = true
}

resource "aws_rds_cluster_instance" "test" {
  apply_immediately  = true
  cluster_identifier = aws_rds_cluster.test.id
  identifier         = %[1]q
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class
  ca_cert_identifier = data.aws_rds_certificate.latest.id
}
`, rName)
}
