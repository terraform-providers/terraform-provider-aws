package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_appmesh_virtual_service", &resource.Sweeper{
		Name: "aws_appmesh_virtual_service",
		F:    testSweepAppmeshVirtualServices,
	})
}

func testSweepAppmeshVirtualServices(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).appmeshconn

	err = conn.ListMeshesPages(&appmesh.ListMeshesInput{}, func(page *appmesh.ListMeshesOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, mesh := range page.Meshes {
			listVirtualServicesInput := &appmesh.ListVirtualServicesInput{
				MeshName: mesh.MeshName,
			}
			meshName := aws.StringValue(mesh.MeshName)

			err := conn.ListVirtualServicesPages(listVirtualServicesInput, func(page *appmesh.ListVirtualServicesOutput, isLast bool) bool {
				if page == nil {
					return !isLast
				}

				for _, virtualService := range page.VirtualServices {
					input := &appmesh.DeleteVirtualServiceInput{
						MeshName:           mesh.MeshName,
						VirtualServiceName: virtualService.VirtualServiceName,
					}
					virtualServiceName := aws.StringValue(virtualService.VirtualServiceName)

					log.Printf("[INFO] Deleting Appmesh Mesh (%s) Virtual Service: %s", meshName, virtualServiceName)
					_, err := conn.DeleteVirtualService(input)

					if err != nil {
						log.Printf("[ERROR] Error deleting Appmesh Mesh (%s) Virtual Service (%s): %s", meshName, virtualServiceName, err)
					}
				}

				return !isLast
			})

			if err != nil {
				log.Printf("[ERROR] Error retrieving Appmesh Mesh (%s) Virtual Services: %s", meshName, err)
			}
		}

		return !isLast
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Appmesh Virtual Service sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error retrieving Appmesh Virtual Services: %s", err)
	}

	return nil
}

func testAccAwsAppmeshVirtualService_virtualNode(t *testing.T) {
	var vs appmesh.VirtualServiceData
	resourceName := "aws_appmesh_virtual_service.test"
	meshName := acctest.RandomWithPrefix("tf-acc-test")
	vnName1 := acctest.RandomWithPrefix("tf-acc-test")
	vnName2 := acctest.RandomWithPrefix("tf-acc-test")
	vsName := fmt.Sprintf("tf-acc-test-%d.mesh.local", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appmesh.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshVirtualServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshVirtualServiceConfig_virtualNode(meshName, vnName1, vnName2, vsName, "aws_appmesh_virtual_node.foo"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualServiceExists(resourceName, &vs),
					resource.TestCheckResourceAttr(resourceName, "name", vsName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					testAccCheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.provider.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.provider.0.virtual_node.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.provider.0.virtual_node.0.virtual_node_name", vnName1),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					testAccCheckResourceAttrAccountID(resourceName, "resource_owner"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualService/%s", meshName, vsName)),
				),
			},
			{
				Config: testAccAppmeshVirtualServiceConfig_virtualNode(meshName, vnName1, vnName2, vsName, "aws_appmesh_virtual_node.bar"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualServiceExists(resourceName, &vs),
					resource.TestCheckResourceAttr(resourceName, "name", vsName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					testAccCheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.provider.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.provider.0.virtual_node.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.provider.0.virtual_node.0.virtual_node_name", vnName2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", meshName, vsName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsAppmeshVirtualService_virtualRouter(t *testing.T) {
	var vs appmesh.VirtualServiceData
	resourceName := "aws_appmesh_virtual_service.test"
	meshName := acctest.RandomWithPrefix("tf-acc-test")
	vrName1 := acctest.RandomWithPrefix("tf-acc-test")
	vrName2 := acctest.RandomWithPrefix("tf-acc-test")
	vsName := fmt.Sprintf("tf-acc-test-%d.mesh.local", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appmesh.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshVirtualServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshVirtualServiceConfig_virtualRouter(meshName, vrName1, vrName2, vsName, "aws_appmesh_virtual_router.foo"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualServiceExists(resourceName, &vs),
					resource.TestCheckResourceAttr(resourceName, "name", vsName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					testAccCheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.provider.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.provider.0.virtual_router.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.provider.0.virtual_router.0.virtual_router_name", vrName1),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					testAccCheckResourceAttrAccountID(resourceName, "resource_owner"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualService/%s", meshName, vsName))),
			},
			{
				Config: testAccAppmeshVirtualServiceConfig_virtualRouter(meshName, vrName1, vrName2, vsName, "aws_appmesh_virtual_router.bar"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualServiceExists(resourceName, &vs),
					resource.TestCheckResourceAttr(resourceName, "name", vsName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					testAccCheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.provider.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.provider.0.virtual_router.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.provider.0.virtual_router.0.virtual_router_name", vrName2),
				),
			},
		},
	})
}

func testAccAwsAppmeshVirtualService_tags(t *testing.T) {
	var vs appmesh.VirtualServiceData
	resourceName := "aws_appmesh_virtual_service.test"
	meshName := acctest.RandomWithPrefix("tf-acc-test")
	vnName1 := acctest.RandomWithPrefix("tf-acc-test")
	vnName2 := acctest.RandomWithPrefix("tf-acc-test")
	vsName := fmt.Sprintf("tf-acc-test-%d.mesh.local", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appmesh.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshVirtualServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshVirtualServiceConfig_tags(meshName, vnName1, vnName2, vsName, "aws_appmesh_virtual_node.foo", "foo", "bar", "good", "bad"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualServiceExists(resourceName, &vs),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.good", "bad"),
				),
			},
			{
				Config: testAccAppmeshVirtualServiceConfig_tags(meshName, vnName1, vnName2, vsName, "aws_appmesh_virtual_node.foo", "foo2", "bar", "good", "bad2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualServiceExists(resourceName, &vs),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.foo2", "bar"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.good", "bad2"),
				),
			},
			{
				Config: testAccAppmeshVirtualServiceConfig_virtualNode(meshName, vnName1, vnName2, vsName, "aws_appmesh_virtual_node.foo"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualServiceExists(resourceName, &vs),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", meshName, vsName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAppmeshVirtualServiceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).appmeshconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appmesh_virtual_service" {
			continue
		}

		_, err := conn.DescribeVirtualService(&appmesh.DescribeVirtualServiceInput{
			MeshName:           aws.String(rs.Primary.Attributes["mesh_name"]),
			VirtualServiceName: aws.String(rs.Primary.Attributes["name"]),
		})
		if isAWSErr(err, appmesh.ErrCodeNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("still exist.")
	}

	return nil
}

func testAccCheckAppmeshVirtualServiceExists(name string, v *appmesh.VirtualServiceData) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).appmeshconn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		resp, err := conn.DescribeVirtualService(&appmesh.DescribeVirtualServiceInput{
			MeshName:           aws.String(rs.Primary.Attributes["mesh_name"]),
			VirtualServiceName: aws.String(rs.Primary.Attributes["name"]),
		})
		if err != nil {
			return err
		}

		*v = *resp.VirtualService

		return nil
	}
}

func testAccAppmeshVirtualServiceConfig_virtualNode(meshName, vnName1, vnName2, vsName, rName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_node" "foo" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {}
}

resource "aws_appmesh_virtual_node" "bar" {
  name      = %[3]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {}
}

resource "aws_appmesh_virtual_service" "test" {
  name      = %[4]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    provider {
      virtual_node {
        virtual_node_name = %[5]s.name
      }
    }
  }
}
`, meshName, vnName1, vnName2, vsName, rName)
}

func testAccAppmeshVirtualServiceConfig_virtualRouter(meshName, vrName1, vrName2, vsName, rName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_router" "foo" {
  name      = %[2]q
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

resource "aws_appmesh_virtual_router" "bar" {
  name      = %[3]q
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
  name      = %[4]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    provider {
      virtual_router {
        virtual_router_name = %[5]s.name
      }
    }
  }
}
`, meshName, vrName1, vrName2, vsName, rName)
}

func testAccAppmeshVirtualServiceConfig_tags(meshName, vnName1, vnName2, vsName, rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_node" "foo" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {}
}

resource "aws_appmesh_virtual_node" "bar" {
  name      = %[3]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {}
}

resource "aws_appmesh_virtual_service" "test" {
  name      = %[4]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    provider {
      virtual_node {
        virtual_node_name = %[5]s.name
      }
    }
  }

  tags = {
    %[6]s = %[7]q
    %[8]s = %[9]q
  }
}
`, meshName, vnName1, vnName2, vsName, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
