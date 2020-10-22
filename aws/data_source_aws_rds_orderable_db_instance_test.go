package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSRdsOrderableDbInstanceDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_rds_orderable_db_instance.test"
	class := "db.t2.small"
	engine := "mysql"
	engineVersion := "5.7.22"
	licenseModel := "general-public-license"
	storageType := "standard"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAWSRdsOrderableDbInstancePreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRdsOrderableDbInstanceDataSourceConfig_basic(class, engine, engineVersion, licenseModel, storageType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "instance_class", class),
					resource.TestCheckResourceAttr(dataSourceName, "engine", engine),
					resource.TestCheckResourceAttr(dataSourceName, "engine_version", engineVersion),
					resource.TestCheckResourceAttr(dataSourceName, "license_model", licenseModel),
					resource.TestCheckResourceAttr(dataSourceName, "storage_type", storageType),
				),
			},
		},
	})
}

func TestAccAWSRdsOrderableDbInstanceDataSource_preferredClass(t *testing.T) {
	dataSourceName := "data.aws_rds_orderable_db_instance.test"
	preferredClass := "db.t2.micro"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAWSRdsOrderableDbInstancePreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRdsOrderableDbInstanceDataSourceConfig_preferredClass(preferredClass),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "instance_class", preferredClass),
				),
			},
		},
	})
}

func TestAccAWSRdsOrderableDbInstanceDataSource_preferredVersion(t *testing.T) {
	dataSourceName := "data.aws_rds_orderable_db_instance.test"
	preferredVersion := "5.7.22"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAWSRdsOrderableDbInstancePreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRdsOrderableDbInstanceDataSourceConfig_preferredVersion(preferredVersion),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "engine_version", preferredVersion),
				),
			},
		},
	})
}

func TestAccAWSRdsOrderableDbInstanceDataSource_preferredClassAndVersion(t *testing.T) {
	dataSourceName := "data.aws_rds_orderable_db_instance.test"
	preferredClass := "db.m3.medium"
	preferredVersion := "5.7.22"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAWSRdsOrderableDbInstancePreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRdsOrderableDbInstanceDataSourceConfig_preferredClassAndVersion(preferredClass, preferredVersion),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "instance_class", preferredClass),
					resource.TestCheckResourceAttr(dataSourceName, "engine_version", preferredVersion),
				),
			},
		},
	})
}

func TestAccAWSRdsOrderableDbInstanceDataSource_supportsEnhancedMonitoring(t *testing.T) {
	dataSourceName := "data.aws_rds_orderable_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAWSRdsOrderableDbInstancePreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRdsOrderableDbInstanceDataSourceConfig_supportsEnhancedMonitoring(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "supports_enhanced_monitoring", "true"),
				),
			},
		},
	})
}

func TestAccAWSRdsOrderableDbInstanceDataSource_supportsIAMDatabaseAuthentication(t *testing.T) {
	dataSourceName := "data.aws_rds_orderable_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAWSRdsOrderableDbInstancePreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRdsOrderableDbInstanceDataSourceConfig_supportsIAMDatabaseAuthentication(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "supports_iam_database_authentication", "true"),
				),
			},
		},
	})
}

func TestAccAWSRdsOrderableDbInstanceDataSource_supportsIops(t *testing.T) {
	dataSourceName := "data.aws_rds_orderable_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAWSRdsOrderableDbInstancePreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRdsOrderableDbInstanceDataSourceConfig_supportsIops(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "supports_iops", "true"),
				),
			},
		},
	})
}

func TestAccAWSRdsOrderableDbInstanceDataSource_supportsKerberosAuthentication(t *testing.T) {
	dataSourceName := "data.aws_rds_orderable_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAWSRdsOrderableDbInstancePreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRdsOrderableDbInstanceDataSourceConfig_supportsKerberosAuthentication(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "supports_kerberos_authentication", "true"),
				),
			},
		},
	})
}

func TestAccAWSRdsOrderableDbInstanceDataSource_supportsPerformanceInsights(t *testing.T) {
	dataSourceName := "data.aws_rds_orderable_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAWSRdsOrderableDbInstancePreCheck(t)
			testAccRDSPerformanceInsightsDefaultVersionPreCheck(t, "mysql")
		},
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRdsOrderableDbInstanceDataSourceConfig_supportsPerformanceInsights(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "supports_performance_insights", "true"),
				),
			},
		},
	})
}

func TestAccAWSRdsOrderableDbInstanceDataSource_supportsStorageAutoscaling(t *testing.T) {
	dataSourceName := "data.aws_rds_orderable_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAWSRdsOrderableDbInstancePreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRdsOrderableDbInstanceDataSourceConfig_supportsStorageAutoscaling(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "supports_storage_autoscaling", "true"),
				),
			},
		},
	})
}

func TestAccAWSRdsOrderableDbInstanceDataSource_supportsStorageEncryption(t *testing.T) {
	dataSourceName := "data.aws_rds_orderable_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAWSRdsOrderableDbInstancePreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRdsOrderableDbInstanceDataSourceConfig_supportsStorageEncryption(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "supports_storage_encryption", "true"),
				),
			},
		},
	})
}

func testAccAWSRdsOrderableDbInstancePreCheck(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).rdsconn

	input := &rds.DescribeOrderableDBInstanceOptionsInput{
		Engine:          aws.String("mysql"),
		EngineVersion:   aws.String("8.0.20"),
		DBInstanceClass: aws.String("db.m5.xlarge"),
	}

	_, err := conn.DescribeOrderableDBInstanceOptions(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSRdsOrderableDbInstanceDataSourceConfig_basic(class, engine, version, license, storage string) string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  instance_class = %q
  engine         = %q
  engine_version = %q
  license_model  = %q
  storage_type   = %q
}
`, class, engine, version, license, storage)
}

func testAccAWSRdsOrderableDbInstanceDataSourceConfig_preferredClass(preferredClass string) string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine         = "mysql"
  engine_version = "5.7.22"
  license_model  = "general-public-license"

  preferred_instance_classes = [
    "db.xyz.xlarge",
    %q,
    "db.t3.small",
  ]
}
`, preferredClass)
}

func testAccAWSRdsOrderableDbInstanceDataSourceConfig_preferredVersion(preferredVersion string) string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine        = "mysql"
  license_model = "general-public-license"
  storage_type  = "standard"

  preferred_engine_versions = [
    "18.42.32",
    %q,
    "not.a.version",
  ]
}
`, preferredVersion)
}

func testAccAWSRdsOrderableDbInstanceDataSourceConfig_preferredClassAndVersion(preferredClass, preferredVersion string) string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine        = "mysql"
  license_model = "general-public-license"

  preferred_instance_classes = [
    "db.xyz.xlarge",
    %q,
    "db.t3.small",
  ]
  preferred_engine_versions = [
    "18.42.32",
    %q,
    "not.a.version",
  ]
}
`, preferredClass, preferredVersion)
}

func testAccAWSRdsOrderableDbInstanceDataSourceConfig_supportsEnhancedMonitoring() string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine                       = "mysql"
  license_model                = "general-public-license"
  storage_type                 = "standard"
  supports_enhanced_monitoring = true

  preferred_engine_versions  = ["5.6.35", "5.6.41", "5.6.44"]
  preferred_instance_classes = ["db.t2.small", "db.t3.medium", "db.t3.large"]
}
`)
}

func testAccAWSRdsOrderableDbInstanceDataSourceConfig_supportsIAMDatabaseAuthentication() string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine                               = "mysql"
  license_model                        = "general-public-license"
  storage_type                         = "standard"
  supports_iam_database_authentication = true

  preferred_engine_versions  = ["5.6.35", "5.6.41", "5.6.44"]
  preferred_instance_classes = ["db.t2.small", "db.t3.medium", "db.t3.large"]
}
`)
}

func testAccAWSRdsOrderableDbInstanceDataSourceConfig_supportsIops() string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine        = "mysql"
  license_model = "general-public-license"
  supports_iops = true

  preferred_engine_versions  = ["8.0.20", "8.0.19", "8.0.17"]
  preferred_instance_classes = ["db.t3.small", "db.t2.xlarge", "db.t2.small"]
}
`)
}

func testAccAWSRdsOrderableDbInstanceDataSourceConfig_supportsKerberosAuthentication() string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine                           = "postgres"
  license_model                    = "postgresql-license"
  storage_type                     = "standard"
  supports_kerberos_authentication = true

  preferred_engine_versions  = ["12.3", "11.1", "10.13"]
  preferred_instance_classes = ["db.m5.xlarge", "db.r5.large", "db.t3.large"]
}
`)
}

func testAccAWSRdsOrderableDbInstanceDataSourceConfig_supportsPerformanceInsights() string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine                        = "mysql"
  license_model                 = "general-public-license"
  supports_performance_insights = true

  preferred_engine_versions  = ["5.6.35", "5.6.41", "5.6.44"]
  preferred_instance_classes = ["db.t2.small", "db.t3.medium", "db.t3.large"]
}
`)
}

func testAccAWSRdsOrderableDbInstanceDataSourceConfig_supportsStorageAutoscaling() string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine                       = "mysql"
  license_model                = "general-public-license"
  supports_storage_autoscaling = true

  preferred_engine_versions  = ["8.0.20", "8.0.19", "5.7.30"]
  preferred_instance_classes = ["db.t3.medium", "db.t2.large", "db.t3.xlarge"]
}
`)
}

func testAccAWSRdsOrderableDbInstanceDataSourceConfig_supportsStorageEncryption() string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine                      = "mysql"
  license_model               = "general-public-license"
  storage_type                = "standard"
  supports_storage_encryption = true

  preferred_engine_versions  = ["5.6.35", "5.6.41", "5.6.44"]
  preferred_instance_classes = ["db.t2.small", "db.t3.medium", "db.t3.large"]
}
`)
}
