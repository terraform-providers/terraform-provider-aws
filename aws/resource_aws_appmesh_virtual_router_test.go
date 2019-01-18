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
