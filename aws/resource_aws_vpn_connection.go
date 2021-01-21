package aws

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"regexp"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/hashcode"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

type XmlVpnConnectionConfig struct {
	Tunnels []XmlIpsecTunnel `xml:"ipsec_tunnel"`
}

type XmlIpsecTunnel struct {
	OutsideAddress   string `xml:"vpn_gateway>tunnel_outside_address>ip_address"`
	BGPASN           string `xml:"vpn_gateway>bgp>asn"`
	BGPHoldTime      int    `xml:"vpn_gateway>bgp>hold_time"`
	PreSharedKey     string `xml:"ike>pre_shared_key"`
	CgwInsideAddress string `xml:"customer_gateway>tunnel_inside_address>ip_address"`
	VgwInsideAddress string `xml:"vpn_gateway>tunnel_inside_address>ip_address"`
}

type TunnelInfo struct {
	Tunnel1Address          string
	Tunnel1CgwInsideAddress string
	Tunnel1VgwInsideAddress string
	Tunnel1PreSharedKey     string
	Tunnel1BGPASN           string
	Tunnel1BGPHoldTime      int
	Tunnel2Address          string
	Tunnel2CgwInsideAddress string
	Tunnel2VgwInsideAddress string
	Tunnel2PreSharedKey     string
	Tunnel2BGPASN           string
	Tunnel2BGPHoldTime      int
}

func (slice XmlVpnConnectionConfig) Len() int {
	return len(slice.Tunnels)
}

func (slice XmlVpnConnectionConfig) Less(i, j int) bool {
	return slice.Tunnels[i].OutsideAddress < slice.Tunnels[j].OutsideAddress
}

func (slice XmlVpnConnectionConfig) Swap(i, j int) {
	slice.Tunnels[i], slice.Tunnels[j] = slice.Tunnels[j], slice.Tunnels[i]
}

func resourceAwsVpnConnection() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsVpnConnectionCreate,
		Read:   resourceAwsVpnConnectionRead,
		Update: resourceAwsVpnConnectionUpdate,
		Delete: resourceAwsVpnConnectionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpn_gateway_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"transit_gateway_id"},
			},

			"customer_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"transit_gateway_attachment_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"transit_gateway_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"vpn_gateway_id"},
			},

			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"enable_acceleration": {
				Type:         schema.TypeBool,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				RequiredWith: []string{"transit_gateway_id"},
			},

			"local_ipv4_network_cidr": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateLocalIpv4NetworkCidr(),
			},

			"local_ipv6_network_cidr": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateLocalIpv6NetworkCidr(),
				RequiredWith: []string{"transit_gateway_id"},
			},

			"remote_ipv4_network_cidr": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateLocalIpv4NetworkCidr(),
			},

			"remote_ipv6_network_cidr": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateLocalIpv6NetworkCidr(),
				RequiredWith: []string{"transit_gateway_id"},
			},

			"static_routes_only": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"tunnel_inside_ip_version": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateTunnelInsideIPVersion(),
			},

			"tunnel1_dpd_timeout_action": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateVpnConnectionTunnelDpdTimeoutAction(),
			},

			"tunnel1_dpd_timeout_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validateVpnConnectionTunnelDpdTimeoutSeconds(),
			},

			"tunnel1_ike_versions": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"tunnel1_phase1_dh_group_numbers": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},

			"tunnel1_phase1_encryption_algorithms": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"tunnel1_phase1_integrity_algorithms": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"tunnel1_phase1_lifetime_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validateVpnConnectionTunnelPhase1LifetimeSeconds(),
			},

			"tunnel1_phase2_dh_group_numbers": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},

			"tunnel1_phase2_encryption_algorithms": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"tunnel1_phase2_integrity_algorithms": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"tunnel1_phase2_lifetime_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validateVpnConnectionTunnelPhase2LifetimeSeconds(),
			},

			"tunnel1_rekey_fuzz_percentage": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validateVpnConnectionTunnelRekeyFuzzPercentage(),
			},

			"tunnel1_rekey_margin_time_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validateVpnConnectionTunnelRekeyMarginTimeSeconds(),
			},

			"tunnel1_replay_window_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validateVpnConnectionTunnelReplayWindowSize(),
			},

			"tunnel1_startup_action": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateVpnConnectionTunnelStartupAction(),
			},

			"tunnel1_inside_cidr": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateVpnConnectionTunnelInsideCIDR(),
			},

			"tunnel1_inside_ipv6_cidr": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateVpnConnectionTunnelInsideIpv6CIDR(),
				RequiredWith: []string{"transit_gateway_id"},
			},

			"tunnel1_preshared_key": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateVpnConnectionTunnelPreSharedKey(),
			},

			"tunnel2_dpd_timeout_action": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateVpnConnectionTunnelDpdTimeoutAction(),
			},

			"tunnel2_dpd_timeout_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validateVpnConnectionTunnelDpdTimeoutSeconds(),
			},

			"tunnel2_ike_versions": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"tunnel2_phase1_dh_group_numbers": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},

			"tunnel2_phase1_encryption_algorithms": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"tunnel2_phase1_integrity_algorithms": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"tunnel2_phase1_lifetime_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validateVpnConnectionTunnelPhase1LifetimeSeconds(),
			},

			"tunnel2_phase2_dh_group_numbers": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},

			"tunnel2_phase2_encryption_algorithms": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"tunnel2_phase2_integrity_algorithms": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"tunnel2_phase2_lifetime_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validateVpnConnectionTunnelPhase2LifetimeSeconds(),
			},

			"tunnel2_rekey_fuzz_percentage": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validateVpnConnectionTunnelRekeyFuzzPercentage(),
			},

			"tunnel2_rekey_margin_time_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validateVpnConnectionTunnelRekeyMarginTimeSeconds(),
			},

			"tunnel2_replay_window_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validateVpnConnectionTunnelReplayWindowSize(),
			},

			"tunnel2_startup_action": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateVpnConnectionTunnelStartupAction(),
			},

			"tunnel2_inside_cidr": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateVpnConnectionTunnelInsideCIDR(),
			},

			"tunnel2_inside_ipv6_cidr": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateVpnConnectionTunnelInsideIpv6CIDR(),
				RequiredWith: []string{"transit_gateway_id"},
			},

			"tunnel2_preshared_key": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateVpnConnectionTunnelPreSharedKey(),
			},

			"tags": tagsSchema(),

			// Begin read only attributes
			"customer_gateway_configuration": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tunnel1_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tunnel1_cgw_inside_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tunnel1_vgw_inside_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tunnel1_bgp_asn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tunnel1_bgp_holdtime": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"tunnel2_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tunnel2_cgw_inside_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tunnel2_vgw_inside_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tunnel2_bgp_asn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tunnel2_bgp_holdtime": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"routes": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"destination_cidr_block": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"source": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"state": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Set: func(v interface{}) int {
					var buf bytes.Buffer
					m := v.(map[string]interface{})
					buf.WriteString(fmt.Sprintf("%s-", m["destination_cidr_block"].(string)))
					buf.WriteString(fmt.Sprintf("%s-", m["source"].(string)))
					buf.WriteString(fmt.Sprintf("%s-", m["state"].(string)))
					return hashcode.String(buf.String())
				},
			},

			"vgw_telemetry": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"accepted_route_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"last_status_change": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"outside_ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"status_message": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Set: func(v interface{}) int {
					var buf bytes.Buffer
					m := v.(map[string]interface{})
					buf.WriteString(fmt.Sprintf("%s-", m["outside_ip_address"].(string)))
					return hashcode.String(buf.String())
				},
			},
		},
	}
}

func resourceAwsVpnConnectionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	// Fill the connection options for the EC2 API
	connectOpts := expandVpnConnectionOptions(d)

	createOpts := &ec2.CreateVpnConnectionInput{
		CustomerGatewayId: aws.String(d.Get("customer_gateway_id").(string)),
		Options:           connectOpts,
		Type:              aws.String(d.Get("type").(string)),
		TagSpecifications: ec2TagSpecificationsFromMap(d.Get("tags").(map[string]interface{}), ec2.ResourceTypeVpnConnection),
	}

	if v, ok := d.GetOk("transit_gateway_id"); ok {
		createOpts.TransitGatewayId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("vpn_gateway_id"); ok {
		createOpts.VpnGatewayId = aws.String(v.(string))
	}

	// Create the VPN Connection
	log.Printf("[DEBUG] Creating vpn connection")
	resp, err := conn.CreateVpnConnection(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating vpn connection: %s", err)
	}

	d.SetId(aws.StringValue(resp.VpnConnection.VpnConnectionId))

	if err := waitForEc2VpnConnectionAvailable(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for VPN connection (%s) to become available: %s", d.Id(), err)
	}

	// Read off the API to populate our RO fields.
	return resourceAwsVpnConnectionRead(d, meta)
}

func vpnConnectionRefreshFunc(conn *ec2.EC2, connectionId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeVpnConnections(&ec2.DescribeVpnConnectionsInput{
			VpnConnectionIds: []*string{aws.String(connectionId)},
		})

		if err != nil {
			if isAWSErr(err, "InvalidVpnConnectionID.NotFound", "") {
				resp = nil
			} else {
				log.Printf("Error on VPNConnectionRefresh: %s", err)
				return nil, "", err
			}
		}

		if resp == nil || len(resp.VpnConnections) == 0 {
			return nil, "", nil
		}

		connection := resp.VpnConnections[0]
		return connection, aws.StringValue(connection.State), nil
	}
}

func resourceAwsVpnConnectionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := conn.DescribeVpnConnections(&ec2.DescribeVpnConnectionsInput{
		VpnConnectionIds: []*string{aws.String(d.Id())},
	})

	if isAWSErr(err, "InvalidVpnConnectionID.NotFound", "") {
		log.Printf("[WARN] EC2 VPN Connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 VPN Connection (%s): %s", d.Id(), err)
	}

	if resp == nil || len(resp.VpnConnections) == 0 || resp.VpnConnections[0] == nil {
		return fmt.Errorf("error reading EC2 VPN Connection (%s): empty response", d.Id())
	}

	if len(resp.VpnConnections) > 1 {
		return fmt.Errorf("error reading EC2 VPN Connection (%s): multiple responses", d.Id())
	}

	vpnConnection := resp.VpnConnections[0]

	if aws.StringValue(vpnConnection.State) == ec2.VpnStateDeleted {
		log.Printf("[WARN] EC2 VPN Connection (%s) already deleted, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	var transitGatewayAttachmentID string
	if vpnConnection.TransitGatewayId != nil {
		input := &ec2.DescribeTransitGatewayAttachmentsInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("resource-id"),
					Values: []*string{vpnConnection.VpnConnectionId},
				},
				{
					Name:   aws.String("resource-type"),
					Values: []*string{aws.String(ec2.TransitGatewayAttachmentResourceTypeVpn)},
				},
				{
					Name:   aws.String("transit-gateway-id"),
					Values: []*string{vpnConnection.TransitGatewayId},
				},
			},
		}

		log.Printf("[DEBUG] Finding EC2 VPN Connection Transit Gateway Attachment: %s", input)
		output, err := conn.DescribeTransitGatewayAttachments(input)

		if err != nil {
			return fmt.Errorf("error finding EC2 VPN Connection (%s) Transit Gateway Attachment: %s", d.Id(), err)
		}

		if output == nil || len(output.TransitGatewayAttachments) == 0 || output.TransitGatewayAttachments[0] == nil {
			return fmt.Errorf("error finding EC2 VPN Connection (%s) Transit Gateway Attachment: empty response", d.Id())
		}

		if len(output.TransitGatewayAttachments) > 1 {
			return fmt.Errorf("error reading EC2 VPN Connection (%s) Transit Gateway Attachment: multiple responses", d.Id())
		}

		transitGatewayAttachmentID = aws.StringValue(output.TransitGatewayAttachments[0].TransitGatewayAttachmentId)
	}

	// Set attributes under the user's control.
	d.Set("vpn_gateway_id", vpnConnection.VpnGatewayId)
	d.Set("customer_gateway_id", vpnConnection.CustomerGatewayId)
	d.Set("transit_gateway_id", vpnConnection.TransitGatewayId)
	d.Set("type", vpnConnection.Type)

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(vpnConnection.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	if vpnConnection.Options != nil {
		if err := d.Set("enable_acceleration", vpnConnection.Options.EnableAcceleration); err != nil {
			return err
		}

		if err := d.Set("local_ipv4_network_cidr", vpnConnection.Options.LocalIpv4NetworkCidr); err != nil {
			return err
		}

		if err := d.Set("local_ipv6_network_cidr", vpnConnection.Options.LocalIpv6NetworkCidr); err != nil {
			return err
		}

		if err := d.Set("remote_ipv4_network_cidr", vpnConnection.Options.RemoteIpv4NetworkCidr); err != nil {
			return err
		}

		if err := d.Set("remote_ipv6_network_cidr", vpnConnection.Options.RemoteIpv6NetworkCidr); err != nil {
			return err
		}

		if err := d.Set("static_routes_only", vpnConnection.Options.StaticRoutesOnly); err != nil {
			return err
		}

		if err := d.Set("tunnel_inside_ip_version", vpnConnection.Options.TunnelInsideIpVersion); err != nil {
			return err
		}
		if err := flattenTunnelOptions(d, vpnConnection); err != nil {
			return err
		}

	} else {
		//If there no Options on the connection then we do not support it
		d.Set("enable_acceleration", false)
		d.Set("local_ipv4_network_cidr", "")
		d.Set("local_ipv6_network_cidr", "")
		d.Set("remote_ipv4_network_cidr", "")
		d.Set("remote_ipv6_network_cidr", "")
		d.Set("static_routes_only", false)
		d.Set("tunnel_inside_ip_version", "")
	}

	// Set read only attributes.
	d.Set("customer_gateway_configuration", vpnConnection.CustomerGatewayConfiguration)
	d.Set("transit_gateway_attachment_id", transitGatewayAttachmentID)

	if vpnConnection.CustomerGatewayConfiguration != nil {
		if tunnelInfo, err := xmlConfigToTunnelInfo(*vpnConnection.CustomerGatewayConfiguration); err != nil {
			log.Printf("[ERR] Error unmarshaling XML configuration for (%s): %s", d.Id(), err)
		} else {
			d.Set("tunnel1_address", tunnelInfo.Tunnel1Address)
			d.Set("tunnel1_cgw_inside_address", tunnelInfo.Tunnel1CgwInsideAddress)
			d.Set("tunnel1_vgw_inside_address", tunnelInfo.Tunnel1VgwInsideAddress)
			d.Set("tunnel1_preshared_key", tunnelInfo.Tunnel1PreSharedKey)
			d.Set("tunnel1_bgp_asn", tunnelInfo.Tunnel1BGPASN)
			d.Set("tunnel1_bgp_holdtime", tunnelInfo.Tunnel1BGPHoldTime)
			d.Set("tunnel2_address", tunnelInfo.Tunnel2Address)
			d.Set("tunnel2_preshared_key", tunnelInfo.Tunnel2PreSharedKey)
			d.Set("tunnel2_cgw_inside_address", tunnelInfo.Tunnel2CgwInsideAddress)
			d.Set("tunnel2_vgw_inside_address", tunnelInfo.Tunnel2VgwInsideAddress)
			d.Set("tunnel2_bgp_asn", tunnelInfo.Tunnel2BGPASN)
			d.Set("tunnel2_bgp_holdtime", tunnelInfo.Tunnel2BGPHoldTime)
		}
	}

	if err := d.Set("vgw_telemetry", telemetryToMapList(vpnConnection.VgwTelemetry)); err != nil {
		return err
	}
	if err := d.Set("routes", routesToMapList(vpnConnection.Routes)); err != nil {
		return err
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "ec2",
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("vpn-connection/%s", d.Id()),
	}.String()

	d.Set("arn", arn)

	return nil
}

func flattenTunnelOptions(d *schema.ResourceData, vpnConnection *ec2.VpnConnection) error {
	if len(vpnConnection.Options.TunnelOptions) >= 1 {
		if err := d.Set("tunnel1_dpd_timeout_action", vpnConnection.Options.TunnelOptions[0].DpdTimeoutAction); err != nil {
			return err
		}

		if err := d.Set("tunnel1_dpd_timeout_seconds", vpnConnection.Options.TunnelOptions[0].DpdTimeoutSeconds); err != nil {
			return err
		}

		ikeVersions := []string{}
		for _, ikeVersion := range vpnConnection.Options.TunnelOptions[0].IkeVersions {
			ikeVersions = append(ikeVersions, *ikeVersion.Value)
		}
		if err := d.Set("tunnel1_ike_versions", ikeVersions); err != nil {
			return err
		}

		phase1DHGroupNumbers := []int64{}
		for _, phase1DHGroupNumber := range vpnConnection.Options.TunnelOptions[0].Phase1DHGroupNumbers {
			phase1DHGroupNumbers = append(phase1DHGroupNumbers, *phase1DHGroupNumber.Value)
		}
		if err := d.Set("tunnel1_phase1_dh_group_numbers", phase1DHGroupNumbers); err != nil {
			return err
		}

		phase1EncAlgorithms := []string{}
		for _, phase1EncAlgorithm := range vpnConnection.Options.TunnelOptions[0].Phase1EncryptionAlgorithms {
			phase1EncAlgorithms = append(phase1EncAlgorithms, *phase1EncAlgorithm.Value)
		}
		if err := d.Set("tunnel1_phase1_encryption_algorithms", phase1EncAlgorithms); err != nil {
			return err
		}

		phase1IntegrityAlgorithms := []string{}
		for _, phase1IntegrityAlgorithm := range vpnConnection.Options.TunnelOptions[0].Phase1IntegrityAlgorithms {
			phase1IntegrityAlgorithms = append(phase1IntegrityAlgorithms, *phase1IntegrityAlgorithm.Value)
		}
		if err := d.Set("tunnel1_phase1_integrity_algorithms", phase1IntegrityAlgorithms); err != nil {
			return err
		}

		if err := d.Set("tunnel1_phase1_lifetime_seconds", vpnConnection.Options.TunnelOptions[0].Phase1LifetimeSeconds); err != nil {
			return err
		}

		phase2DHGroupNumbers := []int64{}
		for _, phase2DHGroupNumber := range vpnConnection.Options.TunnelOptions[0].Phase2DHGroupNumbers {
			phase2DHGroupNumbers = append(phase2DHGroupNumbers, *phase2DHGroupNumber.Value)
		}
		if err := d.Set("tunnel1_phase2_dh_group_numbers", phase2DHGroupNumbers); err != nil {
			return err
		}

		phase2EncAlgorithms := []string{}
		for _, phase2EncAlgorithm := range vpnConnection.Options.TunnelOptions[0].Phase2EncryptionAlgorithms {
			phase2EncAlgorithms = append(phase2EncAlgorithms, *phase2EncAlgorithm.Value)
		}
		if err := d.Set("tunnel1_phase2_encryption_algorithms", phase2EncAlgorithms); err != nil {
			return err
		}

		phase2IntegrityAlgorithms := []string{}
		for _, phase2IntegrityAlgorithm := range vpnConnection.Options.TunnelOptions[0].Phase2IntegrityAlgorithms {
			phase2IntegrityAlgorithms = append(phase2IntegrityAlgorithms, *phase2IntegrityAlgorithm.Value)
		}
		if err := d.Set("tunnel1_phase2_integrity_algorithms", phase2IntegrityAlgorithms); err != nil {
			return err
		}

		if err := d.Set("tunnel1_phase2_lifetime_seconds", vpnConnection.Options.TunnelOptions[0].Phase2LifetimeSeconds); err != nil {
			return err
		}

		if err := d.Set("tunnel1_rekey_fuzz_percentage", vpnConnection.Options.TunnelOptions[0].RekeyFuzzPercentage); err != nil {
			return err
		}

		if err := d.Set("tunnel1_rekey_margin_time_seconds", vpnConnection.Options.TunnelOptions[0].RekeyMarginTimeSeconds); err != nil {
			return err
		}

		if err := d.Set("tunnel1_replay_window_size", vpnConnection.Options.TunnelOptions[0].ReplayWindowSize); err != nil {
			return err
		}

		if err := d.Set("tunnel1_startup_action", vpnConnection.Options.TunnelOptions[0].StartupAction); err != nil {
			return err
		}

		if err := d.Set("tunnel1_inside_cidr", vpnConnection.Options.TunnelOptions[0].TunnelInsideCidr); err != nil {
			return err
		}

		if err := d.Set("tunnel1_inside_ipv6_cidr", vpnConnection.Options.TunnelOptions[0].TunnelInsideIpv6Cidr); err != nil {
			return err
		}
	}
	if len(vpnConnection.Options.TunnelOptions) >= 2 {
		if err := d.Set("tunnel2_dpd_timeout_action", vpnConnection.Options.TunnelOptions[1].DpdTimeoutAction); err != nil {
			return err
		}

		if err := d.Set("tunnel2_dpd_timeout_seconds", vpnConnection.Options.TunnelOptions[1].DpdTimeoutSeconds); err != nil {
			return err
		}

		ikeVersions := []string{}
		for _, ikeVersion := range vpnConnection.Options.TunnelOptions[1].IkeVersions {
			ikeVersions = append(ikeVersions, *ikeVersion.Value)
		}
		if err := d.Set("tunnel2_ike_versions", ikeVersions); err != nil {
			return err
		}

		phase1DHGroupNumbers := []int64{}
		for _, phase1DHGroupNumber := range vpnConnection.Options.TunnelOptions[1].Phase1DHGroupNumbers {
			phase1DHGroupNumbers = append(phase1DHGroupNumbers, *phase1DHGroupNumber.Value)
		}
		if err := d.Set("tunnel2_phase1_dh_group_numbers", phase1DHGroupNumbers); err != nil {
			return err
		}

		phase1EncAlgorithms := []string{}
		for _, phase1EncAlgorithm := range vpnConnection.Options.TunnelOptions[1].Phase1EncryptionAlgorithms {
			phase1EncAlgorithms = append(phase1EncAlgorithms, *phase1EncAlgorithm.Value)
		}

		if err := d.Set("tunnel2_phase1_encryption_algorithms", phase1EncAlgorithms); err != nil {
			return err
		}

		phase1IntegrityAlgorithms := []string{}
		for _, phase1IntegrityAlgorithm := range vpnConnection.Options.TunnelOptions[1].Phase1IntegrityAlgorithms {
			phase1IntegrityAlgorithms = append(phase1IntegrityAlgorithms, *phase1IntegrityAlgorithm.Value)
		}
		if err := d.Set("tunnel2_phase1_integrity_algorithms", phase1IntegrityAlgorithms); err != nil {
			return err
		}

		if err := d.Set("tunnel2_phase1_lifetime_seconds", vpnConnection.Options.TunnelOptions[1].Phase1LifetimeSeconds); err != nil {
			return err
		}

		phase2DHGroupNumbers := []int64{}
		for _, phase2DHGroupNumber := range vpnConnection.Options.TunnelOptions[1].Phase2DHGroupNumbers {
			phase2DHGroupNumbers = append(phase2DHGroupNumbers, *phase2DHGroupNumber.Value)
		}
		if err := d.Set("tunnel2_phase2_dh_group_numbers", phase2DHGroupNumbers); err != nil {
			return err
		}

		phase2EncAlgorithms := []string{}
		for _, phase2EncAlgorithm := range vpnConnection.Options.TunnelOptions[1].Phase2EncryptionAlgorithms {
			phase2EncAlgorithms = append(phase2EncAlgorithms, *phase2EncAlgorithm.Value)
		}

		if err := d.Set("tunnel2_phase2_encryption_algorithms", phase2EncAlgorithms); err != nil {
			return err
		}

		phase2IntegrityAlgorithms := []string{}
		for _, phase2IntegrityAlgorithm := range vpnConnection.Options.TunnelOptions[1].Phase2IntegrityAlgorithms {
			phase2IntegrityAlgorithms = append(phase2IntegrityAlgorithms, *phase2IntegrityAlgorithm.Value)
		}
		if err := d.Set("tunnel2_phase2_integrity_algorithms", phase2IntegrityAlgorithms); err != nil {
			return err
		}

		if err := d.Set("tunnel2_phase2_lifetime_seconds", vpnConnection.Options.TunnelOptions[1].Phase2LifetimeSeconds); err != nil {
			return err
		}

		if err := d.Set("tunnel2_rekey_fuzz_percentage", vpnConnection.Options.TunnelOptions[1].RekeyFuzzPercentage); err != nil {
			return err
		}

		if err := d.Set("tunnel2_rekey_margin_time_seconds", vpnConnection.Options.TunnelOptions[1].RekeyMarginTimeSeconds); err != nil {
			return err
		}

		if err := d.Set("tunnel2_replay_window_size", vpnConnection.Options.TunnelOptions[1].ReplayWindowSize); err != nil {
			return err
		}

		if err := d.Set("tunnel2_startup_action", vpnConnection.Options.TunnelOptions[1].StartupAction); err != nil {
			return err
		}

		if err := d.Set("tunnel2_inside_cidr", vpnConnection.Options.TunnelOptions[1].TunnelInsideCidr); err != nil {
			return err
		}

		if err := d.Set("tunnel2_inside_ipv6_cidr", vpnConnection.Options.TunnelOptions[1].TunnelInsideIpv6Cidr); err != nil {
			return err
		}
	}
	return nil
}

func resourceAwsVpnConnectionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	if err := modifyVpnConnectionOptions(d, conn); err != nil {
		return err
	}

	if err := modifyVpnTunnels(d, conn); err != nil {
		return err
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		vpnConnectionID := d.Id()

		if err := keyvaluetags.Ec2UpdateTags(conn, vpnConnectionID, o, n); err != nil {
			return fmt.Errorf("error updating EC2 VPN Connection (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsVpnConnectionRead(d, meta)
}

func resourceAwsVpnConnectionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	_, err := conn.DeleteVpnConnection(&ec2.DeleteVpnConnectionInput{
		VpnConnectionId: aws.String(d.Id()),
	})

	if isAWSErr(err, "InvalidVpnConnectionID.NotFound", "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting VPN Connection (%s): %s", d.Id(), err)
	}

	if err := waitForEc2VpnConnectionDeletion(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for VPN connection (%s) to delete: %s", d.Id(), err)
	}

	return nil
}

func expandVpnConnectionOptions(d *schema.ResourceData) *ec2.VpnConnectionOptionsSpecification {
	var connectOpts *ec2.VpnConnectionOptionsSpecification = new(ec2.VpnConnectionOptionsSpecification)
	ipv := d.Get("tunnel_inside_ip_version").(string)
	if ipv == "ipv6" {
		if v, ok := d.GetOk("local_ipv6_network_cidr"); ok {
			connectOpts.LocalIpv6NetworkCidr = aws.String(v.(string))
		}

		if v, ok := d.GetOk("remote_ipv6_network_cidr"); ok {
			connectOpts.RemoteIpv6NetworkCidr = aws.String(v.(string))
		}

		connectOpts.TunnelInsideIpVersion = aws.String(ipv)
	} else {
		if v, ok := d.GetOk("local_ipv4_network_cidr"); ok {
			connectOpts.LocalIpv4NetworkCidr = aws.String(v.(string))
		}

		if v, ok := d.GetOk("remote_ipv4_network_cidr"); ok {
			connectOpts.RemoteIpv4NetworkCidr = aws.String(v.(string))
		}

		connectOpts.TunnelInsideIpVersion = aws.String("ipv4")
	}

	if v, ok := d.GetOk("enable_acceleration"); ok {
		connectOpts.EnableAcceleration = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("static_routes_only"); ok {
		connectOpts.StaticRoutesOnly = aws.Bool(v.(bool))
	}

	// Fill the tunnel options for the EC2 API
	connectOpts.TunnelOptions = expandVpnTunnelOptions(d)

	return connectOpts
}

func expandVpnTunnelOptions(d *schema.ResourceData) []*ec2.VpnTunnelOptionsSpecification {
	options := []*ec2.VpnTunnelOptionsSpecification{
		{}, {},
	}

	if v, ok := d.GetOk("tunnel1_dpd_timeout_action"); ok {
		options[0].DPDTimeoutAction = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tunnel2_dpd_timeout_action"); ok {
		options[1].DPDTimeoutAction = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tunnel1_dpd_timeout_seconds"); ok {
		options[0].DPDTimeoutSeconds = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("tunnel2_dpd_timeout_seconds"); ok {
		options[1].DPDTimeoutSeconds = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("tunnel1_ike_versions"); ok {
		l := []*ec2.IKEVersionsRequestListValue{}
		for _, s := range v.(*schema.Set).List() {
			l = append(l, &ec2.IKEVersionsRequestListValue{Value: aws.String(s.(string))})
		}
		options[0].IKEVersions = l
	}

	if v, ok := d.GetOk("tunnel2_ike_versions"); ok {
		l := []*ec2.IKEVersionsRequestListValue{}
		for _, s := range v.(*schema.Set).List() {
			l = append(l, &ec2.IKEVersionsRequestListValue{Value: aws.String(s.(string))})
		}
		options[1].IKEVersions = l
	}

	if v, ok := d.GetOk("tunnel1_phase1_dh_group_numbers"); ok {
		l := []*ec2.Phase1DHGroupNumbersRequestListValue{}
		for _, s := range v.(*schema.Set).List() {
			l = append(l, &ec2.Phase1DHGroupNumbersRequestListValue{Value: aws.Int64(int64(s.(int)))})
		}
		options[0].Phase1DHGroupNumbers = l
	}

	if v, ok := d.GetOk("tunnel2_phase1_dh_group_numbers"); ok {
		l := []*ec2.Phase1DHGroupNumbersRequestListValue{}
		for _, s := range v.(*schema.Set).List() {
			l = append(l, &ec2.Phase1DHGroupNumbersRequestListValue{Value: aws.Int64(int64(s.(int)))})
		}
		options[1].Phase1DHGroupNumbers = l
	}

	if v, ok := d.GetOk("tunnel1_phase1_encryption_algorithms"); ok {
		l := []*ec2.Phase1EncryptionAlgorithmsRequestListValue{}
		for _, s := range v.(*schema.Set).List() {
			l = append(l, &ec2.Phase1EncryptionAlgorithmsRequestListValue{Value: aws.String(s.(string))})
		}
		options[0].Phase1EncryptionAlgorithms = l
	}

	if v, ok := d.GetOk("tunnel2_phase1_encryption_algorithms"); ok {
		l := []*ec2.Phase1EncryptionAlgorithmsRequestListValue{}
		for _, s := range v.(*schema.Set).List() {
			l = append(l, &ec2.Phase1EncryptionAlgorithmsRequestListValue{Value: aws.String(s.(string))})
		}
		options[1].Phase1EncryptionAlgorithms = l
	}

	if v, ok := d.GetOk("tunnel1_phase1_integrity_algorithms"); ok {
		l := []*ec2.Phase1IntegrityAlgorithmsRequestListValue{}
		for _, s := range v.(*schema.Set).List() {
			l = append(l, &ec2.Phase1IntegrityAlgorithmsRequestListValue{Value: aws.String(s.(string))})
		}
		options[0].Phase1IntegrityAlgorithms = l
	}

	if v, ok := d.GetOk("tunnel2_phase1_integrity_algorithms"); ok {
		l := []*ec2.Phase1IntegrityAlgorithmsRequestListValue{}
		for _, s := range v.(*schema.Set).List() {
			l = append(l, &ec2.Phase1IntegrityAlgorithmsRequestListValue{Value: aws.String(s.(string))})
		}
		options[1].Phase1IntegrityAlgorithms = l
	}

	if v, ok := d.GetOk("tunnel1_phase1_lifetime_seconds"); ok {
		options[0].Phase1LifetimeSeconds = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("tunnel2_phase1_lifetime_seconds"); ok {
		options[1].Phase1LifetimeSeconds = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("tunnel1_phase2_dh_group_numbers"); ok {
		l := []*ec2.Phase2DHGroupNumbersRequestListValue{}
		for _, s := range v.(*schema.Set).List() {
			l = append(l, &ec2.Phase2DHGroupNumbersRequestListValue{Value: aws.Int64(int64(s.(int)))})
		}
		options[0].Phase2DHGroupNumbers = l
	}

	if v, ok := d.GetOk("tunnel2_phase2_dh_group_numbers"); ok {
		l := []*ec2.Phase2DHGroupNumbersRequestListValue{}
		for _, s := range v.(*schema.Set).List() {
			l = append(l, &ec2.Phase2DHGroupNumbersRequestListValue{Value: aws.Int64(int64(s.(int)))})
		}
		options[1].Phase2DHGroupNumbers = l
	}

	if v, ok := d.GetOk("tunnel1_phase2_encryption_algorithms"); ok {
		l := []*ec2.Phase2EncryptionAlgorithmsRequestListValue{}
		for _, s := range v.(*schema.Set).List() {
			l = append(l, &ec2.Phase2EncryptionAlgorithmsRequestListValue{Value: aws.String(s.(string))})
		}
		options[0].Phase2EncryptionAlgorithms = l
	}

	if v, ok := d.GetOk("tunnel2_phase2_encryption_algorithms"); ok {
		l := []*ec2.Phase2EncryptionAlgorithmsRequestListValue{}
		for _, s := range v.(*schema.Set).List() {
			l = append(l, &ec2.Phase2EncryptionAlgorithmsRequestListValue{Value: aws.String(s.(string))})
		}
		options[1].Phase2EncryptionAlgorithms = l
	}

	if v, ok := d.GetOk("tunnel1_phase2_integrity_algorithms"); ok {
		l := []*ec2.Phase2IntegrityAlgorithmsRequestListValue{}
		for _, s := range v.(*schema.Set).List() {
			l = append(l, &ec2.Phase2IntegrityAlgorithmsRequestListValue{Value: aws.String(s.(string))})
		}
		options[0].Phase2IntegrityAlgorithms = l
	}

	if v, ok := d.GetOk("tunnel2_phase2_integrity_algorithms"); ok {
		l := []*ec2.Phase2IntegrityAlgorithmsRequestListValue{}
		for _, s := range v.(*schema.Set).List() {
			l = append(l, &ec2.Phase2IntegrityAlgorithmsRequestListValue{Value: aws.String(s.(string))})
		}
		options[1].Phase2IntegrityAlgorithms = l
	}

	if v, ok := d.GetOk("tunnel1_phase2_lifetime_seconds"); ok {
		options[0].Phase2LifetimeSeconds = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("tunnel2_phase2_lifetime_seconds"); ok {
		options[1].Phase2LifetimeSeconds = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("tunnel1_rekey_fuzz_percentage"); ok {
		options[0].RekeyFuzzPercentage = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("tunnel2_rekey_fuzz_percentage"); ok {
		options[1].RekeyFuzzPercentage = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("tunnel1_rekey_margin_time_seconds"); ok {
		options[0].RekeyMarginTimeSeconds = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("tunnel2_rekey_margin_time_seconds"); ok {
		options[1].RekeyMarginTimeSeconds = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("tunnel1_replay_window_size"); ok {
		options[0].ReplayWindowSize = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("tunnel2_replay_window_size"); ok {
		options[1].ReplayWindowSize = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("tunnel1_startup_action"); ok {
		options[0].StartupAction = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tunnel2_startup_action"); ok {
		options[1].StartupAction = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tunnel1_inside_cidr"); ok {
		options[0].TunnelInsideCidr = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tunnel2_inside_cidr"); ok {
		options[1].TunnelInsideCidr = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tunnel1_inside_ipv6_cidr"); ok {
		options[0].TunnelInsideIpv6Cidr = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tunnel2_inside_ipv6_cidr"); ok {
		options[1].TunnelInsideIpv6Cidr = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tunnel1_preshared_key"); ok {
		options[0].PreSharedKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tunnel2_preshared_key"); ok {
		options[1].PreSharedKey = aws.String(v.(string))
	}

	return options
}

// routesToMapList turns the list of routes into a list of maps.
func routesToMapList(routes []*ec2.VpnStaticRoute) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(routes))
	for _, r := range routes {
		staticRoute := make(map[string]interface{})
		staticRoute["destination_cidr_block"] = aws.StringValue(r.DestinationCidrBlock)
		staticRoute["state"] = aws.StringValue(r.State)

		if r.Source != nil {
			staticRoute["source"] = aws.StringValue(r.Source)
		}

		result = append(result, staticRoute)
	}

	return result
}

// telemetryToMapList turns the VGW telemetry into a list of maps.
func telemetryToMapList(telemetry []*ec2.VgwTelemetry) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(telemetry))
	for _, t := range telemetry {
		vgw := make(map[string]interface{})
		vgw["accepted_route_count"] = aws.Int64Value(t.AcceptedRouteCount)
		vgw["outside_ip_address"] = aws.StringValue(t.OutsideIpAddress)
		vgw["status"] = aws.StringValue(t.Status)
		vgw["status_message"] = aws.StringValue(t.StatusMessage)

		// LastStatusChange is a time.Time(). Convert it into a string
		// so it can be handled by schema's type system.
		vgw["last_status_change"] = t.LastStatusChange.Format(time.RFC3339)
		result = append(result, vgw)
	}

	return result
}

func modifyVpnConnectionOptions(d *schema.ResourceData, conn *ec2.EC2) error {
	var connOpts *ec2.ModifyVpnConnectionOptionsInput = new(ec2.ModifyVpnConnectionOptionsInput)
	connChanged := false

	vpnConnectionID := d.Id()

	if d.HasChange("local_ipv4_network_cidr") {
		connChanged = true
		connOpts.LocalIpv4NetworkCidr = aws.String(d.Get("local_ipv4_network_cidr").(string))
	}

	if d.HasChange("local_ipv6_network_cidr") {
		connChanged = true
		connOpts.LocalIpv6NetworkCidr = aws.String(d.Get("local_ipv6_network_cidr").(string))
	}

	if d.HasChange("remote_ipv4_network_cidr") {
		connChanged = true
		connOpts.RemoteIpv4NetworkCidr = aws.String(d.Get("remote_ipv4_network_cidr").(string))
	}

	if d.HasChange("remote_ipv6_network_cidr") {
		connChanged = true
		connOpts.RemoteIpv6NetworkCidr = aws.String(d.Get("remote_ipv6_network_cidr").(string))
	}

	if connChanged {
		connOpts.VpnConnectionId = aws.String(vpnConnectionID)
		_, err := conn.ModifyVpnConnectionOptions(connOpts)
		if err != nil {
			return fmt.Errorf("Error modifying vpn connection options: %s", err)
		}

		if err := waitForEc2VpnConnectionAvailableWhenModifying(conn, vpnConnectionID); err != nil {
			return fmt.Errorf("error waiting for VPN connection (%s) to become available: %s", vpnConnectionID, err)
		}
	}

	return nil
}

func modifyVpnTunnels(d *schema.ResourceData, conn *ec2.EC2) error {
	tun1Changed := false
	tun2Changed := false
	vgwTelemetryTun1Index := 0
	vgwTelemetryTun2Index := 1
	options := []*ec2.ModifyVpnTunnelOptionsSpecification{
		{}, {},
	}

	vpnConnectionID := d.Id()

	if d.HasChange("tunnel1_dpd_timeout_action") {
		tun1Changed = true
		options[0].DPDTimeoutAction = aws.String(d.Get("tunnel1_dpd_timeout_action").(string))
	}

	if d.HasChange("tunnel2_dpd_timeout_action") {
		tun2Changed = true
		options[1].DPDTimeoutAction = aws.String(d.Get("tunnel2_dpd_timeout_action").(string))
	}

	if d.HasChange("tunnel1_dpd_timeout_seconds") {
		tun1Changed = true
		options[0].DPDTimeoutSeconds = aws.Int64(int64(d.Get("tunnel1_dpd_timeout_seconds").(int)))
	}

	if d.HasChange("tunnel2_dpd_timeout_seconds") {
		tun2Changed = true
		options[1].DPDTimeoutSeconds = aws.Int64(int64(d.Get("tunnel2_dpd_timeout_seconds").(int)))
	}

	if d.HasChange("tunnel1_ike_versions") {
		tun1Changed = true
		l := []*ec2.IKEVersionsRequestListValue{}
		for _, s := range d.Get("tunnel1_ike_versions").(*schema.Set).List() {
			l = append(l, &ec2.IKEVersionsRequestListValue{Value: aws.String(s.(string))})
		}
		options[0].IKEVersions = l
	}

	if d.HasChange("tunnel2_ike_versions") {
		tun2Changed = true
		l := []*ec2.IKEVersionsRequestListValue{}
		for _, s := range d.Get("tunnel2_ike_versions").(*schema.Set).List() {
			l = append(l, &ec2.IKEVersionsRequestListValue{Value: aws.String(s.(string))})
		}
		options[1].IKEVersions = l
	}

	if d.HasChange("tunnel1_phase1_dh_group_numbers") {
		tun1Changed = true
		l := []*ec2.Phase1DHGroupNumbersRequestListValue{}
		for _, s := range d.Get("tunnel1_phase1_dh_group_numbers").(*schema.Set).List() {
			l = append(l, &ec2.Phase1DHGroupNumbersRequestListValue{Value: aws.Int64(int64(s.(int)))})
		}
		options[0].Phase1DHGroupNumbers = l
	}

	if d.HasChange("tunnel2_phase1_dh_group_numbers") {
		tun2Changed = true
		l := []*ec2.Phase1DHGroupNumbersRequestListValue{}
		for _, s := range d.Get("tunnel2_phase1_dh_group_numbers").(*schema.Set).List() {
			l = append(l, &ec2.Phase1DHGroupNumbersRequestListValue{Value: aws.Int64(int64(s.(int)))})
		}
		options[1].Phase1DHGroupNumbers = l
	}

	if d.HasChange("tunnel1_phase1_encryption_algorithms") {
		tun1Changed = true
		l := []*ec2.Phase1EncryptionAlgorithmsRequestListValue{}
		for _, s := range d.Get("tunnel1_phase1_encryption_algorithms").(*schema.Set).List() {
			l = append(l, &ec2.Phase1EncryptionAlgorithmsRequestListValue{Value: aws.String(s.(string))})
		}
		options[0].Phase1EncryptionAlgorithms = l
	}

	if d.HasChange("tunnel2_phase1_encryption_algorithms") {
		tun2Changed = true
		l := []*ec2.Phase1EncryptionAlgorithmsRequestListValue{}
		for _, s := range d.Get("tunnel2_phase1_encryption_algorithms").(*schema.Set).List() {
			l = append(l, &ec2.Phase1EncryptionAlgorithmsRequestListValue{Value: aws.String(s.(string))})
		}
		options[1].Phase1EncryptionAlgorithms = l
	}

	if d.HasChange("tunnel1_phase1_integrity_algorithms") {
		tun1Changed = true
		l := []*ec2.Phase1IntegrityAlgorithmsRequestListValue{}
		for _, s := range d.Get("tunnel1_phase1_integrity_algorithms").(*schema.Set).List() {
			l = append(l, &ec2.Phase1IntegrityAlgorithmsRequestListValue{Value: aws.String(s.(string))})
		}
		options[0].Phase1IntegrityAlgorithms = l
	}

	if d.HasChange("tunnel2_phase1_integrity_algorithms") {
		tun2Changed = true
		l := []*ec2.Phase1IntegrityAlgorithmsRequestListValue{}
		for _, s := range d.Get("tunnel2_phase1_integrity_algorithms").(*schema.Set).List() {
			l = append(l, &ec2.Phase1IntegrityAlgorithmsRequestListValue{Value: aws.String(s.(string))})
		}
		options[1].Phase1IntegrityAlgorithms = l
	}

	if d.HasChange("tunnel1_phase1_lifetime_seconds") {
		tun1Changed = true
		options[0].Phase1LifetimeSeconds = aws.Int64(int64(d.Get("tunnel1_phase1_lifetime_seconds").(int)))
	}

	if d.HasChange("tunnel2_phase1_lifetime_seconds") {
		tun2Changed = true
		options[1].Phase1LifetimeSeconds = aws.Int64(int64(d.Get("tunnel2_phase1_lifetime_seconds").(int)))
	}

	if d.HasChange("tunnel1_phase2_dh_group_numbers") {
		tun1Changed = true
		l := []*ec2.Phase2DHGroupNumbersRequestListValue{}
		for _, s := range d.Get("tunnel1_phase2_dh_group_numbers").(*schema.Set).List() {
			l = append(l, &ec2.Phase2DHGroupNumbersRequestListValue{Value: aws.Int64(int64(s.(int)))})
		}
		options[0].Phase2DHGroupNumbers = l
	}

	if d.HasChange("tunnel2_phase2_dh_group_numbers") {
		tun2Changed = true
		l := []*ec2.Phase2DHGroupNumbersRequestListValue{}
		for _, s := range d.Get("tunnel2_phase2_dh_group_numbers").(*schema.Set).List() {
			l = append(l, &ec2.Phase2DHGroupNumbersRequestListValue{Value: aws.Int64(int64(s.(int)))})
		}
		options[1].Phase2DHGroupNumbers = l
	}

	if d.HasChange("tunnel1_phase2_encryption_algorithms") {
		tun1Changed = true
		l := []*ec2.Phase2EncryptionAlgorithmsRequestListValue{}
		for _, s := range d.Get("tunnel1_phase2_encryption_algorithms").(*schema.Set).List() {
			l = append(l, &ec2.Phase2EncryptionAlgorithmsRequestListValue{Value: aws.String(s.(string))})
		}
		options[0].Phase2EncryptionAlgorithms = l
	}

	if d.HasChange("tunnel2_phase2_encryption_algorithms") {
		tun2Changed = true
		l := []*ec2.Phase2EncryptionAlgorithmsRequestListValue{}
		for _, s := range d.Get("tunnel2_phase2_encryption_algorithms").(*schema.Set).List() {
			l = append(l, &ec2.Phase2EncryptionAlgorithmsRequestListValue{Value: aws.String(s.(string))})
		}
		options[1].Phase2EncryptionAlgorithms = l
	}

	if d.HasChange("tunnel1_phase2_integrity_algorithms") {
		tun1Changed = true
		l := []*ec2.Phase2IntegrityAlgorithmsRequestListValue{}
		for _, s := range d.Get("tunnel1_phase2_integrity_algorithms").(*schema.Set).List() {
			l = append(l, &ec2.Phase2IntegrityAlgorithmsRequestListValue{Value: aws.String(s.(string))})
		}
		options[0].Phase2IntegrityAlgorithms = l
	}

	if d.HasChange("tunnel2_phase2_integrity_algorithms") {
		tun2Changed = true
		l := []*ec2.Phase2IntegrityAlgorithmsRequestListValue{}
		for _, s := range d.Get("tunnel2_phase2_integrity_algorithms").(*schema.Set).List() {
			l = append(l, &ec2.Phase2IntegrityAlgorithmsRequestListValue{Value: aws.String(s.(string))})
		}
		options[1].Phase2IntegrityAlgorithms = l
	}

	if d.HasChange("tunnel1_phase2_lifetime_seconds") {
		tun1Changed = true
		options[0].Phase2LifetimeSeconds = aws.Int64(int64(d.Get("tunnel1_phase2_lifetime_seconds").(int)))
	}

	if d.HasChange("tunnel2_phase2_lifetime_seconds") {
		tun2Changed = true
		options[1].Phase2LifetimeSeconds = aws.Int64(int64(d.Get("tunnel2_phase2_lifetime_seconds").(int)))
	}

	if d.HasChange("tunnel1_rekey_fuzz_percentage") {
		tun1Changed = true
		options[0].RekeyFuzzPercentage = aws.Int64(int64(d.Get("tunnel1_rekey_fuzz_percentage").(int)))
	}

	if d.HasChange("tunnel2_rekey_fuzz_percentage") {
		tun2Changed = true
		options[1].RekeyFuzzPercentage = aws.Int64(int64(d.Get("tunnel2_rekey_fuzz_percentage").(int)))
	}

	if d.HasChange("tunnel1_rekey_margin_time_seconds") {
		tun1Changed = true
		options[0].RekeyMarginTimeSeconds = aws.Int64(int64(d.Get("tunnel1_rekey_margin_time_seconds").(int)))
	}

	if d.HasChange("tunnel2_rekey_margin_time_seconds") {
		tun2Changed = true
		options[1].RekeyMarginTimeSeconds = aws.Int64(int64(d.Get("tunnel2_rekey_margin_time_seconds").(int)))
	}

	if d.HasChange("tunnel1_replay_window_size") {
		tun1Changed = true
		options[0].ReplayWindowSize = aws.Int64(int64(d.Get("tunnel1_replay_window_size").(int)))
	}

	if d.HasChange("tunnel2_replay_window_size") {
		tun2Changed = true
		options[1].ReplayWindowSize = aws.Int64(int64(d.Get("tunnel2_replay_window_size").(int)))
	}

	if d.HasChange("tunnel1_startup_action") {
		tun1Changed = true
		options[0].StartupAction = aws.String(d.Get("tunnel1_startup_action").(string))
	}

	if d.HasChange("tunnel2_startup_action") {
		tun2Changed = true
		options[1].StartupAction = aws.String(d.Get("tunnel2_startup_action").(string))
	}

	if tun1Changed {
		if err := modifyVpnTunnelOptions(conn, d.Get("vgw_telemetry").(*schema.Set), vpnConnectionID, vgwTelemetryTun1Index, options[0]); err != nil {
			return err
		}
	}

	if tun2Changed {
		if err := modifyVpnTunnelOptions(conn, d.Get("vgw_telemetry").(*schema.Set), vpnConnectionID, vgwTelemetryTun2Index, options[1]); err != nil {
			return err
		}
	}

	return nil
}

func modifyVpnTunnelOptions(conn *ec2.EC2, vgwTelemetry *schema.Set, vpnConnectionID string, vgwTelemetryTunIndex int, optionsTun *ec2.ModifyVpnTunnelOptionsSpecification) error {
	if v := vgwTelemetry; v.Len() > 0 {
		vpnTunnelOutsideIPAddress := v.List()[vgwTelemetryTunIndex].(map[string]interface{})["outside_ip_address"].(string)

		o := &ec2.ModifyVpnTunnelOptionsInput{
			VpnConnectionId:           aws.String(vpnConnectionID),
			VpnTunnelOutsideIpAddress: aws.String(vpnTunnelOutsideIPAddress),
			TunnelOptions:             optionsTun,
		}

		_, err := conn.ModifyVpnTunnelOptions(o)
		if err != nil {
			return fmt.Errorf("Error modifying vpn tunnel options: %s", err)
		}

		if err := waitForEc2VpnConnectionAvailableWhenModifying(conn, vpnConnectionID); err != nil {
			return fmt.Errorf("error waiting for VPN connection (%s) to become available: %s", vpnConnectionID, err)
		}
	}

	return nil
}

func waitForEc2VpnConnectionAvailable(conn *ec2.EC2, id string) error {
	// Wait for the connection to become available. This has an obscenely
	// high default timeout because AWS VPN connections are notoriously
	// slow at coming up or going down. There's also no point in checking
	// more frequently than every ten seconds.
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.VpnStatePending},
		Target:     []string{ec2.VpnStateAvailable},
		Refresh:    vpnConnectionRefreshFunc(conn, id),
		Timeout:    40 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}

func waitForEc2VpnConnectionAvailableWhenModifying(conn *ec2.EC2, id string) error {
	// Wait for the connection to become available. This has an obscenely
	// high default timeout because AWS VPN connections are notoriously
	// slow at coming up or going down. There's also no point in checking
	// more frequently than every ten seconds.
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"modifying"}, // VPN state modifying const is not available in SDK
		Target:     []string{ec2.VpnStateAvailable},
		Refresh:    vpnConnectionRefreshFunc(conn, id),
		Timeout:    40 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}

func waitForEc2VpnConnectionDeletion(conn *ec2.EC2, id string) error {
	// These things can take quite a while to tear themselves down and any
	// attempt to modify resources they reference (e.g. CustomerGateways or
	// VPN Gateways) before deletion will result in an error. Furthermore,
	// they don't just disappear. The go into "deleted" state. We need to
	// wait to ensure any other modifications the user might make to their
	// VPC stack can safely run.
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.VpnStateDeleting},
		Target:     []string{ec2.VpnStateDeleted},
		Refresh:    vpnConnectionRefreshFunc(conn, id),
		Timeout:    30 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}

func xmlConfigToTunnelInfo(xmlConfig string) (*TunnelInfo, error) {
	var vpnConfig XmlVpnConnectionConfig
	if err := xml.Unmarshal([]byte(xmlConfig), &vpnConfig); err != nil {
		return nil, fmt.Errorf("Error Unmarshalling XML: %s", err)
	}

	// don't expect consistent ordering from the XML
	sort.Sort(vpnConfig)

	tunnelInfo := TunnelInfo{
		Tunnel1Address:          vpnConfig.Tunnels[0].OutsideAddress,
		Tunnel1PreSharedKey:     vpnConfig.Tunnels[0].PreSharedKey,
		Tunnel1CgwInsideAddress: vpnConfig.Tunnels[0].CgwInsideAddress,
		Tunnel1VgwInsideAddress: vpnConfig.Tunnels[0].VgwInsideAddress,
		Tunnel1BGPASN:           vpnConfig.Tunnels[0].BGPASN,
		Tunnel1BGPHoldTime:      vpnConfig.Tunnels[0].BGPHoldTime,
		Tunnel2Address:          vpnConfig.Tunnels[1].OutsideAddress,
		Tunnel2PreSharedKey:     vpnConfig.Tunnels[1].PreSharedKey,
		Tunnel2CgwInsideAddress: vpnConfig.Tunnels[1].CgwInsideAddress,
		Tunnel2VgwInsideAddress: vpnConfig.Tunnels[1].VgwInsideAddress,
		Tunnel2BGPASN:           vpnConfig.Tunnels[1].BGPASN,
		Tunnel2BGPHoldTime:      vpnConfig.Tunnels[1].BGPHoldTime,
	}

	return &tunnelInfo, nil
}

func validateVpnConnectionTunnelPreSharedKey() schema.SchemaValidateFunc {
	return validation.All(
		validation.StringLenBetween(8, 64),
		validation.StringDoesNotMatch(regexp.MustCompile(`^0`), "cannot start with zero character"),
		validation.StringMatch(regexp.MustCompile(`^[0-9a-zA-Z_.]+$`), "can only contain alphanumeric, period and underscore characters"),
	)
}

// https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_VpnTunnelOptionsSpecification.html
// https://docs.aws.amazon.com/vpn/latest/s2svpn/VPNTunnels.html
func validateVpnConnectionTunnelInsideCIDR() schema.SchemaValidateFunc {
	disallowedCidrs := []string{
		"169.254.0.0/30",
		"169.254.1.0/30",
		"169.254.2.0/30",
		"169.254.3.0/30",
		"169.254.4.0/30",
		"169.254.5.0/30",
		"169.254.169.252/30",
	}

	return validation.All(
		validation.IsCIDRNetwork(30, 30),
		validation.StringMatch(regexp.MustCompile(`^169\.254\.`), "must be within 169.254.0.0/16"),
		validation.StringNotInSlice(disallowedCidrs, false),
	)
}

func validateVpnConnectionTunnelInsideIpv6CIDR() schema.SchemaValidateFunc {
	return validation.All(
		validation.IsCIDRNetwork(126, 126),
		validation.StringMatch(regexp.MustCompile(`^fd00:`), "must be within fd00::/8"),
	)
}

func validateLocalIpv4NetworkCidr() schema.SchemaValidateFunc {
	return validation.All(
		validation.IsCIDRNetwork(32, 32),
	)
}

func validateLocalIpv6NetworkCidr() schema.SchemaValidateFunc {
	return validation.All(
		validation.IsCIDRNetwork(128, 128),
	)
}

func validateVpnConnectionTunnelDpdTimeoutAction() schema.SchemaValidateFunc {
	allowedDpdTimeoutActions := []string{
		"clear",
		"none",
		"restart",
	}

	return validation.All(
		validation.StringInSlice(allowedDpdTimeoutActions, false),
	)
}

func validateTunnelInsideIPVersion() schema.SchemaValidateFunc {
	allowedIPVersions := []string{
		"ipv4",
		"ipv6",
	}

	return validation.All(
		validation.StringInSlice(allowedIPVersions, false),
	)
}

func validateVpnConnectionTunnelDpdTimeoutSeconds() schema.SchemaValidateFunc {
	return validation.All(
		//validation.IntBetween(0, 30)
		validation.IntAtLeast(30), // Must be 30 or higher
	)
}

func validateVpnConnectionTunnelPhase1LifetimeSeconds() schema.SchemaValidateFunc {
	return validation.All(
		validation.IntBetween(900, 28800),
	)
}

func validateVpnConnectionTunnelPhase2LifetimeSeconds() schema.SchemaValidateFunc {
	return validation.All(
		validation.IntBetween(900, 3600),
	)
}

func validateVpnConnectionTunnelRekeyFuzzPercentage() schema.SchemaValidateFunc {
	return validation.All(
		validation.IntBetween(0, 100),
	)
}

func validateVpnConnectionTunnelRekeyMarginTimeSeconds() schema.SchemaValidateFunc {
	return validation.All(
		validation.IntBetween(60, 1800),
	)
}

func validateVpnConnectionTunnelReplayWindowSize() schema.SchemaValidateFunc {
	return validation.All(
		validation.IntBetween(64, 2048),
	)
}

func validateVpnConnectionTunnelStartupAction() schema.SchemaValidateFunc {
	allowedStartupAction := []string{
		"add",
		"start",
	}

	return validation.All(
		validation.StringInSlice(allowedStartupAction, false),
	)
}
