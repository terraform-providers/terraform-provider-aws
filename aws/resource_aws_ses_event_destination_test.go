package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSSESEventDestination_basic(t *testing.T) {
	rString := acctest.RandString(8)

	bucketName := fmt.Sprintf("tf-acc-bucket-ses-event-dst-%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_ses_event_dst_%s", rString)
	streamName := fmt.Sprintf("tf_acc_stream_ses_event_dst_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_ses_event_dst_%s", rString)
	topicName := fmt.Sprintf("tf_acc_topic_ses_event_dst_%s", rString)
	sesCfgSetName := fmt.Sprintf("tf_acc_cfg_ses_event_dst_%s", rString)
	sesEventDstNameKinesis := fmt.Sprintf("tf_acc_event_dst_kinesis_%s", rString)
	sesEventDstNameCw := fmt.Sprintf("tf_acc_event_dst_cloudwatch_%s", rString)
	sesEventDstNameSns := fmt.Sprintf("tf_acc_event_dst_sns_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSESEventDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSESEventDestinationConfig(bucketName, roleName, streamName, policyName, topicName,
					sesCfgSetName, sesEventDstNameKinesis, sesEventDstNameCw, sesEventDstNameSns),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESEventDestinationExists("aws_ses_configuration_set.test"),
					resource.TestCheckResourceAttr(
						"aws_ses_event_destination.kinesis", "name", sesEventDstNameKinesis),
					resource.TestCheckResourceAttr(
						"aws_ses_event_destination.cloudwatch", "name", sesEventDstNameCw),
					resource.TestCheckResourceAttr(
						"aws_ses_event_destination.sns", "name", sesEventDstNameSns),
				),
			},
		},
	})
}

func testAccCheckSESEventDestinationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sesConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ses_configuration_set" {
			continue
		}

		response, err := conn.ListConfigurationSets(&ses.ListConfigurationSetsInput{})
		if err != nil {
			return err
		}

		found := false
		for _, element := range response.ConfigurationSets {
			if *element.Name == rs.Primary.ID {
				found = true
			}
		}

		if found {
			return fmt.Errorf("The configuration set still exists")
		}

	}

	return nil

}

func testAccCheckAwsSESEventDestinationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES event destination not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES event destination ID not set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sesConn

		response, err := conn.ListConfigurationSets(&ses.ListConfigurationSetsInput{})
		if err != nil {
			return err
		}

		found := false
		for _, element := range response.ConfigurationSets {
			if *element.Name == rs.Primary.ID {
				found = true
			}
		}

		if !found {
			return fmt.Errorf("The configuration set was not created")
		}

		return nil
	}
}

func testAccAWSSESEventDestinationConfig(bucketName, roleName, streamName, policyName, topicName,
	sesCfgSetName, sesEventDstNameKinesis, sesEventDstNameCw, sesEventDstNameSns string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "%s"
  acl = "private"
}

resource "aws_iam_role" "firehose_role" {
   name = "%s"
   assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    },
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "ses.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_kinesis_firehose_delivery_stream" "test_stream" {
  name = "%s"
  destination = "s3"
  s3_configuration {
    role_arn = "${aws_iam_role.firehose_role.arn}"
    bucket_arn = "${aws_s3_bucket.bucket.arn}"
  }
}

resource "aws_iam_role_policy" "firehose_delivery_policy" {
  name = "%s"
  role = "${aws_iam_role.firehose_role.id}"
  policy = "${data.aws_iam_policy_document.fh_felivery_document.json}"
}

data "aws_iam_policy_document" "fh_felivery_document" {
    statement {
        sid = "GiveSESPermissionToPutFirehose"
        actions = [
            "firehose:PutRecord",
            "firehose:PutRecordBatch",
        ]
        resources = [
            "*",
        ]
    }
}

resource "aws_sns_topic" "ses_destination" {
  name = "%s"
}

resource "aws_ses_configuration_set" "test" {
    name = "%s"
}

resource "aws_ses_event_destination" "kinesis" {
  name = "%s"
  configuration_set_name = "${aws_ses_configuration_set.test.name}"
  enabled = true
  matching_types = ["bounce", "send"]

  kinesis_destination {
    stream_arn = "${aws_kinesis_firehose_delivery_stream.test_stream.arn}"
    role_arn = "${aws_iam_role.firehose_role.arn}"
  }
}

resource "aws_ses_event_destination" "cloudwatch" {
  name = "%s"
  configuration_set_name = "${aws_ses_configuration_set.test.name}"
  enabled = true
  matching_types = ["bounce", "send"]

  cloudwatch_destination {
    default_value = "default"
	dimension_name = "dimension"
	value_source = "emailHeader"
  }

  cloudwatch_destination {
    default_value = "default"
	dimension_name = "ses:source-ip"
	value_source = "messageTag"
  }
}

resource "aws_ses_event_destination" "sns" {
  name = "%s"
  configuration_set_name = "${aws_ses_configuration_set.test.name}"
  enabled = true
  matching_types = ["bounce", "send"]

  sns_destination {
    topic_arn = "${aws_sns_topic.ses_destination.arn}"
  }
}
`, bucketName, roleName, streamName, policyName, topicName,
		sesCfgSetName, sesEventDstNameKinesis, sesEventDstNameCw, sesEventDstNameSns)
}
