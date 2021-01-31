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

func TestAccAWSEMRInstanceGroup_basic(t *testing.T) {
	var ig emr.InstanceGroup
	rInt := acctest.RandInt()

	resourceName := "aws_emr_instance_group.task"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrInstanceGroupConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrInstanceGroupExists(resourceName, &ig),
					resource.TestCheckResourceAttr(resourceName, "autoscaling_policy", ""),
					resource.TestCheckResourceAttr(resourceName, "bid_price", ""),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSEMRInstanceGroupResourceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEMRInstanceGroup_BidPrice(t *testing.T) {
	var ig1, ig2 emr.InstanceGroup
	rInt := acctest.RandInt()

	resourceName := "aws_emr_instance_group.task"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrInstanceGroupConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrInstanceGroupExists(resourceName, &ig1),
					resource.TestCheckResourceAttr(resourceName, "bid_price", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSEMRInstanceGroupResourceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEmrInstanceGroupConfig_BidPrice(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrInstanceGroupExists(resourceName, &ig2),
					resource.TestCheckResourceAttr(resourceName, "bid_price", "0.30"),
					testAccAWSEMRInstanceGroupRecreated(t, &ig1, &ig2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSEMRInstanceGroupResourceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEmrInstanceGroupConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrInstanceGroupExists(resourceName, &ig1),
					resource.TestCheckResourceAttr(resourceName, "bid_price", ""),
					testAccAWSEMRInstanceGroupRecreated(t, &ig1, &ig2),
				),
			},
		},
	})
}

func TestAccAWSEMRInstanceGroup_ConfigurationsJson(t *testing.T) {
	var ig emr.InstanceGroup
	rInt := acctest.RandInt()

	resourceName := "aws_emr_instance_group.task"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrInstanceGroupConfig_ConfigurationsJson(rInt, "partitionName1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrInstanceGroupExists(resourceName, &ig),
					resource.TestCheckResourceAttrSet(resourceName, "configurations_json"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSEMRInstanceGroupResourceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEmrInstanceGroupConfig_ConfigurationsJson(rInt, "partitionName2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrInstanceGroupExists(resourceName, &ig),
					resource.TestCheckResourceAttrSet(resourceName, "configurations_json"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSEMRInstanceGroupResourceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEMRInstanceGroup_AutoScalingPolicy(t *testing.T) {
	var ig emr.InstanceGroup
	rInt := acctest.RandInt()

	resourceName := "aws_emr_instance_group.task"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrInstanceGroupConfig_AutoScalingPolicy(rInt, 1, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrInstanceGroupExists(resourceName, &ig),
					resource.TestCheckResourceAttrSet(resourceName, "autoscaling_policy"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSEMRInstanceGroupResourceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEmrInstanceGroupConfig_AutoScalingPolicy(rInt, 2, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrInstanceGroupExists(resourceName, &ig),
					resource.TestCheckResourceAttrSet(resourceName, "autoscaling_policy"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSEMRInstanceGroupResourceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

// Confirm we can scale down the instance count.
// Regression test for https://github.com/hashicorp/terraform-provider-aws/issues/1264
func TestAccAWSEMRInstanceGroup_InstanceCount(t *testing.T) {
	var ig emr.InstanceGroup
	rInt := acctest.RandInt()

	resourceName := "aws_emr_instance_group.task"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrInstanceGroupConfig_basic(rInt),
				Check:  testAccCheckAWSEmrInstanceGroupExists(resourceName, &ig),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSEMRInstanceGroupResourceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEmrInstanceGroupConfig_zeroCount(rInt),
				Check:  testAccCheckAWSEmrInstanceGroupExists(resourceName, &ig),
			},
		},
	})
}

// Regression test for https://github.com/hashicorp/terraform-provider-aws/issues/1355
func TestAccAWSEMRInstanceGroup_disappears_EmrCluster(t *testing.T) {
	var cluster emr.Cluster
	var ig emr.InstanceGroup
	rInt := acctest.RandInt()
	emrClusterResourceName := "aws_emr_cluster.tf-test-cluster"
	resourceName := "aws_emr_instance_group.task"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrInstanceGroupConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(emrClusterResourceName, &cluster),
					testAccCheckAWSEmrInstanceGroupExists(resourceName, &ig),
					testAccCheckAWSEmrClusterDisappears(&cluster),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSEMRInstanceGroup_EbsConfig_EbsOptimized(t *testing.T) {
	var ig emr.InstanceGroup
	rInt := acctest.RandInt()

	resourceName := "aws_emr_instance_group.task"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrInstanceGroupConfig_ebsConfig(rInt, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrInstanceGroupExists(resourceName, &ig),
					resource.TestCheckResourceAttr(resourceName, "ebs_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSEMRInstanceGroupResourceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEmrInstanceGroupConfig_ebsConfig(rInt, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrInstanceGroupExists(resourceName, &ig),
					resource.TestCheckResourceAttr(resourceName, "ebs_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", "false"),
				),
			},
		},
	})
}

func testAccCheckAWSEmrInstanceGroupDestroy(s *terraform.State) error {
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

		if providerErr, ok := err.(awserr.Error); !ok {
			log.Printf("[ERROR] %v", providerErr)
			return err
		}
	}

	return nil
}

func testAccCheckAWSEmrInstanceGroupExists(name string, ig *emr.InstanceGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No task group id set")
		}

		meta := testAccProvider.Meta()
		conn := meta.(*AWSClient).emrconn
		group, err := fetchEMRInstanceGroup(conn, rs.Primary.Attributes["cluster_id"], rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("EMR error: %v", err)
		}

		if group == nil {
			return fmt.Errorf("No match found for (%s)", name)
		}
		*ig = *group

		return nil
	}
}

func testAccAWSEMRInstanceGroupResourceImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["cluster_id"], rs.Primary.ID), nil
	}
}

func testAccAWSEMRInstanceGroupRecreated(t *testing.T, before, after *emr.InstanceGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(before.Id) == aws.StringValue(after.Id) {
			t.Fatalf("Expected change of Instance Group Ids, but both were %v", aws.StringValue(before.Id))
		}

		return nil
	}
}

const testAccAWSEmrInstanceGroupBase = `
data "aws_availability_zones" "available" {
  # Many instance types are not available in this availability zone
  exclude_zone_ids = ["usw2-az4"]
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_partition" "current" {}

resource "aws_security_group" "allow_all" {
  name        = "allow_all"
  description = "Allow all inbound traffic"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  depends_on = [aws_subnet.main]

  lifecycle {
    ignore_changes = ["ingress", "egress"]
  }
}

resource "aws_vpc" "main" {
  cidr_block           = "168.31.0.0/16"
  enable_dns_hostnames = true
}

resource "aws_subnet" "main" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "168.31.0.0/20"
  vpc_id            = aws_vpc.main.id
}

resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.main.id
}

resource "aws_route_table" "r" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.gw.id
  }
}

resource "aws_main_route_table_association" "a" {
  vpc_id         = aws_vpc.main.id
  route_table_id = aws_route_table.r.id
}

## EMR Cluster Configuration
resource "aws_emr_cluster" "tf-test-cluster" {
  applications                      = ["Spark"]
  autoscaling_role                  = aws_iam_role.emr-autoscaling-role.arn
  configurations                    = "test-fixtures/emr_configurations.json"
  keep_job_flow_alive_when_no_steps = true
  name                              = "tf-test-emr-%[1]d"
  release_label                     = "emr-5.26.0"
  service_role                      = aws_iam_role.iam_emr_default_role.arn

  ec2_attributes {
    subnet_id                         = aws_subnet.main.id
    emr_managed_master_security_group = aws_security_group.allow_all.id
    emr_managed_slave_security_group  = aws_security_group.allow_all.id
    instance_profile                  = aws_iam_instance_profile.emr_profile.arn
  }

  master_instance_group {
    instance_type = "c4.large"
  }

  core_instance_group {
    instance_type  = "c4.large"
    instance_count = 2
  }

  depends_on = [aws_internet_gateway.gw]
}


###

# IAM role for EMR Service
resource "aws_iam_role" "iam_emr_default_role" {
  name = "iam_emr_default_role_%[1]d"

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

resource "aws_iam_role_policy_attachment" "service-attach" {
  role       = aws_iam_role.iam_emr_default_role.id
  policy_arn = aws_iam_policy.iam_emr_default_policy.arn
}

resource "aws_iam_policy" "iam_emr_default_policy" {
  name = "iam_emr_default_policy_%[1]d"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "ec2:AuthorizeSecurityGroupEgress",
            "ec2:AuthorizeSecurityGroupIngress",
            "ec2:CancelSpotInstanceRequests",
            "ec2:CreateNetworkInterface",
            "ec2:CreateSecurityGroup",
            "ec2:CreateTags",
            "ec2:DeleteNetworkInterface",
            "ec2:DeleteSecurityGroup",
            "ec2:DeleteTags",
            "ec2:DescribeAvailabilityZones",
            "ec2:DescribeAccountAttributes",
            "ec2:DescribeDhcpOptions",
            "ec2:DescribeInstanceStatus",
            "ec2:DescribeInstances",
            "ec2:DescribeKeyPairs",
            "ec2:DescribeNetworkAcls",
            "ec2:DescribeNetworkInterfaces",
            "ec2:DescribePrefixLists",
            "ec2:DescribeRouteTables",
            "ec2:DescribeSecurityGroups",
            "ec2:DescribeSpotInstanceRequests",
            "ec2:DescribeSpotPriceHistory",
            "ec2:DescribeSubnets",
            "ec2:DescribeVpcAttribute",
            "ec2:DescribeVpcEndpoints",
            "ec2:DescribeVpcEndpointServices",
            "ec2:DescribeVpcs",
            "ec2:DetachNetworkInterface",
            "ec2:ModifyImageAttribute",
            "ec2:ModifyInstanceAttribute",
            "ec2:RequestSpotInstances",
            "ec2:RevokeSecurityGroupEgress",
            "ec2:RunInstances",
            "ec2:TerminateInstances",
            "ec2:DeleteVolume",
            "ec2:DescribeVolumeStatus",
            "ec2:DescribeVolumes",
            "ec2:DetachVolume",
            "iam:GetRole",
            "iam:GetRolePolicy",
            "iam:ListInstanceProfiles",
            "iam:ListRolePolicies",
            "iam:PassRole",
            "s3:CreateBucket",
            "s3:Get*",
            "s3:List*",
            "sdb:BatchPutAttributes",
            "sdb:Select",
            "sqs:CreateQueue",
            "sqs:Delete*",
            "sqs:GetQueue*",
            "sqs:PurgeQueue",
            "sqs:ReceiveMessage"
        ]
    }]
}
EOT
}

# IAM Role for EC2 Instance Profile
resource "aws_iam_role" "iam_emr_profile_role" {
  name = "iam_emr_profile_role_%[1]d"

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

resource "aws_iam_instance_profile" "emr_profile" {
  name = "emr_profile_%[1]d"
  role = aws_iam_role.iam_emr_profile_role.name
}

resource "aws_iam_role_policy_attachment" "profile-attach" {
  role       = aws_iam_role.iam_emr_profile_role.id
  policy_arn = aws_iam_policy.iam_emr_profile_policy.arn
}

resource "aws_iam_policy" "iam_emr_profile_policy" {
  name = "iam_emr_profile_policy_%[1]d"

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

# IAM Role for autoscaling
resource "aws_iam_role" "emr-autoscaling-role" {
  name               = "EMR_AutoScaling_DefaultRole_%[1]d"
  assume_role_policy = data.aws_iam_policy_document.emr-autoscaling-role-policy.json
}

data "aws_iam_policy_document" "emr-autoscaling-role-policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["elasticmapreduce.${data.aws_partition.current.dns_suffix}", "application-autoscaling.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "emr-autoscaling-role" {
  role       = aws_iam_role.emr-autoscaling-role.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonElasticMapReduceforAutoScalingRole"
}
`

func testAccAWSEmrInstanceGroupConfig_basic(r int) string {
	return fmt.Sprintf(testAccAWSEmrInstanceGroupBase+`
resource "aws_emr_instance_group" "task" {
  cluster_id     = aws_emr_cluster.tf-test-cluster.id
  instance_count = 1
  instance_type  = "c4.large"
}
`, r)
}

func testAccAWSEmrInstanceGroupConfig_BidPrice(r int) string {
	return fmt.Sprintf(testAccAWSEmrInstanceGroupBase+`
resource "aws_emr_instance_group" "task" {
  cluster_id     = aws_emr_cluster.tf-test-cluster.id
  bid_price      = "0.30"
  instance_count = 1
  instance_type  = "c4.large"
}
`, r)
}

func testAccAWSEmrInstanceGroupConfig_ConfigurationsJson(r int, name string) string {
	return fmt.Sprintf(testAccAWSEmrInstanceGroupBase+`
resource "aws_emr_instance_group" "task" {
  cluster_id          = aws_emr_cluster.tf-test-cluster.id
  instance_count      = 1
  instance_type       = "c4.large"
  configurations_json = <<EOF
    [
      {
        "Classification": "yarn-site",
        "Properties": {
          "yarn.nodemanager.node-labels.provider": "config",
          "yarn.nodemanager.node-labels.provider.configured-node-partition": "%s"
        }
      }
    ]
EOF
}
`, r, name)
}

func testAccAWSEmrInstanceGroupConfig_AutoScalingPolicy(r, min, max int) string {
	return fmt.Sprintf(testAccAWSEmrInstanceGroupBase+`
resource "aws_emr_instance_group" "task" {
  cluster_id         = aws_emr_cluster.tf-test-cluster.id
  instance_count     = 1
  instance_type      = "c4.large"
  autoscaling_policy = <<EOT
{
  "Constraints": {
    "MinCapacity": %d,
    "MaxCapacity": %d
  },
  "Rules": [
    {
      "Name": "ScaleOutMemoryPercentage",
      "Description": "Scale out if YARNMemoryAvailablePercentage is less than 15",
      "Action": {
        "SimpleScalingPolicyConfiguration": {
          "AdjustmentType": "CHANGE_IN_CAPACITY",
          "ScalingAdjustment": 1,
          "CoolDown": 300
        }
      },
      "Trigger": {
        "CloudWatchAlarmDefinition": {
          "ComparisonOperator": "LESS_THAN",
          "EvaluationPeriods": 1,
          "MetricName": "YARNMemoryAvailablePercentage",
          "Namespace": "AWS/ElasticMapReduce",
          "Period": 300,
          "Statistic": "AVERAGE",
          "Threshold": 15.0,
          "Unit": "PERCENT"
        }
      }
    }
  ]
}
EOT
}
`, r, min, max)
}

func testAccAWSEmrInstanceGroupConfig_ebsConfig(r int, o bool) string {
	return fmt.Sprintf(testAccAWSEmrInstanceGroupBase+`
resource "aws_emr_instance_group" "task" {
  cluster_id     = aws_emr_cluster.tf-test-cluster.id
  instance_count = 1
  instance_type  = "c4.large"
  ebs_optimized  = %t

  ebs_config {
    size = 10
    type = "gp2"
  }
}
`, r, o)
}

func testAccAWSEmrInstanceGroupConfig_zeroCount(r int) string {
	return fmt.Sprintf(testAccAWSEmrInstanceGroupBase+`
resource "aws_emr_instance_group" "task" {
  cluster_id     = aws_emr_cluster.tf-test-cluster.id
  instance_count = 0
  instance_type  = "c4.large"
}
`, r)
}
