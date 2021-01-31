package aws

import (
	"testing"
)

func TestAccAWSAppmesh_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"GatewayRoute": {
			"basic":      testAccAwsAppmeshGatewayRoute_basic,
			"disappears": testAccAwsAppmeshGatewayRoute_disappears,
			"grpcRoute":  testAccAwsAppmeshGatewayRoute_GrpcRoute,
			"httpRoute":  testAccAwsAppmeshGatewayRoute_HttpRoute,
			"http2Route": testAccAwsAppmeshGatewayRoute_Http2Route,
			"tags":       testAccAwsAppmeshGatewayRoute_Tags,
		},
		"Mesh": {
			"basic":        testAccAwsAppmeshMesh_basic,
			"egressFilter": testAccAwsAppmeshMesh_egressFilter,
			"tags":         testAccAwsAppmeshMesh_tags,
		},
		"Route": {
			"grpcRoute":           testAccAwsAppmeshRoute_grpcRoute,
			"grpcRouteEmptyMatch": testAccAwsAppmeshRoute_grpcRouteEmptyMatch,
			"grpcRouteTimeout":    testAccAwsAppmeshRoute_grpcRouteTimeout,
			"http2Route":          testAccAwsAppmeshRoute_http2Route,
			"http2RouteTimeout":   testAccAwsAppmeshRoute_http2RouteTimeout,
			"httpHeader":          testAccAwsAppmeshRoute_httpHeader,
			"httpRetryPolicy":     testAccAwsAppmeshRoute_httpRetryPolicy,
			"httpRoute":           testAccAwsAppmeshRoute_httpRoute,
			"httpRouteTimeout":    testAccAwsAppmeshRoute_httpRouteTimeout,
			"routePriority":       testAccAwsAppmeshRoute_routePriority,
			"tcpRoute":            testAccAwsAppmeshRoute_tcpRoute,
			"tcpRouteTimeout":     testAccAwsAppmeshRoute_tcpRouteTimeout,
			"tags":                testAccAwsAppmeshRoute_tags,
		},
		"VirtualGateway": {
			"basic":                  testAccAwsAppmeshVirtualGateway_basic,
			"disappears":             testAccAwsAppmeshVirtualGateway_disappears,
			"backendDefaults":        testAccAwsAppmeshVirtualGateway_BackendDefaults,
			"listenerConnectionPool": testAccAwsAppmeshVirtualGateway_ListenerConnectionPool,
			"listenerHealthChecks":   testAccAwsAppmeshVirtualGateway_ListenerHealthChecks,
			"listenerTls":            testAccAwsAppmeshVirtualGateway_ListenerTls,
			"logging":                testAccAwsAppmeshVirtualGateway_Logging,
			"tags":                   testAccAwsAppmeshVirtualGateway_Tags,
		},
		"VirtualNode": {
			"basic":                    testAccAwsAppmeshVirtualNode_basic,
			"disappears":               testAccAwsAppmeshVirtualNode_disappears,
			"backendClientPolicyAcm":   testAccAwsAppmeshVirtualNode_backendClientPolicyAcm,
			"backendClientPolicyFile":  testAccAwsAppmeshVirtualNode_backendClientPolicyFile,
			"backendDefaults":          testAccAwsAppmeshVirtualNode_backendDefaults,
			"cloudMapServiceDiscovery": testAccAwsAppmeshVirtualNode_cloudMapServiceDiscovery,
			"listenerConnectionPool":   testAccAwsAppmeshVirtualNode_listenerConnectionPool,
			"listenerOutlierDetection": testAccAwsAppmeshVirtualNode_listenerOutlierDetection,
			"listenerHealthChecks":     testAccAwsAppmeshVirtualNode_listenerHealthChecks,
			"listenerTimeout":          testAccAwsAppmeshVirtualNode_listenerTimeout,
			"listenerTls":              testAccAwsAppmeshVirtualNode_listenerTls,
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
