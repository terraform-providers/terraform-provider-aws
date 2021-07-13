package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSMskNodes_ClientIp(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_msk_node.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccMskNodeDataSourceClientVPC(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "attached_eni_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "broker_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "client_subnet"),
					resource.TestCheckResourceAttrSet(dataSourceName, "client_vpc_ip_address"),
					resource.TestCheckResourceAttrSet(dataSourceName, "kafka_version"),
					resource.TestCheckResourceAttrSet(dataSourceName, "broker_endpoint"),
					resource.TestCheckResourceAttrSet(dataSourceName, "instance_type"),
					resource.TestCheckResourceAttrSet(dataSourceName, "arn"),
				),
			},
		},
	})
}

func testAccMskNodeDataSourceClientVPC(rName string) string {
	return testAccMskClusterBaseConfig() + fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.2.1"
  number_of_broker_nodes = 3
  broker_node_group_info {
    client_subnets  = [aws_subnet.example_subnet_az1.id, aws_subnet.example_subnet_az2.id, aws_subnet.example_subnet_az3.id]
    ebs_volume_size = 10
    instance_type   = "kafka.t3.small"
    security_groups = [aws_security_group.example_sg.id]
  }
}

data "aws_msk_node" "test" {
  cluster_arn     = aws_msk_cluster.test.arn
  broker_endpoint = trimsuffix(split(",", aws_msk_cluster.test.bootstrap_brokers_tls)[0], ":9094")
}
`, rName)

}
