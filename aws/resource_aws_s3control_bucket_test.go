package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSS3ControlBucket_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_s3control_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3ControlBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3ControlBucketConfig_Bucket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3ControlBucketExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "s3-outposts", regexp.MustCompile(fmt.Sprintf("outpost/[^/]+/bucket/%s", rName))),
					resource.TestCheckResourceAttr(resourceName, "bucket", rName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttrPair(resourceName, "outpost_id", "data.aws_outposts_outpost.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_enabled", "true"),
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

func TestAccAWSS3ControlBucket_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_s3control_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3ControlBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3ControlBucketConfig_Bucket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3ControlBucketExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsS3ControlBucket(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSS3ControlBucket_Tags(t *testing.T) {
	TestAccSkip(t, "S3 Control Bucket resource tagging requires additional eventual consistency handling, see also: https://github.com/hashicorp/terraform-provider-aws/issues/15572")

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_s3control_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3ControlBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3ControlBucketConfig_Tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3ControlBucketExists(resourceName),
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
				Config: testAccAWSS3ControlBucketConfig_Tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3ControlBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSS3ControlBucketConfig_Tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3ControlBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAWSS3ControlBucketDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).s3controlconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3control_bucket" {
			continue
		}

		parsedArn, err := arn.Parse(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error parsing S3 Control Bucket ARN (%s): %w", rs.Primary.ID, err)
		}

		input := &s3control.GetBucketInput{
			AccountId: aws.String(parsedArn.AccountID),
			Bucket:    aws.String(rs.Primary.ID),
		}

		_, err = conn.GetBucket(input)

		if tfawserr.ErrCodeEquals(err, "NoSuchBucket") {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("S3 Control Bucket (%s) still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSS3ControlBucketExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no resource ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).s3controlconn

		parsedArn, err := arn.Parse(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error parsing S3 Control Bucket ARN (%s): %w", rs.Primary.ID, err)
		}

		input := &s3control.GetBucketInput{
			AccountId: aws.String(parsedArn.AccountID),
			Bucket:    aws.String(rs.Primary.ID),
		}

		_, err = conn.GetBucket(input)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccAWSS3ControlBucketConfig_Bucket(rName string) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_s3control_bucket" "test" {
  bucket     = %[1]q
  outpost_id = data.aws_outposts_outpost.test.id
}
`, rName)
}

func testAccAWSS3ControlBucketConfig_Tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_s3control_bucket" "test" {
  bucket     = %[1]q
  outpost_id = data.aws_outposts_outpost.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSS3ControlBucketConfig_Tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_s3control_bucket" "test" {
  bucket     = %[1]q
  outpost_id = data.aws_outposts_outpost.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
