package aws

import (
	"testing"
)

func TestAccAWSAppmesh(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Mesh": {
			"basic":        testAccAwsAppmeshMesh_basic,
			"egressFilter": testAccAwsAppmeshMesh_egressFilter,
			"tags":         testAccAwsAppmeshMesh_tags,
		},
		"Route": {
			"httpHeader":    testAccAwsAppmeshRoute_httpHeader,
			"httpRoute":     testAccAwsAppmeshRoute_httpRoute,
			"tcpRoute":      testAccAwsAppmeshRoute_tcpRoute,
			"routePriority": testAccAwsAppmeshRoute_routePriority,
			"tags":          testAccAwsAppmeshRoute_tags,
		},
		"VirtualNode": {
			"basic":                    testAccAwsAppmeshVirtualNode_basic,
			"cloudMapServiceDiscovery": testAccAwsAppmeshVirtualNode_cloudMapServiceDiscovery,
			"listenerHealthChecks":     testAccAwsAppmeshVirtualNode_listenerHealthChecks,
			"logging":                  testAccAwsAppmeshVirtualNode_logging,
			"tags":                     testAccAwsAppmeshVirtualNode_tags,
		},
		"VirtualRouter": {
			"basic": testAccAwsAppmeshVirtualRouter_basic,
			"tags":  testAccAwsAppmeshVirtualRouter_tags,
		},
		"VirtualService": {
			"virtualNode":   testAccAwsAppmeshVirtualService_virtualNode,
			"virtualRouter": testAccAwsAppmeshVirtualService_virtualRouter,
			"tags":          testAccAwsAppmeshVirtualService_tags,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}
