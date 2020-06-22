package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsVpcEndpoint() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsVpcEndpointRead,

		Schema: map[string]*schema.Schema{
			"cidr_blocks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"dns_entry": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dns_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"hosted_zone_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"filter": ec2CustomFiltersSchema(),
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"network_interface_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"prefix_list_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_dns_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"requester_managed": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"route_table_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"service_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"tags": tagsSchemaComputed(),
			"vpc_endpoint_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsVpcEndpointRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	req := &ec2.DescribeVpcEndpointsInput{}

	if id, ok := d.GetOk("id"); ok {
		req.VpcEndpointIds = aws.StringSlice([]string{id.(string)})
	}

	req.Filters = buildEC2AttributeFilterList(
		map[string]string{
			"vpc-endpoint-state": d.Get("state").(string),
			"vpc-id":             d.Get("vpc_id").(string),
			"service-name":       d.Get("service_name").(string),
		},
	)
	req.Filters = append(req.Filters, buildEC2TagFilterList(
		keyvaluetags.New(d.Get("tags").(map[string]interface{})).Ec2Tags(),
	)...)
	req.Filters = append(req.Filters, buildEC2CustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)
	if len(req.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		req.Filters = nil
	}

	log.Printf("[DEBUG] Reading VPC Endpoint: %s", req)
	respVpce, err := conn.DescribeVpcEndpoints(req)
	if err != nil {
		return fmt.Errorf("error reading VPC Endpoint: %s", err)
	}
	if respVpce == nil || len(respVpce.VpcEndpoints) == 0 {
		return fmt.Errorf("no matching VPC Endpoint found")
	}
	if len(respVpce.VpcEndpoints) > 1 {
		return fmt.Errorf("multiple VPC Endpoints matched; use additional constraints to reduce matches to a single VPC Endpoint")
	}

	vpce := respVpce.VpcEndpoints[0]
	d.SetId(aws.StringValue(vpce.VpcEndpointId))

	serviceName := aws.StringValue(vpce.ServiceName)
	d.Set("service_name", serviceName)
	d.Set("state", vpce.State)
	d.Set("vpc_id", vpce.VpcId)

	respPl, err := conn.DescribePrefixLists(&ec2.DescribePrefixListsInput{
		Filters: buildEC2AttributeFilterList(map[string]string{
			"prefix-list-name": serviceName,
		}),
	})
	if err != nil {
		return fmt.Errorf("error reading Prefix List (%s): %s", serviceName, err)
	}
	if respPl == nil || len(respPl.PrefixLists) == 0 {
		d.Set("cidr_blocks", []interface{}{})
	} else if len(respPl.PrefixLists) > 1 {
		return fmt.Errorf("multiple prefix lists associated with the service name '%s'. Unexpected", serviceName)
	} else {
		pl := respPl.PrefixLists[0]

		d.Set("prefix_list_id", pl.PrefixListId)
		err = d.Set("cidr_blocks", flattenStringList(pl.Cidrs))
		if err != nil {
			return fmt.Errorf("error setting cidr_blocks: %s", err)
		}
	}

	err = d.Set("dns_entry", flattenVpcEndpointDnsEntries(vpce.DnsEntries))
	if err != nil {
		return fmt.Errorf("error setting dns_entry: %s", err)
	}
	err = d.Set("network_interface_ids", flattenStringSet(vpce.NetworkInterfaceIds))
	if err != nil {
		return fmt.Errorf("error setting network_interface_ids: %s", err)
	}
	d.Set("owner_id", vpce.OwnerId)
	policy, err := structure.NormalizeJsonString(aws.StringValue(vpce.PolicyDocument))
	if err != nil {
		return fmt.Errorf("policy contains an invalid JSON: %s", err)
	}
	d.Set("policy", policy)
	d.Set("private_dns_enabled", vpce.PrivateDnsEnabled)
	err = d.Set("route_table_ids", flattenStringSet(vpce.RouteTableIds))
	if err != nil {
		return fmt.Errorf("error setting route_table_ids: %s", err)
	}
	d.Set("requester_managed", vpce.RequesterManaged)
	err = d.Set("security_group_ids", flattenVpcEndpointSecurityGroupIds(vpce.Groups))
	if err != nil {
		return fmt.Errorf("error setting security_group_ids: %s", err)
	}
	err = d.Set("subnet_ids", flattenStringSet(vpce.SubnetIds))
	if err != nil {
		return fmt.Errorf("error setting subnet_ids: %s", err)
	}
	err = d.Set("tags", keyvaluetags.Ec2KeyValueTags(vpce.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map())
	if err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}
	// VPC endpoints don't have types in GovCloud, so set type to default if empty
	if vpceType := aws.StringValue(vpce.VpcEndpointType); vpceType == "" {
		d.Set("vpc_endpoint_type", ec2.VpcEndpointTypeGateway)
	} else {
		d.Set("vpc_endpoint_type", vpceType)
	}

	return nil
}
