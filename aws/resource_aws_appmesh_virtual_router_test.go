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
	resource.AddTestSweepers("aws_appmesh_virtual_router", &resource.Sweeper{
		Name: "aws_appmesh_virtual_router",
		F:    testSweepAppmeshVirtualRouters,
		Dependencies: []string{
			"aws_appmesh_route",
		},
	})
}

func testSweepAppmeshVirtualRouters(region string) error {
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
			listVirtualRoutersInput := &appmesh.ListVirtualRoutersInput{
				MeshName: mesh.MeshName,
			}
			meshName := aws.StringValue(mesh.MeshName)

			err := conn.ListVirtualRoutersPages(listVirtualRoutersInput, func(page *appmesh.ListVirtualRoutersOutput, isLast bool) bool {
				if page == nil {
					return !isLast
				}

				for _, virtualRouter := range page.VirtualRouters {
					input := &appmesh.DeleteVirtualRouterInput{
						MeshName:          mesh.MeshName,
						VirtualRouterName: virtualRouter.VirtualRouterName,
					}
					virtualRouterName := aws.StringValue(virtualRouter.VirtualRouterName)

					log.Printf("[INFO] Deleting Appmesh Mesh (%s) Virtual Router: %s", meshName, virtualRouterName)
					_, err := conn.DeleteVirtualRouter(input)

					if err != nil {
						log.Printf("[ERROR] Error deleting Appmesh Mesh (%s) Virtual Router (%s): %s", meshName, virtualRouterName, err)
					}
				}

				return !isLast
			})

			if err != nil {
				log.Printf("[ERROR] Error retrieving Appmesh Mesh (%s) Virtual Routers: %s", meshName, err)
			}
		}

		return !isLast
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Appmesh Virtual Router sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error retrieving Appmesh Virtual Routers: %s", err)
	}

	return nil
}

func testAccAwsAppmeshVirtualRouter_basic(t *testing.T) {
	var vr appmesh.VirtualRouterData
	resourceName := "aws_appmesh_virtual_router.test"
	meshName := acctest.RandomWithPrefix("tf-acc-test")
	vrName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appmesh.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshVirtualRouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshVirtualRouterConfig_basic(meshName, vrName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualRouterExists(resourceName, &vr),
					resource.TestCheckResourceAttr(resourceName, "name", vrName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					testAccCheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					testAccCheckResourceAttrAccountID(resourceName, "resource_owner"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s", meshName, vrName)),
				),
			},
			{
				Config: testAccAppmeshVirtualRouterConfig_updated(meshName, vrName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualRouterExists(resourceName, &vr),
					resource.TestCheckResourceAttr(resourceName, "name", vrName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					testAccCheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", meshName, vrName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsAppmeshVirtualRouter_tags(t *testing.T) {
	var vr appmesh.VirtualRouterData
	resourceName := "aws_appmesh_virtual_router.test"
	meshName := acctest.RandomWithPrefix("tf-acc-test")
	vrName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appmesh.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshVirtualRouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshVirtualRouterConfig_tags(meshName, vrName, "foo", "bar", "good", "bad"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualRouterExists(
						resourceName, &vr),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.good", "bad"),
				),
			},
			{
				Config: testAccAppmeshVirtualRouterConfig_tags(meshName, vrName, "foo2", "bar", "good", "bad2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualRouterExists(
						resourceName, &vr),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.foo2", "bar"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.good", "bad2"),
				),
			},
			{
				Config: testAccAppmeshVirtualRouterConfig_basic(meshName, vrName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualRouterExists(
						resourceName, &vr),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", meshName, vrName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAppmeshVirtualRouterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).appmeshconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appmesh_virtual_router" {
			continue
		}

		_, err := conn.DescribeVirtualRouter(&appmesh.DescribeVirtualRouterInput{
			MeshName:          aws.String(rs.Primary.Attributes["mesh_name"]),
			VirtualRouterName: aws.String(rs.Primary.Attributes["name"]),
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

func testAccCheckAppmeshVirtualRouterExists(name string, v *appmesh.VirtualRouterData) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).appmeshconn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		resp, err := conn.DescribeVirtualRouter(&appmesh.DescribeVirtualRouterInput{
			MeshName:          aws.String(rs.Primary.Attributes["mesh_name"]),
			VirtualRouterName: aws.String(rs.Primary.Attributes["name"]),
		})
		if err != nil {
			return err
		}

		*v = *resp.VirtualRouter

		return nil
	}
}

func testAccAppmeshVirtualRouterConfig_basic(meshName, vrName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_router" "test" {
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
`, meshName, vrName)
}

func testAccAppmeshVirtualRouterConfig_updated(meshName, vrName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_router" "test" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8081
        protocol = "http"
      }
    }
  }
}
`, meshName, vrName)
}

func testAccAppmeshVirtualRouterConfig_tags(meshName, vrName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_router" "test" {
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

  tags = {
    %[3]s = %[4]q
    %[5]s = %[6]q
  }
}
`, meshName, vrName, tagKey1, tagValue1, tagKey2, tagValue2)
}
