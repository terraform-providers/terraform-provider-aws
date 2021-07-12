package a

import (
	"fmt"
)

func f() {
	/* Passing cases */
	fmt.Sprintf(`
data "aws_ec2_instance_type_offering" "available" {
  filter {
    name   = "instance-type"
    values = ["t3.micro", "t2.micro"]
  }

  preferred_instance_types = ["t3.micro", %q]
}

resource "aws_instance" "test" {
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
}
`, "t2.micro")

	/* Comment ignored cases */

	//lintignore:AWSAT007
	fmt.Sprint(`node_type                     = "cache.t3.small"`)

	fmt.Sprint(`node_type                     = "cache.t3.small"`) //lintignore:AWSAT007

	/* Failing cases */

	fmt.Sprint(`node_type                     = "cache.t3.small"`) // want "avoid hardcoding instance type"

	const config = `  instance_type = "t3.micro" ` // want "avoid hardcoding instance type"

	fmt.Sprint(`dedicated_master_type    = "t2.small.elasticsearch"`) // want "avoid hardcoding instance type"
	fmt.Sprintf(`instance_class      = "db.t2.micro"`)                // want "avoid hardcoding instance type"
	fmt.Sprintf(`replication_instance_class   = "dms.t2.micro"`)      // want "avoid hardcoding instance type"

}
