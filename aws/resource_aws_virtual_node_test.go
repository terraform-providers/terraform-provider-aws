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

func testAccAwsAppmeshVirtualNode_basic(t *testing.T) {
	var vn appmesh.VirtualNodeData
	resourceName := "aws_appmesh_virtual_node.foo"
	meshName := fmt.Sprintf("tf-test-mesh-%d", acctest.RandInt())
	vnName := fmt.Sprintf("tf-test-node-%d", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshVirtualNodeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshVirtualNodeConfig_basic(meshName, vnName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualNodeExists(
						resourceName, &vn),
					resource.TestCheckResourceAttr(
						resourceName, "name", vnName),
					resource.TestCheckResourceAttr(
						resourceName, "mesh_name", meshName),
					resource.TestCheckResourceAttr(
						resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.backends.#", "0"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.listener.#", "0"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.service_discovery.#", "0"),
					resource.TestCheckResourceAttrSet(
						resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(
						resourceName, "last_updated_date"),
					resource.TestMatchResourceAttr(
						resourceName, "arn", regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:appmesh:[^:]+:\\d{12}:mesh/%s/virtualNode/%s", meshName, vnName))),
				),
			},
		},
	})
}

func testAccAwsAppmeshVirtualNode_allAttributes(t *testing.T) {
	var vn appmesh.VirtualNodeData
	resourceName := "aws_appmesh_virtual_node.foo"
	meshName := fmt.Sprintf("tf-test-mesh-%d", acctest.RandInt())
	vnName := fmt.Sprintf("tf-test-node-%d", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshVirtualNodeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshVirtualNodeConfig_allAttributes(meshName, vnName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualNodeExists(
						resourceName, &vn),
					resource.TestCheckResourceAttr(
						resourceName, "name", vnName),
					resource.TestCheckResourceAttr(
						resourceName, "mesh_name", meshName),
					resource.TestCheckResourceAttr(
						resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.backends.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.backends.1255689679", "servicea.simpleapp.local"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.listener.2279702354.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.listener.2279702354.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.listener.2279702354.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.service_discovery.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.service_discovery.0.dns.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.service_discovery.0.dns.0.service_name", "serviceb.simpleapp.local"),
					resource.TestCheckResourceAttrSet(
						resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(
						resourceName, "last_updated_date"),
					resource.TestMatchResourceAttr(
						resourceName, "arn", regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:appmesh:[^:]+:\\d{12}:mesh/%s/virtualNode/%s", meshName, vnName))),
				),
			},
			{
				Config: testAccAppmeshVirtualNodeConfig_allAttributesUpdated(meshName, vnName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualNodeExists(
						resourceName, &vn),
					resource.TestCheckResourceAttr(
						resourceName, "name", vnName),
					resource.TestCheckResourceAttr(
						resourceName, "mesh_name", meshName),
					resource.TestCheckResourceAttr(
						resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.backends.#", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.backends.2665798920", "servicec.simpleapp.local"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.backends.3195445571", "serviced.simpleapp.local"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.listener.563508454.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.listener.563508454.port_mapping.0.port", "8081"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.listener.563508454.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.service_discovery.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.service_discovery.0.dns.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "spec.0.service_discovery.0.dns.0.service_name", "serviceb1.simpleapp.local"),
					resource.TestCheckResourceAttrSet(
						resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(
						resourceName, "last_updated_date"),
					resource.TestMatchResourceAttr(
						resourceName, "arn", regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:appmesh:[^:]+:\\d{12}:mesh/%s/virtualNode/%s", meshName, vnName))),
				),
			},
		},
	})
}

func testAccCheckAppmeshVirtualNodeDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).appmeshconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appmesh_virtual_node" {
			continue
		}

		_, err := conn.DescribeVirtualNode(&appmesh.DescribeVirtualNodeInput{
			MeshName:        aws.String(rs.Primary.Attributes["mesh_name"]),
			VirtualNodeName: aws.String(rs.Primary.Attributes["name"]),
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

func testAccCheckAppmeshVirtualNodeExists(name string, v *appmesh.VirtualNodeData) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).appmeshconn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		resp, err := conn.DescribeVirtualNode(&appmesh.DescribeVirtualNodeInput{
			MeshName:        aws.String(rs.Primary.Attributes["mesh_name"]),
			VirtualNodeName: aws.String(rs.Primary.Attributes["name"]),
		})
		if err != nil {
			return err
		}

		*v = *resp.VirtualNode

		return nil
	}
}

func testAccAppmeshVirtualNodeConfig_basic(meshName, vnName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "foo" {
  name = "%s"
}

resource "aws_appmesh_virtual_node" "foo" {
  name      = "%s"
  mesh_name = "${aws_appmesh_mesh.foo.id}"

  spec {}
}
`, meshName, vnName)
}

func testAccAppmeshVirtualNodeConfig_allAttributes(meshName, vnName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "foo" {
  name = "%s"
}

resource "aws_appmesh_virtual_node" "foo" {
  name      = "%s"
  mesh_name = "${aws_appmesh_mesh.foo.id}"

  spec {
    backends = ["servicea.simpleapp.local"]

    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }

    service_discovery {
      dns {
        service_name = "serviceb.simpleapp.local"
      }
    }
  }
}
`, meshName, vnName)
}

func testAccAppmeshVirtualNodeConfig_allAttributesUpdated(meshName, vnName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "foo" {
  name = "%s"
}

resource "aws_appmesh_virtual_node" "foo" {
  name      = "%s"
  mesh_name = "${aws_appmesh_mesh.foo.id}"

  spec {
    backends = ["servicec.simpleapp.local", "serviced.simpleapp.local"]

    listener {
      port_mapping {
        port     = 8081
        protocol = "http"
      }
    }

    service_discovery {
      dns {
        service_name = "serviceb1.simpleapp.local"
      }
    }
  }
}
`, meshName, vnName)
}
