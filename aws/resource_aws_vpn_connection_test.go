package aws

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

type TunnelOptions struct {
	psk                        string
	tunnelCidr                 string
	dpdTimeoutAction           string
	dpdTimeoutSeconds          int
	ikeVersions                string
	phase1DhGroupNumbers       string
	phase1EncryptionAlgorithms string
	phase1IntegrityAlgorithms  string
	phase1LifetimeSeconds      int
	phase2DhGroupNumbers       string
	phase2EncryptionAlgorithms string
	phase2IntegrityAlgorithms  string
	phase2LifetimeSeconds      int
	rekeyFuzzPercentage        int
	rekeyMarginTimeSeconds     int
	replayWindowSize           int
	startupAction              string
}

func init() {
	resource.AddTestSweepers("aws_vpn_connection", &resource.Sweeper{
		Name: "aws_vpn_connection",
		F:    testSweepEc2VpnConnections,
	})
}

func testSweepEc2VpnConnections(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ec2conn
	input := &ec2.DescribeVpnConnectionsInput{}

	// DescribeVpnConnections does not currently have any form of pagination
	output, err := conn.DescribeVpnConnections(input)

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 VPN Connection sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error retrieving EC2 VPN Connections: %s", err)
	}

	for _, vpnConnection := range output.VpnConnections {
		if aws.StringValue(vpnConnection.State) == ec2.VpnStateDeleted {
			continue
		}

		id := aws.StringValue(vpnConnection.VpnConnectionId)
		input := &ec2.DeleteVpnConnectionInput{
			VpnConnectionId: vpnConnection.VpnConnectionId,
		}

		log.Printf("[INFO] Deleting EC2 VPN Connection: %s", id)

		_, err := conn.DeleteVpnConnection(input)

		if isAWSErr(err, "InvalidVpnConnectionID.NotFound", "") {
			continue
		}

		if err != nil {
			return fmt.Errorf("error deleting EC2 VPN Connection (%s): %s", id, err)
		}

		if err := waitForEc2VpnConnectionDeletion(conn, id); err != nil {
			return fmt.Errorf("error waiting for VPN connection (%s) to delete: %s", id, err)
		}
	}

	return nil
}

func TestAccAWSVpnConnection_basic(t *testing.T) {
	rInt := acctest.RandInt()
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheck(t, ec2.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccAwsVpnConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsVpnConnectionConfig(rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccAwsVpnConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_attachment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "enable_acceleration", "false"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`vpn-connection/vpn-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsVpnConnectionConfigUpdate(rInt, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccAwsVpnConnectionExists(resourceName, &vpn),
				),
			},
		},
	})
}

func TestAccAWSVpnConnection_TransitGatewayID(t *testing.T) {
	var vpn ec2.VpnConnection
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	resourceName := "aws_vpn_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSEc2TransitGateway(t)
		},
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccAwsVpnConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsVpnConnectionConfigTransitGatewayID(rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccAwsVpnConnectionExists(resourceName, &vpn),
					resource.TestMatchResourceAttr(resourceName, "transit_gateway_attachment_id", regexp.MustCompile(`tgw-attach-.+`)),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", transitGatewayResourceName, "id"),
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

func TestAccAWSVpnConnection_Tunnel1InsideCidr(t *testing.T) {
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccAwsVpnConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsVpnConnectionConfigTunnel1InsideCidr(rBgpAsn, "169.254.8.0/30", "169.254.9.0/30"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsVpnConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_inside_cidr", "169.254.8.0/30"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_inside_cidr", "169.254.9.0/30"),
				),
			},
			// NOTE: Import does not currently have access to the Terraform configuration,
			//       so proper tunnel ordering is not guaranteed on import. The import
			//       identifier could potentially be updated to accept optional tunnel
			//       configuration information, however the format for this could be
			//       confusing and/or difficult to implement.
		},
	})
}

func TestAccAWSVpnConnection_Tunnel1InsideIpv6Cidr(t *testing.T) {
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccAwsVpnConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsVpnConnectionConfigTunnel1InsideIpv6Cidr(rBgpAsn, "fd00:2001:db8:2:2d1:81ff:fe41:d200/126", "fd00:2001:db8:2:2d1:81ff:fe41:d204/126"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsVpnConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_inside_ipv6_cidr", "fd00:2001:db8:2:2d1:81ff:fe41:d200/126"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_inside_ipv6_cidr", "fd00:2001:db8:2:2d1:81ff:fe41:d204/126"),
				),
			},
			// NOTE: Import does not currently have access to the Terraform configuration,
			//       so proper tunnel ordering is not guaranteed on import. The import
			//       identifier could potentially be updated to accept optional tunnel
			//       configuration information, however the format for this could be
			//       confusing and/or difficult to implement.
		},
	})
}

func TestAccAWSVpnConnection_Tunnel1PresharedKey(t *testing.T) {
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccAwsVpnConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsVpnConnectionConfigTunnel1PresharedKey(rBgpAsn, "tunnel1presharedkey", "tunnel2presharedkey"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsVpnConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_preshared_key", "tunnel1presharedkey"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_preshared_key", "tunnel2presharedkey"),
				),
			},
			// NOTE: Import does not currently have access to the Terraform configuration,
			//       so proper tunnel ordering is not guaranteed on import. The import
			//       identifier could potentially be updated to accept optional tunnel
			//       configuration information, however the format for this could be
			//       confusing and/or difficult to implement.
		},
	})
}

func TestAccAWSVpnConnection_tunnelOptions(t *testing.T) {
	badCidrRangeErr := regexp.MustCompile(`expected \w+ to not be any of \[[\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\/30\s?]+\]`)
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	tunnel1 := TunnelOptions{
		psk:                        "12345678",
		tunnelCidr:                 "169.254.8.0/30",
		dpdTimeoutAction:           "clear",
		dpdTimeoutSeconds:          30,
		ikeVersions:                "\"ikev1\", \"ikev2\"",
		phase1DhGroupNumbers:       "2, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24",
		phase1EncryptionAlgorithms: "\"AES128\", \"AES256\", \"AES128-GCM-16\", \"AES256-GCM-16\"",
		phase1IntegrityAlgorithms:  "\"SHA1\", \"SHA2-256\", \"SHA2-384\", \"SHA2-512\"",
		phase1LifetimeSeconds:      28800,
		phase2DhGroupNumbers:       "2, 5, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24",
		phase2EncryptionAlgorithms: "\"AES128\", \"AES256\", \"AES128-GCM-16\", \"AES256-GCM-16\"",
		phase2IntegrityAlgorithms:  "\"SHA1\", \"SHA2-256\", \"SHA2-384\", \"SHA2-512\"",
		phase2LifetimeSeconds:      3600,
		rekeyFuzzPercentage:        100,
		rekeyMarginTimeSeconds:     540,
		replayWindowSize:           1024,
		startupAction:              "add",
	}

	tunnel2 := TunnelOptions{
		psk:                        "abcdefgh",
		tunnelCidr:                 "169.254.9.0/30",
		dpdTimeoutAction:           "clear",
		dpdTimeoutSeconds:          30,
		ikeVersions:                "\"ikev1\", \"ikev2\"",
		phase1DhGroupNumbers:       "2, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24",
		phase1EncryptionAlgorithms: "\"AES128\", \"AES256\", \"AES128-GCM-16\", \"AES256-GCM-16\"",
		phase1IntegrityAlgorithms:  "\"SHA1\", \"SHA2-256\", \"SHA2-384\", \"SHA2-512\"",
		phase1LifetimeSeconds:      28800,
		phase2DhGroupNumbers:       "2, 5, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24",
		phase2EncryptionAlgorithms: "\"AES128\", \"AES256\", \"AES128-GCM-16\", \"AES256-GCM-16\"",
		phase2IntegrityAlgorithms:  "\"SHA1\", \"SHA2-256\", \"SHA2-384\", \"SHA2-512\"",
		phase2LifetimeSeconds:      3600,
		rekeyFuzzPercentage:        100,
		rekeyMarginTimeSeconds:     540,
		replayWindowSize:           1024,
		startupAction:              "add",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheck(t, ec2.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccAwsVpnConnectionDestroy,
		Steps: []resource.TestStep{
			// Checking CIDR blocks
			{
				Config:      testAccAwsVpnConnectionConfigSingleTunnelOptions(rBgpAsn, "12345678", "not-a-cidr"),
				ExpectError: regexp.MustCompile(`invalid CIDR address: not-a-cidr`),
			},
			{
				Config:      testAccAwsVpnConnectionConfigSingleTunnelOptions(rBgpAsn, "12345678", "169.254.254.0/31"),
				ExpectError: regexp.MustCompile(`expected "\w+" to contain a network Value with between 30 and 30 significant bits`),
			},
			{
				Config:      testAccAwsVpnConnectionConfigSingleTunnelOptions(rBgpAsn, "12345678", "172.16.0.0/30"),
				ExpectError: regexp.MustCompile(`must be within 169.254.0.0/16`),
			},
			{
				Config:      testAccAwsVpnConnectionConfigSingleTunnelOptions(rBgpAsn, "12345678", "169.254.0.0/30"),
				ExpectError: badCidrRangeErr,
			},
			{
				Config:      testAccAwsVpnConnectionConfigSingleTunnelOptions(rBgpAsn, "12345678", "169.254.1.0/30"),
				ExpectError: badCidrRangeErr,
			},
			{
				Config:      testAccAwsVpnConnectionConfigSingleTunnelOptions(rBgpAsn, "12345678", "169.254.2.0/30"),
				ExpectError: badCidrRangeErr,
			},
			{
				Config:      testAccAwsVpnConnectionConfigSingleTunnelOptions(rBgpAsn, "12345678", "169.254.3.0/30"),
				ExpectError: badCidrRangeErr,
			},
			{
				Config:      testAccAwsVpnConnectionConfigSingleTunnelOptions(rBgpAsn, "12345678", "169.254.4.0/30"),
				ExpectError: badCidrRangeErr,
			},
			{
				Config:      testAccAwsVpnConnectionConfigSingleTunnelOptions(rBgpAsn, "12345678", "169.254.5.0/30"),
				ExpectError: badCidrRangeErr,
			},
			{
				Config:      testAccAwsVpnConnectionConfigSingleTunnelOptions(rBgpAsn, "12345678", "169.254.169.252/30"),
				ExpectError: badCidrRangeErr,
			},

			// Checking PreShared Key
			{
				Config:      testAccAwsVpnConnectionConfigSingleTunnelOptions(rBgpAsn, "1234567", "169.254.254.0/30"),
				ExpectError: regexp.MustCompile(`expected length of \w+ to be in the range \(8 - 64\)`),
			},
			{
				Config:      testAccAwsVpnConnectionConfigSingleTunnelOptions(rBgpAsn, acctest.RandStringFromCharSet(65, acctest.CharSetAlpha), "169.254.254.0/30"),
				ExpectError: regexp.MustCompile(`expected length of \w+ to be in the range \(8 - 64\)`),
			},
			{
				Config:      testAccAwsVpnConnectionConfigSingleTunnelOptions(rBgpAsn, "01234567", "169.254.254.0/30"),
				ExpectError: regexp.MustCompile(`cannot start with zero character`),
			},
			{
				Config:      testAccAwsVpnConnectionConfigSingleTunnelOptions(rBgpAsn, "1234567!", "169.254.254.0/30"),
				ExpectError: regexp.MustCompile(`can only contain alphanumeric, period and underscore characters`),
			},

			// Should pre-check:
			// - local_ipv4_network_cidr
			// - local_ipv6_network_cidr
			// - remote_ipv4_network_cidr
			// - remote_ipv6_network_cidr
			// - tunnel_inside_ip_version
			// - tunnel1_dpd_timeout_action
			// - tunnel1_dpd_timeout_seconds
			// - tunnel1_phase1_lifetime_seconds
			// - tunnel1_phase2_lifetime_seconds
			// - tunnel1_rekey_fuzz_percentage
			// - tunnel1_rekey_margin_time_seconds
			// - tunnel1_replay_window_size
			// - tunnel1_startup_action
			// - tunnel1_inside_cidr
			// - tunnel1_inside_ipv6_cidr

			//Try actual building
			{
				Config: testAccAwsVpnConnectionConfigTunnelOptions(rBgpAsn, "192.168.1.1/32", "192.168.1.2/32", tunnel1, tunnel2),
				Check: resource.ComposeTestCheckFunc(
					testAccAwsVpnConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "static_routes_only", "false"),

					resource.TestCheckResourceAttr(resourceName, "tunnel1_inside_cidr", "169.254.8.0/30"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_preshared_key", "12345678"),

					resource.TestCheckResourceAttr(resourceName, "tunnel2_inside_cidr", "169.254.9.0/30"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_preshared_key", "abcdefgh"),
				),
			},
			// NOTE: Import does not currently have access to the Terraform configuration,
			//       so proper tunnel ordering is not guaranteed on import. The import
			//       identifier could potentially be updated to accept optional tunnel
			//       configuration information, however the format for this could be
			//       confusing and/or difficult to implement.
		},
	})
}

func TestAccAWSVpnConnection_withoutStaticRoutes(t *testing.T) {
	rInt := acctest.RandInt()
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheck(t, ec2.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccAwsVpnConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsVpnConnectionConfigUpdate(rInt, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccAwsVpnConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "static_routes_only", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_acceleration", "false"),
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

func TestAccAWSVpnConnection_withEnableAcceleration(t *testing.T) {
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheck(t, ec2.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccAwsVpnConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsVpnConnectionConfigEnableAcceleration(rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccAwsVpnConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "enable_acceleration", "true"),
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

func TestAccAWSVpnConnection_withIpv6(t *testing.T) {
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheck(t, ec2.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccAwsVpnConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsVpnConnectionConfigIpv6(rBgpAsn, "fd00:2001:db8:2:2d1:81ff:fe41:d201/128", "fd00:2001:db8:2:2d1:81ff:fe41:d202/128", "fd00:2001:db8:2:2d1:81ff:fe41:d200/126", "fd00:2001:db8:2:2d1:81ff:fe41:d204/126"),
				Check: resource.ComposeTestCheckFunc(
					testAccAwsVpnConnectionExists(resourceName, &vpn),
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

func TestAccAWSVpnConnection_tags(t *testing.T) {
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheck(t, ec2.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccAwsVpnConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsVpnConnectionConfigTags1(rBgpAsn, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccAwsVpnConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsVpnConnectionConfigTags2(rBgpAsn, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccAwsVpnConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAwsVpnConnectionConfigTags1(rBgpAsn, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccAwsVpnConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSVpnConnection_disappears(t *testing.T) {
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccAwsVpnConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsVpnConnectionConfig(rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccAwsVpnConnectionExists(resourceName, &vpn),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsVpnConnection(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAwsVpnConnectionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpn_connection" {
			continue
		}

		resp, err := conn.DescribeVpnConnections(&ec2.DescribeVpnConnectionsInput{
			VpnConnectionIds: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() == "InvalidVpnConnectionID.NotFound" {
				// not found
				return nil
			}
			return err
		}

		var vpn *ec2.VpnConnection
		for _, v := range resp.VpnConnections {
			if v.VpnConnectionId != nil && *v.VpnConnectionId == rs.Primary.ID {
				vpn = v
			}
		}

		if vpn == nil {
			// vpn connection not found
			return nil
		}

		if vpn.State != nil && *vpn.State == "deleted" {
			return nil
		}

	}

	return nil
}

func testAccAwsVpnConnectionExists(vpnConnectionResource string, vpnConnection *ec2.VpnConnection) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[vpnConnectionResource]
		if !ok {
			return fmt.Errorf("Not found: %s", vpnConnectionResource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}
		connection, ok := s.RootModule().Resources[vpnConnectionResource]
		if !ok {
			return fmt.Errorf("Not found: %s", vpnConnectionResource)
		}

		ec2conn := testAccProvider.Meta().(*AWSClient).ec2conn

		resp, err := ec2conn.DescribeVpnConnections(&ec2.DescribeVpnConnectionsInput{
			VpnConnectionIds: []*string{aws.String(connection.Primary.ID)},
		})

		if err != nil {
			return err
		}

		*vpnConnection = *resp.VpnConnections[0]

		return nil
	}
}

func TestXmlConfigToTunnelInfo(t *testing.T) {
	testCases := []struct {
		Name                  string
		XML                   string
		Tunnel1PreSharedKey   string
		Tunnel1InsideCidr     string
		Tunnel1InsideIpv6Cidr string
		ExpectError           bool
		ExpectTunnelInfo      TunnelInfo
	}{
		{
			Name: "outside address sort",
			XML:  testAccAwsVpnTunnelInfoXML,
			ExpectTunnelInfo: TunnelInfo{
				Tunnel1Address:          "1.1.1.1",
				Tunnel1BGPASN:           "1111",
				Tunnel1BGPHoldTime:      31,
				Tunnel1CgwInsideAddress: "169.254.11.1",
				Tunnel1PreSharedKey:     "FIRST_KEY",
				Tunnel1VgwInsideAddress: "168.254.11.2",
				Tunnel2Address:          "2.2.2.2",
				Tunnel2BGPASN:           "2222",
				Tunnel2BGPHoldTime:      32,
				Tunnel2CgwInsideAddress: "169.254.12.1",
				Tunnel2PreSharedKey:     "SECOND_KEY",
				Tunnel2VgwInsideAddress: "169.254.12.2",
			},
		},
		{
			Name:                "Tunnel1PreSharedKey",
			XML:                 testAccAwsVpnTunnelInfoXML,
			Tunnel1PreSharedKey: "SECOND_KEY",
			ExpectTunnelInfo: TunnelInfo{
				Tunnel1Address:          "2.2.2.2",
				Tunnel1BGPASN:           "2222",
				Tunnel1BGPHoldTime:      32,
				Tunnel1CgwInsideAddress: "169.254.12.1",
				Tunnel1PreSharedKey:     "SECOND_KEY",
				Tunnel1VgwInsideAddress: "169.254.12.2",
				Tunnel2Address:          "1.1.1.1",
				Tunnel2BGPASN:           "1111",
				Tunnel2BGPHoldTime:      31,
				Tunnel2CgwInsideAddress: "169.254.11.1",
				Tunnel2PreSharedKey:     "FIRST_KEY",
				Tunnel2VgwInsideAddress: "168.254.11.2",
			},
		},
		{
			Name:              "Tunnel1InsideCidr",
			XML:               testAccAwsVpnTunnelInfoXML,
			Tunnel1InsideCidr: "169.254.12.0/30",
			ExpectTunnelInfo: TunnelInfo{
				Tunnel1Address:          "2.2.2.2",
				Tunnel1BGPASN:           "2222",
				Tunnel1BGPHoldTime:      32,
				Tunnel1CgwInsideAddress: "169.254.12.1",
				Tunnel1PreSharedKey:     "SECOND_KEY",
				Tunnel1VgwInsideAddress: "169.254.12.2",
				Tunnel2Address:          "1.1.1.1",
				Tunnel2BGPASN:           "1111",
				Tunnel2BGPHoldTime:      31,
				Tunnel2CgwInsideAddress: "169.254.11.1",
				Tunnel2PreSharedKey:     "FIRST_KEY",
				Tunnel2VgwInsideAddress: "168.254.11.2",
			},
		},
		// IPv6 logic is equivalent to IPv4, so we can reuse configuration, expected, etc.
		{
			Name:                  "Tunnel1InsideIpv6Cidr",
			XML:                   testAccAwsVpnTunnelInfoXML,
			Tunnel1InsideIpv6Cidr: "169.254.12.1",
			ExpectTunnelInfo: TunnelInfo{
				Tunnel1Address:          "2.2.2.2",
				Tunnel1BGPASN:           "2222",
				Tunnel1BGPHoldTime:      32,
				Tunnel1CgwInsideAddress: "169.254.12.1",
				Tunnel1PreSharedKey:     "SECOND_KEY",
				Tunnel1VgwInsideAddress: "169.254.12.2",
				Tunnel2Address:          "1.1.1.1",
				Tunnel2BGPASN:           "1111",
				Tunnel2BGPHoldTime:      31,
				Tunnel2CgwInsideAddress: "169.254.11.1",
				Tunnel2PreSharedKey:     "FIRST_KEY",
				Tunnel2VgwInsideAddress: "168.254.11.2",
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.Name, func(t *testing.T) {
			tunnelInfo, err := xmlConfigToTunnelInfo(testCase.XML, testCase.Tunnel1PreSharedKey, testCase.Tunnel1InsideCidr, testCase.Tunnel1InsideIpv6Cidr)

			if err == nil && testCase.ExpectError {
				t.Fatalf("expected error, got none")
			}

			if err != nil && !testCase.ExpectError {
				t.Fatalf("expected no error, got: %s", err)
			}

			if actual, expected := *tunnelInfo, testCase.ExpectTunnelInfo; !reflect.DeepEqual(actual, expected) { // nosemgrep: prefer-aws-go-sdk-pointer-conversion-assignment
				t.Errorf("expected TunnelInfo:\n%+v\n\ngot:\n%+v\n\n", expected, actual)
			}
		})
	}
}

func testAccAwsVpnConnectionConfig(rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "vpn_gateway" {
  tags = {
    Name = "vpn_gateway"
  }
}

resource "aws_customer_gateway" "customer_gateway" {
  bgp_asn    = %d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = "main-customer-gateway"
  }
}

resource "aws_vpn_connection" "test" {
  vpn_gateway_id      = aws_vpn_gateway.vpn_gateway.id
  customer_gateway_id = aws_customer_gateway.customer_gateway.id
  type                = "ipsec.1"
  static_routes_only  = true
}
`, rBgpAsn)
}

// Change static_routes_only to be false, forcing a refresh.
func testAccAwsVpnConnectionConfigUpdate(rInt, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "vpn_gateway" {
  tags = {
    Name = "vpn_gateway"
  }
}

resource "aws_customer_gateway" "customer_gateway" {
  bgp_asn    = %d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = "main-customer-gateway-%d"
  }
}

resource "aws_vpn_connection" "test" {
  vpn_gateway_id      = aws_vpn_gateway.vpn_gateway.id
  customer_gateway_id = aws_customer_gateway.customer_gateway.id
  type                = "ipsec.1"
  static_routes_only  = false
}
`, rBgpAsn, rInt)
}

func testAccAwsVpnConnectionConfigEnableAcceleration(rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {}
resource "aws_customer_gateway" "customer_gateway" {
  bgp_asn    = %d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"
  tags = {
    Name = "tf-acc-test-ec2-vpn-connection-enable-acceleration"
  }
}
resource "aws_vpn_connection" "test" {
  customer_gateway_id = aws_customer_gateway.customer_gateway.id
  transit_gateway_id  = aws_ec2_transit_gateway.test.id
  type                = "ipsec.1"
  static_routes_only  = false
  enable_acceleration = true
}
`, rBgpAsn)
}

func testAccAwsVpnConnectionConfigIpv6(rBgpAsn int, localIpv6NetworkCidr string, remoteIpv6NetworkCidr string, tunnel1InsideIpv6Cidr string, tunnel2InsideIpv6Cidr string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {}
resource "aws_customer_gateway" "customer_gateway" {
  bgp_asn    = %d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"
  tags = {
    Name = "tf-acc-test-ec2-vpn-connection-enable-acceleration"
  }
}
resource "aws_vpn_connection" "test" {
  customer_gateway_id = aws_customer_gateway.customer_gateway.id
  transit_gateway_id  = aws_ec2_transit_gateway.test.id
  type                = "ipsec.1"
  static_routes_only  = false
  enable_acceleration = false

  local_ipv6_network_cidr  = %[2]q
  remote_ipv6_network_cidr = %[3]q
  tunnel_inside_ip_version = "ipv6"

  tunnel1_inside_ipv6_cidr = %[4]q
  tunnel2_inside_ipv6_cidr = %[5]q
}
`, rBgpAsn, localIpv6NetworkCidr, remoteIpv6NetworkCidr, tunnel1InsideIpv6Cidr, tunnel2InsideIpv6Cidr)
}

func testAccAwsVpnConnectionConfigSingleTunnelOptions(rBgpAsn int, psk string, tunnelCidr string) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "vpn_gateway" {
  tags = {
    Name = "vpn_gateway"
  }
}

resource "aws_customer_gateway" "customer_gateway" {
  bgp_asn    = %d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = "main-customer-gateway"
  }
}

resource "aws_vpn_connection" "test" {
  vpn_gateway_id      = aws_vpn_gateway.vpn_gateway.id
  customer_gateway_id = aws_customer_gateway.customer_gateway.id
  type                = "ipsec.1"
  static_routes_only  = false

  tunnel1_inside_cidr   = "%s"
  tunnel1_preshared_key = "%s"
}
`, rBgpAsn, tunnelCidr, psk)
}

func testAccAwsVpnConnectionConfigTransitGatewayID(rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = "tf-acc-test-ec2-vpn-connection-transit-gateway-id"
  }
}

resource "aws_vpn_connection" "test" {
  customer_gateway_id = aws_customer_gateway.test.id
  transit_gateway_id  = aws_ec2_transit_gateway.test.id
  type                = aws_customer_gateway.test.type
}
`, rBgpAsn)
}

func testAccAwsVpnConnectionConfigTunnel1InsideCidr(rBgpAsn int, tunnel1InsideCidr string, tunnel2InsideCidr string) string {
	return fmt.Sprintf(`
resource "aws_customer_gateway" "test" {
  bgp_asn    = %[1]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"
}

resource "aws_vpn_gateway" "test" {}

resource "aws_vpn_connection" "test" {
  customer_gateway_id = aws_customer_gateway.test.id
  tunnel1_inside_cidr = %[2]q
  tunnel2_inside_cidr = %[3]q
  type                = "ipsec.1"
  vpn_gateway_id      = aws_vpn_gateway.test.id
}
`, rBgpAsn, tunnel1InsideCidr, tunnel2InsideCidr)
}

func testAccAwsVpnConnectionConfigTunnel1InsideIpv6Cidr(rBgpAsn int, tunnel1InsideIpv6Cidr string, tunnel2InsideIpv6Cidr string) string {
	return fmt.Sprintf(`
resource "aws_customer_gateway" "test" {
  bgp_asn    = %[1]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"
}

resource "aws_ec2_transit_gateway" "test" {}

resource "aws_vpn_connection" "test" {
  customer_gateway_id      = aws_customer_gateway.test.id
  transit_gateway_id       = aws_ec2_transit_gateway.test.id
  tunnel_inside_ip_version = "ipv6"
  tunnel1_inside_ipv6_cidr = %[2]q
  tunnel2_inside_ipv6_cidr = %[3]q
  type                     = "ipsec.1"
}
`, rBgpAsn, tunnel1InsideIpv6Cidr, tunnel2InsideIpv6Cidr)
}

func testAccAwsVpnConnectionConfigTunnel1PresharedKey(rBgpAsn int, tunnel1PresharedKey string, tunnel2PresharedKey string) string {
	return fmt.Sprintf(`
resource "aws_customer_gateway" "test" {
  bgp_asn    = %[1]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"
}

resource "aws_vpn_gateway" "test" {}

resource "aws_vpn_connection" "test" {
  customer_gateway_id   = aws_customer_gateway.test.id
  tunnel1_preshared_key = %[2]q
  tunnel2_preshared_key = %[3]q
  type                  = "ipsec.1"
  vpn_gateway_id        = aws_vpn_gateway.test.id
}
`, rBgpAsn, tunnel1PresharedKey, tunnel2PresharedKey)
}

func testAccAwsVpnConnectionConfigTunnelOptions(
	rBgpAsn int,
	localIpv4NetworkCidr string,
	remoteIpv4NetworkCidr string,
	tunnel1 TunnelOptions,
	tunnel2 TunnelOptions,
) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "vpn_gateway" {
  tags = {
    Name = "vpn_gateway"
  }
}

resource "aws_customer_gateway" "customer_gateway" {
  bgp_asn    = %d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = "main-customer-gateway"
  }
}

resource "aws_vpn_connection" "test" {
  vpn_gateway_id      = aws_vpn_gateway.vpn_gateway.id
  customer_gateway_id = aws_customer_gateway.customer_gateway.id
  type                = "ipsec.1"
  static_routes_only  = false

  local_ipv4_network_cidr  = %[2]q
  remote_ipv4_network_cidr = %[3]q

  tunnel1_inside_cidr                  = %[4]q
  tunnel1_preshared_key                = %[5]q
  tunnel1_dpd_timeout_action           = %[6]q
  tunnel1_dpd_timeout_seconds          = %[7]d
  tunnel1_ike_versions                 = [%[8]s]
  tunnel1_phase1_dh_group_numbers      = [%[9]s]
  tunnel1_phase1_encryption_algorithms = [%[10]s]
  tunnel1_phase1_integrity_algorithms  = [%[11]s]
  tunnel1_phase1_lifetime_seconds      = %[12]d
  tunnel1_phase2_dh_group_numbers      = [%[13]s]
  tunnel1_phase2_encryption_algorithms = [%[14]s]
  tunnel1_phase2_integrity_algorithms  = [%[15]s]
  tunnel1_phase2_lifetime_seconds      = %[16]d
  tunnel1_rekey_fuzz_percentage        = %[17]d
  tunnel1_rekey_margin_time_seconds    = %[18]d
  tunnel1_replay_window_size           = %[19]d
  tunnel1_startup_action               = %[20]q

  tunnel2_inside_cidr                  = %[21]q
  tunnel2_preshared_key                = %[22]q
  tunnel2_dpd_timeout_action           = %[23]q
  tunnel2_dpd_timeout_seconds          = %[24]d
  tunnel2_ike_versions                 = [%[25]s]
  tunnel2_phase1_dh_group_numbers      = [%[26]s]
  tunnel2_phase1_encryption_algorithms = [%[27]s]
  tunnel2_phase1_integrity_algorithms  = [%[28]s]
  tunnel2_phase1_lifetime_seconds      = %[29]d
  tunnel2_phase2_dh_group_numbers      = [%[30]s]
  tunnel2_phase2_encryption_algorithms = [%[31]s]
  tunnel2_phase2_integrity_algorithms  = [%[32]s]
  tunnel2_phase2_lifetime_seconds      = %[33]d
  tunnel2_rekey_fuzz_percentage        = %[34]d
  tunnel2_rekey_margin_time_seconds    = %[35]d
  tunnel2_replay_window_size           = %[36]d
  tunnel2_startup_action               = %[37]q
}
`,
		rBgpAsn,
		localIpv4NetworkCidr,
		remoteIpv4NetworkCidr,
		tunnel1.tunnelCidr,
		tunnel1.psk,
		tunnel1.dpdTimeoutAction,
		tunnel1.dpdTimeoutSeconds,
		tunnel1.ikeVersions,
		tunnel1.phase1DhGroupNumbers,
		tunnel1.phase1EncryptionAlgorithms,
		tunnel1.phase1IntegrityAlgorithms,
		tunnel1.phase1LifetimeSeconds,
		tunnel1.phase2DhGroupNumbers,
		tunnel1.phase2EncryptionAlgorithms,
		tunnel1.phase2IntegrityAlgorithms,
		tunnel1.phase2LifetimeSeconds,
		tunnel1.rekeyFuzzPercentage,
		tunnel1.rekeyMarginTimeSeconds,
		tunnel1.replayWindowSize,
		tunnel1.startupAction,
		tunnel2.tunnelCidr,
		tunnel2.psk,
		tunnel2.dpdTimeoutAction,
		tunnel2.dpdTimeoutSeconds,
		tunnel2.ikeVersions,
		tunnel2.phase1DhGroupNumbers,
		tunnel2.phase1EncryptionAlgorithms,
		tunnel2.phase1IntegrityAlgorithms,
		tunnel2.phase1LifetimeSeconds,
		tunnel2.phase2DhGroupNumbers,
		tunnel2.phase2EncryptionAlgorithms,
		tunnel2.phase2IntegrityAlgorithms,
		tunnel2.phase2LifetimeSeconds,
		tunnel2.rekeyFuzzPercentage,
		tunnel2.rekeyMarginTimeSeconds,
		tunnel2.replayWindowSize,
		tunnel2.startupAction)
}

func testAccAwsVpnConnectionConfigTags1(rBgpAsn int, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "vpn_gateway" {
  tags = {
    Name = "vpn_gateway"
  }
}

resource "aws_customer_gateway" "customer_gateway" {
  bgp_asn    = %d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = "main-customer-gateway"
  }
}

resource "aws_vpn_connection" "test" {
  vpn_gateway_id      = aws_vpn_gateway.vpn_gateway.id
  customer_gateway_id = aws_customer_gateway.customer_gateway.id
  type                = "ipsec.1"
  static_routes_only  = true

  tags = {
    %[2]q = %[3]q
  }
}
`, rBgpAsn, tagKey1, tagValue1)
}

func testAccAwsVpnConnectionConfigTags2(rBgpAsn int, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "vpn_gateway" {
  tags = {
    Name = "vpn_gateway"
  }
}

resource "aws_customer_gateway" "customer_gateway" {
  bgp_asn    = %d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = "main-customer-gateway"
  }
}

resource "aws_vpn_connection" "test" {
  vpn_gateway_id      = aws_vpn_gateway.vpn_gateway.id
  customer_gateway_id = aws_customer_gateway.customer_gateway.id
  type                = "ipsec.1"
  static_routes_only  = true

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rBgpAsn, tagKey1, tagValue1, tagKey2, tagValue2)
}

// Test our VPN tunnel config XML parsing
const testAccAwsVpnTunnelInfoXML = `
<vpn_connection id="vpn-abc123">
  <ipsec_tunnel>
    <customer_gateway>
      <tunnel_outside_address>
        <ip_address>22.22.22.22</ip_address>
      </tunnel_outside_address>
      <tunnel_inside_address>
        <ip_address>169.254.12.1</ip_address>
        <network_mask>255.255.255.252</network_mask>
        <network_cidr>30</network_cidr>
      </tunnel_inside_address>
    </customer_gateway>
    <vpn_gateway>
      <tunnel_outside_address>
        <ip_address>2.2.2.2</ip_address>
      </tunnel_outside_address>
      <tunnel_inside_address>
        <ip_address>169.254.12.2</ip_address>
        <network_mask>255.255.255.252</network_mask>
        <network_cidr>30</network_cidr>
      </tunnel_inside_address>
      <bgp>
        <asn>2222</asn>
        <hold_time>32</hold_time>
      </bgp>
    </vpn_gateway>
    <ike>
      <pre_shared_key>SECOND_KEY</pre_shared_key>
    </ike>
  </ipsec_tunnel>
  <ipsec_tunnel>
    <customer_gateway>
      <tunnel_outside_address>
        <ip_address>11.11.11.11</ip_address>
      </tunnel_outside_address>
      <tunnel_inside_address>
        <ip_address>169.254.11.1</ip_address>
        <network_mask>255.255.255.252</network_mask>
        <network_cidr>30</network_cidr>
      </tunnel_inside_address>
    </customer_gateway>
    <vpn_gateway>
      <tunnel_outside_address>
        <ip_address>1.1.1.1</ip_address>
      </tunnel_outside_address>
      <tunnel_inside_address>
        <ip_address>168.254.11.2</ip_address>
        <network_mask>255.255.255.252</network_mask>
        <network_cidr>30</network_cidr>
      </tunnel_inside_address>
      <bgp>
        <asn>1111</asn>
        <hold_time>31</hold_time>
      </bgp>
    </vpn_gateway>
    <ike>
      <pre_shared_key>FIRST_KEY</pre_shared_key>
    </ike>
  </ipsec_tunnel>
</vpn_connection>
`
