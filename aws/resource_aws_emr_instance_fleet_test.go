package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSEMRInstanceFleet_basic(t *testing.T) {
	var fleet emr.InstanceFleet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emr_instance_fleet.task"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrInstanceFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrInstanceFleetConfig(rName),
				Check: resource.ComposeTestCheckFunc(testAccCheckAWSEmrInstanceFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "instance_type_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_on_demand_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_spot_capacity", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSEMRInstanceFleetResourceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEMRInstanceFleet_zero_count(t *testing.T) {
	var fleet emr.InstanceFleet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emr_instance_fleet.task"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrInstanceFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrInstanceFleetConfig(rName),
				Check: resource.ComposeTestCheckFunc(testAccCheckAWSEmrInstanceFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "instance_type_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_on_demand_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_spot_capacity", "0"),
				),
			},
			{
				Config: testAccAWSEmrInstanceFleetConfigZeroCount(rName),
				Check: resource.ComposeTestCheckFunc(testAccCheckAWSEmrInstanceFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "instance_type_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_on_demand_capacity", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_spot_capacity", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSEMRInstanceFleetResourceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEMRInstanceFleet_ebsBasic(t *testing.T) {
	var fleet emr.InstanceFleet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emr_instance_fleet.task"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrInstanceFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrInstanceFleetConfigEbsBasic(rName),
				Check: resource.ComposeTestCheckFunc(testAccCheckAWSEmrInstanceFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "instance_type_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_on_demand_capacity", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_spot_capacity", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSEMRInstanceFleetResourceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEMRInstanceFleet_full(t *testing.T) {
	var fleet emr.InstanceFleet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emr_instance_fleet.task"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrInstanceFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrInstanceFleetConfigFull(rName),
				Check: resource.ComposeTestCheckFunc(testAccCheckAWSEmrInstanceFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "instance_type_configs.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "target_on_demand_capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "target_spot_capacity", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSEMRInstanceFleetResourceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEMRInstanceFleet_disappears(t *testing.T) {
	var fleet emr.InstanceFleet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emr_instance_fleet.task"
	emrClusterResourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrInstanceFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrInstanceFleetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrInstanceFleetExists(resourceName, &fleet),
					// EMR Instance Fleet can only be scaled down and are not removed until the
					// Cluster is removed. Verify EMR Cluster disappearance handling.
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEMRCluster(), emrClusterResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSEmrInstanceFleetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).emrconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_emr_cluster" {
			continue
		}

		params := &emr.DescribeClusterInput{
			ClusterId: aws.String(rs.Primary.ID),
		}

		describe, err := conn.DescribeCluster(params)

		if err == nil {
			if describe.Cluster != nil &&
				*describe.Cluster.Status.State == "WAITING" {
				return fmt.Errorf("EMR Cluster still exists")
			}
		}

		providerErr, ok := err.(awserr.Error)
		if !ok {
			return err
		}

		log.Printf("[ERROR] %v", providerErr)
	}

	return nil
}

func testAccCheckAWSEmrInstanceFleetExists(n string, v *emr.InstanceFleet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No task fleet id set")
		}
		meta := testAccProvider.Meta()
		conn := meta.(*AWSClient).emrconn
		instanceFleets, err := fetchAllEMRInstanceFleets(conn, rs.Primary.Attributes["cluster_id"])
		if err != nil {
			return fmt.Errorf("EMR error: %v", err)
		}

		fleet := findInstanceFleetById(instanceFleets, rs.Primary.ID)
		if fleet == nil {
			return fmt.Errorf("No match found for (%s)", n)
		}
		v = fleet
		return nil
	}
}

func testAccAWSEMRInstanceFleetResourceImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["cluster_id"], rs.Primary.ID), nil
	}
}

const testAccAWSEmrInstanceFleetBase = `
data "aws_availability_zones" "available" {
  # Many instance types are not available in this availability zone
  exclude_zone_ids = ["usw2-az4"]
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "tf-acc-test-emr-cluster"
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-emr-cluster"
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id

  ingress {
    from_port = 0
    protocol  = "-1"
    self      = true
    to_port   = 0
  }

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }

  tags = {
    Name = "tf-acc-test-emr-cluster"
  }

  # EMR will modify ingress rules
  lifecycle {
    ignore_changes = [ingress]
  }
}

resource "aws_subnet" "test" {
  availability_zone       = data.aws_availability_zones.available.names[0]
  cidr_block              = "10.0.0.0/24"
  map_public_ip_on_launch = false
  vpc_id                  = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-emr-cluster"
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }
}

resource "aws_route_table_association" "test" {
  route_table_id = aws_route_table.test.id
  subnet_id      = aws_subnet.test.id
}

data "aws_partition" "current" {}

resource "aws_iam_role" "emr_service" {
  name               = "%[1]s_default_role"
  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "elasticmapreduce.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy_attachment" "emr_service" {
  role       = aws_iam_role.emr_service.id
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonElasticMapReduceRole"
}

resource "aws_iam_instance_profile" "emr_instance_profile" {
  name = "%[1]s_profile"
  role = aws_iam_role.emr_instance_profile.name
}

resource "aws_iam_role" "emr_instance_profile" {
  name               = "%[1]s_profile_role"
  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy_attachment" "emr_instance_profile" {
  role       = aws_iam_role.emr_instance_profile.id
  policy_arn = aws_iam_policy.emr_instance_profile.arn
}

resource "aws_iam_policy" "emr_instance_profile" {
  name   = "%[1]s_profile"
  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "cloudwatch:*",
            "dynamodb:*",
            "ec2:Describe*",
            "elasticmapreduce:Describe*",
            "elasticmapreduce:ListBootstrapActions",
            "elasticmapreduce:ListClusters",
            "elasticmapreduce:ListInstanceGroups",
            "elasticmapreduce:ListInstances",
            "elasticmapreduce:ListSteps",
            "kinesis:CreateStream",
            "kinesis:DeleteStream",
            "kinesis:DescribeStream",
            "kinesis:GetRecords",
            "kinesis:GetShardIterator",
            "kinesis:MergeShards",
            "kinesis:PutRecord",
            "kinesis:SplitShard",
            "rds:Describe*",
            "s3:*",
            "sdb:*",
            "sns:*",
            "sqs:*"
        ]
    }]
}
EOT
}

resource "aws_emr_cluster" "test" {
  name          = "%[1]s"
  release_label = "emr-5.30.1"
  applications  = ["Hadoop", "Hive"]
  log_uri       = "s3n://terraform/testlog/"

  master_instance_fleet {
    instance_type_configs {
      instance_type = "m3.xlarge"
    }

    target_on_demand_capacity = 1
  }

  core_instance_fleet {
    instance_type_configs {
      bid_price_as_percentage_of_on_demand_price = 100

      ebs_config {
        size                 = 100
        type                 = "gp2"
        volumes_per_instance = 1
      }

      instance_type     = "m4.xlarge"
      weighted_capacity = 1
    }
    name                      = "core fleet"
    target_on_demand_capacity = 1
    target_spot_capacity      = 0
  }

  service_role = aws_iam_role.emr_service.arn
  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
  ]

  ec2_attributes {
    subnet_id                         = aws_subnet.test.id
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
  }
}
`

func testAccAWSEmrInstanceFleetConfig(r string) string {
	return fmt.Sprintf(testAccAWSEmrInstanceFleetBase+`
resource "aws_emr_instance_fleet" "task" {
  cluster_id = aws_emr_cluster.test.id

  instance_type_configs {
    instance_type     = "m3.xlarge"
    weighted_capacity = 1
  }

  launch_specifications {
    on_demand_specification {
      allocation_strategy = "lowest-price"
    }
  }

  name                      = "emr_instance_fleet_%[1]s"
  target_on_demand_capacity = 1
  target_spot_capacity      = 0
}
`, r)
}

func testAccAWSEmrInstanceFleetConfigZeroCount(r string) string {
	return fmt.Sprintf(testAccAWSEmrInstanceFleetBase+`
resource "aws_emr_instance_fleet" "task" {
  cluster_id = aws_emr_cluster.test.id

  instance_type_configs {
    instance_type     = "m3.xlarge"
    weighted_capacity = 1
  }

  launch_specifications {
    on_demand_specification {
      allocation_strategy = "lowest-price"
    }
  }

  name                      = "emr_instance_fleet_%[1]s"
  target_on_demand_capacity = 0
  target_spot_capacity      = 0
}
`, r)
}

func testAccAWSEmrInstanceFleetConfigEbsBasic(r string) string {
	return fmt.Sprintf(testAccAWSEmrInstanceFleetBase+`
resource "aws_emr_instance_fleet" "task" {
  cluster_id = aws_emr_cluster.test.id

  instance_type_configs {
    bid_price_as_percentage_of_on_demand_price = 100
    ebs_config {
      size                 = 10
      type                 = "gp2"
      volumes_per_instance = 1
    }
    instance_type     = "m4.xlarge"
    weighted_capacity = 1
  }

  launch_specifications {
    spot_specification {
      allocation_strategy      = "capacity-optimized"
      block_duration_minutes   = 0
      timeout_action           = "SWITCH_TO_ON_DEMAND"
      timeout_duration_minutes = 10
    }
  }

  name                      = "emr_instance_fleet_%[1]s"
  target_on_demand_capacity = 0
  target_spot_capacity      = 1
}
`, r)
}

func testAccAWSEmrInstanceFleetConfigFull(r string) string {
	return fmt.Sprintf(testAccAWSEmrInstanceFleetBase+`
resource "aws_emr_instance_fleet" "task" {
  cluster_id = aws_emr_cluster.test.id

  instance_type_configs {
    bid_price_as_percentage_of_on_demand_price = 100
    ebs_config {
      size                 = 10
      type                 = "gp2"
      volumes_per_instance = 1
    }

    ebs_config {
      size                 = 20
      type                 = "gp2"
      volumes_per_instance = 2
    }

    instance_type     = "m4.xlarge"
    weighted_capacity = 1
  }

  instance_type_configs {
    bid_price_as_percentage_of_on_demand_price = 80

    ebs_config {
      size                 = 10
      type                 = "gp2"
      volumes_per_instance = 1
    }

    instance_type     = "m4.2xlarge"
    weighted_capacity = 2
  }

  launch_specifications {
    spot_specification {
      allocation_strategy      = "capacity-optimized"
      block_duration_minutes   = 0
      timeout_action           = "SWITCH_TO_ON_DEMAND"
      timeout_duration_minutes = 10
    }
  }

  name                      = "emr_instance_fleet_%[1]s"
  target_on_demand_capacity = 2
  target_spot_capacity      = 2
}
`, r)
}
