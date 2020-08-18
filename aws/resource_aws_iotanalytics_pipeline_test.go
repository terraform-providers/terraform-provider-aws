package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iotanalytics"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSIoTAnalyticsPipeline_basic(t *testing.T) {
	rString := acctest.RandString(5)
	resourceName := "aws_iotanalytics_pipeline.pipeline"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTAnalyticsPipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTAnalyticsPipeline_basic(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTAnalyticsPipelineExists_basic("aws_iotanalytics_pipeline.pipeline"),
					resource.TestCheckResourceAttr("aws_iotanalytics_pipeline.pipeline", "name", fmt.Sprintf("test_pipeline_%s", rString)),
					resource.TestCheckResourceAttr("aws_iotanalytics_pipeline.pipeline", "tags.tagKey", "tagValue"),
					testAccCheckAWSIoTAnalyticsPipeline_basic(rString),
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

func testAccCheckAWSIoTAnalyticsPipeline_basic(rString string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).iotanalyticsconn
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iotanalytics_pipeline" {
				continue
			}

			params := &iotanalytics.DescribePipelineInput{
				PipelineName: aws.String(rs.Primary.ID),
			}
			out, err := conn.DescribePipeline(params)

			if err != nil {
				return err
			}

			pipeline := out.Pipeline
			expectedPipelineName := fmt.Sprintf("test_pipeline_%s", rString)
			if *pipeline.Name != expectedPipelineName {
				return fmt.Errorf("Pipeline Name %s is not equal expected name %s", *pipeline.Name, expectedPipelineName)
			}

			if len(pipeline.Activities) != 2 {
				return fmt.Errorf("Pipeline activities length %d is not equal expected length %d", len(pipeline.Activities), 2)
			}

			channelActivity := pipeline.Activities[0].Channel
			if channelActivity == nil {
				return fmt.Errorf("channelActivity is not expected to be nil")
			}
			expectedChannelActivityName := "channel_activity"
			if *channelActivity.Name != expectedChannelActivityName {
				return fmt.Errorf("channelActivity Name %s is not equal expected name %s", *channelActivity.Name, expectedChannelActivityName)
			}

			expectedChannelName := fmt.Sprintf("test_channel_%s", rString)
			if *channelActivity.ChannelName != expectedChannelName {
				return fmt.Errorf("channelActivity Channel Name %s is not equal expected channel name %s", *channelActivity.ChannelName, expectedChannelName)
			}

			expectedNextActivity := "datastore_activity"
			if *channelActivity.Next != expectedNextActivity {
				return fmt.Errorf("channelActivity Next %s is not equal expected next %s", *channelActivity.Next, expectedNextActivity)
			}

			datastoreActivity := pipeline.Activities[1].Datastore
			if datastoreActivity == nil {
				return fmt.Errorf("datastoreActivity is not expected to be nil")
			}

			expectedDatastoreActivityName := "datastore_activity"
			if *datastoreActivity.Name != expectedDatastoreActivityName {
				return fmt.Errorf("datastoreActivity Name %s is not equal expected name %s", *datastoreActivity.Name, expectedChannelActivityName)
			}

			expectedDatastoreName := fmt.Sprintf("test_datastore_%s", rString)
			if *datastoreActivity.DatastoreName != expectedDatastoreName {
				return fmt.Errorf("datastoreActivity Datastore Name %s is not equal expected channel name %s", *datastoreActivity.DatastoreName, expectedDatastoreName)
			}
		}
		return nil
	}
}

func TestAccAWSIoTAnalyticsPipeline_addAttributes(t *testing.T) {
	rString := acctest.RandString(5)
	resourceName := "aws_iotanalytics_pipeline.pipeline"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTAnalyticsPipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTAnalyticsPipeline_addAttributes(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTAnalyticsPipelineExists_basic("aws_iotanalytics_pipeline.pipeline"),
					resource.TestCheckResourceAttr("aws_iotanalytics_pipeline.pipeline", "name", fmt.Sprintf("test_pipeline_%s", rString)),
					testAccCheckAWSIoTAnalyticsPipeline_addAttributes(rString),
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

func testAccCheckAWSIoTAnalyticsPipeline_addAttributes(rString string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).iotanalyticsconn
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iotanalytics_pipeline" {
				continue
			}

			params := &iotanalytics.DescribePipelineInput{
				PipelineName: aws.String(rs.Primary.ID),
			}
			out, err := conn.DescribePipeline(params)

			if err != nil {
				return err
			}

			pipeline := out.Pipeline
			expectedPipelineName := fmt.Sprintf("test_pipeline_%s", rString)
			if *pipeline.Name != expectedPipelineName {
				return fmt.Errorf("Pipeline Name %s is not equal expected name %s", *pipeline.Name, expectedPipelineName)
			}

			if len(pipeline.Activities) != 3 {
				return fmt.Errorf("Pipeline activities length %d is not equal expected length %d", len(pipeline.Activities), 3)
			}

			addAttrActivity := pipeline.Activities[1].AddAttributes
			if addAttrActivity == nil {
				return fmt.Errorf("addAttrActivity is not expected to be nil")
			}
			expectedActivityName := "add_attrs_activity"
			if *addAttrActivity.Name != expectedActivityName {
				return fmt.Errorf("addAttrActivity Name %s is not equal expected name %s", *addAttrActivity.Name, expectedActivityName)
			}

			expectedAttrKey := "key"
			expectedAttrValue := "value"
			value, ok := addAttrActivity.Attributes[expectedAttrKey]
			if !ok {
				return fmt.Errorf("addAttrActivity.Attributes has no attribute under key `%s`", expectedAttrKey)
			}
			if *value != expectedAttrValue {
				return fmt.Errorf("addAttrActivity.Attributes value under key '%s' %s is not equal expected value %s", expectedAttrKey, *value, expectedAttrValue)
			}

			expectedNextActivity := "datastore_activity"
			if *addAttrActivity.Next != expectedNextActivity {
				return fmt.Errorf("addAttrActivity Next %s is not equal expected next %s", *addAttrActivity.Next, expectedNextActivity)
			}
		}
		return nil
	}
}

func TestAccAWSIoTAnalyticsPipeline_removeAttributes(t *testing.T) {
	rString := acctest.RandString(5)
	resourceName := "aws_iotanalytics_pipeline.pipeline"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTAnalyticsPipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTAnalyticsPipeline_removeAttributes(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTAnalyticsPipelineExists_basic("aws_iotanalytics_pipeline.pipeline"),
					resource.TestCheckResourceAttr("aws_iotanalytics_pipeline.pipeline", "name", fmt.Sprintf("test_pipeline_%s", rString)),
					testAccCheckAWSIoTAnalyticsPipeline_removeAttributes(rString),
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

func testAccCheckAWSIoTAnalyticsPipeline_removeAttributes(rString string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).iotanalyticsconn
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iotanalytics_pipeline" {
				continue
			}

			params := &iotanalytics.DescribePipelineInput{
				PipelineName: aws.String(rs.Primary.ID),
			}
			out, err := conn.DescribePipeline(params)

			if err != nil {
				return err
			}

			pipeline := out.Pipeline
			expectedPipelineName := fmt.Sprintf("test_pipeline_%s", rString)
			if *pipeline.Name != expectedPipelineName {
				return fmt.Errorf("Pipeline Name %s is not equal expected name %s", *pipeline.Name, expectedPipelineName)
			}

			if len(pipeline.Activities) != 3 {
				return fmt.Errorf("Pipeline activities length %d is not equal expected length %d", len(pipeline.Activities), 3)
			}

			removeAttrActivity := pipeline.Activities[1].RemoveAttributes
			if removeAttrActivity == nil {
				return fmt.Errorf("removeAttrActivity is not expected to be nil")
			}
			expectedActivityName := "remove_attrs_activity"
			if *removeAttrActivity.Name != expectedActivityName {
				return fmt.Errorf("removeAttrActivity Name %s is not equal expected name %s", *removeAttrActivity.Name, expectedActivityName)
			}

			if len(removeAttrActivity.Attributes) != 1 {
				return fmt.Errorf("removeAttrActivity Attributes length %d is not equal expected length %d", len(removeAttrActivity.Attributes), 1)
			}

			expectedAttributeName := "key"
			if *removeAttrActivity.Attributes[0] != expectedAttributeName {
				return fmt.Errorf("removeAttrActivity.Attributes[0] %s is not equal expected name %s", *removeAttrActivity.Attributes[0], expectedAttributeName)
			}

			expectedNextActivity := "datastore_activity"
			if *removeAttrActivity.Next != expectedNextActivity {
				return fmt.Errorf("removeAttrActivity Next %s is not equal expected next %s", *removeAttrActivity.Next, expectedNextActivity)
			}
		}
		return nil
	}
}

func TestAccAWSIoTAnalyticsPipeline_selectAttributes(t *testing.T) {
	rString := acctest.RandString(5)
	resourceName := "aws_iotanalytics_pipeline.pipeline"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTAnalyticsPipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTAnalyticsPipeline_selectAttributes(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTAnalyticsPipelineExists_basic("aws_iotanalytics_pipeline.pipeline"),
					resource.TestCheckResourceAttr("aws_iotanalytics_pipeline.pipeline", "name", fmt.Sprintf("test_pipeline_%s", rString)),
					testAccCheckAWSIoTAnalyticsPipeline_selectAttributes(rString),
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

func testAccCheckAWSIoTAnalyticsPipeline_selectAttributes(rString string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).iotanalyticsconn
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iotanalytics_pipeline" {
				continue
			}

			params := &iotanalytics.DescribePipelineInput{
				PipelineName: aws.String(rs.Primary.ID),
			}
			out, err := conn.DescribePipeline(params)

			if err != nil {
				return err
			}

			pipeline := out.Pipeline
			expectedPipelineName := fmt.Sprintf("test_pipeline_%s", rString)
			if *pipeline.Name != expectedPipelineName {
				return fmt.Errorf("Pipeline Name %s is not equal expected name %s", *pipeline.Name, expectedPipelineName)
			}

			if len(pipeline.Activities) != 3 {
				return fmt.Errorf("Pipeline activities length %d is not equal expected length %d", len(pipeline.Activities), 3)
			}

			selectAttrActivity := pipeline.Activities[1].SelectAttributes
			if selectAttrActivity == nil {
				return fmt.Errorf("selectAttrActivity is not expected to be nil")
			}
			expectedActivityName := "select_attrs_activity"
			if *selectAttrActivity.Name != expectedActivityName {
				return fmt.Errorf("selectAttrActivity Name %s is not equal expected name %s", *selectAttrActivity.Name, expectedActivityName)
			}

			if len(selectAttrActivity.Attributes) != 1 {
				return fmt.Errorf("selectAttrActivity Attributes length %d is not equal expected length %d", len(selectAttrActivity.Attributes), 1)
			}

			expectedAttributeName := "key"
			if *selectAttrActivity.Attributes[0] != expectedAttributeName {
				return fmt.Errorf("selectAttrActivity.Attributes[0] %s is not equal expected name %s", *selectAttrActivity.Attributes[0], expectedAttributeName)
			}

			expectedNextActivity := "datastore_activity"
			if *selectAttrActivity.Next != expectedNextActivity {
				return fmt.Errorf("selectAttrActivity Next %s is not equal expected next %s", *selectAttrActivity.Next, expectedNextActivity)
			}
		}
		return nil
	}
}

func TestAccAWSIoTAnalyticsPipeline_deviceRegistryEnrich(t *testing.T) {
	rString := acctest.RandString(5)
	resourceName := "aws_iotanalytics_pipeline.pipeline"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTAnalyticsPipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTAnalyticsPipeline_deviceRegistryEnrich(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTAnalyticsPipelineExists_basic("aws_iotanalytics_pipeline.pipeline"),
					resource.TestCheckResourceAttr("aws_iotanalytics_pipeline.pipeline", "name", fmt.Sprintf("test_pipeline_%s", rString)),
					testAccCheckAWSIoTAnalyticsPipeline_deviceRegistryEnrich(rString),
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

func testAccCheckAWSIoTAnalyticsPipeline_deviceRegistryEnrich(rString string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).iotanalyticsconn
		iamconn := testAccProvider.Meta().(*AWSClient).iamconn
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iotanalytics_pipeline" {
				continue
			}

			params := &iotanalytics.DescribePipelineInput{
				PipelineName: aws.String(rs.Primary.ID),
			}
			out, err := conn.DescribePipeline(params)

			if err != nil {
				return err
			}

			pipeline := out.Pipeline
			expectedPipelineName := fmt.Sprintf("test_pipeline_%s", rString)
			if *pipeline.Name != expectedPipelineName {
				return fmt.Errorf("Pipeline Name %s is not equal expected name %s", *pipeline.Name, expectedPipelineName)
			}

			if len(pipeline.Activities) != 3 {
				return fmt.Errorf("Pipeline activities length %d is not equal expected length %d", len(pipeline.Activities), 3)
			}

			deviceRegistryEnrichActivity := pipeline.Activities[1].DeviceRegistryEnrich
			if deviceRegistryEnrichActivity == nil {
				return fmt.Errorf("deviceRegistryEnrichActivity is not expected to be nil")
			}
			expectedActivityName := "device_registry_enrich_activity"
			if *deviceRegistryEnrichActivity.Name != expectedActivityName {
				return fmt.Errorf("deviceRegistryEnrichActivity Name %s is not equal expected %s", *deviceRegistryEnrichActivity.Name, expectedActivityName)
			}

			expectedAttribute := "test_attribute"
			if *deviceRegistryEnrichActivity.Attribute != expectedAttribute {
				return fmt.Errorf("deviceRegistryEnrichActivity Attribute %s is not equal expected %s", *deviceRegistryEnrichActivity.Attribute, expectedAttribute)
			}

			expectedThingName := fmt.Sprintf("test_thing_%s", rString)
			if *deviceRegistryEnrichActivity.ThingName != expectedThingName {
				return fmt.Errorf("deviceRegistryEnrichActivity ThingName %s is not equal expected %s", *deviceRegistryEnrichActivity.Attribute, expectedThingName)
			}

			roleName := fmt.Sprintf("test_role_%s", rString)
			getRoleParams := &iam.GetRoleInput{
				RoleName: aws.String(roleName),
			}
			iamOut, _ := iamconn.GetRole(getRoleParams)
			if *deviceRegistryEnrichActivity.RoleArn != *iamOut.Role.Arn {
				return fmt.Errorf("deviceRegistryEnrichActivity RoleArn %s is not equal expected %s", *deviceRegistryEnrichActivity.RoleArn, *iamOut.Role.Arn)
			}

			expectedNextActivity := "datastore_activity"
			if *deviceRegistryEnrichActivity.Next != expectedNextActivity {
				return fmt.Errorf("deviceRegistryEnrichActivity Next %s is not equal expected %s", *deviceRegistryEnrichActivity.Next, expectedNextActivity)
			}
		}
		return nil
	}
}

func TestAccAWSIoTAnalyticsPipeline_deviceShadowEnrich(t *testing.T) {
	rString := acctest.RandString(5)
	resourceName := "aws_iotanalytics_pipeline.pipeline"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTAnalyticsPipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTAnalyticsPipeline_deviceShadowEnrich(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTAnalyticsPipelineExists_basic("aws_iotanalytics_pipeline.pipeline"),
					resource.TestCheckResourceAttr("aws_iotanalytics_pipeline.pipeline", "name", fmt.Sprintf("test_pipeline_%s", rString)),
					testAccCheckAWSIoTAnalyticsPipeline_deviceShadowEnrich(rString),
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

func testAccCheckAWSIoTAnalyticsPipeline_deviceShadowEnrich(rString string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).iotanalyticsconn
		iamconn := testAccProvider.Meta().(*AWSClient).iamconn
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iotanalytics_pipeline" {
				continue
			}

			params := &iotanalytics.DescribePipelineInput{
				PipelineName: aws.String(rs.Primary.ID),
			}
			out, err := conn.DescribePipeline(params)

			if err != nil {
				return err
			}

			pipeline := out.Pipeline
			expectedPipelineName := fmt.Sprintf("test_pipeline_%s", rString)
			if *pipeline.Name != expectedPipelineName {
				return fmt.Errorf("Pipeline Name %s is not equal expected name %s", *pipeline.Name, expectedPipelineName)
			}

			if len(pipeline.Activities) != 3 {
				return fmt.Errorf("Pipeline activities length %d is not equal expected length %d", len(pipeline.Activities), 3)
			}

			deviceShadowEnrichActivity := pipeline.Activities[1].DeviceShadowEnrich
			if deviceShadowEnrichActivity == nil {
				return fmt.Errorf("deviceShadowEnrichActivity is not expected to be nil")
			}
			expectedActivityName := "device_shadow_enrich_activity"
			if *deviceShadowEnrichActivity.Name != expectedActivityName {
				return fmt.Errorf("deviceShadowEnrichActivity Name %s is not equal expected %s", *deviceShadowEnrichActivity.Name, expectedActivityName)
			}

			expectedAttribute := "test_attribute"
			if *deviceShadowEnrichActivity.Attribute != expectedAttribute {
				return fmt.Errorf("deviceShadowEnrichActivity Attribute %s is not equal expected %s", *deviceShadowEnrichActivity.Attribute, expectedAttribute)
			}

			expectedThingName := fmt.Sprintf("test_thing_%s", rString)
			if *deviceShadowEnrichActivity.ThingName != expectedThingName {
				return fmt.Errorf("deviceShadowEnrichActivity ThingName %s is not equal expected %s", *deviceShadowEnrichActivity.Attribute, expectedThingName)
			}

			roleName := fmt.Sprintf("test_role_%s", rString)
			getRoleParams := &iam.GetRoleInput{
				RoleName: aws.String(roleName),
			}
			iamOut, _ := iamconn.GetRole(getRoleParams)
			if *deviceShadowEnrichActivity.RoleArn != *iamOut.Role.Arn {
				return fmt.Errorf("deviceShadowEnrichActivity RoleArn %s is not equal expected %s", *deviceShadowEnrichActivity.RoleArn, *iamOut.Role.Arn)
			}

			expectedNextActivity := "datastore_activity"
			if *deviceShadowEnrichActivity.Next != expectedNextActivity {
				return fmt.Errorf("deviceShadowEnrichActivity Next %s is not equal expected %s", *deviceShadowEnrichActivity.Next, expectedNextActivity)
			}
		}
		return nil
	}
}

func TestAccCheckAWSIoTAnalyticsPipeline_filter(t *testing.T) {
	rString := acctest.RandString(5)
	resourceName := "aws_iotanalytics_pipeline.pipeline"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTAnalyticsPipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTAnalyticsPipeline_filter(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTAnalyticsPipelineExists_basic("aws_iotanalytics_pipeline.pipeline"),
					resource.TestCheckResourceAttr("aws_iotanalytics_pipeline.pipeline", "name", fmt.Sprintf("test_pipeline_%s", rString)),
					testAccCheckAWSIoTAnalyticsPipeline_filter(rString),
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

func testAccCheckAWSIoTAnalyticsPipeline_filter(rString string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).iotanalyticsconn
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iotanalytics_pipeline" {
				continue
			}

			params := &iotanalytics.DescribePipelineInput{
				PipelineName: aws.String(rs.Primary.ID),
			}
			out, err := conn.DescribePipeline(params)

			if err != nil {
				return err
			}

			pipeline := out.Pipeline
			expectedPipelineName := fmt.Sprintf("test_pipeline_%s", rString)
			if *pipeline.Name != expectedPipelineName {
				return fmt.Errorf("Pipeline Name %s is not equal expected name %s", *pipeline.Name, expectedPipelineName)
			}

			if len(pipeline.Activities) != 3 {
				return fmt.Errorf("Pipeline activities length %d is not equal expected length %d", len(pipeline.Activities), 3)
			}

			filterActivity := pipeline.Activities[1].Filter
			if filterActivity == nil {
				return fmt.Errorf("filterActivity is not expected to be nil")
			}
			expectedActivityName := "filter_activity"
			if *filterActivity.Name != expectedActivityName {
				return fmt.Errorf("filterActivity Name %s is not equal expected %s", *filterActivity.Name, expectedActivityName)
			}

			expectedFilter := "temp > 40 AND hum < 20"
			if *filterActivity.Filter != expectedFilter {
				return fmt.Errorf("filterActivity Filter Name %s is not equal expected %s", *filterActivity.Filter, expectedFilter)
			}

			expectedNextActivity := "datastore_activity"
			if *filterActivity.Next != expectedNextActivity {
				return fmt.Errorf("filterActivity Next %s is not equal expected %s", *filterActivity.Next, expectedNextActivity)
			}
		}
		return nil
	}
}

func TestAccAWSIoTAnalyticsPipeline_lambda(t *testing.T) {
	rString := acctest.RandString(5)
	resourceName := "aws_iotanalytics_pipeline.pipeline"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTAnalyticsPipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTAnalyticsPipeline_lambda(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTAnalyticsPipelineExists_basic("aws_iotanalytics_pipeline.pipeline"),
					resource.TestCheckResourceAttr("aws_iotanalytics_pipeline.pipeline", "name", fmt.Sprintf("test_pipeline_%s", rString)),
					testAccCheckAWSIoTAnalyticsPipeline_lambda(rString),
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

func testAccCheckAWSIoTAnalyticsPipeline_lambda(rString string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).iotanalyticsconn
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iotanalytics_pipeline" {
				continue
			}

			params := &iotanalytics.DescribePipelineInput{
				PipelineName: aws.String(rs.Primary.ID),
			}
			out, err := conn.DescribePipeline(params)

			if err != nil {
				return err
			}

			pipeline := out.Pipeline
			expectedPipelineName := fmt.Sprintf("test_pipeline_%s", rString)
			if *pipeline.Name != expectedPipelineName {
				return fmt.Errorf("Pipeline Name %s is not equal expected name %s", *pipeline.Name, expectedPipelineName)
			}

			if len(pipeline.Activities) != 3 {
				return fmt.Errorf("Pipeline activities length %d is not equal expected length %d", len(pipeline.Activities), 3)
			}

			lambdaActivity := pipeline.Activities[1].Lambda
			if lambdaActivity == nil {
				return fmt.Errorf("lambdaActivity is not expected to be nil")
			}
			expectedActivityName := "lambda_activity"
			if *lambdaActivity.Name != expectedActivityName {
				return fmt.Errorf("lambdaActivity Name %s is not equal expected %s", *lambdaActivity.Name, expectedActivityName)
			}

			expectedLambdaName := "test_lambda"
			if *lambdaActivity.LambdaName != expectedLambdaName {
				return fmt.Errorf("lambdaActivity LambdaName %s is not equal expected %s", *lambdaActivity.LambdaName, expectedLambdaName)
			}

			expectedBatchSize := int64(10)
			if *lambdaActivity.BatchSize != expectedBatchSize {
				return fmt.Errorf("lambdaActivity BatchSize %d is not equal expected %d", *lambdaActivity.BatchSize, expectedBatchSize)
			}

			expectedNextActivity := "datastore_activity"
			if *lambdaActivity.Next != expectedNextActivity {
				return fmt.Errorf("lambdaActivity Next %s is not equal expected %s", *lambdaActivity.Next, expectedNextActivity)
			}
		}
		return nil
	}
}

func TestAccAWSIoTAnalyticsPipeline_math(t *testing.T) {
	rString := acctest.RandString(5)
	resourceName := "aws_iotanalytics_pipeline.pipeline"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTAnalyticsPipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTAnalyticsPipeline_math(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTAnalyticsPipelineExists_basic("aws_iotanalytics_pipeline.pipeline"),
					resource.TestCheckResourceAttr("aws_iotanalytics_pipeline.pipeline", "name", fmt.Sprintf("test_pipeline_%s", rString)),
					testAccCheckAWSIoTAnalyticsPipeline_math(rString),
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

func testAccCheckAWSIoTAnalyticsPipeline_math(rString string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).iotanalyticsconn
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iotanalytics_pipeline" {
				continue
			}

			params := &iotanalytics.DescribePipelineInput{
				PipelineName: aws.String(rs.Primary.ID),
			}
			out, err := conn.DescribePipeline(params)

			if err != nil {
				return err
			}

			pipeline := out.Pipeline
			expectedPipelineName := fmt.Sprintf("test_pipeline_%s", rString)
			if *pipeline.Name != expectedPipelineName {
				return fmt.Errorf("Pipeline Name %s is not equal expected name %s", *pipeline.Name, expectedPipelineName)
			}

			if len(pipeline.Activities) != 3 {
				return fmt.Errorf("Pipeline activities length %d is not equal expected length %d", len(pipeline.Activities), 3)
			}

			mathActivity := pipeline.Activities[1].Math
			if mathActivity == nil {
				return fmt.Errorf("mathActivity is not expected to be nil")
			}
			expectedActivityName := "math_activity"
			if *mathActivity.Name != expectedActivityName {
				return fmt.Errorf("mathActivity Name %s is not equal expected %s", *mathActivity.Name, expectedActivityName)
			}

			expectedMath := "(tempF - 32) / 2"
			if *mathActivity.Math != expectedMath {
				return fmt.Errorf("mathActivity Math %s is not equal expected %s", *mathActivity.Math, expectedMath)
			}

			expectedAttribute := "test_attr"
			if *mathActivity.Attribute != expectedAttribute {
				return fmt.Errorf("mathActivity Attribute %s is not equal expected %s", *mathActivity.Attribute, expectedAttribute)
			}

			expectedNextActivity := "datastore_activity"
			if *mathActivity.Next != expectedNextActivity {
				return fmt.Errorf("mathActivity Next %s is not equal expected %s", *mathActivity.Next, expectedNextActivity)
			}
		}
		return nil
	}
}

func testAccCheckAWSIoTAnalyticsPipelineDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iotanalyticsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iotanalytics_pipeline.pipeline" {
			continue
		}

		params := &iotanalytics.DescribePipelineInput{
			PipelineName: aws.String(rs.Primary.ID),
		}
		_, err := conn.DescribePipeline(params)

		if err != nil {
			if isAWSErr(err, iotanalytics.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return err
		}

		return fmt.Errorf("Expected IoTAnalytics Pipeline to be destroyed, %s found", rs.Primary.ID)

	}

	return nil
}

func testAccCheckAWSIoTAnalyticsPipelineExists_basic(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

const testAccAWSIoTAnalyticsPipelineBasicConfig = `
resource "aws_iotanalytics_channel" "channel" {
	name = "test_channel_%[1]s"
	storage {
		service_managed_s3 {}
	}
	retention_period {
	  unlimited = true
	}
}

resource "aws_iotanalytics_datastore" "datastore" {
	name = "test_datastore_%[1]s"
	storage {
		service_managed_s3 {}
	}
	retention_period {
		unlimited = true
	}
}
`

func testAccAWSIoTAnalyticsPipeline_basic(rString string) string {
	return fmt.Sprintf(testAccAWSIoTAnalyticsPipelineBasicConfig+`
resource "aws_iotanalytics_pipeline" "pipeline" {
  name = "test_pipeline_%[1]s"

  tags = {
	  "tagKey" = "tagValue",
  }

  pipeline_activity {
	  channel {
		name = "channel_activity"
		channel_name = "${aws_iotanalytics_channel.channel.name}"
		next_activity = "datastore_activity"
	  }
  }

  pipeline_activity {
	datastore {
		name = "datastore_activity"
		datastore_name = "${aws_iotanalytics_datastore.datastore.name}"
	}
  }

}
`, rString)
}

func testAccAWSIoTAnalyticsPipeline_addAttributes(rString string) string {
	return fmt.Sprintf(testAccAWSIoTAnalyticsPipelineBasicConfig+`
resource "aws_iotanalytics_pipeline" "pipeline" {
  name = "test_pipeline_%[1]s"
 
  pipeline_activity {
	  channel {
		name = "channel_activity"
		channel_name = "${aws_iotanalytics_channel.channel.name}"
		next_activity = "add_attrs_activity"
	  }
  }

  pipeline_activity {
	add_attributes {
		name = "add_attrs_activity"
		next_activity = "datastore_activity"
		attributes = {"key" = "value"}
	}
  }

  pipeline_activity {
	datastore {
		name = "datastore_activity"
		datastore_name = "${aws_iotanalytics_datastore.datastore.name}"
	}
  }
}
`, rString)
}

func testAccAWSIoTAnalyticsPipeline_removeAttributes(rString string) string {
	return fmt.Sprintf(testAccAWSIoTAnalyticsPipelineBasicConfig+`
resource "aws_iotanalytics_pipeline" "pipeline" {
  name = "test_pipeline_%[1]s"
 
  pipeline_activity {
	  channel {
		name = "channel_activity"
		channel_name = "${aws_iotanalytics_channel.channel.name}"
		next_activity = "remove_attrs_activity"
	  }
  }

  pipeline_activity {
	remove_attributes {
		name = "remove_attrs_activity"
		next_activity = "datastore_activity"
		attributes = ["key"]
	}
  }

  pipeline_activity {
	datastore {
		name = "datastore_activity"
		datastore_name = "${aws_iotanalytics_datastore.datastore.name}"
	}
  }
}
`, rString)
}

func testAccAWSIoTAnalyticsPipeline_selectAttributes(rString string) string {
	return fmt.Sprintf(testAccAWSIoTAnalyticsPipelineBasicConfig+`
resource "aws_iotanalytics_pipeline" "pipeline" {
  name = "test_pipeline_%[1]s"
 
  pipeline_activity {
	  channel {
		name = "channel_activity"
		channel_name = "${aws_iotanalytics_channel.channel.name}"
		next_activity = "select_attrs_activity"
	  }
  }

  pipeline_activity {
	select_attributes {
		name = "select_attrs_activity"
		next_activity = "datastore_activity"
		attributes = ["key"]
	}
  }

  pipeline_activity {
	datastore {
		name = "datastore_activity"
		datastore_name = "${aws_iotanalytics_datastore.datastore.name}"
	}
  }
}
`, rString)
}

func testAccAWSIoTAnalyticsPipeline_deviceRegistryEnrich(rString string) string {
	return fmt.Sprintf(testAccAWSIoTAnalyticsPipelineBasicConfig+`
resource "aws_iam_role" "iotanalytics_role" {
    name = "test_role_%[1]s"
    assume_role_policy = <<EOF
{
    "Version":"2012-10-17",
    "Statement":[{
        "Effect": "Allow",
        "Principal": {
            "Service": "iotanalytics.amazonaws.com"
        },
        "Action": "sts:AssumeRole"
    }]
}
EOF
}
resource "aws_iam_policy" "policy" {
    name = "test_policy_%[1]s"
    path = "/"
    description = "My test policy"
    policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Action": "*",
        "Resource": "*"
    }]
}
EOF
}
resource "aws_iam_policy_attachment" "attach_policy" {
    name = "test_policy_attachment_%[1]s"
    roles = ["${aws_iam_role.iotanalytics_role.name}"]
    policy_arn = "${aws_iam_policy.policy.arn}"
}

resource "aws_iotanalytics_pipeline" "pipeline" {
  name = "test_pipeline_%[1]s"
 
  pipeline_activity {
	  channel {
		name = "channel_activity"
		channel_name = "${aws_iotanalytics_channel.channel.name}"
		next_activity = "device_registry_enrich_activity"
	  }
  }

  pipeline_activity {
	device_registry_enrich {
		name = "device_registry_enrich_activity"
		next_activity = "datastore_activity"
		attribute = "test_attribute"
		role_arn = "${aws_iam_role.iotanalytics_role.arn}"
		thing_name = "test_thing_%[1]s"
	}
  }

  pipeline_activity {
	datastore {
		name = "datastore_activity"
		datastore_name = "${aws_iotanalytics_datastore.datastore.name}"
	}
  }
}
`, rString)
}

func testAccAWSIoTAnalyticsPipeline_deviceShadowEnrich(rString string) string {
	return fmt.Sprintf(testAccAWSIoTAnalyticsPipelineBasicConfig+`

resource "aws_iam_role" "iotanalytics_role" {
    name = "test_role_%[1]s"
    assume_role_policy = <<EOF
{
    "Version":"2012-10-17",
    "Statement":[{
        "Effect": "Allow",
        "Principal": {
            "Service": "iotanalytics.amazonaws.com"
        },
        "Action": "sts:AssumeRole"
    }]
}
EOF
}
resource "aws_iam_policy" "policy" {
    name = "test_policy_%[1]s"
    path = "/"
    description = "My test policy"
    policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Action": "*",
        "Resource": "*"
    }]
}
EOF
}
resource "aws_iam_policy_attachment" "attach_policy" {
    name = "test_policy_attachment_%[1]s"
    roles = ["${aws_iam_role.iotanalytics_role.name}"]
    policy_arn = "${aws_iam_policy.policy.arn}"
}

resource "aws_iotanalytics_pipeline" "pipeline" {
  name = "test_pipeline_%[1]s"
 
  pipeline_activity {
	  channel {
		name = "channel_activity"
		channel_name = "${aws_iotanalytics_channel.channel.name}"
		next_activity = "device_shadow_enrich_activity"
	  }
  }

  pipeline_activity {
	device_shadow_enrich {
		name = "device_shadow_enrich_activity"
		next_activity = "datastore_activity"
		attribute = "test_attribute"
		role_arn = "${aws_iam_role.iotanalytics_role.arn}"
		thing_name = "test_thing_%[1]s"
	}
  }

  pipeline_activity {
	datastore {
		name = "datastore_activity"
		datastore_name = "${aws_iotanalytics_datastore.datastore.name}"
	}
  }
}
`, rString)
}

func testAccAWSIoTAnalyticsPipeline_filter(rString string) string {
	return fmt.Sprintf(testAccAWSIoTAnalyticsPipelineBasicConfig+`
resource "aws_iotanalytics_pipeline" "pipeline" {
  name = "test_pipeline_%[1]s"
 
  pipeline_activity {
	  channel {
		name = "channel_activity"
		channel_name = "${aws_iotanalytics_channel.channel.name}"
		next_activity = "filter_activity"
	  }
  }

  pipeline_activity {
	filter {
		name = "filter_activity"
		filter = "temp > 40 AND hum < 20"
		next_activity = "datastore_activity"
	}
  }

  pipeline_activity {
	datastore {
		name = "datastore_activity"
		datastore_name = "${aws_iotanalytics_datastore.datastore.name}"
	}
  }
}
`, rString)
}

func testAccAWSIoTAnalyticsPipeline_lambda(rString string) string {
	return fmt.Sprintf(testAccAWSIoTAnalyticsPipelineBasicConfig+`
resource "aws_iotanalytics_pipeline" "pipeline" {
  name = "test_pipeline_%[1]s"
 
  pipeline_activity {
	  channel {
		name = "channel_activity"
		channel_name = "${aws_iotanalytics_channel.channel.name}"
		next_activity = "lambda_activity"
	  }
  }

  pipeline_activity {
	lambda {
		name = "lambda_activity"
		lambda_name = "test_lambda"
		batch_size = 10
		next_activity = "datastore_activity"
	}
  }

  pipeline_activity {
	datastore {
		name = "datastore_activity"
		datastore_name = "${aws_iotanalytics_datastore.datastore.name}"
	}
  }
}
`, rString)
}

func testAccAWSIoTAnalyticsPipeline_math(rString string) string {
	return fmt.Sprintf(testAccAWSIoTAnalyticsPipelineBasicConfig+`
resource "aws_iotanalytics_pipeline" "pipeline" {
  name = "test_pipeline_%[1]s"
 
  pipeline_activity {
	  channel {
		name = "channel_activity"
		channel_name = "${aws_iotanalytics_channel.channel.name}"
		next_activity = "math_activity"
	  }
  }

  pipeline_activity {
	math {
		name = "math_activity"
		math = "(tempF - 32) / 2"
		attribute = "test_attr"
		next_activity = "datastore_activity"
	}
  }

  pipeline_activity {
	datastore {
		name = "datastore_activity"
		datastore_name = "${aws_iotanalytics_datastore.datastore.name}"
	}
  }
}
`, rString)
}
