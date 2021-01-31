package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/appmesh/finder"
)

func init() {
	resource.AddTestSweepers("aws_appmesh_virtual_gateway", &resource.Sweeper{
		Name: "aws_appmesh_virtual_gateway",
		F:    testSweepAppmeshVirtualGateways,
		Dependencies: []string{
			"aws_appmesh_gateway_route",
		},
	})
}

func testSweepAppmeshVirtualGateways(region string) error {
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

					log.Printf("[INFO] Deleting App Mesh service mesh (%s) virtual gateway: %s", meshName, virtualGatewayName)
					r := resourceAwsAppmeshVirtualGateway()
					d := r.Data(nil)
					d.SetId("????????????????") // ID not used in Delete.
					d.Set("mesh_name", meshName)
					d.Set("name", virtualGatewayName)
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

func testAccAwsAppmeshVirtualGateway_basic(t *testing.T) {
	var v appmesh.VirtualGatewayData
	resourceName := "aws_appmesh_virtual_gateway.test"
	meshName := acctest.RandomWithPrefix("tf-acc-test")
	vgName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appmesh.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshVirtualGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshVirtualGatewayConfig(meshName, vgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					testAccCheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", vgName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					testAccCheckResourceAttrAccountID(resourceName, "resource_owner"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s", meshName, vgName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", meshName, vgName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsAppmeshVirtualGateway_disappears(t *testing.T) {
	var v appmesh.VirtualGatewayData
	resourceName := "aws_appmesh_virtual_gateway.test"
	meshName := acctest.RandomWithPrefix("tf-acc-test")
	vgName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appmesh.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshVirtualGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshVirtualGatewayConfig(meshName, vgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualGatewayExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsAppmeshVirtualGateway(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAwsAppmeshVirtualGateway_BackendDefaults(t *testing.T) {
	var v appmesh.VirtualGatewayData
	resourceName := "aws_appmesh_virtual_gateway.test"
	meshName := acctest.RandomWithPrefix("tf-acc-test")
	vgName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appmesh.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshVirtualGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshVirtualGatewayConfigBackendDefaults(meshName, vgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					testAccCheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", vgName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.enforce", "true"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.ports.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.ports.*", "8443"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.acm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.file.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.file.0.certificate_chain", "/cert_chain.pem"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					testAccCheckResourceAttrAccountID(resourceName, "resource_owner"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s", meshName, vgName)),
				),
			},
			{
				Config: testAccAppmeshVirtualGatewayConfigBackendDefaultsUpdated(meshName, vgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					testAccCheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", vgName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.enforce", "true"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.ports.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.ports.*", "443"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.ports.*", "8443"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.acm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.file.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.file.0.certificate_chain", "/etc/ssl/certs/cert_chain.pem"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					testAccCheckResourceAttrAccountID(resourceName, "resource_owner"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s", meshName, vgName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", meshName, vgName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsAppmeshVirtualGateway_ListenerConnectionPool(t *testing.T) {
	var v appmesh.VirtualGatewayData
	resourceName := "aws_appmesh_virtual_gateway.test"
	meshName := acctest.RandomWithPrefix("tf-acc-test")
	vgName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appmesh.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshVirtualGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshVirtualGatewayConfigListenerConnectionPool(meshName, vgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					testAccCheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", vgName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.0.grpc.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.0.grpc.0.max_requests", "4"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.0.http.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.0.http2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "grpc"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					testAccCheckResourceAttrAccountID(resourceName, "resource_owner"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s", meshName, vgName)),
				),
			},
			{
				Config: testAccAppmeshVirtualGatewayConfigListenerConnectionPoolUpdated(meshName, vgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					testAccCheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", vgName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.0.grpc.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.0.http.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.0.http.0.max_connections", "8"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.0.http.0.max_pending_requests", "16"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.0.http2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					testAccCheckResourceAttrAccountID(resourceName, "resource_owner"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s", meshName, vgName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", meshName, vgName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsAppmeshVirtualGateway_ListenerHealthChecks(t *testing.T) {
	var v appmesh.VirtualGatewayData
	resourceName := "aws_appmesh_virtual_gateway.test"
	meshName := acctest.RandomWithPrefix("tf-acc-test")
	vgName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appmesh.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshVirtualGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshVirtualGatewayConfigListenerHealthChecks(meshName, vgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					testAccCheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", vgName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.interval_millis", "5000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.path", "/ping"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.protocol", "http2"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.timeout_millis", "2000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.unhealthy_threshold", "5"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "grpc"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					testAccCheckResourceAttrAccountID(resourceName, "resource_owner"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s", meshName, vgName)),
				),
			},
			{
				Config: testAccAppmeshVirtualGatewayConfigListenerHealthChecksUpdated(meshName, vgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					testAccCheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", vgName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.healthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.interval_millis", "7000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.path", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.protocol", "grpc"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.timeout_millis", "3000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.unhealthy_threshold", "9"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					testAccCheckResourceAttrAccountID(resourceName, "resource_owner"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s", meshName, vgName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", meshName, vgName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsAppmeshVirtualGateway_ListenerTls(t *testing.T) {
	var v appmesh.VirtualGatewayData
	var ca acmpca.CertificateAuthority
	resourceName := "aws_appmesh_virtual_gateway.test"
	acmCAResourceName := "aws_acmpca_certificate_authority.test"
	acmCertificateResourceName := "aws_acm_certificate.test"
	meshName := acctest.RandomWithPrefix("tf-acc-test")
	vgName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appmesh.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshVirtualGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshVirtualGatewayConfigListenerTlsFile(meshName, vgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					testAccCheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", vgName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.acm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.file.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.file.0.certificate_chain", "/cert_chain.pem"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.file.0.private_key", "/key.pem"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.mode", "PERMISSIVE"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					testAccCheckResourceAttrAccountID(resourceName, "resource_owner"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s", meshName, vgName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", meshName, vgName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			// We need to create and activate the CA before issuing a certificate.
			{
				Config: testAccAppmeshVirtualGatewayConfigRootCA(vgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAcmpcaCertificateAuthorityExists(acmCAResourceName, &ca),
					testAccCheckAwsAcmpcaCertificateAuthorityActivateCA(&ca),
				),
			},
			{
				Config: testAccAppmeshVirtualGatewayConfigListenerTlsAcm(meshName, vgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					testAccCheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", vgName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.acm.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.tls.0.certificate.0.acm.0.certificate_arn", acmCertificateResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.file.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.mode", "STRICT"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					testAccCheckResourceAttrAccountID(resourceName, "resource_owner"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s", meshName, vgName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", meshName, vgName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppmeshVirtualGatewayConfigListenerTlsAcm(meshName, vgName),
				Check: resource.ComposeTestCheckFunc(
					// CA must be DISABLED for deletion.
					testAccCheckAwsAcmpcaCertificateAuthorityDisableCA(&ca),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAwsAppmeshVirtualGateway_Logging(t *testing.T) {
	var v appmesh.VirtualGatewayData
	resourceName := "aws_appmesh_virtual_gateway.test"
	meshName := acctest.RandomWithPrefix("tf-acc-test")
	vgName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appmesh.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshVirtualGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshVirtualGatewayConfigLogging(meshName, vgName, "/dev/stdout"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					testAccCheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", vgName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.0.file.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.0.file.0.path", "/dev/stdout"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					testAccCheckResourceAttrAccountID(resourceName, "resource_owner"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s", meshName, vgName)),
				),
			},
			{
				Config: testAccAppmeshVirtualGatewayConfigLogging(meshName, vgName, "/tmp/access.log"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					testAccCheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", vgName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.0.file.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.0.file.0.path", "/tmp/access.log"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					testAccCheckResourceAttrAccountID(resourceName, "resource_owner"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s", meshName, vgName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", meshName, vgName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsAppmeshVirtualGateway_Tags(t *testing.T) {
	var v appmesh.VirtualGatewayData
	resourceName := "aws_appmesh_virtual_gateway.test"
	meshName := acctest.RandomWithPrefix("tf-acc-test")
	vgName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appmesh.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppmeshVirtualGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppmeshVirtualGatewayConfigTags1(meshName, vgName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", meshName, vgName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppmeshVirtualGatewayConfigTags2(meshName, vgName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAppmeshVirtualGatewayConfigTags1(meshName, vgName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppmeshVirtualGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAppmeshVirtualGatewayDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).appmeshconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appmesh_virtual_node" {
			continue
		}

		_, err := finder.VirtualGateway(conn, rs.Primary.Attributes["mesh_name"], rs.Primary.Attributes["name"], rs.Primary.Attributes["mesh_owner"])
		if isAWSErr(err, appmesh.ErrCodeNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("App Mesh virtual gateway still exists: %s", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAppmeshVirtualGatewayExists(name string, v *appmesh.VirtualGatewayData) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).appmeshconn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No App Mesh virtual gateway ID is set")
		}

		out, err := finder.VirtualGateway(conn, rs.Primary.Attributes["mesh_name"], rs.Primary.Attributes["name"], rs.Primary.Attributes["mesh_owner"])
		if err != nil {
			return err
		}

		*v = *out

		return nil
	}
}

func testAccAppmeshVirtualGatewayConfig(meshName, vgName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_gateway" "test" {
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
`, meshName, vgName)
}

func testAccAppmeshVirtualGatewayConfigBackendDefaults(meshName, vgName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_gateway" "test" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }

    backend_defaults {
      client_policy {
        tls {
          ports = [8443]

          validation {
            trust {
              file {
                certificate_chain = "/cert_chain.pem"
              }
            }
          }
        }
      }
    }
  }
}
`, meshName, vgName)
}

func testAccAppmeshVirtualGatewayConfigBackendDefaultsUpdated(meshName, vgName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_gateway" "test" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }

    backend_defaults {
      client_policy {
        tls {
          ports = [443, 8443]

          validation {
            trust {
              file {
                certificate_chain = "/etc/ssl/certs/cert_chain.pem"
              }
            }
          }
        }
      }
    }
  }
}
`, meshName, vgName)
}

func testAccAppmeshVirtualGatewayConfigListenerConnectionPool(meshName, vgName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_gateway" "test" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "grpc"
      }

      connection_pool {
        grpc {
          max_requests = 4
        }
      }
    }
  }
}
`, meshName, vgName)
}

func testAccAppmeshVirtualGatewayConfigListenerConnectionPoolUpdated(meshName, vgName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_gateway" "test" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8081
        protocol = "http"
      }

      connection_pool {
        http {
          max_connections      = 8
          max_pending_requests = 16
        }
      }
    }
  }
}
`, meshName, vgName)
}

func testAccAppmeshVirtualGatewayConfigListenerHealthChecks(meshName, vgName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_gateway" "test" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "grpc"
      }

      health_check {
        protocol            = "http2"
        path                = "/ping"
        healthy_threshold   = 3
        unhealthy_threshold = 5
        timeout_millis      = 2000
        interval_millis     = 5000
      }
    }
  }
}
`, meshName, vgName)
}

func testAccAppmeshVirtualGatewayConfigListenerHealthChecksUpdated(meshName, vgName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_gateway" "test" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8081
        protocol = "http"
      }

      health_check {
        protocol            = "grpc"
        port                = 8081
        healthy_threshold   = 4
        unhealthy_threshold = 9
        timeout_millis      = 3000
        interval_millis     = 7000
      }
    }
  }
}
`, meshName, vgName)
}

func testAccAppmeshVirtualGatewayConfigRootCA(rName string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  type                            = "ROOT"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "%[1]s.com"
    }
  }
}
`, rName)
}

func testAccAppmeshVirtualGatewayConfigListenerTlsAcm(meshName, vgName string) string {
	return composeConfig(
		testAccAppmeshVirtualGatewayConfigRootCA(vgName),
		fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_acm_certificate" "test" {
  domain_name               = "test.%[2]s.com"
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn
}

resource "aws_appmesh_virtual_gateway" "test" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }

      tls {
        certificate {
          acm {
            certificate_arn = aws_acm_certificate.test.arn
          }
        }

        mode = "STRICT"
      }
    }
  }
}
`, meshName, vgName))
}

func testAccAppmeshVirtualGatewayConfigListenerTlsFile(meshName, vgName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_gateway" "test" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }

      tls {
        certificate {
          file {
            certificate_chain = "/cert_chain.pem"
            private_key       = "/key.pem"
          }
        }

        mode = "PERMISSIVE"
      }
    }
  }
}
`, meshName, vgName)
}

func testAccAppmeshVirtualGatewayConfigLogging(meshName, vgName, path string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_gateway" "test" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }

    logging {
      access_log {
        file {
          path = %[3]q
        }
      }
    }
  }
}
`, meshName, vgName, path)
}

func testAccAppmeshVirtualGatewayConfigTags1(meshName, vgName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_gateway" "test" {
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
    %[3]q = %[4]q
  }
}
`, meshName, vgName, tagKey1, tagValue1)
}

func testAccAppmeshVirtualGatewayConfigTags2(meshName, vgName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_gateway" "test" {
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
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, meshName, vgName, tagKey1, tagValue1, tagKey2, tagValue2)
}
