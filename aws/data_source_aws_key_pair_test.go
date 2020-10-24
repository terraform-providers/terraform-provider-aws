package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsKeyPair_basic(t *testing.T) {
	keyName, publicKey := "testKey", "ssh-rsa testKey"
	resourceName := "data.aws_key_pair.default"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsKeyPairConfig(keyName, publicKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key_name", keyName),
				),
			},
		},
	})
}

func testAccDataSourceAwsKeyPairConfig(keyName, publicKey string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_key_pair" "default" {

  key_name   = "%s"
  public_key = "%s"

  tags = {
    TestIdentifierSet = "testAccDataSourceAwsKeyPair"
  }
}

data "aws_key_pair" "default" {
	key_name = aws_key_pair.default.key_name
}

`, keyName, publicKey)
}
