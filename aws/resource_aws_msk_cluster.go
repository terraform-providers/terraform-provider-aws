package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsMskCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsMskClusterCreate,
		Read:   resourceAwsMskClusterRead,
		Update: resourceAwsMskClusterUpdate,
		Delete: resourceAwsMskClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers_tls": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"broker_node_group_info": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"az_distribution": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  kafka.BrokerAZDistributionDefault,
							ValidateFunc: validation.StringInSlice([]string{
								kafka.BrokerAZDistributionDefault,
							}, false),
						},
						"client_subnets": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"instance_type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"security_groups": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"ebs_volume_size": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 16384),
						},
					},
				},
			},
			"client_authentication": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"tls": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"certificate_authority_arns": {
										Type:     schema.TypeSet,
										Optional: true,
										ForceNew: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validateArn,
										},
									},
								},
							},
						},
					},
				},
			},
			"cluster_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"configuration_info": {
				Type:             schema.TypeList,
				Optional:         true,
				DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
				MaxItems:         1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
						"revision": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
					},
				},
			},
			"current_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encryption_info": {
				Type:             schema.TypeList,
				Optional:         true,
				DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
				ForceNew:         true,
				MaxItems:         1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"encryption_at_rest_kms_key_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validateArn,
						},
						"encryption_in_transit": {
							Type:             schema.TypeList,
							Optional:         true,
							DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
							ForceNew:         true,
							MaxItems:         1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"client_broker": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										Default:  kafka.ClientBrokerTlsPlaintext,
										ValidateFunc: validation.StringInSlice([]string{
											kafka.ClientBrokerPlaintext,
											kafka.ClientBrokerTlsPlaintext,
											kafka.ClientBrokerTls,
										}, false),
									},
									"in_cluster": {
										Type:     schema.TypeBool,
										Optional: true,
										ForceNew: true,
										Default:  true,
									},
								},
							},
						},
					},
				},
			},
			"enhanced_monitoring": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  kafka.EnhancedMonitoringDefault,
				ValidateFunc: validation.StringInSlice([]string{
					kafka.EnhancedMonitoringDefault,
					kafka.EnhancedMonitoringPerBroker,
					kafka.EnhancedMonitoringPerTopicPerBroker,
				}, true),
			},
			"kafka_version": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"number_of_broker_nodes": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"open_monitoring": {
				Type:             schema.TypeList,
				Optional:         true,
				DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
				MaxItems:         1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"prometheus": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"jmx_exporter": {
										Type:             schema.TypeList,
										Optional:         true,
										DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
										MaxItems:         1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"enabled_in_broker": {
													Type:     schema.TypeBool,
													Required: true,
												},
											},
										},
									},
									"node_exporter": {
										Type:             schema.TypeList,
										Optional:         true,
										DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
										MaxItems:         1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"enabled_in_broker": {
													Type:     schema.TypeBool,
													Required: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"logging_info": {
				Type:             schema.TypeList,
				Optional:         true,
				DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
				MaxItems:         1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"broker_logs": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cloudwatch_logs": {
										Type:             schema.TypeList,
										Optional:         true,
										DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
										MaxItems:         1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"enabled": {
													Type:     schema.TypeBool,
													Required: true,
												},
												"log_group": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
									"firehose": {
										Type:             schema.TypeList,
										Optional:         true,
										DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
										MaxItems:         1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"enabled": {
													Type:     schema.TypeBool,
													Required: true,
												},
												"delivery_stream": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
									"s3": {
										Type:             schema.TypeList,
										Optional:         true,
										DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
										MaxItems:         1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"enabled": {
													Type:     schema.TypeBool,
													Required: true,
												},
												"bucket": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"prefix": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"tags": tagsSchema(),
			"zookeeper_connect_string": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsMskClusterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn

	input := &kafka.CreateClusterInput{
		BrokerNodeGroupInfo:  expandMskClusterBrokerNodeGroupInfo(d.Get("broker_node_group_info").([]interface{})),
		ClientAuthentication: expandMskClusterClientAuthentication(d.Get("client_authentication").([]interface{})),
		ClusterName:          aws.String(d.Get("cluster_name").(string)),
		ConfigurationInfo:    expandMskClusterConfigurationInfo(d.Get("configuration_info").([]interface{})),
		EncryptionInfo:       expandMskClusterEncryptionInfo(d.Get("encryption_info").([]interface{})),
		EnhancedMonitoring:   aws.String(d.Get("enhanced_monitoring").(string)),
		KafkaVersion:         aws.String(d.Get("kafka_version").(string)),
		NumberOfBrokerNodes:  aws.Int64(int64(d.Get("number_of_broker_nodes").(int))),
		OpenMonitoring:       expandMskOpenMonitoring(d.Get("open_monitoring").([]interface{})),
		LoggingInfo:          expandMskLoggingInfo(d.Get("logging_info").([]interface{})),
		Tags:                 keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().KafkaTags(),
	}

	out, err := conn.CreateCluster(input)

	if err != nil {
		return fmt.Errorf("error creating MSK cluster: %s", err)
	}

	d.SetId(aws.StringValue(out.ClusterArn))

	log.Printf("[DEBUG] Waiting for MSK cluster %q to be created", d.Id())
	err = waitForMskClusterCreation(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error waiting for MSK cluster creation (%s): %s", d.Id(), err)
	}

	return resourceAwsMskClusterRead(d, meta)
}

func waitForMskClusterCreation(conn *kafka.Kafka, arn string) error {
	input := &kafka.DescribeClusterInput{
		ClusterArn: aws.String(arn),
	}
	err := resource.Retry(60*time.Minute, func() *resource.RetryError {
		out, err := conn.DescribeCluster(input)
		if err != nil {
			return resource.NonRetryableError(err)
		}
		if out.ClusterInfo != nil {
			if aws.StringValue(out.ClusterInfo.State) == kafka.ClusterStateFailed {
				return resource.NonRetryableError(fmt.Errorf("Cluster creation failed with cluster state %q", kafka.ClusterStateFailed))
			}
			if aws.StringValue(out.ClusterInfo.State) == kafka.ClusterStateActive {
				return nil
			}
		}
		return resource.RetryableError(fmt.Errorf("%q: cluster still creating", arn))
	})
	if isResourceTimeoutError(err) {
		out, err := conn.DescribeCluster(input)
		if err != nil {
			return fmt.Errorf("Error describing MSK cluster state: %s", err)
		}
		if out.ClusterInfo != nil {
			if aws.StringValue(out.ClusterInfo.State) == kafka.ClusterStateFailed {
				return fmt.Errorf("Cluster creation failed with cluster state %q", kafka.ClusterStateFailed)
			}
			if aws.StringValue(out.ClusterInfo.State) == kafka.ClusterStateActive {
				return nil
			}
		}
	}
	if err != nil {
		return fmt.Errorf("Error waiting for MSK cluster creation: %s", err)
	}
	return nil
}

func resourceAwsMskClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	out, err := conn.DescribeCluster(&kafka.DescribeClusterInput{
		ClusterArn: aws.String(d.Id()),
	})
	if err != nil {
		if isAWSErr(err, kafka.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] MSK Cluster (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("failed lookup cluster %s: %s", d.Id(), err)
	}

	brokerOut, err := conn.GetBootstrapBrokers(&kafka.GetBootstrapBrokersInput{
		ClusterArn: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("failed requesting bootstrap broker info for %q : %s", d.Id(), err)
	}

	cluster := out.ClusterInfo

	d.Set("arn", aws.StringValue(cluster.ClusterArn))
	d.Set("bootstrap_brokers", aws.StringValue(brokerOut.BootstrapBrokerString))
	d.Set("bootstrap_brokers_tls", aws.StringValue(brokerOut.BootstrapBrokerStringTls))

	if err := d.Set("broker_node_group_info", flattenMskBrokerNodeGroupInfo(cluster.BrokerNodeGroupInfo)); err != nil {
		return fmt.Errorf("error setting broker_node_group_info: %s", err)
	}

	if err := d.Set("client_authentication", flattenMskClientAuthentication(cluster.ClientAuthentication)); err != nil {
		return fmt.Errorf("error setting configuration_info: %s", err)
	}

	d.Set("cluster_name", aws.StringValue(cluster.ClusterName))

	if err := d.Set("configuration_info", flattenMskConfigurationInfo(cluster.CurrentBrokerSoftwareInfo)); err != nil {
		return fmt.Errorf("error setting configuration_info: %s", err)
	}

	d.Set("current_version", aws.StringValue(cluster.CurrentVersion))
	d.Set("enhanced_monitoring", aws.StringValue(cluster.EnhancedMonitoring))

	if err := d.Set("encryption_info", flattenMskEncryptionInfo(cluster.EncryptionInfo)); err != nil {
		return fmt.Errorf("error setting encryption_info: %s", err)
	}

	d.Set("kafka_version", aws.StringValue(cluster.CurrentBrokerSoftwareInfo.KafkaVersion))
	d.Set("number_of_broker_nodes", aws.Int64Value(cluster.NumberOfBrokerNodes))

	if err := d.Set("tags", keyvaluetags.KafkaKeyValueTags(cluster.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	if err := d.Set("open_monitoring", flattenMskOpenMonitoring(cluster.OpenMonitoring)); err != nil {
		return fmt.Errorf("error setting open_monitoring: %s", err)
	}

	if err := d.Set("logging_info", flattenMskLoggingInfo(cluster.LoggingInfo)); err != nil {
		return fmt.Errorf("error setting logging_info: %s", err)
	}

	d.Set("zookeeper_connect_string", aws.StringValue(cluster.ZookeeperConnectString))

	return nil
}

func resourceAwsMskClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn

	if d.HasChange("broker_node_group_info.0.ebs_volume_size") {
		input := &kafka.UpdateBrokerStorageInput{
			ClusterArn:     aws.String(d.Id()),
			CurrentVersion: aws.String(d.Get("current_version").(string)),
			TargetBrokerEBSVolumeInfo: []*kafka.BrokerEBSVolumeInfo{
				{
					KafkaBrokerNodeId: aws.String("All"),
					VolumeSizeGB:      aws.Int64(int64(d.Get("broker_node_group_info.0.ebs_volume_size").(int))),
				},
			},
		}

		output, err := conn.UpdateBrokerStorage(input)

		if err != nil {
			return fmt.Errorf("error updating MSK Cluster (%s) broker storage: %s", d.Id(), err)
		}

		if output == nil {
			return fmt.Errorf("error updating MSK Cluster (%s) broker storage: empty response", d.Id())
		}

		clusterOperationARN := aws.StringValue(output.ClusterOperationArn)

		if err := waitForMskClusterOperation(conn, clusterOperationARN); err != nil {
			return fmt.Errorf("error waiting for MSK Cluster (%s) operation (%s): %s", d.Id(), clusterOperationARN, err)
		}
	}

	if d.HasChange("number_of_broker_nodes") {
		input := &kafka.UpdateBrokerCountInput{
			ClusterArn:                aws.String(d.Id()),
			CurrentVersion:            aws.String(d.Get("current_version").(string)),
			TargetNumberOfBrokerNodes: aws.Int64(int64(d.Get("number_of_broker_nodes").(int))),
		}

		output, err := conn.UpdateBrokerCount(input)

		if err != nil {
			return fmt.Errorf("error updating MSK Cluster (%s) broker count: %s", d.Id(), err)
		}

		if output == nil {
			return fmt.Errorf("error updating MSK Cluster (%s) broker count: empty response", d.Id())
		}

		clusterOperationARN := aws.StringValue(output.ClusterOperationArn)

		if err := waitForMskClusterOperation(conn, clusterOperationARN); err != nil {
			return fmt.Errorf("error waiting for MSK Cluster (%s) operation (%s): %s", d.Id(), clusterOperationARN, err)
		}
	}

	if d.HasChanges("enhanced_monitoring", "open_monitoring", "logging_info") {
		input := &kafka.UpdateMonitoringInput{
			ClusterArn:         aws.String(d.Id()),
			CurrentVersion:     aws.String(d.Get("current_version").(string)),
			EnhancedMonitoring: aws.String(d.Get("enhanced_monitoring").(string)),
			OpenMonitoring:     expandMskOpenMonitoring(d.Get("open_monitoring").([]interface{})),
			LoggingInfo:        expandMskLoggingInfo(d.Get("logging_info").([]interface{})),
		}

		output, err := conn.UpdateMonitoring(input)

		if err != nil {
			return fmt.Errorf("error updating MSK Cluster (%s) monitoring: %s", d.Id(), err)
		}

		if output == nil {
			return fmt.Errorf("error updating MSK Cluster (%s) monitoring: empty response", d.Id())
		}

		clusterOperationARN := aws.StringValue(output.ClusterOperationArn)

		if err := waitForMskClusterOperation(conn, clusterOperationARN); err != nil {
			return fmt.Errorf("error waiting for MSK Cluster (%s) operation (%s): %s", d.Id(), clusterOperationARN, err)
		}
	}

	if d.HasChange("configuration_info") {
		input := &kafka.UpdateClusterConfigurationInput{
			ClusterArn:        aws.String(d.Id()),
			ConfigurationInfo: expandMskClusterConfigurationInfo(d.Get("configuration_info").([]interface{})),
			CurrentVersion:    aws.String(d.Get("current_version").(string)),
		}

		output, err := conn.UpdateClusterConfiguration(input)

		if err != nil {
			return fmt.Errorf("error updating MSK Cluster (%s) configuration: %s", d.Id(), err)
		}

		if output == nil {
			return fmt.Errorf("error updating MSK Cluster (%s) configuration: empty response", d.Id())
		}

		clusterOperationARN := aws.StringValue(output.ClusterOperationArn)

		if err := waitForMskClusterOperation(conn, clusterOperationARN); err != nil {
			return fmt.Errorf("error waiting for MSK Cluster (%s) operation (%s): %s", d.Id(), clusterOperationARN, err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.KafkaUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating MSK Cluster (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsMskClusterRead(d, meta)

}

func expandMskClusterBrokerNodeGroupInfo(l []interface{}) *kafka.BrokerNodeGroupInfo {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	bngi := &kafka.BrokerNodeGroupInfo{
		BrokerAZDistribution: aws.String(m["az_distribution"].(string)),
		ClientSubnets:        expandStringList(m["client_subnets"].([]interface{})),
		InstanceType:         aws.String(m["instance_type"].(string)),
		SecurityGroups:       expandStringList(m["security_groups"].([]interface{})),
		StorageInfo: &kafka.StorageInfo{
			EbsStorageInfo: &kafka.EBSStorageInfo{
				VolumeSize: aws.Int64(int64(m["ebs_volume_size"].(int))),
			},
		},
	}

	return bngi
}

func expandMskClusterClientAuthentication(l []interface{}) *kafka.ClientAuthentication {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	ca := &kafka.ClientAuthentication{
		Tls: expandMskClusterTls(m["tls"].([]interface{})),
	}

	return ca
}

func expandMskClusterConfigurationInfo(l []interface{}) *kafka.ConfigurationInfo {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	ci := &kafka.ConfigurationInfo{
		Arn:      aws.String(m["arn"].(string)),
		Revision: aws.Int64(int64(m["revision"].(int))),
	}

	return ci
}

func expandMskClusterEncryptionInfo(l []interface{}) *kafka.EncryptionInfo {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	ei := &kafka.EncryptionInfo{
		EncryptionInTransit: expandMskClusterEncryptionInTransit(m["encryption_in_transit"].([]interface{})),
	}

	if v, ok := m["encryption_at_rest_kms_key_arn"]; ok && v.(string) != "" {
		ei.EncryptionAtRest = &kafka.EncryptionAtRest{
			DataVolumeKMSKeyId: aws.String(v.(string)),
		}
	}

	return ei
}

func expandMskClusterEncryptionInTransit(l []interface{}) *kafka.EncryptionInTransit {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	eit := &kafka.EncryptionInTransit{
		ClientBroker: aws.String(m["client_broker"].(string)),
		InCluster:    aws.Bool(m["in_cluster"].(bool)),
	}

	return eit
}

func expandMskClusterTls(l []interface{}) *kafka.Tls {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	tls := &kafka.Tls{
		CertificateAuthorityArnList: expandStringSet(m["certificate_authority_arns"].(*schema.Set)),
	}

	return tls
}

func expandMskOpenMonitoring(l []interface{}) *kafka.OpenMonitoringInfo {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	openMonitoring := &kafka.OpenMonitoringInfo{
		Prometheus: expandMskOpenMonitoringPrometheus(m["prometheus"].([]interface{})),
	}

	return openMonitoring
}

func expandMskOpenMonitoringPrometheus(l []interface{}) *kafka.PrometheusInfo {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	prometheus := &kafka.PrometheusInfo{
		JmxExporter:  expandMskOpenMonitoringPrometheusJmxExporter(m["jmx_exporter"].([]interface{})),
		NodeExporter: expandMskOpenMonitoringPrometheusNodeExporter(m["node_exporter"].([]interface{})),
	}

	return prometheus
}

func expandMskOpenMonitoringPrometheusJmxExporter(l []interface{}) *kafka.JmxExporterInfo {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	jmxExporter := &kafka.JmxExporterInfo{
		EnabledInBroker: aws.Bool(m["enabled_in_broker"].(bool)),
	}

	return jmxExporter
}

func expandMskOpenMonitoringPrometheusNodeExporter(l []interface{}) *kafka.NodeExporterInfo {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	nodeExporter := &kafka.NodeExporterInfo{
		EnabledInBroker: aws.Bool(m["enabled_in_broker"].(bool)),
	}

	return nodeExporter
}

func expandMskLoggingInfo(l []interface{}) *kafka.LoggingInfo {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	loggingInfo := &kafka.LoggingInfo{
		BrokerLogs: expandMskLoggingInfoBrokerLogs(m["broker_logs"].([]interface{})),
	}

	return loggingInfo
}

func expandMskLoggingInfoBrokerLogs(l []interface{}) *kafka.BrokerLogs {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	brokerLogs := &kafka.BrokerLogs{
		CloudWatchLogs: expandMskLoggingInfoBrokerLogsCloudWatchLogs(m["cloudwatch_logs"].([]interface{})),
		Firehose:       expandMskLoggingInfoBrokerLogsFirehose(m["firehose"].([]interface{})),
		S3:             expandMskLoggingInfoBrokerLogsS3(m["s3"].([]interface{})),
	}

	return brokerLogs
}

func expandMskLoggingInfoBrokerLogsCloudWatchLogs(l []interface{}) *kafka.CloudWatchLogs {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	cloudWatchLogs := &kafka.CloudWatchLogs{
		Enabled:  aws.Bool(m["enabled"].(bool)),
		LogGroup: aws.String(m["log_group"].(string)),
	}

	return cloudWatchLogs
}

func expandMskLoggingInfoBrokerLogsFirehose(l []interface{}) *kafka.Firehose {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	firehose := &kafka.Firehose{
		Enabled:        aws.Bool(m["enabled"].(bool)),
		DeliveryStream: aws.String(m["delivery_stream"].(string)),
	}

	return firehose
}

func expandMskLoggingInfoBrokerLogsS3(l []interface{}) *kafka.S3 {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	s3 := &kafka.S3{
		Enabled: aws.Bool(m["enabled"].(bool)),
		Bucket:  aws.String(m["bucket"].(string)),
		Prefix:  aws.String(m["prefix"].(string)),
	}

	return s3
}

func flattenMskBrokerNodeGroupInfo(b *kafka.BrokerNodeGroupInfo) []map[string]interface{} {

	if b == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"az_distribution": aws.StringValue(b.BrokerAZDistribution),
		"client_subnets":  flattenStringList(b.ClientSubnets),
		"instance_type":   aws.StringValue(b.InstanceType),
		"security_groups": flattenStringList(b.SecurityGroups),
	}
	if b.StorageInfo != nil {
		if b.StorageInfo.EbsStorageInfo != nil {
			m["ebs_volume_size"] = int(aws.Int64Value(b.StorageInfo.EbsStorageInfo.VolumeSize))
		}
	}
	return []map[string]interface{}{m}
}

func flattenMskClientAuthentication(ca *kafka.ClientAuthentication) []map[string]interface{} {
	if ca == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"tls": flattenMskTls(ca.Tls),
	}

	return []map[string]interface{}{m}
}

func flattenMskConfigurationInfo(bsi *kafka.BrokerSoftwareInfo) []map[string]interface{} {
	if bsi == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"arn":      aws.StringValue(bsi.ConfigurationArn),
		"revision": aws.Int64Value(bsi.ConfigurationRevision),
	}

	return []map[string]interface{}{m}
}

func flattenMskEncryptionInfo(e *kafka.EncryptionInfo) []map[string]interface{} {
	if e == nil || e.EncryptionAtRest == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"encryption_at_rest_kms_key_arn": aws.StringValue(e.EncryptionAtRest.DataVolumeKMSKeyId),
		"encryption_in_transit":          flattenMskEncryptionInTransit(e.EncryptionInTransit),
	}

	return []map[string]interface{}{m}
}

func flattenMskEncryptionInTransit(eit *kafka.EncryptionInTransit) []map[string]interface{} {
	if eit == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"client_broker": aws.StringValue(eit.ClientBroker),
		"in_cluster":    aws.BoolValue(eit.InCluster),
	}

	return []map[string]interface{}{m}
}

func flattenMskTls(tls *kafka.Tls) []map[string]interface{} {
	if tls == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"certificate_authority_arns": aws.StringValueSlice(tls.CertificateAuthorityArnList),
	}

	return []map[string]interface{}{m}
}

func flattenMskOpenMonitoring(e *kafka.OpenMonitoring) []map[string]interface{} {
	if e == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"prometheus": flattenMskOpenMonitoringPrometheus(e.Prometheus),
	}

	return []map[string]interface{}{m}
}

func flattenMskOpenMonitoringPrometheus(e *kafka.Prometheus) []map[string]interface{} {
	if e == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"jmx_exporter":  flattenMskOpenMonitoringPrometheusJmxExporter(e.JmxExporter),
		"node_exporter": flattenMskOpenMonitoringPrometheusNodeExporter(e.NodeExporter),
	}

	return []map[string]interface{}{m}
}

func flattenMskOpenMonitoringPrometheusJmxExporter(e *kafka.JmxExporter) []map[string]interface{} {
	if e == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"enabled_in_broker": aws.BoolValue(e.EnabledInBroker),
	}

	return []map[string]interface{}{m}
}

func flattenMskOpenMonitoringPrometheusNodeExporter(e *kafka.NodeExporter) []map[string]interface{} {
	if e == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"enabled_in_broker": aws.BoolValue(e.EnabledInBroker),
	}

	return []map[string]interface{}{m}
}

func flattenMskLoggingInfo(e *kafka.LoggingInfo) []map[string]interface{} {
	if e == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"broker_logs": flattenMskLoggingInfoBrokerLogs(e.BrokerLogs),
	}

	return []map[string]interface{}{m}
}

func flattenMskLoggingInfoBrokerLogs(e *kafka.BrokerLogs) []map[string]interface{} {
	if e == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"cloudwatch_logs": flattenMskLoggingInfoBrokerLogsCloudWatchLogs(e.CloudWatchLogs),
		"firehose":        flattenMskLoggingInfoBrokerLogsFirehose(e.Firehose),
		"s3":              flattenMskLoggingInfoBrokerLogsS3(e.S3),
	}

	return []map[string]interface{}{m}
}

func flattenMskLoggingInfoBrokerLogsCloudWatchLogs(e *kafka.CloudWatchLogs) []map[string]interface{} {
	if e == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"enabled":   aws.BoolValue(e.Enabled),
		"log_group": aws.StringValue(e.LogGroup),
	}

	return []map[string]interface{}{m}
}

func flattenMskLoggingInfoBrokerLogsFirehose(e *kafka.Firehose) []map[string]interface{} {
	if e == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"enabled":         aws.BoolValue(e.Enabled),
		"delivery_stream": aws.StringValue(e.DeliveryStream),
	}

	return []map[string]interface{}{m}
}

func flattenMskLoggingInfoBrokerLogsS3(e *kafka.S3) []map[string]interface{} {
	if e == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"enabled": aws.BoolValue(e.Enabled),
		"bucket":  aws.StringValue(e.Bucket),
		"prefix":  aws.StringValue(e.Prefix),
	}

	return []map[string]interface{}{m}
}

func resourceAwsMskClusterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn

	log.Printf("[DEBUG] Deleting MSK cluster: %q", d.Id())
	_, err := conn.DeleteCluster(&kafka.DeleteClusterInput{
		ClusterArn: aws.String(d.Id()),
	})
	if err != nil {
		if isAWSErr(err, kafka.ErrCodeNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("failed deleting MSK cluster %q: %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Waiting for MSK cluster %q to be deleted", d.Id())

	return resourceAwsMskClusterDeleteWaiter(conn, d.Id())
}

func resourceAwsMskClusterDeleteWaiter(conn *kafka.Kafka, arn string) error {
	input := &kafka.DescribeClusterInput{
		ClusterArn: aws.String(arn),
	}
	err := resource.Retry(60*time.Minute, func() *resource.RetryError {
		_, err := conn.DescribeCluster(input)

		if err != nil {
			if isAWSErr(err, kafka.ErrCodeNotFoundException, "") {
				return nil
			}
			return resource.NonRetryableError(err)
		}

		return resource.RetryableError(fmt.Errorf("timeout while waiting for the cluster %q to be deleted", arn))
	})
	if isResourceTimeoutError(err) {
		_, err = conn.DescribeCluster(input)
		if isAWSErr(err, kafka.ErrCodeNotFoundException, "") {
			return nil
		}
	}
	if err != nil {
		return fmt.Errorf("Error waiting for MSK cluster to be deleted: %s", err)
	}
	return nil
}

func mskClusterOperationRefreshFunc(conn *kafka.Kafka, clusterOperationARN string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &kafka.DescribeClusterOperationInput{
			ClusterOperationArn: aws.String(clusterOperationARN),
		}

		output, err := conn.DescribeClusterOperation(input)

		if err != nil {
			return nil, "UPDATE_FAILED", fmt.Errorf("error describing MSK Cluster Operation (%s): %s", clusterOperationARN, err)
		}

		if output == nil || output.ClusterOperationInfo == nil {
			return nil, "UPDATE_FAILED", fmt.Errorf("error describing MSK Cluster Operation (%s): empty response", clusterOperationARN)
		}

		state := aws.StringValue(output.ClusterOperationInfo.OperationState)

		if state == "UPDATE_FAILED" && output.ClusterOperationInfo.ErrorInfo != nil {
			errorInfo := output.ClusterOperationInfo.ErrorInfo
			err := fmt.Errorf("error code: %s, error string: %s", aws.StringValue(errorInfo.ErrorCode), aws.StringValue(errorInfo.ErrorString))
			return output.ClusterOperationInfo, state, err
		}

		return output.ClusterOperationInfo, state, nil
	}
}

func waitForMskClusterOperation(conn *kafka.Kafka, clusterOperationARN string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"PENDING", "UPDATE_IN_PROGRESS"},
		Target:  []string{"UPDATE_COMPLETE"},
		Refresh: mskClusterOperationRefreshFunc(conn, clusterOperationARN),
		Timeout: 60 * time.Minute,
	}

	log.Printf("[DEBUG] Waiting for MSK Cluster Operation (%s) completion", clusterOperationARN)
	_, err := stateConf.WaitForState()

	return err
}
