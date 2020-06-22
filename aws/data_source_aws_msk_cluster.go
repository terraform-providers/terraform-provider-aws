package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
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
			return fmt.Errorf("error listing MSK Clusters: %s", err)
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

	bootstrapBrokersoOutput, err := conn.GetBootstrapBrokers(bootstrapBrokersInput)

	if err != nil {
		return fmt.Errorf("error reading MSK Cluster (%s) bootstrap brokers: %s", aws.StringValue(cluster.ClusterArn), err)
	}

	d.Set("arn", aws.StringValue(cluster.ClusterArn))
	d.Set("bootstrap_brokers", aws.StringValue(bootstrapBrokersoOutput.BootstrapBrokerString))
	d.Set("bootstrap_brokers_tls", aws.StringValue(bootstrapBrokersoOutput.BootstrapBrokerStringTls))
	d.Set("cluster_name", aws.StringValue(cluster.ClusterName))
	d.Set("kafka_version", aws.StringValue(cluster.CurrentBrokerSoftwareInfo.KafkaVersion))
	d.Set("number_of_broker_nodes", aws.Int64Value(cluster.NumberOfBrokerNodes))

	if err := d.Set("tags", keyvaluetags.KafkaKeyValueTags(cluster.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("zookeeper_connect_string", aws.StringValue(cluster.ZookeeperConnectString))

	d.SetId(aws.StringValue(cluster.ClusterArn))

	return nil
}
