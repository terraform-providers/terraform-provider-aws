package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
)

func testAccAwsEc2ClientVpnAuthorizationRule_basic(t *testing.T) {
	var v ec2.AuthorizationRule
	rStr := acctest.RandString(5)
	resourceName := "aws_ec2_client_vpn_authorization_rule.test"
	subnetResourceName := "aws_subnet.test.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ClientVpnAuthorizationRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnAuthorizationRuleConfigBasic(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnAuthorizationRuleExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "target_network_cidr", subnetResourceName, "cidr_block"),
					resource.TestCheckResourceAttr(resourceName, "authorize_all_groups", "true"),
					resource.TestCheckResourceAttr(resourceName, "access_group_id", ""),
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

func testAccAwsEc2ClientVpnAuthorizationRule_groups(t *testing.T) {
	var v1, v2, v3, v4 ec2.AuthorizationRule
	rStr := acctest.RandString(5)
	resource1Name := "aws_ec2_client_vpn_authorization_rule.test1"
	resource2Name := "aws_ec2_client_vpn_authorization_rule.test2"
	subnetResourceName := "aws_subnet.test.0"

	group1Name := "group_one"
	group2Name := "group_two"

	groups1 := map[string]string{
		"test1": group1Name,
	}
	groups2 := map[string]string{
		"test1": group1Name,
		"test2": group2Name,
	}
	groups3 := map[string]string{
		"test2": group2Name,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ClientVpnAuthorizationRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnAuthorizationRuleConfigGroups(rStr, groups1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnAuthorizationRuleExists(resource1Name, &v1),
					resource.TestCheckResourceAttrPair(resource1Name, "target_network_cidr", subnetResourceName, "cidr_block"),
					resource.TestCheckResourceAttr(resource1Name, "authorize_all_groups", "false"),
					resource.TestCheckResourceAttr(resource1Name, "access_group_id", group1Name),
				),
			},
			{
				ResourceName:      resource1Name,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEc2ClientVpnAuthorizationRuleConfigGroups(rStr, groups2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnAuthorizationRuleExists(resource1Name, &v2),
					resource.TestCheckResourceAttrPair(resource1Name, "target_network_cidr", subnetResourceName, "cidr_block"),
					resource.TestCheckResourceAttr(resource1Name, "authorize_all_groups", "false"),
					resource.TestCheckResourceAttr(resource1Name, "access_group_id", group1Name),

					testAccCheckAwsEc2ClientVpnAuthorizationRuleExists(resource2Name, &v3),
					resource.TestCheckResourceAttrPair(resource2Name, "target_network_cidr", subnetResourceName, "cidr_block"),
					resource.TestCheckResourceAttr(resource2Name, "authorize_all_groups", "false"),
					resource.TestCheckResourceAttr(resource2Name, "access_group_id", group2Name),
				),
			},
			{
				ResourceName:      resource2Name,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEc2ClientVpnAuthorizationRuleConfigGroups(rStr, groups3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnAuthorizationRuleExists(resource2Name, &v4),
					resource.TestCheckResourceAttrPair(resource2Name, "target_network_cidr", subnetResourceName, "cidr_block"),
					resource.TestCheckResourceAttr(resource2Name, "authorize_all_groups", "false"),
					resource.TestCheckResourceAttr(resource2Name, "access_group_id", group2Name),
				),
			},
		},
	})
}

func testAccAwsEc2ClientVpnAuthorizationRule_Subnets(t *testing.T) {
	var v1, v2, v3 ec2.AuthorizationRule
	rStr := acctest.RandString(5)
	resource1Name := "aws_ec2_client_vpn_authorization_rule.test1"
	resource2Name := "aws_ec2_client_vpn_authorization_rule.test2"

	subnetCount := 2

	subnetIndex1 := 0
	subnetIndex2 := 1

	case1 := map[string]int{
		"test1": subnetIndex1,
		"test2": subnetIndex2,
	}
	case2 := map[string]int{
		"test2": subnetIndex2,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ClientVpnAuthorizationRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnAuthorizationRuleConfigSubnets(rStr, subnetCount, case1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnAuthorizationRuleExists(resource1Name, &v1),
					resource.TestCheckResourceAttrPair(resource1Name, "target_network_cidr", fmt.Sprintf("aws_subnet.test.%d", subnetIndex1), "cidr_block"),
					resource.TestCheckResourceAttr(resource1Name, "authorize_all_groups", "true"),
					resource.TestCheckResourceAttr(resource1Name, "access_group_id", ""),

					testAccCheckAwsEc2ClientVpnAuthorizationRuleExists(resource2Name, &v2),
					resource.TestCheckResourceAttrPair(resource2Name, "target_network_cidr", fmt.Sprintf("aws_subnet.test.%d", subnetIndex2), "cidr_block"),
					resource.TestCheckResourceAttr(resource2Name, "authorize_all_groups", "true"),
					resource.TestCheckResourceAttr(resource2Name, "access_group_id", ""),
				),
			},
			{
				ResourceName:      resource2Name,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEc2ClientVpnAuthorizationRuleConfigSubnets(rStr, subnetCount, case2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnAuthorizationRuleExists(resource2Name, &v3),
					resource.TestCheckResourceAttrPair(resource2Name, "target_network_cidr", fmt.Sprintf("aws_subnet.test.%d", subnetIndex2), "cidr_block"),
					resource.TestCheckResourceAttr(resource2Name, "authorize_all_groups", "true"),
					resource.TestCheckResourceAttr(resource2Name, "access_group_id", ""),
				),
			},
		},
	})
}

func testAccAwsEc2ClientVpnAuthorizationRule_disappears(t *testing.T) {
	var v ec2.AuthorizationRule
	rStr := acctest.RandString(5)
	resourceName := "aws_ec2_client_vpn_authorization_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ClientVpnAuthorizationRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnAuthorizationRuleConfigBasic(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnAuthorizationRuleExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEc2ClientVpnAuthorizationRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsEc2ClientVpnAuthorizationRuleDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_client_vpn_authorization_rule" {
			continue
		}

		_, err := finder.ClientVpnAuthorizationRuleByID(conn, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Client VPN authorization rule (%s) still exists", rs.Primary.ID)
		}
		if isAWSErr(err, tfec2.ErrCodeClientVpnAuthorizationRuleNotFound, "") || isAWSErr(err, tfec2.ErrCodeClientVpnEndpointIdNotFound, "") {
			continue
		}
		return err
	}

	return nil
}

func testAccCheckAwsEc2ClientVpnAuthorizationRuleExists(name string, assoc *ec2.AuthorizationRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		result, err := finder.ClientVpnAuthorizationRuleByID(conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error reading Client VPN authorization rule (%s): %w", rs.Primary.ID, err)
		}

		if result != nil || len(result.AuthorizationRules) == 1 || result.AuthorizationRules[0] != nil {
			*assoc = *result.AuthorizationRules[0]
			return nil
		}

		return fmt.Errorf("Client VPN network association (%s) not found", rs.Primary.ID)
	}
}

func testAccEc2ClientVpnAuthorizationRuleConfigBasic(rName string) string {
	return composeConfig(
		testAccEc2ClientVpnAuthorizationRuleVpcBase(rName, 1),
		testAccEc2ClientVpnAuthorizationRuleAcmCertificateBase(),
		fmt.Sprintf(`
resource "aws_ec2_client_vpn_authorization_rule" "test" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.test.id
  target_network_cidr    = aws_subnet.test[0].cidr_block
  authorize_all_groups   = true
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

func testAccEc2ClientVpnAuthorizationRuleConfigGroups(rName string, groupNames map[string]string) string {
	var b strings.Builder
	for k, v := range groupNames {
		fmt.Fprintf(&b, `
resource "aws_ec2_client_vpn_authorization_rule" %[1]q {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.test.id
  target_network_cidr    = aws_subnet.test[0].cidr_block
  access_group_id        = %[2]q
}
`, k, v)
	}

	return composeConfig(
		testAccEc2ClientVpnAuthorizationRuleVpcBase(rName, 1),
		testAccEc2ClientVpnAuthorizationRuleAcmCertificateBase(),
		b.String(),
		fmt.Sprintf(`
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
}`, rName))
}

func testAccEc2ClientVpnAuthorizationRuleConfigSubnets(rName string, subnetCount int, groupNames map[string]int) string {
	var b strings.Builder
	for k, v := range groupNames {
		fmt.Fprintf(&b, `
resource "aws_ec2_client_vpn_authorization_rule" %[1]q {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.test.id
  target_network_cidr    = aws_subnet.test[%[2]d].cidr_block
  authorize_all_groups   = true
}
`, k, v)
	}

	return composeConfig(
		testAccEc2ClientVpnAuthorizationRuleVpcBase(rName, subnetCount),
		testAccEc2ClientVpnAuthorizationRuleAcmCertificateBase(),
		b.String(),
		fmt.Sprintf(`
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
}`, rName))
}

func testAccEc2ClientVpnAuthorizationRuleVpcBase(rName string, subnetCount int) string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-subnet-%[1]s"
  }
}

resource "aws_subnet" "test" {
  count                   = %[2]d
  availability_zone       = data.aws_availability_zones.available.names[count.index]
  cidr_block              = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id                  = aws_vpc.test.id
  map_public_ip_on_launch = true

  tags = {
    Name = "tf-acc-subnet-%[1]s"
  }
}
`, rName, subnetCount))
}

func testAccEc2ClientVpnAuthorizationRuleAcmCertificateBase() string {
	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, "example.com")

	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[1]s"
  private_key      = "%[2]s"
}
`, tlsPemEscapeNewlines(certificate), tlsPemEscapeNewlines(key))
}
