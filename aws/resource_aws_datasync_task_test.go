package aws

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/datasync/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
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
		return fmt.Errorf("error getting client: %w", err)
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
			return fmt.Errorf("Error retrieving DataSync Tasks: %w", err)
		}

		if len(output.Tasks) == 0 {
			log.Print("[DEBUG] No DataSync Tasks to sweep")
			return nil
		}

		for _, task := range output.Tasks {
			name := aws.StringValue(task.Name)

			log.Printf("[INFO] Deleting DataSync Task: %s", name)
			input := &datasync.DeleteTaskInput{
				TaskArn: task.TaskArn,
			}

			_, err := conn.DeleteTask(input)

			if isAWSErr(err, datasync.ErrCodeInvalidRequestException, "not found") {
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
		ErrorCheck:   testAccErrorCheck(t, datasync.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncTaskConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task1),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "datasync", regexp.MustCompile(`task/task-.+`)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_log_group_arn", ""),
					resource.TestCheckResourceAttrPair(resourceName, "destination_location_arn", dataSyncDestinationLocationResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.atime", "BEST_EFFORT"),
					resource.TestCheckResourceAttr(resourceName, "options.0.bytes_per_second", "-1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.gid", "INT_VALUE"),
					resource.TestCheckResourceAttr(resourceName, "options.0.log_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "options.0.mtime", "PRESERVE"),
					resource.TestCheckResourceAttr(resourceName, "options.0.overwrite_mode", "ALWAYS"),
					resource.TestCheckResourceAttr(resourceName, "options.0.posix_permissions", "PRESERVE"),
					resource.TestCheckResourceAttr(resourceName, "options.0.preserve_deleted_files", "PRESERVE"),
					resource.TestCheckResourceAttr(resourceName, "options.0.preserve_devices", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "options.0.task_queueing", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "options.0.transfer_mode", "CHANGED"),
					resource.TestCheckResourceAttr(resourceName, "options.0.uid", "INT_VALUE"),
					resource.TestCheckResourceAttr(resourceName, "options.0.verify_mode", "POINT_IN_TIME_CONSISTENT"),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "source_location_arn", dataSyncSourceLocationResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", "0"),
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
		ErrorCheck:   testAccErrorCheck(t, datasync.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncTaskConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task1),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsDataSyncTask(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSDataSyncTask_schedule(t *testing.T) {
	var task1 datasync.DescribeTaskOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		ErrorCheck:   testAccErrorCheck(t, datasync.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncTaskScheduleConfig(rName, "cron(0 12 ? * SUN,WED *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_expression", "cron(0 12 ? * SUN,WED *)"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDataSyncTaskScheduleConfig(rName, "cron(0 12 ? * SUN,MON *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_expression", "cron(0 12 ? * SUN,MON *)"),
				),
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
		ErrorCheck:   testAccErrorCheck(t, datasync.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncTaskConfigCloudWatchLogGroupArn(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_log_group_arn", "aws_cloudwatch_log_group.test1", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDataSyncTaskConfigCloudWatchLogGroupArn2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_log_group_arn", "aws_cloudwatch_log_group.test2", "arn")),
			},
		},
	})
}

func TestAccAWSDataSyncTask_Excludes(t *testing.T) {
	var task1 datasync.DescribeTaskOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		ErrorCheck:   testAccErrorCheck(t, datasync.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncTaskExcludesConfig(rName, "/folder1|/folder2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "excludes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "excludes.0.filter_type", "SIMPLE_PATTERN"),
					resource.TestCheckResourceAttr(resourceName, "excludes.0.value", "/folder1|/folder2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDataSyncTaskExcludesConfig(rName, "/test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "excludes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "excludes.0.filter_type", "SIMPLE_PATTERN"),
					resource.TestCheckResourceAttr(resourceName, "excludes.0.value", "/test"),
				),
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
		ErrorCheck:   testAccErrorCheck(t, datasync.EndpointsID),
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
		ErrorCheck:   testAccErrorCheck(t, datasync.EndpointsID),
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
		ErrorCheck:   testAccErrorCheck(t, datasync.EndpointsID),
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

func TestAccAWSDataSyncTask_DefaultSyncOptions_LogLevel(t *testing.T) {
	var task1, task2 datasync.DescribeTaskOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		ErrorCheck:   testAccErrorCheck(t, datasync.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncTaskConfigDefaultSyncOptionsLogLevel(rName, "OFF"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.log_level", "OFF"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDataSyncTaskConfigDefaultSyncOptionsLogLevel(rName, "BASIC"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task2),
					testAccCheckAWSDataSyncTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.log_level", "BASIC"),
				),
			},
		},
	})
}

func TestAccAWSDataSyncTask_DefaultSyncOptions_OverwriteMode(t *testing.T) {
	var task1, task2 datasync.DescribeTaskOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		ErrorCheck:   testAccErrorCheck(t, datasync.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncTaskConfigDefaultSyncOptionsOverwriteMode(rName, "NEVER"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.overwrite_mode", "NEVER"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDataSyncTaskConfigDefaultSyncOptionsOverwriteMode(rName, "ALWAYS"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task2),
					testAccCheckAWSDataSyncTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.overwrite_mode", "ALWAYS"),
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
		ErrorCheck:   testAccErrorCheck(t, datasync.EndpointsID),
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
		ErrorCheck:   testAccErrorCheck(t, datasync.EndpointsID),
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
		ErrorCheck:   testAccErrorCheck(t, datasync.EndpointsID),
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

func TestAccAWSDataSyncTask_DefaultSyncOptions_TaskQueueing(t *testing.T) {
	var task1, task2 datasync.DescribeTaskOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		ErrorCheck:   testAccErrorCheck(t, datasync.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncTaskConfigDefaultSyncOptionsTaskQueueing(rName, "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.task_queueing", "ENABLED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDataSyncTaskConfigDefaultSyncOptionsTaskQueueing(rName, "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task2),
					testAccCheckAWSDataSyncTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.task_queueing", "DISABLED"),
				),
			},
		},
	})
}

func TestAccAWSDataSyncTask_DefaultSyncOptions_TransferMode(t *testing.T) {
	var task1, task2 datasync.DescribeTaskOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		ErrorCheck:   testAccErrorCheck(t, datasync.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncTaskConfigDefaultSyncOptionsTransferMode(rName, "CHANGED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.transfer_mode", "CHANGED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDataSyncTaskConfigDefaultSyncOptionsTransferMode(rName, "ALL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task2),
					testAccCheckAWSDataSyncTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.transfer_mode", "ALL"),
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
		ErrorCheck:   testAccErrorCheck(t, datasync.EndpointsID),
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
	var task1, task2, task3 datasync.DescribeTaskOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		ErrorCheck:   testAccErrorCheck(t, datasync.EndpointsID),
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
			{
				Config: testAccAWSDataSyncTaskConfigDefaultSyncOptionsVerifyMode(rName, "ONLY_FILES_TRANSFERRED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncTaskExists(resourceName, &task3),
					testAccCheckAWSDataSyncTaskNotRecreated(&task2, &task3),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.verify_mode", "ONLY_FILES_TRANSFERRED"),
				),
			},
		},
	})
}

func TestAccAWSDataSyncTask_Tags(t *testing.T) {
	TestAccSkip(t, "Tagging on creation is inconsistent")
	var task1, task2, task3 datasync.DescribeTaskOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		ErrorCheck:   testAccErrorCheck(t, datasync.EndpointsID),
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

		_, err := finder.TaskByARN(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("DataSync Task %s still exists", rs.Primary.ID)
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

		output, err := finder.TaskByARN(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if aws.StringValue(output.Status) != datasync.TaskStatusAvailable && aws.StringValue(output.Status) != datasync.TaskStatusRunning {
			return fmt.Errorf("Task %q not available or running: last status (%s), error code (%s), error detail: %s",
				rs.Primary.ID, aws.StringValue(output.Status), aws.StringValue(output.ErrorCode), aws.StringValue(output.ErrorDetail))
		}

		*task = *output

		return nil
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
  name = "%[1]s-destination"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": "datasync.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }]
}
POLICY
}

resource "aws_s3_bucket" "destination" {
  bucket        = "%[1]s-destination"
  force_destroy = true
}

resource "aws_iam_role_policy" "destination" {
  role   = aws_iam_role.destination.id
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": [
      "s3:*"
    ],
    "Effect": "Allow",
    "Resource": [
      "${aws_s3_bucket.destination.arn}",
      "${aws_s3_bucket.destination.arn}/*"
    ]
  }]
}
POLICY
}

resource "aws_datasync_location_s3" "destination" {
  s3_bucket_arn = aws_s3_bucket.destination.arn
  subdirectory  = "/destination"

  s3_config {
    bucket_access_role_arn = aws_iam_role.destination.arn
  }

  depends_on = [aws_iam_role_policy.destination]
}
`, rName)
}

func testAccAWSDataSyncTaskConfigSourceLocationNfsBase(rName string) string {
	return composeConfig(
		// Reference: https://docs.aws.amazon.com/datasync/latest/userguide/agent-requirements.html
		testAccAvailableEc2InstanceTypeForAvailabilityZone("aws_subnet.source.availability_zone", "m5.2xlarge", "m5.4xlarge"),
		fmt.Sprintf(`
# Reference: https://docs.aws.amazon.com/datasync/latest/userguide/deploy-agents.html
data "aws_ssm_parameter" "aws_service_datasync_ami" {
  name = "/aws/service/datasync/ami"
}

resource "aws_vpc" "source" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "source" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = aws_vpc.source.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "source" {
  vpc_id = aws_vpc.source.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_route_table" "source" {
  default_route_table_id = aws_vpc.source.default_route_table_id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.source.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "source" {
  name   = %[1]q
  vpc_id = aws_vpc.source.id

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
    Name = %[1]q
  }
}

# EFS as our NFS server
resource "aws_efs_file_system" "source" {}

resource "aws_efs_mount_target" "source" {
  file_system_id  = aws_efs_file_system.source.id
  security_groups = [aws_security_group.source.id]
  subnet_id       = aws_subnet.source.id
}

resource "aws_instance" "source" {
  depends_on = [
    aws_default_route_table.source,
  ]

  ami                         = data.aws_ssm_parameter.aws_service_datasync_ami.value
  associate_public_ip_address = true
  instance_type               = data.aws_ec2_instance_type_offering.available.instance_type
  vpc_security_group_ids      = [aws_security_group.source.id]
  subnet_id                   = aws_subnet.source.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_datasync_agent" "source" {
  ip_address = aws_instance.source.public_ip
  name       = %[1]q
}

resource "aws_datasync_location_nfs" "source" {
  server_hostname = aws_efs_mount_target.source.dns_name
  subdirectory    = "/"

  on_prem_config {
    agent_arns = [aws_datasync_agent.source.arn]
  }
}
`, rName))
}

func testAccAWSDataSyncTaskConfig(rName string) string {
	return composeConfig(
		testAccAWSDataSyncTaskConfigDestinationLocationS3Base(rName),
		testAccAWSDataSyncTaskConfigSourceLocationNfsBase(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn
}
`, rName))
}

func testAccAWSDataSyncTaskScheduleConfig(rName, cron string) string {
	return composeConfig(
		testAccAWSDataSyncTaskConfigDestinationLocationS3Base(rName),
		testAccAWSDataSyncTaskConfigSourceLocationNfsBase(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn

  schedule {
    schedule_expression = %[2]q
  }
}
`, rName, cron))
}

func testAccAWSDataSyncTaskConfigCloudWatchLogGroupArn(rName string) string {
	return composeConfig(
		testAccAWSDataSyncTaskConfigDestinationLocationS3Base(rName),
		testAccAWSDataSyncTaskConfigSourceLocationNfsBase(rName),
		fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test1" {
  name = "%[1]s-1"
}

resource "aws_datasync_task" "test" {
  cloudwatch_log_group_arn = aws_cloudwatch_log_group.test1.arn
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn
}
`, rName))
}

func testAccAWSDataSyncTaskConfigCloudWatchLogGroupArn2(rName string) string {
	return composeConfig(
		testAccAWSDataSyncTaskConfigDestinationLocationS3Base(rName),
		testAccAWSDataSyncTaskConfigSourceLocationNfsBase(rName),
		fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test1" {
  name = "%[1]s-1"
}

resource "aws_cloudwatch_log_group" "test2" {
  name = "%[1]s-2"
}

resource "aws_datasync_task" "test" {
  cloudwatch_log_group_arn = aws_cloudwatch_log_group.test2.arn
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn
}
`, rName))
}

func testAccAWSDataSyncTaskExcludesConfig(rName, value string) string {
	return composeConfig(
		testAccAWSDataSyncTaskConfigDestinationLocationS3Base(rName),
		testAccAWSDataSyncTaskConfigSourceLocationNfsBase(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn

  excludes {
    filter_type = "SIMPLE_PATTERN"
    value       = %[2]q
  }
}
`, rName, value))
}

func testAccAWSDataSyncTaskConfigDefaultSyncOptionsAtimeMtime(rName, atime, mtime string) string {
	return composeConfig(
		testAccAWSDataSyncTaskConfigDestinationLocationS3Base(rName),
		testAccAWSDataSyncTaskConfigSourceLocationNfsBase(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn

  options {
    atime = %[2]q
    mtime = %[3]q
  }
}
`, rName, atime, mtime))
}

func testAccAWSDataSyncTaskConfigDefaultSyncOptionsBytesPerSecond(rName string, bytesPerSecond int) string {
	return composeConfig(
		testAccAWSDataSyncTaskConfigDestinationLocationS3Base(rName),
		testAccAWSDataSyncTaskConfigSourceLocationNfsBase(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn

  options {
    bytes_per_second = %[2]d
  }
}
`, rName, bytesPerSecond))
}

func testAccAWSDataSyncTaskConfigDefaultSyncOptionsGid(rName, gid string) string {
	return composeConfig(
		testAccAWSDataSyncTaskConfigDestinationLocationS3Base(rName),
		testAccAWSDataSyncTaskConfigSourceLocationNfsBase(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn

  options {
    gid = %[2]q
  }
}
`, rName, gid))
}

func testAccAWSDataSyncTaskConfigDefaultSyncOptionsLogLevel(rName, logLevel string) string {
	return composeConfig(
		testAccAWSDataSyncTaskConfigDestinationLocationS3Base(rName),
		testAccAWSDataSyncTaskConfigSourceLocationNfsBase(rName),
		fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_datasync_task" "test" {
  cloudwatch_log_group_arn = aws_cloudwatch_log_group.test.arn
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn

  options {
    log_level = %[2]q
  }
}
`, rName, logLevel))
}

func testAccAWSDataSyncTaskConfigDefaultSyncOptionsOverwriteMode(rName, overwriteMode string) string {
	return composeConfig(
		testAccAWSDataSyncTaskConfigDestinationLocationS3Base(rName),
		testAccAWSDataSyncTaskConfigSourceLocationNfsBase(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn

  options {
    overwrite_mode = %[2]q
  }
}
`, rName, overwriteMode))
}

func testAccAWSDataSyncTaskConfigDefaultSyncOptionsPosixPermissions(rName, posixPermissions string) string {
	return composeConfig(
		testAccAWSDataSyncTaskConfigDestinationLocationS3Base(rName),
		testAccAWSDataSyncTaskConfigSourceLocationNfsBase(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn

  options {
    posix_permissions = %[2]q
  }
}
`, rName, posixPermissions))
}

func testAccAWSDataSyncTaskConfigDefaultSyncOptionsPreserveDeletedFiles(rName, preserveDeletedFiles string) string {
	return composeConfig(
		testAccAWSDataSyncTaskConfigDestinationLocationS3Base(rName),
		testAccAWSDataSyncTaskConfigSourceLocationNfsBase(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn

  options {
    preserve_deleted_files = %[2]q
  }
}
`, rName, preserveDeletedFiles))
}

func testAccAWSDataSyncTaskConfigDefaultSyncOptionsPreserveDevices(rName, preserveDevices string) string {
	return composeConfig(
		testAccAWSDataSyncTaskConfigDestinationLocationS3Base(rName),
		testAccAWSDataSyncTaskConfigSourceLocationNfsBase(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn

  options {
    preserve_devices = %[2]q
  }
}
`, rName, preserveDevices))
}

func testAccAWSDataSyncTaskConfigDefaultSyncOptionsTaskQueueing(rName, taskQueueing string) string {
	return composeConfig(
		testAccAWSDataSyncTaskConfigDestinationLocationS3Base(rName),
		testAccAWSDataSyncTaskConfigSourceLocationNfsBase(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn

  options {
    task_queueing = %[2]q
  }
}
`, rName, taskQueueing))
}

func testAccAWSDataSyncTaskConfigDefaultSyncOptionsTransferMode(rName, transferMode string) string {
	return composeConfig(
		testAccAWSDataSyncTaskConfigDestinationLocationS3Base(rName),
		testAccAWSDataSyncTaskConfigSourceLocationNfsBase(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn

  options {
    transfer_mode = %[2]q
  }
}
`, rName, transferMode))
}

func testAccAWSDataSyncTaskConfigDefaultSyncOptionsUid(rName, uid string) string {
	return composeConfig(
		testAccAWSDataSyncTaskConfigDestinationLocationS3Base(rName),
		testAccAWSDataSyncTaskConfigSourceLocationNfsBase(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn

  options {
    uid = %[2]q
  }
}
`, rName, uid))
}

func testAccAWSDataSyncTaskConfigDefaultSyncOptionsVerifyMode(rName, verifyMode string) string {
	return composeConfig(
		testAccAWSDataSyncTaskConfigDestinationLocationS3Base(rName),
		testAccAWSDataSyncTaskConfigSourceLocationNfsBase(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn

  options {
    verify_mode = %[2]q
  }
}
`, rName, verifyMode))
}

func testAccAWSDataSyncTaskConfigTags1(rName, key1, value1 string) string {
	return composeConfig(
		testAccAWSDataSyncTaskConfigDestinationLocationS3Base(rName),
		testAccAWSDataSyncTaskConfigSourceLocationNfsBase(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, key1, value1))
}

func testAccAWSDataSyncTaskConfigTags2(rName, key1, value1, key2, value2 string) string {
	return composeConfig(
		testAccAWSDataSyncTaskConfigDestinationLocationS3Base(rName),
		testAccAWSDataSyncTaskConfigSourceLocationNfsBase(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, key1, value1, key2, value2))
}
