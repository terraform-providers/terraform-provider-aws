// +build ignore

package main

import (
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/generators/sweepers"
)

func main() {
	sweepers.Run(
		"kafka",
		map[string]sweepers.ResourceType{
			"Cluster": {
				ListerFunction:       "ListAllClusterPages",
				ListerOutputType:     "ListClustersOutput",
				ListerPageField:      "ClusterInfoList",
				ResourceNameFunction: "ClusterName",
			},
		},
	)
}
