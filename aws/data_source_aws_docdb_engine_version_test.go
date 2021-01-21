package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSDocDBEngineVersionDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_docdb_engine_version.test"
	engine := "docdb"
	version := "3.6.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAWSDocDBEngineVersionPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDocDBEngineVersionDataSourceBasicConfig(engine, version),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "engine", engine),
					resource.TestCheckResourceAttr(dataSourceName, "version", version),

					resource.TestCheckResourceAttrSet(dataSourceName, "engine_description"),
					resource.TestMatchResourceAttr(dataSourceName, "exportable_log_types.#", regexp.MustCompile(`^[1-9][0-9]*`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "parameter_group_family"),
					resource.TestCheckResourceAttrSet(dataSourceName, "supports_log_exports_to_cloudwatch"),
					resource.TestCheckResourceAttrSet(dataSourceName, "version_description"),
				),
			},
		},
	})
}

func TestAccAWSDocDBEngineVersionDataSource_preferred(t *testing.T) {
	dataSourceName := "data.aws_docdb_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAWSDocDBEngineVersionPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDocDBEngineVersionDataSourcePreferredConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "version", "3.6.0"),
				),
			},
		},
	})
}

func TestAccAWSDocDBEngineVersionDataSource_defaultOnly(t *testing.T) {
	dataSourceName := "data.aws_docdb_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAWSDocDBEngineVersionPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDocDBEngineVersionDataSourceDefaultOnlyConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "engine", "docdb"),
					resource.TestCheckResourceAttrSet(dataSourceName, "version"),
				),
			},
		},
	})
}

func testAccAWSDocDBEngineVersionPreCheck(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).docdbconn

	input := &docdb.DescribeDBEngineVersionsInput{
		Engine:      aws.String("docdb"),
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

func testAccAWSDocDBEngineVersionDataSourceBasicConfig(engine, version string) string {
	return fmt.Sprintf(`
data "aws_docdb_engine_version" "test" {
  engine  = %q
  version = %q
}
`, engine, version)
}

func testAccAWSDocDBEngineVersionDataSourcePreferredConfig() string {
	return `
data "aws_docdb_engine_version" "test" {
  preferred_versions = ["34.6.1", "3.6.0", "2.6.0"]
}
`
}

func testAccAWSDocDBEngineVersionDataSourceDefaultOnlyConfig() string {
	return `
data "aws_docdb_engine_version" "test" {}
`
}
