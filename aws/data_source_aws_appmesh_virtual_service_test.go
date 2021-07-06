package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAWSAppmeshVirtualService_virtualNode(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appmesh_virtual_service.test"
	dataSourceName := "data.aws_appmesh_virtual_service.test"
	vsName := fmt.Sprintf("tf-acc-test-%d.mesh.local", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appmesh.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, appmesh.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshVirtualServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsAppmeshVirtualServiceDataSourceConfig_virtualNode(rName, vsName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "created_date", dataSourceName, "created_date"),
					resource.TestCheckResourceAttrPair(resourceName, "last_updated_date", dataSourceName, "last_updated_date"),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_name", dataSourceName, "mesh_name"),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_owner", dataSourceName, "mesh_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_owner", dataSourceName, "resource_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.provider.0.virtual_node.0.virtual_node_name", dataSourceName, "spec.0.provider.0.virtual_node.0.virtual_node_name"),
					resource.TestCheckResourceAttrPair(resourceName, "tags", dataSourceName, "tags"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSAppmeshVirtualService_virtualRouter(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appmesh_virtual_service.test"
	dataSourceName := "data.aws_appmesh_virtual_service.test"
	vsName := fmt.Sprintf("tf-acc-test-%d.mesh.local", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appmesh.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, appmesh.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshVirtualServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsAppmeshVirtualServiceDataSourceConfig_virtualRouter(rName, vsName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "created_date", dataSourceName, "created_date"),
					resource.TestCheckResourceAttrPair(resourceName, "last_updated_date", dataSourceName, "last_updated_date"),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_name", dataSourceName, "mesh_name"),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_owner", dataSourceName, "mesh_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_owner", dataSourceName, "resource_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.provider.0.virtual_router.0.virtual_router_name", dataSourceName, "spec.0.provider.0.virtual_router.0.virtual_router_name"),
					resource.TestCheckResourceAttrPair(resourceName, "tags", dataSourceName, "tags"),
				),
			},
		},
	})
}

func testAccCheckAwsAppmeshVirtualServiceDataSourceConfig_virtualNode(rName, vsName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_node" "test" {
  name      = %[1]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {}
}

resource "aws_appmesh_virtual_service" "test" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    provider {
      virtual_node {
        virtual_node_name = aws_appmesh_virtual_node.test.name
      }
    }
  }

  tags = {
    foo  = "bar"
    good = "bad"
  }
}

data "aws_appmesh_virtual_service" "test" {
  name      = aws_appmesh_virtual_service.test.name
  mesh_name = aws_appmesh_mesh.test.name
}
`, rName, vsName)
}

func testAccCheckAwsAppmeshVirtualServiceDataSourceConfig_virtualRouter(rName, vsName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_router" "test" {
  name      = %[1]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }
  }
}

resource "aws_appmesh_virtual_service" "test" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    provider {
      virtual_router {
        virtual_router_name = aws_appmesh_virtual_router.test.name
      }
    }
  }
}

data "aws_appmesh_virtual_service" "test" {
  name       = aws_appmesh_virtual_service.test.name
  mesh_name  = aws_appmesh_mesh.test.name
  mesh_owner = data.aws_caller_identity.current.account_id
}
`, rName, vsName)
}
