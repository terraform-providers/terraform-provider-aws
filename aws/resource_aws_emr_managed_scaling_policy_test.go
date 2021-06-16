package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsEmrManagedScalingPolicy_basic(t *testing.T) {
	resourceName := "aws_emr_managed_scaling_policy.testpolicy"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheckSkipEmrManagedScalingPolicy(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrManagedScalingPolicyDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrManagedScalingPolicy_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrManagedScalingPolicyExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsEmrManagedScalingPolicy_ComputeLimits_MaximumCoreCapacityUnits(t *testing.T) {
	resourceName := "aws_emr_managed_scaling_policy.testpolicy"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheckSkipEmrManagedScalingPolicy(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrManagedScalingPolicyDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrManagedScalingPolicy_ComputeLimits_MaximumCoreCapacityUnits(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrManagedScalingPolicyExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsEmrManagedScalingPolicy_ComputeLimits_MaximumOndemandCapacityUnits(t *testing.T) {
	resourceName := "aws_emr_managed_scaling_policy.testpolicy"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheckSkipEmrManagedScalingPolicy(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrManagedScalingPolicyDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrManagedScalingPolicy_ComputeLimits_MaximumOndemandCapacityUnits(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrManagedScalingPolicyExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsEmrManagedScalingPolicy_disappears(t *testing.T) {
	resourceName := "aws_emr_managed_scaling_policy.testpolicy"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheckSkipEmrManagedScalingPolicy(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrManagedScalingPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrManagedScalingPolicy_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrManagedScalingPolicyExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEMRManagedScalingPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// testAccErrorCheckSkipEmrManagedScalingPolicy skips tests that have error messages indicating unsupported features
func testAccErrorCheckSkipEmrManagedScalingPolicy(t *testing.T) resource.ErrorCheckFunc {
	return testAccErrorCheckSkipMessagesContaining(t,
		"Managed scaling is not available",
	)
}

func testAccAWSEmrManagedScalingPolicy_basic(r string) string {
	return fmt.Sprintf(testAccAWSEmrManagedScalingPolicyBase+`
resource "aws_emr_managed_scaling_policy" "testpolicy" {
  cluster_id = aws_emr_cluster.test.id
  compute_limits {
    unit_type              = "Instances"
    minimum_capacity_units = 1
    maximum_capacity_units = 2
  }
}
`, r)
}

func testAccAWSEmrManagedScalingPolicy_ComputeLimits_MaximumCoreCapacityUnits(r string, maximumCoreCapacityUnits int) string {
	return fmt.Sprintf(testAccAWSEmrManagedScalingPolicyBase+`
resource "aws_emr_managed_scaling_policy" "testpolicy" {
  cluster_id = aws_emr_cluster.test.id
  compute_limits {
    unit_type                   = "Instances"
    minimum_capacity_units      = 1
    maximum_capacity_units      = 2
    maximum_core_capacity_units = %[2]d
  }
}
`, r, maximumCoreCapacityUnits)
}

func testAccAWSEmrManagedScalingPolicy_ComputeLimits_MaximumOndemandCapacityUnits(r string, maximumOndemandCapacityUnits int) string {
	return fmt.Sprintf(testAccAWSEmrManagedScalingPolicyBase+`
resource "aws_emr_managed_scaling_policy" "testpolicy" {
  cluster_id = aws_emr_cluster.test.id
  compute_limits {
    unit_type                       = "Instances"
    minimum_capacity_units          = 1
    maximum_capacity_units          = 2
    maximum_ondemand_capacity_units = %[2]d
  }
}
`, r, maximumOndemandCapacityUnits)
}

func testAccCheckAWSEmrManagedScalingPolicyExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EMR Managed Scaling Policy ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).emrconn
		resp, err := conn.GetManagedScalingPolicy(&emr.GetManagedScalingPolicyInput{
			ClusterId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if resp.ManagedScalingPolicy == nil {
			return fmt.Errorf("EMR Managed Scaling Policy is empty which shouldn't happen")
		}
		return nil
	}
}

func testAccCheckAWSEmrManagedScalingPolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).emrconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_emr_managed_scaling_policy" {
			continue
		}

		resp, err := conn.GetManagedScalingPolicy(&emr.GetManagedScalingPolicyInput{
			ClusterId: aws.String(rs.Primary.ID),
		})

		if tfawserr.ErrMessageContains(err, "InvalidRequestException", "does not exist") {
			continue
		}

		if tfawserr.ErrMessageContains(err, "ValidationException", "A job flow that is shutting down, terminated, or finished may not be modified") {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading EMR Managed Scaling Policy (%s): %w", rs.Primary.ID, err)
		}

		if resp != nil && resp.ManagedScalingPolicy != nil {
			return fmt.Errorf("EMR Managed Scaling Policy (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

const testAccAWSEmrManagedScalingPolicyBase = `
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

resource "aws_iam_role" "emr_service" {
  name = "%[1]s_default_role"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "elasticmapreduce.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy_attachment" "emr_service" {
  role       = aws_iam_role.emr_service.id
  policy_arn = aws_iam_policy.emr_service.arn
}

resource "aws_iam_policy" "emr_service" {
  name = "%[1]s_emr"

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
resource "aws_iam_instance_profile" "emr_instance_profile" {
  name = "%[1]s_profile"
  role = aws_iam_role.emr_instance_profile.name
}

resource "aws_iam_role" "emr_instance_profile" {
  name = "%[1]s_profile_role"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
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
  name = "%[1]s_profile"

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
  applications                      = ["Hadoop", "Hive"]
  keep_job_flow_alive_when_no_steps = true
  log_uri                           = "s3n://terraform/testlog/"
  name                              = "%[1]s"
  release_label                     = "emr-5.30.1"
  service_role                      = aws_iam_role.emr_service.arn

  master_instance_group {
    instance_type = "c4.large"
  }

  core_instance_group {
    instance_count = 1
    instance_type  = "c4.large"
  }

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
