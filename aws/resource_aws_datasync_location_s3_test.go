package aws

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_datasync_location_s3", &resource.Sweeper{
		Name: "aws_datasync_location_s3",
		F:    testSweepDataSyncLocationS3s,
	})
}

func testSweepDataSyncLocationS3s(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).datasyncconn

	input := &datasync.ListLocationsInput{}
	for {
		output, err := conn.ListLocations(input)

		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping DataSync Location S3 sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving DataSync Location S3s: %s", err)
		}

		if len(output.Locations) == 0 {
			log.Print("[DEBUG] No DataSync Location S3s to sweep")
			return nil
		}

		for _, location := range output.Locations {
			uri := aws.StringValue(location.LocationUri)
			if !strings.HasPrefix(uri, "s3://") {
				log.Printf("[INFO] Skipping DataSync Location S3: %s", uri)
				continue
			}
			log.Printf("[INFO] Deleting DataSync Location S3: %s", uri)
			input := &datasync.DeleteLocationInput{
				LocationArn: location.LocationArn,
			}

			_, err := conn.DeleteLocation(input)

			if isAWSErr(err, "InvalidRequestException", "not found") {
				continue
			}

			if err != nil {
				log.Printf("[ERROR] Failed to delete DataSync Location S3 (%s): %s", uri, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}

func TestAccAWSDataSyncLocationS3_basic(t *testing.T) {
	var locationS31 datasync.DescribeLocationS3Output
	rName := acctest.RandomWithPrefix("tf-acc-test")
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_datasync_location_s3.test"
	s3BucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncLocationS3Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncLocationS3Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationS3Exists(resourceName, &locationS31),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "datasync", regexp.MustCompile(`location/loc-.+`)),
					resource.TestCheckResourceAttrPair(resourceName, "s3_bucket_arn", s3BucketResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "s3_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "s3_config.0.bucket_access_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/test/"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestMatchResourceAttr(resourceName, "uri", regexp.MustCompile(`^s3://.+/`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"s3_bucket_arn"},
			},
		},
	})
}

func TestAccAWSDataSyncLocationS3_disappears(t *testing.T) {
	var locationS31 datasync.DescribeLocationS3Output
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_datasync_location_s3.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncLocationS3Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncLocationS3Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationS3Exists(resourceName, &locationS31),
					testAccCheckAWSDataSyncLocationS3Disappears(&locationS31),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSDataSyncLocationS3_Tags(t *testing.T) {
	var locationS31, locationS32, locationS33 datasync.DescribeLocationS3Output
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_datasync_location_s3.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncLocationS3Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncLocationS3ConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationS3Exists(resourceName, &locationS31),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"s3_bucket_arn"},
			},
			{
				Config: testAccAWSDataSyncLocationS3ConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationS3Exists(resourceName, &locationS32),
					testAccCheckAWSDataSyncLocationS3NotRecreated(&locationS31, &locationS32),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSDataSyncLocationS3ConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationS3Exists(resourceName, &locationS33),
					testAccCheckAWSDataSyncLocationS3NotRecreated(&locationS32, &locationS33),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func testAccCheckAWSDataSyncLocationS3Destroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).datasyncconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_datasync_location_s3" {
			continue
		}

		input := &datasync.DescribeLocationS3Input{
			LocationArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeLocationS3(input)

		if isAWSErr(err, "InvalidRequestException", "not found") {
			return nil
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckAWSDataSyncLocationS3Exists(resourceName string, locationS3 *datasync.DescribeLocationS3Output) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).datasyncconn
		input := &datasync.DescribeLocationS3Input{
			LocationArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeLocationS3(input)

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("Location %q does not exist", rs.Primary.ID)
		}

		*locationS3 = *output

		return nil
	}
}

func testAccCheckAWSDataSyncLocationS3Disappears(location *datasync.DescribeLocationS3Output) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).datasyncconn

		input := &datasync.DeleteLocationInput{
			LocationArn: location.LocationArn,
		}

		_, err := conn.DeleteLocation(input)

		return err
	}
}

func testAccCheckAWSDataSyncLocationS3NotRecreated(i, j *datasync.DescribeLocationS3Output) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CreationTime) != aws.TimeValue(j.CreationTime) {
			return errors.New("DataSync Location S3 was recreated")
		}

		return nil
	}
}

func testAccAWSDataSyncLocationS3ConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "datasync.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_s3_bucket" "test" {
  bucket        = %q
  force_destroy = true
}
`, rName, rName)
}

func testAccAWSDataSyncLocationS3Config(rName string) string {
	return testAccAWSDataSyncLocationS3ConfigBase(rName) + fmt.Sprintf(`
resource "aws_datasync_location_s3" "test" {
  s3_bucket_arn = "${aws_s3_bucket.test.arn}"
  subdirectory  = "/test"

  s3_config {
    bucket_access_role_arn = "${aws_iam_role.test.arn}"
  }
}
`)
}

func testAccAWSDataSyncLocationS3ConfigTags1(rName, key1, value1 string) string {
	return testAccAWSDataSyncLocationS3ConfigBase(rName) + fmt.Sprintf(`
resource "aws_datasync_location_s3" "test" {
  s3_bucket_arn = "${aws_s3_bucket.test.arn}"
  subdirectory  = "/test"

  s3_config {
    bucket_access_role_arn = "${aws_iam_role.test.arn}"
  }

  tags = {
    %q = %q
  }
}
`, key1, value1)
}

func testAccAWSDataSyncLocationS3ConfigTags2(rName, key1, value1, key2, value2 string) string {
	return testAccAWSDataSyncLocationS3ConfigBase(rName) + fmt.Sprintf(`
resource "aws_datasync_location_s3" "test" {
  s3_bucket_arn = "${aws_s3_bucket.test.arn}"
  subdirectory  = "/test"

  s3_config {
    bucket_access_role_arn = "${aws_iam_role.test.arn}"
  }

  tags = {
    %q = %q
    %q = %q
  }
}
`, key1, value1, key2, value2)
}
