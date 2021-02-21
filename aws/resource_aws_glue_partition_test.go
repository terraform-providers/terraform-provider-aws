package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/glue/finder"
)

func TestAccAWSGluePartition_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	parValue := acctest.RandString(10)
	resourceName := "aws_glue_partition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGluePartitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGluePartitionBasicConfig(rName, parValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGluePartitionExists(resourceName),
					testAccCheckResourceAttrAccountID(resourceName, "catalog_id"),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "partition_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "partition_values.0", parValue),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "creation_time"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSGluePartition_multipleValues(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	parValue := acctest.RandString(10)
	parValue2 := acctest.RandString(11)
	resourceName := "aws_glue_partition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGluePartitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGluePartitionMultiplePartValueConfig(rName, parValue, parValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGluePartitionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "partition_values.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "partition_values.0", parValue),
					resource.TestCheckResourceAttr(resourceName, "partition_values.1", parValue2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSGluePartition_parameters(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	parValue := acctest.RandString(10)
	resourceName := "aws_glue_partition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGluePartitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGluePartitionParametersConfig1(rName, parValue, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGluePartitionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGluePartitionParametersConfig2(rName, parValue, "key1", "valueUpdated1", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGluePartitionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "parameters.key1", "valueUpdated1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.key2", "value2"),
				),
			},
			{
				Config: testAccGluePartitionParametersConfig1(rName, parValue, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGluePartitionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSGluePartition_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	parValue := acctest.RandString(10)
	resourceName := "aws_glue_partition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGluePartitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGluePartitionBasicConfig(rName, parValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGluePartitionExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsGluePartition(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSGluePartition_disappears_table(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	parValue := acctest.RandString(10)
	resourceName := "aws_glue_partition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGluePartitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGluePartitionBasicConfig(rName, parValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGluePartitionExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsGlueCatalogTable(), "aws_glue_catalog_table.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckGluePartitionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).glueconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_glue_partition" {
			continue
		}

		if _, err := finder.PartitionByValues(conn, rs.Primary.ID); err != nil {
			if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
				continue
			}

			return err
		}
		return fmt.Errorf("still exists")
	}
	return nil
}

func testAccCheckGluePartitionExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).glueconn
		out, err := finder.PartitionByValues(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if out == nil {
			return fmt.Errorf("No Glue Partition Found")
		}

		return nil
	}
}

func testAccGluePartitionConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    bucket_columns            = ["bucket_column_1"]
    compressed                = false
    input_format              = "SequenceFileInputFormat"
    location                  = "my_location"
    number_of_buckets         = 1
    output_format             = "SequenceFileInputFormat"
    stored_as_sub_directories = false

    parameters = {
      param1 = "param1_val"
    }

    columns {
      name    = "my_column_1"
      type    = "int"
      comment = "my_column1_comment"
    }

    columns {
      name    = "my_column_2"
      type    = "string"
      comment = "my_column2_comment"
    }

    ser_de_info {
      name = "ser_de_name"

      parameters = {
        param1 = "param_val_1"
      }

      serialization_library = "org.apache.hadoop.hive.serde2.columnar.ColumnarSerDe"
    }

    sort_columns {
      column     = "my_column_1"
      sort_order = 1
    }

    skewed_info {
      skewed_column_names = [
        "my_column_1",
      ]

      skewed_column_value_location_maps = {
        my_column_1 = "my_column_1_val_loc_map"
      }

      skewed_column_values = [
        "skewed_val_1",
      ]
    }
  }

  partition_keys {
    name    = "my_column_12"
    type    = "date"
    comment = "my_column_1_comment2"
  }
}
`, rName)
}

func testAccGluePartitionBasicConfig(rName, parValue string) string {
	return testAccGluePartitionConfigBase(rName) +
		fmt.Sprintf(`
resource "aws_glue_partition" "test" {
  database_name    = aws_glue_catalog_database.test.name
  table_name       = aws_glue_catalog_table.test.name
  partition_values = ["%[1]s"]
}
`, parValue)
}

func testAccGluePartitionParametersConfig1(rName, parValue, key1, value1 string) string {
	return testAccGluePartitionConfigBase(rName) +
		fmt.Sprintf(`
resource "aws_glue_partition" "test" {
  database_name    = aws_glue_catalog_database.test.name
  table_name       = aws_glue_catalog_table.test.name
  partition_values = ["%[1]s"]

  parameters = {
    %[2]q = %[3]q
  }
}
`, parValue, key1, value1)
}

func testAccGluePartitionParametersConfig2(rName, parValue, key1, value1, key2, value2 string) string {
	return testAccGluePartitionConfigBase(rName) +
		fmt.Sprintf(`
resource "aws_glue_partition" "test" {
  database_name    = aws_glue_catalog_database.test.name
  table_name       = aws_glue_catalog_table.test.name
  partition_values = ["%[1]s"]

  parameters = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, parValue, key1, value1, key2, value2)
}

func testAccGluePartitionMultiplePartValueConfig(rName, parValue, parValue2 string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    bucket_columns            = ["bucket_column_1"]
    compressed                = false
    input_format              = "SequenceFileInputFormat"
    location                  = "my_location"
    number_of_buckets         = 1
    output_format             = "SequenceFileInputFormat"
    stored_as_sub_directories = false

    parameters = {
      param1 = "param1_val"
    }

    columns {
      name    = "my_column_1"
      type    = "int"
      comment = "my_column1_comment"
    }

    columns {
      name    = "my_column_2"
      type    = "string"
      comment = "my_column2_comment"
    }

    ser_de_info {
      name = "ser_de_name"

      parameters = {
        param1 = "param_val_1"
      }

      serialization_library = "org.apache.hadoop.hive.serde2.columnar.ColumnarSerDe"
    }

    sort_columns {
      column     = "my_column_1"
      sort_order = 1
    }

    skewed_info {
      skewed_column_names = [
        "my_column_1",
      ]

      skewed_column_value_location_maps = {
        my_column_1 = "my_column_1_val_loc_map"
      }

      skewed_column_values = [
        "skewed_val_1",
      ]
    }
  }

  partition_keys {
    name    = "my_column_12"
    type    = "date"
    comment = "my_column_1_comment2"
  }

  partition_keys {
    name    = "my_column_11"
    type    = "date"
    comment = "my_column_1_comment2"
  }
}

resource "aws_glue_partition" "test" {
  database_name    = aws_glue_catalog_database.test.name
  table_name       = aws_glue_catalog_table.test.name
  partition_values = ["%[2]s", "%[3]s"]
}
`, rName, parValue, parValue2)
}
