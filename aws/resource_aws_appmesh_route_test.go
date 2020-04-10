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

func testAccAwsAppmeshRoute_httpRoute(t *testing.T) {
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := acctest.RandomWithPrefix("tf-acc-test")
	vrName := acctest.RandomWithPrefix("tf-acc-test")
	vn1Name := acctest.RandomWithPrefix("tf-acc-test")
	vn2Name := acctest.RandomWithPrefix("tf-acc-test")
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshRouteConfig_httpRoute(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.scheme", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccAppmeshRouteConfig_httpRouteUpdated(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/path"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.scheme", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAwsAppmeshRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsAppmeshRoute_tcpRoute(t *testing.T) {
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := acctest.RandomWithPrefix("tf-acc-test")
	vrName := acctest.RandomWithPrefix("tf-acc-test")
	vn1Name := acctest.RandomWithPrefix("tf-acc-test")
	vn2Name := acctest.RandomWithPrefix("tf-acc-test")
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshRouteConfig_tcpRoute(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.action.0.weighted_target.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccAppmeshRouteConfig_tcpRouteUpdated(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.action.0.weighted_target.#", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAwsAppmeshRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsAppmeshRoute_tags(t *testing.T) {
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := acctest.RandomWithPrefix("tf-acc-test")
	vrName := acctest.RandomWithPrefix("tf-acc-test")
	vn1Name := acctest.RandomWithPrefix("tf-acc-test")
	vn2Name := acctest.RandomWithPrefix("tf-acc-test")
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshRouteConfigWithTags(meshName, vrName, vn1Name, vn2Name, rName, "foo", "bar", "good", "bad"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.good", "bad"),
				),
			},
			{
				Config: testAccAppmeshRouteConfigWithTags(meshName, vrName, vn1Name, vn2Name, rName, "foo2", "bar", "good", "bad2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.foo2", "bar"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.good", "bad2"),
				),
			},
			{
				Config: testAccAppmeshRouteConfig_httpRoute(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAwsAppmeshRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsAppmeshRoute_httpHeader(t *testing.T) {
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := acctest.RandomWithPrefix("tf-acc-test")
	vrName := acctest.RandomWithPrefix("tf-acc-test")
	vn1Name := acctest.RandomWithPrefix("tf-acc-test")
	vn2Name := acctest.RandomWithPrefix("tf-acc-test")
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppmeshRouteConfig_httpHeader(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.2366200004.invert", "false"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.2366200004.match.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.2366200004.name", "X-Testing1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", "POST"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.scheme", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccAwsAppmeshRouteConfig_httpHeaderUpdated(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.2971741486.invert", "true"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.2971741486.match.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.2971741486.name", "X-Testing1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.3147536248.invert", "false"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.3147536248.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.3147536248.match.0.exact", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.3147536248.match.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.3147536248.match.0.range.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.3147536248.match.0.range.0.end", "7"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.3147536248.match.0.range.0.start", "2"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.3147536248.match.0.regex", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.3147536248.match.0.suffix", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.3147536248.name", "X-Testing2"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", "PUT"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/path"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.scheme", "https"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAwsAppmeshRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsAppmeshRoute_routePriority(t *testing.T) {
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := acctest.RandomWithPrefix("tf-acc-test")
	vrName := acctest.RandomWithPrefix("tf-acc-test")
	vn1Name := acctest.RandomWithPrefix("tf-acc-test")
	vn2Name := acctest.RandomWithPrefix("tf-acc-test")
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppmeshRouteConfig_routePriority(meshName, vrName, vn1Name, vn2Name, rName, 42),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.scheme", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", "42"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccAwsAppmeshRouteConfig_routePriority(meshName, vrName, vn1Name, vn2Name, rName, 1000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.scheme", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", "1000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAwsAppmeshRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsAppmeshRouteImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s/%s", rs.Primary.Attributes["mesh_name"], rs.Primary.Attributes["virtual_router_name"], rs.Primary.Attributes["name"]), nil
	}
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

func testAccAppmeshRouteConfigBase(meshName, vrName, vn1Name, vn2Name string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_router" "test" {
  name      = %[2]q
  mesh_name = "${aws_appmesh_mesh.test.id}"

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }
  }
}

resource "aws_appmesh_virtual_node" "foo" {
  name      = %[3]q
  mesh_name = "${aws_appmesh_mesh.test.id}"

  spec {}
}

resource "aws_appmesh_virtual_node" "bar" {
  name      = %[4]q
  mesh_name = "${aws_appmesh_mesh.test.id}"

  spec {}
}
`, meshName, vrName, vn1Name, vn2Name)
}

func testAccAppmeshRouteConfig_httpRoute(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return testAccAppmeshRouteConfigBase(meshName, vrName, vn1Name, vn2Name) + fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = "${aws_appmesh_mesh.test.id}"
  virtual_router_name = "${aws_appmesh_virtual_router.test.name}"

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
`, rName)
}

func testAccAppmeshRouteConfig_httpRouteUpdated(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return testAccAppmeshRouteConfigBase(meshName, vrName, vn1Name, vn2Name) + fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = "${aws_appmesh_mesh.test.id}"
  virtual_router_name = "${aws_appmesh_virtual_router.test.name}"

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
`, rName)
}

func testAccAppmeshRouteConfig_tcpRoute(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return testAccAppmeshRouteConfigBase(meshName, vrName, vn1Name, vn2Name) + fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = "${aws_appmesh_mesh.test.id}"
  virtual_router_name = "${aws_appmesh_virtual_router.test.name}"

  spec {
    tcp_route {
      action {
        weighted_target {
          virtual_node = "${aws_appmesh_virtual_node.foo.name}"
          weight       = 100
        }
      }
    }
  }
}
`, rName)
}

func testAccAppmeshRouteConfig_tcpRouteUpdated(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return testAccAppmeshRouteConfigBase(meshName, vrName, vn1Name, vn2Name) + fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = "${aws_appmesh_mesh.test.id}"
  virtual_router_name = "${aws_appmesh_virtual_router.test.name}"

  spec {
    tcp_route {
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
`, rName)
}

func testAccAppmeshRouteConfigWithTags(meshName, vrName, vn1Name, vn2Name, rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAppmeshRouteConfigBase(meshName, vrName, vn1Name, vn2Name) + fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = "${aws_appmesh_mesh.test.id}"
  virtual_router_name = "${aws_appmesh_virtual_router.test.name}"

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

  tags = {
    %[2]s = %[3]q
    %[4]s = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAwsAppmeshRouteConfig_httpHeader(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return testAccAppmeshRouteConfigBase(meshName, vrName, vn1Name, vn2Name) + fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = "${aws_appmesh_mesh.test.id}"
  virtual_router_name = "${aws_appmesh_virtual_router.test.name}"

  spec {
    http_route {
      match {
        prefix = "/"
        method = "POST"
        scheme = "http"

        header {
          name = "X-Testing1"
        }
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
`, rName)
}

func testAccAwsAppmeshRouteConfig_httpHeaderUpdated(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return testAccAppmeshRouteConfigBase(meshName, vrName, vn1Name, vn2Name) + fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = "${aws_appmesh_mesh.test.id}"
  virtual_router_name = "${aws_appmesh_virtual_router.test.name}"

  spec {
    http_route {
      match {
        prefix = "/path"
        method = "PUT"
        scheme = "https"

        header {
          name   = "X-Testing1"
          invert = true
        }

        header {
          name   = "X-Testing2"
          invert = false

          match {
            range {
              start = 2
              end   = 7
            }
          }
        }
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
`, rName)
}

func testAccAwsAppmeshRouteConfig_routePriority(meshName, vrName, vn1Name, vn2Name, rName string, priority int) string {
	return testAccAppmeshRouteConfigBase(meshName, vrName, vn1Name, vn2Name) + fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = "${aws_appmesh_mesh.test.id}"
  virtual_router_name = "${aws_appmesh_virtual_router.test.name}"

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

    priority = %[2]d
  }
}
`, rName, priority)
}
