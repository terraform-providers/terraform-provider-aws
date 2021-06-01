package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/appmesh/finder"
)

func init() {
	resource.AddTestSweepers("aws_appmesh_gateway_route", &resource.Sweeper{
		Name: "aws_appmesh_gateway_route",
		F:    testSweepAppmeshGatewayRoutes,
	})
}

func testSweepAppmeshGatewayRoutes(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).appmeshconn

	var sweeperErrs *multierror.Error

	err = conn.ListMeshesPages(&appmesh.ListMeshesInput{}, func(page *appmesh.ListMeshesOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, mesh := range page.Meshes {
			meshName := aws.StringValue(mesh.MeshName)

			err = conn.ListVirtualGatewaysPages(&appmesh.ListVirtualGatewaysInput{MeshName: mesh.MeshName}, func(page *appmesh.ListVirtualGatewaysOutput, isLast bool) bool {
				if page == nil {
					return !isLast
				}

				for _, virtualGateway := range page.VirtualGateways {
					virtualGatewayName := aws.StringValue(virtualGateway.VirtualGatewayName)

					err = conn.ListGatewayRoutesPages(&appmesh.ListGatewayRoutesInput{MeshName: mesh.MeshName, VirtualGatewayName: virtualGateway.VirtualGatewayName}, func(page *appmesh.ListGatewayRoutesOutput, isLast bool) bool {
						if page == nil {
							return !isLast
						}

						for _, gatewayRoute := range page.GatewayRoutes {
							gatewayRouteName := aws.StringValue(gatewayRoute.GatewayRouteName)

							log.Printf("[INFO] Deleting App Mesh service mesh (%s) virtual gateway (%s) gateway route: %s", meshName, virtualGatewayName, gatewayRouteName)
							r := resourceAwsAppmeshGatewayRoute()
							d := r.Data(nil)
							d.SetId("????????????????") // ID not used in Delete.
							d.Set("mesh_name", meshName)
							d.Set("name", gatewayRouteName)
							d.Set("virtual_gateway_name", virtualGatewayName)
							err := r.Delete(d, client)

							if err != nil {
								log.Printf("[ERROR] %s", err)
								sweeperErrs = multierror.Append(sweeperErrs, err)
								continue
							}
						}

						return !isLast
					})

					if err != nil {
						sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving App Mesh service mesh (%s) virtual gateway (%s) gateway routes: %w", meshName, virtualGatewayName, err))
					}
				}

				return !isLast
			})

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving App Mesh service mesh (%s) virtual gateways: %w", meshName, err))
			}
		}

		return !isLast
	})
	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Appmesh virtual gateway sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving App Mesh virtual gateways: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func testAccAwsAppmeshGatewayRoute_basic(t *testing.T) {
	var v appmesh.GatewayRouteData
	resourceName := "aws_appmesh_gateway_route.test"
	vsResourceName := "aws_appmesh_virtual_service.test.0"
	meshName := acctest.RandomWithPrefix("tf-acc-test")
	vgName := acctest.RandomWithPrefix("tf-acc-test")
	grName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appmesh.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshGatewayRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshGatewayRouteConfigHttpRoute(meshName, vgName, grName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshGatewayRouteExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					testAccCheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vsResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					testAccCheckResourceAttrAccountID(resourceName, "resource_owner"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAwsAppmeshGatewayRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsAppmeshGatewayRoute_disappears(t *testing.T) {
	var v appmesh.GatewayRouteData
	resourceName := "aws_appmesh_gateway_route.test"
	meshName := acctest.RandomWithPrefix("tf-acc-test")
	vgName := acctest.RandomWithPrefix("tf-acc-test")
	grName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appmesh.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshGatewayRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshGatewayRouteConfigHttpRoute(meshName, vgName, grName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshGatewayRouteExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsAppmeshGatewayRoute(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAwsAppmeshGatewayRoute_GrpcRoute(t *testing.T) {
	var v appmesh.GatewayRouteData
	resourceName := "aws_appmesh_gateway_route.test"
	vs1ResourceName := "aws_appmesh_virtual_service.test.0"
	vs2ResourceName := "aws_appmesh_virtual_service.test.1"
	meshName := acctest.RandomWithPrefix("tf-acc-test")
	vgName := acctest.RandomWithPrefix("tf-acc-test")
	grName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appmesh.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshGatewayRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshGatewayRouteConfigGrpcRoute(meshName, vgName, grName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshGatewayRouteExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					testAccCheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.target.0.virtual_service.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.grpc_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs1ResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.service_name", "test1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					testAccCheckResourceAttrAccountID(resourceName, "resource_owner"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
				),
			},
			{
				Config: testAccAppmeshGatewayRouteConfigGrpcRouteUpdated(meshName, vgName, grName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshGatewayRouteExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					testAccCheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.target.0.virtual_service.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.grpc_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs2ResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.service_name", "test2"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					testAccCheckResourceAttrAccountID(resourceName, "resource_owner"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAwsAppmeshGatewayRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsAppmeshGatewayRoute_HttpRoute(t *testing.T) {
	var v appmesh.GatewayRouteData
	resourceName := "aws_appmesh_gateway_route.test"
	vs1ResourceName := "aws_appmesh_virtual_service.test.0"
	vs2ResourceName := "aws_appmesh_virtual_service.test.1"
	meshName := acctest.RandomWithPrefix("tf-acc-test")
	vgName := acctest.RandomWithPrefix("tf-acc-test")
	grName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appmesh.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshGatewayRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshGatewayRouteConfigHttpRoute(meshName, vgName, grName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshGatewayRouteExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					testAccCheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs1ResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					testAccCheckResourceAttrAccountID(resourceName, "resource_owner"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
				),
			},
			{
				Config: testAccAppmeshGatewayRouteConfigHttpRouteUpdated(meshName, vgName, grName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshGatewayRouteExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					testAccCheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs2ResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/users"),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					testAccCheckResourceAttrAccountID(resourceName, "resource_owner"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAwsAppmeshGatewayRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsAppmeshGatewayRoute_Http2Route(t *testing.T) {
	var v appmesh.GatewayRouteData
	resourceName := "aws_appmesh_gateway_route.test"
	vs1ResourceName := "aws_appmesh_virtual_service.test.0"
	vs2ResourceName := "aws_appmesh_virtual_service.test.1"
	meshName := acctest.RandomWithPrefix("tf-acc-test")
	vgName := acctest.RandomWithPrefix("tf-acc-test")
	grName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appmesh.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshGatewayRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshGatewayRouteConfigHttp2Route(meshName, vgName, grName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshGatewayRouteExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					testAccCheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs1ResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					testAccCheckResourceAttrAccountID(resourceName, "resource_owner"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
				),
			},
			{
				Config: testAccAppmeshGatewayRouteConfigHttp2RouteUpdated(meshName, vgName, grName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshGatewayRouteExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					testAccCheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs2ResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.prefix", "/users"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					testAccCheckResourceAttrAccountID(resourceName, "resource_owner"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAwsAppmeshGatewayRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsAppmeshGatewayRoute_Tags(t *testing.T) {
	var v appmesh.GatewayRouteData
	resourceName := "aws_appmesh_gateway_route.test"
	meshName := acctest.RandomWithPrefix("tf-acc-test")
	vgName := acctest.RandomWithPrefix("tf-acc-test")
	grName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appmesh.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshGatewayRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshGatewayRouteConfigTags1(meshName, vgName, grName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshGatewayRouteExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAwsAppmeshGatewayRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppmeshGatewayRouteConfigTags2(meshName, vgName, grName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshGatewayRouteExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAppmeshGatewayRouteConfigTags1(meshName, vgName, grName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshGatewayRouteExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccAwsAppmeshGatewayRouteImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s/%s", rs.Primary.Attributes["mesh_name"], rs.Primary.Attributes["virtual_gateway_name"], rs.Primary.Attributes["name"]), nil
	}
}

func testAccCheckAppmeshGatewayRouteDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).appmeshconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appmesh_gateway_route" {
			continue
		}

		_, err := finder.GatewayRoute(conn, rs.Primary.Attributes["mesh_name"], rs.Primary.Attributes["virtual_gateway_name"], rs.Primary.Attributes["name"], rs.Primary.Attributes["mesh_owner"])
		if isAWSErr(err, appmesh.ErrCodeNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("App Mesh gateway route still exists: %s", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAppmeshGatewayRouteExists(name string, v *appmesh.GatewayRouteData) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).appmeshconn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No App Mesh gateway route ID is set")
		}

		out, err := finder.GatewayRoute(conn, rs.Primary.Attributes["mesh_name"], rs.Primary.Attributes["virtual_gateway_name"], rs.Primary.Attributes["name"], rs.Primary.Attributes["mesh_owner"])
		if err != nil {
			return err
		}

		*v = *out

		return nil
	}
}

func testAccAppmeshGatewayRouteConfigBase(meshName, vgName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_gateway" "test" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.name

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
  count = 2

  name      = "%[2]s-${count.index}"
  mesh_name = aws_appmesh_mesh.test.name

  spec {}
}
`, meshName, vgName)
}

func testAccAppmeshGatewayRouteConfigGrpcRoute(meshName, vgName, grName string) string {
	return composeConfig(testAccAppmeshGatewayRouteConfigBase(meshName, vgName), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    grpc_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[0].name
          }
        }
      }

      match {
        service_name = "test1"
      }
    }
  }
}
`, grName))
}

func testAccAppmeshGatewayRouteConfigGrpcRouteUpdated(meshName, vgName, grName string) string {
	return composeConfig(testAccAppmeshGatewayRouteConfigBase(meshName, vgName), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    grpc_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[1].name
          }
        }
      }

      match {
        service_name = "test2"
      }
    }
  }
}
`, grName))
}

func testAccAppmeshGatewayRouteConfigHttpRoute(meshName, vgName, grName string) string {
	return composeConfig(testAccAppmeshGatewayRouteConfigBase(meshName, vgName), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    http_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[0].name
          }
        }
      }

      match {
        prefix = "/"
      }
    }
  }
}
`, grName))
}

func testAccAppmeshGatewayRouteConfigHttpRouteUpdated(meshName, vgName, grName string) string {
	return composeConfig(testAccAppmeshGatewayRouteConfigBase(meshName, vgName), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    http_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[1].name
          }
        }
      }

      match {
        prefix = "/users"
      }
    }
  }
}
`, grName))
}

func testAccAppmeshGatewayRouteConfigHttp2Route(meshName, vgName, grName string) string {
	return composeConfig(testAccAppmeshGatewayRouteConfigBase(meshName, vgName), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    http2_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[0].name
          }
        }
      }

      match {
        prefix = "/"
      }
    }
  }
}
`, grName))
}

func testAccAppmeshGatewayRouteConfigHttp2RouteUpdated(meshName, vgName, grName string) string {
	return composeConfig(testAccAppmeshGatewayRouteConfigBase(meshName, vgName), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    http2_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[1].name
          }
        }
      }

      match {
        prefix = "/users"
      }
    }
  }
}
`, grName))
}

func testAccAppmeshGatewayRouteConfigTags1(meshName, vgName, grName, tagKey1, tagValue1 string) string {
	return composeConfig(testAccAppmeshGatewayRouteConfigBase(meshName, vgName), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    http_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[0].name
          }
        }
      }

      match {
        prefix = "/"
      }
    }
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, grName, tagKey1, tagValue1))
}

func testAccAppmeshGatewayRouteConfigTags2(meshName, vgName, grName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return composeConfig(testAccAppmeshGatewayRouteConfigBase(meshName, vgName), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    http_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[0].name
          }
        }
      }

      match {
        prefix = "/"
      }
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, grName, tagKey1, tagValue1, tagKey2, tagValue2))
}
