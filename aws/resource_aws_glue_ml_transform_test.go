package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_glue_ml_transform", &resource.Sweeper{
		Name: "aws_glue_ml_transform",
		F:    testSweepGlueMLTransforms,
	})
}

func testSweepGlueMLTransforms(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).glueconn
	var sweeperErrs *multierror.Error

	input := &glue.GetMLTransformsInput{}
	err = conn.GetMLTransformsPages(input, func(page *glue.GetMLTransformsOutput, lastPage bool) bool {
		if len(page.Transforms) == 0 {
			log.Printf("[INFO] No Glue ML Transforms to sweep")
			return false
		}
		for _, transforms := range page.Transforms {
			id := aws.StringValue(transforms.TransformId)

			log.Printf("[INFO] Deleting Glue ML Transform: %s", id)
			r := resourceAwsGlueMLTransform()
			d := r.Data(nil)
			d.SetId(id)
			err := r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}
		return !lastPage
	})
	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Glue ML Transforms sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Glue ML Transforms: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSGlueMLTransform_basic(t *testing.T) {
	var transform glue.GetMLTransformOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_ml_transform.test"
	roleResourceName := "aws_iam_role.test"
	tableResourceName := "aws_glue_catalog_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueMLTransformDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueMLTransformBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueMLTransformExists(resourceName, &transform),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "glue", regexp.MustCompile(`mlTransform/tfm-.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", roleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "input_record_tables.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "input_record_tables.0.database_name", tableResourceName, "database_name"),
					resource.TestCheckResourceAttrPair(resourceName, "input_record_tables.0.table_name", tableResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.transform_type", "FIND_MATCHES"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.0.primary_key_column_name", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.0.accuracy_cost_trade_off", "0"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.0.precision_recall_trade_off", "0"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.0.enforce_provided_labels", "false"),
					resource.TestCheckResourceAttr(resourceName, "timeout", "2880"),
					resource.TestCheckResourceAttr(resourceName, "schema.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "schema.0.data_type", "int"),
					resource.TestCheckResourceAttr(resourceName, "schema.0.name", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "schema.1.data_type", "string"),
					resource.TestCheckResourceAttr(resourceName, "schema.1.name", "my_column_2"),
					resource.TestCheckResourceAttr(resourceName, "label_count", "0"),
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

func TestAccAWSGlueMLTransform_typeFindMatchesFull(t *testing.T) {
	var transform glue.GetMLTransformOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_ml_transform.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueMLTransformDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueMLTransformTypeFindMatchesFullConfig(rName, true, 0.5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueMLTransformExists(resourceName, &transform),
					resource.TestCheckResourceAttr(resourceName, "parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.transform_type", "FIND_MATCHES"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.0.primary_key_column_name", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.0.accuracy_cost_trade_off", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.0.precision_recall_trade_off", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.0.enforce_provided_labels", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSGlueMLTransformTypeFindMatchesFullConfig(rName, false, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueMLTransformExists(resourceName, &transform),
					resource.TestCheckResourceAttr(resourceName, "parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.transform_type", "FIND_MATCHES"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.0.primary_key_column_name", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.0.accuracy_cost_trade_off", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.0.precision_recall_trade_off", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.0.enforce_provided_labels", "false"),
				),
			},
			{
				Config: testAccAWSGlueMLTransformTypeFindMatchesFullConfig(rName, true, 0.5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueMLTransformExists(resourceName, &transform),
					resource.TestCheckResourceAttr(resourceName, "parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.transform_type", "FIND_MATCHES"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.0.primary_key_column_name", "my_column_1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.0.accuracy_cost_trade_off", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.0.precision_recall_trade_off", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.find_matches_parameters.0.enforce_provided_labels", "true"),
				),
			},
		},
	})
}

func TestAccAWSGlueMLTransform_description(t *testing.T) {
	var transform glue.GetMLTransformOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_ml_transform.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueMLTransformDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueMLTransformConfigDescription(rName, "First Description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueMLTransformExists(resourceName, &transform),
					resource.TestCheckResourceAttr(resourceName, "description", "First Description"),
				),
			},
			{
				Config: testAccAWSGlueMLTransformConfigDescription(rName, "Second Description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueMLTransformExists(resourceName, &transform),
					resource.TestCheckResourceAttr(resourceName, "description", "Second Description"),
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

func TestAccAWSGlueMLTransform_glueVersion(t *testing.T) {
	var transform glue.GetMLTransformOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_ml_transform.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueMLTransformDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueMLTransformConfigGlueVersion(rName, "0.9"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueMLTransformExists(resourceName, &transform),
					resource.TestCheckResourceAttr(resourceName, "glue_version", "0.9"),
				),
			},
			{
				Config: testAccAWSGlueMLTransformConfigGlueVersion(rName, "1.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueMLTransformExists(resourceName, &transform),
					resource.TestCheckResourceAttr(resourceName, "glue_version", "1.0"),
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

func TestAccAWSGlueMLTransform_maxRetries(t *testing.T) {
	var transform glue.GetMLTransformOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_ml_transform.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueMLTransformDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSGlueMLTransformConfigMaxRetries(rName, 11),
				ExpectError: regexp.MustCompile(`expected max_retries to be in the range`),
			},
			{
				Config: testAccAWSGlueMLTransformConfigMaxRetries(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueMLTransformExists(resourceName, &transform),
					resource.TestCheckResourceAttr(resourceName, "max_retries", "0"),
				),
			},
			{
				Config: testAccAWSGlueMLTransformConfigMaxRetries(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueMLTransformExists(resourceName, &transform),
					resource.TestCheckResourceAttr(resourceName, "max_retries", "10"),
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

func TestAccAWSGlueMLTransform_tags(t *testing.T) {
	var transform1, transform2, transform3 glue.GetMLTransformOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_ml_transform.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueMLTransformDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueMLTransformConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueMLTransformExists(resourceName, &transform1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSGlueMLTransformConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueMLTransformExists(resourceName, &transform2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSGlueMLTransformConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueMLTransformExists(resourceName, &transform3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSGlueMLTransform_timeout(t *testing.T) {
	var transform glue.GetMLTransformOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_ml_transform.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueMLTransformDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueMLTransformConfigTimeout(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueMLTransformExists(resourceName, &transform),
					resource.TestCheckResourceAttr(resourceName, "timeout", "1"),
				),
			},
			{
				Config: testAccAWSGlueMLTransformConfigTimeout(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueMLTransformExists(resourceName, &transform),
					resource.TestCheckResourceAttr(resourceName, "timeout", "2"),
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

func TestAccAWSGlueMLTransform_workerType(t *testing.T) {
	var transform glue.GetMLTransformOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_ml_transform.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueMLTransformDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueMLTransformConfigWorkerType(rName, "Standard", 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueMLTransformExists(resourceName, &transform),
					resource.TestCheckResourceAttr(resourceName, "worker_type", "Standard"),
					resource.TestCheckResourceAttr(resourceName, "number_of_workers", "1"),
				),
			},
			{
				Config: testAccAWSGlueMLTransformConfigWorkerType(rName, "G.1X", 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueMLTransformExists(resourceName, &transform),
					resource.TestCheckResourceAttr(resourceName, "worker_type", "G.1X"),
					resource.TestCheckResourceAttr(resourceName, "number_of_workers", "2"),
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

func TestAccAWSGlueMLTransform_maxCapacity(t *testing.T) {
	var transform glue.GetMLTransformOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_ml_transform.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueMLTransformDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueMLTransformConfigMaxCapacity(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueMLTransformExists(resourceName, &transform),
					resource.TestCheckResourceAttr(resourceName, "max_capacity", "10"),
				),
			},
			{
				Config: testAccAWSGlueMLTransformConfigMaxCapacity(rName, 15),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueMLTransformExists(resourceName, &transform),
					resource.TestCheckResourceAttr(resourceName, "max_capacity", "15"),
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

func TestAccAWSGlueMLTransform_disappears(t *testing.T) {
	var transform glue.GetMLTransformOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_ml_transform.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueMLTransformDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueMLTransformBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueMLTransformExists(resourceName, &transform),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsGlueMLTransform(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSGlueMLTransformExists(resourceName string, mlTransform *glue.GetMLTransformOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Glue Job ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).glueconn

		output, err := conn.GetMLTransform(&glue.GetMLTransformInput{
			TransformId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("Glue ML Transform (%s) not found", rs.Primary.ID)
		}

		if aws.StringValue(output.TransformId) == rs.Primary.ID {
			*mlTransform = *output
			return nil
		}

		return fmt.Errorf("Glue ML Transform (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckAWSGlueMLTransformDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_glue_ml_transform" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).glueconn

		output, err := conn.GetMLTransform(&glue.GetMLTransformInput{
			TransformId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
				return nil
			}

		}

		if output != nil && aws.StringValue(output.TransformId) == rs.Primary.ID {
			return fmt.Errorf("Glue ML Transform %s still exists", rs.Primary.ID)
		}

		return err
	}

	return nil
}

func testAccAWSGlueMLTransformConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy" "AWSGlueServiceRole" {
  arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSGlueServiceRole"
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
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
POLICY
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = data.aws_iam_policy.AWSGlueServiceRole.arn
  role       = aws_iam_role.test.name
}

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
}
`, rName)
}

func testAccAWSGlueMLTransformBasicConfig(rName string) string {
	return testAccAWSGlueMLTransformConfigBase(rName) + fmt.Sprintf(`
resource "aws_glue_ml_transform" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  input_record_tables {
    database_name = aws_glue_catalog_table.test.database_name
    table_name    = aws_glue_catalog_table.test.name
  }

  parameters {
    transform_type = "FIND_MATCHES"

    find_matches_parameters {
      primary_key_column_name = "my_column_1"
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccAWSGlueMLTransformTypeFindMatchesFullConfig(rName string, enforce bool, tradeOff float64) string {
	return testAccAWSGlueMLTransformConfigBase(rName) + fmt.Sprintf(`
resource "aws_glue_ml_transform" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  input_record_tables {
    database_name = aws_glue_catalog_table.test.database_name
    table_name    = aws_glue_catalog_table.test.name
  }

  parameters {
    transform_type = "FIND_MATCHES"

    find_matches_parameters {
      primary_key_column_name    = "my_column_1"
      enforce_provided_labels    = %[2]t
      precision_recall_trade_off = %[3]g
      accuracy_cost_trade_off    = %[3]g
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, enforce, tradeOff)
}

func testAccAWSGlueMLTransformConfigDescription(rName, description string) string {
	return testAccAWSGlueMLTransformConfigBase(rName) + fmt.Sprintf(`
resource "aws_glue_ml_transform" "test" {
  name        = %[1]q
  description = %[2]q
  role_arn    = aws_iam_role.test.arn

  input_record_tables {
    database_name = aws_glue_catalog_table.test.database_name
    table_name    = aws_glue_catalog_table.test.name
  }

  parameters {
    transform_type = "FIND_MATCHES"

    find_matches_parameters {
      primary_key_column_name = "my_column_1"
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, description)
}

func testAccAWSGlueMLTransformConfigGlueVersion(rName, glueVersion string) string {
	return testAccAWSGlueMLTransformConfigBase(rName) + fmt.Sprintf(`
resource "aws_glue_ml_transform" "test" {
  name         = %[1]q
  glue_version = %[2]q
  role_arn     = aws_iam_role.test.arn

  input_record_tables {
    database_name = aws_glue_catalog_table.test.database_name
    table_name    = aws_glue_catalog_table.test.name
  }

  parameters {
    transform_type = "FIND_MATCHES"

    find_matches_parameters {
      primary_key_column_name = "my_column_1"
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, glueVersion)
}

func testAccAWSGlueMLTransformConfigMaxRetries(rName string, maxRetries int) string {
	return testAccAWSGlueMLTransformConfigBase(rName) + fmt.Sprintf(`
resource "aws_glue_ml_transform" "test" {
  name        = %[1]q
  max_retries = %[2]d
  role_arn    = aws_iam_role.test.arn

  input_record_tables {
    database_name = aws_glue_catalog_table.test.database_name
    table_name    = aws_glue_catalog_table.test.name
  }

  parameters {
    transform_type = "FIND_MATCHES"

    find_matches_parameters {
      primary_key_column_name = "my_column_1"
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, maxRetries)
}

func testAccAWSGlueMLTransformConfigTags1(rName, tagKey1, tagValue1 string) string {
	return testAccAWSGlueMLTransformConfigBase(rName) + fmt.Sprintf(`
resource "aws_glue_ml_transform" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  input_record_tables {
    database_name = aws_glue_catalog_table.test.database_name
    table_name    = aws_glue_catalog_table.test.name
  }

  parameters {
    transform_type = "FIND_MATCHES"

    find_matches_parameters {
      primary_key_column_name = "my_column_1"
    }
  }

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSGlueMLTransformConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAWSGlueMLTransformConfigBase(rName) + fmt.Sprintf(`
resource "aws_glue_ml_transform" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  input_record_tables {
    database_name = aws_glue_catalog_table.test.database_name
    table_name    = aws_glue_catalog_table.test.name
  }

  parameters {
    transform_type = "FIND_MATCHES"

    find_matches_parameters {
      primary_key_column_name = "my_column_1"
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSGlueMLTransformConfigTimeout(rName string, timeout int) string {
	return testAccAWSGlueMLTransformConfigBase(rName) + fmt.Sprintf(`
resource "aws_glue_ml_transform" "test" {
  name     = %[1]q
  timeout  = %[2]d
  role_arn = aws_iam_role.test.arn

  input_record_tables {
    database_name = aws_glue_catalog_table.test.database_name
    table_name    = aws_glue_catalog_table.test.name
  }

  parameters {
    transform_type = "FIND_MATCHES"

    find_matches_parameters {
      primary_key_column_name = "my_column_1"
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, timeout)
}

func testAccAWSGlueMLTransformConfigWorkerType(rName, workerType string, numOfWorkers int) string {
	return testAccAWSGlueMLTransformConfigBase(rName) + fmt.Sprintf(`
resource "aws_glue_ml_transform" "test" {
  name              = %[1]q
  worker_type       = %[2]q
  number_of_workers = %[3]d
  role_arn          = aws_iam_role.test.arn

  input_record_tables {
    database_name = aws_glue_catalog_table.test.database_name
    table_name    = aws_glue_catalog_table.test.name
  }

  parameters {
    transform_type = "FIND_MATCHES"

    find_matches_parameters {
      primary_key_column_name = "my_column_1"
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, workerType, numOfWorkers)
}

func testAccAWSGlueMLTransformConfigMaxCapacity(rName string, maxCapacity float64) string {
	return testAccAWSGlueMLTransformConfigBase(rName) + fmt.Sprintf(`
resource "aws_glue_ml_transform" "test" {
  name         = %[1]q
  max_capacity = %[2]g
  role_arn     = aws_iam_role.test.arn

  input_record_tables {
    database_name = aws_glue_catalog_table.test.database_name
    table_name    = aws_glue_catalog_table.test.name
  }

  parameters {
    transform_type = "FIND_MATCHES"

    find_matches_parameters {
      primary_key_column_name = "my_column_1"
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, maxCapacity)
}
