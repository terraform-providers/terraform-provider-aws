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
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/glue/finder"
)

func init() {
	resource.AddTestSweepers("aws_glue_registry", &resource.Sweeper{
		Name: "aws_glue_registry",
		F:    testSweepGlueRegistry,
	})
}

func testSweepGlueRegistry(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).glueconn

	listOutput, err := conn.ListRegistries(&glue.ListRegistriesInput{})
	if err != nil {
		// Some endpoints that do not support Glue Registrys return InternalFailure
		if testSweepSkipSweepError(err) || isAWSErr(err, "InternalFailure", "") {
			log.Printf("[WARN] Skipping Glue Registry sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Glue Registry: %s", err)
	}
	for _, registry := range listOutput.Registries {
		arn := aws.StringValue(registry.RegistryArn)
		r := resourceAwsGlueRegistry()
		d := r.Data(nil)
		d.SetId(arn)

		err := r.Delete(d, client)
		if err != nil {
			log.Printf("[ERROR] Failed to delete Glue Registry %s: %s", arn, err)
		}
	}
	return nil
}

func TestAccAWSGlueRegistry_basic(t *testing.T) {
	var registry glue.GetRegistryOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_registry.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSGlueRegistry(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueRegistryBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueRegistryExists(resourceName, &registry),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("registry/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "registry_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAWSGlueRegistry_Description(t *testing.T) {
	var registry glue.GetRegistryOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_registry.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSGlueRegistry(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueRegistryDescriptionConfig(rName, "First Description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueRegistryExists(resourceName, &registry),
					resource.TestCheckResourceAttr(resourceName, "description", "First Description"),
				),
			},
			{
				Config: testAccAWSGlueRegistryDescriptionConfig(rName, "Second Description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueRegistryExists(resourceName, &registry),
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

func TestAccAWSGlueRegistry_tags(t *testing.T) {
	var registry glue.GetRegistryOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_registry.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSGlueRegistry(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueRegistryConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueRegistryExists(resourceName, &registry),
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
				Config: testAccAWSGlueRegistryConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueRegistryExists(resourceName, &registry),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSGlueRegistryConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueRegistryExists(resourceName, &registry),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSGlueRegistry_disappears(t *testing.T) {
	var registry glue.GetRegistryOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_registry.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSGlueRegistry(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueRegistryBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueRegistryExists(resourceName, &registry),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsGlueRegistry(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccPreCheckAWSGlueRegistry(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).glueconn

	_, err := conn.ListRegistries(&glue.ListRegistriesInput{})

	// Some endpoints that do not support Glue Registrys return InternalFailure
	if testAccPreCheckSkipError(err) || isAWSErr(err, "InternalFailure", "") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckAWSGlueRegistryExists(resourceName string, registry *glue.GetRegistryOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Glue Registry ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).glueconn
		output, err := finder.RegistryByID(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("Glue Registry (%s) not found", rs.Primary.ID)
		}

		if aws.StringValue(output.RegistryArn) == rs.Primary.ID {
			*registry = *output
			return nil
		}

		return fmt.Errorf("Glue Registry (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckAWSGlueRegistryDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_glue_registry" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).glueconn
		output, err := finder.RegistryByID(conn, rs.Primary.ID)
		if err != nil {
			if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
				return nil
			}

		}

		if output != nil && aws.StringValue(output.RegistryArn) == rs.Primary.ID {
			return fmt.Errorf("Glue Registry %s still exists", rs.Primary.ID)
		}

		return err
	}

	return nil
}

func testAccAWSGlueRegistryDescriptionConfig(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_glue_registry" "test" {
  registry_name = %[1]q
  description   = %[2]q
}
`, rName, description)
}

func testAccAWSGlueRegistryBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_registry" "test" {
  registry_name = %[1]q
}
`, rName)
}

func testAccAWSGlueRegistryConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_glue_registry" "test" {
  registry_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSGlueRegistryConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_glue_registry" "test" {
  registry_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
