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
	resourceName := "aws_appmesh_virtual_router.foo"
	meshName := fmt.Sprintf("tf-test-mesh-%d", acctest.RandInt())
	vrName := fmt.Sprintf("tf-test-router-%d", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshVirtualRouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshVirtualRouterConfig(meshName, vrName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualRouterExists(
						resourceName, &vr),
					resource.TestCheckResourceAttr(
						resourceName, "name", vrName),
					resource.TestCheckResourceAttr(
						resourceName, "mesh_name", meshName),
					resource.TestCheckResourceAttr(
						resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.service_names.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.service_names.423761483", "serviceb.simpleapp.local"),
					resource.TestCheckResourceAttrSet(
						resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(
						resourceName, "last_updated_date"),
					resource.TestMatchResourceAttr(
						resourceName, "arn", regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:appmesh:[^:]+:\\d{12}:mesh/%s/virtualRouter/%s", meshName, vrName))),
				),
			},
			{
				Config: testAccAppmeshVirtualRouterConfig_serviceNamesUpdated(meshName, vrName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualRouterExists(
						resourceName, &vr),
					resource.TestCheckResourceAttr(
						resourceName, "name", vrName),
					resource.TestCheckResourceAttr(
						resourceName, "mesh_name", meshName),
					resource.TestCheckResourceAttr(
						resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.service_names.#", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.service_names.3826429429", "serviceb1.simpleapp.local"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.service_names.3079206513", "serviceb2.simpleapp.local"),
				),
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

func testAccAppmeshVirtualRouterConfig(meshName, vrName string) string {
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
`, meshName, vrName)
}

func testAccAppmeshVirtualRouterConfig_serviceNamesUpdated(meshName, vrName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "foo" {
  name = "%s"
}

resource "aws_appmesh_virtual_router" "foo" {
  name      = "%s"
  mesh_name = "${aws_appmesh_mesh.foo.id}"

  spec {
    service_names = ["serviceb1.simpleapp.local", "serviceb2.simpleapp.local"]
  }
}
`, meshName, vrName)
}
