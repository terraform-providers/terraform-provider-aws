package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker/finder"
)

func init() {
	resource.AddTestSweepers("aws_sagemaker_feature_group", &resource.Sweeper{
		Name: "aws_sagemaker_feature_group",
		F:    testSweepSagemakerFeatureGroups,
	})
}

func testSweepSagemakerFeatureGroups(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).sagemakerconn

	err = conn.ListFeatureGroupsPages(&sagemaker.ListFeatureGroupsInput{}, func(page *sagemaker.ListFeatureGroupsOutput, lastPage bool) bool {
		for _, group := range page.FeatureGroupSummaries {
			name := aws.StringValue(group.FeatureGroupName)

			input := &sagemaker.DeleteFeatureGroupInput{
				FeatureGroupName: group.FeatureGroupName,
			}

			log.Printf("[INFO] Deleting SageMaker Feature Group: %s", name)
			if _, err := conn.DeleteFeatureGroup(input); err != nil {
				log.Printf("[ERROR] Error deleting SageMaker Feature Group (%s): %s", name, err)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker Feature Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error retrieving SageMaker Feature Groups: %w", err)
	}

	return nil
}

func TestAccAWSSagemakerFeatureGroup_basic(t *testing.T) {
	var featureGroup sagemaker.DescribeFeatureGroupOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_feature_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerFeatureGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerFeatureGroupBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerFeatureGroupExists(resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "event_time_feature_name", rName),
					resource.TestCheckResourceAttr(resourceName, "record_identifier_feature_name", rName),
					resource.TestCheckResourceAttr(resourceName, "online_store_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "online_store_config.0.enable_online_store", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.0.feature_name", rName),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.0.feature_type", "String"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("feature-group/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.#", "0"),
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

func TestAccAWSSagemakerFeatureGroup_description(t *testing.T) {
	var featureGroup sagemaker.DescribeFeatureGroupOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_feature_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerFeatureGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerFeatureGroupDescriptionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerFeatureGroupExists(resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
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

func TestAccAWSSagemakerFeatureGroup_tags(t *testing.T) {
	var featureGroup sagemaker.DescribeFeatureGroupOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_feature_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerFeatureGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerFeatureGroupTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerFeatureGroupExists(resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
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
				Config: testAccAWSSagemakerFeatureGroupTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerFeatureGroupExists(resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSSagemakerFeatureGroupTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerFeatureGroupExists(resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSSagemakerFeatureGroup_multipleFeatures(t *testing.T) {
	var featureGroup sagemaker.DescribeFeatureGroupOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_feature_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerFeatureGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerFeatureGroupConfigMultiFeature(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerFeatureGroupExists(resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.0.feature_name", rName),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.0.feature_type", "String"),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.1.feature_name", fmt.Sprintf("%s-2", rName)),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.1.feature_type", "Integral"),
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

func TestAccAWSSagemakerFeatureGroup_onlineConfigSecurityConfig(t *testing.T) {
	var featureGroup sagemaker.DescribeFeatureGroupOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_feature_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerFeatureGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerFeatureGroupOnlineSecurityConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerFeatureGroupExists(resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "online_store_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "online_store_config.0.enable_online_store", "true"),
					resource.TestCheckResourceAttr(resourceName, "online_store_config.0.security_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "online_store_config.0.security_config.0.kms_key_id", "aws_kms_key.test", "arn"),
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

func TestAccAWSSagemakerFeatureGroup_offlineConfig_basic(t *testing.T) {
	var featureGroup sagemaker.DescribeFeatureGroupOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_feature_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerFeatureGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerFeatureGroupOfflineBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerFeatureGroupExists(resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.disable_glue_table_creation", "true"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.s3_storage_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.s3_storage_config.0.s3_uri", fmt.Sprintf("s3://%s/prefix/", rName)),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.data_catalog_config.#", "0"),
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

func TestAccAWSSagemakerFeatureGroup_offlineConfig_createCatalog(t *testing.T) {
	var featureGroup sagemaker.DescribeFeatureGroupOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_feature_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerFeatureGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerFeatureGroupOfflineCreateGlueCatalogConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerFeatureGroupExists(resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.disable_glue_table_creation", "false"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.s3_storage_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.s3_storage_config.0.s3_uri", fmt.Sprintf("s3://%s/prefix/", rName)),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.data_catalog_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.data_catalog_config.0.catalog", "AwsDataCatalog"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.data_catalog_config.0.database", "sagemaker_featurestore"),
					resource.TestMatchResourceAttr(resourceName, "offline_store_config.0.data_catalog_config.0.table_name", regexp.MustCompile(fmt.Sprintf("^%s-", rName))),
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

func TestAccAWSSagemakerFeatureGroup_offlineConfig_providedCatalog(t *testing.T) {
	var featureGroup sagemaker.DescribeFeatureGroupOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_feature_group.test"
	glueTableResourceName := "aws_glue_catalog_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerFeatureGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerFeatureGroupOfflineCreateGlueCatalogConfigProvidedCatalog(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerFeatureGroupExists(resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.disable_glue_table_creation", "true"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.s3_storage_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.s3_storage_config.0.s3_uri", fmt.Sprintf("s3://%s/prefix/", rName)),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.data_catalog_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "offline_store_config.0.data_catalog_config.0.catalog", glueTableResourceName, "catalog_id"),
					resource.TestCheckResourceAttrPair(resourceName, "offline_store_config.0.data_catalog_config.0.database", glueTableResourceName, "database_name"),
					resource.TestCheckResourceAttrPair(resourceName, "offline_store_config.0.data_catalog_config.0.table_name", glueTableResourceName, "name"),
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

func TestAccAWSSagemakerFeatureGroup_disappears(t *testing.T) {
	var featureGroup sagemaker.DescribeFeatureGroupOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_feature_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerFeatureGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerFeatureGroupBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerFeatureGroupExists(resourceName, &featureGroup),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSagemakerFeatureGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSSagemakerFeatureGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_feature_group" {
			continue
		}

		codeRepository, err := finder.FeatureGroupByName(conn, rs.Primary.ID)
		if err != nil {
			return nil
		}

		if aws.StringValue(codeRepository.FeatureGroupName) == rs.Primary.ID {
			return fmt.Errorf("Sagemaker Feature Group %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSSagemakerFeatureGroupExists(n string, codeRepo *sagemaker.DescribeFeatureGroupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker Feature Group ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn
		resp, err := finder.FeatureGroupByName(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*codeRepo = *resp

		return nil
	}
}

func testAccAWSSagemakerFeatureGroupBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}
`, rName)
}

func testAccAWSSagemakerFeatureGroupOfflineBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  acl           = "private"
  force_destroy = true
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test.arn
}

resource "aws_iam_policy" "test" {
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [{
      "Effect" : "Allow",
      "Resource" : [
        "${aws_s3_bucket.test.arn}",
        "${aws_s3_bucket.test.arn}/*"
      ],
      "Action" : [
        "s3:*"
      ]
    }]
  })
}
`, rName)
}

func testAccAWSSagemakerFeatureGroupBasicConfig(rName string) string {
	return testAccAWSSagemakerFeatureGroupBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_feature_group" "test" {
  feature_group_name             = %[1]q
  record_identifier_feature_name = %[1]q
  event_time_feature_name        = %[1]q
  role_arn                       = aws_iam_role.test.arn

  feature_definition {
    feature_name = %[1]q
    feature_type = "String"
  }

  online_store_config {
    enable_online_store = true
  }
}
`, rName)
}

func testAccAWSSagemakerFeatureGroupDescriptionConfig(rName string) string {
	return testAccAWSSagemakerFeatureGroupBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_feature_group" "test" {
  feature_group_name             = %[1]q
  record_identifier_feature_name = %[1]q
  event_time_feature_name        = %[1]q
  role_arn                       = aws_iam_role.test.arn
  description                    = %[1]q

  feature_definition {
    feature_name = %[1]q
    feature_type = "String"
  }

  online_store_config {
    enable_online_store = true
  }
}
`, rName)
}

func testAccAWSSagemakerFeatureGroupConfigMultiFeature(rName string) string {
	return testAccAWSSagemakerFeatureGroupBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_feature_group" "test" {
  feature_group_name             = %[1]q
  record_identifier_feature_name = %[1]q
  event_time_feature_name        = %[1]q
  role_arn                       = aws_iam_role.test.arn

  feature_definition {
    feature_name = %[1]q
    feature_type = "String"
  }

  feature_definition {
    feature_name = "%[1]s-2"
    feature_type = "Integral"
  }

  online_store_config {
    enable_online_store = true
  }
}
`, rName)
}

func testAccAWSSagemakerFeatureGroupOnlineSecurityConfig(rName string) string {
	return testAccAWSSagemakerFeatureGroupBaseConfig(rName) + fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_sagemaker_feature_group" "test" {
  feature_group_name             = %[1]q
  record_identifier_feature_name = %[1]q
  event_time_feature_name        = %[1]q
  role_arn                       = aws_iam_role.test.arn

  feature_definition {
    feature_name = %[1]q
    feature_type = "String"
  }

  online_store_config {
    enable_online_store = true

    security_config {
      kms_key_id = aws_kms_key.test.arn
    }
  }
}
`, rName)
}

func testAccAWSSagemakerFeatureGroupOfflineBasicConfig(rName string) string {
	return testAccAWSSagemakerFeatureGroupBaseConfig(rName) +
		testAccAWSSagemakerFeatureGroupOfflineBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_feature_group" "test" {
  feature_group_name             = %[1]q
  record_identifier_feature_name = %[1]q
  event_time_feature_name        = %[1]q
  role_arn                       = aws_iam_role.test.arn

  feature_definition {
    feature_name = %[1]q
    feature_type = "String"
  }

  offline_store_config {
    disable_glue_table_creation = true

    s3_storage_config {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/prefix/"
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccAWSSagemakerFeatureGroupOfflineCreateGlueCatalogConfig(rName string) string {
	return testAccAWSSagemakerFeatureGroupBaseConfig(rName) +
		testAccAWSSagemakerFeatureGroupOfflineBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_feature_group" "test" {
  feature_group_name             = %[1]q
  record_identifier_feature_name = %[1]q
  event_time_feature_name        = %[1]q
  role_arn                       = aws_iam_role.test.arn

  feature_definition {
    feature_name = %[1]q
    feature_type = "String"
  }

  offline_store_config {
    disable_glue_table_creation = false

    s3_storage_config {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/prefix/"
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccAWSSagemakerFeatureGroupOfflineCreateGlueCatalogConfigProvidedCatalog(rName string) string {
	return testAccAWSSagemakerFeatureGroupBaseConfig(rName) +
		testAccAWSSagemakerFeatureGroupOfflineBaseConfig(rName) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name
}

resource "aws_sagemaker_feature_group" "test" {
  feature_group_name             = %[1]q
  record_identifier_feature_name = %[1]q
  event_time_feature_name        = %[1]q
  role_arn                       = aws_iam_role.test.arn

  feature_definition {
    feature_name = %[1]q
    feature_type = "String"
  }

  offline_store_config {
    disable_glue_table_creation = true

    s3_storage_config {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/prefix/"
    }

    data_catalog_config {
      catalog    = aws_glue_catalog_table.test.catalog_id
      database   = aws_glue_catalog_table.test.database_name
      table_name = aws_glue_catalog_table.test.name
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccAWSSagemakerFeatureGroupTags1(rName, tag1Key, tag1Value string) string {
	return testAccAWSSagemakerFeatureGroupBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_feature_group" "test" {
  feature_group_name             = %[1]q
  record_identifier_feature_name = %[1]q
  event_time_feature_name        = %[1]q
  role_arn                       = aws_iam_role.test.arn

  feature_definition {
    feature_name = %[1]q
    feature_type = "String"
  }

  online_store_config {
    enable_online_store = true
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag1Key, tag1Value)
}

func testAccAWSSagemakerFeatureGroupTags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return testAccAWSSagemakerFeatureGroupBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_feature_group" "test" {
  feature_group_name             = %[1]q
  record_identifier_feature_name = %[1]q
  event_time_feature_name        = %[1]q
  role_arn                       = aws_iam_role.test.arn

  feature_definition {
    feature_name = %[1]q
    feature_type = "String"
  }

  online_store_config {
    enable_online_store = true
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}
