package a

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func f() {
	testSlice := []string{"test"}

	/* Passing cases */

	_ = validation.StringInSlice(testSlice, false)

	_ = validation.StringInSlice(testSlice, true)

	_ = validation.StringInSlice(testFunc(), false)

	/* Comment ignored cases */

	//lintignore:AWSV001
	_ = validation.StringInSlice([]string{"test"}, false)

	_ = validation.StringInSlice([]string{"test"}, false) //lintignore:AWSV001

	/* Failing cases */

	_ = validation.StringInSlice([]string{"test"}, false) // want "prefer AWS Go SDK ENUM_Values\\(\\) function"
}

func testFunc() []string {
	return []string{"test"}
}
