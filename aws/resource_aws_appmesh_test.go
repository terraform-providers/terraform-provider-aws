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
			"grpcRoute":         testAccAwsAppmeshRoute_grpcRoute,
			"grpcRouteTimeout":  testAccAwsAppmeshRoute_grpcRouteTimeout,
			"http2Route":        testAccAwsAppmeshRoute_http2Route,
			"http2RouteTimeout": testAccAwsAppmeshRoute_http2RouteTimeout,
			"httpHeader":        testAccAwsAppmeshRoute_httpHeader,
			"httpRetryPolicy":   testAccAwsAppmeshRoute_httpRetryPolicy,
			"httpRoute":         testAccAwsAppmeshRoute_httpRoute,
			"httpRouteTimeout":  testAccAwsAppmeshRoute_httpRouteTimeout,
			"routePriority":     testAccAwsAppmeshRoute_routePriority,
			"tcpRoute":          testAccAwsAppmeshRoute_tcpRoute,
			"tcpRouteTimeout":   testAccAwsAppmeshRoute_tcpRouteTimeout,
			"tags":              testAccAwsAppmeshRoute_tags,
		},
		"VirtualGateway": {
			"backendDefaults":      testAccAwsAppmeshVirtualGateway_BackendDefaults,
			"basic":                testAccAwsAppmeshVirtualGateway_basic,
			"disappears":           testAccAwsAppmeshVirtualGateway_disappears,
			"listenerHealthChecks": testAccAwsAppmeshVirtualGateway_ListenerHealthChecks,
			"logging":              testAccAwsAppmeshVirtualGateway_Logging,
			"tags":                 testAccAwsAppmeshVirtualGateway_Tags,
			"tls":                  testAccAwsAppmeshVirtualGateway_TLS,
		},
		"VirtualNode": {
			"basic":                    testAccAwsAppmeshVirtualNode_basic,
			"backendDefaults":          testAccAwsAppmeshVirtualNode_backendDefaults,
			"clientPolicyAcm":          testAccAwsAppmeshVirtualNode_clientPolicyAcm,
			"clientPolicyFile":         testAccAwsAppmeshVirtualNode_clientPolicyFile,
			"cloudMapServiceDiscovery": testAccAwsAppmeshVirtualNode_cloudMapServiceDiscovery,
			"listenerHealthChecks":     testAccAwsAppmeshVirtualNode_listenerHealthChecks,
			"listenerTimeout":          testAccAwsAppmeshVirtualNode_listenerTimeout,
			"logging":                  testAccAwsAppmeshVirtualNode_logging,
			"tls":                      testAccAwsAppmeshVirtualNode_tls,
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
