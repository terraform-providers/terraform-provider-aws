package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSDocdbOrderableDbInstanceDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_docdb_orderable_db_instance.test"
	class := "db.t3.medium"
	engine := "docdb"
	engineVersion := "3.6.0"
	license := "na"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDocdbOrderableDbInstance(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDocdbOrderableDbInstanceDataSourceConfigBasic(class, engine, engineVersion, license),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "instance_class", class),
					resource.TestCheckResourceAttr(dataSourceName, "engine", engine),
					resource.TestCheckResourceAttr(dataSourceName, "engine_version", engineVersion),
					resource.TestCheckResourceAttr(dataSourceName, "license_model", license),
				),
			},
		},
	})
}

func TestAccAWSDocdbOrderableDbInstanceDataSource_preferred(t *testing.T) {
	dataSourceName := "data.aws_docdb_orderable_db_instance.test"
	engine := "docdb"
	engineVersion := "3.6.0"
	license := "na"
	preferredOption := "db.r5.large"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDocdbOrderableDbInstance(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDocdbOrderableDbInstanceDataSourceConfigPreferred(engine, engineVersion, license, preferredOption),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "engine", engine),
					resource.TestCheckResourceAttr(dataSourceName, "engine_version", engineVersion),
					resource.TestCheckResourceAttr(dataSourceName, "license_model", license),
					resource.TestCheckResourceAttr(dataSourceName, "instance_class", preferredOption),
				),
			},
		},
	})
}

func testAccPreCheckAWSDocdbOrderableDbInstance(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).docdbconn

	input := &docdb.DescribeOrderableDBInstanceOptionsInput{
		Engine: aws.String("docdb"),
	}

	_, err := conn.DescribeOrderableDBInstanceOptions(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSDocdbOrderableDbInstanceDataSourceConfigBasic(class, engine, version, license string) string {
	return fmt.Sprintf(`
data "aws_docdb_orderable_db_instance" "test" {
  instance_class = %q
  engine         = %q
  engine_version = %q
  license_model  = %q
}
`, class, engine, version, license)
}

func testAccAWSDocdbOrderableDbInstanceDataSourceConfigPreferred(engine, version, license, preferredOption string) string {
	return fmt.Sprintf(`
data "aws_docdb_orderable_db_instance" "test" {
  engine         = %q
  engine_version = %q
  license_model  = %q

  preferred_instance_classes = [
    "db.xyz.xlarge",
    %q,
    "db.t3.small",
  ]
}
`, engine, version, license, preferredOption)
}
