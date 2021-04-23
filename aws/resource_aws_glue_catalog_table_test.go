package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSGlueCatalogTable_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_catalog_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlueTableDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccGlueCatalogTable_basic(rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlueCatalogTableExists(resourceName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("table/%s/%s", rName, rName)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.#", "0"),
					testAccCheckResourceAttrAccountID(resourceName, "catalog_id"),
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

func TestAccAWSGlueCatalogTable_columnParameters(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_catalog_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlueTableDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccGlueCatalogTableColumnParameters(rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlueCatalogTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.0.name", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.0.parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.0.parameters.param2", "param2_val"),
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

func TestAccAWSGlueCatalogTable_full(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	description := "A test table from terraform"
	resourceName := "aws_glue_catalog_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlueTableDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccGlueCatalogTable_full(rName, description),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlueCatalogTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "owner", "my_owner"),
					resource.TestCheckResourceAttr(resourceName, "retention", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.0.name", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.0.type", "int"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.0.comment", "my_column1_comment"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.1.name", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.1.type", "string"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.1.comment", "my_column2_comment"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.location", "my_location"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.input_format", "SequenceFileInputFormat"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.output_format", "SequenceFileInputFormat"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.compressed", "false"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.number_of_buckets", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.ser_de_info.0.name", "ser_de_name"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.ser_de_info.0.parameters.param1", "param_val_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.ser_de_info.0.serialization_library", "org.apache.hadoop.hive.serde2.columnar.ColumnarSerDe"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.bucket_columns.0", "bucket_column_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.sort_columns.0.column", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.sort_columns.0.sort_order", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.parameters.param1", "param1_val"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.skewed_info.0.skewed_column_names.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.skewed_info.0.skewed_column_value_location_maps.my_column_1", "my_column_1_val_loc_map"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.skewed_info.0.skewed_column_values.0", "skewed_val_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.stored_as_sub_directories", "false"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.0.name", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.0.type", "int"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.0.comment", "my_column_1_comment"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.1.name", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.1.type", "string"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.1.comment", "my_column_2_comment"),
					resource.TestCheckResourceAttr(resourceName, "view_original_text", "view_original_text_1"),
					resource.TestCheckResourceAttr(resourceName, "view_expanded_text", "view_expanded_text_1"),
					resource.TestCheckResourceAttr(resourceName, "table_type", "VIRTUAL_VIEW"),
					resource.TestCheckResourceAttr(resourceName, "parameters.param1", "param1_val"),
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

func TestAccAWSGlueCatalogTable_update_addValues(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	description := "A test table from terraform"
	resourceName := "aws_glue_catalog_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlueTableDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccGlueCatalogTable_basic(rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlueCatalogTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:  testAccGlueCatalogTable_full(rName, description),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlueCatalogTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "owner", "my_owner"),
					resource.TestCheckResourceAttr(resourceName, "retention", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.0.name", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.0.type", "int"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.0.comment", "my_column1_comment"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.1.name", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.1.type", "string"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.1.comment", "my_column2_comment"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.location", "my_location"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.input_format", "SequenceFileInputFormat"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.output_format", "SequenceFileInputFormat"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.compressed", "false"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.number_of_buckets", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.ser_de_info.0.name", "ser_de_name"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.ser_de_info.0.parameters.param1", "param_val_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.ser_de_info.0.serialization_library", "org.apache.hadoop.hive.serde2.columnar.ColumnarSerDe"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.bucket_columns.0", "bucket_column_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.sort_columns.0.column", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.sort_columns.0.sort_order", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.parameters.param1", "param1_val"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.skewed_info.0.skewed_column_names.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.skewed_info.0.skewed_column_value_location_maps.my_column_1", "my_column_1_val_loc_map"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.skewed_info.0.skewed_column_values.0", "skewed_val_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.stored_as_sub_directories", "false"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.0.name", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.0.type", "int"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.0.comment", "my_column_1_comment"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.1.name", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.1.type", "string"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.1.comment", "my_column_2_comment"),
					resource.TestCheckResourceAttr(resourceName, "view_original_text", "view_original_text_1"),
					resource.TestCheckResourceAttr(resourceName, "view_expanded_text", "view_expanded_text_1"),
					resource.TestCheckResourceAttr(resourceName, "table_type", "VIRTUAL_VIEW"),
					resource.TestCheckResourceAttr(resourceName, "parameters.param1", "param1_val"),
				),
			},
		},
	})
}

func TestAccAWSGlueCatalogTable_update_replaceValues(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	description := "A test table from terraform"
	resourceName := "aws_glue_catalog_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlueTableDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccGlueCatalogTable_full(rName, description),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlueCatalogTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "owner", "my_owner"),
					resource.TestCheckResourceAttr(resourceName, "retention", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.0.name", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.0.type", "int"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.0.comment", "my_column1_comment"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.1.name", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.1.type", "string"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.1.comment", "my_column2_comment"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.location", "my_location"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.input_format", "SequenceFileInputFormat"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.output_format", "SequenceFileInputFormat"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.compressed", "false"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.number_of_buckets", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.ser_de_info.0.name", "ser_de_name"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.ser_de_info.0.parameters.param1", "param_val_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.ser_de_info.0.serialization_library", "org.apache.hadoop.hive.serde2.columnar.ColumnarSerDe"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.bucket_columns.0", "bucket_column_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.sort_columns.0.column", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.sort_columns.0.sort_order", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.parameters.param1", "param1_val"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.skewed_info.0.skewed_column_names.0", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.skewed_info.0.skewed_column_value_location_maps.my_column_1", "my_column_1_val_loc_map"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.skewed_info.0.skewed_column_values.0", "skewed_val_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.stored_as_sub_directories", "false"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.0.name", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.0.type", "int"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.0.comment", "my_column_1_comment"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.1.name", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.1.type", "string"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.1.comment", "my_column_2_comment"),
					resource.TestCheckResourceAttr(resourceName, "view_original_text", "view_original_text_1"),
					resource.TestCheckResourceAttr(resourceName, "view_expanded_text", "view_expanded_text_1"),
					resource.TestCheckResourceAttr(resourceName, "table_type", "VIRTUAL_VIEW"),
					resource.TestCheckResourceAttr(resourceName, "parameters.param1", "param1_val"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:  testAccGlueCatalogTable_full_replacedValues(rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlueCatalogTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "A test table from terraform2"),
					resource.TestCheckResourceAttr(resourceName, "owner", "my_owner2"),
					resource.TestCheckResourceAttr(resourceName, "retention", "2"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.0.name", "my_column_12"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.0.type", "date"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.0.comment", "my_column1_comment2"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.1.name", "my_column_22"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.1.type", "timestamp"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.columns.1.comment", "my_column2_comment2"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.location", "my_location2"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.input_format", "TextInputFormat"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.output_format", "TextInputFormat"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.compressed", "true"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.number_of_buckets", "12"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.ser_de_info.0.name", "ser_de_name2"),
					resource.TestCheckNoResourceAttr(resourceName, "storage_descriptor.0.ser_de_info.0.parameters.param1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.ser_de_info.0.parameters.param2", "param_val_12"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.ser_de_info.0.serialization_library", "org.apache.hadoop.hive.serde2.columnar.ColumnarSerDe2"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.bucket_columns.0", "bucket_column_12"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.bucket_columns.1", "bucket_column_2"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.sort_columns.0.column", "my_column_12"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.sort_columns.0.sort_order", "0"),
					resource.TestCheckNoResourceAttr(resourceName, "storage_descriptor.0.parameters.param1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.parameters.param12", "param1_val2"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.skewed_info.0.skewed_column_names.0", "my_column_12"),
					resource.TestCheckNoResourceAttr(resourceName, "storage_descriptor.0.skewed_info.0.skewed_column_value_location_maps.my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.skewed_info.0.skewed_column_value_location_maps.my_column_12", "my_column_1_val_loc_map2"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.skewed_info.0.skewed_column_values.0", "skewed_val_12"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.skewed_info.0.skewed_column_values.1", "skewed_val_2"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.stored_as_sub_directories", "true"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.0.name", "my_column_12"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.0.type", "date"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.0.comment", "my_column_1_comment2"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.1.name", "my_column_22"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.1.type", "timestamp"),
					resource.TestCheckResourceAttr(resourceName, "partition_keys.1.comment", "my_column_2_comment2"),
					resource.TestCheckResourceAttr(resourceName, "view_original_text", "view_original_text_12"),
					resource.TestCheckResourceAttr(resourceName, "view_expanded_text", "view_expanded_text_12"),
					resource.TestCheckResourceAttr(resourceName, "table_type", "EXTERNAL_TABLE"),
					//resource.TestCheckResourceAttr(resourceName, "parameters.param1", "param1_val"),
					resource.TestCheckResourceAttr(resourceName, "parameters.param2", "param1_val2"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11784
func TestAccAWSGlueCatalogTable_StorageDescriptor_EmptyConfigurationBlock(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_catalog_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlueTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCatalogTableConfigStorageDescriptorEmptyConfigurationBlock(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlueCatalogTableExists(resourceName),
				),
				// Expect non-empty instead of panic
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11784
func TestAccAWSGlueCatalogTable_StorageDescriptor_SerDeInfo_EmptyConfigurationBlock(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_catalog_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlueTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCatalogTableConfigStorageDescriptorSerDeInfoEmptyConfigurationBlock(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlueCatalogTableExists(resourceName),
				),
				// Expect non-empty instead of panic
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSGlueCatalogTable_StorageDescriptor_SerDeInfo_UpdateValues(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_catalog_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlueTableDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccGlueCatalogTableConfigStorageDescriptorSerDeInfo(rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlueCatalogTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.ser_de_info.0.name", "ser_de_name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:  testAccGlueCatalogTableConfigStorageDescriptorSerDeInfoUpdate(rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlueCatalogTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.ser_de_info.0.parameters.param1", "param_val_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_descriptor.0.ser_de_info.0.serialization_library", "org.apache.hadoop.hive.serde2.columnar.ColumnarSerDe"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11784
func TestAccAWSGlueCatalogTable_StorageDescriptor_SkewedInfo_EmptyConfigurationBlock(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_catalog_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlueTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCatalogTableConfigStorageDescriptorSkewedInfoEmptyConfigurationBlock(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlueCatalogTableExists(resourceName),
				),
				// Expect non-empty instead of panic
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSGlueCatalogTable_partitionIndexesSingle(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_catalog_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlueTableDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccGlueCatalogTablePartitionIndexesSingle(rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlueCatalogTableExists(resourceName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("table/%s/%s", rName, rName)),
					resource.TestCheckResourceAttr(resourceName, "partition_index.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "partition_index.0.index_name", rName),
					resource.TestCheckResourceAttr(resourceName, "partition_index.0.index_status", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "partition_index.0.keys.#", "2"),
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

func TestAccAWSGlueCatalogTable_partitionIndexesMultiple(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_catalog_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlueTableDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccGlueCatalogTablePartitionIndexesMultiple(rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlueCatalogTableExists(resourceName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("table/%s/%s", rName, rName)),
					resource.TestCheckResourceAttr(resourceName, "partition_index.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "partition_index.0.index_name", rName),
					resource.TestCheckResourceAttr(resourceName, "partition_index.0.index_status", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "partition_index.0.keys.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "partition_index.1.index_name", fmt.Sprintf("%s-2", rName)),
					resource.TestCheckResourceAttr(resourceName, "partition_index.1.index_status", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "partition_index.1.keys.#", "1"),
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

func TestAccAWSGlueCatalogTable_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_catalog_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlueTableDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccGlueCatalogTable_basic(rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlueCatalogTableExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsGlueCatalogTable(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccGlueCatalogTable_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name
}
`, rName)
}

func testAccGlueCatalogTable_full(rName, desc string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name               = %[1]q
  database_name      = aws_glue_catalog_database.test.name
  description        = %[2]q
  owner              = "my_owner"
  retention          = 1
  table_type         = "VIRTUAL_VIEW"
  view_expanded_text = "view_expanded_text_1"
  view_original_text = "view_original_text_1"

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
    name    = "my_column_1"
    type    = "int"
    comment = "my_column_1_comment"
  }

  partition_keys {
    name    = "my_column_2"
    type    = "string"
    comment = "my_column_2_comment"
  }

  parameters = {
    param1 = "param1_val"
  }
}
`, rName, desc)
}

func testAccGlueCatalogTable_full_replacedValues(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name               = %[1]q
  database_name      = aws_glue_catalog_database.test.name
  description        = "A test table from terraform2"
  owner              = "my_owner2"
  retention          = 2
  table_type         = "EXTERNAL_TABLE"
  view_expanded_text = "view_expanded_text_12"
  view_original_text = "view_original_text_12"

  storage_descriptor {
    bucket_columns = [
      "bucket_column_12",
      "bucket_column_2",
    ]

    compressed                = true
    input_format              = "TextInputFormat"
    location                  = "my_location2"
    number_of_buckets         = 12
    output_format             = "TextInputFormat"
    stored_as_sub_directories = true

    parameters = {
      param12 = "param1_val2"
    }

    columns {
      name    = "my_column_12"
      type    = "date"
      comment = "my_column1_comment2"
    }

    columns {
      name    = "my_column_22"
      type    = "timestamp"
      comment = "my_column2_comment2"
    }

    ser_de_info {
      name = "ser_de_name2"

      parameters = {
        param2 = "param_val_12"
      }

      serialization_library = "org.apache.hadoop.hive.serde2.columnar.ColumnarSerDe2"
    }

    skewed_info {
      skewed_column_names = [
        "my_column_12",
      ]

      skewed_column_value_location_maps = {
        my_column_12 = "my_column_1_val_loc_map2"
      }

      skewed_column_values = [
        "skewed_val_12",
        "skewed_val_2",
      ]
    }

    sort_columns {
      column     = "my_column_12"
      sort_order = 0
    }
  }

  partition_keys {
    name    = "my_column_12"
    type    = "date"
    comment = "my_column_1_comment2"
  }

  partition_keys {
    name    = "my_column_22"
    type    = "timestamp"
    comment = "my_column_2_comment2"
  }

  parameters = {
    param2 = "param1_val2"
  }
}
`, rName)
}

func testAccGlueCatalogTableConfigStorageDescriptorEmptyConfigurationBlock(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q

  storage_descriptor {}
}
`, rName)
}

func testAccGlueCatalogTableConfigStorageDescriptorSerDeInfoEmptyConfigurationBlock(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q

  storage_descriptor {
    ser_de_info {}
  }
}
`, rName)
}

func testAccGlueCatalogTableConfigStorageDescriptorSerDeInfo(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q

  storage_descriptor {
    ser_de_info {
      name = "ser_de_name"
    }
  }
}
`, rName)
}

func testAccGlueCatalogTableConfigStorageDescriptorSerDeInfoUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q

  storage_descriptor {
    ser_de_info {
      parameters = {
        param1 = "param_val_1"
      }
      serialization_library = "org.apache.hadoop.hive.serde2.columnar.ColumnarSerDe"
    }
  }
}
`, rName)
}

func testAccGlueCatalogTableConfigStorageDescriptorSkewedInfoEmptyConfigurationBlock(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q

  storage_descriptor {
    skewed_info {
      skewed_column_names = []

      skewed_column_value_location_maps = {}
      skewed_column_values              = []
    }
  }
}
`, rName)
}

func testAccGlueCatalogTableColumnParameters(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name               = %[1]q
  database_name      = aws_glue_catalog_database.test.name
  owner              = "my_owner"
  retention          = 1
  table_type         = "VIRTUAL_VIEW"
  view_expanded_text = "view_expanded_text_1"
  view_original_text = "view_original_text_1"

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

      parameters = {
        param2 = "param2_val"
      }
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
    name    = "my_column_1"
    type    = "int"
    comment = "my_column_1_comment"
  }

  parameters = {
    param1 = "param1_val"
  }
}
`, rName)
}

func testAccCheckGlueTableDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).glueconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_glue_catalog_table" {
			continue
		}

		catalogId, dbName, resourceName, err := readAwsGlueTableID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &glue.GetTableInput{
			DatabaseName: aws.String(dbName),
			CatalogId:    aws.String(catalogId),
			Name:         aws.String(resourceName),
		}
		if _, err := conn.GetTable(input); err != nil {
			//Verify the error is what we want
			if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
				continue
			}

			return err
		}
		return fmt.Errorf("still exists")
	}
	return nil
}

func testAccCheckGlueCatalogTableExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		catalogId, dbName, resourceName, err := readAwsGlueTableID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).glueconn
		out, err := conn.GetTable(&glue.GetTableInput{
			CatalogId:    aws.String(catalogId),
			DatabaseName: aws.String(dbName),
			Name:         aws.String(resourceName),
		})

		if err != nil {
			return err
		}

		if out.Table == nil {
			return fmt.Errorf("No Glue Table Found")
		}

		if aws.StringValue(out.Table.Name) != resourceName {
			return fmt.Errorf("Glue Table Mismatch - existing: %q, state: %q",
				aws.StringValue(out.Table.Name), resourceName)
		}

		return nil
	}
}

func testAccGlueCatalogTablePartitionIndexesSingle(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name               = %[1]q
  database_name      = aws_glue_catalog_database.test.name
  owner              = "my_owner"
  retention          = 1
  table_type         = "VIRTUAL_VIEW"
  view_expanded_text = "view_expanded_text_1"
  view_original_text = "view_original_text_1"

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
    name    = "my_column_1"
    type    = "int"
    comment = "my_column_1_comment"
  }

  partition_keys {
    name    = "my_column_2"
    type    = "string"
    comment = "my_column_2_comment"
  }

  parameters = {
    param1 = "param1_val"
  }

  partition_index {
    index_name = %[1]q
    keys       = ["my_column_1", "my_column_2"]
  }
}
`, rName)
}

func testAccGlueCatalogTablePartitionIndexesMultiple(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name               = %[1]q
  database_name      = aws_glue_catalog_database.test.name
  owner              = "my_owner"
  retention          = 1
  table_type         = "VIRTUAL_VIEW"
  view_expanded_text = "view_expanded_text_1"
  view_original_text = "view_original_text_1"

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
    name    = "my_column_1"
    type    = "int"
    comment = "my_column_1_comment"
  }

  partition_keys {
    name    = "my_column_2"
    type    = "string"
    comment = "my_column_2_comment"
  }

  parameters = {
    param1 = "param1_val"
  }

  partition_index {
    index_name = %[1]q
    keys       = ["my_column_1", "my_column_2"]
  }

  partition_index {
    index_name = "%[1]s-2"
    keys       = ["my_column_1"]
  }
}
`, rName)
}
