package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfawsresource"
)

func init() {
	resource.AddTestSweepers("aws_ec2_client_vpn_network_association", &resource.Sweeper{
		Name: "aws_ec2_client_vpn_network_association",
		F:    testSweepEc2ClientVpnNetworkAssociations,
		Dependencies: []string{
			"aws_directory_service_directory",
		},
	})
}

func testSweepEc2ClientVpnNetworkAssociations(region string) error {
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

			input := &ec2.DescribeClientVpnTargetNetworksInput{
				ClientVpnEndpointId: clientVpnEndpoint.ClientVpnEndpointId,
			}
			err := conn.DescribeClientVpnTargetNetworksPages(input, func(page *ec2.DescribeClientVpnTargetNetworksOutput, isLast bool) bool {
				if page == nil {
					return !isLast
				}

				for _, networkAssociation := range page.ClientVpnTargetNetworks {
					networkAssociationID := aws.StringValue(networkAssociation.AssociationId)
					clientVpnEndpointID := aws.StringValue(networkAssociation.ClientVpnEndpointId)

					log.Printf("[INFO] Deleting Client VPN network association (%s,%s)", clientVpnEndpointID, networkAssociationID)
					err := deleteClientVpnNetworkAssociation(conn, networkAssociationID, clientVpnEndpointID)

					if err != nil {
						sweeperErr := fmt.Errorf("error deleting Client VPN network association (%s,%s): %w", clientVpnEndpointID, networkAssociationID, err)
						log.Printf("[ERROR] %s", sweeperErr)
						sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					}
				}

				return !isLast
			})

			if testSweepSkipSweepError(err) {
				log.Printf("[WARN] Skipping Client VPN network association sweeper for %q: %s", region, err)
				return false
			}
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Client VPN network associations: %w", err))
				return false
			}
		}

		return !isLast
	})
	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Client VPN network association sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Client VPN network associations: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func testAccAwsEc2ClientVpnNetworkAssociation_basic(t *testing.T) {
	var assoc ec2.TargetNetwork
	var group ec2.SecurityGroup
	rStr := acctest.RandString(5)
	resourceName := "aws_ec2_client_vpn_network_association.test"
	endpointResourceName := "aws_ec2_client_vpn_endpoint.test"
	subnetResourceName := "aws_subnet.test"
	vpcResourceName := "aws_vpc.test"
	defaultSecurityGroupResourceName := "aws_default_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ClientVpnNetworkAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnNetworkAssociationConfigBasic(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnNetworkAssociationExists(resourceName, &assoc),
					resource.TestMatchResourceAttr(resourceName, "association_id", regexp.MustCompile("^cvpn-assoc-[a-z0-9]+$")),
					resource.TestCheckResourceAttrPair(resourceName, "id", resourceName, "association_id"),
					resource.TestCheckResourceAttrPair(resourceName, "client_vpn_endpoint_id", endpointResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", subnetResourceName, "id"),
					testAccCheckAWSDefaultSecurityGroupExists(defaultSecurityGroupResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					testAccCheckAwsEc2ClientVpnNetworkAssociationSecurityGroupID(resourceName, "security_groups.*", &group),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", vpcResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsEc2ClientVpnNetworkAssociationImportStateIdFunc(resourceName),
			},
		},
	})
}

func testAccAwsEc2ClientVpnNetworkAssociation_disappears(t *testing.T) {
	var assoc ec2.TargetNetwork
	rStr := acctest.RandString(5)
	resourceName := "aws_ec2_client_vpn_network_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ClientVpnNetworkAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnNetworkAssociationConfigBasic(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnNetworkAssociationExists(resourceName, &assoc),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEc2ClientVpnNetworkAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAwsEc2ClientVpnNetworkAssociation_securityGroups(t *testing.T) {
	var assoc1, assoc2 ec2.TargetNetwork
	var group11, group12, group21 ec2.SecurityGroup
	rStr := acctest.RandString(5)
	resourceName := "aws_ec2_client_vpn_network_association.test"
	securityGroup1ResourceName := "aws_security_group.test1"
	securityGroup2ResourceName := "aws_security_group.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckClientVPNSyncronize(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ClientVpnNetworkAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnNetworkAssociationTwoSecurityGroups(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnNetworkAssociationExists(resourceName, &assoc1),
					testAccCheckAWSDefaultSecurityGroupExists(securityGroup1ResourceName, &group11),
					testAccCheckAWSDefaultSecurityGroupExists(securityGroup2ResourceName, &group12),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "2"),
					testAccCheckAwsEc2ClientVpnNetworkAssociationSecurityGroupID(resourceName, "security_groups.*", &group11),
					testAccCheckAwsEc2ClientVpnNetworkAssociationSecurityGroupID(resourceName, "security_groups.*", &group12),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsEc2ClientVpnNetworkAssociationImportStateIdFunc(resourceName),
			},
			{
				Config: testAccEc2ClientVpnNetworkAssociationOneSecurityGroup(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnNetworkAssociationExists(resourceName, &assoc2),
					testAccCheckAWSDefaultSecurityGroupExists(securityGroup1ResourceName, &group21),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					testAccCheckAwsEc2ClientVpnNetworkAssociationSecurityGroupID(resourceName, "security_groups.*", &group21),
				),
			},
		},
	})
}

func testAccCheckAwsEc2ClientVpnNetworkAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_client_vpn_network_association" {
			continue
		}

		resp, _ := conn.DescribeClientVpnTargetNetworks(&ec2.DescribeClientVpnTargetNetworksInput{
			ClientVpnEndpointId: aws.String(rs.Primary.Attributes["client_vpn_endpoint_id"]),
			AssociationIds:      []*string{aws.String(rs.Primary.ID)},
		})

		for _, v := range resp.ClientVpnTargetNetworks {
			if *v.AssociationId == rs.Primary.ID && !(*v.Status.Code == ec2.AssociationStatusCodeDisassociated) {
				return fmt.Errorf("[DESTROY ERROR] Client VPN network association (%s) not deleted", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckAwsEc2ClientVpnNetworkAssociationExists(name string, assoc *ec2.TargetNetwork) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		resp, err := conn.DescribeClientVpnTargetNetworks(&ec2.DescribeClientVpnTargetNetworksInput{
			ClientVpnEndpointId: aws.String(rs.Primary.Attributes["client_vpn_endpoint_id"]),
			AssociationIds:      []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return fmt.Errorf("Error reading Client VPN network association (%s): %w", rs.Primary.ID, err)
		}

		for _, a := range resp.ClientVpnTargetNetworks {
			if *a.AssociationId == rs.Primary.ID && !(*a.Status.Code == ec2.AssociationStatusCodeDisassociated) {
				*assoc = *a
				return nil
			}
		}

		return fmt.Errorf("Client VPN network association (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckAwsEc2ClientVpnNetworkAssociationSecurityGroupID(name, key string, group *ec2.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return tfawsresource.TestCheckTypeSetElemAttr(name, key, aws.StringValue(group.GroupId))(s)
	}
}

func testAccAwsEc2ClientVpnNetworkAssociationImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return tfec2.ClientVpnNetworkAssociationCreateID(rs.Primary.Attributes["client_vpn_endpoint_id"], rs.Primary.ID), nil
	}
}

func testAccEc2ClientVpnNetworkAssociationConfigBasic(rName string) string {
	return composeConfig(
		testAccEc2ClientVpnNetworkAssociationVpcBase(rName),
		testAccEc2ClientVpnNetworkAssociationAcmCertificateBase(),
		fmt.Sprintf(`
resource "aws_ec2_client_vpn_network_association" "test" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.test.id
  subnet_id              = aws_subnet.test.id
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%[1]s"
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
`, rName))
}

func testAccEc2ClientVpnNetworkAssociationTwoSecurityGroups(rName string) string {
	return composeConfig(
		testAccEc2ClientVpnNetworkAssociationVpcBase(rName),
		testAccEc2ClientVpnNetworkAssociationAcmCertificateBase(),
		fmt.Sprintf(`
resource "aws_ec2_client_vpn_network_association" "test" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.test.id
  subnet_id              = aws_subnet.test.id
  security_groups        = [aws_security_group.test1.id, aws_security_group.test2.id]
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%[1]s"
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

resource "aws_security_group" "test1" {
  name        = "terraform_acceptance_test_example_1"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.test.id
}

resource "aws_security_group" "test2" {
  name        = "terraform_acceptance_test_example_2"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.test.id
}
`, rName))
}

func testAccEc2ClientVpnNetworkAssociationOneSecurityGroup(rName string) string {
	return composeConfig(
		testAccEc2ClientVpnNetworkAssociationVpcBase(rName),
		testAccEc2ClientVpnNetworkAssociationAcmCertificateBase(),
		fmt.Sprintf(`
resource "aws_ec2_client_vpn_network_association" "test" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.test.id
  subnet_id              = aws_subnet.test.id
  security_groups        = [aws_security_group.test1.id]
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%[1]s"
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

resource "aws_security_group" "test1" {
  name        = "terraform_acceptance_test_example_1"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.test.id
}

resource "aws_security_group" "test2" {
  name        = "terraform_acceptance_test_example_2"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.test.id
}
`, rName))
}

func testAccEc2ClientVpnNetworkAssociationVpcBase(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  # InvalidParameterValue: AZ us-west-2d is not currently supported. Please choose another az in this region
  exclude_zone_ids = ["usw2-az4"]
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-subnet-%[1]s"
  }
}

resource "aws_default_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  availability_zone       = data.aws_availability_zones.available.names[0]
  cidr_block              = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  vpc_id                  = aws_vpc.test.id
  map_public_ip_on_launch = true

  tags = {
    Name = "tf-acc-subnet-%[1]s"
  }
}
`, rName)
}

func testAccEc2ClientVpnNetworkAssociationAcmCertificateBase() string {
	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, "example.com")

	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[1]s"
  private_key      = "%[2]s"
}
`, tlsPemEscapeNewlines(certificate), tlsPemEscapeNewlines(key))
}
