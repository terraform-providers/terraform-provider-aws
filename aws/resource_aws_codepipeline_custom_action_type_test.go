package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsCodePipelineCustomActionType_basic(t *testing.T) {
	resourceName := "aws_codepipeline_custom_action_type.test"
	rName := acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsCodePipelineCustomActionTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCodePipelineCustomActionType_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCodePipelineCustomActionTypeExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "category", "Build"),
					resource.TestCheckResourceAttr(resourceName, "input_artifact_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "input_artifact_details.0.maximum_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "input_artifact_details.0.minimum_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "output_artifact_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "output_artifact_details.0.maximum_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "input_artifact_details.0.minimum_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "provider_name", fmt.Sprintf("tf-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
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

func TestAccAwsCodePipelineCustomActionType_settings(t *testing.T) {
	resourceName := "aws_codepipeline_custom_action_type.test"
	rName := acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsCodePipelineCustomActionTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCodePipelineCustomActionType_settings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCodePipelineCustomActionTypeExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.entity_url_template", "http://example.com"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.execution_url_template", "http://example.com"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.revision_url_template", "http://example.com"),
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

func TestAccAwsCodePipelineCustomActionType_configurationProperties(t *testing.T) {
	resourceName := "aws_codepipeline_custom_action_type.test"
	rName := acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsCodePipelineCustomActionTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCodePipelineCustomActionType_configurationProperties(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCodePipelineCustomActionTypeExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration_properties.0.description", "tf-test"),
					resource.TestCheckResourceAttr(resourceName, "configuration_properties.0.key", "true"),
					resource.TestCheckResourceAttr(resourceName, "configuration_properties.0.name", fmt.Sprintf("tf-test-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration_properties.0.queryable", "true"),
					resource.TestCheckResourceAttr(resourceName, "configuration_properties.0.required", "true"),
					resource.TestCheckResourceAttr(resourceName, "configuration_properties.0.secret", "false"),
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

func testAccCheckAwsCodePipelineCustomActionTypeExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CodePipeline CustomActionType is set as ID")
		}

		conn := testAccProvider.Meta().(*AWSClient).codepipelineconn

		actionType, err := lookAwsCodePipelineCustomActionType(conn, rs.Primary.ID)
		if err != nil {
			return err
		}
		if actionType == nil {
			return fmt.Errorf("Not found CodePipeline CustomActionType: %s", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckAwsCodePipelineCustomActionTypeDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).codepipelineconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codepipeline_custom_action_type" {
			continue
		}

		actionType, err := lookAwsCodePipelineCustomActionType(conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error reading CodePipeline CustomActionType: %s", err)
		}
		if actionType != nil {
			return fmt.Errorf("CodePipeline CustomActionType still exists: %s", rs.Primary.ID)
		}

		return err
	}

	return nil
}

func testAccAwsCodePipelineCustomActionType_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_codepipeline_custom_action_type" "test" {
  category = "Build"
  input_artifact_details {
    maximum_count = 1
    minimum_count = 0
  }
  output_artifact_details {
    maximum_count = 1
    minimum_count = 0
  }
  provider_name = "tf-%s"
  version = "1"
}
`, rName)
}

func testAccAwsCodePipelineCustomActionType_settings(rName string) string {
	return fmt.Sprintf(`
resource "aws_codepipeline_custom_action_type" "test" {
  category = "Build"
  input_artifact_details {
    maximum_count = 1
    minimum_count = 0
  }
  output_artifact_details {
    maximum_count = 1
    minimum_count = 0
  }
  provider_name = "tf-%s"
  version = "1"
  settings {
    entity_url_template = "http://example.com"
    execution_url_template = "http://example.com"
    revision_url_template = "http://example.com"
  }
}
`, rName)
}

func testAccAwsCodePipelineCustomActionType_configurationProperties(rName string) string {
	return fmt.Sprintf(`
resource "aws_codepipeline_custom_action_type" "test" {
  category = "Build"
  input_artifact_details {
    maximum_count = 1
    minimum_count = 0
  }
  output_artifact_details {
    maximum_count = 1
    minimum_count = 0
  }
  provider_name = "tf-%s"
  version = "1"
  configuration_properties {
    description = "tf-test"
    key = true
    name = "tf-test-%s"
    queryable = true
    required = true
    secret = false
  }
}
`, rName, rName)
}
