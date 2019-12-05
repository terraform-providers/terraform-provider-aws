package aws

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/greengrass"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSGreengrassConnectorDefinition_basic(t *testing.T) {
	rString := acctest.RandString(8)
	resourceName := "aws_greengrass_connector_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGreengrassConnectorDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGreengrassConnectorDefinitionConfig_basic(rString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("connector_definition_%s", rString)),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr("aws_greengrass_connector_definition.test", "tags.tagKey", "tagValue"),
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

func TestAccAWSGreengrassConnectorDefinition_DefinitionVersion(t *testing.T) {
	rString := acctest.RandString(8)
	resourceName := "aws_greengrass_connector_definition.test"

	connector := map[string]interface{}{
		"connector_arn": "arn:aws:greengrass:eu-west-1::/connectors/RaspberryPiGPIO/versions/5",
		"id":            "connector_id",
	}

	parameters := map[string]string{
		"key": "value",
	}
	connector["parameters"] = parameters

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGreengrassConnectorDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGreengrassConnectorDefinitionConfig_definitionVersion(rString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("connector_definition_%s", rString)),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					testAccCheckGreengrassConnector_checkConnector(resourceName, connector),
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

func testAccCheckGreengrassConnector_checkConnector(n string, expectedConnector map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Greengrass Connector Definition ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).greengrassconn

		getConnectorInput := &greengrass.GetConnectorDefinitionInput{
			ConnectorDefinitionId: aws.String(rs.Primary.ID),
		}
		definitionOut, err := conn.GetConnectorDefinition(getConnectorInput)

		if err != nil {
			return err
		}

		getVersionInput := &greengrass.GetConnectorDefinitionVersionInput{
			ConnectorDefinitionId:        aws.String(rs.Primary.ID),
			ConnectorDefinitionVersionId: definitionOut.LatestVersion,
		}
		versionOut, err := conn.GetConnectorDefinitionVersion(getVersionInput)
		if err != nil {
			return err
		}

		connector := versionOut.Definition.Connectors[0]
		expectedConnectorArn := expectedConnector["connector_arn"].(string)
		if *connector.ConnectorArn != expectedConnectorArn {
			return fmt.Errorf("Connector Arn %s is not equal to expected %s", *connector.ConnectorArn, expectedConnectorArn)
		}

		expectedConnectorId := expectedConnector["id"].(string)
		if *connector.Id != expectedConnectorId {
			return fmt.Errorf("Connector ID %s is not equal to expected %s", *connector.Id, expectedConnectorId)
		}

		expectedParameters := expectedConnector["parameters"].(map[string]string)

		parameters := make(map[string]string)
		for k, v := range connector.Parameters {
			parameters[k] = *v
		}

		if !reflect.DeepEqual(parameters, expectedParameters) {
			return fmt.Errorf("Connector Parameters %v is not equal to expected %v", parameters, expectedParameters)
		}

		return nil
	}
}
func testAccCheckAWSGreengrassConnectorDefinitionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).greengrassconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_greengrass_connector_definition" {
			continue
		}

		params := &greengrass.ListConnectorDefinitionsInput{
			MaxResults: aws.String("20"),
		}

		out, err := conn.ListConnectorDefinitions(params)
		if err != nil {
			return err
		}
		for _, definition := range out.Definitions {
			if *definition.Id == rs.Primary.ID {
				return fmt.Errorf("Expected Greengrass Connector Definition to be destroyed, %s found", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccAWSGreengrassConnectorDefinitionConfig_basic(rString string) string {
	return fmt.Sprintf(`
resource "aws_greengrass_connector_definition" "test" {
  name = "connector_definition_%s"

  tags = {
	"tagKey" = "tagValue"
  }
}
`, rString)
}

func testAccAWSGreengrassConnectorDefinitionConfig_definitionVersion(rString string) string {
	return fmt.Sprintf(`
resource "aws_greengrass_connector_definition" "test" {
	name = "connector_definition_%s"
	connector_definition_version {
		connector {
			connector_arn = "arn:aws:greengrass:eu-west-1::/connectors/RaspberryPiGPIO/versions/5"
			id = "connector_id"
			parameters = {
				"key" = "value",
			}
		}
	}
}
`, rString)
}
