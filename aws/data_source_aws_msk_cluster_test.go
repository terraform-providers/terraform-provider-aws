package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSMskClusterDataSource_Name(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_msk_cluster.test"
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMsk(t) },
		ErrorCheck:   testAccErrorCheck(t, kafka.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskClusterDataSourceConfigName(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bootstrap_brokers", resourceName, "bootstrap_brokers"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bootstrap_brokers_sasl_scram", resourceName, "bootstrap_brokers_sasl_scram"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bootstrap_brokers_tls", resourceName, "bootstrap_brokers_tls"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_name", resourceName, "cluster_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "kafka_version", resourceName, "kafka_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "number_of_broker_nodes", resourceName, "number_of_broker_nodes"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "zookeeper_connect_string", resourceName, "zookeeper_connect_string"),
				),
			},
		},
	})
}

func testAccMskClusterDataSourceConfigName(rName string) string {
	return testAccMskClusterBaseConfig() + fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.2.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = [aws_subnet.example_subnet_az1.id, aws_subnet.example_subnet_az2.id, aws_subnet.example_subnet_az3.id]
    ebs_volume_size = 10
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.example_sg.id]
  }

  tags = {
    foo = "bar"
  }
}

data "aws_msk_cluster" "test" {
  cluster_name = aws_msk_cluster.test.cluster_name
}
`, rName)
}
