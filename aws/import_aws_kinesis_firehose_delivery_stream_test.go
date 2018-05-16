package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSKinesisFirehoseDeliveryStream_importBasic(t *testing.T) {
	resName := "aws_kinesis_firehose_delivery_stream.test_stream"
	rInt := acctest.RandInt()

	funcName := fmt.Sprintf("aws_kinesis_firehose_ds_import_%d", rInt)
	policyName := fmt.Sprintf("tf_acc_policy_%d", rInt)
	roleName := fmt.Sprintf("tf_acc_role_%d", rInt)

	config := testAccFirehoseAWSLambdaConfigBasic(funcName, policyName, roleName) +
		fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_extendedS3basic,
			rInt, rInt, rInt, rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				ResourceName:      resName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Ensure we properly error on malformed import IDs
			{
				ResourceName:  resName,
				ImportState:   true,
				ImportStateId: "just-a-name",
				ExpectError:   regexp.MustCompile(`Expected ID in format`),
			},
			{
				ResourceName:  resName,
				ImportState:   true,
				ImportStateId: "arn:aws:firehose:us-east-1:123456789012:missing-slash",
				ExpectError:   regexp.MustCompile(`Expected ID in format`),
			},
		},
	})
}
