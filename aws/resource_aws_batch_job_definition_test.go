package aws

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/batch/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func init() {
	resource.AddTestSweepers("aws_batch_job_definition", &resource.Sweeper{
		Name: "aws_batch_job_definition",
		F:    testSweepBatchJobDefinitions,
		Dependencies: []string{
			"aws_batch_job_queue",
		},
	})
}

func testSweepBatchJobDefinitions(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).batchconn
	input := &batch.DescribeJobDefinitionsInput{
		Status: aws.String("ACTIVE"),
	}
	var sweeperErrs *multierror.Error

	err = conn.DescribeJobDefinitionsPages(input, func(page *batch.DescribeJobDefinitionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, jobDefinition := range page.JobDefinitions {
			arn := aws.StringValue(jobDefinition.JobDefinitionArn)

			log.Printf("[INFO] Deleting Batch Job Definition: %s", arn)
			_, err := conn.DeregisterJobDefinition(&batch.DeregisterJobDefinitionInput{
				JobDefinition: aws.String(arn),
			})
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Batch Job Definition (%s): %w", arn, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})
	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Batch Job Definitions sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Batch Job Definitions: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSBatchJobDefinition_basic(t *testing.T) {
	var jd batch.JobDefinition
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		ErrorCheck:   testAccErrorCheck(t, batch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchJobDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBatchJobDefinitionConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobDefinitionExists(resourceName, &jd),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "batch", regexp.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
					resource.TestCheckResourceAttrSet(resourceName, "container_properties"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "platform_capabilities.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "propagate_tags", "false"),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "revision"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "container"),
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

func TestAccAWSBatchJobDefinition_disappears(t *testing.T) {
	var jd batch.JobDefinition
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		ErrorCheck:   testAccErrorCheck(t, batch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchJobDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBatchJobDefinitionConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobDefinitionExists(resourceName, &jd),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsBatchJobDefinition(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSBatchJobDefinition_PlatformCapabilities_EC2(t *testing.T) {
	var jd batch.JobDefinition
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		ErrorCheck:   testAccErrorCheck(t, batch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchJobDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBatchJobDefinitionConfigCapabilitiesEC2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobDefinitionExists(resourceName, &jd),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "batch", regexp.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
					resource.TestCheckResourceAttrSet(resourceName, "container_properties"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "platform_capabilities.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "platform_capabilities.*", "EC2"),
					resource.TestCheckResourceAttr(resourceName, "propagate_tags", "false"),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "revision"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "container"),
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

func TestAccAWSBatchJobDefinition_PlatformCapabilities_Fargate_ContainerPropertiesDefaults(t *testing.T) {
	var jd batch.JobDefinition
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		ErrorCheck:   testAccErrorCheck(t, batch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchJobDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBatchJobDefinitionConfigCapabilitiesFargateContainerPropertiesDefaults(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobDefinitionExists(resourceName, &jd),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "batch", regexp.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
					resource.TestCheckResourceAttrSet(resourceName, "container_properties"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "platform_capabilities.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "platform_capabilities.*", "FARGATE"),
					resource.TestCheckResourceAttr(resourceName, "propagate_tags", "false"),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "revision"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "container"),
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

func TestAccAWSBatchJobDefinition_PlatformCapabilities_Fargate(t *testing.T) {
	var jd batch.JobDefinition
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		ErrorCheck:   testAccErrorCheck(t, batch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchJobDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBatchJobDefinitionConfigCapabilitiesFargate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobDefinitionExists(resourceName, &jd),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "batch", regexp.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
					resource.TestCheckResourceAttrSet(resourceName, "container_properties"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "platform_capabilities.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "platform_capabilities.*", "FARGATE"),
					resource.TestCheckResourceAttr(resourceName, "propagate_tags", "false"),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "revision"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "container"),
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

func TestAccAWSBatchJobDefinition_ContainerProperties_Advanced(t *testing.T) {
	var jd batch.JobDefinition
	compare := batch.JobDefinition{
		Parameters: map[string]*string{
			"param1": aws.String("val1"),
			"param2": aws.String("val2"),
		},
		RetryStrategy: &batch.RetryStrategy{
			Attempts: aws.Int64(int64(1)),
			EvaluateOnExit: []*batch.EvaluateOnExit{
				{Action: aws.String(strings.ToLower(batch.RetryActionRetry)), OnStatusReason: aws.String("Host EC2*")},
				{Action: aws.String(strings.ToLower(batch.RetryActionExit)), OnReason: aws.String("*")},
			},
		},
		Timeout: &batch.JobTimeout{
			AttemptDurationSeconds: aws.Int64(int64(60)),
		},
		ContainerProperties: &batch.ContainerProperties{
			Command: []*string{aws.String("ls"), aws.String("-la")},
			Environment: []*batch.KeyValuePair{
				{Name: aws.String("VARNAME"), Value: aws.String("VARVAL")},
			},
			Image:  aws.String("busybox"),
			Memory: aws.Int64(int64(512)),
			MountPoints: []*batch.MountPoint{
				{ContainerPath: aws.String("/tmp"), ReadOnly: aws.Bool(false), SourceVolume: aws.String("tmp")},
			},
			ResourceRequirements: []*batch.ResourceRequirement{},
			Secrets:              []*batch.Secret{},
			Ulimits: []*batch.Ulimit{
				{HardLimit: aws.Int64(int64(1024)), Name: aws.String("nofile"), SoftLimit: aws.Int64(int64(1024))},
			},
			Vcpus: aws.Int64(int64(1)),
			Volumes: []*batch.Volume{
				{
					Host: &batch.Host{SourcePath: aws.String("/tmp")},
					Name: aws.String("tmp"),
				},
			},
		},
	}
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		ErrorCheck:   testAccErrorCheck(t, batch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchJobDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBatchJobDefinitionConfigContainerPropertiesAdvanced(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobDefinitionExists(resourceName, &jd),
					testAccCheckBatchJobDefinitionAttributes(&jd, &compare),
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

func TestAccAWSBatchJobDefinition_updateForcesNewResource(t *testing.T) {
	var before, after batch.JobDefinition
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		ErrorCheck:   testAccErrorCheck(t, batch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchJobDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBatchJobDefinitionConfigContainerPropertiesAdvanced(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobDefinitionExists(resourceName, &before),
					testAccCheckBatchJobDefinitionAttributes(&before, nil),
				),
			},
			{
				Config: testAccBatchJobDefinitionConfigContainerPropertiesAdvancedUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobDefinitionExists(resourceName, &after),
					testAccCheckJobDefinitionRecreated(t, &before, &after),
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

func TestAccAWSBatchJobDefinition_Tags(t *testing.T) {
	var jd batch.JobDefinition
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		ErrorCheck:   testAccErrorCheck(t, batch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchJobDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBatchJobDefinitionConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobDefinitionExists(resourceName, &jd),
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
				Config: testAccBatchJobDefinitionConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobDefinitionExists(resourceName, &jd),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccBatchJobDefinitionConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobDefinitionExists(resourceName, &jd),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSBatchJobDefinition_PropagateTags(t *testing.T) {
	var jd batch.JobDefinition
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		ErrorCheck:   testAccErrorCheck(t, batch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchJobDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBatchJobDefinitionPropagateTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobDefinitionExists(resourceName, &jd),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "batch", regexp.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
					resource.TestCheckResourceAttrSet(resourceName, "container_properties"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "platform_capabilities.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "propagate_tags", "true"),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "revision"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "container"),
				),
			},
		},
	})
}

func testAccCheckBatchJobDefinitionExists(n string, jd *batch.JobDefinition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Batch Job Queue ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).batchconn

		jobDefinition, err := finder.JobDefinitionByARN(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*jd = *jobDefinition

		return nil
	}
}

func testAccCheckBatchJobDefinitionAttributes(jd *batch.JobDefinition, compare *batch.JobDefinition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_batch_job_definition" {
				continue
			}
			if aws.StringValue(jd.JobDefinitionArn) != rs.Primary.Attributes["arn"] {
				return fmt.Errorf("Bad Job Definition ARN\n\t expected: %s\n\tgot: %s\n", rs.Primary.Attributes["arn"], aws.StringValue(jd.JobDefinitionArn))
			}
			if compare != nil {
				if compare.Parameters != nil && !reflect.DeepEqual(compare.Parameters, jd.Parameters) {
					return fmt.Errorf("Bad Job Definition Params\n\t expected: %v\n\tgot: %v\n", compare.Parameters, jd.Parameters)
				}
				if compare.RetryStrategy != nil && aws.Int64Value(compare.RetryStrategy.Attempts) != aws.Int64Value(jd.RetryStrategy.Attempts) {
					return fmt.Errorf("Bad Job Definition Retry Strategy\n\t expected: %d\n\tgot: %d\n", aws.Int64Value(compare.RetryStrategy.Attempts), aws.Int64Value(jd.RetryStrategy.Attempts))
				}
				if compare.RetryStrategy != nil && !reflect.DeepEqual(compare.RetryStrategy.EvaluateOnExit, jd.RetryStrategy.EvaluateOnExit) {
					return fmt.Errorf("Bad Job Definition Retry Strategy\n\t expected: %v\n\tgot: %v\n", compare.RetryStrategy.EvaluateOnExit, jd.RetryStrategy.EvaluateOnExit)
				}
				if compare.ContainerProperties != nil && compare.ContainerProperties.Command != nil && !reflect.DeepEqual(compare.ContainerProperties, jd.ContainerProperties) {
					return fmt.Errorf("Bad Job Definition Container Properties\n\t expected: %s\n\tgot: %s\n", compare.ContainerProperties, jd.ContainerProperties)
				}
			}
		}
		return nil
	}
}

func testAccCheckJobDefinitionRecreated(t *testing.T, before, after *batch.JobDefinition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.Int64Value(before.Revision) == aws.Int64Value(after.Revision) {
			t.Fatalf("Expected change of JobDefinition Revisions, but both were %d", aws.Int64Value(before.Revision))
		}
		return nil
	}
}

func testAccCheckBatchJobDefinitionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).batchconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_batch_job_definition" {
			continue
		}

		_, err := finder.JobDefinitionByARN(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Batch Job Definition %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccBatchJobDefinitionConfigContainerPropertiesAdvanced(rName string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  name = %[1]q
  type = "container"
  parameters = {
    param1 = "val1"
    param2 = "val2"
  }
  retry_strategy {
    attempts = 1
    evaluate_on_exit {
      action           = "RETRY"
      on_status_reason = "Host EC2*"
    }
    evaluate_on_exit {
      action    = "exit"
      on_reason = "*"
    }
  }
  timeout {
    attempt_duration_seconds = 60
  }
  container_properties = <<CONTAINER_PROPERTIES
{
    "command": ["ls", "-la"],
    "image": "busybox",
    "memory": 512,
    "vcpus": 1,
    "volumes": [
      {
        "host": {
          "sourcePath": "/tmp"
        },
        "name": "tmp"
      }
    ],
    "environment": [
        {"name": "VARNAME", "value": "VARVAL"}
    ],
    "mountPoints": [
        {
          "sourceVolume": "tmp",
          "containerPath": "/tmp",
          "readOnly": false
        }
    ],
    "ulimits": [
      {
        "hardLimit": 1024,
        "name": "nofile",
        "softLimit": 1024
      }
    ]
}
CONTAINER_PROPERTIES

}
`, rName)
}

func testAccBatchJobDefinitionConfigContainerPropertiesAdvancedUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  name                 = %[1]q
  type                 = "container"
  container_properties = <<CONTAINER_PROPERTIES
{
    "command": ["ls", "-la"],
    "image": "busybox",
    "memory": 1024,
    "vcpus": 1,
    "volumes": [
      {
        "host": {
          "sourcePath": "/tmp"
        },
        "name": "tmp"
      }
    ],
    "environment": [
        {"name": "VARNAME", "value": "VARVAL"}
    ],
    "mountPoints": [
        {
          "sourceVolume": "tmp",
          "containerPath": "/tmp",
          "readOnly": false
        }
    ],
    "ulimits": [
      {
        "hardLimit": 1024,
        "name": "nofile",
        "softLimit": 1024
      }
    ]
}
CONTAINER_PROPERTIES

}
`, rName)
}

func testAccBatchJobDefinitionConfigName(rName string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  container_properties = jsonencode({
    command = ["echo", "test"]
    image   = "busybox"
    memory  = 128
    vcpus   = 1
  })
  name = %[1]q
  type = "container"
}
`, rName)
}

func testAccBatchJobDefinitionConfigCapabilitiesEC2(rName string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  name = %[1]q
  type = "container"

  platform_capabilities = [
    "EC2",
  ]

  container_properties = jsonencode({
    command = ["echo", "test"]
    image   = "busybox"
    memory  = 128
    vcpus   = 1
  })
}
`, rName)
}

func testAccBatchJobDefinitionConfigCapabilitiesFargateContainerPropertiesDefaults(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "ecs_task_execution_role" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.assume_role_policy.json
}

data "aws_iam_policy_document" "assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "ecs_task_execution_role_policy" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_batch_job_definition" "test" {
  name = %[1]q
  type = "container"

  platform_capabilities = [
    "FARGATE",
  ]

  container_properties = <<CONTAINER_PROPERTIES
{
  "image": "busybox",
  "resourceRequirements": [
    {"type": "MEMORY", "value": "512"},
    {"type": "VCPU", "value": "0.25"}
  ],
  "executionRoleArn": "${aws_iam_role.ecs_task_execution_role.arn}"
}
CONTAINER_PROPERTIES
}
`, rName)
}

func testAccBatchJobDefinitionConfigCapabilitiesFargate(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "ecs_task_execution_role" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.assume_role_policy.json
}

data "aws_iam_policy_document" "assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "ecs_task_execution_role_policy" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_batch_job_definition" "test" {
  name = %[1]q
  type = "container"

  platform_capabilities = [
    "FARGATE",
  ]

  container_properties = <<CONTAINER_PROPERTIES
{
  "command": ["echo", "test"],
  "image": "busybox",
  "fargatePlatformConfiguration": {
    "platformVersion": "LATEST"
  },
  "networkConfiguration": {
    "assignPublicIp": "DISABLED"
  },
  "resourceRequirements": [
    {"type": "VCPU", "value": "0.25"},
    {"type": "MEMORY", "value": "512"}
  ],
  "executionRoleArn": "${aws_iam_role.ecs_task_execution_role.arn}"
}
CONTAINER_PROPERTIES
}
`, rName)
}

func testAccBatchJobDefinitionConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  container_properties = jsonencode({
    command = ["echo", "test"]
    image   = "busybox"
    memory  = 128
    vcpus   = 1
  })
  name = %[1]q
  type = "container"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccBatchJobDefinitionConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  container_properties = jsonencode({
    command = ["echo", "test"]
    image   = "busybox"
    memory  = 128
    vcpus   = 1
  })
  name = %[1]q
  type = "container"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccBatchJobDefinitionPropagateTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  container_properties = jsonencode({
    command = ["echo", "test"]
    image   = "busybox"
    memory  = 128
    vcpus   = 1
  })
  name = %[1]q
  type = "container"

  propagate_tags = true
}
`, rName)
}
