package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsElasticSearchDomain() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsElasticSearchDomainRead,

		Schema: map[string]*schema.Schema{
			"access_policies": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"advanced_options": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kibana_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ebs_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ebs_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"iops": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"volume_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"volume_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"encryption_at_rest": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"kms_key_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"node_to_node_encryption": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"cluster_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dedicated_master_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"dedicated_master_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"dedicated_master_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"instance_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"instance_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"zone_awareness_config": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"availability_zone_count": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
						"zone_awareness_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"warm_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"warm_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"warm_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"snapshot_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"automated_snapshot_start_hour": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"vpc_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"availability_zones": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							//Set:      schema.HashString,
						},
						"security_group_ids": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"log_publishing_options": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"log_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cloudwatch_log_group_arn": {
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
			"elasticsearch_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cognito_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"user_pool_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"identity_pool_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"role_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"created": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"deleted": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"processing": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsElasticSearchDomainRead(d *schema.ResourceData, meta interface{}) error {
	esconn := meta.(*AWSClient).esconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	req := &elasticsearchservice.DescribeElasticsearchDomainInput{
		DomainName: aws.String(d.Get("domain_name").(string)),
	}

	resp, err := esconn.DescribeElasticsearchDomain(req)
	if err != nil {
		return fmt.Errorf("error querying elasticsearch_domain: %s", err)
	}

	if resp.DomainStatus == nil {
		return fmt.Errorf("your query returned no results")
	}

	ds := resp.DomainStatus

	d.SetId(*ds.ARN)

	if ds.AccessPolicies != nil && *ds.AccessPolicies != "" {
		policies, err := structure.NormalizeJsonString(*ds.AccessPolicies)
		if err != nil {
			return fmt.Errorf("access policies contain an invalid JSON: %s", err)
		}
		d.Set("access_policies", policies)
	}

	if err := d.Set("advanced_options", pointersMapToStringList(ds.AdvancedOptions)); err != nil {
		return fmt.Errorf("error setting advanced_options: %s", err)
	}

	d.Set("arn", ds.ARN)
	d.Set("domain_id", ds.DomainId)
	d.Set("endpoint", ds.Endpoint)
	d.Set("kibana_endpoint", getKibanaEndpoint(d))

	if err := d.Set("ebs_options", flattenESEBSOptions(ds.EBSOptions)); err != nil {
		return fmt.Errorf("error setting ebs_options: %s", err)
	}

	if err := d.Set("encryption_at_rest", flattenESEncryptAtRestOptions(ds.EncryptionAtRestOptions)); err != nil {
		return fmt.Errorf("error setting encryption_at_rest: %s", err)
	}

	if err := d.Set("node_to_node_encryption", flattenESNodeToNodeEncryptionOptions(ds.NodeToNodeEncryptionOptions)); err != nil {
		return fmt.Errorf("error setting node_to_node_encryption: %s", err)
	}

	if err := d.Set("cluster_config", flattenESClusterConfig(ds.ElasticsearchClusterConfig)); err != nil {
		return fmt.Errorf("error setting cluster_config: %s", err)
	}

	if err := d.Set("snapshot_options", flattenESSnapshotOptions(ds.SnapshotOptions)); err != nil {
		return fmt.Errorf("error setting snapshot_options: %s", err)
	}

	if ds.VPCOptions != nil {
		if err := d.Set("vpc_options", flattenESVPCDerivedInfo(ds.VPCOptions)); err != nil {
			return fmt.Errorf("error setting vpc_options: %s", err)
		}

		endpoints := pointersMapToStringList(ds.Endpoints)
		if err := d.Set("endpoint", endpoints["vpc"]); err != nil {
			return fmt.Errorf("error setting endpoint: %s", err)
		}
		d.Set("kibana_endpoint", getKibanaEndpoint(d))
		if ds.Endpoint != nil {
			return fmt.Errorf("%q: Elasticsearch domain in VPC expected to have null Endpoint value", d.Id())
		}
	} else {
		if ds.Endpoint != nil {
			d.Set("endpoint", aws.StringValue(ds.Endpoint))
			d.Set("kibana_endpoint", getKibanaEndpoint(d))
		}
		if ds.Endpoints != nil {
			return fmt.Errorf("%q: Elasticsearch domain not in VPC expected to have null Endpoints value", d.Id())
		}
	}

	if ds.LogPublishingOptions != nil {
		m := make([]map[string]interface{}, 0)
		for k, val := range ds.LogPublishingOptions {
			mm := map[string]interface{}{}
			mm["log_type"] = k
			if val.CloudWatchLogsLogGroupArn != nil {
				mm["cloudwatch_log_group_arn"] = aws.StringValue(val.CloudWatchLogsLogGroupArn)
			}
			mm["enabled"] = aws.BoolValue(val.Enabled)
			m = append(m, mm)
		}
		d.Set("log_publishing_options", m)
	}

	d.Set("elasticsearch_version", ds.ElasticsearchVersion)

	if err := d.Set("cognito_options", flattenESCognitoOptions(ds.CognitoOptions)); err != nil {
		return fmt.Errorf("error setting cognito_options: %s", err)
	}

	d.Set("created", ds.Created)
	d.Set("deleted", ds.Deleted)

	d.Set("processing", ds.Processing)

	tags, err := keyvaluetags.ElasticsearchserviceListTags(esconn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for Elasticsearch Cluster (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}
