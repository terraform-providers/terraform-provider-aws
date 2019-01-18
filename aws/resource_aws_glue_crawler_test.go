package aws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_glue_crawler", &resource.Sweeper{
		Name: "aws_glue_crawler",
		F:    testSweepGlueCrawlers,
	})
}

func testSweepGlueCrawlers(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).glueconn

	input := &glue.GetCrawlersInput{}
	err = conn.GetCrawlersPages(input, func(page *glue.GetCrawlersOutput, lastPage bool) bool {
		if len(page.Crawlers) == 0 {
			log.Printf("[INFO] No Glue Crawlers to sweep")
			return false
		}
		for _, crawler := range page.Crawlers {
			name := aws.StringValue(crawler.Name)
			if !strings.HasPrefix(name, "tf-acc-test-") {
				log.Printf("[INFO] Skipping Glue Crawler: %s", name)
				continue
			}

			log.Printf("[INFO] Deleting Glue Crawler: %s", name)
			_, err := conn.DeleteCrawler(&glue.DeleteCrawlerInput{
				Name: aws.String(name),
			})
			if err != nil {
				log.Printf("[ERROR] Failed to delete Glue Crawler %s: %s", name, err)
			}
		}
		return !lastPage
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Glue Crawler sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Glue Crawlers: %s", err)
	}

	return nil
}

func TestAccAWSGlueCrawler_DynamodbTarget(t *testing.T) {
	var crawler glue.Crawler
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_DynamodbTarget(rName, "table1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "classifiers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "configuration", ""),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.path", "table1"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "role", rName),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "schedule", ""),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.delete_behavior", "DEPRECATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.update_behavior", "UPDATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "table_prefix", ""),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_DynamodbTarget(rName, "table2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "classifiers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "configuration", ""),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.0.path", "table2"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "role", rName),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "schedule", ""),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.delete_behavior", "DEPRECATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.update_behavior", "UPDATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "table_prefix", ""),
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

func TestAccAWSGlueCrawler_JdbcTarget(t *testing.T) {
	var crawler glue.Crawler
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_JdbcTarget(rName, "database-name/%"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "classifiers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "configuration", ""),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.exclusions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.path", "database-name/%"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "role", rName),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "schedule", ""),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.delete_behavior", "DEPRECATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.update_behavior", "UPDATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "table_prefix", ""),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_JdbcTarget(rName, "database-name/table-name"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "classifiers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "configuration", ""),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.exclusions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.path", "database-name/table-name"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "role", rName),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "schedule", ""),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.delete_behavior", "DEPRECATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.update_behavior", "UPDATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "table_prefix", ""),
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

func TestAccAWSGlueCrawler_JdbcTarget_Exclusions(t *testing.T) {
	var crawler glue.Crawler
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_JdbcTarget_Exclusions(rName, "exclusion1", "exclusion2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.exclusions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.exclusions.0", "exclusion1"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.exclusions.1", "exclusion2"),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_JdbcTarget_Exclusions(rName, "exclusion1", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.exclusions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.exclusions.0", "exclusion1"),
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

func TestAccAWSGlueCrawler_JdbcTarget_Multiple(t *testing.T) {
	var crawler glue.Crawler
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_JdbcTarget_Multiple(rName, "database-name/table1", "database-name/table2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.exclusions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.path", "database-name/table1"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.1.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.1.exclusions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.1.path", "database-name/table2"),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_JdbcTarget(rName, "database-name/table1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.exclusions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.path", "database-name/table1"),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_JdbcTarget_Multiple(rName, "database-name/table1", "database-name/table2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.exclusions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.0.path", "database-name/table1"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.1.connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.1.exclusions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.1.path", "database-name/table2"),
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

func TestAccAWSGlueCrawler_S3Target(t *testing.T) {
	var crawler glue.Crawler
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_S3Target(rName, "s3://bucket1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "classifiers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "configuration", ""),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "role", rName),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.exclusions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.path", "s3://bucket1"),
					resource.TestCheckResourceAttr(resourceName, "schedule", ""),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.delete_behavior", "DEPRECATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.update_behavior", "UPDATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "table_prefix", ""),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_S3Target(rName, "s3://bucket2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "classifiers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "configuration", ""),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_target.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "jdbc_target.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "role", rName),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.exclusions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.path", "s3://bucket2"),
					resource.TestCheckResourceAttr(resourceName, "schedule", ""),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.delete_behavior", "DEPRECATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.update_behavior", "UPDATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(resourceName, "table_prefix", ""),
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

func TestAccAWSGlueCrawler_S3Target_Exclusions(t *testing.T) {
	var crawler glue.Crawler
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_S3Target_Exclusions(rName, "exclusion1", "exclusion2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.exclusions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.exclusions.0", "exclusion1"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.exclusions.1", "exclusion2"),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_S3Target_Exclusions(rName, "exclusion1", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.exclusions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.exclusions.0", "exclusion1"),
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

func TestAccAWSGlueCrawler_S3Target_Multiple(t *testing.T) {
	var crawler glue.Crawler
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_S3Target_Multiple(rName, "s3://bucket1", "s3://bucket2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.exclusions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.path", "s3://bucket1"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.1.exclusions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.1.path", "s3://bucket2"),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_S3Target(rName, "s3://bucket1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.exclusions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.path", "s3://bucket1"),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_S3Target_Multiple(rName, "s3://bucket1", "s3://bucket2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "s3_target.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.exclusions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.0.path", "s3://bucket1"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.1.exclusions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3_target.1.path", "s3://bucket2"),
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

func TestAccAWSGlueCrawler_recreates(t *testing.T) {
	var crawler glue.Crawler
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_S3Target(rName, "s3://bucket1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
				),
			},
			{
				// Simulate deleting the crawler outside Terraform
				PreConfig: func() {
					conn := testAccProvider.Meta().(*AWSClient).glueconn
					input := &glue.DeleteCrawlerInput{
						Name: aws.String(rName),
					}
					_, err := conn.DeleteCrawler(input)
					if err != nil {
						t.Fatalf("error deleting Glue Crawler: %s", err)
					}
				},
				Config:             testAccGlueCrawlerConfig_S3Target(rName, "s3://bucket1"),
				ExpectNonEmptyPlan: true,
				PlanOnly:           true,
			},
		},
	})
}

func TestAccAWSGlueCrawler_Classifiers(t *testing.T) {
	var crawler glue.Crawler
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_Classifiers_Single(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "classifiers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "classifiers.0", rName+"1"),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_Classifiers_Multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "classifiers.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "classifiers.0", rName+"1"),
					resource.TestCheckResourceAttr(resourceName, "classifiers.1", rName+"2"),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_Classifiers_Single(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "classifiers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "classifiers.0", rName+"1"),
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

func TestAccAWSGlueCrawler_Configuration(t *testing.T) {
	var crawler glue.Crawler
	configuration1 := `{"Version": 1.0, "CrawlerOutput": {"Tables": { "AddOrUpdateBehavior": "MergeNewColumns" }}}`
	configuration2 := `{"Version": 1.0, "CrawlerOutput": {"Partitions": { "AddOrUpdateBehavior": "InheritFromTable" }}}`
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_Configuration(rName, configuration1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					testAccCheckAWSGlueCrawlerConfiguration(&crawler, configuration1),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_Configuration(rName, configuration2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					testAccCheckAWSGlueCrawlerConfiguration(&crawler, configuration2),
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

func TestAccAWSGlueCrawler_Description(t *testing.T) {
	var crawler glue.Crawler
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_Description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_Description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
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

func TestAccAWSGlueCrawler_Role_ARN_NoPath(t *testing.T) {
	var crawler glue.Crawler
	iamRoleResourceName := "aws_iam_role.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_Role_ARN_NoPath(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttrPair(resourceName, "role", iamRoleResourceName, "name"),
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

func TestAccAWSGlueCrawler_Role_ARN_Path(t *testing.T) {
	var crawler glue.Crawler
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_Role_ARN_Path(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "role", fmt.Sprintf("path/%s", rName)),
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

func TestAccAWSGlueCrawler_Role_Name_Path(t *testing.T) {
	var crawler glue.Crawler
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_Role_Name_Path(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "role", fmt.Sprintf("path/%s", rName)),
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

func TestAccAWSGlueCrawler_Schedule(t *testing.T) {
	var crawler glue.Crawler
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_Schedule(rName, "cron(0 1 * * ? *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(0 1 * * ? *)"),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_Schedule(rName, "cron(0 2 * * ? *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(0 2 * * ? *)"),
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

func TestAccAWSGlueCrawler_SchemaChangePolicy(t *testing.T) {
	var crawler glue.Crawler
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_SchemaChangePolicy(rName, glue.DeleteBehaviorDeleteFromDatabase, glue.UpdateBehaviorUpdateInDatabase),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.delete_behavior", glue.DeleteBehaviorDeleteFromDatabase),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.update_behavior", glue.UpdateBehaviorUpdateInDatabase),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_SchemaChangePolicy(rName, glue.DeleteBehaviorLog, glue.UpdateBehaviorLog),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.delete_behavior", glue.DeleteBehaviorLog),
					resource.TestCheckResourceAttr(resourceName, "schema_change_policy.0.update_behavior", glue.UpdateBehaviorLog),
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

func TestAccAWSGlueCrawler_TablePrefix(t *testing.T) {
	var crawler glue.Crawler
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_TablePrefix(rName, "prefix1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "table_prefix", "prefix1"),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_TablePrefix(rName, "prefix2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "table_prefix", "prefix2"),
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

func TestAccAWSGlueCrawler_SecurityConfiguration(t *testing.T) {
	var crawler glue.Crawler
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_crawler.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfig_SecurityConfiguration(rName, "security_configuration1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "security_configuration", "security_configuration1"),
				),
			},
			{
				Config: testAccGlueCrawlerConfig_SecurityConfiguration(rName, "security_configuration2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueCrawlerExists(resourceName, &crawler),
					resource.TestCheckResourceAttr(resourceName, "security_configuration", "security_configuration2"),
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

func testAccCheckAWSGlueCrawlerExists(resourceName string, crawler *glue.Crawler) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		glueConn := testAccProvider.Meta().(*AWSClient).glueconn
		out, err := glueConn.GetCrawler(&glue.GetCrawlerInput{
			Name: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if out.Crawler == nil {
			return fmt.Errorf("no Glue Crawler found")
		}

		*crawler = *out.Crawler

		return nil
	}
}

func testAccCheckAWSGlueCrawlerDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_glue_crawler" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).glueconn
		output, err := conn.GetCrawler(&glue.GetCrawlerInput{
			Name: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
				return nil
			}
			return err
		}

		crawler := output.Crawler
		if crawler != nil && aws.StringValue(crawler.Name) == rs.Primary.ID {
			return fmt.Errorf("Glue Crawler %s still exists", rs.Primary.ID)
		}

		return nil
	}

	return nil
}

func testAccCheckAWSGlueCrawlerConfiguration(crawler *glue.Crawler, acctestJSON string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiJSON := aws.StringValue(crawler.Configuration)
		apiJSONBuffer := bytes.NewBufferString("")
		if err := json.Compact(apiJSONBuffer, []byte(apiJSON)); err != nil {
			return fmt.Errorf("unable to compact API configuration JSON: %s", err)
		}

		acctestJSONBuffer := bytes.NewBufferString("")
		if err := json.Compact(acctestJSONBuffer, []byte(acctestJSON)); err != nil {
			return fmt.Errorf("unable to compact acceptance test configuration JSON: %s", err)
		}

		if !jsonBytesEqual(apiJSONBuffer.Bytes(), acctestJSONBuffer.Bytes()) {
			return fmt.Errorf("expected configuration JSON to match %v, received JSON: %v", acctestJSON, apiJSON)
		}
		return nil
	}
}

func testAccGlueCrawlerConfig_Base(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %q
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "glue.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test-AWSGlueServiceRole" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSGlueServiceRole"
  role       = "${aws_iam_role.test.name}"
}
`, rName)
}

func testAccGlueCrawlerConfig_Classifiers_Single(rName string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %q
}

resource "aws_glue_classifier" "test1" {
  name = %q

  grok_classifier {
    classification = "example"
    grok_pattern   = "example"
  }
}

resource "aws_glue_classifier" "test2" {
  name = %q

  grok_classifier {
    classification = "example"
    grok_pattern   = "example"
  }
}

resource "aws_glue_crawler" "test" {
  depends_on = ["aws_iam_role_policy_attachment.test-AWSGlueServiceRole"]

  classifiers   = ["${aws_glue_classifier.test1.id}"]
  name          = %q
  database_name = "${aws_glue_catalog_database.test.name}"
  role          = "${aws_iam_role.test.name}"

  s3_target {
    path = "s3://bucket-name"
  }
}
`, rName, rName+"1", rName+"2", rName)
}

func testAccGlueCrawlerConfig_Classifiers_Multiple(rName string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %q
}

resource "aws_glue_classifier" "test1" {
  name = %q

  grok_classifier {
    classification = "example"
    grok_pattern   = "example"
  }
}

resource "aws_glue_classifier" "test2" {
  name = %q

  grok_classifier {
    classification = "example"
    grok_pattern   = "example"
  }
}

resource "aws_glue_crawler" "test" {
  depends_on = ["aws_iam_role_policy_attachment.test-AWSGlueServiceRole"]

  classifiers   = ["${aws_glue_classifier.test1.id}", "${aws_glue_classifier.test2.id}"]
  name          = %q
  database_name = "${aws_glue_catalog_database.test.name}"
  role          = "${aws_iam_role.test.name}"

  s3_target {
    path = "s3://bucket-name"
  }
}
`, rName, rName+"1", rName+"2", rName)
}

func testAccGlueCrawlerConfig_Configuration(rName, configuration string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %q
}

resource "aws_glue_crawler" "test" {
  depends_on = ["aws_iam_role_policy_attachment.test-AWSGlueServiceRole"]

  configuration = %s
  database_name = "${aws_glue_catalog_database.test.name}"
  name          = %q
  role          = "${aws_iam_role.test.name}"

  s3_target {
    path = "s3://bucket-name"
  }
}
`, rName, strconv.Quote(configuration), rName)
}

func testAccGlueCrawlerConfig_Description(rName, description string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %q
}

resource "aws_glue_crawler" "test" {
  depends_on = ["aws_iam_role_policy_attachment.test-AWSGlueServiceRole"]

  database_name = "${aws_glue_catalog_database.test.name}"
  description   = %q
  name          = %q
  role          = "${aws_iam_role.test.name}"

  s3_target {
    path = "s3://bucket-name"
  }
}
`, rName, description, rName)
}

func testAccGlueCrawlerConfig_DynamodbTarget(rName, path string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %q
}

resource "aws_glue_crawler" "test" {
  depends_on = ["aws_iam_role_policy_attachment.test-AWSGlueServiceRole"]

  database_name = "${aws_glue_catalog_database.test.name}"
  name          = %q
  role          = "${aws_iam_role.test.name}"

  dynamodb_target {
    path = %q
  }
}
`, rName, rName, path)
}

func testAccGlueCrawlerConfig_JdbcTarget(rName, path string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %q
}

resource "aws_glue_connection" "test" {
  name = %q

  connection_properties {
    JDBC_CONNECTION_URL = "jdbc:mysql://terraformacctesting.com/testdatabase"
    PASSWORD            = "testpassword"
    USERNAME            = "testusername"
  }
}

resource "aws_glue_crawler" "test" {
  depends_on = ["aws_iam_role_policy_attachment.test-AWSGlueServiceRole"]

  database_name = "${aws_glue_catalog_database.test.name}"
  name          = %q
  role          = "${aws_iam_role.test.name}"

  jdbc_target {
    connection_name = "${aws_glue_connection.test.name}"
    path            = %q
  }
}
`, rName, rName, rName, path)
}

func testAccGlueCrawlerConfig_JdbcTarget_Exclusions(rName, exclusion1, exclusion2 string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %q
}

resource "aws_glue_connection" "test" {
  name = %q

  connection_properties {
    JDBC_CONNECTION_URL = "jdbc:mysql://terraformacctesting.com/testdatabase"
    PASSWORD            = "testpassword"
    USERNAME            = "testusername"
  }
}

resource "aws_glue_crawler" "test" {
  depends_on = ["aws_iam_role_policy_attachment.test-AWSGlueServiceRole"]

  database_name = "${aws_glue_catalog_database.test.name}"
  name          = %q
  role          = "${aws_iam_role.test.name}"

  jdbc_target {
    connection_name = "${aws_glue_connection.test.name}"
    exclusions      = ["${compact(list(%q, %q))}"]
    path            = "database-name/table1"
  }
}
`, rName, rName, rName, exclusion1, exclusion2)
}

func testAccGlueCrawlerConfig_JdbcTarget_Multiple(rName, path1, path2 string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %q
}

resource "aws_glue_connection" "test" {
  name = %q

  connection_properties {
    JDBC_CONNECTION_URL = "jdbc:mysql://terraformacctesting.com/testdatabase"
    PASSWORD            = "testpassword"
    USERNAME            = "testusername"
  }
}

resource "aws_glue_crawler" "test" {
  depends_on = ["aws_iam_role_policy_attachment.test-AWSGlueServiceRole"]

  database_name = "${aws_glue_catalog_database.test.name}"
  name          = %q
  role          = "${aws_iam_role.test.name}"

  jdbc_target {
    connection_name = "${aws_glue_connection.test.name}"
    path            = %q
  }

  jdbc_target {
    connection_name = "${aws_glue_connection.test.name}"
    path            = %q
  }
}
`, rName, rName, rName, path1, path2)
}

func testAccGlueCrawlerConfig_Role_ARN_NoPath(rName string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %q
}

resource "aws_glue_crawler" "test" {
  depends_on = ["aws_iam_role_policy_attachment.test-AWSGlueServiceRole"]

  database_name = "${aws_glue_catalog_database.test.name}"
  name          = %q
  role          = "${aws_iam_role.test.arn}"

  s3_target {
    path = "s3://bucket-name"
  }
}
`, rName, rName)
}

func testAccGlueCrawlerConfig_Role_ARN_Path(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %q
  path = "/path/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "glue.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test-AWSGlueServiceRole" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSGlueServiceRole"
  role       = "${aws_iam_role.test.name}"
}

resource "aws_glue_catalog_database" "test" {
  name = %q
}

resource "aws_glue_crawler" "test" {
  depends_on = ["aws_iam_role_policy_attachment.test-AWSGlueServiceRole"]

  database_name = "${aws_glue_catalog_database.test.name}"
  name          = %q
  role          = "${aws_iam_role.test.arn}"

  s3_target {
    path = "s3://bucket-name"
  }
}
`, rName, rName, rName)
}

func testAccGlueCrawlerConfig_Role_Name_Path(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %q
  path = "/path/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "glue.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test-AWSGlueServiceRole" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSGlueServiceRole"
  role       = "${aws_iam_role.test.name}"
}

resource "aws_glue_catalog_database" "test" {
  name = %q
}

resource "aws_glue_crawler" "test" {
  depends_on = ["aws_iam_role_policy_attachment.test-AWSGlueServiceRole"]

  database_name = "${aws_glue_catalog_database.test.name}"
  name          = %q
  role          = "${replace(aws_iam_role.test.path, "/^\\//", "")}${aws_iam_role.test.name}"

  s3_target {
    path = "s3://bucket-name"
  }
}
`, rName, rName, rName)
}

func testAccGlueCrawlerConfig_S3Target(rName, path string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %q
}

resource "aws_glue_crawler" "test" {
  depends_on = ["aws_iam_role_policy_attachment.test-AWSGlueServiceRole"]

  database_name = "${aws_glue_catalog_database.test.name}"
  name          = %q
  role          = "${aws_iam_role.test.name}"

  s3_target {
    path = %q
  }
}
`, rName, rName, path)
}

func testAccGlueCrawlerConfig_S3Target_Exclusions(rName, exclusion1, exclusion2 string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %q
}

resource "aws_glue_crawler" "test" {
  depends_on = ["aws_iam_role_policy_attachment.test-AWSGlueServiceRole"]

  database_name = "${aws_glue_catalog_database.test.name}"
  name          = %q
  role          = "${aws_iam_role.test.name}"

  s3_target {
    exclusions = ["${compact(list(%q, %q))}"]
    path       = "s3://bucket1"
  }
}
`, rName, rName, exclusion1, exclusion2)
}

func testAccGlueCrawlerConfig_S3Target_Multiple(rName, path1, path2 string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %q
}

resource "aws_glue_crawler" "test" {
  depends_on = ["aws_iam_role_policy_attachment.test-AWSGlueServiceRole"]

  database_name = "${aws_glue_catalog_database.test.name}"
  name          = %q
  role          = "${aws_iam_role.test.name}"

  s3_target {
    path = %q
  }

  s3_target {
    path = %q
  }
}
`, rName, rName, path1, path2)
}

func testAccGlueCrawlerConfig_Schedule(rName, schedule string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %q
}

resource "aws_glue_crawler" "test" {
  depends_on = ["aws_iam_role_policy_attachment.test-AWSGlueServiceRole"]

  database_name = "${aws_glue_catalog_database.test.name}"
  name          = %q
  role          = "${aws_iam_role.test.name}"
  schedule      = %q

  s3_target {
    path = "s3://bucket-name"
  }
}
`, rName, rName, schedule)
}

func testAccGlueCrawlerConfig_SchemaChangePolicy(rName, deleteBehavior, updateBehavior string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %q
}

resource "aws_glue_crawler" "test" {
  depends_on = ["aws_iam_role_policy_attachment.test-AWSGlueServiceRole"]

  database_name = "${aws_glue_catalog_database.test.name}"
  name          = %q
  role          = "${aws_iam_role.test.name}"

  s3_target {
    path = "s3://bucket-name"
  }

  schema_change_policy {
    delete_behavior = %q
    update_behavior = %q
  }
}
`, rName, rName, deleteBehavior, updateBehavior)
}

func testAccGlueCrawlerConfig_TablePrefix(rName, tablePrefix string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %q
}

resource "aws_glue_crawler" "test" {
  depends_on = ["aws_iam_role_policy_attachment.test-AWSGlueServiceRole"]

  database_name = "${aws_glue_catalog_database.test.name}"
  name          = %q
  role          = "${aws_iam_role.test.name}"
  table_prefix  = %q

  s3_target {
    path = "s3://bucket-name"
  }
}
`, rName, rName, tablePrefix)
}

func testAccGlueCrawlerConfig_SecurityConfiguration(rName, securityConfiguration string) string {
	return testAccGlueCrawlerConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %q
}

resource "aws_glue_security_configuration" "test" {
  name = %q

  encryption_configuration {
    cloudwatch_encryption {
      cloudwatch_encryption_mode = "DISABLED"
    }

    job_bookmarks_encryption {
      job_bookmarks_encryption_mode = "DISABLED"
    }

    s3_encryption {
      s3_encryption_mode = "DISABLED"
    }
  }
}

resource "aws_glue_crawler" "test" {
  depends_on = ["aws_iam_role_policy_attachment.test-AWSGlueServiceRole"]

  database_name          = "${aws_glue_catalog_database.test.name}"
  name                   = %q
  role                   = "${aws_iam_role.test.name}"
  security_configuration = "${aws_glue_security_configuration.test.name}"

  s3_target {
    path = "s3://bucket-name"
  }
}
`, rName, securityConfiguration, rName)
}
