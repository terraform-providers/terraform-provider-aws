package aws

import (
	"errors"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSLightsailDatabase_basic(t *testing.T) {
	var db lightsail.RelationalDatabase
	resourceName := "aws_lightsail_database.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lightsail.EndpointsID, t)
			testAccPreCheckAWSLightsail(t)
		},
		ErrorCheck:   testAccErrorCheck(t, lightsail.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLightsailDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailDatabaseConfigBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailDatabaseExists(resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "blueprint_id", "mysql_8_0"),
					resource.TestCheckResourceAttr(resourceName, "bundle_id", "micro_1_0"),
					resource.TestCheckResourceAttr(resourceName, "master_database_name", "testdatabasename"),
					resource.TestCheckResourceAttr(resourceName, "master_username", "test"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "engine"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_version"),
					resource.TestCheckResourceAttrSet(resourceName, "cpu_count"),
					resource.TestCheckResourceAttrSet(resourceName, "ram_size"),
					resource.TestCheckResourceAttrSet(resourceName, "disk_size"),
					resource.TestCheckResourceAttrSet(resourceName, "master_endpoint_port"),
					resource.TestCheckResourceAttrSet(resourceName, "master_endpoint_address"),
					resource.TestCheckResourceAttrSet(resourceName, "support_code"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
		},
	})
}

func TestAccAWSLightsailDatabase_Name(t *testing.T) {
	var db lightsail.RelationalDatabase
	resourceName := "aws_lightsail_database.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lightsail.EndpointsID, t)
			testAccPreCheckAWSLightsail(t)
		},
		ErrorCheck:    testAccErrorCheck(t, lightsail.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLightsailDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailDatabaseConfigBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailDatabaseExists(resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
		},
	})
}

func TestAccAWSLightsailDatabase_MasterDatabaseName(t *testing.T) {
	var db lightsail.RelationalDatabase
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lightsail.EndpointsID, t)
			testAccPreCheckAWSLightsail(t)
		},
		ErrorCheck:   testAccErrorCheck(t, lightsail.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLightsailDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailDatabaseConfigMasterDatabaseName(rName, "databasename1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailDatabaseExists(resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "master_database_name", "databasename1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
			{
				Config: testAccAWSLightsailDatabaseConfigMasterDatabaseName(rName, "databasename2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailDatabaseExists(resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "master_database_name", "databasename2"),
				),
			},
		},
	})
}

func TestAccAWSLightsailDatabase_MasterUsername(t *testing.T) {
	var db lightsail.RelationalDatabase
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lightsail.EndpointsID, t)
			testAccPreCheckAWSLightsail(t)
		},
		ErrorCheck:   testAccErrorCheck(t, lightsail.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLightsailDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailDatabaseConfigMasterUsername(rName, "username1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailDatabaseExists(resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "master_username", "username1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
			{
				Config: testAccAWSLightsailDatabaseConfigMasterUsername(rName, "username2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailDatabaseExists(resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "master_username", "username2"),
				),
			},
		},
	})
}

func TestAccAWSLightsailDatabase_PreferredBackupWindow(t *testing.T) {
	var db lightsail.RelationalDatabase
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lightsail.EndpointsID, t)
			testAccPreCheckAWSLightsail(t)
		},
		ErrorCheck:   testAccErrorCheck(t, lightsail.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLightsailDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailDatabaseConfigPreferredBackupWindow(rName, "09:30-10:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailDatabaseExists(resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "preferred_backup_window", "09:30-10:00"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
			{
				Config: testAccAWSLightsailDatabaseConfigPreferredBackupWindow(rName, "09:45-10:15"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailDatabaseExists(resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "preferred_backup_window", "09:45-10:15"),
				),
			},
		},
	})
}

func TestAccAWSLightsailDatabase_PreferredMaintenanceWindow(t *testing.T) {
	var db lightsail.RelationalDatabase
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lightsail.EndpointsID, t)
			testAccPreCheckAWSLightsail(t)
		},
		ErrorCheck:   testAccErrorCheck(t, lightsail.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLightsailDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailDatabaseConfigPreferredMaintenanceWindow(rName, "tue:04:30-tue:05:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailDatabaseExists(resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "preferred_maintenance_window", "tue:04:30-tue:05:00"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
			{
				Config: testAccAWSLightsailDatabaseConfigPreferredMaintenanceWindow(rName, "wed:06:00-wed:07:30"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailDatabaseExists(resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "preferred_maintenance_window", "wed:06:00-wed:07:30"),
				),
			},
		},
	})
}

func TestAccAWSLightsailDatabase_PubliclyAccessible(t *testing.T) {
	var db lightsail.RelationalDatabase
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lightsail.EndpointsID, t)
			testAccPreCheckAWSLightsail(t)
		},
		ErrorCheck:   testAccErrorCheck(t, lightsail.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLightsailDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailDatabaseConfigPubliclyAccessible(rName, "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailDatabaseExists(resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
			{
				Config: testAccAWSLightsailDatabaseConfigPubliclyAccessible(rName, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailDatabaseExists(resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "false"),
				),
			},
		},
	})
}

func TestAccAWSLightsailDatabase_BackupRetentionEnabled(t *testing.T) {
	var db lightsail.RelationalDatabase
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lightsail.EndpointsID, t)
			testAccPreCheckAWSLightsail(t)
		},
		ErrorCheck:   testAccErrorCheck(t, lightsail.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLightsailDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailDatabaseConfigBackupRetentionEnabled(rName, "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailDatabaseExists(resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
			{
				Config: testAccAWSLightsailDatabaseConfigBackupRetentionEnabled(rName, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailDatabaseExists(resourceName, &db),
					resource.TestCheckResourceAttr(resourceName, "backup_retention_enabled", "false"),
				),
			},
		},
	})
}

func TestAccAWSLightsailDatabase_FinalSnapshotName(t *testing.T) {
	var db lightsail.RelationalDatabase
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_database.test"
	sName := fmt.Sprintf("%s-snapshot", rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lightsail.EndpointsID, t)
			testAccPreCheckAWSLightsail(t)
		},
		ErrorCheck:   testAccErrorCheck(t, lightsail.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLightsailDatabaseSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailDatabaseConfigFinalSnapshotName(rName, sName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailDatabaseExists(resourceName, &db),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
		},
	})
}

func TestAccAWSLightsailDatabase_Tags(t *testing.T) {
	var db1, db2, db3 lightsail.RelationalDatabase
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lightsail.EndpointsID, t)
			testAccPreCheckAWSLightsail(t)
		},
		IDRefreshName: resourceName,
		ErrorCheck:    testAccErrorCheck(t, lightsail.EndpointsID),
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLightsailDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailDatabaseConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailDatabaseExists(resourceName, &db1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"master_password",
					"skip_final_snapshot",
					"final_snapshot_name",
				},
			},
			{
				Config: testAccAWSLightsailDatabaseConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailDatabaseExists(resourceName, &db2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSLightsailDatabaseConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailDatabaseExists(resourceName, &db3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSLightsailDatabase_disappears(t *testing.T) {
	var db lightsail.RelationalDatabase
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lightsail.EndpointsID, t)
			testAccPreCheckAWSLightsail(t)
		},
		Providers:    testAccProviders,
		ErrorCheck:   testAccErrorCheck(t, lightsail.EndpointsID),
		CheckDestroy: testAccCheckAWSLightsailDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailDatabaseConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailDatabaseExists(resourceName, &db),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsLightsailDatabase(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSLightsailDatabaseExists(n string, res *lightsail.RelationalDatabase) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Lightsail Database ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).lightsailconn

		params := lightsail.GetRelationalDatabaseInput{
			RelationalDatabaseName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetRelationalDatabase(&params)

		if err != nil {
			return err
		}

		if resp == nil || resp.RelationalDatabase == nil {
			return fmt.Errorf("Database (%s) not found", rs.Primary.Attributes["name"])
		}
		*res = *resp.RelationalDatabase
		return nil
	}
}

func testAccCheckAWSLightsailDatabaseDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).lightsailconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lightsail_database" {
			continue
		}

		params := lightsail.GetRelationalDatabaseInput{
			RelationalDatabaseName: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetRelationalDatabase(&params)

		if err == nil {
			return fmt.Errorf("Lightsail Database %q still exists", rs.Primary.ID)
		}

		// Verify the error
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "NotFoundException" {
				return nil
			}
		}
		return err
	}

	return nil
}

func testAccCheckAWSLightsailDatabaseSnapshotDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).lightsailconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lightsail_database" {
			continue
		}

		// Try and delete the snapshot before we check for the cluster not found
		snapshot_identifier := fmt.Sprintf("%s-snapshot", rs.Primary.ID)

		log.Printf("[INFO] Deleting the Snapshot %s", snapshot_identifier)
		_, err := conn.DeleteRelationalDatabaseSnapshot(
			&lightsail.DeleteRelationalDatabaseSnapshotInput{
				RelationalDatabaseSnapshotName: aws.String(snapshot_identifier),
			})

		if err != nil {
			return err
		}

		params := lightsail.GetRelationalDatabaseInput{
			RelationalDatabaseName: aws.String(rs.Primary.ID),
		}

		_, err = conn.GetRelationalDatabase(&params)

		if err == nil {
			return fmt.Errorf("Lightsail Database %q still exists", rs.Primary.ID)
		}

		// Verify the error
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "NotFoundException" {
				return nil
			}
		}
		return err
	}

	return nil
}

func testAccAWSLightsailDatabaseConfigBase() string {
	return `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}
`
}

func testAccAWSLightsailDatabaseConfigBasic(rName string) string {
	return composeConfig(
		testAccAWSLightsailDatabaseConfigBase(),
		fmt.Sprintf(`	
resource "aws_lightsail_database" "test" {
  name                 = %[1]q
  availability_zone    = data.aws_availability_zones.available.names[0]
  master_database_name = "testdatabasename"
  master_password      = "testdatabasepassword"
  master_username      = "test"
  blueprint_id         = "mysql_8_0"
  bundle_id            = "micro_1_0"
  skip_final_snapshot  = true
}
`, rName))
}

func testAccAWSLightsailDatabaseConfigMasterDatabaseName(rName string, masterDatabaseName string) string {
	return composeConfig(
		testAccAWSLightsailDatabaseConfigBase(),
		fmt.Sprintf(`	
resource "aws_lightsail_database" "test" {
  name                 = %[1]q
  availability_zone    = data.aws_availability_zones.available.names[0]
  master_database_name = %[2]q
  master_password      = "testdatabasepassword"
  master_username      = "test"
  blueprint_id         = "mysql_8_0"
  bundle_id            = "micro_1_0"
  skip_final_snapshot  = true
}
`, rName, masterDatabaseName))
}

func testAccAWSLightsailDatabaseConfigMasterUsername(rName string, masterUsername string) string {
	return composeConfig(
		testAccAWSLightsailDatabaseConfigBase(),
		fmt.Sprintf(`	
resource "aws_lightsail_database" "test" {
  name                 = %[1]q
  availability_zone    = data.aws_availability_zones.available.names[0]
  master_database_name = "testdatabasename"
  master_password      = "testdatabasepassword"
  master_username      = %[2]q
  blueprint_id         = "mysql_8_0"
  bundle_id            = "micro_1_0"
  skip_final_snapshot  = true
}
`, rName, masterUsername))
}

func testAccAWSLightsailDatabaseConfigPreferredBackupWindow(rName string, preferredBackupWindow string) string {
	return composeConfig(
		testAccAWSLightsailDatabaseConfigBase(),
		fmt.Sprintf(`	
resource "aws_lightsail_database" "test" {
  name                    = %[1]q
  availability_zone       = data.aws_availability_zones.available.names[0]
  master_database_name    = "testdatabasename"
  master_password         = "testdatabasepassword"
  master_username         = "test"
  blueprint_id            = "mysql_8_0"
  bundle_id               = "micro_1_0"
  preferred_backup_window = %[2]q
  apply_immediately       = true
  skip_final_snapshot     = true
}
`, rName, preferredBackupWindow))
}

func testAccAWSLightsailDatabaseConfigPreferredMaintenanceWindow(rName string, preferredMaintenanceWindow string) string {
	return composeConfig(
		testAccAWSLightsailDatabaseConfigBase(),
		fmt.Sprintf(`	
resource "aws_lightsail_database" "test" {
  name                         = %[1]q
  availability_zone            = data.aws_availability_zones.available.names[0]
  master_database_name         = "testdatabasename"
  master_password              = "testdatabasepassword"
  master_username              = "test"
  blueprint_id                 = "mysql_8_0"
  bundle_id                    = "micro_1_0"
  preferred_maintenance_window = %[2]q
  apply_immediately            = true
  skip_final_snapshot          = true
}
`, rName, preferredMaintenanceWindow))
}

func testAccAWSLightsailDatabaseConfigPubliclyAccessible(rName string, publiclyAccessible string) string {
	return composeConfig(
		testAccAWSLightsailDatabaseConfigBase(),
		fmt.Sprintf(`	
resource "aws_lightsail_database" "test" {
  name                 = %[1]q
  availability_zone    = data.aws_availability_zones.available.names[0]
  master_database_name = "testdatabasename"
  master_password      = "testdatabasepassword"
  master_username      = "test"
  blueprint_id         = "mysql_8_0"
  bundle_id            = "micro_1_0"
  publicly_accessible  = %[2]q
  apply_immediately    = true
  skip_final_snapshot  = true
}
`, rName, publiclyAccessible))
}

func testAccAWSLightsailDatabaseConfigBackupRetentionEnabled(rName string, backupRetentionEnabled string) string {
	return composeConfig(
		testAccAWSLightsailDatabaseConfigBase(),
		fmt.Sprintf(`	
resource "aws_lightsail_database" "test" {
  name                     = %[1]q
  availability_zone        = data.aws_availability_zones.available.names[0]
  master_database_name     = "test"
  master_password          = "testdatabasepassword"
  master_username          = "test"
  blueprint_id             = "mysql_8_0"
  bundle_id                = "micro_1_0"
  backup_retention_enabled = %[2]q
  apply_immediately        = true
  skip_final_snapshot      = true
}
`, rName, backupRetentionEnabled))
}

func testAccAWSLightsailDatabaseConfigFinalSnapshotName(rName string, sName string) string {
	return composeConfig(
		testAccAWSLightsailDatabaseConfigBase(),
		fmt.Sprintf(`	
resource "aws_lightsail_database" "test" {
  name                 = %[1]q
  availability_zone    = data.aws_availability_zones.available.names[0]
  master_database_name = "test"
  master_password      = "testdatabasepassword"
  master_username      = "test"
  blueprint_id         = "mysql_8_0"
  bundle_id            = "micro_1_0"
  final_snapshot_name  = %[2]q
}
`, rName, sName))
}

func testAccAWSLightsailDatabaseConfigTags1(rName string, tagKey1, tagValue1 string) string {
	return composeConfig(
		testAccAWSLightsailDatabaseConfigBase(),
		fmt.Sprintf(`	
resource "aws_lightsail_database" "test" {
  name                 = %[1]q
  availability_zone    = data.aws_availability_zones.available.names[0]
  master_database_name = "testdatabasename"
  master_password      = "testdatabasepassword"
  master_username      = "test"
  blueprint_id         = "mysql_8_0"
  bundle_id            = "micro_1_0"
  skip_final_snapshot  = true
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccAWSLightsailDatabaseConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return composeConfig(
		testAccAWSLightsailDatabaseConfigBase(),
		fmt.Sprintf(`	
resource "aws_lightsail_database" "test" {
  name                 = %[1]q
  availability_zone    = data.aws_availability_zones.available.names[0]
  master_database_name = "testdatabasename"
  master_password      = "testdatabasepassword"
  master_username      = "test"
  blueprint_id         = "mysql_8_0"
  bundle_id            = "micro_1_0"
  skip_final_snapshot  = true
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
