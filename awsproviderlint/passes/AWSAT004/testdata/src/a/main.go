package a

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func f() {
	resourceName := "resource"

	/* Passing cases */
	resource.TestCheckResourceAttr(resourceName, "listener.*.instance_port", "8000")

	/* Comment ignored cases */

	//lintignore:AWSAT004
	resource.TestCheckResourceAttr(resourceName, "listener.206423021.instance_port", "8000")

	resource.TestCheckResourceAttr(resourceName, "listener.206423021.instance_port", "8000") //lintignore:AWSAT004

	/* Failing cases */
	resource.TestCheckResourceAttr(resourceName, "listener.206423021.instance_port", "8000") // want "avoid hardcoded state hashes"
	resource.TestCheckResourceAttr(resourceName, "rule.3061118601.statement.#", "1")         // want "avoid hardcoded state hashes"
}
