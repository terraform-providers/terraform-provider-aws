package aws

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccAWSLakeFormationPermissions_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lakeformation.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPermissionsConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPermissionsExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "principal", roleName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", "CREATE_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "catalog_resource", "true"),
				),
			},
		},
	})
}

func testAccAWSLakeFormationPermissions_dataLocation(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	bucketName := "aws_s3_bucket.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lakeformation.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPermissionsConfig_dataLocation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPermissionsExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "principal", roleName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", "DATA_LOCATION_ACCESS"),
					resource.TestCheckResourceAttr(resourceName, "catalog_resource", "false"),
					resource.TestCheckResourceAttr(resourceName, "data_location.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "data_location.0.arn", bucketName, "arn"),
				),
			},
		},
	})
}

func testAccAWSLakeFormationPermissions_database(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	dbName := "aws_glue_catalog_database.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lakeformation.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPermissionsConfig_database(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPermissionsExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "principal", roleName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "catalog_resource", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "principal", roleName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "database.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "database.0.name", dbName, "name"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", "ALTER"),
					resource.TestCheckResourceAttr(resourceName, "permissions.1", "CREATE_TABLE"),
					resource.TestCheckResourceAttr(resourceName, "permissions.2", "DROP"),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.0", "CREATE_TABLE"),
				),
			},
		},
	})
}

func testAccAWSLakeFormationPermissions_table(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	tableName := "aws_glue_catalog_table.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lakeformation.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPermissionsConfig_table(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPermissionsExists(resourceName),
					resource.TestCheckResourceAttrPair(roleName, "arn", resourceName, "principal"),
					resource.TestCheckResourceAttr(resourceName, "table.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.database_name", tableName, "database_name"),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.name", tableName, "name"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", "ALL"),
				),
			},
		},
	})
}

func testAccAWSLakeFormationPermissions_tableWithColumns(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	tableName := "aws_glue_catalog_table.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lakeformation.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPermissionsConfig_tableWithColumns(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPermissionsExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "principal", roleName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.database_name", tableName, "database_name"),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.name", tableName, "name"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.0", "event"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.1", "timestamp"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", "SELECT"),
				),
			},
		},
	})
}

func testAccCheckAWSLakeFormationPermissionsDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).lakeformationconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lakeformation_permissions" {
			continue
		}

		principal := rs.Primary.Attributes["principal"]
		catalogId := rs.Primary.Attributes["catalog_id"]

		input := &lakeformation.ListPermissionsInput{
			CatalogId: aws.String(catalogId),
			Principal: &lakeformation.DataLakePrincipal{
				DataLakePrincipalIdentifier: aws.String(principal),
			},
		}

		out, err := conn.ListPermissions(input)
		if err == nil {
			fmt.Print(out)
			return fmt.Errorf("Resource still registered: %s %s", catalogId, principal)
		}
	}

	return nil
}

func testAccCheckAWSLakeFormationPermissionsExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).lakeformationconn

		input := &lakeformation.ListPermissionsInput{
			MaxResults: aws.Int64(1),
			Principal: &lakeformation.DataLakePrincipal{
				DataLakePrincipalIdentifier: aws.String(rs.Primary.Attributes["principal"]),
			},
		}

		if v, ok := rs.Primary.Attributes["catalog_resource"]; ok && v != "" && v == "true" {
			input.ResourceType = aws.String(lakeformation.DataLakeResourceTypeCatalog)
			input.Resource = &lakeformation.Resource{
				Catalog: &lakeformation.CatalogResource{},
			}
		}

		if v, ok := rs.Primary.Attributes["data_location.#"]; ok && v != "" && v != "0" {
			input.ResourceType = aws.String(lakeformation.DataLakeResourceTypeDataLocation)
			res := &lakeformation.DataLocationResource{
				ResourceArn: aws.String(rs.Primary.Attributes["data_location.0.arn"]),
			}
			if rs.Primary.Attributes["data_location.0.catalog_id"] != "" {
				res.CatalogId = aws.String(rs.Primary.Attributes["data_location.0.catalog_id"])
			}
			input.Resource = &lakeformation.Resource{
				DataLocation: res,
			}
		}

		if v, ok := rs.Primary.Attributes["database.#"]; ok && v != "" && v != "0" {
			input.ResourceType = aws.String(lakeformation.DataLakeResourceTypeDatabase)
			res := &lakeformation.DatabaseResource{
				Name: aws.String(rs.Primary.Attributes["database.0.name"]),
			}
			if rs.Primary.Attributes["database.0.catalog_id"] != "" {
				res.CatalogId = aws.String(rs.Primary.Attributes["database.0.catalog_id"])
			}
			input.Resource = &lakeformation.Resource{
				Database: res,
			}
		}

		if v, ok := rs.Primary.Attributes["table.#"]; ok && v != "" && v != "0" {
			input.ResourceType = aws.String(lakeformation.DataLakeResourceTypeTable)
			res := &lakeformation.TableResource{
				DatabaseName: aws.String(rs.Primary.Attributes["table.0.database_name"]),
			}
			if rs.Primary.Attributes["table.0.catalog_id"] != "" {
				res.CatalogId = aws.String(rs.Primary.Attributes["table.0.catalog_id"])
			}
			if rs.Primary.Attributes["table.0.name"] != "" {
				res.Name = aws.String(rs.Primary.Attributes["table.0.name"])
			}
			if rs.Primary.Attributes["table.0.wildcard"] == "true" {
				res.TableWildcard = &lakeformation.TableWildcard{}
			}
			input.Resource = &lakeformation.Resource{
				Table: res,
			}
		}

		// ListPermissions does not support getting privileges on a table with columns.
		// Instead, call this operation on the table, and the operation returns the
		// table and the table w columns.
		// https://docs.aws.amazon.com/sdk-for-go/api/service/lakeformation/#ListPermissionsInput
		if v, ok := rs.Primary.Attributes["table_with_columns.#"]; ok && v != "" && v != "0" {
			input.ResourceType = aws.String(lakeformation.DataLakeResourceTypeTable)
			res := &lakeformation.TableResource{
				DatabaseName: aws.String(rs.Primary.Attributes["table_with_columns.0.database_name"]),
				Name:         aws.String(rs.Primary.Attributes["table_with_columns.0.name"]),
			}
			if rs.Primary.Attributes["table.0.catalog_id"] != "" {
				res.CatalogId = aws.String(rs.Primary.Attributes["table.0.catalog_id"])
			}
			input.Resource = &lakeformation.Resource{
				Table: res,
			}
		}

		err := resource.Retry(2*time.Minute, func() *resource.RetryError {
			var err error
			_, err = conn.ListPermissions(input)
			if err != nil {
				if isAWSErr(err, lakeformation.ErrCodeInvalidInputException, "Invalid principal") {
					return resource.RetryableError(err)
				}
				if isAWSErr(err, lakeformation.ErrCodeInvalidInputException, "Grantee has no permissions") {
					return resource.RetryableError(err)
				}
				if isAWSErr(err, lakeformation.ErrCodeInvalidInputException, "register the S3 path") {
					return resource.RetryableError(err)
				}
				if isAWSErr(err, lakeformation.ErrCodeConcurrentModificationException, "") {
					return resource.RetryableError(err)
				}
				if isAWSErr(err, "AccessDeniedException", "is not authorized to access requested permissions") {
					return resource.RetryableError(err)
				}

				return resource.NonRetryableError(fmt.Errorf("unable to get Lake Formation Permissions: %w", err))
			}
			return nil
		})

		if isResourceTimeoutError(err) {
			_, err = conn.ListPermissions(input)
		}

		if err != nil {
			return fmt.Errorf("unable to get Lake Formation permissions (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccAWSLakeFormationPermissionsConfig_basic(rName string) string {
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
`, rName)
}

func testAccAWSLakeFormationPermissionsConfig_dataLocation(rName string) string {
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
`, rName)
}

func testAccAWSLakeFormationPermissionsConfig_database(rName string) string {
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
`, rName)
}

func testAccAWSLakeFormationPermissionsConfig_table(rName string) string {
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
`, rName)
}

func testAccAWSLakeFormationPermissionsConfig_tableWithColumns(rName string) string {
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

  depends_on = ["aws_lakeformation_data_lake_settings.test"]
}
`, rName)
}
