package aws

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSDMSReplicationInstanceDataSource_basic(t *testing.T) {
	resourceName := "aws_dms_replication_instance.test"
	datasourceName := "data.aws_dms_replication_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDmsReplicationInstanceDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "replication_instance_arn", resourceName, "replication_instance_arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "replication_instance_private_ips", resourceName, "replication_instance_private_ips"),
					resource.TestCheckResourceAttrPair(datasourceName, "replication_instance_public_ips", resourceName, "replication_instance_public_ips"),
				),
			},
		},
	})
}

var testAccDmsReplicationInstanceDataSourceConfig = `
resource "aws_dms_replication_instance" "test" {
  allocated_storage            = 50
  apply_immediately            = true
  auto_minor_version_upgrade   = true
  availability_zone            = "us-east-1c"
  engine_version               = "3.4.0"
  multi_az                     = false
  preferred_maintenance_window = "sun:10:30-sun:14:30"
  replication_instance_class   = "dms.t2.medium"
  replication_instance_id      = "instance-id"
  replication_subnet_group_id  = "subnet-group-id"
}

data "aws_dms_replication_instance" "test" {
  filter {
    name   = "replication-instance-id"
    values = [aws_dms_replication_instance.test.id]
  }
}
`
