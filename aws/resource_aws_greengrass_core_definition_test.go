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

func TestAccAWSGreengrassCoreDefinition_basic(t *testing.T) {
	rString := acctest.RandString(8)
	resourceName := "aws_greengrass_core_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGreengrassCoreDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGreengrassCoreDefinitionConfig_basic(rString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("core_definition_%s", rString)),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr("aws_greengrass_core_definition.test", "tags.tagKey", "tagValue"),
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

func TestAccAWSGreengrassCoreDefinition_DefinitionVersion(t *testing.T) {
	rString := acctest.RandString(8)
	resourceName := "aws_greengrass_core_definition.test"

	core := map[string]interface{}{
		"sync_shadow": false,
		"id":          "core_id",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGreengrassCoreDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGreengrassCoreDefinitionConfig_definitionVersion(rString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("core_definition_%s", rString)),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					testAccCheckGreengrassCore_checkCore(resourceName, core),
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

func testAccCheckGreengrassCore_checkCore(n string, expectedCore map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Greengrass Core Definition ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).greengrassconn

		getCoreInput := &greengrass.GetCoreDefinitionInput{
			CoreDefinitionId: aws.String(rs.Primary.ID),
		}
		definitionOut, err := conn.GetCoreDefinition(getCoreInput)

		if err != nil {
			return err
		}

		getVersionInput := &greengrass.GetCoreDefinitionVersionInput{
			CoreDefinitionId:        aws.String(rs.Primary.ID),
			CoreDefinitionVersionId: definitionOut.LatestVersion,
		}
		versionOut, err := conn.GetCoreDefinitionVersion(getVersionInput)
		if err != nil {
			return err
		}

		core := versionOut.Definition.Cores[0]
		expectedSyncShadow := expectedCore["sync_shadow"].(bool)
		if *core.SyncShadow != expectedSyncShadow {
			return fmt.Errorf("Sync Shadow %t is not equal expected %t", *core.SyncShadow, expectedSyncShadow)
		}

		expectedCoreId := expectedCore["id"].(string)
		if *core.Id != expectedCoreId {
			return fmt.Errorf("Core Id %s is not equal expected %s", *core.Id, expectedCoreId)
		}

		expectedCertArn, err := getAttrFromResourceState("aws_iot_certificate.foo_cert", "arn", s)
		if err != nil {
			return err
		}
		if *core.CertificateArn != expectedCertArn {
			return fmt.Errorf("Core Certificate Arn %s is not equal expected %s", *core.CertificateArn, expectedCertArn)
		}

		expectedThingArn, err := getAttrFromResourceState("aws_iot_thing.test", "arn", s)
		if err != nil {
			return err
		}
		if *core.ThingArn != expectedThingArn {
			return fmt.Errorf("Core Thing Arn %s is not equal expected %s", *core.ThingArn, expectedThingArn)
		}

		return nil
	}
}

func testAccCheckAWSGreengrassCoreDefinitionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).greengrassconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_greengrass_core_definition" {
			continue
		}

		params := &greengrass.ListCoreDefinitionsInput{
			MaxResults: aws.String("20"),
		}

		out, err := conn.ListCoreDefinitions(params)
		if err != nil {
			return err
		}
		for _, definition := range out.Definitions {
			if *definition.Id == rs.Primary.ID {
				return fmt.Errorf("Expected Greengrass Core Definition to be destroyed, %s found", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccAWSGreengrassCoreDefinitionConfig_basic(rString string) string {
	return fmt.Sprintf(`
resource "aws_greengrass_core_definition" "test" {
  name = "core_definition_%s"

  tags = {
	"tagKey" = "tagValue"
  } 
}
`, rString)
}

func testAccAWSGreengrassCoreDefinitionConfig_definitionVersion(rString string) string {
	return fmt.Sprintf(`

resource "aws_iot_thing" "test" {
	name = "%[1]s"
}

resource "aws_iot_certificate" "foo_cert" {
	csr = "${file("test-fixtures/iot-csr.pem")}"
	active = true
}
	
resource "aws_greengrass_core_definition" "test" {
	name = "core_definition_%[1]s"
	core_definition_version {
		core {
			certificate_arn = "${aws_iot_certificate.foo_cert.arn}"
			id = "core_id"
			sync_shadow = false
			thing_arn = "${aws_iot_thing.test.arn}"
		}
	}
}
`, rString)
}
