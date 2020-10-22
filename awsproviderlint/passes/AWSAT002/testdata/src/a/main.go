package a

import (
	"fmt"
)

func f() {
	/* Passing cases */
	fmt.Sprintf(`
data "aws_ami" "amzn-ami-minimal-hvm-ebs" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-minimal-hvm-*"]
  }

  filter {
    name   = "root-device-type"
    values = [%q]
  }
}

resource "aws_spot_fleet_request" "test" {
    launch_specification {
        ami = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    }
}
`, "ebs")

	/* Comment ignored cases */

	//lintignore:AWSAT002
	fmt.Sprintf(`ami = "%s"`, "ami-516b9131")

	fmt.Sprintf(`ami = "%s"`, "ami-516b9131")            //lintignore:AWSAT002
	fmt.Sprintf(`ami = "%s"`, "amzn-ami-minimalist-ebs") //lintignore:AWSAT002

	/* Failing cases */

	fmt.Sprintf(`ami = "%s"`, "ami-516b9131") // want "AMI IDs should not be hardcoded"
}
