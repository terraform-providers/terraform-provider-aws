package aws

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/experimental/sync"
)

const clientVpnEndpointDefaultLimit = 5

var testAccEc2ClientVpnEndpointSemaphore sync.Semaphore

func init() {
	testAccEc2ClientVpnEndpointSemaphore = sync.InitializeSemaphore("AWS_EC2_CLIENT_VPN_LIMIT", clientVpnEndpointDefaultLimit)
}

func init() {
	resource.AddTestSweepers("aws_ec2_client_vpn_endpoint", &resource.Sweeper{
		Name: "aws_ec2_client_vpn_endpoint",
		F:    testSweepEc2ClientVpnEndpoints,
		Dependencies: []string{
			"aws_directory_service_directory",
			"aws_ec2_client_vpn_network_association",
		},
	})
}

func testSweepEc2ClientVpnEndpoints(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*AWSClient).ec2conn

	var sweeperErrs *multierror.Error

	input := &ec2.DescribeClientVpnEndpointsInput{}
	err = conn.DescribeClientVpnEndpointsPages(input, func(page *ec2.DescribeClientVpnEndpointsOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, clientVpnEndpoint := range page.ClientVpnEndpoints {
			id := aws.StringValue(clientVpnEndpoint.ClientVpnEndpointId)
			log.Printf("[INFO] Deleting Client VPN endpoint: %s", id)
			err := deleteClientVpnEndpoint(conn, id)
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Client VPN endpoint (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !isLast
	})
	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Client VPN endpoint sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Client VPN endpoints: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

// This is part of an experimental feature, do not use this as a starting point for tests
//   "This place is not a place of honor... no highly esteemed deed is commemorated here... nothing valued is here.
//   What is here was dangerous and repulsive to us. This message is a warning about danger."
//   --  https://hyperallergic.com/312318/a-nuclear-warning-designed-to-last-10000-years/
func TestAccAwsEc2ClientVpn_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Endpoint": {
			"basic":             testAccAwsEc2ClientVpnEndpoint_basic,
			"disappears":        testAccAwsEc2ClientVpnEndpoint_disappears,
			"msAD":              testAccAwsEc2ClientVpnEndpoint_msAD,
			"mutualAuthAndMsAD": testAccAwsEc2ClientVpnEndpoint_mutualAuthAndMsAD,
			"federated":         testAccAwsEc2ClientVpnEndpoint_federated,
			"withLogGroup":      testAccAwsEc2ClientVpnEndpoint_withLogGroup,
			"withDNSServers":    testAccAwsEc2ClientVpnEndpoint_withDNSServers,
			"tags":              testAccAwsEc2ClientVpnEndpoint_tags,
			"splitTunnel":       testAccAwsEc2ClientVpnEndpoint_splitTunnel,
		},
		"AuthorizationRule": {
			"basic":      testAccAwsEc2ClientVpnAuthorizationRule_basic,
			"groups":     testAccAwsEc2ClientVpnAuthorizationRule_groups,
			"Subnets":    testAccAwsEc2ClientVpnAuthorizationRule_Subnets,
			"disappears": testAccAwsEc2ClientVpnAuthorizationRule_disappears,
		},
		"NetworkAssociation": {
			"basic":          testAccAwsEc2ClientVpnNetworkAssociation_basic,
			"disappears":     testAccAwsEc2ClientVpnNetworkAssociation_disappears,
			"securityGroups": testAccAwsEc2ClientVpnNetworkAssociation_securityGroups,
		},
		"Route": {
			"basic":       testAccAwsEc2ClientVpnRoute_basic,
			"description": testAccAwsEc2ClientVpnRoute_description,
			"disappears":  testAccAwsEc2ClientVpnRoute_disappears,
		},
	}

	t.Parallel()
	for group, m := range testCases {
		m := m
		for name, tc := range m {
			tc := tc
			t.Run(fmt.Sprintf("%s_%s", group, name), func(t *testing.T) {
				t.Cleanup(func() {
					if os.Getenv(resource.TestEnvVar) != "" {
						testAccEc2ClientVpnEndpointSemaphore.Notify()
					}
				})
				tc(t)
			})
		}
	}
}

func testAccAwsEc2ClientVpnEndpoint_basic(t *testing.T) {
	var v ec2.ClientVpnEndpoint
	rStr := acctest.RandString(5)
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ClientVpnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfig(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`client-vpn-endpoint/cvpn-endpoint-.+`)),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.0.type", "certificate-authentication"),
					resource.TestCheckResourceAttr(resourceName, "status", ec2.ClientVpnEndpointStatusCodePendingAssociate),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsEc2ClientVpnEndpoint_disappears(t *testing.T) {
	var v ec2.ClientVpnEndpoint
	rStr := acctest.RandString(5)
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ClientVpnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfig(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEc2ClientVpnEndpoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAwsEc2ClientVpnEndpoint_msAD(t *testing.T) {
	var v ec2.ClientVpnEndpoint
	rStr := acctest.RandString(5)
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ClientVpnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfigWithMicrosoftAD(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.0.type", "directory-service-authentication"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsEc2ClientVpnEndpoint_mutualAuthAndMsAD(t *testing.T) {
	var v ec2.ClientVpnEndpoint
	rStr := acctest.RandString(5)
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ClientVpnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfigWithMutualAuthAndMicrosoftAD(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.0.type", "directory-service-authentication"),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.1.type", "certificate-authentication"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsEc2ClientVpnEndpoint_federated(t *testing.T) {
	var v ec2.ClientVpnEndpoint
	rStr := acctest.RandString(5)
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckClientVPNSyncronize(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ClientVpnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfigWithFederatedAuth(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.0.type", "federated-authentication"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsEc2ClientVpnEndpoint_withLogGroup(t *testing.T) {
	var v1, v2 ec2.ClientVpnEndpoint
	rStr := acctest.RandString(5)
	resourceName := "aws_ec2_client_vpn_endpoint.test"
	logGroupResourceName := "aws_cloudwatch_log_group.lg"
	logStreamResourceName := "aws_cloudwatch_log_stream.ls"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ClientVpnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfig(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName, &v1),
				),
			},
			{
				Config: testAccEc2ClientVpnEndpointConfigWithLogGroup(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.0.enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "connection_log_options.0.cloudwatch_log_group", logGroupResourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "connection_log_options.0.cloudwatch_log_stream", logStreamResourceName, "name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsEc2ClientVpnEndpoint_withDNSServers(t *testing.T) {
	var v1, v2 ec2.ClientVpnEndpoint
	rStr := acctest.RandString(5)
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ClientVpnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfig(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName, &v1),
				),
			},
			{
				Config: testAccEc2ClientVpnEndpointConfigWithDNSServers(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName, &v2),
				),
			},
		},
	})
}

func testAccAwsEc2ClientVpnEndpoint_tags(t *testing.T) {
	var v1, v2, v3 ec2.ClientVpnEndpoint
	resourceName := "aws_ec2_client_vpn_endpoint.test"
	rStr := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ClientVpnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfig_tags(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "original"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEc2ClientVpnEndpointConfig_tagsChanged(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "changed"),
				),
			},
			{
				Config: testAccEc2ClientVpnEndpointConfig(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName, &v3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccAwsEc2ClientVpnEndpoint_splitTunnel(t *testing.T) {
	var v1, v2 ec2.ClientVpnEndpoint
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ClientVpnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfigSplitTunnel(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "split_tunnel", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEc2ClientVpnEndpointConfigSplitTunnel(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "split_tunnel", "false"),
				),
			},
		},
	})
}

func testAccPreCheckClientVPNSyncronize(t *testing.T) {
	sync.TestAccPreCheckSyncronize(t, testAccEc2ClientVpnEndpointSemaphore, "Client VPN")
}

func testAccCheckAwsEc2ClientVpnEndpointDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_client_vpn_endpoint" {
			continue
		}

		input := &ec2.DescribeClientVpnEndpointsInput{
			ClientVpnEndpointIds: aws.StringSlice([]string{rs.Primary.ID}),
		}

		resp, _ := conn.DescribeClientVpnEndpoints(input)
		for _, v := range resp.ClientVpnEndpoints {
			if aws.StringValue(v.ClientVpnEndpointId) == rs.Primary.ID {
				return fmt.Errorf("Client VPN endpoint (%s) not deleted", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccCheckAwsEc2ClientVpnEndpointExists(name string, endpoint *ec2.ClientVpnEndpoint) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		input := &ec2.DescribeClientVpnEndpointsInput{
			ClientVpnEndpointIds: aws.StringSlice([]string{rs.Primary.ID}),
		}
		result, err := conn.DescribeClientVpnEndpoints(input)
		if err != nil {
			return err
		}

		if result == nil || len(result.ClientVpnEndpoints) == 0 || result.ClientVpnEndpoints[0] == nil {
			return fmt.Errorf("EC2 Client VPN Endpoint (%s) not found", rs.Primary.ID)
		}

		*endpoint = *result.ClientVpnEndpoints[0]

		return nil
	}
}

func testAccEc2ClientVpnEndpointConfigAcmCertificateBase() string {
	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, "example.com")

	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[1]s"
  private_key      = "%[2]s"
}
`, tlsPemEscapeNewlines(certificate), tlsPemEscapeNewlines(key))
}

func testAccEc2ClientVpnEndpointMsADBase() string {
	return `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  count             = 2
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id            = aws_vpc.test.id
}

resource "aws_directory_service_directory" "test" {
  name     = "vpn.notexample.com"
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}
`
}

func testAccEc2ClientVpnEndpointConfig(rName string) string {
	return testAccEc2ClientVpnEndpointConfigAcmCertificateBase() + fmt.Sprintf(`
resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = aws_acm_certificate.test.arn
  }

  connection_log_options {
    enabled = false
  }
}
`, rName)
}

func testAccEc2ClientVpnEndpointConfigWithLogGroup(rName string) string {
	return testAccEc2ClientVpnEndpointConfigAcmCertificateBase() + fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "lg" {
  name = "terraform-testacc-clientvpn-loggroup-%s"
}

resource "aws_cloudwatch_log_stream" "ls" {
  name           = "${aws_cloudwatch_log_group.lg.name}-stream"
  log_group_name = aws_cloudwatch_log_group.lg.name
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = aws_acm_certificate.test.arn
  }

  connection_log_options {
    enabled               = true
    cloudwatch_log_group  = aws_cloudwatch_log_group.lg.name
    cloudwatch_log_stream = aws_cloudwatch_log_stream.ls.name
  }
}
`, rName, rName)
}

func testAccEc2ClientVpnEndpointConfigWithDNSServers(rName string) string {
	return testAccEc2ClientVpnEndpointConfigAcmCertificateBase() + fmt.Sprintf(`
resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"

  dns_servers = ["8.8.8.8", "8.8.4.4"]

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = aws_acm_certificate.test.arn
  }

  connection_log_options {
    enabled = false
  }
}
`, rName)
}

func testAccEc2ClientVpnEndpointConfigWithMicrosoftAD(rName string) string {
	return testAccEc2ClientVpnEndpointConfigAcmCertificateBase() +
		testAccEc2ClientVpnEndpointMsADBase() + fmt.Sprintf(`
resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                = "directory-service-authentication"
    active_directory_id = aws_directory_service_directory.test.id
  }

  connection_log_options {
    enabled = false
  }
}
`, rName)
}

func testAccEc2ClientVpnEndpointConfigWithMutualAuthAndMicrosoftAD(rName string) string {
	return testAccEc2ClientVpnEndpointConfigAcmCertificateBase() + testAccEc2ClientVpnEndpointMsADBase() + fmt.Sprintf(`
resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                = "directory-service-authentication"
    active_directory_id = aws_directory_service_directory.test.id
  }

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = aws_acm_certificate.test.arn
  }

  connection_log_options {
    enabled = false
  }
}
`, rName)
}

func testAccEc2ClientVpnEndpointConfigWithFederatedAuth(rName string) string {
	return testAccEc2ClientVpnEndpointConfigAcmCertificateBase() + fmt.Sprintf(`
resource "aws_iam_saml_provider" "default" {
  name                   = "myprovider-%s"
  saml_metadata_document = file("./test-fixtures/saml-metadata.xml")
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type              = "federated-authentication"
    saml_provider_arn = aws_iam_saml_provider.default.arn
  }

  connection_log_options {
    enabled = false
  }
}
`, rName, rName)
}

func testAccEc2ClientVpnEndpointConfig_tags(rName string) string {
	return testAccEc2ClientVpnEndpointConfigAcmCertificateBase() + fmt.Sprintf(`
resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = aws_acm_certificate.test.arn
  }

  connection_log_options {
    enabled = false
  }

  tags = {
    Environment = "production"
    Usage       = "original"
  }
}
`, rName)
}

func testAccEc2ClientVpnEndpointConfig_tagsChanged(rName string) string {
	return testAccEc2ClientVpnEndpointConfigAcmCertificateBase() + fmt.Sprintf(`
resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = aws_acm_certificate.test.arn
  }

  connection_log_options {
    enabled = false
  }

  tags = {
    Usage = "changed"
  }
}
`, rName)
}

func testAccEc2ClientVpnEndpointConfigSplitTunnel(rName string, splitTunnel bool) string {
	return testAccEc2ClientVpnEndpointConfigAcmCertificateBase() + fmt.Sprintf(`
resource "aws_ec2_client_vpn_endpoint" "test" {
  client_cidr_block      = "10.0.0.0/16"
  description            = %[1]q
  server_certificate_arn = aws_acm_certificate.test.arn
  split_tunnel           = %[2]t

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = aws_acm_certificate.test.arn
  }

  connection_log_options {
    enabled = false
  }
}
`, rName, splitTunnel)
}
