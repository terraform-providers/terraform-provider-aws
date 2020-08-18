package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iotanalytics"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSIoTAnalyticsDataset_basic(t *testing.T) {
	rString := acctest.RandString(5)
	resourceName := "aws_iotanalytics_dataset.dataset"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTAnalyticsDatasetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTAnalyticsDataset_basic(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTAnalyticsDatasetExists_basic("aws_iotanalytics_dataset.dataset"),
					resource.TestCheckResourceAttr("aws_iotanalytics_dataset.dataset", "name", fmt.Sprintf("test_dataset_%s", rString)),
					resource.TestCheckResourceAttr("aws_iotanalytics_dataset.dataset", "tags.tagKey", "tagValue"),
					testAccCheckAWSIoTAnalyticsDataset_basic(rString),
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

func testAccCheckAWSIoTAnalyticsDataset_basic(rString string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).iotanalyticsconn
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iotanalytics_dataset" {
				continue
			}

			params := &iotanalytics.DescribeDatasetInput{
				DatasetName: aws.String(rs.Primary.ID),
			}
			out, err := conn.DescribeDataset(params)

			if err != nil {
				return err
			}

			dataset := out.Dataset

			action := dataset.Actions[0]
			expectedActionName := "test_action"

			if *action.ActionName != expectedActionName {
				return fmt.Errorf("Expected action.ActionName %s is not equal to %s", expectedActionName, *action.ActionName)
			}

			if action.QueryAction == nil {
				return fmt.Errorf("Expected action.QueryAction is not nil")
			}

			if action.ContainerAction != nil {
				return fmt.Errorf("Expected action.ContainerAction is nil")
			}

			queryAction := action.QueryAction
			expectedSQLQuery := fmt.Sprintf("select * from test_datastore_%s", rString)

			if *queryAction.SqlQuery != expectedSQLQuery {
				return fmt.Errorf("Expected queryAction.SqlQuery %s is not equal to %s", expectedSQLQuery, *queryAction.SqlQuery)
			}

			filters := queryAction.Filters
			if len(filters) != 1 {
				return fmt.Errorf("Expected queryAction.Filters len %d is not equal to %d", 1, len(filters))
			}

			queryFilter := filters[0]
			expectedOffset := int64(30)
			if *queryFilter.DeltaTime.OffsetSeconds != expectedOffset {
				return fmt.Errorf("Expected queryFilter.DeltaTime.OffsetSeconds %d is not equal to %d", expectedOffset, *queryFilter.DeltaTime.OffsetSeconds)
			}

			expectedTimeExpression := "date"
			if *queryFilter.DeltaTime.TimeExpression != expectedTimeExpression {
				return fmt.Errorf("Expected queryFilter.DeltaTime.TimeExpression %s is not equal to %s", expectedTimeExpression, *queryFilter.DeltaTime.TimeExpression)
			}
		}
		return nil
	}
}

func TestAccAWSIoTAnalyticsDataset_triggerSchedule(t *testing.T) {
	rString := acctest.RandString(5)
	resourceName := "aws_iotanalytics_dataset.dataset"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTAnalyticsDatasetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTAnalyticsDataset_triggerSchedule(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTAnalyticsDatasetExists_basic("aws_iotanalytics_dataset.dataset"),
					resource.TestCheckResourceAttr("aws_iotanalytics_dataset.dataset", "name", fmt.Sprintf("test_dataset_%s", rString)),
					testAccCheckAWSIoTAnalyticsDataset_triggerSchedule(rString),
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

func testAccCheckAWSIoTAnalyticsDataset_triggerSchedule(rString string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).iotanalyticsconn
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iotanalytics_dataset" {
				continue
			}

			params := &iotanalytics.DescribeDatasetInput{
				DatasetName: aws.String(rs.Primary.ID),
			}
			out, err := conn.DescribeDataset(params)

			if err != nil {
				return err
			}

			dataset := out.Dataset

			if len(dataset.Triggers) != 1 {
				return fmt.Errorf("Expected 1 elements in dataset.Triggers, but not %d ", len(dataset.Triggers))
			}
			trigger := dataset.Triggers[0]

			if trigger.Dataset != nil {
				return fmt.Errorf("Expected trigger.Dataset is not equal nil")
			}

			expectedScheduleExpression := "cron(0 12 * * ? *)"

			if *trigger.Schedule.Expression != expectedScheduleExpression {
				return fmt.Errorf("Expected trigger.Schedule.Expression %s is not equal to %s", expectedScheduleExpression, *trigger.Schedule.Expression)
			}
		}
		return nil
	}
}

func TestAccAWSIoTAnalyticsDataset_retentionPeriodNumberOfDays(t *testing.T) {
	rString := acctest.RandString(5)
	resourceName := "aws_iotanalytics_dataset.dataset"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTAnalyticsDatasetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTAnalyticsDataset_retentionPeriodNumberOfDays(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTAnalyticsDatasetExists_basic("aws_iotanalytics_dataset.dataset"),
					resource.TestCheckResourceAttr("aws_iotanalytics_dataset.dataset", "name", fmt.Sprintf("test_dataset_%s", rString)),
					testAccCheckAWSIoTAnalyticsDataset_retentionPeriodNumberOfDays(rString),
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

func testAccCheckAWSIoTAnalyticsDataset_retentionPeriodNumberOfDays(rString string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).iotanalyticsconn
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iotanalytics_dataset" {
				continue
			}

			params := &iotanalytics.DescribeDatasetInput{
				DatasetName: aws.String(rs.Primary.ID),
			}
			out, err := conn.DescribeDataset(params)

			if err != nil {
				return err
			}

			dataset := out.Dataset

			retentionPeriod := dataset.RetentionPeriod

			expectedNumberOfDays := int64(6)
			if *retentionPeriod.NumberOfDays != expectedNumberOfDays {
				return fmt.Errorf("Expected retentionPeriod.NumberOfDays %d is not equal to %d", expectedNumberOfDays, *retentionPeriod.NumberOfDays)
			}

			expectedUnlimited := false
			if *retentionPeriod.Unlimited != expectedUnlimited {
				return fmt.Errorf("Expected retentionPeriod.Unlimited %t is not equal to %t", expectedUnlimited, *retentionPeriod.Unlimited)
			}
		}
		return nil
	}
}

func TestAccAWSIoTAnalyticsDataset_retentionPeriodUnlimited(t *testing.T) {
	rString := acctest.RandString(5)
	resourceName := "aws_iotanalytics_dataset.dataset"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTAnalyticsDatasetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTAnalyticsDataset_retentionPeriodUnlimited(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTAnalyticsDatasetExists_basic("aws_iotanalytics_dataset.dataset"),
					resource.TestCheckResourceAttr("aws_iotanalytics_dataset.dataset", "name", fmt.Sprintf("test_dataset_%s", rString)),
					testAccCheckAWSIoTAnalyticsDataset_retentionPeriodUnlimited(rString),
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

func testAccCheckAWSIoTAnalyticsDataset_retentionPeriodUnlimited(rString string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).iotanalyticsconn
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iotanalytics_dataset" {
				continue
			}

			params := &iotanalytics.DescribeDatasetInput{
				DatasetName: aws.String(rs.Primary.ID),
			}
			out, err := conn.DescribeDataset(params)

			if err != nil {
				return err
			}

			dataset := out.Dataset

			retentionPeriod := dataset.RetentionPeriod

			var expectedNumberOfDays *int64
			if retentionPeriod.NumberOfDays != expectedNumberOfDays {
				return fmt.Errorf("Expected retentionPeriod.NumberOfDays %d id not equal to %d", expectedNumberOfDays, *retentionPeriod.NumberOfDays)
			}

			expectedUnlimited := true
			if *retentionPeriod.Unlimited != expectedUnlimited {
				return fmt.Errorf("Expected retentionPeriod.Unlimited %t is not equal to %t", expectedUnlimited, *retentionPeriod.Unlimited)
			}
		}
		return nil
	}
}

func TestAccAWSIoTAnalyticsDataset_versioningConfigurationMaxVersions(t *testing.T) {
	rString := acctest.RandString(5)
	resourceName := "aws_iotanalytics_dataset.dataset"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTAnalyticsDatasetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTAnalyticsDataset_versioningConfigurationMaxVersions(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTAnalyticsDatasetExists_basic("aws_iotanalytics_dataset.dataset"),
					resource.TestCheckResourceAttr("aws_iotanalytics_dataset.dataset", "name", fmt.Sprintf("test_dataset_%s", rString)),
					testAccCheckAWSIoTAnalyticsDataset_versioningConfigurationMaxVersions(rString),
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

func testAccCheckAWSIoTAnalyticsDataset_versioningConfigurationMaxVersions(rString string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).iotanalyticsconn
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iotanalytics_dataset" {
				continue
			}

			params := &iotanalytics.DescribeDatasetInput{
				DatasetName: aws.String(rs.Primary.ID),
			}
			out, err := conn.DescribeDataset(params)

			if err != nil {
				return err
			}

			dataset := out.Dataset

			versioningConfiguration := dataset.VersioningConfiguration

			expectedMaxVersions := int64(5)
			if *versioningConfiguration.MaxVersions != expectedMaxVersions {
				return fmt.Errorf("Expected versioningConfiguration.MaxVersions %d is not equal to %d", expectedMaxVersions, *versioningConfiguration.MaxVersions)
			}

			expectedUnlimited := false
			if *versioningConfiguration.Unlimited != expectedUnlimited {
				return fmt.Errorf("Expected versioningConfiguration.Unlimited %t is not equal to %t", expectedUnlimited, *versioningConfiguration.Unlimited)
			}
		}
		return nil
	}
}

func TestAccAWSIoTAnalyticsDataset_versioningConfigurationUnlimited(t *testing.T) {
	rString := acctest.RandString(5)
	resourceName := "aws_iotanalytics_dataset.dataset"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTAnalyticsDatasetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTAnalyticsDataset_versioningConfigurationUnlimited(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTAnalyticsDatasetExists_basic("aws_iotanalytics_dataset.dataset"),
					resource.TestCheckResourceAttr("aws_iotanalytics_dataset.dataset", "name", fmt.Sprintf("test_dataset_%s", rString)),
					testAccCheckAWSIoTAnalyticsDataset_versioningConfigurationUnlimited(rString),
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

func testAccCheckAWSIoTAnalyticsDataset_versioningConfigurationUnlimited(rString string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).iotanalyticsconn
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iotanalytics_dataset" {
				continue
			}

			params := &iotanalytics.DescribeDatasetInput{
				DatasetName: aws.String(rs.Primary.ID),
			}
			out, err := conn.DescribeDataset(params)

			if err != nil {
				return err
			}

			dataset := out.Dataset

			versioningConfiguration := dataset.VersioningConfiguration

			var expectedMaxVersions *int64
			if versioningConfiguration.MaxVersions != expectedMaxVersions {
				return fmt.Errorf("Expected versioningConfiguration.MaxVersions %d is not equal to %d", expectedMaxVersions, *versioningConfiguration.MaxVersions)
			}

			expectedUnlimited := true
			if *versioningConfiguration.Unlimited != expectedUnlimited {
				return fmt.Errorf("Expected versioningConfiguration.Unlimited %t is not equal to %t", expectedUnlimited, *versioningConfiguration.Unlimited)
			}
		}
		return nil
	}
}

func testAccCheckAWSIoTAnalyticsDatasetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iotanalyticsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iotanalytics_dataset.dataset" {
			continue
		}

		params := &iotanalytics.DescribeDatasetInput{
			DatasetName: aws.String(rs.Primary.ID),
		}
		_, err := conn.DescribeDataset(params)

		if err != nil {
			if isAWSErr(err, iotanalytics.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return err
		}

		return fmt.Errorf("Expected IoTAnalytics Dataset to be destroyed, %s found", rs.Primary.ID)

	}

	return nil
}

func testAccCheckAWSIoTAnalyticsDatasetExists_basic(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func TestAccAWSIoTAnalyticsDataset_contentDeliveryRuleS3Destination(t *testing.T) {
	rString := acctest.RandString(5)
	resourceName := "aws_iotanalytics_dataset.dataset"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTAnalyticsDatasetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTAnalyticsDataset_contentDeliveryRuleDestinationS3(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTAnalyticsDatasetExists_basic("aws_iotanalytics_dataset.dataset"),
					resource.TestCheckResourceAttr("aws_iotanalytics_dataset.dataset", "name", fmt.Sprintf("test_dataset_%s", rString)),
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

func TestAccAWSIoTAnalyticsDataset_contentDeliveryRuleIotEventsDestination(t *testing.T) {
	rString := acctest.RandString(5)
	resourceName := "aws_iotanalytics_dataset.dataset"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTAnalyticsDatasetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTAnalyticsDataset_contentDeliveryRuleDestinationIotEvents(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTAnalyticsDatasetExists_basic("aws_iotanalytics_dataset.dataset"),
					resource.TestCheckResourceAttr("aws_iotanalytics_dataset.dataset", "name", fmt.Sprintf("test_dataset_%s", rString)),
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

const testAccAWSIoTAnalyticsDatasetRole = `
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

func testAccAWSIoTAnalyticsDataset_basic(rString string) string {
	return fmt.Sprintf(testAccAWSIoTAnalyticsDatasetRole+`
resource "aws_iotanalytics_dataset" "dataset" {
  name = "test_dataset_%[1]s"

  tags = {
	  "tagKey" = "tagValue"
  }

  action {
	  name = "test_action"

	  query_action {

		filter {
			delta_time {
				offset_seconds = 30
				time_expression = "date"
			}
		}

		  sql_query = "select * from ${aws_iotanalytics_datastore.datastore.name}"
	  }
  }
}
`, rString)
}

func testAccAWSIoTAnalyticsDataset_triggerSchedule(rString string) string {
	return fmt.Sprintf(testAccAWSIoTAnalyticsDatasetRole+`
resource "aws_iotanalytics_dataset" "dataset" {
  name = "test_dataset_%[1]s"

  action {
	  name = "test_action"

	  query_action {

		filter {
			delta_time {
				offset_seconds = 30
				time_expression = "date"
			}
		}

		  sql_query = "select * from ${aws_iotanalytics_datastore.datastore.name}"
	  }
  }

  trigger {
	schedule {
		expression = "cron(0 12 * * ? *)"
	}
  }
}
`, rString)
}

func testAccAWSIoTAnalyticsDataset_retentionPeriodNumberOfDays(rString string) string {
	return fmt.Sprintf(testAccAWSIoTAnalyticsDatasetRole+`
resource "aws_iotanalytics_dataset" "dataset" {
  name = "test_dataset_%[1]s"

  action {
	  name = "test_action"

	  query_action {

		filter {
			delta_time {
				offset_seconds = 30
				time_expression = "date"
			}
		}

		  sql_query = "select * from ${aws_iotanalytics_datastore.datastore.name}"
	  }
  }

  retention_period {
	number_of_days = 6
  }
}
`, rString)
}

func testAccAWSIoTAnalyticsDataset_retentionPeriodUnlimited(rString string) string {
	return fmt.Sprintf(testAccAWSIoTAnalyticsDatasetRole+`
resource "aws_iotanalytics_dataset" "dataset" {
  name = "test_dataset_%[1]s"

  action {
	  name = "test_action"

	  query_action {

		filter {
			delta_time {
				offset_seconds = 30
				time_expression = "date"
			}
		}

		  sql_query = "select * from ${aws_iotanalytics_datastore.datastore.name}"
	  }
  }

  retention_period {
	unlimited = true
  }
}
`, rString)
}

func testAccAWSIoTAnalyticsDataset_versioningConfigurationMaxVersions(rString string) string {
	return fmt.Sprintf(testAccAWSIoTAnalyticsDatasetRole+`
resource "aws_iotanalytics_dataset" "dataset" {
  name = "test_dataset_%[1]s"

  action {
	  name = "test_action"

	  query_action {

		filter {
			delta_time {
				offset_seconds = 30
				time_expression = "date"
			}
		}

		  sql_query = "select * from ${aws_iotanalytics_datastore.datastore.name}"
	  }
  }

  versioning_configuration {
	max_versions = 5
  }

}
`, rString)
}

func testAccAWSIoTAnalyticsDataset_versioningConfigurationUnlimited(rString string) string {
	return fmt.Sprintf(testAccAWSIoTAnalyticsDatasetRole+`
resource "aws_iotanalytics_dataset" "dataset" {
  name = "test_dataset_%[1]s"

  action {
	  name = "test_action"

	  query_action {

		filter {
			delta_time {
				offset_seconds = 30
				time_expression = "date"
			}
		}

		  sql_query = "select * from ${aws_iotanalytics_datastore.datastore.name}"
	  }
  }
  versioning_configuration {
	unlimited = true
  }
}
`, rString)
}

func testAccAWSIoTAnalyticsDataset_contentDeliveryRuleDestinationIotEvents(rString string) string {
	return fmt.Sprintf(testAccAWSIoTAnalyticsDatasetRole+`

resource "aws_iotevents_input" "test_input" {
  name = "test_input_%[1]s"
  
  definition {
	  attribute {
		json_path = "temperature"
	  }

	  attribute {
		json_path = "humidity"
	  }
  }
}

resource "aws_iotanalytics_dataset" "dataset" {
  name = "test_dataset_%[1]s"

  action {
	  name = "test_action"

	  query_action {

		filter {
			delta_time {
				offset_seconds = 30
				time_expression = "date"
			}
		}

		  sql_query = "select * from ${aws_iotanalytics_datastore.datastore.name}"
	  }
  }

  content_delivery_rule {
      entry_name = ""
	  destination {
		iotevents_destination {
			input_name = "${aws_iotevents_input.test_input.name}"
			role_arn = "${aws_iam_role.iotanalytics_role.arn}"
		}
	  }
  }

}
`, rString)
}

func testAccAWSIoTAnalyticsDataset_contentDeliveryRuleDestinationS3(rString string) string {
	return fmt.Sprintf(testAccAWSIoTAnalyticsDatasetRole+`
resource "aws_iotanalytics_dataset" "dataset" {
  name = "test_dataset_%[1]s"

  action {
	  name = "test_action"

	  query_action {

		filter {
			delta_time {
				offset_seconds = 30
				time_expression = "date"
			}
		}

		  sql_query = "select * from ${aws_iotanalytics_datastore.datastore.name}"
	  }
  }

  content_delivery_rule {
	  entry_name = ""
	  destination {
		s3_destination {
			bucket = "test_bucket"
			key = "test_key"
			role_arn = "${aws_iam_role.iotanalytics_role.arn}"

			glue_configuration {
				database_name = "test_db_name"
				table_name = "test_table_name"
			}
		}
	  }
  }

}
`, rString)
}
