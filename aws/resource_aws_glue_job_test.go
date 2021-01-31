package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_glue_job", &resource.Sweeper{
		Name: "aws_glue_job",
		F:    testSweepGlueJobs,
	})
}

func testSweepGlueJobs(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).glueconn

	input := &glue.GetJobsInput{}
	err = conn.GetJobsPages(input, func(page *glue.GetJobsOutput, lastPage bool) bool {
		if len(page.Jobs) == 0 {
			log.Printf("[INFO] No Glue Jobs to sweep")
			return false
		}
		for _, job := range page.Jobs {
			name := aws.StringValue(job.Name)

			log.Printf("[INFO] Deleting Glue Job: %s", name)
			err := deleteGlueJob(conn, name)
			if err != nil {
				log.Printf("[ERROR] Failed to delete Glue Job %s: %s", name, err)
			}
		}
		return !lastPage
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Glue Job sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Glue Jobs: %s", err)
	}

	return nil
}

func TestAccAWSGlueJob_basic(t *testing.T) {
	var job glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	resourceName := "aws_glue_job.test"
	roleResourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueJobConfig_Required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("job/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "command.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "command.0.script_location", "testscriptlocation"),
					resource.TestCheckResourceAttr(resourceName, "default_arguments.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "non_overridable_arguments.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", roleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout", "2880"),
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

func TestAccAWSGlueJob_Command(t *testing.T) {
	var job glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueJobConfig_Command(rName, "testscriptlocation1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "command.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "command.0.script_location", "testscriptlocation1"),
				),
			},
			{
				Config: testAccAWSGlueJobConfig_Command(rName, "testscriptlocation2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "command.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "command.0.script_location", "testscriptlocation2"),
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

func TestAccAWSGlueJob_DefaultArguments(t *testing.T) {
	var job glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueJobConfig_DefaultArguments(rName, "job-bookmark-disable", "python"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "default_arguments.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "default_arguments.--job-bookmark-option", "job-bookmark-disable"),
					resource.TestCheckResourceAttr(resourceName, "default_arguments.--job-language", "python"),
				),
			},
			{
				Config: testAccAWSGlueJobConfig_DefaultArguments(rName, "job-bookmark-enable", "scala"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "default_arguments.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "default_arguments.--job-bookmark-option", "job-bookmark-enable"),
					resource.TestCheckResourceAttr(resourceName, "default_arguments.--job-language", "scala"),
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

func TestAccAWSGlueJob_nonOverridableArguments(t *testing.T) {
	var job glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueJobnonOverridableArgumentsConfig(rName, "job-bookmark-disable", "python"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "non_overridable_arguments.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "non_overridable_arguments.--job-bookmark-option", "job-bookmark-disable"),
					resource.TestCheckResourceAttr(resourceName, "non_overridable_arguments.--job-language", "python"),
				),
			},
			{
				Config: testAccAWSGlueJobnonOverridableArgumentsConfig(rName, "job-bookmark-enable", "scala"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "non_overridable_arguments.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "non_overridable_arguments.--job-bookmark-option", "job-bookmark-enable"),
					resource.TestCheckResourceAttr(resourceName, "non_overridable_arguments.--job-language", "scala"),
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

func TestAccAWSGlueJob_Description(t *testing.T) {
	var job glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueJobConfig_Description(rName, "First Description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "description", "First Description"),
				),
			},
			{
				Config: testAccAWSGlueJobConfig_Description(rName, "Second Description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "description", "Second Description"),
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

func TestAccAWSGlueJob_GlueVersion(t *testing.T) {
	var job glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueJobConfig_GlueVersion_MaxCapacity(rName, "0.9"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "glue_version", "0.9"),
				),
			},
			{
				Config: testAccAWSGlueJobConfig_GlueVersion_MaxCapacity(rName, "1.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "glue_version", "1.0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSGlueJobConfig_GlueVersion_NumberOfWorkers(rName, "2.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "glue_version", "2.0"),
				),
			},
		},
	})
}

func TestAccAWSGlueJob_ExecutionProperty(t *testing.T) {
	var job glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueJobDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSGlueJobConfig_ExecutionProperty(rName, 0),
				ExpectError: regexp.MustCompile(`expected execution_property.0.max_concurrent_runs to be at least`),
			},
			{
				Config: testAccAWSGlueJobConfig_ExecutionProperty(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "execution_property.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "execution_property.0.max_concurrent_runs", "1"),
				),
			},
			{
				Config: testAccAWSGlueJobConfig_ExecutionProperty(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "execution_property.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "execution_property.0.max_concurrent_runs", "2"),
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

func TestAccAWSGlueJob_MaxRetries(t *testing.T) {
	var job glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueJobDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSGlueJobConfig_MaxRetries(rName, 11),
				ExpectError: regexp.MustCompile(`expected max_retries to be in the range`),
			},
			{
				Config: testAccAWSGlueJobConfig_MaxRetries(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "max_retries", "0"),
				),
			},
			{
				Config: testAccAWSGlueJobConfig_MaxRetries(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "max_retries", "10"),
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

func TestAccAWSGlueJob_NotificationProperty(t *testing.T) {
	var job glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueJobDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSGlueJobConfig_NotificationProperty(rName, 0),
				ExpectError: regexp.MustCompile(`expected notification_property.0.notify_delay_after to be at least`),
			},
			{
				Config: testAccAWSGlueJobConfig_NotificationProperty(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "notification_property.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "notification_property.0.notify_delay_after", "1"),
				),
			},
			{
				Config: testAccAWSGlueJobConfig_NotificationProperty(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "notification_property.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "notification_property.0.notify_delay_after", "2"),
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

func TestAccAWSGlueJob_Tags(t *testing.T) {
	var job1, job2, job3 glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueJobConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job1),
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
				Config: testAccAWSGlueJobConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSGlueJobConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSGlueJob_Timeout(t *testing.T) {
	var job glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueJobConfig_Timeout(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "timeout", "1"),
				),
			},
			{
				Config: testAccAWSGlueJobConfig_Timeout(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "timeout", "2"),
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

func TestAccAWSGlueJob_SecurityConfiguration(t *testing.T) {
	var job glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueJobConfig_SecurityConfiguration(rName, "default_encryption"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "security_configuration", "default_encryption"),
				),
			},
			{
				Config: testAccAWSGlueJobConfig_SecurityConfiguration(rName, "custom_encryption2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "security_configuration", "custom_encryption2"),
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

func TestAccAWSGlueJob_WorkerType(t *testing.T) {
	var job glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueJobConfig_WorkerType(rName, "Standard"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "worker_type", "Standard"),
				),
			},
			{
				Config: testAccAWSGlueJobConfig_WorkerType(rName, "G.1X"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "worker_type", "G.1X"),
				),
			},
			{
				Config: testAccAWSGlueJobConfig_WorkerType(rName, "G.2X"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "worker_type", "G.2X"),
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

func TestAccAWSGlueJob_PythonShell(t *testing.T) {
	var job glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueJobConfig_PythonShell(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "max_capacity", "0.0625"),
					resource.TestCheckResourceAttr(resourceName, "command.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "command.0.script_location", "testscriptlocation"),
					resource.TestCheckResourceAttr(resourceName, "command.0.name", "pythonshell"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSGlueJobConfig_PythonShellWithVersion(rName, "2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "command.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "command.0.script_location", "testscriptlocation"),
					resource.TestCheckResourceAttr(resourceName, "command.0.python_version", "2"),
					resource.TestCheckResourceAttr(resourceName, "command.0.name", "pythonshell"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSGlueJobConfig_PythonShellWithVersion(rName, "3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "command.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "command.0.script_location", "testscriptlocation"),
					resource.TestCheckResourceAttr(resourceName, "command.0.python_version", "3"),
					resource.TestCheckResourceAttr(resourceName, "command.0.name", "pythonshell"),
				),
			},
		},
	})
}

func TestAccAWSGlueJob_MaxCapacity(t *testing.T) {
	var job glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueJobConfig_MaxCapacity(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "max_capacity", "10"),
					resource.TestCheckResourceAttr(resourceName, "command.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "command.0.script_location", "testscriptlocation"),
					resource.TestCheckResourceAttr(resourceName, "command.0.name", "glueetl"),
				),
			},
			{
				Config: testAccAWSGlueJobConfig_MaxCapacity(rName, 15),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "max_capacity", "15"),
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

func TestAccAWSGlueJob_disappears(t *testing.T) {
	var job glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueJobConfig_Required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsGlueJob(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSGlueJobExists(resourceName string, job *glue.Job) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Glue Job ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).glueconn

		output, err := conn.GetJob(&glue.GetJobInput{
			JobName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if output.Job == nil {
			return fmt.Errorf("Glue Job (%s) not found", rs.Primary.ID)
		}

		if aws.StringValue(output.Job.Name) == rs.Primary.ID {
			*job = *output.Job
			return nil
		}

		return fmt.Errorf("Glue Job (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckAWSGlueJobDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_glue_job" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).glueconn

		output, err := conn.GetJob(&glue.GetJobInput{
			JobName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
				return nil
			}

		}

		job := output.Job
		if job != nil && aws.StringValue(job.Name) == rs.Primary.ID {
			return fmt.Errorf("Glue Job %s still exists", rs.Primary.ID)
		}

		return err
	}

	return nil
}

func testAccAWSGlueJobConfig_Base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy" "AWSGlueServiceRole" {
  arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSGlueServiceRole"
}

resource "aws_iam_role" "test" {
  name = "%s"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "glue.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = data.aws_iam_policy.AWSGlueServiceRole.arn
  role       = aws_iam_role.test.name
}
`, rName)
}

func testAccAWSGlueJobConfig_Command(rName, scriptLocation string) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  max_capacity = 10
  name         = "%s"
  role_arn     = aws_iam_role.test.arn

  command {
    script_location = "%s"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, testAccAWSGlueJobConfig_Base(rName), rName, scriptLocation)
}

func testAccAWSGlueJobConfig_DefaultArguments(rName, jobBookmarkOption, jobLanguage string) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  max_capacity = 10
  name         = "%s"
  role_arn     = aws_iam_role.test.arn

  command {
    script_location = "testscriptlocation"
  }

  default_arguments = {
    "--job-bookmark-option" = "%s"
    "--job-language"        = "%s"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, testAccAWSGlueJobConfig_Base(rName), rName, jobBookmarkOption, jobLanguage)
}

func testAccAWSGlueJobnonOverridableArgumentsConfig(rName, jobBookmarkOption, jobLanguage string) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  max_capacity = 10
  name         = "%s"
  role_arn     = aws_iam_role.test.arn

  command {
    script_location = "testscriptlocation"
  }

  non_overridable_arguments = {
    "--job-bookmark-option" = "%s"
    "--job-language"        = "%s"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, testAccAWSGlueJobConfig_Base(rName), rName, jobBookmarkOption, jobLanguage)
}

func testAccAWSGlueJobConfig_Description(rName, description string) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  description  = "%s"
  max_capacity = 10
  name         = "%s"
  role_arn     = aws_iam_role.test.arn

  command {
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, testAccAWSGlueJobConfig_Base(rName), description, rName)
}

func testAccAWSGlueJobConfig_GlueVersion_MaxCapacity(rName, glueVersion string) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  glue_version = "%s"
  max_capacity = 10
  name         = "%s"
  role_arn     = aws_iam_role.test.arn

  command {
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, testAccAWSGlueJobConfig_Base(rName), glueVersion, rName)
}

func testAccAWSGlueJobConfig_GlueVersion_NumberOfWorkers(rName, glueVersion string) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  glue_version      = "%s"
  name              = "%s"
  number_of_workers = 2
  role_arn          = aws_iam_role.test.arn
  worker_type       = "Standard"

  command {
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, testAccAWSGlueJobConfig_Base(rName), glueVersion, rName)
}

func testAccAWSGlueJobConfig_ExecutionProperty(rName string, maxConcurrentRuns int) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  max_capacity = 10
  name         = "%s"
  role_arn     = aws_iam_role.test.arn

  command {
    script_location = "testscriptlocation"
  }

  execution_property {
    max_concurrent_runs = %d
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, testAccAWSGlueJobConfig_Base(rName), rName, maxConcurrentRuns)
}

func testAccAWSGlueJobConfig_MaxRetries(rName string, maxRetries int) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  max_capacity = 10
  max_retries  = %d
  name         = "%s"
  role_arn     = aws_iam_role.test.arn

  command {
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, testAccAWSGlueJobConfig_Base(rName), maxRetries, rName)
}

func testAccAWSGlueJobConfig_NotificationProperty(rName string, notifyDelayAfter int) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  max_capacity = 10
  name         = "%s"
  role_arn     = aws_iam_role.test.arn

  command {
    script_location = "testscriptlocation"
  }

  notification_property {
    notify_delay_after = %d
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, testAccAWSGlueJobConfig_Base(rName), rName, notifyDelayAfter)
}

func testAccAWSGlueJobConfig_Required(rName string) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  max_capacity = 10
  name         = "%s"
  role_arn     = aws_iam_role.test.arn

  command {
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, testAccAWSGlueJobConfig_Base(rName), rName)
}

func testAccAWSGlueJobConfigTags1(rName, tagKey1, tagValue1 string) string {
	return testAccAWSGlueJobConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_job" "test" {
  name              = %[1]q
  number_of_workers = 2
  role_arn          = aws_iam_role.test.arn
  worker_type       = "Standard"

  command {
    script_location = "testscriptlocation"
  }

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSGlueJobConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAWSGlueJobConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_job" "test" {
  name              = %[1]q
  number_of_workers = 2
  role_arn          = aws_iam_role.test.arn
  worker_type       = "Standard"

  command {
    script_location = "testscriptlocation"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSGlueJobConfig_Timeout(rName string, timeout int) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  max_capacity = 10
  name         = "%s"
  role_arn     = aws_iam_role.test.arn
  timeout      = %d

  command {
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, testAccAWSGlueJobConfig_Base(rName), rName, timeout)
}

func testAccAWSGlueJobConfig_SecurityConfiguration(rName string, securityConfiguration string) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  max_capacity           = 10
  name                   = "%s"
  role_arn               = aws_iam_role.test.arn
  security_configuration = "%s"

  command {
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, testAccAWSGlueJobConfig_Base(rName), rName, securityConfiguration)
}

func testAccAWSGlueJobConfig_WorkerType(rName string, workerType string) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  name              = "%s"
  role_arn          = aws_iam_role.test.arn
  worker_type       = "%s"
  number_of_workers = 10

  command {
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, testAccAWSGlueJobConfig_Base(rName), rName, workerType)
}

func testAccAWSGlueJobConfig_PythonShell(rName string) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  name         = "%s"
  role_arn     = aws_iam_role.test.arn
  max_capacity = 0.0625

  command {
    name            = "pythonshell"
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, testAccAWSGlueJobConfig_Base(rName), rName)
}

func testAccAWSGlueJobConfig_PythonShellWithVersion(rName string, pythonVersion string) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  name         = "%s"
  role_arn     = aws_iam_role.test.arn
  max_capacity = 0.0625

  command {
    name            = "pythonshell"
    script_location = "testscriptlocation"
    python_version  = "%s"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, testAccAWSGlueJobConfig_Base(rName), rName, pythonVersion)
}

func testAccAWSGlueJobConfig_MaxCapacity(rName string, maxCapacity float64) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  name         = "%s"
  role_arn     = aws_iam_role.test.arn
  max_capacity = %g

  command {
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, testAccAWSGlueJobConfig_Base(rName), rName, maxCapacity)
}
