package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
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

	prefixes := []string{
		"tf-acc-test-",
	}

	input := &glue.GetJobsInput{}
	err = conn.GetJobsPages(input, func(page *glue.GetJobsOutput, lastPage bool) bool {
		if len(page.Jobs) == 0 {
			log.Printf("[INFO] No Glue Jobs to sweep")
			return false
		}
		for _, job := range page.Jobs {
			skip := true
			name := job.Name
			for _, prefix := range prefixes {
				if strings.HasPrefix(*name, prefix) {
					skip = false
					break
				}
			}
			if skip {
				log.Printf("[INFO] Skipping Glue Job: %s", *name)
				continue
			}

			log.Printf("[INFO] Deleting Glue Job: %s", *name)
			err := deleteGlueJob(conn, *name)
			if err != nil {
				log.Printf("[ERROR] Failed to delete Glue Job %s: %s", *name, err)
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

func TestAccAWSGlueJob_Basic(t *testing.T) {
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
					resource.TestCheckResourceAttr(resourceName, "command.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "command.0.script_location", "testscriptlocation"),
					resource.TestCheckResourceAttr(resourceName, "default_arguments.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestMatchResourceAttr(resourceName, "role_arn", regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:iam::[^:]+:role/%s", rName))),
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

func TestAccAWSGlueJob_AllocatedCapacity(t *testing.T) {
	var job glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueJobDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSGlueJobConfig_AllocatedCapacity(rName, 1),
				ExpectError: regexp.MustCompile(`expected allocated_capacity to be at least`),
			},
			{
				Config: testAccAWSGlueJobConfig_AllocatedCapacity(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "allocated_capacity", "2"),
				),
			},
			{
				Config: testAccAWSGlueJobConfig_AllocatedCapacity(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "allocated_capacity", "3"),
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
data "aws_iam_policy" "AWSGlueServiceRole" {
  arn = "arn:aws:iam::aws:policy/service-role/AWSGlueServiceRole"
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
        "Service": "glue.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "${data.aws_iam_policy.AWSGlueServiceRole.arn}"
  role       = "${aws_iam_role.test.name}"
}
`, rName)
}

func testAccAWSGlueJobConfig_AllocatedCapacity(rName string, allocatedCapacity int) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  allocated_capacity = %d
  name               = "%s"
  role_arn           = "${aws_iam_role.test.arn}"

  command {
    script_location = "testscriptlocation"
  }

  depends_on = ["aws_iam_role_policy_attachment.test"]
}
`, testAccAWSGlueJobConfig_Base(rName), allocatedCapacity, rName)
}

func testAccAWSGlueJobConfig_Command(rName, scriptLocation string) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  name     = "%s"
  role_arn = "${aws_iam_role.test.arn}"

  command {
    script_location = "%s"
  }

  depends_on = ["aws_iam_role_policy_attachment.test"]
}
`, testAccAWSGlueJobConfig_Base(rName), rName, scriptLocation)
}

func testAccAWSGlueJobConfig_DefaultArguments(rName, jobBookmarkOption, jobLanguage string) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  name     = "%s"
  role_arn = "${aws_iam_role.test.arn}"

  command {
    script_location = "testscriptlocation"
  }

  default_arguments = {
    "--job-bookmark-option" = "%s"
    "--job-language"        = "%s"
  }

  depends_on = ["aws_iam_role_policy_attachment.test"]
}
`, testAccAWSGlueJobConfig_Base(rName), rName, jobBookmarkOption, jobLanguage)
}

func testAccAWSGlueJobConfig_Description(rName, description string) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  description = "%s"
  name        = "%s"
  role_arn    = "${aws_iam_role.test.arn}"

  command {
    script_location = "testscriptlocation"
  }

  depends_on = ["aws_iam_role_policy_attachment.test"]
}
`, testAccAWSGlueJobConfig_Base(rName), description, rName)
}

func testAccAWSGlueJobConfig_ExecutionProperty(rName string, maxConcurrentRuns int) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  name     = "%s"
  role_arn = "${aws_iam_role.test.arn}"

  command {
    script_location = "testscriptlocation"
  }

  execution_property {
    max_concurrent_runs = %d
  }

  depends_on = ["aws_iam_role_policy_attachment.test"]
}
`, testAccAWSGlueJobConfig_Base(rName), rName, maxConcurrentRuns)
}

func testAccAWSGlueJobConfig_MaxRetries(rName string, maxRetries int) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  max_retries = %d
  name        = "%s"
  role_arn    = "${aws_iam_role.test.arn}"

  command {
    script_location = "testscriptlocation"
  }

  depends_on = ["aws_iam_role_policy_attachment.test"]
}
`, testAccAWSGlueJobConfig_Base(rName), maxRetries, rName)
}

func testAccAWSGlueJobConfig_Required(rName string) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  name     = "%s"
  role_arn = "${aws_iam_role.test.arn}"

  command {
    script_location = "testscriptlocation"
  }

  depends_on = ["aws_iam_role_policy_attachment.test"]
}
`, testAccAWSGlueJobConfig_Base(rName), rName)
}

func testAccAWSGlueJobConfig_Timeout(rName string, timeout int) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  name        = "%s"
  role_arn    = "${aws_iam_role.test.arn}"
  timeout     = %d

  command {
    script_location = "testscriptlocation"
  }

  depends_on = ["aws_iam_role_policy_attachment.test"]
}
`, testAccAWSGlueJobConfig_Base(rName), rName, timeout)
}

func testAccAWSGlueJobConfig_SecurityConfiguration(rName string, securityConfiguration string) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
 name                   = "%s"
 role_arn               = "${aws_iam_role.test.arn}"
 security_configuration = "%s"

 command {
   script_location = "testscriptlocation"
 }

 depends_on = ["aws_iam_role_policy_attachment.test"]
}
`, testAccAWSGlueJobConfig_Base(rName), rName, securityConfiguration)
}
