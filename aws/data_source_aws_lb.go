package aws

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsLb() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLbRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateArn,
			},

			"arn_suffix": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"internal": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"load_balancer_type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"security_groups": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},

			"subnets": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},

			"subnet_mapping": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"outpost_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"allocation_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"private_ipv4_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ipv6_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"access_logs": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"prefix": {
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

			"enable_deletion_protection": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"enable_http2": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"idle_timeout": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"drop_invalid_header_fields": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"ip_address_type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"customer_owned_ipv4_pool": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsLbRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elbv2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &elbv2.DescribeLoadBalancersInput{}

	if v, ok := d.GetOk("arn"); ok {
		input.LoadBalancerArns = aws.StringSlice([]string{v.(string)})
	} else if v, ok := d.GetOk("name"); ok {
		input.Names = aws.StringSlice([]string{v.(string)})
	}

	var results []*elbv2.LoadBalancer

	err := conn.DescribeLoadBalancersPages(input, func(page *elbv2.DescribeLoadBalancersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		results = append(results, page.LoadBalancers...)

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error retrieving LB: %w", err)
	}

	if len(results) != 1 {
		return fmt.Errorf("Search returned %d results, please revise so only one is returned", len(results))
	}

	lb := results[0]

	d.SetId(aws.StringValue(lb.LoadBalancerArn))

	d.Set("arn", lb.LoadBalancerArn)
	d.Set("arn_suffix", lbSuffixFromARN(lb.LoadBalancerArn))
	d.Set("name", lb.LoadBalancerName)
	d.Set("internal", lb.Scheme != nil && aws.StringValue(lb.Scheme) == "internal")
	d.Set("security_groups", flattenStringList(lb.SecurityGroups))
	d.Set("vpc_id", lb.VpcId)
	d.Set("zone_id", lb.CanonicalHostedZoneId)
	d.Set("dns_name", lb.DNSName)
	d.Set("ip_address_type", lb.IpAddressType)
	d.Set("load_balancer_type", lb.Type)
	d.Set("customer_owned_ipv4_pool", lb.CustomerOwnedIpv4Pool)

	if err := d.Set("subnets", flattenSubnetsFromAvailabilityZones(lb.AvailabilityZones)); err != nil {
		return fmt.Errorf("error setting subnets: %w", err)
	}

	if err := d.Set("subnet_mapping", flattenSubnetMappingsFromAvailabilityZones(lb.AvailabilityZones)); err != nil {
		return fmt.Errorf("error setting subnet_mapping: %w", err)
	}

	tags, err := keyvaluetags.Elbv2ListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for (%s): %w", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	attributesResp, err := conn.DescribeLoadBalancerAttributes(&elbv2.DescribeLoadBalancerAttributesInput{
		LoadBalancerArn: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("error retrieving LB Attributes: %w", err)
	}

	accessLogMap := map[string]interface{}{
		"bucket":  "",
		"enabled": false,
		"prefix":  "",
	}

	for _, attr := range attributesResp.Attributes {
		switch aws.StringValue(attr.Key) {
		case "access_logs.s3.enabled":
			accessLogMap["enabled"] = aws.StringValue(attr.Value) == "true"
		case "access_logs.s3.bucket":
			accessLogMap["bucket"] = aws.StringValue(attr.Value)
		case "access_logs.s3.prefix":
			accessLogMap["prefix"] = aws.StringValue(attr.Value)
		case "idle_timeout.timeout_seconds":
			timeout, err := strconv.Atoi(aws.StringValue(attr.Value))
			if err != nil {
				return fmt.Errorf("error parsing ALB timeout: %w", err)
			}
			d.Set("idle_timeout", timeout)
		case "routing.http.drop_invalid_header_fields.enabled":
			dropInvalidHeaderFieldsEnabled := aws.StringValue(attr.Value) == "true"
			d.Set("drop_invalid_header_fields", dropInvalidHeaderFieldsEnabled)
		case "deletion_protection.enabled":
			protectionEnabled := aws.StringValue(attr.Value) == "true"
			d.Set("enable_deletion_protection", protectionEnabled)
		case "routing.http2.enabled":
			http2Enabled := aws.StringValue(attr.Value) == "true"
			d.Set("enable_http2", http2Enabled)
		case "load_balancing.cross_zone.enabled":
			crossZoneLbEnabled := aws.StringValue(attr.Value) == "true"
			d.Set("enable_cross_zone_load_balancing", crossZoneLbEnabled)
		}
	}

	if err := d.Set("access_logs", []interface{}{accessLogMap}); err != nil {
		return fmt.Errorf("error setting access_logs: %w", err)
	}

	return nil
}
