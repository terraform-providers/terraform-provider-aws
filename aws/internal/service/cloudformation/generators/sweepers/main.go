// +build ignore

package main

import (
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/generators/sweepers"
)

func main() {
	sweepers.Run(
		"cloudformation",
		map[string]sweepers.ResourceType{
			"StackSet": {
				ListerFunction:       "ListAllStackSetsPages",
				ListerOutputType:     "ListStackSetsOutput",
				ListerPageField:      "Summaries",
				ResourceNameFunction: "StackSetName",
			},
		},
	)
}
