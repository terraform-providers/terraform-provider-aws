package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

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
					resource.TestCheckResourceAttr(
						resourceName,
						"name",
						rName,
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"description",
						"",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"location_uri",
						"",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"parameters.%",
						"0",
					),
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
					resource.TestCheckResourceAttr(
						resourceName,
						"description",
						"A test catalog from terraform",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"location_uri",
						"my-location",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"parameters.param1",
						"value1",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"parameters.param2",
						"true",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"parameters.param3",
						"50",
					),
				),
			},
			{
				Config: testAccGlueCatalogDatabase_full(rName, "An updated test catalog from terraform"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlueCatalogDatabaseExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName,
						"description",
						"An updated test catalog from terraform",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"location_uri",
						"my-location",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"parameters.param1",
						"value1",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"parameters.param2",
						"true",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"parameters.param3",
						"50",
					),
				),
			},
		},
	})
}

func TestAccAWSGlueCatalogDatabase_recreates(t *testing.T) {
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
				),
			},
			{
				// Simulate deleting the database outside Terraform
				PreConfig: func() {
					conn := testAccProvider.Meta().(*AWSClient).glueconn
					input := &glue.DeleteDatabaseInput{
						Name: aws.String(rName),
					}
					_, err := conn.DeleteDatabase(input)
					if err != nil {
						t.Fatalf("error deleting Glue Catalog Database: %s", err)
					}
				},
				Config:             testAccGlueCatalogDatabase_basic(rName),
				ExpectNonEmptyPlan: true,
				PlanOnly:           true,
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

		if *out.Database.Name != dbName {
			return fmt.Errorf("Glue Database Mismatch - existing: %q, state: %q",
				*out.Database.Name, dbName)
		}

		return nil
	}
}
