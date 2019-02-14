package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_appmesh_route", &resource.Sweeper{
		Name: "aws_appmesh_route",
		F:    testSweepAppmeshRoutes,
	})
}

func testSweepAppmeshRoutes(region string) error {
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
					listRoutesInput := &appmesh.ListRoutesInput{
						MeshName:          mesh.MeshName,
						VirtualRouterName: virtualRouter.VirtualRouterName,
					}
					virtualRouterName := aws.StringValue(virtualRouter.VirtualRouterName)

					err := conn.ListRoutesPages(listRoutesInput, func(page *appmesh.ListRoutesOutput, isLast bool) bool {
						if page == nil {
							return !isLast
						}

						for _, route := range page.Routes {
							input := &appmesh.DeleteRouteInput{
								MeshName:          mesh.MeshName,
								RouteName:         route.RouteName,
								VirtualRouterName: virtualRouter.VirtualRouterName,
							}
							routeName := aws.StringValue(route.RouteName)

							log.Printf("[INFO] Deleting Appmesh Mesh (%s) Virtual Router (%s) Route: %s", meshName, virtualRouterName, routeName)
							_, err := conn.DeleteRoute(input)

							if err != nil {
								log.Printf("[ERROR] Error deleting Appmesh Mesh (%s) Virtual Router (%s) Route (%s): %s", meshName, virtualRouterName, routeName, err)
							}
						}

						return !isLast
					})

					if err != nil {
						log.Printf("[ERROR] Error retrieving Appmesh Mesh (%s) Virtual Router (%s) Routes: %s", meshName, virtualRouterName, err)
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
			log.Printf("[WARN] Skipping Appmesh Mesh sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error retrieving Appmesh Meshes: %s", err)
	}

	return nil
}

func testAccAwsAppmeshRoute_basic(t *testing.T) {
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.foo"
	meshName := fmt.Sprintf("tf-test-mesh-%d", acctest.RandInt())
	vrName := fmt.Sprintf("tf-test-router-%d", acctest.RandInt())
	rName := fmt.Sprintf("tf-test-route-%d", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshRouteConfig_basic(meshName, vrName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshRouteExists(
						resourceName, &r),
					resource.TestCheckResourceAttr(
						resourceName, "name", rName),
					resource.TestCheckResourceAttr(
						resourceName, "mesh_name", meshName),
					resource.TestCheckResourceAttr(
						resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(
						resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttrSet(
						resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(
						resourceName, "last_updated_date"),
					resource.TestMatchResourceAttr(
						resourceName, "arn", regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:appmesh:[^:]+:\\d{12}:mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName))),
				),
			},
		},
	})
}

func testAccAwsAppmeshRoute_allAttributes(t *testing.T) {
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.foo"
	meshName := fmt.Sprintf("tf-test-mesh-%d", acctest.RandInt())
	vrName := fmt.Sprintf("tf-test-router-%d", acctest.RandInt())
	vn1Name := fmt.Sprintf("tf-test-node-%d", acctest.RandInt())
	vn2Name := fmt.Sprintf("tf-test-node-%d", acctest.RandInt())
	rName := fmt.Sprintf("tf-test-route-%d", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshRouteConfig_allAttributes(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshRouteExists(
						resourceName, &r),
					resource.TestCheckResourceAttr(
						resourceName, "name", rName),
					resource.TestCheckResourceAttr(
						resourceName, "mesh_name", meshName),
					resource.TestCheckResourceAttr(
						resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(
						resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.http_route.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.http_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.http_route.0.action.0.weighted_target.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.http_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttrSet(
						resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(
						resourceName, "last_updated_date"),
					resource.TestMatchResourceAttr(
						resourceName, "arn", regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:appmesh:[^:]+:\\d{12}:mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName))),
				),
			},
			{
				Config: testAccAppmeshRouteConfig_allAttributesUpdated(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshRouteExists(
						resourceName, &r),
					resource.TestCheckResourceAttr(
						resourceName, "name", rName),
					resource.TestCheckResourceAttr(
						resourceName, "mesh_name", meshName),
					resource.TestCheckResourceAttr(
						resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(
						resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.http_route.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.http_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.http_route.0.action.0.weighted_target.#", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.http_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.http_route.0.match.0.prefix", "/path"),
					resource.TestCheckResourceAttrSet(
						resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(
						resourceName, "last_updated_date"),
					resource.TestMatchResourceAttr(
						resourceName, "arn", regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:appmesh:[^:]+:\\d{12}:mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName))),
				),
			},
		},
	})
}

func testAccCheckAppmeshRouteDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).appmeshconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appmesh_route" {
			continue
		}

		_, err := conn.DescribeRoute(&appmesh.DescribeRouteInput{
			MeshName:          aws.String(rs.Primary.Attributes["mesh_name"]),
			RouteName:         aws.String(rs.Primary.Attributes["name"]),
			VirtualRouterName: aws.String(rs.Primary.Attributes["virtual_router_name"]),
		})
		if err != nil {
			if isAWSErr(err, appmesh.ErrCodeNotFoundException, "") {
				return nil
			}
			return err
		}
		return fmt.Errorf("still exist.")
	}

	return nil
}

func testAccCheckAppmeshRouteExists(name string, v *appmesh.RouteData) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).appmeshconn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		resp, err := conn.DescribeRoute(&appmesh.DescribeRouteInput{
			MeshName:          aws.String(rs.Primary.Attributes["mesh_name"]),
			RouteName:         aws.String(rs.Primary.Attributes["name"]),
			VirtualRouterName: aws.String(rs.Primary.Attributes["virtual_router_name"]),
		})
		if err != nil {
			return err
		}

		*v = *resp.Route

		return nil
	}
}

func testAccAppmeshRouteConfig_basic(meshName, vrName, rName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "foo" {
  name = "%s"
}

resource "aws_appmesh_virtual_router" "foo" {
  name      = "%s"
  mesh_name = "${aws_appmesh_mesh.foo.id}"

  spec {
    service_names = ["serviceb.simpleapp.local"]
  }
}

resource "aws_appmesh_route" "foo" {
  name                = "%s"
  mesh_name           = "${aws_appmesh_mesh.foo.id}"
  virtual_router_name = "${aws_appmesh_virtual_router.foo.name}"

  spec {}
}
`, meshName, vrName, rName)
}

func testAccAppmeshRouteConfig_allAttributes(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "foo" {
  name = "%s"
}

resource "aws_appmesh_virtual_router" "foo" {
  name      = "%s"
  mesh_name = "${aws_appmesh_mesh.foo.id}"

  spec {
    service_names = ["serviceb.simpleapp.local"]
  }
}

resource "aws_appmesh_virtual_node" "foo" {
  name      = "%s"
  mesh_name = "${aws_appmesh_mesh.foo.id}"

  spec {}
}

resource "aws_appmesh_virtual_node" "bar" {
  name      = "%s"
  mesh_name = "${aws_appmesh_mesh.foo.id}"

  spec {}
}

resource "aws_appmesh_route" "foo" {
  name                = "%s"
  mesh_name           = "${aws_appmesh_mesh.foo.id}"
  virtual_router_name = "${aws_appmesh_virtual_router.foo.name}"

  spec {
    http_route {
      match {
        prefix = "/"
      }

      action {
        weighted_target {
          virtual_node = "${aws_appmesh_virtual_node.foo.name}"
          weight       = 100
        }
      }
    }
  }
}
`, meshName, vrName, vn1Name, vn2Name, rName)
}

func testAccAppmeshRouteConfig_allAttributesUpdated(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "foo" {
  name = "%s"
}

resource "aws_appmesh_virtual_router" "foo" {
  name      = "%s"
  mesh_name = "${aws_appmesh_mesh.foo.id}"

  spec {
    service_names = ["serviceb.simpleapp.local"]
  }
}

resource "aws_appmesh_virtual_node" "foo" {
  name      = "%s"
  mesh_name = "${aws_appmesh_mesh.foo.id}"

  spec {}
}

resource "aws_appmesh_virtual_node" "bar" {
  name      = "%s"
  mesh_name = "${aws_appmesh_mesh.foo.id}"

  spec {}
}

resource "aws_appmesh_route" "foo" {
  name                = "%s"
  mesh_name           = "${aws_appmesh_mesh.foo.id}"
  virtual_router_name = "${aws_appmesh_virtual_router.foo.name}"

  spec {
    http_route {
      match {
        prefix = "/path"
      }

      action {
        weighted_target {
          virtual_node = "${aws_appmesh_virtual_node.foo.name}"
          weight       = 90
        }

        weighted_target {
          virtual_node = "${aws_appmesh_virtual_node.bar.name}"
          weight       = 10
        }
      }
    }
  }
}
`, meshName, vrName, vn1Name, vn2Name, rName)
}
