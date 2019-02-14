package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_batch_compute_environment", &resource.Sweeper{
		Name: "aws_batch_compute_environment",
		Dependencies: []string{
			"aws_batch_job_queue",
		},
		F: testSweepBatchComputeEnvironments,
	})
}

func testSweepBatchComputeEnvironments(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).batchconn

	prefixes := []string{
		"tf_acc",
	}

	out, err := conn.DescribeComputeEnvironments(&batch.DescribeComputeEnvironmentsInput{})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Batch Compute Environment sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Batch Compute Environments: %s", err)
	}
	for _, computeEnvironment := range out.ComputeEnvironments {
		name := computeEnvironment.ComputeEnvironmentName
		skip := true
		for _, prefix := range prefixes {
			if strings.HasPrefix(*name, prefix) {
				skip = false
				break
			}
		}
		if skip {
			log.Printf("[INFO] Skipping Batch Compute Environment: %s", *name)
			continue
		}

		log.Printf("[INFO] Disabling Batch Compute Environment: %s", *name)
		err := disableBatchComputeEnvironment(*name, 20*time.Minute, conn)
		if err != nil {
			log.Printf("[ERROR] Failed to disable Batch Compute Environment %s: %s", *name, err)
			continue
		}

		log.Printf("[INFO] Deleting Batch Compute Environment: %s", *name)
		err = deleteBatchComputeEnvironment(*name, 20*time.Minute, conn)
		if err != nil {
			log.Printf("[ERROR] Failed to delete Batch Compute Environment %s: %s", *name, err)
		}
	}

	return nil
}

func TestAccAWSBatchComputeEnvironment_createEc2(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigEC2(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(),
				),
			},
		},
	})
}

func TestAccAWSBatchComputeEnvironment_createEc2WithTags(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigEC2WithTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(),
					resource.TestCheckResourceAttr("aws_batch_compute_environment.ec2", "compute_resources.0.tags.%", "1"),
					resource.TestCheckResourceAttr("aws_batch_compute_environment.ec2", "compute_resources.0.tags.Key1", "Value1"),
				),
			},
		},
	})
}

func TestAccAWSBatchComputeEnvironment_createSpot(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigSpot(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(),
				),
			},
		},
	})
}

func TestAccAWSBatchComputeEnvironment_createUnmanaged(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigUnmanaged(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(),
				),
			},
		},
	})
}

func TestAccAWSBatchComputeEnvironment_updateMaxvCpus(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigEC2(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(),
					resource.TestCheckResourceAttr("aws_batch_compute_environment.ec2", "compute_resources.0.max_vcpus", "16"),
				),
			},
			{
				Config: testAccAWSBatchComputeEnvironmentConfigEC2UpdateMaxvCpus(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(),
					resource.TestCheckResourceAttr("aws_batch_compute_environment.ec2", "compute_resources.0.max_vcpus", "32"),
				),
			},
		},
	})
}

func TestAccAWSBatchComputeEnvironment_updateInstanceType(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigEC2(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(),
					resource.TestCheckResourceAttr("aws_batch_compute_environment.ec2", "compute_resources.0.instance_type.#", "1"),
				),
			},
			{
				Config: testAccAWSBatchComputeEnvironmentConfigEC2UpdateInstanceType(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(),
					resource.TestCheckResourceAttr("aws_batch_compute_environment.ec2", "compute_resources.0.instance_type.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSBatchComputeEnvironment_updateComputeEnvironmentName(t *testing.T) {
	rInt := acctest.RandInt()
	expectedName := fmt.Sprintf("tf_acc_test_%d", rInt)
	expectedUpdatedName := fmt.Sprintf("tf_acc_test_updated_%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigEC2(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(),
					resource.TestCheckResourceAttr("aws_batch_compute_environment.ec2", "compute_environment_name", expectedName),
				),
			},
			{
				Config: testAccAWSBatchComputeEnvironmentConfigEC2UpdateComputeEnvironmentName(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(),
					resource.TestCheckResourceAttr("aws_batch_compute_environment.ec2", "compute_environment_name", expectedUpdatedName),
				),
			},
		},
	})
}

func TestAccAWSBatchComputeEnvironment_createEc2WithoutComputeResources(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSBatchComputeEnvironmentConfigEC2WithoutComputeResources(rInt),
				ExpectError: regexp.MustCompile(`One compute environment is expected, but no compute environments are set`),
			},
		},
	})
}

func TestAccAWSBatchComputeEnvironment_createUnmanagedWithComputeResources(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigUnmanagedWithComputeResources(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(),
					resource.TestCheckResourceAttr("aws_batch_compute_environment.unmanaged", "type", "UNMANAGED"),
				),
			},
		},
	})
}

func TestAccAWSBatchComputeEnvironment_createSpotWithoutBidPercentage(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSBatchComputeEnvironmentConfigSpotWithoutBidPercentage(rInt),
				ExpectError: regexp.MustCompile(`ComputeResources.spotIamFleetRole cannot not be null or empty`),
			},
		},
	})
}

func TestAccAWSBatchComputeEnvironment_updateState(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigEC2UpdateState(rInt, batch.CEStateEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(),
					resource.TestCheckResourceAttr("aws_batch_compute_environment.ec2", "state", batch.CEStateEnabled),
				),
			},
			{
				Config: testAccAWSBatchComputeEnvironmentConfigEC2UpdateState(rInt, batch.CEStateDisabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(),
					resource.TestCheckResourceAttr("aws_batch_compute_environment.ec2", "state", batch.CEStateDisabled),
				),
			},
		},
	})
}

func testAccCheckBatchComputeEnvironmentDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).batchconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_batch_compute_environment" {
			continue
		}

		result, err := conn.DescribeComputeEnvironments(&batch.DescribeComputeEnvironmentsInput{
			ComputeEnvironments: []*string{
				aws.String(rs.Primary.ID),
			},
		})

		if err != nil {
			return fmt.Errorf("Error occurred when get compute environment information.")
		}
		if len(result.ComputeEnvironments) == 1 {
			return fmt.Errorf("Compute environment still exists.")
		}

	}

	return nil
}

func testAccCheckAwsBatchComputeEnvironmentExists() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).batchconn

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_batch_compute_environment" {
				continue
			}

			result, err := conn.DescribeComputeEnvironments(&batch.DescribeComputeEnvironmentsInput{
				ComputeEnvironments: []*string{
					aws.String(rs.Primary.ID),
				},
			})

			if err != nil {
				return fmt.Errorf("Error occurred when get compute environment information.")
			}
			if len(result.ComputeEnvironments) == 0 {
				return fmt.Errorf("Compute environment doesn't exists.")
			} else if len(result.ComputeEnvironments) >= 2 {
				return fmt.Errorf("Too many compute environments exist.")
			}
		}

		return nil
	}
}

func testAccAWSBatchComputeEnvironmentConfigBase(rInt int) string {
	return fmt.Sprintf(`

########## ecs_instance_role ##########

resource "aws_iam_role" "ecs_instance_role" {
  name = "tf_acc_test_batch_inst_role_%d"
  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
	{
	    "Action": "sts:AssumeRole",
	    "Effect": "Allow",
	    "Principal": {
		"Service": "ec2.amazonaws.com"
	    }
	}
    ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "ecs_instance_role" {
  role       = "${aws_iam_role.ecs_instance_role.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

resource "aws_iam_instance_profile" "ecs_instance_role" {
  name  = "tf_acc_test_batch_ip_%d"
  role = "${aws_iam_role.ecs_instance_role.name}"
}

########## aws_batch_service_role ##########

resource "aws_iam_role" "aws_batch_service_role" {
  name = "tf_acc_test_batch_svc_role_%d"
  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
	{
	    "Action": "sts:AssumeRole",
	    "Effect": "Allow",
	    "Principal": {
		"Service": "batch.amazonaws.com"
	    }
	}
    ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "aws_batch_service_role" {
  role       = "${aws_iam_role.aws_batch_service_role.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSBatchServiceRole"
}

########## aws_ec2_spot_fleet_role ##########

resource "aws_iam_role" "aws_ec2_spot_fleet_role" {
  name = "tf_acc_test_batch_spot_fleet_role_%d"
  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
	{
	    "Action": "sts:AssumeRole",
	    "Effect": "Allow",
	    "Principal": {
		"Service": "spotfleet.amazonaws.com"
	    }
	}
    ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "aws_ec2_spot_fleet_role" {
  role       = "${aws_iam_role.aws_ec2_spot_fleet_role.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEC2SpotFleetRole"
}

########## security group ##########

resource "aws_security_group" "test_acc" {
  name = "tf_acc_test_batch_sg_%d"
}

########## subnets ##########

resource "aws_vpc" "test_acc" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-batch-compute-environment"
  }
}

resource "aws_subnet" "test_acc" {
  vpc_id = "${aws_vpc.test_acc.id}"
  cidr_block = "10.1.1.0/24"
  tags = {
    Name = "tf-acc-batch-compute-environment"
  }
}
`, rInt, rInt, rInt, rInt, rInt)
}

func testAccAWSBatchComputeEnvironmentConfigEC2(rInt int) string {
	return testAccAWSBatchComputeEnvironmentConfigBase(rInt) + fmt.Sprintf(`
resource "aws_batch_compute_environment" "ec2" {
  compute_environment_name = "tf_acc_test_%d"
  compute_resources {
    instance_role = "${aws_iam_instance_profile.ecs_instance_role.arn}"
    instance_type = [
      "c4.large",
    ]
    max_vcpus = 16
    min_vcpus = 0
    security_group_ids = [
      "${aws_security_group.test_acc.id}"
    ]
    subnets = [
      "${aws_subnet.test_acc.id}"
    ]
    type = "EC2"
  }
  service_role = "${aws_iam_role.aws_batch_service_role.arn}"
  type = "MANAGED"
  depends_on = ["aws_iam_role_policy_attachment.aws_batch_service_role"]
}
`, rInt)
}

func testAccAWSBatchComputeEnvironmentConfigEC2WithTags(rInt int) string {
	return testAccAWSBatchComputeEnvironmentConfigBase(rInt) + fmt.Sprintf(`
resource "aws_batch_compute_environment" "ec2" {
  compute_environment_name = "tf_acc_test_%d"
  compute_resources {
    instance_role = "${aws_iam_instance_profile.ecs_instance_role.arn}"
    instance_type = [
      "c4.large",
    ]
    max_vcpus = 16
    min_vcpus = 0
    security_group_ids = [
      "${aws_security_group.test_acc.id}"
    ]
    subnets = [
      "${aws_subnet.test_acc.id}"
    ]
    type = "EC2"
  tags = {
      Key1 = "Value1"
    }
  }
  service_role = "${aws_iam_role.aws_batch_service_role.arn}"
	type = "MANAGED"
  depends_on = ["aws_iam_role_policy_attachment.aws_batch_service_role"]
}
`, rInt)
}

func testAccAWSBatchComputeEnvironmentConfigSpot(rInt int) string {
	return testAccAWSBatchComputeEnvironmentConfigBase(rInt) + fmt.Sprintf(`
resource "aws_batch_compute_environment" "spot" {
  compute_environment_name = "tf_acc_test_%d"
  compute_resources {
    bid_percentage = 100
    instance_role = "${aws_iam_instance_profile.ecs_instance_role.arn}"
    instance_type = [
      "c4.large",
    ]
    max_vcpus = 16
    min_vcpus = 0
    security_group_ids = [
      "${aws_security_group.test_acc.id}"
    ]
    spot_iam_fleet_role = "${aws_iam_role.aws_ec2_spot_fleet_role.arn}"
    subnets = [
      "${aws_subnet.test_acc.id}"
    ]
    type = "SPOT"
  }
  service_role = "${aws_iam_role.aws_batch_service_role.arn}"
  type = "MANAGED"
  depends_on = ["aws_iam_role_policy_attachment.aws_batch_service_role"]
}
`, rInt)
}

func testAccAWSBatchComputeEnvironmentConfigUnmanaged(rInt int) string {
	return testAccAWSBatchComputeEnvironmentConfigBase(rInt) + fmt.Sprintf(`
resource "aws_batch_compute_environment" "unmanaged" {
  compute_environment_name = "tf_acc_test_%d"
  service_role = "${aws_iam_role.aws_batch_service_role.arn}"
  type = "UNMANAGED"
  depends_on = ["aws_iam_role_policy_attachment.aws_batch_service_role"]
}
`, rInt)
}

func testAccAWSBatchComputeEnvironmentConfigEC2UpdateMaxvCpus(rInt int) string {
	return testAccAWSBatchComputeEnvironmentConfigBase(rInt) + fmt.Sprintf(`
resource "aws_batch_compute_environment" "ec2" {
  compute_environment_name = "tf_acc_test_%d"
  compute_resources {
    instance_role = "${aws_iam_instance_profile.ecs_instance_role.arn}"
    instance_type = [
      "c4.large",
    ]
    max_vcpus = 32
    min_vcpus = 0
    security_group_ids = [
      "${aws_security_group.test_acc.id}"
    ]
    subnets = [
      "${aws_subnet.test_acc.id}"
    ]
    type = "EC2"
  }
  service_role = "${aws_iam_role.aws_batch_service_role.arn}"
  type = "MANAGED"
  depends_on = ["aws_iam_role_policy_attachment.aws_batch_service_role"]
}
`, rInt)
}

func testAccAWSBatchComputeEnvironmentConfigEC2UpdateInstanceType(rInt int) string {
	return testAccAWSBatchComputeEnvironmentConfigBase(rInt) + fmt.Sprintf(`
resource "aws_batch_compute_environment" "ec2" {
  compute_environment_name = "tf_acc_test_%d"
  compute_resources {
    instance_role = "${aws_iam_instance_profile.ecs_instance_role.arn}"
    instance_type = [
      "c4.large",
      "c4.xlarge",
    ]
    max_vcpus = 16
    min_vcpus = 0
    security_group_ids = [
      "${aws_security_group.test_acc.id}"
    ]
    subnets = [
      "${aws_subnet.test_acc.id}"
    ]
    type = "EC2"
  }
  service_role = "${aws_iam_role.aws_batch_service_role.arn}"
  type = "MANAGED"
  depends_on = ["aws_iam_role_policy_attachment.aws_batch_service_role"]
}
`, rInt)
}

func testAccAWSBatchComputeEnvironmentConfigEC2UpdateState(rInt int, state string) string {
	return testAccAWSBatchComputeEnvironmentConfigBase(rInt) + fmt.Sprintf(`
resource "aws_batch_compute_environment" "ec2" {
  compute_environment_name = "tf_acc_test_%d"
  compute_resources {
    instance_role = "${aws_iam_instance_profile.ecs_instance_role.arn}"
    instance_type = [
      "c4.large",
    ]
    max_vcpus = 16
    min_vcpus = 0
    security_group_ids = [
      "${aws_security_group.test_acc.id}"
    ]
    subnets = [
      "${aws_subnet.test_acc.id}"
    ]
    type = "EC2"
  }
  service_role = "${aws_iam_role.aws_batch_service_role.arn}"
  type = "MANAGED"
  state = "%s"
  depends_on = ["aws_iam_role_policy_attachment.aws_batch_service_role"]
}
`, rInt, state)
}

func testAccAWSBatchComputeEnvironmentConfigEC2UpdateComputeEnvironmentName(rInt int) string {
	return testAccAWSBatchComputeEnvironmentConfigBase(rInt) + fmt.Sprintf(`
resource "aws_batch_compute_environment" "ec2" {
  compute_environment_name = "tf_acc_test_updated_%d"
  compute_resources {
    instance_role = "${aws_iam_instance_profile.ecs_instance_role.arn}"
    instance_type = [
      "c4.large",
    ]
    max_vcpus = 16
    min_vcpus = 0
    security_group_ids = [
      "${aws_security_group.test_acc.id}"
    ]
    subnets = [
      "${aws_subnet.test_acc.id}"
    ]
    type = "EC2"
  }
  service_role = "${aws_iam_role.aws_batch_service_role.arn}"
  type = "MANAGED"
  depends_on = ["aws_iam_role_policy_attachment.aws_batch_service_role"]
}
`, rInt)
}

func testAccAWSBatchComputeEnvironmentConfigEC2WithoutComputeResources(rInt int) string {
	return testAccAWSBatchComputeEnvironmentConfigBase(rInt) + fmt.Sprintf(`
resource "aws_batch_compute_environment" "ec2" {
  compute_environment_name = "tf_acc_test_%d"
  service_role = "${aws_iam_role.aws_batch_service_role.arn}"
  type = "MANAGED"
  depends_on = ["aws_iam_role_policy_attachment.aws_batch_service_role"]
}
`, rInt)
}

func testAccAWSBatchComputeEnvironmentConfigUnmanagedWithComputeResources(rInt int) string {
	return testAccAWSBatchComputeEnvironmentConfigBase(rInt) + fmt.Sprintf(`
resource "aws_batch_compute_environment" "unmanaged" {
  compute_environment_name = "tf_acc_test_%d"
  compute_resources {
    instance_role = "${aws_iam_instance_profile.ecs_instance_role.arn}"
    instance_type = [
      "c4.large",
    ]
    max_vcpus = 16
    min_vcpus = 0
    security_group_ids = [
      "${aws_security_group.test_acc.id}"
    ]
    subnets = [
      "${aws_subnet.test_acc.id}"
    ]
    type = "EC2"
  }
  service_role = "${aws_iam_role.aws_batch_service_role.arn}"
  type = "UNMANAGED"
  depends_on = ["aws_iam_role_policy_attachment.aws_batch_service_role"]
}
`, rInt)
}

func testAccAWSBatchComputeEnvironmentConfigSpotWithoutBidPercentage(rInt int) string {
	return testAccAWSBatchComputeEnvironmentConfigBase(rInt) + fmt.Sprintf(`
resource "aws_batch_compute_environment" "ec2" {
  compute_environment_name = "tf_acc_test_%d"
  compute_resources {
    instance_role = "${aws_iam_instance_profile.ecs_instance_role.arn}"
    instance_type = [
      "c4.large",
    ]
    max_vcpus = 16
    min_vcpus = 0
    security_group_ids = [
      "${aws_security_group.test_acc.id}"
    ]
    subnets = [
      "${aws_subnet.test_acc.id}"
    ]
    type = "SPOT"
  }
  service_role = "${aws_iam_role.aws_batch_service_role.arn}"
  type = "MANAGED"
  depends_on = ["aws_iam_role_policy_attachment.aws_batch_service_role"]
}
`, rInt)
}
