package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_batch_job_queue", &resource.Sweeper{
		Name: "aws_batch_job_queue",
		F:    testSweepBatchJobQueues,
	})
}

func testSweepBatchJobQueues(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).batchconn

	out, err := conn.DescribeJobQueues(&batch.DescribeJobQueuesInput{})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Batch Job Queue sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Batch Job Queues: %s", err)
	}
	for _, jobQueue := range out.JobQueues {
		name := jobQueue.JobQueueName

		log.Printf("[INFO] Disabling Batch Job Queue: %s", *name)
		err := disableBatchJobQueue(*name, conn)
		if err != nil {
			log.Printf("[ERROR] Failed to disable Batch Job Queue %s: %s", *name, err)
			continue
		}

		log.Printf("[INFO] Deleting Batch Job Queue: %s", *name)
		err = deleteBatchJobQueue(*name, conn)
		if err != nil {
			log.Printf("[ERROR] Failed to delete Batch Job Queue %s: %s", *name, err)
		}
	}

	return nil
}

func TestAccAWSBatchJobQueue_basic(t *testing.T) {
	var jobQueue1 batch.JobQueueDetail
	resourceName := "aws_batch_job_queue.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchJobQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBatchJobQueueConfigState(rName, batch.JQStateEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobQueueExists(resourceName, &jobQueue1),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("job-queue/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environments.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "priority", "1"),
					resource.TestCheckResourceAttr(resourceName, "state", batch.JQStateEnabled),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAWSBatchJobQueue_disappears(t *testing.T) {
	var jobQueue1 batch.JobQueueDetail
	resourceName := "aws_batch_job_queue.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBatchJobQueueConfigState(rName, batch.JQStateEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobQueueExists(resourceName, &jobQueue1),
					testAccCheckBatchJobQueueDisappears(&jobQueue1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/8083
func TestAccAWSBatchJobQueue_ComputeEnvironments_ExternalOrderUpdate(t *testing.T) {
	var jobQueue1 batch.JobQueueDetail
	resourceName := "aws_batch_job_queue.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchJobQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBatchJobQueueConfigState(rName, batch.JQStateEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobQueueExists(resourceName, &jobQueue1),
					testAccCheckBatchJobQueueComputeEnvironmentOrderUpdate(&jobQueue1),
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

func TestAccAWSBatchJobQueue_Priority(t *testing.T) {
	var jobQueue1, jobQueue2 batch.JobQueueDetail
	resourceName := "aws_batch_job_queue.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchJobQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBatchJobQueueConfigPriority(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobQueueExists(resourceName, &jobQueue1),
					resource.TestCheckResourceAttr(resourceName, "priority", "1"),
				),
			},
			{
				Config: testAccBatchJobQueueConfigPriority(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobQueueExists(resourceName, &jobQueue2),
					resource.TestCheckResourceAttr(resourceName, "priority", "2"),
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

func TestAccAWSBatchJobQueue_State(t *testing.T) {
	var jobQueue1, jobQueue2 batch.JobQueueDetail
	resourceName := "aws_batch_job_queue.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchJobQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBatchJobQueueConfigState(rName, batch.JQStateDisabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobQueueExists(resourceName, &jobQueue1),
					resource.TestCheckResourceAttr(resourceName, "state", batch.JQStateDisabled),
				),
			},
			{
				Config: testAccBatchJobQueueConfigState(rName, batch.JQStateEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobQueueExists(resourceName, &jobQueue2),
					resource.TestCheckResourceAttr(resourceName, "state", batch.JQStateEnabled),
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

func TestAccAWSBatchJobQueue_Tags(t *testing.T) {
	var jobQueue batch.JobQueueDetail
	resourceName := "aws_batch_job_queue.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchJobQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBatchJobQueueConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobQueueExists(resourceName, &jobQueue),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBatchJobQueueConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobQueueExists(resourceName, &jobQueue),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccBatchJobQueueConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobQueueExists(resourceName, &jobQueue),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckBatchJobQueueExists(n string, jq *batch.JobQueueDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		log.Printf("State: %#v", s.RootModule().Resources)
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Batch Job Queue ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).batchconn
		name := rs.Primary.Attributes["name"]
		queue, err := getJobQueue(conn, name)
		if err != nil {
			return err
		}
		if queue == nil {
			return fmt.Errorf("Not found: %s", n)
		}
		*jq = *queue

		return nil
	}
}

func testAccCheckBatchJobQueueDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_batch_job_queue" {
			continue
		}
		conn := testAccProvider.Meta().(*AWSClient).batchconn
		jq, err := getJobQueue(conn, rs.Primary.Attributes["name"])
		if err == nil {
			if jq != nil {
				return fmt.Errorf("Error: Job Queue still exists")
			}
		}
		return nil
	}
	return nil
}

func testAccCheckBatchJobQueueDisappears(jobQueue *batch.JobQueueDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).batchconn
		name := aws.StringValue(jobQueue.JobQueueName)

		err := disableBatchJobQueue(name, conn)
		if err != nil {
			return fmt.Errorf("error disabling Batch Job Queue (%s): %s", name, err)
		}

		return deleteBatchJobQueue(name, conn)
	}
}

// testAccCheckBatchJobQueueComputeEnvironmentOrderUpdate simulates the change of a Compute Environment Order
// An external update to the Batch Job Queue (e.g. console) may trigger changes to the value of the Order
// parameter that do not affect the operation of the queue itself, but the resource logic needs to handle.
// For example, Terraform may set a single Compute Environment with Order 0, but the console updates it to 1.
func testAccCheckBatchJobQueueComputeEnvironmentOrderUpdate(jobQueue *batch.JobQueueDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).batchconn

		input := &batch.UpdateJobQueueInput{
			ComputeEnvironmentOrder: jobQueue.ComputeEnvironmentOrder,
			JobQueue:                jobQueue.JobQueueName,
		}
		name := aws.StringValue(jobQueue.JobQueueName)

		if len(input.ComputeEnvironmentOrder) != 1 {
			return fmt.Errorf("expected one ComputeEnvironmentOrder in Batch Job Queue (%s)", name)
		}

		if aws.Int64Value(input.ComputeEnvironmentOrder[0].Order) != 0 {
			return fmt.Errorf("expected first ComputeEnvironmentOrder in Batch Job Queue (%s) to have existing Order value of 0", name)
		}

		input.ComputeEnvironmentOrder[0].Order = aws.Int64(1)

		_, err := conn.UpdateJobQueue(input)

		if err != nil {
			return fmt.Errorf("error updating Batch Job Queue (%s): %s", name, err)
		}

		return nil
	}
}

func testAccBatchJobQueueConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name               = %[1]q
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

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSBatchServiceRole"
}

resource "aws_iam_role" "ecs_instance_role" {
  name = "%[1]s-ecs"

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
  role       = aws_iam_role.ecs_instance_role.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

resource "aws_iam_instance_profile" "ecs_instance_role" {
  name = aws_iam_role.ecs_instance_role.name
  role = aws_iam_role_policy_attachment.ecs_instance_role.role
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-batch-job-queue"
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = "tf-acc-batch-job-queue"
  }
}

resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q
  service_role             = aws_iam_role.test.arn
  type                     = "MANAGED"

  compute_resources {
    instance_role      = aws_iam_instance_profile.ecs_instance_role.arn
    instance_type      = ["c5", "m5", "r5"]
    max_vcpus          = 1
    min_vcpus          = 0
    security_group_ids = [aws_security_group.test.id]
    subnets            = [aws_subnet.test.id]
    type               = "EC2"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccBatchJobQueueConfigPriority(rName string, priority int) string {
	return composeConfig(
		testAccBatchJobQueueConfigBase(rName),
		fmt.Sprintf(`
resource "aws_batch_job_queue" "test" {
  compute_environments = [aws_batch_compute_environment.test.arn]
  name                 = %[1]q
  priority             = %[2]d
  state                = "ENABLED"
}
`, rName, priority))
}

func testAccBatchJobQueueConfigState(rName string, state string) string {
	return composeConfig(
		testAccBatchJobQueueConfigBase(rName),
		fmt.Sprintf(`
resource "aws_batch_job_queue" "test" {
  compute_environments = [aws_batch_compute_environment.test.arn]
  name                 = %[1]q
  priority             = 1
  state                = %[2]q
}
`, rName, state))
}

func testAccBatchJobQueueConfigTags1(rName, tagKey1, tagValue1 string) string {
	return composeConfig(
		testAccBatchJobQueueConfigBase(rName),
		fmt.Sprintf(`
resource "aws_batch_job_queue" "test" {
  compute_environments = [aws_batch_compute_environment.test.arn]
  name                 = %[1]q
  priority             = 1
  state                = "DISABLED"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccBatchJobQueueConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return composeConfig(
		testAccBatchJobQueueConfigBase(rName),
		fmt.Sprintf(`
resource "aws_batch_job_queue" "test" {
  compute_environments = [aws_batch_compute_environment.test.arn]
  name                 = %[1]q
  priority             = 1
  state                = "DISABLED"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
