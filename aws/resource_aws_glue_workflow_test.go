package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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

func TestAccAWSGlueWorkflow_Basic(t *testing.T) {
	var workflow glue.Workflow

	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
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
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccAWSGlueWorkflow_DefaultRunProperties(t *testing.T) {
	var workflow glue.Workflow

	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
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

	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
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
 name               = "%s"

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
 description        = "%s"
 name               = "%s"
}
`, description, rName)
}

func testAccAWSGlueWorkflowConfig_Required(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_workflow" "test" {
 name               = "%s"
}
`, rName)
}
