package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datapipeline"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSDataPipelinePipeline_basic(t *testing.T) {
	var conf1, conf2 datapipeline.PipelineDescription
	rName1 := fmt.Sprintf("tf-datapipeline-%s", acctest.RandString(5))
	rName2 := fmt.Sprintf("tf-datapipeline-%s", acctest.RandString(5))
	resourceName := "aws_datapipeline_pipeline.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataPipeline(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataPipelinePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataPipelinePipelineConfig(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataPipelinePipelineExists(resourceName, &conf1),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
				),
			},
			{
				Config: testAccAWSDataPipelinePipelineConfig(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataPipelinePipelineExists(resourceName, &conf2),
					testAccCheckAWSDataPipelinePipelineNotEqual(&conf1, &conf2),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
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

func TestAccAWSDataPipelinePipeline_description(t *testing.T) {
	var conf1, conf2 datapipeline.PipelineDescription
	rName := fmt.Sprintf("tf-datapipeline-%s", acctest.RandString(5))
	resourceName := "aws_datapipeline_pipeline.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataPipeline(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataPipelinePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataPipelinePipelineConfigWithDescription(rName, "test description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataPipelinePipelineExists(resourceName, &conf1),
					resource.TestCheckResourceAttr(resourceName, "description", "test description"),
				),
			},
			{
				Config: testAccAWSDataPipelinePipelineConfigWithDescription(rName, "update description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataPipelinePipelineExists(resourceName, &conf2),
					testAccCheckAWSDataPipelinePipelineNotEqual(&conf1, &conf2),
					resource.TestCheckResourceAttr(resourceName, "description", "update description"),
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

func TestAccAWSDataPipelinePipeline_disappears(t *testing.T) {
	var conf datapipeline.PipelineDescription
	rName := fmt.Sprintf("tf-datapipeline-%s", acctest.RandString(5))
	resourceName := "aws_datapipeline_pipeline.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataPipeline(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataPipelinePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataPipelinePipelineConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataPipelinePipelineExists(resourceName, &conf),
					testAccCheckAWSDataPipelinePipelineDisappears(&conf),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSDataPipelinePipeline_tags(t *testing.T) {
	var conf datapipeline.PipelineDescription
	rName := fmt.Sprintf("tf-datapipeline-%s", acctest.RandString(5))
	resourceName := "aws_datapipeline_pipeline.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataPipeline(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataPipelinePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataPipelinePipelineConfigWithTags(rName, "foo", "bar", "fizz", "buzz"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataPipelinePipelineExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.fizz", "buzz"),
				),
			},
			{
				Config: testAccAWSDataPipelinePipelineConfigWithTags(rName, "foo", "bar2", "fizz2", "buzz2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataPipelinePipelineExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.foo", "bar2"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.fizz2", "buzz2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDataPipelinePipelineConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataPipelinePipelineExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccCheckAWSDataPipelinePipelineDisappears(conf *datapipeline.PipelineDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).datapipelineconn
		params := &datapipeline.DeletePipelineInput{
			PipelineId: conf.PipelineId,
		}

		_, err := conn.DeletePipeline(params)
		if err != nil {
			return err
		}
		return waitForDataPipelineDeletion(conn, *conf.PipelineId)
	}
}

func testAccCheckAWSDataPipelinePipelineDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).datapipelineconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_datapipeline_pipeline" {
			continue
		}
		// Try to find the Pipeline
		pipelineDescription, err := resourceAwsDataPipelinePipelineRetrieve(rs.Primary.ID, conn)
		if isAWSErr(err, datapipeline.ErrCodePipelineNotFoundException, "") {
			continue
		} else if isAWSErr(err, datapipeline.ErrCodePipelineDeletedException, "") {
			continue
		}

		if err != nil {
			return err
		}
		if pipelineDescription != nil {
			return fmt.Errorf("Pipeline still exists")
		}
	}

	return nil
}

func testAccCheckAWSDataPipelinePipelineExists(n string, v *datapipeline.PipelineDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DataPipeline ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).datapipelineconn

		pipelineDescription, err := resourceAwsDataPipelinePipelineRetrieve(rs.Primary.ID, conn)

		if err != nil {
			return err
		}
		if pipelineDescription == nil {
			return fmt.Errorf("DataPipeline not found")
		}

		*v = *pipelineDescription
		return nil
	}
}

func testAccPreCheckAWSDataPipeline(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).datapipelineconn

	input := &datapipeline.ListPipelinesInput{}

	_, err := conn.ListPipelines(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckAWSDataPipelinePipelineNotEqual(pipeline1, pipeline2 *datapipeline.PipelineDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(pipeline1.PipelineId) == aws.StringValue(pipeline2.PipelineId) {
			return fmt.Errorf("Pipeline IDs are equal")
		}

		return nil
	}
}

func testAccAWSDataPipelinePipelineConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_datapipeline_pipeline" "default" {
  name = "%[1]s"
}`, rName)

}

func testAccAWSDataPipelinePipelineConfigWithDescription(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_datapipeline_pipeline" "default" {
  name        = "%[1]s"
  description = %[2]q
}`, rName, description)

}

func testAccAWSDataPipelinePipelineConfigWithTags(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_datapipeline_pipeline" "default" {
  name = "%[1]s"

  tags = {
    %[2]s = %[3]q
    %[4]s = %[5]q
  }
}`, rName, tagKey1, tagValue1, tagKey2, tagValue2)

}
