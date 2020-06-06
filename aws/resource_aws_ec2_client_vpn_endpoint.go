package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsEc2ClientVpnEndpoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2ClientVpnEndpointCreate,
		Read:   resourceAwsEc2ClientVpnEndpointRead,
		Delete: resourceAwsEc2ClientVpnEndpointDelete,
		Update: resourceAwsEc2ClientVpnEndpointUpdate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"client_cidr_block": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsCIDR,
			},
			"dns_servers": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"server_certificate_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"split_tunnel": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"transport_protocol": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  ec2.TransportProtocolUdp,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.TransportProtocolTcp,
					ec2.TransportProtocolUdp,
				}, false),
			},
			"authentication_options": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 2,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								ec2.ClientVpnAuthenticationTypeCertificateAuthentication,
								ec2.ClientVpnAuthenticationTypeDirectoryServiceAuthentication,
							}, false),
						},
						"active_directory_id": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"root_certificate_chain_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},
			"connection_log_options": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cloudwatch_log_group": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"cloudwatch_log_stream": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsEc2ClientVpnEndpointCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	req := &ec2.CreateClientVpnEndpointInput{
		ClientCidrBlock:      aws.String(d.Get("client_cidr_block").(string)),
		ServerCertificateArn: aws.String(d.Get("server_certificate_arn").(string)),
		TransportProtocol:    aws.String(d.Get("transport_protocol").(string)),
		SplitTunnel:          aws.Bool(d.Get("split_tunnel").(bool)),
		TagSpecifications:    ec2TagSpecificationsFromMap(d.Get("tags").(map[string]interface{}), ec2.ResourceTypeClientVpnEndpoint),
	}

	if v, ok := d.GetOk("description"); ok {
		req.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("dns_servers"); ok {
		req.DnsServers = expandStringList(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("authentication_options"); ok {
		authOptions := v.([]interface{})
		authRequests := make([]*ec2.ClientVpnAuthenticationRequest, 0, len(authOptions))

		for _, authOpt := range authOptions {
			auth := authOpt.(map[string]interface{})

			authReq := expandEc2ClientVpnAuthenticationRequest(auth)
			authRequests = append(authRequests, authReq)
		}
		req.AuthenticationOptions = authRequests
	}

	if v, ok := d.GetOk("connection_log_options"); ok {
		connLogSet := v.([]interface{})
		attrs := connLogSet[0].(map[string]interface{})

		connLogReq := &ec2.ConnectionLogOptions{
			Enabled: aws.Bool(attrs["enabled"].(bool)),
		}

		if attrs["enabled"].(bool) && attrs["cloudwatch_log_group"].(string) != "" {
			connLogReq.CloudwatchLogGroup = aws.String(attrs["cloudwatch_log_group"].(string))
		}

		if attrs["enabled"].(bool) && attrs["cloudwatch_log_stream"].(string) != "" {
			connLogReq.CloudwatchLogStream = aws.String(attrs["cloudwatch_log_stream"].(string))
		}

		req.ConnectionLogOptions = connLogReq
	}

	resp, err := conn.CreateClientVpnEndpoint(req)

	if err != nil {
		return fmt.Errorf("Error creating Client VPN endpoint: %s", err)
	}

	d.SetId(*resp.ClientVpnEndpointId)

	return resourceAwsEc2ClientVpnEndpointRead(d, meta)
}

func resourceAwsEc2ClientVpnEndpointRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	var err error

	result, err := conn.DescribeClientVpnEndpoints(&ec2.DescribeClientVpnEndpointsInput{
		ClientVpnEndpointIds: []*string{aws.String(d.Id())},
	})

	if isAWSErr(err, "InvalidClientVpnAssociationId.NotFound", "") || isAWSErr(err, "InvalidClientVpnEndpointId.NotFound", "") {
		log.Printf("[WARN] EC2 Client VPN Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error reading Client VPN endpoint: %s", err)
	}

	if result == nil || len(result.ClientVpnEndpoints) == 0 || result.ClientVpnEndpoints[0] == nil {
		log.Printf("[WARN] EC2 Client VPN Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if result.ClientVpnEndpoints[0].Status != nil && aws.StringValue(result.ClientVpnEndpoints[0].Status.Code) == ec2.ClientVpnEndpointStatusCodeDeleted {
		log.Printf("[WARN] EC2 Client VPN Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("description", result.ClientVpnEndpoints[0].Description)
	d.Set("client_cidr_block", result.ClientVpnEndpoints[0].ClientCidrBlock)
	d.Set("server_certificate_arn", result.ClientVpnEndpoints[0].ServerCertificateArn)
	d.Set("transport_protocol", result.ClientVpnEndpoints[0].TransportProtocol)
	d.Set("dns_name", result.ClientVpnEndpoints[0].DnsName)
	d.Set("dns_servers", result.ClientVpnEndpoints[0].DnsServers)

	if result.ClientVpnEndpoints[0].Status != nil {
		d.Set("status", result.ClientVpnEndpoints[0].Status.Code)
	}

	d.Set("split_tunnel", result.ClientVpnEndpoints[0].SplitTunnel)

	err = d.Set("authentication_options", flattenAuthOptsConfig(result.ClientVpnEndpoints[0].AuthenticationOptions))
	if err != nil {
		return fmt.Errorf("error setting authentication_options: %s", err)
	}

	err = d.Set("connection_log_options", flattenConnLoggingConfig(result.ClientVpnEndpoints[0].ConnectionLogOptions))
	if err != nil {
		return fmt.Errorf("error setting connection_log_options: %s", err)
	}

	err = d.Set("tags", keyvaluetags.Ec2KeyValueTags(result.ClientVpnEndpoints[0].Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map())
	if err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "ec2",
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("client-vpn-endpoint/%s", d.Id()),
	}.String()

	d.Set("arn", arn)

	return nil
}

func resourceAwsEc2ClientVpnEndpointDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	_, err := conn.DeleteClientVpnEndpoint(&ec2.DeleteClientVpnEndpointInput{
		ClientVpnEndpointId: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("Error deleting Client VPN endpoint: %s", err)
	}

	return nil
}

func resourceAwsEc2ClientVpnEndpointUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	req := &ec2.ModifyClientVpnEndpointInput{
		ClientVpnEndpointId: aws.String(d.Id()),
	}

	if d.HasChange("description") {
		req.Description = aws.String(d.Get("description").(string))
	}

	if d.HasChange("dns_servers") {
		dnsValue := expandStringList(d.Get("dns_servers").(*schema.Set).List())
		var enabledValue *bool

		if len(dnsValue) > 0 {
			enabledValue = aws.Bool(true)
		} else {
			enabledValue = aws.Bool(false)
		}

		dnsMod := &ec2.DnsServersOptionsModifyStructure{
			CustomDnsServers: dnsValue,
			Enabled:          enabledValue,
		}
		req.DnsServers = dnsMod
	}

	if d.HasChange("server_certificate_arn") {
		req.ServerCertificateArn = aws.String(d.Get("server_certificate_arn").(string))
	}

	if d.HasChange("split_tunnel") {
		req.SplitTunnel = aws.Bool(d.Get("split_tunnel").(bool))
	}

	if d.HasChange("connection_log_options") {
		if v, ok := d.GetOk("connection_log_options"); ok {
			connSet := v.([]interface{})
			attrs := connSet[0].(map[string]interface{})

			connReq := &ec2.ConnectionLogOptions{
				Enabled: aws.Bool(attrs["enabled"].(bool)),
			}

			if attrs["enabled"].(bool) && attrs["cloudwatch_log_group"].(string) != "" {
				connReq.CloudwatchLogGroup = aws.String(attrs["cloudwatch_log_group"].(string))
			}

			if attrs["enabled"].(bool) && attrs["cloudwatch_log_stream"].(string) != "" {
				connReq.CloudwatchLogStream = aws.String(attrs["cloudwatch_log_stream"].(string))
			}

			req.ConnectionLogOptions = connReq
		}
	}

	if _, err := conn.ModifyClientVpnEndpoint(req); err != nil {
		return fmt.Errorf("Error modifying Client VPN endpoint: %s", err)
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.Ec2UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Client VPN Endpoint (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsEc2ClientVpnEndpointRead(d, meta)
}

func flattenConnLoggingConfig(lopts *ec2.ConnectionLogResponseOptions) []map[string]interface{} {
	m := make(map[string]interface{})
	if lopts.CloudwatchLogGroup != nil {
		m["cloudwatch_log_group"] = *lopts.CloudwatchLogGroup
	}
	if lopts.CloudwatchLogStream != nil {
		m["cloudwatch_log_stream"] = *lopts.CloudwatchLogStream
	}
	m["enabled"] = *lopts.Enabled
	return []map[string]interface{}{m}
}

func flattenAuthOptsConfig(aopts []*ec2.ClientVpnAuthentication) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(aopts))
	for _, aopt := range aopts {
		r := map[string]interface{}{
			"type": aws.StringValue(aopt.Type),
		}
		if aopt.MutualAuthentication != nil {
			r["root_certificate_chain_arn"] = aws.StringValue(aopt.MutualAuthentication.ClientRootCertificateChain)
		}
		if aopt.ActiveDirectory != nil {
			r["active_directory_id"] = aws.StringValue(aopt.ActiveDirectory.DirectoryId)
		}
		result = append([]map[string]interface{}{r}, result...)
	}
	return result
}

func expandEc2ClientVpnAuthenticationRequest(data map[string]interface{}) *ec2.ClientVpnAuthenticationRequest {
	req := &ec2.ClientVpnAuthenticationRequest{
		Type: aws.String(data["type"].(string)),
	}

	if data["type"].(string) == ec2.ClientVpnAuthenticationTypeCertificateAuthentication {
		req.MutualAuthentication = &ec2.CertificateAuthenticationRequest{
			ClientRootCertificateChainArn: aws.String(data["root_certificate_chain_arn"].(string)),
		}
	}

	if data["type"].(string) == ec2.ClientVpnAuthenticationTypeDirectoryServiceAuthentication {
		req.ActiveDirectory = &ec2.DirectoryServiceAuthenticationRequest{
			DirectoryId: aws.String(data["active_directory_id"].(string)),
		}
	}

	return req
}
