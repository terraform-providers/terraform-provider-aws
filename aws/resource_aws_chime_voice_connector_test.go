package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/chime"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSChimeVoiceConnector_basic(t *testing.T) {
	var voiceConnector *chime.VoiceConnector

	vcName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_chime_voice_connector.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, chime.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSChimeVoiceConnectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSChimeVoiceConnectorConfig(vcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSChimeVoiceConnectorExists(resourceName, voiceConnector),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("vc-%s", vcName)),
					resource.TestCheckResourceAttr(resourceName, "aws_region", chime.VoiceConnectorAwsRegionUsEast1),
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

func testAccAWSChimeVoiceConnectorConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_chime_voice_connector" "default" {
  name               = "vc-%s"
  require_encryption = true
}
`, name)
}

func testAccCheckAWSChimeVoiceConnectorExists(name string, vc *chime.VoiceConnector) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Chime voice connector ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).chimeconn
		input := &chime.GetVoiceConnectorInput{
			VoiceConnectorId: aws.String(rs.Primary.ID),
		}
		resp, err := conn.GetVoiceConnector(input)
		if err != nil {
			return err
		}

		vc = resp.VoiceConnector

		return nil
	}
}

func testAccCheckAWSChimeVoiceConnectorDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_chime_voice_connector" {
			continue
		}
		conn := testAccProvider.Meta().(*AWSClient).chimeconn
		input := &chime.GetVoiceConnectorInput{
			VoiceConnectorId: aws.String(rs.Primary.ID),
		}
		resp, err := conn.GetVoiceConnector(input)
		if err == nil {
			if resp.VoiceConnector != nil && aws.StringValue(resp.VoiceConnector.Name) != "" {
				return fmt.Errorf("error Chime Voice Connector still exists")
			}
		}
		return nil
	}
	return nil
}
