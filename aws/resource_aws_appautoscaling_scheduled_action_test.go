package aws

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/applicationautoscaling"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSAppautoscalingScheduledAction_dynamo(t *testing.T) {
	ts := time.Now().AddDate(0, 0, 1).Format("2006-01-02T15:04:05")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppautoscalingScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppautoscalingScheduledActionConfig_DynamoDB(acctest.RandString(5), ts),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppautoscalingScheduledActionExists("aws_appautoscaling_scheduled_action.hoge"),
				),
			},
		},
	})
}

func TestAccAWSAppautoscalingScheduledAction_ECS(t *testing.T) {
	ts := time.Now().AddDate(0, 0, 1).Format("2006-01-02T15:04:05")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppautoscalingScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppautoscalingScheduledActionConfig_ECS(acctest.RandString(5), ts),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppautoscalingScheduledActionExists("aws_appautoscaling_scheduled_action.hoge"),
				),
			},
		},
	})
}

func TestAccAWSAppautoscalingScheduledAction_EMR(t *testing.T) {
	ts := time.Now().AddDate(0, 0, 1).Format("2006-01-02T15:04:05")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppautoscalingScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppautoscalingScheduledActionConfig_EMR(acctest.RandString(5), ts),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppautoscalingScheduledActionExists("aws_appautoscaling_scheduled_action.hoge"),
				),
			},
		},
	})
}

func TestAccAWSAppautoscalingScheduledAction_Name_Duplicate(t *testing.T) {
	resourceName := "aws_appautoscaling_scheduled_action.test"
	resourceName2 := "aws_appautoscaling_scheduled_action.test2"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppautoscalingScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppautoscalingScheduledActionConfig_Name_Duplicate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppautoscalingScheduledActionExists(resourceName),
					testAccCheckAwsAppautoscalingScheduledActionExists(resourceName2),
				),
			},
		},
	})
}

func TestAccAWSAppautoscalingScheduledAction_SpotFleet(t *testing.T) {
	ts := time.Now().AddDate(0, 0, 1).Format("2006-01-02T15:04:05")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppautoscalingScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppautoscalingScheduledActionConfig_SpotFleet(acctest.RandString(5), ts, validUntil),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppautoscalingScheduledActionExists("aws_appautoscaling_scheduled_action.hoge"),
				),
			},
		},
	})
}

func testAccCheckAwsAppautoscalingScheduledActionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).appautoscalingconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appautoscaling_scheduled_action" {
			continue
		}

		input := &applicationautoscaling.DescribeScheduledActionsInput{
			ResourceId:           aws.String(rs.Primary.Attributes["resource_id"]),
			ScheduledActionNames: []*string{aws.String(rs.Primary.Attributes["name"])},
			ServiceNamespace:     aws.String(rs.Primary.Attributes["service_namespace"]),
		}
		resp, err := conn.DescribeScheduledActions(input)
		if err != nil {
			return err
		}
		if len(resp.ScheduledActions) > 0 {
			return fmt.Errorf("Appautoscaling Scheduled Action (%s) not deleted", rs.Primary.Attributes["name"])
		}
	}
	return nil
}

func testAccCheckAwsAppautoscalingScheduledActionExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		return nil
	}
}

func testAccAppautoscalingScheduledActionConfig_DynamoDB(rName, ts string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "hoge" {
  name           = "tf-ddb-%s"
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "UserID"

  attribute {
    name = "UserID"
    type = "S"
  }
}

resource "aws_appautoscaling_target" "read" {
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.hoge.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  min_capacity       = 1
  max_capacity       = 10
}

resource "aws_appautoscaling_scheduled_action" "hoge" {
  name               = "tf-appauto-%s"
  service_namespace  = aws_appautoscaling_target.read.service_namespace
  resource_id        = aws_appautoscaling_target.read.resource_id
  scalable_dimension = aws_appautoscaling_target.read.scalable_dimension
  schedule           = "at(%s)"

  scalable_target_action {
    min_capacity = 1
    max_capacity = 10
  }
}
`, rName, rName, ts)
}

func testAccAppautoscalingScheduledActionConfig_ECS(rName, ts string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "hoge" {
  name = "tf-ecs-cluster-%s"
}

resource "aws_ecs_task_definition" "hoge" {
  family = "foobar%s"

  container_definitions = <<EOF
[
  {
    "name": "busybox",
    "image": "busybox:latest",
    "cpu": 10,
    "memory": 128,
    "essential": true
  }
]
EOF
}

resource "aws_ecs_service" "hoge" {
  name            = "tf-ecs-service-%s"
  cluster         = aws_ecs_cluster.hoge.id
  task_definition = aws_ecs_task_definition.hoge.arn
  desired_count   = 1

  deployment_maximum_percent         = 200
  deployment_minimum_healthy_percent = 50
}

resource "aws_appautoscaling_target" "hoge" {
  service_namespace  = "ecs"
  resource_id        = "service/${aws_ecs_cluster.hoge.name}/${aws_ecs_service.hoge.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  min_capacity       = 1
  max_capacity       = 3
}

resource "aws_appautoscaling_scheduled_action" "hoge" {
  name               = "tf-appauto-%s"
  service_namespace  = aws_appautoscaling_target.hoge.service_namespace
  resource_id        = aws_appautoscaling_target.hoge.resource_id
  scalable_dimension = aws_appautoscaling_target.hoge.scalable_dimension
  schedule           = "at(%s)"

  scalable_target_action {
    min_capacity = 1
    max_capacity = 5
  }
}
`, rName, rName, rName, rName, ts)
}

func testAccAppautoscalingScheduledActionConfig_EMR(rName, ts string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  # The requested instance type c4.large is not supported in the requested availability zone.
  exclude_zone_ids = ["usw2-az4"]
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_partition" "current" {}

resource "aws_emr_cluster" "hoge" {
  name          = "tf-emr-%s"
  release_label = "emr-5.4.0"
  applications  = ["Spark"]

  ec2_attributes {
    subnet_id                         = aws_subnet.hoge.id
    emr_managed_master_security_group = aws_security_group.hoge.id
    emr_managed_slave_security_group  = aws_security_group.hoge.id
    instance_profile                  = aws_iam_instance_profile.instance_profile.arn
  }

  master_instance_group {
    instance_type = "c4.large"
  }

  core_instance_group {
    instance_count = 2
    instance_type  = "c4.large"
  }

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = true

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = [aws_main_route_table_association.hoge]

  service_role     = aws_iam_role.emr_role.arn
  autoscaling_role = aws_iam_role.autoscale_role.arn
}

resource "aws_emr_instance_group" "hoge" {
  cluster_id     = aws_emr_cluster.hoge.id
  instance_count = 1
  instance_type  = "c4.large"
}

resource "aws_security_group" "hoge" {
  name        = "tf-sg-%s"
  description = "Allow all inbound traffic"
  vpc_id      = aws_vpc.hoge.id

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

  depends_on = [aws_subnet.hoge]

  lifecycle {
    ignore_changes = [
      ingress,
      egress,
    ]
  }
}

resource "aws_vpc" "hoge" {
  cidr_block           = "168.31.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-appautoscaling-scheduled-action-emr"
  }
}

resource "aws_subnet" "hoge" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "168.31.0.0/20"
  vpc_id            = aws_vpc.hoge.id

  tags = {
    Name = "tf-acc-appautoscaling-scheduled-action"
  }
}

resource "aws_internet_gateway" "hoge" {
  vpc_id = aws_vpc.hoge.id
}

resource "aws_route_table" "hoge" {
  vpc_id = aws_vpc.hoge.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.hoge.id
  }
}

resource "aws_main_route_table_association" "hoge" {
  vpc_id         = aws_vpc.hoge.id
  route_table_id = aws_route_table.hoge.id
}

resource "aws_iam_role" "emr_role" {
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

resource "aws_iam_role_policy_attachment" "emr_role" {
  role       = aws_iam_role.emr_role.id
  policy_arn = aws_iam_policy.emr_policy.arn
}

resource "aws_iam_policy" "emr_policy" {
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
resource "aws_iam_role" "instance_role" {
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

resource "aws_iam_instance_profile" "instance_profile" {
  name = "tf-emr-profile-%s"
  role = aws_iam_role.instance_role.name
}

resource "aws_iam_role_policy_attachment" "instance_role" {
  role       = aws_iam_role.instance_role.id
  policy_arn = aws_iam_policy.instance_policy.arn
}

resource "aws_iam_policy" "instance_policy" {
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
resource "aws_iam_role" "autoscale_role" {
  assume_role_policy = data.aws_iam_policy_document.autoscale_role.json
}

data "aws_iam_policy_document" "autoscale_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["elasticmapreduce.${data.aws_partition.current.dns_suffix}", "application-autoscaling.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "autoscale_role" {
  role       = aws_iam_role.autoscale_role.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonElasticMapReduceforAutoScalingRole"
}

resource "aws_appautoscaling_target" "hoge" {
  service_namespace  = "elasticmapreduce"
  resource_id        = "instancegroup/${aws_emr_cluster.hoge.id}/${aws_emr_instance_group.hoge.id}"
  scalable_dimension = "elasticmapreduce:instancegroup:InstanceCount"
  role_arn           = aws_iam_role.autoscale_role.arn
  min_capacity       = 1
  max_capacity       = 5
}

resource "aws_appautoscaling_scheduled_action" "hoge" {
  name               = "tf-appauto-%s"
  service_namespace  = aws_appautoscaling_target.hoge.service_namespace
  resource_id        = aws_appautoscaling_target.hoge.resource_id
  scalable_dimension = aws_appautoscaling_target.hoge.scalable_dimension
  schedule           = "at(%s)"

  scalable_target_action {
    min_capacity = 1
    max_capacity = 5
  }
}
`, rName, rName, rName, rName, ts)
}

func testAccAppautoscalingScheduledActionConfig_Name_Duplicate(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test2" {
  name           = "%[1]s-2"
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "UserID"

  attribute {
    name = "UserID"
    type = "S"
  }
}

resource "aws_appautoscaling_target" "test2" {
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.test2.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  min_capacity       = 1
  max_capacity       = 10
}

resource "aws_appautoscaling_scheduled_action" "test2" {
  name               = %[1]q
  service_namespace  = aws_appautoscaling_target.test2.service_namespace
  resource_id        = aws_appautoscaling_target.test2.resource_id
  scalable_dimension = aws_appautoscaling_target.test2.scalable_dimension
  schedule           = "cron(0 17 * * ? *)"

  scalable_target_action {
    min_capacity = 1
    max_capacity = 10
  }
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "UserID"

  attribute {
    name = "UserID"
    type = "S"
  }
}

resource "aws_appautoscaling_target" "test" {
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.test.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  min_capacity       = 1
  max_capacity       = 10
}

resource "aws_appautoscaling_scheduled_action" "test" {
  name               = %[1]q
  service_namespace  = aws_appautoscaling_target.test.service_namespace
  resource_id        = aws_appautoscaling_target.test.resource_id
  scalable_dimension = aws_appautoscaling_target.test.scalable_dimension
  schedule           = "cron(0 17 * * ? *)"

  scalable_target_action {
    min_capacity = 1
    max_capacity = 10
  }
}
`, rName)
}

func testAccAppautoscalingScheduledActionConfig_SpotFleet(rName, ts, validUntil string) string {
	return fmt.Sprintf(`
data "aws_ami" "amzn-ami-minimal-hvm-ebs" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-minimal-hvm-*"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role" "fleet_role" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "spotfleet.${data.aws_partition.current.dns_suffix}",
          "ec2.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "fleet_role_policy" {
  role       = aws_iam_role.fleet_role.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonEC2SpotFleetTaggingRole"
}

resource "aws_spot_fleet_request" "hoge" {
  iam_fleet_role                      = aws_iam_role.fleet_role.arn
  spot_price                          = "0.005"
  target_capacity                     = 2
  valid_until                         = %[3]q
  terminate_instances_with_expiration = true

  launch_specification {
    instance_type = "m3.medium"
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  }
}

resource "aws_appautoscaling_target" "hoge" {
  service_namespace  = "ec2"
  resource_id        = "spot-fleet-request/${aws_spot_fleet_request.hoge.id}"
  scalable_dimension = "ec2:spot-fleet-request:TargetCapacity"
  min_capacity       = 1
  max_capacity       = 3
}

resource "aws_appautoscaling_scheduled_action" "hoge" {
  name               = "tf-appauto-%[1]s"
  service_namespace  = aws_appautoscaling_target.hoge.service_namespace
  resource_id        = aws_appautoscaling_target.hoge.resource_id
  scalable_dimension = aws_appautoscaling_target.hoge.scalable_dimension
  schedule           = "at(%[2]s)"

  scalable_target_action {
    min_capacity = 1
    max_capacity = 3
  }
}
`, rName, ts, validUntil)
}
