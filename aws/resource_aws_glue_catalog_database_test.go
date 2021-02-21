package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_glue_catalog_database", &resource.Sweeper{
		Name: "aws_glue_catalog_database",
		F:    testSweepGlueCatalogDatabases,
	})
}

func testSweepGlueCatalogDatabases(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).glueconn

	input := &glue.GetDatabasesInput{}
	err = conn.GetDatabasesPages(input, func(page *glue.GetDatabasesOutput, lastPage bool) bool {
		if len(page.DatabaseList) == 0 {
			log.Printf("[INFO] No Glue Catalog Databases to sweep")
			return false
		}
		for _, database := range page.DatabaseList {
			name := aws.StringValue(database.Name)

			log.Printf("[INFO] Deleting Glue Catalog Database: %s", name)

			r := resourceAwsGlueCatalogDatabase()
			d := r.Data(nil)
			d.SetId("???")
			d.Set("name", name)
			d.Set("catalog_id", database.CatalogId)

			err := r.Delete(d, client)
			if err != nil {
				log.Printf("[ERROR] Failed to delete Glue Catalog Database %s: %s", name, err)
			}
		}
		return !lastPage
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Glue Catalog Database sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Glue Catalog Databases: %s", err)
	}

	return nil
}

func TestAccAWSGlueCatalogDatabase_full(t *testing.T) {
	resourceName := "aws_glue_catalog_database.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlueDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccGlueCatalogDatabase_basic(rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlueCatalogDatabaseExists(resourceName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("database/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", ""), resource.TestCheckResourceAttr(resourceName, "location_uri", ""),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:  testAccGlueCatalogDatabase_full(rName, "A test catalog from terraform"),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlueCatalogDatabaseExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "A test catalog from terraform"),
					resource.TestCheckResourceAttr(resourceName, "location_uri", "my-location"),
					resource.TestCheckResourceAttr(resourceName, "parameters.param1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.param2", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameters.param3", "50"),
				),
			},
			{
				Config: testAccGlueCatalogDatabase_full(rName, "An updated test catalog from terraform"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlueCatalogDatabaseExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "An updated test catalog from terraform"),
					resource.TestCheckResourceAttr(resourceName, "location_uri", "my-location"),
					resource.TestCheckResourceAttr(resourceName, "parameters.param1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.param2", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameters.param3", "50"),
				),
			},
		},
	})
}

func TestAccAWSGlueCatalogDatabase_disappears(t *testing.T) {
	resourceName := "aws_glue_catalog_database.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlueDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCatalogDatabase_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlueCatalogDatabaseExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsGlueCatalogDatabase(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckGlueDatabaseDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).glueconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_glue_catalog_database" {
			continue
		}

		catalogId, dbName, err := readAwsGlueCatalogID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &glue.GetDatabaseInput{
			CatalogId: aws.String(catalogId),
			Name:      aws.String(dbName),
		}
		if _, err := conn.GetDatabase(input); err != nil {
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

func testAccGlueCatalogDatabase_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}
`, rName)
}

func testAccGlueCatalogDatabase_full(rName, desc string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name         = %[1]q
  description  = %[2]q
  location_uri = "my-location"

  parameters = {
    param1 = "value1"
    param2 = true
    param3 = 50
  }
}
`, rName, desc)
}

func testAccCheckGlueCatalogDatabaseExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		catalogId, dbName, err := readAwsGlueCatalogID(rs.Primary.ID)
		if err != nil {
			return err
		}

		glueconn := testAccProvider.Meta().(*AWSClient).glueconn
		out, err := glueconn.GetDatabase(&glue.GetDatabaseInput{
			CatalogId: aws.String(catalogId),
			Name:      aws.String(dbName),
		})

		if err != nil {
			return err
		}

		if out.Database == nil {
			return fmt.Errorf("No Glue Database Found")
		}

		if aws.StringValue(out.Database.Name) != dbName {
			return fmt.Errorf("Glue Database Mismatch - existing: %q, state: %q",
				aws.StringValue(out.Database.Name), dbName)
		}

		return nil
	}
}
