package aws

import (
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSEBSSnapshotImport_basic(t *testing.T) {
	var v ec2.Snapshot
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ebs_snapshot_import.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEbsSnapshotImportDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsSnapshotImportConfigBasic(rName, t),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotImportExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "ec2", regexp.MustCompile(`snapshot/snap-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
				),
			},
		},
	})
}

func TestAccAWSEBSSnapshotImport_tags(t *testing.T) {
	var v ec2.Snapshot
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ebs_snapshot_import.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEbsSnapshotImportDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsSnapshotImportConfigTags1(rName, t, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotImportExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "ec2", regexp.MustCompile(`snapshot/snap-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
				),
			},
			{
				Config: testAccAwsEbsSnapshotImportConfigTags2(rName, t, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotImportExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "ec2", regexp.MustCompile(`snapshot/snap-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
				),
			},
			{
				Config: testAccAwsEbsSnapshotImportConfigTags1(rName, t, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotImportExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "ec2", regexp.MustCompile(`snapshot/snap-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
				),
			},
		},
	})
}

func TestAccAWSEBSSnapshotImport_disappears(t *testing.T) {
	var v ec2.Snapshot
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ebs_snapshot_import.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEbsSnapshotImportDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsSnapshotImportConfigBasic(rName, t),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotImportExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "ec2", regexp.MustCompile(`snapshot/snap-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEbsSnapshotImport(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSEBSSnapshotImport_disappears_S3BucketObject(t *testing.T) {
	var v ec2.Snapshot
	rName := acctest.RandomWithPrefix("tf-acc-test")
	parentResourceName := "aws_s3_bucket_object.image"
	resourceName := "aws_ebs_snapshot_import.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEbsSnapshotImportDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsSnapshotImportConfigBasic(rName, t),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotImportExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "ec2", regexp.MustCompile(`snapshot/snap-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsS3BucketObject(), parentResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSnapshotImportExists(n string, v *ec2.Snapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		request := &ec2.DescribeSnapshotsInput{
			SnapshotIds: []*string{aws.String(rs.Primary.ID)},
		}

		response, err := conn.DescribeSnapshots(request)
		if err == nil {
			if response.Snapshots != nil && len(response.Snapshots) > 0 {
				*v = *response.Snapshots[0]
				return nil
			}
		}
		return fmt.Errorf("Error finding EC2 Snapshot %s", rs.Primary.ID)
	}
}

func testAccCheckAwsEbsSnapshotImportDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ebs_snapshot_import" {
			continue
		}
		input := &ec2.DescribeSnapshotsInput{
			SnapshotIds: []*string{aws.String(rs.Primary.ID)},
		}

		output, err := conn.DescribeSnapshots(input)
		if err != nil {
			if isAWSErr(err, "InvalidSnapshot.NotFound", "") {
				continue
			}
			return err
		}
		if output != nil && len(output.Snapshots) > 0 && aws.StringValue(output.Snapshots[0].SnapshotId) == rs.Primary.ID {
			return fmt.Errorf("EBS Snapshot %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAwsEbsSnapshotImportConfig_Base(t *testing.T) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "images" {
  bucket_prefix = "images-"
  force_destroy = true
}

resource "aws_s3_bucket_object" "image" {
  bucket         = aws_s3_bucket.images.id
  key            = "diskimage.vhd"
  content_base64 = %[1]q
}

# The following resources are for the *vmimport service user*
# See: https://docs.aws.amazon.com/vm-import/latest/userguide/vmie_prereqs.html#vmimport-role
resource "aws_iam_role" "vmimport" {
  assume_role_policy = data.aws_iam_policy_document.vmimport-trust.json
}

resource "aws_iam_role_policy" "vmimport-access" {
  role   = aws_iam_role.vmimport.id
  policy = data.aws_iam_policy_document.vmimport-access.json
}

data "aws_iam_policy_document" "vmimport-access" {
  statement {
    effect = "Allow"
    actions = [
      "s3:GetBucketLocation",
      "s3:GetObject",
      "s3:ListBucket",
    ]
    resources = [
      aws_s3_bucket.images.arn,
      "${aws_s3_bucket.images.arn}/*"
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "ec2:ModifySnapshotAttribute",
      "ec2:CopySnapshot",
      "ec2:RegisterImage",
      "ec2:Describe*"
    ]
    resources = [
      "*"
    ]
  }
}


data "aws_iam_policy_document" "vmimport-trust" {
  statement {
    effect = "Allow"
    principals {
      type        = "Service"
      identifiers = ["vmie.amazonaws.com"]
    }

    actions = [
      "sts:AssumeRole"
    ]

    condition {
      test     = "StringEquals"
      variable = "sts:ExternalId"
      values   = ["vmimport"]
    }
  }
}
`, testAccAwsEbsSnapshotDisk(t))
}

func testAccAwsEbsSnapshotImportConfigBasic(rName string, t *testing.T) string {
	return testAccAwsEbsSnapshotImportConfig_Base(t) + fmt.Sprintf(`
resource "aws_ebs_snapshot_import" "test" {
  disk_container {
    description = %[1]q
    format      = "VHD"
    user_bucket {
      s3_bucket = aws_s3_bucket.images.id
      s3_key    = aws_s3_bucket_object.image.key
    }
  }

  role_name = aws_iam_role.vmimport.name

  timeouts {
    create = "10m"
    delete = "10m"
  }
}
`, rName)
}

func testAccAwsEbsSnapshotImportConfigTags1(rName string, t *testing.T, tagKey1 string, tagValue1 string) string {
	return testAccAwsEbsSnapshotImportConfig_Base(t) + fmt.Sprintf(`
resource "aws_ebs_snapshot_import" "test" {
  disk_container {
    description = %[1]q
    format      = "VHD"
    user_bucket {
      s3_bucket = aws_s3_bucket.images.id
      s3_key    = aws_s3_bucket_object.image.key
    }
  }

  role_name = aws_iam_role.vmimport.name

  timeouts {
    create = "10m"
    delete = "10m"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAwsEbsSnapshotImportConfigTags2(rName string, t *testing.T, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return testAccAwsEbsSnapshotImportConfig_Base(t) + fmt.Sprintf(`
resource "aws_ebs_snapshot_import" "test" {
  disk_container {
    description = %[1]q
    format      = "VHD"
    user_bucket {
      s3_bucket = aws_s3_bucket.images.id
      s3_key    = aws_s3_bucket_object.image.key
    }
  }

  role_name = aws_iam_role.vmimport.name

  timeouts {
    create = "10m"
    delete = "10m"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAwsEbsSnapshotDisk(t *testing.T) string {
	// Take a compressed then base64'd disk image,
	// base64 decode, then decompress, then re-base64
	// the image, so it can be uploaded to s3.

	// little vmdk built by:
	// $ VBoxManage createmedium disk --filename ./image.vhd --sizebytes 512 --format vhd
	// $ cat image.vhd | gzip --best | base64
	b64_compressed := "H4sIAAAAAAACA0vOz0tNLsmsYGBgYGJgZIACJgZ1789hZUn5FQxsDIzhmUbZMHEEzSIIJJj///+QlV1rMXFVnLzHwteXYmWDDfYxjIIhA5IrigsSi4pT/0MBRJSNAZoWGBkUGBj+//9SNhpSo2AUDD+AyPOjYESW/6P1/4gGAAvDpVcACgAA"

	decoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(b64_compressed))

	zr, err := gzip.NewReader(decoder)
	if err != nil {
		t.Fatal(err)
	}

	var out strings.Builder
	encoder := base64.NewEncoder(base64.StdEncoding, &out)

	_, err = io.Copy(encoder, zr)
	if err != nil {
		t.Fatal(err)
	}
	encoder.Close()

	return out.String()
}
