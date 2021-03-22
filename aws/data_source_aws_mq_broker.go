package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mq"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsMqBroker() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsmQBrokerRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authentication_strategy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_minor_version_upgrade": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"broker_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"broker_name"},
			},
			"broker_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"broker_id"},
			},
			"configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"revision": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"deployment_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encryption_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kms_key_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"use_aws_owned_key": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"engine_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"host_instance_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instances": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"console_url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"endpoints": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"ldap_server_metadata": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hosts": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"role_base": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"role_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"role_search_matching": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"role_search_subtree": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"service_account_password": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service_account_username": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"user_base": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"user_role_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"user_search_matching": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"user_search_subtree": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"logs": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				// Ignore missing configuration block
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == "1" && new == "0" {
						return true
					}
					return false
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"general": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"audit": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"maintenance_window_start_time": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"day_of_week": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"time_of_day": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"time_zone": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"publicly_accessible": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"security_groups": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"storage_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"tags": tagsSchemaComputed(),
			"user": {
				Type:     schema.TypeSet,
				Computed: true,
				Set:      resourceAwsMqUserHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"console_access": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"groups": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
							Computed: true,
						},
						"username": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAwsmQBrokerRead(d *schema.ResourceData, meta interface{}) error {
	if brokerId, ok := d.GetOk("broker_id"); ok {
		d.SetId(brokerId.(string))
	} else {
		conn := meta.(*AWSClient).mqconn
		brokerName := d.Get("broker_name").(string)

		input := &mq.ListBrokersInput{}

		err := conn.ListBrokersPages(input, func(page *mq.ListBrokersResponse, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, brokerSummary := range page.BrokerSummaries {
				if brokerSummary == nil {
					continue
				}

				if aws.StringValue(brokerSummary.BrokerName) == brokerName {
					d.Set("broker_id", brokerSummary.BrokerId)
					d.SetId(aws.StringValue(brokerSummary.BrokerId))
					return false
				}
			}

			return !lastPage
		})

		if err != nil {
			return fmt.Errorf("error listing MQ Brokers: %w", err)
		}

		if d.Id() == "" {
			return fmt.Errorf("Failed to determine mq broker: %s", brokerName)
		}
	}

	return resourceAwsMqBrokerRead(d, meta)
}
