package aws

import (
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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

	err = conn.DescribeJobDefinitionsPages(input, func(page *batch.DescribeJobDefinitionsOutput, isLast bool) bool {
		if page == nil {
			return !isLast
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

		return !isLast
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
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBatchJobDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBatchJobDefinitionConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobDefinitionExists(resourceName, &jd),
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

func TestAccAWSBatchJobDefinition_ContainerProperties_Advanced(t *testing.T) {
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
