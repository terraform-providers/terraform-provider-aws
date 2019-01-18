package aws

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
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

	prefixes := []string{
		"tf_acc",
	}

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
		skip := true
		for _, prefix := range prefixes {
			if strings.HasPrefix(*name, prefix) {
				skip = false
				break
			}
		}
		if skip {
			log.Printf("[INFO] Skipping Batch Job Queue: %s", *name)
			continue
		}

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
	var jq batch.JobQueueDetail
	ri := acctest.RandInt()
	config := fmt.Sprintf(testAccBatchJobQueueBasic, ri)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchJobQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobQueueExists("aws_batch_job_queue.test_queue", &jq),
					testAccCheckBatchJobQueueAttributes(&jq),
				),
			},
		},
	})
}

func TestAccAWSBatchJobQueue_disappears(t *testing.T) {
	var jobQueue1 batch.JobQueueDetail
	resourceName := "aws_batch_job_queue.test_queue"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccBatchJobQueueBasic, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobQueueExists(resourceName, &jobQueue1),
					testAccCheckBatchJobQueueDisappears(&jobQueue1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSBatchJobQueue_update(t *testing.T) {
	var jq batch.JobQueueDetail
	ri := acctest.RandInt()
	config := fmt.Sprintf(testAccBatchJobQueueBasic, ri)
	updateConfig := fmt.Sprintf(testAccBatchJobQueueUpdate, ri)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchJobQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobQueueExists("aws_batch_job_queue.test_queue", &jq),
					testAccCheckBatchJobQueueAttributes(&jq),
				),
			},
			{
				Config: updateConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobQueueExists("aws_batch_job_queue.test_queue", &jq),
					testAccCheckBatchJobQueueAttributes(&jq),
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

func testAccCheckBatchJobQueueAttributes(jq *batch.JobQueueDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !strings.HasPrefix(*jq.JobQueueName, "tf_acctest_batch_job_queue") {
			return fmt.Errorf("Bad Job Queue name: %s", *jq.JobQueueName)
		}
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_batch_job_queue" {
				continue
			}
			if *jq.JobQueueArn != rs.Primary.Attributes["arn"] {
				return fmt.Errorf("Bad Job Queue ARN\n\t expected: %s\n\tgot: %s\n", rs.Primary.Attributes["arn"], *jq.JobQueueArn)
			}
			if *jq.State != rs.Primary.Attributes["state"] {
				return fmt.Errorf("Bad Job Queue State\n\t expected: %s\n\tgot: %s\n", rs.Primary.Attributes["state"], *jq.State)
			}
			priority, err := strconv.ParseInt(rs.Primary.Attributes["priority"], 10, 64)
			if err != nil {
				return err
			}
			if *jq.Priority != priority {
				return fmt.Errorf("Bad Job Queue Priority\n\t expected: %s\n\tgot: %d\n", rs.Primary.Attributes["priority"], *jq.Priority)
			}
		}
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

const testAccBatchJobQueueBaseConfig = `
########## ecs_instance_role ##########

resource "aws_iam_role" "ecs_instance_role" {
  name = "ecs_instance_role_%[1]d"
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
  name  = "ecs_instance_role_%[1]d"
  role = "${aws_iam_role.ecs_instance_role.name}"
}

########## aws_batch_service_role ##########

resource "aws_iam_role" "aws_batch_service_role" {
  name = "aws_batch_service_role_%[1]d"
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

########## security group ##########

resource "aws_security_group" "test_acc" {
  name = "aws_batch_compute_environment_security_group_%[1]d"
}

########## subnets ##########

resource "aws_vpc" "test_acc" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-batch-job-queue"
  }
}

resource "aws_subnet" "test_acc" {
  vpc_id = "${aws_vpc.test_acc.id}"
  cidr_block = "10.1.1.0/24"
  tags = {
    Name = "tf-acc-batch-job-queue"
  }
}

resource "aws_batch_compute_environment" "test_environment" {
  compute_environment_name = "tf_acctest_batch_compute_environment_%[1]d"
  compute_resources {
    instance_role = "${aws_iam_role.aws_batch_service_role.arn}"
    instance_type = ["m3.medium"]
    max_vcpus = 1
    min_vcpus = 0
    security_group_ids = ["${aws_security_group.test_acc.id}"]
    subnets = ["${aws_subnet.test_acc.id}"]
    type = "EC2"
  }
  service_role = "${aws_iam_role.aws_batch_service_role.arn}"
  type = "MANAGED"
  depends_on = ["aws_iam_role_policy_attachment.aws_batch_service_role"]
}`

var testAccBatchJobQueueBasic = testAccBatchJobQueueBaseConfig + `
resource "aws_batch_job_queue" "test_queue" {
  name = "tf_acctest_batch_job_queue_%[1]d"
  state = "ENABLED"
  priority = 1
  compute_environments = ["${aws_batch_compute_environment.test_environment.arn}"]
}`

var testAccBatchJobQueueUpdate = testAccBatchJobQueueBaseConfig + `
resource "aws_batch_job_queue" "test_queue" {
  name = "tf_acctest_batch_job_queue_%[1]d"
  state = "DISABLED"
  priority = 2
  compute_environments = ["${aws_batch_compute_environment.test_environment.arn}"]
}`
