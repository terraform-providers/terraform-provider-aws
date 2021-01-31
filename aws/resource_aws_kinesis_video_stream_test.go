package aws

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesisvideo"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSKinesisVideoStream_basic(t *testing.T) {
	var stream kinesisvideo.StreamInfo

	resourceName := "aws_kinesis_video_stream.default"
	rInt1 := acctest.RandInt()
	rInt2 := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(kinesisvideo.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisVideoStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisVideoStreamConfig(rInt1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisVideoStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("terraform-kinesis-video-stream-test-%d", rInt1)),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "kinesisvideo", regexp.MustCompile(fmt.Sprintf("stream/terraform-kinesis-video-stream-test-%d/.+", rInt1))),
					resource.TestCheckResourceAttrSet(resourceName, "creation_time"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKinesisVideoStreamConfig(rInt2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisVideoStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("terraform-kinesis-video-stream-test-%d", rInt2)),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "kinesisvideo", regexp.MustCompile(fmt.Sprintf("stream/terraform-kinesis-video-stream-test-%d/.+", rInt2))),
				),
			},
		},
	})
}

func TestAccAWSKinesisVideoStream_options(t *testing.T) {
	var stream kinesisvideo.StreamInfo

	resourceName := "aws_kinesis_video_stream.default"
	kmsResourceName := "aws_kms_key.default"
	rInt := acctest.RandInt()
	rName1 := acctest.RandString(8)
	rName2 := acctest.RandString(8)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(kinesisvideo.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisVideoStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisVideoStreamConfig_Options(rInt, rName1, "video/h264"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisVideoStreamExists(resourceName, &stream),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "kinesisvideo", regexp.MustCompile(fmt.Sprintf("stream/terraform-kinesis-video-stream-test-%d/.+", rInt))),
					resource.TestCheckResourceAttr(resourceName, "data_retention_in_hours", "1"),
					resource.TestCheckResourceAttr(resourceName, "media_type", "video/h264"),
					resource.TestCheckResourceAttr(resourceName, "device_name", fmt.Sprintf("kinesis-video-device-name-%s", rName1)),
					resource.TestCheckResourceAttrPair(
						resourceName, "kms_key_id",
						kmsResourceName, "id"),
				),
			},
			{
				Config: testAccKinesisVideoStreamConfig_Options(rInt, rName2, "video/h120"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisVideoStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "media_type", "video/h120"),
					resource.TestCheckResourceAttr(resourceName, "device_name", fmt.Sprintf("kinesis-video-device-name-%s", rName2)),
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

func TestAccAWSKinesisVideoStream_Tags(t *testing.T) {
	var stream kinesisvideo.StreamInfo

	resourceName := "aws_kinesis_video_stream.default"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(kinesisvideo.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisVideoStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisVideoStreamConfig_Tags1(rInt, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisVideoStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccKinesisVideoStreamConfig_Tags2(rInt, "key1", "value1", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisVideoStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKinesisVideoStreamConfig_Tags1(rInt, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisVideoStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSKinesisVideoStream_disappears(t *testing.T) {
	var stream kinesisvideo.StreamInfo

	resourceName := "aws_kinesis_video_stream.default"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(kinesisvideo.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisVideoStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisVideoStreamConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisVideoStreamExists(resourceName, &stream),
					testAccCheckKinesisVideoStreamDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckKinesisVideoStreamDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).kinesisvideoconn

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No resource ID is set")
		}

		input := &kinesisvideo.DeleteStreamInput{
			StreamARN:      aws.String(rs.Primary.ID),
			CurrentVersion: aws.String(rs.Primary.Attributes["version"]),
		}

		if _, err := conn.DeleteStream(input); err != nil {
			return err
		}

		stateConf := &resource.StateChangeConf{
			Pending:    []string{kinesisvideo.StatusDeleting},
			Target:     []string{"DELETED"},
			Refresh:    kinesisVideoStreamStateRefresh(conn, rs.Primary.ID),
			Timeout:    15 * time.Minute,
			Delay:      10 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		if _, err := stateConf.WaitForState(); err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKinesisVideoStreamExists(n string, stream *kinesisvideo.StreamInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Kinesis ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).kinesisvideoconn
		describeOpts := &kinesisvideo.DescribeStreamInput{
			StreamARN: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeStream(describeOpts)
		if err != nil {
			return err
		}

		*stream = *resp.StreamInfo

		return nil
	}
}

func testAccCheckKinesisVideoStreamDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_kinesis_video_stream" {
			continue
		}
		conn := testAccProvider.Meta().(*AWSClient).kinesisvideoconn
		describeOpts := &kinesisvideo.DescribeStreamInput{
			StreamARN: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeStream(describeOpts)
		if err == nil {
			if resp.StreamInfo != nil && aws.StringValue(resp.StreamInfo.Status) != "DELETING" {
				return fmt.Errorf("Error Kinesis Video Stream still exists")
			}
		}

		return nil

	}

	return nil
}

func testAccKinesisVideoStreamConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_video_stream" "default" {
  name = "terraform-kinesis-video-stream-test-%d"
}
`, rInt)
}

func testAccKinesisVideoStreamConfig_Options(rInt int, rName, mediaType string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "default" {
  description             = "KMS key 1"
  deletion_window_in_days = 7
}

resource "aws_kinesis_video_stream" "default" {
  name = "terraform-kinesis-video-stream-test-%[1]d"

  data_retention_in_hours = 1
  device_name             = "kinesis-video-device-name-%[2]s"
  kms_key_id              = aws_kms_key.default.id
  media_type              = "%[3]s"
}
`, rInt, rName, mediaType)
}

func testAccKinesisVideoStreamConfig_Tags1(rInt int, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_video_stream" "default" {
  name = "terraform-kinesis-video-stream-test-%d"

  tags = {
    %[2]q = %[3]q
  }
}
`, rInt, tagKey1, tagValue1)
}

func testAccKinesisVideoStreamConfig_Tags2(rInt int, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_video_stream" "default" {
  name = "terraform-kinesis-video-stream-test-%d"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rInt, tagKey1, tagValue1, tagKey2, tagValue2)
}
