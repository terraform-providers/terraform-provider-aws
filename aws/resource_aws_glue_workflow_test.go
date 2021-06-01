package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_glue_workflow", &resource.Sweeper{
		Name: "aws_glue_workflow",
		F:    testSweepGlueWorkflow,
	})
}

func testSweepGlueWorkflow(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).glueconn

	listOutput, err := conn.ListWorkflows(&glue.ListWorkflowsInput{})
	if err != nil {
		// Some endpoints that do not support Glue Workflows return InternalFailure
		if testSweepSkipSweepError(err) || isAWSErr(err, "InternalFailure", "") {
			log.Printf("[WARN] Skipping Glue Workflow sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Glue Workflow: %s", err)
	}
	for _, workflowName := range listOutput.Workflows {
		err := deleteWorkflow(conn, *workflowName)
		if err != nil {
			log.Printf("[ERROR] Failed to delete Glue Workflow %s: %s", *workflowName, err)
		}
	}
	return nil
}

func TestAccAWSGlueWorkflow_basic(t *testing.T) {
	var workflow glue.Workflow

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_workflow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSGlueWorkflow(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueWorkflowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueWorkflowConfig_Required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueWorkflowExists(resourceName, &workflow),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("workflow/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccAWSGlueWorkflow_maxConcurrentRuns(t *testing.T) {
	var workflow glue.Workflow

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_workflow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSGlueWorkflow(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueWorkflowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueWorkflowConfigMaxConcurrentRuns(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueWorkflowExists(resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, "max_concurrent_runs", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSGlueWorkflowConfigMaxConcurrentRuns(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueWorkflowExists(resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, "max_concurrent_runs", "2"),
				),
			},
			{
				Config: testAccAWSGlueWorkflowConfigMaxConcurrentRuns(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueWorkflowExists(resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, "max_concurrent_runs", "1"),
				),
			},
		},
	})
}

func TestAccAWSGlueWorkflow_DefaultRunProperties(t *testing.T) {
	var workflow glue.Workflow

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_workflow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSGlueWorkflow(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueWorkflowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueWorkflowConfig_DefaultRunProperties(rName, "firstPropValue", "secondPropValue"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueWorkflowExists(resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, "default_run_properties.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "default_run_properties.--run-prop1", "firstPropValue"),
					resource.TestCheckResourceAttr(resourceName, "default_run_properties.--run-prop2", "secondPropValue"),
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

func TestAccAWSGlueWorkflow_Description(t *testing.T) {
	var workflow glue.Workflow

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_workflow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSGlueWorkflow(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueWorkflowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueWorkflowConfig_Description(rName, "First Description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueWorkflowExists(resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, "description", "First Description"),
				),
			},
			{
				Config: testAccAWSGlueWorkflowConfig_Description(rName, "Second Description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueWorkflowExists(resourceName, &workflow),
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

func TestAccAWSGlueWorkflow_Tags(t *testing.T) {
	var workflow glue.Workflow
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_workflow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSGlueWorkflow(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueWorkflowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueWorkflowConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueWorkflowExists(resourceName, &workflow),
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
				Config: testAccAWSGlueWorkflowConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueWorkflowExists(resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSGlueWorkflowConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueWorkflowExists(resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSGlueWorkflow_disappears(t *testing.T) {
	var workflow glue.Workflow

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_workflow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSGlueWorkflow(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueWorkflowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueWorkflowConfig_Required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueWorkflowExists(resourceName, &workflow),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsGlueWorkflow(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccPreCheckAWSGlueWorkflow(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).glueconn

	_, err := conn.ListWorkflows(&glue.ListWorkflowsInput{})

	// Some endpoints that do not support Glue Workflows return InternalFailure
	if testAccPreCheckSkipError(err) || isAWSErr(err, "InternalFailure", "") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckAWSGlueWorkflowExists(resourceName string, workflow *glue.Workflow) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Glue Workflow ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).glueconn

		output, err := conn.GetWorkflow(&glue.GetWorkflowInput{
			Name: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if output.Workflow == nil {
			return fmt.Errorf("Glue Workflow (%s) not found", rs.Primary.ID)
		}

		if aws.StringValue(output.Workflow.Name) == rs.Primary.ID {
			*workflow = *output.Workflow
			return nil
		}

		return fmt.Errorf("Glue Workflow (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckAWSGlueWorkflowDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_glue_workflow" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).glueconn

		output, err := conn.GetWorkflow(&glue.GetWorkflowInput{
			Name: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
				return nil
			}

		}

		workflow := output.Workflow
		if workflow != nil && aws.StringValue(workflow.Name) == rs.Primary.ID {
			return fmt.Errorf("Glue Workflow %s still exists", rs.Primary.ID)
		}

		return err
	}

	return nil
}

func testAccAWSGlueWorkflowConfig_DefaultRunProperties(rName, firstPropValue, secondPropValue string) string {
	return fmt.Sprintf(`
resource "aws_glue_workflow" "test" {
  name = "%s"

  default_run_properties = {
    "--run-prop1" = "%s"
    "--run-prop2" = "%s"
  }
}
`, rName, firstPropValue, secondPropValue)
}

func testAccAWSGlueWorkflowConfig_Description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_glue_workflow" "test" {
  description = "%s"
  name        = "%s"
}
`, description, rName)
}

func testAccAWSGlueWorkflowConfig_Required(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_workflow" "test" {
  name = "%s"
}
`, rName)
}

func testAccAWSGlueWorkflowConfigMaxConcurrentRuns(rName string, runs int) string {
	return fmt.Sprintf(`
resource "aws_glue_workflow" "test" {
  name                = %[1]q
  max_concurrent_runs = %[2]d
}
`, rName, runs)
}

func testAccAWSGlueWorkflowConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_glue_workflow" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSGlueWorkflowConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_glue_workflow" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
