package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

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
