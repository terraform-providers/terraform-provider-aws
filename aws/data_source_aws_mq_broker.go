package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mq"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
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
	conn := meta.(*AWSClient).mqconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &mq.ListBrokersInput{}

	var results []*mq.BrokerSummary

	err := conn.ListBrokersPages(input, func(page *mq.ListBrokersResponse, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, brokerSummary := range page.BrokerSummaries {
			if brokerSummary == nil {
				continue
			}

			if v, ok := d.GetOk("broker_id"); ok && v.(string) != aws.StringValue(brokerSummary.BrokerId) {
				continue
			}

			if v, ok := d.GetOk("broker_name"); ok && v.(string) != aws.StringValue(brokerSummary.BrokerName) {
				continue
			}

			results = append(results, brokerSummary)
		}

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error listing MQ Brokers: %w", err)
	}

	if len(results) != 1 {
		return fmt.Errorf("Search returned %d results, please revise so only one is returned", len(results))
	}

	brokerId := aws.StringValue(results[0].BrokerId)

	output, err := conn.DescribeBroker(&mq.DescribeBrokerInput{
		BrokerId: aws.String(brokerId),
	})

	if err != nil {
		return fmt.Errorf("error reading MQ broker (%s): %w", brokerId, err)
	}

	if output == nil {
		return fmt.Errorf("empty response while reading MQ broker (%s)", brokerId)
	}

	d.SetId(brokerId)

	d.Set("arn", output.BrokerArn)
	d.Set("authentication_strategy", output.AuthenticationStrategy)
	d.Set("auto_minor_version_upgrade", output.AutoMinorVersionUpgrade)
	d.Set("broker_id", brokerId)
	d.Set("broker_name", output.BrokerName)
	d.Set("deployment_mode", output.DeploymentMode)
	d.Set("engine_type", output.EngineType)
	d.Set("engine_version", output.EngineVersion)
	d.Set("host_instance_type", output.HostInstanceType)
	d.Set("instances", flattenMqBrokerInstances(output.BrokerInstances))
	d.Set("publicly_accessible", output.PubliclyAccessible)
	d.Set("security_groups", aws.StringValueSlice(output.SecurityGroups))
	d.Set("storage_type", output.StorageType)
	d.Set("subnet_ids", aws.StringValueSlice(output.SubnetIds))

	if err := d.Set("configuration", flattenMqConfiguration(output.Configurations)); err != nil {
		return fmt.Errorf("error setting configuration: %w", err)
	}

	if err := d.Set("encryption_options", flattenMqEncryptionOptions(output.EncryptionOptions)); err != nil {
		return fmt.Errorf("error setting encryption_options: %w", err)
	}

	var password string
	if v, ok := d.GetOk("ldap_server_metadata.0.service_account_password"); ok {
		password = v.(string)
	}

	if err := d.Set("ldap_server_metadata", flattenMQLDAPServerMetadata(output.LdapServerMetadata, password)); err != nil {
		return fmt.Errorf("error setting ldap_server_metadata: %w", err)
	}

	if err := d.Set("logs", flattenMqLogs(output.Logs)); err != nil {
		return fmt.Errorf("error setting logs: %w", err)
	}

	if err := d.Set("maintenance_window_start_time", flattenMqWeeklyStartTime(output.MaintenanceWindowStartTime)); err != nil {
		return fmt.Errorf("error setting maintenance_window_start_time: %w", err)
	}

	rawUsers, err := expandMqUsersForBroker(conn, brokerId, output.Users)

	if err != nil {
		return fmt.Errorf("error retrieving user info for MQ broker (%s): %w", brokerId, err)
	}

	if err := d.Set("user", flattenMqUsers(rawUsers, d.Get("user").(*schema.Set).List())); err != nil {
		return fmt.Errorf("error setting user: %w", err)
	}

	if err := d.Set("tags", keyvaluetags.MqKeyValueTags(output.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}
