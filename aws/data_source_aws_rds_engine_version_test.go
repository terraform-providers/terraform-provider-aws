package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSRDSEngineVersionDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_rds_engine_version.test"
	engine := "oracle-ee"
	version := "19.0.0.0.ru-2020-07.rur-2020-07.r1"
	paramGroup := "oracle-ee-19"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAWSRDSEngineVersionPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSEngineVersionDataSourceBasicConfig(engine, version, paramGroup),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "engine", engine),
					resource.TestCheckResourceAttr(dataSourceName, "version", version),
					resource.TestCheckResourceAttr(dataSourceName, "parameter_group_family", paramGroup),

					resource.TestCheckResourceAttrSet(dataSourceName, "default_character_set"),
					resource.TestCheckResourceAttrSet(dataSourceName, "engine_description"),
					resource.TestMatchResourceAttr(dataSourceName, "exportable_log_types.#", regexp.MustCompile(`^[1-9][0-9]*`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "status"),
					resource.TestMatchResourceAttr(dataSourceName, "supported_character_sets.#", regexp.MustCompile(`^[1-9][0-9]*`)),
					resource.TestMatchResourceAttr(dataSourceName, "supported_feature_names.#", regexp.MustCompile(`^[1-9][0-9]*`)),
					resource.TestMatchResourceAttr(dataSourceName, "supported_modes.#", regexp.MustCompile(`^[0-9]*`)),
					resource.TestMatchResourceAttr(dataSourceName, "supported_timezones.#", regexp.MustCompile(`^[0-9]*`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "supports_global_databases"),
					resource.TestCheckResourceAttrSet(dataSourceName, "supports_log_exports_to_cloudwatch"),
					resource.TestCheckResourceAttrSet(dataSourceName, "supports_parallel_query"),
					resource.TestCheckResourceAttrSet(dataSourceName, "supports_read_replica"),
					resource.TestCheckResourceAttrSet(dataSourceName, "version_description"),
				),
			},
		},
	})
}

func TestAccAWSRDSEngineVersionDataSource_upgradeTargets(t *testing.T) {
	dataSourceName := "data.aws_rds_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAWSRDSEngineVersionPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSEngineVersionDataSourceUpgradeTargetsConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "valid_upgrade_targets.#", regexp.MustCompile(`^[1-9][0-9]*`)),
				),
			},
		},
	})
}

func TestAccAWSRDSEngineVersionDataSource_preferred(t *testing.T) {
	dataSourceName := "data.aws_rds_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAWSRDSEngineVersionPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSEngineVersionDataSourcePreferredConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "version", "5.7.19"),
				),
			},
		},
	})
}

func TestAccAWSRDSEngineVersionDataSource_defaultOnly(t *testing.T) {
	dataSourceName := "data.aws_rds_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAWSRDSEngineVersionPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSEngineVersionDataSourceDefaultOnlyConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "version"),
				),
			},
		},
	})
}

func testAccAWSRDSEngineVersionPreCheck(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).rdsconn

	input := &rds.DescribeDBEngineVersionsInput{
		Engine:      aws.String("mysql"),
		DefaultOnly: aws.Bool(true),
	}

	_, err := conn.DescribeDBEngineVersions(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSRDSEngineVersionDataSourceBasicConfig(engine, version, paramGroup string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine                 = %q
  version                = %q
  parameter_group_family = %q
}
`, engine, version, paramGroup)
}

func testAccAWSRDSEngineVersionDataSourceUpgradeTargetsConfig() string {
	return `
data "aws_rds_engine_version" "test" {
  engine  = "mysql"
  version = "5.7.17"
}
`
}

func testAccAWSRDSEngineVersionDataSourcePreferredConfig() string {
	return `
data "aws_rds_engine_version" "test" {
  engine             = "mysql"
  preferred_versions = ["85.9.12", "5.7.19", "5.7.17"]
}
`
}

func testAccAWSRDSEngineVersionDataSourceDefaultOnlyConfig() string {
	return `
data "aws_rds_engine_version" "test" {
  engine = "mysql"
}
`
}
