package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
)

// GatewayRoute returns the gateway route corresponding to the specified mesh name, virtual gateway name, gateway route name and optional mesh owner.
// Returns an error if no gateway route is found.
func GatewayRoute(conn *appmesh.AppMesh, meshName, virtualGatewayName, gatewayRouteName, meshOwner string) (*appmesh.GatewayRouteData, error) {
	input := &appmesh.DescribeGatewayRouteInput{
		GatewayRouteName:   aws.String(gatewayRouteName),
		MeshName:           aws.String(meshName),
		VirtualGatewayName: aws.String(virtualGatewayName),
	}
	if meshOwner != "" {
		input.MeshOwner = aws.String(meshOwner)
	}

	output, err := conn.DescribeGatewayRoute(input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output.GatewayRoute, nil
}

// VirtualGateway returns the virtual gateway corresponding to the specified mesh name, virtual gateway name and optional mesh owner.
// Returns an error if no virtual gateway is found.
func VirtualGateway(conn *appmesh.AppMesh, meshName, virtualGatewayName, meshOwner string) (*appmesh.VirtualGatewayData, error) {
	input := &appmesh.DescribeVirtualGatewayInput{
		MeshName:           aws.String(meshName),
		VirtualGatewayName: aws.String(virtualGatewayName),
	}
	if meshOwner != "" {
		input.MeshOwner = aws.String(meshOwner)
	}

	output, err := conn.DescribeVirtualGateway(input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output.VirtualGateway, nil
}
