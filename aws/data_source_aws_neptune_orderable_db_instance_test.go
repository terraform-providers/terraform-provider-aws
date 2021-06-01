package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSNeptuneOrderableDbInstanceDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_neptune_orderable_db_instance.test"
	engine := "neptune"
	engineVersion := "1.0.2.2"
	licenseModel := "amazon-license"
	class := "db.t3.medium"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSNeptuneOrderableDbInstance(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneOrderableDbInstanceDataSourceConfigBasic(class, engine, engineVersion, licenseModel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "engine", engine),
					resource.TestCheckResourceAttr(dataSourceName, "engine_version", engineVersion),
					resource.TestCheckResourceAttr(dataSourceName, "license_model", licenseModel),
					resource.TestCheckResourceAttr(dataSourceName, "instance_class", class),
				),
			},
		},
	})
}

func TestAccAWSNeptuneOrderableDbInstanceDataSource_preferred(t *testing.T) {
	dataSourceName := "data.aws_neptune_orderable_db_instance.test"
	engine := "neptune"
	engineVersion := "1.0.3.0"
	licenseModel := "amazon-license"
	preferredOption := "db.r4.2xlarge"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSNeptuneOrderableDbInstance(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneOrderableDbInstanceDataSourceConfigPreferred(engine, engineVersion, licenseModel, preferredOption),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "engine", engine),
					resource.TestCheckResourceAttr(dataSourceName, "engine_version", engineVersion),
					resource.TestCheckResourceAttr(dataSourceName, "license_model", licenseModel),
					resource.TestCheckResourceAttr(dataSourceName, "instance_class", preferredOption),
				),
			},
		},
	})
}

func testAccPreCheckAWSNeptuneOrderableDbInstance(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).neptuneconn

	input := &neptune.DescribeOrderableDBInstanceOptionsInput{
		Engine: aws.String("mysql"),
	}

	_, err := conn.DescribeOrderableDBInstanceOptions(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSNeptuneOrderableDbInstanceDataSourceConfigBasic(class, engine, version, license string) string {
	return fmt.Sprintf(`
data "aws_neptune_orderable_db_instance" "test" {
  instance_class = %q
  engine         = %q
  engine_version = %q
  license_model  = %q
}
`, class, engine, version, license)
}

func testAccAWSNeptuneOrderableDbInstanceDataSourceConfigPreferred(engine, version, license, preferredOption string) string {
	return fmt.Sprintf(`
data "aws_neptune_orderable_db_instance" "test" {
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
