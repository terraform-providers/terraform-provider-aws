package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsElb() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsElbRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"access_logs": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"interval": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"bucket": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"bucket_prefix": {
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

			"availability_zones": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Set:      schema.HashString,
			},

			"connection_draining": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"connection_draining_timeout": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"cross_zone_load_balancing": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"health_check": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"healthy_threshold": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"unhealthy_threshold": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"target": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"interval": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"timeout": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},

			"idle_timeout": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"instances": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Set:      schema.HashString,
			},

			"internal": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"listener": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"instance_port": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"instance_protocol": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"lb_port": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"lb_protocol": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"ssl_certificate_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Set: resourceAwsElbListenerHash,
			},

			"security_groups": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Set:      schema.HashString,
			},

			"source_security_group": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"source_security_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"subnets": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Set:      schema.HashString,
			},

			"tags": tagsSchemaComputed(),

			"zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsElbRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elbconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	lbName := d.Get("name").(string)

	input := &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: aws.StringSlice([]string{lbName}),
	}

	log.Printf("[DEBUG] Reading ELB: %s", input)
	resp, err := conn.DescribeLoadBalancers(input)
	if err != nil {
		return fmt.Errorf("error retrieving LB: %w", err)
	}
	if len(resp.LoadBalancerDescriptions) != 1 {
		return fmt.Errorf("search returned %d results, please revise so only one is returned", len(resp.LoadBalancerDescriptions))
	}
	d.SetId(aws.StringValue(resp.LoadBalancerDescriptions[0].LoadBalancerName))

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "elasticloadbalancing",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("loadbalancer/%s", aws.StringValue(resp.LoadBalancerDescriptions[0].LoadBalancerName)),
	}
	d.Set("arn", arn.String())

	lb := resp.LoadBalancerDescriptions[0]
	ec2conn := meta.(*AWSClient).ec2conn

	describeAttrsOpts := &elb.DescribeLoadBalancerAttributesInput{
		LoadBalancerName: aws.String(d.Id()),
	}
	describeAttrsResp, err := conn.DescribeLoadBalancerAttributes(describeAttrsOpts)
	if err != nil {
		return fmt.Errorf("error retrieving ELB: %w", err)
	}

	lbAttrs := describeAttrsResp.LoadBalancerAttributes

	d.Set("name", lb.LoadBalancerName)
	d.Set("dns_name", lb.DNSName)
	d.Set("zone_id", lb.CanonicalHostedZoneNameID)

	var scheme bool
	if lb.Scheme != nil {
		scheme = aws.StringValue(lb.Scheme) == "internal"
	}
	d.Set("internal", scheme)
	d.Set("availability_zones", flattenStringList(lb.AvailabilityZones))
	d.Set("instances", flattenInstances(lb.Instances))
	d.Set("listener", flattenListeners(lb.ListenerDescriptions))
	d.Set("security_groups", flattenStringList(lb.SecurityGroups))
	if lb.SourceSecurityGroup != nil {
		group := lb.SourceSecurityGroup.GroupName
		if lb.SourceSecurityGroup.OwnerAlias != nil && aws.StringValue(lb.SourceSecurityGroup.OwnerAlias) != "" {
			group = aws.String(aws.StringValue(lb.SourceSecurityGroup.OwnerAlias) + "/" + aws.StringValue(lb.SourceSecurityGroup.GroupName))
		}
		d.Set("source_security_group", group)

		// Manually look up the ELB Security Group ID, since it's not provided
		var elbVpc string
		if lb.VPCId != nil {
			elbVpc = aws.StringValue(lb.VPCId)
			sgId, err := sourceSGIdByName(ec2conn, aws.StringValue(lb.SourceSecurityGroup.GroupName), elbVpc)
			if err != nil {
				return fmt.Errorf("error looking up ELB Security Group ID: %w", err)
			} else {
				d.Set("source_security_group_id", sgId)
			}
		}
	}
	d.Set("subnets", flattenStringList(lb.Subnets))
	if lbAttrs.ConnectionSettings != nil {
		d.Set("idle_timeout", lbAttrs.ConnectionSettings.IdleTimeout)
	}
	d.Set("connection_draining", lbAttrs.ConnectionDraining.Enabled)
	d.Set("connection_draining_timeout", lbAttrs.ConnectionDraining.Timeout)
	d.Set("cross_zone_load_balancing", lbAttrs.CrossZoneLoadBalancing.Enabled)
	if lbAttrs.AccessLog != nil {
		// The AWS API does not allow users to remove access_logs, only disable them.
		// During creation of the ELB, Terraform sets the access_logs to disabled,
		// so there should not be a case where lbAttrs.AccessLog above is nil.

		// Here we do not record the remove value of access_log if:
		// - there is no access_log block in the configuration
		// - the remote access_logs are disabled
		//
		// This indicates there is no access_log in the configuration.
		// - externally added access_logs will be enabled, so we'll detect the drift
		// - locally added access_logs will be in the config, so we'll add to the
		// API/state
		// See https://github.com/hashicorp/terraform/issues/10138
		_, n := d.GetChange("access_logs")
		elbal := lbAttrs.AccessLog
		nl := n.([]interface{})
		if len(nl) == 0 && !aws.BoolValue(elbal.Enabled) {
			elbal = nil
		}
		if err := d.Set("access_logs", flattenAccessLog(elbal)); err != nil {
			return err
		}
	}

	tags, err := keyvaluetags.ElbListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for ELB (%s): %w", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	// There's only one health check, so save that to state as we
	// currently can
	if aws.StringValue(lb.HealthCheck.Target) != "" {
		d.Set("health_check", flattenHealthCheck(lb.HealthCheck))
	}

	return nil
}
