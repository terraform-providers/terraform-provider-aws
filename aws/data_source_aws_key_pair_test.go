package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsKeyPair_basic(t *testing.T) {
	keyName, publicKey := "testKey", "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 email@example.com"
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
