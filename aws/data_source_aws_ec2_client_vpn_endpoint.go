package aws

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws/arn"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsEc2ClientVpnEndpoint() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEc2ClientVpnEndpointRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"authentication_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"active_directory_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"root_certificate_chain_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"client_cidr_block": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_log_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cloudwatch_log_group": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cloudwatch_log_stream": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_servers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"filter": dataSourceFiltersSchema(),
			"security_group_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"server_certificate_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"split_tunnel": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"tags": tagsSchemaComputed(),
			"transport_protocol": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpn_port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"vpn_protocol": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsEc2ClientVpnEndpointRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeClientVpnEndpointsInput{}

	if v, ok := d.GetOk("filter"); ok {
		input.Filters = buildAwsDataSourceFilters(v.(*schema.Set))
	}

	if v, ok := d.GetOk("id"); ok {
		input.ClientVpnEndpointIds = []*string{aws.String(v.(string))}
	}

	log.Printf("[DEBUG] Reading EC2 Transit Gateways: %s", input)

	resp, err := conn.DescribeClientVpnEndpoints(input)
	if err != nil {
		return fmt.Errorf("error getting Client Vpn Endpoint: %v", err)
	}

	if resp == nil || len(resp.ClientVpnEndpoints) == 0 {
		return errors.New("error reading EC2 Client Vpn Endpoint: no results found")
	}

	if len(resp.ClientVpnEndpoints) > 1 {
		return errors.New("error reading EC2 Client Vpn Endpoint: multiple results found, try adjusting search criteria")
	}

	endpoint := resp.ClientVpnEndpoints[0]

	d.SetId(aws.StringValue(endpoint.ClientVpnEndpointId))
	d.Set("dns_name", endpoint.DnsName)
	d.Set("dns_servers", endpoint.DnsServers)
	d.Set("client_cidr_block", endpoint.ClientCidrBlock)
	d.Set("description", endpoint.Description)
	d.Set("server_certificate_arn", endpoint.ServerCertificateArn)
	d.Set("split_tunnel", endpoint.SplitTunnel)
	d.Set("transport_protocol", endpoint.TransportProtocol)
	d.Set("security_group_ids", endpoint.SecurityGroupIds)
	d.Set("vpc_id", endpoint.VpcId)
	d.Set("vpn_port", endpoint.VpnPort)
	d.Set("vpn_protocol", endpoint.VpnProtocol)

	err = d.Set("authentication_options", flattenAuthOptsConfig(endpoint.AuthenticationOptions))
	if err != nil {
		return fmt.Errorf("error setting authentication_options: %s", err)
	}

	err = d.Set("connection_log_options", flattenConnLoggingConfig(endpoint.ConnectionLogOptions))
	if err != nil {
		return fmt.Errorf("error setting connection_log_options: %s", err)
	}

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(endpoint.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("Error setting tags: %s", err)
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
