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

func TestAccAWSStorageGatewaySmbFileShare_Authentication_ActiveDirectory(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewaySmbFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_Authentication_ActiveDirectory(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestMatchResourceAttr(resourceName, "arn", regexp.MustCompile(`^arn:[^:]+:storagegateway:[^:]+:[^:]+:share/share-.+$`)),
					resource.TestCheckResourceAttr(resourceName, "authentication", "ActiveDirectory"),
					resource.TestCheckResourceAttr(resourceName, "default_storage_class", "S3_STANDARD"),
					resource.TestMatchResourceAttr(resourceName, "fileshare_id", regexp.MustCompile(`^share-`)),
					resource.TestMatchResourceAttr(resourceName, "gateway_arn", regexp.MustCompile(`^arn:`)),
					resource.TestCheckResourceAttr(resourceName, "guess_mime_type_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "invalid_user_list.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_arn", ""),
					resource.TestMatchResourceAttr(resourceName, "location_arn", regexp.MustCompile(`^arn:`)),
					resource.TestCheckResourceAttr(resourceName, "object_acl", storagegateway.ObjectACLPrivate),
					resource.TestCheckResourceAttr(resourceName, "read_only", "false"),
					resource.TestCheckResourceAttr(resourceName, "requester_pays", "false"),
					resource.TestMatchResourceAttr(resourceName, "role_arn", regexp.MustCompile(`^arn:`)),
					resource.TestCheckResourceAttr(resourceName, "valid_user_list.#", "0"),
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

func TestAccAWSStorageGatewaySmbFileShare_Authentication_GuestAccess(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewaySmbFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_Authentication_GuestAccess(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestMatchResourceAttr(resourceName, "arn", regexp.MustCompile(`^arn:[^:]+:storagegateway:[^:]+:[^:]+:share/share-.+$`)),
					resource.TestCheckResourceAttr(resourceName, "authentication", "GuestAccess"),
					resource.TestCheckResourceAttr(resourceName, "default_storage_class", "S3_STANDARD"),
					resource.TestMatchResourceAttr(resourceName, "fileshare_id", regexp.MustCompile(`^share-`)),
					resource.TestMatchResourceAttr(resourceName, "gateway_arn", regexp.MustCompile(`^arn:`)),
					resource.TestCheckResourceAttr(resourceName, "guess_mime_type_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "invalid_user_list.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_arn", ""),
					resource.TestMatchResourceAttr(resourceName, "location_arn", regexp.MustCompile(`^arn:`)),
					resource.TestCheckResourceAttr(resourceName, "object_acl", storagegateway.ObjectACLPrivate),
					resource.TestCheckResourceAttr(resourceName, "read_only", "false"),
					resource.TestCheckResourceAttr(resourceName, "requester_pays", "false"),
					resource.TestMatchResourceAttr(resourceName, "role_arn", regexp.MustCompile(`^arn:`)),
					resource.TestCheckResourceAttr(resourceName, "valid_user_list.#", "0"),
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

func TestAccAWSStorageGatewaySmbFileShare_DefaultStorageClass(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewaySmbFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_DefaultStorageClass(rName, "S3_STANDARD_IA"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "default_storage_class", "S3_STANDARD_IA"),
				),
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_DefaultStorageClass(rName, "S3_ONEZONE_IA"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
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

func TestAccAWSStorageGatewaySmbFileShare_GuessMIMETypeEnabled(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewaySmbFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_GuessMIMETypeEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "guess_mime_type_enabled", "false"),
				),
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_GuessMIMETypeEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
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

func TestAccAWSStorageGatewaySmbFileShare_InvalidUserList(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewaySmbFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_InvalidUserList_Single(rName, "invaliduser1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "invalid_user_list.#", "1"),
				),
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_InvalidUserList_Multiple(rName, "invaliduser2", "invaliduser3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "invalid_user_list.#", "2"),
				),
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_InvalidUserList_Single(rName, "invaliduser4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "invalid_user_list.#", "1"),
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

func TestAccAWSStorageGatewaySmbFileShare_KMSEncrypted(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewaySmbFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSStorageGatewaySmbFileShareConfig_KMSEncrypted(rName, true),
				ExpectError: regexp.MustCompile(`KMSKey is missing`),
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_KMSEncrypted(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
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

func TestAccAWSStorageGatewaySmbFileShare_KMSKeyArn(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewaySmbFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_KMSKeyArn(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", "true"),
					resource.TestMatchResourceAttr(resourceName, "kms_key_arn", regexp.MustCompile(`^arn:`)),
				),
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_KMSKeyArn_Update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
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
				Config: testAccAWSStorageGatewaySmbFileShareConfig_KMSEncrypted(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", "false"),
				),
			},
		},
	})
}

func TestAccAWSStorageGatewaySmbFileShare_ObjectACL(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewaySmbFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_ObjectACL(rName, storagegateway.ObjectACLPublicRead),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "object_acl", storagegateway.ObjectACLPublicRead),
				),
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_ObjectACL(rName, storagegateway.ObjectACLPublicReadWrite),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
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

func TestAccAWSStorageGatewaySmbFileShare_ReadOnly(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewaySmbFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_ReadOnly(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "read_only", "false"),
				),
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_ReadOnly(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
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

func TestAccAWSStorageGatewaySmbFileShare_RequesterPays(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewaySmbFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_RequesterPays(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "requester_pays", "false"),
				),
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_RequesterPays(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
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

func TestAccAWSStorageGatewaySmbFileShare_ValidUserList(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewaySmbFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_ValidUserList_Single(rName, "validuser1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "valid_user_list.#", "1"),
				),
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_ValidUserList_Multiple(rName, "validuser2", "validuser3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "valid_user_list.#", "2"),
				),
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_ValidUserList_Single(rName, "validuser4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "valid_user_list.#", "1"),
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

func testAccCheckAWSStorageGatewaySmbFileShareDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).storagegatewayconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_storagegateway_smb_file_share" {
			continue
		}

		input := &storagegateway.DescribeSMBFileSharesInput{
			FileShareARNList: []*string{aws.String(rs.Primary.ID)},
		}

		output, err := conn.DescribeSMBFileShares(input)

		if err != nil {
			if isAWSErr(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified file share was not found.") {
				continue
			}
			return err
		}

		if output != nil && len(output.SMBFileShareInfoList) > 0 && output.SMBFileShareInfoList[0] != nil {
			return fmt.Errorf("Storage Gateway SMB File Share %q still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName string, smbFileShare *storagegateway.SMBFileShareInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).storagegatewayconn
		input := &storagegateway.DescribeSMBFileSharesInput{
			FileShareARNList: []*string{aws.String(rs.Primary.ID)},
		}

		output, err := conn.DescribeSMBFileShares(input)

		if err != nil {
			return err
		}

		if output == nil || len(output.SMBFileShareInfoList) == 0 || output.SMBFileShareInfoList[0] == nil {
			return fmt.Errorf("Storage Gateway SMB File Share %q does not exist", rs.Primary.ID)
		}

		*smbFileShare = *output.SMBFileShareInfoList[0]

		return nil
	}
}

func testAccAWSStorageGateway_SmbFileShare_ActiveDirectoryBase(rName string) string {
	return testAccAWSStorageGatewayGatewayConfig_SmbActiveDirectorySettings(rName) + fmt.Sprintf(`
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
`, rName, rName)
}

func testAccAWSStorageGateway_SmbFileShare_GuestAccessBase(rName string) string {
	return testAccAWSStorageGatewayGatewayConfig_SmbGuestPassword(rName, "smbguestpassword") + fmt.Sprintf(`
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
`, rName, rName)
}

func testAccAWSStorageGatewaySmbFileShareConfig_Authentication_ActiveDirectory(rName string) string {
	return testAccAWSStorageGateway_SmbFileShare_ActiveDirectoryBase(rName) + `
resource "aws_storagegateway_smb_file_share" "test" {
  authentication = "ActiveDirectory"
  gateway_arn    = "${aws_storagegateway_gateway.test.arn}"
  location_arn   = "${aws_s3_bucket.test.arn}"
  role_arn       = "${aws_iam_role.test.arn}"
}
`
}

func testAccAWSStorageGatewaySmbFileShareConfig_Authentication_GuestAccess(rName string) string {
	return testAccAWSStorageGateway_SmbFileShare_GuestAccessBase(rName) + `
resource "aws_storagegateway_smb_file_share" "test" {
  authentication = "GuestAccess"
  gateway_arn    = "${aws_storagegateway_gateway.test.arn}"
  location_arn   = "${aws_s3_bucket.test.arn}"
  role_arn       = "${aws_iam_role.test.arn}"
}
`
}

func testAccAWSStorageGatewaySmbFileShareConfig_DefaultStorageClass(rName, defaultStorageClass string) string {
	return testAccAWSStorageGateway_SmbFileShare_GuestAccessBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication        = "GuestAccess"
  default_storage_class = %q
  gateway_arn           = "${aws_storagegateway_gateway.test.arn}"
  location_arn          = "${aws_s3_bucket.test.arn}"
  role_arn              = "${aws_iam_role.test.arn}"
}
`, defaultStorageClass)
}

func testAccAWSStorageGatewaySmbFileShareConfig_GuessMIMETypeEnabled(rName string, guessMimeTypeEnabled bool) string {
	return testAccAWSStorageGateway_SmbFileShare_GuestAccessBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication          = "GuestAccess"
  gateway_arn             = "${aws_storagegateway_gateway.test.arn}"
  guess_mime_type_enabled = %t
  location_arn            = "${aws_s3_bucket.test.arn}"
  role_arn                = "${aws_iam_role.test.arn}"
}
`, guessMimeTypeEnabled)
}

func testAccAWSStorageGatewaySmbFileShareConfig_InvalidUserList_Single(rName, invalidUser1 string) string {
	return testAccAWSStorageGateway_SmbFileShare_ActiveDirectoryBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Must be ActiveDirectory
  authentication    = "ActiveDirectory"
  gateway_arn       = "${aws_storagegateway_gateway.test.arn}"
  invalid_user_list = [%q]
  location_arn      = "${aws_s3_bucket.test.arn}"
  role_arn          = "${aws_iam_role.test.arn}"
}
`, invalidUser1)
}

func testAccAWSStorageGatewaySmbFileShareConfig_InvalidUserList_Multiple(rName, invalidUser1, invalidUser2 string) string {
	return testAccAWSStorageGateway_SmbFileShare_ActiveDirectoryBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Must be ActiveDirectory
  authentication    = "ActiveDirectory"
  gateway_arn       = "${aws_storagegateway_gateway.test.arn}"
  invalid_user_list = [%q, %q]
  location_arn      = "${aws_s3_bucket.test.arn}"
  role_arn          = "${aws_iam_role.test.arn}"
}
`, invalidUser1, invalidUser2)
}

func testAccAWSStorageGatewaySmbFileShareConfig_KMSEncrypted(rName string, kmsEncrypted bool) string {
	return testAccAWSStorageGateway_SmbFileShare_GuestAccessBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication = "GuestAccess"
  gateway_arn   = "${aws_storagegateway_gateway.test.arn}"
  kms_encrypted = %t
  location_arn  = "${aws_s3_bucket.test.arn}"
  role_arn      = "${aws_iam_role.test.arn}"
}
`, kmsEncrypted)
}

func testAccAWSStorageGatewaySmbFileShareConfig_KMSKeyArn(rName string) string {
	return testAccAWSStorageGateway_SmbFileShare_GuestAccessBase(rName) + `
resource "aws_kms_key" "test" {
  count = 2

  deletion_window_in_days = 7
  description             = "Terraform Acceptance Testing"
}

resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication = "GuestAccess"
  gateway_arn    = "${aws_storagegateway_gateway.test.arn}"
  kms_encrypted  = true
  kms_key_arn    = "${aws_kms_key.test.0.arn}"
  location_arn   = "${aws_s3_bucket.test.arn}"
  role_arn       = "${aws_iam_role.test.arn}"
}
`
}

func testAccAWSStorageGatewaySmbFileShareConfig_KMSKeyArn_Update(rName string) string {
	return testAccAWSStorageGateway_SmbFileShare_GuestAccessBase(rName) + `
resource "aws_kms_key" "test" {
  count = 2

  deletion_window_in_days = 7
  description             = "Terraform Acceptance Testing"
}

resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication = "GuestAccess"
  gateway_arn    = "${aws_storagegateway_gateway.test.arn}"
  kms_encrypted  = true
  kms_key_arn    = "${aws_kms_key.test.1.arn}"
  location_arn   = "${aws_s3_bucket.test.arn}"
  role_arn       = "${aws_iam_role.test.arn}"
}
`
}

func testAccAWSStorageGatewaySmbFileShareConfig_ObjectACL(rName, objectACL string) string {
	return testAccAWSStorageGateway_SmbFileShare_GuestAccessBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication = "GuestAccess"
  gateway_arn    = "${aws_storagegateway_gateway.test.arn}"
  location_arn   = "${aws_s3_bucket.test.arn}"
  object_acl     = %q
  role_arn       = "${aws_iam_role.test.arn}"
}
`, objectACL)
}

func testAccAWSStorageGatewaySmbFileShareConfig_ReadOnly(rName string, readOnly bool) string {
	return testAccAWSStorageGateway_SmbFileShare_GuestAccessBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication = "GuestAccess"
  gateway_arn    = "${aws_storagegateway_gateway.test.arn}"
  location_arn   = "${aws_s3_bucket.test.arn}"
  read_only      = %t
  role_arn       = "${aws_iam_role.test.arn}"
}
`, readOnly)
}

func testAccAWSStorageGatewaySmbFileShareConfig_RequesterPays(rName string, requesterPays bool) string {
	return testAccAWSStorageGateway_SmbFileShare_GuestAccessBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication = "GuestAccess"
  gateway_arn    = "${aws_storagegateway_gateway.test.arn}"
  location_arn   = "${aws_s3_bucket.test.arn}"
  requester_pays = %t
  role_arn       = "${aws_iam_role.test.arn}"
}
`, requesterPays)
}

func testAccAWSStorageGatewaySmbFileShareConfig_ValidUserList_Single(rName, validUser1 string) string {
	return testAccAWSStorageGateway_SmbFileShare_ActiveDirectoryBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Must be ActiveDirectory
  authentication    = "ActiveDirectory"
  gateway_arn     = "${aws_storagegateway_gateway.test.arn}"
  location_arn    = "${aws_s3_bucket.test.arn}"
  role_arn        = "${aws_iam_role.test.arn}"
  valid_user_list = [%q]
}
`, validUser1)
}

func testAccAWSStorageGatewaySmbFileShareConfig_ValidUserList_Multiple(rName, validUser1, validUser2 string) string {
	return testAccAWSStorageGateway_SmbFileShare_ActiveDirectoryBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Must be ActiveDirectory
  authentication    = "ActiveDirectory"
  gateway_arn     = "${aws_storagegateway_gateway.test.arn}"
  location_arn    = "${aws_s3_bucket.test.arn}"
  role_arn        = "${aws_iam_role.test.arn}"
  valid_user_list = [%q, %q]
}
`, validUser1, validUser2)
}
