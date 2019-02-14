package aws

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_datasync_task", &resource.Sweeper{
		Name: "aws_datasync_task",
		F:    testSweepDataSyncTasks,
	})
}

func testSweepDataSyncTasks(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).datasyncconn

	input := &datasync.ListTasksInput{}
	for {
		output, err := conn.ListTasks(input)

		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping DataSync Task sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving DataSync Tasks: %s", err)
		}

		if len(output.Tasks) == 0 {
			log.Print("[DEBUG] No DataSync Tasks to sweep")
			return nil
		}

		for _, task := range output.Tasks {
			name := aws.StringValue(task.Name)
			if !strings.HasPrefix(name, "tf-acc-test") {
				log.Printf("[INFO] Skipping DataSync Task: %s", name)
				continue
			}
			log.Printf("[INFO] Deleting DataSync Task: %s", name)
			input := &datasync.DeleteTaskInput{
				TaskArn: task.TaskArn,
			}

			_, err := conn.DeleteTask(input)

			if isAWSErr(err, "InvalidRequestException", "not found") {
				continue
			}

			if err != nil {
				log.Printf("[ERROR] Failed to delete DataSync Task (%s): %s", name, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}

func TestAccAWSDataSyncTask_basic(t *testing.T) {
	var task1 datasync.DescribeTaskOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSyncDestinationLocationResourceName := "aws_datasync_location_s3.destination"
	dataSyncSourceLocationResourceName := "aws_datasync_location_nfs.source"
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncTaskConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task1),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "datasync", regexp.MustCompile(`task/task-.+`)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_log_group_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.atime", "BEST_EFFORT"),
					resource.TestCheckResourceAttr(resourceName, "options.0.bytes_per_second", "-1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.gid", "INT_VALUE"),
					resource.TestCheckResourceAttr(resourceName, "options.0.mtime", "PRESERVE"),
					resource.TestCheckResourceAttr(resourceName, "options.0.posix_permissions", "PRESERVE"),
					resource.TestCheckResourceAttr(resourceName, "options.0.preserve_deleted_files", "PRESERVE"),
					resource.TestCheckResourceAttr(resourceName, "options.0.preserve_devices", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "options.0.uid", "INT_VALUE"),
					resource.TestCheckResourceAttr(resourceName, "options.0.verify_mode", "POINT_IN_TIME_CONSISTENT"),
					resource.TestCheckResourceAttrPair(resourceName, "destination_location_arn", dataSyncDestinationLocationResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "source_location_arn", dataSyncSourceLocationResourceName, "arn"),
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

func TestAccAWSDataSyncTask_disappears(t *testing.T) {
	var task1 datasync.DescribeTaskOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncTaskConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task1),
					testAccCheckAWSDataSyncTaskDisappears(&task1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSDataSyncTask_CloudWatchLogGroupARN(t *testing.T) {
	var task1 datasync.DescribeTaskOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncTaskConfigCloudWatchLogGroupArn(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttrSet(resourceName, "cloudwatch_log_group_arn"),
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

func TestAccAWSDataSyncTask_DefaultSyncOptions_AtimeMtime(t *testing.T) {
	var task1, task2 datasync.DescribeTaskOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncTaskConfigDefaultSyncOptionsAtimeMtime(rName, "NONE", "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.atime", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "options.0.mtime", "NONE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDataSyncTaskConfigDefaultSyncOptionsAtimeMtime(rName, "BEST_EFFORT", "PRESERVE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task2),
					testAccCheckAWSDataSyncTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.atime", "BEST_EFFORT"),
					resource.TestCheckResourceAttr(resourceName, "options.0.mtime", "PRESERVE"),
				),
			},
		},
	})
}

func TestAccAWSDataSyncTask_DefaultSyncOptions_BytesPerSecond(t *testing.T) {
	var task1, task2 datasync.DescribeTaskOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncTaskConfigDefaultSyncOptionsBytesPerSecond(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.bytes_per_second", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDataSyncTaskConfigDefaultSyncOptionsBytesPerSecond(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task2),
					testAccCheckAWSDataSyncTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.bytes_per_second", "2"),
				),
			},
		},
	})
}

func TestAccAWSDataSyncTask_DefaultSyncOptions_Gid(t *testing.T) {
	var task1, task2 datasync.DescribeTaskOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncTaskConfigDefaultSyncOptionsGid(rName, "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.gid", "NONE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDataSyncTaskConfigDefaultSyncOptionsGid(rName, "INT_VALUE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task2),
					testAccCheckAWSDataSyncTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.gid", "INT_VALUE"),
				),
			},
		},
	})
}

func TestAccAWSDataSyncTask_DefaultSyncOptions_PosixPermissions(t *testing.T) {
	var task1, task2 datasync.DescribeTaskOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncTaskConfigDefaultSyncOptionsPosixPermissions(rName, "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.posix_permissions", "NONE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDataSyncTaskConfigDefaultSyncOptionsPosixPermissions(rName, "PRESERVE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task2),
					testAccCheckAWSDataSyncTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.posix_permissions", "PRESERVE"),
				),
			},
		},
	})
}

func TestAccAWSDataSyncTask_DefaultSyncOptions_PreserveDeletedFiles(t *testing.T) {
	var task1, task2 datasync.DescribeTaskOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncTaskConfigDefaultSyncOptionsPreserveDeletedFiles(rName, "REMOVE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.preserve_deleted_files", "REMOVE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDataSyncTaskConfigDefaultSyncOptionsPreserveDeletedFiles(rName, "PRESERVE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task2),
					testAccCheckAWSDataSyncTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.preserve_deleted_files", "PRESERVE"),
				),
			},
		},
	})
}

func TestAccAWSDataSyncTask_DefaultSyncOptions_PreserveDevices(t *testing.T) {
	var task1, task2 datasync.DescribeTaskOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncTaskConfigDefaultSyncOptionsPreserveDevices(rName, "PRESERVE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.preserve_devices", "PRESERVE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDataSyncTaskConfigDefaultSyncOptionsPreserveDevices(rName, "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task2),
					testAccCheckAWSDataSyncTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.preserve_devices", "NONE"),
				),
			},
		},
	})
}

func TestAccAWSDataSyncTask_DefaultSyncOptions_Uid(t *testing.T) {
	var task1, task2 datasync.DescribeTaskOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncTaskConfigDefaultSyncOptionsUid(rName, "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.uid", "NONE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDataSyncTaskConfigDefaultSyncOptionsUid(rName, "INT_VALUE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task2),
					testAccCheckAWSDataSyncTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.uid", "INT_VALUE"),
				),
			},
		},
	})
}

func TestAccAWSDataSyncTask_DefaultSyncOptions_VerifyMode(t *testing.T) {
	var task1, task2 datasync.DescribeTaskOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncTaskConfigDefaultSyncOptionsVerifyMode(rName, "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.verify_mode", "NONE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDataSyncTaskConfigDefaultSyncOptionsVerifyMode(rName, "POINT_IN_TIME_CONSISTENT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task2),
					testAccCheckAWSDataSyncTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.verify_mode", "POINT_IN_TIME_CONSISTENT"),
				),
			},
		},
	})
}

func TestAccAWSDataSyncTask_Tags(t *testing.T) {
	t.Skip("Tagging on creation is inconsistent")
	var task1, task2, task3 datasync.DescribeTaskOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncTaskConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task1),
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
				Config: testAccAWSDataSyncTaskConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task2),
					testAccCheckAWSDataSyncTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSDataSyncTaskConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task3),
					testAccCheckAWSDataSyncTaskNotRecreated(&task2, &task3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func testAccCheckAWSDataSyncTaskDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).datasyncconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_datasync_task" {
			continue
		}

		input := &datasync.DescribeTaskInput{
			TaskArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeTask(input)

		if isAWSErr(err, "InvalidRequestException", "not found") {
			return nil
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckAWSDataSyncTaskExists(resourceName string, task *datasync.DescribeTaskOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).datasyncconn
		input := &datasync.DescribeTaskInput{
			TaskArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeTask(input)

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("Task %q does not exist", rs.Primary.ID)
		}

		if aws.StringValue(output.Status) != datasync.TaskStatusAvailable && aws.StringValue(output.Status) != datasync.TaskStatusRunning {
			return fmt.Errorf("Task %q not available or running: last status (%s), error code (%s), error detail: %s",
				rs.Primary.ID, aws.StringValue(output.Status), aws.StringValue(output.ErrorCode), aws.StringValue(output.ErrorDetail))
		}

		*task = *output

		return nil
	}
}

func testAccCheckAWSDataSyncTaskDisappears(location *datasync.DescribeTaskOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).datasyncconn

		input := &datasync.DeleteTaskInput{
			TaskArn: location.TaskArn,
		}

		_, err := conn.DeleteTask(input)

		return err
	}
}

func testAccCheckAWSDataSyncTaskNotRecreated(i, j *datasync.DescribeTaskOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.TaskArn) != aws.StringValue(j.TaskArn) {
			return errors.New("DataSync Task was recreated")
		}

		return nil
	}
}

func testAccPreCheckAWSDataSync(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).datasyncconn

	input := &datasync.ListTasksInput{
		MaxResults: aws.Int64(1),
	}

	_, err := conn.ListTasks(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSDataSyncTaskConfigDestinationLocationS3Base(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "destination" {
  name = "%sdestination"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "datasync.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_s3_bucket" "destination" {
  bucket        = "%sdestination"
  force_destroy = true
}

resource "aws_datasync_location_s3" "destination" {
  s3_bucket_arn = "${aws_s3_bucket.destination.arn}"
  subdirectory  = "/destination"

  s3_config {
    bucket_access_role_arn = "${aws_iam_role.destination.arn}"
  }
}
`, rName, rName)
}

func testAccAWSDataSyncTaskConfigSourceLocationNfsBase(rName string) string {
	return fmt.Sprintf(`
data "aws_ami" "source-aws-thinstaller" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["aws-thinstaller-*"]
  }
}

resource "aws_vpc" "source" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "tf-acc-test-datasync-task"
  }
}

resource "aws_vpc_dhcp_options" "source" {
  domain_name_servers = ["AmazonProvidedDNS"]
}

resource "aws_vpc_dhcp_options_association" "source" {
  dhcp_options_id = "${aws_vpc_dhcp_options.source.id}"
  vpc_id          = "${aws_vpc.source.id}"
}

resource "aws_subnet" "source" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = "${aws_vpc.source.id}"

  tags = {
    Name = "tf-acc-test-datasync-task"
  }
}

resource "aws_internet_gateway" "source" {
  vpc_id = "${aws_vpc.source.id}"

  tags = {
    Name = "tf-acc-test-datasync-task"
  }
}

resource "aws_route_table" "source" {
  vpc_id = "${aws_vpc.source.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.source.id}"
  }

  tags = {
    Name = "tf-acc-test-datasync-task"
  }
}

resource "aws_route_table_association" "source" {
  subnet_id      = "${aws_subnet.source.id}"
  route_table_id = "${aws_route_table.source.id}"
}

resource "aws_security_group" "source" {
  vpc_id = "${aws_vpc.source.id}"

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "tf-acc-test-datasync-task"
  }
}

# EFS as our NFS server
resource "aws_efs_file_system" "source" {}

resource "aws_efs_mount_target" "source" {
  file_system_id  = "${aws_efs_file_system.source.id}"
  security_groups = ["${aws_security_group.source.id}"]
  subnet_id       = "${aws_subnet.source.id}"
}

resource "aws_instance" "source" {
  depends_on = ["aws_internet_gateway.source", "aws_vpc_dhcp_options_association.source"]

  ami                         = "${data.aws_ami.source-aws-thinstaller.id}"
  associate_public_ip_address = true
  # Default instance type from sync.sh
  instance_type               = "c5.2xlarge"
  vpc_security_group_ids      = ["${aws_security_group.source.id}"]
  subnet_id                   = "${aws_subnet.source.id}"

  tags = {
    Name = "tf-acc-test-datasync-task"
  }
}

resource "aws_datasync_agent" "source" {
  ip_address = "${aws_instance.source.public_ip}"
  name       = %q
}

# Using EFS File System DNS name due to DNS propagation delays
resource "aws_datasync_location_nfs" "source" {
  depends_on = ["aws_efs_mount_target.source"]

  server_hostname = "${aws_efs_file_system.source.dns_name}"
  subdirectory    = "/"

  on_prem_config {
    agent_arns = ["${aws_datasync_agent.source.arn}"]
  }
}
`, rName)
}

func testAccAWSDataSyncTaskConfig(rName string) string {
	return testAccAWSDataSyncTaskConfigDestinationLocationS3Base(rName) + testAccAWSDataSyncTaskConfigSourceLocationNfsBase(rName) + fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = "${aws_datasync_location_s3.destination.arn}"
  name                     = %q
  source_location_arn      = "${aws_datasync_location_nfs.source.arn}"
}
`, rName)
}

func testAccAWSDataSyncTaskConfigCloudWatchLogGroupArn(rName string) string {
	return testAccAWSDataSyncTaskConfigDestinationLocationS3Base(rName) + testAccAWSDataSyncTaskConfigSourceLocationNfsBase(rName) + fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %q
}

resource "aws_datasync_task" "test" {
  cloudwatch_log_group_arn = "${replace(aws_cloudwatch_log_group.test.arn, ":*", "")}"
  destination_location_arn = "${aws_datasync_location_s3.destination.arn}"
  name                     = %q
  source_location_arn      = "${aws_datasync_location_nfs.source.arn}"
}
`, rName, rName)
}

func testAccAWSDataSyncTaskConfigDefaultSyncOptionsAtimeMtime(rName, atime, mtime string) string {
	return testAccAWSDataSyncTaskConfigDestinationLocationS3Base(rName) + testAccAWSDataSyncTaskConfigSourceLocationNfsBase(rName) + fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = "${aws_datasync_location_s3.destination.arn}"
  name                     = %q
  source_location_arn      = "${aws_datasync_location_nfs.source.arn}"

  options {
    atime = %q
    mtime = %q
  }
}
`, rName, atime, mtime)
}

func testAccAWSDataSyncTaskConfigDefaultSyncOptionsBytesPerSecond(rName string, bytesPerSecond int) string {
	return testAccAWSDataSyncTaskConfigDestinationLocationS3Base(rName) + testAccAWSDataSyncTaskConfigSourceLocationNfsBase(rName) + fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = "${aws_datasync_location_s3.destination.arn}"
  name                     = %q
  source_location_arn      = "${aws_datasync_location_nfs.source.arn}"

  options {
    bytes_per_second = %d
  }
}
`, rName, bytesPerSecond)
}

func testAccAWSDataSyncTaskConfigDefaultSyncOptionsGid(rName, gid string) string {
	return testAccAWSDataSyncTaskConfigDestinationLocationS3Base(rName) + testAccAWSDataSyncTaskConfigSourceLocationNfsBase(rName) + fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = "${aws_datasync_location_s3.destination.arn}"
  name                     = %q
  source_location_arn      = "${aws_datasync_location_nfs.source.arn}"

  options {
    gid = %q
  }
}
`, rName, gid)
}

func testAccAWSDataSyncTaskConfigDefaultSyncOptionsPosixPermissions(rName, posixPermissions string) string {
	return testAccAWSDataSyncTaskConfigDestinationLocationS3Base(rName) + testAccAWSDataSyncTaskConfigSourceLocationNfsBase(rName) + fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = "${aws_datasync_location_s3.destination.arn}"
  name                     = %q
  source_location_arn      = "${aws_datasync_location_nfs.source.arn}"

  options {
    posix_permissions = %q
  }
}
`, rName, posixPermissions)
}

func testAccAWSDataSyncTaskConfigDefaultSyncOptionsPreserveDeletedFiles(rName, preserveDeletedFiles string) string {
	return testAccAWSDataSyncTaskConfigDestinationLocationS3Base(rName) + testAccAWSDataSyncTaskConfigSourceLocationNfsBase(rName) + fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = "${aws_datasync_location_s3.destination.arn}"
  name                     = %q
  source_location_arn      = "${aws_datasync_location_nfs.source.arn}"

  options {
    preserve_deleted_files = %q
  }
}
`, rName, preserveDeletedFiles)
}

func testAccAWSDataSyncTaskConfigDefaultSyncOptionsPreserveDevices(rName, preserveDevices string) string {
	return testAccAWSDataSyncTaskConfigDestinationLocationS3Base(rName) + testAccAWSDataSyncTaskConfigSourceLocationNfsBase(rName) + fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = "${aws_datasync_location_s3.destination.arn}"
  name                     = %q
  source_location_arn      = "${aws_datasync_location_nfs.source.arn}"

  options {
    preserve_devices = %q
  }
}
`, rName, preserveDevices)
}

func testAccAWSDataSyncTaskConfigDefaultSyncOptionsUid(rName, uid string) string {
	return testAccAWSDataSyncTaskConfigDestinationLocationS3Base(rName) + testAccAWSDataSyncTaskConfigSourceLocationNfsBase(rName) + fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = "${aws_datasync_location_s3.destination.arn}"
  name                     = %q
  source_location_arn      = "${aws_datasync_location_nfs.source.arn}"

  options {
    uid = %q
  }
}
`, rName, uid)
}

func testAccAWSDataSyncTaskConfigDefaultSyncOptionsVerifyMode(rName, verifyMode string) string {
	return testAccAWSDataSyncTaskConfigDestinationLocationS3Base(rName) + testAccAWSDataSyncTaskConfigSourceLocationNfsBase(rName) + fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = "${aws_datasync_location_s3.destination.arn}"
  name                     = %q
  source_location_arn      = "${aws_datasync_location_nfs.source.arn}"

  options {
    verify_mode = %q
  }
}
`, rName, verifyMode)
}

func testAccAWSDataSyncTaskConfigTags1(rName, key1, value1 string) string {
	return testAccAWSDataSyncTaskConfigDestinationLocationS3Base(rName) + testAccAWSDataSyncTaskConfigSourceLocationNfsBase(rName) + fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = "${aws_datasync_location_s3.destination.arn}"
  name                     = %q
  source_location_arn      = "${aws_datasync_location_nfs.source.arn}"

  tags = {
    %q = %q
  }
}
`, rName, key1, value1)
}

func testAccAWSDataSyncTaskConfigTags2(rName, key1, value1, key2, value2 string) string {
	return testAccAWSDataSyncTaskConfigDestinationLocationS3Base(rName) + testAccAWSDataSyncTaskConfigSourceLocationNfsBase(rName) + fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = "${aws_datasync_location_s3.destination.arn}"
  name                     = %q
  source_location_arn      = "${aws_datasync_location_nfs.source.arn}"

  tags = {
    %q = %q
    %q = %q
  }
}
`, rName, key1, value1, key2, value2)
}
