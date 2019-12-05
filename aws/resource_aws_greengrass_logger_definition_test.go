package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/greengrass"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSGreengrassLoggerDefinition_basic(t *testing.T) {
	rString := acctest.RandString(8)
	resourceName := "aws_greengrass_logger_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGreengrassLoggerDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGreengrassLoggerDefinitionConfig_basic(rString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("logger_definition_%s", rString)),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr("aws_greengrass_logger_definition.test", "tags.tagKey", "tagValue"),
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

func TestAccAWSGreengrassLoggerDefinition_DefinitionVersion(t *testing.T) {
	rString := acctest.RandString(8)
	resourceName := "aws_greengrass_logger_definition.test"

	logger := map[string]interface{}{
		"component": "GreengrassSystem",
		"type":      "FileSystem",
		"level":     "DEBUG",
		"id":        "test_id",
		"space":     int64(3),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGreengrassLoggerDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGreengrassLoggerDefinitionConfig_definitionVersion(rString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("logger_definition_%s", rString)),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					testAccCheckGreengrassLogger_checkLogger(resourceName, logger),
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

func testAccCheckGreengrassLogger_checkLogger(n string, expectedLogger map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Greengrass Logger Definition ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).greengrassconn

		getLoggerInput := &greengrass.GetLoggerDefinitionInput{
			LoggerDefinitionId: aws.String(rs.Primary.ID),
		}
		definitionOut, err := conn.GetLoggerDefinition(getLoggerInput)

		if err != nil {
			return err
		}

		getVersionInput := &greengrass.GetLoggerDefinitionVersionInput{
			LoggerDefinitionId:        aws.String(rs.Primary.ID),
			LoggerDefinitionVersionId: definitionOut.LatestVersion,
		}
		versionOut, err := conn.GetLoggerDefinitionVersion(getVersionInput)
		if err != nil {
			return err
		}

		logger := versionOut.Definition.Loggers[0]
		expectedComponent := expectedLogger["component"].(string)
		if *logger.Component != expectedComponent {
			return fmt.Errorf("Component %s is not equal expected %s", *logger.Component, expectedLogger)
		}

		expectedLoggerId := expectedLogger["id"].(string)
		if *logger.Id != expectedLoggerId {
			return fmt.Errorf("Logger Id %s is not equal expected %s", *logger.Id, expectedLoggerId)
		}

		expectedLoggerType := expectedLogger["type"].(string)
		if *logger.Type != expectedLoggerType {
			return fmt.Errorf("Logger Type %s is not equal expected %s", *logger.Type, expectedLoggerType)
		}

		expectedLoggerLevel := expectedLogger["level"].(string)
		if *logger.Level != expectedLoggerLevel {
			return fmt.Errorf("Logger Level %s is not equal expected %s", *logger.Level, expectedLoggerLevel)
		}

		expectedLoggerSpace := expectedLogger["space"].(int64)
		if *logger.Space != expectedLoggerSpace {
			return fmt.Errorf("Logger Space %d is not equal expected %d", *logger.Space, expectedLoggerSpace)
		}
		return nil
	}
}

func testAccCheckAWSGreengrassLoggerDefinitionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).greengrassconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_greengrass_logger_definition" {
			continue
		}

		params := &greengrass.ListLoggerDefinitionsInput{
			MaxResults: aws.String("20"),
		}

		out, err := conn.ListLoggerDefinitions(params)
		if err != nil {
			return err
		}
		for _, definition := range out.Definitions {
			if *definition.Id == rs.Primary.ID {
				return fmt.Errorf("Expected Greengrass Logger Definition to be destroyed, %s found", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccAWSGreengrassLoggerDefinitionConfig_basic(rString string) string {
	return fmt.Sprintf(`
resource "aws_greengrass_logger_definition" "test" {
  name = "logger_definition_%s"

  tags = {
	"tagKey" = "tagValue"
  } 
}
`, rString)
}

func testAccAWSGreengrassLoggerDefinitionConfig_definitionVersion(rString string) string {
	return fmt.Sprintf(`
resource "aws_greengrass_logger_definition" "test" {
	name = "logger_definition_%[1]s"
	logger_definition_version {
		logger {
			component = "GreengrassSystem"
			id = "test_id"
			type = "FileSystem"
			level = "DEBUG"
			space = 3	
		}
	}
}
`, rString)
}
