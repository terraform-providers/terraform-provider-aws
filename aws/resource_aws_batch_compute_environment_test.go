package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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

	out, err := conn.DescribeComputeEnvironments(&batch.DescribeComputeEnvironmentsInput{})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Batch Compute Environment sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Batch Compute Environments: %s", err)
	}
	for _, computeEnvironment := range out.ComputeEnvironments {
		name := aws.StringValue(computeEnvironment.ComputeEnvironmentName)

		if aws.StringValue(computeEnvironment.State) == batch.CEStateEnabled {
			log.Printf("[INFO] Disabling Batch Compute Environment: %s", name)
			err := disableBatchComputeEnvironment(name, 20*time.Minute, conn)

			if err != nil {
				log.Printf("[ERROR] Failed to disable Batch Compute Environment %s: %s", name, err)
				continue
			}
		}

		log.Printf("[INFO] Deleting Batch Compute Environment: %s", name)
		err := deleteBatchComputeEnvironment(name, 20*time.Minute, conn)

		if err != nil {
			log.Printf("[ERROR] Failed to delete Batch Compute Environment %s: %s", name, err)
		}
	}

	return nil
}

func TestAccAWSBatchComputeEnvironment_createEc2(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
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

func TestAccAWSBatchComputeEnvironment_createWithNamePrefix(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigNamePrefix(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(),
					resource.TestMatchResourceAttr(
						"aws_batch_compute_environment.ec2", "compute_environment_name", regexp.MustCompile("^tf_acc_test"),
					),
				),
			},
		},
	})
}

func TestAccAWSBatchComputeEnvironment_createEc2WithTags(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_batch_compute_environment.ec2"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigEC2WithTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.Key1", "Value1"),
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

func TestAccAWSBatchComputeEnvironment_createSpot(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
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
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
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
	resourceName := "aws_batch_compute_environment.ec2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigEC2(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
				),
			},
			{
				Config: testAccAWSBatchComputeEnvironmentConfigEC2UpdateMaxvCpus(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "32"),
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

func TestAccAWSBatchComputeEnvironment_updateInstanceType(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_batch_compute_environment.ec2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigEC2(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "1"),
				),
			},
			{
				Config: testAccAWSBatchComputeEnvironmentConfigEC2UpdateInstanceType(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "2"),
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

func TestAccAWSBatchComputeEnvironment_updateComputeEnvironmentName(t *testing.T) {
	rInt := acctest.RandInt()
	expectedName := fmt.Sprintf("tf_acc_test_%d", rInt)
	expectedUpdatedName := fmt.Sprintf("tf_acc_test_updated_%d", rInt)
	resourceName := "aws_batch_compute_environment.ec2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigEC2(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", expectedName),
				),
			},
			{
				Config: testAccAWSBatchComputeEnvironmentConfigEC2UpdateComputeEnvironmentName(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", expectedUpdatedName),
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

func TestAccAWSBatchComputeEnvironment_createEc2WithoutComputeResources(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
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
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
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

func TestAccAWSBatchComputeEnvironment_launchTemplate(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_batch_compute_environment.ec2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigLaunchTemplate(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(),
					resource.TestCheckResourceAttr(resourceName,
						"compute_resources.0.launch_template.#",
						"1"),
					resource.TestCheckResourceAttr(resourceName,
						"compute_resources.0.launch_template.0.launch_template_name",
						fmt.Sprintf("tf_acc_test_%d", rInt)),
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

func TestAccAWSBatchComputeEnvironment_UpdateLaunchTemplate(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_batch_compute_environment.ec2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentUpdateLaunchTemplateInExistingComputeEnvironment(rInt, "$Default"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.0.version", "$Default"),
				),
			},
			{
				Config: testAccAWSBatchComputeEnvironmentUpdateLaunchTemplateInExistingComputeEnvironment(rInt, "$Latest"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.0.version", "$Latest"),
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

func TestAccAWSBatchComputeEnvironment_createSpotWithAllocationStrategy(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigSpotWithAllocationStrategy(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(),
					resource.TestCheckResourceAttr("aws_batch_compute_environment.ec2", "compute_resources.0.allocation_strategy", "BEST_FIT"),
				),
			},
		},
	})
}

func TestAccAWSBatchComputeEnvironment_createSpotWithoutBidPercentage(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
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
	resourceName := "aws_batch_compute_environment.ec2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBatchComputeEnvironmentConfigEC2UpdateState(rInt, batch.CEStateEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(),
					resource.TestCheckResourceAttr(resourceName, "state", batch.CEStateEnabled),
				),
			},
			{
				Config: testAccAWSBatchComputeEnvironmentConfigEC2UpdateState(rInt, batch.CEStateDisabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBatchComputeEnvironmentExists(),
					resource.TestCheckResourceAttr(resourceName, "state", batch.CEStateDisabled),
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

func testAccPreCheckAWSBatch(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).batchconn

	input := &batch.DescribeComputeEnvironmentsInput{}

	_, err := conn.DescribeComputeEnvironments(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSBatchComputeEnvironmentConfigBase(rInt int) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

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
		"Service": "ec2.${data.aws_partition.current.dns_suffix}"
	    }
	}
    ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "ecs_instance_role" {
  role       = "${aws_iam_role.ecs_instance_role.name}"
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

resource "aws_iam_instance_profile" "ecs_instance_role" {
  name = "tf_acc_test_batch_ip_%d"
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
		"Service": "batch.${data.aws_partition.current.dns_suffix}"
	    }
	}
    ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "aws_batch_service_role" {
  role       = "${aws_iam_role.aws_batch_service_role.name}"
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSBatchServiceRole"
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
		"Service": "spotfleet.${data.aws_partition.current.dns_suffix}"
	    }
	}
    ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "aws_ec2_spot_fleet_role" {
  role       = "${aws_iam_role.aws_ec2_spot_fleet_role.name}"
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonEC2SpotFleetTaggingRole"
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
  vpc_id     = "${aws_vpc.test_acc.id}"
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

func testAccAWSBatchComputeEnvironmentConfigNamePrefix(rInt int) string {
	return testAccAWSBatchComputeEnvironmentConfigBase(rInt) + `
resource "aws_batch_compute_environment" "ec2" {
  compute_environment_name_prefix = "tf_acc_test"
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
`
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

func testAccAWSBatchComputeEnvironmentConfigSpotWithAllocationStrategy(rInt int) string {
	return testAccAWSBatchComputeEnvironmentConfigBase(rInt) + fmt.Sprintf(`
resource "aws_batch_compute_environment" "ec2" {
  compute_environment_name = "tf_acc_test_%d"
  compute_resources {
	allocation_strategy = "BEST_FIT"
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

func testAccAWSBatchComputeEnvironmentConfigLaunchTemplate(rInt int) string {
	return testAccAWSBatchComputeEnvironmentConfigBase(rInt) + fmt.Sprintf(`
resource "aws_launch_template" "foo" {
  name = "tf_acc_test_%d"
}

resource "aws_batch_compute_environment" "ec2" {
  compute_environment_name = "tf_acc_test_%d"
  compute_resources {
    instance_role = "${aws_iam_instance_profile.ecs_instance_role.arn}"
    instance_type = [
      "c4.large",
    ]
    launch_template {
			launch_template_name = "${aws_launch_template.foo.name}"
		}
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
`, rInt, rInt)
}

func testAccAWSBatchComputeEnvironmentUpdateLaunchTemplateInExistingComputeEnvironment(rInt int, version string) string {
	return testAccAWSBatchComputeEnvironmentConfigBase(rInt) + fmt.Sprintf(`
resource "aws_launch_template" "foo" {
  name = "tf_acc_test_%d"
}

resource "aws_batch_compute_environment" "ec2" {
  compute_environment_name = "tf_acc_test_%d"
  compute_resources {
    instance_role = "${aws_iam_instance_profile.ecs_instance_role.arn}"
    instance_type = [
      "c4.large",
    ]
    launch_template {
			launch_template_name = "${aws_launch_template.foo.name}"
			version              = "%s"
		}
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
`, rInt, rInt, version)
}
