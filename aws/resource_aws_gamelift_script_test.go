package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_gamelift_script", &resource.Sweeper{
		Name: "aws_gamelift_script",
		F:    testSweepGameliftScripts,
	})
}

func testSweepGameliftScripts(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).gameliftconn

	resp, err := conn.ListScripts(&gamelift.ListScriptsInput{})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Gamelife Script sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error listing Gamelift Scripts: %s", err)
	}

	if len(resp.Scripts) == 0 {
		log.Print("[DEBUG] No Gamelift Scripts to sweep")
		return nil
	}

	log.Printf("[INFO] Found %d Gamelift Scripts", len(resp.Scripts))

	for _, script := range resp.Scripts {
		log.Printf("[INFO] Deleting Gamelift Script %q", *script.ScriptId)
		_, err := conn.DeleteScript(&gamelift.DeleteScriptInput{
			ScriptId: script.ScriptId,
		})
		if err != nil {
			return fmt.Errorf("Error deleting Gamelift Script (%s): %w",
				aws.StringValue(script.ScriptId), err)
		}
	}

	return nil
}

func TestAccAWSGameliftScript_basic(t *testing.T) {
	var conf gamelift.Script
	resourceName := "aws_gamelift_script.test"
	rName := acctest.RandomWithPrefix("acc-test-test")
	rNameUpdated := acctest.RandomWithPrefix("acc-test-test")
	region := testAccGetRegion()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSGameliftScripts(t) },
		ErrorCheck:   testAccErrorCheck(t, gamelift.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGameliftScriptDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftScriptBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftScriptExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "gamelift", regexp.MustCompile(`script/script-.+`)),
					resource.TestCheckResourceAttr(resourceName, "storage_location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_location.0.bucket", fmt.Sprintf("prod-gamescale-scripts-%s", region)),
					resource.TestCheckResourceAttrSet(resourceName, "storage_location.0.key"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file"},
			},
			{
				Config: testAccAWSGameliftScriptBasicConfigUpdated(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftScriptExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "gamelift", regexp.MustCompile(`script/script-.+`)),
					resource.TestCheckResourceAttr(resourceName, "storage_location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_location.0.bucket", fmt.Sprintf("prod-gamescale-scripts-%s", region)),
					resource.TestCheckResourceAttrSet(resourceName, "storage_location.0.key"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSGameliftScript_storageLocation(t *testing.T) {
	var conf gamelift.Script
	resourceName := "aws_gamelift_script.test"
	rName := acctest.RandomWithPrefix("acc-test-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSGameliftScripts(t) },
		ErrorCheck:   testAccErrorCheck(t, gamelift.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGameliftScriptDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftScriptConfigStorageLocation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftScriptExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "storage_location.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_location.0.bucket", "aws_s3_bucket_object.test", "bucket"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_location.0.key", "aws_s3_bucket_object.test", "key"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_location.0.role_arn", "aws_iam_role.test", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSGameliftScriptConfigStorageLocationUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftScriptExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "storage_location.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_location.0.bucket", "aws_s3_bucket_object.test", "bucket"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_location.0.key", "aws_s3_bucket_object.test", "key"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_location.0.role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_location.0.object_version", "aws_s3_bucket_object.test", "version_id"),
				),
			},
		},
	})
}

func TestAccAWSGameliftScript_tags(t *testing.T) {
	var conf gamelift.Script
	resourceName := "aws_gamelift_script.test"
	rName := acctest.RandomWithPrefix("acc-test-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSGameliftScripts(t) },
		ErrorCheck:   testAccErrorCheck(t, gamelift.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGameliftScriptDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftScriptConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftScriptExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file"},
			},
			{
				Config: testAccAWSGameliftScriptConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftScriptExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSGameliftScriptConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftScriptExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSGameliftScript_disappears(t *testing.T) {
	var conf gamelift.Script
	resourceName := "aws_gamelift_script.test"
	rName := acctest.RandomWithPrefix("acc-test-script")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSGameliftScripts(t) },
		ErrorCheck:   testAccErrorCheck(t, gamelift.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGameliftScriptDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftScriptBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftScriptExists(resourceName, &conf),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsGameliftScript(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSGameliftScriptExists(n string, res *gamelift.Script) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Gamelift Script ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).gameliftconn

		req := &gamelift.DescribeScriptInput{
			ScriptId: aws.String(rs.Primary.ID),
		}
		out, err := conn.DescribeScript(req)
		if err != nil {
			return err
		}

		b := out.Script

		if aws.StringValue(b.ScriptId) != rs.Primary.ID {
			return fmt.Errorf("Gamelift Script not found")
		}

		*res = *b

		return nil
	}
}

func testAccCheckAWSGameliftScriptDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).gameliftconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_gamelift_script" {
			continue
		}

		req := gamelift.DescribeScriptInput{
			ScriptId: aws.String(rs.Primary.ID),
		}
		out, err := conn.DescribeScript(&req)
		if err == nil {
			if *out.Script.ScriptId == rs.Primary.ID {
				return fmt.Errorf("Gamelift Script still exists")
			}
		}
		if isAWSErr(err, gamelift.ErrCodeNotFoundException, "") {
			return nil
		}

		return err
	}

	return nil
}

func testAccPreCheckAWSGameliftScripts(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).gameliftconn

	input := &gamelift.ListScriptsInput{}

	_, err := conn.ListScripts(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSGameliftScriptBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_script" "test" {
  name     = %[1]q
  zip_file = "test-fixtures/lambdatest.zip"
}
`, rName)
}

func testAccAWSGameliftScriptBasicConfigUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_script" "test" {
  name     = %[1]q
  zip_file = "test-fixtures/lambdatest_modified.zip"
}
`, rName)
}

func testAccAWSGameliftScriptConfigStorageLocation(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  acl           = "private"
  force_destroy = true

  versioning {
    enabled = true
  }
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "gamelift.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_s3_bucket_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = %[1]q
  source = "test-fixtures/lambdatest.zip"
  etag   = filemd5("test-fixtures/lambdatest.zip")
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
		"s3:GetObject",
		"s3:GetObjectVersion"
      ],
      "Resource": [
        "${aws_s3_bucket.test.arn}/*"
      ]
    }
  ]
}
EOF
}

resource "aws_gamelift_script" "test" {
  name = %[1]q

  storage_location {
    bucket   = aws_s3_bucket_object.test.bucket
    key      = aws_s3_bucket_object.test.key
    role_arn = aws_iam_role.test.arn
  }
}
`, rName)
}

func testAccAWSGameliftScriptConfigStorageLocationUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  acl           = "private"
  force_destroy = true

  versioning {
    enabled = true
  }
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "gamelift.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_s3_bucket_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = %[1]q
  source = "test-fixtures/lambdatest.zip"
  etag   = filemd5("test-fixtures/lambdatest.zip")
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
		"s3:GetObject",
		"s3:GetObjectVersion"
      ],
      "Resource": [
        "${aws_s3_bucket.test.arn}/*"
      ]
    }
  ]
}
EOF
}

resource "aws_gamelift_script" "test" {
  name = %[1]q

  storage_location {
    bucket         = aws_s3_bucket_object.test.bucket
    key            = aws_s3_bucket_object.test.key
    role_arn       = aws_iam_role.test.arn
    object_version = aws_s3_bucket_object.test.version_id
  }
}
`, rName)
}

func testAccAWSGameliftScriptConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_script" "test" {
  name     = %[1]q
  zip_file = "test-fixtures/lambdatest.zip"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSGameliftScriptConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_script" "test" {
  name     = %[1]q
  zip_file = "test-fixtures/lambdatest.zip"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
