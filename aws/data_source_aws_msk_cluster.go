package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsMskCluster() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsMskClusterRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers_sasl_iam": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers_sasl_scram": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers_tls": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"kafka_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"number_of_broker_nodes": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"tags": tagsSchemaComputed(),
			"zookeeper_connect_string": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsMskClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	listClustersInput := &kafka.ListClustersInput{
		ClusterNameFilter: aws.String(d.Get("cluster_name").(string)),
	}

	var clusters []*kafka.ClusterInfo
	for {
		listClustersOutput, err := conn.ListClusters(listClustersInput)

		if err != nil {
			return fmt.Errorf("error listing MSK Clusters: %w", err)
		}

		if listClustersOutput == nil {
			break
		}

		clusters = append(clusters, listClustersOutput.ClusterInfoList...)

		if aws.StringValue(listClustersOutput.NextToken) == "" {
			break
		}

		listClustersInput.NextToken = listClustersOutput.NextToken
	}

	if len(clusters) == 0 {
		return fmt.Errorf("error reading MSK Cluster: no results found")
	}

	if len(clusters) > 1 {
		return fmt.Errorf("error reading MSK Cluster: multiple results found, try adjusting search criteria")
	}

	cluster := clusters[0]

	bootstrapBrokersInput := &kafka.GetBootstrapBrokersInput{
		ClusterArn: cluster.ClusterArn,
	}

	bootstrapBrokersOutput, err := conn.GetBootstrapBrokers(bootstrapBrokersInput)

	if err != nil {
		return fmt.Errorf("error reading MSK Cluster (%s) bootstrap brokers: %w", aws.StringValue(cluster.ClusterArn), err)
	}

	d.Set("arn", cluster.ClusterArn)
	d.Set("bootstrap_brokers", sortMskClusterEndpoints(aws.StringValue(bootstrapBrokersOutput.BootstrapBrokerString)))
	d.Set("bootstrap_brokers_sasl_iam", sortMskClusterEndpoints(aws.StringValue(bootstrapBrokersOutput.BootstrapBrokerStringSaslIam)))
	d.Set("bootstrap_brokers_sasl_scram", sortMskClusterEndpoints(aws.StringValue(bootstrapBrokersOutput.BootstrapBrokerStringSaslScram)))
	d.Set("bootstrap_brokers_tls", sortMskClusterEndpoints(aws.StringValue(bootstrapBrokersOutput.BootstrapBrokerStringTls)))
	d.Set("cluster_name", cluster.ClusterName)
	d.Set("kafka_version", cluster.CurrentBrokerSoftwareInfo.KafkaVersion)
	d.Set("number_of_broker_nodes", cluster.NumberOfBrokerNodes)

	if err := d.Set("tags", keyvaluetags.KafkaKeyValueTags(cluster.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	d.Set("zookeeper_connect_string", sortMskClusterEndpoints(aws.StringValue(cluster.ZookeeperConnectString)))

	d.SetId(aws.StringValue(cluster.ClusterArn))

	return nil
}
