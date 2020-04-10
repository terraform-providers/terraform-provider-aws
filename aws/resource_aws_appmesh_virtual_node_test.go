package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	resource.AddTestSweepers("aws_appmesh_virtual_node", &resource.Sweeper{
		Name: "aws_appmesh_virtual_node",
		F:    testSweepAppmeshVirtualNodes,
	})
}

func testSweepAppmeshVirtualNodes(region string) error {
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
			listVirtualNodesInput := &appmesh.ListVirtualNodesInput{
				MeshName: mesh.MeshName,
			}
			meshName := aws.StringValue(mesh.MeshName)

			err := conn.ListVirtualNodesPages(listVirtualNodesInput, func(page *appmesh.ListVirtualNodesOutput, isLast bool) bool {
				if page == nil {
					return !isLast
				}

				for _, virtualNode := range page.VirtualNodes {
					input := &appmesh.DeleteVirtualNodeInput{
						MeshName:        mesh.MeshName,
						VirtualNodeName: virtualNode.VirtualNodeName,
					}
					virtualNodeName := aws.StringValue(virtualNode.VirtualNodeName)

					log.Printf("[INFO] Deleting Appmesh Mesh (%s) Virtual Node: %s", meshName, virtualNodeName)
					_, err := conn.DeleteVirtualNode(input)

					if err != nil {
						log.Printf("[ERROR] Error deleting Appmesh Mesh (%s) Virtual Node (%s): %s", meshName, virtualNodeName, err)
					}
				}

				return !isLast
			})

			if err != nil {
				log.Printf("[ERROR] Error retrieving Appmesh Mesh (%s) Virtual Nodes: %s", meshName, err)
			}
		}

		return !isLast
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Appmesh Virtual Node sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error retrieving Appmesh Virtual Nodes: %s", err)
	}

	return nil
}

func testAccAwsAppmeshVirtualNode_basic(t *testing.T) {
	var vn appmesh.VirtualNodeData
	resourceName := "aws_appmesh_virtual_node.test"
	meshName := acctest.RandomWithPrefix("tf-acc-test")
	vnName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshVirtualNodeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshVirtualNodeConfig_basic(meshName, vnName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualNodeExists(resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, "name", vnName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualNode/%s", meshName, vnName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", meshName, vnName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsAppmeshVirtualNode_cloudMapServiceDiscovery(t *testing.T) {
	var vn appmesh.VirtualNodeData
	resourceName := "aws_appmesh_virtual_node.test"
	nsResourceName := "aws_service_discovery_http_namespace.test"
	meshName := acctest.RandomWithPrefix("tf-acc-test")
	vnName := acctest.RandomWithPrefix("tf-acc-test")
	// Avoid 'config is invalid: last character of "name" must be a letter' for aws_service_discovery_http_namespace.
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(20, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshVirtualNodeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshVirtualNodeConfig_cloudMapServiceDiscovery(meshName, vnName, rName, "Key1", "Value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualNodeExists(resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, "name", vnName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.aws_cloud_map.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.aws_cloud_map.0.attributes.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.aws_cloud_map.0.attributes.Key1", "Value1"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.service_discovery.0.aws_cloud_map.0.namespace_name", nsResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.aws_cloud_map.0.service_name", rName),
				),
			},
			{
				Config: testAccAppmeshVirtualNodeConfig_cloudMapServiceDiscovery(meshName, vnName, rName, "Key1", "Value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualNodeExists(resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, "name", vnName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.aws_cloud_map.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.aws_cloud_map.0.attributes.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.aws_cloud_map.0.attributes.Key1", "Value2"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.service_discovery.0.aws_cloud_map.0.namespace_name", nsResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.aws_cloud_map.0.service_name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", meshName, vnName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsAppmeshVirtualNode_listenerHealthChecks(t *testing.T) {
	var vn appmesh.VirtualNodeData
	resourceName := "aws_appmesh_virtual_node.test"
	meshName := acctest.RandomWithPrefix("tf-acc-test")
	vnName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshVirtualNodeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshVirtualNodeConfig_listenerHealthChecks(meshName, vnName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualNodeExists(resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, "name", vnName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend.2622272660.virtual_service.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend.2622272660.virtual_service.0.virtual_service_name", "servicea.simpleapp.local"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.433446196.health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.433446196.health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.433446196.health_check.0.interval_millis", "5000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.433446196.health_check.0.path", "/ping"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.433446196.health_check.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.433446196.health_check.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.433446196.health_check.0.timeout_millis", "2000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.433446196.health_check.0.unhealthy_threshold", "5"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.433446196.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.433446196.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.433446196.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.0.hostname", "serviceb.simpleapp.local"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualNode/%s", meshName, vnName)),
				),
			},
			{
				Config: testAccAppmeshVirtualNodeConfig_listenerHealthChecksUpdated(meshName, vnName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualNodeExists(resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, "name", vnName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend.2576932631.virtual_service.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend.2576932631.virtual_service.0.virtual_service_name", "servicec.simpleapp.local"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend.2025248115.virtual_service.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend.2025248115.virtual_service.0.virtual_service_name", "serviced.simpleapp.local"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.3446683576.health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.3446683576.health_check.0.healthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.3446683576.health_check.0.interval_millis", "7000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.3446683576.health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.3446683576.health_check.0.protocol", "tcp"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.3446683576.health_check.0.timeout_millis", "3000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.3446683576.health_check.0.unhealthy_threshold", "9"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.3446683576.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.3446683576.port_mapping.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.3446683576.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.0.hostname", "serviceb1.simpleapp.local"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualNode/%s", meshName, vnName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", meshName, vnName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsAppmeshVirtualNode_logging(t *testing.T) {
	var vn appmesh.VirtualNodeData
	resourceName := "aws_appmesh_virtual_node.test"
	meshName := acctest.RandomWithPrefix("tf-acc-test")
	vnName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshVirtualNodeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshVirtualNodeConfig_logging(meshName, vnName, "/dev/stdout"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualNodeExists(resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, "name", vnName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.0.file.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.0.file.0.path", "/dev/stdout"),
				),
			},
			{
				Config: testAccAppmeshVirtualNodeConfig_logging(meshName, vnName, "/tmp/access.log"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualNodeExists(resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, "name", vnName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.0.file.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.0.file.0.path", "/tmp/access.log"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", meshName, vnName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsAppmeshVirtualNode_tags(t *testing.T) {
	var vn appmesh.VirtualNodeData
	resourceName := "aws_appmesh_virtual_node.test"
	meshName := acctest.RandomWithPrefix("tf-acc-test")
	vnName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshVirtualNodeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshVirtualNodeConfig_tags(meshName, vnName, "foo", "bar", "good", "bad"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualNodeExists(resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.good", "bad"),
				),
			},
			{
				Config: testAccAppmeshVirtualNodeConfig_tags(meshName, vnName, "foo2", "bar", "good", "bad2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualNodeExists(resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo2", "bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.good", "bad2"),
				),
			},
			{
				Config: testAccAppmeshVirtualNodeConfig_basic(meshName, vnName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualNodeExists(resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", meshName, vnName),
				ImportState:       true,
				ImportStateVerify: true,
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

func testAccAppmeshVirtualNodeConfig_mesh(rName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}
`, rName)
}

func testAccAppmeshVirtualNodeConfig_basic(meshName, vnName string) string {
	return testAccAppmeshVirtualNodeConfig_mesh(meshName) + fmt.Sprintf(`
resource "aws_appmesh_virtual_node" "test" {
  name      = %[1]q
  mesh_name = "${aws_appmesh_mesh.test.id}"

  spec {}
}
`, vnName)
}

func testAccAppmeshVirtualNodeConfig_cloudMapServiceDiscovery(meshName, vnName, rName, attrKey, attrValue string) string {
	return testAccAppmeshVirtualNodeConfig_mesh(meshName) + fmt.Sprintf(`
resource "aws_service_discovery_http_namespace" "test" {
  name = %[2]q
}

resource "aws_appmesh_virtual_node" "test" {
  name      = %[1]q
  mesh_name = "${aws_appmesh_mesh.test.id}"

  spec {
    backend {
      virtual_service {
        virtual_service_name = "servicea.simpleapp.local"
      }
    }

    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }

    service_discovery {
      aws_cloud_map {
        attributes = {
          %[3]s = %[4]q
        }

        service_name   = %[2]q
        namespace_name = "${aws_service_discovery_http_namespace.test.name}"
      }
    }
  }
}
`, vnName, rName, attrKey, attrValue)
}

func testAccAppmeshVirtualNodeConfig_listenerHealthChecks(meshName, vnName string) string {
	return testAccAppmeshVirtualNodeConfig_mesh(meshName) + fmt.Sprintf(`
resource "aws_appmesh_virtual_node" "test" {
  name      = %[1]q
  mesh_name = "${aws_appmesh_mesh.test.id}"

  spec {
    backend {
      virtual_service {
        virtual_service_name = "servicea.simpleapp.local"
      }
    }

    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }

      health_check {
        protocol            = "http"
        path                = "/ping"
        healthy_threshold   = 3
        unhealthy_threshold = 5
        timeout_millis      = 2000
        interval_millis     = 5000
      }
    }

    service_discovery {
      dns {
        hostname = "serviceb.simpleapp.local"
      }
    }
  }
}
`, vnName)
}

func testAccAppmeshVirtualNodeConfig_listenerHealthChecksUpdated(meshName, vnName string) string {
	return testAccAppmeshVirtualNodeConfig_mesh(meshName) + fmt.Sprintf(`
resource "aws_appmesh_virtual_node" "test" {
  name      = %[1]q
  mesh_name = "${aws_appmesh_mesh.test.id}"

  spec {
    backend {
      virtual_service {
        virtual_service_name = "servicec.simpleapp.local"
      }
    }

    backend {
      virtual_service {
        virtual_service_name = "serviced.simpleapp.local"
      }
    }

    listener {
      port_mapping {
        port     = 8081
        protocol = "http"
      }

      health_check {
        protocol            = "tcp"
        port                = 8081
        healthy_threshold   = 4
        unhealthy_threshold = 9
        timeout_millis      = 3000
        interval_millis     = 7000
      }
    }

    service_discovery {
      dns {
        hostname = "serviceb1.simpleapp.local"
      }
    }
  }
}
`, vnName)
}

func testAccAppmeshVirtualNodeConfig_logging(meshName, vnName, path string) string {
	return testAccAppmeshVirtualNodeConfig_mesh(meshName) + fmt.Sprintf(`
resource "aws_appmesh_virtual_node" "test" {
  name      = %[1]q
  mesh_name = "${aws_appmesh_mesh.test.id}"

  spec {
    backend {
      virtual_service {
        virtual_service_name = "servicea.simpleapp.local"
      }
    }

    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }

    logging {
      access_log {
        file {
          path = %[2]q
        }
      }
    }

    service_discovery {
      dns {
        hostname = "serviceb.simpleapp.local"
      }
    }
  }
}
`, vnName, path)
}

func testAccAppmeshVirtualNodeConfig_tags(meshName, vnName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAppmeshVirtualNodeConfig_mesh(meshName) + fmt.Sprintf(`
resource "aws_appmesh_virtual_node" "test" {
  name      = %[1]q
  mesh_name = "${aws_appmesh_mesh.test.id}"

  spec {}

  tags = {
    %[2]s = %[3]q
    %[4]s = %[5]q
  }
}
`, vnName, tagKey1, tagValue1, tagKey2, tagValue2)
}
