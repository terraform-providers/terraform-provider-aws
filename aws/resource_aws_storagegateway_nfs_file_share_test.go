package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSStorageGatewayNfsFileShare_basic(t *testing.T) {
	var nfsFileShare storagegateway.NFSFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_nfs_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayNfsFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayNfsFileShareConfig_Required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayNfsFileShareExists(resourceName, &nfsFileShare),
					resource.TestMatchResourceAttr(resourceName, "arn", regexp.MustCompile(`^arn:[^:]+:storagegateway:[^:]+:[^:]+:share/share-.+$`)),
					resource.TestCheckResourceAttr(resourceName, "client_list.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_list.217649824", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "default_storage_class", "S3_STANDARD"),
					resource.TestMatchResourceAttr(resourceName, "fileshare_id", regexp.MustCompile(`^share-`)),
					resource.TestMatchResourceAttr(resourceName, "gateway_arn", regexp.MustCompile(`^arn:`)),
					resource.TestCheckResourceAttr(resourceName, "guess_mime_type_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_arn", ""),
					resource.TestMatchResourceAttr(resourceName, "location_arn", regexp.MustCompile(`^arn:`)),
					resource.TestCheckResourceAttr(resourceName, "nfs_file_share_defaults.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "object_acl", storagegateway.ObjectACLPrivate),
					resource.TestCheckResourceAttr(resourceName, "read_only", "false"),
					resource.TestCheckResourceAttr(resourceName, "requester_pays", "false"),
					resource.TestMatchResourceAttr(resourceName, "role_arn", regexp.MustCompile(`^arn:`)),
					resource.TestCheckResourceAttr(resourceName, "squash", "RootSquash"),
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

func TestAccAWSStorageGatewayNfsFileShare_ClientList(t *testing.T) {
	var nfsFileShare storagegateway.NFSFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_nfs_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayNfsFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayNfsFileShareConfig_ClientList_Single(rName, "1.1.1.1/32"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayNfsFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "client_list.#", "1"),
				),
			},
			{
				Config: testAccAWSStorageGatewayNfsFileShareConfig_ClientList_Multiple(rName, "2.2.2.2/32", "3.3.3.3/32"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayNfsFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "client_list.#", "2"),
				),
			},
			{
				Config: testAccAWSStorageGatewayNfsFileShareConfig_ClientList_Single(rName, "4.4.4.4/32"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayNfsFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "client_list.#", "1"),
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

func TestAccAWSStorageGatewayNfsFileShare_DefaultStorageClass(t *testing.T) {
	var nfsFileShare storagegateway.NFSFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_nfs_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayNfsFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayNfsFileShareConfig_DefaultStorageClass(rName, "S3_STANDARD_IA"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayNfsFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "default_storage_class", "S3_STANDARD_IA"),
				),
			},
			{
				Config: testAccAWSStorageGatewayNfsFileShareConfig_DefaultStorageClass(rName, "S3_ONEZONE_IA"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayNfsFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "default_storage_class", "S3_ONEZONE_IA"),
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

func TestAccAWSStorageGatewayNfsFileShare_GuessMIMETypeEnabled(t *testing.T) {
	var nfsFileShare storagegateway.NFSFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_nfs_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayNfsFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayNfsFileShareConfig_GuessMIMETypeEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayNfsFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "guess_mime_type_enabled", "false"),
				),
			},
			{
				Config: testAccAWSStorageGatewayNfsFileShareConfig_GuessMIMETypeEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayNfsFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "guess_mime_type_enabled", "true"),
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

func TestAccAWSStorageGatewayNfsFileShare_KMSEncrypted(t *testing.T) {
	var nfsFileShare storagegateway.NFSFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_nfs_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayNfsFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSStorageGatewayNfsFileShareConfig_KMSEncrypted(rName, true),
				ExpectError: regexp.MustCompile(`KMSKey is missing`),
			},
			{
				Config: testAccAWSStorageGatewayNfsFileShareConfig_KMSEncrypted(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayNfsFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", "false"),
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

func TestAccAWSStorageGatewayNfsFileShare_KMSKeyArn(t *testing.T) {
	var nfsFileShare storagegateway.NFSFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_nfs_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayNfsFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayNfsFileShareConfig_KMSKeyArn(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayNfsFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", "true"),
					resource.TestMatchResourceAttr(resourceName, "kms_key_arn", regexp.MustCompile(`^arn:`)),
				),
			},
			{
				Config: testAccAWSStorageGatewayNfsFileShareConfig_KMSKeyArn_Update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayNfsFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", "true"),
					resource.TestMatchResourceAttr(resourceName, "kms_key_arn", regexp.MustCompile(`^arn:`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSStorageGatewayNfsFileShareConfig_KMSEncrypted(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayNfsFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", "false"),
				),
			},
		},
	})
}

func TestAccAWSStorageGatewayNfsFileShare_NFSFileShareDefaults(t *testing.T) {
	var nfsFileShare storagegateway.NFSFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_nfs_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayNfsFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayNfsFileShareConfig_NFSFileShareDefaults(rName, "0700", "0600", 1, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayNfsFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "nfs_file_share_defaults.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "nfs_file_share_defaults.0.directory_mode", "0700"),
					resource.TestCheckResourceAttr(resourceName, "nfs_file_share_defaults.0.file_mode", "0600"),
					resource.TestCheckResourceAttr(resourceName, "nfs_file_share_defaults.0.group_id", "1"),
					resource.TestCheckResourceAttr(resourceName, "nfs_file_share_defaults.0.owner_id", "2"),
				),
			},
			{
				Config: testAccAWSStorageGatewayNfsFileShareConfig_NFSFileShareDefaults(rName, "0770", "0660", 3, 4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayNfsFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "nfs_file_share_defaults.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "nfs_file_share_defaults.0.directory_mode", "0770"),
					resource.TestCheckResourceAttr(resourceName, "nfs_file_share_defaults.0.file_mode", "0660"),
					resource.TestCheckResourceAttr(resourceName, "nfs_file_share_defaults.0.group_id", "3"),
					resource.TestCheckResourceAttr(resourceName, "nfs_file_share_defaults.0.owner_id", "4"),
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

func TestAccAWSStorageGatewayNfsFileShare_ObjectACL(t *testing.T) {
	var nfsFileShare storagegateway.NFSFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_nfs_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayNfsFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayNfsFileShareConfig_ObjectACL(rName, storagegateway.ObjectACLPublicRead),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayNfsFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "object_acl", storagegateway.ObjectACLPublicRead),
				),
			},
			{
				Config: testAccAWSStorageGatewayNfsFileShareConfig_ObjectACL(rName, storagegateway.ObjectACLPublicReadWrite),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayNfsFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "object_acl", storagegateway.ObjectACLPublicReadWrite),
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

func TestAccAWSStorageGatewayNfsFileShare_ReadOnly(t *testing.T) {
	var nfsFileShare storagegateway.NFSFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_nfs_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayNfsFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayNfsFileShareConfig_ReadOnly(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayNfsFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "read_only", "false"),
				),
			},
			{
				Config: testAccAWSStorageGatewayNfsFileShareConfig_ReadOnly(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayNfsFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "read_only", "true"),
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

func TestAccAWSStorageGatewayNfsFileShare_RequesterPays(t *testing.T) {
	var nfsFileShare storagegateway.NFSFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_nfs_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayNfsFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayNfsFileShareConfig_RequesterPays(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayNfsFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "requester_pays", "false"),
				),
			},
			{
				Config: testAccAWSStorageGatewayNfsFileShareConfig_RequesterPays(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayNfsFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "requester_pays", "true"),
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

func TestAccAWSStorageGatewayNfsFileShare_Squash(t *testing.T) {
	var nfsFileShare storagegateway.NFSFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_nfs_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayNfsFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayNfsFileShareConfig_Squash(rName, "NoSquash"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayNfsFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "squash", "NoSquash"),
				),
			},
			{
				Config: testAccAWSStorageGatewayNfsFileShareConfig_Squash(rName, "AllSquash"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayNfsFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "squash", "AllSquash"),
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

func testAccCheckAWSStorageGatewayNfsFileShareDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).storagegatewayconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_storagegateway_nfs_file_share" {
			continue
		}

		input := &storagegateway.DescribeNFSFileSharesInput{
			FileShareARNList: []*string{aws.String(rs.Primary.ID)},
		}

		output, err := conn.DescribeNFSFileShares(input)

		if err != nil {
			if isAWSErr(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified file share was not found.") {
				continue
			}
			return err
		}

		if output != nil && len(output.NFSFileShareInfoList) > 0 && output.NFSFileShareInfoList[0] != nil {
			return fmt.Errorf("Storage Gateway NFS File Share %q still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccCheckAWSStorageGatewayNfsFileShareExists(resourceName string, nfsFileShare *storagegateway.NFSFileShareInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).storagegatewayconn
		input := &storagegateway.DescribeNFSFileSharesInput{
			FileShareARNList: []*string{aws.String(rs.Primary.ID)},
		}

		output, err := conn.DescribeNFSFileShares(input)

		if err != nil {
			return err
		}

		if output == nil || len(output.NFSFileShareInfoList) == 0 || output.NFSFileShareInfoList[0] == nil {
			return fmt.Errorf("Storage Gateway NFS File Share %q does not exist", rs.Primary.ID)
		}

		*nfsFileShare = *output.NFSFileShareInfoList[0]

		return nil
	}
}

func testAccAWSStorageGateway_S3FileShareBase(rName string) string {
	return testAccAWSStorageGateway_FileGatewayBase(rName) + fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "storagegateway.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "test" {
  role = "${aws_iam_role.test.name}"
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "s3:*"
      ],
      "Resource": [
        "${aws_s3_bucket.test.arn}",
        "${aws_s3_bucket.test.arn}/*"
      ],
      "Effect": "Allow"
    }
  ]
}
POLICY
}

resource "aws_s3_bucket" "test" {
  bucket        = %q
  force_destroy = true
}

resource "aws_storagegateway_gateway" "test" {
  depends_on = ["aws_iam_role_policy.test"]

  gateway_ip_address = "${aws_instance.test.public_ip}"
  gateway_name       = %q
  gateway_timezone   = "GMT"
  gateway_type       = "FILE_S3"
}
`, rName, rName, rName)
}

func testAccAWSStorageGatewayNfsFileShareConfig_Required(rName string) string {
	return testAccAWSStorageGateway_S3FileShareBase(rName) + `
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list  = ["0.0.0.0/0"]
  gateway_arn  = "${aws_storagegateway_gateway.test.arn}"
  location_arn = "${aws_s3_bucket.test.arn}"
  role_arn     = "${aws_iam_role.test.arn}"
}
`
}

func testAccAWSStorageGatewayNfsFileShareConfig_ClientList_Single(rName, clientList1 string) string {
	return testAccAWSStorageGateway_S3FileShareBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list           = [%q]
  gateway_arn           = "${aws_storagegateway_gateway.test.arn}"
  location_arn          = "${aws_s3_bucket.test.arn}"
  role_arn              = "${aws_iam_role.test.arn}"
}
`, clientList1)
}

func testAccAWSStorageGatewayNfsFileShareConfig_ClientList_Multiple(rName, clientList1, clientList2 string) string {
	return testAccAWSStorageGateway_S3FileShareBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list           = [%q, %q]
  gateway_arn           = "${aws_storagegateway_gateway.test.arn}"
  location_arn          = "${aws_s3_bucket.test.arn}"
  role_arn              = "${aws_iam_role.test.arn}"
}
`, clientList1, clientList2)
}

func testAccAWSStorageGatewayNfsFileShareConfig_DefaultStorageClass(rName, defaultStorageClass string) string {
	return testAccAWSStorageGateway_S3FileShareBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list           = ["0.0.0.0/0"]
  default_storage_class = %q
  gateway_arn           = "${aws_storagegateway_gateway.test.arn}"
  location_arn          = "${aws_s3_bucket.test.arn}"
  role_arn              = "${aws_iam_role.test.arn}"
}
`, defaultStorageClass)
}

func testAccAWSStorageGatewayNfsFileShareConfig_GuessMIMETypeEnabled(rName string, guessMimeTypeEnabled bool) string {
	return testAccAWSStorageGateway_S3FileShareBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list             = ["0.0.0.0/0"]
  gateway_arn             = "${aws_storagegateway_gateway.test.arn}"
  guess_mime_type_enabled = %t
  location_arn            = "${aws_s3_bucket.test.arn}"
  role_arn                = "${aws_iam_role.test.arn}"
}
`, guessMimeTypeEnabled)
}

func testAccAWSStorageGatewayNfsFileShareConfig_KMSEncrypted(rName string, kmsEncrypted bool) string {
	return testAccAWSStorageGateway_S3FileShareBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list   = ["0.0.0.0/0"]
  gateway_arn   = "${aws_storagegateway_gateway.test.arn}"
  kms_encrypted = %t
  location_arn  = "${aws_s3_bucket.test.arn}"
  role_arn      = "${aws_iam_role.test.arn}"
}
`, kmsEncrypted)
}

func testAccAWSStorageGatewayNfsFileShareConfig_KMSKeyArn(rName string) string {
	return testAccAWSStorageGateway_S3FileShareBase(rName) + `
resource "aws_kms_key" "test" {
  count = 2

  deletion_window_in_days = 7
  description             = "Terraform Acceptance Testing"
}

resource "aws_storagegateway_nfs_file_share" "test" {
  client_list             = ["0.0.0.0/0"]
  gateway_arn             = "${aws_storagegateway_gateway.test.arn}"
  kms_encrypted           = true
  kms_key_arn             = "${aws_kms_key.test.0.arn}"
  location_arn            = "${aws_s3_bucket.test.arn}"
  role_arn                = "${aws_iam_role.test.arn}"
}
`
}

func testAccAWSStorageGatewayNfsFileShareConfig_KMSKeyArn_Update(rName string) string {
	return testAccAWSStorageGateway_S3FileShareBase(rName) + `
resource "aws_kms_key" "test" {
  count = 2

  deletion_window_in_days = 7
  description             = "Terraform Acceptance Testing"
}

resource "aws_storagegateway_nfs_file_share" "test" {
  client_list             = ["0.0.0.0/0"]
  gateway_arn             = "${aws_storagegateway_gateway.test.arn}"
  kms_encrypted           = true
  kms_key_arn             = "${aws_kms_key.test.1.arn}"
  location_arn            = "${aws_s3_bucket.test.arn}"
  role_arn                = "${aws_iam_role.test.arn}"
}
`
}

func testAccAWSStorageGatewayNfsFileShareConfig_NFSFileShareDefaults(rName, directoryMode, fileMode string, groupID, ownerID int) string {
	return testAccAWSStorageGateway_S3FileShareBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list           = ["0.0.0.0/0"]
  gateway_arn           = "${aws_storagegateway_gateway.test.arn}"
  location_arn          = "${aws_s3_bucket.test.arn}"
  role_arn              = "${aws_iam_role.test.arn}"

  nfs_file_share_defaults {
    directory_mode = %q
    file_mode      = %q
    group_id       = %d
    owner_id       = %d
  }
}
`, directoryMode, fileMode, groupID, ownerID)
}

func testAccAWSStorageGatewayNfsFileShareConfig_ObjectACL(rName, objectACL string) string {
	return testAccAWSStorageGateway_S3FileShareBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list  = ["0.0.0.0/0"]
  gateway_arn  = "${aws_storagegateway_gateway.test.arn}"
  location_arn = "${aws_s3_bucket.test.arn}"
  object_acl   = %q
  role_arn     = "${aws_iam_role.test.arn}"
}
`, objectACL)
}

func testAccAWSStorageGatewayNfsFileShareConfig_ReadOnly(rName string, readOnly bool) string {
	return testAccAWSStorageGateway_S3FileShareBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list   = ["0.0.0.0/0"]
  gateway_arn   = "${aws_storagegateway_gateway.test.arn}"
  location_arn  = "${aws_s3_bucket.test.arn}"
  read_only     = %t
  role_arn      = "${aws_iam_role.test.arn}"
}
`, readOnly)
}

func testAccAWSStorageGatewayNfsFileShareConfig_RequesterPays(rName string, requesterPays bool) string {
	return testAccAWSStorageGateway_S3FileShareBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list    = ["0.0.0.0/0"]
  gateway_arn    = "${aws_storagegateway_gateway.test.arn}"
  location_arn   = "${aws_s3_bucket.test.arn}"
  requester_pays = %t
  role_arn       = "${aws_iam_role.test.arn}"
}
`, requesterPays)
}

func testAccAWSStorageGatewayNfsFileShareConfig_Squash(rName, squash string) string {
	return testAccAWSStorageGateway_S3FileShareBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list  = ["0.0.0.0/0"]
  gateway_arn  = "${aws_storagegateway_gateway.test.arn}"
  location_arn = "${aws_s3_bucket.test.arn}"
  role_arn     = "${aws_iam_role.test.arn}"
  squash       = %q
}
`, squash)
}
