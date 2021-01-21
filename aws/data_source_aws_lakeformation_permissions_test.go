package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testAccAWSLakeFormationPermissionsDataSource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_permissions.test"
	dataSourceName := "data.aws_lakeformation_permissions.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lakeformation.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPermissionsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "principal", dataSourceName, "principal"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.#", dataSourceName, "permissions.#"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.0", dataSourceName, "permissions.0"),
					resource.TestCheckResourceAttrPair(resourceName, "catalog_resource", dataSourceName, "catalog_resource"),
				),
			},
		},
	})
}

func testAccAWSLakeFormationPermissionsDataSource_dataLocation(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_permissions.test"
	dataSourceName := "data.aws_lakeformation_permissions.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lakeformation.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPermissionsDataSourceConfig_dataLocation(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "principal", dataSourceName, "principal"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.#", dataSourceName, "permissions.#"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.0", dataSourceName, "permissions.0"),
					resource.TestCheckResourceAttrPair(resourceName, "data_location.#", dataSourceName, "data_location.#"),
					resource.TestCheckResourceAttrPair(resourceName, "data_location.0.arn", dataSourceName, "data_location.0.arn"),
				),
			},
		},
	})
}

func testAccAWSLakeFormationPermissionsDataSource_database(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_permissions.test"
	dataSourceName := "data.aws_lakeformation_permissions.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lakeformation.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPermissionsDataSourceConfig_database(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "principal", dataSourceName, "principal"),
					resource.TestCheckResourceAttrPair(resourceName, "database.#", dataSourceName, "database.#"),
					resource.TestCheckResourceAttrPair(resourceName, "database.0.name", dataSourceName, "database.0.name"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.#", dataSourceName, "permissions.#"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.0", dataSourceName, "permissions.0"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.1", dataSourceName, "permissions.1"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.2", dataSourceName, "permissions.2"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions_with_grant_option.#", dataSourceName, "permissions_with_grant_option.#"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions_with_grant_option.0", dataSourceName, "permissions_with_grant_option.0"),
				),
			},
		},
	})
}

func testAccAWSLakeFormationPermissionsDataSource_table(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_permissions.test"
	dataSourceName := "data.aws_lakeformation_permissions.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lakeformation.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPermissionsDataSourceConfig_table(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "principal", dataSourceName, "principal"),
					resource.TestCheckResourceAttrPair(resourceName, "table.#", dataSourceName, "table.#"),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.database_name", dataSourceName, "table.0.database_name"),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.name", dataSourceName, "table.0.name"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.#", dataSourceName, "permissions.#"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.0", dataSourceName, "permissions.0"),
				),
			},
		},
	})
}

func testAccAWSLakeFormationPermissionsDataSource_tableWithColumns(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_permissions.test"
	dataSourceName := "data.aws_lakeformation_permissions.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lakeformation.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPermissionsDataSourceConfig_tableWithColumns(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "principal", dataSourceName, "principal"),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.#", dataSourceName, "table_with_columns.#"),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.database_name", dataSourceName, "table_with_columns.0.database_name"),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.name", dataSourceName, "table_with_columns.0.name"),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.column_names.#", dataSourceName, "table_with_columns.0.column_names.#"),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.column_names.0", dataSourceName, "table_with_columns.0.column_names.0"),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.column_names.1", dataSourceName, "table_with_columns.0.column_names.1"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.#", dataSourceName, "permissions.#"),
					resource.TestCheckResourceAttrPair(resourceName, "permissions.0", dataSourceName, "permissions.0"),
				),
			},
		},
	})
}

func testAccAWSLakeFormationPermissionsDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lakeformation.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

data "aws_caller_identity" "current" {}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_caller_identity.current.arn]
}

resource "aws_lakeformation_permissions" "test" {
  principal        = aws_iam_role.test.arn
  permissions      = ["CREATE_DATABASE"]
  catalog_resource = true
}

data "aws_lakeformation_permissions" "test" {
  principal        = aws_lakeformation_permissions.test.principal
  catalog_resource = true
}
`, rName)
}

func testAccAWSLakeFormationPermissionsDataSourceConfig_dataLocation(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "glue.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  acl           = "private"
  force_destroy = true
}

resource "aws_lakeformation_resource" "test" {
  arn = aws_s3_bucket.test.arn
}

data "aws_caller_identity" "current" {}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_caller_identity.current.arn]
}

resource "aws_lakeformation_permissions" "test" {
  principal   = aws_iam_role.test.arn
  permissions = ["DATA_LOCATION_ACCESS"]

  data_location {
    arn = aws_s3_bucket.test.arn
  }
}

data "aws_lakeformation_permissions" "test" {
  principal = aws_lakeformation_permissions.test.principal

  data_location {
    arn = aws_s3_bucket.test.arn
  }
}
`, rName)
}

func testAccAWSLakeFormationPermissionsDataSourceConfig_database(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "glue.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

data "aws_caller_identity" "current" {}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_caller_identity.current.arn]
}

resource "aws_lakeformation_permissions" "test" {
  permissions                   = ["ALTER", "CREATE_TABLE", "DROP"]
  permissions_with_grant_option = ["CREATE_TABLE"]
  principal                     = aws_iam_role.test.arn

  database {
    name = aws_glue_catalog_database.test.name
  }

  depends_on = ["aws_lakeformation_data_lake_settings.test"]
}

data "aws_lakeformation_permissions" "test" {
  principal = aws_lakeformation_permissions.test.principal

  database {
    name = aws_glue_catalog_database.test.name
  }
}
`, rName)
}

func testAccAWSLakeFormationPermissionsDataSourceConfig_table(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "glue.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

data "aws_caller_identity" "current" {}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_caller_identity.current.arn]
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["ALL"]
  principal   = aws_iam_role.test.arn

  table {
    database_name = aws_glue_catalog_table.test.database_name
    name          = aws_glue_catalog_table.test.name
  }
}

data "aws_lakeformation_permissions" "test" {
  principal = aws_lakeformation_permissions.test.principal

  table {
    database_name = aws_glue_catalog_table.test.database_name
    name          = aws_glue_catalog_table.test.name
  }
}
`, rName)
}

func testAccAWSLakeFormationPermissionsDataSourceConfig_tableWithColumns(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "glue.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

data "aws_caller_identity" "current" {}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    columns {
      name = "event"
      type = "string"
    }
    columns {
      name = "timestamp"
      type = "date"
    }
    columns {
      name = "value"
      type = "double"
    }
  }
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_caller_identity.current.arn]
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["SELECT"]
  principal   = aws_iam_role.test.arn

  table_with_columns {
    database_name = aws_glue_catalog_table.test.database_name
    name          = aws_glue_catalog_table.test.name
    column_names  = ["event", "timestamp"]
  }
}

data "aws_lakeformation_permissions" "test" {
  principal = aws_lakeformation_permissions.test.principal

  table_with_columns {
    database_name = aws_glue_catalog_table.test.database_name
    name          = aws_glue_catalog_table.test.name
    column_names  = ["event", "timestamp"]
  }
}
`, rName)
}
