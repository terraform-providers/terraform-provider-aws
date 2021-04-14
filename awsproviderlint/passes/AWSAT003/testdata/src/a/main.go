package a

import (
	"fmt"
)

func f() {
	/* Passing cases */
	fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = %q
}  
`, "10.0.0.0/24")

	fmt.Sprintf(`
resource "aws_config_configuration_aggregator" "example" {
  name = %[1]q

  account_aggregation_source {
    account_ids = [data.aws_caller_identity.current.account_id]
    regions     = [endpoints.UsWest2RegionID]
  }
}

data "aws_caller_identity" "current" {}
`, "rName")

	/* Comment ignored cases */

	//lintignore:AWSAT003
	fmt.Sprintf(`"af-south-1":     %q,`, "525921808201")

	fmt.Sprintf(`"af-south-1":     %q,`, "525921808201") //lintignore:AWSAT003

	/* Failing cases */

	fmt.Println(`availability_zone = "us-west-2a"`) // want "regions should not be hardcoded"

	fmt.Sprintf(`regions      = ["us-west-2"]`) // want "regions should not be hardcoded"
}
