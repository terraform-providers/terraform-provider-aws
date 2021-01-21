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

func TestAccAWSS3ControlBucketPolicy_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_s3control_bucket_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3ControlBucketPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3ControlBucketPolicyConfig_Policy(rName, "s3-outposts:*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3ControlBucketPolicyExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3control_bucket.test", "arn"),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`s3-outposts:\*`)),
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

func TestAccAWSS3ControlBucketPolicy_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_s3control_bucket_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3ControlBucketPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3ControlBucketPolicyConfig_Policy(rName, "s3-outposts:*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3ControlBucketPolicyExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsS3ControlBucketPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSS3ControlBucketPolicy_Policy(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_s3control_bucket_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3ControlBucketPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3ControlBucketPolicyConfig_Policy(rName, "s3-outposts:GetObject"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3ControlBucketPolicyExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`s3-outposts:GetObject`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSS3ControlBucketPolicyConfig_Policy(rName, "s3-outposts:PutObject"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3ControlBucketPolicyExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`s3-outposts:PutObject`)),
				),
			},
		},
	})
}

func testAccCheckAWSS3ControlBucketPolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).s3controlconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3control_bucket_policy" {
			continue
		}

		parsedArn, err := arn.Parse(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error parsing S3 Control Bucket ARN (%s): %w", rs.Primary.ID, err)
		}

		input := &s3control.GetBucketPolicyInput{
			AccountId: aws.String(parsedArn.AccountID),
			Bucket:    aws.String(rs.Primary.ID),
		}

		_, err = conn.GetBucketPolicy(input)

		if tfawserr.ErrCodeEquals(err, "NoSuchBucket") {
			continue
		}

		if tfawserr.ErrCodeEquals(err, "NoSuchBucketPolicy") {
			continue
		}

		if tfawserr.ErrCodeEquals(err, "NoSuchOutpost") {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("S3 Control Bucket Policy (%s) still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSS3ControlBucketPolicyExists(resourceName string) resource.TestCheckFunc {
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

		input := &s3control.GetBucketPolicyInput{
			AccountId: aws.String(parsedArn.AccountID),
			Bucket:    aws.String(rs.Primary.ID),
		}

		_, err = conn.GetBucketPolicy(input)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccAWSS3ControlBucketPolicyConfig_Policy(rName, action string) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_s3control_bucket" "test" {
  bucket     = %[1]q
  outpost_id = data.aws_outposts_outpost.test.id
}

resource "aws_s3control_bucket_policy" "test" {
  bucket = aws_s3control_bucket.test.arn
  policy = jsonencode({
    Id = "testBucketPolicy"
    Statement = [
      {
        Action = %[2]q
        Effect = "Deny"
        Principal = {
          AWS = "*"
        }
        Resource = "${aws_s3control_bucket.test.arn}/object/test"
        Sid      = "st1"
      }
    ]
    Version = "2012-10-17"
  })
}
`, rName, action)
}
