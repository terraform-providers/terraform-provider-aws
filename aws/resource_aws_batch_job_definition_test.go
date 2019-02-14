package aws

import (
	"fmt"
	"strings"
	"testing"

	"reflect"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSBatchJobDefinition_basic(t *testing.T) {
	var jd batch.JobDefinition
	compare := batch.JobDefinition{
		Parameters: map[string]*string{
			"param1": aws.String("val1"),
			"param2": aws.String("val2"),
		},
		RetryStrategy: &batch.RetryStrategy{
			Attempts: aws.Int64(int64(1)),
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
			Ulimits: []*batch.Ulimit{
				{HardLimit: aws.Int64(int64(1024)), Name: aws.String("nofile"), SoftLimit: aws.Int64(int64(1024))},
			},
			Volumes: []*batch.Volume{
				{
					Host: &batch.Host{SourcePath: aws.String("/tmp")},
					Name: aws.String("tmp"),
				},
			},
			Vcpus: aws.Int64(int64(1)),
		},
	}
	ri := acctest.RandInt()
	config := fmt.Sprintf(testAccBatchJobDefinitionBaseConfig, ri)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchJobDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobDefinitionExists("aws_batch_job_definition.test", &jd),
					testAccCheckBatchJobDefinitionAttributes(&jd, &compare),
				),
			},
		},
	})
}

func TestAccAWSBatchJobDefinition_updateForcesNewResource(t *testing.T) {
	var before batch.JobDefinition
	var after batch.JobDefinition
	ri := acctest.RandInt()
	config := fmt.Sprintf(testAccBatchJobDefinitionBaseConfig, ri)
	updateConfig := fmt.Sprintf(testAccBatchJobDefinitionUpdateConfig, ri)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchJobDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobDefinitionExists("aws_batch_job_definition.test", &before),
					testAccCheckBatchJobDefinitionAttributes(&before, nil),
				),
			},
			{
				Config: updateConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobDefinitionExists("aws_batch_job_definition.test", &after),
					testAccCheckJobDefinitionRecreated(t, &before, &after),
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
		arn := rs.Primary.Attributes["arn"]
		def, err := getJobDefinition(conn, arn)
		if err != nil {
			return err
		}
		if def == nil {
			return fmt.Errorf("Not found: %s", n)
		}
		*jd = *def

		return nil
	}
}

func testAccCheckBatchJobDefinitionAttributes(jd *batch.JobDefinition, compare *batch.JobDefinition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !strings.HasPrefix(*jd.JobDefinitionName, "tf_acctest_batch_job_definition") {
			return fmt.Errorf("Bad Job Definition name: %s", *jd.JobDefinitionName)
		}
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_batch_job_definition" {
				continue
			}
			if *jd.JobDefinitionArn != rs.Primary.Attributes["arn"] {
				return fmt.Errorf("Bad Job Definition ARN\n\t expected: %s\n\tgot: %s\n", rs.Primary.Attributes["arn"], *jd.JobDefinitionArn)
			}
			if compare != nil {
				if compare.Parameters != nil && !reflect.DeepEqual(compare.Parameters, jd.Parameters) {
					return fmt.Errorf("Bad Job Definition Params\n\t expected: %v\n\tgot: %v\n", compare.Parameters, jd.Parameters)
				}
				if compare.RetryStrategy != nil && *compare.RetryStrategy.Attempts != *jd.RetryStrategy.Attempts {
					return fmt.Errorf("Bad Job Definition Retry Strategy\n\t expected: %d\n\tgot: %d\n", *compare.RetryStrategy.Attempts, *jd.RetryStrategy.Attempts)
				}
				if compare.ContainerProperties != nil && compare.ContainerProperties.Command != nil && !reflect.DeepEqual(compare.ContainerProperties, jd.ContainerProperties) {
					return fmt.Errorf("Bad Job Definition Container Properties\n\t expected: %s\n\tgot: %s\n", compare.ContainerProperties, jd.ContainerProperties)
				}
			}
		}
		return nil
	}
}

func testAccCheckJobDefinitionRecreated(t *testing.T,
	before, after *batch.JobDefinition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *before.Revision == *after.Revision {
			t.Fatalf("Expected change of JobDefinition Revisions, but both were %v", before.Revision)
		}
		return nil
	}
}

func testAccCheckBatchJobDefinitionDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_batch_job_definition" {
			continue
		}
		conn := testAccProvider.Meta().(*AWSClient).batchconn
		js, err := getJobDefinition(conn, rs.Primary.Attributes["arn"])
		if err == nil && js != nil {
			if *js.Status == "ACTIVE" {
				return fmt.Errorf("Error: Job Definition still active")
			}
		}
		return nil
	}
	return nil
}

const testAccBatchJobDefinitionBaseConfig = `
resource "aws_batch_job_definition" "test" {
	name = "tf_acctest_batch_job_definition_%[1]d"
	type = "container"
	parameters = {
		param1 = "val1"
		param2 = "val2"
	}
	retry_strategy {
		attempts = 1
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
`

const testAccBatchJobDefinitionUpdateConfig = `
resource "aws_batch_job_definition" "test" {
	name = "tf_acctest_batch_job_definition_%[1]d"
	type = "container"
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
`
