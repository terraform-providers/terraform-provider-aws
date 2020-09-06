package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSSSMMaintenanceWindowTask_basic(t *testing.T) {
	var before, after ssm.MaintenanceWindowTask
	resourceName := "aws_ssm_maintenance_window_task.test"

	name := acctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMMaintenanceWindowTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMMaintenanceWindowTaskBasicConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowTaskExists(resourceName, &before),
				),
			},
			{
				Config: testAccAWSSSMMaintenanceWindowTaskBasicConfigUpdate(name, "test description", "RUN_COMMAND", "AWS-InstallPowerShellModule", 3, 3, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowTaskExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("maintenance-window-task-%s", name)),
					resource.TestCheckResourceAttr(resourceName, "description", "test description"),
					resource.TestCheckResourceAttr(resourceName, "task_type", "RUN_COMMAND"),
					resource.TestCheckResourceAttr(resourceName, "task_arn", "AWS-InstallPowerShellModule"),
					resource.TestCheckResourceAttr(resourceName, "priority", "3"),
					resource.TestCheckResourceAttr(resourceName, "max_concurrency", "3"),
					resource.TestCheckResourceAttr(resourceName, "max_errors", "2"),
					testAccCheckAwsSsmWindowsTaskNotRecreated(t, &before, &after),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSSSMMaintenanceWindowTaskImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSSMMaintenanceWindowTask_updateForcesNewResource(t *testing.T) {
	var before, after ssm.MaintenanceWindowTask
	name := acctest.RandString(10)
	resourceName := "aws_ssm_maintenance_window_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMMaintenanceWindowTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMMaintenanceWindowTaskBasicConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowTaskExists(resourceName, &before),
				),
			},
			{
				Config: testAccAWSSSMMaintenanceWindowTaskBasicConfigUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowTaskExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", "TestMaintenanceWindowTask"),
					resource.TestCheckResourceAttr(resourceName, "description", "This resource is for test purpose only"),
					testAccCheckAwsSsmWindowsTaskRecreated(t, &before, &after),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSSSMMaintenanceWindowTaskImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSSMMaintenanceWindowTask_TaskInvocationAutomationParameters(t *testing.T) {
	var task ssm.MaintenanceWindowTask
	resourceName := "aws_ssm_maintenance_window_task.test"

	name := acctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMMaintenanceWindowTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMMaintenanceWindowTaskAutomationConfig(name, "$DEFAULT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowTaskExists(resourceName, &task),
					resource.TestCheckResourceAttr(resourceName, "task_invocation_parameters.0.automation_parameters.0.document_version", "$DEFAULT"),
				),
			},
			{
				Config: testAccAWSSSMMaintenanceWindowTaskAutomationConfigUpdate(name, "$LATEST"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowTaskExists(resourceName, &task),
					resource.TestCheckResourceAttr(resourceName, "task_invocation_parameters.0.automation_parameters.0.document_version", "$LATEST"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSSSMMaintenanceWindowTaskImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSSMMaintenanceWindowTask_TaskInvocationLambdaParameters(t *testing.T) {
	var task ssm.MaintenanceWindowTask
	resourceName := "aws_ssm_maintenance_window_task.test"
	rString := acctest.RandString(8)
	rInt := acctest.RandInt()

	funcName := fmt.Sprintf("tf_acc_lambda_func_tags_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_tags_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_tags_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_tags_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMMaintenanceWindowTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMMaintenanceWindowTaskLambdaConfig(funcName, policyName, roleName, sgName, rString, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowTaskExists(resourceName, &task),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSSSMMaintenanceWindowTaskImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSSMMaintenanceWindowTask_TaskInvocationRunCommandParameters(t *testing.T) {
	var task ssm.MaintenanceWindowTask
	resourceName := "aws_ssm_maintenance_window_task.test"
	serviceRoleResourceName := "aws_iam_role.test"
	s3BucketResourceName := "aws_s3_bucket.foo"

	name := acctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMMaintenanceWindowTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMMaintenanceWindowTaskRunCommandConfig(name, "test comment", 30),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowTaskExists(resourceName, &task),
					resource.TestCheckResourceAttrPair(resourceName, "service_role_arn", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "task_invocation_parameters.0.run_command_parameters.0.service_role_arn", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "task_invocation_parameters.0.run_command_parameters.0.comment", "test comment"),
					resource.TestCheckResourceAttr(resourceName, "task_invocation_parameters.0.run_command_parameters.0.timeout_seconds", "30"),
				),
			},
			{
				Config: testAccAWSSSMMaintenanceWindowTaskRunCommandConfigUpdate(name, "test comment update", 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowTaskExists(resourceName, &task),
					resource.TestCheckResourceAttr(resourceName, "task_invocation_parameters.0.run_command_parameters.0.comment", "test comment update"),
					resource.TestCheckResourceAttr(resourceName, "task_invocation_parameters.0.run_command_parameters.0.timeout_seconds", "60"),
					resource.TestCheckResourceAttrPair(resourceName, "task_invocation_parameters.0.run_command_parameters.0.output_s3_bucket", s3BucketResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSSSMMaintenanceWindowTaskImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSSMMaintenanceWindowTask_TaskInvocationStepFunctionParameters(t *testing.T) {
	var task ssm.MaintenanceWindowTask
	resourceName := "aws_ssm_maintenance_window_task.test"
	rString := acctest.RandString(8)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMMaintenanceWindowTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMMaintenanceWindowTaskStepFunctionConfig(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowTaskExists(resourceName, &task),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSSSMMaintenanceWindowTaskImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSSMMaintenanceWindowTask_emptyNotificationConfig(t *testing.T) {
	var task ssm.MaintenanceWindowTask
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ssm_maintenance_window_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMMaintenanceWindowTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMMaintenanceWindowTaskConfigEmptyNotifcationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowTaskExists(resourceName, &task),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsSsmWindowsTaskNotRecreated(t *testing.T,
	before, after *ssm.MaintenanceWindowTask) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(before.WindowTaskId) != aws.StringValue(after.WindowTaskId) {
			t.Fatalf("Unexpected change of Windows Task IDs, but both were %s and %s", aws.StringValue(before.WindowTaskId), aws.StringValue(after.WindowTaskId))
		}
		return nil
	}
}

func testAccCheckAwsSsmWindowsTaskRecreated(t *testing.T,
	before, after *ssm.MaintenanceWindowTask) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before.WindowTaskId == after.WindowTaskId {
			t.Fatalf("Expected change of Windows Task IDs, but both were %v", before.WindowTaskId)
		}
		return nil
	}
}

func testAccCheckAWSSSMMaintenanceWindowTaskExists(n string, task *ssm.MaintenanceWindowTask) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSM Maintenance Window Task Window ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ssmconn

		resp, err := conn.DescribeMaintenanceWindowTasks(&ssm.DescribeMaintenanceWindowTasksInput{
			WindowId: aws.String(rs.Primary.Attributes["window_id"]),
		})
		if err != nil {
			return err
		}

		for _, i := range resp.Tasks {
			if *i.WindowTaskId == rs.Primary.ID {
				*task = *i
				return nil
			}
		}

		return fmt.Errorf("No AWS SSM Maintenance window task found")
	}
}

func testAccCheckAWSSSMMaintenanceWindowTaskDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ssmconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ssm_maintenance_window_target" {
			continue
		}

		out, err := conn.DescribeMaintenanceWindowTasks(&ssm.DescribeMaintenanceWindowTasksInput{
			WindowId: aws.String(rs.Primary.Attributes["window_id"]),
		})

		if err != nil {
			// Verify the error is what we want
			if ae, ok := err.(awserr.Error); ok && ae.Code() == "DoesNotExistException" {
				continue
			}
			return err
		}

		if len(out.Tasks) > 0 {
			return fmt.Errorf("Expected AWS SSM Maintenance Task to be gone, but was still found")
		}

		return nil
	}

	return nil
}

func testAccAWSSSMMaintenanceWindowTaskImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["window_id"], rs.Primary.ID), nil
	}
}

func testAccAWSSSMMaintenanceWindowTaskConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff   = 1
  duration = 3
  name     = %[1]q
  schedule = "cron(0 16 ? * TUE *)"
}

resource "aws_ssm_maintenance_window_target" "test" {
  name          = %[1]q
  resource_type = "INSTANCE"
  window_id     = aws_ssm_maintenance_window.test.id

  targets {
    key    = "tag:Name"
    values = ["tf-acc-test"]
  }
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "events.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY

}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.name

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "ssm:*",
    "Resource": "*"
  }
}
POLICY

}
`, rName)
}

func testAccAWSSSMMaintenanceWindowTaskBasicConfig(rName string) string {
	return fmt.Sprintf(testAccAWSSSMMaintenanceWindowTaskConfigBase(rName) + `

resource "aws_ssm_maintenance_window_task" "test" {
  window_id        = aws_ssm_maintenance_window.test.id
  task_type        = "RUN_COMMAND"
  task_arn         = "AWS-RunShellScript"
  priority         = 1
  service_role_arn = aws_iam_role.test.arn
  max_concurrency  = "2"
  max_errors       = "1"

  targets {
    key    = "WindowTargetIds"
    values = [aws_ssm_maintenance_window_target.test.id]
  }

  task_invocation_parameters {
    run_command_parameters {
      parameter {
        name   = "commands"
        values = ["pwd"]
      }
    }
  }
}
`)
}

func testAccAWSSSMMaintenanceWindowTaskBasicConfigUpdate(rName, description, taskType, taskArn string, priority, maxConcurrency, maxErrors int) string {
	return fmt.Sprintf(testAccAWSSSMMaintenanceWindowTaskConfigBase(rName)+`

resource "aws_ssm_maintenance_window_task" "test" {
  window_id        = aws_ssm_maintenance_window.test.id
  task_type        = "%[2]s"
  task_arn         = "%[3]s"
  name             = "maintenance-window-task-%[1]s"
  description      = "%[4]s"
  priority         = %[5]d
  service_role_arn = aws_iam_role.ssm_role_update.arn
  max_concurrency  = %[6]d
  max_errors       = %[7]d

  targets {
    key    = "WindowTargetIds"
    values = [aws_ssm_maintenance_window_target.test.id]
  }

  task_invocation_parameters {
    run_command_parameters {
      parameter {
        name   = "commands"
        values = ["pwd"]
      }
    }
  }
}

resource "aws_iam_role" "ssm_role_update" {
  name = "ssm-role-update-%[1]s"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "events.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY

}

resource "aws_iam_role_policy" "bar" {
  name = "ssm_role_policy_update_%[1]s"
  role = aws_iam_role.ssm_role_update.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "ssm:*",
    "Resource": "*"
  }
}
EOF

}
`, rName, taskType, taskArn, description, priority, maxConcurrency, maxErrors)
}

func testAccAWSSSMMaintenanceWindowTaskBasicConfigUpdated(rName string) string {
	return fmt.Sprintf(testAccAWSSSMMaintenanceWindowTaskConfigBase(rName) + `

resource "aws_ssm_maintenance_window_task" "test" {
  window_id        = aws_ssm_maintenance_window.test.id
  task_type        = "RUN_COMMAND"
  task_arn         = "AWS-RunShellScript"
  priority         = 1
  name             = "TestMaintenanceWindowTask"
  description      = "This resource is for test purpose only"
  service_role_arn = aws_iam_role.test.arn
  max_concurrency  = "2"
  max_errors       = "1"

  targets {
    key    = "WindowTargetIds"
    values = [aws_ssm_maintenance_window_target.test.id]
  }

  task_invocation_parameters {
    run_command_parameters {
      parameter {
        name   = "commands"
        values = ["date"]
      }
    }
  }
}
`)
}

func testAccAWSSSMMaintenanceWindowTaskConfigEmptyNotifcationConfig(rName string) string {
	return fmt.Sprintf(testAccAWSSSMMaintenanceWindowTaskConfigBase(rName) + `

resource "aws_ssm_maintenance_window_task" "test" {
  window_id        = aws_ssm_maintenance_window.test.id
  task_type        = "RUN_COMMAND"
  task_arn         = "AWS-CreateImage"
  priority         = 1
  service_role_arn = aws_iam_role.test.arn
  max_concurrency  = "2"
  max_errors       = "1"

  targets {
    key    = "WindowTargetIds"
    values = [aws_ssm_maintenance_window_target.test.id]
  }

  task_invocation_parameters {
    run_command_parameters {
      timeout_seconds = 600

      notification_config {}

      parameter {
        name   = "Operation"
        values = ["Install"]
      }
    }
  }
}
`)
}

func testAccAWSSSMMaintenanceWindowTaskAutomationConfig(rName, version string) string {
	return fmt.Sprintf(testAccAWSSSMMaintenanceWindowTaskConfigBase(rName)+`

resource "aws_ssm_maintenance_window_task" "test" {
  window_id        = aws_ssm_maintenance_window.test.id
  task_type        = "AUTOMATION"
  task_arn         = "AWS-CreateImage"
  priority         = 1
  service_role_arn = aws_iam_role.test.arn
  max_concurrency  = "2"
  max_errors       = "1"

  targets {
    key    = "WindowTargetIds"
    values = [aws_ssm_maintenance_window_target.test.id]
  }

  task_invocation_parameters {
    automation_parameters {
      document_version = "%[2]s"

      parameter {
        name   = "InstanceId"
        values = ["{{TARGET_ID}}"]
      }

      parameter {
        name   = "NoReboot"
        values = ["false"]
      }
    }
  }
}
`, rName, version)
}

func testAccAWSSSMMaintenanceWindowTaskAutomationConfigUpdate(rName, version string) string {
	return fmt.Sprintf(testAccAWSSSMMaintenanceWindowTaskConfigBase(rName)+`
resource "aws_s3_bucket" "foo" {
  bucket        = "tf-s3-%[1]s"
  acl           = "private"
  force_destroy = true
}

resource "aws_ssm_maintenance_window_task" "test" {
  window_id        = aws_ssm_maintenance_window.test.id
  task_type        = "AUTOMATION"
  task_arn         = "AWS-CreateImage"
  priority         = 1
  service_role_arn = aws_iam_role.test.arn
  max_concurrency  = "2"
  max_errors       = "1"

  targets {
    key    = "WindowTargetIds"
    values = [aws_ssm_maintenance_window_target.test.id]
  }

  task_invocation_parameters {
    automation_parameters {
      document_version = "%[2]s"

      parameter {
        name   = "InstanceId"
        values = ["{{TARGET_ID}}"]
      }

      parameter {
        name   = "NoReboot"
        values = ["false"]
      }
    }
  }
}
`, rName, version)
}

func testAccAWSSSMMaintenanceWindowTaskLambdaConfig(funcName, policyName, roleName, sgName, rName string, rInt int) string {
	return fmt.Sprintf(testAccAWSLambdaConfigBasic(funcName, policyName, roleName, sgName)+
		testAccAWSSSMMaintenanceWindowTaskConfigBase(rName)+`

resource "aws_ssm_maintenance_window_task" "test" {
  window_id        = aws_ssm_maintenance_window.test.id
  task_type        = "LAMBDA"
  task_arn         = aws_lambda_function.test.arn
  priority         = 1
  service_role_arn = aws_iam_role.test.arn
  max_concurrency  = "2"
  max_errors       = "1"

  targets {
    key    = "WindowTargetIds"
    values = [aws_ssm_maintenance_window_target.test.id]
  }

  task_invocation_parameters {
    lambda_parameters {
      client_context = base64encode(jsonencode({
        key1 = "value1"
        key2 = "value2"
        key3 = "value3"
      }))
      payload = jsonencode({
        number = %[2]d
      })
    }
  }
}
`, rName, rInt)
}

func testAccAWSSSMMaintenanceWindowTaskRunCommandConfig(rName, comment string, timeoutSeconds int) string {
	return fmt.Sprintf(testAccAWSSSMMaintenanceWindowTaskConfigBase(rName)+`

resource "aws_ssm_maintenance_window_task" "test" {
  window_id        = aws_ssm_maintenance_window.test.id
  task_type        = "RUN_COMMAND"
  task_arn         = "AWS-RunShellScript"
  priority         = 1
  service_role_arn = aws_iam_role.test.arn
  max_concurrency  = "2"
  max_errors       = "1"

  targets {
    key    = "WindowTargetIds"
    values = [aws_ssm_maintenance_window_target.test.id]
  }

  task_invocation_parameters {
    run_command_parameters {
      comment            = "%[2]s"
      document_hash      = sha256("COMMAND")
      document_hash_type = "Sha256"
      service_role_arn   = aws_iam_role.test.arn
      timeout_seconds    = %[3]d

      parameter {
        name   = "commands"
        values = ["date"]
      }
    }
  }
}
`, rName, comment, timeoutSeconds)
}

func testAccAWSSSMMaintenanceWindowTaskRunCommandConfigUpdate(rName, comment string, timeoutSeconds int) string {
	return fmt.Sprintf(testAccAWSSSMMaintenanceWindowTaskConfigBase(rName)+`
resource "aws_s3_bucket" "foo" {
  bucket        = "tf-s3-%[1]s"
  acl           = "private"
  force_destroy = true
}

resource "aws_ssm_maintenance_window_task" "test" {
  window_id        = aws_ssm_maintenance_window.test.id
  task_type        = "RUN_COMMAND"
  task_arn         = "AWS-RunShellScript"
  priority         = 1
  service_role_arn = aws_iam_role.test.arn
  max_concurrency  = "2"
  max_errors       = "1"

  targets {
    key    = "WindowTargetIds"
    values = [aws_ssm_maintenance_window_target.test.id]
  }

  task_invocation_parameters {
    run_command_parameters {
      comment              = "%[2]s"
      document_hash        = sha256("COMMAND")
      document_hash_type   = "Sha256"
      service_role_arn     = aws_iam_role.test.arn
      timeout_seconds      = %[3]d
      output_s3_bucket     = aws_s3_bucket.foo.id
      output_s3_key_prefix = "foo"

      parameter {
        name   = "commands"
        values = ["date"]
      }
    }
  }
}
`, rName, comment, timeoutSeconds)
}

func testAccAWSSSMMaintenanceWindowTaskStepFunctionConfig(rName string) string {
	return testAccAWSSSMMaintenanceWindowTaskConfigBase(rName) + fmt.Sprintf(`
resource "aws_sfn_activity" "test" {
  name = %[1]q
}

resource "aws_ssm_maintenance_window_task" "test" {
  window_id        = aws_ssm_maintenance_window.test.id
  task_type        = "STEP_FUNCTIONS"
  task_arn         = aws_sfn_activity.test.id
  priority         = 1
  service_role_arn = aws_iam_role.test.arn
  max_concurrency  = "2"
  max_errors       = "1"

  targets {
    key    = "WindowTargetIds"
    values = [aws_ssm_maintenance_window_target.test.id]
  }

  task_invocation_parameters {
    step_functions_parameters {
      input = jsonencode({
        key1 = "value1"
        key2 = "value2"
        key3 = "value3"
      })
      name = "tf-step-function-%[1]s"
    }
  }
}
`, rName)
}
