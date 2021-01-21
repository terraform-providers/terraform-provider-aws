package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsRoute53ResolverEndpoint() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsRoute53ResolverEndpointRead,

		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"values": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},

			"direction": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"resolver_endpoint_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_addresses": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Set:      schema.HashString,
			},
		},
	}
}

func dataSourceAwsRoute53ResolverEndpointRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn
	req := &route53resolver.ListResolverEndpointsInput{}

	resolvers := make([]*route53resolver.ResolverEndpoint, 0)

	rID, rIDOk := d.GetOk("resolver_endpoint_id")
	filters, filtersOk := d.GetOk("filter")

	if filtersOk {
		req.Filters = buildR53ResolverTagFilters(filters.(*schema.Set))
	}

	for {
		resp, err := conn.ListResolverEndpoints(req)

		if err != nil {
			return fmt.Errorf("Error Reading Route53 Resolver Endpoints: %s", req)
		}

		if len(resp.ResolverEndpoints) == 0 && filtersOk {
			return fmt.Errorf("Your query returned no results. Please change your search criteria and try again")
		}

		if len(resp.ResolverEndpoints) > 1 && !rIDOk {
			return fmt.Errorf("Your query returned more than one resolver. Please change your search criteria and try again")
		}

		if rIDOk {
			for _, r := range resp.ResolverEndpoints {
				if aws.StringValue(r.Id) == rID {
					resolvers = append(resolvers, r)
					break
				}
			}
		} else {
			resolvers = append(resolvers, resp.ResolverEndpoints[0])
		}

		if len(resolvers) == 0 {
			return fmt.Errorf("The ID provided could not be found")
		}

		resolver := resolvers[0]

		d.SetId(aws.StringValue(resolver.Id))
		d.Set("resolver_endpoint_id", resolver.Id)
		d.Set("arn", aws.StringValue(resolver.Arn))
		d.Set("status", aws.StringValue(resolver.Status))
		d.Set("name", aws.StringValue(resolver.Name))
		d.Set("vpc_id", aws.StringValue(resolver.HostVPCId))
		d.Set("direction", aws.StringValue(resolver.Direction))

		if resp.NextToken == nil {
			break
		}

		req.NextToken = resp.NextToken
	}

	params := &route53resolver.ListResolverEndpointIpAddressesInput{
		ResolverEndpointId: aws.String(d.Id()),
	}

	ipAddresses := []interface{}{}

	for {
		ip, err := conn.ListResolverEndpointIpAddresses(params)

		if err != nil {
			return fmt.Errorf("error getting Route53 Resolver endpoint (%s) IP Addresses: %s", d.Id(), err)
		}

		for _, vIPAddresses := range ip.IpAddresses {
			ipAddresses = append(ipAddresses, aws.StringValue(vIPAddresses.Ip))
		}

		d.Set("ip_addresses", ipAddresses)

		if ip.NextToken == nil {
			break
		}

		params.NextToken = ip.NextToken
	}

	return nil
}

func buildR53ResolverTagFilters(set *schema.Set) []*route53resolver.Filter {
	var filters []*route53resolver.Filter

	for _, v := range set.List() {
		m := v.(map[string]interface{})
		var filterValues []*string
		for _, e := range m["values"].([]interface{}) {
			filterValues = append(filterValues, aws.String(e.(string)))
		}
		filters = append(filters, &route53resolver.Filter{
			Name:   aws.String(m["name"].(string)),
			Values: filterValues,
		})
	}

	return filters
}
