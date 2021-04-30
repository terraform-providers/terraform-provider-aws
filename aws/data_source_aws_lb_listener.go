package aws

import (
	"errors"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsLbListener() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLbListenerRead,

		Schema: map[string]*schema.Schema{
			"alpn_policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"load_balancer_arn", "port"},
			},
			"certificate_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_action": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"authenticate_cognito": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"authentication_request_extra_params": {
										Type:     schema.TypeMap,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"on_unauthenticated_request": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"scope": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"session_cookie_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"session_timeout": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"user_pool_arn": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"user_pool_client_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"user_pool_domain": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"authenticate_oidc": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"authentication_request_extra_params": {
										Type:     schema.TypeMap,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"authorization_endpoint": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"client_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"client_secret": {
										Type:      schema.TypeString,
										Computed:  true,
										Sensitive: true,
									},
									"issuer": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"on_unauthenticated_request": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"scope": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"session_cookie_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"session_timeout": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"token_endpoint": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"user_info_endpoint": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"fixed_response": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"content_type": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"message_body": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"status_code": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"forward": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"stickiness": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"duration": {
													Type:     schema.TypeInt,
													Computed: true,
												},
												"enabled": {
													Type:     schema.TypeBool,
													Computed: true,
												},
											},
										},
									},
									"target_group": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"arn": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"weight": {
													Type:     schema.TypeInt,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						"order": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"redirect": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"host": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"path": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"port": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"protocol": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"query": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"status_code": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"target_group_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"load_balancer_arn": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"arn"},
			},
			"port": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"arn"},
			},
			"protocol": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ssl_policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsLbListenerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elbv2conn

	input := &elbv2.DescribeListenersInput{}

	if v, ok := d.GetOk("arn"); ok {
		input.ListenerArns = aws.StringSlice([]string{v.(string)})
	} else {
		lbArn, lbOk := d.GetOk("load_balancer_arn")
		_, portOk := d.GetOk("port")

		if !lbOk || !portOk {
			return errors.New("both load_balancer_arn and port must be set")
		}

		input.LoadBalancerArn = aws.String(lbArn.(string))
	}

	var results []*elbv2.Listener

	err := conn.DescribeListenersPages(input, func(page *elbv2.DescribeListenersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, l := range page.Listeners {
			if l == nil {
				continue
			}

			if v, ok := d.GetOk("port"); ok && v.(int) != int(aws.Int64Value(l.Port)) {
				continue
			}

			results = append(results, l)
		}

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error reading Listener: %w", err)
	}

	if len(results) != 1 {
		return fmt.Errorf("Search returned %d results, please revise so only one is returned", len(results))
	}

	listener := results[0]

	d.SetId(aws.StringValue(listener.ListenerArn))
	d.Set("arn", listener.ListenerArn)
	d.Set("load_balancer_arn", listener.LoadBalancerArn)
	d.Set("port", listener.Port)
	d.Set("protocol", listener.Protocol)
	d.Set("ssl_policy", listener.SslPolicy)

	if listener.Certificates != nil && len(listener.Certificates) == 1 && listener.Certificates[0] != nil {
		d.Set("certificate_arn", listener.Certificates[0].CertificateArn)
	}

	if listener.AlpnPolicy != nil && len(listener.AlpnPolicy) == 1 && listener.AlpnPolicy[0] != nil {
		d.Set("alpn_policy", listener.AlpnPolicy[0])
	}

	sort.Slice(listener.DefaultActions, func(i, j int) bool {
		return aws.Int64Value(listener.DefaultActions[i].Order) < aws.Int64Value(listener.DefaultActions[j].Order)
	})

	if err := d.Set("default_action", flattenLbListenerActions(d, listener.DefaultActions)); err != nil {
		return fmt.Errorf("error setting default_action: %w", err)
	}

	return nil
}
