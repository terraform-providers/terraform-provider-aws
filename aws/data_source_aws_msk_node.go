package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsMskNode() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsMskNodeRead,

		Schema: map[string]*schema.Schema{
			"attached_eni_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"broker_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"client_subnet": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_vpc_ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kafka_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"broker_endpoint": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"instance_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceAwsMskNodeRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn

	listNodesInput := &kafka.ListNodesInput{
		ClusterArn: aws.String(d.Get("cluster_arn").(string)),
	}

	var nodes []*kafka.NodeInfo
	for {
		listNodesOutput, err := conn.ListNodes(listNodesInput)

		if err != nil {
			return fmt.Errorf("error listing MSK Cluster Nodes: %s", err)
		}

		if listNodesOutput == nil {
			break
		}

		nodes = append(nodes, listNodesOutput.NodeInfoList...)

		if aws.StringValue(listNodesOutput.NextToken) == "" {
			break
		}

		listNodesInput.NextToken = listNodesOutput.NextToken
	}

	if len(nodes) == 0 {
		return fmt.Errorf("error reading MSK Nodes: no results found")
	}

	var brokerNodes []*kafka.NodeInfo

	for _, broker := range nodes {
		if int(*broker.BrokerNodeInfo.BrokerId) == d.Get("broker_id").(int) {
			brokerNodes = append(brokerNodes, broker)
		} else if *broker.BrokerNodeInfo.Endpoints[0] == d.Get("broker_endpoint").(string) {
			brokerNodes = append(brokerNodes, broker)
		}
	}

	if len(brokerNodes) < 1 {
		return fmt.Errorf("error reading MSK Nodes: node not found, try adjusting search criteria")
	}

	node := brokerNodes[0]

	d.Set("attached_eni_id", aws.StringValue(node.BrokerNodeInfo.AttachedENIId))
	d.Set("broker_id", int(aws.Float64Value(node.BrokerNodeInfo.BrokerId)))
	d.Set("client_subnet", aws.StringValue(node.BrokerNodeInfo.ClientSubnet))
	d.Set("client_vpc_ip_address", aws.StringValue(node.BrokerNodeInfo.ClientVpcIpAddress))
	d.Set("kafka_version", aws.StringValue(node.BrokerNodeInfo.CurrentBrokerSoftwareInfo.KafkaVersion))
	d.Set("broker_endpoint", aws.StringValue(node.BrokerNodeInfo.Endpoints[0]))
	d.Set("instance_type", aws.StringValue(node.InstanceType))
	d.Set("arn", aws.StringValue(node.NodeARN))
	d.SetId(aws.StringValue(node.NodeARN))

	return nil
}
